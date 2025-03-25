// Package context provides the execution context for workflows and activities.
package context

import (
	"time"
)

// ParallelGroup provides a way to execute multiple activities and workflows in parallel.
type ParallelGroup interface {
	// Add activities to the group
	Activity(name string) ActivityBuilder

	// Add workflows to the group
	Workflow(name string) WorkflowBuilder

	// Execution control
	Wait() error
	WaitWithTimeout(timeout time.Duration) error
	Cancel()
}

// parallelGroupImpl implements the ParallelGroup interface.
type parallelGroupImpl struct {
	workflowContext *workflowContextImpl
	activities      map[string]ActivityBuilder
	workflows       map[string]WorkflowBuilder
	futures         []interface{} // ActivityFuture or WorkflowFuture
}

// Activity adds an activity builder to the group.
func (pg *parallelGroupImpl) Activity(name string) ActivityBuilder {
	builder := pg.workflowContext.Activity(name)
	pg.activities[name] = builder
	return builder
}

// Workflow adds a workflow builder to the group.
func (pg *parallelGroupImpl) Workflow(name string) WorkflowBuilder {
	builder := pg.workflowContext.Workflow(name)
	pg.workflows[name] = builder
	return builder
}

// Wait waits for all activities and workflows in the group to complete.
func (pg *parallelGroupImpl) Wait() error {
	// Wait for all futures to complete
	for _, future := range pg.futures {
		switch f := future.(type) {
		case ActivityFuture:
			if err := f.Get(); err != nil {
				return err
			}
		case WorkflowFuture:
			if err := f.Get(); err != nil {
				return err
			}
		}
	}
	return nil
}

// WaitWithTimeout waits for all activities and workflows in the group to complete
// or until the timeout expires.
func (pg *parallelGroupImpl) WaitWithTimeout(timeout time.Duration) error {
	// Set a deadline for all futures
	deadline := time.After(timeout)

	// Create a map of channels to results
	results := make(map[<-chan struct{}]interface{})
	remaining := len(pg.futures)

	// Collect all OnResults channels
	for _, future := range pg.futures {
		var ch <-chan struct{}
		switch f := future.(type) {
		case ActivityFuture:
			ch = f.OnResults()
		case WorkflowFuture:
			ch = f.OnResults()
		}
		results[ch] = future
	}

	// Wait for all futures or timeout
	for remaining > 0 {
		select {
		case <-deadline:
			pg.Cancel()
			return ErrTimeout
		default:
			// Wait for any future to complete
			for ch, future := range results {
				select {
				case <-ch:
					// Future completed
					switch f := future.(type) {
					case ActivityFuture:
						if err := f.Error(); err != nil {
							return err
						}
					case WorkflowFuture:
						if err := f.Error(); err != nil {
							return err
						}
					}
					delete(results, ch)
					remaining--
				default:
					// Future not yet completed
				}
			}

			// Sleep a bit to avoid busy waiting
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}

// Cancel cancels all activities and workflows in the group.
func (pg *parallelGroupImpl) Cancel() {
	// Cancel all futures
	for _, future := range pg.futures {
		switch f := future.(type) {
		case ActivityFuture:
			f.Cancel()
		case WorkflowFuture:
			f.Cancel()
		}
	}
}

// AddFuture adds a future to the group for tracking.
// This is an internal method used by ActivityBuilder and WorkflowBuilder.
func (pg *parallelGroupImpl) AddFuture(future interface{}) {
	pg.futures = append(pg.futures, future)
}
