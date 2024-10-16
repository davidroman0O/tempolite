package tempolite

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/retrypool"
)

type transactionTask[T Identifier] struct {
	ctx         TransactionContext[T]
	sagaID      string
	executionID string
	stepIndex   int
	handlerName string
	isLast      bool
	next        func() error
	compensate  func() error
}

func (tp *Tempolite[T]) createTransactionPool() *retrypool.Pool[*transactionTask[T]] {
	opts := []retrypool.Option[*transactionTask[T]]{
		retrypool.WithAttempts[*transactionTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.transactionOnSuccess),
		retrypool.WithOnTaskFailure(tp.transactionOnFailure),
		retrypool.WithPanicHandler(tp.transactionOnPanic),
		retrypool.WithOnRetry(tp.transactionOnRetry),
	}

	workers := []retrypool.Worker[*transactionTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, transactionWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(tp.ctx, workers, opts...)
}

func (tp *Tempolite[T]) transactionOnSuccess(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]]) {
	log.Printf("Transaction step success: %d %s %s", workerID, task.Data().executionID, task.Data().handlerName)

	_, err := tp.client.SagaExecution.
		UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusCompleted).
		Save(tp.ctx)
	if err != nil {
		log.Printf("Failed to update saga execution status: %v", err)
	} else {
		log.Printf("Updated saga execution status to completed: %s", task.Data().executionID)
	}

	if task.Data().isLast {
		_, err := tp.client.Saga.
			UpdateOneID(task.Data().sagaID).
			SetStatus(saga.StatusCompleted).
			Save(tp.ctx)
		if err != nil {
			log.Printf("Failed to update saga status to completed: %v", err)
		} else {
			log.Printf("Updated saga status to completed: %s", task.Data().sagaID)
		}
	} else if task.Data().next != nil {
		if err := task.Data().next(); err != nil {
			log.Printf("Failed to dispatch next transaction task: %v", err)
		} else {
			log.Printf("Dispatched next transaction task: %s", task.Data().handlerName)
		}
	}
}

func (tp *Tempolite[T]) transactionOnFailure(controller retrypool.WorkerController[*transactionTask[T]], workerID int, worker retrypool.Worker[*transactionTask[T]], task *retrypool.TaskWrapper[*transactionTask[T]], err error) retrypool.DeadTaskAction {
	log.Printf("Transaction step failed: %v", err)

	_, updateErr := tp.client.SagaExecution.UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusFailed).
		Save(tp.ctx)
	if updateErr != nil {
		log.Printf("Failed to update saga execution status: %v", updateErr)
	}

	if task.Data().compensate != nil {
		if err := task.Data().compensate(); err != nil {
			log.Printf("Failed to dispatch compensation task: %v", err)
		}
	}

	return retrypool.DeadTaskActionDoNothing
}

func (tp *Tempolite[T]) transactionOnPanic(task *transactionTask[T], v interface{}, stackTrace string) {
	log.Printf("Transaction task panicked: %v\nStack trace: %s", v, stackTrace)

	_, err := tp.client.SagaExecution.UpdateOneID(task.executionID).
		SetStatus(sagaexecution.StatusFailed).
		Save(tp.ctx)
	if err != nil {
		log.Printf("Failed to update saga execution status after panic: %v", err)
	}

	_, err = tp.client.Saga.UpdateOneID(task.sagaID).
		SetStatus(saga.StatusFailed).
		Save(tp.ctx)
	if err != nil {
		log.Printf("Failed to update saga status to failed after panic: %v", err)
	}
}

func (tp *Tempolite[T]) transactionOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*transactionTask[T]]) {
	log.Printf("Transaction task retry attempt %d for saga execution %s: %v",
		attempt, task.Data().executionID, err)
}

type transactionWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w transactionWorker[T]) Run(ctx context.Context, data *transactionTask[T]) error {
	log.Printf("Executing transaction step: %s", data.handlerName)

	sagaHandlerInfo, ok := w.tp.sagas.Load(data.sagaID)
	if !ok {
		return fmt.Errorf("saga handler info not found for ID: %s", data.sagaID)
	}

	sagaDef := sagaHandlerInfo.(*SagaDefinition[T])
	step := sagaDef.Steps[data.stepIndex]

	result, err := step.Transaction(data.ctx)
	if err != nil {
		return err
	}

	log.Printf("Transaction step completed: %s, Result: %v", data.handlerName, result)

	return nil
}
