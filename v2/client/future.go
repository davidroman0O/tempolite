// Package client provides the external API for interacting with the Tempolite engine.
package client

import (
	"errors"
	"sync"
	"time"
)

// Future represents the result of an asynchronous operation.
// It provides methods for waiting on and retrieving results.
type Future struct {
	client     *Client
	workflowID string
	operation  string
	done       chan struct{}
	err        error
	result     interface{}
	mu         sync.RWMutex
	completed  bool
}

// NewFuture creates a new future for an asynchronous operation.
func NewFuture(client *Client, workflowID, operation string) *Future {
	return &Future{
		client:     client,
		workflowID: workflowID,
		operation:  operation,
		done:       make(chan struct{}),
	}
}

// Wait blocks until the operation completes or the context is cancelled.
func (f *Future) Wait() error {
	<-f.done
	return f.err
}

// WaitWithTimeout blocks until the operation completes, the timeout expires, or the context is cancelled.
func (f *Future) WaitWithTimeout(timeout time.Duration) error {
	select {
	case <-f.done:
		return f.err
	case <-time.After(timeout):
		return errors.New("operation timed out")
	}
}

// Done returns a channel that is closed when the operation completes.
func (f *Future) Done() <-chan struct{} {
	return f.done
}

// Err returns any error that occurred during the operation.
func (f *Future) Err() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.err
}

// Get retrieves the result of the operation.
// It blocks until the operation completes and populates the provided pointer with the result.
func (f *Future) Get(result interface{}) error {
	if err := f.Wait(); err != nil {
		return err
	}

	// TODO: Use serialization to populate the result with f.result
	// This would use the internal serialization utilities to convert
	// the result to the appropriate type

	return nil
}

// OnResults returns a channel that is closed when the operation completes.
func (f *Future) OnResults() <-chan struct{} {
	return f.done
}

// Cancel attempts to cancel the operation.
func (f *Future) Cancel() error {
	// Check if already completed
	f.mu.RLock()
	if f.completed {
		f.mu.RUnlock()
		return errors.New("operation already completed")
	}
	f.mu.RUnlock()

	// Attempt to cancel the operation via the client
	// This is just a placeholder - real implementation would
	// call something like client.CancelWorkflow(workflowID)
	return nil
}

// IsDone returns whether the operation has completed.
func (f *Future) IsDone() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.completed
}

// complete marks the future as complete with the given result and error.
// This is called internally by the Tempolite engine.
func (f *Future) complete(result interface{}, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.completed {
		return // Already completed
	}

	f.result = result
	f.err = err
	f.completed = true
	close(f.done)
}

// BatchFuture represents the result of an operation performed on multiple workflows.
type BatchFuture struct {
	futures map[string]*Future
	done    chan struct{}
	err     error
	mu      sync.RWMutex
}

// NewBatchFuture creates a new batch future.
func NewBatchFuture(futures map[string]*Future) *BatchFuture {
	bf := &BatchFuture{
		futures: futures,
		done:    make(chan struct{}),
	}

	// Start a goroutine to wait for all futures to complete
	go func() {
		for _, future := range futures {
			<-future.Done()
		}
		close(bf.done)
	}()

	return bf
}

// Wait blocks until all operations complete.
func (bf *BatchFuture) Wait() error {
	<-bf.done
	return bf.err
}

// Done returns a channel that is closed when all operations complete.
func (bf *BatchFuture) Done() <-chan struct{} {
	return bf.done
}

// Results returns a map of workflow IDs to errors.
func (bf *BatchFuture) Results() map[string]error {
	results := make(map[string]error)
	for id, future := range bf.futures {
		results[id] = future.Err()
	}
	return results
}

// Cancel attempts to cancel all operations.
func (bf *BatchFuture) Cancel() {
	for _, future := range bf.futures {
		future.Cancel()
	}
}
