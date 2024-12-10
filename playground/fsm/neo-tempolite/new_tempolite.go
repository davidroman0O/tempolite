package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	future       Future           // Future for tracking execution
	queueName    string           // Name of the queue this task belongs to
	continued    bool
	resume       bool
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

	name  string
	count int

	orchestrators *retrypool.Pool[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]]

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

func (q *QueueInstance) Pause(id WorkflowEntityID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	worker, exists := q.processingWorkers[id]
	if !exists {
		return fmt.Errorf("entity %d not found", id)
	}
	worker.Pause()
	return nil
}

func NewQueueInstance(ctx context.Context, db Database, registry *Registry, name string, count int, opt ...queueOption) (*QueueInstance, error) {
	q := &QueueInstance{
		name:              name,
		count:             count,
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

	workers := []retrypool.Worker[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]]{}
	for i := 0; i < count; i++ {
		workers = append(workers, NewQueueWorker(q))
	}

	q.orchestrators = retrypool.New(
		ctx,
		workers,
		retrypool.WithAttempts[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](1),
		retrypool.WithDelay[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](time.Second/2),
		retrypool.WithOnNewDeadTask[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](
			func(task *retrypool.DeadTask[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]], idx int) {
				errs := errors.New("failed to process request")
				for _, e := range task.Errors {
					errs = errors.Join(errs, e)
				}
				task.Data.CompleteWithError(errs)
				_, err := q.orchestrators.PullDeadTask(idx)
				if err != nil {
					// too bad
					log.Printf("failed to pull dead task: %v", err)
				}
			}),
	)

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
		return queueSize > 0 || processingCount > 0
	}, time.Second)
}

func (qi *QueueInstance) Submit(workflowFunc interface{}, options *WorkflowOptions, opts *preparationOptions, args ...interface{}) (Future, *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse], <-chan struct{}, error) {

	workflowEntity, err := PrepareWorkflow(qi.registry, qi.database, workflowFunc, options, opts, args...)
	if err != nil {
		log.Printf("failed to prepare workflow: %v", err)
		return nil, nil, nil, err
	}

	dbFuture := NewRuntimeFuture()
	dbFuture.setEntityID(FutureEntityWithWorkflowID(workflowEntity.ID))

	queued := retrypool.NewQueuedNotification()

	task := retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](
		&WorkflowRequest{
			workflowID:   workflowEntity.ID,
			workflowFunc: workflowFunc,
			options:      options,
			args:         args,
			future:       dbFuture,
			queueName:    qi.name,
		},
	)

	if err := qi.orchestrators.Submit(task, retrypool.WithQueued[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](queued)); err != nil {
		return nil, nil, nil, err
	}

	return dbFuture, task, queued.Done(), nil
}

func (qi *QueueInstance) SubmitResume(entityID WorkflowEntityID) (Future, *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse], <-chan struct{}, error) {
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

	// Create future for tracking execution
	dbFuture := NewRuntimeFuture()
	dbFuture.setEntityID(FutureEntityWithWorkflowID(workflowEntity.ID))

	queued := retrypool.NewQueuedNotification()

	// Create resume request
	task := retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](
		&WorkflowRequest{
			workflowID:   workflowEntity.ID,
			workflowFunc: handler.Handler, // Use the actual handler function
			queueName:    qi.name,
			continued:    false, // Not a continuation
			resume:       true,  // Mark this as a resume
			future:       dbFuture,
		},
	)

	if err := qi.orchestrators.Submit(task, retrypool.WithQueued[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](queued)); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to submit resume task: %w", err)
	}

	return dbFuture, task, queued.Done(), nil
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

	instance.mu.Lock()
	instance.workers[qw.ID] = qw
	instance.mu.Unlock()

	onStart := func(ctx context.Context) {

	}

	onEnd := func(ctx context.Context) {

	}

	qw.onStart = onStart
	qw.onEnd = onEnd

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
	return qw
}

func (w *QueueWorker) OnStart(ctx context.Context) {
	// log.Printf("Queue worker started for queue %s", w.queueInstance.name)
}

func (w *QueueWorker) Run(ctx context.Context, task *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]) error {

	// Mark entity as processing
	w.queueInstance.mu.Lock()
	if _, exists := w.queueInstance.processingWorkers[task.Request.workflowID]; exists {
		w.queueInstance.mu.Unlock()
		task.Request.future.SetError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		task.CompleteWithError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		return fmt.Errorf("entity %d is already being processed", task.Request.workflowID)
	}
	w.queueInstance.processingWorkers[task.Request.workflowID] = w
	w.queueInstance.mu.Unlock()

	// Clean up when done
	defer func() {
		w.queueInstance.mu.Lock()
		delete(w.queueInstance.processingWorkers, task.Request.workflowID)
		w.queueInstance.mu.Unlock()
	}()

	if _, err := w.orchestrator.registry.RegisterWorkflow(task.Request.workflowFunc); err != nil {
		task.Request.future.SetError(err)
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

	if err != nil {
		task.Request.future.SetError(err)
		task.CompleteWithError(err)
		// return fmt.Errorf("failed to execute workflow: %w", err)
		return nil
	}

	var results []interface{}
	// Get the results and update the DatabaseFuture
	if results, err = future.GetResults(); err != nil {
		task.Request.future.SetError(err)
		task.CompleteWithError(err)
		//	it's fine, we have an successful error
		return nil
	}

	task.Request.future.SetResult(results)
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
	Name        string
	WorkerCount int
}

// TempoliteOption is a function type for configuring Tempolite
type TempoliteOption func(*Tempolite) error

// createCrossWorkflowHandler creates the handler function for cross-queue workflow communication
func (t *Tempolite) createCrossWorkflowHandler() crossQueueWorkflowHandler {
	return func(queueName string, workflowID WorkflowEntityID, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
		t.mu.RLock()
		queue, ok := t.queueInstances[queueName]
		t.mu.RUnlock()
		if !ok {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(fmt.Errorf("queue %s not found", queueName))
			return futureErr
		}

		future := NewRuntimeFuture()

		if err := queue.orchestrators.Submit(
			retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
				workflowFunc: workflowFunc,
				options:      options,
				workflowID:   workflowID,
				args:         args,
				future:       future,
				queueName:    queueName,
				continued:    false,
			})); err != nil {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(err)
			return futureErr
		}

		return future
	}
}

func (t *Tempolite) createContinueAsNewHandler() crossQueueContinueAsNewHandler {
	return func(queueName string, workflowID WorkflowEntityID, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
		t.mu.Lock()
		queue, ok := t.queueInstances[queueName]
		t.mu.Unlock()
		if !ok {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(fmt.Errorf("queue %s not found", queueName))
			return futureErr
		}

		future := NewRuntimeFuture()

		// Handle version inheritance through ContinueAsNew
		if err := queue.orchestrators.Submit(
			retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
				workflowID:   workflowID,
				workflowFunc: workflowFunc,
				options:      options,
				args:         args,
				future:       future,
				queueName:    queueName,
				continued:    true, // Mark this as a continuation
			})); err != nil {
			futureErr := NewRuntimeFuture()
			futureErr.SetError(err)
			return futureErr
		}

		return future
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
			Name:        t.defaultQueue,
			WorkerCount: 1,
		}); err != nil {
			t.mu.Unlock()
			cancel()
			return nil, err
		}
	}
	t.mu.Unlock()

	return t, nil
}

func (t *Tempolite) createQueueLocked(config QueueConfig) error {
	if config.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be greater than 0")
	}

	queueInstance, err := NewQueueInstance(
		t.ctx,
		t.database,
		t.registry,
		config.Name,
		config.WorkerCount,
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

	future, _, queued, err := queue.Submit(workflowFunc, options, nil, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to submit workflow to queue %s: %w", queueName, err)
	}

	<-queued

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

func (t *Tempolite) Pause(queueName string, id WorkflowEntityID) error {
	t.mu.RLock()
	queue, exists := t.queueInstances[queueName]
	t.mu.RUnlock()
	if !exists {
		return fmt.Errorf("queue %s not found", queueName)
	}
	return queue.Pause(id)
}

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

	<-queued

	return future, nil
}

func (t *Tempolite) ScaleQueue(queueName string, delta int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	instance, exists := t.queueInstances[queueName]
	if !exists {
		return fmt.Errorf("queue %s not found", queueName)
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	logger.Info(t.ctx, "Scaling queue", "queueName", queueName, "target", delta)

	if delta > 0 {
		// Add workers
		for i := 0; i < delta; i++ {
			worker := NewQueueWorker(instance)
			instance.orchestrators.AddWorker(worker)
			logger.Info(t.ctx, "Added worker to queue", "queueName", queueName)
		}
		instance.count += delta
	} else if delta < 0 {
		// Remove workers
		count := -delta
		if count > instance.count {
			return fmt.Errorf("cannot remove %d workers, only %d available", count, instance.count)
		}
		workers := instance.orchestrators.ListWorkers()
		for i := 0; i < count && i < len(workers); i++ {
			logger.Info(t.ctx, "Removing worker from queue", "queueName", queueName)
			if err := instance.orchestrators.RemoveWorker(workers[len(workers)-1-i].ID); err != nil {
				return fmt.Errorf("failed to remove worker: %w", err)
			}
			logger.Info(t.ctx, "Removed worker from queue %s", queueName)
		}
		instance.count -= count
	}

	logger.Info(t.ctx, "Queue scaled", "queueName", queueName, "target", delta)

	return nil
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

	t.mu.RLock()
	for _, instance := range t.queueInstances {
		instance := instance
		shutdown.Go(func() error {
			return instance.Wait()
		})
	}
	t.mu.RUnlock()

	return shutdown.Wait()
}

// Option functions
func WithQueue(config QueueConfig) TempoliteOption {
	return func(t *Tempolite) error {
		return t.createQueueLocked(config)
	}
}

func WithDefaultQueueWorkers(count int) TempoliteOption {
	return func(t *Tempolite) error {
		return t.createQueueLocked(QueueConfig{
			Name:        "default",
			WorkerCount: count,
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
