package main

import (
	"context"
	"errors"
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
	select {
	case <-ctx.Done():
		fmt.Println("Activity cancelled")
		return ctx.Err()
	case <-time.After(duration):
		log.Printf("Activity completed")
		return nil
	}
}

var failed atomic.Bool

// WorkflowWithTimeouts demonstrates using durations for both workflows and activities
func WorkflowWithTimeouts(ctx tempolite.WorkflowContext) error {
	// Meaning that you can retry, but if we decide to cancel the workflow then you have to stop
	select {
	case <-ctx.Done():
		fmt.Println("Workflow cancelled")
		return ctx.Err()
	default:
		activity := LongRunningActivity{}

		fmt.Println("Workflow starting")

		// First activity: Should complete within duration
		log.Println("Starting first activity (should complete)")
		err := ctx.Activity(
			"quick-task",
			activity.Run,
			tempolite.ActivityConfig(
				tempolite.WithActivityContextDuration("2s"),
				tempolite.WithActivityRetryMaximumAttempts(0),
			),
			3*time.Second,
		).Get()
		if err != nil {
			return err
		}

		if !failed.Load() {
			failed.Store(true)
			return errors.Join(fmt.Errorf("failed first activity"))
		}

		// Second activity: Should exceed duration and fail
		log.Println("Starting second activity (should timeout)")
		err = ctx.Activity(
			"slow-task",
			activity.Run,
			tempolite.ActivityConfig(
				tempolite.WithActivityContextDuration("1s"),
				tempolite.WithActivityRetryMaximumAttempts(0),
			),
			3*time.Second,
		).Get()
		if err != nil {
			log.Printf("Second activity failed as expected: %v", err)
			// Continue workflow execution despite activity timeout
			return err
		}

		log.Println("Workflow completing")
		return nil
	}
}

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
			tempolite.WithWorkflowContextDuration("0s"),
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
