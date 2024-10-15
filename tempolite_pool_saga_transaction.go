package tempolite

import (
	"context"
	"log"

	"github.com/davidroman0O/retrypool"
)

type transactionTask[T Identifier] struct {
	ctx         TransactionContext[T]
	handler     interface{}
	handlerName HandlerIdentity
}

func (tp *Tempolite[T]) createTransactionPool() *retrypool.Pool[*transactionTask[T]] {
	opts := []retrypool.Option[*transactionTask[T]]{
		retrypool.WithAttempts[*transactionTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.transactiontOnSuccess),
		retrypool.WithOnTaskFailure(tp.transactiontOnFailure),
		retrypool.WithPanicHandler(tp.transactiontOnPanic),
		retrypool.WithOnRetry(tp.transactiontOnRetry),
	}

	workers := []retrypool.Worker[*transactionTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, transactionWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite[T]) transactiontOnPanic(task *transactionTask[T], v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite[T]) transactiontOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*transactionTask[T]]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite[T]) transactiontOnSuccess(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]]) {
	log.Printf("transactiontOnSuccess: %d %s %s %s %s", workerID, task.Data().ctx.RunID(), task.Data().ctx.EntityID(), task.Data().ctx.ExecutionID(), task.Data().handlerName)

}

func (tp *Tempolite[T]) transactiontOnFailure(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]], err error) retrypool.DeadTaskAction {
	log.Printf("transactiontOnFailure: %v", err)

	return retrypool.DeadTaskActionDoNothing
}

type transactionWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w transactionWorker[T]) Run(ctx context.Context, data *transactionTask[T]) error {
	log.Printf("transactionWorker: %s", data.handlerName)

	return nil
}
