package engine

import (
	"context"
	"fmt"
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
	e := &Engine{
		ctx:          ctx,
		workerQueues: make(map[string]*queues.Queue),
	}

	if e.registry, err = builder(); err != nil {
		return nil, err
	}

	e.db = repository.NewRepository(
		ctx,
		client)

	if e.scheduler, err = schedulers.New(
		ctx,
		e.db,
		e.registry,
		e.GetQueue,
	); err != nil {
		return nil, err
	}

	e.commands = commands.New(e.ctx, e.db, e.registry)
	e.queries = queries.New(e.ctx, e.db, e.registry)

	if err := e.AddQueue("default"); err != nil {
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
	// TODO: add options for initial workers
	if defaultQueue, err = queues.New(
		e.ctx,
		queue,
		e.registry,
		e.db,
		e.commands,
		e.queries,
	); err != nil {
		return err
	}
	e.scheduler.AddQueue(queue)
	e.mu.Lock()
	e.workerQueues[queue] = defaultQueue
	e.mu.Unlock()
	return nil
}

func (e *Engine) Shutdown() error {
	shutdown := errgroup.Group{}

	shutdown.Go(func() error {
		fmt.Println("Shutting down scheduler")
		e.scheduler.Stop()
		return nil
	})

	shutdown.Go(func() error {
		fmt.Println("Shutting down info")
		e.queries.Stop()
		return nil
	})

	for n, q := range e.workerQueues {
		shutdown.Go(func() error {
			fmt.Println("Shutting down queue", n)
			return q.Shutdown()
		})
	}

	defer fmt.Println("Engine shutdown complete")
	return shutdown.Wait()
}

func (e *Engine) Scale(queue string, targets map[string]int) error {
	e.mu.Lock()
	q := e.workerQueues[queue]
	e.mu.Unlock()
	return q.Scale(targets)
}

// Create a new run workflow
func (e *Engine) Workflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var id types.WorkflowID
	var err error
	if id, err = e.commands.CommandWorkflow(workflowFunc, options, params...); err != nil {
		return e.queries.QueryNoWorkflow(err)
	}
	return e.queries.QueryWorfklow(workflowFunc, id)
}
