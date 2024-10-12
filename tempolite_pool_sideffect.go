package tempolite

import (
	"context"
	"log"

	"github.com/davidroman0O/retrypool"
)

type sideEffectTask[T Identifier] struct {
	ctx         SideEffectContext[T]
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
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

}

func (tp *Tempolite[T]) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask[T]], workerID int, worker retrypool.Worker[*sideEffectTask[T]], task *retrypool.TaskWrapper[*sideEffectTask[T]], err error) retrypool.DeadTaskAction {

	return retrypool.DeadTaskActionDoNothing
}

type sideEffectWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w sideEffectWorker[T]) Run(ctx context.Context, data *sideEffectTask[T]) error {

	return nil
}
