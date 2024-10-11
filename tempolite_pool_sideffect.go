package tempolite

import (
	"context"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/retrypool"
)

type sideEffectTask struct {
	ctx         SideEffectContext
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
}

func (tp *Tempolite) createSideEffectPool() *retrypool.Pool[*sideEffectTask] {
	opts := []retrypool.Option[*sideEffectTask]{
		retrypool.WithAttempts[*sideEffectTask](1),
		retrypool.WithOnTaskSuccess(tp.sideEffectOnSuccess),
		retrypool.WithOnTaskFailure(tp.sideEffectOnFailure),
		retrypool.WithPanicHandler(tp.sideEffectOnPanic),
		retrypool.WithOnRetry(tp.sideEffectOnRetry),
	}

	workers := []retrypool.Worker[*sideEffectTask]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, sideEffectWorker{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)

}

func (tp *Tempolite) sideEffectOnPanic(task *sideEffectTask, v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite) sideEffectOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*sideEffectTask]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite) sideEffectOnSuccess(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask]) {
	log.Printf("sideEffectOnSuccess: %d", workerID)
	if _, err := tp.client.SideEffectExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(sideeffectexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnSuccess: SideEffectExecution.Update failed: %v", err)
	}
	if _, err := tp.client.SideEffect.UpdateOneID(task.Data().ctx.sideEffectID).SetStatus(sideeffect.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnSuccess: SideEffect.UpdateOneID failed: %v", err)
	}
}

func (tp *Tempolite) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask], err error) retrypool.DeadTaskAction {
	// printf with err + retryCount + maxRetry
	log.Printf("sideEffectOnFailure: %v, %d, %d", err, task.Data().retryCount, task.Data().maxRetry)

	if _, err := tp.client.SideEffectExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(sideeffectexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnFailure: SideEffectExecution.Update failed: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if _, err := tp.client.SideEffect.UpdateOneID(task.Data().ctx.sideEffectID).SetStatus(sideeffect.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnFailure: sideEffect.UpdateOneID failed: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

type sideEffectWorker struct {
	id int
	tp *Tempolite
}

func (w sideEffectWorker) Run(ctx context.Context, data *sideEffectTask) error {
	log.Printf("sideEffectWorker: %s, %v", data.handlerName, data.params)

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

	if _, err := w.tp.client.SideEffectExecution.UpdateOneID(data.ctx.executionID).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("sideEffectworker: SideEffectExecution.Update failed: %v", err)
	}

	return errRes
}
