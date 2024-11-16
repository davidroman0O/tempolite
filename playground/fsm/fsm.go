package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qmuntal/stateless"
	"github.com/stephenfire/go-rtl"
)

type state string

// FSM states
const (
	StateIdle          state = "Idle"
	StateExecuting     state = "Executing"
	StateCompleted     state = "Completed"
	StateFailed        state = "Failed"
	StateRetried       state = "Retried"
	StatePaused        state = "Paused"
	StateTransactions  state = "Transactions"
	StateCompensations state = "Compensations"
)

type trigger string

// FSM triggers
const (
	TriggerStart      trigger = "Start"
	TriggerComplete   trigger = "Complete"
	TriggerFail       trigger = "Fail"
	TriggerPause      trigger = "Pause"
	TriggerResume     trigger = "Resume"
	TriggerCompensate trigger = "Compensate"
)

// Serialization functions
func ConvertInputsForSerialization(executionInputs []interface{}) ([][]byte, error) {
	inputs := [][]byte{}

	for _, input := range executionInputs {
		buf := new(bytes.Buffer)

		// Get the real value
		if reflect.TypeOf(input).Kind() == reflect.Ptr {
			input = reflect.ValueOf(input).Elem().Interface()
		}

		if err := rtl.Encode(input, buf); err != nil {
			return nil, err
		}
		inputs = append(inputs, buf.Bytes())
	}

	return inputs, nil
}

func ConvertOutputsForSerialization(executionOutputs []interface{}) ([][]byte, error) {
	outputs := [][]byte{}

	for _, output := range executionOutputs {
		buf := new(bytes.Buffer)

		// Get the real value
		if reflect.TypeOf(output).Kind() == reflect.Ptr {
			output = reflect.ValueOf(output).Elem().Interface()
		}

		if err := rtl.Encode(output, buf); err != nil {
			return nil, err
		}
		outputs = append(outputs, buf.Bytes())
	}

	return outputs, nil
}

func ConvertInputsFromSerialization(handlerInfo HandlerInfo, executionInputs [][]byte) ([]interface{}, error) {
	inputs := []interface{}{}

	for idx, inputType := range handlerInfo.ParamTypes {
		buf := bytes.NewBuffer(executionInputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(inputType).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		inputs = append(inputs, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return inputs, nil
}

func ConvertOutputsFromSerialization(handlerInfo HandlerInfo, executionOutputs [][]byte) ([]interface{}, error) {
	output := []interface{}{}

	for idx, outputType := range handlerInfo.ReturnTypes {
		buf := bytes.NewBuffer(executionOutputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(outputType).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}

func ConvertSingleOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
	buf := bytes.NewBuffer(executionOutput)

	decodedObj := reflect.New(outputType).Interface()

	if err := rtl.Decode(buf, decodedObj); err != nil {
		return nil, err
	}

	return reflect.ValueOf(decodedObj).Elem().Interface(), nil
}

// Future represents an asynchronous result.
type Future struct {
	results    []interface{}
	err        error
	done       chan struct{}
	workflowID int
}

func NewFuture(workflowID int) *Future {
	return &Future{
		done:       make(chan struct{}),
		workflowID: workflowID,
	}
}

func (f *Future) Get(outputs ...interface{}) error {
	log.Printf("Future.Get called")
	<-f.done
	if f.err != nil {
		log.Printf("Future.Get returning error: %v", f.err)
		return f.err
	}
	if len(outputs) > len(f.results) {
		return fmt.Errorf("number of outputs requested exceeds number of results")
	}
	for i := 0; i < len(outputs); i++ {
		if outputs[i] != nil && f.results[i] != nil {
			reflect.ValueOf(outputs[i]).Elem().Set(reflect.ValueOf(f.results[i]))
			log.Printf("Future.Get setting output[%d]: %v", i, f.results[i])
		}
	}
	return nil
}

func (f *Future) setResult(results []interface{}) {
	log.Printf("Future.setResult called with results: %v", results)
	f.results = results
	close(f.done)
}

func (f *Future) setError(err error) {
	log.Printf("Future.setError called with error: %v", err)
	f.err = err
	close(f.done)
}

// RetryPolicy defines retry behavior configuration for API usage.
type RetryPolicy struct {
	MaxAttempts        int
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaxInterval        time.Duration
}

// RetryState tracks the state of retries.
type RetryState struct {
	Attempts int `json:"attempts"`
}

// Internal retry policy matching the database schema, using int64 for intervals.
type retryPolicyInternal struct {
	MaxAttempts        int     `json:"max_attempts"`
	InitialInterval    int64   `json:"initial_interval"` // Stored in nanoseconds
	BackoffCoefficient float64 `json:"backoff_coefficient"`
	MaxInterval        int64   `json:"max_interval"` // Stored in nanoseconds
}

// WorkflowOptions provides options for workflows, including retry policies and version overrides.
type WorkflowOptions struct {
	RetryPolicy      *RetryPolicy
	VersionOverrides map[string]int // Map of changeID to forced version
}

// ActivityOptions provides options for activities, including retry policies.
type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

// WorkflowContext provides context for workflow execution.
type WorkflowContext struct {
	orchestrator *Orchestrator
	ctx          context.Context
	workflowID   int
	stepID       string
	options      *WorkflowOptions
}

var ErrPaused = errors.New("execution paused")

func (ctx *WorkflowContext) checkPause() error {
	if ctx.orchestrator.IsPaused() {
		log.Printf("WorkflowContext detected orchestrator is paused")
		return ErrPaused
	}
	return nil
}

func (ctx *WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options WorkflowOptions, args ...interface{}) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Workflow paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)
	future := NewFuture(0)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for workflow: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			return future
		}
		latestExecution := ctx.orchestrator.db.GetLatestExecution(entity.ID)
		if latestExecution != nil && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Output != nil {
			// Deserialize output
			outputs, err := ConvertOutputsFromSerialization(*handlerInfo, latestExecution.WorkflowExecutionData.Output)
			if err != nil {
				log.Printf("Error deserializing outputs: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}
	}

	// Register workflow on-the-fly
	handler, err := ctx.orchestrator.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		log.Printf("Error registering workflow: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert inputs to [][]byte
	inputBytes, err := ConvertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal
	if options.RetryPolicy != nil {
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        options.RetryPolicy.MaxAttempts,
			InitialInterval:    options.RetryPolicy.InitialInterval.Nanoseconds(),
			BackoffCoefficient: options.RetryPolicy.BackoffCoefficient,
			MaxInterval:        options.RetryPolicy.MaxInterval.Nanoseconds(),
		}
	}

	// Create WorkflowData
	workflowData := &WorkflowData{
		Duration:    "",
		Paused:      false,
		Resumable:   false,
		RetryState:  &RetryState{Attempts: 0},
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Type:         string(EntityTypeWorkflow),
		Status:       string(StatusPending),
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)
	future.workflowID = entity.ID

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		RunID:             ctx.orchestrator.runID,
		ParentEntityID:    ctx.workflowID,
		ChildEntityID:     entity.ID,
		ParentExecutionID: 0,
		ChildExecutionID:  0,
		ParentStepID:      ctx.stepID,
		ChildStepID:       stepID,
		ParentType:        string(EntityTypeWorkflow),
		ChildType:         string(EntityTypeWorkflow),
	}
	ctx.orchestrator.db.AddHierarchy(hierarchy)

	subWorkflowInstance := &WorkflowInstance{
		stepID:       stepID,
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
		workflowID:   entity.ID,
		options:      options,
		entity:       entity,
		entityID:     entity.ID,
	}
	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

func (ctx *WorkflowContext) Activity(stepID string, activityFunc interface{}, options ActivityOptions, args ...interface{}) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Activity paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewFuture(0)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for activity: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			return future
		}
		latestExecution := ctx.orchestrator.db.GetLatestExecution(entity.ID)
		if latestExecution != nil && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Output != nil {
			// Deserialize output
			outputs, err := ConvertOutputsFromSerialization(*handlerInfo, latestExecution.ActivityExecutionData.Output)
			if err != nil {
				log.Printf("Error deserializing outputs: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}
	}

	// Register activity on-the-fly
	handler, err := ctx.orchestrator.registry.RegisterActivity(activityFunc)
	if err != nil {
		log.Printf("Error registering activity: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert inputs to [][]byte
	inputBytes, err := ConvertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal
	if options.RetryPolicy != nil {
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        options.RetryPolicy.MaxAttempts,
			InitialInterval:    options.RetryPolicy.InitialInterval.Nanoseconds(),
			BackoffCoefficient: options.RetryPolicy.BackoffCoefficient,
			MaxInterval:        options.RetryPolicy.MaxInterval.Nanoseconds(),
		}
	}

	// Create ActivityData
	activityData := &ActivityData{
		Timeout:     0,
		MaxAttempts: 1,
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Type:         string(EntityTypeActivity),
		Status:       string(StatusPending),
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ActivityData: activityData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)
	future.workflowID = entity.ID

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		RunID:             ctx.orchestrator.runID,
		ParentEntityID:    ctx.workflowID,
		ChildEntityID:     entity.ID,
		ParentExecutionID: 0,
		ChildExecutionID:  0,
		ParentStepID:      ctx.stepID,
		ChildStepID:       stepID,
		ParentType:        string(EntityTypeWorkflow),
		ChildType:         string(EntityTypeActivity),
	}
	ctx.orchestrator.db.AddHierarchy(hierarchy)

	activityInstance := &ActivityInstance{
		stepID:       stepID,
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
		workflowID:   ctx.workflowID,
		options:      options,
		entity:       entity,
		entityID:     entity.ID,
	}
	ctx.orchestrator.addActivityInstance(activityInstance)
	activityInstance.Start()

	return future
}

func (ctx *WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}, options WorkflowOptions) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.SideEffect paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.SideEffect called with stepID: %s", stepID)
	future := NewFuture(0)

	// Get the return type of the sideEffectFunc
	sideEffectFuncType := reflect.TypeOf(sideEffectFunc)
	if sideEffectFuncType.Kind() != reflect.Func || sideEffectFuncType.NumOut() == 0 {
		err := fmt.Errorf("side effect function must be a function that returns at least one value")
		log.Printf("Error: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}
	returnTypes := make([]reflect.Type, sideEffectFuncType.NumOut())
	for i := 0; i < sideEffectFuncType.NumOut(); i++ {
		returnTypes[i] = sideEffectFuncType.Out(i)
	}

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		latestExecution := ctx.orchestrator.db.GetLatestExecution(entity.ID)
		if latestExecution != nil && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Output != nil {
			// Deserialize result using the returnTypes
			outputs, err := ConvertOutputsFromSerialization(HandlerInfo{ReturnTypes: returnTypes}, latestExecution.SideEffectExecutionData.Output)
			if err != nil {
				log.Printf("Error deserializing side effect result: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}
	}

	// Register side effect on-the-fly
	handlerName := getFunctionName(sideEffectFunc)

	// Side Effects are anonymous functions, we don't need to register them at all
	handler := HandlerInfo{
		HandlerName: handlerName,
		Handler:     sideEffectFunc,
		ReturnTypes: returnTypes,
	}

	// Use the retry policy from options or default
	var retryPolicy *RetryPolicy
	if options.RetryPolicy != nil {
		retryPolicy = options.RetryPolicy
	} else {
		// Default retry policy with MaxAttempts=1
		retryPolicy = &RetryPolicy{
			MaxAttempts:     1,
			InitialInterval: 0,
		}
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal
	if retryPolicy != nil {
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        retryPolicy.MaxAttempts,
			InitialInterval:    retryPolicy.InitialInterval.Nanoseconds(),
			BackoffCoefficient: retryPolicy.BackoffCoefficient,
			MaxInterval:        retryPolicy.MaxInterval.Nanoseconds(),
		}
	}

	// Create SideEffectData
	sideEffectData := &SideEffectData{
		Input:       nil, // No input for side effects in this example
		RetryPolicy: internalRetryPolicy,
	}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:         stepID,
		HandlerName:    handlerName,
		Type:           string(EntityTypeSideEffect),
		Status:         string(StatusPending),
		RunID:          ctx.orchestrator.runID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		SideEffectData: sideEffectData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)
	future.workflowID = entity.ID

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		RunID:             ctx.orchestrator.runID,
		ParentEntityID:    ctx.workflowID,
		ChildEntityID:     entity.ID,
		ParentExecutionID: 0,
		ChildExecutionID:  0,
		ParentStepID:      ctx.stepID,
		ChildStepID:       stepID,
		ParentType:        string(EntityTypeWorkflow),
		ChildType:         string(EntityTypeSideEffect),
	}
	ctx.orchestrator.db.AddHierarchy(hierarchy)

	sideEffectInstance := &SideEffectInstance{
		stepID:         stepID,
		sideEffectFunc: sideEffectFunc,
		future:         future,
		ctx:            ctx.ctx,
		orchestrator:   ctx.orchestrator,
		workflowID:     ctx.workflowID,
		entity:         entity,
		entityID:       entity.ID,
		options:        options,
		returnTypes:    returnTypes,
		handlerName:    handlerName,
		handler:        handler,
	}
	ctx.orchestrator.addSideEffectInstance(sideEffectInstance)
	sideEffectInstance.Start()

	return future
}

// SagaContext provides context for saga execution.
type SagaContext struct {
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	stepID       string
}

func (ctx *WorkflowContext) Saga(stepID string, saga *SagaDefinition) *SagaInfo {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Saga paused at stepID: %s", stepID)
		sagaInfo := &SagaInfo{
			err:  err,
			done: make(chan struct{}),
		}
		close(sagaInfo.done)
		return sagaInfo
	}

	// Execute the saga
	return ctx.orchestrator.executeSaga(ctx, stepID, saga)
}

// ActivityContext provides context for activity execution.
type ActivityContext struct {
	ctx context.Context
}

type HandlerIdentity string

func (h HandlerIdentity) String() string {
	return string(h)
}

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamsKinds     []reflect.Kind
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type
	ReturnKinds     []reflect.Kind
	NumIn           int
	NumOut          int
}

type EntityType string

const (
	EntityTypeWorkflow   EntityType = "Workflow"
	EntityTypeActivity   EntityType = "Activity"
	EntityTypeSideEffect EntityType = "SideEffect"
	EntityTypeSaga       EntityType = "Saga"
)

type EntityStatus string

const (
	StatusPending   EntityStatus = "Pending"
	StatusRunning   EntityStatus = "Running"
	StatusCompleted EntityStatus = "Completed"
	StatusFailed    EntityStatus = "Failed"
	StatusRetried   EntityStatus = "Retried"
	StatusPaused    EntityStatus = "Paused"
)

// ExecutionData structures matching the ent schemas
type WorkflowExecutionData struct {
	Error  string   `json:"error,omitempty"`
	Output [][]byte `json:"output,omitempty"`
}

type ActivityExecutionData struct {
	Heartbeats       [][]byte   `json:"heartbeats,omitempty"`
	LastHeartbeat    *time.Time `json:"last_heartbeat,omitempty"`
	Progress         []byte     `json:"progress,omitempty"`
	ExecutionDetails []byte     `json:"execution_details,omitempty"`
	ErrMessage       string     `json:"err_message,omitempty"`
	Output           [][]byte   `json:"output,omitempty"`
}

type SideEffectExecutionData struct {
	Output           [][]byte   `json:"output,omitempty"`
	EffectTime       *time.Time `json:"effect_time,omitempty"`
	EffectMetadata   []byte     `json:"effect_metadata,omitempty"`
	ExecutionContext []byte     `json:"execution_context,omitempty"`
}

type SagaExecutionData struct {
	Error  string   `json:"error,omitempty"`
	Output [][]byte `json:"output,omitempty"`
}

// Entity represents the base entity for workflows, activities, and side effects.
type Entity struct {
	ID             int
	HandlerName    string
	Type           string // "Workflow", "Activity", etc.
	Status         string // "Pending", "Running", etc.
	StepID         string
	RunID          int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Executions     []*Execution
	WorkflowData   *WorkflowData
	ActivityData   *ActivityData
	SideEffectData *SideEffectData
	SagaData       *SagaData
	Paused         bool
	Resumable      bool
	HandlerInfo    *HandlerInfo
}

// Execution represents a single execution attempt of an entity.
type Execution struct {
	ID                      int
	EntityID                int
	StartedAt               time.Time
	CompletedAt             *time.Time
	Status                  string // "Pending", "Running", etc.
	Attempt                 int
	Error                   string // Error message
	CreatedAt               time.Time
	UpdatedAt               time.Time
	WorkflowExecutionData   *WorkflowExecutionData
	ActivityExecutionData   *ActivityExecutionData
	SideEffectExecutionData *SideEffectExecutionData
	SagaExecutionData       *SagaExecutionData
}

// Hierarchy tracks parent-child relationships between entities.
type Hierarchy struct {
	RunID             int
	ParentEntityID    int
	ChildEntityID     int
	ParentExecutionID int
	ChildExecutionID  int
	ParentStepID      string
	ChildStepID       string
	ParentType        string
	ChildType         string
}

// Version tracks entity versions.
type Version struct {
	ID       int
	EntityID int                    // The Entity ID (workflow ID)
	ChangeID string                 // The unique identifier for the code change
	Version  int                    // The version
	Data     map[string]interface{} // Additional data if needed
}

// Run represents a workflow execution group.
type Run struct {
	ID        int
	Status    string // "Pending", "Running", "Completed", etc.
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*Entity
}

// Data structures matching the ent schemas
type WorkflowData struct {
	Duration    string               `json:"duration,omitempty"`
	Paused      bool                 `json:"paused"`
	Resumable   bool                 `json:"resumable"`
	RetryState  *RetryState          `json:"retry_state"`
	RetryPolicy *retryPolicyInternal `json:"retry_policy"`
	Input       [][]byte             `json:"input,omitempty"`
}

type ActivityData struct {
	Timeout      int64                `json:"timeout,omitempty"`
	MaxAttempts  int                  `json:"max_attempts"`
	ScheduledFor *time.Time           `json:"scheduled_for,omitempty"`
	RetryPolicy  *retryPolicyInternal `json:"retry_policy"`
	Input        [][]byte             `json:"input,omitempty"`
	Output       [][]byte             `json:"output,omitempty"`
}

type SideEffectData struct {
	Input       [][]byte             `json:"input,omitempty"`
	RetryPolicy *retryPolicyInternal `json:"retry_policy"`
}

type SagaData struct {
	Compensating     bool                 `json:"compensating"`
	CompensationData [][]byte             `json:"compensation_data,omitempty"`
	RetryState       *RetryState          `json:"retry_state"`
	RetryPolicy      *retryPolicyInternal `json:"retry_policy"`
}

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	results      []interface{}
	err          error
	fsm          *stateless.StateMachine
	future       *Future
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	options      WorkflowOptions
	entity       *Entity
	entityID     int
	executionID  int
	execution    *Execution // Current execution
	//isRoot       bool        // Indicates if this is the root workflow instance
}

func (wi *WorkflowInstance) Start() {
	// Initialize the FSM
	wi.fsm = stateless.NewStateMachine(StateIdle)
	wi.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	wi.fsm.Configure(StateExecuting).
		OnEntry(wi.executeWorkflow).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	wi.fsm.Configure(StateCompleted).
		OnEntry(wi.onCompleted)

	wi.fsm.Configure(StateFailed).
		OnEntry(wi.onFailed)

	wi.fsm.Configure(StatePaused).
		OnEntry(wi.onPaused).
		Permit(TriggerResume, StateExecuting)

	// Start the FSM
	go wi.fsm.Fire(TriggerStart)
}

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, _ ...interface{}) error {
	wi.executeWithRetry()
	return nil
}

func (wi *WorkflowInstance) executeWithRetry() {
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration

	if wi.options.RetryPolicy != nil {
		maxAttempts = wi.options.RetryPolicy.MaxAttempts
		initialInterval = wi.options.RetryPolicy.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		if wi.orchestrator.IsPaused() {
			log.Printf("WorkflowInstance %s is paused", wi.stepID)
			wi.entity.Status = string(StatusPaused)
			wi.entity.Paused = true
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &Execution{
			EntityID:  wi.entity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = wi.orchestrator.db.AddExecution(execution)
		wi.entity.Executions = append(wi.entity.Executions, execution)
		executionID := execution.ID
		wi.executionID = executionID // Store execution ID
		wi.execution = execution

		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.entityID, executionID)

		err := wi.runWorkflow(execution)
		if errors.Is(err, ErrPaused) {
			wi.entity.Status = string(StatusPaused)
			wi.entity.Paused = true
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			wi.entity.Status = string(StatusCompleted)
			wi.orchestrator.db.UpdateExecution(execution)
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			wi.err = err
			wi.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.entityID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				wi.entity.Status = string(StatusFailed)
				wi.orchestrator.db.UpdateEntity(wi.entity)
				wi.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (wi *WorkflowInstance) runWorkflow(execution *Execution) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d, Execution ID: %d) runWorkflow attempt %d", wi.stepID, wi.entityID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := wi.orchestrator.db.GetLatestExecution(wi.entity.ID)
	if latestExecution != nil && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", wi.entity.ID)
		outputs, err := ConvertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Output)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return err
		}
		wi.results = outputs
		return nil
	}

	handler := wi.handler
	f := handler.Handler

	ctxWorkflow := &WorkflowContext{
		orchestrator: wi.orchestrator,
		ctx:          wi.ctx,
		workflowID:   wi.entity.ID,
		stepID:       wi.stepID,
		options:      &wi.options,
	}

	// Convert inputs from serialization
	inputs, err := ConvertInputsFromSerialization(handler, wi.entity.WorkflowData.Input)
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
			log.Printf("Panic in workflow: %v", r)
			wi.err = fmt.Errorf("panic: %v", r)
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		wi.err = wi.ctx.Err()
		return wi.err
	default:
	}

	results := reflect.ValueOf(f).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		wi.err = err
		return err
	}

	errInterface := results[numOut-1].Interface()

	if errInterface != nil {
		log.Printf("Workflow returned error: %v", errInterface)
		wi.err = errInterface.(error)

		// Serialize error
		errorMessage := wi.err.Error()

		// Create WorkflowExecutionData
		workflowExecutionData := &WorkflowExecutionData{
			Error:  errorMessage,
			Output: nil,
		}

		// Update the execution with the execution data
		wi.execution.WorkflowExecutionData = workflowExecutionData
		wi.orchestrator.db.UpdateExecution(wi.execution)

		return wi.err
	} else {
		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				log.Printf("Workflow returned result [%d]: %v", i, result)
				outputs = append(outputs, result)
			}
		}

		wi.results = outputs

		// Serialize output
		outputBytes, err := ConvertOutputsForSerialization(wi.results)
		if err != nil {
			log.Printf("Error serializing output: %v", err)
			return err
		}

		// Create WorkflowExecutionData
		workflowExecutionData := &WorkflowExecutionData{
			Error:  "",
			Output: outputBytes,
		}

		// Update the execution with the execution data
		wi.execution.WorkflowExecutionData = workflowExecutionData
		wi.orchestrator.db.UpdateExecution(wi.execution)

		return nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.entityID)
	if wi.future != nil {
		wi.future.setResult(wi.results)
	}

	// If this is the root workflow, update the Run status to Completed
	if wi.orchestrator.rootWf == wi {
		run := wi.orchestrator.db.GetRun(wi.orchestrator.runID)
		if run != nil {
			run.Status = string(StatusCompleted)
			run.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateRun(run)
		}
	}

	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.entityID)
	if wi.future != nil {
		wi.future.setError(wi.err)
	}

	// If this is the root workflow, update the Run status to Failed
	if wi.orchestrator.rootWf == wi {
		run := wi.orchestrator.db.GetRun(wi.orchestrator.runID)
		if run != nil {
			run.Status = string(StatusFailed)
			run.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateRun(run)
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

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	results      []interface{}
	err          error
	fsm          *stateless.StateMachine
	future       *Future
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	options      ActivityOptions
	entity       *Entity
	entityID     int
	executionID  int
	execution    *Execution // Current execution
}

func (ai *ActivityInstance) Start() {
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

	// Start the FSM
	go ai.fsm.Fire(TriggerStart)
}

func (ai *ActivityInstance) executeActivity(_ context.Context, _ ...interface{}) error {
	ai.executeWithRetry()
	return nil
}

func (ai *ActivityInstance) executeWithRetry() {
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration

	if ai.options.RetryPolicy != nil {
		maxAttempts = ai.options.RetryPolicy.MaxAttempts
		initialInterval = ai.options.RetryPolicy.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		if ai.orchestrator.IsPaused() {
			log.Printf("ActivityInstance %s is paused", ai.stepID)
			ai.entity.Status = string(StatusPaused)
			ai.entity.Paused = true
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &Execution{
			EntityID:  ai.entity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = ai.orchestrator.db.AddExecution(execution)
		ai.entity.Executions = append(ai.entity.Executions, execution)
		executionID := execution.ID
		ai.executionID = executionID // Store execution ID
		ai.execution = execution

		log.Printf("Executing activity %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.entityID, executionID)

		err := ai.runActivity(execution)
		if errors.Is(err, ErrPaused) {
			ai.entity.Status = string(StatusPaused)
			ai.entity.Paused = true
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			ai.entity.Status = string(StatusCompleted)
			ai.orchestrator.db.UpdateExecution(execution)
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			ai.err = err
			ai.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying activity %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", ai.stepID, ai.entityID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				ai.entity.Status = string(StatusFailed)
				ai.orchestrator.db.UpdateEntity(ai.entity)
				ai.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (ai *ActivityInstance) runActivity(execution *Execution) error {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d) runActivity attempt %d", ai.stepID, ai.entityID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := ai.orchestrator.db.GetLatestExecution(ai.entity.ID)
	if latestExecution != nil && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", ai.entity.ID)
		outputs, err := ConvertOutputsFromSerialization(ai.handler, latestExecution.ActivityExecutionData.Output)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return err
		}
		ai.results = outputs
		return nil
	}

	handler := ai.handler
	f := handler.Handler

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in activity: %v", r)
			log.Printf("Panic in activity: %v", err)
			ai.err = err
		}
	}()

	// Convert inputs from serialization
	inputs, err := ConvertInputsFromSerialization(handler, ai.entity.ActivityData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return err
	}

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range inputs {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	select {
	case <-ai.ctx.Done():
		log.Printf("Context cancelled in activity")
		ai.err = ai.ctx.Err()
		return ai.err
	default:
	}

	results := reflect.ValueOf(f).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("activity %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		ai.err = err
		return err
	}

	errInterface := results[numOut-1].Interface()
	if errInterface != nil {
		log.Printf("Activity returned error: %v", errInterface)
		ai.err = errInterface.(error)

		// Serialize error
		errorMessage := ai.err.Error()

		// Create ActivityExecutionData
		activityExecutionData := &ActivityExecutionData{
			ErrMessage: errorMessage,
			Output:     nil,
		}

		// Update the execution with the execution data
		ai.execution.ActivityExecutionData = activityExecutionData
		ai.orchestrator.db.UpdateExecution(ai.execution)

		return ai.err
	} else {
		outputs := []interface{}{}
		if numOut > 1 {
			for i := 0; i < numOut-1; i++ {
				result := results[i].Interface()
				log.Printf("Activity returned result [%d]: %v", i, result)
				outputs = append(outputs, result)
			}
		}
		ai.results = outputs

		// Serialize output
		outputBytes, err := ConvertOutputsForSerialization(ai.results)
		if err != nil {
			log.Printf("Error serializing output: %v", err)
			return err
		}

		// Create ActivityExecutionData
		activityExecutionData := &ActivityExecutionData{
			Output: outputBytes,
		}

		// Update the execution with the execution data
		ai.execution.ActivityExecutionData = activityExecutionData
		ai.orchestrator.db.UpdateExecution(ai.execution)

		return nil
	}
}

func (ai *ActivityInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onCompleted called", ai.stepID, ai.entityID)
	if ai.future != nil {
		ai.future.setResult(ai.results)
	}
	return nil
}

func (ai *ActivityInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onFailed called", ai.stepID, ai.entityID)
	if ai.future != nil {
		ai.future.setError(ai.err)
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

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
	stepID         string
	sideEffectFunc interface{}
	results        []interface{}
	err            error
	fsm            *stateless.StateMachine
	future         *Future
	ctx            context.Context
	orchestrator   *Orchestrator
	workflowID     int
	entity         *Entity
	entityID       int
	executionID    int
	options        WorkflowOptions
	execution      *Execution // Current execution
	returnTypes    []reflect.Type
	handlerName    string
	handler        HandlerInfo
}

func (sei *SideEffectInstance) Start() {
	// Initialize the FSM
	sei.fsm = stateless.NewStateMachine(StateIdle)
	sei.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	sei.fsm.Configure(StateExecuting).
		OnEntry(sei.executeSideEffect).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	sei.fsm.Configure(StateCompleted).
		OnEntry(sei.onCompleted)

	sei.fsm.Configure(StateFailed).
		OnEntry(sei.onFailed)

	sei.fsm.Configure(StatePaused).
		OnEntry(sei.onPaused).
		Permit(TriggerResume, StateExecuting)

	// Start the FSM
	go sei.fsm.Fire(TriggerStart)
}

func (sei *SideEffectInstance) executeSideEffect(_ context.Context, _ ...interface{}) error {
	sei.executeWithRetry()
	return nil
}

func (sei *SideEffectInstance) executeWithRetry() {
	var attempt int
	var retryPolicy *RetryPolicy

	// Get the retry policy from the entity data
	if sei.entity.SideEffectData != nil && sei.entity.SideEffectData.RetryPolicy != nil {
		retryPolicy = &RetryPolicy{
			MaxAttempts:        sei.entity.SideEffectData.RetryPolicy.MaxAttempts,
			InitialInterval:    time.Duration(sei.entity.SideEffectData.RetryPolicy.InitialInterval),
			BackoffCoefficient: sei.entity.SideEffectData.RetryPolicy.BackoffCoefficient,
			MaxInterval:        time.Duration(sei.entity.SideEffectData.RetryPolicy.MaxInterval),
		}
	}

	if retryPolicy == nil {
		// Default retry policy with MaxAttempts=1
		retryPolicy = &RetryPolicy{
			MaxAttempts:     1,
			InitialInterval: 0,
		}
	}

	for attempt = 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
		if sei.orchestrator.IsPaused() {
			log.Printf("SideEffectInstance %s is paused", sei.stepID)
			sei.entity.Status = string(StatusPaused)
			sei.entity.Paused = true
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &Execution{
			EntityID:  sei.entity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = sei.orchestrator.db.AddExecution(execution)
		sei.entity.Executions = append(sei.entity.Executions, execution)
		executionID := execution.ID
		sei.executionID = executionID // Store execution ID
		sei.execution = execution

		log.Printf("Executing side effect %s (Entity ID: %d, Execution ID: %d)", sei.stepID, sei.entityID, executionID)

		err := sei.runSideEffect(execution)
		if errors.Is(err, ErrPaused) {
			sei.entity.Status = string(StatusPaused)
			sei.entity.Paused = true
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			sei.entity.Status = string(StatusCompleted)
			sei.orchestrator.db.UpdateExecution(execution)
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			sei.err = err
			sei.orchestrator.db.UpdateExecution(execution)
			if attempt < retryPolicy.MaxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying side effect %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", sei.stepID, sei.entityID, executionID, attempt+1, retryPolicy.MaxAttempts, retryPolicy.InitialInterval)
				time.Sleep(retryPolicy.InitialInterval)
			} else {
				// Max attempts reached
				sei.entity.Status = string(StatusFailed)
				sei.orchestrator.db.UpdateEntity(sei.entity)
				sei.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (sei *SideEffectInstance) runSideEffect(execution *Execution) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d, Execution ID: %d) runSideEffect attempt %d", sei.stepID, sei.entityID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := sei.orchestrator.db.GetLatestExecution(sei.entity.ID)
	if latestExecution != nil && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", sei.entity.ID)
		outputs, err := ConvertOutputsFromSerialization(HandlerInfo{ReturnTypes: sei.returnTypes}, latestExecution.SideEffectExecutionData.Output)
		if err != nil {
			log.Printf("Error deserializing side effect result: %v", err)
			return err
		}
		sei.results = outputs
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in side effect: %v", r)
			log.Printf("Panic in side effect: %v", err)
			sei.err = err
		}
	}()

	select {
	case <-sei.ctx.Done():
		log.Printf("Context cancelled in side effect")
		sei.err = sei.ctx.Err()
		return sei.err
	default:
	}

	// Retrieve the function from handler
	f := sei.handler.Handler

	argsValues := []reflect.Value{}
	results := reflect.ValueOf(f).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("side effect should return at least a value")
		log.Printf("Error: %v", err)
		sei.err = err
		return err
	}

	outputs := []interface{}{}
	for i := 0; i < numOut; i++ {
		result := results[i].Interface()
		log.Printf("Side effect returned result [%d]: %v", i, result)
		outputs = append(outputs, result)
	}
	sei.results = outputs

	// Serialize output
	outputBytes, err := ConvertOutputsForSerialization(sei.results)
	if err != nil {
		log.Printf("Error serializing output: %v", err)
		return err
	}

	// Create SideEffectExecutionData
	sideEffectExecutionData := &SideEffectExecutionData{
		Output: outputBytes,
	}

	// Update the execution with the execution data
	sei.execution.SideEffectExecutionData = sideEffectExecutionData
	sei.orchestrator.db.UpdateExecution(sei.execution)

	return nil
}

func (sei *SideEffectInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onCompleted called", sei.stepID, sei.entityID)
	if sei.future != nil {
		sei.future.setResult(sei.results)
	}
	return nil
}

func (sei *SideEffectInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onFailed called", sei.stepID, sei.entityID)
	if sei.future != nil {
		sei.future.setError(sei.err)
	}
	return nil
}

func (sei *SideEffectInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onPaused called", sei.stepID, sei.entityID)
	if sei.future != nil {
		sei.future.setError(ErrPaused)
	}
	return nil
}

// Saga types and implementations

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

type SagaInfo struct {
	err    error
	result interface{}
	done   chan struct{}
}

func (s *SagaInfo) Get() error {
	<-s.done
	return s.err
}

func (o *Orchestrator) executeSaga(ctx *WorkflowContext, stepID string, saga *SagaDefinition) *SagaInfo {
	sagaInfo := &SagaInfo{
		done: make(chan struct{}),
	}

	// Create a new Entity for the Saga
	entity := &Entity{
		StepID:    stepID,
		Type:      string(EntityTypeSaga),
		Status:    string(StatusPending),
		RunID:     o.runID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		SagaData:  &SagaData{},
	}

	// Add the entity to the database
	entity = o.db.AddEntity(entity)

	// Record hierarchy relationship
	hierarchy := &Hierarchy{
		RunID:             o.runID,
		ParentEntityID:    ctx.workflowID,
		ChildEntityID:     entity.ID,
		ParentExecutionID: 0,
		ChildExecutionID:  0,
		ParentStepID:      ctx.stepID,
		ChildStepID:       stepID,
		ParentType:        string(EntityTypeWorkflow),
		ChildType:         string(EntityTypeSaga),
	}
	o.db.AddHierarchy(hierarchy)

	// Create a SagaInstance and start it
	sagaInstance := &SagaInstance{
		saga:         saga,
		ctx:          ctx.ctx,
		orchestrator: o,
		workflowID:   ctx.workflowID,
		stepID:       stepID,
		sagaInfo:     sagaInfo,
		entity:       entity,
	}

	o.addSagaInstance(sagaInstance)
	sagaInstance.Start()

	return sagaInfo
}

type TransactionContext struct {
	ctx context.Context
}

type CompensationContext struct {
	ctx context.Context
}

type SagaInstance struct {
	saga          *SagaDefinition
	ctx           context.Context
	orchestrator  *Orchestrator
	workflowID    int
	stepID        string
	sagaInfo      *SagaInfo
	entity        *Entity
	fsm           *stateless.StateMachine
	err           error
	currentStep   int
	compensations []int // Indices of steps to compensate
	mu            sync.Mutex
}

// Start initializes and starts the Saga FSM.
func (si *SagaInstance) Start() {
	si.fsm = stateless.NewStateMachine(StateIdle)

	// Configure FSM states and transitions
	si.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateTransactions)

	si.fsm.Configure(StateTransactions).
		OnEntry(si.executeTransactions).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateCompensations)

	si.fsm.Configure(StateCompensations).
		OnEntry(si.executeCompensations).
		Permit(TriggerCompensate, StateFailed)

	si.fsm.Configure(StateCompleted).
		OnEntry(si.onCompleted)

	si.fsm.Configure(StateFailed).
		OnEntry(si.onFailed)

	// Start the FSM
	_ = si.fsm.Fire(TriggerStart)
}

// executeTransactions executes the saga transactions sequentially.
func (si *SagaInstance) executeTransactions(ctx context.Context, args ...interface{}) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	for si.currentStep < len(si.saga.Steps) {
		step := si.saga.Steps[si.currentStep]
		_, err := step.Transaction(TransactionContext{
			ctx: si.ctx,
		})
		if err != nil {
			// Transaction failed
			log.Printf("Transaction failed at step %d: %v", si.currentStep, err)
			// Record the steps up to the last successful one
			if si.currentStep > 0 {
				si.compensations = si.compensations[:si.currentStep]
			}
			si.err = fmt.Errorf("transaction failed at step %d: %v", si.currentStep, err)
			_ = si.fsm.Fire(TriggerFail)
			return nil
		}
		// Record the successful transaction
		si.compensations = append(si.compensations, si.currentStep)
		si.currentStep++
	}
	// All transactions succeeded
	_ = si.fsm.Fire(TriggerComplete)
	return nil
}

// executeCompensations compensates executed transactions in reverse order.
func (si *SagaInstance) executeCompensations(ctx context.Context, args ...interface{}) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Compensate in reverse order
	for i := len(si.compensations) - 1; i >= 0; i-- {
		stepIndex := si.compensations[i]
		step := si.saga.Steps[stepIndex]
		_, err := step.Compensation(CompensationContext{
			ctx: si.ctx,
		})
		if err != nil {
			// Compensation failed
			log.Printf("Compensation failed for step %d: %v", stepIndex, err)
			si.err = fmt.Errorf("compensation failed at step %d: %v", stepIndex, err)
			break // Exit after first compensation failure
		}
	}
	// All compensations completed (successfully or not)
	_ = si.fsm.Fire(TriggerCompensate)
	return nil
}

// onCompleted handles the completion of the Saga.
func (si *SagaInstance) onCompleted(ctx context.Context, args ...interface{}) error {
	// Update entity status to Completed
	si.entity.Status = string(StatusCompleted)
	si.orchestrator.db.UpdateEntity(si.entity)

	// Notify SagaInfo
	si.sagaInfo.err = nil
	close(si.sagaInfo.done)
	log.Println("Saga completed successfully")
	return nil
}

// onFailed handles the failure of the Saga.
func (si *SagaInstance) onFailed(ctx context.Context, args ...interface{}) error {
	// Update entity status to Failed
	si.entity.Status = string(StatusFailed)
	si.orchestrator.db.UpdateEntity(si.entity)

	// Set the error in SagaInfo
	if si.err != nil {
		si.sagaInfo.err = si.err
	} else {
		si.sagaInfo.err = errors.New("saga execution failed")
	}

	// Mark parent Workflow as Failed
	parentEntity := si.orchestrator.db.GetEntity(si.workflowID)
	if parentEntity != nil {
		parentEntity.Status = string(StatusFailed)
		si.orchestrator.db.UpdateEntity(parentEntity)
	}

	close(si.sagaInfo.done)
	log.Printf("Saga failed with error: %v", si.sagaInfo.err)
	return nil
}

func (o *Orchestrator) addSagaInstance(si *SagaInstance) {
	o.sagasMu.Lock()
	o.sagas = append(o.sagas, si)
	o.sagasMu.Unlock()
}

// Database interface defines methods for interacting with the data store.
type Database interface {
	// Run methods
	AddRun(run *Run) *Run
	GetRun(id int) *Run
	UpdateRun(run *Run)

	// Entity methods
	AddEntity(entity *Entity) *Entity
	GetEntity(id int) *Entity
	UpdateEntity(entity *Entity)
	GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) *Entity
	GetChildEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *Entity

	// Execution methods
	AddExecution(execution *Execution) *Execution
	GetExecution(id int) *Execution
	UpdateExecution(execution *Execution)
	GetLatestExecution(entityID int) *Execution

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy)
	GetHierarchy(parentID, childID int) *Hierarchy

	// Version methods
	GetVersion(entityID int, changeID string) *Version
	SetVersion(version *Version) *Version
}

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs        map[int]*Run
	entities    map[int]*Entity
	executions  map[int]*Execution
	versions    map[int]*Version
	hierarchies []*Hierarchy
	mu          sync.Mutex
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{
		runs:       make(map[int]*Run),
		entities:   make(map[int]*Entity),
		executions: make(map[int]*Execution),
		versions:   make(map[int]*Version),
	}
}

var (
	runIDCounter       int64
	entityIDCounter    int64
	executionIDCounter int64
	versionIDCounter   int64
)

func (db *DefaultDatabase) AddRun(run *Run) *Run {
	db.mu.Lock()
	defer db.mu.Unlock()
	runID := int(atomic.AddInt64(&runIDCounter, 1))
	run.ID = runID
	db.runs[run.ID] = run
	return run
}

func (db *DefaultDatabase) GetRun(id int) *Run {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.runs[id]
}

func (db *DefaultDatabase) UpdateRun(run *Run) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.runs[run.ID] = run
}

func (db *DefaultDatabase) AddEntity(entity *Entity) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&entityIDCounter, 1))
	entity.ID = entityID
	db.entities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetEntity(id int) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.entities[id]
}

func (db *DefaultDatabase) UpdateEntity(entity *Entity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.entities[entity.ID] = entity
}

func (db *DefaultDatabase) AddExecution(execution *Execution) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&executionIDCounter, 1))
	execution.ID = executionID
	db.executions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetExecution(id int) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.executions[id]
}

func (db *DefaultDatabase) UpdateExecution(execution *Execution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.executions[execution.ID] = execution
}

func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.hierarchies = append(db.hierarchies, hierarchy)
}

func (db *DefaultDatabase) GetHierarchy(parentID, childID int) *Hierarchy {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentID && h.ChildEntityID == childID {
			return h
		}
	}
	return nil
}

func (db *DefaultDatabase) GetLatestExecution(entityID int) *Execution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

func (db *DefaultDatabase) GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, entity := range db.entities {
		if entity.ID == workflowID && entity.StepID == stepID {
			return entity
		}
	}
	return nil
}

func (db *DefaultDatabase) GetChildEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			return db.entities[childEntityID]
		}
	}
	return nil
}

func (db *DefaultDatabase) GetVersion(entityID int, changeID string) *Version {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, version := range db.versions {
		if version.EntityID == entityID && version.ChangeID == changeID {
			return version
		}
	}
	return nil
}

func (db *DefaultDatabase) SetVersion(version *Version) *Version {
	db.mu.Lock()
	defer db.mu.Unlock()
	versionID := int(atomic.AddInt64(&versionIDCounter, 1))
	version.ID = versionID
	db.versions[version.ID] = version
	return version
}

// Clear removes all Runs that are 'Completed' and their associated data.
func (db *DefaultDatabase) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	runsToDelete := []int{}
	entitiesToDelete := map[int]*Entity{}
	executionsToDelete := map[int]*Execution{}
	hierarchiesToKeep := []*Hierarchy{}
	versionsToDelete := map[int]*Version{}

	// Find Runs to delete
	for runID, run := range db.runs {
		if run.Status == string(StatusCompleted) {
			runsToDelete = append(runsToDelete, runID)
			// Collect Entities associated with the Run
			for _, entity := range run.Entities {
				entitiesToDelete[entity.ID] = entity
			}
		}
	}

	// Collect Executions associated with Entities to delete
	for execID, execution := range db.executions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = execution
		}
	}

	// Collect Versions associated with Entities to delete
	for versionID, version := range db.versions {
		if _, exists := entitiesToDelete[version.EntityID]; exists {
			versionsToDelete[versionID] = version
		}
	}

	// Filter Hierarchies to keep only those not associated with Entities to delete
	for _, hierarchy := range db.hierarchies {
		if _, parentExists := entitiesToDelete[hierarchy.ParentEntityID]; parentExists {
			continue
		}
		if _, childExists := entitiesToDelete[hierarchy.ChildEntityID]; childExists {
			continue
		}
		hierarchiesToKeep = append(hierarchiesToKeep, hierarchy)
	}

	// Delete Runs
	for _, runID := range runsToDelete {
		delete(db.runs, runID)
	}

	// Delete Entities
	for entityID := range entitiesToDelete {
		delete(db.entities, entityID)
	}

	// Delete Executions
	for execID := range executionsToDelete {
		delete(db.executions, execID)
	}

	// Delete Versions
	for versionID := range versionsToDelete {
		delete(db.versions, versionID)
	}

	// Replace hierarchies with the filtered ones
	db.hierarchies = hierarchiesToKeep
}

// Orchestrator orchestrates the execution of workflows and activities.
type Orchestrator struct {
	db            Database
	registry      *Registry
	rootWf        *WorkflowInstance
	ctx           context.Context
	cancel        context.CancelFunc
	instances     []*WorkflowInstance
	activities    []*ActivityInstance
	sideEffects   []*SideEffectInstance
	sagas         []*SagaInstance
	instancesMu   sync.Mutex
	activitiesMu  sync.Mutex
	sideEffectsMu sync.Mutex
	sagasMu       sync.Mutex
	err           error
	runID         int
	paused        bool
	pausedMu      sync.Mutex
}

func NewOrchestrator(ctx context.Context, db Database, registry *Registry, runID int) *Orchestrator {
	log.Printf("NewOrchestrator called")
	ctx, cancel := context.WithCancel(ctx)
	o := &Orchestrator{
		db:       db,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
		runID:    runID,
	}
	return o
}

func (o *Orchestrator) Workflow(workflowFunc interface{}, options WorkflowOptions, args ...interface{}) *Future {
	handler, err := o.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		log.Printf("Error registering workflow: %v", err)
		return NewFuture(0)
	}

	if o.runID == 0 {
		// Create a new Run
		run := &Run{
			Status:    string(StatusPending),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		run = o.db.AddRun(run)
		o.runID = run.ID
	}

	// Update Run status to Running
	run := o.db.GetRun(o.runID)
	if run != nil {
		run.Status = string(StatusRunning)
		run.UpdatedAt = time.Now()
		o.db.UpdateRun(run)
	}

	// Convert inputs to [][]byte
	inputBytes, err := ConvertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		return NewFuture(0)
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal
	if options.RetryPolicy != nil {
		internalRetryPolicy = &retryPolicyInternal{
			MaxAttempts:        options.RetryPolicy.MaxAttempts,
			InitialInterval:    options.RetryPolicy.InitialInterval.Nanoseconds(),
			BackoffCoefficient: options.RetryPolicy.BackoffCoefficient,
			MaxInterval:        options.RetryPolicy.MaxInterval.Nanoseconds(),
		}
	}

	// Create WorkflowData
	workflowData := &WorkflowData{
		Duration:    "",
		Paused:      false,
		Resumable:   false,
		RetryState:  &RetryState{Attempts: 0},
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:       "root",
		HandlerName:  handler.HandlerName,
		Type:         string(EntityTypeWorkflow),
		Status:       string(StatusPending),
		RunID:        o.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = o.db.AddEntity(entity)

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:       "root",
		handler:      handler,
		input:        args,
		ctx:          o.ctx,
		orchestrator: o,
		workflowID:   entity.ID,
		options:      options,
		entity:       entity,
		entityID:     entity.ID, // Store entity ID
		//isRoot:       true,
	}

	future := NewFuture(entity.ID)
	instance.future = future

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return future
}

func (o *Orchestrator) Pause() {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = true
	o.cancel()
	log.Printf("Orchestrator paused")
}

func (o *Orchestrator) IsPaused() bool {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	return o.paused
}

func (o *Orchestrator) Resume(entityID int) *Future {
	o.pausedMu.Lock()
	o.paused = false
	o.pausedMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	o.ctx = ctx
	o.cancel = cancel

	// Retrieve the workflow entity
	entity := o.db.GetEntity(entityID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", entityID)
		return NewFuture(0)
	}

	// Set the runID from the entity
	o.runID = entity.RunID

	// Update the entity's paused state
	entity.Paused = false
	entity.Resumable = true
	o.db.UpdateEntity(entity)

	// Update Run status to Running
	run := o.db.GetRun(o.runID)
	if run != nil {
		run.Status = string(StatusRunning)
		run.UpdatedAt = time.Now()
		o.db.UpdateRun(run)
	}

	// Retrieve the handler
	handlerInfo := entity.HandlerInfo
	if handlerInfo == nil {
		log.Printf("No handler info found for workflow: %s", entity.HandlerName)
		return NewFuture(0)
	}
	handler := *handlerInfo

	// Convert inputs from serialization
	inputs, err := ConvertInputsFromSerialization(handler, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return NewFuture(0)
	}

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:       entity.StepID,
		handler:      handler,
		input:        inputs,
		ctx:          o.ctx,
		orchestrator: o,
		workflowID:   entity.ID,
		options: WorkflowOptions{
			RetryPolicy: &RetryPolicy{
				MaxAttempts:        int(entity.WorkflowData.RetryPolicy.MaxAttempts),
				InitialInterval:    time.Duration(entity.WorkflowData.RetryPolicy.InitialInterval),
				BackoffCoefficient: entity.WorkflowData.RetryPolicy.BackoffCoefficient,
				MaxInterval:        time.Duration(entity.WorkflowData.RetryPolicy.MaxInterval),
			},
			VersionOverrides: nil, // Adjust if necessary
		},
		entity:   entity,
		entityID: entity.ID,
	}

	future := NewFuture(entity.ID)
	instance.future = future

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return future
}

func (o *Orchestrator) stopWithError(err error) {
	o.err = err
	if o.rootWf != nil && o.rootWf.fsm != nil {
		o.rootWf.fsm.Fire(TriggerFail)
	}
}

func (o *Orchestrator) addWorkflowInstance(wi *WorkflowInstance) {
	o.instancesMu.Lock()
	o.instances = append(o.instances, wi)
	o.instancesMu.Unlock()
}

func (o *Orchestrator) addActivityInstance(ai *ActivityInstance) {
	o.activitiesMu.Lock()
	o.activities = append(o.activities, ai)
	o.activitiesMu.Unlock()
}

func (o *Orchestrator) addSideEffectInstance(sei *SideEffectInstance) {
	o.sideEffectsMu.Lock()
	o.sideEffects = append(o.sideEffects, sei)
	o.sideEffectsMu.Unlock()
}

func (o *Orchestrator) Wait() {
	log.Printf("Orchestrator.Wait called")

	if o.rootWf == nil {
		log.Printf("No root workflow to execute")
		return
	}

	// Wait for root workflow to complete
	for {
		state := o.rootWf.fsm.MustState()
		log.Printf("Root Workflow FSM state: %s", state)
		if state == StateCompleted || state == StateFailed {
			break
		}
		select {
		case <-o.ctx.Done():
			log.Printf("Context cancelled: %v", o.ctx.Err())
			o.rootWf.fsm.Fire(TriggerFail)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Root workflow has completed or failed
	// The Run's status should have been updated in onCompleted or onFailed of the root workflow

	// Check if rootWf has any error
	if o.rootWf.err != nil {
		fmt.Printf("Root workflow failed with error: %v\n", o.rootWf.err)
	} else {
		fmt.Printf("Root workflow completed successfully with results: %v\n", o.rootWf.results)
	}
}

// Registry holds registered workflows, activities, and side effects.
type Registry struct {
	workflows   map[string]HandlerInfo
	activities  map[string]HandlerInfo
	sideEffects map[string]HandlerInfo
	mu          sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		workflows:   make(map[string]HandlerInfo),
		activities:  make(map[string]HandlerInfo),
		sideEffects: make(map[string]HandlerInfo),
	}
}

func (r *Registry) RegisterWorkflow(workflowFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(workflowFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.workflows[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(workflowFunc)
	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("workflow must be a function")
		return HandlerInfo{}, err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
		return HandlerInfo{}, err
	}

	expectedContextType := reflect.TypeOf(&WorkflowContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of workflow function must be *WorkflowContext")
		return HandlerInfo{}, err
	}

	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	numOut := handlerType.NumOut()
	if numOut == 0 {
		err := fmt.Errorf("workflow function must return at least an error")
		return HandlerInfo{}, err
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		err := fmt.Errorf("last return value of workflow function must be error")
		return HandlerInfo{}, err
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         workflowFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1,
		NumOut:          numOut - 1,
	}

	r.workflows[funcName] = handler
	return handler, nil
}

func (r *Registry) RegisterActivity(activityFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(activityFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.activities[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(activityFunc)
	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("activity must be a function")
		return HandlerInfo{}, err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
		return HandlerInfo{}, err
	}

	expectedContextType := reflect.TypeOf(ActivityContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of activity function must be ActivityContext")
		return HandlerInfo{}, err
	}

	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	numOut := handlerType.NumOut()
	if numOut == 0 {
		err := fmt.Errorf("activity function must return at least an error")
		return HandlerInfo{}, err
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		err := fmt.Errorf("last return value of activity function must be error")
		return HandlerInfo{}, err
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         activityFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1,
		NumOut:          numOut - 1,
	}

	r.activities[funcName] = handler
	return handler, nil
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Example Saga Steps
type ReserveInventorySaga struct {
	Data int
}

func (s ReserveInventorySaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Printf("ReserveInventorySaga Transaction called with Data: %d", s.Data)
	// Simulate successful transaction
	return nil, nil
}

func (s ReserveInventorySaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Printf("ReserveInventorySaga Compensation called")
	// Simulate compensation
	return nil, nil
}

type ProcessPaymentSaga struct {
	Data int
}

func (s ProcessPaymentSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Transaction called with Data: %d", s.Data)
	// Simulate failure in transaction
	// if s.Data%2 == 0 {
	// 	return nil, fmt.Errorf("Payment processing failed")
	// }
	return nil, nil
}

func (s ProcessPaymentSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Compensation called")
	// Simulate compensation
	return nil, nil
}

type UpdateLedgerSaga struct {
	Data int
}

func (s UpdateLedgerSaga) Transaction(ctx TransactionContext) (interface{}, error) {
	log.Printf("UpdateLedgerSaga Transaction called with Data: %d", s.Data)
	// Simulate successful transaction
	return nil, nil
}

func (s UpdateLedgerSaga) Compensation(ctx CompensationContext) (interface{}, error) {
	log.Printf("UpdateLedgerSaga Compensation called")
	// Simulate compensation
	return nil, nil
}

// Example Activity
func SomeActivity(ctx ActivityContext, data int) (int, error) {
	log.Printf("SomeActivity called with data: %d", data)
	select {
	case <-time.After(time.Second * 2):
		// Simulate processing
	case <-ctx.ctx.Done():
		log.Printf("SomeActivity context cancelled")
		return -1, ctx.ctx.Err()
	}
	result := data * 3
	log.Printf("SomeActivity returning result: %d", result)
	return result, nil
}

func SubSubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	log.Printf("SubSubSubWorkflow returning result: %d", result)
	return result, nil
}

func SubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	if err := ctx.Workflow("subsubsubworkflow-step", SubSubSubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, result).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubSubWorkflow returning result: %d", result)
	return result, nil
}

var subWorkflowFailed atomic.Bool

func SubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from activity: %v", err)
		return -1, err
	}

	if subWorkflowFailed.Load() {
		subWorkflowFailed.Store(false)
		return -1, fmt.Errorf("subworkflow failed on purpose")
	}

	if err := ctx.Workflow("subsubworkflow-step", SubSubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, result).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubWorkflow returning result: %d", result)
	return result, nil
}

func SubManyWorkflow(ctx *WorkflowContext) (int, int, error) {
	return 1, 2, nil
}

func Workflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("Workflow called with data: %d", data)
	var value int

	select {
	case <-ctx.ctx.Done():
		log.Printf("Workflow context cancelled")
		return -1, ctx.ctx.Err()
	default:
	}

	// Versioning example
	changeID := "ChangeIDCalculateTax"
	const DefaultVersion = 0
	version := ctx.GetVersion(changeID, DefaultVersion, 1)
	if version == DefaultVersion {
		log.Printf("Using default version logic")
		// Original logic
	} else {
		log.Printf("Using new version logic")
		// New logic
	}

	var shouldDouble bool
	if err := ctx.SideEffect("side-effect-step", func() bool {
		log.Printf("Side effect called")
		return rand.Float32() < 0.5
	}, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        1,
			InitialInterval:    0,
			BackoffCoefficient: 1.0,
			MaxInterval:        0,
		},
	}).Get(&shouldDouble); err != nil {
		log.Printf("Workflow encountered error from side effect: %v", err)
		return -1, err
	}

	if err := ctx.Workflow("subworkflow-step", SubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&value); err != nil {
		log.Printf("Workflow encountered error from sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("Workflow received value from sub-workflow: %d", value)

	// Build and execute a saga
	orderData := value
	sagaBuilder := NewSaga()
	sagaBuilder.AddStep(ReserveInventorySaga{Data: orderData})
	sagaBuilder.AddStep(ProcessPaymentSaga{Data: orderData})
	sagaBuilder.AddStep(UpdateLedgerSaga{Data: orderData})

	saga, err := sagaBuilder.Build()
	if err != nil {
		return -1, fmt.Errorf("failed to build saga: %w", err)
	}

	err = ctx.Saga("process-order", saga).Get()
	if err != nil {
		return -1, fmt.Errorf("saga execution failed: %w", err)
	}

	result := value + data
	if shouldDouble {
		result *= 2
	}

	var a int
	var b int
	if err := ctx.Workflow("a-b", SubManyWorkflow, WorkflowOptions{}).Get(&a, &b); err != nil {
		log.Printf("Workflow encountered error from sub-many-workflow: %v", err)
		return -1, err
	}

	log.Printf("Workflow returning result: %d - %v %v", result, a, b)
	return result, nil
}

// GetVersion retrieves or sets a version for a changeID.
func (ctx *WorkflowContext) GetVersion(changeID string, minSupported, maxSupported int) int {
	// First check if version is overridden in options
	if ctx.options != nil && ctx.options.VersionOverrides != nil {
		if forcedVersion, ok := ctx.options.VersionOverrides[changeID]; ok {
			return forcedVersion
		}
	}

	version := ctx.orchestrator.db.GetVersion(ctx.workflowID, changeID)
	if version != nil {
		return version.Version
	}

	// If version not found, create a new version.
	newVersion := &Version{
		EntityID: ctx.workflowID,
		ChangeID: changeID,
		Version:  maxSupported,
	}

	ctx.orchestrator.db.SetVersion(newVersion)
	return newVersion.Version
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("main started")

	database := NewDefaultDatabase()
	registry := NewRegistry()
	registry.RegisterWorkflow(Workflow) // Only register the root workflow

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a new orchestrator with runID 0 (which will create a new run)
	orchestrator := NewOrchestrator(ctx, database, registry, 0)

	subWorkflowFailed.Store(true)

	future := orchestrator.Workflow(Workflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
		VersionOverrides: map[string]int{
			"ChangeIDCalculateTax": 1, // Force version 1
		},
	}, 40)

	// Simulate pause after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		log.Printf("Pausing orchestrator")
		orchestrator.Pause()
	})

	// Wait for some time to simulate the orchestrator being paused
	time.Sleep(3 * time.Second)

	log.Printf("Resuming orchestrator")
	// Retrieve the runID from the previous orchestrator
	runID := orchestrator.runID
	newOrchestrator := NewOrchestrator(ctx, database, registry, runID)

	future = newOrchestrator.Resume(future.workflowID)

	var result int
	if err := future.Get(&result); err != nil {
		log.Printf("Workflow failed with error: %v", err)
	} else {
		log.Printf("Workflow completed with result: %d", result)
	}

	newOrchestrator.Wait()

	// After execution, clear completed Runs to free up memory
	database.Clear()

	log.Printf("main finished")
}
