package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

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
func (ctx WorkflowContext) GetVersion(changeID string, minSupported, maxSupported int) int {
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
func (ctx WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Workflow paused at stepID: %s", stepID)
		future := NewDatabaseFuture(ctx.orchestrator.ctx, ctx.orchestrator.db)
		future.setEntityID(ctx.workflowID)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepIDAndType(ctx.workflowID, stepID, EntityTypeWorkflow)
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
		latestExecution := ctx.orchestrator.db.GetLatestExecution(entity.ID)
		if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.WorkflowExecutionData != nil && latestExecution.WorkflowExecutionData.Outputs != nil {
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
		if latestExecution != nil && latestExecution.Status == ExecutionStatusFailed && latestExecution.Error != "" {
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
		queue := ctx.orchestrator.db.GetQueueByName(options.Queue)
		if queue == nil {
			err := fmt.Errorf("queue %s not found", options.Queue)
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
		entity = ctx.orchestrator.db.AddEntity(entity)
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
	entity = ctx.orchestrator.db.AddEntity(entity)
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
