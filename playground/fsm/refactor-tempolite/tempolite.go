package tempolite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"math/rand"

	"github.com/qmuntal/stateless"
	"github.com/sasha-s/go-deadlock"
	"github.com/stephenfire/go-rtl"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/retrypool/logs"
	"golang.org/x/sync/errgroup"
)

func init() {
	maxprocs.Set()

	deadlock.Opts.DeadlockTimeout = time.Second * 2 // Time to wait before reporting a potential deadlock
	deadlock.Opts.OnPotentialDeadlock = func() {
		// You can customize the behavior when a potential deadlock is detected
		log.Println("POTENTIAL DEADLOCK DETECTED!")
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
		log.Printf("Goroutine stack dump:\n%s", buf)
	}
}

/// FILE: ./activities.go

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

// ActivityInstance represents an instance of an activity execution.
type ActivityInstance struct {
	mu                deadlock.Mutex
	stepID            string
	handler           HandlerInfo
	input             []interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            *RuntimeFuture
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	options           *ActivityOptions
	entity            *Entity
	entityID          int
	executionID       int
	execution         *Execution // Current execution
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

func (ai *ActivityInstance) Start() {
	ai.mu.Lock()
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
	ai.mu.Unlock()

	// Start the FSM
	go func() {
		ai.fsm.Fire(TriggerStart)
	}()
}

func (ai *ActivityInstance) executeActivity(_ context.Context, _ ...interface{}) error {
	return ai.executeWithRetry()
}

func (ai *ActivityInstance) executeWithRetry() error {
	var err error

	var attempt int
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration

	if ai.entity.RetryPolicy != nil {
		rp := ai.entity.RetryPolicy
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

	for attempt = ai.entity.RetryState.Attempts + 1; attempt <= maxAttempts; attempt++ {
		if ai.orchestrator.IsPaused() {
			log.Printf("ActivityInstance %s is paused", ai.stepID)
			ai.entity.Status = StatusPaused
			ai.entity.Paused = true
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.mu.Lock()
			ai.fsm.Fire(TriggerPause)
			ai.mu.Unlock()
			return nil
		}

		// Update RetryState
		ai.entity.RetryState.Attempts = attempt
		ai.orchestrator.db.UpdateEntity(ai.entity)

		// Create Execution without ID
		execution := &Execution{
			EntityID:  ai.entity.ID,
			Status:    ExecutionStatusRunning,
			Attempt:   attempt,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entity:    ai.entity,
		}
		// Add execution to database, which assigns the ID
		if err = ai.orchestrator.db.AddExecution(execution); err != nil {
			log.Printf("Error adding execution: %v", err)
			ai.entity.Status = StatusFailed
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.mu.Lock()
			ai.fsm.Fire(TriggerFail)
			ai.mu.Unlock()
			return nil
		}
		ai.entity.Executions = append(ai.entity.Executions, execution)
		executionID := execution.ID
		ai.executionID = executionID // Store execution ID
		ai.execution = execution

		// Now that we have executionID, we can create the hierarchy
		// But only if parentExecutionID is available (non-zero)
		if ai.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             ai.orchestrator.runID,
				ParentEntityID:    ai.parentEntityID,
				ChildEntityID:     ai.entity.ID,
				ParentExecutionID: ai.parentExecutionID,
				ChildExecutionID:  ai.executionID,
				ParentStepID:      ai.parentStepID,
				ChildStepID:       ai.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeActivity),
			}
			ai.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing activity %s (Entity ID: %d, Execution ID: %d)", ai.stepID, ai.entity.ID, executionID)

		err := ai.runActivity(execution)
		if errors.Is(err, ErrPaused) {
			ai.entity.Status = StatusPaused
			ai.entity.Paused = true
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.mu.Lock()
			ai.fsm.Fire(TriggerPause)
			ai.mu.Unlock()
			return nil
		}
		if err == nil {
			// Success
			execution.Status = ExecutionStatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			ai.entity.Status = StatusCompleted
			ai.orchestrator.db.UpdateExecution(execution)
			ai.orchestrator.db.UpdateEntity(ai.entity)
			ai.mu.Lock()
			ai.fsm.Fire(TriggerComplete)
			ai.mu.Unlock()
			return nil
		} else {
			execution.Status = ExecutionStatusFailed
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			ai.err = err
			ai.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {

				// Calculate next interval
				nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(attempt-1)))
				if nextInterval > maxInterval {
					nextInterval = maxInterval
				}
				log.Printf("Retrying activity %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", ai.stepID, ai.entity.ID, executionID, attempt+1, maxAttempts, nextInterval)
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached
				ai.entity.Status = StatusFailed
				ai.orchestrator.db.UpdateEntity(ai.entity)
				ai.mu.Lock()
				ai.fsm.Fire(TriggerFail)
				ai.mu.Unlock()
				return nil
			}
		}
	}
	return nil
}

// Sub-function within executeWithRetry

func (ai *ActivityInstance) runActivity(execution *Execution) error {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d) runActivity attempt %d", ai.stepID, ai.entity.ID, execution.ID, execution.Attempt)

	// You only need to lock when accessing shared data
	var err error
	var latestExecution *Execution
	latestExecution, err = ai.orchestrator.db.GetLatestExecution(ai.entity.ID)
	if err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err
	}

	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted &&
		latestExecution.ActivityExecutionData != nil &&
		latestExecution.ActivityExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", ai.entity.ID)
		outputs, err := convertOutputsFromSerialization(ai.handler, latestExecution.ActivityExecutionData.Outputs)
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
	inputs, err := convertInputsFromSerialization(handler, ai.entity.ActivityData.Input)
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
		// Update execution error
		execution.Error = errorMessage

		// Create ActivityExecutionData
		activityExecutionData := &ActivityExecutionData{
			Outputs: nil,
		}

		// Update the execution with the execution data
		execution.ActivityExecutionData = activityExecutionData
		ai.orchestrator.db.UpdateExecution(execution)

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
			Outputs: outputBytes,
		}

		// Update the execution with the execution data
		execution.ActivityExecutionData = activityExecutionData
		ai.orchestrator.db.UpdateExecution(execution)

		return nil
	}
}

func (ai *ActivityInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onCompleted called", ai.stepID, ai.entity.ID)
	if ai.future != nil {
		ai.mu.Lock()
		results := ai.results
		ai.mu.Unlock()
		ai.future.setResult(results)
	}
	return nil
}

func (ai *ActivityInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onFailed called", ai.stepID, ai.entity.ID)
	if ai.future != nil {
		ai.mu.Lock()
		err := ai.err
		ai.mu.Unlock()
		ai.future.setError(err)
	}
	return nil
}

func (ai *ActivityInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onPaused called", ai.stepID, ai.entity.ID)
	if ai.future != nil {
		ai.future.setError(ErrPaused)
	}
	return nil
}

/// FILE: ./context_workflows.go

// Data structs matching the ent schemas
type WorkflowData struct {
	Duration  string   `json:"duration,omitempty"`
	Paused    bool     `json:"paused"`
	Resumable bool     `json:"resumable"`
	Input     [][]byte `json:"input,omitempty"`
	Attempt   int      `json:"attempt"`
}

// ExecutionData structs matching the ent schemas
type WorkflowExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

// WorkflowOptions provides options for workflows, including retry policies and version overrides.
type WorkflowOptions struct {
	Queue            string
	RetryPolicy      *RetryPolicy
	VersionOverrides map[string]int // Map of changeID to forced version
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

func (ac WorkflowContext) Done() <-chan struct{} {
	return ac.ctx.Done()
}

func (ac WorkflowContext) Err() error {
	return ac.ctx.Err()
}

func (ctx WorkflowContext) checkPause() error {
	if ctx.orchestrator.IsPaused() {
		log.Printf("WorkflowContext detected orchestrator is paused")
		return ErrPaused
	}
	return nil
}

// GetVersion retrieves or sets a version for a changeID.
func (ctx WorkflowContext) GetVersion(changeID string, minSupported, maxSupported int) (int, error) {

	var err error

	// First check if version is overridden in options
	if ctx.options != nil && ctx.options.VersionOverrides != nil {
		if forcedVersion, ok := ctx.options.VersionOverrides[changeID]; ok {
			return forcedVersion, nil
		}
	}

	var version *Version
	if version, err = ctx.orchestrator.db.GetVersion(ctx.workflowID, changeID); err != nil {

		// If version not found, create a new version.
		newVersion := &Version{
			EntityID: ctx.workflowID,
			ChangeID: changeID,
			Version:  maxSupported,
		}

		if err := ctx.orchestrator.db.SetVersion(newVersion); err != nil {
			return 0, err
		}

		return newVersion.Version, nil
	}

	return version.Version, nil

}

// Workflow creates a sub-workflow.
func (ctx WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
	var err error

	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Workflow paused at stepID: %s", stepID)
		future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
		future.setEntityID(ctx.workflowID)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)

	// Check if result already exists in the database
	var entity *Entity
	if entity, err = ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepIDAndType(ctx.workflowID, stepID, EntityTypeWorkflow); err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			log.Printf("Error getting entity: %v", err)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(ctx.workflowID)
			future.setError(err)
			return future
		}
	}

	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for workflow: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(entity.ID)
			future.setError(err)
			return future
		}

		var latestExecution *Execution
		if latestExecution, err = ctx.orchestrator.db.GetLatestExecution(entity.ID); err != nil {
			log.Printf("Error getting latest execution: %v", err)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(entity.ID)
			future.setError(err)
			return future
		}

		if latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
			// Deserialize output
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(entity.ID)
			outputs, err := convertOutputsFromSerialization(*handlerInfo, latestExecution.WorkflowExecutionData.Outputs)
			if err != nil {
				log.Printf("Error deserializing outputs: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}

		if latestExecution.Status == ExecutionStatusFailed && latestExecution.Error != "" {
			log.Printf("Workflow %s has failed execution with error: %s", stepID, latestExecution.Error)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(entity.ID)
			future.setError(errors.New(latestExecution.Error))
			return future
		}
	}

	// Register workflow on-the-fly
	handler, err := ctx.orchestrator.registry.RegisterWorkflow(workflowFunc)
	if err != nil {
		log.Printf("Error registering workflow: %v", err)
		future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
		future.setEntityID(ctx.workflowID)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	// Convert inputs to [][]byte
	inputBytes, err := convertInputsForSerialization(args)
	if err != nil {
		log.Printf("Error converting inputs: %v", err)
		future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
		future.setEntityID(ctx.workflowID)
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
		Duration:  "",
		Paused:    false,
		Resumable: false,
		Input:     inputBytes,
		Attempt:   1,
	}

	// Create RetryState
	retryState := &RetryState{Attempts: 0}

	// Check for queue routing
	if options != nil && options.Queue != "" {

		// Find or get queue by name
		var queue *Queue
		if queue, err = ctx.orchestrator.db.GetQueueByName(options.Queue); err != nil {
			log.Printf("Error getting queue: %v", err)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setEntityID(ctx.workflowID)
			future.setError(err)
			return future
		}

		// Create a new Entity with queue assignment
		entity = &Entity{
			StepID:       stepID,
			HandlerName:  handler.HandlerName,
			Status:       StatusPending,
			Type:         EntityTypeWorkflow,
			RunID:        ctx.orchestrator.runID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			WorkflowData: workflowData,
			RetryPolicy:  internalRetryPolicy,
			RetryState:   retryState,
			HandlerInfo:  &handler,
			Paused:       false,
			Resumable:    false,
			QueueID:      queue.ID,
			Queue:        queue,
		}

		// Add the entity to the database
		if err = ctx.orchestrator.db.AddEntity(entity); err != nil {
			log.Printf("Error adding entity: %v", err)
			future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
			future.setError(err)
			future.setEntityID(entity.ID)
			return future
		}

		future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
		future.setEntityID(entity.ID) // track that one
		return future
	}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Status:       StatusPending,
		Type:         EntityTypeWorkflow,
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
		RetryPolicy:  internalRetryPolicy,
		RetryState:   retryState,
		HandlerInfo:  &handler,
		Paused:       false,
		Resumable:    false,
	}

	// Add the entity to the database, which assigns the ID
	if err = ctx.orchestrator.db.AddEntity(entity); err != nil {
		log.Printf("Error adding entity: %v", err)
		future := NewRuntimeFuture()
		future.setError(err)
		future.setEntityID(entity.ID)
		return future
	}

	future := NewRuntimeFuture()
	future.setEntityID(entity.ID) // track that one

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
		entity:            entity,
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

// ContinueAsNew allows a workflow to continue as new with the given function and arguments.
func (ctx WorkflowContext) ContinueAsNew(options *WorkflowOptions, args ...interface{}) error {
	return &ContinueAsNewError{
		Options: options,
		Args:    args,
	}
}

/// FILE: ./context_workflows_activities.go

// Activity creates an activity.
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

	// Check if result already exists in the database
	var entity *Entity
	if entity, err = ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepIDAndType(ctx.workflowID, stepID, EntityTypeActivity); err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			log.Printf("Error getting entity: %v", err)
			future.setError(err)
			return future
		}
	}

	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for activity: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setEntityID(entity.ID)
			future.setError(err)
			return future
		}

		var latestExecution *Execution
		if latestExecution, err = ctx.orchestrator.db.GetLatestExecution(entity.ID); err != nil {
			log.Printf("Error getting latest execution: %v", err)
			future.setEntityID(entity.ID)
			future.setError(err)
			return future
		}

		if latestExecution.Status == ExecutionStatusCompleted && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Outputs != nil {
			// Deserialize output
			outputs, err := convertOutputsFromSerialization(*handlerInfo, latestExecution.ActivityExecutionData.Outputs)
			if err != nil {
				log.Printf("Error deserializing outputs: %v", err)
				future.setEntityID(entity.ID)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}

		if latestExecution.Status == ExecutionStatusFailed && latestExecution.Error != "" {
			log.Printf("Activity %s has failed execution with error: %s", stepID, latestExecution.Error)
			future.setEntityID(entity.ID)
			future.setError(errors.New(latestExecution.Error))
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
		MaxAttempts: internalRetryPolicy.MaxAttempts,
		Input:       inputBytes,
		Attempt:     1,
	}

	// Create RetryState
	retryState := &RetryState{Attempts: 0}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:       stepID,
		HandlerName:  handler.HandlerName,
		Status:       StatusPending,
		Type:         EntityTypeActivity,
		RunID:        ctx.orchestrator.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ActivityData: activityData,
		RetryPolicy:  internalRetryPolicy,
		RetryState:   retryState,
		HandlerInfo:  &handler,
		Paused:       false,
		Resumable:    false,
	}

	// Add the entity to the database, which assigns the ID
	if err = ctx.orchestrator.db.AddEntity(entity); err != nil {
		log.Printf("Error adding entity: %v", err)
		future.setError(err)
		return future
	}

	future.setEntityID(entity.ID)

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
		entity:            entity,
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

/// FILE: ./context_workflows_saga.go

// SagaContext provides context for saga execution.
type SagaContext struct {
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	stepID       string
}

func (ctx WorkflowContext) Saga(stepID string, saga *SagaDefinition) *SagaInfo {
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

/// FILE: ./context_workflows_sideffects.go

// SideEffect executes a side effect function
func (ctx WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}) *RuntimeFuture {
	var err error

	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.SideEffect paused at stepID: %s", stepID)
		future := NewRuntimeFuture()
		future.setEntityID(ctx.workflowID)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.SideEffect called with stepID: %s", stepID)
	future := NewRuntimeFuture()
	future.setEntityID(ctx.workflowID)

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
	var entity *Entity
	if entity, err = ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepIDAndType(ctx.workflowID, stepID, EntityTypeSideEffect); err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			log.Printf("Error getting side effect entity: %v", err)
			future.setError(err)
			return future
		}
	}

	if entity != nil {
		var latestExecution *Execution
		if latestExecution, err = ctx.orchestrator.db.GetLatestExecution(entity.ID); err != nil {
			log.Printf("Error getting latest execution: %v", err)
			future.setError(err)
			return future
		}

		if latestExecution.Status == ExecutionStatusCompleted && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Outputs != nil {
			// Deserialize result using the returnTypes
			outputs, err := convertOutputsFromSerialization(HandlerInfo{ReturnTypes: returnTypes}, latestExecution.SideEffectExecutionData.Outputs)
			if err != nil {
				log.Printf("Error deserializing side effect result: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}

		if latestExecution.Status == ExecutionStatusFailed && latestExecution.Error != "" {
			log.Printf("SideEffect %s has failed execution with error: %s", stepID, latestExecution.Error)
			future.setError(errors.New(latestExecution.Error))
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
	// Default retry policy with MaxAttempts=1
	retryPolicy = &RetryPolicy{
		MaxAttempts:        1,
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaxInterval:        5 * time.Minute,
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
		// No fields as per ent schema
	}

	// Create RetryState
	retryState := &RetryState{Attempts: 0}

	// Create a new Entity without ID (database assigns it)
	entity = &Entity{
		StepID:         stepID,
		HandlerName:    handlerName,
		Status:         StatusPending,
		Type:           EntityTypeSideEffect,
		RunID:          ctx.orchestrator.runID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		SideEffectData: sideEffectData,
		HandlerInfo:    &handler,
		RetryPolicy:    internalRetryPolicy,
		RetryState:     retryState,
		Paused:         false,
		Resumable:      false,
	}

	// Add the entity to the database, which assigns the ID
	if err = ctx.orchestrator.db.AddEntity(entity); err != nil {
		log.Printf("Error adding side effect entity: %v", err)
		future.setError(err)
		return future
	}

	future.setEntityID(entity.ID)

	// Prepare to create the side effect instance
	sideEffectInstance := &SideEffectInstance{
		stepID:            stepID,
		sideEffectFunc:    sideEffectFunc,
		future:            future,
		ctx:               ctx.ctx,
		orchestrator:      ctx.orchestrator,
		workflowID:        ctx.workflowID,
		entityID:          entity.ID,
		options:           nil,
		returnTypes:       returnTypes,
		handlerName:       handlerName,
		handler:           handler,
		parentExecutionID: ctx.executionID,
		parentEntityID:    ctx.workflowID,
		parentStepID:      ctx.stepID,
		entity:            entity,
	}

	// Start the side effect instance
	ctx.orchestrator.addSideEffectInstance(sideEffectInstance)
	sideEffectInstance.Start()

	return future
}

/// FILE: ./futures.go

type Future interface {
	Get(out ...interface{}) error
	setEntityID(entityID int)
	setError(err error)
	setResult(results []interface{})
	WorkflowID() int
}

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
	var err error
	var entity *Entity
	if entity, err = f.database.GetEntity(f.entityID); err != nil {
		f.err = fmt.Errorf("entity not found: %d", f.entityID)
		return true
	}

	fmt.Println("\t entity", entity.ID, " status: ", entity.Status)
	var latestExec *Execution
	// TODO: should we check for pause?
	switch entity.Status {
	case StatusCompleted:
		// Get results from latest execution
		if latestExec, err = f.database.GetLatestExecution(f.entityID); err != nil {
			f.err = err
			return true
		}
		if latestExec.WorkflowExecutionData != nil {
			outputs, err := convertOutputsFromSerialization(*entity.HandlerInfo,
				latestExec.WorkflowExecutionData.Outputs)
			if err != nil {
				f.err = err
				return true
			}
			f.results = outputs
			return true
		}
		// TODO: no output, maybe we don't care?
		return true
	case StatusFailed:
		if latestExec, err = f.database.GetLatestExecution(f.entityID); err != nil {
			f.err = err
			return true
		}

		f.err = errors.New(latestExec.Error)

		return true
	case StatusPaused:
		f.err = ErrPaused
		return true
	case StatusCancelled:
		f.err = errors.New("workflow was cancelled")
		return true
	}

	return false
}

// RuntimeFuture represents an asynchronous result.
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
	f.workflowID = entityID
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

/// FILE: ./orchestrator.go

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
	mu            deadlock.Mutex
	instancesMu   deadlock.RWMutex
	activitiesMu  deadlock.Mutex
	sideEffectsMu deadlock.Mutex
	sagasMu       deadlock.Mutex
	err           error
	runID         int
	paused        bool
	pausedMu      deadlock.Mutex
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

// prepareWorkflowEntity creates the necessary database records for a new workflow
func (o *Orchestrator) prepareWorkflowEntity(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (*Entity, error) {
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
		run := &Run{
			Status:    string(StatusPending),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err = o.db.AddRun(run); err != nil {
			return nil, fmt.Errorf("failed to create run: %w", err)
		}
		o.runID = run.ID
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
		retryPolicy = &retryPolicyInternal{
			MaxAttempts:        1,
			InitialInterval:    time.Second.Nanoseconds(),
			BackoffCoefficient: 2.0,
			MaxInterval:        (5 * time.Minute).Nanoseconds(),
		}
	}

	queueName := "default"

	if options != nil && options.Queue != "" {
		queueName = options.Queue
	}

	var queue *Queue
	if queue, err = o.db.GetQueueByName(queueName); err != nil {
		return nil, fmt.Errorf("failed to get queue %s: %w", queueName, err)
	}

	// Create Entity
	entity := &Entity{
		StepID:      "root",
		HandlerName: handler.HandlerName,
		Status:      StatusPending,
		Type:        EntityTypeWorkflow,
		RunID:       o.runID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		QueueID:     queue.ID,
		Queue:       queue,
		WorkflowData: &WorkflowData{
			Input:     inputBytes,
			Attempt:   1,
			Paused:    false,
			Resumable: false,
		},
		HandlerInfo: &handler,
		RetryState: &RetryState{
			Attempts: 0,
		},
		RetryPolicy: retryPolicy,
	}

	// Add entity to database
	if err = o.db.AddEntity(entity); err != nil {
		return nil, fmt.Errorf("failed to add entity: %w", err)
	}

	// // Create initial execution
	// execution := &Execution{
	// 	EntityID:  entity.ID,
	// 	Status:    ExecutionStatusPending,
	// 	Attempt:   1,
	// 	StartedAt: time.Now(),
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// 	Entity:    entity,
	// }
	// o.db.AddExecution(execution)

	return entity, nil
}

// ExecuteWithEntity starts a workflow using an existing entity ID
func (o *Orchestrator) ExecuteWithEntity(entityID int) (*RuntimeFuture, error) {
	// Get the entity and verify it exists
	var err error
	var entity *Entity
	if entity, err = o.db.GetEntity(entityID); err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	// Set orchestrator's runID from entity
	o.runID = entity.RunID

	// Verify it's a workflow
	if entity.Type != EntityTypeWorkflow {
		return nil, fmt.Errorf("entity %d is not a workflow", entityID)
	}

	// Get handler info
	handler := entity.HandlerInfo
	if handler == nil {
		return nil, fmt.Errorf("no handler info for entity %d", entityID)
	}

	// Convert inputs
	inputs, err := convertInputsFromSerialization(*handler, entity.WorkflowData.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert inputs: %w", err)
	}

	// Create workflow instance
	instance := &WorkflowInstance{
		stepID:       entity.StepID,
		handler:      *handler,
		input:        inputs,
		ctx:          o.ctx,
		orchestrator: o,
		workflowID:   entity.ID,
		entity:       entity,
		entityID:     entity.ID,
	}

	// If this is a sub-workflow, set parent info from hierarchies
	var hierarchies []*Hierarchy
	if hierarchies, err = o.db.GetHierarchiesByChildEntity(entityID); err != nil {
		return nil, fmt.Errorf("failed to get hierarchies: %w", err)
	}
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

	// very important to notice the orchestrator of the real root workflow
	o.rootWf = instance

	o.addWorkflowInstance(instance)
	go instance.Start()

	return future, nil
}

func (o *Orchestrator) Execute(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *RuntimeFuture {
	// Create entity and related records
	entity, err := o.prepareWorkflowEntity(workflowFunc, options, args)
	if err != nil {
		future := NewRuntimeFuture()
		future.setError(err)
		return future
	}

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

func (o *Orchestrator) Pause() {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	o.paused = true
	log.Printf("Orchestrator paused")
}

func (o *Orchestrator) IsPaused() bool {
	o.pausedMu.Lock()
	defer o.pausedMu.Unlock()
	return o.paused
}

func (o *Orchestrator) Resume(entityID int) *RuntimeFuture {
	var err error
	future := NewRuntimeFuture()
	future.setEntityID(entityID)

	o.pausedMu.Lock()
	o.paused = false
	o.pausedMu.Unlock()

	// Retrieve the workflow entity
	var entity *Entity
	if entity, err = o.db.GetEntity(entityID); err != nil {
		log.Printf("Error getting entity: %v", err)
		future.setError(err)
		return future
	}

	// Set the runID from the entity
	o.runID = entity.RunID

	// Update the entity state
	entity.Paused = false
	entity.Status = StatusPending
	entity.Resumable = true
	o.db.UpdateEntity(entity)

	// Update Run status
	var run *Run
	if run, err = o.db.GetRun(o.runID); err != nil {
		log.Printf("Error getting run: %v", err)
		future.setError(err)
		return future
	}

	run.Status = string(StatusPending)
	run.UpdatedAt = time.Now()
	if err = o.db.UpdateRun(run); err != nil {
		log.Printf("Error updating run: %v", err)
		future.setError(err)
		return future
	}

	// Get parent execution if this is a sub-workflow
	var parentExecID, parentEntityID int
	var hierarchies []*Hierarchy
	if hierarchies, err = o.db.GetHierarchiesByChildEntity(entityID); err != nil {
		log.Printf("Error getting hierarchies: %v", err)
		future.setError(err)
		return future
	}

	if len(hierarchies) > 0 {
		parentExecID = hierarchies[0].ParentExecutionID
		parentEntityID = hierarchies[0].ParentEntityID
	}

	// Retrieve the handler
	handlerInfo := entity.HandlerInfo
	if handlerInfo == nil {
		log.Printf("No handler info found for workflow: %s", entity.HandlerName)
		future.setError(fmt.Errorf("no handler info found for workflow: %s", entity.HandlerName))
		return future
	}
	handler := *handlerInfo

	// Convert inputs from serialization
	inputs, err := convertInputsFromSerialization(handler, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		future.setError(err)
		return future
	}

	// Create a new WorkflowInstance with parent info
	instance := &WorkflowInstance{
		stepID:            entity.StepID,
		handler:           handler,
		input:             inputs,
		ctx:               o.ctx,
		orchestrator:      o,
		workflowID:        entity.ID,
		entity:            entity,
		options:           nil,
		entityID:          entity.ID,
		parentExecutionID: parentExecID,
		parentEntityID:    parentEntityID,
		parentStepID:      entity.StepID,
	}

	future.setEntityID(entity.ID)
	instance.future = future

	// Lock orchestrator's instancesMu before modifying shared state
	o.instancesMu.Lock()
	o.rootWf = instance
	o.instancesMu.Unlock()

	// Start the instance
	instance.Start()

	return future
}

func (o *Orchestrator) stopWithError(err error) {
	o.err = err
	if o.rootWf != nil && o.rootWf.fsm != nil {
		o.instancesMu.Lock()
		o.rootWf.fsm.Fire(TriggerFail)
		o.instancesMu.Unlock()
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

func (o *Orchestrator) Wait() error {
	log.Printf("Orchestrator.Wait called")
	var err error

	// Lock instancesMu before accessing rootWf
	o.instancesMu.RLock()
	if o.rootWf == nil {
		o.instancesMu.RUnlock()
		log.Printf("No root workflow to execute")
		return fmt.Errorf("no root workflow to execute")
	}
	rootWf := o.rootWf
	o.instancesMu.RUnlock()

	lastLogTime := time.Now()
	// Wait for root workflow to complete
	for {
		rootWf.mu.Lock()
		if rootWf.fsm == nil {
			rootWf.mu.Unlock()
			continue
		}
		state := rootWf.fsm.MustState()
		rootWf.mu.Unlock()

		// Log the state every 500ms
		currentTime := time.Now()
		if currentTime.Sub(lastLogTime) >= 500*time.Millisecond {
			log.Printf("Root Workflow FSM state: %s", state)
			lastLogTime = currentTime
		}
		if state == StateCompleted || state == StateFailed || state == StatePaused {
			break
		}
		select {
		case <-o.ctx.Done():
			log.Printf("Context cancelled: %v", o.ctx.Err())
			rootWf.mu.Lock()
			rootWf.fsm.Fire(TriggerFail)
			rootWf.mu.Unlock()
			return o.ctx.Err()
		default:
			runtime.Gosched()
		}
	}

	// Root workflow has completed or failed
	// Get final status safely
	status, err := o.db.GetEntityStatus(o.rootWf.entityID)
	if err != nil {
		log.Printf("error getting entity status: %v", err)
		return err
	}

	var latestExecution *Execution
	if latestExecution, err = o.db.GetLatestExecution(o.rootWf.entityID); err != nil {
		log.Printf("error getting latest execution: %v", err)
		return err
	}

	switch status {
	case StatusCompleted:
		o.rootWf.mu.Lock()
		results := o.rootWf.results
		o.rootWf.mu.Unlock()
		if results != nil && len(results) > 0 {
			fmt.Printf("Root workflow completed successfully with results: %v\n", results)
		} else {
			fmt.Printf("Root workflow completed successfully\n")
		}
	case StatusPaused:
		fmt.Printf("Root workflow paused\n")
	case StatusFailed:
		var errMsg string
		if latestExecution.Error != "" {
			errMsg = latestExecution.Error
		} else {
			o.rootWf.mu.Lock()
			if o.rootWf.err != nil {
				errMsg = o.rootWf.err.Error()
			} else {
				errMsg = "unknown error"
			}
			o.rootWf.mu.Unlock()
		}
		fmt.Printf("Root workflow failed with error: %v\n", errMsg)
	default:
		fmt.Printf("Root workflow ended with status: %s\n", status)
	}
	return nil
}

// Retry retries a failed root workflow by creating a new entity and execution.
func (o *Orchestrator) Retry(workflowID int) *RuntimeFuture {

	var err error

	future := NewRuntimeFuture()
	future.setEntityID(workflowID)

	// Retrieve the workflow entity
	var entity *Entity
	if entity, err = o.db.GetEntity(workflowID); err != nil {
		log.Printf("Error getting entity: %v", err)
		future.setError(err)
		return future
	}

	// Set the runID from the entity
	o.runID = entity.RunID

	// Copy inputs
	inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		future.setError(err)
		return future
	}

	// Create a new WorkflowData
	workflowData := &WorkflowData{
		Duration:  "",
		Paused:    false,
		Resumable: false,
		Input:     entity.WorkflowData.Input,
		Attempt:   1,
	}

	// Create RetryState
	retryState := &RetryState{Attempts: 0}

	// Create a new Entity
	newEntity := &Entity{
		StepID:       entity.StepID,
		HandlerName:  entity.HandlerName,
		Status:       StatusPending,
		Type:         EntityTypeWorkflow,
		RunID:        o.runID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WorkflowData: workflowData,
		RetryPolicy:  entity.RetryPolicy,
		RetryState:   retryState,
		HandlerInfo:  entity.HandlerInfo,
		Paused:       false,
		Resumable:    false,
	}

	if err = o.db.AddEntity(newEntity); err != nil {
		log.Printf("Error adding entity: %v", err)
		future.setError(err)
		return future
	}

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:            newEntity.StepID,
		handler:           *entity.HandlerInfo,
		input:             inputs,
		ctx:               o.ctx,
		orchestrator:      o,
		workflowID:        newEntity.ID,
		options:           nil, // You might want to use the same options or set new ones
		entity:            newEntity,
		entityID:          newEntity.ID,
		parentExecutionID: 0,
		parentEntityID:    0,
		parentStepID:      "",
	}

	instance.future = future

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return future
}

func (o *Orchestrator) executeSaga(ctx WorkflowContext, stepID string, saga *SagaDefinition) *SagaInfo {
	sagaInfo := &SagaInfo{
		done: make(chan struct{}),
	}
	var err error

	// Create a new Entity
	entity := &Entity{
		StepID:    stepID,
		Status:    StatusPending,
		Type:      EntityTypeSaga,
		RunID:     o.runID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		SagaData:  &SagaData{},
	}
	entity.HandlerInfo = &HandlerInfo{}

	// Add the entity to the database
	if err = o.db.AddEntity(entity); err != nil {
		log.Printf("Error adding entity: %v", err)
		return nil
	}

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

func (o *Orchestrator) GetWorkflow(id int) (*WorkflowInfo, error) {
	var err error

	var entity *Entity
	if entity, err = o.db.GetEntity(id); err != nil {
		return nil, fmt.Errorf("error getting entity: %v", err)
	}

	var run *Run
	if run, err = o.db.GetRun(entity.RunID); err != nil {
		return nil, fmt.Errorf("error getting run: %v", err)
	}

	info := &WorkflowInfo{
		EntityID: entity.ID,
		Status:   entity.Status,
		Run: &RunInfo{
			RunID:  entity.RunID,
			Status: run.Status,
		},
	}

	return info, nil
}

// WaitForContinuations waits for all continuations of a workflow to complete
func (o *Orchestrator) WaitForContinuations(originalID int) error {
	log.Printf("Waiting for all continuations starting from workflow %d", originalID)
	// var err error
	currentID := originalID

	for {
		// Lock instancesMu before accessing rootWf
		o.instancesMu.RLock()
		rootWf := o.rootWf
		o.instancesMu.RUnlock()

		if rootWf == nil {
			// If rootWf is not yet initialized, wait a bit and retry
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check the state of rootWf
		rootWf.mu.Lock()
		state := rootWf.fsm.MustState()
		rootWf.mu.Unlock()

		// Wait for workflow to complete
		for state != StateCompleted && state != StateFailed {
			select {
			case <-o.ctx.Done():
				log.Printf("Context cancelled: %v", o.ctx.Err())
				rootWf.mu.Lock()
				rootWf.fsm.Fire(TriggerFail)
				rootWf.mu.Unlock()
				return o.ctx.Err()
			default:
				runtime.Gosched()
			}
			rootWf.mu.Lock()
			state = rootWf.fsm.MustState()
			rootWf.mu.Unlock()
		}

		// Check if this workflow initiated a continuation
		hasEntity, err := o.db.HasEntity(currentID)
		if err != nil {
			return fmt.Errorf("error checking for entity: %v", err)
		}

		if !hasEntity {
			return fmt.Errorf("workflow %d not found", currentID)
		}

		latestExecution, err := o.db.GetLatestExecution(currentID)
		if err != nil {
			return fmt.Errorf("error getting latest execution: %v", err)
		}

		if latestExecution.Status == ExecutionStatusCompleted {
			// Look for any child workflow that was created via ContinueAsNew
			childEntity, err := o.db.GetChildEntityByParentEntityIDAndStepIDAndType(currentID, "root", EntityTypeWorkflow)
			if err != nil {
				if errors.Is(err, ErrEntityNotFound) {
					log.Printf("No more continuations found after workflow %d", currentID)
					return nil
				}
				return err
			}

			currentID = childEntity.ID

			// Update the rootWf to the new workflow instance
			o.instancesMu.Lock()
			var found bool
			for _, wi := range o.instances {
				if wi.entityID == currentID {
					// Ensure wi.fsm is initialized
					wi.mu.Lock()
					if wi.fsm != nil {
						o.rootWf = wi
						found = true
					}
					wi.mu.Unlock()
					break
				}
			}
			o.instancesMu.Unlock()

			if !found {
				// Wait for the new workflow instance to be added and initialized
				time.Sleep(100 * time.Millisecond)
				continue
			}

			log.Printf("Found continuation: workflow %d", currentID)
		} else if latestExecution.Status == ExecutionStatusFailed {
			return fmt.Errorf("workflow %d failed", currentID)
		} else {
			return nil
		}
	}
}

/// FILE: ./queue.go

type queueRequestType string

var (
	queueRequestTypePause   queueRequestType = "pause"
	queueRequestTypeResume  queueRequestType = "resume"
	queueRequestTypeExecute queueRequestType = "execute"
)

type taskRequest struct {
	requestType queueRequestType
	entityID    int
}

// QueueManager manages a single queue and its workers
type QueueManager struct {
	mu              deadlock.RWMutex
	name            string
	pool            *retrypool.Pool[*QueueTask]
	database        Database
	registry        *Registry
	ctx             context.Context
	cancel          context.CancelFunc
	workerCount     int
	orchestrator    *Orchestrator
	cache           map[int]*QueueWorker
	entitiesWorkers map[int]int
	busy            map[int]struct{}
	free            map[int]struct{}
	requestPool     *retrypool.Pool[*retrypool.RequestResponse[*taskRequest, struct{}]]
}

// QueueTask represents a workflow execution task
type QueueTask struct {
	workflowFunc interface{}      // The workflow function to execute
	options      *WorkflowOptions // Workflow execution options
	args         []interface{}    // Arguments for the workflow
	future       *RuntimeFuture   // Future for the task
	queueName    string           // Name of the queue this task belongs to
	entityID     int              // Add this field
}

// QueueInfo provides information about a queue's state
type QueueInfo struct {
	Name            string
	WorkerCount     int
	PendingTasks    int
	ProcessingTasks int
	FailedTasks     int
}

// QueueWorker represents a worker in a queue
type QueueWorker struct {
	ID           int // automatically set by the pool * magic *
	queueName    string
	orchestrator *Orchestrator
	ctx          context.Context
	mu           deadlock.Mutex
	onStartTask  func(*QueueWorker, *QueueTask)
	onEndTask    func(*QueueWorker, *QueueTask)
	onStart      func(*QueueWorker)
}

func (w *QueueWorker) OnStart(ctx context.Context) {
	if w.onStart != nil {
		w.onStart(w)
	}
}

func (w *QueueWorker) Run(ctx context.Context, task *QueueTask) error {

	fmt.Println("\t workflow started on queue", w.queueName, task.entityID)
	if w.onStartTask != nil {
		w.onStartTask(w, task)
	}
	defer func() {
		if w.onEndTask != nil {
			w.onEndTask(w, task)
		}
	}()

	// Now execute the workflow using our existing orchestrator
	ftre, err := w.orchestrator.ExecuteWithEntity(task.entityID)
	if err != nil {
		if task.future != nil {
			task.future.setError(err)
		}
		return fmt.Errorf("failed to execute workflow task %d: %w", task.entityID, err)
	}

	if err := w.orchestrator.Wait(); err != nil {
		if task.future != nil {
			task.future.setError(err)
		}
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if errors.Is(err, ErrPaused) {
			return nil
		}
		return err
	}

	fmt.Println("\t workflow waiting GET")
	// Wait for completion and propagate error/result
	if err := ftre.Get(); err != nil {
		if task.future != nil {
			task.future.setError(err)
		}
		if errors.Is(err, context.Canceled) && errors.Is(err, ErrPaused) {
			return nil
		}
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// if optional future is set, propagate results
	if task.future != nil {
		w.orchestrator.rootWf.mu.Lock()
		task.future.setResult(ftre.results)
		w.orchestrator.rootWf.mu.Unlock()
	}

	fmt.Println("\t workflow done")

	return nil
}

func newQueueManager(ctx context.Context, name string, workerCount int, registry *Registry, db Database) *QueueManager {
	ctx, cancel := context.WithCancel(ctx)
	var qm *QueueManager
	qm = &QueueManager{
		name:            name,
		database:        db,
		registry:        registry,
		ctx:             ctx,
		cancel:          cancel,
		workerCount:     workerCount,
		orchestrator:    NewOrchestrator(ctx, db, registry),
		cache:           make(map[int]*QueueWorker),
		free:            make(map[int]struct{}),
		busy:            make(map[int]struct{}),
		entitiesWorkers: make(map[int]int),
		requestPool: retrypool.New(
			ctx,
			[]retrypool.Worker[*retrypool.RequestResponse[*taskRequest, struct{}]]{},
			retrypool.WithLogLevel[*retrypool.RequestResponse[*taskRequest, struct{}]](logs.LevelDebug),
			retrypool.WithAttempts[*retrypool.RequestResponse[*taskRequest, struct{}]](3),
			retrypool.WithDelay[*retrypool.RequestResponse[*taskRequest, struct{}]](time.Second/2),
			retrypool.WithOnNewDeadTask[*retrypool.RequestResponse[*taskRequest, struct{}]](func(task *retrypool.DeadTask[*retrypool.RequestResponse[*taskRequest, struct{}]], idx int) {
				errs := errors.New("failed to process request")
				for _, e := range task.Errors {
					errs = errors.Join(errs, e)
				}
				task.Data.CompleteWithError(errs)
				_, err := qm.requestPool.PullDeadTask(idx)
				if err != nil {
					// too bad
					log.Printf("failed to pull dead task: %v", err)
				}
			}),
		),
	}

	qm.mu.Lock()
	poolRequest := qm.requestPool
	qm.pool = qm.createWorkerPool(workerCount)
	qm.mu.Unlock()

	// TODO: how much?
	poolRequest.AddWorker(&queueWorkerRequests{
		qm: qm,
	})

	poolRequest.AddWorker(&queueWorkerRequests{
		qm: qm,
	})

	fmt.Println("Queue manager", name, "created with", workerCount, "workers", qm.pool.AvailableWorkers(), "available workers", qm.requestPool.AvailableWorkers(), "available request workers")

	return qm
}

type queueWorkerRequests struct {
	ID int
	qm *QueueManager
}

func (w *queueWorkerRequests) OnStart(ctx context.Context) {
	fmt.Println("queueWorkerRequests started")
}

func (w *queueWorkerRequests) Run(ctx context.Context, task *retrypool.RequestResponse[*taskRequest, struct{}]) error {

	var err error

	fmt.Println("\t queueWorkerRequests", task.Request.requestType, task.Request.entityID)
	defer fmt.Println("\t queueWorkerRequests done", task.Request.requestType, task.Request.entityID)

	switch task.Request.requestType {
	case queueRequestTypePause:

		workerID, ok := w.qm.entitiesWorkers[task.Request.entityID]
		if !ok {
			return fmt.Errorf("entity %d not found", task.Request.entityID)
		}
		task.Complete(struct{}{})
		w.qm.cache[workerID].orchestrator.Pause()

	case queueRequestTypeResume:

		if w.qm.pool.AvailableWorkers() == 0 {
			return fmt.Errorf("no available workers for entity %d", task.Request.entityID)
		}

		var freeWorkers []int
		for workerID := range w.qm.free {
			freeWorkers = append(freeWorkers, workerID)
		}
		if len(freeWorkers) == 0 {
			return fmt.Errorf("no free workers available")
		}

		randomWorkerID := freeWorkers[rand.Intn(len(freeWorkers))]
		worker := w.qm.cache[randomWorkerID]
		task.Complete(struct{}{})
		worker.orchestrator.Resume(task.Request.entityID)

	case queueRequestTypeExecute:

		// Verify entity exists and is ready for execution
		var entity *Entity
		if entity, err = w.qm.database.GetEntity(task.Request.entityID); err != nil {
			task.CompleteWithError(fmt.Errorf("failed to get entity %d: %w", task.Request.entityID, err))
			return fmt.Errorf("failed to get entity %d: %w", task.Request.entityID, err)
		}

		if entity == nil {
			task.CompleteWithError(fmt.Errorf("entity not found: %d", task.Request.entityID))
			return fmt.Errorf("entity not found: %d", task.Request.entityID)
		}

		if entity.Status != StatusPending {
			task.CompleteWithError(fmt.Errorf("entity %d is not in pending status", task.Request.entityID))
			return fmt.Errorf("entity %d is not in pending status", task.Request.entityID)
		}

		// Get handler info
		if entity.HandlerInfo == nil {
			task.CompleteWithError(fmt.Errorf("no handler info for entity %d", task.Request.entityID))
			return fmt.Errorf("no handler info for entity %d", task.Request.entityID)
		}

		// Convert stored inputs
		inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
		if err != nil {
			task.CompleteWithError(fmt.Errorf("failed to convert inputs: %w", err))
			return fmt.Errorf("failed to convert inputs: %w", err)
		}

		// Create task for execution
		queueTask := &QueueTask{
			workflowFunc: entity.HandlerInfo.Handler,
			options:      nil, // Could be stored in entity if needed
			args:         inputs,
			queueName:    w.qm.name,
			entityID:     task.Request.entityID,
		}

		processed := retrypool.NewProcessedNotification()
		queued := retrypool.NewQueuedNotification()

		// Submit to worker pool
		if err := w.qm.pool.Submit(
			queueTask,
			retrypool.WithBeingProcessed[*QueueTask](processed),
			retrypool.WithQueued[*QueueTask](queued),
		); err != nil {
			task.CompleteWithError(fmt.Errorf("failed to submit entity %d to queue %s: %w", task.Request.entityID, w.qm.name, err))
			return fmt.Errorf("failed to submit entity %d to queue %s: %w", task.Request.entityID, w.qm.name, err)
		}

		fmt.Println("workflow queuing", w.qm.name, task.Request.entityID, w.qm.pool.AvailableWorkers(), w.qm.pool.QueueSize(), w.qm.pool.ProcessingCount(), queued)
		<-queued.Done()
		fmt.Println("workflow submitted to queue", w.qm.name, task.Request.entityID, w.qm.pool.AvailableWorkers(), w.qm.pool.QueueSize(), w.qm.pool.ProcessingCount(), processed)
		<-processed.Done()

		task.Complete(struct{}{})

		return nil
	}
	return nil
}

func (qm *QueueManager) Pause(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		requestType: queueRequestTypePause,
		entityID:    id,
	})
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	return task
}

func (qm *QueueManager) AvailableWorkers() int {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.pool.AvailableWorkers()
}

// Resume entity on available worker
func (qm *QueueManager) Resume(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		entityID:    id,
		requestType: queueRequestTypeResume,
	})
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	return task
}

func (qm *QueueManager) onTaskStart(worker *QueueWorker, task *QueueTask) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.entitiesWorkers[task.entityID] = worker.ID
	qm.cache[worker.ID] = worker
	qm.busy[worker.ID] = struct{}{}
	delete(qm.free, worker.ID)
	fmt.Println("\t onTaskStart", qm.entitiesWorkers, qm.busy, qm.free)
}

func (qm *QueueManager) onTaskEnd(worker *QueueWorker, task *QueueTask) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	delete(qm.entitiesWorkers, task.entityID)
	delete(qm.busy, worker.ID)
	qm.free[worker.ID] = struct{}{}
	fmt.Println("\t onTaskEnd", qm.entitiesWorkers, qm.busy, qm.free)
}

func (qm *QueueManager) GetEntities() []int {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	entities := make([]int, 0, len(qm.cache))
	for id := range qm.cache {
		entities = append(entities, id)
	}
	return entities
}

func (qm *QueueManager) createWorkerPool(count int) *retrypool.Pool[*QueueTask] {
	workers := make([]retrypool.Worker[*QueueTask], count)
	for i := 0; i < count; i++ {
		workers[i] = &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
			queueName:    qm.name,
			onStartTask:  qm.onTaskStart,
			onEndTask:    qm.onTaskEnd,
			onStart: func(w *QueueWorker) {
				qm.free[w.ID] = struct{}{}
				qm.cache[w.ID] = w
				fmt.Println("worker started", w.ID)
			},
		}
	}

	pool := retrypool.New(qm.ctx, workers, []retrypool.Option[*QueueTask]{
		retrypool.WithAttempts[*QueueTask](1),
		retrypool.WithLogLevel[*QueueTask](logs.LevelDebug),
		retrypool.WithOnTaskFailure[*QueueTask](qm.handleTaskFailure),
	}...)

	return pool
}

// Starts the automatic queue pulling process
func (qm *QueueManager) Start() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-ticker.C:
			qm.checkAndProcessPending()
		}
	}
}

func (qm *QueueManager) checkAndProcessPending() {
	var err error

	// Get current queue state
	info := qm.GetInfo()
	availableSlots := info.WorkerCount - info.ProcessingTasks

	if availableSlots <= 0 {
		return
	}

	// Get queue first
	var queue *Queue
	if queue, err = qm.database.GetQueueByName(qm.name); err != nil {
		log.Printf("Failed to get queue %s: %v", qm.name, err)
		return
	}

	// Find pending workflows for this queue
	var pendingEntities []*Entity
	if pendingEntities, err = qm.database.FindPendingWorkflowsByQueue(queue.ID); err != nil {
		log.Printf("Failed to get pending workflows for queue %s: %v", qm.name, err)
		return
	}

	if len(pendingEntities) == 0 {
		// fmt.Println("Queue", qm.name, "has no pending tasks", info.PendingTasks, "pending tasks and", availableSlots, "available slots")
		return
	}

	// fmt.Println("Queue", qm.name, "has", info.PendingTasks, "pending tasks and", availableSlots, "available slots and", len(pendingEntities), "pending entities")

	for i := 0; i < min(availableSlots, len(pendingEntities)); i++ {
		entity := pendingEntities[i]

		// Recheck entity state from database
		var freshEntity *Entity
		if freshEntity, err = qm.database.GetEntity(entity.ID); err != nil {
			log.Printf("Failed to get entity %d: %v", entity.ID, err)
			continue
		}

		if freshEntity.Status != StatusPending {
			continue
		}

		task := qm.ExecuteWorkflow(freshEntity.ID)

		// Wait for processing to start before continuing
		select {
		case <-task.Done():
			if task.Err() != nil {
				log.Printf("Failed to execute workflow %d on queue %s: %v",
					freshEntity.ID, qm.name, task.Err())
			}
			// Processing started
		case <-qm.ctx.Done():
			return
		}
	}
}

func (am *QueueManager) CreateWorkflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) (int, error) {
	// Register handler before creating entity
	if _, err := am.registry.RegisterWorkflow(workflowFunc); err != nil {
		return 0, fmt.Errorf("failed to register workflow: %w", err)
	}

	// Prepare the workflow entity
	entity, err := am.orchestrator.prepareWorkflowEntity(workflowFunc, options, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare workflow entity: %w", err)
	}

	return entity.ID, nil
}

func (qm *QueueManager) ExecuteWorkflow(id int) *retrypool.RequestResponse[*taskRequest, struct{}] {
	task := retrypool.NewRequestResponse[*taskRequest, struct{}](&taskRequest{
		entityID:    id,
		requestType: queueRequestTypeExecute,
	})
	qm.mu.Lock()
	qm.requestPool.Submit(task)
	qm.mu.Unlock()
	return task
}

func (qm *QueueManager) handleTaskFailure(controller retrypool.WorkerController[*QueueTask], workerID int, worker retrypool.Worker[*QueueTask], data *QueueTask, retries int, totalDuration time.Duration, timeLimit time.Duration, maxDuration time.Duration, scheduledTime time.Time, triedWorkers map[int]bool, taskErrors []error, durations []time.Duration, queuedAt []time.Time, processedAt []time.Time, err error) retrypool.DeadTaskAction {
	if errors.Is(err, ErrPaused) {
		return retrypool.DeadTaskActionDoNothing
	}
	log.Printf("Task failed in queue %s: %v", qm.name, err)
	return retrypool.DeadTaskActionDoNothing
}

func (qm *QueueManager) AddWorkers(count int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := 0; i < count; i++ {
		worker := &QueueWorker{
			orchestrator: NewOrchestrator(qm.ctx, qm.database, qm.registry),
			ctx:          qm.ctx,
			queueName:    qm.name,
			onStartTask:  qm.onTaskStart,
			onEndTask:    qm.onTaskEnd,
			onStart: func(w *QueueWorker) {
				qm.free[w.ID] = struct{}{}
				qm.cache[w.ID] = w
			},
		}
		qm.pool.AddWorker(worker)
	}
	qm.workerCount += count

	return nil
}

func (qm *QueueManager) RemoveWorkers(count int) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if count > qm.workerCount {
		return fmt.Errorf("cannot remove %d workers, only %d available", count, qm.workerCount)
	}

	workers := qm.pool.ListWorkers()
	for i := 0; i < count && i < len(workers); i++ {
		delete(qm.cache, workers[len(workers)-1-i].ID)
		delete(qm.busy, workers[len(workers)-1-i].ID)
		delete(qm.free, workers[len(workers)-1-i].ID)
		if err := qm.pool.RemoveWorker(workers[len(workers)-1-i].ID); err != nil {
			return fmt.Errorf("failed to remove worker: %w", err)
		}
	}
	qm.workerCount -= count

	return nil
}

func (qm *QueueManager) GetInfo() *QueueInfo {
	metrics := qm.pool.Metrics()
	return &QueueInfo{
		Name:            qm.name,
		WorkerCount:     qm.workerCount,
		PendingTasks:    qm.pool.QueueSize(),
		ProcessingTasks: qm.pool.ProcessingCount(),
		FailedTasks:     int(metrics.TasksFailed),
	}
}

func (qm *QueueManager) pauseAll() {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	for workerID := range qm.busy {
		worker := qm.cache[workerID]
		worker.orchestrator.Pause()
	}
}

func (qm *QueueManager) Close() error {
	go func() {
		<-time.After(5 * time.Second) // grace period
		fmt.Println("Queue manager", qm.name, "force closing")
		qm.cancel()
	}()

	shutdown := errgroup.Group{}

	shutdown.Go(func() error {
		err := qm.requestPool.Close()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			fmt.Println("Request pool closed", err)
			return err
		}
		return nil
	})

	shutdown.Go(func() error {
		qm.pauseAll()
		err := qm.pool.Close()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			fmt.Println("Queue", qm.name, "closed", err)
			return err
		}
		return nil
	})
	defer qm.cancel()
	return shutdown.Wait()
}

func (qm *QueueManager) Wait() error {
	return qm.pool.WaitWithCallback(qm.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		log.Printf("Wait Queue %s - Workers: %d, Pending: %d, Processing: %d, Failed: %d",
			qm.name, qm.workerCount, queueSize, processingCount, deadTaskCount)
		return queueSize > 0 || processingCount > 0
	}, time.Second)
}

/// FILE: ./registry.go

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

// Registry holds registered workflows, activities, and side effects.
type Registry struct {
	workflows   map[string]HandlerInfo
	activities  map[string]HandlerInfo
	sideEffects map[string]HandlerInfo
	mu          deadlock.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		workflows:   make(map[string]HandlerInfo),
		activities:  make(map[string]HandlerInfo),
		sideEffects: make(map[string]HandlerInfo),
	}
}

func (r *Registry) GetWorkflow(name string) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handler, ok := r.workflows[name]
	return handler, ok
}

func (r *Registry) GetActivity(name string) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handler, ok := r.activities[name]
	return handler, ok
}

func (r *Registry) GetWorkflowFunc(f interface{}) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := getFunctionName(f)
	handler, ok := r.workflows[name]
	return handler, ok
}

func (r *Registry) GetActivityFunc(f interface{}) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := getFunctionName(f)
	handler, ok := r.activities[name]
	return handler, ok
}

// TODO: add description schema so later features
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

	expectedContextType := reflect.TypeOf(WorkflowContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of workflow function must be WorkflowContext")
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

/// FILE: ./sagas.go

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

type SagaInstance struct {
	mu                deadlock.Mutex
	saga              *SagaDefinition
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	stepID            string
	sagaInfo          *SagaInfo
	entity            *Entity
	fsm               *stateless.StateMachine
	err               error
	currentStep       int
	compensations     []int // Indices of steps to compensate
	executionID       int
	execution         *Execution
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

func (si *SagaInstance) Start() {
	// Initialize the FSM
	si.mu.Lock()
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
	si.mu.Unlock()

	// Start the FSM without holding the mutex
	go func() {
		if err := si.fsm.Fire(TriggerStart); err != nil {
			log.Printf("Error starting saga: %v", err)
		}
	}()
}

func (si *SagaInstance) executeSaga(_ context.Context, _ ...interface{}) error {
	// Create Execution without ID
	execution := &Execution{
		EntityID:  si.entity.ID,
		Attempt:   1,
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Add execution to database, which assigns the ID
	if err := si.orchestrator.db.AddExecution(execution); err != nil {
		log.Printf("Error adding execution: %v", err)
		return err
	}
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
	// Lock only around shared data access
	// Remove or minimize the usage of si.mu.Lock() if possible
	for si.currentStep < len(si.saga.Steps) {
		// Lock before accessing or modifying shared data
		si.mu.Lock()
		currentStep := si.currentStep
		si.mu.Unlock()

		step := si.saga.Steps[currentStep]
		// Perform the transaction without holding the lock
		_, err := step.Transaction(TransactionContext{
			ctx: si.ctx,
		})
		if err != nil {
			// Handle error
			si.mu.Lock()
			si.err = fmt.Errorf("transaction failed at step %d: %v", currentStep, err)
			si.mu.Unlock()
			si.executeCompensations()
			return
		}
		// Update shared data
		si.mu.Lock()
		si.compensations = append(si.compensations, currentStep)
		si.currentStep++
		si.mu.Unlock()
	}

	si.fsm.Fire(TriggerComplete)
}

func (si *SagaInstance) executeCompensations() {
	// Create a separate slice for tracking compensation errors
	var compensationErrors []error

	// Compensate in reverse order
	for i := len(si.compensations) - 1; i >= 0; i-- {
		stepIndex := si.compensations[i]
		step := si.saga.Steps[stepIndex]

		log.Printf("Executing compensation for step %d", stepIndex)

		_, err := step.Compensation(CompensationContext{
			ctx: si.ctx,
		})
		if err != nil {
			// Log compensation failure but continue with other compensations
			log.Printf("Compensation failed for step %d: %v", stepIndex, err)
			compensationErrors = append(compensationErrors, fmt.Errorf("compensation failed at step %d: %v", stepIndex, err))
		}
	}

	// If there were any compensation errors, append them to the saga error
	if len(compensationErrors) > 0 {
		var errMsg string
		if si.err != nil {
			errMsg = si.err.Error() + "; "
		}
		errMsg += "compensation errors: "
		for i, err := range compensationErrors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		si.err = errors.New(errMsg)
	}

	// Update the execution status
	si.execution.Status = ExecutionStatusFailed
	si.execution.Error = si.err.Error()
	si.orchestrator.db.UpdateExecution(si.execution)

	// Fire the fail trigger after compensations are complete
	si.fsm.Fire(TriggerFail)
}

func (si *SagaInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	// Update entity status to Completed
	si.entity.Status = StatusCompleted
	si.orchestrator.db.UpdateEntity(si.entity)

	// Update execution status to Completed
	si.execution.Status = ExecutionStatusCompleted
	completedAt := time.Now()
	si.execution.CompletedAt = &completedAt
	si.orchestrator.db.UpdateExecution(si.execution)

	// Notify SagaInfo
	si.sagaInfo.err = nil
	close(si.sagaInfo.done)
	log.Println("Saga completed successfully")
	return nil
}

func (si *SagaInstance) onFailed(_ context.Context, _ ...interface{}) error {
	var err error

	// Update entity status to Failed
	si.entity.Status = StatusFailed
	si.orchestrator.db.UpdateEntity(si.entity)

	// Update execution status to Failed
	si.execution.Status = ExecutionStatusFailed
	completedAt := time.Now()
	si.execution.CompletedAt = &completedAt
	si.orchestrator.db.UpdateExecution(si.execution)

	// Set the error in SagaInfo
	if si.err != nil {
		si.sagaInfo.err = si.err
	} else {
		si.sagaInfo.err = errors.New("saga execution failed")
	}

	// Mark parent Workflow as Failed
	var parentEntity *Entity
	if parentEntity, err = si.orchestrator.db.GetEntity(si.workflowID); err != nil {
		log.Printf("Error getting parent entity %d: %v", si.workflowID, err)
		return err
	}

	parentEntity.Status = StatusFailed
	si.orchestrator.db.UpdateEntity(parentEntity)

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

/// FILE: ./sideffects.go

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
	mu                deadlock.Mutex
	stepID            string
	sideEffectFunc    interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            *RuntimeFuture
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	entityID          int
	executionID       int
	options           *WorkflowOptions
	execution         *Execution // Current execution
	returnTypes       []reflect.Type
	handlerName       string
	handler           HandlerInfo
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
	entity            *Entity
}

func (sei *SideEffectInstance) Start() {
	sei.mu.Lock()

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
	sei.mu.Unlock()

	// Start the FSM
	go func() {
		sei.fsm.Fire(TriggerStart)
	}()
}

func (sei *SideEffectInstance) executeSideEffect(_ context.Context, _ ...interface{}) error {
	return sei.executeWithRetry()
}

func (sei *SideEffectInstance) executeWithRetry() error {
	var err error

	var attempt int
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration

	if sei.entity.RetryPolicy != nil {
		rp := sei.entity.RetryPolicy
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

	for attempt = sei.entity.RetryState.Attempts + 1; attempt <= maxAttempts; attempt++ {
		if sei.orchestrator.IsPaused() {
			log.Printf("SideEffectInstance %s is paused", sei.stepID)
			sei.entity.Status = StatusPaused
			sei.entity.Paused = true
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.mu.Lock()
			sei.fsm.Fire(TriggerPause)
			sei.mu.Unlock()
			return nil
		}

		// Update RetryState
		sei.entity.RetryState.Attempts = attempt
		sei.orchestrator.db.UpdateEntity(sei.entity)

		// Create Execution without ID
		execution := &Execution{
			EntityID:  sei.entity.ID,
			Status:    ExecutionStatusRunning,
			Attempt:   attempt,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entity:    sei.entity,
		}

		// Add execution to database, which assigns the ID
		if err = sei.orchestrator.db.AddExecution(execution); err != nil {
			sei.entity.Status = StatusFailed
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.mu.Lock()
			sei.fsm.Fire(TriggerFail)
			sei.mu.Unlock()
			return nil
		}
		sei.entity.Executions = append(sei.entity.Executions, execution)
		executionID := execution.ID
		sei.executionID = executionID // Store execution ID
		sei.execution = execution

		// Now that we have executionID, we can create the hierarchy
		// But only if parentExecutionID is available (non-zero)
		if sei.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             sei.orchestrator.runID,
				ParentEntityID:    sei.parentEntityID,
				ChildEntityID:     sei.entity.ID,
				ParentExecutionID: sei.parentExecutionID,
				ChildExecutionID:  sei.executionID,
				ParentStepID:      sei.parentStepID,
				ChildStepID:       sei.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeSideEffect),
			}
			sei.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing side effect %s (Entity ID: %d, Execution ID: %d)", sei.stepID, sei.entity.ID, executionID)

		err := sei.runSideEffect(execution)
		if errors.Is(err, ErrPaused) {
			sei.entity.Status = StatusPaused
			sei.entity.Paused = true
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.mu.Lock()
			sei.fsm.Fire(TriggerPause)
			sei.mu.Unlock()
			return nil
		}
		if err == nil {
			// Success
			execution.Status = ExecutionStatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			sei.entity.Status = StatusCompleted
			sei.orchestrator.db.UpdateExecution(execution)
			sei.orchestrator.db.UpdateEntity(sei.entity)
			sei.mu.Lock()
			sei.fsm.Fire(TriggerComplete)
			sei.mu.Unlock()
			return nil
		} else {
			execution.Status = ExecutionStatusFailed
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			sei.err = err
			sei.orchestrator.db.UpdateExecution(execution)
			if attempt < maxAttempts {

				// Calculate next interval
				nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(attempt-1)))
				if nextInterval > maxInterval {
					nextInterval = maxInterval
				}
				log.Printf("Retrying side effect %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", sei.stepID, sei.entity.ID, executionID, attempt+1, maxAttempts, nextInterval)
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached
				sei.entity.Status = StatusFailed
				sei.orchestrator.db.UpdateEntity(sei.entity)
				sei.mu.Lock()
				sei.fsm.Fire(TriggerFail)
				sei.mu.Unlock()
				return nil
			}
		}
	}
	return nil
}

func (sei *SideEffectInstance) runSideEffect(execution *Execution) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d, Execution ID: %d) runSideEffect attempt %d", sei.stepID, sei.entity.ID, execution.ID, execution.Attempt)
	var err error

	// Check if result already exists in the database
	latestExecution, err := sei.orchestrator.db.GetLatestExecution(sei.entity.ID)
	if err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err
	}

	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted &&
		latestExecution.SideEffectExecutionData != nil &&
		latestExecution.SideEffectExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", sei.entity.ID)
		outputs, err := convertOutputsFromSerialization(
			HandlerInfo{ReturnTypes: sei.returnTypes},
			latestExecution.SideEffectExecutionData.Outputs,
		)
		if err != nil {
			log.Printf("Error deserializing side effect result: %v", err)
			return err
		}
		sei.mu.Lock()
		sei.results = outputs
		sei.mu.Unlock()
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in side effect: %v", r)
			log.Printf("Panic in side effect: %v", err)
			sei.mu.Lock()
			sei.err = err
			sei.mu.Unlock()
		}
	}()

	select {
	case <-sei.ctx.Done():
		log.Printf("Context cancelled in side effect")
		sei.mu.Lock()
		sei.err = sei.ctx.Err()
		sei.mu.Unlock()
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
		sei.mu.Lock()
		sei.err = err
		sei.mu.Unlock()
		return err
	}

	// Collect results without holding the lock
	outputs := make([]interface{}, numOut)
	for i := 0; i < numOut; i++ {
		result := results[i].Interface()
		log.Printf("Side effect returned result [%d]: %v", i, result)
		outputs[i] = result
	}

	// Now set results while holding the lock
	sei.mu.Lock()
	sei.results = outputs
	sei.mu.Unlock()

	// Serialize output
	outputBytes, err := convertOutputsForSerialization(sei.results)
	if err != nil {
		log.Printf("Error serializing output: %v", err)
		return err
	}

	// Create SideEffectExecutionData
	sideEffectExecutionData := &SideEffectExecutionData{
		Outputs: outputBytes,
	}

	// Update the execution with the execution data
	execution.SideEffectExecutionData = sideEffectExecutionData
	sei.orchestrator.db.UpdateExecution(execution)

	return nil
}

func (sei *SideEffectInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onCompleted called", sei.stepID, sei.entity.ID)
	if sei.future != nil {
		sei.mu.Lock()
		results := sei.results
		sei.mu.Unlock()
		sei.future.setResult(results)
	}
	return nil
}

func (sei *SideEffectInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onFailed called", sei.stepID, sei.entity.ID)
	if sei.future != nil {
		sei.mu.Lock()
		err := sei.err
		sei.mu.Unlock()
		sei.future.setError(err)
	}
	return nil
}

func (sei *SideEffectInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onPaused called", sei.stepID, sei.entity.ID)
	if sei.future != nil {
		sei.future.setError(ErrPaused)
	}
	return nil
}

/// FILE: ./states_triggers.go

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

/// FILE: ./tempolite.go

// Tempolite is the main API interface for the workflow engine
type Tempolite struct {
	mu       deadlock.RWMutex
	registry *Registry
	database Database
	queues   map[string]*QueueManager
	ctx      context.Context
	cancel   context.CancelFunc
	defaultQ string
}

// QueueConfig holds configuration for a queue
type QueueConfig struct {
	Name        string
	WorkerCount int
}

// TempoliteOption is a function type for configuring Tempolite
type TempoliteOption func(*Tempolite)

func New(ctx context.Context, db Database, options ...TempoliteOption) (*Tempolite, error) {
	ctx, cancel := context.WithCancel(ctx)
	t := &Tempolite{
		registry: NewRegistry(),
		database: db,
		queues:   make(map[string]*QueueManager),
		ctx:      ctx,
		cancel:   cancel,
		defaultQ: "default",
	}

	// Apply options BEFORE creating default queue
	for _, opt := range options {
		opt(t)
	}

	// Only create default queue if it doesn't exist
	t.mu.Lock()
	if _, exists := t.queues[t.defaultQ]; !exists {
		if err := t.createQueueLocked(QueueConfig{
			Name:        t.defaultQ,
			WorkerCount: 1,
		}); err != nil {
			if !errors.Is(err, ErrQueueExists) {
				t.mu.Unlock()
				return nil, err
			}
		}
	}
	t.mu.Unlock()

	return t, nil
}

// createQueueLocked creates a queue while holding the lock
func (t *Tempolite) createQueueLocked(config QueueConfig) error {

	if config.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be greater than 0")
	}
	var err error

	_, err = t.database.GetQueueByName(config.Name)
	if errors.Is(err, ErrQueueNotFound) {
		// Create queue in database
		queue := &Queue{
			Name: config.Name,
		}
		if err = t.database.AddQueue(queue); err != nil {
			return fmt.Errorf("failed to create queue in database: %w", err)
		}
	}

	if _, exists := t.queues[config.Name]; exists {
		return ErrQueueExists
	}

	// Create queue manager
	manager := newQueueManager(t.ctx, config.Name, config.WorkerCount, t.registry, t.database)
	t.queues[config.Name] = manager

	go manager.Start() // starts the pulling

	return nil
}

func (t *Tempolite) createQueue(config QueueConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.createQueueLocked(config)
}

// RegisterWorkflow registers a workflow with the registry
func (t *Tempolite) RegisterWorkflow(workflowFunc interface{}) error {
	_, err := t.registry.RegisterWorkflow(workflowFunc)
	return err
}

// Workflow executes a workflow on the default queue
func (t *Tempolite) Workflow(workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *DatabaseFuture {
	return t.executeWorkflow(t.defaultQ, workflowFunc, options, args...)
}

// executeWorkflow executes a workflow on a specific queue
func (t *Tempolite) executeWorkflow(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) *DatabaseFuture {

	// If options specify a different queue, use that
	if options != nil && options.Queue != "" {
		queueName = options.Queue
	}

	// Verify queue exists
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		f := NewDatabaseFuture(t.ctx, t.database)
		f.setError(fmt.Errorf("queue %s does not exist", queueName))
		return f
	}

	future := NewDatabaseFuture(t.ctx, t.database)

	id, err := qm.CreateWorkflow(workflowFunc, options, args...)
	if err != nil {
		future.setError(err)
		return future
	}

	// We don't manage execution on Tempolite layer since it has to go through a pulling
	future.setEntityID(id)

	return future
}

func (t *Tempolite) AddQueueWorkers(queueName string, count int) error {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.AddWorkers(count)
}

func (t *Tempolite) RemoveQueueWorkers(queueName string, count int) error {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.RemoveWorkers(count)
}

func (t *Tempolite) GetQueueInfo(queueName string) (*QueueInfo, error) {
	t.mu.RLock()
	qm, exists := t.queues[queueName]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("queue %s does not exist", queueName)
	}

	return qm.GetInfo(), nil
}

func (t *Tempolite) Wait() error {
	var err error
	shutdown := errgroup.Group{}

	t.mu.RLock()
	for _, qm := range t.queues {
		qm := qm
		shutdown.Go(func() error {
			fmt.Println("Waiting for queue", qm.name)
			defer fmt.Println("Queue", qm.name, "shutdown complete")
			// Ignore context.Canceled during wait
			if err := qm.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		})
	}
	t.mu.RUnlock()

	if err = shutdown.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return errors.Join(err, fmt.Errorf("failed to wait for queues"))
	}
	fmt.Println("All queues shutdown complete")

	select {
	case <-t.ctx.Done():
		if !errors.Is(t.ctx.Err(), context.Canceled) {
			return t.ctx.Err()
		}
	default:
	}

	return nil
}

func (t *Tempolite) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Start graceful shutdown
	var errs []error
	for name, qm := range t.queues {
		if err := qm.Close(); err != nil {
			// Only append non-context-canceled errors
			if !errors.Is(err, context.Canceled) {
				errs = append(errs, fmt.Errorf("failed to close queue %s: %w", name, err))
			}
		}
	}

	// Set up grace period in background
	done := make(chan struct{})
	go func() {
		defer close(done)

		shutdown := errgroup.Group{}
		// Wait for them to close
		for _, qm := range t.queues {
			qm := qm
			shutdown.Go(func() error {
				fmt.Println("Waiting for queue", qm.name)
				defer fmt.Println("Queue", qm.name, "shutdown complete")
				err := qm.Wait()
				if err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil
			})
		}

		if err := shutdown.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			errs = append(errs, fmt.Errorf("failed to wait for queues: %w", err))
		}
	}()

	// Wait for either grace period or clean shutdown
	select {
	case <-done:
		fmt.Println("Clean shutdown completed")
	case <-time.After(5 * time.Second):
		fmt.Println("Grace period expired, forcing shutdown")
		t.cancel()
	}

	// Only return error if we have non-context-canceled errors
	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}
	return nil
}

// GetQueues returns all available queue names
func (t *Tempolite) GetQueues() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	queues := make([]string, 0, len(t.queues))
	for name := range t.queues {
		queues = append(queues, name)
	}
	return queues
}

// WithQueue adds a new queue configuration
func WithQueue(config QueueConfig) TempoliteOption {
	return func(t *Tempolite) {
		if err := t.createQueue(config); err != nil {
			panic(fmt.Sprintf("failed to create queue %s: %v", config.Name, err))
		}
	}
}

// WithDefaultQueueWorkers sets the number of workers for the default queue
func WithDefaultQueueWorkers(count int) TempoliteOption {
	return func(t *Tempolite) {
		t.mu.Lock()
		defer t.mu.Unlock()

		// Remove existing default queue if it exists
		if existing, exists := t.queues[t.defaultQ]; exists {
			existing.Close()
			delete(t.queues, t.defaultQ)
		}

		// Create new default queue with specified worker count
		if err := t.createQueueLocked(QueueConfig{
			Name:        t.defaultQ,
			WorkerCount: count,
		}); err != nil {
			panic(fmt.Sprintf("failed to configure default queue: %v", err))
		}
	}
}

// min function used in checkAndProcessPending
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

/// FILE: ./types.go

func getInternalRetryPolicy(options *WorkflowOptions) *retryPolicyInternal {
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
		return &retryPolicyInternal{
			MaxAttempts:        rp.MaxAttempts,
			InitialInterval:    rp.InitialInterval.Nanoseconds(),
			BackoffCoefficient: rp.BackoffCoefficient,
			MaxInterval:        rp.MaxInterval.Nanoseconds(),
		}
	}
	// Default RetryPolicy
	return &retryPolicyInternal{
		MaxAttempts:        1,
		InitialInterval:    time.Second.Nanoseconds(),
		BackoffCoefficient: 2.0,
		MaxInterval:        (5 * time.Minute).Nanoseconds(),
	}
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

var DefaultVersion int = 0

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
	ID                int
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

// Run represents a workflow execution group
type Run struct {
	ID          int
	Status      string // "Pending", "Running", "Completed", etc.
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []*Entity
	Hierarchies []*Hierarchy
}

// EntityType defines the type of the entity
type EntityType string

const (
	EntityTypeWorkflow   EntityType = "Workflow"
	EntityTypeActivity   EntityType = "Activity"
	EntityTypeSaga       EntityType = "Saga"
	EntityTypeSideEffect EntityType = "SideEffect"
)

// EntityStatus defines the status of the entity
type EntityStatus string

const (
	StatusPending   EntityStatus = "Pending"
	StatusQueued    EntityStatus = "Queued"
	StatusRunning   EntityStatus = "Running"
	StatusPaused    EntityStatus = "Paused"
	StatusCancelled EntityStatus = "Cancelled"
	StatusCompleted EntityStatus = "Completed"
	StatusFailed    EntityStatus = "Failed"
)

// Entity represents an executable component
type Entity struct {
	ID             int
	HandlerName    string
	Type           EntityType
	Status         EntityStatus
	StepID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RunID          int
	Run            *Run
	Executions     []*Execution
	QueueID        int    // optional
	Queue          *Queue // optional
	Versions       []*Version
	WorkflowData   *WorkflowData   // optional
	ActivityData   *ActivityData   // optional
	SagaData       *SagaData       // optional
	SideEffectData *SideEffectData // optional
	HandlerInfo    *HandlerInfo    // handler metadata
	RetryState     *RetryState
	RetryPolicy    *retryPolicyInternal
	Paused         bool
	Resumable      bool
}

// ExecutionStatus defines the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "Pending"
	ExecutionStatusQueued    ExecutionStatus = "Queued"
	ExecutionStatusRunning   ExecutionStatus = "Running"
	ExecutionStatusRetried   ExecutionStatus = "Retried"
	ExecutionStatusPaused    ExecutionStatus = "Paused"
	ExecutionStatusCancelled ExecutionStatus = "Cancelled"
	ExecutionStatusCompleted ExecutionStatus = "Completed"
	ExecutionStatusFailed    ExecutionStatus = "Failed"
)

// Execution represents a single execution attempt
type Execution struct {
	ID                      int
	EntityID                int
	Entity                  *Entity
	StartedAt               time.Time
	CompletedAt             *time.Time
	Status                  ExecutionStatus
	Error                   string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	WorkflowExecutionData   *WorkflowExecutionData   // optional
	ActivityExecutionData   *ActivityExecutionData   // optional
	SagaExecutionData       *SagaExecutionData       // optional
	SideEffectExecutionData *SideEffectExecutionData // optional
	Attempt                 int
}

// Queue represents a work queue
type Queue struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*Entity
}

type ActivityData struct {
	Timeout      int64      `json:"timeout,omitempty"`
	MaxAttempts  int        `json:"max_attempts"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	Input        [][]byte   `json:"input,omitempty"`
	Output       [][]byte   `json:"output,omitempty"`
	Attempt      int        `json:"attempt"`
}

type SagaData struct {
	Compensating     bool     `json:"compensating"`
	CompensationData [][]byte `json:"compensation_data,omitempty"`
}

type SideEffectData struct {
	// No fields as per ent schema
}

type ActivityExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type SagaExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Output        [][]byte   `json:"output,omitempty"`
	HasOutput     bool       `json:"hasOutput"`
}

type SideEffectExecutionData struct {
	Outputs [][]byte `json:"outputs,omitempty"`
}

type RunInfo struct {
	RunID  int
	Status string
}

type WorkflowInfo struct {
	EntityID int
	Status   EntityStatus
	Run      *RunInfo
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

// ActivityOptions provides options for activities, including retry policies.
type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

var ErrPaused = errors.New("execution paused")

// ContinueAsNewError indicates that the workflow should restart with new inputs.
type ContinueAsNewError struct {
	Options *WorkflowOptions
	Args    []interface{}
}

func (e *ContinueAsNewError) Error() string {
	return "workflow is continuing as new"
}

/// FILE: ./workflows.go

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	mu                deadlock.Mutex // TODO: need to support mutex to fix all those data races
	stepID            string
	handler           HandlerInfo
	input             []interface{}
	results           []interface{}
	err               error
	fsm               *stateless.StateMachine
	future            Future
	ctx               context.Context
	orchestrator      *Orchestrator
	workflowID        int
	options           *WorkflowOptions
	entity            *Entity
	entityID          int
	executionID       int
	execution         *Execution // Current execution
	continueAsNew     *ContinueAsNewError
	parentExecutionID int
	parentEntityID    int
	parentStepID      string
}

func (wi *WorkflowInstance) MustState() any {
	wi.mu.Lock()
	defer wi.mu.Unlock()
	if wi.fsm == nil {
		return nil
	}
	return wi.fsm.MustState()
}

func (wi *WorkflowInstance) Start() {

	wi.mu.Lock()
	wi.fsm = stateless.NewStateMachine(StateIdle)
	wi.mu.Unlock()

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

	log.Printf("Starting workflow: %s (Entity ID: %d)", wi.stepID, wi.entity.ID)

	// Start the FSM without holding wi.mu
	go func() {
		if err := wi.fsm.Fire(TriggerStart); err != nil {
			log.Printf("Error starting workflow: %v", err)
		}
	}()
}

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) executeWorkflow called", wi.stepID, wi.entity.ID)
	return wi.executeWithRetry()
}

func (wi *WorkflowInstance) executeWithRetry() error {
	var err error
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration
	var backoffCoefficient float64
	var maxInterval time.Duration

	log.Printf("WorkflowInstance %s (Entity ID: %d) executeWithRetry called", wi.stepID, wi.entity.ID)

	if wi.entity.RetryPolicy != nil {
		rp := wi.entity.RetryPolicy
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

	// Somehow we need to also check if `wi.entity.Resumable` is true and allow to pass that for loop
	log.Printf("WorkflowInstance %s (Entity ID: %d) executeWithRetry maxAttempts: %d, initialInterval: %v, backoffCoefficient: %f, maxInterval: %v", wi.stepID, wi.entity.ID, maxAttempts, initialInterval, backoffCoefficient, maxInterval)
	attempt = wi.entity.RetryState.Attempts + 1

	for {
		if wi.orchestrator.IsPaused() {
			log.Printf("WorkflowInstance %s is paused", wi.stepID)
			wi.entity.Status = StatusPaused
			wi.entity.Paused = true
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerPause)
			return nil
		}

		// Check if maximum attempts have been reached and not resumable
		if attempt > maxAttempts && !wi.entity.Resumable {
			wi.entity.Status = StatusFailed
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerFail)
			log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) failed after %d attempts", wi.stepID, wi.entity.ID, wi.executionID, attempt-1)
			return nil
		}

		// Update RetryState without changing attempts values beyond maxAttempts
		if attempt <= maxAttempts {
			wi.entity.RetryState.Attempts = attempt
		}
		wi.orchestrator.db.UpdateEntity(wi.entity)

		// Create Execution
		execution := &Execution{
			EntityID:  wi.entity.ID,
			Status:    ExecutionStatusRunning,
			Attempt:   attempt,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entity:    wi.entity,
		}
		if err = wi.orchestrator.db.AddExecution(execution); err != nil {
			log.Printf("Error adding execution: %v", err)
			wi.entity.Status = StatusFailed
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerFail)
			return nil
		}
		wi.entity.Executions = append(wi.entity.Executions, execution)
		wi.executionID = execution.ID
		wi.execution = execution

		// Create hierarchy if parentExecutionID is available
		if wi.parentExecutionID != 0 {
			hierarchy := &Hierarchy{
				RunID:             wi.orchestrator.runID,
				ParentEntityID:    wi.parentEntityID,
				ChildEntityID:     wi.entity.ID,
				ParentExecutionID: wi.parentExecutionID,
				ChildExecutionID:  wi.executionID,
				ParentStepID:      wi.parentStepID,
				ChildStepID:       wi.stepID,
				ParentType:        string(EntityTypeWorkflow),
				ChildType:         string(EntityTypeWorkflow),
			}
			wi.orchestrator.db.AddHierarchy(hierarchy)
		}

		log.Printf("Executing workflow %s (Entity ID: %d, Execution ID: %d)", wi.stepID, wi.entity.ID, wi.executionID)

		err := wi.runWorkflow(execution)
		if errors.Is(err, ErrPaused) {
			wi.entity.Status = StatusPaused
			wi.entity.Paused = true
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerPause)
			log.Printf("WorkflowInstance %s is paused post-runWorkflow", wi.stepID)
			return nil
		}
		if err == nil {
			// Success
			execution.Status = ExecutionStatusCompleted
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			wi.entity.Status = StatusCompleted
			wi.orchestrator.db.UpdateExecution(execution)
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.fsm.Fire(TriggerComplete)
			log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) completed successfully", wi.stepID, wi.entity.ID, wi.executionID)
			return nil
		} else {
			execution.Status = ExecutionStatusFailed
			completedAt := time.Now()
			execution.CompletedAt = &completedAt
			execution.Error = err.Error()
			wi.err = err
			wi.orchestrator.db.UpdateExecution(execution)

			// Calculate next interval
			nextInterval := time.Duration(float64(initialInterval) * math.Pow(backoffCoefficient, float64(attempt-1)))
			if nextInterval > maxInterval {
				nextInterval = maxInterval
			}

			if attempt < maxAttempts || wi.entity.Resumable {
				log.Printf("Retrying workflow %s (Entity ID: %d, Execution ID: %d), attempt %d/%d after %v", wi.stepID, wi.entity.ID, wi.executionID, attempt+1, maxAttempts, nextInterval)
				time.Sleep(nextInterval)
			} else {
				// Max attempts reached and not resumable
				wi.entity.Status = StatusFailed
				wi.orchestrator.db.UpdateEntity(wi.entity)
				wi.fsm.Fire(TriggerFail)
				log.Printf("Workflow %s (Entity ID: %d, Execution ID: %d) failed after %d attempts", wi.stepID, wi.entity.ID, wi.executionID, attempt)
				return nil
			}
		}

		// Increment attempt only if less than maxAttempts
		if attempt < maxAttempts {
			attempt++
		}
	}
}

// Sub-function to run the workflow within retry loop
func (wi *WorkflowInstance) runWorkflow(execution *Execution) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d, Execution ID: %d) runWorkflow attempt %d", wi.stepID, wi.entity.ID, execution.ID, execution.Attempt)
	var err error
	// Check if result already exists in the database
	var latestExecution *Execution
	if latestExecution, err = wi.orchestrator.db.GetLatestExecution(wi.entity.ID); err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err
	}
	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", wi.entity.ID)
		outputs, err := convertOutputsFromSerialization(wi.handler, latestExecution.WorkflowExecutionData.Outputs)
		if err != nil {
			log.Printf("Error deserializing outputs: %v", err)
			return err
		}
		wi.results = outputs
		return nil
	}

	handler := wi.handler
	f := handler.Handler

	ctxWorkflow := WorkflowContext{
		orchestrator: wi.orchestrator,
		ctx:          wi.ctx,
		workflowID:   wi.entity.ID,
		stepID:       wi.stepID,
		options:      wi.options,
		executionID:  wi.executionID,
	}

	// Convert inputs from serialization
	inputs, err := convertInputsFromSerialization(handler, wi.entity.WorkflowData.Input)
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
			wi.err = fmt.Errorf("panic: %v", r)
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		wi.err = wi.ctx.Err()
		return wi.err
	case <-time.After(0):
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
			wi.mu.Lock()
			wi.err = errInterface.(error)
			wi.mu.Unlock()

			// Serialize error
			errorMessage := wi.err.Error()

			// Update execution error
			execution.Error = errorMessage

			// Create WorkflowExecutionData
			workflowExecutionData := &WorkflowExecutionData{
				Outputs: nil,
			}

			// Update the execution with the execution data
			execution.WorkflowExecutionData = workflowExecutionData
			wi.orchestrator.db.UpdateExecution(execution)

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

		// Create WorkflowExecutionData
		workflowExecutionData := &WorkflowExecutionData{
			Outputs: outputBytes,
		}

		// Update the execution with the execution data
		execution.WorkflowExecutionData = workflowExecutionData
		wi.orchestrator.db.UpdateExecution(execution)

		return nil
	}
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onCompleted called", wi.stepID, wi.entity.ID)
	var err error

	if wi.continueAsNew != nil {
		// Handle ContinueAsNew
		o := wi.orchestrator

		// Use the existing handler
		handler := wi.handler

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
		internalRetryPolicy := getInternalRetryPolicy(wi.continueAsNew.Options)

		// Create WorkflowData
		workflowData := &WorkflowData{
			Duration:  "",
			Paused:    false,
			Resumable: false,
			Input:     inputBytes,
			Attempt:   1,
		}

		// Create RetryState
		retryState := &RetryState{Attempts: 0}

		// Create a new Entity
		newEntity := &Entity{
			StepID:       wi.entity.StepID, // Use the same stepID
			HandlerName:  handler.HandlerName,
			Status:       StatusPending,
			Type:         EntityTypeWorkflow,
			RunID:        wi.entity.RunID, // Share the same RunID
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			WorkflowData: workflowData,
			RetryPolicy:  internalRetryPolicy,
			RetryState:   retryState,
			HandlerInfo:  &handler,
			Paused:       false,
			Resumable:    false,
		}
		// Add the entity to the database
		if err = o.db.AddEntity(newEntity); err != nil {
			return err
		}

		// Create a new WorkflowInstance
		newInstance := &WorkflowInstance{
			stepID:            wi.stepID,
			handler:           handler,
			input:             wi.continueAsNew.Args,
			ctx:               o.ctx,
			orchestrator:      o,
			workflowID:        newEntity.ID,
			options:           wi.continueAsNew.Options,
			entity:            newEntity,
			entityID:          newEntity.ID,
			parentExecutionID: wi.parentExecutionID,
			parentEntityID:    wi.parentEntityID,
			parentStepID:      wi.parentStepID,
		}

		// Start the new workflow instance to initialize fsm
		newInstance.Start()

		// Determine if this is the root workflow
		o.instancesMu.Lock()
		isRootWorkflow := (wi == o.rootWf)
		if isRootWorkflow {
			// Update rootWf
			o.rootWf = newInstance
		}
		o.instancesMu.Unlock()

		// Add the new instance
		o.addWorkflowInstance(newInstance)

		if isRootWorkflow {
			// Complete the future immediately
			if wi.future != nil {
				wi.future.setResult(wi.results)
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

		// Check if this is the root workflow
		wi.orchestrator.instancesMu.RLock()
		isRootWorkflow := (wi == wi.orchestrator.rootWf)
		wi.orchestrator.instancesMu.RUnlock()

		// If this is the root workflow, update the Run status to Completed
		if isRootWorkflow {
			// Update the run status
			var run *Run
			if run, err = wi.orchestrator.db.GetRun(wi.orchestrator.runID); err != nil {
				log.Printf("Error getting run: %v", err)
				wi.err = err
				if wi.future != nil {
					wi.future.setError(wi.err)
				}
				return nil
			}

			run.Status = string(StatusCompleted)
			run.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateRun(run)

			wi.orchestrator.db.UpdateEntityStatus(wi.entity.ID, StatusCompleted)
			wi.execution.Status = ExecutionStatusCompleted
			wi.execution.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateExecution(wi.execution)
			fmt.Println("Entity", wi.entity.ID, "completed")
			fmt.Println("Execution", wi.execution.ID, "completed")
		}
	}
	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.entity.ID)
	var err error
	if wi.future != nil {
		wi.mu.Lock()
		errCopy := wi.err
		wi.mu.Unlock()
		wi.future.setError(errCopy)
	}
	// If this is the root workflow, update the Run status to Failed
	if wi.orchestrator.rootWf == wi {
		var run *Run
		if run, err = wi.orchestrator.db.GetRun(wi.orchestrator.runID); err != nil {
			log.Printf("Error getting run: %v", err)
			return err
		}

		run.Status = string(StatusFailed)
		run.UpdatedAt = time.Now()
		wi.orchestrator.db.UpdateRun(run)
	}
	return nil
}

func (wi *WorkflowInstance) onPaused(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onPaused called", wi.stepID, wi.entity.ID)
	if wi.future != nil {
		wi.future.setError(ErrPaused)
	}
	return nil
}

/// FILE: ./database.go

var ErrQueueExists = errors.New("queue already exists")
var ErrQueueNotFound = errors.New("queue not found")

// Database interface defines methods for interacting with the data store.
// All methods return errors to handle future implementations that might have errors.
type Database interface {
	// Run methods
	AddRun(run *Run) error
	GetRun(id int) (*Run, error)
	UpdateRun(run *Run) error

	// Version methods
	GetVersion(entityID int, changeID string) (*Version, error)
	SetVersion(version *Version) error

	// Hierarchy methods
	AddHierarchy(hierarchy *Hierarchy) error
	GetHierarchy(parentID, childID int) (*Hierarchy, error)
	GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error)

	// Entity methods
	AddEntity(entity *Entity) error
	HasEntity(id int) (bool, error)
	GetEntity(id int) (*Entity, error)
	UpdateEntity(entity *Entity) error
	GetEntityStatus(id int) (EntityStatus, error)
	UpdateEntityStatus(id int, status EntityStatus) error
	GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) (*Entity, error)
	GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) (*Entity, error)
	FindPendingWorkflowsByQueue(queueID int) ([]*Entity, error)

	// Execution methods
	AddExecution(execution *Execution) error
	GetExecution(id int) (*Execution, error)
	UpdateExecution(execution *Execution) error
	GetLatestExecution(entityID int) (*Execution, error)

	// Queue methods
	AddQueue(queue *Queue) error
	GetQueue(id int) (*Queue, error)
	GetQueueByName(name string) (*Queue, error)
	UpdateQueue(queue *Queue) error
	ListQueues() ([]*Queue, error)

	// Clear removes all Runs that are 'Completed' and their associated data.
	Clear() error
}

/// FILE: ./database_memory.go

// DefaultDatabase is an in-memory implementation of Database.
type DefaultDatabase struct {
	runs        map[int]*Run
	versions    map[int]*Version
	hierarchies map[int]*Hierarchy
	entities    map[int]*Entity
	executions  map[int]*Execution
	queues      map[int]*Queue
	mu          deadlock.RWMutex
}

func NewDefaultDatabase() *DefaultDatabase {
	db := &DefaultDatabase{
		runs:        make(map[int]*Run),
		versions:    make(map[int]*Version),
		hierarchies: make(map[int]*Hierarchy),
		entities:    make(map[int]*Entity),
		executions:  make(map[int]*Execution),
		queues:      make(map[int]*Queue),
	}
	db.queues[1] = &Queue{
		ID:        1,
		Name:      "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entities:  []*Entity{},
	}
	return db
}

var (
	runIDCounter       int64
	versionIDCounter   int64
	entityIDCounter    int64
	executionIDCounter int64
	hierarchyIDCounter int64
	queueIDCounter     int64 = 1 // Starting from 1 for the default queue.
)

// Run methods
func (db *DefaultDatabase) AddRun(run *Run) error {
	runID := int(atomic.AddInt64(&runIDCounter, 1))
	run.ID = runID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.runs[run.ID] = run
	return nil
}

func (db *DefaultDatabase) GetRun(id int) (*Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	run, exists := db.runs[id]
	if !exists {
		return nil, errors.New("run not found")
	}

	return db.copyRun(run), nil
}

func (db *DefaultDatabase) UpdateRun(run *Run) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.runs[run.ID]; !exists {
		return errors.New("run not found")
	}

	db.runs[run.ID] = run
	return nil
}

func (db *DefaultDatabase) copyRun(run *Run) *Run {
	copyRun := *run
	copyRun.Entities = make([]*Entity, len(run.Entities))
	copy(copyRun.Entities, run.Entities)
	return &copyRun
}

// Version methods
func (db *DefaultDatabase) GetVersion(entityID int, changeID string) (*Version, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, version := range db.versions {
		if version.EntityID == entityID && version.ChangeID == changeID {
			return version, nil
		}
	}
	return nil, errors.New("version not found")
}

func (db *DefaultDatabase) SetVersion(version *Version) error {
	versionID := int(atomic.AddInt64(&versionIDCounter, 1))
	version.ID = versionID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.versions[version.ID] = version
	return nil
}

// Hierarchy methods
func (db *DefaultDatabase) AddHierarchy(hierarchy *Hierarchy) error {
	hierarchyID := int(atomic.AddInt64(&hierarchyIDCounter, 1))
	hierarchy.ID = hierarchyID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.hierarchies[hierarchy.ID] = hierarchy
	return nil
}

func (db *DefaultDatabase) GetHierarchy(parentID, childID int) (*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, h := range db.hierarchies {
		if h.ParentEntityID == parentID && h.ChildEntityID == childID {
			return h, nil
		}
	}
	return nil, errors.New("hierarchy not found")
}

func (db *DefaultDatabase) GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []*Hierarchy
	for _, h := range db.hierarchies {
		if h.ChildEntityID == childEntityID {
			result = append(result, h)
		}
	}
	return result, nil
}

// Entity methods
func (db *DefaultDatabase) AddEntity(entity *Entity) error {
	entityID := int(atomic.AddInt64(&entityIDCounter, 1))
	entity.ID = entityID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.entities[entity.ID] = entity

	// Add the entity to its Run
	run, exists := db.runs[entity.RunID]
	if exists {
		run.Entities = append(run.Entities, entity)
		db.runs[entity.RunID] = run
	}

	return nil
}

var ErrEntityNotFound = errors.New("entity not found")

func (db *DefaultDatabase) GetEntity(id int) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	entity, exists := db.entities[id]
	if !exists {
		return nil, errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	copy := *entity
	return &copy, nil
}

func (db *DefaultDatabase) HasEntity(id int) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, exists := db.entities[id]
	return exists, nil
}

func (db *DefaultDatabase) UpdateEntity(entity *Entity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.entities[entity.ID]; !exists {
		return errors.Join(fmt.Errorf("entity %d", entity.ID), ErrEntityNotFound)
	}

	db.entities[entity.ID] = entity

	// Update the entity in its Run's Entities slice
	run, exists := db.runs[entity.RunID]
	if exists {
		for i, e := range run.Entities {
			if e.ID == entity.ID {
				run.Entities[i] = entity
				db.runs[run.ID] = run
				break
			}
		}
	}

	return nil
}

func (db *DefaultDatabase) GetEntityStatus(id int) (EntityStatus, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	entity, exists := db.entities[id]
	if !exists {
		return "", errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	return entity.Status, nil
}

func (db *DefaultDatabase) UpdateEntityStatus(id int, status EntityStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	entity, exists := db.entities[id]
	if !exists {
		return errors.Join(fmt.Errorf("entity %d", id), ErrEntityNotFound)
	}
	entity.Status = status
	entity.UpdatedAt = time.Now()
	return nil
}

func (db *DefaultDatabase) GetEntityByWorkflowIDAndStepID(workflowID int, stepID string) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, entity := range db.entities {
		if entity.RunID == workflowID && entity.StepID == stepID {
			return entity, nil
		}
	}
	return nil, errors.Join(fmt.Errorf(
		"entity with workflow ID %d and step ID %s", workflowID, stepID,
	), ErrEntityNotFound)
}

func (db *DefaultDatabase) GetChildEntityByParentEntityIDAndStepIDAndType(parentEntityID int, stepID string, entityType EntityType) (*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, hierarchy := range db.hierarchies {
		if hierarchy.ParentEntityID == parentEntityID && hierarchy.ChildStepID == stepID {
			childEntityID := hierarchy.ChildEntityID
			if entity, exists := db.entities[childEntityID]; exists && entity.Type == entityType {
				return entity, nil
			}
		}
	}
	return nil, errors.Join(fmt.Errorf(
		"child entity with parent entity ID %d, step ID %s, and type %s", parentEntityID, stepID, entityType,
	), ErrEntityNotFound)
}

func (db *DefaultDatabase) FindPendingWorkflowsByQueue(queueID int) ([]*Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []*Entity
	for _, entity := range db.entities {
		if entity.Type == EntityTypeWorkflow &&
			entity.Status == StatusPending &&
			entity.QueueID == queueID {
			result = append(result, entity)
		}
	}

	// Sort by creation time
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

// Execution methods
func (db *DefaultDatabase) AddExecution(execution *Execution) error {
	executionID := int(atomic.AddInt64(&executionIDCounter, 1))
	execution.ID = executionID

	db.mu.Lock()
	defer db.mu.Unlock()

	db.executions[execution.ID] = execution
	return nil
}

func (db *DefaultDatabase) GetExecution(id int) (*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	execution, exists := db.executions[id]
	if !exists {
		return nil, errors.New("execution not found")
	}
	return execution, nil
}

func (db *DefaultDatabase) UpdateExecution(execution *Execution) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.executions[execution.ID]; !exists {
		return errors.New("execution not found")
	}

	db.executions[execution.ID] = execution
	return nil
}

func (db *DefaultDatabase) GetLatestExecution(entityID int) (*Execution, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var latestExecution *Execution
	for _, execution := range db.executions {
		if execution.EntityID == entityID {
			if latestExecution == nil || execution.ID > latestExecution.ID {
				latestExecution = execution
			}
		}
	}
	if latestExecution != nil {
		return latestExecution, nil
	}
	return nil, errors.New("execution not found")
}

// Queue methods
func (db *DefaultDatabase) AddQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if queue with same name exists
	for _, q := range db.queues {
		if q.Name == queue.Name {
			return errors.New("queue already exists")
		}
	}

	queueID := int(atomic.AddInt64(&queueIDCounter, 1))
	queue.ID = queueID
	queue.CreatedAt = time.Now()
	queue.UpdatedAt = time.Now()

	db.queues[queue.ID] = queue
	return nil
}

func (db *DefaultDatabase) GetQueue(id int) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queue, exists := db.queues[id]
	if !exists {
		return nil, errors.Join(ErrQueueNotFound)
	}
	copy := *queue
	return &copy, nil
}

func (db *DefaultDatabase) GetQueueByName(name string) (*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, queue := range db.queues {
		if queue.Name == name {
			return queue, nil
		}
	}
	return nil, errors.Join(ErrQueueNotFound)
}

func (db *DefaultDatabase) UpdateQueue(queue *Queue) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if existing, exists := db.queues[queue.ID]; exists {
		existing.UpdatedAt = time.Now()
		existing.Name = queue.Name
		existing.Entities = queue.Entities
		db.queues[queue.ID] = existing
		return nil
	}
	return errors.Join(ErrQueueNotFound)
}

func (db *DefaultDatabase) ListQueues() ([]*Queue, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	queues := make([]*Queue, 0, len(db.queues))
	for _, q := range db.queues {
		queues = append(queues, q)
	}
	return queues, nil
}

// Clear removes all Runs that are 'Completed' or 'Failed' and their associated data.
func (db *DefaultDatabase) Clear() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	runsToDelete := []int{}
	entitiesToDelete := map[int]bool{}
	executionsToDelete := map[int]bool{}
	hierarchiesToKeep := map[int]*Hierarchy{}
	versionsToDelete := map[int]*Version{}

	// Find Runs to delete
	for runID, run := range db.runs {
		if run.Status == string(StatusCompleted) || run.Status == string(StatusFailed) {
			runsToDelete = append(runsToDelete, runID)
			// Collect Entities associated with the Run
			for _, entity := range run.Entities {
				entitiesToDelete[entity.ID] = true
			}
		}
	}

	// Collect Executions associated with Entities to delete
	for execID, execution := range db.executions {
		if _, exists := entitiesToDelete[execution.EntityID]; exists {
			executionsToDelete[execID] = true
		}
	}

	// Collect Versions associated with Entities to delete
	for versionID, version := range db.versions {
		if _, exists := entitiesToDelete[version.EntityID]; exists {
			versionsToDelete[versionID] = version
		}
	}

	// Filter Hierarchies to keep only those not associated with Entities to delete
	for hid, hierarchy := range db.hierarchies {
		if _, parentExists := entitiesToDelete[hierarchy.ParentEntityID]; parentExists {
			continue
		}
		if _, childExists := entitiesToDelete[hierarchy.ChildEntityID]; childExists {
			continue
		}
		hierarchiesToKeep[hid] = hierarchy
	}

	// Delete Runs
	for _, runID := range runsToDelete {
		delete(db.runs, runID)
	}

	// Delete Entities and Executions
	for entityID := range entitiesToDelete {
		delete(db.entities, entityID)
	}
	for execID := range executionsToDelete {
		delete(db.executions, execID)
	}

	// Delete Versions
	for versionID := range versionsToDelete {
		delete(db.versions, versionID)
	}

	// Replace hierarchies with the filtered ones
	db.hierarchies = hierarchiesToKeep

	return nil
}
