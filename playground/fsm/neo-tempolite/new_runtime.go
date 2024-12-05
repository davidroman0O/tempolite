package tempolite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
	"github.com/sasha-s/go-deadlock"
	"github.com/sethvargo/go-retry"
	"github.com/stephenfire/go-rtl"
	"go.uber.org/automaxprocs/maxprocs"
)

/// The Orchestrator/Workflow engine works like a stack of event - similar to an event log - where each execution is a new event in the stack of an entity.
/// Entity got executions, executions represent the attempts at running the entity. We cannot determine the amount of execution by counting because we could pause/resume all the time which lead to stupid queries.
///
/// Every component got an Entity/Execution model with specific defined structures: WorkflowEntity/WorkflowExecution, ActivityEntity/ActivityExecution, SideEffectEntity/SideEffectExecution, SagaEntity/SagaExecution.
/// The entity is the main model that holds the state of the execution, the execution is the attempt at running the entity. The RestryState is the truth of the real attempts.
///
/// If you ContinueAsNew: you will get a new workflow entity with a new execution and a new retry state.
/// If you pause and then resume a workflow: you will have the same workflow entity but a new execution will be created keeping the previous one in a paused state.
///
/// The runtime state is represented by an Orchestrator which is your lowest API to interact with one workflow.

/// TODO: might add future.setError everywhere we return within fsm

var logger Logger = NewDefaultLogger(slog.LevelInfo, TextFormat)

func init() {
	maxprocs.Set()
	deadlock.Opts.DeadlockTimeout = time.Second * 2
	deadlock.Opts.OnPotentialDeadlock = func() {
		logger.Error(context.Background(), "POTENTIAL DEADLOCK DETECTED!")
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
	}
}

var (
	ErrWorkflowPanicked   = errors.New("workflow panicked")
	ErrActivityPanicked   = errors.New("activity panicked")
	ErrSideEffectPanicked = errors.New("side effect panicked")
	ErrSagaPanicked       = errors.New("saga panicked")

	ErrWorkflowFailed   = errors.New("workflow failed")
	ErrActivityFailed   = errors.New("activity failed")
	ErrSideEffectFailed = errors.New("side effect failed")
	ErrSagaFailed       = errors.New("saga failed")
	ErrSagaCompensated  = errors.New("saga compensated")
)

type Future interface {
	Get(out ...interface{}) error
	GetResults() ([]interface{}, error)
	setEntityID(entityID int)
	setError(err error)
	setResult(results []interface{})
	WorkflowID() int
	IsPaused() bool
}

type crossWorkflow func(queueName string, workflowID int, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future

type continueAsNew func(queueName string, workflowID int, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future

type signalNew func(workflowID int, signal string) Future

type TransactionContext struct {
	ctx         context.Context
	db          Database
	contextID   int
	executionID int
	stepIndex   int
}

var ErrMustPointer = errors.New("value must be a pointer")
var ErrEncoding = errors.New("failed to encode value")
var ErrSagaGetValue = errors.New("failed to get saga value")

var ErrTransactionContext = errors.New("failed transaction context")

func (tc *TransactionContext) Store(key string, value interface{}) error {
	var err error

	logger.Debug(tc.ctx, "transaction storing value", "transaction_context.key", key, "transaction_context.value", value, "transaction_context.execution_id", tc.executionID, "transaction_context.step_index", tc.stepIndex, "transaction_context.context_id", tc.contextID)

	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		err := errors.Join(ErrTransactionContext, ErrMustPointer)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key)
		return err
	}

	buf := new(bytes.Buffer)
	if err = rtl.Encode(reflect.ValueOf(value).Elem().Interface(), buf); err != nil {
		err := errors.Join(ErrTransactionContext, ErrEncoding, err)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key, "error", err)
		return err
	}

	if _, err = tc.db.SetSagaValue(tc.executionID, key, buf.Bytes()); err != nil {
		err := errors.Join(ErrTransactionContext, err)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key, "error", err)
		return err
	}

	logger.Debug(tc.ctx, "transaction stored value", "transaction_context.key", key, "transaction_context.value", value, "transaction_context.execution_id", tc.executionID, "transaction_context.step_index", tc.stepIndex, "transaction_context.context_id", tc.contextID)

	return nil
}

func (tc *TransactionContext) Load(key string, value interface{}) error {
	var data []byte
	var err error

	logger.Debug(tc.ctx, "transaction loading value", "transaction_context.key", key, "transaction_context.execution_id", tc.executionID, "transaction_context.step_index", tc.stepIndex, "transaction_context.context_id", tc.contextID)

	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		err := errors.Join(ErrTransactionContext, ErrMustPointer)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key)
		return err
	}

	if data, err = tc.db.GetSagaValueByExecutionID(tc.executionID, key); err != nil {
		err := errors.Join(ErrSagaGetValue, err)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key, "error", err)
		return err
	}

	if err := rtl.Decode(bytes.NewBuffer(data), value); err != nil {
		err := errors.Join(ErrTransactionContext, ErrEncoding, err)
		logger.Error(tc.ctx, err.Error(), "transaction_context.key", key, "error", err)
		return err
	}

	logger.Debug(tc.ctx, "transaction loaded value", "transaction_context.key", key, "transaction_context.value", value, "transaction_context.execution_id", tc.executionID, "transaction_context.step_index", tc.stepIndex, "transaction_context.context_id", tc.contextID)

	return nil
}

type CompensationContext struct {
	ctx         context.Context
	db          Database
	contextID   int
	executionID int
	stepIndex   int
}

var ErrCompensationContext = errors.New("failed compensation context")

func (cc *CompensationContext) Load(key string, value interface{}) error {
	var data []byte
	var err error

	logger.Debug(cc.ctx, "compensation loading value", "compensation_context.key", key, "compensation_context.execution_id", cc.executionID, "compensation_context.step_index", cc.stepIndex, "compensation_context.context_id", cc.contextID)

	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		err := errors.Join(ErrCompensationContext, ErrMustPointer)
		logger.Error(cc.ctx, err.Error(), "compensation_context.key", key)
		return err
	}

	if data, err = cc.db.GetSagaValueByExecutionID(cc.executionID, key); err != nil {
		err := errors.Join(ErrSagaGetValue, err)
		logger.Error(cc.ctx, err.Error(), "compensation_context.key", key, "error", err)
		return err
	}

	if err := rtl.Decode(bytes.NewBuffer(data), value); err != nil {
		err := errors.Join(ErrCompensationContext, ErrEncoding, err)
		logger.Error(cc.ctx, err.Error(), "compensation_context.key", key, "error", err)
		return err
	}

	logger.Debug(cc.ctx, "compensation loaded value", "compensation_context.key", key, "compensation_context.value", value, "compensation_context.execution_id", cc.executionID, "compensation_context.step_index", cc.stepIndex, "compensation_context.context_id", cc.contextID)

	return nil
}

// GetStepIndex returns the current step index
func (tc *TransactionContext) GetStepIndex() int {
	return tc.stepIndex
}

// GetStepIndex returns the current step index
func (cc *CompensationContext) GetStepIndex() int {
	return cc.stepIndex
}

type SagaStep interface {
	Transaction(ctx TransactionContext) error
	Compensation(ctx CompensationContext) error
}

type FunctionStep struct {
	transactionFn  func(ctx TransactionContext) error
	compensationFn func(ctx CompensationContext) error
}

func (s *FunctionStep) Transaction(ctx TransactionContext) error {
	return s.transactionFn(ctx)
}

func (s *FunctionStep) Compensation(ctx CompensationContext) error {
	return s.compensationFn(ctx)
}

type SagaDefinition struct {
	Steps       []SagaStep
	HandlerInfo *SagaHandlerInfo
}

type SagaDefinitionBuilder struct {
	steps []SagaStep
}

type SagaHandlerInfo struct {
	TransactionInfo  []HandlerInfo
	CompensationInfo []HandlerInfo
}

// NewSaga creates a new builder instance.
func NewSaga() *SagaDefinitionBuilder {
	return &SagaDefinitionBuilder{
		steps: make([]SagaStep, 0),
	}
}

func (b *SagaDefinitionBuilder) Add(transaction func(TransactionContext) error, compensation func(CompensationContext) error) *SagaDefinitionBuilder {
	b.steps = append(b.steps, &FunctionStep{
		transactionFn:  transaction,
		compensationFn: compensation,
	})
	return b
}

// AddStep adds a saga step to the builder.
func (b *SagaDefinitionBuilder) AddStep(step SagaStep) *SagaDefinitionBuilder {
	b.steps = append(b.steps, step)
	return b
}

var ErrTransactionMethodNotFound = errors.New("transaction method not found")
var ErrCompensationMethodNotFound = errors.New("compensation method not found")

// Build creates a SagaDefinition with the HandlerInfo included.
func (b *SagaDefinitionBuilder) Build() (*SagaDefinition, error) {

	logger.Debug(context.Background(), "building saga definition", "saga_builder.steps", len(b.steps))

	sagaInfo := &SagaHandlerInfo{
		TransactionInfo:  make([]HandlerInfo, len(b.steps)),
		CompensationInfo: make([]HandlerInfo, len(b.steps)),
	}

	for i, step := range b.steps {
		stepType := reflect.TypeOf(step)
		originalType := stepType // Keep original type for handler name
		isPtr := stepType.Kind() == reflect.Ptr

		// Get the base type for method lookup
		if isPtr {
			stepType = stepType.Elem()
		}

		// Try to find methods on both pointer and value receivers
		var transactionMethod, compensationMethod reflect.Method
		var transactionOk, compensationOk bool

		// First try the original type (whether pointer or value)
		if transactionMethod, transactionOk = originalType.MethodByName("Transaction"); !transactionOk {
			// If not found and original wasn't a pointer, try pointer
			if !isPtr {
				if ptrMethod, ok := reflect.PtrTo(stepType).MethodByName("Transaction"); ok {
					transactionMethod = ptrMethod
					transactionOk = true
				}
			}
		}

		if compensationMethod, compensationOk = originalType.MethodByName("Compensation"); !compensationOk {
			// If not found and original wasn't a pointer, try pointer
			if !isPtr {
				if ptrMethod, ok := reflect.PtrTo(stepType).MethodByName("Compensation"); ok {
					compensationMethod = ptrMethod
					compensationOk = true
				}
			}
		}

		if !transactionOk {
			err := errors.Join(ErrTransactionMethodNotFound, fmt.Errorf("not found for step %d", i))
			logger.Error(context.Background(), err.Error(), "saga_builder.step", i)
			return nil, err
		}
		if !compensationOk {
			err := errors.Join(ErrCompensationMethodNotFound, fmt.Errorf("not found for step %d", i))
			logger.Error(context.Background(), err.Error(), "saga_builder.step", i)
			return nil, err
		}

		// Use the actual type name for the handler
		typeName := stepType.Name()
		if isPtr {
			typeName = "*" + typeName
		}

		// Note: Since Transaction/Compensation now only return error, we modify the analysis
		transactionInfo := HandlerInfo{
			HandlerName:     typeName + ".Transaction",
			HandlerLongName: HandlerIdentity(typeName),
			Handler:         transactionMethod.Func.Interface(),
			ParamTypes:      []reflect.Type{reflect.TypeOf(TransactionContext{})},
			ReturnTypes:     []reflect.Type{}, // No return types except error
			NumIn:           1,
			NumOut:          0,
		}

		compensationInfo := HandlerInfo{
			HandlerName:     typeName + ".Compensation",
			HandlerLongName: HandlerIdentity(typeName),
			Handler:         compensationMethod.Func.Interface(),
			ParamTypes:      []reflect.Type{reflect.TypeOf(CompensationContext{})},
			ReturnTypes:     []reflect.Type{}, // No return types except error
			NumIn:           1,
			NumOut:          0,
		}

		sagaInfo.TransactionInfo[i] = transactionInfo
		sagaInfo.CompensationInfo[i] = compensationInfo
	}

	logger.Debug(context.Background(), "built saga definition", "saga_builder.steps", len(b.steps), "saga_builder.transaction_info", len(sagaInfo.TransactionInfo), "saga_builder.compensation_info", len(sagaInfo.CompensationInfo))

	return &SagaDefinition{
		Steps:       b.steps,
		HandlerInfo: sagaInfo,
	}, nil
}

// ActivityContext provides context for activity execution.
type ActivityContext struct {
	ctx context.Context
}

func (ac ActivityContext) Done() <-chan struct{} {
	return ac.ctx.Done()
}

func (ac ActivityContext) Err() error {
	return ac.ctx.Err()
}

// WorkflowContext provides context for workflow execution.
// TODO: add a function to know how much was i retried!
type WorkflowContext struct {
	db                  Database
	registry            *Registry
	ctx                 context.Context
	state               StateTracker
	tracker             InstanceTracker
	debug               Debug
	runID               int
	queueID             int
	stepID              string
	workflowID          int
	workflowExecutionID int
	options             WorkflowOptions

	onCrossQueueWorkflow crossWorkflow
	onContinueAsNew      continueAsNew
	onSignalNew          signalNew
}

func (ctx WorkflowContext) checkPause() error {
	if ctx.state.isPaused() {
		logger.Debug(ctx.ctx, "workflow is paused", "workflow_context.workflow_id", ctx.workflowID, "workflow_context.execution_id", ctx.workflowExecutionID, "workflow_context.step_id", ctx.stepID, "workflow_context.queue_id", ctx.queueID, "workflow_context.run_id", ctx.runID)
		return ErrPaused
	}
	return nil
}

// ContinueAsNew allows a workflow to continue as new with the given function and arguments.
// It also mean the workflow is successful and will be marked as completed.
func (ctx WorkflowContext) ContinueAsNew(options *WorkflowOptions, args ...interface{}) error {

	if options == nil {
		options = &WorkflowOptions{}
	}

	if options.Queue == "" {
		options.Queue = ctx.options.Queue
	}

	if options.RetryPolicy == nil {
		options.RetryPolicy = DefaultRetryPolicy()
	}

	logger.Debug(ctx.ctx, "workflow continuing as new", "workflow_context.workflow_id", ctx.workflowID, "workflow_context.execution_id", ctx.workflowExecutionID, "workflow_context.step_id", ctx.stepID, "workflow_context.queue_id", ctx.queueID, "workflow_context.run_id", ctx.runID)

	return &ContinueAsNewError{
		Options: options,
		Args:    args,
	}
}

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	ctx      context.Context
	mu       deadlock.Mutex
	db       Database
	tracker  InstanceTracker
	state    StateTracker
	registry *Registry
	debug    Debug

	handler HandlerInfo

	fsm *stateless.StateMachine

	options WorkflowOptions
	future  Future

	stepID      string
	runID       int
	workflowID  int
	executionID int
	dataID      int
	queueID     int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string

	onCrossWorkflow crossWorkflow
	onContinueAsNew continueAsNew
	onSignalNew     signalNew
}

// Future implementation of direct calling
type RuntimeFuture struct {
	mu         sync.Mutex
	results    []interface{}
	err        error
	done       chan struct{}
	workflowID int
	once       sync.Once
	continueAs int
}

func (f *RuntimeFuture) WorkflowID() int {
	return f.workflowID
}

func NewRuntimeFuture() Future {
	return &RuntimeFuture{
		done: make(chan struct{}),
	}
}

func (f *RuntimeFuture) IsPaused() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if errors.Is(f.err, ErrPaused) {
		return true
	}
	return false
}

func (f *RuntimeFuture) setEntityID(entityID int) {
	f.mu.Lock()
	f.workflowID = entityID
	logger.Debug(context.Background(), "future set entity id", "future.entity_id", entityID, "future.workflow_id", f.workflowID)
	f.mu.Unlock()
}

func (f *RuntimeFuture) setResult(results []interface{}) {
	f.mu.Lock()
	f.results = results
	logger.Debug(context.Background(), "future set results", "future.results", len(results), "future.workflow_id", f.workflowID)
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

func (f *RuntimeFuture) setError(err error) {
	f.mu.Lock()
	f.err = err
	logger.Debug(context.Background(), "future set error", "future.error", err, "future.workflow_id", f.workflowID)
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

func (f *RuntimeFuture) GetResults() ([]interface{}, error) {
	logger.Debug(context.Background(), "future wait results", "future.workflow_id", f.workflowID)

	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()

	logger.Debug(context.Background(), "future get results", "future.results", len(f.results), "future.workflow_id", f.workflowID)

	if f.err != nil {
		logger.Debug(context.Background(), "future get results error", "future.error", f.err, "future.workflow_id", f.workflowID)
		return nil, f.err
	}

	// Return a copy of results to prevent modification
	results := make([]interface{}, len(f.results))
	copy(results, f.results)

	return results, nil
}

var ErrGetResults = errors.New("cannot get results")

func (f *RuntimeFuture) Get(out ...interface{}) error {

	logger.Debug(context.Background(), "future wait get", "future.workflow_id", f.workflowID)

	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		logger.Debug(context.Background(), "future get error", "future.error", f.err, "future.workflow_id", f.workflowID)
		return f.err
	}

	if len(out) == 0 {
		logger.Debug(context.Background(), "future get no output", "future.workflow_id", f.workflowID)
		return nil
	}

	if len(out) > len(f.results) {
		err := errors.Join(ErrGetResults, fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results)))
		logger.Error(context.Background(), err.Error(), "future.workflow_id", f.workflowID)
		return err
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			err := errors.Join(ErrGetResults, ErrMustPointer)
			logger.Error(context.Background(), err.Error(), "future.workflow_id", f.workflowID)
			return err
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			err := errors.Join(ErrGetResults, fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i))
			logger.Error(context.Background(), err.Error(), "future.workflow_id", f.workflowID)
			return err
		}

		val.Set(result)
	}

	return nil
}

type StateTracker interface {
	isPaused() bool
	// isActive() bool
	setUnpause()
	// setActive()
	// setInactive()
}

type Debug interface {
	canStackTrace() bool
}

type InstanceTracker interface {
	onPause() error
	addWorkflowInstance(wi *WorkflowInstance)
	addActivityInstance(ai *ActivityInstance)
	addSideEffectInstance(sei *SideEffectInstance)
	addSagaInstance(si *SagaInstance)
	newRoot(root *WorkflowInstance)
	reset()
	findSagaInstance(stepID string) (*SagaInstance, error)
}

// Orchestrator orchestrates the execution of workflows and activities.
type Orchestrator struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu            deadlock.Mutex
	instancesMu   deadlock.RWMutex
	activitiesMu  deadlock.Mutex
	sideEffectsMu deadlock.Mutex
	sagasMu       deadlock.Mutex

	registry *Registry
	db       Database

	runID int

	rootWf *WorkflowInstance // root workflow instance

	instances   []*WorkflowInstance
	activities  []*ActivityInstance
	sideEffects []*SideEffectInstance
	sagas       []*SagaInstance

	err error

	paused   bool
	pausedMu deadlock.Mutex

	// active atomic.Bool

	displayStackTrace bool

	onCrossWorkflow crossWorkflow
	onContinueAsNew continueAsNew
	onSignalNew     signalNew
}

type orchestratorConfig struct {
	onCrossWorkflow crossWorkflow
	onContinueAsNew continueAsNew
	onSignalNew     signalNew
}

type OrchestratorOption func(*orchestratorConfig)

func WithCrossWorkflow(fn crossWorkflow) OrchestratorOption {
	return func(c *orchestratorConfig) {
		c.onCrossWorkflow = fn
	}
}

func WithContinueAsNew(fn continueAsNew) OrchestratorOption {
	return func(c *orchestratorConfig) {
		c.onContinueAsNew = fn
	}
}

func WithSignalNew(fn signalNew) OrchestratorOption {
	return func(c *orchestratorConfig) {
		c.onSignalNew = fn
	}
}

func NewOrchestrator(ctx context.Context, db Database, registry *Registry, opt ...OrchestratorOption) *Orchestrator {

	ctx, cancel := context.WithCancel(ctx)
	o := &Orchestrator{
		db:       db,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,

		displayStackTrace: true,
	}

	cfg := &orchestratorConfig{}
	for _, o := range opt {
		o(cfg)
	}

	if cfg.onCrossWorkflow != nil {
		o.onCrossWorkflow = cfg.onCrossWorkflow
	}
	if cfg.onContinueAsNew != nil {
		o.onContinueAsNew = cfg.onContinueAsNew
	}
	if cfg.onSignalNew != nil {
		o.onSignalNew = cfg.onSignalNew
	}

	return o
}

// func (o *Orchestrator) setActive() {
// 	o.active.Store(true)
// }

// func (o *Orchestrator) setInactive() {
// 	o.active.Store(false)
// }

func (o *Orchestrator) findSagaInstance(stepID string) (*SagaInstance, error) {
	logger.Debug(o.ctx, "orchestrator find saga instance", "orchestrator.find_saga_instance.step_id", stepID)
	for _, saga := range o.sagas {
		if saga.stepID == stepID {
			return saga, nil
		}
	}
	err := errors.Join(ErrOrchestrator, fmt.Errorf("not found step %s", stepID))
	logger.Error(o.ctx, err.Error(), "orchestrator.find_saga_instance.step_id", stepID)
	return nil, err
}

func (o *Orchestrator) reset() {
	o.rootWf = nil
	o.instances = []*WorkflowInstance{}
	o.activities = []*ActivityInstance{}
	o.sideEffects = []*SideEffectInstance{}
	o.sagas = []*SagaInstance{}
}

// Request to pause the current orchestration
// Pause != Cancel
// Pause is controlled mechanism to stop the execution of the current orchestration by slowly stopping the execution of the current workflow.
func (o *Orchestrator) Pause() {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = true

}

// Cancel the orchestrator will create a cascade of cancellation to all the instances.
// Pause != Cancel
func (o *Orchestrator) Cancel() {
	o.cancel()
}

var ErrOrchestrator = errors.New("failed orchestrator")

func (o *Orchestrator) GetStatus(id int) (EntityStatus, error) {

	logger.Debug(o.ctx, "orchestrator get status", "orchestrator.get_status.workflow_id", id)

	o.mu.Lock()
	var status EntityStatus
	if err := o.db.GetWorkflowEntityProperties(id, GetWorkflowEntityStatus(&status)); err != nil {
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get workflow entity properties: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.get_status.workflow_id", id)
		return "", err
	}
	o.mu.Unlock()
	return status, nil
}

func (o *Orchestrator) WaitFor(id int, status EntityStatus) error {

	logger.Debug(o.ctx, "orchestrator wait for", "orchestrator.wait_for.workflow_id", id, "orchestrator.wait_for.status", status)

	for {
		s, err := o.GetStatus(id)
		if err != nil {
			err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get status: %w", err))
			logger.Error(o.ctx, err.Error(), "orchestrator.wait_for.workflow_id", id)
			return err
		}
		if s == status {
			return nil
		}
		<-time.After(100 * time.Millisecond) // TODO: should be configurable
	}
}

// func (o *Orchestrator) WaitActive() {
// for o.active.Load() { // it is still active, you have to wait
// 	<-time.After(100 * time.Millisecond)
// }
// }

func (o *Orchestrator) isPaused() bool {
	logger.Debug(o.ctx, "orchestrator is paused", "orchestrator.is_paused", o.paused)
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	return o.paused
}

func (o *Orchestrator) setUnpause() {
	logger.Debug(o.ctx, "orchestrator set unpause")
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = false
}

// func (o *Orchestrator) isActive() bool {
// 	return o.active.Load()
// }

func (o *Orchestrator) canStackTrace() bool {
	logger.Debug(o.ctx, "orchestrator can stack trace", "orchestrator.can_stack_trace", o.displayStackTrace)
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.displayStackTrace
}

var ErrWorkflowPaused = errors.New("workflow is paused")
var ErrRegistryHandlerNotFound = errors.New("handler not found")
var ErrSerialization = errors.New("failed to serialize")

func (o *Orchestrator) Resume(entityID int) Future {
	// TODO: add timeout
	// for o.active.Load() { // it is still active, you have to wait
	// 	<-time.After(100 * time.Millisecond)
	// }

	logger.Debug(o.ctx, "orchestrator resume", "orchestrator.resume.workflow_id", entityID)

	var err error

	// resume, heh? was it paused?
	var status EntityStatus
	if err := o.db.GetWorkflowEntityProperties(entityID, GetWorkflowEntityStatus(&status)); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get workflow entity properties: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	if status != StatusPaused {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, ErrPaused, ErrWorkflowPaused, fmt.Errorf("workflow is not paused: %s", status))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	var workflowEntity *WorkflowEntity
	if workflowEntity, err = o.db.
		GetWorkflowEntity(
			entityID,
			WorkflowEntityWithData(),
			WorkflowEntityWithQueue(),
		); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get workflow entity: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	var latestExecution *WorkflowExecution
	if latestExecution, err = o.db.GetWorkflowExecutionLatestByEntityID(entityID); err != nil {
		future := NewRuntimeFuture()
		future.setError(fmt.Errorf("failed to get latest execution: %w", err))
		return future
	}

	// Get parent execution if this is a sub-workflow
	var parentExecID, parentEntityID int
	var parentStepID string
	var hierarchies []*Hierarchy
	if hierarchies, err = o.db.GetHierarchiesByChildEntity(entityID); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get hierarchies: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	if len(hierarchies) > 0 {
		parentExecID = hierarchies[0].ParentExecutionID
		parentEntityID = hierarchies[0].ParentEntityID
		parentStepID = hierarchies[0].ParentStepID
	}

	handler, ok := o.registry.GetWorkflow(workflowEntity.HandlerName)
	if !ok {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, ErrRegistryHandlerNotFound, fmt.Errorf("handler %s not found", workflowEntity.HandlerName))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	inputs, err := convertInputsFromSerialization(handler, workflowEntity.WorkflowData.Inputs)
	if err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestrator, ErrSerialization, fmt.Errorf("failed to convert inputs: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	// now we need to re-create a new WorkflowInstance for the execution we left behind

	instance := &WorkflowInstance{
		ctx:      o.ctx, // share context with orchestrator
		db:       o.db,  // share db
		tracker:  o,
		state:    o,
		registry: o.registry,
		debug:    o,

		handler: handler,

		runID:       workflowEntity.RunID,
		workflowID:  entityID,
		stepID:      workflowEntity.StepID,
		executionID: latestExecution.ID,
		dataID:      workflowEntity.WorkflowData.ID,
		queueID:     workflowEntity.QueueID,
		options: WorkflowOptions{
			RetryPolicy: &RetryPolicy{
				MaxAttempts: workflowEntity.RetryPolicy.MaxAttempts,
				MaxInterval: time.Duration(workflowEntity.RetryPolicy.MaxInterval),
			},
			Queue: workflowEntity.Edges.Queue.Name,
		},

		parentExecutionID: parentExecID,
		parentEntityID:    parentEntityID,
		parentStepID:      parentStepID,

		onCrossWorkflow: o.onCrossWorkflow,
		onContinueAsNew: o.onContinueAsNew,
		onSignalNew:     o.onSignalNew,
	}

	// Create Future and start instance
	future := NewRuntimeFuture()
	future.setEntityID(workflowEntity.ID)
	instance.future = future

	// Very important to notice the orchestrator of the real root workflow
	o.rootWf = instance

	// if !o.isActive() {
	// 	future.setError(fmt.Errorf("orchestrator is stopping"))
	// 	return future
	// }

	o.setUnpause()
	o.reset()

	o.addWorkflowInstance(instance)

	// Not paused anymore
	if err = o.db.SetWorkflowDataProperties(
		entityID,
		SetWorkflowDataPaused(false),
		SetWorkflowDataResumable(false),
	); err != nil {
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to set workflow data properties: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	// Not paused anymore
	if err = o.db.SetWorkflowEntityProperties(
		entityID,
		SetWorkflowEntityStatus(StatusRunning),
	); err != nil {
		err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to set workflow entity status: %w", err))
		logger.Error(o.ctx, err.Error(), "orchestrator.resume.workflow_id", entityID)
		future.setError(err)
		return future
	}

	logger.Debug(o.ctx, "orchestrator resumed", "orchestrator.resumed.workflow_id", entityID)

	go instance.Start(inputs)

	return future
}

type preparationOptions struct {
	// Mainly for continuation
	parentRunID               int
	parentWorkflowID          int
	parentWorkflowExecutionID int
}

var ErrRegistryRegisteration = errors.New("failed to register workflow")

var ErrPreparation = errors.New("failed to prepare workflow")

// Preparing the creation of a new root workflow instance so it can exists in the database, we might decide depending of which systems of used if we want to pull the workflows or execute it directly.
func prepareWorkflow(registry *Registry, db Database, workflowFunc interface{}, workflowOptions *WorkflowOptions, opts *preparationOptions, args ...interface{}) (*WorkflowEntity, error) {

	logger.Debug(context.Background(), "preparing workflow", "workflow_func", getFunctionName(workflowFunc))

	// Register workflow if needed
	handler, err := registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		err := errors.Join(ErrPreparation, ErrRegistryRegisteration, fmt.Errorf("failed to register workflow: %w", err))
		logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc))
		return nil, err
	}

	isRoot := true

	// Convert inputs for storage
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		err := errors.Join(ErrPreparation, ErrSerialization, fmt.Errorf("failed to serialize inputs: %w", err))
		logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc))
		return nil, err
	}

	opttionsPreparation := &preparationOptions{}
	if opts != nil {
		if opts.parentRunID != 0 {
			opttionsPreparation.parentRunID = opts.parentRunID
			isRoot = false
		}
		if opts.parentWorkflowID != 0 {
			opttionsPreparation.parentWorkflowID = opts.parentWorkflowID
			isRoot = false
		}
		if opts.parentWorkflowExecutionID != 0 {
			opttionsPreparation.parentWorkflowExecutionID = opts.parentWorkflowExecutionID
			isRoot = false
		}
	}

	var continuedID int
	var continuedExecutionID int

	if opttionsPreparation.parentWorkflowExecutionID != 0 {

		continuedID = opttionsPreparation.parentWorkflowID
	}
	if opttionsPreparation.parentWorkflowID != 0 {

		continuedExecutionID = opttionsPreparation.parentWorkflowExecutionID
	}

	var runID int
	if opttionsPreparation.parentRunID == 0 {
		logger.Debug(context.Background(), "preparing workflow got no parentRunID", "workflow_func", getFunctionName(workflowFunc))
		// Create or get Run
		if runID, err = db.AddRun(&Run{
			Status: RunStatusPending,
		}); err != nil {
			err := errors.Join(ErrPreparation, fmt.Errorf("failed to create run: %w", err))
			logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc))
			return nil, err
		}
	} else {
		logger.Debug(context.Background(), "preparing workflow got parentRunID", "workflow_func", getFunctionName(workflowFunc), "parentRunID", opttionsPreparation.parentRunID)
		runID = opttionsPreparation.parentRunID
	}

	// Convert retry policy if provided
	var retryPolicy *retryPolicyInternal
	if workflowOptions != nil && workflowOptions.RetryPolicy != nil {
		logger.Debug(context.Background(), "preparing workflow got retry policy", "workflow_func", getFunctionName(workflowFunc), "retry_policy.max_attempts", workflowOptions.RetryPolicy.MaxAttempts, "retry_policy.max_interval", workflowOptions.RetryPolicy.MaxInterval)
		rp := workflowOptions.RetryPolicy
		retryPolicy = &retryPolicyInternal{
			MaxAttempts: rp.MaxAttempts,
			MaxInterval: rp.MaxInterval.Nanoseconds(),
		}
	} else {
		logger.Debug(context.Background(), "preparing workflow got no retry policy (default)", "workflow_func", getFunctionName(workflowFunc))
		// Default retry policy
		retryPolicy = DefaultRetryPolicyInternal()
	}

	var queueName string = DefaultQueue
	if workflowOptions != nil && workflowOptions.Queue != "" {
		logger.Debug(context.Background(), "preparing workflow got queue name", "workflow_func", getFunctionName(workflowFunc), "queue_name", workflowOptions.Queue)
		queueName = workflowOptions.Queue
	}

	var queue *Queue
	if queue, err = db.GetQueueByName(queueName); err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to get queue %s: %w", queueName, err))
		logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc), "queue_name", queueName)
		return nil, err
	}

	data := &WorkflowData{
		Inputs:    inputBytes,
		Paused:    false,
		Resumable: false,
		IsRoot:    isRoot,
	}

	if continuedID != 0 {
		data.ContinuedFrom = &continuedID
	}
	if continuedExecutionID != 0 {
		data.ContinuedExecutionFrom = &continuedExecutionID
	}

	var workflowID int
	// We're not even storing the HandlerInfo since it is a runtime thing
	if workflowID, err = db.
		AddWorkflowEntity(
			&WorkflowEntity{
				BaseEntity: BaseEntity{
					HandlerName: handler.HandlerName,
					Status:      StatusPending,
					Type:        EntityWorkflow,
					QueueID:     queue.ID,
					StepID:      "root",
					RunID:       runID,
					RetryPolicy: *retryPolicy,
					RetryState:  RetryState{Attempts: 0},
				},
				WorkflowData: data,
			}); err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to add workflow entity: %w", err))
		logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc), "queue_name", queueName)
		return nil, err
	}

	var entity *WorkflowEntity
	if entity, err = db.GetWorkflowEntity(workflowID); err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to get workflow entity: %w", err))
		logger.Error(context.Background(), err.Error(), "workflow_func", getFunctionName(workflowFunc), "queue_name", queueName)
		return nil, err
	}

	return entity, nil
}

var ErrOrchestratorExecution = errors.New("failed to execute workflow")

// Execute starts the execution of a workflow directly.
func (o *Orchestrator) Execute(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
	var entity *WorkflowEntity
	var err error
	logger.Debug(o.ctx, "orchestrator execute", "workflow_func", getFunctionName(workflowFunc))

	// Create entity and related records
	if entity, err = prepareWorkflow(o.registry, o.db, workflowFunc, options, nil, args...); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrOrchestratorExecution, err)
		logger.Error(o.ctx, err.Error(), "workflow_func", getFunctionName(workflowFunc))
		future.setError(err)
		return future
	}

	// Execute using the entity
	future, err := o.ExecuteWithEntity(entity.ID)
	if err != nil {
		f := NewRuntimeFuture()
		f.setEntityID(entity.ID)
		err := errors.Join(ErrOrchestratorExecution, err, fmt.Errorf("failed to execute workflow with entity %d", entity.ID))
		logger.Error(o.ctx, err.Error(), "workflow_func", getFunctionName(workflowFunc), "entity_id", entity.ID)
		f.setError(err)
		return f
	}

	logger.Debug(o.ctx, "orchestrator executed", "workflow_func", getFunctionName(workflowFunc), "entity_id", entity.ID)

	return future
}

var ErrOrchestratorExecuteEntity = errors.New("failed to execute workflow with entity")

// ExecuteWithEntity starts a workflow using an existing entity ID
func (o *Orchestrator) ExecuteWithEntity(entityID int) (Future, error) {
	var err error

	logger.Debug(o.ctx, "orchestrator execute with entity", "entity_id", entityID)

	// Get the entity and verify it exists
	var entity *WorkflowEntity
	if entity, err = o.db.GetWorkflowEntity(
		entityID,
		WorkflowEntityWithQueue(), // we need the edge to the queue
		WorkflowEntityWithData(),
	); err != nil {
		err := errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get workflow entity: %w", err))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	var isRoot bool
	if err = o.db.GetWorkflowDataProperties(entityID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		err := errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get workflow data properties: %w", err))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	// Set orchestrator's runID from entity
	o.runID = entity.RunID

	// Verify it's a workflow
	if entity.Type != EntityWorkflow {
		err := errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("entity %d is not a workflow", entityID))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	var ok bool
	var handlerInfo HandlerInfo
	if handlerInfo, ok = o.registry.GetWorkflow(entity.HandlerName); !ok {
		err := errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get handler info"))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	// Convert inputs
	inputs, err := convertInputsFromSerialization(handlerInfo, entity.WorkflowData.Inputs)
	if err != nil {
		err := errors.Join(ErrOrchestratorExecuteEntity, ErrSerialization, fmt.Errorf("failed to convert inputs: %w", err))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	// Create workflow instance
	instance := &WorkflowInstance{
		ctx:      o.ctx, // share context with orchestrator
		db:       o.db,  // share db
		tracker:  o,
		state:    o,
		registry: o.registry,
		debug:    o,

		// inputs: inputs,
		stepID: entity.StepID,

		runID:      o.runID,
		workflowID: entity.ID,
		// executionID: , // will be filled at first attempt
		dataID:  entity.WorkflowData.ID,
		queueID: entity.QueueID,

		handler: handlerInfo,

		// rebuild the workflow options based on the database
		options: WorkflowOptions{
			RetryPolicy: &RetryPolicy{
				MaxAttempts: entity.RetryPolicy.MaxAttempts,
				MaxInterval: time.Duration(entity.RetryPolicy.MaxInterval),
			},
			Queue: entity.Edges.Queue.Name,
			// VersionOverrides: , // TODO: implement in WorkflowOption and save it
		},

		onCrossWorkflow: o.onCrossWorkflow,
		onContinueAsNew: o.onContinueAsNew,
		onSignalNew:     o.onSignalNew,
	}

	// If this is a sub-workflow, set parent info from hierarchies
	var hierarchies []*Hierarchy
	if hierarchies, err = o.db.GetHierarchiesByChildEntity(entityID); err != nil {
		err := errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get hierarchies: %w", err))
		logger.Error(o.ctx, err.Error(), "entity_id", entityID)
		return nil, err
	}

	// If we have parents only
	if len(hierarchies) > 0 {
		h := hierarchies[0]
		instance.parentExecutionID = h.ParentExecutionID
		instance.parentEntityID = h.ParentEntityID
		instance.parentStepID = h.ParentStepID
	}

	// Create Future and start instance
	future := NewRuntimeFuture()
	future.setEntityID(entity.ID)
	instance.future = future

	// Very important to notice the orchestrator of the real root workflow
	o.rootWf = instance

	// if !o.isActive() {
	// 	future.setError(errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("orchestrator is already active")))
	// 	return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("orchestrator is already active"))
	// }

	// o.setActive()

	o.addWorkflowInstance(instance)

	logger.Debug(o.ctx, "orchestrator execute with entity", "entity_id", entityID)
	// We don't start the FSM within the same goroutine as the caller
	go instance.Start(inputs) // root workflow OR ContinueAsNew starts in a goroutine

	return future, nil
}

func (o *Orchestrator) addWorkflowInstance(wi *WorkflowInstance) {
	o.instancesMu.Lock()
	o.instances = append(o.instances, wi)
	o.instancesMu.Unlock()
}

func (o *Orchestrator) addActivityInstance(wi *ActivityInstance) {
	o.activitiesMu.Lock()
	o.activities = append(o.activities, wi)
	o.activitiesMu.Unlock()
}

func (o *Orchestrator) addSideEffectInstance(wi *SideEffectInstance) {
	o.sideEffectsMu.Lock()
	o.sideEffects = append(o.sideEffects, wi)
	o.sideEffectsMu.Unlock()
}

func (o *Orchestrator) addSagaInstance(wi *SagaInstance) {
	o.sagasMu.Lock()
	o.sagas = append(o.sagas, wi)
	o.sagasMu.Unlock()
}

// onPause will check all the instances and pause them
func (o *Orchestrator) onPause() error {

	logger.Debug(o.ctx, "orchestrator on pause")
	var err error

	for _, workflow := range o.instances {

		// TODO: pause: we should call the orchestrator to pause everything is running or pending

		var status EntityStatus
		if err = o.db.GetWorkflowEntityProperties(workflow.workflowID, GetWorkflowEntityStatus(&status)); err != nil {
			err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to get workflow entity status: %w", err))
			logger.Error(o.ctx, err.Error(), "workflow_id", workflow.workflowID)
			return err
		}

		switch status {
		case StatusRunning, StatusPending:
			if err = o.db.
				SetWorkflowEntityProperties(
					workflow.workflowID,
					SetWorkflowEntityStatus(StatusPaused),
				); err != nil {
				err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to set workflow entity status: %w", err))
				logger.Error(o.ctx, err.Error(), "workflow_id", workflow.workflowID)
				return err
			}

			if err = o.db.
				SetWorkflowDataProperties(
					workflow.workflowID,
					SetWorkflowDataPaused(true),
					SetWorkflowDataResumable(false),
				); err != nil {
				err := errors.Join(ErrOrchestrator, fmt.Errorf("failed to set workflow data paused: %w", err))
				logger.Error(o.ctx, err.Error(), "workflow_id", workflow.workflowID)
				return err
			}
		}
	}

	/// Those comments are temporary, we can't pause activities, side effects or sagas

	// for _, activity := range o.activities {

	// 	var status EntityStatus

	// 	if err = o.db.GetActivityEntityProperties(activity.entityID, GetActivityEntityStatus(&status)); err != nil {
	// 		return fmt.Errorf("failed to get activity entity status: %w", err)
	// 	}

	// 	// If running, we will let it completed and end
	// 	// If failed, we will let it failed
	// 	// If pending, we will pause its
	// 	switch status {
	// 	case StatusPending:

	// 		var execStatus ExecutionStatus

	// 		if err = o.db.GetActivityExecutionProperties(activity.executionID, GetActivityExecutionStatus(&execStatus)); err != nil {
	// 			return fmt.Errorf("failed to get activity execution status: %w", err)
	// 		}

	// 		if err = o.db.
	// 			SetActivityEntityProperties(
	// 				activity.entityID,
	// 				SetActivityEntityStatus(StatusPaused),
	// 			); err != nil {
	// 			return fmt.Errorf("failed to set activity entity status: %w", err)
	// 		}
	// 	}
	// }

	// for _, sideEffect := range o.sideEffects {

	// 	var status EntityStatus

	// 	if err = o.db.GetSideEffectEntityProperties(sideEffect.entityID, GetSideEffectEntityStatus(&status)); err != nil {
	// 		return fmt.Errorf("failed to get side effect entity status: %w", err)
	// 	}

	// 	// If running, we will let it completed and end
	// 	// If failed, we will let it failed
	// 	// If pending, we will pause its
	// 	switch status {
	// 	case StatusPending:
	// 		if err = o.db.
	// 			SetSideEffectEntityProperties(
	// 				sideEffect.entityID,
	// 				SetSideEffectEntityStatus(StatusPaused),
	// 			); err != nil {
	// 			return fmt.Errorf("failed to set side effect entity status: %w", err)
	// 		}
	// 	}

	// }

	// for _, saga := range o.sagas {
	// 	if err = o.db.
	// 		SetSagaEntityProperties(
	// 			saga.workflowID,
	// 			SetSagaEntityStatus(StatusPaused),
	// 		); err != nil {
	// 		return fmt.Errorf("failed to set saga entity status: %w", err)
	// 	}
	// }

	// destroy the current instances
	o.reset()

	return nil
}

func (o *Orchestrator) newRoot(root *WorkflowInstance) {
	o.rootWf = root
}

////////////////////////// WorkflowInstance

var ErrWorkflowInstance = errors.New("failed workflow instance")

// Starting the workflow instance
// Ideally we're already within a goroutine
func (wi *WorkflowInstance) Start(inputs []interface{}) error {

	logger.Debug(wi.ctx, "workflow instance start", "workflow_id", wi.workflowID)

	wi.mu.Lock()
	wi.fsm = stateless.NewStateMachine(StateIdle)
	fsm := wi.fsm
	wi.mu.Unlock()

	fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateExecuting).
		OnEntry(wi.executeWorkflow).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateCompleted).
		OnEntry(wi.onCompleted)

	fsm.Configure(StateFailed).
		OnEntry(wi.onFailed)

	fsm.Configure(StatePaused).
		OnEntry(wi.onPaused).
		Permit(TriggerResume, StateExecuting)

	logger.Debug(wi.ctx, "workflow instance fsm configured", "workflow_id", wi.workflowID)
	// Start the FSM without holding wi.mu
	if err := fsm.Fire(TriggerStart, inputs); err != nil {
		err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to start FSM: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		if !errors.Is(err, ErrPaused) {
			wi.future.setError(err)
		}
	}

	var isRoot bool
	if err := wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		err := errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to get workflow data properties: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	// free up the orchestrator
	if isRoot {
		logger.Debug(wi.ctx, "workflow instance is root", "workflow_id", wi.workflowID)
		// // set to inactive
		// if wi.state.isActive() {
		// 	wi.state.setInactive()
		// }
		// if was paused, then we know it is definietly paused
		// we can remove the paused state
		if wi.state.isPaused() {
			wi.state.setUnpause()
		}
	}

	return nil
}

var ErrWorkflowInstanceExecution = errors.New("failed to execute workflow instance")

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, args ...interface{}) error {

	logger.Debug(wi.ctx, "workflow instance execute workflow", "workflow_id", wi.workflowID)

	if len(args) != 1 {
		err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("WorkflowInstance executeWorkflow expected 1 argument, got %d", len(args)))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	var err error
	var inputs []interface{}
	var ok bool

	var maxAttempts uint64
	var maxInterval time.Duration
	var retryState RetryState

	maxAttempts, maxInterval = getRetryPolicyOrDefault(wi.options.RetryPolicy)

	if inputs, ok = args[0].([]interface{}); !ok {
		err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("WorkflowInstance executeActivity expected argument to be []interface{}, got %T", args[0]))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to get workflow data properties: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	logger.Debug(wi.ctx, "workflow instance execute workflow retries", "workflow_id", wi.workflowID)
	if err := retry.Do(
		wi.ctx,
		retry.
			WithMaxRetries(
				maxAttempts,
				retry.NewConstant(maxInterval)),
		// only the execution might change
		func(ctx context.Context) error {
			logger.Debug(wi.ctx, "attempt execute workflow", "workflow_id", wi.workflowID)

			// now we can create and set the execution id
			if wi.executionID, err = wi.db.AddWorkflowExecution(&WorkflowExecution{
				BaseExecution: BaseExecution{
					EntityID: wi.workflowID,
					Status:   ExecutionStatusRunning,
				},
				WorkflowExecutionData: &WorkflowExecutionData{},
			}); err != nil {
				err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to add workflow execution: %w", err))
				logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
				return err
			}

			if !isRoot {
				logger.Debug(wi.ctx, "workflow instance is not root", "workflow_id", wi.workflowID)
				// Check if hierarchy already exists for this execution
				hierarchies, err := wi.db.GetHierarchiesByChildEntity(wi.workflowID)
				if err != nil || len(hierarchies) == 0 {
					logger.Debug(wi.ctx, "workflow instance hierarchy does not exist", "workflow_id", wi.workflowID)
					// Only create hierarchy if none exists
					if _, err = wi.db.AddHierarchy(&Hierarchy{
						RunID:             wi.runID,
						ParentStepID:      wi.parentStepID,
						ParentEntityID:    wi.parentEntityID,
						ParentExecutionID: wi.parentExecutionID,
						ChildStepID:       wi.stepID,
						ChildEntityID:     wi.workflowID,
						ChildExecutionID:  wi.executionID,
						ParentType:        EntityWorkflow,
						ChildType:         EntityWorkflow,
					}); err != nil {
						err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to add hierarchy: %w", err))
						logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
						return err
					}
				} else {
					logger.Debug(wi.ctx, "workflow instance hierarchy exists", "workflow_id", wi.workflowID)
					// TODO: hum that's a bug no? we append all the time, we shouldn't update the hierarchy
					// Update existing hierarchy with new execution ID
					hierarchy := hierarchies[0]
					hierarchy.ChildExecutionID = wi.executionID
					if err = wi.db.UpdateHierarchy(hierarchy); err != nil {
						err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to update hierarchy: %w", err))
						logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
						return err
					}
				}
			}

			logger.Debug(wi.ctx, "workflow instance attempt run workflow", "workflow_id", wi.workflowID)

			// Run the real workflow
			workflowOutput, workflowErr := wi.runWorkflow(inputs)

			// success case
			if workflowErr == nil {
				logger.Debug(wi.ctx, "workflow instance run workflow success", "workflow_id", wi.workflowID)
				// Success
				wi.mu.Lock()

				// if we detected a pause
				if workflowOutput.Paused {
					logger.Debug(wi.ctx, "workflow instance detected pause", "workflow_id", wi.workflowID)
					if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusPaused)); err != nil {
						wi.mu.Unlock()
						err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status paused: %w", err))
						logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
						return err
					}

					wi.fsm.Fire(TriggerPause)
				} else {
					logger.Debug(wi.ctx, "workflow instance completed", "workflow_id", wi.workflowID)
					// normal case
					if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
						wi.mu.Unlock()
						err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status completed: %w", err))
						logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
						return err
					}

					wi.fsm.Fire(TriggerComplete, workflowOutput)
				}

				wi.mu.Unlock()

				return nil
			} else {
				logger.Debug(wi.ctx, "workflow instance run workflow failed", "workflow_id", wi.workflowID)
				// real failure
				wi.mu.Lock()

				// Execution failed
				setters := []WorkflowExecutionPropertySetter{
					SetWorkflowExecutionStatus(ExecutionStatusFailed),
					SetWorkflowExecutionError(workflowErr.Error()), // error message
				}

				if workflowOutput != nil && workflowOutput.StrackTrace != nil {
					setters = append(setters, SetWorkflowExecutionStackTrace(*workflowOutput.StrackTrace)) // eventual stack trace
				}

				// you failed but maybe the next time it will be better
				if err = wi.db.SetWorkflowExecutionProperties(
					wi.executionID,
					setters...,
				); err != nil {
					wi.mu.Unlock()
					err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status failed: %w", err))
					logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
					return err
				}

				wi.mu.Unlock()
			}

			// Update the attempt counter
			retryState.Attempts++
			if err = wi.db.
				SetWorkflowEntityProperties(wi.workflowID,
					SetWorkflowEntityRetryState(retryState)); err != nil {
				err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow entity retry state: %w", err))
				logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
				return err
			}

			logger.Debug(wi.ctx, "workflow instance failed (might be retried)", "workflow_id", wi.workflowID, "retry_attempts", retryState.Attempts)

			logger.Error(wi.ctx, workflowErr.Error(), "workflow_id", wi.workflowID)
			return retry.RetryableError(workflowErr) // directly return the workflow error, if the execution failed we will retry eventually
		}); err != nil {
		// Max attempts reached and not resumable
		wi.mu.Lock()
		// the whole workflow entity failed
		err := errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to execute workflow instance: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		wi.fsm.Fire(TriggerFail, err)

		wi.mu.Unlock()
	}
	return nil
}

var ErrWorkflowInstanceRun = errors.New("failed to run workflow instance")

// YOU DONT CHANGE THE STATUS OF THE WORKFLOW HERE
// Whatever happen in the runWorkflow is interpreted for form the WorkflowOutput or an error
func (wi *WorkflowInstance) runWorkflow(inputs []interface{}) (outputs *WorkflowOutput, err error) {

	logger.Debug(wi.ctx, "workflow instance run workflow", "workflow_id", wi.workflowID)

	// Get the latest execution, which may be nil if none exist
	latestExecution, err := wi.db.GetWorkflowExecutionLatestByEntityID(wi.workflowID)
	if err != nil {
		// Only return if there's an actual error)
		err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to get latest workflow execution: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return nil, err
	}

	// No need to run it
	// Technically, that's handled at the contest level
	// TODO: is that really necessary?!?!
	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
		logger.Debug(wi.ctx, "workflow instance run workflow already completed", "workflow_id", wi.workflowID)
		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Outputs)
		if err != nil {
			err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to convert outputs from serialization: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return nil, err
		}
		return &WorkflowOutput{
			Outputs: outputs,
		}, nil
	}

	// Get handler early with proper mutex handling
	wi.mu.Lock()
	handler := wi.handler
	wFunc := handler.Handler
	wi.mu.Unlock()

	ctxWorkflow := WorkflowContext{
		db:                  wi.db,
		ctx:                 wi.ctx,
		state:               wi.state,
		tracker:             wi.tracker,
		debug:               wi.debug,
		registry:            wi.registry,
		options:             wi.options,
		runID:               wi.runID,
		queueID:             wi.queueID,
		stepID:              wi.stepID,
		workflowID:          wi.workflowID,
		workflowExecutionID: wi.executionID,

		onCrossQueueWorkflow: wi.onCrossWorkflow,
		onContinueAsNew:      wi.onContinueAsNew,
		onSignalNew:          wi.onSignalNew,
	}

	var workflowInputs [][]byte
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataInputs(&workflowInputs)); err != nil {
		err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to get workflow data inputs: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return nil, err
	}

	argsValues := []reflect.Value{reflect.ValueOf(ctxWorkflow)}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if wi.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			wi.mu.Lock()

			// allow the caller to know about the panic
			err = errors.Join(ErrWorkflowInstanceRun, ErrWorkflowPanicked)
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			// due to the `err` it will trigger the fail state
			outputs = &WorkflowOutput{
				StrackTrace: &stackTrace,
			}

			wi.mu.Unlock()
		}
	}()

	select {
	case <-wi.ctx.Done():
		err := errors.Join(ErrWorkflowInstanceRun, wi.ctx.Err(), fmt.Errorf("workflow instance context done: %w", wi.ctx.Err()))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return nil, err
	default:
	}

	results := reflect.ValueOf(wFunc).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("function %s should return at least an error", handler.HandlerName))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return nil, err
	}

	errInterface := results[numOut-1].Interface()

	// could be a fake error to trigger a continue as new
	var continueAsNewErr *ContinueAsNewError
	var ok bool
	if continueAsNewErr, ok = errInterface.(*ContinueAsNewError); ok {
		// We treat it as normal completion, return nil to proceed
		errInterface = nil
	}

	var realError error
	var pausedDetected bool

	if realError, ok = errInterface.(error); ok {
		// could be a fake error to trigger a pause
		if errors.Is(realError, ErrPaused) {
			realError = nil // mute the error
			pausedDetected = true
		}
	}

	if realError != nil {
		logger.Debug(wi.ctx, "workflow instance run workflow failed", "workflow_id", wi.workflowID)

		// Update execution error
		if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionError(realError.Error())); err != nil {
			// wi.mu.Unlock()
			err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution error: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return nil, err
		}

		// Emptied the output data
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(nil)); err != nil {
			// wi.mu.Unlock()
			err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution data outputs: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return nil, err
		}

		return nil, realError
	} else {
		logger.Debug(wi.ctx, "workflow instance run workflow success", "workflow_id", wi.workflowID)

		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				outputs = append(outputs, result)
			}
		}

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(outputs)
		if err != nil {
			err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to convert outputs for serialization: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return nil, err
		}

		// save result
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(outputBytes)); err != nil {
			// wi.mu.Unlock()
			err := errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution data outputs: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return nil, err
		}

		// We do not take risk to sent back an instruction
		if continueAsNewErr != nil {
			return &WorkflowOutput{
				Outputs:              outputs,
				ContinueAsNewOptions: continueAsNewErr.Options,
				ContinueAsNewArgs:    continueAsNewErr.Args,
				Paused:               pausedDetected,
			}, nil
		}

		return &WorkflowOutput{
			Outputs: outputs,
			Paused:  pausedDetected,
		}, nil
	}
}

var ErrWorkflowInstanceCompleted = errors.New("failed to complete workflow instance")

func (wi *WorkflowInstance) onCompleted(_ context.Context, args ...interface{}) error {

	logger.Debug(wi.ctx, "workflow instance on completed", "workflow_id", wi.workflowID)

	if len(args) != 1 {
		err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("WorkflowInstance onCompleted expected 1 argument, got %d", len(args)))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	var ok bool
	var workflowOutput *WorkflowOutput
	if workflowOutput, ok = args[0].(*WorkflowOutput); !ok {
		err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("WorkflowInstance onCompleted expected argument to be *WorkflowOutput, got %T", args[0]))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	var err error

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to get workflow data properties: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	// Handle ContinueAsNew
	// We re-create a new workflow entity to that related to that workflow
	// It will run on another worker on it's own, we are use preparing it
	if workflowOutput.ContinueAsNewOptions != nil {

		logger.Debug(wi.ctx, "workflow instance continue as new", "workflow_id", wi.workflowID)

		wi.mu.Lock()
		handler := wi.handler
		copyID := wi.workflowID
		wi.mu.Unlock()

		var workflowEntity *WorkflowEntity

		if workflowEntity, err = prepareWorkflow(
			wi.registry,
			wi.db,
			handler.Handler,
			workflowOutput.ContinueAsNewOptions,
			&preparationOptions{
				parentWorkflowID:          copyID,
				parentWorkflowExecutionID: wi.executionID,
				parentRunID:               wi.runID,
			},
			workflowOutput.ContinueAsNewArgs...); err != nil {
			err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to prepare workflow: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return err
		}

		// Use the callback if available
		// Otherwise, you will have to use ExecuteWithEntity to manually trigger the workflow
		if wi.onContinueAsNew != nil {
			logger.Debug(wi.ctx, "workflow instance created continue as new", "workflow_id", wi.workflowID)
			queue := wi.options.Queue
			if workflowOutput.ContinueAsNewOptions.Queue != "" {
				queue = workflowOutput.ContinueAsNewOptions.Queue
			}

			// Get the handler info to get the original workflow func

			_ = wi.onContinueAsNew( // Just start it running
				queue,
				workflowEntity.ID,
				handler.Handler,
				workflowOutput.ContinueAsNewOptions,
				workflowOutput.ContinueAsNewArgs...,
			)
		} else {
			logger.Debug(wi.ctx, "workflow instance continue as new require manual execution", "workflow_id", wi.workflowID)
		}
		// end continue as new
	}

	if isRoot {
		logger.Debug(wi.ctx, "workflow instance is root for completion", "workflow_id", wi.workflowID)
		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusCompleted)); err != nil {
			err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to set run status completed: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return err
		}

		if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
			err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to set workflow execution status completed: %w", err))
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return err
		}
	}

	if err = wi.db.SetWorkflowEntityProperties(wi.workflowID, SetWorkflowEntityStatus(StatusCompleted)); err != nil {
		// wi.mu.Unlock()
		err := errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to set workflow entity status: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	// Normal completion
	if wi.future != nil {
		logger.Debug(wi.ctx, "workflow instance completed", "workflow_id", wi.workflowID)
		wi.future.setResult(workflowOutput.Outputs)
	}

	return nil
}

var ErrWorkflowInstanceFailed = errors.New("failed to fail workflow instance")

func (wi *WorkflowInstance) onFailed(_ context.Context, args ...interface{}) error {

	logger.Debug(wi.ctx, "workflow instance on failed", "workflow_id", wi.workflowID)

	var err error

	if len(args) != 1 {
		err := errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("WorkfloInstance onFailed expected 1 argument, got %d", len(args)))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	err = args[0].(error)

	joinedErr := errors.Join(ErrWorkflowFailed, err)

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		err := errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to get workflow data properties: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	if wi.ctx.Err() != nil {
		if errors.Is(wi.ctx.Err(), context.Canceled) {
			if err := wi.db.SetWorkflowExecutionProperties(wi.workflowID, SetWorkflowExecutionStatus(ExecutionStatusCancelled)); err != nil {
				err := errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set Workflow execution status cancelled: %w", err))
				logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
				return err
			}
			if err := wi.db.SetWorkflowEntityProperties(wi.workflowID, SetWorkflowEntityStatus(StatusCancelled)); err != nil {
				err := errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set Workflow entity status cancelled: %w", err))
				logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
				return err
			}

			// If this is the root workflow, update the Run status to Failed
			if isRoot {
				// Entity is now a failure
				if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusCancelled)); err != nil {
					err := errors.Join(ErrWorkflowInstanceFailed, err)
					logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
					return err
				}
			}

			if wi.future != nil {
				logger.Debug(wi.ctx, "workflow instance cancelled", "workflow_id", wi.workflowID, "error", wi.ctx.Err())
				wi.future.setError(wi.ctx.Err())
			}

			return nil
		}
	}

	// Since the executions failed many times and was already changed
	if err = wi.db.
		SetWorkflowEntityProperties(
			wi.workflowID,
			SetWorkflowEntityStatus(StatusFailed),
		); err != nil {
		err := errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set workflow entity status: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		return err
	}

	// If this is the root workflow, update the Run status to Failed
	if isRoot {
		// Entity is now a failure
		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusFailed)); err != nil {
			err := errors.Join(ErrWorkflowInstanceFailed, err)
			logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
			return err
		}
	}

	if wi.future != nil {
		err := errors.Join(ErrWorkflowInstanceFailed, joinedErr)
		logger.Debug(wi.ctx, "workflow instance failed", "workflow_id", wi.workflowID, "error", err)
		wi.future.setError(err)
	}

	logger.Debug(wi.ctx, "workflow instance failed", "workflow_id", wi.workflowID, "error", err)

	return nil
}

var ErrWorkflowInstancePaused = errors.New("failed to pause workflow instance")

func (wi *WorkflowInstance) onPaused(_ context.Context, _ ...interface{}) error {

	logger.Debug(wi.ctx, "workflow instance on paused", "workflow_id", wi.workflowID)

	if wi.future != nil {
		err := errors.Join(ErrWorkflowInstancePaused, ErrPaused)
		logger.Debug(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		wi.future.setError(err)
	}

	var err error

	// Tell the orchestrator to manage the case
	if err = wi.tracker.onPause(); err != nil {
		err := errors.Join(ErrWorkflowInstancePaused, fmt.Errorf("failed to pause orchestrator: %w", err))
		logger.Error(wi.ctx, err.Error(), "workflow_id", wi.workflowID)
		if wi.future != nil {
			wi.future.setError(err)
		}
		return err
	}

	logger.Debug(wi.ctx, "workflow instance paused", "workflow_id", wi.workflowID)

	return nil
}

///////////// Activity

var ErrActivityContext = errors.New("failed activity context")

func (ctx WorkflowContext) Activity(stepID string, activityFunc interface{}, options *ActivityOptions, args ...interface{}) Future {
	var err error

	logger.Debug(ctx.ctx, "activity context", "workflow_id", ctx.workflowID, "step_id", stepID)

	if err = ctx.checkPause(); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrActivityContext, err)
		if !errors.Is(err, ErrPaused) {
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		}
		future.setError(err)
		return future
	}

	future := NewRuntimeFuture()

	// Register activity on-the-fly
	handler, err := ctx.registry.RegisterActivity(activityFunc)
	if err != nil {
		err := errors.Join(ErrActivityContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	var activityEntityID int

	// Get all hierarchies for this workflow+stepID combination
	var hierarchies []*Hierarchy
	if hierarchies, err = ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, stepID, EntityActivity); err != nil {
		if !errors.Is(err, ErrHierarchyNotFound) {
			err := errors.Join(ErrActivityContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}
	}

	// Convert inputs to [][]byte
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		err := errors.Join(ErrActivityContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	// If we have existing hierarchies
	if len(hierarchies) > 0 {
		logger.Debug(ctx.ctx, "activity context existing", "workflow_id", ctx.workflowID, "step_id", stepID)
		// Check if we have multiple different entity IDs (shouldn't happen)
		activityEntityID = hierarchies[0].ChildEntityID
		for _, h := range hierarchies[1:] {
			if h.ChildEntityID != activityEntityID {
				// TODO: wat?!
			}
		}

		// If not completed, reuse the existing activity entity for retry
		future.setEntityID(activityEntityID)

		// Check the latest execution for completion
		var activityExecution *ActivityExecution
		if activityExecution, err = ctx.db.GetActivityExecution(hierarchies[0].ChildExecutionID); err != nil {
			err := errors.Join(ErrActivityContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		// If we have a completed execution, return its result
		if activityExecution.Status == ExecutionStatusCompleted {

			var activityExecutionData *ActivityExecutionData
			if activityExecutionData, err = ctx.db.GetActivityExecutionData(activityExecution.ID); err != nil {
				err := errors.Join(ErrActivityContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			outputs, err := convertOutputsFromSerialization(handler, activityExecutionData.Outputs)
			if err != nil {
				err := errors.Join(ErrActivityContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			logger.Debug(ctx.ctx, "activity context completed", "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setResult(outputs)
			return future
		}
	} else {
		// No existing activity entity found, create a new one
		logger.Debug(ctx.ctx, "activity context create new", "workflow_id", ctx.workflowID, "step_id", stepID)
		// Convert API RetryPolicy to internal retry policy
		var internalRetryPolicy *retryPolicyInternal = &retryPolicyInternal{}

		// Handle options and defaults
		if options != nil && options.RetryPolicy != nil {
			rp := options.RetryPolicy
			// Fill default values if zero
			if rp.MaxAttempts == 0 {
				internalRetryPolicy.MaxAttempts = 0 // no retries
			}
			if rp.MaxInterval == 0 {
				internalRetryPolicy.MaxInterval = rp.MaxInterval.Nanoseconds()
			}
		} else {
			// Default RetryPolicy
			internalRetryPolicy = DefaultRetryPolicyInternal()
		}

		if activityEntityID, err = ctx.db.AddActivityEntity(&ActivityEntity{
			BaseEntity: BaseEntity{
				RunID:       ctx.runID,
				QueueID:     ctx.queueID,
				Type:        EntityActivity,
				HandlerName: handler.HandlerName,
				Status:      StatusPending,
				RetryPolicy: *internalRetryPolicy,
				RetryState:  RetryState{Attempts: 0},
				StepID:      stepID,
			},
			ActivityData: &ActivityData{
				Inputs: inputBytes,
			},
		}, ctx.workflowID); err != nil {
			err := errors.Join(ErrActivityContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		future.setEntityID(activityEntityID)
	}

	logger.Debug(ctx.ctx, "activity context start", "workflow_id", ctx.workflowID, "step_id", stepID)

	var data *ActivityData
	if data, err = ctx.db.GetActivityDataByEntityID(activityEntityID); err != nil {
		err := errors.Join(ErrActivityContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
	}

	var cancel context.CancelFunc
	ctx.ctx, cancel = context.WithCancel(ctx.ctx)

	activityInstance := &ActivityInstance{
		ctx:     ctx.ctx,
		cancel:  cancel,
		db:      ctx.db,
		tracker: ctx.tracker,
		state:   ctx.state,
		debug:   ctx.debug,

		parentExecutionID: ctx.workflowExecutionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,

		handler: handler,

		future: future, // connect when the result is done or not

		workflowID:  ctx.workflowID,
		stepID:      stepID,
		entityID:    activityEntityID,
		executionID: 0, // will be set later
		dataID:      data.ID,
		queueID:     ctx.queueID,
		runID:       ctx.runID,
	}

	if options != nil {
		activityInstance.options = *options
	}

	inputs, err := convertInputsFromSerialization(handler, inputBytes)
	if err != nil {
		err := errors.Join(ErrActivityContext, ErrSerialization, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	ctx.tracker.addActivityInstance(activityInstance)
	logger.Debug(ctx.ctx, "activity context start activity", "workflow_id", ctx.workflowID, "step_id", stepID)
	activityInstance.Start(inputs) // synchronous because we want to know when it's done for real, the future will sent back the data anyway

	return future
}

type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	ctx     context.Context
	cancel  context.CancelFunc
	mu      deadlock.Mutex
	db      Database
	tracker InstanceTracker
	state   StateTracker
	debug   Debug

	handler HandlerInfo

	fsm *stateless.StateMachine

	options ActivityOptions
	future  Future

	runID       int
	stepID      string
	workflowID  int
	entityID    int
	executionID int
	dataID      int
	queueID     int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

var ErrActivityInstance = errors.New("failed activity instance")

func (ai *ActivityInstance) Start(inputs []interface{}) {
	logger.Debug(ai.ctx, "activity instance start", "workflow_id", ai.workflowID, "step_id", ai.stepID)
	// ai.mu.Lock()
	// Initialize the FSM
	ai.fsm = stateless.NewStateMachine(StateIdle)
	ai.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	ai.fsm.Configure(StateExecuting).
		OnEntry(ai.executeActivity).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	ai.fsm.Configure(StateCompleted).
		OnEntry(ai.onCompleted)

	ai.fsm.Configure(StateFailed).
		OnEntry(ai.onFailed)

	ai.fsm.Configure(StatePaused).
		OnEntry(ai.onPaused).
		Permit(TriggerResume, StateExecuting)
	// ai.mu.Unlock()

	logger.Debug(ai.ctx, "activity instance fsm configured", "workflow_id", ai.workflowID, "step_id", ai.stepID)
	// Start the FSM
	if err := ai.fsm.Fire(TriggerStart, inputs); err != nil {
		err := errors.Join(ErrActivityInstance, err)
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		ai.future.setError(err)
	}
}

var ErrActivityInstanceExecute = errors.New("failed to execute activity instance")

func (ai *ActivityInstance) executeActivity(_ context.Context, args ...interface{}) error {

	logger.Debug(ai.ctx, "activity instance execute", "workflow_id", ai.workflowID, "step_id", ai.stepID)

	if len(args) != 1 {
		err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("ActivityInstance executeActivity expected 1 argument, got %d", len(args)))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}

	var err error
	var maxAttempts uint64
	var maxInterval time.Duration
	var inputs []interface{}
	var ok bool
	var retryState RetryState

	// pp.Println(wi.options.RetryPolicy)
	maxAttempts, maxInterval = getRetryPolicyOrDefault(ai.options.RetryPolicy)

	if inputs, ok = args[0].([]interface{}); !ok {
		err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("ActivityInstance executeActivity expected argument to be []interface{}, got %T", args[0]))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}

	if err := retry.Do(
		ai.ctx,
		retry.
			WithMaxRetries(
				maxAttempts,
				retry.NewConstant(maxInterval)),
		// only the execution might change
		func(ctx context.Context) error {

			logger.Debug(ai.ctx, "activity instance execute retry", "workflow_id", ai.workflowID, "step_id", ai.stepID)

			// now we can create and set the execution id
			if ai.executionID, err = ai.db.AddActivityExecution(&ActivityExecution{
				BaseExecution: BaseExecution{
					EntityID: ai.entityID,
					Status:   ExecutionStatusRunning,
				},
				ActivityExecutionData: &ActivityExecutionData{},
			}); err != nil {
				ai.mu.Lock()
				err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to add activity execution: %w", err))
				logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
				ai.fsm.Fire(TriggerFail, err)
				ai.mu.Unlock()
				return err
			}

			// Hierarchy
			if _, err = ai.db.AddHierarchy(&Hierarchy{
				RunID:             ai.runID,
				ParentStepID:      ai.parentStepID,
				ParentEntityID:    ai.parentEntityID,
				ParentExecutionID: ai.parentExecutionID,
				ChildStepID:       ai.stepID,
				ChildEntityID:     ai.entityID,
				ChildExecutionID:  ai.executionID,
				ParentType:        EntityWorkflow,
				ChildType:         EntityActivity,
			}); err != nil {
				// ai.mu.Unlock()
				err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to add hierarchy: %w", err))
				logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
				return err
			}

			// run the real activity
			activityOutput, activityErr := ai.runActivity(inputs)

			if activityErr == nil {
				logger.Debug(ai.ctx, "activity instance execute completed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

				ai.mu.Lock()

				if activityOutput.Paused {
					logger.Debug(ai.ctx, "activity instance execute paused", "workflow_id", ai.workflowID, "step_id", ai.stepID)

					if err = ai.db.SetActivityExecutionProperties(
						ai.executionID,
						SetActivityExecutionStatus(ExecutionStatusPaused),
					); err != nil {
						ai.mu.Unlock()
						err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution status: %w", err))
						logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
						return err
					}
					ai.fsm.Fire(TriggerPause)

				} else {
					logger.Debug(ai.ctx, "activity instance execute completed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

					if err = ai.db.SetActivityExecutionProperties(
						ai.executionID,
						SetActivityExecutionStatus(ExecutionStatusCompleted),
					); err != nil {
						ai.mu.Unlock()
						err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution status: %w", err))
						logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
						return err
					}

					ai.fsm.Fire(TriggerComplete, activityOutput)
				}

				ai.mu.Unlock()

				return nil
			} else {
				logger.Debug(ai.ctx, "activity instance execute failed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

				if errors.Is(activityErr, context.Canceled) {
					// Context cancelled
					ai.mu.Lock()
					err := errors.Join(ErrActivityInstanceExecute, activityErr)
					ai.fsm.Fire(TriggerFail, err)
					ai.mu.Unlock()
					return activityErr
				}

				ai.mu.Lock()

				setters := []ActivityExecutionPropertySetter{
					SetActivityExecutionStatus(ExecutionStatusFailed),
					SetActivityExecutionError(activityErr.Error()),
				}

				// could be context cancelled
				if activityErr != nil && activityOutput != nil {
					if activityOutput.StrackTrace != nil {
						setters = append(setters, SetActivityExecutionStackTrace(*activityOutput.StrackTrace))
					}
				}

				if err = ai.db.SetActivityExecutionProperties(
					ai.executionID,
					setters...,
				); err != nil {
					ai.mu.Unlock()
					err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution properties: %w", err))
					logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
					return err
				}

				ai.mu.Unlock()

			}

			// Update the attempt counter
			retryState.Attempts++
			if err = ai.db.
				SetActivityEntityProperties(ai.entityID,
					SetActivityEntityRetryState(retryState)); err != nil {
				err := errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activty entity retry state: %w", err))
				logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
				return err
			}

			return retry.RetryableError(activityErr) // directly return the activty error, if the execution failed we will retry eventually
		}); err != nil {
		// Max attempts reached and not resumable
		ai.mu.Lock()
		err := errors.Join(ErrActivityInstanceExecute, err)
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		// the whole activty entity failed
		ai.fsm.Fire(TriggerFail, err)

		ai.mu.Unlock()
	}
	return nil
}

var ErrActivityInstanceRun = errors.New("failed to run activity instance")

// Sub-function within executeWithRetry
func (ai *ActivityInstance) runActivity(inputs []interface{}) (outputs *ActivityOutput, err error) {

	logger.Debug(ai.ctx, "activity instance run", "workflow_id", ai.workflowID, "step_id", ai.stepID)

	activityExecution, err := ai.db.GetActivityExecutionLatestByEntityID(ai.entityID)
	if err != nil {
		err := errors.Join(ErrActivityInstanceRun, err)
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return nil, err
	}

	if activityExecution != nil && activityExecution.Status == ExecutionStatusCompleted &&
		activityExecution.ActivityExecutionData != nil &&
		activityExecution.ActivityExecutionData.Outputs != nil {

		outputsData, err := convertOutputsFromSerialization(ai.handler, activityExecution.ActivityExecutionData.Outputs)
		if err != nil {
			err := errors.Join(ErrActivityInstanceRun, err)
			logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			return nil, err
		}
		return &ActivityOutput{
			Outputs: outputsData,
		}, nil
	}

	handler := ai.handler
	activityFunc := handler.Handler

	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if ai.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			ai.mu.Lock()
			// allow the caller to know about the panic
			err = errors.Join(ErrActivityInstanceRun, ErrActivityPanicked)
			logger.Error(ai.ctx, fmt.Sprintf("activity panicked: %v", r), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			outputs = &ActivityOutput{
				StrackTrace: &stackTrace,
			}
			// due to the `err` it will trigger the fail state
			ai.mu.Unlock()
		}
	}()

	var activityInputs [][]byte
	if err = ai.db.GetActivityDataProperties(ai.entityID, GetActivityDataInputs(&activityInputs)); err != nil {
		err := errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to get activity data inputs: %w", err))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return nil, err
	}

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	select {
	case <-ai.ctx.Done():
		err := errors.Join(ErrActivityInstanceRun, ai.ctx.Err())
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return nil, err
	default:
	}

	results := reflect.ValueOf(activityFunc).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := errors.Join(ErrActivityInstanceRun, fmt.Errorf("activity %s should return at least an error", handler.HandlerName))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return nil, err
	}

	errInterface := results[numOut-1].Interface()
	if errInterface != nil {

		logger.Debug(ai.ctx, "activity instance run failed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

		errActivity := errInterface.(error)

		// Serialize error
		errorMessage := errActivity.Error()

		// Update execution error
		if err = ai.db.SetActivityExecutionProperties(
			ai.entityID,
			SetActivityExecutionError(errorMessage)); err != nil {
			err := errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution error: %w", err))
			logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			return nil, err
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(nil)); err != nil {
			err := errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution data outputs: %w", err))
			logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			return nil, err
		}

		return nil, errors.Join(ErrActivityInstanceRun, errActivity)
	} else {

		logger.Debug(ai.ctx, "activity instance run completed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				outputs = append(outputs, result)
			}
		}

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(outputs)
		if err != nil {
			err := errors.Join(ErrActivityInstanceRun, err)
			logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			return nil, err
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(outputBytes)); err != nil {
			err := errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution data outputs: %w", err))
			logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
			return nil, err
		}

		return &ActivityOutput{
			Outputs: outputs,
		}, nil
	}
}

var ErrActivityInstanceCompleted = errors.New("failed to complete activity instance")

func (ai *ActivityInstance) onCompleted(_ context.Context, args ...interface{}) error {

	logger.Debug(ai.ctx, "activity instance completed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

	if len(args) != 1 {
		err := errors.Join(ErrActivityInstanceCompleted, fmt.Errorf("activity instance onCompleted expected 1 argument, got %d", len(args)))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}

	if ai.future != nil {
		activtyOutput := args[0].(*ActivityOutput)
		logger.Debug(ai.ctx, "activity instance completed", "workflow_id", ai.workflowID, "step_id", ai.stepID)
		ai.future.setResult(activtyOutput.Outputs)
	}

	if err := ai.db.SetActivityEntityProperties(ai.entityID, SetActivityEntityStatus(StatusCompleted)); err != nil {
		err := errors.Join(ErrActivityInstanceCompleted, fmt.Errorf("failed to set activity entity status: %w", err))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}

	return nil
}

var ErrActivityInstanceFailed = errors.New("failed to fail activity instance")

func (ai *ActivityInstance) onFailed(_ context.Context, args ...interface{}) error {

	logger.Debug(ai.ctx, "activity instance failed", "workflow_id", ai.workflowID, "step_id", ai.stepID)

	if len(args) != 1 {
		err := errors.Join(ErrActivityInstanceFailed, fmt.Errorf("activity instance onFailed expected 1 argument, got %d", len(args)))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}
	if ai.future != nil {
		err := args[0].(error)
		err = errors.Join(ErrActivityInstanceFailed, err)
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		ai.future.setError(err)
	}

	if ai.ctx.Err() != nil {
		if errors.Is(ai.ctx.Err(), context.Canceled) {
			logger.Debug(ai.ctx, "activity instance cancelled", "workflow_id", ai.workflowID, "step_id", ai.stepID)
			if err := ai.db.SetActivityExecutionProperties(ai.entityID, SetActivityExecutionStatus(ExecutionStatusCancelled)); err != nil {
				err := errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity execution status: %w", err))
				logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
				return err
			}
			if err := ai.db.SetActivityEntityProperties(ai.entityID, SetActivityEntityStatus(StatusCancelled)); err != nil {
				err := errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity entity status: %w", err))
				logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
				return err
			}
			return nil
		}
	}

	if err := ai.db.SetActivityEntityProperties(
		ai.entityID,
		SetActivityEntityStatus(StatusFailed),
	); err != nil {
		err := errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity entity status: %w", err))
		logger.Error(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		return err
	}

	return nil
}

var ErrActivityInstancePaused = errors.New("failed to pause activity instance")

// In theory, the behavior of the pause doesn't exists
// The Activity is supposed to finish what it was doing, the developer own the responsibility to detect the cancellation of it's own context.
func (ai *ActivityInstance) onPaused(_ context.Context, _ ...interface{}) error {
	logger.Debug(ai.ctx, "activity instance paused", "workflow_id", ai.workflowID, "step_id", ai.stepID)
	if ai.future != nil {
		err := errors.Join(ErrActivityInstancePaused, ErrPaused)
		logger.Debug(ai.ctx, err.Error(), "workflow_id", ai.workflowID, "step_id", ai.stepID)
		ai.future.setError(err)
	}
	return nil
}

/////////////////////////////////// Side Effects

type SideEffectOutput struct {
	Outputs     []interface{}
	StrackTrace *string
	Paused      bool
}

var ErrSideEffectContext = errors.New("failed side effect context")

func (ctx WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}) Future {
	var err error
	logger.Debug(ctx.ctx, "side effect context", "workflow_id", ctx.workflowID, "step_id", stepID)

	if err = ctx.checkPause(); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrSideEffectContext, err)
		if !errors.Is(err, ErrPaused) {
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		}
		future.setError(err)
		return future
	}

	future := NewRuntimeFuture()

	// Register side effect on-the-fly
	handler, err := ctx.registry.RegisterSideEffect(sideEffectFunc)
	if err != nil {
		err := errors.Join(ErrSideEffectContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	var sideEffectEntityID int

	// First check if there's an existing side effect entity for this workflow and step
	var sideEffectEntities []*SideEffectEntity
	if sideEffectEntities, err = ctx.db.GetSideEffectEntities(ctx.workflowID, SideEffectEntityWithData()); err != nil {
		err := errors.Join(ErrSideEffectContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		if !errors.Is(err, ErrSideEffectEntityNotFound) {
			future.setError(err)
			return future
		}
	}

	// Find matching entity by stepID
	var entity *SideEffectEntity
	for _, e := range sideEffectEntities {
		if e.StepID == stepID {
			entity = e
			sideEffectEntityID = e.ID
			break
		}
	}

	// If we found an existing entity, check for completed executions
	if entity != nil {
		execs, err := ctx.db.GetSideEffectExecutions(entity.ID)
		if err != nil {
			err := errors.Join(ErrSideEffectContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		// Look for a successful execution
		for _, exec := range execs {
			if exec.Status == ExecutionStatusCompleted {
				// Get the execution data
				execData, err := ctx.db.GetSideEffectExecutionDataByExecutionID(exec.ID)
				if err != nil {
					err := errors.Join(ErrSideEffectContext, err)
					logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
					future.setError(err)
					return future
				}
				if execData != nil && len(execData.Outputs) > 0 {
					// Reuse the existing result
					outputs, err := convertOutputsFromSerialization(handler, execData.Outputs)
					if err != nil {
						err := errors.Join(ErrSideEffectContext, err)
						logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
						future.setError(err)
						return future
					}
					logger.Debug(ctx.ctx, "side effect context completed", "workflow_id", ctx.workflowID, "step_id", stepID)
					future.setEntityID(sideEffectEntityID)
					future.setResult(outputs)
					return future
				}
			}
		}
	}

	// No existing completed execution found, proceed with creating new entity/execution
	if entity == nil {
		logger.Debug(ctx.ctx, "side effect context create new", "workflow_id", ctx.workflowID, "step_id", stepID)
		if sideEffectEntityID, err = ctx.db.AddSideEffectEntity(&SideEffectEntity{
			BaseEntity: BaseEntity{
				RunID:       ctx.runID,
				QueueID:     ctx.queueID,
				Type:        EntitySideEffect,
				HandlerName: handler.HandlerName,
				Status:      StatusPending,
				RetryPolicy: retryPolicyInternal{
					MaxAttempts: 0, // No retries for side effects
					MaxInterval: time.Second.Nanoseconds(),
				},
				RetryState: RetryState{Attempts: 0},
				StepID:     stepID,
			},
			SideEffectData: &SideEffectData{},
		}, ctx.workflowID); err != nil {
			err := errors.Join(ErrSideEffectContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}
	}

	future.setEntityID(sideEffectEntityID)

	var data *SideEffectData
	if data, err = ctx.db.GetSideEffectDataByEntityID(sideEffectEntityID); err != nil {
		err := errors.Join(ErrSideEffectContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID, "entity_id", sideEffectEntityID)
		future.setError(err)
		return future
	}

	instance := &SideEffectInstance{
		ctx:     ctx.ctx,
		db:      ctx.db,
		tracker: ctx.tracker,
		state:   ctx.state,
		debug:   ctx.debug,

		parentExecutionID: ctx.workflowExecutionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,

		handler: handler,
		future:  future,

		workflowID:  ctx.workflowID,
		stepID:      stepID,
		entityID:    sideEffectEntityID,
		executionID: 0, // will be set later
		dataID:      data.ID,
		runID:       ctx.runID,
	}

	ctx.tracker.addSideEffectInstance(instance)

	logger.Debug(ctx.ctx, "side effect context start", "workflow_id", ctx.workflowID, "step_id", stepID)
	instance.Start()

	return future
}

var ErrSideEffectInstance = errors.New("failed side effect instance")

type SideEffectInstance struct {
	ctx     context.Context
	mu      deadlock.Mutex
	db      Database
	tracker InstanceTracker
	state   StateTracker
	debug   Debug

	handler HandlerInfo
	future  Future
	fsm     *stateless.StateMachine

	stepID            string
	runID             int
	workflowID        int
	entityID          int
	dataID            int
	executionID       int
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

func (si *SideEffectInstance) Start() {
	si.mu.Lock()
	si.fsm = stateless.NewStateMachine(StateIdle)
	fsm := si.fsm
	si.mu.Unlock()

	fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateExecuting).
		OnEntry(si.executeSideEffect).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateCompleted).
		OnEntry(si.onCompleted)

	fsm.Configure(StateFailed).
		OnEntry(si.onFailed)

	fsm.Configure(StatePaused).
		OnEntry(si.onPaused).
		Permit(TriggerResume, StateExecuting)

	if err := fsm.Fire(TriggerStart); err != nil {
		if !errors.Is(err, ErrPaused) {
			si.future.setError(errors.Join(ErrSideEffectInstance, err))
		}
	}
}

var ErrSideEffectInstanceExecute = errors.New("failed to execute side effect instance")

func (si *SideEffectInstance) executeSideEffect(_ context.Context, args ...interface{}) error {
	var err error

	logger.Debug(si.ctx, "side effect instance execute", "workflow_id", si.workflowID, "step_id", si.stepID)

	// Create execution before trying side effect
	if si.executionID, err = si.db.AddSideEffectExecution(&SideEffectExecution{
		BaseExecution: BaseExecution{
			EntityID: si.entityID,
			Status:   ExecutionStatusRunning,
		},
		SideEffectExecutionData: &SideEffectExecutionData{},
	}); err != nil {
		err := errors.Join(ErrSideEffectInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setError(err)
		}
		return err
	}

	// Add hierarchy
	if _, err = si.db.AddHierarchy(&Hierarchy{
		RunID:             si.runID,
		ParentStepID:      si.parentStepID,
		ParentEntityID:    si.parentEntityID,
		ParentExecutionID: si.parentExecutionID,
		ChildStepID:       si.stepID,
		ChildEntityID:     si.entityID,
		ChildExecutionID:  si.executionID,
		ParentType:        EntityWorkflow,
		ChildType:         EntitySideEffect,
	}); err != nil {
		err := errors.Join(ErrSideEffectInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setError(err)
		}
		return err
	}

	// Run side effect just once
	output, execErr := si.runSideEffect()

	si.mu.Lock()
	defer si.mu.Unlock()

	if execErr == nil {
		if err = si.db.SetSideEffectExecutionProperties(
			si.executionID,
			SetSideEffectExecutionStatus(ExecutionStatusCompleted)); err != nil {
			err := errors.Join(ErrSideEffectInstanceExecute, err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			if si.future != nil {
				si.future.setError(err)
			}
			return err
		}

		logger.Debug(si.ctx, "side effect instance execute completed", "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setResult(output.Outputs)
		}
		si.fsm.Fire(TriggerComplete, output)
		return nil
	}

	// Handle error case
	setters := []SideEffectExecutionPropertySetter{
		SetSideEffectExecutionStatus(ExecutionStatusFailed),
		SetSideEffectExecutionError(execErr.Error()),
	}

	if output != nil && output.StrackTrace != nil {
		setters = append(setters, SetSideEffectExecutionStackTrace(*output.StrackTrace))
	}

	if err = si.db.SetSideEffectExecutionProperties(si.executionID, setters...); err != nil {
		err := errors.Join(ErrSideEffectInstanceExecute, execErr, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setError(err)
		}
		return err
	}

	if err = si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusFailed)); err != nil {
		err := errors.Join(ErrSideEffectInstanceExecute, execErr, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setError(err)
		}
		return err
	}

	logger.Error(si.ctx, "side effect instance execute failed", "workflow_id", si.workflowID, "step_id", si.stepID)
	if si.future != nil {
		si.future.setError(execErr)
	}
	si.fsm.Fire(TriggerFail, execErr)
	return errors.Join(ErrSideEffectInstanceExecute, execErr)
}

var ErrSideEffectRun = errors.New("failed to run side effect")

func (si *SideEffectInstance) runSideEffect() (output *SideEffectOutput, err error) {
	logger.Debug(si.ctx, "side effect instance run", "workflow_id", si.workflowID, "step_id", si.stepID)

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			err = errors.Join(ErrSideEffectRun, ErrSideEffectPanicked)
			logger.Error(si.ctx, fmt.Sprintf("side effect panicked: %v", r), "workflow_id", si.workflowID, "step_id", si.stepID)
		}
	}()

	select {
	case <-si.ctx.Done():
		err := errors.Join(ErrSideEffectRun, si.ctx.Err())
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return nil, err
	default:
	}

	results := reflect.ValueOf(si.handler.Handler).Call(nil)
	outputs := make([]interface{}, len(results))
	for i, res := range results {
		outputs[i] = res.Interface()
	}

	outputBytes, err := convertOutputsForSerialization(outputs)
	if err != nil {
		err := errors.Join(ErrSideEffectRun, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return nil, err
	}

	if err = si.db.SetSideEffectExecutionDataProperties(
		si.entityID,
		SetSideEffectExecutionDataOutputs(outputBytes)); err != nil {
		err := errors.Join(ErrSideEffectRun, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return nil, err
	}

	return &SideEffectOutput{
		Outputs: outputs,
	}, nil
}

var ErrSideEffectInstanceCompleted = errors.New("failed to complete side effect instance")

func (si *SideEffectInstance) onCompleted(_ context.Context, args ...interface{}) error {

	logger.Debug(si.ctx, "side effect instance completed", "workflow_id", si.workflowID, "step_id", si.stepID)

	if len(args) != 1 {
		err := errors.Join(ErrSideEffectInstanceCompleted, fmt.Errorf("SideEffectInstance onCompleted expected 1 argument, got %d", len(args)))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	var ok bool
	var sideEffectOutput *SideEffectOutput
	if sideEffectOutput, ok = args[0].(*SideEffectOutput); !ok {
		err := errors.Join(ErrSideEffectInstanceCompleted, fmt.Errorf("SideEffectInstance onCompleted expected *SideEffectOutput"))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	if err := si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusCompleted)); err != nil {
		err := errors.Join(ErrSideEffectInstanceCompleted, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	logger.Debug(si.ctx, "side effect instance completed", "workflow_id", si.workflowID, "step_id", si.stepID)
	if si.future != nil {
		si.future.setResult(sideEffectOutput.Outputs)
	}

	return nil
}

var ErrSideEffectInstanceFailed = errors.New("failed to fail side effect instance")

func (si *SideEffectInstance) onFailed(_ context.Context, args ...interface{}) error {

	logger.Debug(si.ctx, "side effect instance failed", "workflow_id", si.workflowID, "step_id", si.stepID)

	if len(args) != 1 {
		err := errors.Join(ErrSideEffectInstanceFailed, fmt.Errorf("SideEffectInstance onFailed expected 1 argument, got %d", len(args)))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	err := args[0].(error)

	logger.Error(si.ctx, "side effect instance failed", "workflow_id", si.workflowID, "step_id", si.stepID)
	if si.future != nil {
		si.future.setError(errors.Join(ErrSideEffectFailed, err))
	}

	if si.ctx.Err() != nil {
		if errors.Is(si.ctx.Err(), context.Canceled) {
			if err := si.db.SetSideEffectExecutionProperties(
				si.executionID,
				SetSideEffectExecutionStatus(ExecutionStatusCancelled)); err != nil {
				err := errors.Join(ErrSideEffectInstanceFailed, err)
				logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
				return err
			}
			if err := si.db.SetSideEffectEntityProperties(
				si.entityID,
				SetSideEffectEntityStatus(StatusCancelled)); err != nil {
				err := errors.Join(ErrSideEffectInstanceFailed, err)
				logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
				return err
			}
			return nil
		}
	}

	if err := si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusFailed)); err != nil {
		err := errors.Join(ErrSideEffectInstanceFailed, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	return nil
}

var ErrSideEffectInstancePaused = errors.New("failed to pause side effect instance")

func (si *SideEffectInstance) onPaused(_ context.Context, _ ...interface{}) error {

	logger.Debug(si.ctx, "side effect instance paused", "workflow_id", si.workflowID, "step_id", si.stepID)

	if si.future != nil {
		si.future.setError(errors.Join(ErrSideEffectInstancePaused, ErrPaused))
	}

	var err error
	// Tell the orchestrator to manage the case
	if err = si.tracker.onPause(); err != nil {
		err := errors.Join(ErrSideEffectInstancePaused, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		if si.future != nil {
			si.future.setError(err)
		}
		return err
	}

	return nil
}

//////////////////////////////////// SAGA

var ErrSagaContext = errors.New("failed saga context")

func (ctx WorkflowContext) Saga(stepID string, saga *SagaDefinition) Future {
	var err error

	logger.Debug(ctx.ctx, "saga context", "workflow_id", ctx.workflowID, "step_id", stepID)

	if err = ctx.checkPause(); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrSagaContext, err)
		if !errors.Is(err, ErrPaused) {
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		}
		future.setError(err)
		return future
	}

	var existingSaga *SagaInstance
	// First check if we already have a saga instance for this stepID
	if existingSaga, err = ctx.tracker.findSagaInstance(stepID); err == nil {
		return existingSaga.future
	}

	future := NewRuntimeFuture()

	// Get existing saga entity if any through hierarchy
	var entityID int
	hierarchies, err := ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, stepID, EntitySaga)
	if err == nil && len(hierarchies) > 0 {
		// Reuse existing entity
		entityID = hierarchies[0].ChildEntityID
	} else {
		// Create new entity
		internalRetryPolicy := DefaultRetryPolicyInternal()
		entity := &SagaEntity{
			BaseEntity: BaseEntity{
				StepID:      stepID,
				Status:      StatusPending,
				Type:        EntitySaga,
				RunID:       ctx.runID,
				QueueID:     ctx.queueID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				RetryPolicy: *internalRetryPolicy,
				RetryState:  RetryState{Attempts: 0},
			},
			SagaData: &SagaData{},
		}

		if entityID, err = ctx.db.AddSagaEntity(entity, ctx.workflowID); err != nil {
			err := errors.Join(ErrSagaContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}
	}

	var data *SagaData
	if data, err = ctx.db.GetSagaDataByEntityID(entityID); err != nil {
		err := errors.Join(ErrSagaContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID, "entity_id", entityID)
		future.setError(err)
		return future
	}

	// Create saga instance without starting it
	sagaInstance := &SagaInstance{
		ctx:               ctx.ctx,
		db:                ctx.db,
		tracker:           ctx.tracker,
		state:             ctx.state,
		debug:             ctx.debug,
		saga:              saga,
		stepID:            stepID,
		runID:             ctx.runID,
		workflowID:        ctx.workflowID,
		entityID:          entityID,
		dataID:            data.ID,
		queueID:           ctx.queueID,
		future:            future,
		parentExecutionID: ctx.workflowExecutionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
	}

	future.setEntityID(entityID)

	// Add instance to tracker
	ctx.tracker.addSagaInstance(sagaInstance)

	// Only start if not already started/completed
	var status EntityStatus
	if err := ctx.db.GetSagaEntityProperties(entityID, GetSagaEntityStatus(&status)); err != nil {
		err := errors.Join(ErrSagaContext, fmt.Errorf("failed to get saga status: %w", err))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	if status == StatusPending {
		logger.Debug(ctx.ctx, "saga context start", "workflow_id", ctx.workflowID, "step_id", stepID)
		sagaInstance.Start()
	}

	return future
}

var ErrSagaContextCompensate = errors.New("failed to compensate saga context")

func (ctx WorkflowContext) CompensateSaga(stepID string) error {

	logger.Debug(ctx.ctx, "compensate saga context", "workflow_id", ctx.workflowID, "step_id", stepID)

	// Find the saga instance through the tracker
	sagaInst, err := ctx.tracker.findSagaInstance(stepID)
	if err != nil {
		err := errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to find saga for step %s: %w", stepID, err))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		return err
	}

	// Verify saga is in completed state
	var status EntityStatus
	if err := ctx.db.GetSagaEntityProperties(sagaInst.entityID, GetSagaEntityStatus(&status)); err != nil {
		err := errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to get saga status: %w", err))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		return err
	}
	if status != StatusCompleted {
		err := errors.Join(ErrSagaContextCompensate, fmt.Errorf("cannot compensate saga in %s state", status))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		return err
	}

	// Trigger the FSM compensation
	if err := sagaInst.fsm.Fire(TriggerCompensate, nil); err != nil {
		err := errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to trigger compensation: %w", err))
		logger.Debug(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		return err
	}

	return nil
}

// SagaInstance represents an instance of a saga execution.
// Saga that compensates should have the status COMPENSATED
// Saga that succeed should have the status COMPLETED
// Saga that panics during a compensation should have the status FAILED (that means it's very very critical)
// Saga cannot be retried, Saga cannot be paused
type SagaInstance struct {
	ctx     context.Context
	mu      deadlock.Mutex
	db      Database
	tracker InstanceTracker
	state   StateTracker
	debug   Debug

	handler HandlerInfo
	saga    *SagaDefinition
	fsm     *stateless.StateMachine
	future  Future

	stepID      string
	runID       int
	workflowID  int
	entityID    int
	executionID int
	dataID      int
	queueID     int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

var ErrSagaInstance = errors.New("failed saga instance")

func (si *SagaInstance) Start() {
	logger.Debug(si.ctx, "saga instance start", "workflow_id", si.workflowID, "step_id", si.stepID)
	si.mu.Lock()
	si.fsm = stateless.NewStateMachine(StateIdle)
	fsm := si.fsm
	si.mu.Unlock()

	fsm.Configure(StateIdle).
		Permit(TriggerStart, StateTransactions)

	fsm.Configure(StateTransactions).
		OnEntry(si.executeTransactions).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerCompensate, StateCompensations).
		Permit(TriggerFail, StateFailed)

	fsm.Configure(StateCompensations).
		OnEntry(si.executeCompensations).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed)

	fsm.Configure(StateCompleted).
		OnEntry(si.onCompleted).
		Permit(TriggerCompensate, StateCompensations)

	fsm.Configure(StateFailed).
		OnEntry(si.onFailed)

	logger.Debug(si.ctx, "saga instance start", "workflow_id", si.workflowID, "step_id", si.stepID)
	if err := fsm.Fire(TriggerStart); err != nil {
		err := errors.Join(ErrSagaInstance, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.future.setError(err)
	}
}

var ErrSagaInstanceExecute = errors.New("failed to execute saga instance")

func (si *SagaInstance) executeTransactions(_ context.Context, _ ...interface{}) (err error) {
	// Capture panics at the function level
	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			// Update last execution status to failed
			lastExec, _ := si.db.GetSagaExecutions(si.entityID)
			if len(lastExec) > 0 {
				lastStepID := lastExec[len(lastExec)-1].ID
				if updateErr := si.db.SetSagaExecutionProperties(
					lastStepID,
					SetSagaExecutionStatus(ExecutionStatusFailed),
					SetSagaExecutionError(fmt.Sprintf("panic: %v", r)),
					SetSagaExecutionStackTrace(stackTrace)); updateErr != nil {
					err := errors.Join(ErrSagaInstanceExecute, updateErr)
					logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
				}
			}

			// Trigger compensation with the panic error
			panicErr := fmt.Errorf("step panicked: %v", r)
			err := errors.Join(ErrSagaInstanceExecute, ErrSagaPanicked, panicErr)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			si.fsm.Fire(TriggerCompensate, err)
		}
	}()

	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusRunning)); err != nil {
		err := errors.Join(ErrSagaInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	// Execute transactions
	for stepIdx := 0; stepIdx < len(si.saga.Steps); stepIdx++ {
		stepExec := &SagaExecution{
			BaseExecution: BaseExecution{
				EntityID: si.entityID,
				Status:   ExecutionStatusRunning,
			},
			ExecutionType: ExecutionTypeTransaction,
			SagaExecutionData: &SagaExecutionData{
				StepIndex: stepIdx,
			},
		}

		var stepExecID int
		var err error
		if stepExecID, err = si.db.AddSagaExecution(stepExec); err != nil {
			err := errors.Join(ErrSagaInstanceExecute, err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			si.fsm.Fire(TriggerFail, err)
			return nil
		}

		step := si.saga.Steps[stepIdx]
		txContext := TransactionContext{
			ctx:         si.ctx,
			db:          si.db,
			executionID: stepExecID,
			stepIndex:   stepIdx,
		}

		err = step.Transaction(txContext)
		if err != nil {
			if updateErr := si.db.SetSagaExecutionProperties(
				stepExecID,
				SetSagaExecutionStatus(ExecutionStatusFailed),
				SetSagaExecutionError(err.Error())); updateErr != nil {
				err := errors.Join(ErrSagaInstanceExecute, updateErr)
				logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			}

			err := errors.Join(ErrSagaInstanceExecute, err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			// Pass the error to the compensation state
			si.fsm.Fire(TriggerCompensate, err)
			return nil
		}

		if err := si.db.SetSagaExecutionProperties(
			stepExecID,
			SetSagaExecutionStatus(ExecutionStatusCompleted)); err != nil {
			err := errors.Join(ErrSagaInstanceExecute, err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			si.fsm.Fire(TriggerFail, err)
			return nil
		}
	}

	// All steps completed successfully
	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusCompleted)); err != nil {
		err := errors.Join(ErrSagaInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	si.fsm.Fire(TriggerComplete)
	return nil
}

var ErrSagaInstanceCompensate = errors.New("failed to compensate saga instance")

func (si *SagaInstance) executeCompensations(_ context.Context, args ...interface{}) (err error) {
	logger.Debug(si.ctx, "saga instance compensate", "workflow_id", si.workflowID, "step_id", si.stepID)
	// Recover from panics in the compensation phase
	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			// Update last execution status to failed
			lastExec, _ := si.db.GetSagaExecutions(si.entityID)
			if len(lastExec) > 0 {
				lastStepID := lastExec[len(lastExec)-1].ID
				if updateErr := si.db.SetSagaExecutionProperties(
					lastStepID,
					SetSagaExecutionStatus(ExecutionStatusFailed),
					SetSagaExecutionError(fmt.Sprintf("panic: %v", r)),
					SetSagaExecutionStackTrace(stackTrace)); updateErr != nil {
					err := errors.Join(ErrSagaInstanceCompensate, updateErr)
					logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)

				}
			}

			// Set saga to failed status
			if err := si.db.SetSagaEntityProperties(
				si.entityID,
				SetSagaEntityStatus(StatusFailed)); err != nil {
				err := errors.Join(ErrSagaInstanceCompensate, fmt.Errorf("compensation panicked: %v", r))
				logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			}

			err := errors.Join(ErrSagaInstanceCompensate, ErrSagaPanicked, fmt.Errorf("compensation panicked: %v", r))
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			// Critical failure - compensation panicked
			si.fsm.Fire(TriggerFail, err)
		}
	}()

	// Extract the original transaction error if it was passed
	var transactionErr error
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			transactionErr = err
		}
	}

	executions, err := si.db.GetSagaExecutions(si.entityID)
	if err != nil {
		err := errors.Join(ErrSagaInstanceCompensate, transactionErr, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	// Filter and sort completed transaction executions
	var completedTransactions []*SagaExecution
	for _, exec := range executions {
		if exec.Status == ExecutionStatusCompleted && exec.ExecutionType == ExecutionTypeTransaction {
			completedTransactions = append(completedTransactions, exec)
		}
	}

	// Sort by StepIndex in descending order for reverse compensation
	sort.Slice(completedTransactions, func(i, j int) bool {
		return completedTransactions[i].SagaExecutionData.StepIndex > completedTransactions[j].SagaExecutionData.StepIndex
	})

	compensationFailed := false
	var compensationErr error

	// Execute compensations in reverse order
	for _, txExec := range completedTransactions {
		stepIdx := txExec.SagaExecutionData.StepIndex

		compExec := &SagaExecution{
			BaseExecution: BaseExecution{
				EntityID: si.entityID,
				Status:   ExecutionStatusRunning,
			},
			ExecutionType: ExecutionTypeCompensation,
			SagaExecutionData: &SagaExecutionData{
				StepIndex: stepIdx,
			},
		}

		var compExecID int
		if compExecID, err = si.db.AddSagaExecution(compExec); err != nil {
			compensationFailed = true
			compensationErr = errors.Join(compensationErr, err)
			continue
		}

		step := si.saga.Steps[stepIdx]
		compContext := CompensationContext{
			ctx:         si.ctx,
			db:          si.db,
			executionID: txExec.ID,
			stepIndex:   stepIdx,
		}

		err = step.Compensation(compContext)
		if err != nil {
			compensationFailed = true
			compensationErr = errors.Join(compensationErr, err)

			if updateErr := si.db.SetSagaExecutionProperties(
				compExecID,
				SetSagaExecutionStatus(ExecutionStatusFailed),
				SetSagaExecutionError(err.Error())); updateErr != nil {

			}
			continue
		}

		if err := si.db.SetSagaExecutionProperties(
			compExecID,
			SetSagaExecutionStatus(ExecutionStatusCompleted)); err != nil {
			compensationFailed = true
			compensationErr = errors.Join(compensationErr, err)
		}
	}

	if compensationFailed {
		// Set saga to failed status since compensation failed
		if err := si.db.SetSagaEntityProperties(
			si.entityID,
			SetSagaEntityStatus(StatusFailed)); err != nil {
			err := errors.Join(ErrSagaInstanceCompensate, transactionErr, compensationErr, err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			si.fsm.Fire(TriggerFail, err)
			return nil
		}

		err := errors.Join(ErrSagaInstanceCompensate, transactionErr, compensationErr)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	// All compensations completed successfully
	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusCompensated)); err != nil {
		err := errors.Join(ErrSagaInstanceCompensate, transactionErr, compensationErr, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	logger.Error(si.ctx, "saga instance compensate completed", "workflow_id", si.workflowID, "step_id", si.stepID)

	si.fsm.Fire(TriggerComplete, transactionErr)
	return nil
}

var ErrSagaInstanceCompleted = errors.New("failed to complete saga instance")

func (si *SagaInstance) onCompleted(_ context.Context, args ...interface{}) error {
	logger.Debug(si.ctx, "saga instance completed", "workflow_id", si.workflowID, "step_id", si.stepID)
	var status EntityStatus
	if err := si.db.GetSagaEntityProperties(
		si.entityID,
		GetSagaEntityStatus(&status)); err != nil {
		err := errors.Join(ErrSagaInstanceCompleted, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	if si.future != nil {
		if status == StatusCompensated {
			// If we have a transaction error from the previous state, include it
			var transactionErr error
			if len(args) > 0 {
				if err, ok := args[0].(error); ok {
					transactionErr = err
				}
			}
			err := errors.Join(ErrSagaInstanceCompleted, ErrSagaCompensated, transactionErr)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
			si.future.setError(err)
		} else {
			si.future.setResult(nil)
		}
	}
	logger.Debug(si.ctx, "saga instance completed", "workflow_id", si.workflowID, "step_id", si.stepID)
	return nil
}

var ErrSagaInstanceFailed = errors.New("failed to fail saga instance")

func (si *SagaInstance) onFailed(_ context.Context, args ...interface{}) error {
	logger.Debug(si.ctx, "saga instance failed", "workflow_id", si.workflowID, "step_id", si.stepID)
	var err error
	if len(args) > 0 {
		if e, ok := args[0].(error); ok {
			err = e
		}
	}

	// Mark saga as failed
	if updateErr := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusFailed)); updateErr != nil {
		err := errors.Join(ErrSagaInstanceFailed, fmt.Errorf("failed to update saga status: %w", updateErr))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.workflowID, "step_id", si.stepID)
		return err
	}

	logger.Error(si.ctx, "saga instance failed", "workflow_id", si.workflowID, "step_id", si.stepID)
	if si.future != nil {
		si.future.setError(errors.Join(ErrSagaFailed, err))
	}
	return nil
}

///////////////////////////////////////////////////// Sub Workflow

var ErrWorkflowContext = errors.New("failed workflow context")

// Workflow creates or retrieves a sub-workflow associated with this workflow context
func (ctx WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
	var err error

	// Check for pause first, consistent with other context methods
	if err = ctx.checkPause(); err != nil {
		future := NewRuntimeFuture()
		err := errors.Join(ErrWorkflowContext, err)
		if !errors.Is(err, ErrPaused) {
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		}
		future.setError(err)
		return future
	}

	if options == nil {
		options = &WorkflowOptions{
			Queue: ctx.options.Queue,
		}
	} else if options.Queue == "" {
		options.Queue = ctx.options.Queue
	}

	// Check if this is a cross-queue workflow
	if options != nil && options.Queue != ctx.options.Queue {

		// TODO: check if already exists and return the data, probably just move it below

		// If we have a cross-queue handler, use it
		if ctx.onCrossQueueWorkflow != nil {
			var workflowEntity *WorkflowEntity

			handler, err := ctx.registry.RegisterWorkflow(workflowFunc)
			if err != nil {
				future := NewRuntimeFuture()
				err := errors.Join(ErrWorkflowContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			if workflowEntity, err = prepareWorkflow(
				ctx.registry,
				ctx.db,
				handler.Handler,
				options,
				&preparationOptions{
					parentWorkflowID:          ctx.workflowID,
					parentWorkflowExecutionID: ctx.workflowExecutionID,
					parentRunID:               ctx.runID,
				},
				args...); err != nil {

				future := NewRuntimeFuture()
				err := errors.Join(ErrWorkflowContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			return ctx.onCrossQueueWorkflow(options.Queue, workflowEntity.ID, workflowFunc, options, args...)
		}
		// If no handler is set, return an error
		future := NewRuntimeFuture()
		err := errors.Join(ErrWorkflowContext, fmt.Errorf("cross-queue workflow execution not supported: from %s to %s", ctx.options.Queue, options.Queue))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	future := NewRuntimeFuture()

	// Register workflow on-the-fly (like other components)
	handler, err := ctx.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		err := errors.Join(ErrWorkflowContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	var workflowEntityID int

	// Check for existing hierarchies like we do for activities/sagas
	var hierarchies []*Hierarchy
	if hierarchies, err = ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, stepID, EntityWorkflow); err != nil {
		err := errors.Join(ErrWorkflowContext, err)
		if !errors.Is(err, ErrHierarchyNotFound) {
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}
	}

	// Convert inputs for storage
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		err := errors.Join(ErrWorkflowContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	// If we have existing hierarchies
	if len(hierarchies) > 0 {
		workflowEntityID = hierarchies[0].ChildEntityID

		logger.Debug(ctx.ctx, "workflow context existing", "workflow_id", ctx.workflowID, "step_id", stepID)

		// Check the latest execution for completion
		var workflowExecution *WorkflowExecution
		if workflowExecution, err = ctx.db.GetWorkflowExecutionLatestByEntityID(workflowEntityID); err != nil {
			err := errors.Join(ErrWorkflowContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		future.setEntityID(workflowEntityID)

		// If we have a completed execution, return its result
		if workflowExecution.Status == ExecutionStatusCompleted {

			var workflowExecutionData *WorkflowExecutionData
			if workflowExecutionData, err = ctx.db.GetWorkflowExecutionDataByExecutionID(workflowExecution.ID); err != nil {
				err := errors.Join(ErrWorkflowContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			outputs, err := convertOutputsFromSerialization(handler, workflowExecutionData.Outputs)
			if err != nil {
				err := errors.Join(ErrWorkflowContext, err)
				logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
				future.setError(err)
				return future
			}

			logger.Debug(ctx.ctx, "workflow context completed", "workflow_id", ctx.workflowID, "step_id", stepID)

			future.setEntityID(workflowEntityID)
			future.setResult(outputs)
			return future
		}
	} else {
		// Create new workflow entity as child
		logger.Debug(ctx.ctx, "workflow context new", "workflow_id", ctx.workflowID, "step_id", stepID)

		// Convert retry policy if provided
		var internalRetryPolicy *retryPolicyInternal
		if options != nil && options.RetryPolicy != nil {
			internalRetryPolicy = ToInternalRetryPolicy(options.RetryPolicy)
		} else {
			internalRetryPolicy = DefaultRetryPolicyInternal()
		}

		// Use parent's queue if not specified
		var queueName string = ctx.options.Queue
		if options != nil && options.Queue != "" {
			queueName = options.Queue
		}

		var queue *Queue
		if queue, err = ctx.db.GetQueueByName(queueName); err != nil {
			err := errors.Join(ErrWorkflowContext, fmt.Errorf("failed to get queue %s: %w", queueName, err))
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		// Create the workflow entity
		if workflowEntityID, err = ctx.db.AddWorkflowEntity(&WorkflowEntity{
			BaseEntity: BaseEntity{
				RunID:       ctx.runID,
				QueueID:     queue.ID,
				Type:        EntityWorkflow,
				HandlerName: handler.HandlerName,
				Status:      StatusPending,
				RetryPolicy: *internalRetryPolicy,
				RetryState:  RetryState{Attempts: 0},
				StepID:      stepID,
			},
			WorkflowData: &WorkflowData{
				Inputs:    inputBytes,
				Paused:    false,
				Resumable: false,
				IsRoot:    false, // This is a child workflow
			},
		}); err != nil {
			err := errors.Join(ErrWorkflowContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}

		future.setEntityID(workflowEntityID)

		// Add hierarchy relationship
		if _, err = ctx.db.AddHierarchy(&Hierarchy{
			RunID:             ctx.runID,
			ParentStepID:      ctx.stepID,
			ParentEntityID:    ctx.workflowID,
			ParentExecutionID: ctx.workflowExecutionID,
			ChildStepID:       stepID,
			ChildEntityID:     workflowEntityID,
			ParentType:        EntityWorkflow,
			ChildType:         EntityWorkflow,
		}); err != nil {
			err := errors.Join(ErrWorkflowContext, err)
			logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
			future.setError(err)
			return future
		}
	}

	var data *WorkflowData
	if data, err = ctx.db.GetWorkflowDataByEntityID(workflowEntityID); err != nil {
		err := errors.Join(ErrWorkflowContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	workflowInstance := &WorkflowInstance{
		ctx:      ctx.ctx,
		db:       ctx.db,
		tracker:  ctx.tracker,
		state:    ctx.state,
		registry: ctx.registry,
		debug:    ctx.debug,

		parentExecutionID: ctx.workflowExecutionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,

		handler: handler,
		future:  future,

		workflowID:  workflowEntityID,
		stepID:      stepID,
		executionID: 0, // will be set later
		dataID:      data.ID,
		queueID:     ctx.queueID,
		runID:       ctx.runID,
		options:     ctx.options, // Inherit parent options unless overridden

		onCrossWorkflow: ctx.onCrossQueueWorkflow,
		onContinueAsNew: ctx.onContinueAsNew,
		onSignalNew:     ctx.onSignalNew,
	}

	// Override options if provided
	if options != nil {
		workflowInstance.options = *options
	}

	inputs, err := convertInputsFromSerialization(handler, inputBytes)
	if err != nil {
		err := errors.Join(ErrWorkflowContext, err)
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", stepID)
		future.setError(err)
		return future
	}

	logger.Debug(ctx.ctx, "workflow instance start", "workflow_id", ctx.workflowID, "step_id", stepID)
	ctx.tracker.addWorkflowInstance(workflowInstance)
	workflowInstance.Start(inputs)

	return future
}

// /////////////////////////////////////////////////// Signals

var ErrSignalContext = errors.New("failed signal context")

// Signal allows workflow to receive named signals with type-safe output
func (ctx WorkflowContext) Signal(name string, output interface{}) error {
	logger.Debug(ctx.ctx, "signal context", "workflow_id", ctx.workflowID, "step_id", ctx.stepID, "signal_name", name)

	// Validate output is a pointer
	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		err := errors.Join(ErrSignalContext, ErrMustPointer, fmt.Errorf("output must be a pointer"))
		logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", ctx.stepID, "signal_name", name)
		return err
	}

	outputType := reflect.TypeOf(output).Elem()

	// First check existing hierarchies for completed signals
	hierarchies, err := ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, name, EntitySignal)
	if err == nil && len(hierarchies) > 0 {
		for _, h := range hierarchies {
			var status EntityStatus
			if err := ctx.db.GetSignalEntityProperties(h.ChildEntityID,
				GetSignalEntityStatus(&status)); err == nil && status == StatusCompleted {
				// Get existing successful signal
				if exec, err := ctx.db.GetSignalExecutionLatestByEntityID(h.ChildEntityID); err == nil {
					if data, err := ctx.db.GetSignalExecutionDataByExecutionID(exec.ID); err == nil &&
						data != nil && len(data.Value) > 0 {

						// Convert stored value to requested type
						result, err := convertSingleOutputFromSerialization(outputType, data.Value)
						if err != nil {
							err := errors.Join(ErrSignalContext, fmt.Errorf("failed to decode signal value: %w", err))
							logger.Error(ctx.ctx, err.Error(), "workflow_id", ctx.workflowID, "step_id", ctx.stepID, "signal_name", name)
							return err
						}

						logger.Debug(ctx.ctx, "signal context already completed", "workflow_id", ctx.workflowID, "step_id", ctx.stepID, "signal_name", name)

						// Set output value
						reflect.ValueOf(output).Elem().Set(reflect.ValueOf(result))
						return nil
					}
				}
			}
		}
	}

	// Create new signal instance
	instance := &SignalInstance{
		ctx:     ctx.ctx,
		db:      ctx.db,
		tracker: ctx.tracker,
		state:   ctx.state,
		debug:   ctx.debug,

		runID:             ctx.runID,
		parentExecutionID: ctx.workflowExecutionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
		queueID:           ctx.queueID,

		output:     output,
		outputType: outputType,
		name:       name,
		onSignal:   ctx.onSignalNew,
		done:       make(chan error, 1),
	}

	logger.Debug(ctx.ctx, "signal instance start", "workflow_id", ctx.workflowID, "step_id", ctx.stepID, "signal_name", name)

	return instance.Start()
}

var ErrSignalInstance = errors.New("failed signal instance")

type SignalInstance struct {
	ctx     context.Context
	db      Database
	tracker InstanceTracker
	state   StateTracker
	debug   Debug
	fsm     *stateless.StateMachine
	mu      deadlock.Mutex

	runID             int
	entityID          int
	executionID       int
	name              string
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
	queueID           int

	output     interface{}
	outputType reflect.Type
	done       chan error

	onSignal signalNew
}

func (si *SignalInstance) Start() error {

	logger.Debug(si.ctx, "signal instance start", "workflow_id", si.runID, "step_id", si.parentStepID, "signal_name", si.name)

	si.mu.Lock()
	si.fsm = stateless.NewStateMachine(StateIdle)
	fsm := si.fsm
	si.mu.Unlock()

	fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateExecuting).
		OnEntry(si.executeSignal).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	fsm.Configure(StateCompleted).
		OnEntry(si.onCompleted)

	fsm.Configure(StateFailed).
		OnEntry(si.onFailed)

	fsm.Configure(StatePaused).
		OnEntry(si.onPaused).
		Permit(TriggerResume, StateExecuting)

	logger.Debug(si.ctx, "signal instance start fsm configured", "workflow_id", si.runID, "step_id", si.parentStepID, "signal_name", si.name)
	if err := fsm.Fire(TriggerStart); err != nil {
		err := errors.Join(ErrSignalInstance, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID, "signal_name", si.name)
		return err
	}

	return <-si.done
}

var ErrSignalInstanceExecute = errors.New("failed to execute signal instance")

func (si *SignalInstance) executeSignal(_ context.Context, _ ...interface{}) error {

	logger.Debug(si.ctx, "signal instance execute", "workflow_id", si.runID, "step_id", si.parentStepID, "signal_name", si.name)

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			si.mu.Lock()
			err := errors.Join(ErrSignalInstanceExecute, fmt.Errorf("signal panicked: %v", r))
			logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID, "signal_name", si.name)
			si.done <- err
			si.fsm.Fire(TriggerFail, err)
			si.mu.Unlock()
		}
	}()

	// Create entity
	entityID, err := si.db.AddSignalEntity(&SignalEntity{
		BaseEntity: BaseEntity{
			RunID:       si.runID,
			QueueID:     si.queueID,
			Type:        EntitySignal,
			HandlerName: si.name,
			Status:      StatusPending,
			RetryPolicy: retryPolicyInternal{MaxAttempts: 0},
			RetryState:  RetryState{Attempts: 0},
			StepID:      si.name,
		},
		SignalData: &SignalData{
			Name: si.name,
		},
	}, si.parentEntityID)

	if err != nil {
		err := errors.Join(ErrSignalInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		si.done <- err
		si.fsm.Fire(TriggerFail, err)
		return nil
	}
	si.entityID = entityID

	// Create execution
	exec := &SignalExecution{
		BaseExecution: BaseExecution{
			EntityID: si.entityID,
			Status:   ExecutionStatusRunning,
		},
		SignalExecutionData: &SignalExecutionData{},
	}

	if si.executionID, err = si.db.AddSignalExecution(exec); err != nil {
		err := errors.Join(ErrSignalInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		si.done <- err
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	// Create hierarchy
	if _, err = si.db.AddHierarchy(&Hierarchy{
		RunID:             si.runID,
		ParentStepID:      si.parentStepID,
		ParentEntityID:    si.parentEntityID,
		ParentExecutionID: si.parentExecutionID,
		ChildStepID:       si.name,
		ChildEntityID:     si.entityID,
		ChildExecutionID:  si.executionID,
		ParentType:        EntityWorkflow,
		ChildType:         EntitySignal,
	}); err != nil {
		err := errors.Join(ErrSignalInstanceExecute, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
		return nil
	}

	if si.onSignal == nil {
		err := fmt.Errorf("signal handler not implemented")
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		si.done <- err
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	future := si.onSignal(si.parentEntityID, si.name)
	res, err := future.GetResults()
	if err != nil {
		err := fmt.Errorf("unable to process signal: %w", err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		si.done <- err
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	if len(res) > 0 {
		valueBytes, err := convertOutputsForSerialization(res)
		if err != nil {
			err := fmt.Errorf("failed to serialize signal value: %w", err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
			si.done <- err
			si.fsm.Fire(TriggerFail, err)
			return nil
		}

		// Store value
		if err := si.db.SetSignalExecutionDataProperties(
			si.executionID,
			SetSignalExecutionDataValue(valueBytes[0])); err != nil {
			err := fmt.Errorf("failed to store signal value: %w", err)
			logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
			si.done <- err
			si.fsm.Fire(TriggerFail, err)
			return nil
		}

		// Convert and set output
		result, err := convertSingleOutputFromSerialization(si.outputType, valueBytes[0])
		if err != nil {
			err := errors.Join(ErrSignalInstanceExecute, ErrSerialization, fmt.Errorf("failed to convert signal value: %w", err))
			logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
			si.done <- err
			si.fsm.Fire(TriggerFail, err)
			return nil
		}

		reflect.ValueOf(si.output).Elem().Set(reflect.ValueOf(result))
	}

	logger.Debug(si.ctx, "signal instance completed", "workflow_id", si.runID, "step_id", si.parentStepID)
	si.done <- nil
	si.fsm.Fire(TriggerComplete)
	return nil
}

var ErrSignalInstanceCompleted = errors.New("failed to complete signal instance")

func (si *SignalInstance) onCompleted(_ context.Context, args ...interface{}) error {
	logger.Debug(si.ctx, "signal instance completed", "workflow_id", si.runID, "step_id", si.parentStepID)
	if err := si.db.SetSignalEntityProperties(
		si.entityID,
		SetSignalEntityStatus(StatusCompleted)); err != nil {
		err := errors.Join(ErrSignalInstanceCompleted, fmt.Errorf("failed to set signal entity status: %w", err))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		return err
	}

	if err := si.db.SetSignalExecutionProperties(
		si.executionID,
		SetSignalExecutionStatus(ExecutionStatusCompleted)); err != nil {
		err := errors.Join(ErrSignalInstanceCompleted, fmt.Errorf("failed to set signal execution status: %w", err))
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		return err
	}

	return nil
}

var ErrSignalInstanceFailed = errors.New("failed to fail signal instance")

func (si *SignalInstance) onFailed(_ context.Context, args ...interface{}) error {
	logger.Debug(si.ctx, "signal instance failed", "workflow_id", si.runID, "step_id", si.parentStepID)
	var err error
	if len(args) > 0 {
		err = args[0].(error)
	}

	if si.ctx.Err() != nil {
		if errors.Is(si.ctx.Err(), context.Canceled) {
			if err := si.db.SetSignalExecutionProperties(
				si.executionID,
				SetSignalExecutionStatus(ExecutionStatusCancelled)); err != nil {
				err := errors.Join(ErrSignalInstanceFailed, err)
				logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
				return err
			}
			if err := si.db.SetSignalEntityProperties(
				si.entityID,
				SetSignalEntityStatus(StatusCancelled)); err != nil {
				err := errors.Join(ErrSignalInstanceFailed, err)
				logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
				return err
			}
			return nil
		}
	}

	if err := si.db.SetSignalEntityProperties(
		si.entityID,
		SetSignalEntityStatus(StatusFailed)); err != nil {
		err := errors.Join(ErrSignalInstanceFailed, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		return err
	}

	if err := si.db.SetSignalExecutionProperties(
		si.executionID,
		SetSignalExecutionStatus(ExecutionStatusFailed),
		SetSignalExecutionError(err.Error())); err != nil {
		err := errors.Join(ErrSignalInstanceFailed, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		return err
	}

	return nil
}

var ErrSignalInstancePaused = errors.New("failed to pause signal instance")

func (si *SignalInstance) onPaused(_ context.Context, _ ...interface{}) error {
	logger.Debug(si.ctx, "signal instance paused", "workflow_id", si.runID, "step_id", si.parentStepID)
	var err error
	if err = si.tracker.onPause(); err != nil {
		err := errors.Join(ErrSignalInstancePaused, err)
		logger.Error(si.ctx, err.Error(), "workflow_id", si.runID, "step_id", si.parentStepID)
		return err
	}
	return nil
}
