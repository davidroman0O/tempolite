package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/davidroman0O/tempolite"
)

// Counter to track concurrent executions
var concurrentExecutions atomic.Int32

// SubWorkflowData represents the input for sub-workflows
type SubWorkflowData struct {
	ID       int
	TaskName string
}

// ProcessTask is a simple activity that simulates work
func ProcessTask(ctx tempolite.ActivityContext, data SubWorkflowData) error {
	count := concurrentExecutions.Add(1)
	defer concurrentExecutions.Add(-1)

	log.Printf("[Activity] Processing task %s (ID: %d): %d",
		data.TaskName, data.ID, count)

	// Simulate some work
	time.Sleep(5 * time.Second)
	return nil
}

// SubWorkflow processes a single task
func SubWorkflow(ctx tempolite.WorkflowContext, data SubWorkflowData) error {
	log.Printf("[SubWorkflow] Starting task %s (ID: %d)", data.TaskName, data.ID)
	<-time.After(1 * time.Second)
	if err := ctx.Activity("process-task", ProcessTask, data).Get(); err != nil {
		return fmt.Errorf("failed to process task: %w", err)
	}

	log.Printf("[SubWorkflow] Completed task %s (ID: %d)", data.TaskName, data.ID)
	return nil
}

// MainWorkflow orchestrates multiple sub-workflows
func MainWorkflow(ctx tempolite.WorkflowContext, numTasks int) error {
	log.Printf("[MainWorkflow] Starting workflow with %d tasks", numTasks)

	workflowInfos := []*tempolite.WorkflowInfo{}
	// Create tasks sequentially but they'll be processed by the single worker in the specialized queue
	for i := 0; i < numTasks; i++ {
		taskData := SubWorkflowData{
			ID:       i + 1,
			TaskName: fmt.Sprintf("Task-%d", i+1),
		}

		stepID := fmt.Sprintf("sub-workflow-%d", i+1)

		// Execute sub-workflow in specialized queue with one worker
		workflowInfos = append(
			workflowInfos,
			ctx.Workflow(
				stepID,
				SubWorkflow,
				tempolite.WorkflowConfig(
					tempolite.WithQueue("specialized-queue"),
				),
				taskData))
	}

	// In theory will run one by one
	for idx, v := range workflowInfos {
		if err := v.Get(); err != nil {
			return fmt.Errorf("failed to wait for sub-workflow %v: %w", idx, err)
		}
	}

	log.Printf("[MainWorkflow] All tasks completed successfully")
	return nil
}

func main() {
	// Initialize registry with workflows and activities
	registry := tempolite.NewRegistry().
		Workflow(MainWorkflow).
		Workflow(SubWorkflow).
		Activity(ProcessTask).
		Build()

	// Create Tempolite instance with a specialized queue
	tp, err := tempolite.New(
		context.Background(),
		registry,
		tempolite.WithPath("./db/tempolite-queues.db"),
		tempolite.WithDestructive(),
		tempolite.WithQueueConfig(
			tempolite.NewQueue(
				"specialized-queue",
				tempolite.WithWorkflowWorkers(1),
				tempolite.WithActivityWorkers(1),
			),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Start the main workflow on the default queue
	log.Println("Starting main workflow on default queue")
	if err := tp.Workflow("main", MainWorkflow, nil, 5).Get(); err != nil {
		log.Fatalf("Main workflow failed: %v", err)
	}

	// Wait for all workflows to complete
	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows: %v", err)
	}

	log.Println("All workflows completed successfully")
}
