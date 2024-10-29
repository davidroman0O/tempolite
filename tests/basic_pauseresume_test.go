package tests

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestPauseResume$ ./tests
func TestPauseResume(t *testing.T) {
	progressCounter := atomic.Int32{}

	slowActivity := func(ctx tempolite.ActivityContext) error {
		progressCounter.Add(1)
		time.Sleep(2 * time.Second)
		return nil
	}

	workflowFn := func(ctx tempolite.WorkflowContext) error {
		// Multiple activities to ensure we can pause between them
		if err := ctx.Activity("step1", slowActivity, nil).Get(); err != nil {
			return err
		}

		if err := ctx.Activity("step2", slowActivity, nil).Get(); err != nil {
			return err
		}

		if err := ctx.Activity("step3", slowActivity, nil).Get(); err != nil {
			return err
		}

		return nil
	}

	t.Run("pause and resume workflow", func(t *testing.T) {
		cfg := NewTestConfig(t, "pause-resume")
		cfg.Registry.
			Workflow(workflowFn).
			Activity(slowActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		progressCounter.Store(0)

		// Start workflow
		workflowInfo := tp.Workflow(workflowFn, nil)

		// Wait briefly for the first activity to start
		time.Sleep(500 * time.Millisecond)

		// Pause the workflow
		err := tp.PauseWorkflow(workflowInfo.WorkflowID)
		if err != nil {
			t.Fatalf("Failed to pause workflow: %v", err)
		}

		// Record progress at pause
		pauseProgress := progressCounter.Load()

		// Wait to ensure no progress while paused
		time.Sleep(2 * time.Second)

		if currentProgress := progressCounter.Load(); currentProgress != pauseProgress {
			t.Errorf("Progress continued while paused: got %d, want %d", currentProgress, pauseProgress)
		}

		// Resume the workflow
		err = tp.ResumeWorkflow(workflowInfo.WorkflowID)
		if err != nil {
			t.Fatalf("Failed to resume workflow: %v", err)
		}

		// Wait for completion
		err = workflowInfo.Get()
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if err := WaitWithTimeout(t, tp, 10*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}

		// Verify all activities completed
		finalProgress := progressCounter.Load()
		if finalProgress != 3 {
			t.Errorf("Expected 3 activities to complete, got %d", finalProgress)
		}
	})

	t.Run("list paused workflows", func(t *testing.T) {
		cfg := NewTestConfig(t, "list-paused")
		cfg.Registry.
			Workflow(workflowFn).
			Activity(slowActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		workflowInfo1 := tp.Workflow(workflowFn, nil)
		workflowInfo2 := tp.Workflow(workflowFn, nil)

		// Pause both workflows
		if err := tp.PauseWorkflow(workflowInfo1.WorkflowID); err != nil {
			t.Fatalf("Failed to pause workflow 1: %v", err)
		}
		if err := tp.PauseWorkflow(workflowInfo2.WorkflowID); err != nil {
			t.Fatalf("Failed to pause workflow 2: %v", err)
		}

		// Get list of paused workflows
		pausedWorkflows, err := tp.ListPausedWorkflows()
		if err != nil {
			t.Fatalf("Failed to list paused workflows: %v", err)
		}

		if len(pausedWorkflows) != 2 {
			t.Errorf("Expected 2 paused workflows, got %d", len(pausedWorkflows))
		}

		// Resume one workflow
		if err := tp.ResumeWorkflow(workflowInfo1.WorkflowID); err != nil {
			t.Fatalf("Failed to resume workflow 1: %v", err)
		}

		// Check paused workflows again
		pausedWorkflows, err = tp.ListPausedWorkflows()
		if err != nil {
			t.Fatalf("Failed to list paused workflows: %v", err)
		}

		if len(pausedWorkflows) != 1 {
			t.Errorf("Expected 1 paused workflow after resume, got %d", len(pausedWorkflows))
		}

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflow completion: %v", err)
		}
	})
}
