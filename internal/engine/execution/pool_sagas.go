package execution

import (
	"context"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

type SagaRequest struct{}
type SagaReponse struct{}

type PoolSagas struct {
	ctx   context.Context
	queue string
}

func NewSagasExecutor(ctx context.Context, queue string) PoolSagas {
	p := PoolSagas{
		ctx:   ctx,
		queue: queue,
	}
	return p
}

func (w PoolSagas) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[SagaRequest, SagaReponse]]) {

}

func (w PoolSagas) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[SagaRequest, SagaReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[SagaRequest, SagaReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[SagaRequest, SagaReponse]]) {

}

func (w PoolSagas) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[SagaRequest, SagaReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[SagaRequest, SagaReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[SagaRequest, SagaReponse]], err error) retrypool.DeadTaskAction {
	return retrypool.DeadTaskActionDoNothing
}

func (w PoolSagas) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[SagaRequest, SagaReponse]]) {

}

func (w PoolSagas) OnPanic(task *retrypool.RequestResponse[SagaRequest, SagaReponse], v interface{}, stackTrace string) {

}

func (w PoolSagas) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {

}

type WorkerSagas struct {
	ctx   context.Context
	queue string
}

func NewSagasWorker(ctx context.Context, queue string, registry *registry.Registry) WorkerSagas {
	w := WorkerSagas{
		ctx:   ctx,
		queue: queue,
	}
	return w
}

func (w WorkerSagas) Run(ctx context.Context, data *retrypool.RequestResponse[SagaRequest, SagaReponse]) error {
	return nil
}
