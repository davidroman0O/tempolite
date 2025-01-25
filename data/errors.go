package data

import "errors"

// Error definitions
var (
	ErrWorkflowEntityNotFound      = errors.New("workflow entity not found")
	ErrActivityEntityNotFound      = errors.New("activity entity not found")
	ErrSagaEntityNotFound          = errors.New("saga entity not found")
	ErrSideEffectEntityNotFound    = errors.New("side effect entity not found")
	ErrWorkflowExecutionNotFound   = errors.New("workflow execution not found")
	ErrActivityExecutionNotFound   = errors.New("activity execution not found")
	ErrSagaExecutionNotFound       = errors.New("saga execution not found")
	ErrSideEffectExecutionNotFound = errors.New("side effect execution not found")
	ErrRunNotFound                 = errors.New("run not found")
	ErrVersionNotFound             = errors.New("version not found")
	ErrHierarchyNotFound           = errors.New("hierarchy not found")
	ErrQueueNotFound               = errors.New("queue not found")
	ErrQueueExists                 = errors.New("queue already exists")
	ErrSagaValueNotFound           = errors.New("saga value not found")
)

var (
	// Core errors
	ErrWorkflowPanicked   = errors.New("workflow panicked")
	ErrActivityPanicked   = errors.New("activity panicked")
	ErrSideEffectPanicked = errors.New("side effect panicked")
	ErrSagaPanicked       = errors.New("saga panicked")

	// Fsm errors
	ErrWorkflowFailed   = errors.New("workflow failed")
	ErrActivityFailed   = errors.New("activity failed")
	ErrSideEffectFailed = errors.New("side effect failed")
	ErrSagaFailed       = errors.New("saga failed")
	ErrSagaCompensated  = errors.New("saga compensated")

	// io
	ErrMustPointer   = errors.New("value must be a pointer")
	ErrEncoding      = errors.New("failed to encode value")
	ErrSerialization = errors.New("failed to serialize")

	// Saga errors
	ErrTransactionContext         = errors.New("failed transaction context")
	ErrCompensationContext        = errors.New("failed compensation context")
	ErrTransactionMethodNotFound  = errors.New("transaction method not found")
	ErrCompensationMethodNotFound = errors.New("compensation method not found")
	ErrSagaGetValue               = errors.New("failed to get saga value")

	// Orchestrator errors
	ErrOrchestrator              = errors.New("failed orchestrator")
	ErrGetResults                = errors.New("cannot get results")
	ErrWorkflowPaused            = errors.New("workflow is paused")
	ErrRegistryHandlerNotFound   = errors.New("handler not found")
	ErrOrchestratorExecution     = errors.New("failed to execute workflow")
	ErrOrchestratorExecuteEntity = errors.New("failed to execute workflow with entity")
	ErrRegistryRegisteration     = errors.New("failed to register workflow")
	ErrPreparation               = errors.New("failed to prepare workflow")

	// Context errors
	ErrActivityContext   = errors.New("failed activity context")
	ErrSideEffectContext = errors.New("failed side effect context")
	ErrSagaContext       = errors.New("failed saga context")
	ErrWorkflowContext   = errors.New("failed workflow context")
	ErrSignalContext     = errors.New("failed signal context")

	// Workflow lifecycle errors
	ErrWorkflowInstance          = errors.New("failed workflow instance")
	ErrWorkflowInstanceExecution = errors.New("failed to execute workflow instance")
	ErrWorkflowInstanceRun       = errors.New("failed to run workflow instance")
	ErrWorkflowInstanceCompleted = errors.New("failed to complete workflow instance")
	ErrWorkflowInstanceFailed    = errors.New("failed to fail workflow instance")
	ErrWorkflowInstancePaused    = errors.New("failed to pause workflow instance")

	// Activity lifecycle errors
	ErrActivityInstance          = errors.New("failed activity instance")
	ErrActivityInstanceExecute   = errors.New("failed to execute activity instance")
	ErrActivityInstanceRun       = errors.New("failed to run activity instance")
	ErrActivityInstanceCompleted = errors.New("failed to complete activity instance")
	ErrActivityInstanceFailed    = errors.New("failed to fail activity instance")
	ErrActivityInstancePaused    = errors.New("failed to pause activity instance")

	// SideEffect lifecycle errors
	ErrSideEffectInstance          = errors.New("failed side effect instance")
	ErrSideEffectInstanceExecute   = errors.New("failed to execute side effect instance")
	ErrSideEffectRun               = errors.New("failed to run side effect")
	ErrSideEffectInstanceCompleted = errors.New("failed to complete side effect instance")
	ErrSideEffectInstanceFailed    = errors.New("failed to fail side effect instance")
	ErrSideEffectInstancePaused    = errors.New("failed to pause side effect instance")

	// Saga lifecycle errors
	ErrSagaContextCompensate  = errors.New("failed to compensate saga context")
	ErrSagaInstance           = errors.New("failed saga instance")
	ErrSagaInstanceExecute    = errors.New("failed to execute saga instance")
	ErrSagaInstanceCompensate = errors.New("failed to compensate saga instance")
	ErrSagaInstanceCompleted  = errors.New("failed to complete saga instance")
	ErrSagaInstanceFailed     = errors.New("failed to fail saga instance")

	// Signal lifecycle errors
	ErrSignalInstance          = errors.New("failed signal instance")
	ErrSignalInstanceExecute   = errors.New("failed to execute signal instance")
	ErrSignalInstanceCompleted = errors.New("failed to complete signal instance")
	ErrSignalInstanceFailed    = errors.New("failed to fail signal instance")
	ErrSignalInstancePaused    = errors.New("failed to pause signal instance")

	ErrSagaInstanceNotFound = errors.New("saga instance not found")
)
