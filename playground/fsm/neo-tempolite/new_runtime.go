package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/k0kubun/pp/v3"
	"github.com/qmuntal/stateless"
	"github.com/sasha-s/go-deadlock"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	maxprocs.Set()
	deadlock.Opts.DeadlockTimeout = time.Second * 2
	deadlock.Opts.OnPotentialDeadlock = func() {
		log.Println("POTENTIAL DEADLOCK DETECTED!")
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
		log.Printf("Goroutine stack dump:\n%s", buf)
	}
}

var (
	ErrWorkflowPanicked = errors.New("workflow panicked")
	ErrActivityPanicked = errors.New("activity panicked")
	ErrActivityFailed   = errors.New("activity failed")
	ErrWorkflowFailed   = errors.New("workflow failed")
)

type Future interface {
	Get(out ...interface{}) error
	setEntityID(entityID int)
	setError(err error)
	setResult(results []interface{})
	WorkflowID() int
}

// TransactionContext provides context for transaction execution in Sagas.
type TransactionContext struct {
	ctx context.Context
}

// CompensationContext provides context for compensation execution in Sagas.
type CompensationContext struct {
	ctx context.Context
}

type SagaStep interface {
	Transaction(ctx TransactionContext) (interface{}, error)
	Compensation(ctx CompensationContext) (interface{}, error)
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

// AddStep adds a saga step to the builder.
func (b *SagaDefinitionBuilder) AddStep(step SagaStep) *SagaDefinitionBuilder {
	b.steps = append(b.steps, step)
	return b
}

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
			return nil, fmt.Errorf("Transaction method not found for step %d", i)
		}
		if !compensationOk {
			return nil, fmt.Errorf("Compensation method not found for step %d", i)
		}

		// Use the actual type name for the handler
		typeName := stepType.Name()
		if isPtr {
			typeName = "*" + typeName
		}

		transactionInfo, err := analyzeMethod(transactionMethod, typeName)
		if err != nil {
			return nil, fmt.Errorf("error analyzing Transaction method for step %d: %w", i, err)
		}

		compensationInfo, err := analyzeMethod(compensationMethod, typeName)
		if err != nil {
			return nil, fmt.Errorf("error analyzing Compensation method for step %d: %w", i, err)
		}

		sagaInfo.TransactionInfo[i] = transactionInfo
		sagaInfo.CompensationInfo[i] = compensationInfo
	}

	return &SagaDefinition{
		Steps:       b.steps,
		HandlerInfo: sagaInfo,
	}, nil
}

func analyzeMethod(method reflect.Method, name string) (HandlerInfo, error) {
	methodType := method.Type

	if methodType.NumIn() < 2 {
		return HandlerInfo{}, fmt.Errorf("method must have at least two parameters (receiver and context)")
	}

	paramTypes := make([]reflect.Type, methodType.NumIn()-2)
	paramKinds := make([]reflect.Kind, methodType.NumIn()-2)
	for i := 2; i < methodType.NumIn(); i++ {
		paramTypes[i-2] = methodType.In(i)
		paramKinds[i-2] = methodType.In(i).Kind()
	}

	returnTypes := make([]reflect.Type, methodType.NumOut()-1)
	returnKinds := make([]reflect.Kind, methodType.NumOut()-1)
	for i := 0; i < methodType.NumOut()-1; i++ {
		returnTypes[i] = methodType.Out(i)
		returnKinds[i] = methodType.Out(i).Kind()
	}

	handlerName := fmt.Sprintf("%s.%s", name, method.Name)

	return HandlerInfo{
		HandlerName:     handlerName,
		HandlerLongName: HandlerIdentity(name),
		Handler:         method.Func.Interface(),
		ParamTypes:      paramTypes,
		ParamsKinds:     paramKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           methodType.NumIn() - 2,  // Exclude receiver and context
		NumOut:          methodType.NumOut() - 1, // Exclude error
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
}

func (ctx WorkflowContext) checkPause() error {
	if ctx.state.IsPaused() {
		log.Printf("WorkflowContext detected orchestrator is paused")
		return ErrPaused
	}
	return nil
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

	continueAsNew *ContinueAsNewError

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	ctx     context.Context
	mu      deadlock.Mutex
	db      Database
	tracker InstanceTracker
	state   StateTracker
	debug   Debug

	handler HandlerInfo

	// results []interface{}
	// err     error

	fsm *stateless.StateMachine

	options ActivityOptions
	future  Future

	runID       int
	stepID      string
	workflowID  int
	entityID    int
	executionID int
	// dataID      int
	queueID int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
	ctx   context.Context
	mu    deadlock.Mutex
	debug Debug

	handlerName    string
	handler        HandlerInfo
	returnTypes    []reflect.Type // since we're doing it on-the-fly
	sideEffectFunc interface{}

	results []interface{}
	err     error

	fsm    *stateless.StateMachine
	future *Future

	stepID      string
	workflowID  int
	entityID    int
	executionID int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

// SagaInstance represents an instance of a saga execution.
type SagaInstance struct {
	ctx   context.Context
	mu    deadlock.Mutex
	debug Debug

	saga        SagaDefinition
	fsm         *stateless.StateMachine
	future      Future
	currentStep int

	err error

	compensations []int // Indices of steps to compensate

	stepID      string
	workflowID  int
	executionID int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

// Future implementation of direct calling
type RuntimeFuture struct {
	mu         sync.Mutex
	results    []interface{}
	err        error
	done       chan struct{}
	workflowID int
	once       sync.Once
}

func (f *RuntimeFuture) WorkflowID() int {
	return f.workflowID
}

func NewRuntimeFuture() *RuntimeFuture {
	return &RuntimeFuture{
		done: make(chan struct{}),
	}
}

func (f *RuntimeFuture) setEntityID(entityID int) {
	f.mu.Lock()
	f.workflowID = entityID
	f.mu.Unlock()
}

func (f *RuntimeFuture) setResult(results []interface{}) {
	log.Printf("Future.setResult called with results: %v", results)
	f.mu.Lock()
	f.results = results
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

func (f *RuntimeFuture) setError(err error) {
	log.Printf("Future.setError called with error: %v", err)
	f.mu.Lock()
	f.err = err
	f.mu.Unlock()
	f.once.Do(func() {
		close(f.done)
	})
}

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
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}

// Future implementation using the database to track states
type DatabaseFuture struct {
	ctx      context.Context
	entityID int
	database Database
	results  []interface{}
	err      error
}

func NewDatabaseFuture(ctx context.Context, database Database) *DatabaseFuture {
	return &DatabaseFuture{
		ctx:      ctx,
		database: database,
	}
}

func (f *DatabaseFuture) setEntityID(entityID int) {
	f.entityID = entityID
}

func (f *DatabaseFuture) setError(err error) {
	f.err = err
}

func (f *DatabaseFuture) WorkflowID() int {
	return f.entityID
}

func (f *DatabaseFuture) Get(out ...interface{}) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return errors.New("context cancelled")
		case <-ticker.C:
			if f.entityID == 0 {
				continue
			}
			if f.err != nil {
				defer func() { // if we re-use the same future, we must start from a fresh state
					f.err = nil
					f.results = nil
				}()
				return f.err
			}
			if completed := f.checkCompletion(); completed {
				defer func() { // if we re-use the same future, we must start from a fresh state
					f.err = nil
					f.results = nil
				}()
				return f.handleResults(out...)
			}
			continue
		}
	}
}

func (f *DatabaseFuture) setResult(results []interface{}) {
	f.results = results
}

func (f *DatabaseFuture) handleResults(out ...interface{}) error {
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	if len(out) > len(f.results) {
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}

func (f *DatabaseFuture) checkCompletion() bool {
	// var err error
	// var entity *Entity
	// if entity, err = f.database.GetEntity(f.entityID); err != nil {
	// 	f.err = fmt.Errorf("entity not found: %d", f.entityID)
	// 	return true
	// }

	// fmt.Println("\t entity", entity.ID, " status: ", entity.Status)
	// var latestExec *Execution
	// // TODO: should we check for pause?
	// switch entity.Status {
	// case StatusCompleted:
	// 	// Get results from latest execution
	// 	if latestExec, err = f.database.GetLatestExecution(f.entityID); err != nil {
	// 		f.err = err
	// 		return true
	// 	}

	// 	var outputs []interface{}
	// 	if entity.HandlerInfo == nil {
	// 		f.err = fmt.Errorf("handler info not found for entity %d", entity.ID)
	// 		return true
	// 	}

	// 	switch entity.Type {
	// 	case EntityTypeWorkflow:
	// 		if latestExec.WorkflowExecutionData != nil && latestExec.WorkflowExecutionData.Outputs != nil {
	// 			outputs, err = convertOutputsFromSerialization(*entity.HandlerInfo, latestExec.WorkflowExecutionData.Outputs)
	// 			if err != nil {
	// 				f.err = err
	// 				return true
	// 			}
	// 			f.results = outputs
	// 		}
	// 	case EntityTypeActivity:
	// 		if latestExec.ActivityExecutionData != nil && latestExec.ActivityExecutionData.Outputs != nil {
	// 			outputs, err = convertOutputsFromSerialization(*entity.HandlerInfo, latestExec.ActivityExecutionData.Outputs)
	// 			if err != nil {
	// 				f.err = err
	// 				return true
	// 			}
	// 			f.results = outputs
	// 		}
	// 	case EntityTypeSideEffect:
	// 		if latestExec.SideEffectExecutionData != nil && latestExec.SideEffectExecutionData.Outputs != nil {
	// 			outputs, err = convertOutputsFromSerialization(*entity.HandlerInfo, latestExec.SideEffectExecutionData.Outputs)
	// 			if err != nil {
	// 				f.err = err
	// 				return true
	// 			}
	// 			f.results = outputs
	// 		}
	// 	default:
	// 		f.err = fmt.Errorf("unsupported entity type: %s", entity.Type)
	// 		return true
	// 	}

	// 	// If outputs are successfully retrieved
	// 	return true
	// case StatusFailed:
	// 	if latestExec, err = f.database.GetLatestExecution(f.entityID); err != nil {
	// 		f.err = err
	// 		return true
	// 	}

	// 	f.err = errors.New(latestExec.Error)

	// 	return true
	// case StatusPaused:
	// 	f.err = ErrPaused
	// 	return true
	// case StatusCancelled:
	// 	f.err = errors.New("workflow was cancelled")
	// 	return true
	// }

	return false
}

type StateTracker interface {
	IsPaused() bool
	TogglePause()
}

type Debug interface {
	canStackTrace() bool
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

	rootWf      *WorkflowInstance // root workflow instance
	instances   []*WorkflowInstance
	activities  []*ActivityInstance
	sideEffects []*SideEffectInstance
	sagas       []*SagaInstance

	err error

	paused   bool
	pausedMu deadlock.Mutex

	displayStackTrace bool
}

func NewOrchestrator(ctx context.Context, db Database, registry *Registry) *Orchestrator {
	log.Printf("NewOrchestrator called")
	ctx, cancel := context.WithCancel(ctx)
	o := &Orchestrator{
		db:       db,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,

		displayStackTrace: true,
	}
	return o
}

func (o *Orchestrator) IsPaused() bool {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	return o.paused
}

func (o *Orchestrator) TogglePause() {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = !o.paused
}

func (o *Orchestrator) canStackTrace() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.displayStackTrace
}

// Preparing the creation of a new root workflow instance so it can exists in the database, we might decide depending of which systems of used if we want to pull the workflows or execute it directly.
func (o *Orchestrator) prepareRootWorkflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (*WorkflowEntity, error) {

	// Register workflow if needed
	handler, err := o.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to register workflow: %w", err)
	}

	// Convert inputs for storage
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize inputs: %w", err)
	}

	// Create or get Run
	if o.runID == 0 {
		if o.runID, err = o.db.AddRun(&Run{
			Status: RunStatusPending,
		}); err != nil {
			return nil, fmt.Errorf("failed to create run: %w", err)
		}
	}

	pp.Println(options)

	// Convert retry policy if provided
	var retryPolicy *retryPolicyInternal
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		retryPolicy = &retryPolicyInternal{
			MaxAttempts:        rp.MaxAttempts,
			InitialInterval:    rp.InitialInterval.Nanoseconds(),
			BackoffCoefficient: rp.BackoffCoefficient,
			MaxInterval:        rp.MaxInterval.Nanoseconds(),
		}
	} else {
		// Default retry policy
		retryPolicy = DefaultRetryPolicyInternal()
	}

	var queueName string = DefaultQueue
	if options != nil && options.Queue != "" {
		queueName = options.Queue
	}

	var queue *Queue
	if queue, err = o.db.GetQueueByName(queueName); err != nil {
		return nil, fmt.Errorf("failed to get queue %s: %w", queueName, err)
	}

	pp.Println(queue)

	var workflowID int
	// We're not even storing the HandlerInfo since it is a runtime thing
	if workflowID, err = o.db.
		AddWorkflowEntity(
			&WorkflowEntity{
				BaseEntity: BaseEntity{
					HandlerName: handler.HandlerName,
					Status:      StatusPending,
					Type:        EntityWorkflow,
					QueueID:     queue.ID,
					StepID:      "root",
					RunID:       o.runID,
					RetryPolicy: *retryPolicy,
					RetryState:  RetryState{Attempts: 0},
				},
				WorkflowData: &WorkflowData{
					Inputs:    inputBytes,
					Paused:    false,
					Resumable: false,
					IsRoot:    true,
				},
			}); err != nil {
		return nil, fmt.Errorf("failed to add workflow entity: %w", err)
	}

	return o.db.GetWorkflowEntity(workflowID)
}

// Execute starts the execution of a workflow directly.
func (o *Orchestrator) Execute(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *RuntimeFuture {
	// Create entity and related records
	entity, err := o.prepareRootWorkflow(workflowFunc, options, args...)
	if err != nil {
		future := NewRuntimeFuture()
		future.setError(err)
		return future
	}

	pp.Println("getted", entity)

	// Execute using the entity
	future, err := o.ExecuteWithEntity(entity.ID)
	if err != nil {
		f := NewRuntimeFuture()
		f.setEntityID(entity.ID)
		f.setError(err)
		return f
	}

	return future
}

// ExecuteWithEntity starts a workflow using an existing entity ID
func (o *Orchestrator) ExecuteWithEntity(entityID int) (*RuntimeFuture, error) {
	// Get the entity and verify it exists
	var err error

	var entity *WorkflowEntity
	if entity, err = o.db.GetWorkflowEntity(
		entityID,
		WorkflowEntityWithQueue(), // we need the edge to the queue
		WorkflowEntityWithData(),
	); err != nil {
		return nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	var isRoot bool

	if err = o.db.GetWorkflowDataProperties(entityID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return nil, fmt.Errorf("failed to get workflow data properties: %w", err)
	}

	// Set orchestrator's runID from entity
	o.runID = entity.RunID

	// Verify it's a workflow
	if entity.Type != EntityWorkflow {
		return nil, fmt.Errorf("entity %d is not a workflow", entityID)
	}

	var ok bool
	var handlerInfo HandlerInfo
	if handlerInfo, ok = o.registry.GetWorkflow(entity.HandlerName); !ok {
		return nil, fmt.Errorf("failed to get handler info")
	}

	// Convert inputs
	inputs, err := convertInputsFromSerialization(handlerInfo, entity.WorkflowData.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert inputs: %w", err)
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
				MaxAttempts:        entity.RetryPolicy.MaxAttempts,
				InitialInterval:    time.Duration(entity.RetryPolicy.InitialInterval),
				BackoffCoefficient: entity.RetryPolicy.BackoffCoefficient,
				MaxInterval:        time.Duration(entity.RetryPolicy.MaxInterval),
			},
			Queue: entity.Edges.Queue.Name,
			// VersionOverrides: , // TODO: implement in WorkflowOption and save it
		},
	}

	// If this is a sub-workflow, set parent info from hierarchies
	var hierarchies []*Hierarchy
	if hierarchies, err = o.db.GetHierarchiesByChildEntity(entityID); err != nil {
		return nil, fmt.Errorf("failed to get hierarchies: %w", err)
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

	o.addWorkflowInstance(instance)

	// We don't start the FSM within the same goroutine as the caller
	if isRoot {
		go instance.Start(inputs) // root workflow OR ContinueAsNew starts in a goroutine
	} else {
		instance.Start(inputs) // sub-workflow directly starts
	}

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

func (o *Orchestrator) newRoot(root *WorkflowInstance) {
	o.rootWf = root
}

type InstanceTracker interface {
	addWorkflowInstance(wi *WorkflowInstance)
	addActivityInstance(ai *ActivityInstance)
	addSideEffectInstance(sei *SideEffectInstance)
	addSagaInstance(si *SagaInstance)
	newRoot(root *WorkflowInstance)
}

////////////////////////// WorkflowInstance

// Starting the workflow instance
// Ideally we're already within a goroutine
func (wi *WorkflowInstance) Start(inputs []interface{}) {

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

	log.Printf("Starting workflow: %s (Entity ID: %d)", wi.stepID, wi.workflowID)

	// Start the FSM without holding wi.mu
	if err := fsm.Fire(TriggerStart, inputs); err != nil {
		log.Printf("Error starting workflow: %v", err)
	}
}

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, args ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) executeWorkflow called", wi.stepID, wi.workflowID)
	inputs := args[0].([]interface{})
	return wi.executeWithRetry(inputs)
}

// YOU CAN ONLY CHANGE THE EXECUTION STATUS HERE
func (wi *WorkflowInstance) executeWithRetry(inputs []interface{}) error {
	var err error
	var retryState RetryState
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration
	var paused bool
	var resumable bool

	wi.mu.Lock()

	maxAttempts, initialInterval, backoffCoefficient, maxInterval = getRetryPolicyOrDefault(wi.options.RetryPolicy)

	if err = wi.db.GetWorkflowEntityProperties(wi.workflowID, GetWorkflowEntityRetryState(&retryState)); err != nil {
		return err
	}

	wi.mu.Unlock()

	for {
		wi.mu.Lock()

		// Check if paused
		if err = wi.db.GetWorkflowDataProperties(
			wi.workflowID,
			GetWorkflowDataPaused(&paused),
			GetWorkflowDataResumable(&resumable),
		); err != nil {
			wi.mu.Unlock()
			return fmt.Errorf("failed to get workflow execution data: %w", err)
		}

		if paused {
			log.Printf("WorkflowInstance %s is paused", wi.stepID)

			wi.fsm.Fire(TriggerPause)
			wi.mu.Unlock()
			return nil
		}
		wi.mu.Unlock()

		// Check if maximum attempts have been reached and not resumable
		wi.mu.Lock()
		if retryState.Attempts > maxAttempts && !resumable {

			// Only execution here
			if err = wi.db.
				SetWorkflowExecutionProperties(
					wi.workflowID,
					SetWorkflowExecutionStatus(ExecutionStatusFailed),
					SetWorkflowExecutionError(fmt.Errorf("workflow %s failed after %d attempts", wi.stepID, retryState.Attempts-1).Error()),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			wi.fsm.Fire(TriggerFail)
			log.Printf("Workflow %s (Entity ID: %d) failed after %d attempts", wi.stepID, wi.workflowID, retryState.Attempts-1)
			wi.mu.Unlock()
			return nil
		}
		wi.mu.Unlock()

		// now we can create and set the execution id
		if wi.executionID, err = wi.db.AddWorkflowExecution(&WorkflowExecution{
			BaseExecution: BaseExecution{
				EntityID: wi.workflowID,
				Status:   ExecutionStatusRunning,
			},
			WorkflowExecutionData: &WorkflowExecutionData{},
		}); err != nil {
			wi.mu.Lock()
			log.Printf("Error adding execution: %v", err)
			wi.fsm.Fire(TriggerFail)
			wi.mu.Unlock()
		}

		// Create hierarchy if parentExecutionID is available
		wi.mu.Lock()
		if wi.parentExecutionID != 0 {

			if _, err = wi.db.AddHierarchy(&Hierarchy{
				RunID: wi.runID,

				ParentEntityID:    wi.parentEntityID,
				ParentExecutionID: wi.parentExecutionID,

				ChildEntityID:    wi.workflowID,
				ChildExecutionID: wi.executionID,

				ChildStepID:  wi.stepID,
				ParentStepID: wi.parentStepID,

				ParentType: EntityWorkflow,
				ChildType:  EntityWorkflow,
			}); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to add hierarchy: %w", err)
			}

		}
		wi.mu.Unlock()

		// Now we have an execution, so we can update the RetryState
		wi.mu.Lock()
		if retryState.Attempts <= maxAttempts {
			if err = wi.db.
				SetWorkflowEntityProperties(
					wi.workflowID,
					SetWorkflowEntityRetryState(retryState),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution properties: %w", err)
			}
		}
		wi.mu.Unlock()

		wi.mu.Lock()
		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.workflowID, wi.executionID)
		wi.mu.Unlock()

		ouputs, workflowErr := wi.runWorkflow(inputs)
		if errors.Is(err, ErrPaused) {
			wi.mu.Lock()

			wi.fsm.Fire(TriggerPause)
			log.Printf("WorkflowInstance %s is paused post-runWorkflow", wi.stepID)
			wi.mu.Unlock()
			return nil
		}

		if workflowErr == nil {
			// Success
			wi.mu.Lock()

			if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			wi.fsm.Fire(TriggerComplete, ouputs)
			log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) completed successfully", wi.stepID, wi.workflowID, wi.executionID)
			wi.mu.Unlock()
			return nil
		} else {
			wi.mu.Lock()

			// you failed but maybe the next time it will be better
			if err = wi.db.SetWorkflowExecutionProperties(
				wi.executionID,
				SetWorkflowExecutionStatus(ExecutionStatusFailed),
				SetWorkflowExecutionError(workflowErr.Error()),
			); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			wi.mu.Unlock()

			if retryState.Attempts < maxAttempts || resumable {
				// Calculate next interval
				nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(retryState.Attempts-1)))
				if nextInterval > maxInterval {
					nextInterval = maxInterval
				}
				wi.mu.Lock()
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.workflowID, wi.executionID, retryState.Attempts+1, maxAttempts, nextInterval)
				wi.mu.Unlock()
				// try again
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached and not resumable
				wi.mu.Lock()
				// the whole workflow entity failed
				wi.fsm.Fire(TriggerFail, workflowErr)
				log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) failed after %d attempts", wi.stepID, wi.workflowID, wi.executionID, retryState.Attempts)
				wi.mu.Unlock()
				return nil
			}
		}

		// Increment attempt for the next loop iteration
		retryState.Attempts++
	}
}

// YOU DONT CHANGE THE STATUS OF THE WORKFLOW HERE
func (wi *WorkflowInstance) runWorkflow(inputs []interface{}) (outputs []interface{}, err error) {
	log.Printf("WorkflowInstance %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.workflowID, wi.executionID)

	// Get the latest execution, which may be nil if none exist
	latestExecution, err := wi.db.GetWorkflowExecutionLatestByEntityID(wi.workflowID)
	if err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return nil, err // Only return if there's an actual error
	}

	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", wi.workflowID)
		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Outputs)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return nil, err
		}
		return outputs, nil
	}

	handler := wi.handler
	wFunc := handler.Handler

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
	}

	var workflowInputs [][]byte
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataInputs(&workflowInputs)); err != nil {
		log.Printf("Error getting workflow data inputs: %v", err)
		return nil, err
	}

	argsValues := []reflect.Value{reflect.ValueOf(ctxWorkflow)}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	log.Printf("Executing workflow: %s with args: %v", handler.HandlerName, inputs)

	defer func() {
		if r := recover(); r != nil {
			// Capture the stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if wi.debug.canStackTrace() {
				fmt.Println(stackTrace)
				log.Printf("Panic in workflow: %v", r)
			}

			joinedErr := errors.New("workflow panicked")
			joinedErr = errors.Join(joinedErr, fmt.Errorf("stack trace: %s", stackTrace))

			wi.mu.Lock()

			// allow the caller to know about the panic
			err = joinedErr
			// due to the `err` it will trigger the fail state

			wi.mu.Unlock()
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		return nil, wi.ctx.Err()
	default:
	}

	results := reflect.ValueOf(wFunc).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		return nil, err
	}

	errInterface := results[numOut-1].Interface()

	if errInterface != nil {
		if continueErr, ok := errInterface.(*ContinueAsNewError); ok {
			log.Printf("Workflow requested ContinueAsNew")
			// We treat it as normal completion, return nil to proceed
			return nil, continueErr
		} else {
			log.Printf("Workflow returned error: %v", errInterface)

			realError := errInterface.(error)

			// Update execution error
			if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionError(realError.Error())); err != nil {
				wi.mu.Unlock()
				return nil, fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			// Emptied the output data
			if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(nil)); err != nil {
				wi.mu.Unlock()
				return nil, fmt.Errorf("failed to set workflow execution data outputs: %w", err)
			}

			return nil, realError
		}
	} else {
		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				log.Printf("Workflow returned result [%d]: %v", i, result)
				outputs = append(outputs, result)
			}
		}

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(outputs)
		if err != nil {
			log.Printf("Error serializing output: %v", err)
			return nil, err
		}

		// save result
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(outputBytes)); err != nil {
			wi.mu.Unlock()
			return nil, fmt.Errorf("failed to set workflow execution data outputs: %w", err)
		}

		return outputs, nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, args ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.workflowID)

	if len(args) != 1 {
		return fmt.Errorf("WorkfloInstance onCompleted expected 1 argument, got %d", len(args))
	}

	outputs := args[0].([]interface{})

	var err error
	var isRoot bool

	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return fmt.Errorf("failed to get workflow data properties: %w", err)
	}

	// Handle ContinueAsNew
	// We re-create a new workflow entity to that related to that workflow
	// It will run on another worker on it's own, we are use preparing it
	if wi.continueAsNew != nil {

		wi.mu.Lock()
		handler := wi.handler
		wi.mu.Unlock()

		// Convert inputs to [][]byte
		inputBytes, err := convertInputsForSerialization(wi.continueAsNew.Args)
		if err != nil {
			log.Printf("Error converting inputs in ContinueAsNew: %v", err)
			if wi.future != nil {
				wi.future.setError(err)
			}
			return err
		}

		// Convert API RetryPolicy to internal retry policy
		internalRetryPolicy := ToInternalRetryPolicy(wi.continueAsNew.Options.RetryPolicy)

		copyID := wi.workflowID

		var workflowID int
		if workflowID, err = wi.db.AddWorkflowEntity(&WorkflowEntity{
			BaseEntity: BaseEntity{
				// we still depends on the run
				RunID: wi.runID,
				// continue as new mean on the same queue
				// you have to make another sub-workflow on a different queue
				QueueID:     wi.queueID,
				Type:        EntityWorkflow,
				HandlerName: handler.HandlerName,
				Status:      StatusPending,
				RetryPolicy: *internalRetryPolicy,
				RetryState:  RetryState{Attempts: 0},
				// We have to use the same stepID
				// - if we are root, there is no problem, it's normal to have root
				// - if we are not root, that mean we are a child, so we have to use the same stepID
				// I need to make a diagram for that... it's badly explained
				StepID: wi.stepID,
			},
			WorkflowData: &WorkflowData{
				Duration:      "",
				Paused:        false,
				Resumable:     false,
				Inputs:        inputBytes,
				ContinuedFrom: &copyID, // parent workflow
			},
		}); err != nil {
			log.Printf("Error adding workflow entity: %v", err)
			return err
		}

		var workflowExecutionID int
		if workflowExecutionID, err = wi.db.AddWorkflowExecution(&WorkflowExecution{
			BaseExecution: BaseExecution{
				Status:   ExecutionStatusPending,
				EntityID: workflowID,
			},
			WorkflowExecutionData: &WorkflowExecutionData{},
		}); err != nil {
			log.Printf("Error adding workflow execution: %v", err)
			return err
		}

		// if !isRoot {
		if _, err = wi.db.
			AddHierarchy(&Hierarchy{
				RunID:             wi.runID,
				ParentEntityID:    wi.parentEntityID,
				ParentExecutionID: wi.parentExecutionID,
				ChildEntityID:     workflowID,
				ChildExecutionID:  workflowExecutionID,
				ParentStepID:      wi.stepID,
				ChildStepID:       wi.stepID, // both use the same since it's a continue as new
				ParentType:        EntityWorkflow,
				ChildType:         EntityWorkflow,
			}); err != nil {
			log.Printf("Error adding hierarchy: %v", err)
			return err
		}
		// }

		// var queueName string = DefaultQueue
		// if err = wi.db.GetQueueProperties(wi.queueID, GetQueueName(&queueName)); err != nil {
		// 	log.Printf("Error getting queue properties: %v", err)
		// 	return err
		// }

		// // Create a new WorkflowInstance
		// newInstance := &WorkflowInstance{
		// 	ctx:      wi.ctx,
		// 	db:       wi.db,
		// 	tracker:  wi.tracker,
		// 	state:    wi.state,
		// 	registry: wi.registry,
		// 	debug:    wi.debug,

		// 	inputs: wi.continueAsNew.Args,
		// 	stepID: wi.stepID,

		// 	runID:       wi.runID,
		// 	workflowID:  workflowID,
		// 	executionID: workflowExecutionID,
		// 	queueID:     wi.queueID,

		// 	parentStepID:      wi.stepID,
		// 	parentEntityID:    wi.workflowID,
		// 	parentExecutionID: wi.executionID,

		// 	handler: handler,

		// 	options: WorkflowOptions{
		// 		RetryPolicy: &RetryPolicy{
		// 			MaxAttempts:        internalRetryPolicy.MaxAttempts,
		// 			InitialInterval:    time.Duration(internalRetryPolicy.InitialInterval),
		// 			BackoffCoefficient: internalRetryPolicy.BackoffCoefficient,
		// 			MaxInterval:        time.Duration(internalRetryPolicy.MaxInterval),
		// 		},
		// 		Queue: queueName,
		// 		// VersionOverrides: , // TODO: implement in WorkflowOption and save it
		// 	},
		// }

		return nil
	}

	// Normal completion
	if wi.future != nil {
		wi.future.setResult(outputs)
	}

	if isRoot {

		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusCompleted)); err != nil {
			log.Printf("Error setting run status: %v", err)
			return err
		}

		if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
			log.Printf("Error setting workflow execution status: %v", err)
			return err
		}

		fmt.Println("Entity", wi.workflowID, "completed")
		fmt.Println("Execution", wi.executionID, "completed")
	}

	if err = wi.db.SetWorkflowEntityProperties(wi.workflowID, SetWorkflowEntityStatus(StatusCompleted)); err != nil {
		wi.mu.Unlock()
		return fmt.Errorf("failed to set workflow entity status: %w", err)
	}

	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, args ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.workflowID)
	var err error

	if len(args) != 1 {
		return fmt.Errorf("WorkfloInstance onFailed expected 1 argument, got %d", len(args))
	}

	err = args[0].(error)

	joinedErr := errors.Join(ErrWorkflowFailed, err)

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return fmt.Errorf("failed to get workflow data properties: %w", err)
	}

	// Since the executions failed many times and was already changed
	if err = wi.db.
		SetWorkflowEntityProperties(
			wi.workflowID,
			SetWorkflowEntityStatus(StatusPaused),
		); err != nil {
		return fmt.Errorf("failed to set workflow entity status: %w", err)
	}

	if wi.future != nil {
		wi.future.setError(joinedErr)
	}

	// If this is the root workflow, update the Run status to Failed
	if isRoot {
		// Entity is now a failure
		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusFailed)); err != nil {
			log.Printf("Error setting run status: %v", err)
			return err
		}
	}

	return nil
}

func (wi *WorkflowInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onPaused called", wi.stepID, wi.workflowID)
	if wi.future != nil {
		wi.future.setError(ErrPaused)
	}
	var err error

	if err = wi.db.
		SetWorkflowEntityProperties(
			wi.workflowID,
			SetWorkflowEntityStatus(StatusPaused),
		); err != nil {
		return fmt.Errorf("failed to set workflow entity status: %w", err)
	}

	if err = wi.db.
		SetWorkflowDataProperties(
			wi.workflowID,
			SetWorkflowDataPaused(true),
		); err != nil {
		return fmt.Errorf("failed to set workflow data paused: %w", err)
	}

	return nil
}

///////////// Activity

func (ctx WorkflowContext) Activity(stepID string, activityFunc interface{}, options *ActivityOptions, args ...interface{}) *RuntimeFuture {
	var err error

	if err = ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Activity paused at stepID: %s", stepID)
		future := NewRuntimeFuture()
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewRuntimeFuture()

	// Register activity on-the-fly
	handler, err := ctx.registry.RegisterActivity(activityFunc)
	if err != nil {
		log.Printf("Error registering activity: %v", err)
		future.setError(err)
		return future
	}

	// Do my parent already had me?
	var hierarchy *Hierarchy
	if hierarchy, err = ctx.db.GetHierarchyByParentEntity(ctx.workflowID, stepID, EntityActivity); err != nil {
		if !errors.Is(err, ErrHierarchyNotFound) {
			log.Printf("Error getting entity: %v", err)
			future.setError(err)
			return future
		}
	}

	// We might have a result from a previous execution
	if hierarchy != nil {
		var activityExecution *ActivityExecution
		if activityExecution, err = ctx.db.GetActivityExecution(hierarchy.ChildExecutionID); err != nil {
			log.Printf("Error getting activity execution: %v", err)
			future.setError(err)
			return future
		}

		// If was completed, return the result
		if activityExecution.Status == ExecutionStatusCompleted {

			log.Printf("Activity %s (Entity ID: %d) already completed", stepID, hierarchy.ChildEntityID)

			var activityExecutionData *ActivityExecutionData
			if activityExecutionData, err = ctx.db.GetActivityExecutionData(activityExecution.ID); err != nil {
				log.Printf("Error getting activity execution data: %v", err)
				future.setError(err)
				return future
			}

			outputs, err := convertOutputsFromSerialization(handler, activityExecutionData.Outputs)
			if err != nil {
				log.Printf("Error converting outputs: %v", err)
				future.setError(err)
				return future
			}

			log.Printf("Activity %s (Entity ID: %d) outputs: %v", stepID, hierarchy.ChildEntityID, outputs)
			future.setResult(outputs)
			return future
		}
	}

	// Convert inputs to [][]byte
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future.setError(err)
		// TODO: should i trigger the "stopWithError"
		return future
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal

	// Handle options and defaults
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		// Fill default values if zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1 // no retries
		}
		if rp.InitialInterval == 0 {
			rp.InitialInterval = time.Second
		}
		if rp.BackoffCoefficient == 0 {
			rp.BackoffCoefficient = 2.0
		}
		if rp.MaxInterval == 0 {
			rp.MaxInterval = 5 * time.Minute
		}
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        rp.MaxAttempts,
			InitialInterval:    rp.InitialInterval.Nanoseconds(),
			BackoffCoefficient: rp.BackoffCoefficient,
			MaxInterval:        rp.MaxInterval.Nanoseconds(),
		}
	} else {
		// Default RetryPolicy
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        1,
			InitialInterval:    time.Second.Nanoseconds(),
			BackoffCoefficient: 2.0,
			MaxInterval:        (5 * time.Minute).Nanoseconds(),
		}
	}

	var activityEntityID int
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
	}); err != nil {
		log.Printf("Error adding activity entity: %v", err)
		future.setError(err)
		return future
	}

	future.setEntityID(activityEntityID)

	activityInstance := &ActivityInstance{
		ctx:     ctx.ctx,
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
		queueID:     ctx.queueID,
		runID:       ctx.runID,
	}

	if options != nil {
		activityInstance.options = *options
	}

	inputs, err := convertInputsFromSerialization(handler, inputBytes)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
	}

	ctx.tracker.addActivityInstance(activityInstance)
	activityInstance.Start(inputs) // synchronous because we want to know when it's done for real, the future will sent back the data anyway

	return future
}

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

	fmt.Println("Starting activity: ", ai.stepID)
	// Start the FSM
	ai.fsm.Fire(TriggerStart, inputs)
}

func (ai *ActivityInstance) executeActivity(_ context.Context, args ...interface{}) error {
	if len(args) != 1 {
		return fmt.Errorf("ActivityInstance executeActivity expected 1 argument, got %d", len(args))
	}
	return ai.executeWithRetry(args[0].([]interface{}))
}

// TODO: we need to copy the same algorithm as the workflowinstance here
func (ai *ActivityInstance) executeWithRetry(inputs []interface{}) error {
	var err error

	var retryState RetryState
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration
	fmt.Println("ActivityInstance executeWithRetry called")

	ai.mu.Lock()
	if ai.parentExecutionID == 0 {
		fmt.Println("ActivityInstance parentExecutionID is required (cannot execute an activity without a parent)")
		return fmt.Errorf("activity parentExecutionID is required (cannot execute an activity without a parent)")
	}

	maxAttempts, initialInterval, backoffCoefficient, maxInterval = getRetryPolicyOrDefault(ai.options.RetryPolicy)

	if err = ai.db.GetActivityEntityProperties(ai.entityID, GetActivityEntityRetryState(&retryState)); err != nil {
		return err
	}
	ai.mu.Unlock()

	fmt.Println("ActivityInstance executeWithRetry loop retrystate:", retryState)
	for {
		ai.mu.Lock()
		if ai.state.IsPaused() {
			// It means that will have to put the Activity from Paused -> Pending and let the orchestrator execute
			log.Printf("ActivityInstance %s is paused", ai.stepID)

			// TODO: pause/resume: please set me to Pending when you resume
			if err = ai.db.SetActivityEntityProperties(
				ai.entityID,
				SetActivityEntityStatus(StatusPaused),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set Activity entity status: %w", err)
			}

			ai.fsm.Fire(TriggerPause)
			ai.mu.Unlock()
			return nil
		}
		ai.mu.Unlock()

		// Check if maximum attempts have been reached
		ai.mu.Lock()
		if retryState.Attempts > maxAttempts {
			// Only execution here
			if err = ai.db.SetActivityExecutionProperties(
				ai.executionID,
				SetActivityExecutionStatus(ExecutionStatusFailed),
				SetActivityExecutionError(fmt.Errorf("activity %s failed after %d attempts", ai.stepID, retryState.Attempts-1).Error()),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set activity execution status: %w", err)
			}

			ai.fsm.Fire(TriggerFail, err)
			log.Printf("Activity %s (Entity ID: %d) failed after %d attempts", ai.stepID, ai.entityID, retryState.Attempts-1)
			ai.mu.Unlock()
			return nil
		}
		ai.mu.Unlock()

		// now we can create and set the execution id
		var activityExecutionID int
		if activityExecutionID, err = ai.db.AddActivityExecution(&ActivityExecution{
			BaseExecution: BaseExecution{
				EntityID: ai.entityID,
				Status:   ExecutionStatusRunning,
			},
			ActivityExecutionData: &ActivityExecutionData{},
		}); err != nil {
			ai.mu.Lock()
			log.Printf("Error adding execution: %v", err)
			ai.fsm.Fire(TriggerFail, err)
			ai.mu.Unlock()
			return err
		}

		ai.mu.Lock()
		ai.executionID = activityExecutionID // Store execution ID

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
			ai.mu.Unlock()
			return fmt.Errorf("failed to add hierarchy: %w", err)
		}

		// Update RetryState
		if retryState.Attempts <= maxAttempts {
			if err = ai.db.SetActivityEntityProperties(
				ai.entityID,
				SetActivityEntityRetryState(retryState),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set activity properties: %w", err)
			}
		}

		log.Printf("Executing activity %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.entityID, ai.executionID)
		ai.mu.Unlock()

		outputs, activityErr := ai.runActivity(inputs)
		fmt.Println("run activity", err)
		if errors.Is(err, ErrPaused) {
			// TODO: IS THAT EVEN A CASE?! I DON'T THINK SO
			// REVIEW AND DELETE
			ai.mu.Lock()
			if err = ai.db.SetActivityEntityProperties(
				ai.entityID,
				SetActivityEntityStatus(StatusPaused),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set Activity entity status: %w", err)
			}
			ai.fsm.Fire(TriggerPause)
			ai.mu.Unlock()
			return nil
		}

		if activityErr == nil {
			// Success
			ai.mu.Lock()
			if err = ai.db.SetActivityExecutionProperties(
				ai.executionID,
				SetActivityExecutionStatus(ExecutionStatusCompleted),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set activity execution status: %w", err)
			}

			ai.fsm.Fire(TriggerComplete, outputs)
			ai.mu.Unlock()
			return nil
		} else {
			// Failed
			// We should only intervene in the Execution part
			ai.mu.Lock()
			if err = ai.db.SetActivityExecutionProperties(
				ai.executionID,
				SetActivityExecutionStatus(ExecutionStatusFailed),
				SetActivityExecutionError(activityErr.Error()),
			); err != nil {
				ai.mu.Unlock()
				return fmt.Errorf("failed to set activity execution status: %w", err)
			}

			// retry management
			if retryState.Attempts < maxAttempts {
				// Calculate next interval
				nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(retryState.Attempts-1)))
				if nextInterval > maxInterval {
					nextInterval = maxInterval
				}
				log.Printf("Retrying activity %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v",
					ai.stepID, ai.entityID, ai.executionID, retryState.Attempts+1, maxAttempts, nextInterval)
				ai.mu.Unlock()
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached
				ai.fsm.Fire(TriggerFail, activityErr)
				ai.mu.Unlock()
				return nil
			}
		}

		// Increment attempt for next iteration
		retryState.Attempts++
	}
}

// Sub-function within executeWithRetry
func (ai *ActivityInstance) runActivity(inputs []interface{}) (outputs []interface{}, err error) {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.entityID, ai.entityID)

	activityExecution, err := ai.db.GetActivityExecutionLatestByEntityID(ai.entityID)
	if err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return nil, err
	}

	if activityExecution != nil && activityExecution.Status == ExecutionStatusCompleted &&
		activityExecution.ActivityExecutionData != nil &&
		activityExecution.ActivityExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", ai.entityID)
		outputs, err = convertOutputsFromSerialization(ai.handler, activityExecution.ActivityExecutionData.Outputs)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return nil, err
		}
		return outputs, nil
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
				log.Printf("Panic in activity: %v", r)
			}

			ai.mu.Lock()
			// allow the caller to know about the panic
			err = errors.Join(ErrActivityPanicked, fmt.Errorf("stack trace: %s", stackTrace))
			// due to the `err` it will trigger the fail state
			ai.mu.Unlock()
		}
	}()

	var activityInputs [][]byte
	if err = ai.db.GetActivityDataProperties(ai.entityID, GetActivityDataInputs(&activityInputs)); err != nil {
		log.Printf("Error getting activity data inputs: %v", err)
		return nil, err
	}

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	select {
	case <-ai.ctx.Done():
		log.Printf("Context cancelled in activity")
		return nil, ai.ctx.Err()
	default:
	}

	results := reflect.ValueOf(activityFunc).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("activity %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		return nil, err
	}

	errInterface := results[numOut-1].Interface()
	if errInterface != nil {
		log.Printf("Activity returned error: %v", errInterface)
		err = errInterface.(error)

		// Serialize error
		errorMessage := err.Error()

		// Update execution error
		if err = ai.db.SetActivityExecutionProperties(
			ai.entityID,
			SetActivityExecutionError(errorMessage)); err != nil {
			return nil, fmt.Errorf("failed to set activity execution error: %w", err)
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(nil)); err != nil {
			return nil, fmt.Errorf("failed to set activity execution data outputs: %w", err)
		}

		return nil, err
	} else {
		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				log.Printf("Activity returned result [%d]: %v", i, result)
				outputs = append(outputs, result)
			}
		}

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(outputs)
		if err != nil {
			log.Printf("Error serializing output: %v", err)
			return nil, err
		}

		// Update the execution with the execution data
		if err = ai.db.SetActivityExecutionDataProperties(
			ai.entityID,
			SetActivityExecutionDataOutputs(outputBytes)); err != nil {
			return nil, fmt.Errorf("failed to set activity execution data outputs: %w", err)
		}

		return outputs, nil
	}
}

func (ai *ActivityInstance) onCompleted(_ context.Context, args ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onCompleted called", ai.stepID, ai.entityID)
	if len(args) != 1 {
		return fmt.Errorf("activity instance onCompleted expected 1 argument, got %d", len(args))
	}
	if ai.future != nil {
		ouputs := args[0].([]interface{})
		ai.future.setResult(ouputs)
	}
	return nil
}

func (ai *ActivityInstance) onFailed(_ context.Context, args ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onFailed called", ai.stepID, ai.entityID)
	if len(args) != 1 {
		return fmt.Errorf("activity instance onFailed expected 1 argument, got %d", len(args))
	}
	if ai.future != nil {
		err := args[0].(error)
		ai.future.setError(err)
	}
	return nil
}

func (ai *ActivityInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onPaused called", ai.stepID, ai.entityID)
	if ai.future != nil {
		ai.future.setError(ErrPaused)
	}
	return nil
}
