package tempolite

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executionunit"
	"github.com/davidroman0O/go-tempolite/ent/task"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
)

type TaskPool struct {
	pool *retrypool.Pool[*ent.Task]
	tp   *Tempolite
	typ  string
}

func NewTaskPool(tp *Tempolite, count int, typ string) *TaskPool {
	taskPool := &TaskPool{
		tp:  tp,
		typ: typ,
	}

	workers := make([]retrypool.Worker[*ent.Task], count)
	for i := 0; i < count; i++ {
		workers[i] = &TaskWorker{ID: i, tp: tp, typ: typ}
	}

	opts := []retrypool.Option[*ent.Task]{
		retrypool.WithOnTaskSuccess[*ent.Task](taskPool.onTaskSuccess),
		retrypool.WithOnTaskFailure[*ent.Task](taskPool.onTaskFailure),
		retrypool.WithOnRetry[*ent.Task](taskPool.onTaskRetry),
		retrypool.WithPanicHandler[*ent.Task](taskPool.onTaskPanic),
		retrypool.WithOnNewDeadTask[*ent.Task](taskPool.onNewDeadTask),
	}

	taskPool.pool = retrypool.New(tp.ctx, workers, opts...)

	return taskPool
}

type TaskWorker struct {
	ID  int
	tp  *Tempolite
	typ string
}

func (w *TaskWorker) Run(ctx context.Context, job *ent.Task) error {
	log.Printf("Running %s task with ID %s on worker %d", w.typ, job.ID, w.ID)

	// Use a new context for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the task has already been processed
	latestJob, err := w.tp.client.Task.Get(dbCtx, job.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("task not found: %s", job.ID)
		}
		return fmt.Errorf("failed to fetch latest task state: %w", err)
	}

	if latestJob.Status != task.StatusPending && latestJob.Status != task.StatusInProgress {
		log.Printf("Task %s has already been processed. Current status: %s", job.ID, latestJob.Status)
		return nil
	}

	execUnit, err := latestJob.QueryExecutionUnit().Only(dbCtx)
	if err != nil {
		return w.updateTaskFailure(dbCtx, latestJob, nil, fmt.Errorf("failed to fetch execution unit for task %s: %w", latestJob.ID, err))
	}

	handlerInfo, exists := w.tp.getHandler(latestJob.HandlerName)
	if !exists {
		return w.updateTaskFailure(dbCtx, latestJob, execUnit, fmt.Errorf("no handler registered with name: %s", latestJob.HandlerName))
	}

	param, err := handlerInfo.toInterface(latestJob.Payload)
	if err != nil {
		return w.updateTaskFailure(dbCtx, latestJob, execUnit, fmt.Errorf("failed to unmarshal task payload: %w", err))
	}

	handlerCtx := HandlerContext{
		Context:            ctx,
		tp:                 w.tp,
		executionUnitID:    execUnit.ID,
		executionContextID: execUnit.QueryExecutionContext().OnlyX(dbCtx).ID,
	}

	result, err := callHandler(handlerInfo, handlerCtx, param)
	if err != nil {
		return w.updateTaskFailure(dbCtx, latestJob, execUnit, err)
	}

	return w.updateTaskResult(dbCtx, latestJob, execUnit, result)
}

func (w *TaskWorker) updateTaskFailure(ctx context.Context, job *ent.Task, execUnit *ent.ExecutionUnit, err error) error {
	tx, txErr := w.tp.client.Tx(ctx)
	if txErr != nil {
		log.Printf("Failed to start transaction: %v", txErr)
		return fmt.Errorf("failed to start transaction: %w", txErr)
	}
	defer tx.Rollback()

	errorBytes, _ := json.Marshal(err.Error())

	updatedJob, updateErr := tx.Task.UpdateOne(job).
		SetStatus(task.StatusFailed).
		SetError(errorBytes).
		SetCompletedAt(time.Now()).
		Save(ctx)
	if updateErr != nil {
		log.Printf("Failed to update task: %v", updateErr)
		return fmt.Errorf("failed to update task: %w", updateErr)
	}

	if execUnit != nil {
		_, updateErr = tx.ExecutionUnit.UpdateOneID(execUnit.ID).
			SetStatus(executionunit.StatusFailed).
			SetEndTime(time.Now()).
			Save(ctx)
		if updateErr != nil {
			log.Printf("Failed to update execution unit: %v", updateErr)
			return fmt.Errorf("failed to update execution unit: %w", updateErr)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("Failed to commit transaction: %v", commitErr)
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	log.Printf("Task %s failed: %v", updatedJob.ID, err)
	return err
}

func (w *TaskWorker) updateTaskResult(ctx context.Context, job *ent.Task, execUnit *ent.ExecutionUnit, result interface{}) error {
	tx, err := w.tp.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result: %w", err)
	}

	if _, err := tx.Task.UpdateOne(job).
		SetStatus(task.StatusCompleted).
		SetResult(resultBytes).
		SetCompletedAt(time.Now()).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if _, err := tx.ExecutionUnit.UpdateOne(execUnit).
		SetStatus(executionunit.StatusCompleted).
		SetEndTime(time.Now()).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to update execution unit: %w", err)
	}

	return tx.Commit()
}

func (p *TaskPool) onTaskSuccess(controller retrypool.WorkerController[*ent.Task], workerID int, worker retrypool.Worker[*ent.Task], task *retrypool.TaskWrapper[*ent.Task]) {
	log.Printf("%s task completed successfully: %s", p.typ, task.Data().ID)
}

func (p *TaskPool) onTaskFailure(controller retrypool.WorkerController[*ent.Task], workerID int, worker retrypool.Worker[*ent.Task], job *retrypool.TaskWrapper[*ent.Task], err error) retrypool.DeadTaskAction {
	log.Printf("%s task failed: %s, error: %v", p.typ, job.Data().ID, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := p.tp.client.Tx(ctx)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}
	defer tx.Rollback()

	failedTask, err := tx.Task.Get(ctx, job.Data().ID)
	if err != nil {
		log.Printf("Failed to get failed task: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	execUnit, err := failedTask.QueryExecutionUnit().Only(ctx)
	if err != nil {
		log.Printf("Failed to get execution unit: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if execUnit.RetryCount < execUnit.MaxRetries {
		execUnit, err = tx.ExecutionUnit.UpdateOneID(execUnit.ID).
			SetRetryCount(execUnit.RetryCount + 1).
			SetStatus(executionunit.StatusPending).
			Save(ctx)
		if err != nil {
			log.Printf("Failed to update execution unit: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		newJob, err := tx.Task.
			Create().
			SetID(uuid.New().String()).
			SetType(failedTask.Type).
			SetHandlerName(failedTask.HandlerName).
			SetPayload(failedTask.Payload).
			SetStatus(task.StatusPending).
			SetCreatedAt(time.Now()).
			SetExecutionUnitID(execUnit.ID).
			Save(ctx)
		if err != nil {
			log.Printf("Failed to create new task for retry: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		log.Printf("%s task %s scheduled for retry. Attempt %d/%d", p.typ, newJob.ID, execUnit.RetryCount, execUnit.MaxRetries)

		if err := p.Dispatch(newJob); err != nil {
			log.Printf("Failed to dispatch retry task: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		return retrypool.DeadTaskActionForceRetry
	}

	// Update the status of the execution unit to failed
	_, err = tx.ExecutionUnit.UpdateOne(execUnit).
		SetStatus(executionunit.StatusFailed).
		Save(ctx)
	if err != nil {
		log.Printf("Failed to update execution unit status: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}

	log.Printf("%s task %s has reached max retries. Marked as failed.", p.typ, job.Data().ID)
	return retrypool.DeadTaskActionAddToDeadTasks
}

func (p *TaskPool) onTaskRetry(attempt int, err error, task *retrypool.TaskWrapper[*ent.Task]) {
	log.Printf("%s task retrying: %s, attempt: %d, error: %v", p.typ, task.Data().ID, attempt, err)
}

func (p *TaskPool) onTaskPanic(task *ent.Task, v interface{}, stackTrace string) {
	log.Printf("%s task panicked: %s, panic: %v", p.typ, task.ID, v)
	log.Println(stackTrace)
}

func (p *TaskPool) onNewDeadTask(task *retrypool.DeadTask[*ent.Task]) {
	log.Printf("%s task is dead: %s", p.typ, task.Data.ID)
}

func (p *TaskPool) Dispatch(task *ent.Task) error {
	return p.pool.Dispatch(task)
}

func (p *TaskPool) QueueSize() int {
	return p.pool.QueueSize()
}

func (p *TaskPool) ProcessingCount() int {
	return p.pool.ProcessingCount()
}

func (p *TaskPool) DeadTaskCount() int {
	return p.pool.DeadTaskCount()
}

func callHandler(handlerInfo HandlerInfo, ctx HandlerContext, param interface{}) (interface{}, error) {
	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramValue := reflect.ValueOf(param)
	if paramValue.Type() != handlerInfo.ParamType {
		return nil, fmt.Errorf("parameter type mismatch: expected %v, got %v", handlerInfo.ParamType, paramValue.Type())
	}
	args := []reflect.Value{reflect.ValueOf(ctx), paramValue}
	results := handlerValue.Call(args)

	if len(results) == 0 {
		return nil, nil
	}

	var err error
	if errVal := results[len(results)-1]; !errVal.IsNil() {
		err = errVal.Interface().(error)
	}

	if len(results) == 2 && !results[0].IsNil() {
		return results[0].Interface(), err
	}

	return nil, err
}
