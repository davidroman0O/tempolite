package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
)

// WorkflowInstance represents an instance of a workflow execution.
type WorkflowInstance struct {
	mu                sync.Mutex // TODO: need to support mutex to fix all those data races
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
	return wi.fsm.MustState()
}

func (wi *WorkflowInstance) Start() {
	wi.mu.Lock()
	defer wi.mu.Unlock()

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

	fmt.Println("Starting workflow: ", wi.stepID, wi.entity.ID, wi.entity.Status)

	log.Printf("Starting workflow: %s (Entity ID: %d)", wi.stepID, wi.entity.ID)
	if err := wi.fsm.Fire(TriggerStart); err != nil {
		log.Printf("Error starting workflow: %v", err)
	}
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
			wi.err = errInterface.(error)

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

		wi.results = outputs

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

		// Use the existing handler instead of registering a new one
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
			Duration:  "",
			Paused:    false,
			Resumable: false,
			Input:     inputBytes,
			Attempt:   1,
		}

		// Create RetryState
		retryState := &RetryState{Attempts: 0}

		// Create a new Entity without ID (database assigns it)
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
		// Add the entity to the database, which assigns the ID
		if err = o.db.AddEntity(newEntity); err != nil {
			return err
		}

		// Create a new WorkflowInstance
		newInstance := &WorkflowInstance{
			stepID:            wi.stepID,
			handler:           handler, // Use existing handler
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

			wi.entity.Status = StatusCompleted
			wi.entity.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateEntity(wi.entity)
			wi.execution.Status = ExecutionStatusCompleted
			wi.execution.UpdatedAt = time.Now()
			wi.orchestrator.db.UpdateExecution(wi.execution)
			fmt.Println("Entity ", wi.entity.ID, " completed")
			fmt.Println("Execution ", wi.execution.ID, " completed")
		}
	}
	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s (Entity ID: %d) onFailed called", wi.stepID, wi.entity.ID)
	var err error
	if wi.future != nil {
		wi.future.setError(wi.err)
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
