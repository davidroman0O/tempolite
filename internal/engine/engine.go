package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/engine/schedulers"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
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
	info         *info.InfoClock
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

	e.info = info.New(ctx)

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
	if defaultQueue, err = queues.New(e.ctx, queue, e.registry, e.db); err != nil {
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
		e.info.Stop()
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
	if id, err = e.commandWorkflow(workflowFunc, options, params...); err != nil {
		return e.queryNoWorkflow(err)
	}
	return e.queryWorfklow(workflowFunc, id)
}

func (e *Engine) commandWorkflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		return types.NoWorkflowID, err
	}

	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		return types.NoWorkflowID, err
	}

	if err = e.registry.VerifyParamsMatching(types.HandlerInfo(workflow), params...); err != nil {
		return types.NoWorkflowID, err
	}

	var tx *ent.Tx
	if tx, err = e.db.Tx(); err != nil {
		return types.NoWorkflowID, err
	}

	var workflowInfo *repository.WorkflowInfo

	var runInfo *repository.RunInfo
	if runInfo, err = e.db.Runs().Create(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	retryPolicyConfig := schema.RetryPolicy{
		MaxAttempts: 1,
	}

	queueName := "default"
	var queueInfo *repository.QueueInfo

	if queueInfo, err = e.db.Queues().GetByName(tx, queueName); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	duration := ""

	if options != nil {
		config := types.WorkflowConfig{}
		for _, opt := range options {
			opt(&config)
		}
		if config.QueueName != "" {
			queueName = config.QueueName
		}
		if config.RetryMaximumAttempts >= 0 {
			retryPolicyConfig.MaxAttempts = config.RetryMaximumAttempts
		}
		if config.RetryInitialInterval >= 0 {
			retryPolicyConfig.InitialInterval = config.RetryInitialInterval
		}
		// TODO: implement layet
		// if config.RetryBackoffCoefficient >= 0 {
		// 	retryPolicyConfig.BackoffCoefficient = config.RetryBackoffCoefficient
		// }
		// if config.MaximumInterval >= 0 {
		// 	retryPolicyConfig.MaxAttempts = config.MaximumInterval
		// }
		if config.Duration != "" {
			duration = config.Duration
		}

	}

	// Extremely important conversion
	serializableParams, err := io.ConvertInputsForSerialization(params)
	if err != nil {
		return types.NoWorkflowID, err
	}

	if workflowInfo, err = e.db.
		Workflows().
		Create(
			tx,
			repository.CreateWorkflowInput{
				RunID:       runInfo.ID,
				HandlerName: identity.String(),
				StepID:      "root",
				RetryPolicy: &retryPolicyConfig,
				Input:       serializableParams,
				QueueID:     queueInfo.ID,
				Duration:    duration, // if empty it won't be set
			}); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	if err = tx.Commit(); err != nil {
		return types.NoWorkflowID, err
	}

	return types.WorkflowID(workflowInfo.ID), nil
}

func (e *Engine) queryWorfklow(workflowFunc interface{}, id types.WorkflowID) *info.WorkflowInfo {
	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		return e.queryNoWorkflow(err)
	}
	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		return e.queryNoWorkflow(err)
	}
	return info.NewWorkflowInfo(id, types.HandlerInfo(workflow), e.db, e.info)
}

func (e *Engine) queryNoWorkflow(err error) *info.WorkflowInfo {
	return info.NewWorkflowInfoWithError(err)
}
