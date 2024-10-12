package tempolite

import (
	"context"
	"log"

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

}

func (tp *Tempolite) sideEffectOnFailure(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask], err error) retrypool.DeadTaskAction {

	return retrypool.DeadTaskActionDoNothing
}

type sideEffectWorker struct {
	id int
	tp *Tempolite
}

func (w sideEffectWorker) Run(ctx context.Context, data *sideEffectTask) error {

	return nil
}
