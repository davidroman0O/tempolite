// Package context provides the execution context for workflows and activities.
// It contains the primary interfaces for workflow definition and execution.
package context

import (
	"time"
)

// WorkflowContext is the primary interface for workflow functions.
// It provides methods for executing activities, child workflows, side effects,
// waiting for signals, creating sagas, and other workflow operations.
type WorkflowContext interface {
	// Context operations
	WithTimeout(timeout time.Duration) WorkflowContext
	WithRetryPolicy(policy RetryPolicy) WorkflowContext

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
	Logger() Logger

	// Versioning
	Version() int
	WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor

	// Utilities
	ParallelGroup() ParallelGroup
	RegisterQuery(name string, fn interface{}) error
}

// workflowContextImpl implements the WorkflowContext interface.
type workflowContextImpl struct {
	workflowID      string
	executionID     string
	currentVersion  int
	currentActivity string
	// Internal references to the engine, storage, etc.
	engine interface{}
	logger Logger
}

// NewWorkflowContext creates a new workflow context.
func NewWorkflowContext(workflowID, executionID string, engine interface{}, logger Logger) WorkflowContext {
	return &workflowContextImpl{
		workflowID:     workflowID,
		executionID:    executionID,
		currentVersion: 0,
		engine:         engine,
		logger:         logger,
	}
}

// WithTimeout returns a new workflow context with the specified timeout.
func (wc *workflowContextImpl) WithTimeout(timeout time.Duration) WorkflowContext {
	// Create a new context with timeout
	return wc
}

// WithRetryPolicy returns a new workflow context with the specified retry policy.
func (wc *workflowContextImpl) WithRetryPolicy(policy RetryPolicy) WorkflowContext {
	// Create a new context with retry policy
	return wc
}

// Activity creates an activity builder for the specified activity.
func (wc *workflowContextImpl) Activity(name string) ActivityBuilder {
	return &activityBuilderImpl{
		workflowContext: wc,
		activityName:    name,
	}
}

// Workflow creates a workflow builder for the specified workflow.
func (wc *workflowContextImpl) Workflow(name string) WorkflowBuilder {
	return &workflowBuilderImpl{
		workflowContext: wc,
		workflowName:    name,
	}
}

// SideEffect creates a side effect builder for the specified side effect.
func (wc *workflowContextImpl) SideEffect(name string) SideEffectBuilder {
	return &sideEffectBuilderImpl{
		workflowContext: wc,
		sideEffectName:  name,
	}
}

// WaitForSignal creates a signal builder for waiting for the specified signal.
func (wc *workflowContextImpl) WaitForSignal(name string) SignalBuilder {
	return &signalBuilderImpl{
		workflowContext: wc,
		signalName:      name,
	}
}

// Saga creates a saga builder for the specified saga.
func (wc *workflowContextImpl) Saga(name string) SagaBuilder {
	return &sagaBuilderImpl{
		workflowContext: wc,
		sagaName:        name,
	}
}

// Sleep pauses the workflow for the specified duration.
func (wc *workflowContextImpl) Sleep(duration time.Duration) error {
	// Sleep implementation
	return nil
}

// ContinueAsNew continues the workflow as a new execution with the specified arguments.
func (wc *workflowContextImpl) ContinueAsNew(args ...interface{}) error {
	// Continue-as-new implementation
	return nil
}

// Logger returns the logger for the workflow.
func (wc *workflowContextImpl) Logger() Logger {
	return wc.logger
}

// Version returns the current version number for the workflow.
func (wc *workflowContextImpl) Version() int {
	return wc.currentVersion
}

// WithVersion returns a version executor for the specified version.
func (wc *workflowContextImpl) WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor {
	return &versionExecutorImpl{
		workflowContext: wc,
		version:         version,
		fn:              fn,
		next:            nil,
	}
}

// ParallelGroup creates a new parallel group.
func (wc *workflowContextImpl) ParallelGroup() ParallelGroup {
	return &parallelGroupImpl{
		workflowContext: wc,
		activities:      make(map[string]ActivityBuilder),
		workflows:       make(map[string]WorkflowBuilder),
	}
}

// RegisterQuery registers a query handler for the workflow.
func (wc *workflowContextImpl) RegisterQuery(name string, fn interface{}) error {
	// Register query handler
	return nil
}

// RetryPolicy defines how retries are handled.
type RetryPolicy struct {
	MaxAttempts int
	MaxInterval int64
}

// Logger is the interface for logging within Tempolite.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// VersionExecutor is used for executing versioned code.
type VersionExecutor interface {
	WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor
	Execute() error
}

// versionExecutorImpl implements the VersionExecutor interface.
type versionExecutorImpl struct {
	workflowContext *workflowContextImpl
	version         int
	fn              func(WorkflowContext) error
	next            *versionExecutorImpl
}

// WithVersion adds another version handler to the executor.
func (ve *versionExecutorImpl) WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor {
	next := &versionExecutorImpl{
		workflowContext: ve.workflowContext,
		version:         version,
		fn:              fn,
		next:            nil,
	}

	ve.next = next
	return next
}

// Execute executes the version handler.
func (ve *versionExecutorImpl) Execute() error {
	if ve.workflowContext.currentVersion == ve.version {
		return ve.fn(ve.workflowContext)
	}

	if ve.next != nil {
		return ve.next.Execute()
	}

	return nil
}
