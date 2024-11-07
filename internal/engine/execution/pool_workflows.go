package execution

import (
	"context"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

type WorkflowRequest struct{}
type WorkflowReponse struct{}

type PoolWorkflows struct {
	ctx   context.Context
	queue string
}

func NewWorkflowsExecutor(ctx context.Context, queue string) PoolWorkflows {
	p := PoolWorkflows{
		ctx:   ctx,
		queue: queue,
	}
	return p
}

func (w PoolWorkflows) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {

}

func (w PoolWorkflows) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {

}

func (w PoolWorkflows) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]], err error) retrypool.DeadTaskAction {
	return retrypool.DeadTaskActionDoNothing
}

func (w PoolWorkflows) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]]) {

}

func (w PoolWorkflows) OnPanic(task *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse], v interface{}, stackTrace string) {

}

func (w PoolWorkflows) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {

}

type WorkerWorkflows struct {
	ctx   context.Context
	queue string
}

func NewWorkflowsWorker(ctx context.Context, queue string, registry *registry.Registry) WorkerWorkflows {
	w := WorkerWorkflows{
		ctx:   ctx,
		queue: queue,
	}
	return w
}

func (w WorkerWorkflows) Run(ctx context.Context, data *retrypool.RequestResponse[WorkflowRequest, WorkflowReponse]) error {
	return nil
}
