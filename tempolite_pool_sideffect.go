package tempolite

import (
	"context"
	"reflect"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

type sideEffectTask[T Identifier] struct {
	ctx         SideEffectContext[T]
	handler     interface{}
	handlerName HandlerIdentity
}

func (tp *Tempolite[T]) createSideEffectPool(countWorkers int) *retrypool.Pool[*sideEffectTask[T]] {
	opts := []retrypool.Option[*sideEffectTask[T]]{
		retrypool.WithAttempts[*sideEffectTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.sideEffectOnSuccess),
		retrypool.WithOnTaskFailure(tp.sideEffectOnFailure),
		retrypool.WithPanicHandler(tp.sideEffectOnPanic),
		retrypool.WithOnRetry(tp.sideEffectOnRetry),
		retrypool.WithPanicWorker[*sideEffectTask[T]](tp.sideEffectWorkerPanic),
	}

	workers := []retrypool.Worker[*sideEffectTask[T]]{}

	for i := 0; i < countWorkers; i++ {
		workers = append(workers, sideEffectWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite[T]) sideEffectWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "sideEffect pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "sideEffect pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) sideEffectOnPanic(task *sideEffectTask[T], v interface{}, stackTrace string) {
	tp.logger.Debug(tp.ctx, "sideEffect pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "sideEffect pool task panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) sideEffectOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*sideEffectTask[T]]) {
	tp.logger.Debug(tp.ctx, "sideEffect pool task retry", "attempt", attempt, "error", err)
}

func (tp *Tempolite[T]) sideEffectOnSuccess(controller retrypool.WorkerController[*sideEffectTask[T]], workerID int, worker retrypool.Worker[*sideEffectTask[T]], task *retrypool.TaskWrapper[*sideEffectTask[T]]) {

	tp.logger.Debug(tp.ctx, "sideEffect task on success", "workerID", workerID, "runID", task.Data().ctx.RunID(), "entityID", task.Data().ctx.EntityID(), "executionID", task.Data().ctx.ExecutionID(), "handlerName", task.Data().handlerName)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for side effect success", "error", err)
		return
	}

	if _, err := tx.SideEffectExecution.UpdateOneID(task.Data().ctx.ExecutionID()).SetStatus(sideeffectexecution.StatusCompleted).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "sideEffect task on success: SideEffectExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if _, err := tx.SideEffect.UpdateOneID(task.Data().ctx.EntityID()).SetStatus(sideeffect.StatusCompleted).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "sideEffect task on success: SideEffect.UpdateOneID failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction for side effect success", "error", err)
	}
}

func (tp *Tempolite[T]) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask[T]], workerID int, worker retrypool.Worker[*sideEffectTask[T]], task *retrypool.TaskWrapper[*sideEffectTask[T]], err error) retrypool.DeadTaskAction {

	tp.logger.Debug(tp.ctx, "sideEffect task on failure", "workerID", workerID, "runID", task.Data().ctx.RunID(), "entityID", task.Data().ctx.EntityID(), "executionID", task.Data().ctx.ExecutionID(), "handlerName", task.Data().handlerName)

	tx, txErr := tp.client.Tx(tp.ctx)
	if txErr != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for side effect failure", "error", txErr)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if _, err := tx.SideEffectExecution.UpdateOneID(task.Data().ctx.ExecutionID()).SetStatus(sideeffectexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "sideEffect task on failure: SideEffectExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if _, err := tx.SideEffect.UpdateOneID(task.Data().ctx.EntityID()).SetStatus(sideeffect.StatusFailed).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "sideEffect task on failure: SideEffect.UpdateOneID failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction for side effect failure", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

type sideEffectWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w sideEffectWorker[T]) Run(ctx context.Context, data *sideEffectTask[T]) error {
	w.tp.logger.Debug(data.ctx, "sideEffect pool worker run", "executionID", data.ctx.executionID, "handler", data.handlerName)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}
	returnedValues := reflect.ValueOf(data.handler).Call(values)

	res := make([]interface{}, len(returnedValues))
	for i, v := range returnedValues {
		res[i] = v.Interface()
	}

	tx, err := w.tp.client.Tx(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(data.ctx, "Failed to start transaction for updating side effect execution output", "error", err)
		return err
	}

	if _, err := tx.SideEffectExecution.UpdateOneID(data.ctx.ExecutionID()).SetOutput(res).Save(w.tp.ctx); err != nil {
		w.tp.logger.Error(data.ctx, "sideEffect pool worker run: SideEffectExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			w.tp.logger.Error(data.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		w.tp.logger.Error(data.ctx, "Failed to commit transaction for updating side effect execution output", "error", err)
		return err
	}

	return nil
}
