package tempolite

import (
	"context"
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/retrypool"
)

type compensationTask[T Identifier] struct {
	ctx         CompensationContext[T]
	sagaID      string
	executionID string
	stepIndex   int
	handlerName string
	isLast      bool
	next        func() error
}

func (tp *Tempolite[T]) createCompensationPool() *retrypool.Pool[*compensationTask[T]] {
	opts := []retrypool.Option[*compensationTask[T]]{
		retrypool.WithAttempts[*compensationTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.compensationOnSuccess),
		retrypool.WithOnTaskFailure(tp.compensationOnFailure),
		retrypool.WithPanicHandler(tp.compensationOnPanic),
		retrypool.WithOnRetry(tp.compensationOnRetry),
		retrypool.WithPanicWorker[*compensationTask[T]](tp.compensationWorkerPanic),
	}

	workers := []retrypool.Worker[*compensationTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, compensationWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(tp.ctx, workers, opts...)
}

func (tp *Tempolite[T]) compensationWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "compensation pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "compensation pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) compensationOnSuccess(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]]) {

	tp.logger.Debug(tp.ctx, "compensation task on success", "workerID", workerID, "executionID", task.Data().executionID, "handlerName", task.Data().handlerName)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga execution status", "error", err)
		return
	}

	_, err = tx.SagaExecution.UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusCompleted).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "compensation task on success: SagaExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if task.Data().isLast {
		_, err := tx.Saga.UpdateOneID(task.Data().sagaID).
			SetStatus(saga.StatusCompensated).
			Save(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "compensation task on success: Saga.Update failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return
		}
	} else if task.Data().next != nil {
		if err := tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
			return
		}
		if err := task.Data().next(); err != nil {
			// TODO: should update the entity status to failed
			tp.logger.Error(tp.ctx, "Failed to dispatch next compensation task", "error", err)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
	}
}

func (tp *Tempolite[T]) compensationOnFailure(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]], err error) retrypool.DeadTaskAction {

	tp.logger.Debug(tp.ctx, "compensation task on failure", "workerID", workerID, "executionID", task.Data().executionID, "handlerName", task.Data().handlerName)

	tx, txErr := tp.client.Tx(tp.ctx)
	if txErr != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating saga status", "error", txErr)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	_, updateErr := tp.client.Saga.UpdateOneID(task.Data().sagaID).
		SetStatus(saga.StatusFailed).
		Save(tp.ctx)
	if updateErr != nil {
		tp.logger.Error(tp.ctx, "Failed to update saga status to failed", "error", updateErr)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionAddToDeadTasks
}

func (tp *Tempolite[T]) compensationOnPanic(task *compensationTask[T], v interface{}, stackTrace string) {

	tp.logger.Debug(tp.ctx, "compensation pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "compensation pool task panicked", "stackTrace", stackTrace)

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

func (tp *Tempolite[T]) compensationOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*compensationTask[T]]) {
	tp.logger.Debug(tp.ctx, "compensation task retry", "attempt", attempt, "error", err)
}

type compensationWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w compensationWorker[T]) Run(ctx context.Context, data *compensationTask[T]) error {
	w.tp.logger.Debug(ctx, "Executing compensation step", "handlerName", data.handlerName)

	sagaHandlerInfo, ok := w.tp.sagas.Load(data.sagaID)
	if !ok {
		w.tp.logger.Error(ctx, "saga handler info not found", "sagaID", data.sagaID)
		return fmt.Errorf("saga handler info not found for ID: %s", data.sagaID)
	}
	sagaDef := sagaHandlerInfo.(*SagaDefinition[T])
	step := sagaDef.Steps[data.stepIndex]

	result, err := step.Compensation(data.ctx)
	if err != nil {
		w.tp.logger.Error(ctx, "Compensation step failed", "handlerName", data.handlerName, "error", err)
		return err
	}

	w.tp.logger.Debug(ctx, "Compensation step completed", "handlerName", data.handlerName, "result", result)

	return nil
}
