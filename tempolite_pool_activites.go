package tempolite

import (
	"context"
	"fmt"
	"reflect"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent/activity"
	"github.com/davidroman0O/tempolite/ent/activityexecution"
)

type activityTask struct {
	ctx         ActivityContext
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
}

func (tp *Tempolite) createActivityPool(countWorkers int) *retrypool.Pool[*activityTask] {
	opts := []retrypool.Option[*activityTask]{
		retrypool.WithAttempts[*activityTask](1),
		retrypool.WithOnTaskSuccess(tp.activityOnSuccess),
		retrypool.WithOnTaskFailure(tp.activityOnFailure),
		retrypool.WithPanicHandler(tp.activityOnPanic),
		retrypool.WithOnRetry(tp.activityOnRetry),
		retrypool.WithPanicWorker[*activityTask](tp.activityWorkerPanic),
	}

	workers := []retrypool.Worker[*activityTask]{}

	for i := 0; i < countWorkers; i++ {
		workers = append(workers, activityWorker{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)

}

func (tp *Tempolite) activityWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "activity pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "activity pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite) activityOnPanic(task *activityTask, v interface{}, stackTrace string) {
	tp.logger.Debug(tp.ctx, "activity pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "activity pool task panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite) activityOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*activityTask]) {
	tp.logger.Debug(tp.ctx, "activity pool task retry", "attempt", attempt, "error", err)
}

func (tp *Tempolite) activityOnSuccess(controller retrypool.WorkerController[*activityTask], workerID int, worker retrypool.Worker[*activityTask], task *retrypool.TaskWrapper[*activityTask]) {
	tp.logger.Debug(tp.ctx, "activity task on success", "workerID", workerID, "executionID", task.Data().ctx.executionID, "handlerName", task.Data().handlerName)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating activity status", "error", err)
		return
	}

	if _, err := tx.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on success: ActivityExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if _, err := tx.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusCompleted).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on success: activity.UpdateOneID failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction", "error", err)
	}
}

func (tp *Tempolite) activityOnFailure(controller retrypool.WorkerController[*activityTask], workerID int, worker retrypool.Worker[*activityTask], task *retrypool.TaskWrapper[*activityTask], taskErr error) retrypool.DeadTaskAction {
	// printf with err + retryCount + maxRetry
	tp.logger.Error(tp.ctx, "activity task on failure", "workerID", workerID, "executionID", task.Data().ctx.executionID, "handlerName", task.Data().handlerName)

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().maxRetry > 0 && task.Data().retryCount < task.Data().maxRetry {
		task.Data().retryCount++
		// fmt.Println("retry it the task: ", err)
		// Just by creating a new activity execution, we're incrementing the total count of executions which is the retry count in the database
		tx, err := tp.client.Tx(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Failed to start transaction for activity retry", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if _, err := tx.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "activity task on failure: ActivityExecution.Update failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if _, err := tx.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "activity task on failure: activity.UpdateOneID failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to commit transaction for activity retry", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := task.Data().retry(); err != nil {
			// TODO: should change the status of the activity to failed
			tp.logger.Error(tp.ctx, "activity task on failure: retry failed", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		return retrypool.DeadTaskActionDoNothing
	}

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for updating activity status to failed", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if _, err := tx.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusFailed).SetError(taskErr.Error()).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on failure: ActivityExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}
	if _, err := tx.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusFailed).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on failure: activity.UpdateOneID failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction for updating activity status to failed", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

type activityWorker struct {
	id int
	tp *Tempolite
}

func (w activityWorker) Run(ctx context.Context, data *activityTask) error {

	w.tp.logger.Debug(w.tp.ctx, "activityWorker: Run", "handlerName", data.handlerName, "params", data.params)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}

	for _, v := range data.params {
		values = append(values, reflect.ValueOf(v))
	}

	returnedValues := reflect.ValueOf(data.handler).Call(values)

	var res []interface{}
	var errRes error
	if len(returnedValues) > 0 {
		res = make([]interface{}, len(returnedValues)-1)
		for i := 0; i < len(returnedValues)-1; i++ {
			res[i] = returnedValues[i].Interface()
		}
		if !returnedValues[len(returnedValues)-1].IsNil() {
			errRes = returnedValues[len(returnedValues)-1].Interface().(error)
		}
	}

	var value any
	var ok bool
	var activityInfo Activity

	if value, ok = w.tp.activities.Load(data.handlerName); ok {
		if activityInfo, ok = value.(Activity); !ok {
			w.tp.logger.Error(data.ctx, "activity pool worker: activity not found", "activityID", data.ctx.activityID, "executionID", data.ctx.executionID, "handler", data.handlerName)
			return fmt.Errorf("activity %s not found", data.handlerName)
		}
	} else {
		w.tp.logger.Error(data.ctx, "activity pool worker: activity not found", "activityID", data.ctx.activityID, "executionID", data.ctx.executionID, "handler", data.handlerName)
		return fmt.Errorf("activity %s not found", data.handlerName)
	}

	serializableOutput, err := w.tp.convertOutputsForSerialization(HandlerInfo(activityInfo), res)
	if err != nil {
		w.tp.logger.Error(data.ctx, "activity pool worker: convertOutputsForSerialization failed", "error", err)
		return err
	}

	tx, err := w.tp.client.Tx(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Failed to start transaction for updating activity execution output", "error", err)
		return err
	}

	if _, err := tx.ActivityExecution.UpdateOneID(data.ctx.executionID).SetOutput(serializableOutput).Save(w.tp.ctx); err != nil {
		w.tp.logger.Error(w.tp.ctx, "activityWorker: ActivityExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		w.tp.logger.Error(w.tp.ctx, "Failed to commit transaction for updating activity execution output", "error", err)
		return err
	}

	w.tp.logger.Debug(w.tp.ctx, "activityWorker: Run", "handlerName", data.handlerName, "output", res)

	return errRes
}
