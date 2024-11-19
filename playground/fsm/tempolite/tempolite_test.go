package tempolite

import (
	"context"
	"errors"
	"testing"
)

func TestTempoliteBasic(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	workflow := func(ctx WorkflowContext) error {
		t.Log("workflow started")
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

	t.Log("registering workflow")
	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("failed to register workflow: %v", err)
	}

	t.Log("call for new workflow")
	future := engine.Workflow(workflow, nil)

	t.Log("waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	t.Log("workflow completed successfully")
	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteNewQueue(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	workflow := func(ctx WorkflowContext) error {
		t.Log("workflow started")
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

	t.Log("workflow completed successfully")
	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteNewQueueWorkflow(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	workflow := func(ctx WorkflowContext) error {
		t.Log("workflow started")
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
		t.Fatalf("failed to run workflow: %v", err)
	}

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	t.Log("workflow completed successfully")
	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}

func TestTempoliteCrossQueueWorkflow(t *testing.T) {
	database := NewDefaultDatabase()
	ctx := context.Background()

	crossWorkflow := func(ctx WorkflowContext) error {
		t.Log("cross workflow started")
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		t.Log("workflow started")
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
		t.Fatalf("failed to register workflow: %v", err)
	}

	future := engine.Workflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("failed to run workflow: %v", err)
	}

	t.Log("workflow completed successfully")
	if err := engine.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to wait for engine: %v", err)
	}
}
