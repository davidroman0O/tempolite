package tempolite

import (
	"fmt"
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
)

func (tp *Tempolite) schedulerExecutionActivity() {
	defer close(tp.schedulerExecutionActivityDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			pendingActivities, err := tp.client.ActivityExecution.Query().
				Where(activityexecution.StatusEQ(activityexecution.StatusPending)).
				Order(ent.Asc(activityexecution.FieldStartedAt)).WithActivity().
				Limit(1).All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: ActivityExecution.Query failed: %v", err)
				continue
			}

			if len(pendingActivities) == 0 {
				continue
			}

			var value any
			var ok bool

			for _, act := range pendingActivities {

				var activityEntity *ent.Activity
				if activityEntity, err = tp.client.Activity.Get(tp.ctx, act.Edges.Activity.ID); err != nil {

					log.Printf("scheduler: Activity.Get failed: %v", err)
					continue
				}

				if value, ok = tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
					var activityHandlerInfo Activity
					if activityHandlerInfo, ok = value.(Activity); !ok {
						// could be development bug
						log.Printf("scheduler: activity %s is not handler info", activityEntity.HandlerName)
						continue
					}

					inputs := []interface{}{}

					for idx, rawInput := range activityEntity.Input {
						inputType := activityHandlerInfo.ParamTypes[idx]

						inputKind := activityHandlerInfo.ParamsKinds[idx]

						realInput, err := convertIO(rawInput, inputType, inputKind)
						if err != nil {
							log.Printf("scheduler: convertInput failed: %v", err)
							continue
						}

						inputs = append(inputs, realInput)
					}

					contextActivity := ActivityContext{
						tp:          tp,
						activityID:  activityEntity.ID,
						executionID: act.ID,
						runID:       act.RunID,
					}

					task := &activityTask{
						ctx:         contextActivity,
						handlerName: activityHandlerInfo.HandlerLongName,
						handler:     activityHandlerInfo.Handler,
						params:      inputs,
						maxRetry:    activityEntity.RetryPolicy.MaximumAttempts,
						retryCount:  0,
					}

					retryIt := func() error {

						// create a new execution for the same activity
						var activityExecution *ent.ActivityExecution
						if activityExecution, err = tp.client.ActivityExecution.
							Create().
							SetID(uuid.NewString()).
							SetRunID(act.RunID).
							SetActivity(activityEntity).
							Save(tp.ctx); err != nil {
							log.Printf("ERROR scheduler: ActivityExecution.Create failed: %v", err)
							return err
						}

						// update the current execution id
						task.ctx.executionID = activityExecution.ID
						task.retryCount++

						return nil
					}

					task.retry = retryIt

					// query the count of how many activity execution exists related to the activityEntity
					// > but but why are you getting the count?!?!
					// well maybe if we crashed, then when re-enqueueing the activity, we can prepare the retry count and continue our work
					total, err := tp.client.ActivityExecution.Query().Where(activityexecution.HasActivityWith(activity.IDEQ(activityEntity.ID))).Count(tp.ctx)
					if err != nil {
						log.Printf("scheduler: ActivityExecution.Query failed: %v", err)
						continue
					}

					log.Printf("scheduler: total: %d", total)

					// If it's not me
					if total > 1 {
						task.retryCount = total
					}

					log.Printf("scheduler: Dispatching activity %s with params: %v", activityEntity.HandlerName, activityEntity.Input)

					if err := tp.activityPool.Dispatch(task); err != nil {
						log.Printf("scheduler: Dispatch failed: %v", err)
						continue
					}

					if _, err = tp.client.ActivityExecution.UpdateOneID(act.ID).SetStatus(activityexecution.StatusRunning).Save(tp.ctx); err != nil {
						log.Printf("scheduler: ActivityExecution.UpdateOneID failed: %v", err)
						continue
					}

				} else {
					log.Printf("scheduler: Activity %s not found", act.Edges.Activity.HandlerName)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}

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
				Limit(1).
				All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
				continue
			}

			if len(pendingWorkflows) == 0 {
				continue
			}

			var value any
			var ok bool

			for _, pendingWorkflowExecution := range pendingWorkflows {

				var workflowEntity *ent.Workflow
				if workflowEntity, err = tp.client.Workflow.Get(tp.ctx, pendingWorkflowExecution.Edges.Workflow.ID); err != nil {
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

						realInput, err := convertIO(rawInput, inputType, inputKind)
						if err != nil {
							log.Printf("scheduler: convertInput failed: %v", err)
							continue
						}

						inputs = append(inputs, realInput)
					}

					contextWorkflow := WorkflowContext{
						tp:           tp,
						workflowID:   workflowEntity.ID,
						executionID:  pendingWorkflowExecution.ID,
						runID:        pendingWorkflowExecution.RunID,
						workflowType: workflowEntity.Identity,
					}

					task := &workflowTask{
						ctx:         contextWorkflow,
						handlerName: workflowHandlerInfo.HandlerLongName,
						handler:     workflowHandlerInfo.Handler,
						params:      inputs,
						maxRetry:    workflowEntity.RetryPolicy.MaximumAttempts,
						retryCount:  0,
					}

					// On retry, we will have to create a new workflow exection
					retryIt := func() error {

						fmt.Println("\t ==Create new workflow from", workflowEntity.HandlerName, pendingWorkflowExecution.ID)

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

						// now we notify the workflow enity that we're working
						if _, err = tp.client.Workflow.UpdateOneID(contextWorkflow.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
							log.Printf("scheduler: Workflow.UpdateOneID failed: %v", err)
						}

						return nil
					}

					task.retry = retryIt

					// // query the count of how many workflow execution exists related to the workflowEntity
					// // > but but why are you getting the count?!?!
					// // well maybe if we crashed, then when re-enqueueing the workflow, we can prepare the retry count and continue our work
					// total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(workflowEntity.ID))).Count(tp.ctx)
					// if err != nil {
					// 	log.Printf("scheduler: WorkflowExecution.Query failed: %v", err)
					// 	continue
					// }

					// log.Printf("scheduler: total: %d", total)

					// // If it's not me
					// if total > 1 {
					// 	task.retryCount = total
					// }

					log.Printf("scheduler: Dispatching workflow %s of id %v exec id %v with params: %v", workflowEntity.HandlerName, workflowEntity.ID, contextWorkflow.executionID, workflowEntity.Input)

					if err := tp.workflowPool.
						Dispatch(
							task,
							retrypool.WithImmediateRetry[*workflowTask](),
						); err != nil {
						log.Printf("scheduler: Dispatch failed: %v", err)
						continue
					}

					if _, err = tp.client.WorkflowExecution.UpdateOneID(pendingWorkflowExecution.ID).SetStatus(workflowexecution.StatusRunning).Save(tp.ctx); err != nil {
						// TODO: could be a problem if not really dispatched
						log.Printf("scheduler: WorkflowExecution.UpdateOneID failed: %v", err)
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
