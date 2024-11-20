package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/retrypool/logs"
)

/// TODO: after that we will see on tempolite which works differently, more data-based using the goroutine and triggering the function then

type queueRequestType string

var (
	queueRequestTypePause   queueRequestType = "pause"
	queueRequestTypeResume  queueRequestType = "resume"
	queueRequestTypeExecute queueRequestType = "execute"
)

type taskRequest struct {
	requestType queueRequestType
	entityID    int
}

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

	workerCounter int

	// which orchestrator is handling which entity ID
	cache           map[int]*QueueWorker
	entitiesWorkers map[int]int
	busy            map[int]struct{}
	free            map[int]struct{}

	requestPool *retrypool.Pool[*retrypool.RequestResponse[*taskRequest, struct{}]]
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
	ID           int // automatically set by the pool * magic *
	queueName    string
	orchestrator *Orchestrator
	ctx          context.Context
	mu           sync.Mutex
	onStartTask  func(*QueueWorker, *QueueTask)
	onEndTask    func(*QueueWorker, *QueueTask)
}

func (w *QueueWorker) Run(ctx context.Context, task *QueueTask) error {

	fmt.Println("\t workflow started on queue", w.queueName, task.entityID)
	if w.onStartTask != nil {
		w.onStartTask(w, task)
	}
	defer func() {
		if w.onEndTask != nil {
			w.onEndTask(w, task)
		}
	}()

	// Now execute the workflow using our existing orchestrator
	ftre, err := w.orchestrator.ExecuteWithEntity(task.entityID)
	if err != nil {
		if task.future != nil {
			task.future.setError(err)
		}
		return fmt.Errorf("failed to execute workflow task %d: %w", task.entityID, err)
	}

	if err := w.orchestrator.Wait(); err != nil {
		if task.future != nil {
			task.future.setError(err)
		}
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
		if task.future != nil {
			task.future.setError(err)
		}
		if errors.Is(err, context.Canceled) && errors.Is(err, ErrPaused) {
			return nil
		}
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
	var qm *QueueManager
	qm = &QueueManager{
		name:            name,
		database:        db,
		registry:        registry,
		ctx:             ctx,
		cancel:          cancel,
		workerCount:     workerCount,
		orchestrator:    NewOrchestrator(ctx, db, registry),
		cache:           make(map[int]*QueueWorker),
		free:            make(map[int]struct{}),
		busy:            make(map[int]struct{}),
		entitiesWorkers: make(map[int]int),
		requestPool: retrypool.New(
			ctx,
			[]retrypool.Worker[*retrypool.RequestResponse[*taskRequest, struct{}]]{},
			retrypool.WithAttempts[*retrypool.RequestResponse[*taskRequest, struct{}]](3),
			retrypool.WithDelay[*retrypool.RequestResponse[*taskRequest, struct{}]](time.Second/2),
			retrypool.WithOnNewDeadTask[*retrypool.RequestResponse[*taskRequest, struct{}]](func(task *retrypool.DeadTask[*retrypool.RequestResponse[*taskRequest, struct{}]], idx int) {
				errs := errors.New("failed to process request")
				for _, e := range task.Errors {
					errs = errors.Join(errs, e)
				}
				task.Data.CompleteWithError(errs)
				_, err := qm.requestPool.PullDeadTask(idx)
				if err != nil {
					// too bad
					log.Printf("failed to pull dead task: %v", err)
				}
			}),
		),
	}

	qm.pool = qm.createWorkerPool(workerCount)

	// TODO: how much?
	qm.requestPool.AddWorker(&queueWorkerRequests{
		qm: qm,
	})
	qm.requestPool.AddWorker(&queueWorkerRequests{
		qm: qm,
	})

	return qm
}

type queueWorkerRequests struct {
	qm *QueueManager
}

func (w *queueWorkerRequests) Run(ctx context.Context, task *retrypool.RequestResponse[*taskRequest, struct{}]) error {

	w.qm.mu.Lock()
	defer w.qm.mu.Unlock()
	fmt.Println("\t queueWorkerRequests", task.Request.requestType, task.Request.entityID)
	defer fmt.Println("\t queueWorkerRequests done", task.Request.requestType, task.Request.entityID)

	switch task.Request.requestType {
	case queueRequestTypePause:

		workerID, ok := w.qm.entitiesWorkers[task.Request.entityID]
		if !ok {
			return fmt.Errorf("entity %d not found", task.Request.entityID)
		}
		w.qm.cache[workerID].orchestrator.Pause()
		task.Complete(struct{}{})

	case queueRequestTypeResume:

		if w.qm.pool.AvailableWorkers() == 0 {
			return fmt.Errorf("no available workers for entity %d", task.Request.entityID)
		}

		var freeWorkers []int
		for workerID := range w.qm.free {
			freeWorkers = append(freeWorkers, workerID)
		}
		if len(freeWorkers) == 0 {
			return fmt.Errorf("no free workers available")
		}

		randomWorkerID := freeWorkers[rand.Intn(len(freeWorkers))]
		worker := w.qm.cache[randomWorkerID]
		worker.orchestrator.Resume(task.Request.entityID)
		task.Complete(struct{}{})

	case queueRequestTypeExecute:

		// Verify entity exists and is ready for execution
		entity := w.qm.database.GetEntity(task.Request.entityID)

		if entity == nil {
			task.CompleteWithError(fmt.Errorf("entity not found: %d", task.Request.entityID))
			return fmt.Errorf("entity not found: %d", task.Request.entityID)
		}

		if entity.Status != StatusPending {
			task.CompleteWithError(fmt.Errorf("entity %d is not in pending status", task.Request.entityID))
			return fmt.Errorf("entity %d is not in pending status", task.Request.entityID)
		}

		// Get handler info
		if entity.HandlerInfo == nil {
			task.CompleteWithError(fmt.Errorf("no handler info for entity %d", task.Request.entityID))
			return fmt.Errorf("no handler info for entity %d", task.Request.entityID)
		}

		// Convert stored inputs
		inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
		if err != nil {
			task.CompleteWithError(fmt.Errorf("failed to convert inputs: %w", err))
			return fmt.Errorf("failed to convert inputs: %w", err)
		}

		// Create task for execution
		queueTask := &QueueTask{
			workflowFunc: entity.HandlerInfo.Handler,
			options:      nil, // Could be stored in entity if needed
			args:         inputs,
			queueName:    w.qm.name,
			entityID:     task.Request.entityID,
		}

		processed := retrypool.NewProcessedNotification()

		// Submit to worker pool
		if err := w.qm.pool.Submit(queueTask, retrypool.WithBeingProcessed[*QueueTask](processed)); err != nil {
			task.CompleteWithError(fmt.Errorf("failed to submit entity %d to queue %s: %w", task.Request.entityID, w.qm.name, err))
			return fmt.Errorf("failed to submit entity %d to queue %s: %w", task.Request.entityID, w.qm.name, err)
		}

		<-processed

		task.Complete(struct{}{})

		return nil
	}
	return nil
}

func (qm *QueueManager) Pause(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		requestType: queueRequestTypePause,
		entityID:    id,
	})
	fmt.Println("\t PAUSE")
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	fmt.Println("\t PAUSE2")
	return task
}

func (qm *QueueManager) AvailableWorkers() int {
	return qm.pool.AvailableWorkers()
}

// Resume entity on available worker
func (qm *QueueManager) Resume(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		entityID:    id,
		requestType: queueRequestTypeResume,
	})
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	return task
}

func (qm *QueueManager) onTaskStart(worker *QueueWorker, task *QueueTask) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.entitiesWorkers[task.entityID] = worker.ID
	qm.cache[worker.ID] = worker
	qm.busy[worker.ID] = struct{}{}
	delete(qm.free, worker.ID)
	fmt.Println("\t onTaskStart", qm.entitiesWorkers, qm.busy, qm.free)
}

func (qm *QueueManager) onTaskEnd(worker *QueueWorker, task *QueueTask) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	delete(qm.entitiesWorkers, task.entityID)
	delete(qm.busy, worker.ID)
	qm.free[worker.ID] = struct{}{}
	fmt.Println("\t onTaskEnd", qm.entitiesWorkers, qm.busy, qm.free)
}

func (qm *QueueManager) GetEntities() []int {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	entities := make([]int, 0, len(qm.cache))
	for id := range qm.cache {
		entities = append(entities, id)
	}
	return entities
}

func (qm *QueueManager) createWorkerPool(count int) *retrypool.Pool[*QueueTask] {
	workers := make([]retrypool.Worker[*QueueTask], count)
	for i := 0; i < count; i++ {
		workers[i] = &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
			queueName:    qm.name,
			onStartTask:  qm.onTaskStart,
			onEndTask:    qm.onTaskEnd,
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
		// fmt.Println("Queue", qm.name, "has no pending tasks", info.PendingTasks, "pending tasks and", availableSlots, "available slots")
		return
	}

	// fmt.Println("Queue", qm.name, "has", info.PendingTasks, "pending tasks and", availableSlots, "available slots and", len(pendingEntities), "pending entities")

	for i := 0; i < min(availableSlots, len(pendingEntities)); i++ {
		entity := pendingEntities[i]

		// Recheck entity state from database
		freshEntity := qm.database.GetEntity(entity.ID)
		if freshEntity == nil || freshEntity.Status != StatusPending {
			continue
		}

		task := qm.ExecuteWorkflow(freshEntity.ID)

		// Wait for processing to start before continuing
		select {
		case <-task.Done():
			if task.Err() != nil {
				log.Printf("Failed to execute workflow %d on queue %s: %v",
					freshEntity.ID, qm.name, task.Err())
			}
			// Processing started
		case <-qm.ctx.Done():
			return
		}
	}
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

func (qm *QueueManager) ExecuteWorkflow(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		entityID:    id,
		requestType: queueRequestTypeExecute,
	})
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	return task
}

func (qm *QueueManager) handleTaskFailure(controller retrypool.WorkerController[*QueueTask], workerID int, worker retrypool.Worker[*QueueTask], task *retrypool.TaskWrapper[*QueueTask], err error) retrypool.DeadTaskAction {
	if errors.Is(err, ErrPaused) {
		return retrypool.DeadTaskActionDoNothing
	}
	log.Printf("Task failed in queue %s: %v", qm.name, err)
	return retrypool.DeadTaskActionDoNothing
}

func (qm *QueueManager) AddWorkers(count int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := 0; i < count; i++ {
		worker := &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
			queueName:    qm.name,
			onStartTask:  qm.onTaskStart,
			onEndTask:    qm.onTaskEnd,
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
		delete(qm.cache, workers[len(workers)-1-i].ID)
		delete(qm.busy, workers[len(workers)-1-i].ID)
		delete(qm.free, workers[len(workers)-1-i].ID)
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
