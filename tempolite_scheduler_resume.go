package tempolite

import (
	"fmt"
	"log"
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
				if _, err := tp.client.Workflow.UpdateOne(wf).SetIsReady(false).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Error updating workflow", "workflowID", wf.ID, "error", err)
				}
			}
		}
	}
}

func (tp *Tempolite[T]) redispatchWorkflow(id WorkflowID) error {

	wf, err := tp.client.Workflow.Get(tp.ctx, id.String())
	if err != nil {
		tp.logger.Error(tp.ctx, "Error fetching workflow", "workflowID", id, "error", err)
		return fmt.Errorf("error fetching workflow: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Redispatching workflow", "workflowID", id)
	tp.logger.Debug(tp.ctx, "Querying workflow execution", "workflowID", id)

	wfEx, err := tp.client.WorkflowExecution.Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).
		WithWorkflow().
		Order(ent.Desc(workflowexecution.FieldStartedAt)).
		First(tp.ctx)

	if err != nil {
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
		tp.logger.Error(tp.ctx, "Workflow handler not found", "workflowID", id)
		return fmt.Errorf("workflow handler not found for %s", wf.Identity)
	}

	workflowHandler, ok := handlerInfo.(Workflow)
	if !ok {
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

		// create a new execution for the same workflow
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tp.client.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(wfEx.RunID).
			SetWorkflow(wf).
			Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating workflow execution during retry", "workflowID", id, "error", err)
			return err
		}

		task.ctx.executionID = workflowExecution.ID
		task.retryCount++
		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "retryCount", task.retryCount)

		tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "status", workflow.StatusRunning)
		// now we notify the workflow enity that we're working
		if _, err = tp.client.Workflow.UpdateOneID(ctx.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
			log.Printf("scheduler: Workflow.UpdateOneID failed: %v", err)
		}

		return nil
	}

	task.retry = retryIt

	tp.logger.Debug(tp.ctx, "Fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
	total, err := tp.client.WorkflowExecution.
		Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).Count(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "error", err)
		return err
	}

	tp.logger.Debug(tp.ctx, "Total previously created workflow execution", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "total", total)

	// If it's not me
	if total > 1 {
		task.retryCount = total
	}

	tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
	return tp.workflowPool.Dispatch(task)
}
