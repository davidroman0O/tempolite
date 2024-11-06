package tempolite

import (
	"time"

	"github.com/davidroman0O/retrypool"
)

// Base task properties shared across all task types
type BaseTempoliteTask struct {
	EntityID    string
	EntityType  string
	RunID       string
	StepID      string
	Status      retrypool.TaskStatus
	ExecutionID string
	HandlerName string
	QueueName   string
	MaxRetry    int
	RetryCount  int
	QueuedAt    []time.Time
	ProcessedAt []time.Time
	Durations   []time.Duration
	ScheduledAt time.Time
}

// Specific task type for workflows
type WorkflowTask struct {
	BaseTempoliteTask
	WorkflowID string
	IsPaused   bool
	Params     []interface{}
}

// Specific task type for activities
type ActivityTask struct {
	BaseTempoliteTask
	ActivityID string
	Params     []interface{}
}

// Specific task type for transactions and compensations
type TransactionTask struct {
	BaseTempoliteTask
	SagaID string
}

// Specific task type for transactions and compensations
type CompensationTask struct {
	BaseTempoliteTask
	SagaID string
}

type WorkerInfo struct {
	WorkerID        int
	Type            string
	ProcessingTasks []interface{} // []WorkflowTask or []ActivityTask or []TransactionTask or []CompensationTask or []SideEffect
	QueuedTasks     []interface{} // []WorkflowTask or []ActivityTask or []TransactionTask or []CompensationTask or []SideEffect
}

// Pool statistics and workers
type TempoliteRetryPool struct {
	QueueSize  int
	Processing int
	DeadTasks  int
	Workers    map[int]*WorkerInfo // Changed to map for direct worker lookup
}

// Queue information including all pool types
type TempoliteQueue struct {
	Name             string
	WorkflowPool     TempoliteRetryPool
	ActivityPool     TempoliteRetryPool
	TransactionPool  TempoliteRetryPool
	CompensationPool TempoliteRetryPool
	Status           QueueStatus
}

// Overall Tempolite system information
type TempoliteInfo struct {
	Queues []TempoliteQueue
}

// Implementation of the Info method
func (tp *Tempolite) Info() *TempoliteInfo {
	ti := &TempoliteInfo{
		Queues: []TempoliteQueue{},
	}

	tp.queues.Range(func(key, value interface{}) bool {
		queueName := key.(string)
		queue := value.(*QueueWorkers)

		value, ok := tp.queueStatus.Load(queueName)
		if !ok {
			tp.logger.Error(tp.ctx, "Error getting queue status", "queueName", queueName)
			return false
		}

		tq := TempoliteQueue{
			Name:   queueName,
			Status: QueueStatus(value.(QueueStatus)),
			WorkflowPool: TempoliteRetryPool{
				QueueSize:  queue.Workflows.QueueSize(),
				Processing: queue.Workflows.ProcessingCount(),
				DeadTasks:  queue.Workflows.DeadTaskCount(),
				Workers:    make(map[int]*WorkerInfo),
			},
			ActivityPool: TempoliteRetryPool{
				QueueSize:  queue.Activities.QueueSize(),
				Processing: queue.Activities.ProcessingCount(),
				DeadTasks:  queue.Activities.DeadTaskCount(),
				Workers:    make(map[int]*WorkerInfo),
			},
			TransactionPool: TempoliteRetryPool{
				QueueSize:  queue.Transactions.QueueSize(),
				Processing: queue.Transactions.ProcessingCount(),
				DeadTasks:  queue.Transactions.DeadTaskCount(),
				Workers:    make(map[int]*WorkerInfo),
			},
			CompensationPool: TempoliteRetryPool{
				QueueSize:  queue.Compensations.QueueSize(),
				Processing: queue.Compensations.ProcessingCount(),
				DeadTasks:  queue.Compensations.DeadTaskCount(),
				Workers:    make(map[int]*WorkerInfo),
			},
		}

		// Initialize worker maps for each pool
		for _, workerID := range queue.Workflows.GetWorkerIDs() {
			tq.WorkflowPool.Workers[workerID] = &WorkerInfo{
				WorkerID:        workerID,
				Type:            "workflow",
				ProcessingTasks: []interface{}{},
				QueuedTasks:     []interface{}{},
			}
		}
		for _, workerID := range queue.Activities.GetWorkerIDs() {
			tq.ActivityPool.Workers[workerID] = &WorkerInfo{
				WorkerID:        workerID,
				Type:            "activity",
				ProcessingTasks: []interface{}{},
				QueuedTasks:     []interface{}{},
			}
		}
		for _, workerID := range queue.Transactions.GetWorkerIDs() {
			tq.TransactionPool.Workers[workerID] = &WorkerInfo{
				WorkerID:        workerID,
				Type:            "transaction",
				ProcessingTasks: []interface{}{},
				QueuedTasks:     []interface{}{},
			}
		}
		for _, workerID := range queue.Compensations.GetWorkerIDs() {
			tq.CompensationPool.Workers[workerID] = &WorkerInfo{
				WorkerID:        workerID,
				Type:            "compensation",
				ProcessingTasks: []interface{}{},
				QueuedTasks:     []interface{}{},
			}
		}

		// Collect workflow tasks per worker
		queue.Workflows.RangeTasks(func(data *retrypool.TaskWrapper[*workflowTask], workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.WorkflowPool.Workers[workerID]; exists {

				switch status {
				case retrypool.TaskStatusProcessing:
					workerInfo.ProcessingTasks = append(workerInfo.ProcessingTasks, WorkflowTask{
						BaseTempoliteTask: BaseTempoliteTask{
							EntityID:    data.Data().ctx.EntityID(),
							EntityType:  data.Data().ctx.EntityType(),
							RunID:       data.Data().ctx.RunID(),
							StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().ctx.executionID,
							HandlerName: string(data.Data().handlerName),
							QueueName:   data.Data().queueName,
							MaxRetry:    data.Data().maxRetry,
							RetryCount:  data.Data().retryCount,
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						Params:     data.Data().params,
						WorkflowID: data.Data().ctx.workflowID,
						IsPaused:   data.Data().isPaused,
					})
				case retrypool.TaskStatusQueued:
					workerInfo.QueuedTasks = append(workerInfo.QueuedTasks, WorkflowTask{
						BaseTempoliteTask: BaseTempoliteTask{
							EntityID:    data.Data().ctx.EntityID(),
							EntityType:  data.Data().ctx.EntityType(),
							RunID:       data.Data().ctx.RunID(),
							StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().ctx.executionID,
							HandlerName: string(data.Data().handlerName),
							QueueName:   data.Data().queueName,
							MaxRetry:    data.Data().maxRetry,
							RetryCount:  data.Data().retryCount,
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						Params:     data.Data().params,
						WorkflowID: data.Data().ctx.workflowID,
						IsPaused:   data.Data().isPaused,
					})
				}

			}
			return true
		})

		// Collect activity tasks per worker
		queue.Activities.RangeTasks(func(data *retrypool.TaskWrapper[*activityTask], workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.ActivityPool.Workers[workerID]; exists {

				switch status {
				case retrypool.TaskStatusProcessing:
					workerInfo.ProcessingTasks = append(workerInfo.ProcessingTasks, ActivityTask{
						BaseTempoliteTask: BaseTempoliteTask{
							EntityID:    data.Data().ctx.EntityID(),
							EntityType:  data.Data().ctx.EntityType(),
							RunID:       data.Data().ctx.RunID(),
							StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().ctx.executionID,
							HandlerName: string(data.Data().handlerName),
							QueueName:   data.Data().queueName,
							MaxRetry:    data.Data().maxRetry,
							RetryCount:  data.Data().retryCount,
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						Params:     data.Data().params,
						ActivityID: data.Data().ctx.activityID,
					})
				case retrypool.TaskStatusQueued:
					workerInfo.QueuedTasks = append(workerInfo.QueuedTasks, ActivityTask{
						BaseTempoliteTask: BaseTempoliteTask{
							EntityID:    data.Data().ctx.EntityID(),
							EntityType:  data.Data().ctx.EntityType(),
							RunID:       data.Data().ctx.RunID(),
							StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().ctx.executionID,
							HandlerName: string(data.Data().handlerName),
							QueueName:   data.Data().queueName,
							MaxRetry:    data.Data().maxRetry,
							RetryCount:  data.Data().retryCount,
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						Params:     data.Data().params,
						ActivityID: data.Data().ctx.activityID,
					})
				}

			}
			return true
		})

		// Collect transaction tasks per worker
		queue.Transactions.RangeTasks(func(data *retrypool.TaskWrapper[*transactionTask], workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.TransactionPool.Workers[workerID]; exists {
				switch status {
				case retrypool.TaskStatusProcessing:
					workerInfo.ProcessingTasks = append(workerInfo.ProcessingTasks, TransactionTask{
						BaseTempoliteTask: BaseTempoliteTask{
							// EntityID:    data.Data().ctx.EntityID(),
							EntityType: data.Data().ctx.EntityType(),
							// RunID:       data.Data().ctx.RunID(),
							// StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().executionID,
							HandlerName: string(data.Data().handlerName),
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						SagaID: data.Data().sagaID,
					})
				case retrypool.TaskStatusQueued:
					workerInfo.QueuedTasks = append(workerInfo.QueuedTasks, TransactionTask{
						BaseTempoliteTask: BaseTempoliteTask{
							// EntityID:    data.Data().ctx.EntityID(),
							EntityType: data.Data().ctx.EntityType(),
							// RunID:       data.Data().ctx.RunID(),
							// StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().executionID,
							HandlerName: string(data.Data().handlerName),
							QueuedAt:    data.QueuedAt(),
							ProcessedAt: data.ProcessedAt(),
							Durations:   data.Durations(),
							ScheduledAt: data.ScheduledTime(),
						},
						SagaID: data.Data().sagaID,
					})
				}

			}
			return true
		})

		// Collect compensation tasks per worker
		queue.Compensations.RangeTasks(func(data *retrypool.TaskWrapper[*compensationTask], workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.CompensationPool.Workers[workerID]; exists {
				switch status {
				case retrypool.TaskStatusProcessing:
					workerInfo.ProcessingTasks = append(workerInfo.ProcessingTasks, TransactionTask{
						BaseTempoliteTask: BaseTempoliteTask{
							// EntityID:    data.Data().ctx.EntityID(),
							EntityType: data.Data().ctx.EntityType(),
							// RunID:       data.Data().ctx.RunID(),
							// StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().executionID,
							HandlerName: string(data.Data().handlerName),
						},
						SagaID: data.Data().sagaID,
					})
				case retrypool.TaskStatusQueued:
					workerInfo.QueuedTasks = append(workerInfo.QueuedTasks, TransactionTask{
						BaseTempoliteTask: BaseTempoliteTask{
							// EntityID:    data.Data().ctx.EntityID(),
							EntityType: data.Data().ctx.EntityType(),
							// RunID:       data.Data().ctx.RunID(),
							// StepID:      data.Data().ctx.StepID(),
							Status:      retrypool.TaskStatus(status),
							ExecutionID: data.Data().executionID,
							HandlerName: string(data.Data().handlerName),
						},
						SagaID: data.Data().sagaID,
					})
				}

			}
			return true
		})

		ti.Queues = append(ti.Queues, tq)
		return true
	})

	return ti
}

type WorkerPoolInfo struct {
	Workers    []int
	DeadTasks  int
	QueueSize  int
	Processing int
}

// Worker pool info functions
func (tp *Tempolite) GetWorkflowsInfo(queue string) (*WorkerPoolInfo, error) {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return nil, err
	}

	workers, err := tp.ListWorkersWorkflow(queue)
	if err != nil {
		return nil, err
	}

	return &WorkerPoolInfo{
		Workers:    workers,
		DeadTasks:  pool.DeadTaskCount(),
		QueueSize:  pool.QueueSize(),
		Processing: pool.ProcessingCount(),
	}, nil
}

func (tp *Tempolite) GetActivitiesInfo(queue string) (*WorkerPoolInfo, error) {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return nil, err
	}

	workers, err := tp.ListWorkersActivity(queue)
	if err != nil {
		return nil, err
	}

	return &WorkerPoolInfo{
		Workers:    workers,
		DeadTasks:  pool.DeadTaskCount(),
		QueueSize:  pool.QueueSize(),
		Processing: pool.ProcessingCount(),
	}, nil
}

func (tp *Tempolite) GetSideEffectsInfo(queue string) (*WorkerPoolInfo, error) {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return nil, err
	}

	workers, err := tp.ListWorkersSideEffect(queue)
	if err != nil {
		return nil, err
	}

	return &WorkerPoolInfo{
		Workers:    workers,
		DeadTasks:  pool.DeadTaskCount(),
		QueueSize:  pool.QueueSize(),
		Processing: pool.ProcessingCount(),
	}, nil
}

func (tp *Tempolite) GetTransactionsInfo(queue string) (*WorkerPoolInfo, error) {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return nil, err
	}

	workers, err := tp.ListWorkersTransaction(queue)
	if err != nil {
		return nil, err
	}

	return &WorkerPoolInfo{
		Workers:    workers,
		DeadTasks:  pool.DeadTaskCount(),
		QueueSize:  pool.QueueSize(),
		Processing: pool.ProcessingCount(),
	}, nil
}

func (tp *Tempolite) GetCompensationsInfo(queue string) (*WorkerPoolInfo, error) {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return nil, err
	}

	workers, err := tp.ListWorkersCompensation(queue)
	if err != nil {
		return nil, err
	}

	return &WorkerPoolInfo{
		Workers:    workers,
		DeadTasks:  pool.DeadTaskCount(),
		QueueSize:  pool.QueueSize(),
		Processing: pool.ProcessingCount(),
	}, nil
}
