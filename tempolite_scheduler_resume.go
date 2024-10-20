package tempolite

import (
	"fmt"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/google/uuid"
)

// When workflows were paused, but will be flagged as IsReady
func (tp *Tempolite[T]) resumeWorkflowsWorker() {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	defer close(tp.resumeWorkflowsWorkerDone)
	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "Resume workflows worker due to context done")
			return
		case <-tp.resumeWorkflowsWorkerDone:
			tp.logger.Debug(tp.ctx, "Resume workflows worker done")
			return
		case <-ticker.C:
			workflows, err := tp.client.Workflow.Query().
				Where(
					workflow.StatusEQ(workflow.StatusRunning),
					workflow.IsPausedEQ(false),
					workflow.IsReadyEQ(true),
				).All(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Error querying workflows to resume", "error", err)
				continue
			}

			for _, wf := range workflows {
				tp.logger.Debug(tp.ctx, "Resuming workflow", "workflowID", wf.ID)
				if err := tp.redispatchWorkflow(WorkflowID(wf.ID)); err != nil {
					tp.logger.Error(tp.ctx, "Error redispatching workflow", "workflowID", wf.ID, "error", err)
				}

				tx, err := tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error starting transaction for updating workflow", "workflowID", wf.ID, "error", err)
					continue
				}
				tp.logger.Debug(tp.ctx, "Started transaction for updating workflow", "workflowID", wf.ID)

				_, err = tx.Workflow.UpdateOne(wf).SetIsReady(false).Save(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error updating workflow", "workflowID", wf.ID, "error", err)
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						tp.logger.Error(tp.ctx, "Error rolling back transaction", "workflowID", wf.ID, "error", rollbackErr)
					} else {
						tp.logger.Debug(tp.ctx, "Successfully rolled back transaction", "workflowID", wf.ID)
					}
					continue
				}

				if err := tx.Commit(); err != nil {
					tp.logger.Error(tp.ctx, "Error committing transaction for updating workflow", "workflowID", wf.ID, "error", err)
				} else {
					tp.logger.Debug(tp.ctx, "Successfully committed transaction for updating workflow", "workflowID", wf.ID)
				}
			}
		}
	}
}

func (tp *Tempolite[T]) redispatchWorkflow(id WorkflowID) error {
	// Start a transaction
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error starting transaction", "error", err)
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Ensure the transaction is rolled back in case of an error
	defer func() {
		if r := recover(); r != nil {
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction after panic", "error", rerr)
			}
			panic(r)
		}
	}()

	wf, err := tx.Workflow.Get(tp.ctx, id.String())
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Error fetching workflow", "workflowID", id, "error", err)
		return fmt.Errorf("error fetching workflow: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Redispatching workflow", "workflowID", id)
	tp.logger.Debug(tp.ctx, "Querying workflow execution", "workflowID", id)

	wfEx, err := tx.WorkflowExecution.Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).
		WithWorkflow().
		Order(ent.Desc(workflowexecution.FieldStartedAt)).
		First(tp.ctx)

	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Error querying workflow execution", "workflowID", id, "error", err)
		return fmt.Errorf("error querying workflow execution: %w", err)
	}

	// Create WorkflowContext
	ctx := WorkflowContext[T]{
		tp:           tp,
		workflowID:   wf.ID,
		executionID:  wfEx.ID,
		runID:        wfEx.RunID,
		workflowType: wf.Identity,
		stepID:       wf.StepID,
	}

	tp.logger.Debug(tp.ctx, "Getting handler info for workflow", "workflowID", id, "handlerIdentity", wf.Identity)

	// Fetch the workflow handler
	handlerInfo, ok := tp.workflows.Load(HandlerIdentity(wf.Identity))
	if !ok {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Workflow handler not found", "workflowID", id)
		return fmt.Errorf("workflow handler not found for %s", wf.Identity)
	}

	workflowHandler, ok := handlerInfo.(Workflow)
	if !ok {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Invalid workflow handler", "workflowID", id)
		return fmt.Errorf("invalid workflow handler for %s", wf.Identity)
	}

	inputs := []interface{}{}

	tp.logger.Debug(tp.ctx, "Converting inputs", "workflowID", id, "inputs", len(wf.Input))
	// TODO: we can probably parallelize this
	for idx, rawInput := range wf.Input {
		inputType := workflowHandler.ParamTypes[idx]
		inputKind := workflowHandler.ParamsKinds[idx]

		realInput, err := convertIO(rawInput, inputType, inputKind)
		if err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			tp.logger.Error(tp.ctx, "Error converting input", "workflowID", id, "error", err)
			return err
		}

		inputs = append(inputs, realInput)
	}

	tp.logger.Debug(tp.ctx, "Creating workflow task", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

	// Create and dispatch the workflow task
	task := &workflowTask[T]{
		ctx:         ctx,
		handlerName: workflowHandler.HandlerLongName,
		handler:     workflowHandler.Handler,
		params:      inputs,
		maxRetry:    wf.RetryPolicy.MaximumAttempts,
	}

	retryIt := func() error {
		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

		// Start a new transaction for the retry operation
		retryTx, err := tp.client.Tx(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error starting transaction for retry", "workflowID", id, "error", err)
			return fmt.Errorf("error starting transaction for retry: %w", err)
		}

		// create a new execution for the same workflow
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = retryTx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(wfEx.RunID).
			SetWorkflow(wf).
			Save(tp.ctx); err != nil {
			if rerr := retryTx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			tp.logger.Error(tp.ctx, "Error creating workflow execution during retry", "workflowID", id, "error", err)
			return err
		}

		task.ctx.executionID = workflowExecution.ID
		task.retryCount++
		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "retryCount", task.retryCount)

		tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "status", workflow.StatusRunning)
		// now we notify the workflow entity that we're working
		if _, err = retryTx.Workflow.UpdateOneID(ctx.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
			if rerr := retryTx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			tp.logger.Error(tp.ctx, "Error updating workflow status during retry", "workflowID", id, "error", err)
			return fmt.Errorf("error updating workflow status during retry: %w", err)
		}

		// Commit the retry transaction
		if err = retryTx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Error committing retry transaction", "workflowID", id, "error", err)
			return fmt.Errorf("error committing retry transaction: %w", err)
		}

		return nil
	}

	task.retry = retryIt

	tp.logger.Debug(tp.ctx, "Fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
	total, err := tx.WorkflowExecution.
		Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).Count(tp.ctx)
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Error fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "error", err)
		return err
	}

	tp.logger.Debug(tp.ctx, "Total previously created workflow execution", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "total", total)

	// If it's not me
	if total > 1 {
		task.retryCount = total
	}

	tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

	// Dispatch the workflow task
	if err := tp.workflowPool.Dispatch(task); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		tp.logger.Error(tp.ctx, "Error dispatching workflow task", "workflowID", id, "error", err)
		return fmt.Errorf("error dispatching workflow task: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
