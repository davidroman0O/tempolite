package tempolite

import (
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) schedulerExecutionWorkflow() {
	defer close(tp.schedulerExecutionWorkflowDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			pendingWorkflows, err := tp.client.WorkflowExecution.Query().
				Where(workflowexecution.StatusEQ(workflowexecution.StatusPending)).
				Order(ent.Asc(workflowexecution.FieldStartedAt)).
				WithWorkflow().
				Limit(1).
				All(tp.ctx)
			if err != nil {
				continue
			}

			tp.schedulerWorkflowStarted.Store(true)

			if len(pendingWorkflows) == 0 {
				continue
			}

			var value any
			var ok bool

			for _, pendingWorkflowExecution := range pendingWorkflows {

				tp.logger.Debug(tp.ctx, "Scheduler workflow execution: pending workflow", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID, "workflow_execution_id", pendingWorkflowExecution.ID)

				var workflowEntity *ent.Workflow
				if workflowEntity, err = tp.client.Workflow.Get(tp.ctx, pendingWorkflowExecution.Edges.Workflow.ID); err != nil {
					// todo: maybe we can tag the execution as not executable
					tp.logger.Error(tp.ctx, "Scheduler workflow execution: workflow.Get failed", "error", err)
					continue
				}

				if value, ok = tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
					var workflowHandlerInfo Workflow
					if workflowHandlerInfo, ok = value.(Workflow); !ok {
						// could be development bug
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: workflow handler info not found", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID)
						continue
					}

					inputs := []interface{}{}

					// TODO: we can probably parallelize this
					for idx, rawInput := range workflowEntity.Input {
						inputType := workflowHandlerInfo.ParamTypes[idx]
						inputKind := workflowHandlerInfo.ParamsKinds[idx]

						realInput, err := convertIO(rawInput, inputType, inputKind)
						if err != nil {
							tp.logger.Error(tp.ctx, "Scheduler workflow execution: convertInput failed", "error", err)
							continue
						}

						inputs = append(inputs, realInput)
					}

					contextWorkflow := WorkflowContext[T]{
						tp:           tp,
						workflowID:   workflowEntity.ID,
						executionID:  pendingWorkflowExecution.ID,
						runID:        pendingWorkflowExecution.RunID,
						workflowType: workflowEntity.Identity,
						stepID:       workflowEntity.StepID,
					}

					task := &workflowTask[T]{
						ctx:         contextWorkflow,
						handlerName: workflowHandlerInfo.HandlerLongName,
						handler:     workflowHandlerInfo.Handler,
						params:      inputs,
						maxRetry:    workflowEntity.RetryPolicy.MaximumAttempts,
						retryCount:  0,
					}

					// On retry, we will have to create a new workflow exection
					retryIt := func() error {

						// create a new execution for the same workflow
						var workflowExecution *ent.WorkflowExecution
						if workflowExecution, err = tp.client.WorkflowExecution.
							Create().
							SetID(uuid.NewString()).
							SetRunID(pendingWorkflowExecution.RunID).
							SetWorkflow(workflowEntity).
							Save(tp.ctx); err != nil {
							tp.logger.Error(tp.ctx, "Scheduler workflow execution: retry create new execution failed", "error", err)
							return err
						}

						tp.logger.Debug(tp.ctx, "Scheduler workflow execution: retrying", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID, "workflow_execution_id", workflowExecution.ID)

						task.ctx.executionID = workflowExecution.ID
						task.retryCount++

						tp.logger.Debug(tp.ctx, "Scheduler workflow execution: retrying", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID, "workflow_execution_id", workflowExecution.ID, "retry_count", task.retryCount)

						// now we notify the workflow enity that we're working
						if _, err = tp.client.Workflow.UpdateOneID(contextWorkflow.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
							tp.logger.Error(tp.ctx, "Scheduler workflow execution: retry update workflow status failed", "error", err)
							return err
						}

						return nil
					}

					task.retry = retryIt

					// query the count of how many workflow execution exists related to the workflowEntity
					// > but but why are you getting the count?!?!
					// well maybe if we crashed, then when re-enqueueing the workflow, we can prepare the retry count and continue our work
					total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(workflowEntity.ID))).Count(tp.ctx)
					if err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: WorkflowExecution.Query total failed", "error", err)
						continue
					}

					// If it's not me
					if total > 1 {
						task.retryCount = total
					}

					tp.logger.Debug(tp.ctx, "Scheduler workflow execution: Dispatching", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID, "workflow_execution_id", pendingWorkflowExecution.ID)

					if err := tp.workflowPool.
						Dispatch(
							task,
							retrypool.WithImmediateRetry[*workflowTask[T]](),
						); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Dispatch failed", "error", err)
						continue
					}

					if _, err = tp.client.WorkflowExecution.UpdateOneID(pendingWorkflowExecution.ID).SetStatus(workflowexecution.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: WorkflowExecution.UpdateOneID failed", "error", err)
						continue
					}

					if _, err = tp.client.Workflow.UpdateOneID(workflowEntity.ID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: Workflow.UpdateOneID failed", "error", err)
						continue
					}

				} else {
					tp.logger.Error(tp.ctx, "Workflow not found", "workflow_id", pendingWorkflowExecution.Edges.Workflow.ID)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
