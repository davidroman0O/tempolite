package tempolite

import (
	"context"
	"fmt"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/sagaexecution"
)

type compensationTask struct {
	ctx         CompensationContext
	sagaID      string
	executionID string
	stepIndex   int
	handlerName string
	isLast      bool
	next        func() error
}

func (tp *Tempolite) createCompensationPool(queue string, countWorkers int) (*retrypool.Pool[*compensationTask], error) {
	opts := []retrypool.Option[*compensationTask]{
		retrypool.WithAttempts[*compensationTask](1),
		retrypool.WithOnTaskSuccess(tp.compensationOnSuccess),
		retrypool.WithOnTaskFailure(tp.compensationOnFailure),
		retrypool.WithPanicHandler(tp.compensationOnPanic),
		retrypool.WithOnRetry(tp.compensationOnRetry),
		retrypool.WithPanicWorker[*compensationTask](tp.compensationWorkerPanic),
	}

	workers := []retrypool.Worker[*compensationTask]{}

	tp.logger.Debug(tp.ctx, "createCompensationPool", "queue", queue, "countWorkers", countWorkers)

	for i := 0; i < countWorkers; i++ {
		id, err := tp.getWorkerCompensationID(queue)
		if err != nil {
			return nil, fmt.Errorf("failed to generate compensation worker ID: %w", err)
		}
		workers = append(workers, compensationWorker{id: id, tp: tp})
	}

	return retrypool.New(tp.ctx, workers, opts...), nil
}

func (tp *Tempolite) compensationWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "compensation pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "compensation pool worker panicked", "err", err, "stackTrace", stackTrace)
}

func (tp *Tempolite) compensationOnSuccess(controller retrypool.WorkerController[*compensationTask], workerID int, worker retrypool.Worker[*compensationTask], task *retrypool.TaskWrapper[*compensationTask]) {

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

func (tp *Tempolite) compensationOnFailure(controller retrypool.WorkerController[*compensationTask], workerID int, worker retrypool.Worker[*compensationTask], task *retrypool.TaskWrapper[*compensationTask], err error) retrypool.DeadTaskAction {

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

func (tp *Tempolite) compensationOnPanic(task *compensationTask, v interface{}, stackTrace string) {

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

func (tp *Tempolite) compensationOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*compensationTask]) {
	tp.logger.Debug(tp.ctx, "compensation task retry", "attempt", attempt, "error", err)
}

type compensationWorker struct {
	id int
	tp *Tempolite
}

func (w compensationWorker) Run(ctx context.Context, data *compensationTask) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		w.tp.logger.Debug(ctx, "Executing compensation step", "handlerName", data.handlerName)

		sagaHandlerInfo, ok := w.tp.sagas.Load(data.sagaID)
		if !ok {
			w.tp.logger.Error(ctx, "saga handler info not found", "sagaID", data.sagaID)
			return fmt.Errorf("saga handler info not found for ID: %s", data.sagaID)
		}
		sagaDef := sagaHandlerInfo.(*SagaDefinition)
		step := sagaDef.Steps[data.stepIndex]

		result, err := step.Compensation(data.ctx)
		if err != nil {
			w.tp.logger.Error(ctx, "Compensation step failed", "handlerName", data.handlerName, "error", err)
			return err
		}

		w.tp.logger.Debug(ctx, "Compensation step completed", "handlerName", data.handlerName, "result", result)

		return nil
	}
}
