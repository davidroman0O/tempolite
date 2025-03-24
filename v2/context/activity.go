// Package context provides the execution context for workflows and activities.
package context

import (
	"time"
)

// ActivityBuilder provides a fluent API for configuring and executing activities.
type ActivityBuilder interface {
	// Configuration
	WithRetries(policy RetryPolicy) ActivityBuilder
	WithTimeout(timeout time.Duration) ActivityBuilder
	WithHeartbeat(interval time.Duration) ActivityBuilder
	WithQueue(queue string) ActivityBuilder

	// Execution
	Run(fn interface{}, args ...interface{}) ActivityFuture
	RunIf(condition bool, fn interface{}, args ...interface{}) ActivityFuture
	RunWhen(predicate func() bool, fn interface{}, args ...interface{}) ActivityFuture
}

// ActivityFuture represents the result of an asynchronous activity execution.
type ActivityFuture interface {
	// Result handling
	Get(results ...interface{}) error
	OnResults() <-chan struct{}
	Cancel() error
	IsDone() bool

	// Status information
	Status() string
	Error() error
}

// ActivityContext is the context provided to activity functions.
type ActivityContext interface {
	// Context operations
	WithTimeout(timeout time.Duration) ActivityContext

	// Logging
	Logger() Logger

	// Heartbeating
	Heartbeat(details ...interface{}) error

	// Context information
	WorkflowID() string
	ActivityID() string
	RetryCount() int
}

// activityContextImpl implements the ActivityContext interface.
type activityContextImpl struct {
	workflowID string
	activityID string
	retryCount int
	logger     Logger
}

// NewActivityContext creates a new activity context.
func NewActivityContext(workflowID, activityID string, retryCount int, logger Logger) ActivityContext {
	return &activityContextImpl{
		workflowID: workflowID,
		activityID: activityID,
		retryCount: retryCount,
		logger:     logger,
	}
}

// WithTimeout returns a new activity context with the specified timeout.
func (ac *activityContextImpl) WithTimeout(timeout time.Duration) ActivityContext {
	// Create a new context with timeout
	return ac
}

// Logger returns the logger for the activity.
func (ac *activityContextImpl) Logger() Logger {
	return ac.logger
}

// Heartbeat sends a heartbeat for the activity.
func (ac *activityContextImpl) Heartbeat(details ...interface{}) error {
	// Send heartbeat
	return nil
}

// WorkflowID returns the ID of the workflow that initiated the activity.
func (ac *activityContextImpl) WorkflowID() string {
	return ac.workflowID
}

// ActivityID returns the ID of the activity.
func (ac *activityContextImpl) ActivityID() string {
	return ac.activityID
}

// RetryCount returns the current retry count for the activity.
func (ac *activityContextImpl) RetryCount() int {
	return ac.retryCount
}

// activityBuilderImpl implements the ActivityBuilder interface.
type activityBuilderImpl struct {
	workflowContext *workflowContextImpl
	activityName    string
	retryPolicy     *RetryPolicy
	timeout         time.Duration
	heartbeat       time.Duration
	queue           string
}

// WithRetries configures the retry policy for the activity.
func (ab *activityBuilderImpl) WithRetries(policy RetryPolicy) ActivityBuilder {
	ab.retryPolicy = &policy
	return ab
}

// WithTimeout configures the timeout for the activity.
func (ab *activityBuilderImpl) WithTimeout(timeout time.Duration) ActivityBuilder {
	ab.timeout = timeout
	return ab
}

// WithHeartbeat configures the heartbeat interval for the activity.
func (ab *activityBuilderImpl) WithHeartbeat(interval time.Duration) ActivityBuilder {
	ab.heartbeat = interval
	return ab
}

// WithQueue configures the queue for the activity.
func (ab *activityBuilderImpl) WithQueue(queue string) ActivityBuilder {
	ab.queue = queue
	return ab
}

// Run executes the activity with the given function and arguments.
func (ab *activityBuilderImpl) Run(fn interface{}, args ...interface{}) ActivityFuture {
	return &activityFutureImpl{
		activityBuilder: ab,
		activityFn:      fn,
		args:            args,
	}
}

// RunIf executes the activity if the condition is true.
func (ab *activityBuilderImpl) RunIf(condition bool, fn interface{}, args ...interface{}) ActivityFuture {
	if !condition {
		return &activityFutureImpl{
			activityBuilder: ab,
			completed:       true,
		}
	}
	return ab.Run(fn, args...)
}

// RunWhen executes the activity when the predicate returns true.
func (ab *activityBuilderImpl) RunWhen(predicate func() bool, fn interface{}, args ...interface{}) ActivityFuture {
	if !predicate() {
		return &activityFutureImpl{
			activityBuilder: ab,
			completed:       true,
		}
	}
	return ab.Run(fn, args...)
}

// activityFutureImpl implements the ActivityFuture interface.
type activityFutureImpl struct {
	activityBuilder *activityBuilderImpl
	activityFn      interface{}
	args            []interface{}
	result          interface{}
	err             error
	completed       bool
	doneCh          chan struct{}
}

// Get retrieves the result of the activity.
func (af *activityFutureImpl) Get(results ...interface{}) error {
	if af.completed {
		return af.err
	}

	// Wait for activity to complete and get result
	return nil
}

// OnResults returns a channel that is closed when the activity completes.
func (af *activityFutureImpl) OnResults() <-chan struct{} {
	if af.doneCh == nil {
		af.doneCh = make(chan struct{})
		if af.completed {
			close(af.doneCh)
		}
	}
	return af.doneCh
}

// Cancel attempts to cancel the activity.
func (af *activityFutureImpl) Cancel() error {
	// Cancel the activity
	return nil
}

// IsDone returns whether the activity has completed.
func (af *activityFutureImpl) IsDone() bool {
	return af.completed
}

// Status returns the status of the activity.
func (af *activityFutureImpl) Status() string {
	if af.completed {
		if af.err != nil {
			return "Failed"
		}
		return "Completed"
	}
	return "Running"
}

// Error returns any error that occurred during the activity.
func (af *activityFutureImpl) Error() error {
	return af.err
}

// complete marks the future as complete with the given result and error.
func (af *activityFutureImpl) complete(result interface{}, err error) {
	af.result = result
	af.err = err
	af.completed = true
	if af.doneCh != nil {
		close(af.doneCh)
	}
}
