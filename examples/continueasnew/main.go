package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/go-tempolite"
)

type CustomIdentifier string

func LongRunningWorkflow(ctx tempolite.WorkflowContext[CustomIdentifier], iteration int) (int, error) {
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
	return 0, ctx.ContinueAsNew(ctx, "continue-workflow", iteration+1)
}

func main() {
	tp, err := tempolite.New[CustomIdentifier](
		context.Background(),
		tempolite.NewRegistry[CustomIdentifier]().
			Workflow(LongRunningWorkflow).
			Build(),
		tempolite.WithPath("./tempolite_long_running.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	workflowInfo := tp.Workflow("long-running", LongRunningWorkflow, nil, 0)

	// Simulate sending signals to the workflow
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			err := tp.PublishSignal(workflowInfo.WorkflowID, "continue-signal", true)
			if err != nil {
				log.Printf("Failed to publish signal: %v", err)
			}
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
}
