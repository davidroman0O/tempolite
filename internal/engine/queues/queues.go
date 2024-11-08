package queues

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/execution"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"golang.org/x/sync/errgroup"
)

type Queue struct {
	ctx               context.Context
	workflowsWorker   *execution.WorkerPool[execution.WorkflowRequest, execution.WorkflowReponse]
	activitiesWorker  *execution.WorkerPool[execution.ActivityRequest, execution.ActivityReponse]
	sideEffectsWorker *execution.WorkerPool[execution.SideEffectRequest, execution.SideEffectReponse]
	sagasWorker       *execution.WorkerPool[execution.SagaRequest, execution.SagaReponse]
}

func New(
	ctx context.Context,
	queue string,
	registry *registry.Registry,
	db repository.Repository,
) (*Queue, error) {

	tx, err := db.Tx()
	if err != nil {
		return nil, err
	}

	if _, err = db.Queues().GetByName(tx, "default"); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			fmt.Println("Creating default queue")
			if _, errCreate := db.Queues().Create(tx, "default"); errCreate != nil {
				if errRollback := tx.Rollback(); errRollback != nil {
					return nil, errRollback
				}
				return nil, errCreate
			}
		}
		// If the queue already exists, we can ignore the error
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	q := &Queue{
		ctx: ctx,
		workflowsWorker: execution.NewWorkerPool(
			ctx,
			queue,
			execution.NewWorkflowsExecutor(ctx, queue),
			func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]] {
				return execution.NewWorkflowsWorker(ctx, queue, registry)
			},
		),
		activitiesWorker: execution.NewWorkerPool(
			ctx,
			queue,
			execution.NewActivitiesExecutor(ctx, queue),
			func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.ActivityRequest, execution.ActivityReponse]] {
				return execution.NewActivitiesWorker(ctx, queue, registry)
			},
		),
		sideEffectsWorker: execution.NewWorkerPool(
			ctx,
			queue,
			execution.NewSideEffectsExecutor(ctx, queue),
			func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.SideEffectRequest, execution.SideEffectReponse]] {
				return execution.NewSideEffectsWorker(ctx, queue, registry)
			},
		),
		sagasWorker: execution.NewWorkerPool(
			ctx,
			queue,
			execution.NewSagasExecutor(ctx, queue),
			func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.SagaRequest, execution.SagaReponse]] {
				return execution.NewSagasWorker(ctx, queue, registry)
			},
		),
	}

	// TODO: change with option of initial workers configs
	q.AddWorker()

	return q, nil
}

func (q *Queue) SubmitWorkflow(request *repository.WorkflowInfo) error {
	return q.workflowsWorker.Submit(
		&retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]{Request: execution.WorkflowRequest{}},
	)
}

func (q *Queue) SubmitActivity(request *repository.ActivityInfo) error {
	return q.activitiesWorker.Submit(
		&retrypool.RequestResponse[execution.ActivityRequest, execution.ActivityReponse]{Request: execution.ActivityRequest{}},
	)
}

func (q *Queue) SubmitSideEffect(request *repository.SideEffectInfo) error {
	return q.sideEffectsWorker.Submit(
		&retrypool.RequestResponse[execution.SideEffectRequest, execution.SideEffectReponse]{Request: execution.SideEffectRequest{}},
	)
}

func (q *Queue) SubmitSaga(request *repository.SagaInfo) error {
	return q.sagasWorker.Submit(
		&retrypool.RequestResponse[execution.SagaRequest, execution.SagaReponse]{Request: execution.SagaRequest{}},
	)
}

func (q *Queue) AvailableWorkflowWorkers() int {
	return q.workflowsWorker.AvailableWorkers()
}

func (q *Queue) AvailableActivityWorkers() int {
	return q.activitiesWorker.AvailableWorkers()
}

func (q *Queue) AvailableSideEffectWorkers() int {
	return q.sideEffectsWorker.AvailableWorkers()
}

func (q *Queue) AvailableSagaWorkers() int {
	return q.sagasWorker.AvailableWorkers()
}

func (q *Queue) Wait() error {
	waitErrGroup := errgroup.Group{}

	waitErrGroup.Go(func() error {
		return q.workflowsWorker.Wait()
	})

	waitErrGroup.Go(func() error {
		return q.activitiesWorker.Wait()
	})

	waitErrGroup.Go(func() error {
		return q.sideEffectsWorker.Wait()
	})

	waitErrGroup.Go(func() error {
		return q.sagasWorker.Wait()
	})

	return waitErrGroup.Wait()
}

func (q *Queue) Shutdown() error {

	shutdownErrGroup := errgroup.Group{}

	shutdownErrGroup.Go(func() error {
		fmt.Println("\t Shutting down workflows worker")
		return q.workflowsWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		fmt.Println("\t Shutting down activities worker")
		return q.activitiesWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		fmt.Println("\t Shutting down side effects worker")
		return q.sideEffectsWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		fmt.Println("\t Shutting down sagas worker")
		return q.sagasWorker.Shutdown()
	})

	defer fmt.Println("Queue shutdown complete")
	return shutdownErrGroup.Wait()
}

func (q *Queue) AddWorker() {
	q.workflowsWorker.AddWorker()
	q.activitiesWorker.AddWorker()
	q.sideEffectsWorker.AddWorker()
	q.sagasWorker.AddWorker()
}

func (q *Queue) RemoveWorker() error {
	errG := errgroup.Group{}

	errG.Go(func() error {
		return q.workflowsWorker.RemoveWorker()
	})

	errG.Go(func() error {
		return q.activitiesWorker.RemoveWorker()
	})

	errG.Go(func() error {
		return q.sideEffectsWorker.RemoveWorker()
	})

	errG.Go(func() error {
		return q.sagasWorker.RemoveWorker()
	})

	return errG.Wait()
}
