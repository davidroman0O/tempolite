package tests

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestContinueAsNew$ ./tests
func TestContinueAsNew(t *testing.T) {
	executionCount := atomic.Int32{}

	workflowFn := func(ctx tempolite.WorkflowContext, iteration int, maxIterations int) (int, error) {
		executionCount.Add(1)

		if iteration >= maxIterations {
			return iteration, nil
		}

		return 0, ctx.ContinueAsNew(ctx, "continue-step", nil, iteration+1, maxIterations)
	}

	// go test -v -count=1 -timeout 30s -run ^TestContinueAsNew$/^basic_continuation$ ./tests
	t.Run("basic continuation", func(t *testing.T) {
		cfg := NewTestConfig(t, "continue-basic")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		executionCount.Store(0)

		var result int
		workflowInfo := tp.Workflow(workflowFn, nil, 0, 3)

		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		continued := true
		for continued {
			workflowInfo = workflowInfo.GetContinuation()
			if err := workflowInfo.Get(&result); err != nil {
				t.Fatalf("Workflow execution failed: %v", err)
			}
			continued = workflowInfo.IsContinued
		}

		if result != 3 {
			t.Errorf("Expected final iteration to be 3, got %d", result)
		}

		if count := executionCount.Load(); count != 4 {
			t.Errorf("Expected 4 executions (0->1->2->3), got %d", count)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	// go test -timeout 30s -run ^TestContinueAsNew$/^execution_chain_tracking$ ./tests
	t.Run("execution chain tracking", func(t *testing.T) {
		cfg := NewTestConfig(t, "continue-tracking")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		executionCount.Store(0)

		workflowInfo := tp.Workflow(workflowFn, nil, 0, 2)

		var result int
		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		// Get the latest workflow execution
		latestID, err := tp.GetLatestWorkflowExecution(workflowInfo.WorkflowID)
		if err != nil {
			t.Fatalf("Failed to get latest workflow execution: %v", err)
		}

		if latestID == workflowInfo.WorkflowID {
			t.Error("Latest workflow ID should be different from original after continue as new")
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	// go test -timeout 30s -run ^TestContinueAsNew$/^continue_with_side_effects$ ./tests
	t.Run("continue with side effects", func(t *testing.T) {
		sideEffectCount := atomic.Int32{}

		workflowWithSideEffect := func(ctx tempolite.WorkflowContext, iteration int) (int, error) {
			var value int
			err := ctx.SideEffect("test-effect", func(sCtx tempolite.SideEffectContext) int {
				sideEffectCount.Add(1)
				return iteration * 2
			}).Get(&value)

			if err != nil {
				return 0, err
			}

			if iteration < 2 {
				return 0, ctx.ContinueAsNew(ctx, "continue-step", nil, iteration+1)
			}

			return value, nil
		}

		cfg := NewTestConfig(t, "continue-side-effects")
		cfg.Registry.Workflow(workflowWithSideEffect)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		sideEffectCount.Store(0)

		var result int
		workflowInfo := tp.Workflow(workflowWithSideEffect, nil, 0)
		if err := workflowInfo.Get(&result); err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		continued := true
		for continued {
			workflowInfo = workflowInfo.GetContinuation()
			if err := workflowInfo.Get(&result); err != nil {
				t.Fatalf("Workflow execution failed: %v", err)
			}
			continued = workflowInfo.IsContinued
		}

		expectedExecutions := 3 // One for each iteration (0, 1, 2)
		if count := sideEffectCount.Load(); count != int32(expectedExecutions) {
			t.Errorf("Expected %d side effect executions, got %d", expectedExecutions, count)
		}

		if result != 4 { // Final iteration (2) * 2
			t.Errorf("Expected final result to be 4, got %d", result)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})
}
