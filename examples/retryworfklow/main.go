package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/davidroman0O/go-tempolite"
)

type MyIdentifier string

var shouldFail atomic.Bool

func exampleWorkflow(ctx tempolite.WorkflowContext[MyIdentifier], input int) (int, error) {
	log.Printf("Starting workflow with input: %d", input)

	// Simulate some work
	time.Sleep(time.Second)

	if shouldFail.Load() {
		return 0, fmt.Errorf("workflow failed")
	}

	result := input * 2
	return result, nil
}

func main() {
	// Initialize Tempolite
	tp, err := tempolite.New[MyIdentifier](
		context.Background(),
		tempolite.NewRegistry[MyIdentifier]().
			Workflow(exampleWorkflow).
			Build(),
		tempolite.WithPath("./db/tempolite-retryworklow.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Tempolite: %v", err)
	}
	defer tp.Close()

	// Set the workflow to fail on the first run
	shouldFail.Store(true)

	// Execute the workflow
	workflowInfo := tp.Workflow(MyIdentifier("example-step"), exampleWorkflow, nil, 42)
	var result int
	if err := workflowInfo.Get(&result); err != nil {
		log.Printf("Workflow execution failed as expected: %v", err)

		// Set the workflow to succeed on retry
		shouldFail.Store(false)

		// Retry the workflow
		log.Printf("Retrying workflow with ID: %s", workflowInfo.WorkflowID)
		newWorkflowInfo := tp.RetryWorkflow(workflowInfo.WorkflowID)

		// Wait for the retried workflow to complete
		if err := newWorkflowInfo.Get(&result); err != nil {
			log.Fatalf("Retried workflow failed unexpectedly: %v", err)
		} else {
			log.Printf("Retried workflow completed successfully with result: %d", result)
		}
	} else {
		log.Fatalf("Workflow should have failed on the first run")
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}

	// Once finished, you can check the DB yourself
}
