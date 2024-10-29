package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite"
)

// ProcessOrderActivity represents a long-running task
type ProcessOrderActivity struct{}

func (p ProcessOrderActivity) Run(ctx tempolite.ActivityContext, orderID string) error {
	log.Printf("Processing order %s...", orderID)
	time.Sleep(2 * time.Second) // Simulate long-running process
	return nil
}

// SendNotificationActivity represents another task
type SendNotificationActivity struct{}

func (n SendNotificationActivity) Run(ctx tempolite.ActivityContext, orderID string) error {
	log.Printf("Sending notification for order %s...", orderID)
	time.Sleep(1 * time.Second) // Simulate notification sending
	return nil
}

// OrderWorkflow demonstrates a workflow that can be paused and resumed
func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
	log.Printf("Starting workflow for order %s", orderID)

	// First activity
	processor := ProcessOrderActivity{}
	if err := ctx.Activity("process-order", processor.Run, nil, orderID).Get(); err != nil {
		return fmt.Errorf("failed to process order: %w", err)
	}

	log.Printf("First activity completed for order %s", orderID)

	// Second activity
	notifier := SendNotificationActivity{}
	if err := ctx.Activity("send-notification", notifier.Run, nil, orderID).Get(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("Second activity completed for order %s", orderID)
	return nil
}

func main() {
	// Initialize Tempolite
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(OrderWorkflow).
			Activity(ProcessOrderActivity{}.Run).
			Activity(SendNotificationActivity{}.Run).
			Build(),
		tempolite.WithPath("./db/tempolite-pause-resume.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Start the workflow
	orderID := "ORDER-001"
	workflowInfo := tp.Workflow(OrderWorkflow, nil, orderID)

	// Wait a moment for the first activity to start
	time.Sleep(1 * time.Second)

	// Pause the workflow
	log.Println("Pausing workflow...")
	if err := tp.PauseWorkflow(workflowInfo.WorkflowID); err != nil {
		log.Fatalf("Failed to pause workflow: %v", err)
	}

	// List paused workflows
	pausedWorkflows, err := tp.ListPausedWorkflows()
	if err != nil {
		log.Fatalf("Failed to list paused workflows: %v", err)
	}
	log.Printf("Paused workflows: %v", pausedWorkflows)

	// Wait for a while (simulating external conditions)
	log.Println("Workflow is paused. Waiting 5 seconds before resuming...")
	time.Sleep(5 * time.Second)

	// Resume the workflow
	log.Println("Resuming workflow...")
	if err := tp.ResumeWorkflow(workflowInfo.WorkflowID); err != nil {
		log.Fatalf("Failed to resume workflow: %v", err)
	}

	// Wait for workflow completion
	if err := workflowInfo.Get(); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}

	log.Println("Workflow completed successfully!")
}
