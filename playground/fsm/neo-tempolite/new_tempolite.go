package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/sasha-s/go-deadlock"
)

// DatabaseFuture implements the Future interface for cross-queue workflow communication
type DatabaseFuture struct {
	ctx        context.Context
	entityID   int
	database   Database
	registry   *Registry
	results    []interface{}
	err        error
	continueAs int
}

func NewDatabaseFuture(ctx context.Context, database Database, registry *Registry) *DatabaseFuture {
	return &DatabaseFuture{
		ctx:      ctx,
		database: database,
		registry: registry,
	}
}

func (f *DatabaseFuture) setEntityID(entityID int) {
	f.entityID = entityID
}

func (f *DatabaseFuture) setError(err error) {
	f.err = err
}

func (f *DatabaseFuture) WorkflowID() int {
	return f.entityID
}

func (f *DatabaseFuture) Get(out ...interface{}) error {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return f.ctx.Err()
		case <-ticker.C:
			if f.entityID == 0 {
				continue
			}
			if f.err != nil {
				defer func() { // if we re-use the same future, we must start from a fresh state
					f.err = nil
					f.results = nil
				}()
				return f.err
			}
			if completed := f.checkCompletion(); completed {
				defer func() { // if we re-use the same future, we must start from a fresh state
					f.err = nil
					f.results = nil
				}()
				return f.handleResults(out...)
			}
		}
	}
}

func (f *DatabaseFuture) checkCompletion() bool {
	var status EntityStatus
	if err := f.database.GetWorkflowEntityProperties(f.entityID, GetWorkflowEntityStatus(&status)); err != nil {
		f.err = fmt.Errorf("failed to get workflow entity: %w", err)
		return true
	}

	switch status {
	case StatusCompleted:
		// Get latest execution
		latestExec, err := f.database.GetWorkflowExecutionLatestByEntityID(f.entityID)
		if err != nil {
			f.err = fmt.Errorf("failed to get latest execution: %w", err)
			return true
		}

		// Get execution data with outputs
		if latestExec.WorkflowExecutionData != nil {
			var outputs [][]byte
			if err := f.database.GetWorkflowExecutionDataProperties(latestExec.ID,
				GetWorkflowExecutionDataOutputs(&outputs)); err != nil {
				f.err = fmt.Errorf("failed to get execution outputs: %w", err)
				return true
			}

			var handlerName string
			if err := f.database.GetWorkflowEntityProperties(f.entityID,
				GetWorkflowEntityHandlerName(&handlerName)); err != nil {
				f.err = fmt.Errorf("failed to get handler info: %w", err)
				return true
			}

			// Now using registry instead of database to get workflow handler
			handler, ok := f.registry.GetWorkflow(handlerName)
			if !ok {
				f.err = fmt.Errorf("handler %s not found", handlerName)
				return true
			}

			results, err := convertOutputsFromSerialization(handler, outputs)
			if err != nil {
				f.err = fmt.Errorf("failed to deserialize outputs: %w", err)
				return true
			}

			f.results = results
			return true
		}

	case StatusFailed:
		latestExec, err := f.database.GetWorkflowExecutionLatestByEntityID(f.entityID)
		if err != nil {
			f.err = fmt.Errorf("failed to get latest execution: %w", err)
		} else {
			var execError string
			if err := f.database.GetWorkflowExecutionProperties(latestExec.ID,
				GetWorkflowExecutionError(&execError)); err != nil {
				f.err = fmt.Errorf("failed to get execution error: %w", err)
			} else {
				f.err = errors.New(execError)
			}
		}
		return true

	case StatusPaused:
		f.err = ErrPaused
		return true

	case StatusCancelled:
		f.err = errors.New("workflow was cancelled")
		return true
	}

	return false
}

func (f *DatabaseFuture) handleResults(out ...interface{}) error {
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	if len(out) > len(f.results) {
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)",
			len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d",
				result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}

func (f *DatabaseFuture) GetResults() ([]interface{}, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return nil, errors.New("context cancelled")
		case <-ticker.C:
			if f.entityID == 0 {
				continue
			}
			if f.err != nil {
				return nil, f.err
			}
			if completed := f.checkCompletion(); completed {
				if f.err != nil {
					return nil, f.err
				}
				return f.results, nil
			}
			continue
		}
	}
}

func (f *DatabaseFuture) setResult(results []interface{}) {
	f.results = results
}

func (f *DatabaseFuture) setContinueAs(continueAs int) {
	f.continueAs = continueAs
}

func (f *DatabaseFuture) ContinuedAsNew() bool {
	return f.continueAs != 0
}

func (f *DatabaseFuture) ContinuedAs() int {
	return f.continueAs
}

func (f *DatabaseFuture) IsPaused() bool {
	var status EntityStatus
	err := f.database.GetWorkflowEntityProperties(f.entityID, GetWorkflowEntityStatus(&status))
	if err != nil {
		return false
	}
	return status == StatusPaused
}

// WorkflowRequest represents a workflow execution request
type WorkflowRequest struct {
	workflowID   int
	workflowFunc interface{}      // The workflow function
	options      *WorkflowOptions // Workflow options
	args         []interface{}    // Arguments for the workflow
	future       *DatabaseFuture  // Future for tracking execution
	queueName    string           // Name of the queue this task belongs to
}

type WorkflowResponse struct {
	entityID int // ID of the workflow entity to execute
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

	processing map[int]struct{}
}

func NewQueueInstance(ctx context.Context, db Database, registry *Registry, name string, count int) *QueueInstance {
	q := &QueueInstance{
		name:       name,
		count:      count,
		registry:   registry,
		database:   db,
		ctx:        ctx,
		processing: make(map[int]struct{}),
	}

	workers := []retrypool.Worker[*retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]]{}
	for i := 0; i < count; i++ {
		workers = append(workers, NewQueueWorker(q))
	}

	// TODO: add the panic management
	q.orchestrators = retrypool.New(ctx, workers)

	return q
}

func (qi *QueueInstance) Submit(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (*DatabaseFuture, *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse], error) {

	workflowEntity, err := prepareWorkflow(qi.registry, qi.database, workflowFunc, options, args...)
	if err != nil {
		log.Printf("failed to prepare workflow: %v", err)
		return nil, nil, err
	}

	dbFuture := NewDatabaseFuture(qi.ctx, qi.database, qi.registry)
	dbFuture.setEntityID(workflowEntity.ID)

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

	if err := qi.orchestrators.Submit(task); err != nil {
		return nil, nil, err
	}

	return dbFuture, task, nil
}

// QueueWorker implements the retrypool.Worker interface
type QueueWorker struct {
	ID            int
	queueInstance *QueueInstance
	orchestrator  *Orchestrator
}

func NewQueueWorker(instance *QueueInstance) *QueueWorker {
	return &QueueWorker{
		queueInstance: instance,
		orchestrator:  NewOrchestrator(instance.ctx, instance.database, instance.registry),
	}
}

func (w *QueueWorker) OnStart(ctx context.Context) {
	log.Printf("Queue worker started for queue %s", w.queueInstance.name)
}

func (w *QueueWorker) Run(ctx context.Context, task *retrypool.RequestResponse[*WorkflowRequest, *WorkflowResponse]) error {

	// Mark entity as processing
	w.queueInstance.mu.Lock()
	if _, exists := w.queueInstance.processing[task.Request.workflowID]; exists {
		w.queueInstance.mu.Unlock()
		task.Request.future.setError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		task.CompleteWithError(fmt.Errorf("entity %d is already being processed", task.Request.workflowID))
		return fmt.Errorf("entity %d is already being processed", task.Request.workflowID)
	}
	w.queueInstance.processing[task.Request.workflowID] = struct{}{}
	w.queueInstance.mu.Unlock()

	// Clean up when done
	defer func() {
		w.queueInstance.mu.Lock()
		delete(w.queueInstance.processing, task.Request.workflowID)
		w.queueInstance.mu.Unlock()
	}()

	if _, err := w.orchestrator.registry.RegisterWorkflow(task.Request.workflowFunc); err != nil {
		task.Request.future.setError(err)
		task.CompleteWithError(err)
		return fmt.Errorf("failed to register workflow: %w", err)
	}

	// Execute the workflow using the orchestrator
	future, err := w.orchestrator.ExecuteWithEntity(task.Request.workflowID)
	if err != nil {
		task.Request.future.setError(err)
		task.CompleteWithError(err)
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	var results []interface{}
	// Get the results and update the DatabaseFuture
	if results, err = future.GetResults(); err != nil {
		task.Request.future.setError(err)
		task.CompleteWithError(err)
		return err
	}

	task.Request.future.setResult(results)
	task.Complete(&WorkflowResponse{
		entityID: task.Request.workflowID,
	})

	return nil
}

// // Tempolite is the main orchestration engine
// type Tempolite struct {
// 	mu             deadlock.RWMutex
// 	queueInstances map[string]*QueueInstance
// 	registry       *Registry
// 	database       Database
// 	ctx            context.Context
// 	cancel         context.CancelFunc
// 	defaultQueue   string
// }

// // QueueConfig holds configuration for a queue
// type QueueConfig struct {
// 	Name        string
// 	WorkerCount int
// }

// // TempoliteOption is a function type for configuring Tempolite
// type TempoliteOption func(*Tempolite) error

// func New(ctx context.Context, db Database, options ...TempoliteOption) (*Tempolite, error) {
// 	ctx, cancel := context.WithCancel(ctx)
// 	t := &Tempolite{
// 		registry:       NewRegistry(),
// 		database:       db,
// 		queueInstances: make(map[string]*QueueInstance),
// 		ctx:            ctx,
// 		cancel:         cancel,
// 		defaultQueue:   "default",
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

// 	// Create queue in database if it doesn't exist
// 	if _, err := t.database.GetQueueByName(config.Name); err != nil {
// 		if errors.Is(err, ErrQueueNotFound) {
// 			queue := &Queue{
// 				Name:      config.Name,
// 				CreatedAt: time.Now(),
// 				UpdatedAt: time.Now(),
// 			}
// 			if err := t.database.AddQueue(queue); err != nil {
// 				return fmt.Errorf("failed to create queue in database: %w", err)
// 			}
// 		} else {
// 			return fmt.Errorf("failed to check queue existence: %w", err)
// 		}
// 	}

// 	// Create queue instance
// 	queueCtx, queueCancel := context.WithCancel(t.ctx)
// 	instance := &QueueInstance{
// 		name:       config.Name,
// 		workers:    config.WorkerCount,
// 		registry:   t.registry,
// 		database:   t.database,
// 		ctx:        queueCtx,
// 		cancel:     queueCancel,
// 		processing: make(map[int]struct{}),
// 	}

// 	// Create worker pool
// 	workers := make([]retrypool.Worker[*WorkflowTask], config.WorkerCount)
// 	for i := 0; i < config.WorkerCount; i++ {
// 		workers[i] = NewQueueWorker(instance)
// 	}

// 	instance.pool = retrypool.New(queueCtx, workers,
// 		retrypool.WithAttempts[*WorkflowTask](1),
// 		retrypool.WithLogLevel[*WorkflowTask](logs.LevelDebug),
// 		retrypool.WithDelay[*WorkflowTask](time.Second),
// 	)

// 	t.queueInstances[config.Name] = instance
// 	go instance.Start()

// 	return nil
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
// 			instance.pool.AddWorker(worker)
// 		}
// 		instance.workers += delta
// 	} else if delta < 0 {
// 		// Remove workers
// 		count := -delta
// 		if count > instance.workers {
// 			return fmt.Errorf("cannot remove %d workers, only %d available", count, instance.workers)
// 		}
// 		workers := instance.pool.ListWorkers()
// 		for i := 0; i < count && i < len(workers); i++ {
// 			if err := instance.pool.RemoveWorker(workers[len(workers)-1-i].ID); err != nil {
// 				return fmt.Errorf("failed to remove worker: %w", err)
// 			}
// 		}
// 		instance.workers -= count
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
// 		instance.mu.Lock()
// 		instance.cancel()
// 		if err := instance.pool.Close(); err != nil {
// 			errs = append(errs, fmt.Errorf("failed to close queue %s: %w", instance.name, err))
// 		}
// 		instance.mu.Unlock()
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
// 			return instance.pool.WaitWithCallback(instance.ctx,
// 				func(queueSize, processingCount, deadTaskCount int) bool {
// 					return queueSize > 0 || processingCount > 0
// 				},
// 				time.Second,
// 			)
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
