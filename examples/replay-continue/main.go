package main

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/davidroman0O/tempolite"
)

var shouldSucceed atomic.Bool

func workflowContinueReplay(ctx tempolite.WorkflowContext) (string, error) {
	if shouldSucceed.Load() {
		return "Success", nil
	}
	// When you replay a workflow, it should not continue as new
	return "Continue", ctx.ContinueAsNew(ctx, "should-stop", nil)
}

func main() {
	// Initialize Tempolite
	registry := tempolite.NewRegistry().
		Workflow(workflowContinueReplay).
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

	shouldSucceed.Store(true)
	log.Println("Running workflow for the first time (should succeed)")
	workflowInfo := tp.Workflow("test", workflowContinueReplay, nil)

	var result string
	err = workflowInfo.Get(&result)
	if err != nil {
		log.Printf("Workflow failed as expected: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Wait failed: %v", err)
	}

	// Now set the flag to true for the replay
	shouldSucceed.Store(false)
	log.Println("Replaying workflow (should not continue)")

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
