package tempolite

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/davidroman0O/retrypool"
// 	"github.com/sasha-s/go-deadlock"
// 	"golang.org/x/sync/errgroup"
// )

// // WorkflowRequest represents a workflow execution request
// type WorkflowRequest struct {
// 	workflowID   int
// 	workflowFunc interface{}      // The workflow function
// 	options      *WorkflowOptions // Workflow options
// 	args         []interface{}    // Arguments for the workflow
// 	future       Future           // Future for tracking execution
// 	queueName    string           // Name of the queue this task belongs to
// 	continued    bool
// }

// type WorkflowResponse struct {
// 	entityID int // ID of the workflow entity to execute
// }

// // QueueInstance manages a queue and its worker pool
// type QueueInstance struct {
// 	mu deadlock.RWMutex

// 	ctx    context.Context
// 	cancel context.CancelFunc

// 	registry *Registry
// 	database Database

// 	name  string
// 	count int

// 	orchestrators *retrypool.Pool[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]]

// 	processingWorkers map[int]struct{}
// 	freeWorkers       map[int]struct{}

// 	onCrossWorkflow crossQueueWorkflowHandler
// 	onContinueAsNew crossQueueContinueAsNewHandler
// 	onSignalNew     workflowSignalHandler
// }

// type queueConfig struct {
// 	onCrossWorkflow crossQueueWorkflowHandler
// 	onContinueAsNew crossQueueContinueAsNewHandler
// 	onSignalNew     workflowSignalHandler
// }

// type queueOption func(*queueConfig)

// func WithCrossWorkflowHandler(handler crossQueueWorkflowHandler) queueOption {
// 	return func(c *queueConfig) {
// 		c.onCrossWorkflow = handler
// 	}
// }

// func WithContinueAsNewHandler(handler crossQueueContinueAsNewHandler) queueOption {
// 	return func(c *queueConfig) {
// 		c.onContinueAsNew = handler
// 	}
// }

// func WithSignalNewHandler(handler workflowSignalHandler) queueOption {
// 	return func(c *queueConfig) {
// 		c.onSignalNew = handler
// 	}
// }

// func NewQueueInstance(ctx context.Context, db Database, registry *Registry, name string, count int, opt ...queueOption) (*QueueInstance, error) {
// 	q := &QueueInstance{
// 		name:              name,
// 		count:             count,
// 		registry:          registry,
// 		database:          db,
// 		ctx:               ctx,
// 		processingWorkers: make(map[int]struct{}),
// 		freeWorkers:       make(map[int]struct{}),
// 	}

// 	cfg := &queueConfig{}

// 	for _, o := range opt {
// 		o(cfg)
// 	}

// 	if cfg.onCrossWorkflow != nil {
// 		q.onCrossWorkflow = cfg.onCrossWorkflow
// 	}
// 	if cfg.onContinueAsNew != nil {
// 		q.onContinueAsNew = cfg.onContinueAsNew
// 	}
// 	if cfg.onSignalNew != nil {
// 		q.onSignalNew = cfg.onSignalNew
// 	}

// 	_, err := db.AddQueue(&Queue{Name: name})
// 	if err != nil {
// 		if !errors.Is(err, ErrQueueExists) {
// 			return nil, err
// 		}
// 	}

// 	workers := []retrypool.Worker[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]]{}
// 	for i := 0; i < count; i++ {
// 		workers = append(workers, NewQueueWorker(q))
// 	}

// 	q.orchestrators = retrypool.New(
// 		ctx,
// 		workers,
// 		retrypool.WithAttempts[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](3),
// 		retrypool.WithDelay[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](time.Second/2),
// 		retrypool.WithOnNewDeadTask[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]](
// 			func(task *retrypool.DeadTask[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]], idx int) {
// 				errs := errors.New("failed to process request")
// 				for _, e := range task.Errors {
// 					errs = errors.Join(errs, e)
// 				}
// 				task.Data.CompleteWithError(errs)
// 				_, err := q.orchestrators.PullDeadTask(idx)
// 				if err != nil {
// 					// too bad
// 					log.Printf("failed to pull dead task: %v", err)
// 				}
// 			}),
// 	)

// 	return q, nil
// }

// func (qi *QueueInstance) Close() error {
// 	if err := qi.orchestrators.Close(); err != nil {
// 		if err != context.Canceled {
// 			return fmt.Errorf("failed to close orchestrators: %w", err)
// 		}
// 	}
// 	return nil
// }

// func (qi *QueueInstance) Wait() error {
// 	return qi.orchestrators.WaitWithCallback(qi.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
// 		return queueSize > 0 || processingCount > 0
// 	}, time.Second)
// }

// func (qi *QueueInstance) Submit(workflowFunc interface{}, options *WorkflowOptions, opts *preparationOptions, args ...interface{}) (Future, *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse], error) {

// 	workflowEntity, err := prepareWorkflow(qi.registry, qi.database, workflowFunc, options, opts, args...)
// 	if err != nil {
// 		log.Printf("failed to prepare workflow: %v", err)
// 		return nil, nil, err
// 	}

// 	dbFuture := NewRuntimeFuture()
// 	dbFuture.setEntityID(workflowEntity.ID)

// 	task := retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](
// 		&WorkflowRequest{
// 			workflowID:   workflowEntity.ID,
// 			workflowFunc: workflowFunc,
// 			options:      options,
// 			args:         args,
// 			future:       dbFuture,
// 			queueName:    qi.name,
// 		},
// 	)

// 	if err := qi.orchestrators.Submit(task); err != nil {
// 		return nil, nil, err
// 	}

// 	return dbFuture, task, nil
// }

// type onStartFunc func(context.Context)
// type onEndFunc func(context.Context)

// // QueueWorker implements the retrypool.Worker interface
// type QueueWorker struct {
// 	ID            int
// 	queueInstance *QueueInstance
// 	orchestrator  *Orchestrator

// 	onStart onStartFunc
// 	onEnd   onEndFunc
// }

// func NewQueueWorker(instance *QueueInstance) *QueueWorker {
// 	qw := &QueueWorker{
// 		queueInstance: instance,
// 	}
// 	opts := []OrchestratorOption{}
// 	if instance.onCrossWorkflow != nil {
// 		opts = append(opts, WithCrossWorkflow(instance.onCrossWorkflow))
// 	}
// 	if instance.onContinueAsNew != nil {
// 		opts = append(opts, WithContinueAsNew(instance.onContinueAsNew))
// 	}
// 	if instance.onSignalNew != nil {
// 		opts = append(opts, WithSignalNew(instance.onSignalNew))
// 	}
// 	qw.orchestrator = NewOrchestrator(instance.ctx, instance.database, instance.registry, opts...)
// 	return qw
// }

// func (w *QueueWorker) OnStart(ctx context.Context) {
// 	// log.Printf("Queue worker started for queue %s", w.queueInstance.name)
// }

// func (w *QueueWorker) Run(ctx context.Context, task *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]) error {

// 	// Mark entity as processing
// 	w.queueInstance.mu.Lock()
// 	if _, exists := w.queueInstance.processingWorkers[task.Request.workflowID]; exists {
// 		w.queueInstance.mu.Unlock()
// 		task.Request.future.setError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
// 		task.CompleteWithError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
// 		return fmt.Errorf("entity %d is already being processed", task.Request.workflowID)
// 	}
// 	w.queueInstance.processingWorkers[task.Request.workflowID] = struct{}{}
// 	w.queueInstance.mu.Unlock()

// 	// Clean up when done
// 	defer func() {
// 		w.queueInstance.mu.Lock()
// 		delete(w.queueInstance.processingWorkers, task.Request.workflowID)
// 		w.queueInstance.mu.Unlock()
// 	}()

// 	if _, err := w.orchestrator.registry.RegisterWorkflow(task.Request.workflowFunc); err != nil {
// 		task.Request.future.setError(err)
// 		task.CompleteWithError(err)
// 		return fmt.Errorf("failed to register workflow: %w", err)
// 	}

// 	// Execute the workflow using the orchestrator
// 	future, err := w.orchestrator.ExecuteWithEntity(task.Request.workflowID)
// 	if err != nil {
// 		task.Request.future.setError(err)
// 		task.CompleteWithError(err)
// 		return fmt.Errorf("failed to execute workflow: %w", err)
// 	}

// 	var results []interface{}
// 	// Get the results and update the DatabaseFuture
// 	if results, err = future.GetResults(); err != nil {
// 		task.Request.future.setError(err)
// 		task.CompleteWithError(err)
// 		return err
// 	}

// 	task.Request.future.setResult(results)
// 	task.Complete(&WorkflowResponse{
// 		entityID: task.Request.workflowID,
// 	})

// 	return nil
// }

// // Tempolite is the main orchestration engine
// type Tempolite struct {
// 	mu             deadlock.RWMutex
// 	queueInstances map[string]*QueueInstance
// 	registry       *Registry
// 	database       Database
// 	ctx            context.Context
// 	cancel         context.CancelFunc
// 	defaultQueue   string

// 	signals map[string]Future
// }

// // QueueConfig holds configuration for a queue
// type QueueConfig struct {
// 	Name        string
// 	WorkerCount int
// }

// // TempoliteOption is a function type for configuring Tempolite
// type TempoliteOption func(*Tempolite) error

// // createCrossWorkflowHandler creates the handler function for cross-queue workflow communication
// func (t *Tempolite) createCrossWorkflowHandler() crossQueueWorkflowHandler {
// 	return func(queueName string, workflowID int, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
// 		t.mu.RLock()
// 		queue, ok := t.queueInstances[queueName]
// 		t.mu.RUnlock()
// 		if !ok {
// 			futureErr := NewRuntimeFuture()
// 			futureErr.setError(fmt.Errorf("queue %s not found", queueName))
// 			return futureErr
// 		}

// 		future := NewRuntimeFuture()

// 		if err := queue.orchestrators.Submit(
// 			retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
// 				workflowFunc: workflowFunc,
// 				options:      options,
// 				workflowID:   workflowID,
// 				args:         args,
// 				future:       future,
// 				queueName:    queueName,
// 				continued:    false,
// 			})); err != nil {
// 			futureErr := NewRuntimeFuture()
// 			futureErr.setError(err)
// 			return futureErr
// 		}

// 		return future
// 	}
// }

// func (t *Tempolite) createContinueAsNewHandler() crossQueueContinueAsNewHandler {
// 	return func(queueName string, workflowID int, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
// 		t.mu.Lock()
// 		queue, ok := t.queueInstances[queueName]
// 		t.mu.Unlock()
// 		if !ok {
// 			futureErr := NewRuntimeFuture()
// 			futureErr.setError(fmt.Errorf("queue %s not found", queueName))
// 			return futureErr
// 		}

// 		future := NewRuntimeFuture()

// 		// Handle version inheritance through ContinueAsNew
// 		if err := queue.orchestrators.Submit(
// 			retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
// 				workflowID:   workflowID,
// 				workflowFunc: workflowFunc,
// 				options:      options,
// 				args:         args,
// 				future:       future,
// 				queueName:    queueName,
// 				continued:    true, // Mark this as a continuation
// 			})); err != nil {
// 			futureErr := NewRuntimeFuture()
// 			futureErr.setError(err)
// 			return futureErr
// 		}

// 		return future
// 	}
// }

// func (t *Tempolite) createSignalNewHandler() workflowSignalHandler {
// 	return func(workflowID int, signal string) Future {

// 		t.mu.Lock()
// 		future := NewRuntimeFuture()
// 		future.setEntityID(workflowID)
// 		t.signals[signal] = future
// 		t.mu.Unlock()

// 		return future
// 	}
// }

// func (t *Tempolite) PublishSignal(workflowID int, signal string, value interface{}) error {
// 	t.mu.RLock()
// 	future, ok := t.signals[signal]
// 	t.mu.RUnlock()
// 	if !ok {
// 		return fmt.Errorf("signal %s not found", signal)
// 	}
// 	future.setResult([]interface{}{value})
// 	return nil
// }

// func New(ctx context.Context, db Database, options ...TempoliteOption) (*Tempolite, error) {
// 	ctx, cancel := context.WithCancel(ctx)
// 	t := &Tempolite{
// 		registry:       NewRegistry(),
// 		database:       db,
// 		queueInstances: make(map[string]*QueueInstance),
// 		ctx:            ctx,
// 		cancel:         cancel,
// 		defaultQueue:   "default",
// 		signals:        make(map[string]Future),
// 	}

// 	// Apply options before creating default queue
// 	for _, opt := range options {
// 		if err := opt(t); err != nil {
// 			cancel()
// 			return nil, err
// 		}
// 	}

// 	// Create default queue if it doesn't exist
// 	t.mu.Lock()
// 	if _, exists := t.queueInstances[t.defaultQueue]; !exists {
// 		if err := t.createQueueLocked(QueueConfig{
// 			Name:        t.defaultQueue,
// 			WorkerCount: 1,
// 		}); err != nil {
// 			t.mu.Unlock()
// 			cancel()
// 			return nil, err
// 		}
// 	}
// 	t.mu.Unlock()

// 	return t, nil
// }

// func (t *Tempolite) createQueueLocked(config QueueConfig) error {
// 	if config.WorkerCount <= 0 {
// 		return fmt.Errorf("worker count must be greater than 0")
// 	}

// 	queueInstance, err := NewQueueInstance(
// 		t.ctx,
// 		t.database,
// 		t.registry,
// 		config.Name,
// 		config.WorkerCount,
// 		WithCrossWorkflowHandler(t.createCrossWorkflowHandler()),
// 		WithContinueAsNewHandler(t.createContinueAsNewHandler()),
// 		WithSignalNewHandler(t.createSignalNewHandler()),
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to create queue instance: %w", err)
// 	}

// 	t.queueInstances[config.Name] = queueInstance
// 	return nil
// }

// func (t *Tempolite) Execute(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (Future, error) {
// 	t.mu.RLock()
// 	queue, exists := t.queueInstances[queueName]
// 	t.mu.RUnlock()

// 	if !exists {
// 		return nil, fmt.Errorf("queue %s not found", queueName)
// 	}

// 	future, _, err := queue.Submit(workflowFunc, options, nil, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to submit workflow to queue %s: %w", queueName, err)
// 	}

// 	return future, nil
// }

// // ExecuteDefault executes a workflow on the default queue
// func (t *Tempolite) ExecuteDefault(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (Future, error) {
// 	return t.Execute(t.defaultQueue, workflowFunc, options, args...)
// }

// func (t *Tempolite) CreateQueue(config QueueConfig) error {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()
// 	return t.createQueueLocked(config)
// }

// func (t *Tempolite) ScaleQueue(queueName string, delta int) error {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	instance, exists := t.queueInstances[queueName]
// 	if !exists {
// 		return fmt.Errorf("queue %s not found", queueName)
// 	}

// 	instance.mu.Lock()
// 	defer instance.mu.Unlock()

// 	if delta > 0 {
// 		// Add workers
// 		for i := 0; i < delta; i++ {
// 			worker := NewQueueWorker(instance)
// 			instance.orchestrators.AddWorker(worker)
// 		}
// 		instance.count += delta
// 	} else if delta < 0 {
// 		// Remove workers
// 		count := -delta
// 		if count > instance.count {
// 			return fmt.Errorf("cannot remove %d workers, only %d available", count, instance.count)
// 		}
// 		workers := instance.orchestrators.ListWorkers()
// 		for i := 0; i < count && i < len(workers); i++ {
// 			if err := instance.orchestrators.RemoveWorker(workers[len(workers)-1-i].ID); err != nil {
// 				return fmt.Errorf("failed to remove worker: %w", err)
// 			}
// 		}
// 		instance.count -= count
// 	}

// 	return nil
// }

// func (t *Tempolite) Close() error {
// 	t.cancel()

// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	// Close all queues
// 	var errs []error
// 	for _, instance := range t.queueInstances {
// 		instance.Close()
// 	}

// 	if len(errs) > 0 {
// 		return fmt.Errorf("errors during shutdown: %v", errs)
// 	}
// 	return nil
// }

// func (t *Tempolite) Wait() error {
// 	shutdown := errgroup.Group{}

// 	t.mu.RLock()
// 	for _, instance := range t.queueInstances {
// 		instance := instance
// 		shutdown.Go(func() error {
// 			return instance.Wait()
// 		})
// 	}
// 	t.mu.RUnlock()

// 	return shutdown.Wait()
// }

// // Option functions
// func WithQueue(config QueueConfig) TempoliteOption {
// 	return func(t *Tempolite) error {
// 		return t.createQueueLocked(config)
// 	}
// }

// func WithDefaultQueueWorkers(count int) TempoliteOption {
// 	return func(t *Tempolite) error {
// 		return t.createQueueLocked(QueueConfig{
// 			Name:        "default",
// 			WorkerCount: count,
// 		})
// 	}
// }
