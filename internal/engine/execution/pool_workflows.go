package execution

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/davidroman0O/retrypool"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/cq/commands"
	"github.com/davidroman0O/tempolite/internal/engine/cq/queries"
	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type WorkflowRequest struct {
	WorkflowInfo  *repository.WorkflowInfo
	Retry         func() error
	PauseDetected bool
}

type WorkflowReponse struct{}

type PoolWorkflows struct {
	ctx   context.Context
	queue string
	db    repository.Repository
}

func NewWorkflowsExecutor(ctx context.Context, queue string, db repository.Repository) PoolWorkflows {
	logs.Debug(ctx, "Create new workflow executor")
	p := PoolWorkflows{
		ctx:   ctx,
		queue: queue,
		db:    db,
	}
	return p
}

func (w PoolWorkflows) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {
	logs.Debug(w.ctx, "OnRetry workflow task", "attempt", attempt, "error", err)
}

func (w PoolWorkflows) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {
	logs.Debug(w.ctx, "OnSuccess workflow task", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
	if task.Data().Request.PauseDetected {
		tx, err := w.db.Tx()
		if err != nil {
			logs.Error(w.ctx, "OnSuccess workflow error on transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return
		}
		if err := w.db.Workflows().UpdateExecutionPaused(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
			tx.Rollback()
			logs.Error(w.ctx, "OnSuccess workflow error on update execution paused", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return
		}
		if err := tx.Commit(); err != nil {
			logs.Error(w.ctx, "OnSuccess workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return
		}
		logs.Debug(w.ctx, "OnSuccess workflow execution paused", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return
	}

	tx, err := w.db.Tx()
	if err != nil {
		logs.Error(w.ctx, "OnSuccess workflow error on transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return
	}
	if err := w.db.Workflows().UpdateExecutionSuccess(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
		tx.Rollback()
		logs.Error(w.ctx, "OnSuccess workflow error on update execution success", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return
	}
	if err := tx.Commit(); err != nil {
		logs.Error(w.ctx, "OnSuccess workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return
	}
	logs.Debug(w.ctx, "OnSuccess workflow execution success", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
}

func (w PoolWorkflows) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], err error) retrypool.DeadTaskAction {

	logs.Debug(w.ctx, "OnFailure workflow task", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
	var state *schema.RetryState
	var policy schema.RetryPolicy = task.Data().Request.WorkflowInfo.Data.RetryPolicy

	tx, err := w.db.Tx()
	if err != nil {
		logs.Error(w.ctx, "OnFailure workflow error on transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if state, err = w.db.Workflows().GetRetryState(tx, task.Data().Request.WorkflowInfo.ID); err != nil {
		tx.Rollback()
		logs.Error(w.ctx, "OnFailure workflow error on get retry state", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if state.Attempts <= policy.MaxAttempts {

		if err = w.db.Workflows().IncrementRetryAttempt(tx, task.Data().Request.WorkflowInfo.ID); err != nil {
			tx.Rollback()
			logs.Error(w.ctx, "OnFailure workflow error on increment retry attempt", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err = w.db.Workflows().UpdateExecutionRetried(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
			tx.Rollback()
			logs.Error(w.ctx, "OnFailure workflow error on update execution retried", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		// Trigger the attempt to retry
		if err = task.Data().Request.Retry(); err != nil {
			tx.Rollback()
			logs.Error(w.ctx, "OnFailure workflow error on retry", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err = tx.Commit(); err != nil {
			logs.Error(w.ctx, "OnFailure workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		logs.Debug(w.ctx, "OnFailure workflow execution retried", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		// We literally do nothing, we already dispatched a new workflow execution
		return retrypool.DeadTaskActionDoNothing
	}

	if err = w.db.Workflows().UpdateExecutionFailed(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
		tx.Rollback()
		logs.Error(w.ctx, "OnFailure workflow error on update execution failed", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err = tx.Commit(); err != nil {
		logs.Error(w.ctx, "OnFailure workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	logs.Error(w.ctx, "OnFailure workflow execution failed", "queue", w.queue, "handlerName", task.Data().Request.WorkflowInfo.HandlerName, "workerID", workerID, "workflowID", task.Data().Request.WorkflowInfo.ID, "executionID", task.Data().Request.WorkflowInfo.Execution.ID, "stepID", task.Data().Request.WorkflowInfo.StepID)
	return retrypool.DeadTaskActionDoNothing
}

func (w PoolWorkflows) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {
	logs.Debug(w.ctx, "OnDeadTask workflow task", "queue", w.queue, "handlerName", task.Data.Request.WorkflowInfo.HandlerName, "workflowID", task.Data.Request.WorkflowInfo.ID, "executionID", task.Data.Request.WorkflowInfo.Execution.ID, "stepID", task.Data.Request.WorkflowInfo.StepID)
}

func (w PoolWorkflows) OnPanic(task *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse], v interface{}, stackTrace string) {
	logs.Debug(w.ctx, "OnPanic workflow task", "queue", w.queue, "handlerName", task.Request.WorkflowInfo.HandlerName, "workflowID", task.Request.WorkflowInfo.ID, "executionID", task.Request.WorkflowInfo.Execution.ID, "stepID", task.Request.WorkflowInfo.StepID)
}

func (w PoolWorkflows) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {
	logs.Debug(w.ctx, "OnExecutorPanic workflow task", "queue", w.queue, "workerID", worker)
}

type WorkerWorkflows struct {
	ctx      context.Context
	queue    string
	registry *registry.Registry
	db       repository.Repository

	commands *commands.Commands
	queries  *queries.Queries
}

func NewWorkflowsWorker(
	ctx context.Context,
	queue string,
	registry *registry.Registry,
	db repository.Repository,
	commands *commands.Commands,
	queries *queries.Queries,
) WorkerWorkflows {
	w := WorkerWorkflows{
		ctx:      ctx,
		queue:    queue,
		registry: registry,
		db:       db,
		commands: commands,
		queries:  queries,
	}
	return w
}

func (w WorkerWorkflows) Run(ctx context.Context, data *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]) error {

	logs.Debug(w.ctx, "Run workflow task", "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
	identity := types.HandlerIdentity(data.Request.WorkflowInfo.HandlerName)
	var workflow types.Workflow
	var err error

	logs.Debug(w.ctx, "Run workflow task getting workflow identity", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
	if workflow, err = w.registry.GetWorkflow(identity); err != nil {
		logs.Error(w.ctx, "Run workflow error on get workflow", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
		return err
	}

	logs.Debug(w.ctx, "Run workflow task creating new workflow context", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
	contextWorkflow := tempoliteContext.NewWorkflowContext(
		ctx,
		data.Request.WorkflowInfo.ID,
		data.Request.WorkflowInfo.Execution.ID,
		data.Request.WorkflowInfo.RunID,
		data.Request.WorkflowInfo.StepID,
		w.queue,
		identity,
		struct {
			*commands.Commands
			*queries.Queries
		}{
			w.commands,
			w.queries,
		},
	)

	logs.Debug(w.ctx, "Run workflow task converting inputs from serialization", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
	params, err := io.ConvertInputsFromSerialization(types.HandlerInfo(workflow), data.Request.WorkflowInfo.Data.Input)
	if err != nil {
		logs.Error(w.ctx, "Run workflow error on convert inputs from serialization", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
		return err
	}

	values := []reflect.Value{reflect.ValueOf(contextWorkflow)}

	for _, v := range params {
		values = append(values, reflect.ValueOf(v))
	}

	var txQueueUpdate *ent.Tx
	logs.Debug(w.ctx, "Worker Workflow Pending updating execution pending to running", "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
	if txQueueUpdate, err = w.db.Tx(); err != nil {
		logs.Error(w.ctx, "Worker Workflow Pending error creating transaction", "error", err, "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
		return err
	}
	if err := w.db.Workflows().UpdateExecutionPendingToRunning(txQueueUpdate, data.Request.WorkflowInfo.Execution.ID); err != nil {
		if errors.Is(err, context.Canceled) {
			logs.Debug(w.ctx, "Worker Workflow Pending update execution pending to running context canceled", "error", err, "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
			return nil
		}
		if err := txQueueUpdate.Rollback(); err != nil {
			logs.Error(w.ctx, "Worker Workflow Pending error rolling back transaction", "error", err, "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
			return err
		}
		logs.Error(w.ctx, "Worker Workflow Pending error updating execution pending to running", "error", err, "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
		return err
	}

	if ctx.Err() != nil {
		logs.Debug(w.ctx, "Worker Workflow Pending context canceled", "error", ctx.Err(), "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
		txQueueUpdate.Rollback()
		return ctx.Err()
	}

	if err := txQueueUpdate.Commit(); err != nil {
		logs.Error(w.ctx, "Worker Workflow Pending error committing transaction", "error", err, "queue", w.queue, "workflowID", data.Request.WorkflowInfo.ID, "stepID", data.Request.WorkflowInfo.StepID, "queueID", data.Request.WorkflowInfo.QueueID, "runID", data.Request.WorkflowInfo.RunID)
		return err
	}

	logs.Debug(w.ctx, "Run workflow task calling handler", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
	fmt.Println("\t Calling handler", workflow.HandlerName, "with values", values)
	returnedValues := reflect.ValueOf(workflow.Handler).Call(values)

	fmt.Println("\t Returned values", returnedValues, "from", workflow.HandlerName)

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

	// If pause detected, there is nothing to update YET
	if errors.Is(errRes, types.ErrWorkflowPaused) {
		data.Request.PauseDetected = true
		logs.Debug(w.ctx, "Run workflow task pause detected", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
		return nil
	}

	// Classic update of the execution data
	// Error or output

	if errRes != nil {
		{
			tx, err := w.db.Tx()
			if err != nil {
				logs.Error(w.ctx, "Run workflow error on transaction", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
				return err
			}

			if err := w.db.Workflows().UpdateExecutionDataError(tx, data.Request.WorkflowInfo.Execution.ID, errRes.Error()); err != nil {
				tx.Rollback()
				logs.Error(w.ctx, "Run workflow error on update execution data error", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
				return err
			}

			if err := tx.Commit(); err != nil {
				logs.Error(w.ctx, "Run workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
				return err
			}

			logs.Debug(w.ctx, "Run workflow execution error", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
			return errRes
		}
	}

	output, err := io.ConvertOutputsForSerialization(res)
	if err != nil {
		logs.Error(w.ctx, "Run workflow error on convert outputs for serialization", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
		return err
	}

	{
		tx, err := w.db.Tx()
		if err != nil {
			logs.Error(w.ctx, "Run workflow error on transaction", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
			return err
		}

		if err := w.db.Workflows().UpdateExecutionDataOuput(tx, data.Request.WorkflowInfo.Execution.ID, output); err != nil {
			tx.Rollback()
			logs.Error(w.ctx, "Run workflow error on update execution data output", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
			return err
		}

		if err := tx.Commit(); err != nil {
			logs.Error(w.ctx, "Run workflow error on commit transaction", "error", err, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)
			return err
		}
	}

	logs.Debug(w.ctx, "Run workflow execution success", "identity", identity, "queue", w.queue, "handlerName", data.Request.WorkflowInfo.HandlerName, "workflowID", data.Request.WorkflowInfo.ID, "executionID", data.Request.WorkflowInfo.Execution.ID, "stepID", data.Request.WorkflowInfo.StepID)

	return nil
}
