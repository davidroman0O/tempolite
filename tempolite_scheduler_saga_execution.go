package tempolite

import (
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/google/uuid"
)

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

							_, err := tp.client.Saga.UpdateOne(sagaExecution.Edges.Saga).
								SetStatus(saga.StatusCompensated).
								Save(tp.ctx)
							// No compensation needed if no transactions succeeded
							return err
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
