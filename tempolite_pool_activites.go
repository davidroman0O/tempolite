package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/retrypool"
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

func (tp *Tempolite) createActivityPool() *retrypool.Pool[*activityTask] {
	opts := []retrypool.Option[*activityTask]{
		retrypool.WithAttempts[*activityTask](1),
		retrypool.WithOnTaskSuccess(tp.activityOnSuccess),
		retrypool.WithOnTaskFailure(tp.activityOnFailure),
	}
	return retrypool.New(
		tp.ctx,
		[]retrypool.Worker[*activityTask]{
			activityWorker{
				tp: tp,
			},
		},
		opts...)
}

func (tp *Tempolite) activityOnSuccess(controller retrypool.WorkerController[*activityTask], workerID int, worker retrypool.Worker[*activityTask], task *retrypool.TaskWrapper[*activityTask]) {
	log.Printf("activityOnSuccess: %d", workerID)
	if _, err := tp.client.ActivityExecution.Update().Where(activityexecution.IDEQ(task.Data().ctx.executionID)).SetStatus(activityexecution.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("activityOnSuccess: ActivityExecution.Update failed: %v", err)
	}
}

func (tp *Tempolite) activityOnFailure(controller retrypool.WorkerController[*activityTask], workerID int, worker retrypool.Worker[*activityTask], task *retrypool.TaskWrapper[*activityTask], err error) retrypool.DeadTaskAction {

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().retryCount < task.Data().maxRetry {
		task.Data().retryCount++
		fmt.Println("retry it the task: ", err)
		// Just by creating a new workflow execution, we're incrementing the total count of executions which is the retry count in the database
		if _, err := tp.client.ActivityExecution.Update().SetStatus(activityexecution.StatusRetried).Save(tp.ctx); err != nil {
			log.Printf("activityOnFailure: ActivityExecution.Update failed: %v", err)
		}
		if err := task.Data().retry(); err != nil {
			log.Printf("activityOnFailure: retry failed: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}
		return retrypool.DeadTaskActionDoNothing
	}

	if _, err := tp.client.ActivityExecution.Update().SetStatus(activityexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		log.Printf("activityOnFailure: ActivityExecution.Update failed: %v", err)
	}

	return retrypool.DeadTaskActionDoNothing
}

type activityWorker struct {
	tp *Tempolite
}

func (w activityWorker) Run(ctx context.Context, data *activityTask) error {
	log.Printf("activityWorker: %s, %v", data.handlerName, data.params)

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

	if _, err := w.tp.client.ActivityExecution.Update().Where(activityexecution.IDEQ(data.ctx.executionID)).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("activityworker: ActivityExecution.Update failed: %v", err)
	}

	return errRes
}
