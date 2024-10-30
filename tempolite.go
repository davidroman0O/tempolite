package tempolite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/workflow"

	dbSQL "database/sql"
)

/// TODO: change the registry functions
/// - register structs with directly a pointer of the struct, we should guarantee will destroy that pointer
/// - I don't want to see that AsXXX functions anymore
///
/// TODO: change enqueue functions, i don't want to see As functions anymore, you either put the function or an instance of the struct
///
/// In documentation warn that struct activities need a particular care since the developer might introduce even more non-deteministic code, it should be used for struct that hold a client/api but not a value what might change the output given the same inputs since it activities won't be replayed if sucessful.
///

var errWorkflowPaused = errors.New("workflow is paused")

type QueueStatus string

const (
	QueueStatusPending QueueStatus = "pending"
	QueueStatusRunning QueueStatus = "running"
	QueueStatusClosing QueueStatus = "closing"
)

// TODO: to be renamed as tempoliteEngine since it will be used to rotate/roll new databases when its database size is too big
type Tempolite struct {
	db     *dbSQL.DB
	client *ent.Client

	// Registry management
	workflows   sync.Map // map[HandlerIdentity]Workflow
	activities  sync.Map // map[HandlerIdentity]Activity
	sideEffects sync.Map // map[string]SideEffect
	sagas       sync.Map // map[string]*SagaDefinition

	queueStatus sync.Map // map[string]bool

	queuePoolWorkflowCounter     sync.Map // map[string]*atomic.Int64
	queuePoolActivityCounter     sync.Map
	queuePoolSideEffectCounter   sync.Map
	queuePoolTransactionCounter  sync.Map
	queuePoolCompensationCounter sync.Map

	// Additional queues
	queues sync.Map // map[string]*QueuePools

	ctx    context.Context
	cancel context.CancelFunc

	// Per-queue scheduler status
	queueSchedulerStatus sync.Map // map[string]*schedulerStatus

	// Scheduler done channels for default queue
	resumeWorkflowsWorkerDone chan struct{}

	// Per-queue done channels
	schedulerDone sync.Map // map[string]map[string]chan struct{} // queueName -> schedulerType -> done

	// Per-queue worker ID counters
	workerCounters sync.Map // map[string]map[string]*atomic.Int32 // queueName -> poolType -> counter

	// Versioning
	versionCache sync.Map

	logger Logger
}

func New(ctx context.Context, registry *Registry, opts ...tempoliteOption) (*Tempolite, error) {
	cfg := tempoliteConfig{
		workerConfig: WorkerConfig{
			InitialWorkflowsWorkers:    5,
			InitialActivityWorkers:     5,
			InitialSideEffectWorkers:   5,
			InitialTransctionWorkers:   5,
			InitialCompensationWorkers: 5,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger = NewDefaultLogger(cfg.defaultLogLevel)
	}

	ctx, cancel := context.WithCancel(ctx)

	optsComfy := []comfylite3.ComfyOption{}

	var firstTime bool
	if cfg.path != nil {
		cfg.logger.Debug(ctx, "Database got a path", "path", *cfg.path)
		cfg.logger.Debug(ctx, "Checking if first time or not", "path", *cfg.path)
		info, err := os.Stat(*cfg.path)
		if err == nil && !info.IsDir() {
			firstTime = false
		} else {
			firstTime = true
		}
		cfg.logger.Debug(ctx, "Fist timer check %v", firstTime)
		if cfg.destructive {
			cfg.logger.Debug(ctx, "Destructive option triggered", "firstTime", firstTime)
			if err := os.Remove(*cfg.path); err != nil {
				if !os.IsNotExist(err) {
					cfg.logger.Error(ctx, "Error removing file", "error", err)
					cancel()
					return nil, err
				}
			}
		}
		cfg.logger.Debug(ctx, "Creating directory recursively if necessary", "path", *cfg.path)
		if err := os.MkdirAll(filepath.Dir(*cfg.path), os.ModePerm); err != nil {
			cfg.logger.Error(ctx, "Error creating directory", "error", err)
			cancel()
			return nil, err
		}
		optsComfy = append(optsComfy, comfylite3.WithPath(*cfg.path))
	} else {
		cfg.logger.Debug(ctx, "Memory database option")
		optsComfy = append(optsComfy, comfylite3.WithMemory())
		firstTime = true
	}

	cfg.logger.Debug(ctx, "Opening/Creating database")
	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		cfg.logger.Error(ctx, "Error opening/creating database", "error", err)
		cancel()
		return nil, err
	}

	//	TODO: think how should I allow sql.DB option
	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)

	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	if firstTime || (cfg.destructive && cfg.path != nil) {
		cfg.logger.Debug(ctx, "Creating schema")
		if err = client.Schema.Create(ctx); err != nil {
			cfg.logger.Error(ctx, "Error creating schema", "error", err)
			cancel()
			return nil, err
		}
	}

	tp := &Tempolite{
		ctx:                       ctx,
		db:                        db,
		client:                    client,
		cancel:                    cancel,
		resumeWorkflowsWorkerDone: make(chan struct{}),
		logger:                    cfg.logger,
	}

	tp.logger.Debug(ctx, "Registering workflows functions", "total", len(registry.workflowsFunc))
	// Register components
	for _, workflow := range registry.workflowsFunc {
		if err := tp.registerWorkflow(workflow); err != nil {
			return nil, err
		}
	}

	tp.logger.Debug(ctx, "Registering activities", "total", len(registry.activities))
	for _, activity := range registry.activities {
		if err := tp.registerActivity(activity); err != nil {
			tp.logger.Error(ctx, "Error registering activity", "error", err)
			return nil, err
		}
	}

	tp.logger.Debug(ctx, "Registering default queue")
	// priority
	if err := tp.createQueue("default", cfg.workerConfig.InitialWorkflowsWorkers, cfg.workerConfig.InitialActivityWorkers, cfg.workerConfig.InitialSideEffectWorkers, cfg.workerConfig.InitialTransctionWorkers, cfg.workerConfig.InitialCompensationWorkers); err != nil {
		tp.logger.Error(ctx, "Error creating default queue", "error", err)
		return nil, err
	}

	for _, queue := range cfg.queues {
		tp.logger.Debug(ctx, "Registering additional queue", "name", queue.Name)
		if err := tp.createQueue(queue.Name, queue.WorkflowWorkers, queue.ActivityWorkers, queue.SideEffectWorkers, queue.TransactionWorkers, queue.CompensationWorkers); err != nil {
			tp.logger.Error(ctx, "Error creating additional queue", "error", err)
			return nil, err
		}
	}

	tp.logger.Debug(ctx, "Starting resume workflows worker")
	// go tp.resumeWorkflowsWorker()

	return tp, nil
}

type QueueWorkers struct {
	Workflows     *retrypool.Pool[*workflowTask]
	Activities    *retrypool.Pool[*activityTask]
	SideEffects   *retrypool.Pool[*sideEffectTask]
	Transactions  *retrypool.Pool[*transactionTask]
	Compensations *retrypool.Pool[*compensationTask]
	Done          chan struct{}
}

func (tp *Tempolite) createQueue(name string, workflowWorkers, activityWorkers,
	sideEffectWorkers, transactionWorkers, compensationWorkers int) error {

	// Validate queue name
	if name == "" {
		tp.logger.Error(tp.ctx, "Queue name cannot be empty")
		return fmt.Errorf("queue name cannot be empty")
	}

	tp.queueStatus.Store(name, QueueStatusPending)

	tp.logger.Debug(tp.ctx, "Creating queue", "name", name)

	tp.logger.Debug(tp.ctx, "Creating workflow pool", "name", name, "workers", workflowWorkers)
	// Initialize pools
	workflowPool, err := tp.createWorkflowPool(name, workflowWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating workflow pool", "error", err)
		tp.queueStatus.Delete(name)
		return fmt.Errorf("failed to create workflow pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating activity pool", "name", name, "workers", activityWorkers)
	activityPool, err := tp.createActivityPool(name, activityWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating activity pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up the workflow pool before returning
		workflowPool.Close()
		return fmt.Errorf("failed to create activity pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating side effect pool", "name", name, "workers", sideEffectWorkers)
	sideEffectPool, err := tp.createSideEffectPool(name, sideEffectWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating side effect pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		workflowPool.Close()
		activityPool.Close()
		return fmt.Errorf("failed to create side effect pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction pool", "name", name, "workers", transactionWorkers)
	transactionPool, err := tp.createTransactionPool(name, transactionWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		workflowPool.Close()
		activityPool.Close()
		sideEffectPool.Close()
		return fmt.Errorf("failed to create transaction pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating compensation pool", "name", name, "workers", compensationWorkers)
	compensationPool, err := tp.createCompensationPool(name, compensationWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating compensation pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		workflowPool.Close()
		activityPool.Close()
		sideEffectPool.Close()
		transactionPool.Close()
		return fmt.Errorf("failed to create compensation pool: %w", err)
	}

	done := make(chan struct{})

	workers := &QueueWorkers{
		Workflows:     workflowPool,
		Activities:    activityPool,
		SideEffects:   sideEffectPool,
		Transactions:  transactionPool,
		Compensations: compensationPool,
		Done:          done,
	}

	tp.logger.Debug(tp.ctx, "Storing queue workers", "name", name)
	// Store the queue workers
	if _, loaded := tp.queues.LoadOrStore(name, workers); loaded {
		tp.logger.Error(tp.ctx, "Queue already exists", "name", name)
		tp.queueStatus.Delete(name)
		// Clean up all pools if queue already exists
		workflowPool.Close()
		activityPool.Close()
		sideEffectPool.Close()
		transactionPool.Close()
		compensationPool.Close()
		return fmt.Errorf("queue %s already exists", name)
	}

	tp.logger.Debug(tp.ctx, "Storing queue counters", "name", name)
	// Initialize queue counters
	tp.initQueueCounters(name)

	if tp.resumeRunningWorkflows(name) != nil {
		tp.logger.Error(tp.ctx, "Error resuming running workflows")
		return fmt.Errorf("failed to resume running workflows")
	}

	tp.logger.Debug(tp.ctx, "Starting queue-specific schedulers")
	// Start queue-specific schedulers
	go tp.schedulerExecutionWorkflowForQueue(name, done)
	go tp.schedulerExecutionActivityForQueue(name, done)
	go tp.schedulerExecutionSideEffectForQueue(name, done)
	go tp.schedulerExecutionSagaForQueue(name, done)
	go tp.schedulerResumeRunningWorkflows(name, done)

	tp.queueStatus.Store(name, QueueStatusRunning)

	tp.logger.Debug(tp.ctx, "Created queue", "name", name,
		"workflowWorkers", workflowWorkers,
		"activityWorkers", activityWorkers,
		"sideEffectWorkers", sideEffectWorkers,
		"transactionWorkers", transactionWorkers,
		"compensationWorkers", compensationWorkers)

	return nil
}

func (tp *Tempolite) removeQueue(name string) error {
	value, ok := tp.queues.LoadAndDelete(name)
	if !ok {
		return fmt.Errorf("queue %s not found", name)
	}

	tp.queueStatus.Store(name, QueueStatusClosing)

	workers := value.(*QueueWorkers)
	close(workers.Done) // Signal schedulers to stop

	// Close all pools
	workers.Workflows.Close()
	workers.Activities.Close()
	workers.SideEffects.Close()
	workers.Transactions.Close()
	workers.Compensations.Close()

	// Clean up queue counters
	tp.queuePoolWorkflowCounter.Delete(name)
	tp.queuePoolActivityCounter.Delete(name)
	tp.queuePoolSideEffectCounter.Delete(name)
	tp.queuePoolTransactionCounter.Delete(name)
	tp.queuePoolCompensationCounter.Delete(name)

	tp.logger.Debug(tp.ctx, "Removed queue", "name", name)
	tp.queueStatus.Delete(name)
	return nil
}

func (tp *Tempolite) getWorkflowPoolQueue(queueName string) (*retrypool.Pool[*workflowTask], error) {
	if queueName == "" {
		queueName = "default"
	}
	if value, ok := tp.queues.Load(queueName); ok {
		queue := value.(*QueueWorkers)
		return queue.Workflows, nil
	}
	return nil, fmt.Errorf("workflow queue %s not found", queueName)
}

func (tp *Tempolite) getActivityPoolQueue(queueName string) (*retrypool.Pool[*activityTask], error) {
	if queueName == "" {
		queueName = "default"
	}
	if value, ok := tp.queues.Load(queueName); ok {
		queue := value.(*QueueWorkers)
		return queue.Activities, nil
	}
	return nil, fmt.Errorf("activity queue %s not found", queueName)
}

func (tp *Tempolite) getSideEffectPoolQueue(queueName string) (*retrypool.Pool[*sideEffectTask], error) {
	if queueName == "" {
		queueName = "default"
	}
	if value, ok := tp.queues.Load(queueName); ok {
		queue := value.(*QueueWorkers)
		return queue.SideEffects, nil
	}
	return nil, fmt.Errorf("side effect queue %s not found", queueName)
}

func (tp *Tempolite) getTransactionPoolQueue(queueName string) (*retrypool.Pool[*transactionTask], error) {
	if queueName == "" {
		queueName = "default"
	}
	if value, ok := tp.queues.Load(queueName); ok {
		queue := value.(*QueueWorkers)
		return queue.Transactions, nil
	}
	return nil, fmt.Errorf("transaction queue %s not found", queueName)
}

func (tp *Tempolite) getCompensationPoolQueue(queueName string) (*retrypool.Pool[*compensationTask], error) {
	if queueName == "" {
		queueName = "default"
	}
	if value, ok := tp.queues.Load(queueName); ok {
		queue := value.(*QueueWorkers)
		return queue.Compensations, nil
	}
	return nil, fmt.Errorf("compensation queue %s not found", queueName)
}

type closeConfig struct {
	duration time.Duration
}

type closeOption func(*closeConfig)

func WithCloseDuration(duration time.Duration) closeOption {
	return func(cfg *closeConfig) {
		cfg.duration = duration
	}
}

func (tp *Tempolite) Close(opts ...closeOption) error {
	tp.logger.Debug(tp.ctx, "Starting Tempolite shutdown sequence")

	cfg := closeConfig{
		duration: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// First cancel the context to stop new operations
	tp.cancel()

	// Now close all pools in each queue
	tp.queues.Range(func(key, value interface{}) bool {
		queueName := key.(string)
		tp.logger.Debug(tp.ctx, "Closing queue pools", "queueName", queueName)
		tp.removeQueue(queueName)
		return true
	})

	// Short wait to allow schedulers to stop
	<-time.After(100 * time.Millisecond)

	// Create a channel to coordinate shutdown
	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)

		// First pause all active workflows using the existing PauseWorkflow function
		workflows, err := tp.client.Workflow.Query().
			Where(
				workflow.StatusEQ(workflow.StatusRunning),
				workflow.IsPausedEQ(false),
			).
			All(tp.ctx)

		if err == nil {
			for _, wf := range workflows {
				if err := tp.PauseWorkflow(WorkflowID(wf.ID)); err != nil {
					tp.logger.Error(tp.ctx, "Error pausing workflow during shutdown",
						"workflowID", wf.ID, "error", err)
				}
			}
		}

	}()

	// Wait for shutdown with timeout
	select {
	case <-shutdownDone:
		tp.logger.Debug(tp.ctx, "Tempolite shutdown completed successfully")
		return nil
	case <-time.After(cfg.duration):
		tp.logger.Error(tp.ctx, "Tempolite shutdown timed out")
		return fmt.Errorf("shutdown timed out")
	}
}

func (tp *Tempolite) preparePauseClosingWorkflow(ctx context.Context) error {

	// Get all active workflows
	activeWorkflows, err := tp.client.Workflow.Query().
		Where(
			workflow.StatusEQ(workflow.StatusRunning),
			workflow.IsPausedEQ(false),
			workflow.IsReadyEQ(false),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query active workflows: %w", err)
	}

	// Pause each workflow
	for _, wf := range activeWorkflows {
		if err := tp.PauseWorkflow(WorkflowID(wf.ID)); err != nil {
			tp.logger.Error(ctx, "Failed to pause workflow during shutdown", "workflowID", wf.ID, "error", err)
			continue
		}
	}

	return nil
}

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

// Worker information including all tasks assigned to it
type WorkerInfo struct {
	WorkerID      int
	Workflows     []WorkflowTask
	Activities    []ActivityTask
	Transactions  []TransactionTask
	Compensations []TransactionTask
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
				WorkerID:      workerID,
				Workflows:     []WorkflowTask{},
				Activities:    []ActivityTask{},
				Transactions:  []TransactionTask{},
				Compensations: []TransactionTask{},
			}
		}
		for _, workerID := range queue.Activities.GetWorkerIDs() {
			tq.ActivityPool.Workers[workerID] = &WorkerInfo{
				WorkerID:      workerID,
				Workflows:     []WorkflowTask{},
				Activities:    []ActivityTask{},
				Transactions:  []TransactionTask{},
				Compensations: []TransactionTask{},
			}
		}
		for _, workerID := range queue.Transactions.GetWorkerIDs() {
			tq.TransactionPool.Workers[workerID] = &WorkerInfo{
				WorkerID:      workerID,
				Workflows:     []WorkflowTask{},
				Activities:    []ActivityTask{},
				Transactions:  []TransactionTask{},
				Compensations: []TransactionTask{},
			}
		}
		for _, workerID := range queue.Compensations.GetWorkerIDs() {
			tq.CompensationPool.Workers[workerID] = &WorkerInfo{
				WorkerID:      workerID,
				Workflows:     []WorkflowTask{},
				Activities:    []ActivityTask{},
				Transactions:  []TransactionTask{},
				Compensations: []TransactionTask{},
			}
		}

		// Collect workflow tasks per worker
		queue.Workflows.RangeTasks(func(data *workflowTask, workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.WorkflowPool.Workers[workerID]; exists {
				workerInfo.Workflows = append(workerInfo.Workflows, WorkflowTask{
					BaseTempoliteTask: BaseTempoliteTask{
						EntityID:    data.ctx.EntityID(),
						EntityType:  data.ctx.EntityType(),
						RunID:       data.ctx.RunID(),
						StepID:      data.ctx.StepID(),
						Status:      retrypool.TaskStatus(status),
						ExecutionID: data.ctx.executionID,
						HandlerName: string(data.handlerName),
						QueueName:   data.queueName,
						MaxRetry:    data.maxRetry,
						RetryCount:  data.retryCount,
					},
					Params:     data.params,
					WorkflowID: data.ctx.workflowID,
					IsPaused:   data.isPaused,
				})
			}
			return true
		})

		// Collect activity tasks per worker
		queue.Activities.RangeTasks(func(data *activityTask, workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.ActivityPool.Workers[workerID]; exists {
				workerInfo.Activities = append(workerInfo.Activities, ActivityTask{
					BaseTempoliteTask: BaseTempoliteTask{
						EntityID:    data.ctx.EntityID(),
						EntityType:  data.ctx.EntityType(),
						RunID:       data.ctx.RunID(),
						StepID:      data.ctx.StepID(),
						Status:      retrypool.TaskStatus(status),
						ExecutionID: data.ctx.executionID,
						HandlerName: string(data.handlerName),
						QueueName:   data.queueName,
						MaxRetry:    data.maxRetry,
						RetryCount:  data.retryCount,
					},
					Params:     data.params,
					ActivityID: data.ctx.activityID,
				})
			}
			return true
		})

		// Collect transaction tasks per worker
		queue.Transactions.RangeTasks(func(data *transactionTask, workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.TransactionPool.Workers[workerID]; exists {
				workerInfo.Transactions = append(workerInfo.Transactions, TransactionTask{
					BaseTempoliteTask: BaseTempoliteTask{
						// EntityID:    data.ctx.EntityID(),
						EntityType: data.ctx.EntityType(),
						// RunID:       data.ctx.RunID(),
						// StepID:      data.ctx.StepID(),
						Status:      retrypool.TaskStatus(status),
						ExecutionID: data.executionID,
						HandlerName: string(data.handlerName),
					},
					SagaID: data.sagaID,
				})
			}
			return true
		})

		// Collect compensation tasks per worker
		queue.Compensations.RangeTasks(func(data *compensationTask, workerID int, status retrypool.TaskStatus) bool {
			if workerInfo, exists := tq.CompensationPool.Workers[workerID]; exists {
				workerInfo.Compensations = append(workerInfo.Compensations, TransactionTask{
					BaseTempoliteTask: BaseTempoliteTask{
						// EntityID:    data.ctx.EntityID(),
						EntityType: data.ctx.EntityType(),
						// RunID:       data.ctx.RunID(),
						// StepID:      data.ctx.StepID(),
						Status:      retrypool.TaskStatus(status),
						ExecutionID: data.executionID,
						HandlerName: string(data.handlerName),
					},
					SagaID: data.sagaID,
				})
			}
			return true
		})

		ti.Queues = append(ti.Queues, tq)
		return true
	})

	return ti
}
func (tp *Tempolite) getWorkerWorkflowID(queue string) (int, error) {
	counter, _ := tp.queuePoolWorkflowCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerActivityID(queue string) (int, error) {
	counter, _ := tp.queuePoolActivityCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerSideEffectID(queue string) (int, error) {
	counter, _ := tp.queuePoolSideEffectCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerTransactionID(queue string) (int, error) {
	counter, _ := tp.queuePoolTransactionCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) getWorkerCompensationID(queue string) (int, error) {
	counter, _ := tp.queuePoolCompensationCounter.LoadOrStore(queue, &atomic.Int64{})
	return int(counter.(*atomic.Int64).Add(1)), nil
}

func (tp *Tempolite) initQueueCounters(queue string) {
	tp.queuePoolWorkflowCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolActivityCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolSideEffectCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolTransactionCounter.Store(queue, &atomic.Int64{})
	tp.queuePoolCompensationCounter.Store(queue, &atomic.Int64{})
}

type WorkerPoolInfo struct {
	Workers    []int
	DeadTasks  int
	QueueSize  int
	Processing int
}

func (tp *Tempolite) AddWorkerWorkflow(queue string) error {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerWorkflowID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(workflowWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerActivity(queue string) error {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return err
	}

	id, err := tp.getWorkerActivityID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(activityWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerSideEffect(queue string) error {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerSideEffectID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(sideEffectWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerTransaction(queue string) error {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerTransactionID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(transactionWorker{id: id, tp: tp})
	return nil
}

func (tp *Tempolite) AddWorkerCompensation(queue string) error {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return err
	}
	id, err := tp.getWorkerCompensationID(queue)
	if err != nil {
		return err
	}
	pool.AddWorker(compensationWorker{id: id, tp: tp})
	return nil
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

func (tp *Tempolite) RemoveWorkerWorkflowByID(queue string, id int) error {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return err
	}
	pool.ForceClose()
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerActivityByID(queue string, id int) error {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerSideEffectByID(queue string, id int) error {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerTransactionByID(queue string, id int) error {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) RemoveWorkerCompensationByID(queue string, id int) error {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return err
	}
	return pool.RemoveWorker(id)
}

func (tp *Tempolite) ListWorkersWorkflow(queue string) ([]int, error) {
	pool, err := tp.getWorkflowPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersActivity(queue string) ([]int, error) {
	pool, err := tp.getActivityPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersSideEffect(queue string) ([]int, error) {
	pool, err := tp.getSideEffectPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersTransaction(queue string) ([]int, error) {
	pool, err := tp.getTransactionPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListWorkersCompensation(queue string) ([]int, error) {
	pool, err := tp.getCompensationPoolQueue(queue)
	if err != nil {
		return nil, err
	}
	return pool.GetWorkerIDs(), nil
}

func (tp *Tempolite) ListQueues() []string {
	queues := []string{}
	tp.queues.Range(func(key, _ interface{}) bool {
		queues = append(queues, key.(string))
		return true
	})
	return queues
}

type waitItem struct {
	QueueName string
	Workers   *QueueWorkers
}

func (tp *Tempolite) Wait() error {
	<-time.After(1 * time.Second)

	// make channe of array of waitItem
	newWaiters := make(chan []waitItem)

	go func() {

		ticker := time.NewTicker(time.Second)
		for {
			select {
			case _, ok := <-newWaiters:
				if !ok {
					ticker.Stop()
					// closed
					return
				}
			case <-ticker.C:
				waiters := []waitItem{}
				tp.queues.Range(func(key, value interface{}) bool {
					queueName := key.(string)
					queue := value.(*QueueWorkers)
					waiters = append(waiters, waitItem{QueueName: queueName, Workers: queue})
					return true
				})
				newWaiters <- waiters
			}
		}
	}()

	go func() {
		finished := false
		for !finished {
			<-time.After(1 * time.Second)
			select {
			case waiters := <-newWaiters:
				allZero := true
				states := make([]bool, len(waiters))
				for i, waiter := range waiters {
					queueName := waiter.QueueName
					queue := waiter.Workers
					tp.logger.Debug(tp.ctx, "Waiting queue", "queueName", queueName)

					checkPool := func(q string) func(queueSize, processingCount, deadTaskCount int) bool {
						return func(queueSize, processingCount, deadTaskCount int) bool {
							tp.logger.Debug(tp.ctx, fmt.Sprintf("Checking %s pool on queue %s", q, queueName), "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
							if queueSize > 0 || processingCount > 0 || deadTaskCount > 0 {
								states[i] = false
							} else {
								states[i] = true
							}
							return false // no wait, we just consult
						}
					}

					queue.Workflows.WaitWithCallback(tp.ctx, checkPool("workfow"), 1)
					queue.Activities.WaitWithCallback(tp.ctx, checkPool("activities"), 1)
					queue.SideEffects.WaitWithCallback(tp.ctx, checkPool("side effects"), 1)
					queue.Transactions.WaitWithCallback(tp.ctx, checkPool("transactions"), 1)
					queue.Compensations.WaitWithCallback(tp.ctx, checkPool("compensations"), 1)
				}
				for _, state := range states {
					if !state {
						allZero = false
						break
					}
				}
				if allZero {
					finished = true
					close(newWaiters)
				}
			}
		}
	}()

	return nil
}

// func (tp *Tempolite) convertInputs(handlerInfo HandlerInfo, executionInputs []interface{}) ([]interface{}, error) {
// 	tp.logger.Debug(tp.ctx, "Converting inputs", "handlerInfo", handlerInfo)
// 	outputs := []interface{}{}
// 	// TODO: we can probably parallelize this
// 	for idx, rawInputs := range executionInputs {
// 		inputType := handlerInfo.ParamTypes[idx]
// 		inputKind := handlerInfo.ParamsKinds[idx]
// 		tp.logger.Debug(tp.ctx, "Converting input", "inputType", inputType, "inputKind", inputKind, "rawInputs", rawInputs)
// 		realInput, err := convertIO(rawInputs, inputType, inputKind)
// 		if err != nil {
// 			tp.logger.Error(tp.ctx, "Error converting input", "error", err)
// 			return nil, err
// 		}
// 		outputs = append(outputs, realInput)
// 	}
// 	return outputs, nil
// }

// func (tp *Tempolite) convertOuputs(handlerInfo HandlerInfo, executionOutput []interface{}) ([]interface{}, error) {
// 	tp.logger.Debug(tp.ctx, "Converting outputs", "handlerInfo", handlerInfo)
// 	outputs := []interface{}{}
// 	// TODO: we can probably parallelize this
// 	for idx, rawOutput := range executionOutput {
// 		ouputType := handlerInfo.ReturnTypes[idx]
// 		outputKind := handlerInfo.ReturnKinds[idx]
// 		tp.logger.Debug(tp.ctx, "Converting output", "ouputType", ouputType, "outputKind", outputKind, "rawOutput", rawOutput)
// 		realOutput, err := convertIO(rawOutput, ouputType, outputKind)
// 		if err != nil {
// 			tp.logger.Error(tp.ctx, "Error converting output", "error", err)
// 			return nil, err
// 		}
// 		outputs = append(outputs, realOutput)
// 	}
// 	return outputs, nil
// }

func (tp *Tempolite) verifyHandlerAndParams(handlerInfo HandlerInfo, params []interface{}) error {

	if len(params) != handlerInfo.NumIn {
		tp.logger.Error(tp.ctx, "Parameter count mismatch", "handlerName", handlerInfo.HandlerLongName, "expected", handlerInfo.NumIn, "got", len(params))
		return fmt.Errorf("parameter count mismatch (you probably put the wrong handler): expected %d, got %d", handlerInfo.NumIn, len(params))
	}

	for idx, param := range params {
		if reflect.TypeOf(param) != handlerInfo.ParamTypes[idx] {
			tp.logger.Error(tp.ctx, "Parameter type mismatch", "handlerName", handlerInfo.HandlerLongName, "expected", handlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
			return fmt.Errorf("parameter type mismatch (you probably put the wrong handler) at index %d: expected %s, got %s", idx, handlerInfo.ParamTypes[idx], reflect.TypeOf(param))
		}
	}

	return nil
}

func (tp *Tempolite) GetLatestWorkflowExecution(originalWorkflowID WorkflowID) (WorkflowID, error) {
	tp.logger.Debug(tp.ctx, "Getting latest workflow execution", "originalWorkflowID", originalWorkflowID)

	currentID := originalWorkflowID
	for {
		w, err := tp.client.Workflow.Query().
			Where(workflow.ID(string(currentID))).
			Only(tp.ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				tp.logger.Error(tp.ctx, "Workflow not found", "workflowID", currentID)
				return "", fmt.Errorf("workflow not found: %w", err)
			}
			tp.logger.Error(tp.ctx, "Error querying workflow", "error", err)
			return "", fmt.Errorf("error querying workflow: %w", err)
		}

		// Check if this workflow has been continued
		continuedWorkflow, err := tp.client.Workflow.Query().
			Where(workflow.ContinuedFromID(w.ID)).
			Only(tp.ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				// This is the latest workflow in the chain
				tp.logger.Debug(tp.ctx, "Found latest workflow execution", "latestWorkflowID", currentID)
				return currentID, nil
			}
			tp.logger.Error(tp.ctx, "Error querying continued workflow", "error", err)
			return "", fmt.Errorf("error querying continued workflow: %w", err)
		}

		// Move to the next workflow in the chain
		currentID = WorkflowID(continuedWorkflow.ID)
	}
}

func (tp *Tempolite) ListPausedWorkflows() ([]WorkflowID, error) {
	pausedWorkflows, err := tp.client.Workflow.Query().
		Where(workflow.IsPausedEQ(true)).
		All(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error listing paused workflows", "error", err)
		return nil, err
	}

	ids := make([]WorkflowID, len(pausedWorkflows))
	for i, wf := range pausedWorkflows {
		ids[i] = WorkflowID(wf.ID)
	}

	return ids, nil
}

func (tp *Tempolite) PauseCloseWorkflow(id WorkflowID) error {
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return err
	}
	_, err = tx.Workflow.UpdateOneID(id.String()).
		SetIsPaused(true).
		SetIsReady(false).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error pausing workflow", "workflowID", id, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
		return err
	}

	// Wait for workflow to reach paused status
	return tp.waitWorkflowStatus(tp.ctx, id, workflow.StatusPaused)
}

func (tp *Tempolite) PauseWorkflow(id WorkflowID) error {
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return err
	}
	_, err = tx.Workflow.UpdateOneID(id.String()).
		SetIsPaused(true).
		SetIsReady(false).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error pausing workflow", "workflowID", id, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
		return err
	}

	// Wait for workflow to reach paused status
	return tp.waitWorkflowStatus(tp.ctx, id, workflow.StatusPaused)
}

func (tp *Tempolite) ResumeWorkflow(id WorkflowID) error {
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return err
	}
	_, err = tx.Workflow.UpdateOneID(id.String()).
		SetIsPaused(false).
		SetIsReady(true).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error resuming workflow", "workflowID", id, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
		return err
	}

	// Wait for workflow to reach running status
	return tp.waitWorkflowStatus(tp.ctx, id, workflow.StatusRunning)
}

func (tp *Tempolite) waitWorkflowStatus(ctx context.Context, id WorkflowID, status workflow.Status) error {
	ticker := time.NewTicker(time.Second / 16)
	for {
		select {
		case <-ctx.Done():
			tp.logger.Error(tp.ctx, "waitWorkflowStatus context done")
			return ctx.Err()
		case <-ticker.C:
			_, err := tp.client.Workflow.Query().Where(workflow.ID(id.String()), workflow.StatusEQ(status)).Only(tp.ctx)
			if err != nil {
				continue
			}
			return nil
		}
	}
}

func (tp *Tempolite) CancelWorkflow(id WorkflowID) error {

	tp.logger.Debug(tp.ctx, "CancelWorkflow", "workflowID", id)

	if err := tp.PauseWorkflow(id); err != nil {
		tp.logger.Error(tp.ctx, "Error pausing workflow", "workflowID", id, "error", err)
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	waiting := make(chan string)

	// fmt.Println("cancel waiting")
	go func() {
		if err := tp.waitWorkflowStatus(ctx, id, workflow.StatusPaused); err == nil {
			waiting <- "paused"
		}
	}()

	go func() {
		if err := tp.waitWorkflowStatus(ctx, id, workflow.StatusCompleted); err == nil {
			waiting <- "completed"
		}
	}()

	var result string

	select {
	case result = <-waiting:
		tp.logger.Debug(tp.ctx, "waitted for workflow", "workflowID", id, "result", result)
		cancel()
	case <-ctx.Done():
		tp.logger.Error(tp.ctx, "cancel waiting context done")
	}

	close(waiting)

	wf, err := tp.client.Workflow.Get(tp.ctx, id.String())
	if err != nil {
		tp.logger.Error(tp.ctx, "Error getting workflow", "workflowID", id, "error", err)
		return err
	}

	queueWorkflow, err := tp.getWorkflowPoolQueue(wf.QueueName)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error getting workflow queue", "workflowID", id, "error", err)
		return err
	}

	switch result {
	case "paused":
		// fmt.Println("ranging")
		queueWorkflow.RangeTasks(func(data *workflowTask, workerID int, status retrypool.TaskStatus) bool {
			tp.logger.Debug(tp.ctx, "task still within the workflow", "workflowID", data.ctx.workflowID, "id", id.String())
			if data.ctx.workflowID == id.String() {
				queueWorkflow.InterruptWorker(workerID, retrypool.WithForcePanic(), retrypool.WithRemoveTask())
			}
			return true
		})
		tp.logger.Debug(tp.ctx, "workflow tasks removed", "workflowID", id)
	case "completed":
		tp.logger.Debug(tp.ctx, "workflow already completed", "workflowID", id)
		return nil
	}

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return err
	}
	_, err = tx.Workflow.UpdateOneID(id.String()).
		SetStatus(workflow.StatusCancelled).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error cancelling workflow", "workflowID", id, "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "workflowID", id, "error", err)
		return err
	}

	return err
}
