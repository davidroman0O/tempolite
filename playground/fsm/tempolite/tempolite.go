package tempolite

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Tempolite is the main API interface for the workflow engine
type Tempolite struct {
	mu       sync.RWMutex
	registry *Registry
	database Database
	queues   map[string]*QueueManager
	ctx      context.Context
	cancel   context.CancelFunc
	defaultQ string
}

// QueueConfig holds configuration for a queue
type QueueConfig struct {
	Name        string
	WorkerCount int
}

// TempoliteOption is a function type for configuring Tempolite
type TempoliteOption func(*Tempolite)

func New(ctx context.Context, db Database, options ...TempoliteOption) (*Tempolite, error) {
	ctx, cancel := context.WithCancel(ctx)
	t := &Tempolite{
		registry: NewRegistry(),
		database: db,
		queues:   make(map[string]*QueueManager),
		ctx:      ctx,
		cancel:   cancel,
		defaultQ: "default",
	}

	// Apply options BEFORE creating default queue
	for _, opt := range options {
		opt(t)
	}

	// Only create default queue if it doesn't exist
	t.mu.Lock()
	if _, exists := t.queues[t.defaultQ]; !exists {
		if err := t.createQueueLocked(QueueConfig{
			Name:        t.defaultQ,
			WorkerCount: 1,
		}); err != nil {
			t.mu.Unlock()
			return nil, err
		}
	}
	t.mu.Unlock()

	return t, nil
}

// createQueueLocked creates a queue while holding the lock
func (t *Tempolite) createQueueLocked(config QueueConfig) error {
	if config.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be greater than 0")
	}

	if _, exists := t.queues[config.Name]; exists {
		return fmt.Errorf("queue %s already exists", config.Name)
	}

	// Create queue in database
	queue := &Queue{
		Name: config.Name,
	}
	queue = t.database.AddQueue(queue)

	// Create queue manager
	manager := newQueueManager(t.ctx, config.Name, config.WorkerCount, t.registry, t.database)
	t.queues[config.Name] = manager

	return nil
}

func (t *Tempolite) createQueue(config QueueConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.createQueueLocked(config)
}

// RegisterWorkflow registers a workflow with the registry
func (t *Tempolite) RegisterWorkflow(workflowFunc interface{}) error {
	_, err := t.registry.RegisterWorkflow(workflowFunc)
	return err
}

// Workflow executes a workflow on the default queue
func (t *Tempolite) Workflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *DatabaseFuture {
	return t.ExecuteWorkflow(t.defaultQ, workflowFunc, options, args...)
}

// ExecuteWorkflow executes a workflow on a specific queue
func (t *Tempolite) ExecuteWorkflow(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *DatabaseFuture {
	// Verify queue exists
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		f := NewDatabaseFuture(t.ctx, t.database)
		f.setErr(fmt.Errorf("queue %s does not exist", queueName))
		return f
	}

	// Create orchestrator instance for entity creation
	orchestrator := NewOrchestrator(t.ctx, t.database, t.registry)

	// Prepare workflow entity
	entity, err := orchestrator.prepareWorkflowEntity(workflowFunc, options, args)
	if err != nil {
		f := NewDatabaseFuture(t.ctx, t.database)
		f.setErr(err)
		return f
	}

	// Create database-watching future
	future := NewDatabaseFuture(t.ctx, t.database)
	future.setEntityID(entity.ID)

	if err := qm.ExecuteDatabaseWorkflow(entity.ID); err != nil {
		future.setErr(err)
		return future
	}

	return future
}

func (t *Tempolite) AddQueueWorkers(queueName string, count int) error {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.AddWorkers(count)
}

func (t *Tempolite) RemoveQueueWorkers(queueName string, count int) error {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.RemoveWorkers(count)
}

func (t *Tempolite) GetQueueInfo(queueName string) (*QueueInfo, error) {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.GetInfo(), nil
}

func (t *Tempolite) Wait() error {
	shutdown := errgroup.Group{}

	for _, qm := range t.queues {
		qm := qm
		shutdown.Go(func() error {
			return qm.Wait()
		})
	}

	if err := shutdown.Wait(); err != nil {
		return errors.Join(err, fmt.Errorf("failed to wait for queues"))
	}

	<-t.ctx.Done()

	return nil
}

func (t *Tempolite) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()

	var errs []error
	for name, qm := range t.queues {
		if err := qm.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close queue %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing queues: %v", errs)
	}

	return nil
}

// GetQueues returns all available queue names
func (t *Tempolite) GetQueues() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	queues := make([]string, 0, len(t.queues))
	for name := range t.queues {
		queues = append(queues, name)
	}
	return queues
}

// WithQueue adds a new queue configuration
func WithQueue(config QueueConfig) TempoliteOption {
	return func(t *Tempolite) {
		if err := t.createQueue(config); err != nil {
			panic(fmt.Sprintf("failed to create queue %s: %v", config.Name, err))
		}
	}
}

// WithDefaultQueueWorkers sets the number of workers for the default queue
func WithDefaultQueueWorkers(count int) TempoliteOption {
	return func(t *Tempolite) {
		t.mu.Lock()
		defer t.mu.Unlock()

		// Remove existing default queue if it exists
		if existing, exists := t.queues[t.defaultQ]; exists {
			existing.Close()
			delete(t.queues, t.defaultQ)
		}

		// Create new default queue with specified worker count
		if err := t.createQueueLocked(QueueConfig{
			Name:        t.defaultQ,
			WorkerCount: count,
		}); err != nil {
			panic(fmt.Sprintf("failed to configure default queue: %v", err))
		}
	}
}
