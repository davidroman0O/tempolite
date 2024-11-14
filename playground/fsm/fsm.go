package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qmuntal/stateless"
)

const (
	// FSM states
	StateIdle      = "Idle"
	StateExecuting = "Executing"
	StateCompleted = "Completed"
	StateFailed    = "Failed"
	StateRetried   = "Retried"

	// FSM triggers
	TriggerStart    = "Start"
	TriggerComplete = "Complete"
	TriggerFail     = "Fail"
)

// Future represents an asynchronous result.
type Future struct {
	result interface{}
	err    error
	done   chan struct{}
}

func NewFuture() *Future {
	return &Future{
		done: make(chan struct{}),
	}
}

func (f *Future) Get(result interface{}) error {
	log.Printf("Future.Get called")
	<-f.done
	if f.err != nil {
		log.Printf("Future.Get returning error: %v", f.err)
		return f.err
	}
	if result != nil && f.result != nil {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(f.result))
		log.Printf("Future.Get returning result: %v", f.result)
	}
	return nil
}

func (f *Future) setResult(result interface{}) {
	log.Printf("Future.setResult called with result: %v", result)
	f.result = result
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

// WorkflowOptions provides options for workflows, including retry policies.
type WorkflowOptions struct {
	RetryPolicy *RetryPolicy
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
}

func (ctx *WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options WorkflowOptions, args ...interface{}) *Future {
	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)
	future := NewFuture()

	// Check if result already exists in the database
	if result, ok := ctx.orchestrator.db.GetResult(ctx.workflowID, stepID); ok {
		log.Printf("Result found in database for workflowID: %d, stepID: %s", ctx.workflowID, stepID)
		future.setResult(result)
		return future
	}

	// Register workflow on-the-fly
	handler, err := ctx.orchestrator.registerWorkflow(workflowFunc)
	if err != nil {
		log.Printf("Error registering workflow: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:      stepID,
		HandlerName: handler.HandlerName,
		Type:        EntityTypeWorkflow,
		Status:      StatusPending,
		RunID:       ctx.orchestrator.runID,
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
		Input:       args,
		RetryPolicy: internalRetryPolicy,
		RetryState:  &RetryState{Attempts: 0},
	}

	entity.Data = workflowData

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		ParentEntityID: ctx.workflowID,
		ChildEntityID:  entity.ID,
		ParentStepID:   ctx.stepID,
		ChildStepID:    stepID,
		ParentType:     EntityTypeWorkflow,
		ChildType:      EntityTypeWorkflow,
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
		entityID:     entity.ID, // Store entity ID
	}

	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

func (ctx *WorkflowContext) Activity(stepID string, activityFunc interface{}, options ActivityOptions, args ...interface{}) *Future {
	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewFuture()

	// Check if result already exists in the database
	if result, ok := ctx.orchestrator.db.GetResult(ctx.workflowID, stepID); ok {
		log.Printf("Result found in database for workflowID: %d, stepID: %s", ctx.workflowID, stepID)
		future.setResult(result)
		return future
	}

	// Register activity on-the-fly
	handler, err := ctx.orchestrator.registerActivity(activityFunc)
	if err != nil {
		log.Printf("Error registering activity: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:      stepID,
		HandlerName: handler.HandlerName,
		Type:        EntityTypeActivity,
		Status:      StatusPending,
		RunID:       ctx.orchestrator.runID,
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
		Input:       args,
		RetryPolicy: internalRetryPolicy,
	}

	entity.Data = activityData

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		ParentEntityID: ctx.workflowID,
		ChildEntityID:  entity.ID,
		ParentStepID:   ctx.stepID,
		ChildStepID:    stepID,
		ParentType:     EntityTypeWorkflow,
		ChildType:      EntityTypeActivity,
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
		entityID:     entity.ID, // Store entity ID
	}

	go activityInstance.executeWithRetry()

	return future
}

func (ctx *WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}) *Future {
	log.Printf("WorkflowContext.SideEffect called with stepID: %s", stepID)
	future := NewFuture()

	// Check if result already exists in the database
	if result, ok := ctx.orchestrator.db.GetResult(ctx.workflowID, stepID); ok {
		log.Printf("Result found in database for workflowID: %d, stepID: %s", ctx.workflowID, stepID)
		future.setResult(result)
		return future
	}

	// SideEffect has a default retry policy with MaxAttempts=1
	retryPolicy := &RetryPolicy{
		MaxAttempts:     1,
		InitialInterval: 0,
	}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:      stepID,
		HandlerName: getFunctionName(sideEffectFunc),
		Type:        EntityTypeSideEffect,
		Status:      StatusPending,
		RunID:       ctx.orchestrator.runID,
	}

	// Create SideEffectData
	sideEffectData := &SideEffectData{
		Input: nil, // No input for side effects in this example
	}

	entity.Data = sideEffectData

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddEntity(entity)

	// Record hierarchy relationship using parent stepID
	hierarchy := &Hierarchy{
		ParentEntityID: ctx.workflowID,
		ChildEntityID:  entity.ID,
		ParentStepID:   ctx.stepID,
		ChildStepID:    stepID,
		ParentType:     EntityTypeWorkflow,
		ChildType:      EntityTypeSideEffect,
	}
	ctx.orchestrator.db.AddHierarchy(hierarchy)

	go func() {
		var attempt int
		for attempt = 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
			// Create Execution without ID
			execution := &Execution{
				EntityID:  entity.ID,
				Attempt:   attempt,
				Status:    StatusRunning,
				StartedAt: time.Now(),
			}
			// Add execution to database, which assigns the ID
			execution = ctx.orchestrator.db.AddExecution(execution)
			entity.Executions = append(entity.Executions, execution)
			executionID := execution.ID

			log.Printf("Executing side effect %s (Entity ID: %d, Execution ID: %d)", stepID, entity.ID, executionID)

			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic in side effect: %v", r)
					log.Printf("Panic in side effect: %v", err)
					execution.Status = StatusFailed
					completedAt := time.Now()
					execution.CompletedAt = &completedAt
					execution.Error = err
					entity.Status = StatusFailed
					ctx.orchestrator.db.UpdateExecution(execution)
					ctx.orchestrator.db.UpdateEntity(entity)
					future.setError(err)
					ctx.orchestrator.stopWithError(err)
				}
			}()

			argsValues := []reflect.Value{}
			results := reflect.ValueOf(sideEffectFunc).Call(argsValues)
			numOut := len(results)
			if numOut == 0 {
				err := fmt.Errorf("side effect should return at least a value")
				log.Printf("Error: %v", err)
				execution.Status = StatusFailed
				completedAt := time.Now()
				execution.CompletedAt = &completedAt
				execution.Error = err
				entity.Status = StatusFailed
				ctx.orchestrator.db.UpdateExecution(execution)
				ctx.orchestrator.db.UpdateEntity(entity)
				future.setError(err)
				ctx.orchestrator.stopWithError(err)
				return
			}

			result := results[0].Interface()
			log.Printf("Side effect returned result: %v", result)
			future.setResult(result)

			// Store the result in the database
			ctx.orchestrator.db.SetResult(entity.ID, result)

			execution.Status = StatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			entity.Status = StatusCompleted
			entity.Result = result
			ctx.orchestrator.db.UpdateExecution(execution)
			ctx.orchestrator.db.UpdateEntity(entity)
			return
		}
		// Max attempts reached
		entity.Status = StatusFailed
		ctx.orchestrator.db.UpdateEntity(entity)
		future.setError(fmt.Errorf("side effect failed after %d attempts", retryPolicy.MaxAttempts))
	}()

	return future
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
)

type EntityStatus string

const (
	StatusPending   EntityStatus = "Pending"
	StatusRunning   EntityStatus = "Running"
	StatusCompleted EntityStatus = "Completed"
	StatusFailed    EntityStatus = "Failed"
	StatusRetried   EntityStatus = "Retried"
)

// Entity represents the base entity for workflows, activities, and side effects.
type Entity struct {
	ID          int
	StepID      string
	HandlerName string
	Type        EntityType
	Status      EntityStatus
	RunID       int
	Result      interface{}
	Executions  []*Execution
	Data        interface{} // Could be WorkflowData, ActivityData, etc.
}

// Execution represents a single execution attempt of an entity.
type Execution struct {
	ID            int
	EntityID      int
	Attempt       int
	Status        EntityStatus
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         error
	ExecutionData interface{} // Could be WorkflowExecutionData, ActivityExecutionData, etc.
}

// Hierarchy tracks parent-child relationships between entities.
type Hierarchy struct {
	ParentEntityID int
	ChildEntityID  int
	ParentStepID   string
	ChildStepID    string
	ParentType     EntityType
	ChildType      EntityType
}

// Data structures matching the ent schemas
type WorkflowData struct {
	Duration    string               `json:"duration,omitempty"`
	Paused      bool                 `json:"paused"`
	Resumable   bool                 `json:"resumable"`
	RetryState  *RetryState          `json:"retry_state"`
	RetryPolicy *retryPolicyInternal `json:"retry_policy"`
	Input       []interface{}        `json:"input,omitempty"`
}

type ActivityData struct {
	Timeout      int64                `json:"timeout,omitempty"`
	MaxAttempts  int                  `json:"max_attempts"`
	ScheduledFor *time.Time           `json:"scheduled_for,omitempty"`
	RetryPolicy  *retryPolicyInternal `json:"retry_policy"`
	Input        []interface{}        `json:"input,omitempty"`
	Output       []interface{}        `json:"output,omitempty"`
}

type SideEffectData struct {
	Input  []interface{} `json:"input,omitempty"`
	Output []interface{} `json:"output,omitempty"`
}

// ExecutionData structures
type WorkflowExecutionData struct {
	Error  string        `json:"error,omitempty"`
	Output []interface{} `json:"output,omitempty"`
}

type ActivityExecutionData struct {
	Heartbeats       []interface{} `json:"heartbeats,omitempty"`
	LastHeartbeat    *time.Time    `json:"last_heartbeat,omitempty"`
	Progress         interface{}   `json:"progress,omitempty"`
	ExecutionDetails interface{}   `json:"execution_details,omitempty"`
}

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	result       interface{}
	err          error
	fsm          *stateless.StateMachine
	future       *Future
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	options      WorkflowOptions
	entity       *Entity
	entityID     int // Added entity ID for logging
	executionID  int // Added execution ID for logging
}

func (wi *WorkflowInstance) Start() {
	// Initialize the FSM
	wi.fsm = stateless.NewStateMachine(StateIdle)
	wi.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting)

	wi.fsm.Configure(StateExecuting).
		OnEntry(wi.executeWorkflow).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed)

	wi.fsm.Configure(StateCompleted).
		OnEntry(wi.onCompleted)

	wi.fsm.Configure(StateFailed).
		OnEntry(wi.onFailed)

	// Start the FSM
	wi.fsm.Fire(TriggerStart)
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
		// Create Execution without ID
		execution := &Execution{
			EntityID:  wi.entity.ID,
			Attempt:   attempt,
			Status:    StatusRunning,
			StartedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = wi.orchestrator.db.AddExecution(execution)
		wi.entity.Executions = append(wi.entity.Executions, execution)
		executionID := execution.ID
		wi.executionID = executionID // Store execution ID

		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.entityID, executionID)

		err := wi.runWorkflow(execution)
		if err == nil {
			// Success
			execution.Status = StatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			wi.entity.Status = StatusCompleted
			wi.entity.Result = wi.result
			wi.orchestrator.db.UpdateExecution(execution)
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = StatusFailed
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err
			wi.err = err
			wi.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {
				execution.Status = StatusRetried
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.entityID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				wi.entity.Status = StatusFailed
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
	if result, ok := wi.orchestrator.db.GetResult(wi.entity.ID, wi.stepID); ok {
		log.Printf("Result found in database for entity ID: %d", wi.entity.ID)
		wi.result = result
		return nil
	}

	// Register workflow on-the-fly
	handler, ok := wi.orchestrator.registry.workflows[wi.handler.HandlerName]
	if !ok {
		err := fmt.Errorf("error getting handler for workflow: %s", wi.handler.HandlerName)
		return err
	}

	f := handler.Handler

	ctxWorkflow := &WorkflowContext{
		orchestrator: wi.orchestrator,
		ctx:          wi.ctx,
		workflowID:   wi.entity.ID,
		stepID:       wi.stepID,
	}

	argsValues := []reflect.Value{reflect.ValueOf(ctxWorkflow)}
	for _, arg := range wi.input {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	log.Printf("Executing workflow: %s with args: %v", handler.HandlerName, wi.input)

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
		return wi.err
	} else {
		if numOut > 1 {
			result := results[0].Interface()
			log.Printf("Workflow returned result: %v", result)
			wi.result = result
		}
		// Store the result in the database
		wi.orchestrator.db.SetResult(wi.entity.ID, wi.result)
		return nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.entityID)
	if wi.future != nil {
		wi.future.setResult(wi.result)
	}
	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.entityID)
	if wi.future != nil {
		wi.future.setError(wi.err)
	}
	return nil
}

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	result       interface{}
	err          error
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	options      ActivityOptions
	entity       *Entity
	future       *Future
	entityID     int // Added entity ID for logging
	executionID  int // Added execution ID for logging
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
		// Create Execution without ID
		execution := &Execution{
			EntityID:  ai.entity.ID,
			Attempt:   attempt,
			Status:    StatusRunning,
			StartedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = ai.orchestrator.db.AddExecution(execution)
		ai.entity.Executions = append(ai.entity.Executions, execution)
		executionID := execution.ID
		ai.executionID = executionID // Store execution ID

		log.Printf("Executing activity %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.entityID, executionID)

		err := ai.runActivity(execution)
		if err == nil {
			// Success
			execution.Status = StatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			ai.entity.Status = StatusCompleted
			ai.entity.Result = ai.result
			ai.orchestrator.db.UpdateExecution(execution)
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.future.setResult(ai.result)
			return
		} else {
			execution.Status = StatusFailed
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err
			ai.err = err
			ai.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {
				execution.Status = StatusRetried
				log.Printf("Retrying activity %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", ai.stepID, ai.entityID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				ai.entity.Status = StatusFailed
				ai.orchestrator.db.UpdateEntity(ai.entity)
				ai.future.setError(err)
				return
			}
		}
	}
}

func (ai *ActivityInstance) runActivity(execution *Execution) error {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d) runActivity attempt %d", ai.stepID, ai.entityID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	if result, ok := ai.orchestrator.db.GetResult(ai.entity.ID, ai.stepID); ok {
		log.Printf("Result found in database for entity ID: %d", ai.entity.ID)
		ai.result = result
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

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range ai.input {
		argsValues = append(argsValues, reflect.ValueOf(arg))
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
		return ai.err
	} else {
		var result interface{}
		if numOut > 1 {
			result = results[0].Interface()
			log.Printf("Activity returned result: %v", result)
		}
		ai.result = result
		// Store the result in the database
		ai.orchestrator.db.SetResult(ai.entity.ID, result)
		return nil
	}
}

// Database interface defines methods for interacting with the data store.
type Database interface {
	// Entity methods
	AddEntity(entity *Entity) *Entity
	GetEntity(id int) *Entity
	UpdateEntity(entity *Entity)

	// Execution methods
	AddExecution(execution *Execution) *Execution
	GetExecution(id int) *Execution
	UpdateExecution(execution *Execution)

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy)
	GetHierarchy(parentID, childID int) *Hierarchy

	// Methods to store and retrieve results
	GetResult(entityID int, stepID string) (interface{}, bool)
	SetResult(entityID int, result interface{})
}

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	entities    map[int]*Entity
	executions  map[int]*Execution
	hierarchies []*Hierarchy
	results     map[int]interface{} // Map entityID to result
	mu          sync.Mutex
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{
		entities:   make(map[int]*Entity),
		executions: make(map[int]*Execution),
		results:    make(map[int]interface{}),
	}
}

var (
	entityIDCounter    int64
	executionIDCounter int64
)

func (db *DefaultDatabase) AddEntity(entity *Entity) *Entity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&entityIDCounter, 1))
	entity.ID = entityID
	db.entities[entity.ID] = entity
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

func (db *DefaultDatabase) GetResult(entityID int, stepID string) (interface{}, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	result, ok := db.results[entityID]
	return result, ok
}

func (db *DefaultDatabase) SetResult(entityID int, result interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.results[entityID] = result
}

// Orchestrator orchestrates the execution of workflows and activities.
type Orchestrator struct {
	db          Database
	rootWf      *WorkflowInstance
	ctx         context.Context
	instances   []*WorkflowInstance
	instancesMu sync.Mutex
	registry    *Registry
	err         error
	runID       int
}

func NewOrchestrator(db Database, ctx context.Context) *Orchestrator {
	log.Printf("NewOrchestrator called")
	o := &Orchestrator{
		db:       db,
		ctx:      ctx,
		registry: newRegistry(),
		runID:    generateRunID(),
	}
	return o
}

var runIDCounter int64

func generateRunID() int {
	return int(atomic.AddInt64(&runIDCounter, 1))
}

func (o *Orchestrator) Workflow(workflowFunc interface{}, options WorkflowOptions, args ...interface{}) error {
	// Register the workflow if not already registered
	handler, err := o.registerWorkflow(workflowFunc)
	if err != nil {
		return err
	}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:      "root",
		HandlerName: handler.HandlerName,
		Type:        EntityTypeWorkflow,
		Status:      StatusPending,
		RunID:       o.runID,
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
		Input:       args,
		RetryPolicy: internalRetryPolicy,
		RetryState:  &RetryState{Attempts: 0},
	}

	entity.Data = workflowData

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
	}

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return nil
}

func (o *Orchestrator) RegisterWorkflow(workflowFunc interface{}) error {
	_, err := o.registerWorkflow(workflowFunc)
	return err
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

	// Check if rootWf has any error
	if o.rootWf.err != nil {
		fmt.Printf("Root workflow failed with error: %v\n", o.rootWf.err)
	} else {
		fmt.Printf("Root workflow completed successfully with result: %v\n", o.rootWf.result)
	}
}

// Registry holds registered workflows and activities.
type Registry struct {
	workflows  map[string]HandlerInfo
	activities map[string]HandlerInfo
	mu         sync.Mutex
}

func newRegistry() *Registry {
	return &Registry{
		workflows:  make(map[string]HandlerInfo),
		activities: make(map[string]HandlerInfo),
	}
}

func (o *Orchestrator) registerWorkflow(workflowFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(workflowFunc)
	o.registry.mu.Lock()
	defer o.registry.mu.Unlock()

	// Check if already registered
	if handler, ok := o.registry.workflows[funcName]; ok {
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

	o.registry.workflows[funcName] = handler
	return handler, nil
}

func (o *Orchestrator) registerActivity(activityFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(activityFunc)
	o.registry.mu.Lock()
	defer o.registry.mu.Unlock()

	// Check if already registered
	if handler, ok := o.registry.activities[funcName]; ok {
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

	o.registry.activities[funcName] = handler
	return handler, nil
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Example Activity
func SomeActivity(ctx ActivityContext, data int) (int, error) {
	log.Printf("SomeActivity called with data: %d", data)
	select {
	case <-time.After(0):
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

func Workflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("Workflow called with data: %d", data)
	var value int

	select {
	case <-ctx.ctx.Done():
		log.Printf("Workflow context cancelled")
		return -1, ctx.ctx.Err()
	default:
	}

	var shouldDouble bool
	if err := ctx.SideEffect("side-effect-step", func() bool {
		log.Printf("Side effect called")
		return rand.Float32() < 0.5
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

	result := value + data
	if shouldDouble {
		result *= 2
	}
	log.Printf("Workflow returning result: %d", result)
	return result, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("main started")

	database := NewDefaultDatabase()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orchestrator := NewOrchestrator(database, ctx)
	orchestrator.RegisterWorkflow(Workflow)

	subWorkflowFailed.Store(true)

	err := orchestrator.Workflow(Workflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, 40)
	if err != nil {
		log.Fatalf("Error starting workflow: %v", err)
	}

	orchestrator.Wait()

	log.Printf("main finished")
}
