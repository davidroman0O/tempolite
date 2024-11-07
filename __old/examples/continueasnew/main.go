package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite"
)

func LongRunningWorkflow(ctx tempolite.WorkflowContext, iteration int) (int, error) {
	if iteration >= 3 {
		return iteration, nil
	}

	// Wait for a signal
	signal := ctx.Signal("continue-signal")
	var shouldContinue bool
	if err := signal.Receive(ctx, &shouldContinue); err != nil {
		return 0, err
	}

	if !shouldContinue {
		return iteration, nil
	}

	// Continue as new
	return 0, ctx.ContinueAsNew(ctx, "continue-workflow", nil, iteration+1)
}

func main() {
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(LongRunningWorkflow).
			Build(),
		tempolite.WithPath("./db/tempolite-continue-as-new.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	workflowInfo := tp.Workflow(LongRunningWorkflow, nil, 0)

	// Simulate sending signals to the workflow
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)

			// Get the latest workflow execution ID before publishing the signal
			latestWorkflowID, err := tp.GetLatestWorkflowExecution(workflowInfo.WorkflowID)
			if err != nil {
				log.Printf("Failed to get latest workflow execution: %v\n", err)
				continue
			}

			err = tp.PublishSignal(latestWorkflowID, "continue-signal", true)
			if err != nil {
				log.Printf("Failed to publish signal: %v\n", err)
			}
			log.Printf("Published signal to workflow: %s\n", latestWorkflowID)
		}
	}()

	var result int
	if err := workflowInfo.Get(&result); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Printf("Workflow completed after %d iterations\n", result)

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for tasks to complete: %v", err)
	}

	latestWorkflowID, err := tp.GetLatestWorkflowExecution(workflowInfo.WorkflowID)
	if err != nil {
		log.Printf("Failed to get latest workflow execution: %v\n", err)
		return
	}
	fmt.Println("Latest workflow is", latestWorkflowID)
}
