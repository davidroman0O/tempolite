package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/davidroman0O/tempolite"
)

var shouldSucceed atomic.Bool

func undeterministicWorkflow(ctx tempolite.WorkflowContext) (string, error) {
	if shouldSucceed.Load() {
		return "Success", nil
	}
	return "", fmt.Errorf("simulated failure")
}

func main() {
	// Initialize Tempolite
	registry := tempolite.NewRegistry().
		Workflow(undeterministicWorkflow).
		Build()

	tp, err := tempolite.New(
		context.Background(),
		registry,
		tempolite.WithPath("./db/tempolite-replay-example.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// First run - this should fail
	shouldSucceed.Store(false)
	log.Println("Running workflow for the first time (should fail)")
	workflowInfo := tp.Workflow(undeterministicWorkflow, nil)

	var result string
	err = workflowInfo.Get(&result)
	if err != nil {
		log.Printf("Workflow failed as expected: %v", err)
	} else {
		log.Fatalf("Workflow unexpectedly succeeded with result: %s", result)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Wait failed: %v", err)
	}

	// Now set the flag to true for the replay
	shouldSucceed.Store(true)
	log.Println("Replaying workflow (should succeed)")

	// Replay the workflow
	replayedWorkflowInfo := tp.ReplayWorkflow(workflowInfo.WorkflowID)

	// Wait for the replayed workflow to complete
	err = replayedWorkflowInfo.Get(&result)
	if err != nil {
		log.Fatalf("Replayed workflow failed: %v", err)
	}

	log.Printf("Replayed workflow succeeded with result: %s", result)

	if err := tp.Wait(); err != nil {
		log.Fatalf("Wait failed: %v", err)
	}
}
