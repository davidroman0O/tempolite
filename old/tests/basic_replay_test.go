package tests

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestReplay$ ./tests
func TestReplay(t *testing.T) {
	t.Run("basic replay", func(t *testing.T) {
		sideEffectCount := atomic.Int32{}
		activityCount := atomic.Int32{}

		testActivity := func(ctx tempolite.ActivityContext, input int) (int, error) {
			activityCount.Add(1)
			return input * 2, nil
		}

		workflowFn := func(ctx tempolite.WorkflowContext, input int) (int, error) {
			var sideEffectValue int
			err := ctx.SideEffect("random-value", func(sCtx tempolite.SideEffectContext) int {
				sideEffectCount.Add(1)
				return 42
			}).Get(&sideEffectValue)

			if err != nil {
				return 0, err
			}

			var activityResult int
			err = ctx.Activity("test-activity", testActivity, nil, input).Get(&activityResult)
			if err != nil {
				return 0, err
			}

			return sideEffectValue + activityResult, nil
		}

		cfg := NewTestConfig(t, "replay-basic")
		cfg.Registry.
			Workflow(workflowFn).
			Activity(testActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		sideEffectCount.Store(0)
		activityCount.Store(0)

		// First execution
		var result int
		workflowInfo := tp.Workflow(workflowFn, nil, 5)

		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Initial workflow execution failed: %v", err)
		}

		initialSideEffects := sideEffectCount.Load()
		initialActivities := activityCount.Load()

		if initialSideEffects != 1 {
			t.Errorf("Expected 1 side effect execution, got %d", initialSideEffects)
		}

		if initialActivities != 1 {
			t.Errorf("Expected 1 activity execution, got %d", initialActivities)
		}

		// Reset counters for replay
		sideEffectCount.Store(0)
		activityCount.Store(0)

		// Replay workflow
		var replayResult int
		err = tp.ReplayWorkflow(workflowInfo.WorkflowID).Get(&replayResult)
		if err != nil {
			t.Fatalf("Workflow replay failed: %v", err)
		}

		// Verify no side effects or activities were executed during replay
		if count := sideEffectCount.Load(); count != 0 {
			t.Errorf("Side effect executed during replay: got %d executions", count)
		}

		if count := activityCount.Load(); count != 0 {
			t.Errorf("Activity executed during replay: got %d executions", count)
		}

		// Results should match
		if replayResult != result {
			t.Errorf("Replay result mismatch: got %d, want %d", replayResult, result)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})

	t.Run("replay with signals", func(t *testing.T) {
		workflowFn := func(ctx tempolite.WorkflowContext) (string, error) {
			signal := ctx.Signal("test-signal")

			var value string
			if err := signal.Receive(ctx, &value); err != nil {
				return "", err
			}

			return "signal: " + value, nil
		}

		cfg := NewTestConfig(t, "replay-signals")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		workflowInfo := tp.Workflow(workflowFn, nil)

		// Send signal
		go func() {
			time.Sleep(time.Second)
			err := tp.PublishSignal(workflowInfo.WorkflowID, "test-signal", "hello")
			if err != nil {
				t.Errorf("Failed to publish signal: %v", err)
			}
		}()

		var result string
		err := workflowInfo.Get(&result)
		if err != nil {
			t.Fatalf("Initial workflow execution failed: %v", err)
		}

		// Replay workflow
		var replayResult string
		err = tp.ReplayWorkflow(workflowInfo.WorkflowID).Get(&replayResult)
		if err != nil {
			t.Fatalf("Workflow replay failed: %v", err)
		}

		if replayResult != result {
			t.Errorf("Replay result mismatch: got %q, want %q", replayResult, result)
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})
}
