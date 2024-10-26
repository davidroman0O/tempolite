package tempolite

import (
	"context"
	"fmt"
	"reflect"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

type sideEffectTask struct {
	ctx         SideEffectContext
	handler     interface{}
	handlerName HandlerIdentity
}

func (tp *Tempolite) createSideEffectPool(countWorkers int) *retrypool.Pool[*sideEffectTask] {
	opts := []retrypool.Option[*sideEffectTask]{
		retrypool.WithAttempts[*sideEffectTask](1),
		retrypool.WithOnTaskSuccess(tp.sideEffectOnSuccess),
		retrypool.WithOnTaskFailure(tp.sideEffectOnFailure),
		retrypool.WithPanicHandler(tp.sideEffectOnPanic),
		retrypool.WithOnRetry(tp.sideEffectOnRetry),
		retrypool.WithPanicWorker[*sideEffectTask](tp.sideEffectWorkerPanic),
	}

	workers := []retrypool.Worker[*sideEffectTask]{}

	for i := 0; i < countWorkers; i++ {
		workers = append(workers, sideEffectWorker{id: tp.getWorkerSideEffectID(), tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite) sideEffectWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "sideEffect pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "sideEffect pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite) sideEffectOnPanic(task *sideEffectTask, v interface{}, stackTrace string) {
	tp.logger.Debug(tp.ctx, "sideEffect pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "sideEffect pool task panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite) sideEffectOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*sideEffectTask]) {
	tp.logger.Debug(tp.ctx, "sideEffect pool task retry", "attempt", attempt, "error", err)
}

func (tp *Tempolite) sideEffectOnSuccess(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask]) {

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

func (tp *Tempolite) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask], err error) retrypool.DeadTaskAction {

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

type sideEffectWorker struct {
	id int
	tp *Tempolite
}

func (w sideEffectWorker) Run(ctx context.Context, data *sideEffectTask) error {
	w.tp.logger.Debug(data.ctx, "sideEffect pool worker run", "executionID", data.ctx.executionID, "handler", data.handlerName)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}
	returnedValues := reflect.ValueOf(data.handler).Call(values)

	res := make([]interface{}, len(returnedValues))
	for i, v := range returnedValues {
		res[i] = v.Interface()
	}

	var value any
	var ok bool
	var sideEffectInfo SideEffect

	if value, ok = w.tp.sideEffects.Load(data.ctx.sideEffectID); ok {
		if sideEffectInfo, ok = value.(SideEffect); !ok {
			w.tp.logger.Error(data.ctx, "sideEffect pool worker: sideEffect not found", "sideEffectID", data.ctx.sideEffectID, "executionID", data.ctx.executionID, "handler", data.handlerName)
			return fmt.Errorf("sideEffect %s not found", data.handlerName)
		}
	} else {
		w.tp.logger.Error(data.ctx, "sideEffect pool worker: sideEffect not found", "sideEffectID", data.ctx.sideEffectID, "executionID", data.ctx.executionID, "handler", data.handlerName)
		return fmt.Errorf("sideEffect %s not found", data.handlerName)
	}

	serializableOutput, err := w.tp.convertOutputsForSerialization(HandlerInfo(sideEffectInfo), res)
	if err != nil {
		w.tp.logger.Error(data.ctx, "sideEffect pool worker: convertOutputsForSerialization failed", "error", err)
		return err
	}

	tx, err := w.tp.client.Tx(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(data.ctx, "Failed to start transaction for updating side effect execution output", "error", err)
		return err
	}

	if _, err := tx.SideEffectExecution.UpdateOneID(data.ctx.ExecutionID()).SetOutput(serializableOutput).Save(w.tp.ctx); err != nil {
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
