package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestWorkflowVersioning$ ./tests
func TestWorkflowVersioning(t *testing.T) {
	const (
		changeID = "TestFeature"
	)

	workflowFn := func(ctx tempolite.WorkflowContext, maxVersion int) (string, error) {
		version := ctx.GetVersion(changeID, tempolite.DefaultVersion, maxVersion)

		switch version {
		case tempolite.DefaultVersion:
			return "original", nil
		case 1:
			return "version1", nil
		case 2:
			return "version2", nil
		default:
			return "unknown", nil
		}
	}

	t.Run("version progression", func(t *testing.T) {
		cfg := NewTestConfig(t, "versioning")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		testCases := []struct {
			maxVersion int
			want       string
		}{
			{maxVersion: tempolite.DefaultVersion, want: "original"},
			{maxVersion: 1, want: "version1"},
			{maxVersion: 2, want: "version2"},
		}

		var lastWorkflowID tempolite.WorkflowID
		for i, tc := range testCases {
			var result string
			workflowInfo := tp.Workflow(workflowFn, nil, tc.maxVersion)

			err := workflowInfo.Get(&result)
			if err != nil {
				t.Fatalf("Workflow execution %d failed: %v", i, err)
			}

			if result != tc.want {
				t.Errorf("Version %d: got %q, want %q", tc.maxVersion, result, tc.want)
			}

			lastWorkflowID = workflowInfo.WorkflowID

			if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
				t.Fatalf("Failed to wait for workflow completion: %v", err)
			}
		}

		// Verify version consistency with replay
		var replayResult string
		err := tp.ReplayWorkflow(lastWorkflowID).Get(&replayResult)
		if err != nil {
			t.Fatalf("Replay failed: %v", err)
		}

		if replayResult != "version2" {
			t.Errorf("Replay result: got %q, want %q", replayResult, "version2")
		}
	})

	t.Run("version compatibility", func(t *testing.T) {
		compatWorkflow := func(ctx tempolite.WorkflowContext) (string, error) {
			// Test min/max version compatibility
			v1 := ctx.GetVersion("Feature1", 1, 2)
			v2 := ctx.GetVersion("Feature1", 1, 2) // Same feature, should return same version

			if v1 != v2 {
				return "", fmt.Errorf("version mismatch: got %d and %d", v1, v2)
			}

			return fmt.Sprintf("version%d", v1), nil
		}

		cfg := NewTestConfig(t, "version-compat")
		cfg.Registry.Workflow(compatWorkflow)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		var result string
		err := tp.Workflow(compatWorkflow, nil).Get(&result)
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		// Version should be stable within same workflow execution
		if result != "version2" {
			t.Errorf("Got %q, want version2", result)
		}
	})
}
