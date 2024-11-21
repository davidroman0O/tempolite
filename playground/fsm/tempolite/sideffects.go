package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"time"

	"github.com/qmuntal/stateless"
)

// SideEffectInstance represents an instance of a side effect execution.
type SideEffectInstance struct {
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
			sei.fsm.Fire(TriggerPause)
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
			sei.fsm.Fire(TriggerFail)
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
			sei.fsm.Fire(TriggerPause)
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
			sei.fsm.Fire(TriggerComplete)
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
				sei.fsm.Fire(TriggerFail)
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
	var latestExecution *Execution
	if latestExecution, err = sei.orchestrator.db.GetLatestExecution(sei.entity.ID); err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err
	}

	if latestExecution.Status == ExecutionStatusCompleted && latestExecution.SideEffectExecutionData != nil && latestExecution.SideEffectExecutionData.Outputs != nil {
		log.Printf("Result found in database for entity ID: %d", sei.entity.ID)
		outputs, err := convertOutputsFromSerialization(HandlerInfo{ReturnTypes: sei.returnTypes}, latestExecution.SideEffectExecutionData.Outputs)
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
		sei.future.setResult(sei.results)
	}
	return nil
}

func (sei *SideEffectInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("SideEffectInstance %s (Entity ID: %d) onFailed called", sei.stepID, sei.entity.ID)
	if sei.future != nil {
		sei.future.setError(sei.err)
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
