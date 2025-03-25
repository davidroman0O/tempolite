// Package context provides the execution context for workflows and activities.
package context

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")

	// ErrCancelled is returned when an operation is cancelled.
	ErrCancelled = errors.New("operation cancelled")

	// ErrSignalTimeout is returned when waiting for a signal times out.
	ErrSignalTimeout = errors.New("signal timeout")

	// ErrWorkflowNotFound is returned when a workflow is not found.
	ErrWorkflowNotFound = errors.New("workflow not found")

	// ErrActivityNotFound is returned when an activity is not found.
	ErrActivityNotFound = errors.New("activity not found")

	// ErrWorkflowPaused is returned when attempting to update a paused workflow.
	ErrWorkflowPaused = errors.New("workflow is paused")

	// ErrWorkflowCancelled is returned when attempting to update a cancelled workflow.
	ErrWorkflowCancelled = errors.New("workflow is cancelled")
)

// NewError creates a new error with the given message.
func NewError(message string) error {
	return errors.New(message)
}

// NewErrorf creates a new error with the given format and arguments.
func NewErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// ErrorWithCause creates a new error that wraps the given cause.
func ErrorWithCause(message string, cause error) error {
	return fmt.Errorf("%s: %w", message, cause)
}

// IsTimeout returns whether the given error is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsCancelled returns whether the given error is a cancelled error.
func IsCancelled(err error) bool {
	return errors.Is(err, ErrCancelled)
}

// IsSignalTimeout returns whether the given error is a signal timeout error.
func IsSignalTimeout(err error) bool {
	return errors.Is(err, ErrSignalTimeout)
}

// IsWorkflowNotFound returns whether the given error is a workflow not found error.
func IsWorkflowNotFound(err error) bool {
	return errors.Is(err, ErrWorkflowNotFound)
}

// IsActivityNotFound returns whether the given error is an activity not found error.
func IsActivityNotFound(err error) bool {
	return errors.Is(err, ErrActivityNotFound)
}

// IsWorkflowPaused returns whether the given error is a workflow paused error.
func IsWorkflowPaused(err error) bool {
	return errors.Is(err, ErrWorkflowPaused)
}

// IsWorkflowCancelled returns whether the given error is a workflow cancelled error.
func IsWorkflowCancelled(err error) bool {
	return errors.Is(err, ErrWorkflowCancelled)
}
