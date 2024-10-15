package tempolite

import (
	"context"
	"log"

	"github.com/davidroman0O/retrypool"
)

type compensationTask[T Identifier] struct {
	ctx         CompensationContext[T]
	handler     interface{}
	handlerName HandlerIdentity
}

func (tp *Tempolite[T]) createCompensationPool() *retrypool.Pool[*compensationTask[T]] {
	opts := []retrypool.Option[*compensationTask[T]]{
		retrypool.WithAttempts[*compensationTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.compensationOnSuccess),
		retrypool.WithOnTaskFailure(tp.compensationOnFailure),
		retrypool.WithPanicHandler(tp.compensationOnPanic),
		retrypool.WithOnRetry(tp.compensationOnRetry),
	}

	workers := []retrypool.Worker[*compensationTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, compensationWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite[T]) compensationOnPanic(task *compensationTask[T], v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite[T]) compensationOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*compensationTask[T]]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite[T]) compensationOnSuccess(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]]) {
	log.Printf("compensationOnSuccess: %d %s %s %s %s", workerID, task.Data().ctx.RunID(), task.Data().ctx.EntityID(), task.Data().ctx.ExecutionID(), task.Data().handlerName)

}

func (tp *Tempolite[T]) compensationOnFailure(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]], err error) retrypool.DeadTaskAction {
	log.Printf("compensationOnFailure: %v", err)

	return retrypool.DeadTaskActionDoNothing
}

type compensationWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w compensationWorker[T]) Run(ctx context.Context, data *compensationTask[T]) error {
	log.Printf("compensationWorker: %s", data.handlerName)

	return nil
}
