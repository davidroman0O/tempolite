package tempolite

import (
	"context"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/retrypool"
)

type activityTask[T Identifier] struct {
	ctx         ActivityContext[T]
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
}

func (tp *Tempolite[T]) createActivityPool() *retrypool.Pool[*activityTask[T]] {
	opts := []retrypool.Option[*activityTask[T]]{
		retrypool.WithAttempts[*activityTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.activityOnSuccess),
		retrypool.WithOnTaskFailure(tp.activityOnFailure),
		retrypool.WithPanicHandler(tp.activityOnPanic),
		retrypool.WithOnRetry(tp.activityOnRetry),
		retrypool.WithPanicWorker[*activityTask[T]](tp.activityWorkerPanic),
	}

	workers := []retrypool.Worker[*activityTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, activityWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)

}

func (tp *Tempolite[T]) activityWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "activity pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "activity pool worker panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) activityOnPanic(task *activityTask[T], v interface{}, stackTrace string) {
	tp.logger.Debug(tp.ctx, "activity pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "activity pool task panicked", "stackTrace", stackTrace)
}

func (tp *Tempolite[T]) activityOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*activityTask[T]]) {
	tp.logger.Debug(tp.ctx, "activity pool task retry", "attempt", attempt, "error", err)
}

func (tp *Tempolite[T]) activityOnSuccess(controller retrypool.WorkerController[*activityTask[T]], workerID int, worker retrypool.Worker[*activityTask[T]], task *retrypool.TaskWrapper[*activityTask[T]]) {
	tp.logger.Debug(tp.ctx, "activity task on success", "workerID", workerID, "executionID", task.Data().ctx.executionID, "handlerName", task.Data().handlerName)

	if _, err := tp.client.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on success: ActivityExecution.Update failed", "error", err)
	}

	if _, err := tp.client.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusCompleted).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on success: activity.UpdateOneID failed", "error", err)
	}
}

func (tp *Tempolite[T]) activityOnFailure(controller retrypool.WorkerController[*activityTask[T]], workerID int, worker retrypool.Worker[*activityTask[T]], task *retrypool.TaskWrapper[*activityTask[T]], err error) retrypool.DeadTaskAction {
	// printf with err + retryCount + maxRetry
	tp.logger.Error(tp.ctx, "activity task on failure", "workerID", workerID, "executionID", task.Data().ctx.executionID, "handlerName", task.Data().handlerName)

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().maxRetry > 0 && task.Data().retryCount < task.Data().maxRetry {
		task.Data().retryCount++
		// fmt.Println("retry it the task: ", err)
		// Just by creating a new activity execution, we're incrementing the total count of executions which is the retry count in the database
		if _, err := tp.client.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "activity task on failure: ActivityExecution.Update failed", "error", err)
		}

		if _, err := tp.client.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "activity task on failure: activity.UpdateOneID failed", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := task.Data().retry(); err != nil {
			// TODO: should change the status of the activity to failed
			tp.logger.Error(tp.ctx, "activity task on failure: retry failed", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		return retrypool.DeadTaskActionDoNothing
	}

	tp.logger.Debug(tp.ctx, "activity task on failure", "workerID", workerID, "executionID", task.Data().ctx.executionID, "handlerName", task.Data().handlerName)

	if _, err := tp.client.ActivityExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(activityexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on failure: ActivityExecution.Update failed", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}
	if _, err := tp.client.Activity.UpdateOneID(task.Data().ctx.activityID).SetStatus(activity.StatusFailed).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "activity task on failure: activity.UpdateOneID failed", "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

type activityWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w activityWorker[T]) Run(ctx context.Context, data *activityTask[T]) error {

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

	if _, err := w.tp.client.ActivityExecution.UpdateOneID(data.ctx.executionID).SetOutput(res).Save(w.tp.ctx); err != nil {
		w.tp.logger.Error(w.tp.ctx, "activityWorker: ActivityExecution.Update failed", "error", err)
		return err
	}

	w.tp.logger.Debug(w.tp.ctx, "activityWorker: Run", "handlerName", data.handlerName, "output", res)

	return errRes
}
