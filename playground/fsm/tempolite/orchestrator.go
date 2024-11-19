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
		run = o.db.AddRun(run)
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

	// Create Entity
	entity := &Entity{
		StepID:      "root",
		HandlerName: handler.HandlerName,
		Status:      StatusPending,
		Type:        EntityTypeWorkflow,
		RunID:       o.runID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
	entity = o.db.AddEntity(entity)

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
	entity := o.db.GetEntity(entityID)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %d", entityID)
	}

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
	hierarchies := o.db.GetHierarchiesByChildEntity(entityID)
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
	instance.Start()

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

	// Resuming isn't retrying! We need to create a new execution to the entity while avoiding increasing the attempt count.

	future := NewRuntimeFuture()

	o.pausedMu.Lock()
	o.paused = false
	o.pausedMu.Unlock()

	// Create new context
	ctx, cancel := context.WithCancel(context.Background())
	o.ctx = ctx
	o.cancel = cancel

	// Retrieve the workflow entity
	entity := o.db.GetEntity(entityID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", entityID)
		future.setError(fmt.Errorf("no workflow found with ID: %d", entityID))
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
	run := o.db.GetRun(o.runID)
	if run != nil {
		run.Status = string(StatusPending)
		run.UpdatedAt = time.Now()
		o.db.UpdateRun(run)
	}

	// Get parent execution if this is a sub-workflow
	var parentExecID, parentEntityID int
	hierarchies := o.db.GetHierarchiesByChildEntity(entityID)
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
		stepID:       entity.StepID,
		handler:      handler,
		input:        inputs,
		ctx:          ctx,
		orchestrator: o,
		workflowID:   entity.ID,
		entity:       entity,
		// TODO: might need to find them back tho
		options:           nil,
		entityID:          entity.ID,
		parentExecutionID: parentExecID,
		parentEntityID:    parentEntityID,
		parentStepID:      entity.StepID,
	}

	future.setEntityID(entity.ID)
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

func (o *Orchestrator) Wait() error {
	log.Printf("Orchestrator.Wait called")

	if o.rootWf == nil {
		log.Printf("No root workflow to execute")
		return fmt.Errorf("no root workflow to execute")
	}

	lastLogTime := time.Now()
	// Wait for root workflow to complete
	for {
		// not initialized yet
		if o.rootWf.fsm == nil {
			continue
		}
		state := o.rootWf.fsm.MustState()
		// Log the state every 500ms
		currentTime := time.Now()
		if currentTime.Sub(lastLogTime) >= 500*time.Millisecond {
			log.Printf("Root Workflow FSM state: %s", state)
			lastLogTime = currentTime
		}
		if state == StateCompleted || state == StateFailed {
			break
		}
		select {
		case <-o.ctx.Done():
			log.Printf("Context cancelled: %v", o.ctx.Err())
			o.rootWf.fsm.Fire(TriggerFail)
			return o.ctx.Err()
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
		return fmt.Errorf("could not find root workflow entity")
	}

	latestExecution := o.db.GetLatestExecution(entity.ID)
	if latestExecution == nil {
		log.Printf("Could not find latest execution for root workflow")
		return fmt.Errorf("could not find latest execution for root workflow")
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
	return nil
}

// Retry retries a failed root workflow by creating a new entity and execution.
func (o *Orchestrator) Retry(workflowID int) *RuntimeFuture {

	future := NewRuntimeFuture()
	future.setEntityID(workflowID)

	// Retrieve the workflow entity
	entity := o.db.GetEntity(workflowID)
	if entity == nil {
		log.Printf("No workflow found with ID: %d", workflowID)
		future.setError(fmt.Errorf("no workflow found with ID: %d", workflowID))
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

func (o *Orchestrator) GetWorkflow(id int) (*WorkflowInfo, error) {
	entity := o.db.GetEntity(id)
	if entity == nil {
		return nil, fmt.Errorf("workflow entity not found with ID: %d", id)
	}

	run := o.db.GetRun(entity.RunID)

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
// In orchestrator.go
func (o *Orchestrator) WaitForContinuations(originalID int) error {
	log.Printf("Waiting for all continuations starting from workflow %d", originalID)

	currentID := originalID

	for {
		// Wait for current workflow
		state := o.rootWf.fsm.MustState()
		for state != StateCompleted && state != StateFailed {
			select {
			case <-o.ctx.Done():
				return o.ctx.Err()
			default:
				runtime.Gosched()
			}
			state = o.rootWf.fsm.MustState()
		}

		// Check if this workflow initiated a continuation
		entity := o.db.GetEntity(currentID)
		if entity == nil {
			return fmt.Errorf("entity not found: %d", currentID)
		}

		latestExecution := o.db.GetLatestExecution(currentID)
		if latestExecution == nil {
			return fmt.Errorf("no execution found for entity: %d", currentID)
		}

		if latestExecution.Status == ExecutionStatusCompleted {
			// Look for any child workflow that was created via ContinueAsNew
			children := o.db.GetChildEntityByParentEntityIDAndStepIDAndType(currentID, "root", EntityTypeWorkflow)
			if children == nil {
				// No more continuations, we're done
				log.Printf("No more continuations found after workflow %d", currentID)
				return nil
			}
			currentID = children.ID
			log.Printf("Found continuation: workflow %d", currentID)
		} else if latestExecution.Status == ExecutionStatusFailed {
			return fmt.Errorf("workflow %d failed", currentID)
		}
	}
}
