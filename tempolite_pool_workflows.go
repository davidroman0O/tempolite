package tempolite

import (
	"context"
	"fmt"
	"reflect"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

type workflowTask struct {
	ctx         WorkflowContext
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
	isPaused    bool
	queueName   string
}

func (tp *Tempolite) createWorkflowPool(queue string, countWorkers int) (*retrypool.Pool[*workflowTask], error) {

	opts := []retrypool.Option[*workflowTask]{
		retrypool.WithAttempts[*workflowTask](1),
		retrypool.WithOnTaskSuccess(tp.workflowOnSuccess),
		retrypool.WithOnTaskFailure(tp.workflowOnFailure),
		retrypool.WithPanicHandler(tp.workflowOnPanic),
		retrypool.WithOnRetry(tp.workflowOnRetry),
		retrypool.WithPanicWorker[*workflowTask](tp.workflowWorkerPanic),
	}

	tp.logger.Debug(tp.ctx, "creating workflow pool", "queue", queue, "workers", countWorkers)

	workers := []retrypool.Worker[*workflowTask]{}

	for i := 0; i < countWorkers; i++ {
		id, err := tp.getWorkerWorkflowID(queue)
		if err != nil {
			return nil, fmt.Errorf("failed to generate workflow worker ID: %w", err)
		}
		workers = append(workers, workflowWorker{id: id, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...), nil
}

func (tp *Tempolite) workflowWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	tp.logger.Debug(tp.ctx, "workflow pool worker panicked", "workerID", workerID, "error", err)
	tp.logger.Error(tp.ctx, "workflow pool worker panicked", "err", err, "stackTrace", stackTrace)
	fmt.Println(err)
	fmt.Println(stackTrace)
}

func (tp *Tempolite) workflowOnPanic(task *workflowTask, v interface{}, stackTrace string) {
	tp.logger.Debug(tp.ctx, "workflow pool task panicked", "task", task, "error", v)
	tp.logger.Error(tp.ctx, "workflow pool task panicked", "stackTrace", stackTrace)
	fmt.Println(v)
	fmt.Println(stackTrace)
}

func (tp *Tempolite) workflowOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*workflowTask]) {
	tp.logger.Debug(tp.ctx, "workflow pool task retry", "attempt", attempt, "error", err)
}

func (tp *Tempolite) workflowOnSuccess(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask]) {

	if task.Data().isPaused {
		tp.logger.Debug(tp.ctx, "workflow pool task paused", "workflowID", task.Data().ctx.workflowID)
		// Update workflow status to paused in the database
		tx, err := tp.client.Tx(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Failed to start transaction for pausing workflow", "workflowID", task.Data().ctx.workflowID, "error", err)
			return
		}
		if _, err := tx.Workflow.UpdateOneID(task.Data().ctx.workflowID).
			SetIsPaused(true).
			SetIsReady(false).
			Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Failed to update workflow pause status", "workflowID", task.Data().ctx.workflowID, "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return
		}
		if err := tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to commit transaction for pausing workflow", "workflowID", task.Data().ctx.workflowID, "error", err)
		}
		return
	}

	tp.logger.Debug(tp.ctx, "workflow pool task success", "workflowID", task.Data().ctx.workflowID, "executionID", task.Data().ctx.executionID)

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for workflow success", "workflowID", task.Data().ctx.workflowID, "error", err)
		return
	}

	if _, err := tx.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Failed to update workflow execution status", "executionID", task.Data().ctx.executionID, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if _, err := tx.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusCompleted).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Failed to update workflow status", "workflowID", task.Data().ctx.workflowID, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction for workflow success", "workflowID", task.Data().ctx.workflowID, "error", err)
	}
}

func (tp *Tempolite) workflowOnFailure(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask], taskErr error) retrypool.DeadTaskAction {

	tp.logger.Debug(tp.ctx, "workflow pool task failed", "workflowID", task.Data().ctx.workflowID, "executionID", task.Data().ctx.executionID, "retryCount", task.Data().retryCount, "maxRetry", task.Data().maxRetry)

	total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(task.Data().ctx.workflowID))).Count(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "workflow pool task failed to count workflow executions", "workflowID", task.Data().ctx.workflowID, "error", err)
	}

	total = total - 1 // removing myself

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().maxRetry > 0 && total < task.Data().maxRetry {
		tx, err := tp.client.Tx(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Failed to start transaction for workflow retry", "workflowID", task.Data().ctx.workflowID, "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		// Just by creating a new workflow execution, we're incrementing the total count of executions which is the retry count in the database
		if _, err := tx.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "workflow pool task: WorkflowExecution.Update failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if _, err := tx.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusRetried).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "workflow pool task: Workflow.UpdateOneID failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Failed to commit transaction for workflow retry", "workflowID", task.Data().ctx.workflowID, "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err := task.Data().retry(); err != nil {
			tp.logger.Error(tp.ctx, "workflow pool task: retry failed", "error", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		tp.logger.Debug(tp.ctx, "workflow pool task: retry success", "workflowID", task.Data().ctx.workflowID, "executionID", task.Data().ctx.executionID)

		return retrypool.DeadTaskActionForceRetry
	}

	// we won't re-try the workflow, probably was an error somewhere
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Failed to start transaction for workflow failure", "workflowID", task.Data().ctx.workflowID, "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if taskErr != nil {
		if _, err := tx.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusFailed).SetError(taskErr.Error()).Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "workflow pool task: WorkflowExecution.Update failed", "error", taskErr)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}
	} else {
		// not retrying and there is no error
		if _, err := tx.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusFailed).SetError("unknown error").Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "workflow pool task: WorkflowExecution.Update failed", "error", err)
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
			}
			return retrypool.DeadTaskActionAddToDeadTasks
		}
	}

	if _, err := tx.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusFailed).Save(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "workflow pool task: Workflow.UpdateOneID failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Failed to commit transaction for workflow failure", "workflowID", task.Data().ctx.workflowID, "error", err)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

type workflowWorker struct {
	id int
	tp *Tempolite
}

func (w workflowWorker) Run(ctx context.Context, data *workflowTask) error {

	w.tp.logger.Debug(data.ctx, "workflow pool worker run", "workflowID", data.ctx.workflowID, "executionID", data.ctx.executionID, "handler", data.handlerName)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}

	for _, v := range data.params {
		values = append(values, reflect.ValueOf(v))
	}

	returnedValues := reflect.ValueOf(data.handler).Call(values)

	var res []interface{}
	var errRes error
	if len(returnedValues) > 0 {
		res = make([]interface{}, len(returnedValues)-1)
		for i := 0; i < len(returnedValues)-1; i++ {
			res[i] = returnedValues[i].Interface()
		}
		if !returnedValues[len(returnedValues)-1].IsNil() {
			errRes = returnedValues[len(returnedValues)-1].Interface().(error)
		}
	}

	if errRes == errWorkflowPaused {
		data.isPaused = true
		w.tp.logger.Debug(data.ctx, "workflow pool worker paused", "workflowID", data.ctx.workflowID, "executionID", data.ctx.executionID)
		return nil
	}

	var value any
	var ok bool
	var workflowInfo Workflow

	if value, ok = w.tp.workflows.Load(data.handlerName); ok {
		if workflowInfo, ok = value.(Workflow); !ok {
			w.tp.logger.Error(data.ctx, "workflow pool worker: workflow not found", "workflowID", data.ctx.workflowID, "executionID", data.ctx.executionID, "handler", data.handlerName)
			return fmt.Errorf("workflow %s not found", data.handlerName)
		}
	} else {
		w.tp.logger.Error(data.ctx, "workflow pool worker: workflow not found", "workflowID", data.ctx.workflowID, "executionID", data.ctx.executionID, "handler", data.handlerName)
		return fmt.Errorf("workflow %s not found", data.handlerName)
	}

	serializableOutput, err := w.tp.convertOutputsForSerialization(HandlerInfo(workflowInfo), res)
	if err != nil {
		w.tp.logger.Error(data.ctx, "workflow pool worker: convertOutputsForSerialization failed", "error", err)
		return err
	}

	// fmt.Println("output to save", res, errRes)
	tx, err := w.tp.client.Tx(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(data.ctx, "Failed to start transaction for updating workflow execution output", "error", err)
		return err
	}

	if _, err := tx.WorkflowExecution.UpdateOneID(data.ctx.executionID).SetOutput(serializableOutput).Save(w.tp.ctx); err != nil {
		w.tp.logger.Error(data.ctx, "workflowWorker: WorkflowExecution.Update failed", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			w.tp.logger.Error(data.ctx, "Failed to rollback transaction", "error", rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		w.tp.logger.Error(data.ctx, "Failed to commit transaction for updating workflow execution output", "error", err)
		return err
	}

	return errRes
}
