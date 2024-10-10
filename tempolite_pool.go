package tempolite

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/handlerexecution"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
)

type HandlerTaskPool struct {
	pool *retrypool.Pool[*ent.HandlerTask]
}

func NewHandlerTaskPool(tp *Tempolite, count int) *HandlerTaskPool {
	handlerPool := &HandlerTaskPool{}

	workers := make([]retrypool.Worker[*ent.HandlerTask], count)
	for i := 0; i < count; i++ {
		workers[i] = &HandlerWorker{ID: i, tp: tp}
		log.Printf("Created handler worker %d", i)
	}

	opts := []retrypool.Option[*ent.HandlerTask]{
		retrypool.WithOnTaskSuccess[*ent.HandlerTask](tp.onHandlerTaskSuccess),
		retrypool.WithOnTaskFailure[*ent.HandlerTask](tp.onHandlerTaskFailure),
		retrypool.WithOnRetry[*ent.HandlerTask](tp.onHandlerTaskRetry),
		retrypool.WithAttempts[*ent.HandlerTask](1),
		retrypool.WithPanicHandler[*ent.HandlerTask](tp.onHandlerTaskPanic),
		retrypool.WithOnNewDeadTask[*ent.HandlerTask](tp.onDeadHandlerTask),
	}

	handlerPool.pool = retrypool.New(tp.ctx, workers, opts...)

	return handlerPool
}

type HandlerWorker struct {
	ID int
	tp *Tempolite
}

func (h *HandlerWorker) Run(ctx context.Context, task *ent.HandlerTask) error {
	switch task.TaskType {
	case "handler":
		return h.runHandler(ctx, task)
	case "side_effect":
		return h.runSideEffect(ctx, task)
	case "transaction":
		return h.runTransaction(ctx, task)
	case "compensation":
		return h.runCompensation(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}
}

func (h *HandlerWorker) runHandler(ctx context.Context, task *ent.HandlerTask) error {
	log.Printf("Running handler task with ID %s on worker %d", task.ID, h.ID)

	handlerExec, err := h.tp.client.HandlerExecution.Query().
		Where(handlerexecution.HasTasksWith(handlertask.ID(task.ID))).
		WithExecutionContext().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch handler execution for task %s: %w", task.ID, err)
	}

	if handlerExec.Edges.ExecutionContext == nil {
		return fmt.Errorf("handler execution %s has no associated execution context", handlerExec.ID)
	}

	handlerInfo, exists := h.tp.getHandler(task.HandlerName)
	if !exists {
		return h.updateTaskFailure(ctx, task, fmt.Errorf("no handler registered with name: %s", task.HandlerName))
	}

	param, err := handlerInfo.ToInterface(task.Payload)
	if err != nil {
		return h.updateTaskFailure(ctx, task, fmt.Errorf("failed to unmarshal task payload: %v", err))
	}

	handlerCtx := HandlerContext{
		Context:            ctx,
		tp:                 h.tp,
		handlerExecutionID: handlerExec.ID,
		executionContextID: handlerExec.Edges.ExecutionContext.ID,
	}

	log.Printf("Calling handler for task ID %s", task.ID)
	results := handlerInfo.GetFn().Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	var res interface{}
	var errRes error
	if len(results) > 0 && !results[len(results)-1].IsNil() {
		errRes = results[len(results)-1].Interface().(error)
	}
	if len(results) == 2 && !results[0].IsNil() {
		res = results[0].Interface()
	}

	if errRes != nil {
		return h.updateTaskFailure(ctx, task, errRes)
	}

	return h.updateTaskResult(ctx, task, handlerExec, res, nil)
}

func (h *HandlerWorker) runSideEffect(ctx context.Context, task *ent.HandlerTask) error {
	handlerExec, err := h.tp.client.HandlerExecution.Query().
		Where(handlerexecution.HasTasksWith(handlertask.ID(task.ID))).
		WithExecutionContext().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch handler execution for task %s: %w", task.ID, err)
	}

	if handlerExec.Edges.ExecutionContext == nil {
		return fmt.Errorf("handler execution %s has no associated execution context", handlerExec.ID)
	}

	handlerInfo, exists := h.tp.getHandler(task.HandlerName)
	if !exists {
		return h.updateTaskFailure(ctx, task, fmt.Errorf("no handler registered with name: %s", task.HandlerName))
	}

	sideEffect, ok := handlerInfo.Handler.(SideEffect)
	if !ok {
		return h.updateTaskFailure(ctx, task, fmt.Errorf("handler %s is not a SideEffect", task.HandlerName))
	}

	sideEffectCtx := SideEffectContext{
		Context:            ctx,
		tp:                 h.tp,
		handlerExecutionID: handlerExec.ID,
		executionContextID: handlerExec.Edges.ExecutionContext.ID,
	}

	res, err := sideEffect.Run(sideEffectCtx)
	if err != nil {
		return h.updateTaskFailure(ctx, task, err)
	}

	resultBytes, _ := json.Marshal(res)

	// Save the result to SideEffectResult
	_, err = h.tp.client.SideEffectResult.
		Create().
		SetID(uuid.New().String()).
		SetExecutionContextID(handlerExec.Edges.ExecutionContext.ID).
		SetName(task.HandlerName).
		SetResult(resultBytes).
		Save(ctx)
	if err != nil {
		return h.updateTaskFailure(ctx, task, fmt.Errorf("failed to save side effect result: %v", err))
	}

	return h.updateTaskResult(ctx, task, handlerExec, res, nil)
}

func (h *HandlerWorker) runTransaction(ctx context.Context, task *ent.HandlerTask) error {
	// Implement transaction logic for SagaStep here
	return nil // Placeholder
}

func (h *HandlerWorker) runCompensation(ctx context.Context, task *ent.HandlerTask) error {
	// Implement compensation logic for SagaStep here
	return nil // Placeholder
}

func (h *HandlerWorker) updateTaskResult(ctx context.Context, task *ent.HandlerTask, handlerExec *ent.HandlerExecution, result interface{}, handlerErr error) error {
	tx, err := h.tp.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	status := handlertask.StatusCompleted
	if handlerErr != nil {
		status = handlertask.StatusFailed
	}

	resultBytes, _ := json.Marshal(result)
	var errorBytes []byte
	if handlerErr != nil {
		errorBytes, _ = json.Marshal(handlerErr.Error())
	}

	_, err = tx.HandlerTask.UpdateOne(task).
		SetStatus(status).
		SetResult(resultBytes).
		SetError(errorBytes).
		SetCompletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	_, err = tx.HandlerExecution.UpdateOne(handlerExec).
		SetStatus(handlerexecution.Status(status.String())).
		SetEndTime(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update handler execution: %w", err)
	}

	return tx.Commit()
}

func (h *HandlerWorker) updateTaskFailure(ctx context.Context, task *ent.HandlerTask, err error) error {
	errorBytes, _ := json.Marshal(err.Error())

	tx, err := h.tp.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.HandlerTask.UpdateOne(task).
		SetStatus(handlertask.StatusFailed).
		SetError(errorBytes).
		SetCompletedAt(time.Now()).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if _, err := tx.HandlerExecution.UpdateOneID(task.Edges.HandlerExecution.ID).
		SetStatus(handlerexecution.StatusFailed).
		SetEndTime(time.Now()).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to update handler execution: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After updating the task as failed, we return the original error
	// This will trigger the retry mechanism in onHandlerTaskFailure
	return err
}

func (tp *Tempolite) onHandlerTaskSuccess(controller retrypool.WorkerController[*ent.HandlerTask], workerID int, worker retrypool.Worker[*ent.HandlerTask], task *retrypool.TaskWrapper[*ent.HandlerTask]) {
	log.Printf("Task completed successfully %s", task.Data().HandlerName)
}

func (tp *Tempolite) onHandlerTaskFailure(controller retrypool.WorkerController[*ent.HandlerTask], workerID int, worker retrypool.Worker[*ent.HandlerTask], task *retrypool.TaskWrapper[*ent.HandlerTask], err error) retrypool.DeadTaskAction {
	log.Printf("Task failed: %v", err)

	handlerExec, err := tp.client.HandlerExecution.Query().
		Where(handlerexecution.HasTasksWith(handlertask.ID(task.Data().ID))).
		Only(tp.ctx)
	if err != nil {
		log.Printf("Failed to get handler execution: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if handlerExec.RetryCount < handlerExec.MaxRetries {
		updatedHandlerExec, err := tp.client.HandlerExecution.
			UpdateOne(handlerExec).
			SetRetryCount(handlerExec.RetryCount + 1).
			SetStatus(handlerexecution.StatusPending).
			Save(tp.ctx)
		if err != nil {
			log.Printf("Failed to update handler execution: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		newTask, err := tp.client.HandlerTask.
			Create().
			SetHandlerExecution(updatedHandlerExec).
			SetHandlerName(task.Data().HandlerName).
			SetPayload(task.Data().Payload).
			SetTaskType(task.Data().TaskType).
			SetStatus(handlertask.StatusPending).
			SetCreatedAt(time.Now()).
			Save(tp.ctx)
		if err != nil {
			log.Printf("Failed to create new task for retry: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		log.Printf("Task %s scheduled for retry. Attempt %d/%d", newTask.ID, updatedHandlerExec.RetryCount, updatedHandlerExec.MaxRetries)

		// Dispatch the new task
		if err := tp.handlerTaskPool.pool.Dispatch(newTask); err != nil {
			log.Printf("Failed to dispatch retry task: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		return retrypool.DeadTaskActionForceRetry
	}

	log.Printf("Task %s has reached max retries. Marked as failed.", task.Data().ID)
	return retrypool.DeadTaskActionAddToDeadTasks
}

func (tp *Tempolite) onHandlerTaskRetry(attempt int, err error, task *retrypool.TaskWrapper[*ent.HandlerTask]) {
	log.Printf("Task retrying: %v", err)
}

func (tp *Tempolite) onHandlerTaskPanic(task *ent.HandlerTask, v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite) onDeadHandlerTask(task *retrypool.DeadTask[*ent.HandlerTask]) {
	log.Printf("Task is dead: %v", task)
}
