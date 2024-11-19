package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/retrypool/logs"
)

// QueueManager manages a single queue and its workers
type QueueManager struct {
	name        string
	pool        *retrypool.Pool[*QueueTask]
	database    Database
	registry    *Registry
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	workerCount int

	orchestrator *Orchestrator
}

// QueueTask represents a workflow execution task
type QueueTask struct {
	workflowFunc interface{}      // The workflow function to execute
	options      *WorkflowOptions // Workflow execution options
	args         []interface{}    // Arguments for the workflow
	future       *RuntimeFuture   // Future for the task
	queueName    string           // Name of the queue this task belongs to
	entityID     int              // Add this field
}

// QueueInfo provides information about a queue's state
type QueueInfo struct {
	Name            string
	WorkerCount     int
	PendingTasks    int
	ProcessingTasks int
	FailedTasks     int
}

// QueueWorker represents a worker in a queue
type QueueWorker struct {
	queueName    string
	orchestrator *Orchestrator
	ctx          context.Context
}

func (w *QueueWorker) Run(ctx context.Context, task *QueueTask) error {

	fmt.Println("\t workflow started on queue", w.queueName)

	// Now execute the workflow using our existing orchestrator
	ftre, err := w.orchestrator.ExecuteWithEntity(task.entityID)
	if err != nil {
		return fmt.Errorf("failed to execute workflow task %d: %w", task.entityID, err)
	}

	if err := w.orchestrator.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if errors.Is(err, ErrPaused) {
			return nil
		}
		return err
	}

	fmt.Println("\t workflow waiting GET")
	// Wait for completion and propagate error/result
	if err := ftre.Get(); err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// if optional future is set, propagate results
	if task.future != nil {
		task.future.setResult(ftre.results)
	}

	fmt.Println("\t workflow done")

	return nil
}

func newQueueManager(ctx context.Context, name string, workerCount int, registry *Registry, db Database) *QueueManager {
	ctx, cancel := context.WithCancel(ctx)
	qm := &QueueManager{
		name:         name,
		database:     db,
		registry:     registry,
		ctx:          ctx,
		cancel:       cancel,
		workerCount:  workerCount,
		orchestrator: NewOrchestrator(ctx, db, registry),
	}

	qm.pool = qm.createWorkerPool(workerCount)

	return qm
}

func (qm *QueueManager) createWorkerPool(count int) *retrypool.Pool[*QueueTask] {
	workers := make([]retrypool.Worker[*QueueTask], count)
	for i := 0; i < count; i++ {
		workers[i] = &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
			queueName:    qm.name,
		}
	}
	return retrypool.New(qm.ctx, workers, []retrypool.Option[*QueueTask]{
		retrypool.WithAttempts[*QueueTask](1),
		retrypool.WithLogLevel[*QueueTask](logs.LevelDebug),
		retrypool.WithOnTaskFailure[*QueueTask](qm.handleTaskFailure),
	}...)
}

// Starts the automatic queue pulling process
func (qm *QueueManager) Start() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-ticker.C:
			qm.checkAndProcessPending()
		}
	}
}

func (qm *QueueManager) checkAndProcessPending() {
	// Get current queue state
	info := qm.GetInfo()
	availableSlots := info.WorkerCount - info.ProcessingTasks

	if availableSlots <= 0 {
		return
	}

	// Get queue first
	queue := qm.database.GetQueueByName(qm.name)
	if queue == nil {
		return
	}

	// Find pending workflows for this queue
	pendingEntities := qm.database.FindPendingWorkflowsByQueue(queue.ID)
	if len(pendingEntities) == 0 {
		return
	}

	for i := 0; i < min(availableSlots, len(pendingEntities)); i++ {
		entity := pendingEntities[i]

		// Recheck entity state from database
		freshEntity := qm.database.GetEntity(entity.ID)
		if freshEntity == nil || freshEntity.Status != StatusPending {
			continue
		}

		// Update status atomically
		freshEntity.Status = StatusQueued
		qm.database.UpdateEntity(freshEntity)

		// Only use BeingProcessed notification
		processed := retrypool.NewProcessedNotification()

		if err := qm.ExecuteDatabaseWorkflow(freshEntity.ID, processed); err != nil {
			log.Printf("Failed to execute workflow %d on queue %s: %v",
				freshEntity.ID, qm.name, err)
			// Reset status if execution failed
			freshEntity.Status = StatusPending
			qm.database.UpdateEntity(freshEntity)
			continue
		}

		// Wait for processing to start before continuing
		select {
		case <-processed:
			// Processing started
		case <-qm.ctx.Done():
			return
		}
	}
}

func (qm *QueueManager) ExecuteRuntimeWorkflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *RuntimeFuture {
	future := NewRuntimeFuture()

	// Handler registration check before submitting
	if _, err := qm.registry.RegisterWorkflow(workflowFunc); err != nil {
		future.setError(err)
		return future
	}

	// Create the workflow entity first
	entity, err := qm.orchestrator.prepareWorkflowEntity(workflowFunc, options, args...)
	if err != nil {
		future.setError(fmt.Errorf("failed to prepare workflow entity: %w", err))
		return future
	}

	task := &QueueTask{
		workflowFunc: workflowFunc,
		options:      options,
		args:         args,
		queueName:    qm.name,
		future:       future,
		entityID:     entity.ID, // Set the entity ID in the task
	}

	// Submit task to worker pool
	if err := qm.pool.Submit(task); err != nil {
		future.setError(fmt.Errorf("failed to submit task to queue %s: %w", qm.name, err))
		return future
	}

	return future
}

func (am *QueueManager) CreateWorkflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (int, error) {
	// Register handler before creating entity
	if _, err := am.registry.RegisterWorkflow(workflowFunc); err != nil {
		return 0, fmt.Errorf("failed to register workflow: %w", err)
	}

	// Prepare the workflow entity
	entity, err := am.orchestrator.prepareWorkflowEntity(workflowFunc, options, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare workflow entity: %w", err)
	}

	return entity.ID, nil
}

func (qm *QueueManager) ExecuteDatabaseWorkflow(id int, processed chan struct{}) error {
	// Verify entity exists and is ready for execution
	entity := qm.database.GetEntity(id)
	if entity == nil {
		return fmt.Errorf("entity not found: %d", id)
	}

	if entity.Status != StatusPending {
		return fmt.Errorf("entity %d is not in pending status", id)
	}

	// Get handler info
	if entity.HandlerInfo == nil {
		return fmt.Errorf("no handler info for entity %d", id)
	}

	// Convert stored inputs
	inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
	if err != nil {
		return fmt.Errorf("failed to convert inputs: %w", err)
	}

	// Create task for execution
	task := &QueueTask{
		workflowFunc: entity.HandlerInfo.Handler,
		options:      nil, // Could be stored in entity if needed
		args:         inputs,
		queueName:    qm.name,
		entityID:     id,
	}

	// Submit to worker pool
	if err := qm.pool.Submit(task, retrypool.WithBeingProcessed[*QueueTask](processed)); err != nil {
		return fmt.Errorf("failed to submit entity %d to queue %s: %w", id, qm.name, err)
	}

	return nil
}

func (qm *QueueManager) handleTaskFailure(controller retrypool.WorkerController[*QueueTask], workerID int, worker retrypool.Worker[*QueueTask], task *retrypool.TaskWrapper[*QueueTask], err error) retrypool.DeadTaskAction {
	log.Printf("Task failed in queue %s: %v", qm.name, err)
	return retrypool.DeadTaskActionAddToDeadTasks
}

func (qm *QueueManager) AddWorkers(count int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := 0; i < count; i++ {
		worker := &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
		}
		qm.pool.AddWorker(worker)
	}
	qm.workerCount += count

	return nil
}

func (qm *QueueManager) RemoveWorkers(count int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if count > qm.workerCount {
		return fmt.Errorf("cannot remove %d workers, only %d available", count, qm.workerCount)
	}

	workers := qm.pool.ListWorkers()
	for i := 0; i < count && i < len(workers); i++ {
		if err := qm.pool.RemoveWorker(workers[len(workers)-1-i].ID); err != nil {
			return fmt.Errorf("failed to remove worker: %w", err)
		}
	}
	qm.workerCount -= count

	return nil
}

func (qm *QueueManager) GetInfo() *QueueInfo {
	metrics := qm.pool.Metrics()
	return &QueueInfo{
		Name:            qm.name,
		WorkerCount:     qm.workerCount,
		PendingTasks:    qm.pool.QueueSize(),
		ProcessingTasks: qm.pool.ProcessingCount(),
		FailedTasks:     int(metrics.TasksFailed),
	}
}

func (qm *QueueManager) Close() error {
	go func() {
		<-time.After(5 * time.Second) // grace period
		qm.cancel()
	}()
	return qm.pool.Close()
}

func (qm *QueueManager) Wait() error {
	return qm.pool.WaitWithCallback(qm.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		log.Printf("Wait Queue %s - Workers: %d, Pending: %d, Processing: %d, Failed: %d",
			qm.name, qm.workerCount, queueSize, processingCount, deadTaskCount)
		return queueSize > 0 || processingCount > 0
	}, time.Second)
}
