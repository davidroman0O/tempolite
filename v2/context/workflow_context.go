package context

import (
	"time"

	"github.com/davidroman0O/tempolite/v2/internal"
)

// WorkflowContext is the primary interface for workflow functions.
// It provides methods for executing activities, child workflows, side effects,
// waiting for signals, creating sagas, and other workflow operations.
type WorkflowContext interface {
	// Context operations
	WithTimeout(timeout time.Duration) WorkflowContext
	WithRetryPolicy(RetryPolicy) WorkflowContext
	WithOptions(WorkflowOptions) WorkflowContext

	// Core operations
	Activity(name string) ActivityBuilder
	Workflow(name string) WorkflowBuilder
	SideEffect(name string) SideEffectBuilder
	WaitForSignal(name string) SignalBuilder
	Saga(name string) SagaBuilder

	// Workflow control
	Sleep(duration time.Duration) error
	ContinueAsNew(args ...interface{}) error

	// Logging
	Logger() internal.Logger

	// Versioning
	Version() int
	WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor

	// Utilities
	ParallelGroup() ParallelGroup
	RegisterQuery(name string, fn interface{}) error

	// Workflow information
	WorkflowID() string
	ExecutionID() string
}

// WorkflowOptions contains configuration options for workflow execution
type WorkflowOptions struct {
	Timeout      time.Duration
	RetryPolicy  RetryPolicy
	Queue        string
	TaskPriority int
}

// RetryPolicy defines how workflows and activities should be retried
type RetryPolicy struct {
	MaxAttempts        uint
	InitialInterval    time.Duration
	MaxInterval        time.Duration
	BackoffCoefficient float64
	MaxRetryDuration   time.Duration
	NonRetryableErrors []string
}

// VersionExecutor allows for conditional execution based on workflow versions
type VersionExecutor interface {
	WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor
	Execute() error
}

// versionExecutorImpl implements the VersionExecutor interface
type versionExecutorImpl struct {
	workflowContext WorkflowContext
	version         int
	fn              func(WorkflowContext) error
	next            VersionExecutor
}

// WithVersion adds another version-specific handler
func (ve *versionExecutorImpl) WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor {
	ve.next = &versionExecutorImpl{
		workflowContext: ve.workflowContext,
		version:         version,
		fn:              fn,
	}
	return ve.next
}

// Execute runs the appropriate version handler
func (ve *versionExecutorImpl) Execute() error {
	if ve.workflowContext.Version() == ve.version {
		return ve.fn(ve.workflowContext)
	}

	if ve.next != nil {
		return ve.next.Execute()
	}

	return nil
}
