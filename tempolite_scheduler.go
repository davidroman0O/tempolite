package tempolite

import (
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) schedulerExecutionActivity() {
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

			tp.schedulerActivityStarted.Store(true)

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

					contextActivity := ActivityContext[T]{
						tp:          tp,
						activityID:  activityEntity.ID,
						executionID: act.ID,
						runID:       act.RunID,
						stepID:      activityEntity.StepID,
					}

					task := &activityTask[T]{
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

func (tp *Tempolite[T]) schedulerExecutionSideEffect() {
	defer close(tp.schedulerSideEffectDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:
			pendingSideEffects, err := tp.client.SideEffectExecution.Query().
				Where(sideeffectexecution.StatusEQ(sideeffectexecution.StatusPending)).
				Order(ent.Asc(sideeffectexecution.FieldStartedAt)).WithSideEffect().
				WithSideEffect().
				Limit(1).All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: SideEffectExecution.Query failed: %v", err)
				continue
			}

			tp.schedulerSideEffectStarted.Store(true)

			if len(pendingSideEffects) == 0 {
				continue
			}

			for _, se := range pendingSideEffects {
				sideEffectInfo, ok := tp.sideEffects.Load(se.Edges.SideEffect.ID)
				if !ok {
					log.Printf("scheduler: SideEffect %s not found", se.Edges.SideEffect.ID)
					continue
				}

				sideEffect := sideEffectInfo.(SideEffect)

				contextSideEffect := SideEffectContext[T]{
					tp:           tp,
					sideEffectID: se.Edges.SideEffect.ID,
					executionID:  se.ID,
					// runID:        se.RunID,
					stepID: se.Edges.SideEffect.StepID,
				}

				task := &sideEffectTask[T]{
					ctx:         contextSideEffect,
					handlerName: sideEffect.HandlerLongName,
					handler:     sideEffect.Handler,
				}

				log.Printf("scheduler: Dispatching side effect %s", se.Edges.SideEffect.HandlerName)

				if err := tp.sideEffectPool.Dispatch(task); err != nil {
					log.Printf("scheduler: Dispatch failed: %v", err)
					continue
				}

				if _, err = tp.client.SideEffectExecution.UpdateOneID(se.ID).SetStatus(sideeffectexecution.StatusRunning).Save(tp.ctx); err != nil {
					log.Printf("scheduler: SideEffectExecution.UpdateOneID failed: %v", err)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}

// We want to make a chain reaction of transactions and compensations
//
// T1 --next--> T2 --next--> T3
//
//	  |             |
//	compensate     compensate
//	  |             |
//	  v             v
//	 C1 --next-->  C2 --next--> C1
//
// # Which mean
//
// Scenario 1: All Transactions Succeed
// T1 --next--> T2 --next--> T3
// (No compensations needed)
//
// Scenario 2: Failure at T3
// T1 --next--> T2 --next--> T3
//
//	     |
//	 [Failure]
//	     |
//	C2 <-- C1
//
// Scenario 3: Failure at T2
// T1 --next--> T2
//
//	    |
//	[Failure]
//	    |
//	   C1
//
// Scenario 4: Failure at T1
// T1
// |
// [Failure]
// (No compensations, since no successful transactions)
//
// # So the complete flow would look like that here
//
// T1 --next--> T2 --next--> T3
//
//	|             |             |
//
// compensate    compensate    compensate
//
//	 |             |             |
//	 v             v             v
//	C1 <--next-- C2 <--next-- C3
func (tp *Tempolite[T]) schedulerExecutionSaga() {
	defer close(tp.schedulerSagaDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:
			pendingSagas, err := tp.client.SagaExecution.Query().
				Where(sagaexecution.StatusEQ(sagaexecution.StatusPending)).
				Order(ent.Asc(sagaexecution.FieldStartedAt)).
				WithSaga().
				Limit(1).
				All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: Saga.Query failed: %v", err)
				continue
			}

			tp.schedulerSagaStarted.Store(true)

			for _, sagaExecution := range pendingSagas {
				sagaHandlerInfo, ok := tp.sagas.Load(sagaExecution.Edges.Saga.ID)
				if !ok {
					log.Printf("scheduler: SagaHandlerInfo not found for ID: %s", sagaExecution.Edges.Saga.ID)
					continue
				}

				sagaDef := sagaHandlerInfo.(*SagaDefinition[T])
				transactionTasks := make([]*transactionTask[T], len(sagaDef.HandlerInfo.TransactionInfo))
				compensationTasks := make([]*compensationTask[T], len(sagaDef.HandlerInfo.CompensationInfo))

				// Prepare all the transactions and compensation to be orchestrated in a chain reaction
				{

					var lastSuccessfulIndex = -1 // Initialize with -1 indicating no transactions have succeeded yet

					// Prepare and link tasks
					for i := 0; i < len(sagaDef.Steps); i++ {
						transactionTasks[i] = &transactionTask[T]{
							ctx:         TransactionContext[T]{},
							sagaID:      sagaExecution.Edges.Saga.ID,
							executionID: sagaExecution.ID,
							stepIndex:   i,
							handlerName: sagaDef.HandlerInfo.TransactionInfo[i].HandlerName,
							isLast:      i == len(sagaDef.Steps)-1,
						}

						// Create compensation tasks for all steps
						compensationTasks[i] = &compensationTask[T]{
							ctx:         CompensationContext[T]{},
							sagaID:      sagaExecution.Edges.Saga.ID,
							executionID: sagaExecution.ID,
							stepIndex:   i,
							handlerName: sagaDef.HandlerInfo.CompensationInfo[i].HandlerName,
							isLast:      i == 0,
						}

						// Link transaction to next transaction
						if i < len(sagaDef.Steps)-1 {
							nextIndex := i + 1
							transactionTasks[i].next = func() error {
								log.Printf("scheduler: Dispatching next transaction task: %s", transactionTasks[nextIndex].handlerName)
								var transactionExecution *ent.SagaExecution
								if transactionExecution, err = tp.client.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeTransaction).
									SetHandlerName(sagaDef.HandlerInfo.TransactionInfo[nextIndex].HandlerName).
									SetSequence(nextIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									log.Printf("scheduler: Failed to create next transaction task: %v", err)
									return err
								}

								transactionTasks[nextIndex].executionID = transactionExecution.ID

								// Before moving to the next transaction, update the last successful index
								lastSuccessfulIndex = i
								return tp.transactionPool.Dispatch(transactionTasks[nextIndex])
							}
						} else {
							// For the last transaction
							transactionTasks[i].next = func() error {
								// Update last successful index as this is the last transaction
								lastSuccessfulIndex = i
								return nil // No next transaction
							}
						}

						// Link transaction to its compensation (will adjust this later)
					}

					// Adjust the compensate function after setting up the tasks
					for i := 0; i < len(sagaDef.Steps); i++ {
						transactionTasks[i].compensate = func() error {
							if lastSuccessfulIndex >= 0 {

								var compensationExecution *ent.SagaExecution
								if compensationExecution, err = tp.client.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeCompensation).
									SetHandlerName(sagaDef.HandlerInfo.CompensationInfo[lastSuccessfulIndex].HandlerName).
									SetSequence(lastSuccessfulIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									log.Printf("scheduler: Failed to create compensation task: %v", err)
									return err
								}

								compensationTasks[lastSuccessfulIndex].executionID = compensationExecution.ID

								// Start compensation from the last successful transaction
								return tp.compensationPool.Dispatch(compensationTasks[lastSuccessfulIndex])
							}
							// No compensation needed if no transactions succeeded
							return nil
						}
					}

					// Link compensation tasks in reverse order
					for i := len(sagaDef.Steps) - 1; i >= 0; i-- {
						if i > 0 {
							prevIndex := i - 1
							compensationTasks[i].next = func() error {

								var compensationExecution *ent.SagaExecution
								if compensationExecution, err = tp.client.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeCompensation).
									SetHandlerName(sagaDef.HandlerInfo.CompensationInfo[prevIndex].HandlerName).
									SetSequence(prevIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									log.Printf("scheduler: Failed to create next compensation task: %v", err)
									return err
								}

								compensationTasks[prevIndex].executionID = compensationExecution.ID

								return tp.compensationPool.Dispatch(compensationTasks[prevIndex])
							}
						} else {
							// Optionally, for the first compensation task, you might decide to loop back or end the chain
							compensationTasks[i].next = nil
						}
					}

				}

				// Dispatch the first transaction task
				if err := tp.transactionPool.Dispatch(transactionTasks[0]); err != nil {
					log.Printf("scheduler: Failed to dispatch first transaction task: %v", err)
					continue
				}

				if _, err := tp.client.Saga.UpdateOne(sagaExecution.Edges.Saga).
					SetStatus(saga.StatusRunning).
					Save(tp.ctx); err != nil {
					log.Printf("scheduler: Failed to update saga status: %v", err)
				}
			}

			runtime.Gosched()
		}
	}
}
