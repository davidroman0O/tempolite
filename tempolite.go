package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/featureflagversion"
	"github.com/davidroman0O/go-tempolite/ent/run"
	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/go-tempolite/ent/schema"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"

	dbSQL "database/sql"
)

/// TODO: change the registry functions
/// - register structs with directly a pointer of the struct, we should guarantee will destroy that pointer
/// - I don't want to see that AsXXX functions anymore
///
/// TODO: change enqueue functions, i don't want to see As[T] functions anymore, you either put the function or an instance of the struct
///
/// In documentation warn that struct activities need a particular care since the developer might introduce even more non-deteministic code, it should be used for struct that hold a client/api but not a value what might change the output given the same inputs since it activities won't be replayed if sucessful.
///

// Trade-off of Tempolite vs Temporal
// Supporting the same similar concepts but now the exact similar features while having less time and resources implies that knowing how Workflows/Activities/Sideffect are behaving in a deterministic and non-deterministic environment, it is crucial
// to assign some kind of identifying value to each of the calls so the whole system can works with minimum overhead and just be reliable and predictable. For now it should works for sqlite, but one day... i want the same with no db.
type Identifier interface {
	~string | ~int
}

var errWorkflowPaused = errors.New("workflow is paused")

// TODO: to be renamed as tempoliteEngine since it will be used to rotate/roll new databases when its database size is too big
type Tempolite[T Identifier] struct {
	db     *dbSQL.DB
	client *ent.Client

	// Workflows are pre-registered
	// Deterministic function that should be replayed exactly the same way if successful without triggering code since it will have only the results of the previous execution
	workflows                      sync.Map
	workflowPool                   *retrypool.Pool[*workflowTask[T]]
	schedulerExecutionWorkflowDone chan struct{}

	// Activities are pre-registered
	// Un-deterministic function
	activities                     sync.Map
	activityPool                   *retrypool.Pool[*activityTask[T]]
	schedulerExecutionActivityDone chan struct{}

	// SideEffects are dynamically cached
	// You have to use side effects to guide the flow of a workflow
	// Help to manage un-deterministic code within a workflow
	sideEffects             sync.Map
	sideEffectPool          *retrypool.Pool[*sideEffectTask[T]]
	schedulerSideEffectDone chan struct{}

	// Saga are dynamically cached
	// Simplify the management of the Sagas, if we were to use Activities we would be subject to trickery to manage the flow
	sagas             sync.Map
	transactionPool   *retrypool.Pool[*transactionTask[T]]
	compensationPool  *retrypool.Pool[*compensationTask[T]]
	schedulerSagaDone chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	// Versioning system cache
	versionCache sync.Map

	schedulerWorkflowStarted   atomic.Bool
	schedulerActivityStarted   atomic.Bool
	schedulerSideEffectStarted atomic.Bool
	schedulerSagaStarted       atomic.Bool

	resumeWorkflowsWorkerDone chan struct{}

	logger Logger
}

type tempoliteConfig struct {
	path        *string
	destructive bool
	logger      Logger
}

type tempoliteOption func(*tempoliteConfig)

func WithLogger(logger Logger) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.logger = logger
	}
}

func WithPath(path string) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = &path
	}
}

func WithMemory() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = nil
	}
}

func WithDestructive() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.destructive = true
	}
}

func New[T Identifier](ctx context.Context, registry *Registry[T], opts ...tempoliteOption) (*Tempolite[T], error) {
	cfg := tempoliteConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger = NewDefaultLogger()
	}

	ctx, cancel := context.WithCancel(ctx)

	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithBuffer(77777),
	}

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

	tp := &Tempolite[T]{
		ctx:                            ctx,
		db:                             db,
		client:                         client,
		cancel:                         cancel,
		schedulerExecutionWorkflowDone: make(chan struct{}),
		schedulerExecutionActivityDone: make(chan struct{}),
		schedulerSideEffectDone:        make(chan struct{}),
		schedulerSagaDone:              make(chan struct{}),
		resumeWorkflowsWorkerDone:      make(chan struct{}),
		logger:                         cfg.logger,
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
		if err := tp.registerActivityFunc(activity); err != nil {
			return nil, err
		}
	}

	tp.logger.Debug(ctx, "Registering activities functions", "total", len(registry.activitiesFunc))
	for _, activity := range registry.activitiesFunc {
		if err := tp.registerActivityFunc(activity); err != nil {
			return nil, err
		}
	}

	tp.logger.Debug(ctx, "Creating pools")
	tp.workflowPool = tp.createWorkflowPool()
	tp.activityPool = tp.createActivityPool()
	tp.sideEffectPool = tp.createSideEffectPool()
	tp.transactionPool = tp.createTransactionPool()
	tp.compensationPool = tp.createCompensationPool()

	tp.logger.Debug(ctx, "Starting scheduler side effect")
	go tp.schedulerExecutionSideEffect()
	tp.logger.Debug(ctx, "Starting scheduler workflow executions")
	go tp.schedulerExecutionWorkflow()
	tp.logger.Debug(ctx, "Starting scheduler activity executions")
	go tp.schedulerExecutionActivity()
	tp.logger.Debug(ctx, "Starting scheduler saga executions")
	go tp.schedulerExecutionSaga()
	tp.logger.Debug(ctx, "Starting resume workflows worker")
	go tp.resumeWorkflowsWorker()

	return tp, nil
}

// When workflows were paused, but will be flagged as IsReady
func (tp *Tempolite[T]) resumeWorkflowsWorker() {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	defer close(tp.resumeWorkflowsWorkerDone)
	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "Resume workflows worker due to context done")
			return
		case <-tp.resumeWorkflowsWorkerDone:
			tp.logger.Debug(tp.ctx, "Resume workflows worker done")
			return
		case <-ticker.C:
			workflows, err := tp.client.Workflow.Query().
				Where(
					workflow.StatusEQ(workflow.StatusRunning),
					workflow.IsPausedEQ(false),
					workflow.IsReadyEQ(true),
				).All(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Error querying workflows to resume", "error", err)
				continue
			}

			for _, wf := range workflows {
				tp.logger.Debug(tp.ctx, "Resuming workflow", "workflowID", wf.ID)
				if err := tp.redispatchWorkflow(WorkflowID(wf.ID)); err != nil {
					tp.logger.Error(tp.ctx, "Error redispatching workflow", "workflowID", wf.ID, "error", err)
				}
				if _, err := tp.client.Workflow.UpdateOne(wf).SetIsReady(false).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Error updating workflow", "workflowID", wf.ID, "error", err)
				}
			}
		}
	}
}

func (tp *Tempolite[T]) redispatchWorkflow(id WorkflowID) error {

	wf, err := tp.client.Workflow.Get(tp.ctx, id.String())
	if err != nil {
		tp.logger.Error(tp.ctx, "Error fetching workflow", "workflowID", id, "error", err)
		return fmt.Errorf("error fetching workflow: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Redispatching workflow", "workflowID", id)
	tp.logger.Debug(tp.ctx, "Querying workflow execution", "workflowID", id)

	wfEx, err := tp.client.WorkflowExecution.Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).
		WithWorkflow().
		Order(ent.Desc(workflowexecution.FieldStartedAt)).
		First(tp.ctx)

	if err != nil {
		tp.logger.Error(tp.ctx, "Error querying workflow execution", "workflowID", id, "error", err)
		return fmt.Errorf("error querying workflow execution: %w", err)
	}

	// Create WorkflowContext
	ctx := WorkflowContext[T]{
		tp:           tp,
		workflowID:   wf.ID,
		executionID:  wfEx.ID,
		runID:        wfEx.RunID,
		workflowType: wf.Identity,
		stepID:       wf.StepID,
	}

	tp.logger.Debug(tp.ctx, "Getting handler info for workflow", "workflowID", id, "handlerIdentity", wf.Identity)

	// Fetch the workflow handler
	handlerInfo, ok := tp.workflows.Load(HandlerIdentity(wf.Identity))
	if !ok {
		tp.logger.Error(tp.ctx, "Workflow handler not found", "workflowID", id)
		return fmt.Errorf("workflow handler not found for %s", wf.Identity)
	}

	workflowHandler, ok := handlerInfo.(Workflow)
	if !ok {
		tp.logger.Error(tp.ctx, "Invalid workflow handler", "workflowID", id)
		return fmt.Errorf("invalid workflow handler for %s", wf.Identity)
	}

	inputs := []interface{}{}

	tp.logger.Debug(tp.ctx, "Converting inputs", "workflowID", id, "inputs", len(wf.Input))
	// TODO: we can probably parallelize this
	for idx, rawInput := range wf.Input {
		inputType := workflowHandler.ParamTypes[idx]
		inputKind := workflowHandler.ParamsKinds[idx]

		realInput, err := convertIO(rawInput, inputType, inputKind)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error converting input", "workflowID", id, "error", err)
			return err
		}

		inputs = append(inputs, realInput)
	}

	tp.logger.Debug(tp.ctx, "Creating workflow task", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

	// Create and dispatch the workflow task
	task := &workflowTask[T]{
		ctx:         ctx,
		handlerName: workflowHandler.HandlerLongName,
		handler:     workflowHandler.Handler,
		params:      inputs,
		maxRetry:    wf.RetryPolicy.MaximumAttempts,
	}

	retryIt := func() error {

		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)

		// create a new execution for the same workflow
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tp.client.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(wfEx.RunID).
			SetWorkflow(wf).
			Save(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating workflow execution during retry", "workflowID", id, "error", err)
			return err
		}

		task.ctx.executionID = workflowExecution.ID
		task.retryCount++
		tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "retryCount", task.retryCount)

		tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "status", workflow.StatusRunning)
		// now we notify the workflow enity that we're working
		if _, err = tp.client.Workflow.UpdateOneID(ctx.workflowID).SetStatus(workflow.StatusRunning).Save(tp.ctx); err != nil {
			log.Printf("scheduler: Workflow.UpdateOneID failed: %v", err)
		}

		return nil
	}

	task.retry = retryIt

	tp.logger.Debug(tp.ctx, "Fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
	total, err := tp.client.WorkflowExecution.
		Query().
		Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(wf.ID))).Count(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error fetching total execution workflow related to workflow entity", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "error", err)
		return err
	}

	tp.logger.Debug(tp.ctx, "Total previously created workflow execution", "workflowID", id, "handlerName", workflowHandler.HandlerLongName, "total", total)

	// If it's not me
	if total > 1 {
		task.retryCount = total
	}

	tp.logger.Debug(tp.ctx, "Dispatching workflow", "workflowID", id, "handlerName", workflowHandler.HandlerLongName)
	return tp.workflowPool.Dispatch(task)
}

func (tp *Tempolite[T]) getOrCreateVersion(workflowType, workflowID, changeID string, minSupported, maxSupported int) (int, error) {
	key := fmt.Sprintf("%s-%s", workflowType, changeID)

	tp.logger.Debug(tp.ctx, "Checking cache for version", "key", key)

	// Check cache first
	if cachedVersion, ok := tp.versionCache.Load(key); ok {
		version := cachedVersion.(int)
		tp.logger.Debug(tp.ctx, "Found cached version", "version", version, "key", key)
		// Update version if necessary
		if version < maxSupported {
			version = maxSupported
			tp.versionCache.Store(key, version)
			tp.logger.Debug(tp.ctx, "Updated cached version", "version", version, "key", key)
		}
		tp.logger.Debug(tp.ctx, "Returning cached version", "version", version, "key", key)
		return version, nil
	}

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction when getting or creating version", "error", err)
		return 0, err
	}
	defer tx.Rollback()

	tp.logger.Debug(tp.ctx, "Querying feature flag version", "workflowType", workflowType, "changeID", changeID)
	v, err := tx.FeatureFlagVersion.Query().
		Where(featureflagversion.WorkflowTypeEQ(workflowType)).
		Where(featureflagversion.ChangeIDEQ(changeID)).
		Only(tp.ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			tp.logger.Error(tp.ctx, "Error querying feature flag version", "error", err)
			return 0, err
		}
		tp.logger.Debug(tp.ctx, "Creating new version", "workflowType", workflowType, "changeID", changeID, "version", minSupported)
		// Version not found, create a new one
		log.Printf("Creating new version for key: %s with value: %d", key, minSupported)
		v, err = tx.FeatureFlagVersion.Create().
			SetWorkflowType(workflowType).
			SetWorkflowID(workflowID).
			SetChangeID(changeID).
			SetVersion(minSupported).
			Save(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error creating feature flag version", "error", err)
			return 0, err
		}
		tp.logger.Debug(tp.ctx, "Created new version", "workflowType", workflowType, "changeID", changeID, "version", minSupported)
	} else {
		tp.logger.Debug(tp.ctx, "Found existing version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
		// Update the version if maxSupported is greater
		if v.Version < maxSupported {
			v.Version = maxSupported
			v, err = tx.FeatureFlagVersion.UpdateOne(v).
				SetVersion(v.Version).
				SetWorkflowID(workflowID).
				Save(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Error updating feature flag version", "error", err)
				return 0, err
			}
			tp.logger.Debug(tp.ctx, "Updated version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
		}
	}

	tp.logger.Debug(tp.ctx, "Committing transaction version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "error", err)
		return 0, err
	}

	// Update cache
	tp.versionCache.Store(key, v.Version)
	tp.logger.Debug(tp.ctx, "Stored version in cache", "version", v.Version, "key", key)

	return v.Version, nil
}

func (tp *Tempolite[T]) Close() {
	tp.logger.Debug(tp.ctx, "Closing Tempolite")
	tp.cancel() // it will stops other systems
	tp.workflowPool.Close()
	tp.activityPool.Close()
	tp.sideEffectPool.Close()
	tp.transactionPool.Close()
	tp.compensationPool.Close()
}

func (tp *Tempolite[T]) Wait() error {
	<-time.After(1 * time.Second)

	activityDone := make(chan error)
	workflowDone := make(chan error)
	sideEffectDone := make(chan error)
	transactionDone := make(chan error)
	compensationDone := make(chan error)

	doneSignals := []chan error{activityDone, workflowDone, sideEffectDone, transactionDone, compensationDone}

	tp.logger.Debug(tp.ctx, "Waiting for scheduler to start")
	for !tp.schedulerWorkflowStarted.Load() || !tp.schedulerActivityStarted.Load() || !tp.schedulerSideEffectStarted.Load() || !tp.schedulerSagaStarted.Load() {
		runtime.Gosched()
	}
	tp.logger.Debug(tp.ctx, "Waiting for scheduler to start done")

	go func() {
		defer close(activityDone)
		defer tp.logger.Debug(tp.ctx, "Finished waiting for activities")
		activityDone <- tp.activityPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			tp.logger.Debug(tp.ctx, "Wait Activity Pool", "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
			tp.activityPool.RangeTasks(func(data *activityTask[T], workerID int, status retrypool.TaskStatus) bool {
				tp.logger.Debug(tp.ctx, "Activity Pool RangeTask", "workerID", workerID, "status", status, "task", data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	go func() {
		defer close(workflowDone)
		defer tp.logger.Debug(tp.ctx, "Finished waiting for workflows")
		workflowDone <- tp.workflowPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			tp.logger.Debug(tp.ctx, "Wait Workflow Pool", "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
			tp.workflowPool.RangeTasks(func(data *workflowTask[T], workerID int, status retrypool.TaskStatus) bool {
				tp.logger.Debug(tp.ctx, "Workflow Pool RangeTask", "workerID", workerID, "status", status, "task", data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	go func() {
		defer close(sideEffectDone)
		defer tp.logger.Debug(tp.ctx, "Finished waiting for side effects")
		sideEffectDone <- tp.sideEffectPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			tp.logger.Debug(tp.ctx, "Wait SideEffect Pool", "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
			tp.sideEffectPool.RangeTasks(func(data *sideEffectTask[T], workerID int, status retrypool.TaskStatus) bool {
				tp.logger.Debug(tp.ctx, "SideEffect Pool RangeTask", "workerID", workerID, "status", status, "task", data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	go func() {
		defer close(transactionDone)
		defer tp.logger.Debug(tp.ctx, "Finished waiting for transactions")
		transactionDone <- tp.transactionPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			tp.logger.Debug(tp.ctx, "Wait Transaction Pool", "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
			tp.transactionPool.RangeTasks(func(data *transactionTask[T], workerID int, status retrypool.TaskStatus) bool {
				tp.logger.Debug(tp.ctx, "Transaction Pool RangeTask", "workerID", workerID, "status", status, "task", data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	go func() {
		defer close(compensationDone)
		defer tp.logger.Debug(tp.ctx, "Finished waiting for compensations")
		compensationDone <- tp.compensationPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			tp.logger.Debug(tp.ctx, "Wait Compensation Pool", "queueSize", queueSize, "processingCount", processingCount, "deadTaskCount", deadTaskCount)
			tp.compensationPool.RangeTasks(func(data *compensationTask[T], workerID int, status retrypool.TaskStatus) bool {
				tp.logger.Debug(tp.ctx, "Compensation Pool RangeTask", "workerID", workerID, "status", status, "task", data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	tp.logger.Debug(tp.ctx, "Waiting for done signals")
	for _, doneSignal := range doneSignals {
		if err := <-doneSignal; err != nil {
			tp.logger.Error(tp.ctx, "Error waiting for done signal", "error", err)
			return err
		}
	}

	tp.logger.Debug(tp.ctx, "Done waiting for signals")

	return nil
}

func (tp *Tempolite[T]) convertInputs(handlerInfo HandlerInfo, executionInputs []interface{}) ([]interface{}, error) {
	tp.logger.Debug(tp.ctx, "Converting inputs", "handlerInfo", handlerInfo)
	outputs := []interface{}{}
	// TODO: we can probably parallelize this
	for idx, rawInputs := range executionInputs {
		inputType := handlerInfo.ParamTypes[idx]
		inputKind := handlerInfo.ParamsKinds[idx]
		tp.logger.Debug(tp.ctx, "Converting input", "inputType", inputType, "inputKind", inputKind, "rawInputs", rawInputs)
		realInput, err := convertIO(rawInputs, inputType, inputKind)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error converting input", "error", err)
			return nil, err
		}
		outputs = append(outputs, realInput)
	}
	return outputs, nil
}

func (tp *Tempolite[T]) convertOuputs(handlerInfo HandlerInfo, executionOutput []interface{}) ([]interface{}, error) {
	tp.logger.Debug(tp.ctx, "Converting outputs", "handlerInfo", handlerInfo)
	outputs := []interface{}{}
	// TODO: we can probably parallelize this
	for idx, rawOutput := range executionOutput {
		ouputType := handlerInfo.ReturnTypes[idx]
		outputKind := handlerInfo.ReturnKinds[idx]
		tp.logger.Debug(tp.ctx, "Converting output", "ouputType", ouputType, "outputKind", outputKind, "rawOutput", rawOutput)
		realOutput, err := convertIO(rawOutput, ouputType, outputKind)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error converting output", "error", err)
			return nil, err
		}
		outputs = append(outputs, realOutput)
	}
	return outputs, nil
}

func (tp *Tempolite[T]) verifyHandlerAndParams(handlerInfo HandlerInfo, params []interface{}) error {

	if len(params) != handlerInfo.NumIn {
		tp.logger.Error(tp.ctx, "Parameter count mismatch", "expected", handlerInfo.NumIn, "got", len(params))
		return fmt.Errorf("parameter count mismatch (you probably put the wrong handler): expected %d, got %d", handlerInfo.NumIn, len(params))
	}

	for idx, param := range params {
		if reflect.TypeOf(param) != handlerInfo.ParamTypes[idx] {
			tp.logger.Error(tp.ctx, "Parameter type mismatch", "expected", handlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
			return fmt.Errorf("parameter type mismatch (you probably put the wrong handler) at index %d: expected %s, got %s", idx, handlerInfo.ParamTypes[idx], reflect.TypeOf(param))
		}
	}

	return nil
}

func (tp *Tempolite[T]) enqueueActivity(ctx WorkflowContext[T], stepID T, longName HandlerIdentity, params ...interface{}) (ActivityID, error) {

	tp.logger.Debug(tp.ctx, "EnqueueActivity", "stepID", stepID, "longName", longName)
	switch ctx.EntityType() {
	case "workflow":
		// nothing
	default:
		tp.logger.Error(tp.ctx, "Context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error

	// Check for existing activity with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeActivity),
		).
		First(tp.ctx)
	if err == nil {
		// todo: is there a way to just get the status?
		act, err := tp.client.Activity.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting activity", "error", err)
			return "", err
		}
		if act.Status == activity.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Activity already completed", "activityID", exists.ChildEntityID)
			return ActivityID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating activity", "longName", longName)

	var value any
	var ok bool
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching activity handler", "longName", longName)
	if value, ok = tp.activities.Load(longName); ok {
		var activityHandlerInfo Activity
		if activityHandlerInfo, ok = value.(Activity); !ok {
			// could be development bug
			tp.logger.Error(tp.ctx, "Activity is not handler info", "longName", longName)
			return "", fmt.Errorf("activity %s is not handler info", longName)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "activityHandlerInfo", activityHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(activityHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create activity", "longName", longName)
		// Proceed to create a new activity and activity execution
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating activity entity", "longName", longName, "stepID", stepID)
		var activityEntity *ent.Activity
		if activityEntity, err = tx.
			Activity.
			Create().
			SetID(uuid.NewString()).
			SetStepID(fmt.Sprint(stepID)).
			SetIdentity(string(longName)).
			SetHandlerName(activityHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating activity entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating activity entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Created activity execution", "activityID", activityEntity.ID, "longName", longName, "stepID", stepID)
		// Create activity execution with the deterministic ID
		var activityExecution *ent.ActivityExecution
		if activityExecution, err = tx.ActivityExecution.
			Create().
			SetID(activityEntity.ID). // Use the deterministic activity ID
			SetRunID(ctx.RunID()).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating activity execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating activity execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Created activity relationship", "activityID", activityEntity.ID, "actvityExecution", activityExecution.ID, "longName", longName, "stepID", stepID)
		// Add execution relationship
		if _, err := tx.ExecutionRelationship.Create().
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(activityEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(activityEntity.ID).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(fmt.Sprint(stepID)).
			// types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeActivity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Committing transaction creating activity", "longName", longName)
		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Activity created", "activityID", activityEntity.ID, "longName", longName, "stepID", stepID)
		return ActivityID(activityEntity.ID), nil

	} else {
		tp.logger.Error(tp.ctx, "Activity not found", "longName", longName)
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite[T]) enqueueActivityFunc(ctx WorkflowContext[T], stepID T, activityFunc interface{}, params ...interface{}) (ActivityID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	tp.logger.Debug(tp.ctx, "Enqueue ActivityFunc", "stepID", stepID, "handlerIdentity", handlerIdentity)
	return tp.enqueueActivity(ctx, stepID, handlerIdentity, params...)
}

func (tp *Tempolite[T]) enqueueWorkflow(ctx TempoliteContext, stepID T, workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with sub-workflow creation
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	// Check for existing sub-workflow with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeWorkflow),
		).
		First(tp.ctx)
	if err == nil {
		tp.logger.Debug(tp.ctx, "Existing sub-workflow found", "childEntityID", exists.ChildEntityID)
		// todo: is there a way to just get the status?
		act, err := tp.client.Workflow.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting workflow", "error", err)
			return "", err
		}
		if act.Status == workflow.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Sub-workflow already completed", "workflowID", exists.ChildEntityID)
			return WorkflowID(exists.ChildEntityID), nil
		}
	} else {
		if !ent.IsNotFound(err) {
			tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
			return "", fmt.Errorf("error checking for existing stepID: %w", err)
		}
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching workflow handler", "handlerIdentity", handlerIdentity)
	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			// could be development bug
			tp.logger.Error(tp.ctx, "Workflow is not handler info", "handlerIdentity", handlerIdentity)
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "workflowHandlerInfo", workflowHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		if len(params) != workflowHandlerInfo.NumIn {
			tp.logger.Error(tp.ctx, "Parameter count mismatch", "expected", workflowHandlerInfo.NumIn, "got", len(params))
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				tp.logger.Error(tp.ctx, "Parameter type mismatch", "expected", workflowHandlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create workflow", "handlerIdentity", handlerIdentity)
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction to create workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow entity")
		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(uuid.NewString()).
			SetStepID(fmt.Sprint(stepID)).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow execution")
		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(ctx.RunID()).
			SetWorkflow(workflowEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow relationship")
		if _, err := tx.ExecutionRelationship.Create().
			// run id
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(workflowEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(workflowExecution.ID).
			// Types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeWorkflow).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(fmt.Sprint(stepID)).
			//
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Committing transaction creating workflow", "handlerIdentity", handlerIdentity)
		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error committing transaction creating workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Workflow created", "workflowID", workflowEntity.ID, "handlerIdentity", handlerIdentity)
		return WorkflowID(workflowEntity.ID), nil

	} else {
		tp.logger.Error(tp.ctx, "Workflow not found", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite[T]) Workflow(stepID T, workflowFunc interface{}, params ...interface{}) *WorkflowInfo[T] {
	tp.logger.Debug(tp.ctx, "Workflow", "stepID", stepID)
	id, err := tp.executeWorkflow(stepID, workflowFunc, params...)
	return tp.getWorkflowRoot(id, err)
}

func (tp *Tempolite[T]) executeWorkflow(stepID T, workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching workflow handler", "handlerIdentity", handlerIdentity)
	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			tp.logger.Error(tp.ctx, "Workflow is not handler info", "handlerIdentity", handlerIdentity)
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "workflowHandlerInfo", workflowHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		if len(params) != workflowHandlerInfo.NumIn {
			tp.logger.Error(tp.ctx, "Parameter count mismatch", "expected", workflowHandlerInfo.NumIn, "got", len(params))
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				tp.logger.Error(tp.ctx, "Parameter type mismatch", "expected", workflowHandlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create workflow", "handlerIdentity", handlerIdentity)
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction to create workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating root run entity")
		var runEntity *ent.Run
		// root Run entity that anchored the workflow entity despite retries
		if runEntity, err = tx.
			Run.
			Create().
			SetID(uuid.NewString()).    // immutable
			SetRunID(uuid.NewString()). // can change if that flow change due to a retry
			SetType(run.TypeWorkflow).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating root run entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating root run entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow entity")
		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(runEntity.RunID).
			SetStatus(workflow.StatusPending).
			SetStepID(fmt.Sprint(stepID)).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow execution")
		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(runEntity.ID).
			SetWorkflow(workflowEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "workflow execution created", "workflowExecutionID", workflowExecution.ID, "handlerIdentity", handlerIdentity)

		tp.logger.Debug(tp.ctx, "Updating run entity with workflow")
		if _, err = tx.Run.UpdateOneID(runEntity.ID).SetWorkflow(workflowEntity).Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction updating run entity with workflow", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error updating run entity with workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Commiting transaction creating workflow", "handlerIdentity", handlerIdentity)
		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error committing transaction creating workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Workflow created", "workflowID", workflowEntity.ID, "handlerIdentity", handlerIdentity)
		//	we're outside of the execution model, so we care about the workflow entity
		return WorkflowID(workflowEntity.ID), nil
	} else {
		tp.logger.Error(tp.ctx, "Workflow not found", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite[T]) GetWorkflow(id WorkflowID) *WorkflowInfo[T] {
	tp.logger.Debug(tp.ctx, "GetWorkflow", "workflowID", id)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflowRoot(id WorkflowID, err error) *WorkflowInfo[T] {
	tp.logger.Debug(tp.ctx, "getWorkflowRoot", "workflowID", id, "error", err)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflow(ctx TempoliteContext, id WorkflowID, err error) *WorkflowInfo[T] {
	tp.logger.Debug(tp.ctx, "getWorkflow", "workflowID", id, "error", err)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflowExecution(ctx TempoliteContext, id WorkflowExecutionID, err error) *WorkflowExecutionInfo[T] {
	tp.logger.Debug(tp.ctx, "getWorkflowExecution", "workflowExecutionID", id, "error", err)
	info := WorkflowExecutionInfo[T]{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite[T]) GetActivity(id ActivityID) (*ActivityInfo[T], error) {
	tp.logger.Debug(tp.ctx, "GetActivity", "activityID", id)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
	}
	return &info, nil
}

func (tp *Tempolite[T]) getActivity(ctx TempoliteContext, id ActivityID, err error) *ActivityInfo[T] {
	tp.logger.Debug(tp.ctx, "getActivity", "activityID", id, "error", err)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getActivityExecution(ctx TempoliteContext, id ActivityExecutionID, err error) *ActivityExecutionInfo[T] {
	tp.logger.Debug(tp.ctx, "getActivityExecution", "activityExecutionID", id, "error", err)
	info := ActivityExecutionInfo[T]{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tp *Tempolite[T]) GetSideEffect(id SideEffectID) *SideEffectInfo[T] {
	tp.logger.Debug(tp.ctx, "GetSideEffect", "sideEffectID", id)
	return &SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
	}
}

func (tp *Tempolite[T]) getSideEffect(ctx TempoliteContext, id SideEffectID, err error) *SideEffectInfo[T] {
	tp.logger.Debug(tp.ctx, "getSideEffect", "sideEffectID", id, "error", err)
	info := SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
		err:      err,
	}
	return &info
}

// That's how we should use it
//
// ```go
// var value int
//
//	err := ctx.SideEffect("eventual switch", func(ctx SideEffectContext[testIdentifier]) int {
//		return 420
//	}).Get(&value)
//
// ```
func (tp *Tempolite[T]) enqueueSideEffect(ctx TempoliteContext, stepID T, sideEffectHandler interface{}) (SideEffectID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with side effect creation
	default:
		tp.logger.Error(tp.ctx, "creating side effect context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error
	var tx *ent.Tx

	// Generate a unique identifier for the side effect function
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffectHandler).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	tp.logger.Debug(tp.ctx, "EnqueueSideEffect", "stepID", stepID, "handlerIdentity", handlerIdentity)

	tp.logger.Debug(tp.ctx, "check for existing side effect", "stepID", stepID)
	// Check for existing side effect with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSideEffect),
		).
		First(tp.ctx)
	if err == nil {
		sideEffectEntity, err := tp.client.SideEffect.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting side effect", "error", err)
			return "", err
		}
		if sideEffectEntity.Status == sideeffect.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Side effect already completed", "sideEffectID", exists.ChildEntityID)
			return SideEffectID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction to create side effect", "handlerIdentity", handlerIdentity)
	if tx, err = tp.client.Tx(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Creating side effect entity", "handlerIdentity", handlerIdentity)
	sideEffectEntity, err := tx.SideEffect.
		Create().
		SetID(uuid.NewString()).
		SetStepID(fmt.Sprint(stepID)).
		SetIdentity(string(handlerIdentity)).
		SetHandlerName(funcName).
		SetStatus(sideeffect.StatusPending).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect entity", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating side effect entity", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Creating side effect execution")
	sideEffectExecution, err := tx.SideEffectExecution.
		Create().
		SetID(uuid.NewString()).
		// SetRunID(ctx.RunID()).
		SetSideEffect(sideEffectEntity).
		SetStatus(sideeffectexecution.StatusPending).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect execution", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating side effect execution", "error", err)
		return "", fmt.Errorf("failed to create side effect execution: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(sideEffectEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(sideEffectExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSideEffect).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "analyze side effect handler", "handlerIdentity", handlerIdentity)

	// Analyze the side effect handler
	handlerType := reflect.TypeOf(sideEffectHandler)
	if handlerType.Kind() != reflect.Func {
		tp.logger.Error(tp.ctx, "Side effect must be a function", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect must be a function")
	}

	if handlerType.NumIn() != 1 || handlerType.In(0) != reflect.TypeOf(SideEffectContext[T]{}) {
		tp.logger.Error(tp.ctx, "Side effect function must have exactly one input parameter of type SideEffectContext[T]", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect function must have exactly one input parameter of type SideEffectContext[T]")
	}

	// Collect all return types
	numOut := handlerType.NumOut()
	if numOut == 0 {
		tp.logger.Error(tp.ctx, "Side effect function must return at least one value", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect function must return at least one value")
	}

	returnTypes := make([]reflect.Type, numOut)
	returnKinds := make([]reflect.Kind, numOut)
	for i := 0; i < numOut; i++ {
		returnTypes[i] = handlerType.Out(i)
		returnKinds[i] = handlerType.Out(i).Kind()
	}

	tp.logger.Debug(tp.ctx, "Caching side effect info", "handlerIdentity", handlerIdentity)
	// Cache the side effect info
	tp.sideEffects.Store(sideEffectEntity.ID, SideEffect{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         sideEffectHandler,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumOut:          numOut,
	})

	tp.logger.Debug(tp.ctx, "Committing transaction creating side effect", "handlerIdentity", handlerIdentity)
	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			tp.sideEffects.Delete(sideEffectEntity.ID)
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect", "error", err)
			return "", err
		}
		tp.sideEffects.Delete(sideEffectEntity.ID)
		tp.logger.Error(tp.ctx, "Error committing transaction creating side effect", "error", err)
		return "", err
	}
	tp.logger.Debug(tp.ctx, "Side effect created", "sideEffectID", sideEffectEntity.ID, "handlerIdentity", handlerIdentity)

	log.Printf("Enqueued side effect %s with ID %s", funcName, sideEffectEntity.ID)
	return SideEffectID(sideEffectEntity.ID), nil
}

func (tp *Tempolite[T]) enqueueSideEffectFunc(ctx TempoliteContext, stepID T, sideEffect interface{}) (SideEffectID, error) {
	tp.logger.Debug(tp.ctx, "EnqueueSideEffectFunc", "stepID", stepID)
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffect).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSideEffect(ctx, stepID, handlerIdentity)
}

func (tp *Tempolite[T]) saga(ctx TempoliteContext, stepID T, saga *SagaDefinition[T]) *SagaInfo[T] {
	tp.logger.Debug(tp.ctx, "Saga", "stepID", stepID)
	id, err := tp.enqueueSaga(ctx, stepID, saga)
	return tp.getSaga(id, err)
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite[T]) enqueueSaga(ctx TempoliteContext, stepID T, sagaDef *SagaDefinition[T]) (SagaID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with saga creation
	default:
		tp.logger.Error(tp.ctx, "creating saga context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error
	var tx *ent.Tx
	tp.logger.Debug(tp.ctx, "EnqueueSaga", "stepID", stepID)

	tp.logger.Debug(tp.ctx, "check for existing saga", "stepID", stepID)
	// Check for existing saga with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSaga),
		).
		First(tp.ctx)
	if err == nil {
		sagaEntity, err := tp.client.Saga.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting saga", "error", err)
			return "", err
		}
		if sagaEntity.Status == saga.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Saga already completed", "sagaID", exists.ChildEntityID)
			return SagaID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction to create saga")
	if tx, err = tp.client.Tx(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
		return "", err
	}

	// Convert SagaHandlerInfo to SagaDefinitionData
	sagaDefData := schema.SagaDefinitionData{
		Steps: make([]schema.SagaStepPair, len(sagaDef.HandlerInfo.TransactionInfo)),
	}

	for i, txInfo := range sagaDef.HandlerInfo.TransactionInfo {
		sagaDefData.Steps[i] = schema.SagaStepPair{
			TransactionHandlerName:  txInfo.HandlerName,
			CompensationHandlerName: sagaDef.HandlerInfo.CompensationInfo[i].HandlerName,
		}
	}

	tp.logger.Debug(tp.ctx, "Creating saga entity")
	sagaEntity, err := tx.Saga.
		Create().
		SetID(uuid.NewString()).
		SetRunID(ctx.RunID()).
		SetStepID(fmt.Sprint(stepID)).
		SetStatus(saga.StatusPending).
		SetSagaDefinition(sagaDefData).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating saga entity", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating saga entity", "error", err)
		return "", err
	}

	// Create the first transaction step
	firstStep := sagaDefData.Steps[0]
	sagaExecution, err := tx.SagaExecution.
		Create().
		SetID(uuid.NewString()).
		SetStatus(sagaexecution.StatusPending).
		SetStepType(sagaexecution.StepTypeTransaction).
		SetHandlerName(firstStep.TransactionHandlerName).
		SetSequence(0).
		SetSaga(sagaEntity).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating first saga execution step", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating first saga execution step", "error", err)
		return "", fmt.Errorf("failed to create first saga execution step: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(sagaEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(sagaExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSaga).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Committing transaction creating saga")
	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating saga", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error committing transaction creating saga", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Saga created", "sagaID", sagaEntity.ID)
	// store the cache
	tp.sagas.Store(sagaEntity.ID, sagaDef)

	tp.logger.Debug(tp.ctx, "Created saga", "sagaID", sagaEntity.ID)
	return SagaID(sagaEntity.ID), nil
}

func (tp *Tempolite[T]) getSaga(id SagaID, err error) *SagaInfo[T] {
	tp.logger.Debug(tp.ctx, "getSaga", "sagaID", id, "error", err)
	return &SagaInfo[T]{
		tp:     tp,
		SagaID: id,
		err:    err,
	}
}

func (tp *Tempolite[T]) ListPausedWorkflows() ([]WorkflowID, error) {
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

func (tp *Tempolite[T]) PauseWorkflow(id WorkflowID) error {
	_, err := tp.client.Workflow.UpdateOneID(id.String()).
		SetIsPaused(true).
		SetIsReady(false).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error pausing workflow", "workflowID", id, "error", err)
		return err
	}
	return nil
}

func (tp *Tempolite[T]) ResumeWorkflow(id WorkflowID) error {
	_, err := tp.client.Workflow.UpdateOneID(id.String()).
		SetIsPaused(false).
		SetIsReady(true).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error resuming workflow", "workflowID", id, "error", err)
		return err
	}
	return nil
}

func (tp *Tempolite[T]) waitWorkflowStatus(ctx context.Context, id WorkflowID, status workflow.Status) error {
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

func (tp *Tempolite[T]) CancelWorkflow(id WorkflowID) error {

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

	switch result {
	case "paused":
		// fmt.Println("ranging")
		tp.workflowPool.RangeTasks(func(data *workflowTask[T], workerID int, status retrypool.TaskStatus) bool {
			tp.logger.Debug(tp.ctx, "task still within the workflow", "workflowID", data.ctx.workflowID, "id", id.String())
			if data.ctx.workflowID == id.String() {
				tp.workflowPool.InterruptWorker(workerID, retrypool.WithForcePanic(), retrypool.WithRemoveTask())
			}
			return true
		})
		tp.logger.Debug(tp.ctx, "workflow tasks removed", "workflowID", id)
	case "completed":
		tp.logger.Debug(tp.ctx, "workflow already completed", "workflowID", id)
		return nil
	}

	_, err := tp.client.Workflow.UpdateOneID(id.String()).
		SetStatus(workflow.StatusCancelled).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error cancelling workflow", "workflowID", id, "error", err)
	}

	return err
}
