package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/davidroman0O/tempolite"
)

var executionCount atomic.Int32

// Activity
type CalculateActivity struct{}

func (c CalculateActivity) Run(ctx tempolite.ActivityContext, x, y int) (int, error) {
	executionCount.Add(1)
	return x + y, nil
}

// Saga steps
type SagaStep1 struct{}

func (s SagaStep1) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	executionCount.Add(1)
	return "Step 1 completed", nil
}
func (s SagaStep1) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	return "Step 1 compensated", nil
}

type SagaStep2 struct{}

func (s SagaStep2) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	executionCount.Add(1)
	return "Step 2 completed", nil
}
func (s SagaStep2) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	return "Step 2 compensated", nil
}

// Main workflow
func complexWorkflow(ctx tempolite.WorkflowContext) (string, error) {
	var result int

	// Side effect
	err := ctx.SideEffect("random-number", func(ctx tempolite.SideEffectContext) int {
		executionCount.Add(1)
		return 42 // In a real scenario, this would be non-deterministic
	}).Get(&result)
	if err != nil {
		return "", fmt.Errorf("side effect failed: %w", err)
	}

	// Activity
	var activityResult int
	err = ctx.Activity("calculate", activityStruct.Run, nil, result, 10).Get(&activityResult)
	if err != nil {
		return "", fmt.Errorf("activity failed: %w", err)
	}

	// Saga
	sagaBuilder := tempolite.NewSaga()
	sagaBuilder.AddStep(SagaStep1{})
	sagaBuilder.AddStep(SagaStep2{})
	saga, _ := sagaBuilder.Build()

	err = ctx.Saga("complex-saga", saga).Get()
	if err != nil {
		return "", fmt.Errorf("saga failed: %w", err)
	}

	return fmt.Sprintf("Workflow completed with result: %d", activityResult), nil
}

var activityStruct CalculateActivity = CalculateActivity{}

func main() {

	registry := tempolite.NewRegistry().
		Workflow(complexWorkflow).
		Activity(activityStruct.Run).
		Build()

	tp, err := tempolite.New(
		context.Background(),
		registry,
		tempolite.WithPath("./db/tempolite-advanced-replay.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// First run
	executionCount.Store(0)
	log.Println("Running workflow for the first time")
	workflowInfo := tp.Workflow(complexWorkflow, nil)

	var result string
	err = workflowInfo.Get(&result)
	if err != nil {
		log.Fatalf("Workflow failed unexpectedly: %v", err)
	}
	log.Printf("Workflow succeeded with result: %s", result)
	log.Printf("Execution count: %d", executionCount.Load())

	if err := tp.Wait(); err != nil {
		log.Fatalf("Wait failed: %v", err)
	}

	// Now replay the workflow
	executionCount.Store(0)
	log.Println("Replaying workflow")

	replayedWorkflowInfo := tp.ReplayWorkflow(workflowInfo.WorkflowID)

	var replayResult string
	err = replayedWorkflowInfo.Get(&replayResult)
	if err != nil {
		log.Fatalf("Replayed workflow failed: %v", err)
	}

	log.Printf("Replayed workflow result: %s", replayResult)
	log.Printf("Execution count after replay: %d", executionCount.Load())

	if result != replayResult {
		log.Fatalf("Replay result doesn't match original result")
	}

	if executionCount.Load() != 0 {
		log.Fatalf("Execution count should be 0 for replay, but got: %d", executionCount.Load())
	}

	log.Println("Replay successful and matched original execution")

	if err := tp.Wait(); err != nil {
		log.Fatalf("Wait failed: %v", err)
	}
}
