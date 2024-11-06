package tempolite

import (
	"context"
	"errors"
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/sagaexecution"
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
func (tp *Tempolite) schedulerExecutionSagaForQueue(queueName string, done chan struct{}) {

	queueTransaction, err := tp.getTransactionPoolQueue(queueName)
	if err != nil {
		tp.logger.Error(tp.ctx, "Scheduler transaction execution: getTransactionPoolQueue failed", "error", err)
		return
	}

	queueCompensation, err := tp.getCompensationPoolQueue(queueName)
	if err != nil {
		tp.logger.Error(tp.ctx, "Scheduler compensation execution: getCompensationPoolQueue failed", "error", err)
		return
	}

	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "scheduler saga execution: context done", "queue", queueName)
			return
		case <-done:
			tp.logger.Debug(tp.ctx, "scheduler saga execution: done signal", "queue", queueName)
			return
		default:

			queueWorkersRaw, ok := tp.queues.Load(queueName)
			if !ok {
				continue
			}
			queueWorkers := queueWorkersRaw.(*QueueWorkers)

			// For transactions
			txWorkerIDs := queueWorkers.Transactions.GetWorkerIDs()
			txAvailableSlots := len(txWorkerIDs) - queueWorkers.Transactions.ProcessingCount()
			if txAvailableSlots <= 0 {
				runtime.Gosched()
				continue
			}

			// For compensations
			compWorkerIDs := queueWorkers.Compensations.GetWorkerIDs()
			compAvailableSlots := len(compWorkerIDs) - queueWorkers.Compensations.ProcessingCount()
			if compAvailableSlots <= 0 {
				runtime.Gosched()
				continue
			}

			pendingSagas, err := tp.client.SagaExecution.Query().
				Where(
					sagaexecution.StatusEQ(sagaexecution.StatusPending),
					sagaexecution.HasSagaWith(saga.QueueNameEQ(queueName)),
				).
				Order(ent.Asc(sagaexecution.FieldStartedAt)).
				WithSaga().
				Limit(min(txAvailableSlots, compAvailableSlots)).
				All(tp.ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					tp.logger.Debug(tp.ctx, "scheduler saga execution: context canceled", "queue", queueName)
					return
				}
				tp.logger.Error(tp.ctx, "Scheduler saga execution: SagaExecution.Query failed", "error", err)
				continue
			}

			for _, sagaExecution := range pendingSagas {
				sagaHandlerInfo, ok := tp.sagas.Load(sagaExecution.Edges.Saga.ID)
				if !ok {
					tp.logger.Error(tp.ctx, "Scheduler saga execution: SagaHandlerInfo not found", "sagaID", sagaExecution.Edges.Saga.ID)
					continue
				}

				sagaDef := sagaHandlerInfo.(*SagaDefinition)
				transactionTasks := make([]*transactionTask, len(sagaDef.HandlerInfo.TransactionInfo))
				compensationTasks := make([]*compensationTask, len(sagaDef.HandlerInfo.CompensationInfo))

				// Prepare and link tasks
				{
					var lastSuccessfulIndex = -1 // Initialize with -1 indicating no transactions have succeeded yet

					// Prepare and link tasks
					for i := 0; i < len(sagaDef.Steps); i++ {
						transactionTasks[i] = &transactionTask{
							ctx:         TransactionContext{},
							sagaID:      sagaExecution.Edges.Saga.ID,
							executionID: sagaExecution.ID,
							stepIndex:   i,
							handlerName: sagaDef.HandlerInfo.TransactionInfo[i].HandlerName,
							isLast:      i == len(sagaDef.Steps)-1,
						}

						// Create compensation tasks for all steps
						compensationTasks[i] = &compensationTask{
							ctx:         CompensationContext{},
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
								tp.logger.Debug(tp.ctx, "scheduler: Dispatching next transaction task", "handlerName", transactionTasks[nextIndex].handlerName)
								tx, err := tp.client.Tx(tp.ctx)
								if err != nil {
									tp.logger.Error(tp.ctx, "Failed to start transaction for next transaction task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Started transaction for next transaction task")
								var transactionExecution *ent.SagaExecution
								if transactionExecution, err = tx.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeTransaction).
									SetHandlerName(sagaDef.HandlerInfo.TransactionInfo[nextIndex].HandlerName).
									SetSequence(nextIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to create next transaction task", "error", err)
									if rerr := tx.Rollback(); rerr != nil {
										tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
									} else {
										tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
									}
									return err
								}
								if err = tx.Commit(); err != nil {
									tp.logger.Error(tp.ctx, "Failed to commit transaction for next transaction task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Successfully committed transaction for next transaction task")

								transactionTasks[nextIndex].executionID = transactionExecution.ID

								// Before moving to the next transaction, update the last successful index
								lastSuccessfulIndex = i
								return queueTransaction.Submit(transactionTasks[nextIndex])
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
								tx, err := tp.client.Tx(tp.ctx)
								if err != nil {
									tp.logger.Error(tp.ctx, "Failed to start transaction for compensation task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Started transaction for compensation task")
								var compensationExecution *ent.SagaExecution
								if compensationExecution, err = tx.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeCompensation).
									SetHandlerName(sagaDef.HandlerInfo.CompensationInfo[lastSuccessfulIndex].HandlerName).
									SetSequence(lastSuccessfulIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to create compensation task", "error", err)
									if rerr := tx.Rollback(); rerr != nil {
										tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
									} else {
										tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
									}
									return err
								}
								if err = tx.Commit(); err != nil {
									tp.logger.Error(tp.ctx, "Failed to commit transaction for compensation task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Successfully committed transaction for compensation task")

								compensationTasks[lastSuccessfulIndex].executionID = compensationExecution.ID

								// Start compensation from the last successful transaction
								return queueCompensation.Submit(compensationTasks[lastSuccessfulIndex])
							}

							tx, err := tp.client.Tx(tp.ctx)
							if err != nil {
								tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga status", "error", err)
								return err
							}
							tp.logger.Debug(tp.ctx, "Started transaction for updating saga status")
							_, err = tx.Saga.UpdateOne(sagaExecution.Edges.Saga).
								SetStatus(saga.StatusCompensated).
								Save(tp.ctx)
							if err != nil {
								tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to update saga status", "error", err)
								if rerr := tx.Rollback(); rerr != nil {
									tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
								} else {
									tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
								}
								return err
							}
							if err = tx.Commit(); err != nil {
								tp.logger.Error(tp.ctx, "Failed to commit transaction for updating saga status", "error", err)
								return err
							}
							tp.logger.Debug(tp.ctx, "Successfully committed transaction for updating saga status")
							return nil
						}
					}

					// Link compensation tasks in reverse order
					for i := len(sagaDef.Steps) - 1; i >= 0; i-- {
						if i > 0 {
							prevIndex := i - 1
							compensationTasks[i].next = func() error {
								tx, err := tp.client.Tx(tp.ctx)
								if err != nil {
									tp.logger.Error(tp.ctx, "Failed to start transaction for next compensation task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Started transaction for next compensation task")
								var compensationExecution *ent.SagaExecution
								if compensationExecution, err = tx.SagaExecution.Create().
									SetID(uuid.NewString()).
									SetStatus(sagaexecution.StatusRunning).
									SetStepType(sagaexecution.StepTypeCompensation).
									SetHandlerName(sagaDef.HandlerInfo.CompensationInfo[prevIndex].HandlerName).
									SetSequence(prevIndex).
									SetSaga(sagaExecution.Edges.Saga).
									Save(tp.ctx); err != nil {
									tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to create next compensation task", "error", err)
									if rerr := tx.Rollback(); rerr != nil {
										tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
									} else {
										tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
									}
									return err
								}
								if err = tx.Commit(); err != nil {
									tp.logger.Error(tp.ctx, "Failed to commit transaction for next compensation task", "error", err)
									return err
								}
								tp.logger.Debug(tp.ctx, "Successfully committed transaction for next compensation task")

								compensationTasks[prevIndex].executionID = compensationExecution.ID

								return queueCompensation.Submit(compensationTasks[prevIndex])
							}
						} else {
							// Optionally, for the first compensation task, you might decide to loop back or end the chain
							compensationTasks[i].next = nil
						}
					}

				}

				// Submit the first transaction task
				if err := queueTransaction.Submit(transactionTasks[0]); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to dispatch first transaction task", "error", err)
					tx, err := tp.client.Tx(tp.ctx)
					if err != nil {
						tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga status after dispatch failure", "error", err)
						continue
					}
					tp.logger.Debug(tp.ctx, "Started transaction for updating saga status after dispatch failure")
					if _, err := tx.Saga.UpdateOne(sagaExecution.Edges.Saga).
						SetStatus(saga.StatusFailed).
						SetError(err.Error()).
						Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to update saga status", "error", err)
						if rerr := tx.Rollback(); rerr != nil {
							tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
						} else {
							tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
						}
						continue
					}
					if err := tx.Commit(); err != nil {
						tp.logger.Error(tp.ctx, "Failed to commit transaction for updating saga status after dispatch failure", "error", err)
						continue
					}
					tp.logger.Debug(tp.ctx, "Successfully committed transaction for updating saga status after dispatch failure")
					continue
				}

				tx, err := tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga status to running", "error", err)
					continue
				}
				tp.logger.Debug(tp.ctx, "Started transaction for updating saga status to running")
				if _, err := tx.Saga.UpdateOne(sagaExecution.Edges.Saga).
					SetStatus(saga.StatusRunning).
					Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler saga execution: Failed to update saga status to running", "error", err)
					if rerr := tx.Rollback(); rerr != nil {
						tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
					} else {
						tp.logger.Debug(tp.ctx, "Successfully rolled back transaction")
					}
					continue
				}
				if err := tx.Commit(); err != nil {
					tp.logger.Error(tp.ctx, "Failed to commit transaction for updating saga status to running", "error", err)
					continue
				}
				tp.logger.Debug(tp.ctx, "Successfully committed transaction for updating saga status to running")
			}

			runtime.Gosched()
		}
	}
}
