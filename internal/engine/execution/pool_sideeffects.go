package execution

import (
	"context"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

type SideEffectRequest struct{}
type SideEffectReponse struct{}

type PoolSideEffects struct {
	ctx   context.Context
	queue string
}

func NewSideEffectsExecutor(ctx context.Context, queue string) PoolSideEffects {
	p := PoolSideEffects{
		ctx:   ctx,
		queue: queue,
	}
	return p
}

func (w PoolSideEffects) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]]) {

}

func (w PoolSideEffects) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]]) {

}

func (w PoolSideEffects) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]], err error) retrypool.DeadTaskAction {
	return retrypool.DeadTaskActionDoNothing
}

func (w PoolSideEffects) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]]) {

}

func (w PoolSideEffects) OnPanic(task *retrypool.RequestResponse[SideEffectRequest, SideEffectReponse], v interface{}, stackTrace string) {

}

func (w PoolSideEffects) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {

}

type WorkerSideEffects struct {
	ctx   context.Context
	queue string
}

func NewSideEffectsWorker(ctx context.Context, queue string, registry *registry.Registry) WorkerSideEffects {
	w := WorkerSideEffects{
		ctx:   ctx,
		queue: queue,
	}
	return w
}

func (w WorkerSideEffects) Run(ctx context.Context, data *retrypool.RequestResponse[SideEffectRequest, SideEffectReponse]) error {
	return nil
}
