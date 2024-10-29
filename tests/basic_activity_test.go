package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

type TestActivity struct {
	MultiplyBy int
}

func (a TestActivity) Run(ctx tempolite.ActivityContext, value int) (int, error) {
	return value * a.MultiplyBy, nil
}

// go test -v -count=1 -timeout 30s -run ^TestActivityExecution$ ./tests
func TestActivityExecution(t *testing.T) {
	activity := TestActivity{MultiplyBy: 2}

	workflowFn := func(ctx tempolite.WorkflowContext, input int) (int, error) {
		var result int
		err := ctx.Activity("multiply", activity.Run, nil, input).Get(&result)
		if err != nil {
			return 0, err
		}
		return result, nil
	}

	cfg := NewTestConfig(t, "activity-execution")
	cfg.Registry.
		Workflow(workflowFn).
		Activity(activity.Run)

	tp := NewTestTempolite(t, cfg)
	defer tp.Close()

	testCases := []struct {
		name  string
		input int
		want  int
	}{
		{
			name:  "multiply by 2",
			input: 5,
			want:  10,
		},
		{
			name:  "multiply zero",
			input: 0,
			want:  0,
		},
		{
			name:  "multiply negative",
			input: -3,
			want:  -6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result int
			err := tp.Workflow(
				workflowFn,
				nil,
				tc.input,
			).Get(&result)

			if err != nil {
				t.Fatalf("Workflow execution failed: %v", err)
			}

			if result != tc.want {
				t.Errorf("got %v, want %v", result, tc.want)
			}

			if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
				t.Fatalf("Failed to wait for workflow completion: %v", err)
			}
		})
	}
}

// go test -v -count=1 -timeout 30s -run ^TestActivityRetry$ ./tests
func TestActivityRetry(t *testing.T) {
	attempts := 0

	activity := func(ctx tempolite.ActivityContext, input string) (string, error) {
		attempts++
		if attempts == 1 {
			return "", fmt.Errorf("intentional activity failure")
		}
		return input + " processed", nil
	}

	workflowFn := func(ctx tempolite.WorkflowContext, input string) (string, error) {
		var result string
		err := ctx.Activity("retry-activity", activity, nil, input).Get(&result)
		if err != nil {
			return "", err
		}
		return result, nil
	}

	cfg := NewTestConfig(t, "activity-retry")
	cfg.Registry.
		Workflow(workflowFn).
		Activity(activity)

	tp := NewTestTempolite(t, cfg)
	defer tp.Close()

	var result string
	err := tp.Workflow(
		workflowFn,
		tempolite.WorkflowConfig(
			tempolite.WithWorkflowRetryMaximumAttempts(2),
		),
		"test",
	).Get(&result)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	if result != "test processed" {
		t.Errorf("Expected 'test processed', got '%s'", result)
	}

	if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
		t.Fatalf("Failed to wait for workflow completion: %v", err)
	}
}
