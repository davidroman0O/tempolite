package execution

import (
	"context"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

type ActivityRequest struct{}
type ActivityReponse struct{}

type PoolActivities struct {
	ctx   context.Context
	queue string
}

func NewActivitiesExecutor(ctx context.Context, queue string) PoolActivities {
	p := PoolActivities{
		ctx:   ctx,
		queue: queue,
	}
	return p
}

func (w PoolActivities) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]]) {

}

func (w PoolActivities) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]]) {

}

func (w PoolActivities) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]], err error) retrypool.DeadTaskAction {
	return retrypool.DeadTaskActionDoNothing
}

func (w PoolActivities) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[ActivityRequest, ActivityReponse]]) {

}

func (w PoolActivities) OnPanic(task *retrypool.RequestResponse[ActivityRequest, ActivityReponse], v interface{}, stackTrace string) {

}

func (w PoolActivities) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {

}

type WorkerActivities struct {
	ctx   context.Context
	queue string
}

func NewActivitiesWorker(ctx context.Context, queue string, registry *registry.Registry) WorkerActivities {
	w := WorkerActivities{
		ctx:   ctx,
		queue: queue,
	}
	return w
}

func (w WorkerActivities) Run(ctx context.Context, data *retrypool.RequestResponse[ActivityRequest, ActivityReponse]) error {
	return nil
}
