package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite"
)

// LongRunningActivity simulates a task that takes time to complete
type LongRunningActivity struct{}

func (l LongRunningActivity) Run(ctx tempolite.ActivityContext, stepName string) error {
	log.Printf("Executing step: %s", stepName)
	time.Sleep(2 * time.Second) // Simulate work
	return nil
}

// PersistentWorkflow demonstrates a workflow that can survive application restarts
func PersistentWorkflow(ctx tempolite.WorkflowContext, processID string) error {
	log.Printf("Starting workflow for process %s", processID)

	activity := LongRunningActivity{}

	// Step 1
	if err := ctx.Activity("step-1", activity.Run, nil, "First Step").Get(); err != nil {
		return fmt.Errorf("step 1 failed: %w", err)
	}
	log.Printf("Completed step 1 for process %s", processID)

	// Step 2
	if err := ctx.Activity("step-2", activity.Run, nil, "Second Step").Get(); err != nil {
		return fmt.Errorf("step 2 failed: %w", err)
	}
	log.Printf("Completed step 2 for process %s", processID)

	// Step 3
	if err := ctx.Activity("step-3", activity.Run, nil, "Third Step").Get(); err != nil {
		return fmt.Errorf("step 3 failed: %w", err)
	}
	log.Printf("Completed step 3 for process %s", processID)

	log.Printf("Workflow completed for process %s", processID)
	return nil
}

// initTempolite creates a new Tempolite instance
func initTempolite() (*tempolite.Tempolite, error) {
	return tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(PersistentWorkflow).
			Activity(LongRunningActivity{}.Run).
			Build(),
		tempolite.WithPath("./db/tempolite-persistence.db"),
	)
}

func main() {
	// First instance - Start the workflow
	log.Println("=== Starting first instance ===")
	tp1, err := initTempolite()
	if err != nil {
		log.Fatalf("Failed to create first Tempolite instance: %v", err)
	}

	processID := "PROC-001"
	workflowInfo := tp1.Workflow(PersistentWorkflow, nil, processID)

	// Wait for the first activity to complete
	time.Sleep(3 * time.Second)

	log.Println("\n=== Shutting down first instance ===")
	tp1.Close()

	// Wait a moment to simulate application restart
	log.Println("\n=== Simulating application restart ===")
	time.Sleep(2 * time.Second)

	// Second instance - Resume the workflow
	log.Println("\n=== Starting second instance ===")
	tp2, err := initTempolite()
	if err != nil {
		log.Fatalf("Failed to create second Tempolite instance: %v", err)
	}
	defer tp2.Close()

	// Get the workflow information using the same ID
	resumedWorkflow := tp2.GetWorkflow(workflowInfo.WorkflowID)

	// Wait for completion
	if err := resumedWorkflow.Get(); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	if err := tp2.Wait(); err != nil {
		log.Fatalf("Error waiting for workflow to complete: %v", err)
	}

	log.Println("\n=== Workflow completed successfully ===")
}
