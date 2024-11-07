package tempolite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/workflow"

	dbSQL "database/sql"
)

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
		cfg.logger.Debug(ctx, "Fist timer check", "firstTime", firstTime)
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
		if err := workflowPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		return fmt.Errorf("failed to create activity pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating side effect pool", "name", name, "workers", sideEffectWorkers)
	sideEffectPool, err := tp.createSideEffectPool(name, sideEffectWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating side effect pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		if err := workflowPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := activityPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		return fmt.Errorf("failed to create side effect pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction pool", "name", name, "workers", transactionWorkers)
	transactionPool, err := tp.createTransactionPool(name, transactionWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		if err := workflowPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := activityPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := sideEffectPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		return fmt.Errorf("failed to create transaction pool: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating compensation pool", "name", name, "workers", compensationWorkers)
	compensationPool, err := tp.createCompensationPool(name, compensationWorkers)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating compensation pool", "error", err)
		tp.queueStatus.Delete(name)
		// Clean up previous pools
		if err := workflowPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := activityPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := sideEffectPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := transactionPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
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
		if err := workflowPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := activityPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := sideEffectPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := transactionPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
		if err := compensationPool.Shutdown(); err != nil {
			if err != context.Canceled {
				return err
			}
		}
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
	if err := workers.Workflows.Shutdown(); err != nil {
		if err != context.Canceled {
			return err
		}
	}
	if err := workers.Activities.Shutdown(); err != nil {
		if err != context.Canceled {
			return err
		}
	}
	if err := workers.SideEffects.Shutdown(); err != nil {
		if err != context.Canceled {
			return err
		}
	}
	if err := workers.Transactions.Shutdown(); err != nil {
		if err != context.Canceled {
			return err
		}
	}
	if err := workers.Compensations.Shutdown(); err != nil {
		if err != context.Canceled {
			return err
		}
	}

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

	waiting := make(chan workflow.Status)

	// fmt.Println("cancel waiting")
	go func() {
		if err := tp.waitWorkflowStatus(ctx, id, workflow.StatusPaused); err == nil {
			waiting <- workflow.StatusPaused
		}
	}()

	go func() {
		if err := tp.waitWorkflowStatus(ctx, id, workflow.StatusCompleted); err == nil {
			waiting <- workflow.StatusCompleted
		}
	}()

	var result workflow.Status

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
		queueWorkflow.RangeTasks(func(data *retrypool.TaskWrapper[*workflowTask], workerID int, status retrypool.TaskStatus) bool {
			tp.logger.Debug(tp.ctx, "task still within the workflow", "workflowID", data.Data().ctx.workflowID, "id", id.String())
			if data.Data().ctx.workflowID == id.String() {
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
