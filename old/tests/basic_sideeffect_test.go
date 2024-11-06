package tests

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestSideEffects$ ./tests
func TestSideEffects(t *testing.T) {
	// Counter to verify side effect execution
	executionCount := atomic.Int32{}

	workflowFn := func(ctx tempolite.WorkflowContext) (int, error) {
		var result int
		err := ctx.SideEffect("random-value", func(ctx tempolite.SideEffectContext) int {
			executionCount.Add(1)
			return 42
		}).Get(&result)

		if err != nil {
			return 0, err
		}
		return result, nil
	}

	t.Run("side effect execution", func(t *testing.T) {
		cfg := NewTestConfig(t, "side-effects")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		// First execution
		var result int
		workflowInfo := tp.Workflow(workflowFn, nil)

		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if result != 42 {
			t.Errorf("got %v, want 42", result)
		}

		// Original execution
		if count := executionCount.Load(); count != 1 {
			t.Errorf("Expected 1 execution, got %d", count)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}

		// Test replay using the actual workflow ID
		executionCount.Store(0)
		replayInfo := tp.ReplayWorkflow(workflowInfo.WorkflowID)
		if err := replayInfo.Get(&result); err != nil {
			t.Fatalf("Replay failed: %v", err)
		}

		// Side effect should not execute during replay
		if count := executionCount.Load(); count != 0 {
			t.Errorf("Expected 0 executions during replay, got %d", count)
		}
	})
}
