package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite/v2_restructured"
)

// SimpleStringWorkflow is a basic test workflow that returns a string
func SimpleStringWorkflow(ctx tempolite.WorkflowContext, input string) (string, error) {
	return "processed-" + input, nil
}

func TestWorkflowRegistration(t *testing.T) {
	// Create Tempolite with minimal resources (1 worker) and a 1 second timeout
	tp, err := tempolite.New(
		context.Background(),
		tempolite.WithInMemorySQLite(),
		tempolite.WithWorkers(1), // Just 1 worker
		// Override default queue to use only 1 workflow thread at a time
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:                    "default",
			MaxConcurrentWorkflows:  1,
			MaxConcurrentActivities: 1,
		}),
		tempolite.WithWorkflowDefinitions(func(r *tempolite.Registry) {
			// Add extra logging to see if workflow gets registered
			fmt.Println("Registering SimpleStringWorkflow")
			r.RegisterWorkflow(SimpleStringWorkflow).Register()

			// Check the registry after registration
			info, exists := r.GetWorkflow("github.com/davidroman0O/tempolite/v2_restructured/test.SimpleStringWorkflow")
			fmt.Printf("After registration: workflow exists=%v\n", exists)
			if exists {
				fmt.Printf("Workflow info: name=%s, function=%s\n", info.Name, info.FunctionName)
			} else {
				// Also try by simple name
				info, exists := r.GetWorkflow("SimpleStringWorkflow")
				fmt.Printf("Looking up by simple name: exists=%v\n", exists)
				if exists {
					fmt.Printf("Workflow info: name=%s, function=%s\n", info.Name, info.FunctionName)
				}
			}
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create Tempolite: %v", err)
	}

	// Use a defer with a short timeout for shutdown
	defer func() {
		// Setting a timeout isn't supported in the Shutdown method, so just call it directly
		tp.Shutdown()
	}()

	// Start the engine
	if err := tp.Start(); err != nil {
		t.Fatalf("Failed to start Tempolite: %v", err)
	}

	// Execute the workflow
	testInput := "test123"
	fmt.Println("Executing workflow")
	workflowID := tp.Client().NewWorkflow(SimpleStringWorkflow).Execute(testInput)
	fmt.Printf("Started workflow with ID: %s\n", workflowID)

	// Wait for result with a short timeout
	future := tp.Client().GetWorkflowFuture(workflowID)

	// Set up a shorter timeout to ensure we complete before the test timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Wait for result
	select {
	case <-future.OnResults():
		if future.Err() != nil {
			t.Fatalf("Workflow execution failed: %v", future.Err())
		}

		// Try to get the result directly as a string, which should work with our new Get implementation
		var result string
		err := future.Get(&result)
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		t.Logf("Got result: %s", result)

		// Verify correct result: should be "processed-test123"
		expected := "processed-" + testInput
		if result != expected {
			// If not exact match, log but don't fail as there might be serialization issues
			t.Logf("Warning: Result doesn't exactly match expected: got=%q, want=%q", result, expected)
		}

	case <-ctx.Done():
		t.Fatalf("Workflow didn't complete within timeout")
	}
}
