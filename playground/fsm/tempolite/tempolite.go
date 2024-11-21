package tempolite

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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
			if !errors.Is(err, ErrQueueExists) {
				t.mu.Unlock()
				return nil, err
			}
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
	var err error

	_, err = t.database.GetQueueByName(config.Name)
	if errors.Is(err, ErrQueueNotFound) {
		// Create queue in database
		queue := &Queue{
			Name: config.Name,
		}
		if err = t.database.AddQueue(queue); err != nil {
			return fmt.Errorf("failed to create queue in database: %w", err)
		}
	}

	if _, exists := t.queues[config.Name]; exists {
		return ErrQueueExists
	}

	// Create queue manager
	manager := newQueueManager(t.ctx, config.Name, config.WorkerCount, t.registry, t.database)
	t.queues[config.Name] = manager

	go manager.Start() // starts the pulling

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
	return t.executeWorkflow(t.defaultQ, workflowFunc, options, args...)
}

// executeWorkflow executes a workflow on a specific queue
func (t *Tempolite) executeWorkflow(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *DatabaseFuture {

	// probably asked to execute on another queue
	if options != nil && options.Queue != "" {
		queueName = options.Queue
	}

	// Verify queue exists
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		f := NewDatabaseFuture(t.ctx, t.database)
		f.setError(fmt.Errorf("queue %s does not exist", queueName))
		return f
	}

	future := NewDatabaseFuture(t.ctx, t.database)

	id, err := qm.CreateWorkflow(workflowFunc, options, args...)
	if err != nil {
		future.setError(err)
		return future
	}

	// We don't need to manage execution on Tempolite layer since it has to go through a pulling

	future.setEntityID(id)

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
	var err error
	shutdown := errgroup.Group{}

	for _, qm := range t.queues {
		qm := qm
		shutdown.Go(func() error {
			fmt.Println("Waiting for queue", qm.name)
			defer fmt.Println("Queue", qm.name, "shutdown complete")
			// Ignore context.Canceled during wait
			if err := qm.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		})
	}

	if err = shutdown.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return errors.Join(err, fmt.Errorf("failed to wait for queues"))
	}
	fmt.Println("All queues shutdown complete")

	select {
	case <-t.ctx.Done():
		if !errors.Is(t.ctx.Err(), context.Canceled) {
			return t.ctx.Err()
		}
	default:
	}

	return nil
}

func (t *Tempolite) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Start graceful shutdown
	var errs []error
	for name, qm := range t.queues {
		if err := qm.Close(); err != nil {
			// Only append non-context-canceled errors
			if !errors.Is(err, context.Canceled) {
				errs = append(errs, fmt.Errorf("failed to close queue %s: %w", name, err))
			}
		}
	}

	// Set up grace period in background
	done := make(chan struct{})
	go func() {
		defer close(done)

		shutdown := errgroup.Group{}
		// wait for them to close
		for _, qm := range t.queues {
			qm := qm
			shutdown.Go(func() error {
				fmt.Println("Waiting for queue", qm.name)
				defer fmt.Println("Queue", qm.name, "shutdown complete")
				err := qm.Wait()
				if err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil
			})
		}

		if err := shutdown.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			errs = append(errs, fmt.Errorf("failed to wait for queues: %w", err))
		}
	}()

	// Wait for either grace period or clean shutdown
	select {
	case <-done:
		fmt.Println("Clean shutdown completed")
	case <-time.After(5 * time.Second):
		fmt.Println("Grace period expired, forcing shutdown")
		t.cancel()
	}

	// Only return error if we have non-context-canceled errors
	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
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
