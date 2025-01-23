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
