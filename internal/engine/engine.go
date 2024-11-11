package engine

import (
	"context"
	"sync"

	"github.com/davidroman0O/tempolite/internal/engine/cq/commands"
	"github.com/davidroman0O/tempolite/internal/engine/cq/queries"
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/engine/schedulers"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	ctx          context.Context
	mu           sync.Mutex
	registry     *registry.Registry
	workerQueues map[string]*queues.Queue
	db           repository.Repository
	scheduler    *schedulers.Scheduler
	commands     *commands.Commands
	queries      *queries.Queries
}

func New(
	ctx context.Context,
	builder registry.RegistryBuildFn,
	client *ent.Client,
) (*Engine, error) {
	var err error
	logs.Debug(ctx, "Creating engine")
	e := &Engine{
		ctx:          ctx,
		workerQueues: make(map[string]*queues.Queue),
	}

	logs.Debug(ctx, "Creating registry")
	if e.registry, err = builder(); err != nil {
		logs.Error(ctx, "Error creating registry", "error", err)
		return nil, err
	}

	logs.Debug(ctx, "Creating repository")
	e.db = repository.NewRepository(
		ctx,
		client)

	logs.Debug(ctx, "Creating scheduler")
	if e.scheduler, err = schedulers.New(
		ctx,
		e.db,
		e.registry,
		e.GetQueue,
	); err != nil {
		logs.Error(ctx, "Error creating scheduler", "error", err)
		return nil, err
	}

	e.commands = commands.New(e.ctx, e.db, e.registry)
	e.queries = queries.New(e.ctx, e.db, e.registry)

	if err := e.AddQueue("default"); err != nil {
		logs.Error(ctx, "Error creating default queue", "error", err)
		return nil, err
	}

	return e, nil
}

func (e *Engine) GetQueue(queue string) *queues.Queue {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.workerQueues[queue]
}

func (e *Engine) AddQueue(queue string) error {
	var err error
	var defaultQueue *queues.Queue
	logs.Debug(e.ctx, "Adding queue", "queue", queue)
	// TODO: add options for initial workers
	if defaultQueue, err = queues.New(
		e.ctx,
		queue,
		e.registry,
		e.db,
		e.commands,
		e.queries,
	); err != nil {
		logs.Error(e.ctx, "Error creating queue", "queue", queue, "error", err)
		return err
	}
	e.scheduler.AddQueue(queue)
	e.mu.Lock()
	e.workerQueues[queue] = defaultQueue
	e.mu.Unlock()
	logs.Debug(e.ctx, "Queue added", "queue", queue)
	return nil
}

func (e *Engine) Shutdown() error {
	shutdown := errgroup.Group{}
	logs.Debug(e.ctx, "Shutting down engine")

	logs.Debug(e.ctx, "Shutting down scheduler")
	e.scheduler.Stop()
	logs.Debug(e.ctx, "Scheduler shutdown complete")

	shutdown.Go(func() error {
		logs.Debug(e.ctx, "Shutting down queries")
		e.queries.Stop()
		logs.Debug(e.ctx, "Queries shutdown complete")
		return nil
	})

	for n, q := range e.workerQueues {
		shutdown.Go(func() error {
			logs.Debug(e.ctx, "Shutting down queue", "queue", n)
			defer logs.Debug(e.ctx, "Queue shutdown complete", "queue", n)
			return q.Shutdown()
		})
	}

	defer logs.Debug(e.ctx, "Engine shutdown complete")

	logs.Debug(e.ctx, "Waiting for shutdown")
	return shutdown.Wait()
}

func (e *Engine) Scale(queue string, targets map[string]int) error {
	logs.Debug(e.ctx, "Preparing to scale", "queue", queue, "targets", targets)
	e.mu.Lock()
	q := e.workerQueues[queue]
	e.mu.Unlock()
	logs.Debug(e.ctx, "Scaling queue", "queue", queue, "targets", targets)
	return q.Scale(targets)
}

// Create a new run workflow
//
// # Pseudo code
//
// ```
// => Workflow
// -     - create workflow entity
// -     - create workflow execution
// -
// -     if fail
// -     if max attempt < attemps => retry
// -         - create workflow execution
// -
// - 		if fail
// - 		if max attempt < attemps => retry
// - 			- create workflow execution
// - 		else
// - 			- workflow failed
// - else
// - 	- workflow failed
// ```
func (e *Engine) Workflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var id types.WorkflowID
	var err error
	logs.Debug(e.ctx, "Creating workflow")
	if id, err = e.commands.CommandWorkflow(workflowFunc, options, params...); err != nil {
		logs.Error(e.ctx, "Error creating workflow", "error", err)
		return e.queries.QueryNoWorkflow(err)
	}
	logs.Debug(e.ctx, "Workflow created", "id", id)
	return e.queries.QueryWorfklow(workflowFunc, id)
}

// IF you cancelled a Tempolite instance, you need to a new WorkflowInfo instance
func (e *Engine) GetWorkflow(id types.WorkflowID) *info.WorkflowInfo {
	logs.Debug(e.ctx, "Engine Getting workflow", id)
	return e.queries.GetWorkflowInfo(id)
}
