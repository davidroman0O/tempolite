package tempolite

import (
	"context"
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/retrypool"
)

type transactionTask[T Identifier] struct {
	ctx         TransactionContext[T]
	sagaID      string
	executionID string
	stepIndex   int
	handlerName string
	isLast      bool
	next        func() error
	compensate  func() error
}

func (tp *Tempolite[T]) createTransactionPool() *retrypool.Pool[*transactionTask[T]] {
	opts := []retrypool.Option[*transactionTask[T]]{
		retrypool.WithAttempts[*transactionTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.transactionOnSuccess),
		retrypool.WithOnTaskFailure(tp.transactionOnFailure),
		retrypool.WithPanicHandler(tp.transactionOnPanic),
		retrypool.WithOnRetry(tp.transactionOnRetry),
		retrypool.WithPanicWorker[*transactionTask[T]](tp.transactionWorkerPanic),
	}

	workers := []retrypool.Worker[*transactionTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, transactionWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(tp.ctx, workers, opts...)
}

func (tp *Tempolite[T]) transactionWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "transaction pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "transaction pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) transactionOnSuccess(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]]) {

	tp.logger.Debug(tp.ctx, "transaction task on success", "workerID", workerID, "executionID", task.Data().executionID, "handlerName", task.Data().handlerName)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga execution status", "error", err)
		return
	}

	_, err = tx.SagaExecution.
		UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusCompleted).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "transaction task on success: SagaExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if task.Data().isLast {
		_, err := tx.Saga.UpdateOneID(task.Data().sagaID).
			SetStatus(saga.StatusCompleted).
			Save(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Failed to update saga status to completed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return
		}
		tp.logger.Debug(tp.ctx, "Updated saga status to completed", "sagaID", task.Data().sagaID)
	} else if task.Data().next != nil {
		if err := tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
			return
		}
		if err := task.Data().next(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to dispatch next transaction task", "error", err)
		} else {
			tp.logger.Debug(tp.ctx, "Dispatched next transaction task", "handlerName", task.Data().handlerName)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
	}
}

func (tp *Tempolite[T]) transactionOnFailure(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]], err error) retrypool.DeadTaskAction {

	tp.logger.Debug(tp.ctx, "transaction task on failure", "workerID", workerID, "executionID", task.Data().executionID, "handlerName", task.Data().handlerName)

	tx, txErr := tp.client.Tx(tp.ctx)
	if txErr != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga execution status", "error", txErr)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	_, updateErr := tx.SagaExecution.UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusFailed).
		Save(tp.ctx)
	if updateErr != nil {
		tp.logger.Error(tp.ctx, "Failed to update saga execution status", "error", updateErr)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if task.Data().compensate != nil {
		if err := task.Data().compensate(); err != nil {
			// TODO: should we update the saga?
			tp.logger.Error(tp.ctx, "Failed to dispatch compensation task", "error", err)
		}
	}

	return retrypool.DeadTaskActionDoNothing
}

func (tp *Tempolite[T]) transactionOnPanic(task *transactionTask[T], v interface{}, stackTrace string) {

	tp.logger.Debug(tp.ctx, "transaction pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "transaction pool task panicked", "stackTrace", stackTrace)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga execution status after panic", "error", err)
		return
	}

	_, err = tx.SagaExecution.UpdateOneID(task.executionID).
		SetStatus(sagaexecution.StatusFailed).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to update saga execution status after panic", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	_, err = tx.Saga.UpdateOneID(task.sagaID).
		SetStatus(saga.StatusFailed).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to update saga status to failed after panic", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction after panic", "error", err)
	}
}

func (tp *Tempolite[T]) transactionOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*transactionTask[T]]) {
	tp.logger.Debug(tp.ctx, "transaction task retry", "attempt", attempt, "error", err)
}

type transactionWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w transactionWorker[T]) Run(ctx context.Context, data *transactionTask[T]) error {
	w.tp.logger.Debug(ctx, "Executing transaction step", "handlerName", data.handlerName)

	sagaHandlerInfo, ok := w.tp.sagas.Load(data.sagaID)
	if !ok {
		w.tp.logger.Error(ctx, "saga handler info not found", "sagaID", data.sagaID)
		return fmt.Errorf("saga handler info not found for ID: %s", data.sagaID)
	}

	sagaDef := sagaHandlerInfo.(*SagaDefinition[T])
	step := sagaDef.Steps[data.stepIndex]

	result, err := step.Transaction(data.ctx)
	if err != nil {
		w.tp.logger.Error(ctx, "Transaction step failed", "handlerName", data.handlerName, "error", err)
		return err
	}

	w.tp.logger.Debug(ctx, "Transaction step completed", "handlerName", data.handlerName, "result", result)

	return nil
}
