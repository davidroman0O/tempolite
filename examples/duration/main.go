package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/davidroman0O/tempolite"
)

// LongRunningActivity simulates a task that might exceed its duration limit
type LongRunningActivity struct{}

func (l LongRunningActivity) Run(ctx tempolite.ActivityContext, duration time.Duration) error {
	log.Printf("Activity starting, will run for %v", duration)
	time.Sleep(duration)
	log.Printf("Activity completed")
	return nil
}

var failed atomic.Bool

// WorkflowWithTimeouts demonstrates using durations for both workflows and activities
func WorkflowWithTimeouts(ctx tempolite.WorkflowContext) error {
	activity := LongRunningActivity{}

	fmt.Println("Workflow starting")

	// First activity: Should complete within duration
	log.Println("Starting first activity (should complete)")
	err := ctx.Activity(
		"quick-task",
		activity.Run,
		tempolite.ActivityConfig(
			tempolite.WithActivityDuration("2s"),
			tempolite.WithActivityRetryMaximumAttempts(0),
		),
		3*time.Second,
	).Get()
	if err != nil {
		return fmt.Errorf("first activity failed: %w", err)
	}

	if !failed.Load() {
		failed.Store(true)
		return fmt.Errorf("first activity failed")
	}

	// Second activity: Should exceed duration and fail
	log.Println("Starting second activity (should timeout)")
	err = ctx.Activity(
		"slow-task",
		activity.Run,
		tempolite.ActivityConfig(
			tempolite.WithActivityDuration("1s"),
			tempolite.WithActivityRetryMaximumAttempts(0),
		),
		3*time.Second,
	).Get()
	if err != nil {
		log.Printf("Second activity failed as expected: %v", err)
		// Continue workflow execution despite activity timeout
	}

	log.Println("Workflow completing")
	return nil
}

// TODO: that feature doesn't works YET
func main() {
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(WorkflowWithTimeouts).
			Activity(LongRunningActivity{}.Run).
			Build(),
		tempolite.WithPath("./db/tempolite-duration.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Start the workflow with a total duration limit of 5 seconds
	workflowInfo := tp.Workflow(
		WorkflowWithTimeouts,
		tempolite.WorkflowConfig(
			tempolite.WithWorkflowDuration("1s"),
			tempolite.WithWorkflowRetryMaximumAttempts(1),
		),
	)

	if err := workflowInfo.Get(); err != nil {
		log.Printf("Workflow execution failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflow to complete: %v", err)
	}
}
