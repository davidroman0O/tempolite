// Package context provides the execution context for workflows and activities.
package context

import (
	"time"
)

// WorkflowBuilder provides a fluent API for configuring and executing child workflows.
type WorkflowBuilder interface {
	// Configuration
	WithRetries(policy RetryPolicy) WorkflowBuilder
	WithTimeout(timeout time.Duration) WorkflowBuilder
	WithQueue(queue string) WorkflowBuilder

	// Execution
	Run(fn interface{}, args ...interface{}) WorkflowFuture
	RunIf(condition bool, fn interface{}, args ...interface{}) WorkflowFuture
	RunWhen(predicate func() bool, fn interface{}, args ...interface{}) WorkflowFuture
}

// WorkflowFuture represents the result of an asynchronous workflow execution.
type WorkflowFuture interface {
	// Result handling
	Get(results ...interface{}) error
	OnResults() <-chan struct{}
	Cancel() error
	IsDone() bool

	// Status information
	Status() string
	Error() error
	ID() string
}

// workflowBuilderImpl implements the WorkflowBuilder interface.
type workflowBuilderImpl struct {
	workflowContext *workflowContextImpl
	workflowName    string
	retryPolicy     *RetryPolicy
	timeout         time.Duration
	queue           string
}

// WithRetries configures the retry policy for the workflow.
func (wb *workflowBuilderImpl) WithRetries(policy RetryPolicy) WorkflowBuilder {
	wb.retryPolicy = &policy
	return wb
}

// WithTimeout configures the timeout for the workflow.
func (wb *workflowBuilderImpl) WithTimeout(timeout time.Duration) WorkflowBuilder {
	wb.timeout = timeout
	return wb
}

// WithQueue configures the queue for the workflow.
func (wb *workflowBuilderImpl) WithQueue(queue string) WorkflowBuilder {
	wb.queue = queue
	return wb
}

// Run executes the workflow with the given function and arguments.
func (wb *workflowBuilderImpl) Run(fn interface{}, args ...interface{}) WorkflowFuture {
	return &workflowFutureImpl{
		workflowBuilder: wb,
		workflowFn:      fn,
		args:            args,
	}
}

// RunIf executes the workflow if the condition is true.
func (wb *workflowBuilderImpl) RunIf(condition bool, fn interface{}, args ...interface{}) WorkflowFuture {
	if !condition {
		return &workflowFutureImpl{
			workflowBuilder: wb,
			completed:       true,
		}
	}
	return wb.Run(fn, args...)
}

// RunWhen executes the workflow when the predicate returns true.
func (wb *workflowBuilderImpl) RunWhen(predicate func() bool, fn interface{}, args ...interface{}) WorkflowFuture {
	if !predicate() {
		return &workflowFutureImpl{
			workflowBuilder: wb,
			completed:       true,
		}
	}
	return wb.Run(fn, args...)
}

// workflowFutureImpl implements the WorkflowFuture interface.
type workflowFutureImpl struct {
	workflowBuilder *workflowBuilderImpl
	workflowFn      interface{}
	args            []interface{}
	result          interface{}
	err             error
	completed       bool
	doneCh          chan struct{}
	workflowID      string
}

// Get retrieves the result of the workflow.
func (wf *workflowFutureImpl) Get(results ...interface{}) error {
	if wf.completed {
		return wf.err
	}

	// Wait for workflow to complete and get result
	return nil
}

// OnResults returns a channel that is closed when the workflow completes.
func (wf *workflowFutureImpl) OnResults() <-chan struct{} {
	if wf.doneCh == nil {
		wf.doneCh = make(chan struct{})
		if wf.completed {
			close(wf.doneCh)
		}
	}
	return wf.doneCh
}

// Cancel attempts to cancel the workflow.
func (wf *workflowFutureImpl) Cancel() error {
	// Cancel the workflow
	return nil
}

// IsDone returns whether the workflow has completed.
func (wf *workflowFutureImpl) IsDone() bool {
	return wf.completed
}

// Status returns the status of the workflow.
func (wf *workflowFutureImpl) Status() string {
	if wf.completed {
		if wf.err != nil {
			return "Failed"
		}
		return "Completed"
	}
	return "Running"
}

// Error returns any error that occurred during the workflow.
func (wf *workflowFutureImpl) Error() error {
	return wf.err
}

// ID returns the ID of the workflow.
func (wf *workflowFutureImpl) ID() string {
	return wf.workflowID
}

// complete marks the future as complete with the given result and error.
func (wf *workflowFutureImpl) complete(result interface{}, err error) {
	wf.result = result
	wf.err = err
	wf.completed = true
	if wf.doneCh != nil {
		close(wf.doneCh)
	}
}
