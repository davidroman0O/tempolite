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
			ai.fsm.Fire(TriggerPause)
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
			ai.fsm.Fire(TriggerFail)
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
			ai.fsm.Fire(TriggerPause)
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
			ai.fsm.Fire(TriggerComplete)
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
				ai.fsm.Fire(TriggerFail)
				return nil
			}
		}
	}
	return nil
}

// Sub-function within executeWithRetry
func (ai *ActivityInstance) runActivity(execution *Execution) error {
	log.Printf("ActivityInstance %s (Entity ID: %d, Execution ID: %d) runActivity attempt %d", ai.stepID, ai.entity.ID, execution.ID, execution.Attempt)
	var err error
	// Check if result already exists in the database
	var latestExecution *Execution
	if latestExecution, err = ai.orchestrator.db.GetLatestExecution(ai.entity.ID); err != nil {
		log.Printf("Error getting latest execution: %v", err)
		return err
	}

	if latestExecution != nil && latestExecution.Status == ExecutionStatusCompleted && latestExecution.ActivityExecutionData != nil && latestExecution.ActivityExecutionData.Outputs != nil {
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
		ai.future.setResult(ai.results)
	}
	return nil
}

func (ai *ActivityInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("ActivityInstance %s (Entity ID: %d) onFailed called", ai.stepID, ai.entity.ID)
	if ai.future != nil {
		ai.future.setError(ai.err)
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
