package tests

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestErrorHandling$ ./tests
func TestErrorHandling(t *testing.T) {
	panicCount := atomic.Int32{}
	errorCount := atomic.Int32{}

	panicActivity := func(ctx tempolite.ActivityContext) error {
		panicCount.Add(1)
		panic("intentional panic")
	}

	errorActivity := func(ctx tempolite.ActivityContext) error {
		errorCount.Add(1)
		return fmt.Errorf("intentional error")
	}

	workflowWithPanic := func(ctx tempolite.WorkflowContext) error {
		return ctx.Activity("panic-step", panicActivity, tempolite.ActivityConfig(tempolite.WithActivityRetryMaximumAttempts(0))).Get()
	}

	workflowWithError := func(ctx tempolite.WorkflowContext) error {
		return ctx.Activity("error-step", errorActivity, tempolite.ActivityConfig(tempolite.WithActivityRetryMaximumAttempts(0))).Get()
	}

	// go test -v -count=1 -timeout 30s -run ^TestErrorHandling$/^activity_panic_recovery$ ./tests
	t.Run("activity panic recovery", func(t *testing.T) {
		cfg := NewTestConfig(t, "panic-recovery")
		cfg.Registry.
			Workflow(workflowWithPanic).
			Activity(panicActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		panicCount.Store(0)

		err := tp.Workflow(workflowWithPanic, tempolite.WorkflowConfig(tempolite.WithWorkflowRetryMaximumAttempts(0))).Get()
		if err == nil {
			t.Fatal("Expected workflow to fail due to panic")
		}

		if panicCount.Load() != 1 {
			t.Errorf("Expected panic activity to execute once, got %d", panicCount.Load())
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	// go test -v -count=1 -timeout 30s -run ^TestErrorHandling$/^activity_error_handling$ ./tests
	t.Run("activity error handling", func(t *testing.T) {
		cfg := NewTestConfig(t, "error-handling")
		cfg.Registry.
			Workflow(workflowWithError).
			Activity(errorActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		errorCount.Store(0)

		err := tp.Workflow(workflowWithError, tempolite.WorkflowConfig(tempolite.WithWorkflowRetryMaximumAttempts(0))).Get()
		if err == nil {
			t.Fatal("Expected workflow to fail due to error")
		}

		if errorCount.Load() != 1 {
			t.Errorf("Expected error activity to execute once, got %d", errorCount.Load())
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	// go test -v -count=1 -timeout 30s -run ^TestErrorHandling$/^workflow_error_propagation$ ./tests
	t.Run("workflow error propagation", func(t *testing.T) {
		nestedWorkflow := func(ctx tempolite.WorkflowContext) error {
			return fmt.Errorf("nested workflow error")
		}

		parentWorkflow := func(ctx tempolite.WorkflowContext) error {
			return ctx.Workflow("nested", nestedWorkflow, nil).Get()
		}

		cfg := NewTestConfig(t, "error-propagation")
		cfg.Registry.
			Workflow(parentWorkflow).
			Workflow(nestedWorkflow)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		err := tp.Workflow(parentWorkflow, nil).Get()
		if err == nil {
			t.Fatal("Expected workflow to fail due to nested error")
		}

		if err.Error() != "nested workflow error" {
			t.Errorf("Unexpected error message: %v", err)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})
}
