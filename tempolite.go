package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"
)

/// At the Tempolite level, if we want to benefit from pause/resume and other features, we need to pre-register all workflow functions, only the root workflows. Which is one constraint due to the runtime/registry/database isolation. We should be able to resume a workflow at any time, even publish a signal on a paused workflow, then later on resume it, and it should continue on the next operation.

// WorkflowRequest represents a workflow execution request
type WorkflowRequest struct {
	workflowID   WorkflowEntityID
	workflowFunc interface{}      // The workflow function
	options      *WorkflowOptions // Workflow options
	args         []interface{}    // Arguments for the workflow
	queueName    string           // Name of the queue this task belongs to
	continued    bool
	resume       bool
	enqueue      bool
	chnFuture    chan Future
}

type WorkflowResponse struct {
	ID WorkflowEntityID // ID of the workflow entity to execute
}

// QueueInstance manages a queue and its worker pool
type QueueInstance struct {
	mu deadlock.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	registry *Registry
	database Database

	name         string
	maxRuns      int32
	maxWorkflows int32

	// We use a blocking pool because the orchestrator will lock workers of the pool, we will have one Run as group per pool
	orchestrators *retrypool.BlockingPool[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], RunID, WorkflowEntityID]

	workerMu          deadlock.RWMutex
	processingWorkers map[WorkflowEntityID]*QueueWorker
	freeWorkers       map[int]struct{}
	workers           map[int]*QueueWorker

	onCrossWorkflow crossQueueWorkflowHandler
	onContinueAsNew crossQueueContinueAsNewHandler
	onSignalNew     workflowNewSignalHandler
	onSignalRemove  workflowRemoveSignalHandler
}

type queueConfig struct {
	onCrossWorkflow crossQueueWorkflowHandler
	onContinueAsNew crossQueueContinueAsNewHandler
	onSignalNew     workflowNewSignalHandler
	onSignalRemove  workflowRemoveSignalHandler
}

type queueOption func(*queueConfig)

func WithCrossWorkflowHandler(handler crossQueueWorkflowHandler) queueOption {
	return func(c *queueConfig) {
		c.onCrossWorkflow = handler
	}
}

func WithContinueAsNewHandler(handler crossQueueContinueAsNewHandler) queueOption {
	return func(c *queueConfig) {
		c.onContinueAsNew = handler
	}
}

func WithSignalNewHandler(handler workflowNewSignalHandler) queueOption {
	return func(c *queueConfig) {
		c.onSignalNew = handler
	}
}

func WithSignalRemoveHandler(handler workflowRemoveSignalHandler) queueOption {
	return func(c *queueConfig) {
		c.onSignalRemove = handler
	}
}

func (q *QueueInstance) metrics() retrypool.BlockingMetricsSnapshot[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], RunID] {
	return q.orchestrators.GetMetricsSnapshot()
}

func (q *QueueInstance) Pause(id WorkflowEntityID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	worker, exists := q.processingWorkers[id]
	if !exists {

		// // TODO: We have a problem, when we want to pause a newly created workflow, its not yet running so there might not be an execution for it
		// // TODO: the  "WorkflowExecutions" got "Running" and since it is not running yet it shouldn't be there, the main workflow is already locking the worker

		// _, err := q.database.GetWorkflowEntity(id)
		// if err != nil {
		// 	err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to get workflow entity: %w", err))
		// 	logger.Error(q.ctx, err.Error(), "workflow_id", id)
		// 	return err
		// }

		// snap := q.orchestrators.GetMetricsSnapshot()
		// fmt.Println("snap", snap.Queues)

		// if err := q.database.
		// 	SetWorkflowEntityProperties(
		// 		id,
		// 		SetWorkflowEntityStatus(StatusPaused),
		// 	); err != nil {
		// 	err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to set workflow entity status: %w", err))
		// 	logger.Error(q.ctx, err.Error(), "workflow_id", id)
		// 	return err
		// }

		// exec, err := q.database.GetWorkflowExecutionLatestByEntityID(id)
		// if err != nil {
		// 	err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to get workflow execution: %w", err))
		// 	logger.Error(q.ctx, err.Error(), "workflow_id", id)
		// 	return err
		// }

		// if err := q.database.
		// 	SetWorkflowExecutionProperties(
		// 		exec.ID,
		// 		SetWorkflowExecutionStatus(ExecutionStatusPaused),
		// 	); err != nil {
		// 	err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to set workflow execution status: %w", err))
		// 	logger.Error(q.ctx, err.Error(), "workflow_id", id, "execution_id", exec.ID)
		// 	return err
		// }

		return fmt.Errorf("workflow %d is not being processed", id)
	}
	worker.Pause()
	return nil
}

func NewQueueInstance(ctx context.Context, db Database, registry *Registry, name string, maxRuns int, maxWorkflows int, opt ...queueOption) (*QueueInstance, error) {
	q := &QueueInstance{
		name:              name,
		maxRuns:           int32(maxRuns),
		maxWorkflows:      int32(maxWorkflows),
		registry:          registry,
		database:          db,
		ctx:               ctx,
		processingWorkers: make(map[WorkflowEntityID]*QueueWorker),
		freeWorkers:       make(map[int]struct{}),
		workers:           make(map[int]*QueueWorker),
	}

	cfg := &queueConfig{}

	for _, o := range opt {
		o(cfg)
	}

	if cfg.onCrossWorkflow != nil {
		q.onCrossWorkflow = cfg.onCrossWorkflow
	}
	if cfg.onContinueAsNew != nil {
		q.onContinueAsNew = cfg.onContinueAsNew
	}
	if cfg.onSignalNew != nil {
		q.onSignalNew = cfg.onSignalNew
	}
	if cfg.onSignalRemove != nil {
		q.onSignalRemove = cfg.onSignalRemove
	}

	_, err := db.AddQueue(&Queue{Name: name})
	if err != nil {
		if !errors.Is(err, ErrQueueExists) {
			return nil, err
		}
	}

	workers := []retrypool.Worker[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]]{}
	for i := 0; i < maxRuns; i++ {
		workers = append(workers, NewQueueWorker(q))
	}

	logger := retrypool.NewLogger(slog.LevelDebug)
	logger.Enable()

	options := []retrypool.BlockingPoolOption[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]]{

		retrypool.WithBlockingWorkerFactory(func() retrypool.Worker[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]] {
			return NewQueueWorker(q)
		}),
		// TODO: make a OnNewSession or OnRemoveSession to increase or decrease that number
		retrypool.WithBlockingMaxActivePools[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](maxRuns),
	}

	if maxWorkflows > 0 {
		options = append(options, retrypool.WithBlockingMaxWorkersPerPool[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](maxWorkflows))
	}

	q.orchestrators, err = retrypool.NewBlockingPool[
		*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], RunID, WorkflowEntityID](
		ctx,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// q.orchestrators = retrypool.New(
	// 	ctx,
	// 	workers,
	// 	retrypool.WithLogger[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](
	// 		logger,
	// 	),
	// 	retrypool.WithAttempts[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](1),
	// 	retrypool.WithOnPanic[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](
	// 		func(recovery interface{}, stackTrace string) {
	// 			fmt.Println("PANIC", recovery, stackTrace)
	// 		},
	// 	),
	// 	retrypool.WithRoundRobinDistribution[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](),
	// 	retrypool.WithOnWorkerPanic[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](
	// 		func(workerID int, recovery interface{}, stackTrace string) {
	// 			fmt.Println("WORKER PANIC", workerID, recovery, stackTrace)
	// 		},
	// 	),
	// 	// retrypool.WithDelay[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](time.Second/2),
	// 	retrypool.WithOnDeadTask[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](
	// 		func(deadTaskIndex int) {
	// 			fmt.Println("DEADTASK", deadTaskIndex)
	// 			// deadTask, err := q.orchestrators.GetDeadTask(deadTaskIndex)
	// 			// if err != nil {
	// 			// 	// too bad
	// 			// 	log.Printf("failed to get dead task: %v", err)
	// 			// 	return
	// 			// }
	// 			// errs := errors.New("failed to process request")
	// 			// for _, e := range deadTask.Errors {
	// 			// 	errs = errors.Join(errs, e)
	// 			// }
	// 			// deadTask.Data.CompleteWithError(errs)
	// 			// _, err = q.orchestrators.PullDeadTask(deadTaskIndex)
	// 			// if err != nil {
	// 			// 	// too bad
	// 			// 	log.Printf("failed to pull dead task: %v", err)
	// 			// }
	// 		},
	// 	),
	// 	// retrypool.WithOnNewDeadTask[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](
	// 	// 	func(task *retrypool.DeadTask[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]], idx int) {
	// 	// 		errs := errors.New("failed to process request")
	// 	// 		for _, e := range task.Errors {
	// 	// 			errs = errors.Join(errs, e)
	// 	// 		}
	// 	// 		task.Data.CompleteWithError(errs)
	// 	// 		_, err := q.orchestrators.PullDeadTask(idx)
	// 	// 		if err != nil {
	// 	// 			// too bad
	// 	// 			log.Printf("failed to pull dead task: %v", err)
	// 	// 		}
	// 	// 	}),
	// )

	return q, nil
}

func (qi *QueueInstance) Close() error {
	if err := qi.orchestrators.Close(); err != nil {
		if err != context.Canceled {
			return fmt.Errorf("failed to close orchestrators: %w", err)
		}
	}
	return nil
}

func (qi *QueueInstance) Wait() error {
	return qi.orchestrators.WaitWithCallback(qi.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		logger.Debug(qi.ctx, "waiting for queue to finish", "queue", qi.name, "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
		return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
	}, time.Second)
}

func (qi *QueueInstance) Submit(workflowFunc interface{}, options *WorkflowOptions, opts *preparationOptions, args ...interface{}) (Future, *retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], *retrypool.QueuedNotification, error) {

	workflowEntity, err := PrepareWorkflow(qi.registry, qi.database, workflowFunc, options, opts, args...)
	if err != nil {
		log.Printf("failed to prepare workflow: %v", err)
		return nil, nil, nil, err
	}

	chnFuture := make(chan Future, 1)

	// queuenotification := retrypool.NewQueuedNotification()

	task := retrypool.NewBlockingRequestResponse[*WorkflowRequest, *WorkflowResponse](
		&WorkflowRequest{
			workflowID:   workflowEntity.ID,
			workflowFunc: workflowFunc,
			options:      options,
			args:         args,
			chnFuture:    chnFuture,
			queueName:    qi.name,
		},
		workflowEntity.RunID,
		workflowEntity.ID,
	)

	if err := qi.orchestrators.
		Submit(
			task,
			// retrypool.WithBlockingQueueNotification[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](queuenotification),
		); err != nil {
		fmt.Println("failed to submit task", err)
		return nil, nil, nil, err
	}

	return <-chnFuture, task, nil, nil
}

func (qi *QueueInstance) SubmitResume(entityID WorkflowEntityID) (Future, *retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], *retrypool.QueuedNotification, error) {
	// Get workflow entity with queue info
	workflowEntity, err := qi.database.GetWorkflowEntity(entityID, WorkflowEntityWithQueue())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	// Verify the workflow is actually paused
	var status EntityStatus
	if err := qi.database.GetWorkflowEntityProperties(entityID, GetWorkflowEntityStatus(&status)); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get workflow status: %w", err)
	}
	if status != StatusPaused {
		return nil, nil, nil, fmt.Errorf("workflow %d is not paused (current status: %s)", entityID, status)
	}

	// Get handler info
	handler, ok := qi.registry.GetWorkflow(workflowEntity.HandlerName)
	if !ok {
		return nil, nil, nil, fmt.Errorf("workflow handler %s not found", workflowEntity.HandlerName)
	}

	chnFuture := make(chan Future, 1)

	queuenotification := retrypool.NewQueuedNotification()

	// Create resume request
	task := retrypool.NewBlockingRequestResponse[*WorkflowRequest, *WorkflowResponse](
		&WorkflowRequest{
			workflowID:   workflowEntity.ID,
			workflowFunc: handler.Handler, // Use the actual handler function
			queueName:    qi.name,
			continued:    false, // Not a continuation
			resume:       true,  // Mark this as a resume
			chnFuture:    chnFuture,
		},
		workflowEntity.RunID,
		workflowEntity.ID,
	)

	if err := qi.orchestrators.
		Submit(
			task,
			retrypool.WithBlockingQueueNotification[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](queuenotification),
		); err != nil {
		fmt.Println("failed to submit task", err)
		return nil, nil, nil, fmt.Errorf("failed to submit resume task: %w", err)
	}

	return <-chnFuture, task, queuenotification, nil

}

// Defered Workflows aren't paused but in pending state, you have to manually enqueue them
func (qi *QueueInstance) SubmitEnqueue(entityID WorkflowEntityID) (Future, *retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], *retrypool.QueuedNotification, error) {
	// Get workflow entity with queue info
	workflowEntity, err := qi.database.GetWorkflowEntity(entityID, WorkflowEntityWithQueue())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	// Verify the workflow is actually paused
	var status EntityStatus
	if err := qi.database.GetWorkflowEntityProperties(entityID, GetWorkflowEntityStatus(&status)); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get workflow status: %w", err)
	}
	if status != StatusPending {
		return nil, nil, nil, fmt.Errorf("workflow %d is not pending (current status: %s)", entityID, status)
	}

	// Get handler info
	handler, ok := qi.registry.GetWorkflow(workflowEntity.HandlerName)
	if !ok {
		return nil, nil, nil, fmt.Errorf("workflow handler %s not found", workflowEntity.HandlerName)
	}

	chnFuture := make(chan Future, 1)

	queuenotification := retrypool.NewQueuedNotification()

	// Create resume request
	task := retrypool.NewBlockingRequestResponse[*WorkflowRequest, *WorkflowResponse](
		&WorkflowRequest{
			workflowID:   workflowEntity.ID,
			workflowFunc: handler.Handler, // Use the actual handler function
			queueName:    qi.name,
			continued:    false, // Not a continuation
			resume:       false, // Mark this as a resume
			enqueue:      true,  // Mark this as an enqueue
			chnFuture:    chnFuture,
		},
		workflowEntity.RunID,
		workflowEntity.ID,
	)

	if err := qi.orchestrators.
		Submit(
			task,
			retrypool.WithBlockingQueueNotification[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]](queuenotification),
		); err != nil {
		fmt.Println("failed to submit task", err)
		return nil, nil, nil, fmt.Errorf("failed to submit resume task: %w", err)
	}
	// pp.Println(qi.orchestrators.GetMetricsSnapshot())

	return <-chnFuture, task, queuenotification, nil
}

type onStartFunc func(context.Context)
type onEndFunc func(context.Context)

// QueueWorker implements the retrypool.Worker interface
type QueueWorker struct {
	ID            int
	queueInstance *QueueInstance
	orchestrator  *Orchestrator

	onStart onStartFunc
	onEnd   onEndFunc
}

func (w *QueueWorker) Pause() {
	w.orchestrator.Pause()
}

func NewQueueWorker(instance *QueueInstance) *QueueWorker {
	qw := &QueueWorker{
		queueInstance: instance,
	}

	// Create orchestrator first
	opts := []OrchestratorOption{}
	if instance.onCrossWorkflow != nil {
		opts = append(opts, WithCrossWorkflow(instance.onCrossWorkflow))
	}
	if instance.onContinueAsNew != nil {
		opts = append(opts, WithContinueAsNew(instance.onContinueAsNew))
	}
	if instance.onSignalNew != nil {
		opts = append(opts, WithSignalNew(instance.onSignalNew))
	}
	if instance.onSignalRemove != nil {
		opts = append(opts, WithSignalRemove(instance.onSignalRemove))
	}
	qw.orchestrator = NewOrchestrator(instance.ctx, instance.database, instance.registry, opts...)

	// Add worker to maps under a separate lock
	instance.workerMu.Lock()
	qw.ID = len(instance.workers) + 1
	instance.workers[qw.ID] = qw
	instance.workerMu.Unlock()

	qw.onStart = func(ctx context.Context) {}
	qw.onEnd = func(ctx context.Context) {}

	return qw
}

func (w *QueueWorker) OnStart(ctx context.Context) {
	logger.Info(ctx, "Queue worker started", "queue", w.queueInstance.name)
	// log.Printf("Queue worker started for queue %s", w.queueInstance.name)
}

func (w *QueueWorker) OnStop(ctx context.Context) {
	logger.Info(ctx, "Queue worker stopped", "queue", w.queueInstance.name)
	// log.Printf("Queue worker started for queue %s", w.queueInstance.name)
}

func (w *QueueWorker) OnRemove(ctx context.Context) {
	logger.Info(ctx, "Queue worker removed", "queue", w.queueInstance.name)
	// log.Printf("Queue worker started for queue %s", w.queueInstance.name)
}

func (w *QueueWorker) Run(ctx context.Context, task *retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID]) error {

	// Mark entity as processing
	w.queueInstance.workerMu.Lock()
	if _, exists := w.queueInstance.processingWorkers[task.Request.workflowID]; exists {
		w.queueInstance.workerMu.Unlock()
		future := NewRuntimeFuture()
		future.SetError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		task.Request.chnFuture <- future
		close(task.Request.chnFuture)
		task.CompleteWithError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		return nil
	}
	w.queueInstance.processingWorkers[task.Request.workflowID] = w
	w.queueInstance.workerMu.Unlock()

	// Clean up when done
	defer func() {
		w.queueInstance.workerMu.Lock()
		delete(w.queueInstance.processingWorkers, task.Request.workflowID)
		w.queueInstance.workerMu.Unlock()
	}()

	if _, err := w.orchestrator.registry.RegisterWorkflow(task.Request.workflowFunc); err != nil {
		future := NewRuntimeFuture()
		future.SetError(err)
		task.Request.chnFuture <- future
		close(task.Request.chnFuture)
		task.CompleteWithError(err)
		return fmt.Errorf("failed to register workflow: %w", err)
	}
	var future Future
	var err error

	// Execute the workflow using the orchestrator
	if task.Request.resume {
		// Use Resume for paused workflows
		future = w.orchestrator.Resume(task.Request.workflowID)
	} else {
		// Use ExecuteWithEntity for normal workflows
		future, err = w.orchestrator.ExecuteWithEntity(task.Request.workflowID)
	}

	task.Request.chnFuture <- future
	close(task.Request.chnFuture)

	if err != nil {
		future.SetError(err)
		task.CompleteWithError(err)
		// return fmt.Errorf("failed to execute workflow: %w", err)
		return nil
	}

	var results []interface{}
	// Get the results and update the DatabaseFuture
	if results, err = future.GetResults(); err != nil {
		future.SetError(err)
		task.CompleteWithError(err)
		//	it's fine, we have an successful error
		return nil
	}

	future.SetResult(results)
	task.Complete(&WorkflowResponse{
		ID: task.Request.workflowID,
	})

	return nil
}

// Tempolite is the main orchestration engine
type Tempolite struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu deadlock.RWMutex

	registry *Registry
	database Database

	queueInstances map[string]*QueueInstance

	defaultQueue string

	// TODO: when pause/cancel/resume check that we re-add/delete those futures
	signals map[string]Future
}

// QueueConfig holds configuration for a queue
type QueueConfig struct {
	Name         string
	MaxRuns      int
	MaxWorkflows int
}

// TempoliteOption is a function type for configuring Tempolite
type TempoliteOption func(*Tempolite) error

// createCrossWorkflowHandler creates the handler function for cross-queue workflow communication
func (t *Tempolite) createCrossWorkflowHandler() crossQueueWorkflowHandler {
	return func(queueName string, workflowID WorkflowEntityID, runID RunID, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
		t.mu.RLock()
		queue, ok := t.queueInstances[queueName]
		t.mu.RUnlock()
		if !ok {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(fmt.Errorf("queue %s not found", queueName))
			return futureErr
		}

		// since we work with a cross-workflow, we are resposible for some options
		// we simply get the information we need, set it to the future and return it
		if options.DeferExecution {
			future := NewRuntimeFuture()
			wrk, err := t.database.GetWorkflowEntity(workflowID, WorkflowEntityWithData())
			if err != nil {
				future.SetError(fmt.Errorf("failed to get workflow entity: %w", err))
				return future
			}
			future.setEntityID(FutureEntityWithWorkflowID(workflowID))
			future.setParentWorkflowID(*wrk.WorkflowData.WorkflowFrom)
			future.setParentWorkflowExecutionID(*wrk.WorkflowData.WorkflowExecutionFrom)
			return future
		}

		chnFuture := make(chan Future, 1)

		if err := queue.orchestrators.Submit(
			retrypool.NewBlockingRequestResponse[*WorkflowRequest, *WorkflowResponse](
				&WorkflowRequest{
					workflowFunc: workflowFunc,
					options:      options,
					workflowID:   workflowID,
					args:         args,
					chnFuture:    chnFuture,
					queueName:    queueName,
					continued:    false,
				},
				runID,
				workflowID,
			)); err != nil {
			fmt.Println("failed to submit task", err)
			futureErr := NewRuntimeFuture()
			futureErr.SetError(err)
			return futureErr
		}

		return <-chnFuture
	}
}

func (t *Tempolite) createContinueAsNewHandler() crossQueueContinueAsNewHandler {
	return func(queueName string, workflowID WorkflowEntityID, runID RunID, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
		t.mu.Lock()
		queue, ok := t.queueInstances[queueName]
		t.mu.Unlock()
		if !ok {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(fmt.Errorf("queue %s not found", queueName))
			return futureErr
		}

		chnFuture := make(chan Future, 1)

		// Handle version inheritance through ContinueAsNew
		if err := queue.orchestrators.Submit(
			retrypool.NewBlockingRequestResponse[*WorkflowRequest, *WorkflowResponse](
				&WorkflowRequest{
					workflowID:   workflowID,
					workflowFunc: workflowFunc,
					options:      options,
					args:         args,
					chnFuture:    chnFuture,
					queueName:    queueName,
					continued:    true, // Mark this as a continuation
				},
				runID,
				workflowID,
			)); err != nil {
			fmt.Println("failed to submit task", err)
			futureErr := NewRuntimeFuture()
			futureErr.SetError(err)
			return futureErr
		}

		return <-chnFuture
	}
}

func (t *Tempolite) createSignalNewHandler() workflowNewSignalHandler {
	return func(workflowID WorkflowEntityID, workflowExecutionID WorkflowExecutionID, signalEntityID SignalEntityID, signalExecutionID SignalExecutionID, signal string, future Future) error {

		t.mu.Lock()
		t.signals[signal] = future
		t.mu.Unlock()

		return nil
	}
}

func (t *Tempolite) createSignalRemoveHandler() workflowRemoveSignalHandler {
	return func(workflowID WorkflowEntityID, workflowExecutionID WorkflowExecutionID, signalEntityID SignalEntityID, signalExecutionID SignalExecutionID, signal string) error {
		t.mu.Lock()
		delete(t.signals, signal)
		t.mu.Unlock()
		return nil
	}
}

func (t *Tempolite) PublishSignal(workflowID WorkflowEntityID, signal string, value interface{}) error {
	t.mu.RLock()
	future, exists := t.signals[signal]
	t.mu.RUnlock()

	// First try runtime signals
	if exists {
		future.SetResult([]interface{}{value})
		return nil
	}

	// If no runtime signal found, check database for pending signals
	logger.Debug(t.ctx, "checking database for pending signal",
		"workflow_id", workflowID,
		"signal_name", signal)

	// Get hierarchies for this workflow+signal combination
	hierarchies, err := t.database.GetHierarchiesByParentEntityAndStep(int(workflowID), signal, EntitySignal)
	if err != nil {
		if errors.Is(err, ErrHierarchyNotFound) {
			// No signal exists for this workflow/signal combination
			logger.Debug(t.ctx, "no signal found",
				"workflow_id", workflowID,
				"signal_name", signal)
			return nil
		}
		return fmt.Errorf("failed to check for existing signals: %w", err)
	}

	if len(hierarchies) == 0 {
		// No hierarchies found
		// sorry it doesn't exists bro
		return fmt.Errorf("no hierarchies found for workflow %d and signal %s", workflowID, signal)
	}

	// Get the signal entity
	signalEntityID := SignalEntityID(hierarchies[0].ChildEntityID)

	// Check signal status
	var status EntityStatus
	if err := t.database.GetSignalEntityProperties(signalEntityID, GetSignalEntityStatus(&status)); err != nil {
		return fmt.Errorf("failed to get signal status: %w", err)
	}

	logger.Debug(t.ctx, "Found signal in database", "signal_entity_id", signalEntityID, "status", status)

	// Only proceed if signal is pending or running
	if status != StatusPending && status != StatusRunning {
		logger.Debug(t.ctx, "signal not in valid state for publishing",
			"workflow_id", workflowID,
			"signal_name", signal,
			"status", status)
		return fmt.Errorf("signal not in valid state for publishing: %s", status)
	}

	// Get latest execution
	latestExec, err := t.database.GetSignalExecutionLatestByEntityID(signalEntityID)
	if err != nil {
		return fmt.Errorf("failed to get latest signal execution: %w", err)
	}

	// Convert value to bytes
	valueBytes, err := convertOutputsForSerialization([]interface{}{value})
	if err != nil {
		return fmt.Errorf("failed to serialize signal value: %w", err)
	}

	// Store the value
	if err := t.database.SetSignalExecutionDataPropertiesByExecutionID(
		latestExec.ID,
		SetSignalExecutionDataValue(valueBytes[0])); err != nil {
		return fmt.Errorf("failed to store signal value: %w", err)
	}

	// Update execution status
	if err := t.database.SetSignalExecutionProperties(
		latestExec.ID,
		SetSignalExecutionStatus(ExecutionStatusCompleted)); err != nil {
		return fmt.Errorf("failed to update signal execution status: %w", err)
	}

	// Update entity status
	if err := t.database.SetSignalEntityProperties(
		signalEntityID,
		SetSignalEntityStatus(StatusCompleted)); err != nil {
		return fmt.Errorf("failed to update signal entity status: %w", err)
	}

	logger.Debug(t.ctx, "published signal to database",
		"workflow_id", workflowID,
		"signal_name", signal,
		"signal_entity_id", signalEntityID,
		"signal_execution_id", latestExec.ID)

	return nil
}

func New(ctx context.Context, db Database, options ...TempoliteOption) (*Tempolite, error) {
	ctx, cancel := context.WithCancel(ctx)
	t := &Tempolite{
		registry:       NewRegistry(),
		database:       db,
		queueInstances: make(map[string]*QueueInstance),
		ctx:            ctx,
		cancel:         cancel,
		defaultQueue:   "default",
		signals:        make(map[string]Future),
	}

	// Apply options before creating default queue
	for _, opt := range options {
		if err := opt(t); err != nil {
			cancel()
			return nil, err
		}
	}

	// Create default queue if it doesn't exist
	t.mu.Lock()
	if _, exists := t.queueInstances[t.defaultQueue]; !exists {
		if err := t.createQueueLocked(QueueConfig{
			Name:    t.defaultQueue,
			MaxRuns: 2,
		}); err != nil {
			t.mu.Unlock()
			cancel()
			return nil, err
		}
	}
	t.mu.Unlock()

	return t, nil
}

func (t *Tempolite) RegisterWorkflow(handler interface{}) error {
	_, err := t.registry.RegisterWorkflow(handler)
	return err
}

func (t *Tempolite) createQueueLocked(config QueueConfig) error {
	if config.MaxRuns <= 0 {
		return fmt.Errorf("worker count must be greater than 0")
	}

	queueInstance, err := NewQueueInstance(
		t.ctx,
		t.database,
		t.registry,
		config.Name,
		config.MaxRuns,
		config.MaxWorkflows,
		WithCrossWorkflowHandler(t.createCrossWorkflowHandler()),
		WithContinueAsNewHandler(t.createContinueAsNewHandler()),
		WithSignalNewHandler(t.createSignalNewHandler()),
		WithSignalRemoveHandler(t.createSignalRemoveHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to create queue instance: %w", err)
	}

	t.queueInstances[config.Name] = queueInstance
	return nil
}

func (t *Tempolite) Execute(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (Future, error) {
	t.mu.RLock()
	queue, exists := t.queueInstances[queueName]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("queue %s not found", queueName)
	}

	future, _, _, err := queue.Submit(workflowFunc, options, nil, args...)
	if err != nil {
		fmt.Println("failed to submit task", err)
		return nil, fmt.Errorf("failed to submit workflow to queue %s: %w", queueName, err)
	}

	// <-queued.Done()

	if err := future.WaitForIDs(t.ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for workflow IDs: %w", err)
	}

	return future, nil
}

// ExecuteDefault executes a workflow on the default queue
func (t *Tempolite) ExecuteDefault(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (Future, error) {
	return t.Execute(t.defaultQueue, workflowFunc, options, args...)
}

func (t *Tempolite) CreateQueue(config QueueConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.createQueueLocked(config)
}

func (t *Tempolite) Metrics() map[string]retrypool.BlockingMetricsSnapshot[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], RunID] {
	t.mu.RLock()
	defer t.mu.RUnlock()
	metrics := make(map[string]retrypool.BlockingMetricsSnapshot[*retrypool.BlockingRequestResponse[*WorkflowRequest, *WorkflowResponse, RunID, WorkflowEntityID], RunID])
	for name, queue := range t.queueInstances {
		metrics[name] = queue.metrics()
	}
	return metrics
}

// To respect the deterministic principles of the Workflow you need to avoid dynamically creating sub-workflows that you cannot control.
// For example, creating a for-loop that creates many many workflows and wait for them all within a workflow is not recommended.
// You should instead defer the execution of those workflows and store their IDs as a result of the parent workflow and then Enqueue manually the sub-workflows.
func (t *Tempolite) Enqueue(id WorkflowEntityID) (Future, error) {
	// Get workflow entity with queue info
	workflowEntity, err := t.database.GetWorkflowEntity(id, WorkflowEntityWithQueue())
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	if workflowEntity.Status != StatusPending {
		return nil, fmt.Errorf("workflow %d is not in pending state", id)
	}

	// Determine target queue
	queueName := workflowEntity.Edges.Queue.Name
	if queueName == "" {
		queueName = t.defaultQueue
	}

	// Get queue instance
	t.mu.RLock()
	queue, exists := t.queueInstances[queueName]
	t.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("queue %s not found", queueName)
	}

	// Submit enqueue request to queue
	future, _, queued, err := queue.SubmitEnqueue(id)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue workflow: %w", err)
	}

	<-queued.Done()

	return future, nil
}

// Pause a running workflow at the next context call
func (t *Tempolite) Pause(queueName string, id WorkflowEntityID) error {
	t.mu.RLock()
	queue, exists := t.queueInstances[queueName]
	t.mu.RUnlock()
	if !exists {
		return fmt.Errorf("queue %s not found", queueName)
	}
	return queue.Pause(id)
}

// Resume a paused workflow
func (t *Tempolite) Resume(id WorkflowEntityID) (Future, error) {
	// Get workflow entity with queue info
	workflowEntity, err := t.database.GetWorkflowEntity(id, WorkflowEntityWithQueue())
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	// Determine target queue
	queueName := workflowEntity.Edges.Queue.Name
	if queueName == "" {
		queueName = t.defaultQueue
	}

	// Get queue instance
	t.mu.RLock()
	queue, exists := t.queueInstances[queueName]
	t.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("queue %s not found", queueName)
	}

	// Submit resume request to queue
	future, _, queued, err := queue.SubmitResume(id)
	if err != nil {
		return nil, fmt.Errorf("failed to resume workflow: %w", err)
	}

	<-queued.Done()

	return future, nil
}

func (t *Tempolite) CountQueue(queueName string, status EntityStatus) (int, error) {
	q, err := t.database.GetQueueByName(queueName)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue: %w", err)
	}
	return t.database.CountWorkflowEntityByQueueByStatus(q.ID, status)
}

func (t *Tempolite) Scale(queueName string, targetCount int) error {
	t.mu.RLock()
	instance, exists := t.queueInstances[queueName]
	t.mu.RUnlock()
	if !exists {
		return fmt.Errorf("queue %s not found", queueName)
	}

	instance.orchestrators.SetConcurrentPools(targetCount)
	// TODO: also need to be able to scale concurrent workers per pool

	return nil
}

func (t *Tempolite) GetWorkflow(id WorkflowEntityID) (*WorkflowEntity, error) {
	return t.database.GetWorkflowEntity(
		id,
		// TODO: should we expose the options?!
		WorkflowEntityWithData(),
		WorkflowEntityWithVersion(),
		WorkflowEntityWithQueue(),
		WorkflowEntityWithRun(),
	)
}

func (t *Tempolite) Get(id WorkflowEntityID) (Future, error) {
	// Get workflow entity with queue info
	workflowEntity, err := t.database.GetWorkflowEntity(id, WorkflowEntityWithQueue())
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}
	var handler HandlerInfo
	var ok bool
	if handler, ok = t.registry.GetWorkflow(workflowEntity.HandlerName); !ok {
		return nil, fmt.Errorf("workflow handler %s not found", workflowEntity.HandlerName)
	}
	var future Future = NewRuntimeFuture()
	future.setEntityID(FutureEntityWithWorkflowID(workflowEntity.ID))

	exec, err := t.database.GetWorkflowExecutionLatestByEntityID(workflowEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution: %w", err)
	}

	data, err := t.database.GetWorkflowExecutionDataByExecutionID(exec.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution data: %w", err)
	}

	results, err := convertOutputsFromSerialization(handler, data.Outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert outputs: %w", err)
	}

	future.SetResult(results)

	return future, nil
}

func (t *Tempolite) Close() error {
	t.cancel()

	t.mu.Lock()
	defer t.mu.Unlock()

	// Close all queues
	var errs []error
	for _, instance := range t.queueInstances {
		instance.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}
	return nil
}

func (t *Tempolite) Wait() error {
	shutdown := errgroup.Group{}

	shutdown.Go(func() error {
		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()
		// defer fmt.Println("shutting down")
		for {
			select {
			case <-timer.C:
				// pp.Println(t.Metrics())
				hasStillTasks := false
				t.mu.RLock()
				for _, instance := range t.queueInstances {
					metrics := instance.metrics()
					// task queues
					for _, poolMetric := range metrics.Metrics {
						for _, worker := range poolMetric.Workers {
							if worker.HasTask {
								hasStillTasks = true
							}
						}
						for _, v := range poolMetric.TaskQueues {
							if v > 0 {
								hasStillTasks = true
							}
						}
					}
				}
				t.mu.RUnlock()
				if !hasStillTasks {
					// fmt.Println("no more tasks")
					return nil
				}
				// fmt.Println("still tasks")
				continue
			case <-t.ctx.Done():
				return nil
			}
		}
	})

	return shutdown.Wait()
}

// Option functions
func WithQueue(config QueueConfig) TempoliteOption {
	return func(t *Tempolite) error {
		return t.createQueueLocked(config)
	}
}

func WithDefaultQueueWorkers(maxConcurrentRuns int, maxConcurrentWorkflows int) TempoliteOption {
	return func(t *Tempolite) error {
		return t.createQueueLocked(QueueConfig{
			Name:         "default",
			MaxRuns:      maxConcurrentRuns,      // pools
			MaxWorkflows: maxConcurrentWorkflows, // workers per pool, -1 unlimieted
		})
	}
}

func WithWorkflows(workflowFunc ...interface{}) TempoliteOption {
	return func(t *Tempolite) error {
		for _, wf := range workflowFunc {
			if _, err := t.registry.RegisterWorkflow(wf); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithLog(log Logger) TempoliteOption {
	return func(t *Tempolite) error {
		logger = log
		return nil
	}
}
