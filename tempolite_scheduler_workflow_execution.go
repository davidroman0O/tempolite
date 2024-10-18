package tempolite

import (
	"log"
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
				log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
				continue
			}

			tp.schedulerWorkflowStarted.Store(true)

			if len(pendingWorkflows) == 0 {
				continue
			}

			var value any
			var ok bool

			// fmt.Println("pendingWorkflows: ", pendingWorkflows)

			for _, pendingWorkflowExecution := range pendingWorkflows {

				// fmt.Println("pendingWorkflowExecution: ", pendingWorkflowExecution)

				var workflowEntity *ent.Workflow
				if workflowEntity, err = tp.client.Workflow.Get(tp.ctx, pendingWorkflowExecution.Edges.Workflow.ID); err != nil {
					// todo: maybe we can tag the execution as not executable
					log.Printf("scheduler: Workflow.Get failed: %v", err)
					continue
				}

				// fmt.Println("workflowEntity: ", workflowEntity)

				if value, ok = tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
					var workflowHandlerInfo Workflow
					if workflowHandlerInfo, ok = value.(Workflow); !ok {
						// could be development bug
						log.Printf("scheduler: workflow %s is not handler info", workflowEntity.HandlerName)
						continue
					}

					inputs := []interface{}{}

					// TODO: we can probably parallelize this
					for idx, rawInput := range workflowEntity.Input {
						inputType := workflowHandlerInfo.ParamTypes[idx]
						inputKind := workflowHandlerInfo.ParamsKinds[idx]

						realInput, err := convertIO(rawInput, inputType, inputKind)
						if err != nil {
							log.Printf("scheduler: convertInput failed: %v", err)
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

						// fmt.Println("\t ==Create new workflow from", workflowEntity.HandlerName, pendingWorkflowExecution.ID)

						// create a new execution for the same workflow
						var workflowExecution *ent.WorkflowExecution
						if workflowExecution, err = tp.client.WorkflowExecution.
							Create().
							SetID(uuid.NewString()).
							SetRunID(pendingWorkflowExecution.RunID).
							SetWorkflow(workflowEntity).
							Save(tp.ctx); err != nil {
							return err
						}

						task.ctx.executionID = workflowExecution.ID
						task.retryCount++

						log.Printf("scheduler: retrying (%d) workflow %s of id %v exec id %v with params: %v", task.retryCount, workflowEntity.HandlerName, workflowEntity.ID, contextWorkflow.executionID, workflowEntity.Input)

						// now we notify the workflow enity that we're working
						if _, err = tp.client.Workflow.UpdateOneID(contextWorkflow.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
							log.Printf("scheduler: Workflow.UpdateOneID failed: %v", err)
						}

						return nil
					}

					task.retry = retryIt

					// query the count of how many workflow execution exists related to the workflowEntity
					// > but but why are you getting the count?!?!
					// well maybe if we crashed, then when re-enqueueing the workflow, we can prepare the retry count and continue our work
					total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(workflowEntity.ID))).Count(tp.ctx)
					if err != nil {
						log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
						continue
					}

					log.Printf("scheduler: total: %d", total)

					// If it's not me
					if total > 1 {
						task.retryCount = total
					}

					log.Printf("scheduler: Dispatching workflow %s of id %v exec id %v with params: %v", workflowEntity.HandlerName, workflowEntity.ID, contextWorkflow.executionID, workflowEntity.Input)

					if err := tp.workflowPool.
						Dispatch(
							task,
							retrypool.WithImmediateRetry[*workflowTask[T]](),
						); err != nil {
						log.Printf("scheduler: Dispatch failed: %v", err)
						continue
					}

					if _, err = tp.client.WorkflowExecution.UpdateOneID(pendingWorkflowExecution.ID).SetStatus(workflowexecution.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						log.Printf("scheduler: WorkflowExecution.UpdateOneID failed: %v", err)
						continue
					}

					if _, err = tp.client.Workflow.UpdateOneID(workflowEntity.ID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						log.Printf("scheduler: Workflow.UpdateOneID failed: %v", err)
						continue
					}

				} else {
					log.Printf("scheduler: Workflow %s not found", pendingWorkflowExecution.Edges.Workflow.HandlerName)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
