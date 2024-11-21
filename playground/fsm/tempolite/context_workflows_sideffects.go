package tempolite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

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
	// if options != nil && options.RetryPolicy != nil {
	// 	rp := options.RetryPolicy
	// 	// Fill defaults where zero
	// 	if rp.MaxAttempts == 0 {
	// 		rp.MaxAttempts = 1
	// 	}
	// 	if rp.InitialInterval == 0 {
	// 		rp.InitialInterval = time.Second
	// 	}
	// 	if rp.BackoffCoefficient == 0 {
	// 		rp.BackoffCoefficient = 2.0
	// 	}
	// 	if rp.MaxInterval == 0 {
	// 		rp.MaxInterval = 5 * time.Minute
	// 	}
	// 	retryPolicy = rp
	// } else {
	// Default retry policy with MaxAttempts=1
	retryPolicy = &RetryPolicy{
		MaxAttempts:        1,
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaxInterval:        5 * time.Minute,
	}
	// }

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
		stepID:         stepID,
		sideEffectFunc: sideEffectFunc,
		future:         future,
		ctx:            ctx.ctx,
		orchestrator:   ctx.orchestrator,
		workflowID:     ctx.workflowID,
		entityID:       entity.ID,
		options:        nil,
		// options:           options,
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
