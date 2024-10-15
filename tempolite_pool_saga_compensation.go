package tempolite

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/retrypool"
)

type compensationTask[T Identifier] struct {
	ctx         CompensationContext[T]
	sagaID      string
	executionID string
	stepIndex   int
	handlerName string
	isLast      bool
	next        func() error
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

	return retrypool.New(tp.ctx, workers, opts...)
}

func (tp *Tempolite[T]) compensationOnSuccess(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]]) {
	log.Printf("Compensation step success: %d %s %s", workerID, task.Data().executionID, task.Data().handlerName)

	_, err := tp.client.SagaExecution.UpdateOneID(task.Data().executionID).
		SetStatus(sagaexecution.StatusCompleted).
		Save(tp.ctx)
	if err != nil {
		log.Printf("Failed to update saga execution status: %v", err)
	}

	if task.Data().isLast {
		_, err := tp.client.Saga.UpdateOneID(task.Data().sagaID).
			SetStatus(saga.StatusCompensated).
			Save(tp.ctx)
		if err != nil {
			log.Printf("Failed to update saga status to compensated: %v", err)
		}
	} else if task.Data().next != nil {
		if err := task.Data().next(); err != nil {
			log.Printf("Failed to dispatch next compensation task: %v", err)
		}
	}
}

func (tp *Tempolite[T]) compensationOnFailure(controller retrypool.WorkerController[*compensationTask[T]], workerID int, worker retrypool.Worker[*compensationTask[T]], task *retrypool.TaskWrapper[*compensationTask[T]], err error) retrypool.DeadTaskAction {
	log.Printf("Compensation step failed: %v", err)

	_, updateErr := tp.client.Saga.UpdateOneID(task.Data().sagaID).
		SetStatus(saga.StatusFailed).
		Save(tp.ctx)
	if updateErr != nil {
		log.Printf("Failed to update saga status to failed: %v", updateErr)
	}

	return retrypool.DeadTaskActionAddToDeadTasks
}

func (tp *Tempolite[T]) compensationOnPanic(task *compensationTask[T], v interface{}, stackTrace string) {
	log.Printf("Compensation task panicked: %v\nStack trace: %s", v, stackTrace)

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

func (tp *Tempolite[T]) compensationOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*compensationTask[T]]) {
	log.Printf("Compensation task retry attempt %d for saga execution %s: %v",
		attempt, task.Data().executionID, err)
}

type compensationWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w compensationWorker[T]) Run(ctx context.Context, data *compensationTask[T]) error {
	log.Printf("Executing compensation step: %s", data.handlerName)

	sagaHandlerInfo, ok := w.tp.sagas.Load(data.sagaID)
	if !ok {
		return fmt.Errorf("saga handler info not found for ID: %s", data.sagaID)
	}
	sagaDef := sagaHandlerInfo.(*SagaDefinition[T])
	step := sagaDef.Steps[data.stepIndex]

	result, err := step.Compensation(data.ctx)
	if err != nil {
		return err
	}

	log.Printf("Compensation step completed: %s, Result: %v", data.handlerName, result)

	return nil
}
