package queues

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/cq/commands"
	"github.com/davidroman0O/tempolite/internal/engine/cq/queries"
	"github.com/davidroman0O/tempolite/internal/engine/execution"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/pkg/logs"
	"golang.org/x/sync/errgroup"
)

type Queue struct {
	ctx               context.Context
	workflowsWorker   *execution.WorkerPool[execution.WorkflowRequest, execution.WorkflowReponse]
	activitiesWorker  *execution.WorkerPool[execution.ActivityRequest, execution.ActivityReponse]
	sideEffectsWorker *execution.WorkerPool[execution.SideEffectRequest, execution.SideEffectReponse]
	sagasWorker       *execution.WorkerPool[execution.SagaRequest, execution.SagaReponse]
	scaleMu           sync.Mutex

	commands  *commands.Commands
	queries   *queries.Queries
	queueName string
}

func New(
	ctx context.Context,
	queue string,
	registry *registry.Registry,
	db repository.Repository,
	commands *commands.Commands,
	queries *queries.Queries,
) (*Queue, error) {

	logs.Debug(ctx, "Creating queue", "queue", queue)
	tx, err := db.Tx()
	if err != nil {
		logs.Error(ctx, "New Queue error creating transaction", "error", err, "queue", queue)
		return nil, err
	}

	if _, err = db.Queues().GetByName(tx, "default"); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			fmt.Println("Creating default queue")
			if _, errCreate := db.Queues().Create(tx, "default"); errCreate != nil {
				if errRollback := tx.Rollback(); errRollback != nil {
					logs.Error(ctx, "New Queue error rolling back transaction", "error", errRollback, "queue", queue)
					return nil, errRollback
				}
				logs.Error(ctx, "New Queue error creating default queue", "error", errCreate, "queue", queue)
				return nil, errCreate
			}
		}
		// If the queue already exists, we can ignore the error
	}

	if err := tx.Commit(); err != nil {
		logs.Error(ctx, "New Queue error committing transaction", "error", err, "queue", queue)
		return nil, err
	}

	logs.Debug(ctx, "New Queue creating queue", "queue", queue)
	q := &Queue{
		ctx:       ctx,
		commands:  commands,
		queries:   queries,
		queueName: queue,
	}

	logs.Debug(ctx, "New Queue creating workflow worker", "queue", queue)
	q.workflowsWorker = execution.NewWorkerPool(
		ctx,
		queue,
		execution.NewWorkflowsExecutor(ctx, queue, db),
		func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]] {
			return execution.NewWorkflowsWorker(
				ctx,
				queue,
				registry,
				db,
				commands,
				queries,
			)
		},
	)

	logs.Debug(ctx, "New Queue creating activities worker", "queue", queue)
	q.activitiesWorker = execution.NewWorkerPool(
		ctx,
		queue,
		execution.NewActivitiesExecutor(ctx, queue),
		func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.ActivityRequest, execution.ActivityReponse]] {
			return execution.NewActivitiesWorker(ctx, queue, registry)
		},
	)

	logs.Debug(ctx, "New Queue creating side effects worker", "queue", queue)
	q.sideEffectsWorker = execution.NewWorkerPool(
		ctx,
		queue,
		execution.NewSideEffectsExecutor(ctx, queue),
		func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.SideEffectRequest, execution.SideEffectReponse]] {
			return execution.NewSideEffectsWorker(ctx, queue, registry)
		},
	)

	logs.Debug(ctx, "New Queue creating sagas worker", "queue", queue)
	q.sagasWorker = execution.NewWorkerPool(
		ctx,
		queue,
		execution.NewSagasExecutor(ctx, queue),
		func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[execution.SagaRequest, execution.SagaReponse]] {
			return execution.NewSagasWorker(ctx, queue, registry)
		},
	)

	// TODO: change with option of initial workers configs
	q.AddWorker()

	return q, nil
}

func (q *Queue) submitRetryWorkflow(data *retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]) error {
	logs.Debug(q.ctx, "Queue submit retry workflow", "queue", q.queueName)
	if err := q.workflowsWorker.Submit(data); err != nil {
		logs.Error(q.ctx, "Queue submit retry workflow error", "error", err, "queue", q.queueName)
		return err
	}
	return nil
}

func (q *Queue) SubmitWorkflow(task *retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]) (chan struct{}, error) {

	processed := retrypool.NewProcessedNotification()

	logs.Debug(q.ctx, "Queue submit workflow", "queue", q.queueName)
	if err := q.workflowsWorker.Submit(
		task,
		retrypool.WithBeingProcessed[*retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]](processed),
	); err != nil {
		logs.Error(q.ctx, "Queue submit workflow error", "error", err, "queue", q.queueName)
		return nil, err
	}

	return processed, nil
}

func (q *Queue) SubmitActivity(request *repository.ActivityInfo) error {
	logs.Debug(q.ctx, "Queue submit activity", "queue", q.queueName)
	return q.activitiesWorker.Submit(
		&retrypool.RequestResponse[execution.ActivityRequest, execution.ActivityReponse]{Request: execution.ActivityRequest{}},
	)
}

func (q *Queue) SubmitSideEffect(request *repository.SideEffectInfo) error {
	logs.Debug(q.ctx, "Queue submit side effect", "queue", q.queueName)
	return q.sideEffectsWorker.Submit(
		&retrypool.RequestResponse[execution.SideEffectRequest, execution.SideEffectReponse]{Request: execution.SideEffectRequest{}},
	)
}

func (q *Queue) SubmitSaga(request *repository.SagaInfo) error {
	logs.Debug(q.ctx, "Queue submit saga", "queue", q.queueName)
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
		logs.Debug(q.ctx, "Shutting down queue workflow worker", "queue", q.queueName)
		return q.workflowsWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		logs.Debug(q.ctx, "Shutting down queue activities worker", "queue", q.queueName)
		return q.activitiesWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		logs.Debug(q.ctx, "Shutting down queue side effects worker", "queue", q.queueName)
		return q.sideEffectsWorker.Shutdown()
	})

	shutdownErrGroup.Go(func() error {
		logs.Debug(q.ctx, "Shutting down queue sagas worker", "queue", q.queueName)
		return q.sagasWorker.Shutdown()
	})

	defer logs.Debug(q.ctx, "Queue shutdown complete")
	return shutdownErrGroup.Wait()
}

func (q *Queue) AddWorker() {
	logs.Debug(q.ctx, "Adding workers to queue", "queue", q.queueName)
	q.workflowsWorker.AddWorker()
	q.activitiesWorker.AddWorker()
	q.sideEffectsWorker.AddWorker()
	q.sagasWorker.AddWorker()
}

func (q *Queue) RemoveWorker() error {
	errG := errgroup.Group{}
	logs.Debug(q.ctx, "Removing workers from queue", "queue", q.queueName)

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

func (q *Queue) Scale(targets map[string]int) error {
	q.scaleMu.Lock()
	defer q.scaleMu.Unlock()

	type pool struct {
		name      string
		available func() int
		add       func() int
		remove    func() error
	}

	pools := []pool{
		{
			name:      "workflows",
			available: q.workflowsWorker.AvailableWorkers,
			add:       q.workflowsWorker.AddWorker,
			remove:    q.workflowsWorker.RemoveWorker,
		},
		{
			name:      "activities",
			available: q.activitiesWorker.AvailableWorkers,
			add:       q.activitiesWorker.AddWorker,
			remove:    q.activitiesWorker.RemoveWorker,
		},
		{
			name:      "sideEffects",
			available: q.sideEffectsWorker.AvailableWorkers,
			add:       q.sideEffectsWorker.AddWorker,
			remove:    q.sideEffectsWorker.RemoveWorker,
		},
		{
			name:      "sagas",
			available: q.sagasWorker.AvailableWorkers,
			add:       q.sagasWorker.AddWorker,
			remove:    q.sagasWorker.RemoveWorker,
		},
	}

	var wg sync.WaitGroup

	for _, p := range pools {
		target, ok := targets[p.name]
		if !ok {
			continue
		}

		p := p // capture loop variable
		wg.Add(1)

		go func(p pool, target int) {
			defer wg.Done()
			for {
				current := p.available()
				// fmt.Println("Current workers for", p.name, ":", current)
				if current == target {
					// fmt.Println("Already at target workers for", p.name)
					break
				} else if current < target {
					// fmt.Println("Adding worker to", p.name, "pool")
					p.add()
				} else if current > target {
					// fmt.Println("Removing worker from", p.name, "pool")
					err := p.remove()
					if err != nil {
						if errors.Is(err, retrypool.ErrAlreadyRemovingWorker) {
							// Wait for the worker count to decrease
							for {
								time.Sleep(100 * time.Millisecond)
								newCurrent := p.available()
								if newCurrent < current {
									break
								}
							}
						} else if err.Error() == "no workers to remove" {
							// fmt.Println("No workers to remove from", p.name)
							break
						} else {
							logs.Error(q.ctx, "Error removing worker from pool", "error", err, "pool", p.name)
							break
						}
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}(p, target)
	}

	wg.Wait()

	logs.Debug(q.ctx, "Queue scaled", "queue", q.queueName)
	logs.Debug(q.ctx, "Available workflows workers:", q.AvailableWorkflowWorkers())
	logs.Debug(q.ctx, "Available activities workers:", q.AvailableActivityWorkers())
	logs.Debug(q.ctx, "Available side effects workers:", q.AvailableSideEffectWorkers())
	logs.Debug(q.ctx, "Available sagas workers:", q.AvailableSagaWorkers())

	return nil
}
