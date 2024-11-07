package tempolite

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/executionrelationship"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
	"github.com/google/uuid"
)

// The resume scheduler will simply create a new execution entity and let's the workflow replay itself from the beginning.
// As per the logic of a workflow, all components should be idempotent and should be able to handle the same input multiple times, if a component is successful it will return the same output.
func (tp *Tempolite) schedulerResumeRunningWorkflows(queueName string, done chan struct{}) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "Resume running workflows scheduler stopped due to context done", "queue", queueName)
			return
		case <-done:
			tp.logger.Debug(tp.ctx, "Resume running workflows scheduler stopped due to done signal", "queue", queueName)
			return
		case <-ticker.C:
			// Get pool for this queue
			queue, ok := tp.queues.Load(queueName)
			if !ok {
				runtime.Gosched()
				continue
			}
			queueWorkers := queue.(*QueueWorkers)

			// Check available capacity
			workerIDs := queueWorkers.Workflows.GetWorkerIDs()
			availableSlots := len(workerIDs) - queueWorkers.Workflows.ProcessingCount()
			if availableSlots <= 0 {
				runtime.Gosched()
				continue
			}

			// Find workflows that are running but not ready to be processed
			workflows, err := tp.client.Workflow.Query().
				Where(
					workflow.StatusEQ(workflow.StatusPaused),
					workflow.IsPausedEQ(false),
					workflow.IsReadyEQ(true),
					workflow.QueueNameEQ(queueName),
				).
				Limit(availableSlots).
				All(tp.ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					tp.logger.Debug(tp.ctx, "scheduler resume running workflow execution: context canceled", "queue", queueName)
					return
				}
				tp.logger.Error(tp.ctx, "Error querying workflows", "queue", queueName, "error", err)
				continue
			}

			if len(workflows) > 0 {
				fmt.Println("Resuming", len(workflows), "workflows")
				fmt.Println(workflows)
			}

			for _, wf := range workflows {
				// Start transaction
				tx, err := tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error starting transaction", "workflowID", wf.ID, "error", err)
					continue
				}

				// Set workflow as ready
				_, err = tx.Workflow.UpdateOneID(wf.ID).
					SetIsReady(false).
					SetStatus(workflow.StatusRunning).
					Save(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error updating workflow ready state", "workflowID", wf.ID, "error", err)
					if rerr := tx.Rollback(); rerr != nil {
						tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
					}
					continue
				}

				// Create a new workflow execution
				// That's where we will let the determinism doing its thing
				wfExec, err := tx.WorkflowExecution.Create().
					SetID(uuid.NewString()).
					SetRunID(wf.ID).
					SetStatus(workflowexecution.StatusRunning). // VERY IMPORTANT otherwise the other scheduler for new root workflows will take it
					SetWorkflow(wf).
					SetQueueName(queueName).
					Save(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error creating workflow execution", "workflowID", wf.ID, "error", err)
					continue
				}

				// Create workflow task
				value, ok := tp.workflows.Load(HandlerIdentity(wf.Identity))
				if !ok {
					tp.logger.Error(tp.ctx, "Workflow handler not found", "workflowID", wf.ID)
					continue
				}

				workflowHandler, ok := value.(Workflow)
				if !ok {
					tp.logger.Error(tp.ctx, "Invalid workflow handler", "workflowID", wf.ID)
					continue
				}

				inputs, err := tp.convertInputsFromSerialization(HandlerInfo(workflowHandler), wf.Input)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error converting inputs", "workflowID", wf.ID, "error", err)
					continue
				}

				execRelation, err := tx.ExecutionRelationship.Query().
					Where(
						executionrelationship.And(
							executionrelationship.ParentEntityIDEQ(wf.ID),
						),
					).
					First(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error querying execution relationship", "workflowID", wf.ID, "error", err)
					if rerr := tx.Rollback(); rerr != nil {
						tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
					}
					continue
				}

				contextWorkflow := WorkflowContext{
					tp:              tp,
					workflowID:      wf.ID,
					executionID:     wfExec.ID,
					runID:           execRelation.RunID,
					workflowType:    wf.Identity,
					stepID:          wf.StepID,
					handlerIdentity: HandlerIdentity(wf.Identity),
					queueName:       queueName,
				}

				task := &workflowTask{
					ctx:         contextWorkflow,
					handlerName: workflowHandler.HandlerLongName,
					handler:     workflowHandler.Handler,
					params:      inputs,
					maxRetry:    wf.RetryPolicy.MaximumAttempts,
					queueName:   queueName,
				}

				retryIt := func() error {
					retryTX, err := tp.client.Tx(tp.ctx)
					if err != nil {
						return err
					}
					var workflowExecution *ent.WorkflowExecution
					if workflowExecution, err = retryTX.WorkflowExecution.
						Create().
						SetID(uuid.NewString()).
						SetRunID(wfExec.RunID).
						SetWorkflow(wf).
						Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: retry create new execution failed", "error", err)
						return retryTX.Rollback()
					}

					tp.logger.Debug(tp.ctx, "Scheduler workflow execution: retrying", "workflow_id", wfExec.Edges.Workflow.ID, "workflow_execution_id", workflowExecution.ID)

					task.ctx.executionID = workflowExecution.ID
					task.retryCount++

					tp.logger.Debug(tp.ctx, "Scheduler workflow execution: retrying", "workflow_id", wfExec.Edges.Workflow.ID, "workflow_execution_id", workflowExecution.ID, "retry_count", task.retryCount)

					if _, err = retryTX.Workflow.UpdateOneID(contextWorkflow.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: retry update workflow status failed", "error", err)
						return retryTX.Rollback()
					}

					return retryTX.Commit()
				}

				task.retry = retryIt

				total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).Count(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: WorkflowExecution.Query total failed", "error", err)
					continue
				}

				if total > 1 {
					task.retryCount = total
				}

				whenBeingDispatched := retrypool.NewProcessedNotification()

				opts := []retrypool.TaskOption[*workflowTask]{
					retrypool.WithImmediateRetry[*workflowTask](),
					retrypool.WithBeingProcessed[*workflowTask](whenBeingDispatched),
				}

				if wf.MaxDuration != "" {
					d, err := time.ParseDuration(wf.MaxDuration)
					if err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to parse max duration", "error", err)
						continue
					}
					opts = append(opts, retrypool.WithTimeLimit[*workflowTask](d))
				}

				if err := queueWorkers.Workflows.Submit(
					task,
					opts...,
				); err != nil {
					tp.logger.Error(tp.ctx, "Error dispatching workflow task", "workflowID", wf.ID, "error", err)
					tx, txErr := tp.client.Tx(tp.ctx)
					if txErr != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to start transaction", "error", txErr)
						continue
					}
					if _, err = tx.WorkflowExecution.UpdateOneID(wfExec.ID).SetStatus(workflowexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: WorkflowExecution.UpdateOneID failed when failed to dispatch", "error", err)
						if rollbackErr := tx.Rollback(); rollbackErr != nil {
							tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to rollback transaction", "error", rollbackErr)
						}
						continue
					}
					if _, err = tx.Workflow.UpdateOneID(wf.ID).SetStatus(workflow.StatusFailed).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Workflow.UpdateOneID failed when failed to dispatch", "error", err)
						if rollbackErr := tx.Rollback(); rollbackErr != nil {
							tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to rollback transaction", "error", rollbackErr)
						}
						continue
					}
					if err = tx.Commit(); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to commit transaction", "error", err)
					}
					continue
				}

				<-whenBeingDispatched

				tx, err = tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to start transaction", "error", err)
					continue
				}
				if _, err = tx.WorkflowExecution.UpdateOneID(wfExec.ID).SetStatus(workflowexecution.StatusRunning).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: WorkflowExecution.UpdateOneID failed", "error", err)
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to rollback transaction", "error", rollbackErr)
					}
					continue
				}
				if _, err = tx.Workflow.UpdateOneID(wf.ID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: Workflow.UpdateOneID failed", "error", err)
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to rollback transaction", "error", rollbackErr)
					}
					continue
				}

				if err = tx.Commit(); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: Failed to commit transaction", "error", err)
				}

				tp.logger.Debug(tp.ctx, "Successfully resumed workflow", "workflowID", wf.ID)
			}

			runtime.Gosched()
		}
	}
}
