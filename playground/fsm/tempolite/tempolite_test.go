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
