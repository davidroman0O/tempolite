package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
)

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
	mu                sync.Mutex
	executionID       int
	execution         *Execution
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
	execution := &Execution{
		EntityID:  si.entity.ID,
		Attempt:   1,
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Add execution to database, which assigns the ID
	execution = si.orchestrator.db.AddExecution(execution)
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

			// Update execution error
			si.execution.Error = si.err.Error()
			si.orchestrator.db.UpdateExecution(si.execution)

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
	parentEntity := si.orchestrator.db.GetEntity(si.workflowID)
	if parentEntity != nil {
		parentEntity.Status = StatusFailed
		si.orchestrator.db.UpdateEntity(parentEntity)
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
