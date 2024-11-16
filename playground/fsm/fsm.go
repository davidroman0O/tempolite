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

// State and Trigger definitions
type state string

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

const (
	TriggerStart      trigger = "Start"
	TriggerComplete   trigger = "Complete"
	TriggerFail       trigger = "Fail"
	TriggerPause      trigger = "Pause"
	TriggerResume     trigger = "Resume"
	TriggerCompensate trigger = "Compensate"
)

// Base struct for common fields
type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RetryPolicy defines retry behavior configuration
type RetryPolicy struct {
	MaxAttempts        int
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaxInterval        time.Duration
}

// RetryState tracks the state of retries
type RetryState struct {
	Attempts int `json:"attempts"`
}

// retryPolicyInternal matches the database schema, using int64 for intervals
type retryPolicyInternal struct {
	MaxAttempts        int     `json:"max_attempts"`
	InitialInterval    int64   `json:"initial_interval"`
	BackoffCoefficient float64 `json:"backoff_coefficient"`
	MaxInterval        int64   `json:"max_interval"`
}

// Version tracks entity versions
type Version struct {
	ID       int
	EntityID int
	ChangeID string
	Version  int
	Data     map[string]interface{} // Additional data if needed
}

// Hierarchy tracks parent-child relationships between entities
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

// Data structs matching the ent schemas
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

type SagaData struct {
	Compensating     bool                 `json:"compensating"`
	CompensationData [][]byte             `json:"compensation_data,omitempty"`
	RetryState       *RetryState          `json:"retry_state"`
	RetryPolicy      *retryPolicyInternal `json:"retry_policy"`
}

type SideEffectData struct {
	Input       [][]byte             `json:"input,omitempty"`
	Output      [][]byte             `json:"output,omitempty"`
	RetryPolicy *retryPolicyInternal `json:"retry_policy"`
}

// ExecutionData structs matching the ent schemas
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

type SagaExecutionData struct {
	Error  string   `json:"error,omitempty"`
	Output [][]byte `json:"output,omitempty"`
}

type SideEffectExecutionData struct {
	EffectTime       *time.Time `json:"effect_time,omitempty"`
	EffectMetadata   []byte     `json:"effect_metadata,omitempty"`
	ExecutionContext []byte     `json:"execution_context,omitempty"`
	Result           []byte     `json:"result,omitempty"`
	Output           [][]byte   `json:"output,omitempty"`
}

// Run represents a workflow execution group
type Run struct {
	ID          int
	Status      string // "Pending", "Running", "Completed", etc.
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []interface{} // Can hold any entity type
	Hierarchies []*Hierarchy
}

// WorkflowEntity represents a workflow
type WorkflowEntity struct {
	ID           int
	HandlerName  string
	StepID       string
	Status       string // "Pending", "Running", etc.
	RunID        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Executions   []*WorkflowExecution
	WorkflowData *WorkflowData
	Paused       bool
	Resumable    bool
	HandlerInfo  *HandlerInfo
}

// ActivityEntity represents an activity
type ActivityEntity struct {
	ID           int
	HandlerName  string
	StepID       string
	Status       string // "Pending", "Running", etc.
	RunID        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Executions   []*ActivityExecution
	ActivityData *ActivityData
	Paused       bool
	Resumable    bool
	HandlerInfo  *HandlerInfo
}

// SagaEntity represents a saga
type SagaEntity struct {
	ID          int
	StepID      string
	Status      string // "Pending", "Running", etc.
	RunID       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Executions  []*SagaExecution
	SagaData    *SagaData
	Paused      bool
	Resumable   bool
	HandlerInfo *SagaHandlerInfo
}

// SideEffectEntity represents a side effect
type SideEffectEntity struct {
	ID                int
	HandlerName       string
	StepID            string
	Status            string // "Pending", "Running", etc.
	RunID             int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Executions        []*SideEffectExecution
	SideEffectData    *SideEffectData
	Paused            bool
	Resumable         bool
	HandlerInfo       *HandlerInfo
	ParentExecutionID int
	ParentEntityID    int
	ParentStepID      string
}

// WorkflowExecution represents an execution attempt of a workflow
type WorkflowExecution struct {
	ID                    int
	EntityID              int
	StartedAt             time.Time
	CompletedAt           *time.Time
	Status                string // "Pending", "Running", etc.
	Attempt               int
	Error                 string // Error message
	CreatedAt             time.Time
	UpdatedAt             time.Time
	WorkflowExecutionData *WorkflowExecutionData
}

// ActivityExecution represents an execution attempt of an activity
type ActivityExecution struct {
	ID                    int
	EntityID              int
	StartedAt             time.Time
	CompletedAt           *time.Time
	Status                string // "Pending", "Running", etc.
	Attempt               int
	Error                 string // Error message
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ActivityExecutionData *ActivityExecutionData
}

// SagaExecution represents an execution attempt of a saga
type SagaExecution struct {
	ID                int
	EntityID          int
	StartedAt         time.Time
	CompletedAt       *time.Time
	Status            string // "Pending", "Running", etc.
	Attempt           int
	Error             string // Error message
	CreatedAt         time.Time
	UpdatedAt         time.Time
	SagaExecutionData *SagaExecutionData
}

// SideEffectExecution represents an execution attempt of a side effect
type SideEffectExecution struct {
	ID                      int
	EntityID                int
	StartedAt               time.Time
	CompletedAt             *time.Time
	Status                  string // "Pending", "Running", etc.
	Attempt                 int
	Error                   string // Error message
	CreatedAt               time.Time
	UpdatedAt               time.Time
	SideEffectExecutionData *SideEffectExecutionData
}

// Serialization functions
func convertInputsForSerialization(executionInputs []interface{}) ([][]byte, error) {
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

func convertOutputsForSerialization(executionOutputs []interface{}) ([][]byte, error) {
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

func convertInputsFromSerialization(handlerInfo HandlerInfo, executionInputs [][]byte) ([]interface{}, error) {
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

func convertOutputsFromSerialization(handlerInfo HandlerInfo, executionOutputs [][]byte) ([]interface{}, error) {
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

func convertSingleOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
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

func (f *Future) WorkflowID() int {
	return f.workflowID
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
	executionID  int // Add this to keep track of the current execution ID
}

var ErrPaused = errors.New("execution paused")

func (ctx *WorkflowContext) checkPause() error {
	if ctx.orchestrator.IsPaused() {
		log.Printf("WorkflowContext detected orchestrator is paused")
		return ErrPaused
	}
	return nil
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

// Workflow creates a sub-workflow.
func (ctx *WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Workflow paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)
	future := NewFuture(0)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildWorkflowEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for workflow: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			return future
		}
		latestExecution := ctx.orchestrator.db.GetLatestWorkflowExecution(entity.ID)
		if latestExecution != nil && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Output != nil {
			// Deserialize output
			outputs, err := convertOutputsFromSerialization(*handlerInfo, latestExecution.WorkflowExecutionData.Output)
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
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal

	// Handle options and defaults
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		// Fill default values if zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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

	// Create WorkflowData
	workflowData := &WorkflowData{
		Duration:    "",
		Paused:      false,
		Resumable:   false,
		RetryState:  &RetryState{Attempts: 0},
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new WorkflowEntity without ID (database assigns it)
	entity = &WorkflowEntity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Status:       string(StatusPending),
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddWorkflowEntity(entity)
	future.workflowID = entity.ID

	// Prepare to create the sub-workflow instance
	subWorkflowInstance := &WorkflowInstance{
		stepID:            stepID,
		handler:           handler,
		input:             args,
		future:            future,
		ctx:               ctx.ctx,
		orchestrator:      ctx.orchestrator,
		workflowID:        entity.ID,
		options:           options,
		workflowEntity:    entity,
		entityID:          entity.ID,
		parentExecutionID: ctx.executionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
	}

	// Start the sub-workflow instance
	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

// Activity creates an activity.
func (ctx *WorkflowContext) Activity(stepID string, activityFunc interface{}, options *ActivityOptions, args ...interface{}) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Activity paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewFuture(0)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildActivityEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for activity: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			return future
		}
		latestExecution := ctx.orchestrator.db.GetLatestActivityExecution(entity.ID)
		if latestExecution != nil && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Output != nil {
			// Deserialize output
			outputs, err := convertOutputsFromSerialization(*handlerInfo, latestExecution.ActivityExecutionData.Output)
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
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal

	// Handle options and defaults
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		// Fill default values if zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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

	// Create ActivityData
	activityData := &ActivityData{
		Timeout:     0,
		MaxAttempts: 1,
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new ActivityEntity without ID (database assigns it)
	entity = &ActivityEntity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Status:       string(StatusPending),
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ActivityData: activityData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddActivityEntity(entity)
	future.workflowID = entity.ID

	// Prepare to create the activity instance
	activityInstance := &ActivityInstance{
		stepID:            stepID,
		handler:           handler,
		input:             args,
		future:            future,
		ctx:               ctx.ctx,
		orchestrator:      ctx.orchestrator,
		workflowID:        ctx.workflowID,
		options:           options,
		activityEntity:    entity,
		entityID:          entity.ID,
		parentExecutionID: ctx.executionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
	}

	// Start the activity instance
	ctx.orchestrator.addActivityInstance(activityInstance)
	activityInstance.Start()

	return future
}

// SideEffect executes a side effect function
func (ctx *WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}, options *WorkflowOptions) *Future {
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
	entity := ctx.orchestrator.db.GetChildSideEffectEntityByParentEntityIDAndStepID(ctx.workflowID, stepID)
	if entity != nil {
		latestExecution := ctx.orchestrator.db.GetLatestSideEffectExecution(entity.ID)
		if latestExecution != nil && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Output != nil {
			// Deserialize result using the returnTypes
			outputs, err := convertOutputsFromSerialization(HandlerInfo{ReturnTypes: returnTypes}, latestExecution.SideEffectExecutionData.Output)
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
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		// Fill defaults where zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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
		retryPolicy = rp
	} else {
		// Default retry policy with MaxAttempts=1
		retryPolicy = &RetryPolicy{
			MaxAttempts:        1,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		}
	}

	// Convert API RetryPolicy to internal retry policy
	internalRetryPolicy := &retryPolicyInternal{
		MaxAttempts:        retryPolicy.MaxAttempts,
		InitialInterval:    retryPolicy.InitialInterval.Nanoseconds(),
		BackoffCoefficient: retryPolicy.BackoffCoefficient,
		MaxInterval:        retryPolicy.MaxInterval.Nanoseconds(),
	}

	// Create SideEffectData
	sideEffectData := &SideEffectData{
		Input:       nil, // No input for side effects in this example
		RetryPolicy: internalRetryPolicy,
	}

	// Create a new SideEffectEntity without ID (database assigns it)
	entity = &SideEffectEntity{
		StepID:         stepID,
		HandlerName:    handlerName,
		Status:         string(StatusPending),
		RunID:          ctx.orchestrator.runID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		SideEffectData: sideEffectData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = ctx.orchestrator.db.AddSideEffectEntity(entity)
	future.workflowID = entity.ID

	// Prepare to create the side effect instance
	sideEffectInstance := &SideEffectInstance{
		stepID:            stepID,
		sideEffectFunc:    sideEffectFunc,
		future:            future,
		ctx:               ctx.ctx,
		orchestrator:      ctx.orchestrator,
		workflowID:        ctx.workflowID,
		entityID:          entity.ID,
		options:           options,
		returnTypes:       returnTypes,
		handlerName:       handlerName,
		handler:           handler,
		parentExecutionID: ctx.executionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
		sideEffectEntity:  entity,
	}

	// Start the side effect instance
	ctx.orchestrator.addSideEffectInstance(sideEffectInstance)
	sideEffectInstance.Start()

	return future
}

// ContinueAsNew allows a workflow to continue as new with the given function and arguments.
func (ctx *WorkflowContext) ContinueAsNew(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) error {
	return &ContinueAsNewError{
		WorkflowFunc: workflowFunc,
		Options:      options,
		Args:         args,
	}
}

// ContinueAsNewError indicates that the workflow should restart with new inputs.
type ContinueAsNewError struct {
	WorkflowFunc interface{}
	Options      *WorkflowOptions
	Args         []interface{}
}

func (e *ContinueAsNewError) Error() string {
	return "workflow is continuing as new"
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

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	stepID            string
	handler           HandlerInfo
	input             []interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            *Future
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	options           *WorkflowOptions
	workflowEntity    *WorkflowEntity
	entityID          int
	executionID       int
	execution         *WorkflowExecution // Current execution
	continueAsNew     *ContinueAsNewError
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
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

	if wi.options != nil && wi.options.RetryPolicy != nil {
		rp := wi.options.RetryPolicy
		// Fill default values where zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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
		maxAttempts = rp.MaxAttempts
		initialInterval = rp.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		if wi.orchestrator.IsPaused() {
			log.Printf("WorkflowInstance %s is paused", wi.stepID)
			wi.workflowEntity.Status = string(StatusPaused)
			wi.workflowEntity.Paused = true
			wi.orchestrator.db.UpdateWorkflowEntity(wi.workflowEntity)
			wi.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &WorkflowExecution{
			EntityID:  wi.workflowEntity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = wi.orchestrator.db.AddWorkflowExecution(execution)
		wi.workflowEntity.Executions = append(wi.workflowEntity.Executions, execution)
		executionID := execution.ID
		wi.executionID = executionID // Store execution ID
		wi.execution = execution

		// Now that we have executionID, we can create the hierarchy
		// But only if parentExecutionID is available (non-zero)
		if wi.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             wi.orchestrator.runID,
				ParentEntityID:    wi.parentEntityID,
				ChildEntityID:     wi.workflowEntity.ID,
				ParentExecutionID: wi.parentExecutionID,
				ChildExecutionID:  wi.executionID,
				ParentStepID:      wi.parentStepID,
				ChildStepID:       wi.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeWorkflow),
			}
			wi.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.workflowEntity.ID, executionID)

		err := wi.runWorkflow(execution)
		if errors.Is(err, ErrPaused) {
			wi.workflowEntity.Status = string(StatusPaused)
			wi.workflowEntity.Paused = true
			wi.orchestrator.db.UpdateWorkflowEntity(wi.workflowEntity)
			wi.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			wi.workflowEntity.Status = string(StatusCompleted)
			wi.orchestrator.db.UpdateWorkflowExecution(execution)
			wi.orchestrator.db.UpdateWorkflowEntity(wi.workflowEntity)
			wi.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			wi.err = err
			wi.orchestrator.db.UpdateWorkflowExecution(execution)
			if attempt < maxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.workflowEntity.ID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				wi.workflowEntity.Status = string(StatusFailed)
				wi.orchestrator.db.UpdateWorkflowEntity(wi.workflowEntity)
				wi.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (wi *WorkflowInstance) runWorkflow(execution *WorkflowExecution) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d, Execution ID: %d) runWorkflow attempt %d", wi.stepID, wi.workflowEntity.ID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := wi.orchestrator.db.GetLatestWorkflowExecution(wi.workflowEntity.ID)
	if latestExecution != nil && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", wi.workflowEntity.ID)
		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Output)
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
		workflowID:   wi.workflowEntity.ID,
		stepID:       wi.stepID,
		options:      wi.options,
		executionID:  wi.executionID,
	}

	// Convert inputs from serialization
	inputs, err := convertInputsFromSerialization(handler, wi.workflowEntity.WorkflowData.Input)
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
		if continueErr, ok := errInterface.(*ContinueAsNewError); ok {
			log.Printf("Workflow requested ContinueAsNew")
			wi.continueAsNew = continueErr
			// We treat it as normal completion, return nil to proceed
			return nil
		} else {
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
			wi.orchestrator.db.UpdateWorkflowExecution(wi.execution)

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

		wi.results = outputs

		// Serialize output
		outputBytes, err := convertOutputsForSerialization(wi.results)
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
		wi.orchestrator.db.UpdateWorkflowExecution(wi.execution)

		return nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.workflowEntity.ID)
	if wi.continueAsNew != nil {
		// Handle ContinueAsNew
		o := wi.orchestrator
		newHandler, err := o.registry.RegisterWorkflow(wi.continueAsNew.WorkflowFunc)
		if err != nil {
			log.Printf("Error registering workflow in ContinueAsNew: %v", err)
			wi.err = err
			if wi.future != nil {
				wi.future.setError(wi.err)
			}
			return nil
		}
		// Convert inputs to [][]byte
		inputBytes, err := convertInputsForSerialization(wi.continueAsNew.Args)
		if err != nil {
			log.Printf("Error converting inputs in ContinueAsNew: %v", err)
			wi.err = err
			if wi.future != nil {
				wi.future.setError(wi.err)
			}
			return nil
		}

		// Convert API RetryPolicy to internal retry policy
		var internalRetryPolicy *retryPolicyInternal

		if wi.continueAsNew.Options != nil && wi.continueAsNew.Options.RetryPolicy != nil {
			rp := wi.continueAsNew.Options.RetryPolicy
			// Fill default values if zero
			if rp.MaxAttempts == 0 {
				rp.MaxAttempts = 1
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

		// Create WorkflowData
		workflowData := &WorkflowData{
			Duration:    "",
			Paused:      false,
			Resumable:   false,
			RetryState:  &RetryState{Attempts: 0},
			RetryPolicy: internalRetryPolicy,
			Input:       inputBytes,
		}

		// Create a new WorkflowEntity without ID (database assigns it)
		newEntity := &WorkflowEntity{
			StepID:       wi.workflowEntity.StepID, // Use the same stepID
			HandlerName:  newHandler.HandlerName,
			Status:       string(StatusPending),
			RunID:        wi.workflowEntity.RunID, // Share the same RunID
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			WorkflowData: workflowData,
		}
		newEntity.HandlerInfo = &newHandler
		// Add the entity to the database, which assigns the ID
		newEntity = o.db.AddWorkflowEntity(newEntity)

		// Create a new WorkflowInstance
		newInstance := &WorkflowInstance{
			stepID:            wi.stepID,
			handler:           newHandler,
			input:             wi.continueAsNew.Args,
			ctx:               o.ctx,
			orchestrator:      o,
			workflowID:        newEntity.ID,
			options:           wi.continueAsNew.Options,
			workflowEntity:    newEntity,
			entityID:          newEntity.ID,
			parentExecutionID: wi.parentExecutionID,
			parentEntityID:    wi.parentEntityID,
			parentStepID:      wi.parentStepID,
		}
		if wi == o.rootWf {
			// Root workflow
			o.rootWf = newInstance
			// Complete the future immediately
			if wi.future != nil {
				// Set empty results as per instruction
				wi.future.setResult(wi.results)
			}
		} else {
			// Sub-workflow
			// Pass the future to the newInstance
			newInstance.future = wi.future
		}
		o.addWorkflowInstance(newInstance)
		newInstance.Start()
	} else {
		// Normal completion
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
	}
	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.workflowEntity.ID)
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
	log.Printf("WorkflowInstance %s (Entity ID: %d) onPaused called", wi.stepID, wi.workflowEntity.ID)
	if wi.future != nil {
		wi.future.setError(ErrPaused)
	}
	return nil
}

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	stepID            string
	handler           HandlerInfo
	input             []interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            *Future
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	options           *ActivityOptions
	activityEntity    *ActivityEntity
	entityID          int
	executionID       int
	execution         *ActivityExecution // Current execution
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
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

	if ai.options != nil && ai.options.RetryPolicy != nil {
		rp := ai.options.RetryPolicy
		// Fill default values if zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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
		maxAttempts = rp.MaxAttempts
		initialInterval = rp.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		if ai.orchestrator.IsPaused() {
			log.Printf("ActivityInstance %s is paused", ai.stepID)
			ai.activityEntity.Status = string(StatusPaused)
			ai.activityEntity.Paused = true
			ai.orchestrator.db.UpdateActivityEntity(ai.activityEntity)
			ai.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &ActivityExecution{
			EntityID:  ai.activityEntity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = ai.orchestrator.db.AddActivityExecution(execution)
		ai.activityEntity.Executions = append(ai.activityEntity.Executions, execution)
		executionID := execution.ID
		ai.executionID = executionID // Store execution ID
		ai.execution = execution

		// Now that we have executionID, we can create the hierarchy
		// But only if parentExecutionID is available (non-zero)
		if ai.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             ai.orchestrator.runID,
				ParentEntityID:    ai.parentEntityID,
				ChildEntityID:     ai.activityEntity.ID,
				ParentExecutionID: ai.parentExecutionID,
				ChildExecutionID:  ai.executionID,
				ParentStepID:      ai.parentStepID,
				ChildStepID:       ai.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeActivity),
			}
			ai.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing activity %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.activityEntity.ID, executionID)

		err := ai.runActivity(execution)
		if errors.Is(err, ErrPaused) {
			ai.activityEntity.Status = string(StatusPaused)
			ai.activityEntity.Paused = true
			ai.orchestrator.db.UpdateActivityEntity(ai.activityEntity)
			ai.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			ai.activityEntity.Status = string(StatusCompleted)
			ai.orchestrator.db.UpdateActivityExecution(execution)
			ai.orchestrator.db.UpdateActivityEntity(ai.activityEntity)
			ai.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			ai.err = err
			ai.orchestrator.db.UpdateActivityExecution(execution)
			if attempt < maxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying activity %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", ai.stepID, ai.activityEntity.ID, executionID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				ai.activityEntity.Status = string(StatusFailed)
				ai.orchestrator.db.UpdateActivityEntity(ai.activityEntity)
				ai.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (ai *ActivityInstance) runActivity(execution *ActivityExecution) error {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d) runActivity attempt %d", ai.stepID, ai.activityEntity.ID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := ai.orchestrator.db.GetLatestActivityExecution(ai.activityEntity.ID)
	if latestExecution != nil && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", ai.activityEntity.ID)
		outputs, err := convertOutputsFromSerialization(ai.handler, latestExecution.ActivityExecutionData.Output)
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
	inputs, err := convertInputsFromSerialization(handler, ai.activityEntity.ActivityData.Input)
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
		ai.orchestrator.db.UpdateActivityExecution(ai.execution)

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
		outputBytes, err := convertOutputsForSerialization(ai.results)
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
		ai.orchestrator.db.UpdateActivityExecution(ai.execution)

		return nil
	}
}

func (ai *ActivityInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onCompleted called", ai.stepID, ai.activityEntity.ID)
	if ai.future != nil {
		ai.future.setResult(ai.results)
	}
	return nil
}

func (ai *ActivityInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onFailed called", ai.stepID, ai.activityEntity.ID)
	if ai.future != nil {
		ai.future.setError(ai.err)
	}
	return nil
}

func (ai *ActivityInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onPaused called", ai.stepID, ai.activityEntity.ID)
	if ai.future != nil {
		ai.future.setError(ErrPaused)
	}
	return nil
}

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
	stepID            string
	sideEffectFunc    interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            *Future
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	entityID          int
	executionID       int
	options           *WorkflowOptions
	execution         *SideEffectExecution // Current execution
	returnTypes       []reflect.Type
	handlerName       string
	handler           HandlerInfo
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
	sideEffectEntity  *SideEffectEntity
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
	if sei.sideEffectEntity.SideEffectData != nil && sei.sideEffectEntity.SideEffectData.RetryPolicy != nil {
		rp := sei.sideEffectEntity.SideEffectData.RetryPolicy
		retryPolicy = &RetryPolicy{
			MaxAttempts:        rp.MaxAttempts,
			InitialInterval:    time.Duration(rp.InitialInterval),
			BackoffCoefficient: rp.BackoffCoefficient,
			MaxInterval:        time.Duration(rp.MaxInterval),
		}
	}

	if retryPolicy == nil {
		// Default retry policy with MaxAttempts=1
		retryPolicy = &RetryPolicy{
			MaxAttempts:        1,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		}
	}

	// Fill default values where zero
	if retryPolicy.MaxAttempts == 0 {
		retryPolicy.MaxAttempts = 1
	}
	if retryPolicy.InitialInterval == 0 {
		retryPolicy.InitialInterval = time.Second
	}
	if retryPolicy.BackoffCoefficient == 0 {
		retryPolicy.BackoffCoefficient = 2.0
	}
	if retryPolicy.MaxInterval == 0 {
		retryPolicy.MaxInterval = 5 * time.Minute
	}

	for attempt = 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
		if sei.orchestrator.IsPaused() {
			log.Printf("SideEffectInstance %s is paused", sei.stepID)
			sei.sideEffectEntity.Status = string(StatusPaused)
			sei.sideEffectEntity.Paused = true
			sei.orchestrator.db.UpdateSideEffectEntity(sei.sideEffectEntity)
			sei.fsm.Fire(TriggerPause)
			return
		}

		// Create Execution without ID
		execution := &SideEffectExecution{
			EntityID:  sei.sideEffectEntity.ID,
			Attempt:   attempt,
			Status:    string(StatusRunning),
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Add execution to database, which assigns the ID
		execution = sei.orchestrator.db.AddSideEffectExecution(execution)
		sei.sideEffectEntity.Executions = append(sei.sideEffectEntity.Executions, execution)
		executionID := execution.ID
		sei.executionID = executionID // Store execution ID
		sei.execution = execution

		// Now that we have executionID, we can create the hierarchy
		// But only if parentExecutionID is available (non-zero)
		if sei.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             sei.orchestrator.runID,
				ParentEntityID:    sei.parentEntityID,
				ChildEntityID:     sei.sideEffectEntity.ID,
				ParentExecutionID: sei.parentExecutionID,
				ChildExecutionID:  sei.executionID,
				ParentStepID:      sei.parentStepID,
				ChildStepID:       sei.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeSideEffect),
			}
			sei.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing side effect %s (Entity ID: %d, Execution ID: %d)", sei.stepID, sei.sideEffectEntity.ID, executionID)

		err := sei.runSideEffect(execution)
		if errors.Is(err, ErrPaused) {
			sei.sideEffectEntity.Status = string(StatusPaused)
			sei.sideEffectEntity.Paused = true
			sei.orchestrator.db.UpdateSideEffectEntity(sei.sideEffectEntity)
			sei.fsm.Fire(TriggerPause)
			return
		}
		if err == nil {
			// Success
			execution.Status = string(StatusCompleted)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			sei.sideEffectEntity.Status = string(StatusCompleted)
			sei.orchestrator.db.UpdateSideEffectExecution(execution)
			sei.orchestrator.db.UpdateSideEffectEntity(sei.sideEffectEntity)
			sei.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = string(StatusFailed)
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			sei.err = err
			sei.orchestrator.db.UpdateSideEffectExecution(execution)
			if attempt < retryPolicy.MaxAttempts {
				execution.Status = string(StatusRetried)
				log.Printf("Retrying side effect %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", sei.stepID, sei.sideEffectEntity.ID, executionID, attempt+1, retryPolicy.MaxAttempts, retryPolicy.InitialInterval)
				time.Sleep(retryPolicy.InitialInterval)
			} else {
				// Max attempts reached
				sei.sideEffectEntity.Status = string(StatusFailed)
				sei.orchestrator.db.UpdateSideEffectEntity(sei.sideEffectEntity)
				sei.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (sei *SideEffectInstance) runSideEffect(execution *SideEffectExecution) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d, Execution ID: %d) runSideEffect attempt %d", sei.stepID, sei.sideEffectEntity.ID, execution.ID, execution.Attempt)

	// Check if result already exists in the database
	latestExecution := sei.orchestrator.db.GetLatestSideEffectExecution(sei.sideEffectEntity.ID)
	if latestExecution != nil && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Output != nil {
		log.Printf("Result found in database for entity ID: %d", sei.sideEffectEntity.ID)
		outputs, err := convertOutputsFromSerialization(HandlerInfo{ReturnTypes: sei.returnTypes}, latestExecution.SideEffectExecutionData.Output)
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
	outputBytes, err := convertOutputsForSerialization(sei.results)
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
	sei.orchestrator.db.UpdateSideEffectExecution(sei.execution)

	return nil
}

func (sei *SideEffectInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onCompleted called", sei.stepID, sei.sideEffectEntity.ID)
	if sei.future != nil {
		sei.future.setResult(sei.results)
	}
	return nil
}

func (sei *SideEffectInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onFailed called", sei.stepID, sei.sideEffectEntity.ID)
	if sei.future != nil {
		sei.future.setError(sei.err)
	}
	return nil
}

func (sei *SideEffectInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onPaused called", sei.stepID, sei.sideEffectEntity.ID)
	if sei.future != nil {
		sei.future.setError(ErrPaused)
	}
	return nil
}

type TransactionContext struct {
	ctx context.Context
}

type CompensationContext struct {
	ctx context.Context
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

	// Create a new SagaEntity
	entity := &SagaEntity{
		StepID:    stepID,
		Status:    string(StatusPending),
		RunID:     o.runID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		SagaData:  &SagaData{},
	}
	entity.HandlerInfo = saga.HandlerInfo

	// Add the entity to the database
	entity = o.db.AddSagaEntity(entity)

	// Prepare to create the SagaInstance
	sagaInstance := &SagaInstance{
		saga:              saga,
		ctx:               ctx.ctx,
		orchestrator:      o,
		workflowID:        ctx.workflowID,
		stepID:            stepID,
		sagaInfo:          sagaInfo,
		entity:            entity,
		parentExecutionID: ctx.executionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
	}

	// Start the SagaInstance
	o.addSagaInstance(sagaInstance)
	sagaInstance.Start()

	return sagaInfo
}

type SagaInstance struct {
	saga              *SagaDefinition
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	stepID            string
	sagaInfo          *SagaInfo
	entity            *SagaEntity
	fsm               *stateless.StateMachine
	err               error
	currentStep       int
	compensations     []int // Indices of steps to compensate
	mu                sync.Mutex
	executionID       int
	execution         *SagaExecution
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

func (si *SagaInstance) Start() {
	// Initialize the FSM
	si.fsm = stateless.NewStateMachine(StateIdle)
	si.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting).
		Permit(TriggerPause, StatePaused)

	si.fsm.Configure(StateExecuting).
		OnEntry(si.executeSaga).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed).
		Permit(TriggerPause, StatePaused)

	si.fsm.Configure(StateCompleted).
		OnEntry(si.onCompleted)

	si.fsm.Configure(StateFailed).
		OnEntry(si.onFailed)

	si.fsm.Configure(StatePaused).
		OnEntry(si.onPaused).
		Permit(TriggerResume, StateExecuting)

	// Start the FSM
	go si.fsm.Fire(TriggerStart)
}

func (si *SagaInstance) executeSaga(_ context.Context, _ ...interface{}) error {
	// Create Execution without ID
	execution := &SagaExecution{
		EntityID:  si.entity.ID,
		Attempt:   1,
		Status:    string(StatusRunning),
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Add execution to database, which assigns the ID
	execution = si.orchestrator.db.AddSagaExecution(execution)
	si.entity.Executions = append(si.entity.Executions, execution)
	executionID := execution.ID
	si.executionID = executionID // Store execution ID
	si.execution = execution

	// Now that we have executionID, we can create the hierarchy
	// But only if parentExecutionID is available (non-zero)
	if si.parentExecutionID != 0 {
		hierarchy := &Hierarchy{
			RunID:             si.orchestrator.runID,
			ParentEntityID:    si.parentEntityID,
			ChildEntityID:     si.entity.ID,
			ParentExecutionID: si.parentExecutionID,
			ChildExecutionID:  si.executionID,
			ParentStepID:      si.parentStepID,
			ChildStepID:       si.stepID,
			ParentType:        string(EntityTypeWorkflow),
			ChildType:         string(EntityTypeSaga),
		}
		si.orchestrator.db.AddHierarchy(hierarchy)
	}

	// Execute the saga logic
	si.executeWithRetry()
	return nil
}

func (si *SagaInstance) executeWithRetry() {
	si.executeTransactions()
}

func (si *SagaInstance) executeTransactions() {
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
			si.fsm.Fire(TriggerFail)
			return
		}
		// Record the successful transaction
		si.compensations = append(si.compensations, si.currentStep)
		si.currentStep++
	}

	// All transactions succeeded
	si.fsm.Fire(TriggerComplete)
}

func (si *SagaInstance) executeCompensations() {
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
	si.fsm.Fire(TriggerFail)
}

func (si *SagaInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	// Update entity status to Completed
	si.entity.Status = string(StatusCompleted)
	si.orchestrator.db.UpdateSagaEntity(si.entity)

	// Update execution status to Completed
	si.execution.Status = string(StatusCompleted)
	completedAt := time.Now()
	si.execution.CompletedAt = &completedAt
	si.orchestrator.db.UpdateSagaExecution(si.execution)

	// Notify SagaInfo
	si.sagaInfo.err = nil
	close(si.sagaInfo.done)
	log.Println("Saga completed successfully")
	return nil
}

func (si *SagaInstance) onFailed(_ context.Context, _ ...interface{}) error {
	// Update entity status to Failed
	si.entity.Status = string(StatusFailed)
	si.orchestrator.db.UpdateSagaEntity(si.entity)

	// Update execution status to Failed
	si.execution.Status = string(StatusFailed)
	completedAt := time.Now()
	si.execution.CompletedAt = &completedAt
	si.orchestrator.db.UpdateSagaExecution(si.execution)

	// Set the error in SagaInfo
	if si.err != nil {
		si.sagaInfo.err = si.err
	} else {
		si.sagaInfo.err = errors.New("saga execution failed")
	}

	// Mark parent Workflow as Failed
	parentEntity := si.orchestrator.db.GetWorkflowEntity(si.workflowID)
	if parentEntity != nil {
		parentEntity.Status = string(StatusFailed)
		si.orchestrator.db.UpdateWorkflowEntity(parentEntity)
	}

	close(si.sagaInfo.done)
	log.Printf("Saga failed with error: %v", si.sagaInfo.err)
	return nil
}

func (si *SagaInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("SagaInstance %s (Entity ID: %d) onPaused called", si.stepID, si.entity.ID)
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

	// Version methods
	GetVersion(entityID int, changeID string) *Version
	SetVersion(version *Version) *Version

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy)
	GetHierarchy(parentID, childID int) *Hierarchy

	// WorkflowEntity methods
	AddWorkflowEntity(entity *WorkflowEntity) *WorkflowEntity
	GetWorkflowEntity(id int) *WorkflowEntity
	UpdateWorkflowEntity(entity *WorkflowEntity)
	GetWorkflowEntityByWorkflowIDAndStepID(workflowID int, stepID string) *WorkflowEntity
	GetChildWorkflowEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *WorkflowEntity

	// ActivityEntity methods
	AddActivityEntity(entity *ActivityEntity) *ActivityEntity
	GetActivityEntity(id int) *ActivityEntity
	UpdateActivityEntity(entity *ActivityEntity)
	GetActivityEntityByWorkflowIDAndStepID(workflowID int, stepID string) *ActivityEntity
	GetChildActivityEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *ActivityEntity

	// SideEffectEntity methods
	AddSideEffectEntity(entity *SideEffectEntity) *SideEffectEntity
	GetSideEffectEntity(id int) *SideEffectEntity
	UpdateSideEffectEntity(entity *SideEffectEntity)
	GetSideEffectEntityByWorkflowIDAndStepID(workflowID int, stepID string) *SideEffectEntity
	GetChildSideEffectEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *SideEffectEntity

	// SagaEntity methods
	AddSagaEntity(entity *SagaEntity) *SagaEntity
	GetSagaEntity(id int) *SagaEntity
	UpdateSagaEntity(entity *SagaEntity)
	GetSagaEntityByWorkflowIDAndStepID(workflowID int, stepID string) *SagaEntity
	GetChildSagaEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *SagaEntity

	// WorkflowExecution methods
	AddWorkflowExecution(execution *WorkflowExecution) *WorkflowExecution
	GetWorkflowExecution(id int) *WorkflowExecution
	UpdateWorkflowExecution(execution *WorkflowExecution)
	GetLatestWorkflowExecution(entityID int) *WorkflowExecution

	// ActivityExecution methods
	AddActivityExecution(execution *ActivityExecution) *ActivityExecution
	GetActivityExecution(id int) *ActivityExecution
	UpdateActivityExecution(execution *ActivityExecution)
	GetLatestActivityExecution(entityID int) *ActivityExecution

	// SideEffectExecution methods
	AddSideEffectExecution(execution *SideEffectExecution) *SideEffectExecution
	GetSideEffectExecution(id int) *SideEffectExecution
	UpdateSideEffectExecution(execution *SideEffectExecution)
	GetLatestSideEffectExecution(entityID int) *SideEffectExecution

	// SagaExecution methods
	AddSagaExecution(execution *SagaExecution) *SagaExecution
	GetSagaExecution(id int) *SagaExecution
	UpdateSagaExecution(execution *SagaExecution)
	GetLatestSagaExecution(entityID int) *SagaExecution
}

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs                 map[int]*Run
	versions             map[int]*Version
	hierarchies          []*Hierarchy
	workflowEntities     map[int]*WorkflowEntity
	activityEntities     map[int]*ActivityEntity
	sideEffectEntities   map[int]*SideEffectEntity
	sagaEntities         map[int]*SagaEntity
	workflowExecutions   map[int]*WorkflowExecution
	activityExecutions   map[int]*ActivityExecution
	sideEffectExecutions map[int]*SideEffectExecution
	sagaExecutions       map[int]*SagaExecution
	mu                   sync.Mutex
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{
		runs:                 make(map[int]*Run),
		versions:             make(map[int]*Version),
		workflowEntities:     make(map[int]*WorkflowEntity),
		activityEntities:     make(map[int]*ActivityEntity),
		sideEffectEntities:   make(map[int]*SideEffectEntity),
		sagaEntities:         make(map[int]*SagaEntity),
		workflowExecutions:   make(map[int]*WorkflowExecution),
		activityExecutions:   make(map[int]*ActivityExecution),
		sideEffectExecutions: make(map[int]*SideEffectExecution),
		sagaExecutions:       make(map[int]*SagaExecution),
	}
}

var (
	runIDCounter                 int64
	versionIDCounter             int64
	workflowEntityIDCounter      int64
	activityEntityIDCounter      int64
	sideEffectEntityIDCounter    int64
	sagaEntityIDCounter          int64
	workflowExecutionIDCounter   int64
	activityExecutionIDCounter   int64
	sideEffectExecutionIDCounter int64
	sagaExecutionIDCounter       int64
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

// WorkflowEntity methods
func (db *DefaultDatabase) AddWorkflowEntity(entity *WorkflowEntity) *WorkflowEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&workflowEntityIDCounter, 1))
	entity.ID = entityID
	db.workflowEntities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetWorkflowEntity(id int) *WorkflowEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.workflowEntities[id]
}

func (db *DefaultDatabase) UpdateWorkflowEntity(entity *WorkflowEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.workflowEntities[entity.ID] = entity
}

func (db *DefaultDatabase) GetChildWorkflowEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *WorkflowEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.workflowEntities[childEntityID]; exists {
				return entity
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) GetWorkflowEntityByWorkflowIDAndStepID(workflowID int, stepID string) *WorkflowEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	if entity, exists := db.workflowEntities[workflowID]; exists && entity.StepID == stepID {
		return entity
	}
	return nil
}

// ActivityEntity methods
func (db *DefaultDatabase) AddActivityEntity(entity *ActivityEntity) *ActivityEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&activityEntityIDCounter, 1))
	entity.ID = entityID
	db.activityEntities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetActivityEntity(id int) *ActivityEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.activityEntities[id]
}

func (db *DefaultDatabase) UpdateActivityEntity(entity *ActivityEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.activityEntities[entity.ID] = entity
}

func (db *DefaultDatabase) GetChildActivityEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *ActivityEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.activityEntities[childEntityID]; exists {
				return entity
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) GetActivityEntityByWorkflowIDAndStepID(workflowID int, stepID string) *ActivityEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	if entity, exists := db.activityEntities[workflowID]; exists && entity.StepID == stepID {
		return entity
	}
	return nil
}

// SideEffectEntity methods
func (db *DefaultDatabase) AddSideEffectEntity(entity *SideEffectEntity) *SideEffectEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&sideEffectEntityIDCounter, 1))
	entity.ID = entityID
	db.sideEffectEntities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetSideEffectEntity(id int) *SideEffectEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.sideEffectEntities[id]
}

func (db *DefaultDatabase) UpdateSideEffectEntity(entity *SideEffectEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectEntities[entity.ID] = entity
}

func (db *DefaultDatabase) GetChildSideEffectEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *SideEffectEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.sideEffectEntities[childEntityID]; exists {
				return entity
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) GetSideEffectEntityByWorkflowIDAndStepID(workflowID int, stepID string) *SideEffectEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	if entity, exists := db.sideEffectEntities[workflowID]; exists && entity.StepID == stepID {
		return entity
	}
	return nil
}

// SagaEntity methods
func (db *DefaultDatabase) AddSagaEntity(entity *SagaEntity) *SagaEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	entityID := int(atomic.AddInt64(&sagaEntityIDCounter, 1))
	entity.ID = entityID
	db.sagaEntities[entity.ID] = entity

	// Add the entity to its Run
	run := db.runs[entity.RunID]
	if run != nil {
		run.Entities = append(run.Entities, entity)
	}
	return entity
}

func (db *DefaultDatabase) GetSagaEntity(id int) *SagaEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.sagaEntities[id]
}

func (db *DefaultDatabase) UpdateSagaEntity(entity *SagaEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaEntities[entity.ID] = entity
}

func (db *DefaultDatabase) GetChildSagaEntityByParentEntityIDAndStepID(parentEntityID int, stepID string) *SagaEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.sagaEntities[childEntityID]; exists {
				return entity
			}
		}
	}
	return nil
}

func (db *DefaultDatabase) GetSagaEntityByWorkflowIDAndStepID(workflowID int, stepID string) *SagaEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	if entity, exists := db.sagaEntities[workflowID]; exists && entity.StepID == stepID {
		return entity
	}
	return nil
}

// WorkflowExecution methods
func (db *DefaultDatabase) AddWorkflowExecution(execution *WorkflowExecution) *WorkflowExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&workflowExecutionIDCounter, 1))
	execution.ID = executionID
	db.workflowExecutions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetWorkflowExecution(id int) *WorkflowExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.workflowExecutions[id]
}

func (db *DefaultDatabase) UpdateWorkflowExecution(execution *WorkflowExecution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.workflowExecutions[execution.ID] = execution
}

func (db *DefaultDatabase) GetLatestWorkflowExecution(entityID int) *WorkflowExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *WorkflowExecution
	for _, execution := range db.workflowExecutions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

// ActivityExecution methods
func (db *DefaultDatabase) AddActivityExecution(execution *ActivityExecution) *ActivityExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&activityExecutionIDCounter, 1))
	execution.ID = executionID
	db.activityExecutions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetActivityExecution(id int) *ActivityExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.activityExecutions[id]
}

func (db *DefaultDatabase) UpdateActivityExecution(execution *ActivityExecution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.activityExecutions[execution.ID] = execution
}

func (db *DefaultDatabase) GetLatestActivityExecution(entityID int) *ActivityExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *ActivityExecution
	for _, execution := range db.activityExecutions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

// SideEffectExecution methods
func (db *DefaultDatabase) AddSideEffectExecution(execution *SideEffectExecution) *SideEffectExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&sideEffectExecutionIDCounter, 1))
	execution.ID = executionID
	db.sideEffectExecutions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetSideEffectExecution(id int) *SideEffectExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.sideEffectExecutions[id]
}

func (db *DefaultDatabase) UpdateSideEffectExecution(execution *SideEffectExecution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffectExecutions[execution.ID] = execution
}

func (db *DefaultDatabase) GetLatestSideEffectExecution(entityID int) *SideEffectExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *SideEffectExecution
	for _, execution := range db.sideEffectExecutions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

// SagaExecution methods
func (db *DefaultDatabase) AddSagaExecution(execution *SagaExecution) *SagaExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	executionID := int(atomic.AddInt64(&sagaExecutionIDCounter, 1))
	execution.ID = executionID
	db.sagaExecutions[execution.ID] = execution
	return execution
}

func (db *DefaultDatabase) GetSagaExecution(id int) *SagaExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.sagaExecutions[id]
}

func (db *DefaultDatabase) UpdateSagaExecution(execution *SagaExecution) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.sagaExecutions[execution.ID] = execution
}

func (db *DefaultDatabase) GetLatestSagaExecution(entityID int) *SagaExecution {
	db.mu.Lock()
	defer db.mu.Unlock()
	var latestExecution *SagaExecution
	for _, execution := range db.sagaExecutions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	return latestExecution
}

// Clear removes all Runs that are 'Completed' and their associated data.
func (db *DefaultDatabase) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	runsToDelete := []int{}
	entitiesToDelete := map[int]string{} // EntityID to EntityType
	executionsToDelete := map[int]string{}
	hierarchiesToKeep := []*Hierarchy{}
	versionsToDelete := map[int]*Version{}

	// Find Runs to delete
	for runID, run := range db.runs {
		if run.Status == string(StatusCompleted) {
			runsToDelete = append(runsToDelete, runID)
			// Collect Entities associated with the Run
			for _, entity := range run.Entities {
				switch e := entity.(type) {
				case *WorkflowEntity:
					entitiesToDelete[e.ID] = string(EntityTypeWorkflow)
				case *ActivityEntity:
					entitiesToDelete[e.ID] = string(EntityTypeActivity)
				case *SideEffectEntity:
					entitiesToDelete[e.ID] = string(EntityTypeSideEffect)
				case *SagaEntity:
					entitiesToDelete[e.ID] = string(EntityTypeSaga)
				}
			}
		}
	}

	// Collect Executions associated with Entities to delete
	for execID, execution := range db.workflowExecutions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = string(EntityTypeWorkflow)
		}
	}
	for execID, execution := range db.activityExecutions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = string(EntityTypeActivity)
		}
	}
	for execID, execution := range db.sideEffectExecutions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = string(EntityTypeSideEffect)
		}
	}
	for execID, execution := range db.sagaExecutions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = string(EntityTypeSaga)
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

	// Delete Entities and Executions
	for entityID, entityType := range entitiesToDelete {
		switch EntityType(entityType) {
		case EntityTypeWorkflow:
			delete(db.workflowEntities, entityID)
		case EntityTypeActivity:
			delete(db.activityEntities, entityID)
		case EntityTypeSideEffect:
			delete(db.sideEffectEntities, entityID)
		case EntityTypeSaga:
			delete(db.sagaEntities, entityID)
		}
	}
	for execID, entityType := range executionsToDelete {
		switch EntityType(entityType) {
		case EntityTypeWorkflow:
			delete(db.workflowExecutions, execID)
		case EntityTypeActivity:
			delete(db.activityExecutions, execID)
		case EntityTypeSideEffect:
			delete(db.sideEffectExecutions, execID)
		case EntityTypeSaga:
			delete(db.sagaExecutions, execID)
		}
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

func (o *Orchestrator) Workflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *Future {
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
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		return NewFuture(0)
	}

	// Convert API RetryPolicy to internal retry policy
	var internalRetryPolicy *retryPolicyInternal

	// Handle options and defaults
	if options != nil && options.RetryPolicy != nil {
		rp := options.RetryPolicy
		// Fill default values if zero
		if rp.MaxAttempts == 0 {
			rp.MaxAttempts = 1
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

	// Create WorkflowData
	workflowData := &WorkflowData{
		Duration:    "",
		Paused:      false,
		Resumable:   false,
		RetryState:  &RetryState{Attempts: 0},
		RetryPolicy: internalRetryPolicy,
		Input:       inputBytes,
	}

	// Create a new WorkflowEntity without ID (database assigns it)
	entity := &WorkflowEntity{
		StepID:       "root",
		HandlerName:  handler.HandlerName,
		Status:       string(StatusPending),
		RunID:        o.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
	}
	entity.HandlerInfo = &handler

	// Add the entity to the database, which assigns the ID
	entity = o.db.AddWorkflowEntity(entity)

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:            "root",
		handler:           handler,
		input:             args,
		ctx:               o.ctx,
		orchestrator:      o,
		workflowID:        entity.ID,
		options:           options,
		workflowEntity:    entity,
		entityID:          entity.ID, // Store entity ID
		parentExecutionID: 0,         // Root has no parent execution
		parentEntityID:    0,         // Root has no parent entity
		parentStepID:      "",        // Root has no parent step
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
	entity := o.db.GetWorkflowEntity(entityID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", entityID)
		return NewFuture(0)
	}

	// Set the runID from the entity
	o.runID = entity.RunID

	// Update the entity's paused state
	entity.Paused = false
	entity.Resumable = true
	o.db.UpdateWorkflowEntity(entity)

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
	inputs, err := convertInputsFromSerialization(handler, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return NewFuture(0)
	}

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:         entity.StepID,
		handler:        handler,
		input:          inputs,
		ctx:            o.ctx,
		orchestrator:   o,
		workflowID:     entity.ID,
		workflowEntity: entity,
		options: &WorkflowOptions{
			RetryPolicy: &RetryPolicy{
				MaxAttempts:        int(entity.WorkflowData.RetryPolicy.MaxAttempts),
				InitialInterval:    time.Duration(entity.WorkflowData.RetryPolicy.InitialInterval),
				BackoffCoefficient: entity.WorkflowData.RetryPolicy.BackoffCoefficient,
				MaxInterval:        time.Duration(entity.WorkflowData.RetryPolicy.MaxInterval),
			},
			VersionOverrides: nil, // Adjust if necessary
		},
		entityID:          entity.ID,
		parentExecutionID: 0, // Since we're resuming the root workflow
		parentEntityID:    0,
		parentStepID:      "",
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

// Retry retries a failed root workflow by creating a new entity and execution.
func (o *Orchestrator) Retry(workflowID int) *Future {
	// Retrieve the workflow entity
	entity := o.db.GetWorkflowEntity(workflowID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", workflowID)
		return NewFuture(0)
	}

	// Copy inputs
	inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return NewFuture(0)
	}

	// Create a new WorkflowData
	workflowData := &WorkflowData{
		Duration:    "",
		Paused:      false,
		Resumable:   false,
		RetryState:  &RetryState{Attempts: 0},
		RetryPolicy: entity.WorkflowData.RetryPolicy,
		Input:       entity.WorkflowData.Input,
	}

	// Create a new WorkflowEntity
	newEntity := &WorkflowEntity{
		StepID:       entity.StepID,
		HandlerName:  entity.HandlerName,
		Status:       string(StatusPending),
		RunID:        o.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
		HandlerInfo:  entity.HandlerInfo,
	}
	newEntity = o.db.AddWorkflowEntity(newEntity)

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:            newEntity.StepID,
		handler:           *entity.HandlerInfo,
		input:             inputs,
		ctx:               o.ctx,
		orchestrator:      o,
		workflowID:        newEntity.ID,
		options:           nil, // You might want to use the same options or set new ones
		workflowEntity:    newEntity,
		entityID:          newEntity.ID,
		parentExecutionID: 0,
		parentEntityID:    0,
		parentStepID:      "",
	}

	future := NewFuture(newEntity.ID)
	instance.future = future

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return future
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
	if err := ctx.Activity("activity-step", SomeActivity, &ActivityOptions{
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
	if err := ctx.Activity("activity-step", SomeActivity, &ActivityOptions{
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
	if err := ctx.Workflow("subsubsubworkflow-step", SubSubSubWorkflow, &WorkflowOptions{
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
	if err := ctx.Activity("activity-step", SomeActivity, &ActivityOptions{
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

	if err := ctx.Workflow("subsubworkflow-step", SubSubWorkflow, &WorkflowOptions{
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

var continueAsNewCalled atomic.Bool

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
	}, &WorkflowOptions{
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

	if !continueAsNewCalled.Load() {
		continueAsNewCalled.Store(true)
		log.Printf("Using ContinueAsNew to restart workflow")
		err := ctx.ContinueAsNew(Workflow, nil, data*2)
		if err != nil {
			return -1, err
		}
		return 0, nil // This line won't be executed
	}

	if err := ctx.Workflow("subworkflow-step", SubWorkflow, &WorkflowOptions{
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

	var a int
	var b int
	if err := ctx.Workflow("a-b", SubManyWorkflow, nil).Get(&a, &b); err != nil {
		log.Printf("Workflow encountered error from sub-many-workflow: %v", err)
		return -1, err
	}

	log.Printf("Workflow returning result: %d - %v %v", result, a, b)
	return result, nil
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
	continueAsNewCalled.Store(false)

	future := orchestrator.Workflow(Workflow, &WorkflowOptions{
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

	future = newOrchestrator.Resume(future.WorkflowID())

	var result int
	if err := future.Get(&result); err != nil {
		log.Printf("Workflow failed with error: %v", err)
	} else {
		log.Printf("Workflow completed with result: %d", result)
	}

	newOrchestrator.Wait()

	log.Printf("Retrying workflow")
	retryFuture := newOrchestrator.Retry(future.WorkflowID())

	if err := retryFuture.Get(&result); err != nil {
		log.Printf("Retried workflow failed with error: %v", err)
	} else {
		log.Printf("Retried workflow completed with result: %d", result)
	}

	// After execution, clear completed Runs to free up memory
	database.Clear()

	log.Printf("main finished")
}
