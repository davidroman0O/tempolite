package tempolite

import (
	"fmt"
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/google/uuid"
)

func (tp *Tempolite) schedulerExeutionWorkflow() {
	defer close(tp.schedulerExecutionWorkflowDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			pendingWorkflows, err := tp.client.WorkflowExecution.Query().
				Where(workflowexecution.StatusEQ(workflowexecution.StatusPending)).
				Order(ent.Asc(workflowexecution.FieldStartedAt)).WithWorkflow().
				Limit(1).All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
				continue
			}

			if len(pendingWorkflows) == 0 {
				continue
			}

			var value any
			var ok bool

			for _, wkflw := range pendingWorkflows {

				var workflowEntity *ent.Workflow
				if workflowEntity, err = tp.client.Workflow.Get(tp.ctx, wkflw.Edges.Workflow.ID); err != nil {
					// todo: maybe we can tag the execution as not executable
					log.Printf("scheduler: Workflow.Get failed: %v", err)
					continue
				}

				fmt.Println("workflowEntity: ", workflowEntity)

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

						realInput, err := convertInput(rawInput, inputType, inputKind)
						if err != nil {
							log.Printf("scheduler: convertInput failed: %v", err)
							continue
						}

						inputs = append(inputs, realInput)
					}

					task := &workflowTask{
						executionID: wkflw.ID,
						handlerName: workflowHandlerInfo.HandlerLongName,
						handler:     workflowHandlerInfo.Handler,
						params:      inputs,
						maxRetry:    workflowEntity.RetryPolicy.MaximumAttempts,
						retryCount:  0,
					}

					retryIt := func() error {
						var runEntity *ent.Run
						if runEntity, err = tp.client.Run.Get(tp.ctx, wkflw.RunID); err != nil {
							return err
						}
						// create a new execution for the same workflow
						var workflowExecution *ent.WorkflowExecution
						if workflowExecution, err = tp.client.WorkflowExecution.
							Create().
							SetID(uuid.NewString()).
							SetRunID(runEntity.RunID).
							SetWorkflow(workflowEntity).
							Save(tp.ctx); err != nil {
							return err
						}
						task.executionID = workflowExecution.ID
						task.retryCount++

						if err := tp.workflowPool.Dispatch(task); err != nil {
							log.Printf("scheduler: Dispatch failed: %v", err)
							return err
						}

						return nil
					}

					task.retry = retryIt

					// // query the count of how many workflow execution exists related to the workflowEntity
					// total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(workflowEntity.ID))).Count(tp.ctx)
					// if err != nil {
					// 	log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
					// 	continue
					// }

					// task.retryCount = total

					log.Printf("scheduler: Dispatching workflow %s with params: %v", workflowEntity.HandlerName, workflowEntity.Input)
					if err := tp.workflowPool.Dispatch(task); err != nil {
						log.Printf("scheduler: Dispatch failed: %v", err)
						continue
					}

					if _, err = tp.client.WorkflowExecution.UpdateOneID(wkflw.ID).SetStatus(workflowexecution.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						log.Printf("scheduler: WorkflowExecution.UpdateOneID failed: %v", err)
						continue
					}

				} else {
					log.Printf("scheduler: Workflow %s not found", wkflw.Edges.Workflow.HandlerName)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
