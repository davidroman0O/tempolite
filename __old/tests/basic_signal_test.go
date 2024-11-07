package tests

import (
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestSignals$ ./tests
func TestSignals(t *testing.T) {
	workflowFn := func(ctx tempolite.WorkflowContext) (string, error) {
		signal := ctx.Signal("test-signal")

		var value string
		if err := signal.Receive(ctx, &value); err != nil {
			return "", err
		}

		return value + " received", nil
	}

	t.Run("basic signal handling", func(t *testing.T) {
		cfg := NewTestConfig(t, "signals")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		workflowInfo := tp.Workflow(workflowFn, nil)

		// Send signal in a goroutine
		go func() {
			time.Sleep(time.Second)
			err := tp.PublishSignal(workflowInfo.WorkflowID, "test-signal", "test message")
			if err != nil {
				t.Errorf("Failed to publish signal: %v", err)
			}
		}()

		var result string
		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		expected := "test message received"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	t.Run("multiple signals", func(t *testing.T) {
		multiSignalWorkflow := func(ctx tempolite.WorkflowContext) ([]string, error) {
			signal1 := ctx.Signal("signal-1")
			signal2 := ctx.Signal("signal-2")

			var value1, value2 string
			if err := signal1.Receive(ctx, &value1); err != nil {
				return nil, err
			}
			if err := signal2.Receive(ctx, &value2); err != nil {
				return nil, err
			}

			return []string{value1, value2}, nil
		}

		cfg := NewTestConfig(t, "multiple-signals")
		cfg.Registry.Workflow(multiSignalWorkflow)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		workflowInfo := tp.Workflow(multiSignalWorkflow, nil)

		// Send signals in goroutines
		go func() {
			time.Sleep(time.Second)
			_ = tp.PublishSignal(workflowInfo.WorkflowID, "signal-1", "first")
			time.Sleep(time.Second)
			_ = tp.PublishSignal(workflowInfo.WorkflowID, "signal-2", "second")
		}()

		var results []string
		err := workflowInfo.Get(&results)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		if results[0] != "first" || results[1] != "second" {
			t.Errorf("got %v, want [first second]", results)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})
}
