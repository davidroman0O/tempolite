package tempolite

import (
	"context"
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
}

// QueueTask represents a workflow execution task
type QueueTask struct {
	workflowFunc interface{}
	options      *WorkflowOptions
	args         []interface{}
	future       *Future
	queueName    string
	registry     *Registry
	database     Database
}

// QueueInfo provides information about a queue's state
type QueueInfo struct {
	Name            string
	WorkerCount     int
	PendingTasks    int
	ProcessingTasks int
	FailedTasks     int
}

func newQueueManager(ctx context.Context, name string, workerCount int, registry *Registry, db Database) *QueueManager {
	ctx, cancel := context.WithCancel(ctx)
	qm := &QueueManager{
		name:        name,
		database:    db,
		registry:    registry,
		ctx:         ctx,
		cancel:      cancel,
		workerCount: workerCount,
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
		}
	}

	return retrypool.New(qm.ctx, workers, []retrypool.Option[*QueueTask]{
		retrypool.WithAttempts[*QueueTask](1),
		retrypool.WithLogLevel[*QueueTask](logs.LevelDebug),
		retrypool.WithOnTaskFailure[*QueueTask](qm.handleTaskFailure),
	}...)
}

func (qm *QueueManager) ExecuteWorkflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *Future {
	future := NewFuture(0)

	// Handler registration check before submitting
	if _, err := qm.registry.RegisterWorkflow(workflowFunc); err != nil {
		future.setError(err)
		return future
	}

	task := &QueueTask{
		workflowFunc: workflowFunc,
		options:      options,
		args:         args,
		future:       future,
		queueName:    qm.name,
		registry:     qm.registry,
		database:     qm.database,
	}

	if err := qm.pool.Submit(task); err != nil {
		future.setError(fmt.Errorf("failed to submit task to queue %s: %w", qm.name, err))
	}

	return future
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
	qm.cancel()
	return qm.pool.Close()
}

func (qm *QueueManager) Wait() error {
	return qm.pool.WaitWithCallback(qm.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		log.Printf("Wait Queue %s - Workers: %d, Pending: %d, Processing: %d, Failed: %d",
			qm.name, qm.workerCount, queueSize, processingCount, deadTaskCount)
		return queueSize > 0 || processingCount > 0
	}, time.Second)
}
