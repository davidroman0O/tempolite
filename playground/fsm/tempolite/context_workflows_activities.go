package tempolite

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// Activity creates an activity.
func (ctx WorkflowContext) Activity(stepID string, activityFunc interface{}, options *ActivityOptions, args ...interface{}) *Future {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Activity paused at stepID: %s", stepID)
		future := NewFuture(0)
		future.setError(err)
		return future
	}

	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewFuture(0)

	// Check if result already exists in the database
	entity := ctx.orchestrator.db.GetChildEntityByParentEntityIDAndStepIDAndType(ctx.workflowID, stepID, EntityTypeActivity)
	if entity != nil {
		handlerInfo := entity.HandlerInfo
		if handlerInfo == nil {
			err := fmt.Errorf("handler not found for activity: %s", entity.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			return future
		}
		latestExecution := ctx.orchestrator.db.GetLatestExecution(entity.ID)
		if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Outputs != nil {
			// Deserialize output
			outputs, err := convertOutputsFromSerialization(*handlerInfo, latestExecution.ActivityExecutionData.Outputs)
			if err != nil {
				log.Printf("Error deserializing outputs: %v", err)
				future.setError(err)
				return future
			}
			future.setResult(outputs)
			return future
		}
		if latestExecution != nil && latestExecution.Status == ExecutionStatusFailed && latestExecution.Error != "" {
			log.Printf("Activity %s has failed execution with error: %s", stepID, latestExecution.Error)
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
	entity = ctx.orchestrator.db.AddEntity(entity)
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
