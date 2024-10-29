package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

type WorkflowInput struct {
	Message string
	Value   int
}

type WorkflowOutput struct {
	Result string
	Count  int
}

// go test -v -count=1 -timeout 30s -run ^TestBasicWorkflow$ ./tests
func TestBasicWorkflow(t *testing.T) {
	// Define test workflow
	workflowFn := func(ctx tempolite.WorkflowContext, input WorkflowInput) (WorkflowOutput, error) {
		return WorkflowOutput{
			Result: input.Message,
			Count:  input.Value,
		}, nil
	}

	// Create test configuration
	cfg := NewTestConfig(t, "basic-workflow")
	cfg.Registry.Workflow(workflowFn)

	// Create Tempolite instance
	tp := NewTestTempolite(t, cfg)
	defer tp.Close()

	// Test cases
	testCases := []struct {
		name  string
		input WorkflowInput
		want  WorkflowOutput
	}{
		{
			name: "simple workflow execution",
			input: WorkflowInput{
				Message: "Hello",
				Value:   42,
			},
			want: WorkflowOutput{
				Result: "Hello",
				Count:  42,
			},
		},
		{
			name: "zero value workflow execution",
			input: WorkflowInput{
				Message: "",
				Value:   0,
			},
			want: WorkflowOutput{
				Result: "",
				Count:  0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output WorkflowOutput
			err := tp.Workflow(
				workflowFn,
				nil,
				tc.input,
			).Get(&output)

			if err != nil {
				t.Fatalf("Workflow execution failed: %v", err)
			}

			if output != tc.want {
				t.Errorf("got %v, want %v", output, tc.want)
			}

			if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
				t.Fatalf("Failed to wait for workflow completion: %v", err)
			}
		})
	}
}

// go test -v -count=1 -timeout 30s -run ^TestWorkflowRetry$ ./tests
func TestWorkflowRetry(t *testing.T) {
	attempts := 0
	workflowFn := func(ctx tempolite.WorkflowContext, input string) (string, error) {
		attempts++
		if attempts == 1 {
			return "", fmt.Errorf("intentional failure")
		}
		return input + " success", nil
	}

	cfg := NewTestConfig(t, "workflow-retry")
	cfg.Registry.Workflow(workflowFn)

	tp := NewTestTempolite(t, cfg)
	defer tp.Close()

	var result string
	workflowInfo := tp.Workflow(
		workflowFn,
		tempolite.WorkflowConfig(
			tempolite.WithWorkflowRetryMaximumAttempts(2),
		),
		"test",
	)

	err := workflowInfo.Get(&result)
	if err != nil {
		t.Fatalf("Initial workflow execution failed: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	if result != "test success" {
		t.Errorf("Expected 'test success', got '%s'", result)
	}

	if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
		t.Fatalf("Failed to wait for workflow completion: %v", err)
	}
}
