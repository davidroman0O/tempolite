package execution

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/davidroman0O/retrypool"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
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
	p := PoolWorkflows{
		ctx:   ctx,
		queue: queue,
		db:    db,
	}
	return p
}

func (w PoolWorkflows) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {

}

func (w PoolWorkflows) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {
	if task.Data().Request.PauseDetected {
		tx, err := w.db.Tx()
		if err != nil {
			return
		}
		if err := w.db.Workflows().UpdateExecutionPaused(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
			tx.Rollback()
			return
		}
		if err := tx.Commit(); err != nil {
			return
		}
		return
	}
	fmt.Println("Success", task.Data().Request.WorkflowInfo.ID)
	tx, err := w.db.Tx()
	if err != nil {
		return
	}
	if err := w.db.Workflows().UpdateExecutionSuccess(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
		tx.Rollback()
		return
	}
	if err := tx.Commit(); err != nil {
		return
	}
}

func (w PoolWorkflows) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], err error) retrypool.DeadTaskAction {

	var state *schema.RetryState
	var policy schema.RetryPolicy = task.Data().Request.WorkflowInfo.Data.RetryPolicy

	tx, err := w.db.Tx()
	if err != nil {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if state, err = w.db.Workflows().GetRetryState(tx, task.Data().Request.WorkflowInfo.ID); err != nil {
		tx.Rollback()
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if state.Attempts <= policy.MaxAttempts {

		if err = w.db.Workflows().IncrementRetryAttempt(tx, task.Data().Request.WorkflowInfo.ID); err != nil {
			tx.Rollback()
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err = w.db.Workflows().UpdateExecutionRetried(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
			tx.Rollback()
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		// Trigger the attempt to retry
		if err = task.Data().Request.Retry(); err != nil {
			tx.Rollback()
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		if err = tx.Commit(); err != nil {
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		// We literally do nothing, we already dispatched a new workflow execution
		return retrypool.DeadTaskActionDoNothing
	}

	if err = w.db.Workflows().UpdateExecutionFailed(tx, task.Data().Request.WorkflowInfo.Execution.ID); err != nil {
		tx.Rollback()
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	if err = tx.Commit(); err != nil {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionDoNothing
}

func (w PoolWorkflows) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {

}

func (w PoolWorkflows) OnPanic(task *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse], v interface{}, stackTrace string) {

}

func (w PoolWorkflows) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {

}

type WorkerWorkflows struct {
	ctx      context.Context
	queue    string
	registry *registry.Registry
	db       repository.Repository
}

func NewWorkflowsWorker(ctx context.Context, queue string, registry *registry.Registry, db repository.Repository) WorkerWorkflows {
	w := WorkerWorkflows{
		ctx:      ctx,
		queue:    queue,
		registry: registry,
		db:       db,
	}
	return w
}

func (w WorkerWorkflows) Run(ctx context.Context, data *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]) error {

	identity := types.HandlerIdentity(data.Request.WorkflowInfo.HandlerName)
	var workflow types.Workflow
	var err error

	if workflow, err = w.registry.GetWorkflow(identity); err != nil {
		return err
	}

	contextWorkflow := tempoliteContext.NewWorkflowContext(
		ctx,
		data.Request.WorkflowInfo.ID,
		data.Request.WorkflowInfo.Execution.ID,
		data.Request.WorkflowInfo.RunID,
		data.Request.WorkflowInfo.StepID,
		w.queue,
		identity,
	)

	params, err := io.ConvertInputsFromSerialization(types.HandlerInfo(workflow), data.Request.WorkflowInfo.Data.Input)
	if err != nil {
		return err
	}

	values := []reflect.Value{reflect.ValueOf(contextWorkflow)}

	for _, v := range params {
		values = append(values, reflect.ValueOf(v))
	}

	returnedValues := reflect.ValueOf(workflow.Handler).Call(values)

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
		return nil
	}

	// Classic update of the execution data
	// Error or output

	if errRes != nil {
		tx, err := w.db.Tx()
		if err != nil {
			return err
		}

		fmt.Println(data.Request.WorkflowInfo.Execution.ID, errRes.Error())
		if err := w.db.Workflows().UpdateExecutionDataError(tx, data.Request.WorkflowInfo.Execution.ID, errRes.Error()); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		return errRes
	}

	output, err := io.ConvertOutputsForSerialization(res)
	if err != nil {
		return err
	}

	tx, err := w.db.Tx()
	if err != nil {
		return err
	}

	fmt.Println(data.Request.WorkflowInfo.Execution.ID, output)
	if err := w.db.Workflows().UpdateExecutionDataOuput(tx, data.Request.WorkflowInfo.Execution.ID, output); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Println("Workflow", data.Request.WorkflowInfo.Execution.ID, "executed successfully")

	return nil
}
