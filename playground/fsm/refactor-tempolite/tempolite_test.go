package tempolite

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestTempoliteBasic(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	var executed uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
		return nil
	}

	engine, err := New(ctx, database)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer func() {
		if err := engine.Close(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("failed to close engine: %v", err)
		}
	}()

	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	// Check if the workflow was executed
	if atomic.LoadUint32(&executed) == 0 {
		t.Fatalf("workflow was not executed")
	}

	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteNewQueue(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	var executed uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
		return nil
	}

	engine, err := New(
		ctx,
		database,
		WithQueue(QueueConfig{
			Name:        "test-queue",
			WorkerCount: 1,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() {
		if err := engine.Close(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("failed to close engine: %v", err)
		}
	}()

	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	// Check if the workflow was executed
	if atomic.LoadUint32(&executed) == 0 {
		t.Fatalf("workflow was not executed")
	}

	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteNewQueueWorkflow(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	var executed uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
		return nil
	}

	engine, err := New(
		ctx,
		database,
		WithQueue(QueueConfig{
			Name:        "test-queue",
			WorkerCount: 1,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() {
		if err := engine.Close(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("failed to close engine: %v", err)
		}
	}()

	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)
	futureTest := engine.Workflow(workflow, &WorkflowOptions{
		Queue: "test-queue",
	})

	if err := futureTest.Get(); err != nil {
		t.Fatalf("failed to run workflow on test-queue: %v", err)
	}

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow on default queue: %v", err)
	}

	// Check if the workflow was executed
	if atomic.LoadUint32(&executed) == 0 {
		t.Fatalf("workflow was not executed")
	}

	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteCrossQueueWorkflow(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	var crossWorkflowExecuted uint32 // Use atomic variable

	crossWorkflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&crossWorkflowExecuted, 1) // Atomically set executed to true
		return nil
	}

	var workflowExecuted uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&workflowExecuted, 1) // Atomically set executed to true
		if err := ctx.Workflow("queued", crossWorkflow, &WorkflowOptions{
			Queue: "test-queue",
		}).Get(); err != nil {
			return err
		}
		return nil
	}

	engine, err := New(
		ctx,
		database,
		WithQueue(QueueConfig{
			Name:        "test-queue",
			WorkerCount: 1,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() {
		if err := engine.Close(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("failed to close engine: %v", err)
		}
	}()

	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}
	if err := engine.RegisterWorkflow(crossWorkflow); err != nil {
		t.Fatalf("failed to register crossWorkflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	// Check if the workflows were executed
	if atomic.LoadUint32(&workflowExecuted) == 0 {
		t.Fatalf("workflow was not executed")
	}

	if atomic.LoadUint32(&crossWorkflowExecuted) == 0 {
		t.Fatalf("cross workflow was not executed")
	}

	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteCrossQueueWorkflowActivity(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	var activityExecuted uint32     // Use atomic variable
	var onPurposeFailure uint32 = 1 // Use atomic variable to track failure

	activity := func(ctx ActivityContext) error {
		atomic.StoreUint32(&activityExecuted, 1) // Atomically set executed to true
		if atomic.LoadUint32(&onPurposeFailure) == 1 {
			atomic.StoreUint32(&onPurposeFailure, 0)
			return errors.New("on purpose failure")
		}
		return nil
	}

	var crossWorkflowExecuted uint32 // Use atomic variable

	crossWorkflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&crossWorkflowExecuted, 1) // Atomically set executed to true
		if err := ctx.Activity("activity", activity, &ActivityOptions{
			RetryPolicy: &RetryPolicy{
				MaxAttempts: 2,
			},
		}).Get(); err != nil {
			return err
		}
		return nil
	}

	var workflowExecuted uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&workflowExecuted, 1) // Atomically set executed to true
		if err := ctx.Workflow("queued", crossWorkflow, &WorkflowOptions{
			Queue: "test-queue",
		}).Get(); err != nil {
			return err
		}
		return nil
	}

	engine, err := New(
		ctx,
		database,
		WithQueue(QueueConfig{
			Name:        "test-queue",
			WorkerCount: 1,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer func() {
		if err := engine.Close(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("failed to close engine: %v", err)
		}
	}()

	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}
	if err := engine.RegisterWorkflow(crossWorkflow); err != nil {
		t.Fatalf("failed to register crossWorkflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	// Check if the activity was retried
	if atomic.LoadUint32(&onPurposeFailure) != 0 {
		t.Fatalf("activity was not retried")
	}

	// Check if the workflows and activity were executed
	if atomic.LoadUint32(&workflowExecuted) == 0 {
		t.Fatalf("workflow was not executed")
	}

	if atomic.LoadUint32(&crossWorkflowExecuted) == 0 {
		t.Fatalf("cross workflow was not executed")
	}

	if atomic.LoadUint32(&activityExecuted) == 0 {
		t.Fatalf("activity was not executed")
	}

	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

// TODO: make test when we pause/resume
// TODO: make test when we restart Tempolite
