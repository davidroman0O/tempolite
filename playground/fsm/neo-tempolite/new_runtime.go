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
	ctx                 context.Context
	stepID              string
	workflowID          int
	workflowExecutionID int
	options             WorkflowOptions
}

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	ctx     context.Context
	mu      deadlock.Mutex
	db      Database
	tracker InstanceTracker

	handler HandlerInfo

	inputs  []interface{}
	results []interface{}
	err     error

	fsm *stateless.StateMachine

	options WorkflowOptions
	future  Future

	stepID      string
	runID       int
	workflowID  int
	entityID    int
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
	ctx context.Context
	mu  deadlock.Mutex

	handler HandlerInfo

	input   []interface{}
	results []interface{}
	err     error

	fsm *stateless.StateMachine

	options ActivityOptions
	future  Future

	stepID      string
	workflowID  int
	entityID    int
	executionID int

	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
	ctx context.Context
	mu  deadlock.Mutex

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
	ctx context.Context
	mu  deadlock.Mutex

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
}

func NewOrchestrator(ctx context.Context, db Database, registry *Registry) *Orchestrator {
	log.Printf("NewOrchestrator called")
	ctx, cancel := context.WithCancel(ctx)
	o := &Orchestrator{
		db:       db,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
	}
	return o
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
					Attempt:   1,
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

	var workflowExecutionID int
	if workflowExecutionID, err = o.db.AddWorkflowExecution(&WorkflowExecution{
		BaseExecution: BaseExecution{
			Status:   ExecutionStatusPending,
			EntityID: entityID,
			Attempt:  1, // TODO: i don't think i need to anymore
		},
		WorkflowExecutionData: &WorkflowExecutionData{},
	}); err != nil {
		log.Printf("Error adding workflow execution: %v", err)
		return nil, err
	}

	// Create workflow instance
	instance := &WorkflowInstance{
		ctx:     o.ctx, // share context with orchestrator
		db:      o.db,  // share db
		tracker: o,

		inputs: inputs,
		stepID: entity.StepID,

		runID:       o.runID,
		workflowID:  entity.ID,
		entityID:    entity.ID,
		executionID: workflowExecutionID,
		dataID:      entity.WorkflowData.ID,
		queueID:     entity.QueueID,

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
		go instance.Start() // root workflow OR ContinueAsNew starts in a goroutine
	} else {
		instance.Start() // sub-workflow directly starts
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

/// WorkflowInstance

// Starting the workflow instance
// Ideally we're already within a goroutine
func (wi *WorkflowInstance) Start() {

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

	log.Printf("Starting workflow: %s (Entity ID: %d)", wi.stepID, wi.entityID)

	// Start the FSM without holding wi.mu
	if err := fsm.Fire(TriggerStart); err != nil {
		log.Printf("Error starting workflow: %v", err)
	}
}

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) executeWorkflow called", wi.stepID, wi.entityID)
	return wi.executeWithRetry()
}

func (wi *WorkflowInstance) executeWithRetry() error {
	var err error
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration
	var paused bool
	var resumable bool
	var retryStateAtempts int

	wi.mu.Lock()
	if wi.options.RetryPolicy != nil {
		rp := wi.options.RetryPolicy
		maxAttempts = rp.MaxAttempts
		initialInterval = time.Duration(rp.InitialInterval)
		backoffCoefficient = rp.BackoffCoefficient
		maxInterval = time.Duration(rp.MaxInterval)
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = time.Second
		backoffCoefficient = 2.0
		maxInterval = 5 * time.Minute
	}

	fmt.Println("executionID", wi.executionID)
	if err = wi.db.GetWorkflowExecutionProperties(wi.executionID, GetWorkflowExecutionAttempt(&retryStateAtempts)); err != nil {
		wi.mu.Unlock()
		return fmt.Errorf("failed to get workflow execution properties: %w", err)
	}

	wi.mu.Unlock()

	for {
		wi.mu.Lock()

		// Check if paused
		if err = wi.db.GetWorkflowDataProperties(
			wi.entityID,
			GetWorkflowDataPaused(&paused),
			GetWorkflowDataResumable(&resumable),
		); err != nil {
			wi.mu.Unlock()
			return fmt.Errorf("failed to get workflow execution data: %w", err)
		}

		if paused {
			log.Printf("WorkflowInstance %s is paused", wi.stepID)

			if err = wi.db.
				SetWorkflowEntityProperties(
					wi.entityID,
					SetWorkflowEntityStatus(StatusPaused),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			if err = wi.db.
				SetWorkflowDataProperties(
					wi.entityID,
					SetWorkflowDataPaused(true),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow data paused: %w", err)
			}

			wi.fsm.Fire(TriggerPause)
			wi.mu.Unlock()
			return nil
		}
		wi.mu.Unlock()

		// Check if maximum attempts have been reached and not resumable
		wi.mu.Lock()
		if attempt > maxAttempts && !resumable {

			if err = wi.db.
				SetWorkflowEntityProperties(
					wi.entityID,
					SetWorkflowEntityStatus(StatusFailed),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			wi.fsm.Fire(TriggerFail)
			log.Printf("Workflow %s (Entity ID: %d) failed after %d attempts", wi.stepID, wi.entityID, attempt-1)
			wi.mu.Unlock()
			return nil
		}
		wi.mu.Unlock()

		// now we can create and set the execution id
		if wi.executionID, err = wi.db.AddWorkflowExecution(&WorkflowExecution{
			BaseExecution: BaseExecution{
				EntityID: wi.entityID,
				Status:   ExecutionStatusRunning,
				Attempt:  attempt,
			},
			WorkflowExecutionData: &WorkflowExecutionData{},
		}); err != nil {
			wi.mu.Lock()
			log.Printf("Error adding execution: %v", err)
			if err = wi.db.
				SetWorkflowEntityProperties(
					wi.entityID,
					SetWorkflowEntityStatus(StatusFailed),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}
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

				ChildEntityID:    wi.entityID,
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
		if attempt <= maxAttempts {
			if err = wi.db.
				SetWorkflowExecutionProperties(
					wi.executionID,
					SetWorkflowExecutionAttempt(attempt),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution properties: %w", err)
			}
		}
		wi.mu.Unlock()

		wi.mu.Lock()
		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.entityID, wi.executionID)
		wi.mu.Unlock()

		err = wi.runWorkflow()
		if errors.Is(err, ErrPaused) {
			wi.mu.Lock()

			if err = wi.db.
				SetWorkflowEntityProperties(
					wi.entityID,
					SetWorkflowEntityStatus(StatusPaused),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			if err = wi.db.
				SetWorkflowDataProperties(
					wi.entityID,
					SetWorkflowDataPaused(true),
				); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow data paused: %w", err)
			}

			wi.fsm.Fire(TriggerPause)
			log.Printf("WorkflowInstance %s is paused post-runWorkflow", wi.stepID)
			wi.mu.Unlock()
			return nil
		}

		if err == nil {
			// Success
			wi.mu.Lock()

			if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionStatus(ExecutionStatusCompleted)); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			if err = wi.db.SetWorkflowEntityProperties(wi.entityID, SetWorkflowEntityStatus(StatusCompleted)); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow entity status: %w", err)
			}

			wi.fsm.Fire(TriggerComplete)
			log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) completed successfully", wi.stepID, wi.entityID, wi.executionID)
			wi.mu.Unlock()
			return nil
		} else {
			wi.mu.Lock()

			if err = wi.db.SetWorkflowExecutionProperties(
				wi.executionID,
				SetWorkflowExecutionStatus(ExecutionStatusFailed),
				SetWorkflowExecutionError(err.Error()),
			); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			wi.err = err

			wi.mu.Unlock()

			if attempt < maxAttempts || resumable {
				// Calculate next interval
				nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(attempt-1)))
				if nextInterval > maxInterval {
					nextInterval = maxInterval
				}
				wi.mu.Lock()
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.entityID, wi.executionID, attempt+1, maxAttempts, nextInterval)
				wi.mu.Unlock()
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached and not resumable
				wi.mu.Lock()

				if err = wi.db.SetWorkflowEntityProperties(wi.entityID, SetWorkflowEntityStatus(StatusFailed)); err != nil {
					wi.mu.Unlock()
					return fmt.Errorf("failed to set workflow entity status: %w", err)
				}

				wi.fsm.Fire(TriggerFail)
				log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) failed after %d attempts", wi.stepID, wi.entityID, wi.executionID, attempt)
				wi.mu.Unlock()
				return nil
			}
		}

		// Increment attempt for the next loop iteration
		attempt++
	}
}

func (wi *WorkflowInstance) runWorkflow() error {
	log.Printf("WorkflowInstance %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.entityID, wi.executionID)
	var err error

	// Get the latest execution, which may be nil if none exist
	latestExecution, err := wi.db.GetWorkflowExecutionLatestByEntityID(wi.entityID)
	if err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err // Only return if there's an actual error
	}

	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", wi.entityID)
		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Outputs)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return err
		}
		wi.mu.Lock()
		wi.results = outputs
		wi.mu.Unlock()
		return nil
	}

	handler := wi.handler
	wFunc := handler.Handler

	ctxWorkflow := WorkflowContext{
		db:                  wi.db,
		ctx:                 wi.ctx,
		options:             wi.options,
		stepID:              wi.stepID,
		workflowID:          wi.entityID,
		workflowExecutionID: wi.executionID,
	}

	var workflowInputs [][]byte
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataInputs(&workflowInputs)); err != nil {
		log.Printf("Error getting workflow data inputs: %v", err)
		return err
	}

	// Convert inputs from serialization
	inputs, err := convertInputsFromSerialization(handler, workflowInputs)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return err
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
			fmt.Println(stackTrace)

			log.Printf("Panic in workflow: %v", r)
			wi.mu.Lock()
			wi.err = fmt.Errorf("panic: %v", r)
			wi.mu.Unlock()
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		wi.mu.Lock()
		wi.err = wi.ctx.Err()
		wi.mu.Unlock()
		return wi.err
	default:
	}

	results := reflect.ValueOf(wFunc).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		wi.mu.Lock()
		wi.err = err
		wi.mu.Unlock()
		return err
	}

	errInterface := results[numOut-1].Interface()

	if errInterface != nil {
		if continueErr, ok := errInterface.(*ContinueAsNewError); ok {
			log.Printf("Workflow requested ContinueAsNew")
			wi.mu.Lock()
			wi.continueAsNew = continueErr
			wi.mu.Unlock()
			// We treat it as normal completion, return nil to proceed
			return nil
		} else {
			log.Printf("Workflow returned error: %v", errInterface)
			wi.mu.Lock()
			wi.err = errInterface.(error)
			wi.mu.Unlock()

			// Serialize error
			errorMessage := wi.err.Error()

			// Update execution error
			if err = wi.db.SetWorkflowExecutionProperties(wi.executionID, SetWorkflowExecutionError(errorMessage)); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution status: %w", err)
			}

			// Emptied the output data
			if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(nil)); err != nil {
				wi.mu.Unlock()
				return fmt.Errorf("failed to set workflow execution data outputs: %w", err)
			}

			return wi.err
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

		wi.mu.Lock()
		wi.results = outputs
		wi.mu.Unlock()

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(wi.results)
		if err != nil {
			log.Printf("Error serializing output: %v", err)
			return err
		}

		// save result
		if err = wi.db.SetWorkflowExecutionDataProperties(wi.dataID, SetWorkflowExecutionDataOutputs(outputBytes)); err != nil {
			wi.mu.Unlock()
			return fmt.Errorf("failed to set workflow execution data outputs: %w", err)
		}

		return nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.entityID)
	var err error
	var isRoot bool

	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return fmt.Errorf("failed to get workflow data properties: %w", err)
	}

	if wi.continueAsNew != nil {
		// Handle ContinueAsNew

		wi.mu.Lock()
		handler := wi.handler
		wi.mu.Unlock()

		// Convert inputs to [][]byte
		inputBytes, err := convertInputsForSerialization(wi.continueAsNew.Args)
		if err != nil {
			log.Printf("Error converting inputs in ContinueAsNew: %v", err)
			wi.mu.Lock()
			wi.err = err
			wi.mu.Unlock()
			if wi.future != nil {
				wi.future.setError(wi.err)
			}
			return nil
		}

		// Convert API RetryPolicy to internal retry policy
		internalRetryPolicy := ToInternalRetryPolicy(wi.continueAsNew.Options.RetryPolicy)

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
				Duration:  "",
				Paused:    false,
				Resumable: false,
				Inputs:    inputBytes,
				Attempt:   1,
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
				Attempt:  1, // TODO: i don't think i need to anymore
			},
			WorkflowExecutionData: &WorkflowExecutionData{},
		}); err != nil {
			log.Printf("Error adding workflow execution: %v", err)
			return err
		}

		if !isRoot {
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
		}

		var queueName string = DefaultQueue
		if err = wi.db.GetQueueProperties(wi.queueID, GetQueueName(&queueName)); err != nil {
			log.Printf("Error getting queue properties: %v", err)
			return err
		}

		// Create a new WorkflowInstance
		newInstance := &WorkflowInstance{
			ctx:     wi.ctx,
			db:      wi.db,
			tracker: wi.tracker,

			inputs: wi.continueAsNew.Args,
			stepID: wi.stepID,

			runID:       wi.runID,
			workflowID:  workflowID,
			entityID:    workflowID,
			executionID: workflowExecutionID,
			queueID:     wi.queueID,

			parentStepID:      wi.stepID,
			parentEntityID:    wi.entityID,
			parentExecutionID: wi.executionID,

			handler: handler,

			options: WorkflowOptions{
				RetryPolicy: &RetryPolicy{
					MaxAttempts:        internalRetryPolicy.MaxAttempts,
					InitialInterval:    time.Duration(internalRetryPolicy.InitialInterval),
					BackoffCoefficient: internalRetryPolicy.BackoffCoefficient,
					MaxInterval:        time.Duration(internalRetryPolicy.MaxInterval),
				},
				Queue: queueName,
				// VersionOverrides: , // TODO: implement in WorkflowOption and save it
			},
		}

		// Since we hold the write lock, we can directly access o.rootWf
		if isRoot {
			wi.tracker.newRoot(newInstance)
		}

		// Add the new instance to the orchestrator's instances
		wi.tracker.addWorkflowInstance(newInstance)

		// Now we can start the new instance
		newInstance.Start()

		if isRoot {
			// Complete the future immediately
			if wi.future != nil {
				wi.mu.Lock()
				wi.future.setResult(wi.results)
				wi.mu.Unlock()
			}
		} else {
			// Sub-workflow
			// Pass the future to the newInstance
			newInstance.future = wi.future
		}

	} else {
		// Normal completion
		if wi.future != nil {
			wi.mu.Lock()
			resultsCopy := wi.results
			wi.mu.Unlock()
			wi.future.setResult(resultsCopy)
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

			fmt.Println("Entity", wi.entityID, "completed")
			fmt.Println("Execution", wi.executionID, "completed")
		}
	}

	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.entityID)
	var err error

	var isRoot bool
	if err = wi.db.GetWorkflowDataProperties(wi.dataID, GetWorkflowDataIsRoot(&isRoot)); err != nil {
		return fmt.Errorf("failed to get workflow data properties: %w", err)
	}

	if wi.future != nil {
		wi.mu.Lock()
		errCopy := wi.err
		wi.mu.Unlock()
		wi.future.setError(errCopy)
	}
	// If this is the root workflow, update the Run status to Failed
	if isRoot {
		if err = wi.db.SetRunProperties(wi.runID, SetRunStatus(RunStatusFailed)); err != nil {
			log.Printf("Error setting run status: %v", err)
			return err
		}
	}
	return nil
}

func (wi *WorkflowInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onPaused called", wi.stepID, wi.entityID)
	if wi.future != nil {
		wi.future.setError(ErrPaused)
	}
	return nil
}
