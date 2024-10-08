package tempolite

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/taskcontext"
	"github.com/davidroman0O/retrypool"
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

// Quite simple, we get the handler function, get the type of the second parameter, create a new instance of it, unmarshal the payload into it, and call the handler function with the HandlerContext and the parameter.
// Then simply marshal the result and error, and update the task with the result and error. Let the scheduler work!
func (h *HandlerWorker) Run(ctx context.Context, task *ent.HandlerTask) error {
	log.Printf("Running task with ID %s on worker %d", task.ID, h.ID)

	handlerInfo, exists := h.tp.getHandler(task.HandlerName)
	if !exists {
		log.Printf("No handler registered with name: %s", task.HandlerName)
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	/// When we panic, we tends to: failure + retry + panic
	/// When we fail, we tends to: failure + retry

	param, err := handlerInfo.ToInterface(task.Payload)
	if err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	if task.Edges.ExecutionContext == nil {
		return fmt.Errorf("task with ID %s has no execution context", task.ID)
	}

	if task.Edges.Node == nil {
		return fmt.Errorf("task with ID %s has no node", task.ID)
	}

	handlerCtx := HandlerContext{
		Context:            ctx,
		tp:                 h.tp,
		taskID:             task.ID,
		nodeID:             task.Edges.Node.ID,
		executionContextID: task.Edges.ExecutionContext.ID,
	}

	log.Printf("Calling handler for task ID %s", task.ID)
	results := handlerInfo.GetFn().Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	if len(results) > 0 && !results[len(results)-1].IsNil() {
		log.Printf("Handler for task ID %s returned an error: %v", task.ID, results[len(results)-1].Interface())
		return results[len(results)-1].Interface().(error)
	}

	var res interface{}
	var errRes error
	// it should be error
	if len(results) == 1 {
		if !results[0].IsNil() {
			errRes = results[0].Interface().(error)
		}
	} else if len(results) == 2 {
		// should be data + error
		if !results[0].IsNil() {
			res = results[0].Interface()
		}
		if !results[1].IsNil() {
			errRes = results[1].Interface().(error)
		}
	} else {
		return fmt.Errorf("handler function must return 1 or 2 values")
	}

	log.Printf("Task with ID %s completed successfully on worker %d", task.ID, h.ID)
	log.Printf("Task result: %v", res)
	log.Printf("Task error: %v", errRes)

	var bytesRes []byte
	if bytesRes, err = json.Marshal(res); err != nil {
		log.Printf("Failed to marshal task result: %v", err)
		return fmt.Errorf("failed to marshal task result: %v", err)
	}

	var bytesErr []byte
	if bytesErr, err = json.Marshal(errRes); err != nil {
		log.Printf("Failed to marshal task error: %v", err)
		return fmt.Errorf("failed to marshal task error: %v", err)
	}

	if err = h.tp.
		client.
		HandlerTask.
		Update().
		SetStatus(handlertask.StatusCompleted).
		SetError(bytesErr).
		SetResult(bytesRes).
		Where(
			handlertask.ID(task.ID),
		).
		Exec(ctx); err != nil {
		log.Printf("Failed to update task with ID %s: %v", task.ID, err)
	}

	return nil
}

func (tp *Tempolite) onHandlerTaskSuccess(controller retrypool.WorkerController[*ent.HandlerTask], workerID int, worker retrypool.Worker[*ent.HandlerTask], task *retrypool.TaskWrapper[*ent.HandlerTask]) {
	log.Printf("Task completed successfully")
}

func (tp *Tempolite) onHandlerTaskFailure(controller retrypool.WorkerController[*ent.HandlerTask], workerID int, worker retrypool.Worker[*ent.HandlerTask], task *retrypool.TaskWrapper[*ent.HandlerTask], err error) retrypool.DeadTaskAction {
	log.Printf("Task failed: %v", err)
	var parseErr error
	var bytesErr []byte

	if bytesErr, parseErr = json.Marshal(err); parseErr != nil {
		log.Printf("Failed to marshal task error: %v", parseErr)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	taskContext, err := tp.client.TaskContext.Query().Where(
		taskcontext.ID(task.Data().Edges.TaskContext.ID),
	).First(tp.ctx)
	if err != nil {
		log.Printf("Failed to get task context: %v", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	// TODO: test retry mechanism
	if taskContext.RetryCount < taskContext.MaxRetry {
		if err = tp.client.
			TaskContext.
			Update().
			SetRetryCount(taskContext.RetryCount + 1).
			Where(taskcontext.ID(taskContext.ID)).
			Exec(tp.ctx); err != nil {
			log.Printf("Failed to update task context: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}
		return retrypool.DeadTaskActionForceRetry
	}

	if err = tp.client.
		HandlerTask.
		Update().
		SetError(bytesErr).
		SetStatus(handlertask.StatusFailed).
		Where(handlertask.ID(task.Data().ID)).
		Exec(tp.ctx); err != nil {
		log.Printf("Failed to update task with ID %s: %v", task.Data().ID, err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionRetry
}

// Post failure, we retry the task
func (tp *Tempolite) onHandlerTaskRetry(attempt int, err error, task *retrypool.TaskWrapper[*ent.HandlerTask]) {
	// TODO: we need to create a new task on a new execution context??
	log.Printf("Task retrying: %v", err)
}

// Notification of panic when failure
func (tp *Tempolite) onHandlerTaskPanic(task *ent.HandlerTask, v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

// Notification of dead task, no retry anymore
func (tp *Tempolite) onDeadHandlerTask(task *retrypool.DeadTask[*ent.HandlerTask]) {
	log.Printf("Task is dead: %v", task)
}
