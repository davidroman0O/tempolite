package tempolite

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/retrypool"

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
