package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/retrypool"
)

type sideEffectTask[T Identifier] struct {
	ctx         SideEffectContext[T]
	handler     interface{}
	handlerName HandlerIdentity
}

func (tp *Tempolite[T]) createSideEffectPool() *retrypool.Pool[*sideEffectTask[T]] {
	opts := []retrypool.Option[*sideEffectTask[T]]{
		retrypool.WithAttempts[*sideEffectTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.sideEffectOnSuccess),
		retrypool.WithOnTaskFailure(tp.sideEffectOnFailure),
		retrypool.WithPanicHandler(tp.sideEffectOnPanic),
		retrypool.WithOnRetry(tp.sideEffectOnRetry),
	}

	workers := []retrypool.Worker[*sideEffectTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, sideEffectWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite[T]) sideEffectOnPanic(task *sideEffectTask[T], v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite[T]) sideEffectOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*sideEffectTask[T]]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite[T]) sideEffectOnSuccess(controller retrypool.WorkerController[*sideEffectTask[T]], workerID int, worker retrypool.Worker[*sideEffectTask[T]], task *retrypool.TaskWrapper[*sideEffectTask[T]]) {
	log.Printf("sideEffectOnSuccess: %d %s %s %s %s", workerID, task.Data().ctx.RunID(), task.Data().ctx.EntityID(), task.Data().ctx.ExecutionID(), task.Data().handlerName)

	if _, err := tp.client.SideEffectExecution.UpdateOneID(task.Data().ctx.ExecutionID()).SetStatus(sideeffectexecution.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnSuccess: SideEffectExecution.Update failed: %v", err)
	}

	if _, err := tp.client.SideEffect.UpdateOneID(task.Data().ctx.EntityID()).SetStatus(sideeffect.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnSuccess: SideEffect.UpdateOneID failed: %v", err)
	}
}

func (tp *Tempolite[T]) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask[T]], workerID int, worker retrypool.Worker[*sideEffectTask[T]], task *retrypool.TaskWrapper[*sideEffectTask[T]], err error) retrypool.DeadTaskAction {
	log.Printf("sideEffectOnFailure: %v", err)

	if _, err := tp.client.SideEffectExecution.UpdateOneID(task.Data().ctx.ExecutionID()).SetStatus(sideeffectexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnFailure: SideEffectExecution.Update failed: %v", err)
	}

	if _, err := tp.client.SideEffect.UpdateOneID(task.Data().ctx.EntityID()).SetStatus(sideeffect.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("ERROR sideEffectOnFailure: SideEffect.UpdateOneID failed: %v", err)
	}

	return retrypool.DeadTaskActionDoNothing
}

type sideEffectWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w sideEffectWorker[T]) Run(ctx context.Context, data *sideEffectTask[T]) error {
	log.Printf("sideEffectWorker: %s", data.handlerName)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}
	returnedValues := reflect.ValueOf(data.handler).Call(values)

	res := make([]interface{}, len(returnedValues))
	for i, v := range returnedValues {
		res[i] = v.Interface()
	}

	fmt.Println("\t === res", res)

	if _, err := w.tp.client.SideEffectExecution.UpdateOneID(data.ctx.ExecutionID()).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("sideEffectWorker: SideEffectExecution.Update failed: %v", err)
		return err
	}

	return nil
}
