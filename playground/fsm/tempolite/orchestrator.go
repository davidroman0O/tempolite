package tempolite

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

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
		Duration:  "",
		Paused:    false,
		Resumable: false,
		Input:     inputBytes,
		Attempt:   1,
	}

	// Create RetryState
	retryState := &RetryState{Attempts: 0}

	// Create a new Entity without ID (database assigns it)
	entity := &Entity{
		StepID:       "root",
		HandlerName:  handler.HandlerName,
		Status:       StatusPending,
		Type:         EntityTypeWorkflow,
		RunID:        o.runID,
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
	entity = o.db.AddEntity(entity)

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:            "root",
		handler:           handler,
		input:             args,
		ctx:               o.ctx,
		orchestrator:      o,
		workflowID:        entity.ID,
		options:           options,
		entity:            entity,
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
		entity:         entity,
		options:        nil, // Adjust options as needed
		entityID:       entity.ID,
		parentEntityID: 0,
		parentStepID:   "",
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
			// tbf i don't know yet which one
			// time.Sleep(5 * time.Millisecond)
			runtime.Gosched()
		}
	}

	// Root workflow has completed or failed
	// The Run's status should have been updated in onCompleted or onFailed of the root workflow

	// Get final status from the database instead of relying on error field
	entity := o.db.GetEntity(o.rootWf.entityID)
	if entity == nil {
		log.Printf("Could not find root workflow entity")
		return
	}

	latestExecution := o.db.GetLatestExecution(entity.ID)
	if latestExecution == nil {
		log.Printf("Could not find latest execution for root workflow")
		return
	}

	switch entity.Status {
	case StatusCompleted:
		if o.rootWf.results != nil && len(o.rootWf.results) > 0 {
			fmt.Printf("Root workflow completed successfully with results: %v\n", o.rootWf.results)
		} else {
			fmt.Printf("Root workflow completed successfully\n")
		}
	case StatusFailed:
		var errMsg string
		if latestExecution.Error != "" {
			errMsg = latestExecution.Error
		} else if o.rootWf.err != nil {
			errMsg = o.rootWf.err.Error()
		} else {
			errMsg = "unknown error"
		}
		fmt.Printf("Root workflow failed with error: %v\n", errMsg)
	default:
		fmt.Printf("Root workflow ended with status: %s\n", entity.Status)
	}
}

// Retry retries a failed root workflow by creating a new entity and execution.
func (o *Orchestrator) Retry(workflowID int) *Future {
	// Retrieve the workflow entity
	entity := o.db.GetEntity(workflowID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", workflowID)
		return NewFuture(0)
	}

	// Set the runID from the entity
	o.runID = entity.RunID

	// Copy inputs
	inputs, err := convertInputsFromSerialization(*entity.HandlerInfo, entity.WorkflowData.Input)
	if err != nil {
		log.Printf("Error converting inputs from serialization: %v", err)
		return NewFuture(0)
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
	newEntity = o.db.AddEntity(newEntity)

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

	future := NewFuture(newEntity.ID)
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
	entity = o.db.AddEntity(entity)

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
