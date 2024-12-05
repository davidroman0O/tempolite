package tempolite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
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

func init() {
	maxprocs.Set()
	deadlock.Opts.DeadlockTimeout = time.Second * 2
	deadlock.Opts.OnPotentialDeadlock = func() {
		log.Println("POTENTIAL DEADLOCK DETECTED!")
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

func (tc *TransactionContext) Store(key string, value interface{}) error {
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return ErrMustPointer
	}

	buf := new(bytes.Buffer)
	if err := rtl.Encode(reflect.ValueOf(value).Elem().Interface(), buf); err != nil {
		return errors.Join(ErrEncoding, err)
	}

	_, err := tc.db.SetSagaValue(tc.executionID, key, buf.Bytes())
	return err
}

func (tc *TransactionContext) Load(key string, value interface{}) error {
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return ErrMustPointer
	}

	data, err := tc.db.GetSagaValueByExecutionID(tc.executionID, key)
	if err != nil {
		return errors.Join(ErrSagaGetValue, err)
	}

	return rtl.Decode(bytes.NewBuffer(data), value)
}

type CompensationContext struct {
	ctx         context.Context
	db          Database
	contextID   int
	executionID int
	stepIndex   int
}

func (cc *CompensationContext) Load(key string, value interface{}) error {
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return ErrMustPointer
	}

	data, err := cc.db.GetSagaValueByExecutionID(cc.executionID, key)
	if err != nil {
		return errors.Join(ErrSagaGetValue, err)
	}

	return rtl.Decode(bytes.NewBuffer(data), value)
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
			return nil, errors.Join(ErrTransactionMethodNotFound, fmt.Errorf("not found for step %d", i))
		}
		if !compensationOk {
			return nil, errors.Join(ErrCompensationMethodNotFound, fmt.Errorf("not found for step %d", i))
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
	f.mu.Unlock()
}

func (f *RuntimeFuture) setResult(results []interface{}) {

	f.mu.Lock()
	f.results = results
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

func (f *RuntimeFuture) setError(err error) {

	f.mu.Lock()
	f.err = err
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

func (f *RuntimeFuture) GetResults() ([]interface{}, error) {
	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.err != nil {
		return nil, f.err
	}

	// Return a copy of results to prevent modification
	results := make([]interface{}, len(f.results))
	copy(results, f.results)
	return results, nil
}

var ErrGetResults = errors.New("cannot get results")

func (f *RuntimeFuture) Get(out ...interface{}) error {
	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	if len(out) > len(f.results) {
		return errors.Join(ErrGetResults, fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results)))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return errors.Join(ErrGetResults, fmt.Errorf("output parameter %d must be a pointer", i))
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return errors.Join(ErrGetResults, fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i))
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

var ErrOrchestratorSagaInstanceNotFound = errors.New("saga instance not found")

func (o *Orchestrator) findSagaInstance(stepID string) (*SagaInstance, error) {
	for _, saga := range o.sagas {
		if saga.stepID == stepID {
			return saga, nil
		}
	}
	return nil, errors.Join(ErrOrchestratorSagaInstanceNotFound, fmt.Errorf("not found step %s", stepID))
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

func (o *Orchestrator) GetStatus(id int) (EntityStatus, error) {
	o.mu.Lock()
	var status EntityStatus
	if err := o.db.GetWorkflowEntityProperties(id, GetWorkflowEntityStatus(&status)); err != nil {

		return "", err
	}
	o.mu.Unlock()
	return status, nil
}

func (o *Orchestrator) WaitFor(id int, status EntityStatus) error {
	for {
		s, err := o.GetStatus(id)
		if err != nil {
			return err
		}
		if s == status {
			return nil
		}
		<-time.After(100 * time.Millisecond)
	}
}

// func (o *Orchestrator) WaitActive() {
// for o.active.Load() { // it is still active, you have to wait
// 	<-time.After(100 * time.Millisecond)
// }
// }

func (o *Orchestrator) isPaused() bool {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	return o.paused
}

func (o *Orchestrator) setUnpause() {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = false
}

// func (o *Orchestrator) isActive() bool {
// 	return o.active.Load()
// }

func (o *Orchestrator) canStackTrace() bool {
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

	var err error

	// resume, heh? was it paused?
	var status EntityStatus
	if err := o.db.GetWorkflowEntityProperties(entityID, GetWorkflowEntityStatus(&status)); err != nil {
		future := NewRuntimeFuture()
		future.setError(fmt.Errorf("failed to get workflow entity: %w", err))
		return future
	}

	if status != StatusPaused {
		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrPaused, ErrWorkflowPaused))
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
		future.setError(fmt.Errorf("failed to get workflow entity: %w", err))
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
		future.setError(errors.Join(ErrRegistryHandlerNotFound, fmt.Errorf("handler %s not found", workflowEntity.HandlerName)))
		return future
	}

	inputs, err := convertInputsFromSerialization(handler, workflowEntity.WorkflowData.Inputs)
	if err != nil {
		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrSerialization, fmt.Errorf("failed to convert inputs: %w", err)))
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
		future.setError(fmt.Errorf("failed to get workflow data properties: %w", err))
		return future
	}

	// Not paused anymore
	if err = o.db.SetWorkflowEntityProperties(
		entityID,
		SetWorkflowEntityStatus(StatusRunning),
	); err != nil {
		future.setError(fmt.Errorf("failed to set workflow entity status: %w", err))
		return future
	}

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

	// Register workflow if needed
	handler, err := registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		return nil, errors.Join(ErrPreparation, ErrRegistryRegisteration, fmt.Errorf("failed to register workflow: %w", err))
	}

	isRoot := true

	// Convert inputs for storage
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		return nil, errors.Join(ErrPreparation, ErrSerialization, fmt.Errorf("failed to serialize inputs: %w", err))
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
		// Create or get Run
		if runID, err = db.AddRun(&Run{
			Status: RunStatusPending,
		}); err != nil {
			return nil, errors.Join(ErrPreparation, fmt.Errorf("failed to create run: %w", err))
		}
	} else {
		runID = opttionsPreparation.parentRunID
	}

	// Convert retry policy if provided
	var retryPolicy *retryPolicyInternal
	if workflowOptions != nil && workflowOptions.RetryPolicy != nil {
		rp := workflowOptions.RetryPolicy
		retryPolicy = &retryPolicyInternal{
			MaxAttempts: rp.MaxAttempts,
			MaxInterval: rp.MaxInterval.Nanoseconds(),
		}
	} else {
		// Default retry policy
		retryPolicy = DefaultRetryPolicyInternal()
	}

	var queueName string = DefaultQueue
	if workflowOptions != nil && workflowOptions.Queue != "" {
		queueName = workflowOptions.Queue
	}

	var queue *Queue
	if queue, err = db.GetQueueByName(queueName); err != nil {
		return nil, errors.Join(ErrPreparation, fmt.Errorf("failed to get queue %s: %w", queueName, err))
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
		return nil, errors.Join(ErrPreparation, fmt.Errorf("failed to add workflow entity: %w", err))
	}

	return db.GetWorkflowEntity(workflowID)
}

var ErrOrchestratorExecution = errors.New("failed to execute workflow")

// Execute starts the execution of a workflow directly.
func (o *Orchestrator) Execute(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
	// Create entity and related records
	entity, err := prepareWorkflow(o.registry, o.db, workflowFunc, options, nil, args...)
	if err != nil {
		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrOrchestratorExecution, err))
		return future
	}

	// Execute using the entity
	future, err := o.ExecuteWithEntity(entity.ID)
	if err != nil {
		f := NewRuntimeFuture()
		f.setEntityID(entity.ID)
		f.setError(errors.Join(ErrOrchestratorExecution, err))
		return f
	}

	return future
}

var ErrOrchestratorExecuteEntity = errors.New("failed to execute workflow with entity")

// ExecuteWithEntity starts a workflow using an existing entity ID
func (o *Orchestrator) ExecuteWithEntity(entityID int) (Future, error) {
	// Get the entity and verify it exists
	var err error

	var entity *WorkflowEntity
	if entity, err = o.db.GetWorkflowEntity(
		entityID,
		WorkflowEntityWithQueue(), // we need the edge to the queue
		WorkflowEntityWithData(),
	); err != nil {
		return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get workflow entity: %w", err))
	}

	var isRoot bool
	if err = o.db.GetWorkflowDataProperties(entityID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get workflow data properties: %w", err))
	}

	// Set orchestrator's runID from entity
	o.runID = entity.RunID

	// Verify it's a workflow
	if entity.Type != EntityWorkflow {
		return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("entity %d is not a workflow", entityID))
	}

	var ok bool
	var handlerInfo HandlerInfo
	if handlerInfo, ok = o.registry.GetWorkflow(entity.HandlerName); !ok {
		return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get handler info"))
	}

	// Convert inputs
	inputs, err := convertInputsFromSerialization(handlerInfo, entity.WorkflowData.Inputs)
	if err != nil {
		return nil, errors.Join(ErrOrchestratorExecuteEntity, ErrSerialization, fmt.Errorf("failed to convert inputs: %w", err))
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
		return nil, errors.Join(ErrOrchestratorExecuteEntity, fmt.Errorf("failed to get hierarchies: %w", err))
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

	var err error

	for _, workflow := range o.instances {

		// TODO: pause: we should call the orchestrator to pause everything is running or pending

		var status EntityStatus
		if err = o.db.GetWorkflowEntityProperties(workflow.workflowID, GetWorkflowEntityStatus(&status)); err != nil {
			return fmt.Errorf("failed to get workflow entity status: %w", err)
		}

		switch status {
		case StatusRunning, StatusPending:
			if err = o.db.
				SetWorkflowEntityProperties(
					workflow.workflowID,
					SetWorkflowEntityStatus(StatusPaused),
				); err != nil {
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			if err = o.db.
				SetWorkflowDataProperties(
					workflow.workflowID,
					SetWorkflowDataPaused(true),
					SetWorkflowDataResumable(false),
				); err != nil {
				return fmt.Errorf("failed to set workflow data paused: %w", err)
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

	// Start the FSM without holding wi.mu
	if err := fsm.Fire(TriggerStart, inputs); err != nil {
		if !errors.Is(err, ErrPaused) {
			wi.future.setError(errors.Join(ErrWorkflowInstance, err))
		}
	}

	var isRoot bool
	if err := wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return errors.Join(ErrWorkflowInstance, fmt.Errorf("failed to get workflow data properties: %w", err))
	}

	// free up the orchestrator
	if isRoot {
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

	if len(args) != 1 {
		return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("WorkflowInstance executeActivity expected 1 argument, got %d", len(args)))
	}

	var err error
	var inputs []interface{}
	var ok bool

	var maxAttempts uint64
	var maxInterval time.Duration
	var retryState RetryState

	maxAttempts, maxInterval = getRetryPolicyOrDefault(wi.options.RetryPolicy)

	if inputs, ok = args[0].([]interface{}); !ok {
		return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("WorkflowInstance executeActivity expected argument to be []interface{}, got %T", args[0]))
	}

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to get workflow data properties: %w", err))
	}

	if err := retry.Do(
		wi.ctx,
		retry.
			WithMaxRetries(
				maxAttempts,
				retry.NewConstant(maxInterval)),
		// only the execution might change
		func(ctx context.Context) error {

			// now we can create and set the execution id
			if wi.executionID, err = wi.db.AddWorkflowExecution(&WorkflowExecution{
				BaseExecution: BaseExecution{
					EntityID: wi.workflowID,
					Status:   ExecutionStatusRunning,
				},
				WorkflowExecutionData: &WorkflowExecutionData{},
			}); err != nil {
				return err
			}

			if !isRoot {
				// Check if hierarchy already exists for this execution
				hierarchies, err := wi.db.GetHierarchiesByChildEntity(wi.workflowID)
				if err != nil || len(hierarchies) == 0 {
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
						return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to add hierarchy: %w", err))
					}
				} else {
					// Update existing hierarchy with new execution ID
					hierarchy := hierarchies[0]
					hierarchy.ChildExecutionID = wi.executionID
					if err = wi.db.UpdateHierarchy(hierarchy); err != nil {
						return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to update hierarchy: %w", err))
					}
				}
			}

			// Run the real workflow
			workflowOutput, workflowErr := wi.runWorkflow(inputs)

			// success case
			if workflowErr == nil {
				// Success
				wi.mu.Lock()

				// if we detected a pause
				if workflowOutput.Paused {
					if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusPaused)); err != nil {
						wi.mu.Unlock()
						return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status paused: %w", err))
					}
					wi.fsm.Fire(TriggerPause)

				} else {
					// normal case
					if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
						wi.mu.Unlock()
						return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status to completed: %w", err))
					}

					wi.fsm.Fire(TriggerComplete, workflowOutput)

				}

				wi.mu.Unlock()

				return nil
			} else {
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
					return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow execution status real failure error: %w", err))
				}

				wi.mu.Unlock()
			}

			// Update the attempt counter
			retryState.Attempts++
			if err = wi.db.
				SetWorkflowEntityProperties(wi.workflowID,
					SetWorkflowEntityRetryState(retryState)); err != nil {
				return errors.Join(ErrWorkflowInstanceExecution, fmt.Errorf("failed to set workflow entity retry state: %w", err))
			}

			return retry.RetryableError(workflowErr) // directly return the workflow error, if the execution failed we will retry eventually
		}); err != nil {
		// Max attempts reached and not resumable
		wi.mu.Lock()
		// the whole workflow entity failed
		wi.fsm.Fire(TriggerFail, err)

		wi.mu.Unlock()
	}
	return nil
}

var ErrWorkflowInstanceRun = errors.New("failed to run workflow instance")

// YOU DONT CHANGE THE STATUS OF THE WORKFLOW HERE
// Whatever happen in the runWorkflow is interpreted for form the WorkflowOutput or an error
func (wi *WorkflowInstance) runWorkflow(inputs []interface{}) (outputs *WorkflowOutput, err error) {

	// Get the latest execution, which may be nil if none exist
	latestExecution, err := wi.db.GetWorkflowExecutionLatestByEntityID(wi.workflowID)
	if err != nil {
		// Only return if there's an actual error)
		return nil, errors.Join(ErrWorkflowInstanceRun, err)
	}

	// No need to run it
	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {

		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Outputs)
		if err != nil {

			return nil, errors.Join(ErrWorkflowInstanceRun, err)
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

		return nil, errors.Join(ErrWorkflowInstanceRun, err)
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

			}

			wi.mu.Lock()

			// allow the caller to know about the panic
			err = ErrWorkflowPanicked
			// due to the `err` it will trigger the fail state
			outputs = &WorkflowOutput{
				StrackTrace: &stackTrace,
			}

			wi.mu.Unlock()
		}
	}()

	select {
	case <-wi.ctx.Done():

		return nil, wi.ctx.Err()
	default:
	}

	results := reflect.ValueOf(wFunc).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)

		return nil, errors.Join(ErrWorkflowInstanceRun, err)
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

		// Update execution error
		if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionError(realError.Error())); err != nil {
			// wi.mu.Unlock()
			return nil, errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution status error: %w", err))
		}

		// Emptied the output data
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(nil)); err != nil {
			// wi.mu.Unlock()
			return nil, errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution data outputs: %w", err))
		}

		return nil, realError
	} else {
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

			return nil, errors.Join(ErrWorkflowInstanceRun, err)
		}

		// save result
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(outputBytes)); err != nil {
			// wi.mu.Unlock()
			return nil, errors.Join(ErrWorkflowInstanceRun, fmt.Errorf("failed to set workflow execution data outputs: %w", err))
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

	if len(args) != 1 {
		return errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("WorkflowInstance onCompleted expected 1 argument, got %d", len(args)))
	}

	var ok bool
	var workflowOutput *WorkflowOutput
	if workflowOutput, ok = args[0].(*WorkflowOutput); !ok {
		return errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("WorkflowInstance onCompleted expected argument to be *WorkflowOutput, got %T", args[0]))
	}

	var err error

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to get workflow data properties: %w", err))
	}

	// Handle ContinueAsNew
	// We re-create a new workflow entity to that related to that workflow
	// It will run on another worker on it's own, we are use preparing it
	if workflowOutput.ContinueAsNewOptions != nil {

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

			return err
		}

		// Use the callback if available
		// Otherwise, you will have to use ExecuteWithEntity to manually trigger the workflow
		if wi.onContinueAsNew != nil {
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
		}
		// end continue as new
	}

	if isRoot {

		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusCompleted)); err != nil {

			return err
		}

		if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {

			return err
		}

	}

	if err = wi.db.SetWorkflowEntityProperties(wi.workflowID, SetWorkflowEntityStatus(StatusCompleted)); err != nil {
		// wi.mu.Unlock()
		return errors.Join(ErrWorkflowInstanceCompleted, fmt.Errorf("failed to set workflow entity status: %w", err))
	}

	// Normal completion
	if wi.future != nil {
		wi.future.setResult(workflowOutput.Outputs)
	}

	return nil
}

var ErrWorkflowInstanceFailed = errors.New("failed to fail workflow instance")

func (wi *WorkflowInstance) onFailed(_ context.Context, args ...interface{}) error {

	var err error

	if len(args) != 1 {
		return errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("WorkfloInstance onFailed expected 1 argument, got %d", len(args)))
	}

	err = args[0].(error)

	joinedErr := errors.Join(ErrWorkflowFailed, err)

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to get workflow data properties: %w", err))
	}

	if wi.ctx.Err() != nil {
		if errors.Is(wi.ctx.Err(), context.Canceled) {
			if err := wi.db.SetWorkflowExecutionProperties(wi.workflowID, SetWorkflowExecutionStatus(ExecutionStatusCancelled)); err != nil {
				return errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set Workflow execution status cancelled: %w", err))
			}
			if err := wi.db.SetWorkflowEntityProperties(wi.workflowID, SetWorkflowEntityStatus(StatusCancelled)); err != nil {
				return errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set Workflow entity status: %w", err))
			}

			// If this is the root workflow, update the Run status to Failed
			if isRoot {
				// Entity is now a failure
				if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusCancelled)); err != nil {
					return errors.Join(ErrWorkflowInstanceFailed, err)
				}
			}

			if wi.future != nil {
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
		return errors.Join(ErrWorkflowInstanceFailed, fmt.Errorf("failed to set workflow entity status: %w", err))
	}

	// If this is the root workflow, update the Run status to Failed
	if isRoot {
		// Entity is now a failure
		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusFailed)); err != nil {
			return errors.Join(ErrWorkflowInstanceFailed, err)
		}
	}

	if wi.future != nil {
		wi.future.setError(errors.Join(ErrWorkflowInstanceFailed, joinedErr))
	}

	return nil
}

var ErrWorkflowInstancePaused = errors.New("failed to pause workflow instance")

func (wi *WorkflowInstance) onPaused(_ context.Context, _ ...interface{}) error {

	if wi.future != nil {
		wi.future.setError(ErrPaused)
	}
	var err error

	// Tell the orchestrator to manage the case
	if err = wi.tracker.onPause(); err != nil {
		if wi.future != nil {
			wi.future.setError(err)
		}
		return errors.Join(ErrWorkflowInstancePaused, fmt.Errorf("failed to pause orchestrator: %w", err))
	}

	return nil
}

///////////// Activity

var ErrActivityContext = errors.New("failed activity context")

func (ctx WorkflowContext) Activity(stepID string, activityFunc interface{}, options *ActivityOptions, args ...interface{}) Future {
	var err error

	if err = ctx.checkPause(); err != nil {

		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrActivityContext, err))
		return future
	}

	future := NewRuntimeFuture()

	// Register activity on-the-fly
	handler, err := ctx.registry.RegisterActivity(activityFunc)
	if err != nil {

		future.setError(errors.Join(ErrActivityContext, err))
		return future
	}

	var activityEntityID int

	// Get all hierarchies for this workflow+stepID combination
	var hierarchies []*Hierarchy
	if hierarchies, err = ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, stepID, EntityActivity); err != nil {
		if !errors.Is(err, ErrHierarchyNotFound) {
			future.setError(errors.Join(ErrActivityContext, err))
			return future
		}
	}

	// Convert inputs to [][]byte
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {

		future.setError(errors.Join(ErrActivityContext, err))
		return future
	}

	// If we have existing hierarchies
	if len(hierarchies) > 0 {
		// Check if we have multiple different entity IDs (shouldn't happen)
		activityEntityID = hierarchies[0].ChildEntityID
		for _, h := range hierarchies[1:] {
			if h.ChildEntityID != activityEntityID {

			}
		}

		// Check the latest execution for completion
		var activityExecution *ActivityExecution
		if activityExecution, err = ctx.db.GetActivityExecution(hierarchies[0].ChildExecutionID); err != nil {

			future.setError(errors.Join(ErrActivityContext, err))
			return future
		}

		// If we have a completed execution, return its result
		if activityExecution.Status == ExecutionStatusCompleted {

			var activityExecutionData *ActivityExecutionData
			if activityExecutionData, err = ctx.db.GetActivityExecutionData(activityExecution.ID); err != nil {

				future.setError(errors.Join(ErrActivityContext, err))
				return future
			}

			outputs, err := convertOutputsFromSerialization(handler, activityExecutionData.Outputs)
			if err != nil {

				future.setError(errors.Join(ErrActivityContext, err))
				return future
			}

			future.setResult(outputs)
			return future
		}

		// If not completed, reuse the existing activity entity for retry
		future.setEntityID(activityEntityID)
	} else {
		// No existing activity entity found, create a new one

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

			future.setError(errors.Join(ErrActivityContext, err))
			return future
		}

		future.setEntityID(activityEntityID)
	}

	var data *ActivityData
	if data, err = ctx.db.GetActivityDataByEntityID(activityEntityID); err != nil {

		future.setError(errors.Join(ErrActivityContext, err))
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

	}

	ctx.tracker.addActivityInstance(activityInstance)
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

	// Start the FSM
	if err := ai.fsm.Fire(TriggerStart, inputs); err != nil {
		ai.future.setError(errors.Join(ErrActivityInstance, err))
	}
}

var ErrActivityInstanceExecute = errors.New("failed to execute activity instance")

func (ai *ActivityInstance) executeActivity(_ context.Context, args ...interface{}) error {
	if len(args) != 1 {
		return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("ActivityInstance executeActivity expected 1 argument, got %d", len(args)))
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
		return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("ActivityInstance executeActivity expected argument to be []interface{}, got %T", args[0]))
	}

	if err := retry.Do(
		ai.ctx,
		retry.
			WithMaxRetries(
				maxAttempts,
				retry.NewConstant(maxInterval)),
		// only the execution might change
		func(ctx context.Context) error {

			// now we can create and set the execution id
			if ai.executionID, err = ai.db.AddActivityExecution(&ActivityExecution{
				BaseExecution: BaseExecution{
					EntityID: ai.entityID,
					Status:   ExecutionStatusRunning,
				},
				ActivityExecutionData: &ActivityExecutionData{},
			}); err != nil {
				ai.mu.Lock()

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
				return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to add hierarchy: %w", err))
			}

			// run the real activity
			activityOutput, activityErr := ai.runActivity(inputs)

			if activityErr == nil {

				ai.mu.Lock()

				if activityOutput.Paused {

					if err = ai.db.SetActivityExecutionProperties(
						ai.executionID,
						SetActivityExecutionStatus(ExecutionStatusPaused),
					); err != nil {
						ai.mu.Unlock()
						return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution status: %w", err))
					}
					ai.fsm.Fire(TriggerPause)

				} else {

					if err = ai.db.SetActivityExecutionProperties(
						ai.executionID,
						SetActivityExecutionStatus(ExecutionStatusCompleted),
					); err != nil {
						ai.mu.Unlock()
						return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution status: %w", err))
					}

					ai.fsm.Fire(TriggerComplete, activityOutput)
				}

				ai.mu.Unlock()

				return nil
			} else {

				if errors.Is(activityErr, context.Canceled) {
					// Context cancelled
					ai.mu.Lock()
					ai.fsm.Fire(TriggerFail, activityErr)
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
					return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activity execution status: %w", err))
				}

				ai.mu.Unlock()

			}

			// Update the attempt counter
			retryState.Attempts++
			if err = ai.db.
				SetActivityEntityProperties(ai.entityID,
					SetActivityEntityRetryState(retryState)); err != nil {
				return errors.Join(ErrActivityInstanceExecute, fmt.Errorf("failed to set activty entity retry state: %w", err))
			}

			return retry.RetryableError(activityErr) // directly return the activty error, if the execution failed we will retry eventually
		}); err != nil {
		// Max attempts reached and not resumable
		ai.mu.Lock()
		// the whole activty entity failed
		ai.fsm.Fire(TriggerFail, err)

		ai.mu.Unlock()
	}
	return nil
}

var ErrActivityInstanceRun = errors.New("failed to run activity instance")

// Sub-function within executeWithRetry
func (ai *ActivityInstance) runActivity(inputs []interface{}) (outputs *ActivityOutput, err error) {

	activityExecution, err := ai.db.GetActivityExecutionLatestByEntityID(ai.entityID)
	if err != nil {

		return nil, errors.Join(ErrActivityInstanceRun, err)
	}

	if activityExecution != nil && activityExecution.Status == ExecutionStatusCompleted &&
		activityExecution.ActivityExecutionData != nil &&
		activityExecution.ActivityExecutionData.Outputs != nil {

		outputsData, err := convertOutputsFromSerialization(ai.handler, activityExecution.ActivityExecutionData.Outputs)
		if err != nil {

			return nil, errors.Join(ErrActivityInstanceRun, err)
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

			}

			ai.mu.Lock()
			// allow the caller to know about the panic
			err = ErrActivityPanicked
			outputs = &ActivityOutput{
				StrackTrace: &stackTrace,
			}
			// due to the `err` it will trigger the fail state
			ai.mu.Unlock()
		}
	}()

	var activityInputs [][]byte
	if err = ai.db.GetActivityDataProperties(ai.entityID, GetActivityDataInputs(&activityInputs)); err != nil {

		return nil, errors.Join(ErrActivityInstanceRun, err)
	}

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	select {
	case <-ai.ctx.Done():

		return nil, errors.Join(ErrActivityInstanceRun, ai.ctx.Err())
	default:
	}

	results := reflect.ValueOf(activityFunc).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("activity %s should return at least an error", handler.HandlerName)

		return nil, errors.Join(ErrActivityInstanceRun, err)
	}

	errInterface := results[numOut-1].Interface()
	if errInterface != nil {

		errActivity := errInterface.(error)

		// Serialize error
		errorMessage := errActivity.Error()

		// Update execution error
		if err = ai.db.SetActivityExecutionProperties(
			ai.entityID,
			SetActivityExecutionError(errorMessage)); err != nil {
			return nil, errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution error: %w", err))
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(nil)); err != nil {
			return nil, errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution data outputs: %w", err))
		}

		return nil, errors.Join(ErrActivityInstanceRun, errActivity)
	} else {
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

			return nil, errors.Join(ErrActivityInstanceRun, err)
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(outputBytes)); err != nil {
			return nil, errors.Join(ErrActivityInstanceRun, fmt.Errorf("failed to set activity execution data outputs: %w", err))
		}

		return &ActivityOutput{
			Outputs: outputs,
		}, nil
	}
}

var ErrActivityInstanceCompleted = errors.New("failed to complete activity instance")

func (ai *ActivityInstance) onCompleted(_ context.Context, args ...interface{}) error {

	if len(args) != 1 {
		return errors.Join(ErrActivityInstanceCompleted, fmt.Errorf("activity instance onCompleted expected 1 argument, got %d", len(args)))
	}

	if ai.future != nil {
		activtyOutput := args[0].(*ActivityOutput)
		ai.future.setResult(activtyOutput.Outputs)
	}

	if err := ai.db.SetActivityEntityProperties(ai.entityID, SetActivityEntityStatus(StatusCompleted)); err != nil {
		return errors.Join(ErrActivityInstanceCompleted, fmt.Errorf("failed to set activity entity status: %w", err))
	}

	return nil
}

var ErrActivityInstanceFailed = errors.New("failed to fail activity instance")

func (ai *ActivityInstance) onFailed(_ context.Context, args ...interface{}) error {

	if len(args) != 1 {
		return errors.Join(ErrActivityInstanceFailed, fmt.Errorf("activity instance onFailed expected 1 argument, got %d", len(args)))
	}
	if ai.future != nil {
		err := args[0].(error)
		ai.future.setError(errors.Join(ErrActivityInstanceFailed, err))
	}

	if ai.ctx.Err() != nil {
		if errors.Is(ai.ctx.Err(), context.Canceled) {
			if err := ai.db.SetActivityExecutionProperties(ai.entityID, SetActivityExecutionStatus(ExecutionStatusCancelled)); err != nil {
				return errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity execution status: %w", err))
			}
			if err := ai.db.SetActivityEntityProperties(ai.entityID, SetActivityEntityStatus(StatusCancelled)); err != nil {
				return errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity entity status: %w", err))
			}
			return nil
		}
	}

	if err := ai.db.SetActivityEntityProperties(
		ai.entityID,
		SetActivityEntityStatus(StatusFailed),
	); err != nil {
		return errors.Join(ErrActivityInstanceFailed, fmt.Errorf("failed to set activity entity status: %w", err))
	}

	return nil
}

var ErrActivityInstancePaused = errors.New("failed to pause activity instance")

// In theory, the behavior of the pause doesn't exists
// The Activity is supposed to finish what it was doing, the developer own the responsibility to detect the cancellation of it's own context.
func (ai *ActivityInstance) onPaused(_ context.Context, _ ...interface{}) error {

	if ai.future != nil {
		ai.future.setError(errors.Join(ErrActivityInstancePaused, ErrPaused))
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

	if err = ctx.checkPause(); err != nil {

		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrActivityInstancePaused, err))
		return future
	}

	future := NewRuntimeFuture()

	// Register side effect on-the-fly
	handler, err := ctx.registry.RegisterSideEffect(sideEffectFunc)
	if err != nil {

		future.setError(errors.Join(ErrActivityInstancePaused, err))
		return future
	}

	var sideEffectEntityID int

	// First check if there's an existing side effect entity for this workflow and step
	var sideEffectEntities []*SideEffectEntity
	if sideEffectEntities, err = ctx.db.GetSideEffectEntities(ctx.workflowID, SideEffectEntityWithData()); err != nil {
		if !errors.Is(err, ErrSideEffectEntityNotFound) {
			future.setError(errors.Join(ErrActivityInstancePaused, err))
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
			future.setError(errors.Join(ErrActivityInstancePaused, err))
			return future
		}

		// Look for a successful execution
		for _, exec := range execs {
			if exec.Status == ExecutionStatusCompleted {
				// Get the execution data
				execData, err := ctx.db.GetSideEffectExecutionDataByExecutionID(exec.ID)
				if err != nil {
					future.setError(errors.Join(ErrActivityInstancePaused, err))
					return future
				}
				if execData != nil && len(execData.Outputs) > 0 {
					// Reuse the existing result
					outputs, err := convertOutputsFromSerialization(handler, execData.Outputs)
					if err != nil {
						future.setError(errors.Join(ErrActivityInstancePaused, err))
						return future
					}
					future.setEntityID(sideEffectEntityID)
					future.setResult(outputs)
					return future
				}
			}
		}
	}

	// No existing completed execution found, proceed with creating new entity/execution
	if entity == nil {
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

			future.setError(errors.Join(ErrActivityInstancePaused, err))
			return future
		}
	}

	future.setEntityID(sideEffectEntityID)

	var data *SideEffectData
	if data, err = ctx.db.GetSideEffectDataByEntityID(sideEffectEntityID); err != nil {

		future.setError(errors.Join(ErrActivityInstancePaused, err))
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

	// Create execution before trying side effect
	if si.executionID, err = si.db.AddSideEffectExecution(&SideEffectExecution{
		BaseExecution: BaseExecution{
			EntityID: si.entityID,
			Status:   ExecutionStatusRunning,
		},
		SideEffectExecutionData: &SideEffectExecutionData{},
	}); err != nil {
		if si.future != nil {
			si.future.setError(errors.Join(ErrSideEffectInstanceExecute, err))
		}
		return errors.Join(ErrSideEffectInstanceExecute, err)
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
		if si.future != nil {
			si.future.setError(fmt.Errorf("failed to add hierarchy: %w", err))
		}
		return fmt.Errorf("failed to add hierarchy: %w", err)
	}

	// Run side effect just once
	output, execErr := si.runSideEffect()

	si.mu.Lock()
	defer si.mu.Unlock()

	if execErr == nil {
		if err = si.db.SetSideEffectExecutionProperties(
			si.executionID,
			SetSideEffectExecutionStatus(ExecutionStatusCompleted)); err != nil {
			if si.future != nil {
				si.future.setError(errors.Join(ErrSideEffectInstanceExecute, err))
			}
			return errors.Join(ErrSideEffectInstanceExecute, err)
		}

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
		if si.future != nil {
			si.future.setError(execErr)
		}
		return errors.Join(ErrSideEffectInstanceExecute, execErr)
	}

	if err = si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusFailed)); err != nil {
		if si.future != nil {
			si.future.setError(execErr)
		}
		return errors.Join(ErrSideEffectInstanceExecute, execErr)
	}

	if si.future != nil {
		si.future.setError(execErr)
	}
	si.fsm.Fire(TriggerFail, execErr)
	return errors.Join(ErrSideEffectInstanceExecute, execErr)
}

var ErrSideEffectRun = errors.New("failed to run side effect")

func (si *SideEffectInstance) runSideEffect() (output *SideEffectOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			err = errors.Join(ErrSideEffectRun, ErrSideEffectPanicked)
		}
	}()

	select {
	case <-si.ctx.Done():
		return nil, errors.Join(ErrSideEffectRun, si.ctx.Err())
	default:
	}

	results := reflect.ValueOf(si.handler.Handler).Call(nil)
	outputs := make([]interface{}, len(results))
	for i, res := range results {
		outputs[i] = res.Interface()
	}

	outputBytes, err := convertOutputsForSerialization(outputs)
	if err != nil {
		return nil, errors.Join(ErrSideEffectRun, err)
	}

	if err = si.db.SetSideEffectExecutionDataProperties(
		si.entityID,
		SetSideEffectExecutionDataOutputs(outputBytes)); err != nil {
		return nil, errors.Join(ErrSideEffectRun, err)
	}

	return &SideEffectOutput{
		Outputs: outputs,
	}, nil
}

var ErrSideEffectInstanceCompleted = errors.New("failed to complete side effect instance")

func (si *SideEffectInstance) onCompleted(_ context.Context, args ...interface{}) error {

	if len(args) != 1 {
		return errors.Join(ErrSideEffectInstanceCompleted, fmt.Errorf("SideEffectInstance onCompleted expected 1 argument"))
	}

	var ok bool
	var sideEffectOutput *SideEffectOutput
	if sideEffectOutput, ok = args[0].(*SideEffectOutput); !ok {
		return errors.Join(ErrSideEffectInstanceCompleted, fmt.Errorf("SideEffectInstance onCompleted expected *SideEffectOutput"))
	}

	if err := si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusCompleted)); err != nil {
		return errors.Join(ErrSideEffectInstanceCompleted, err)
	}

	if si.future != nil {
		si.future.setResult(sideEffectOutput.Outputs)
	}

	return nil
}

var ErrSideEffectInstanceFailed = errors.New("failed to fail side effect instance")

func (si *SideEffectInstance) onFailed(_ context.Context, args ...interface{}) error {

	if len(args) != 1 {
		return errors.Join(ErrSideEffectInstanceFailed, fmt.Errorf("activity instance onFailed expected 1 argument, got %d", len(args)))
	}

	err := args[0].(error)

	if si.future != nil {
		si.future.setError(errors.Join(ErrSideEffectFailed, err))
	}

	if si.ctx.Err() != nil {
		if errors.Is(si.ctx.Err(), context.Canceled) {
			if err := si.db.SetSideEffectExecutionProperties(
				si.executionID,
				SetSideEffectExecutionStatus(ExecutionStatusCancelled)); err != nil {
				return errors.Join(ErrSideEffectInstanceFailed, err)
			}
			if err := si.db.SetSideEffectEntityProperties(
				si.entityID,
				SetSideEffectEntityStatus(StatusCancelled)); err != nil {
				return errors.Join(ErrSideEffectInstanceFailed, err)
			}
			return nil
		}
	}

	if err := si.db.SetSideEffectEntityProperties(
		si.entityID,
		SetSideEffectEntityStatus(StatusFailed)); err != nil {
		return errors.Join(ErrSideEffectInstanceFailed, err)
	}

	return nil
}

var ErrSideEffectInstancePaused = errors.New("failed to pause side effect instance")

func (si *SideEffectInstance) onPaused(_ context.Context, _ ...interface{}) error {

	if si.future != nil {
		si.future.setError(errors.Join(ErrSideEffectInstancePaused, ErrPaused))
	}

	var err error
	// Tell the orchestrator to manage the case
	if err = si.tracker.onPause(); err != nil {
		if si.future != nil {
			si.future.setError(errors.Join(ErrSideEffectInstancePaused, err))
		}
		return errors.Join(ErrSideEffectInstancePaused, fmt.Errorf("failed to pause orchestrator: %w", err))
	}

	return nil
}

//////////////////////////////////// SAGA

var ErrSagaContext = errors.New("failed saga context")

func (ctx WorkflowContext) Saga(stepID string, saga *SagaDefinition) Future {
	var err error

	if err = ctx.checkPause(); err != nil {

		future := NewRuntimeFuture()
		future.setError(errors.Join(ErrSagaContext, err))
		return future
	}

	// First check if we already have a saga instance for this stepID
	if existingSaga, err := ctx.tracker.findSagaInstance(stepID); err == nil {
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

			future.setError(errors.Join(ErrSagaContext, err))
			return future
		}
	}

	var data *SagaData
	if data, err = ctx.db.GetSagaDataByEntityID(entityID); err != nil {

		future.setError(errors.Join(ErrSagaContext, err))
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
		future.setError(errors.Join(ErrSagaContext, err))
		return future
	}

	if status == StatusPending {
		sagaInstance.Start()
	}

	return future
}

var ErrSagaContextCompensate = errors.New("failed to compensate saga context")

func (ctx WorkflowContext) CompensateSaga(stepID string) error {
	// Find the saga instance through the tracker
	sagaInst, err := ctx.tracker.findSagaInstance(stepID)
	if err != nil {
		return errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to find saga for step %s: %w", stepID, err))
	}

	// Verify saga is in completed state
	var status EntityStatus
	if err := ctx.db.GetSagaEntityProperties(sagaInst.entityID, GetSagaEntityStatus(&status)); err != nil {
		return errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to get saga status: %w", err))
	}
	if status != StatusCompleted {
		return errors.Join(ErrSagaContextCompensate, fmt.Errorf("cannot compensate saga in %s state", status))
	}

	// Trigger the FSM compensation
	if err := sagaInst.fsm.Fire(TriggerCompensate, nil); err != nil {
		return errors.Join(ErrSagaContextCompensate, fmt.Errorf("failed to trigger compensation: %w", err))
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

	if err := fsm.Fire(TriggerStart); err != nil {
		si.future.setError(errors.Join(ErrSagaInstance, err))
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

				}
			}

			// Trigger compensation with the panic error
			panicErr := fmt.Errorf("step panicked: %v", r)
			si.fsm.Fire(TriggerCompensate, errors.Join(ErrSagaInstanceExecute, ErrSagaPanicked, panicErr))
		}
	}()

	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusRunning)); err != nil {
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

			}

			// Pass the error to the compensation state
			si.fsm.Fire(TriggerCompensate, err)
			return nil
		}

		if err := si.db.SetSagaExecutionProperties(
			stepExecID,
			SetSagaExecutionStatus(ExecutionStatusCompleted)); err != nil {
			si.fsm.Fire(TriggerFail, err)
			return nil
		}
	}

	// All steps completed successfully
	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusCompleted)); err != nil {
		si.fsm.Fire(TriggerFail, err)
		return nil
	}

	si.fsm.Fire(TriggerComplete)
	return nil
}

var ErrSagaInstanceCompensate = errors.New("failed to compensate saga instance")

func (si *SagaInstance) executeCompensations(_ context.Context, args ...interface{}) (err error) {
	// Recover from panics in the compensation phase
	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {

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

				}
			}

			// Set saga to failed status
			if err := si.db.SetSagaEntityProperties(
				si.entityID,
				SetSagaEntityStatus(StatusFailed)); err != nil {

			}

			// Critical failure - compensation panicked
			si.fsm.Fire(TriggerFail, errors.Join(ErrSagaInstanceCompensate, ErrSagaPanicked, fmt.Errorf("compensation panicked: %v", r)))
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
		si.fsm.Fire(TriggerFail, errors.Join(ErrSagaInstanceCompensate, transactionErr, err))
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

		}

		si.fsm.Fire(TriggerFail, errors.Join(ErrSagaInstanceCompensate, transactionErr, compensationErr))
		return nil
	}

	// All compensations completed successfully
	if err := si.db.SetSagaEntityProperties(
		si.entityID,
		SetSagaEntityStatus(StatusCompensated)); err != nil {
		si.fsm.Fire(TriggerFail, errors.Join(ErrSagaInstanceCompensate, transactionErr, compensationErr, err))
		return nil
	}

	si.fsm.Fire(TriggerComplete, transactionErr)
	return nil
}

func (si *SagaInstance) onCompleted(_ context.Context, args ...interface{}) error {
	var status EntityStatus
	if err := si.db.GetSagaEntityProperties(
		si.entityID,
		GetSagaEntityStatus(&status)); err != nil {
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
			si.future.setError(errors.Join(ErrSagaCompensated, transactionErr))
		} else {
			si.future.setResult(nil)
		}
	}
	return nil
}

var ErrSagaInstanceFailed = errors.New("failed to fail saga instance")

func (si *SagaInstance) onFailed(_ context.Context, args ...interface{}) error {
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
		return errors.Join(ErrSagaInstanceFailed, fmt.Errorf("failed to update saga status: %w", updateErr))
	}

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
		future.setError(errors.Join(ErrWorkflowContext, err))
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
				future.setError(errors.Join(ErrWorkflowContext, err))
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
				future.setError(errors.Join(ErrWorkflowContext, err))
				return future
			}

			return ctx.onCrossQueueWorkflow(options.Queue, workflowEntity.ID, workflowFunc, options, args...)
		}
		// If no handler is set, return an error
		future := NewRuntimeFuture()
		future.setError(
			errors.Join(
				ErrWorkflowContext,
				fmt.Errorf("cross-queue workflow execution not supported: from %s to %s",
					ctx.options.Queue, options.Queue)))
		return future
	}

	future := NewRuntimeFuture()

	// Register workflow on-the-fly (like other components)
	handler, err := ctx.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		future.setError(errors.Join(ErrWorkflowContext, err))
		return future
	}

	var workflowEntityID int

	// Check for existing hierarchies like we do for activities/sagas
	var hierarchies []*Hierarchy
	if hierarchies, err = ctx.db.GetHierarchiesByParentEntityAndStep(ctx.workflowID, stepID, EntityWorkflow); err != nil {
		if !errors.Is(err, ErrHierarchyNotFound) {
			future.setError(errors.Join(ErrWorkflowContext, err))
			return future
		}
	}

	// Convert inputs for storage
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		future.setError(errors.Join(ErrWorkflowContext, err))
		return future
	}

	// If we have existing hierarchies
	if len(hierarchies) > 0 {
		workflowEntityID = hierarchies[0].ChildEntityID

		// Check the latest execution for completion
		var workflowExecution *WorkflowExecution
		if workflowExecution, err = ctx.db.GetWorkflowExecutionLatestByEntityID(workflowEntityID); err != nil {
			future.setError(errors.Join(ErrWorkflowContext, err))
			return future
		}

		// If we have a completed execution, return its result
		if workflowExecution.Status == ExecutionStatusCompleted {

			var workflowExecutionData *WorkflowExecutionData
			if workflowExecutionData, err = ctx.db.GetWorkflowExecutionDataByExecutionID(workflowExecution.ID); err != nil {
				future.setError(errors.Join(ErrWorkflowContext, err))
				return future
			}

			outputs, err := convertOutputsFromSerialization(handler, workflowExecutionData.Outputs)
			if err != nil {
				future.setError(errors.Join(ErrWorkflowContext, err))
				return future
			}

			future.setEntityID(workflowEntityID)
			future.setResult(outputs)
			return future
		}

		future.setEntityID(workflowEntityID)
	} else {
		// Create new workflow entity as child

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
			future.setError(fmt.Errorf("failed to get queue %s: %w", queueName, err))
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
			future.setError(errors.Join(ErrWorkflowContext, err))
			return future
		}

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
			future.setError(fmt.Errorf("failed to add hierarchy: %w", err))
			return future
		}
		future.setEntityID(workflowEntityID)
	}

	var data *WorkflowData
	if data, err = ctx.db.GetWorkflowDataByEntityID(workflowEntityID); err != nil {
		future.setError(errors.Join(ErrWorkflowContext, err))
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

		future.setError(errors.Join(ErrWorkflowContext, err))
		return future
	}

	ctx.tracker.addWorkflowInstance(workflowInstance)
	workflowInstance.Start(inputs)

	return future
}

// /////////////////////////////////////////////////// Signals

var ErrSignalContext = errors.New("failed signal context")

// Signal allows workflow to receive named signals with type-safe output
func (ctx WorkflowContext) Signal(name string, output interface{}) error {
	// Validate output is a pointer
	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return errors.Join(ErrSignalContext, fmt.Errorf("output must be a pointer"))
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
							return errors.Join(ErrSignalContext, fmt.Errorf("failed to decode signal value: %w", err))
						}

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

	if err := fsm.Fire(TriggerStart); err != nil {
		return errors.Join(ErrSignalInstance, fmt.Errorf("failed to start signal: %w", err))
	}

	return <-si.done
}

var ErrSignalInstanceExecute = errors.New("failed to execute signal instance")

func (si *SignalInstance) executeSignal(_ context.Context, _ ...interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if si.debug.canStackTrace() {
				fmt.Println(stackTrace)
			}

			si.mu.Lock()
			err := fmt.Errorf("signal panicked: %v", r)
			si.done <- err
			si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
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
		err = fmt.Errorf("failed to create signal entity: %w", err)
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
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
		err = fmt.Errorf("failed to add signal execution: %w", err)
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
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
		err = fmt.Errorf("failed to add hierarchy: %w", err)
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
		return nil
	}

	if si.onSignal == nil {
		err = fmt.Errorf("signal handler not implemented")
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
		return nil
	}

	future := si.onSignal(si.parentEntityID, si.name)
	res, err := future.GetResults()
	if err != nil {
		err = fmt.Errorf("unable to process signal: %w", err)
		si.done <- err
		si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
		return nil
	}

	if len(res) > 0 {
		valueBytes, err := convertOutputsForSerialization(res)
		if err != nil {
			err = fmt.Errorf("failed to serialize signal value: %w", err)
			si.done <- err
			si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
			return nil
		}

		// Store value
		if err := si.db.SetSignalExecutionDataProperties(
			si.executionID,
			SetSignalExecutionDataValue(valueBytes[0])); err != nil {
			err = fmt.Errorf("failed to store signal value: %w", err)
			si.done <- err
			si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
			return nil
		}

		// Convert and set output
		result, err := convertSingleOutputFromSerialization(si.outputType, valueBytes[0])
		if err != nil {
			err = fmt.Errorf("failed to convert signal value: %w", err)
			si.done <- err
			si.fsm.Fire(TriggerFail, errors.Join(ErrSignalInstanceExecute, err))
			return nil
		}

		reflect.ValueOf(si.output).Elem().Set(reflect.ValueOf(result))
	}

	si.done <- nil
	si.fsm.Fire(TriggerComplete)
	return nil
}

var ErrSignalInstanceCompleted = errors.New("failed to complete signal instance")

func (si *SignalInstance) onCompleted(_ context.Context, args ...interface{}) error {
	if err := si.db.SetSignalEntityProperties(
		si.entityID,
		SetSignalEntityStatus(StatusCompleted)); err != nil {
		return errors.Join(ErrSignalInstanceCompleted, fmt.Errorf("failed to set signal entity status: %w", err))
	}

	if err := si.db.SetSignalExecutionProperties(
		si.executionID,
		SetSignalExecutionStatus(ExecutionStatusCompleted)); err != nil {
		return errors.Join(ErrSignalInstanceCompleted, fmt.Errorf("failed to set signal execution status: %w", err))
	}

	return nil
}

var ErrSignalInstanceFailed = errors.New("failed to fail signal instance")

func (si *SignalInstance) onFailed(_ context.Context, args ...interface{}) error {
	var err error
	if len(args) > 0 {
		err = args[0].(error)
	}

	if si.ctx.Err() != nil {
		if errors.Is(si.ctx.Err(), context.Canceled) {
			if err := si.db.SetSignalExecutionProperties(
				si.executionID,
				SetSignalExecutionStatus(ExecutionStatusCancelled)); err != nil {
				return errors.Join(ErrSignalInstanceFailed, err)
			}
			if err := si.db.SetSignalEntityProperties(
				si.entityID,
				SetSignalEntityStatus(StatusCancelled)); err != nil {
				return errors.Join(ErrSignalInstanceFailed, err)
			}
			return nil
		}
	}

	if err := si.db.SetSignalEntityProperties(
		si.entityID,
		SetSignalEntityStatus(StatusFailed)); err != nil {
		return errors.Join(ErrSignalInstanceFailed, err)
	}

	if err := si.db.SetSignalExecutionProperties(
		si.executionID,
		SetSignalExecutionStatus(ExecutionStatusFailed),
		SetSignalExecutionError(err.Error())); err != nil {
		return errors.Join(ErrSignalInstanceFailed, err)
	}

	return nil
}

var ErrSignalInstancePaused = errors.New("failed to pause signal instance")

func (si *SignalInstance) onPaused(_ context.Context, _ ...interface{}) error {
	var err error
	if err = si.tracker.onPause(); err != nil {
		return errors.Join(ErrSignalInstancePaused, fmt.Errorf("failed to pause orchestrator: %w", err))
	}
	return nil
}
