package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/davidroman0O/tempolite/playground/fsm/tempolite"
)

// Saga types and implementations

// Example Saga Steps
type ReserveInventorySaga struct {
	Data int
}

func (s ReserveInventorySaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("ReserveInventorySaga Transaction called with Data: %d", s.Data)
	// Simulate successful transaction
	return nil, nil
}

func (s ReserveInventorySaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("ReserveInventorySaga Compensation called")
	// Simulate compensation
	return nil, nil
}

type ProcessPaymentSaga struct {
	Data int
}

func (s ProcessPaymentSaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Transaction called with Data: %d", s.Data)
	// Simulate failure in transaction
	// if s.Data%2 == 0 {
	// 	return nil, fmt.Errorf("Payment processing failed")
	// }
	return nil, nil
}

func (s ProcessPaymentSaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Compensation called")
	// Simulate compensation
	return nil, nil
}

type UpdateLedgerSaga struct {
	Data int
}

func (s UpdateLedgerSaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("UpdateLedgerSaga Transaction called with Data: %d", s.Data)
	// Simulate successful transaction
	return nil, nil
}

func (s UpdateLedgerSaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("UpdateLedgerSaga Compensation called")
	// Simulate compensation
	return nil, nil
}

// Example Activity
func SomeActivity(ctx tempolite.ActivityContext, data int) (int, error) {
	log.Printf("SomeActivity called with data: %d", data)
	select {
	case <-time.After(time.Second * 2):
		// Simulate processing
	case <-ctx.Done():
		log.Printf("SomeActivity context cancelled")
		return -1, ctx.Err()
	}
	result := data * 3
	log.Printf("SomeActivity returning result: %d", result)
	return result, nil
}

func SubSubSubWorkflow(ctx *tempolite.WorkflowContext, data int) (int, error) {
	log.Printf("SubSubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, &tempolite.ActivityOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	log.Printf("SubSubSubWorkflow returning result: %d", result)
	return result, nil
}

func SubSubWorkflow(ctx *tempolite.WorkflowContext, data int) (int, error) {
	log.Printf("SubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, &tempolite.ActivityOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	if err := ctx.Workflow("subsubsubworkflow-step", SubSubSubWorkflow, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, result).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubSubWorkflow returning result: %d", result)
	return result, nil
}

var subWorkflowFailed atomic.Bool

func SubWorkflow(ctx *tempolite.WorkflowContext, data int) (int, error) {
	log.Printf("SubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, &tempolite.ActivityOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from activity: %v", err)
		return -1, err
	}

	if subWorkflowFailed.Load() {
		subWorkflowFailed.Store(false)
		return -1, fmt.Errorf("subworkflow failed on purpose")
	}

	if err := ctx.Workflow("subsubworkflow-step", SubSubWorkflow, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, result).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubWorkflow returning result: %d", result)
	return result, nil
}

func SubManyWorkflow(ctx *tempolite.WorkflowContext) (int, int, error) {
	return 1, 2, nil
}

var continueAsNewCalled atomic.Bool

func Workflow(ctx *tempolite.WorkflowContext, data int) (int, error) {
	log.Printf("Workflow called with data: %d", data)
	var value int

	select {
	case <-ctx.Done():
		log.Printf("Workflow context cancelled")
		return -1, ctx.Err()
	default:
	}

	// Versioning example
	changeID := "ChangeIDCalculateTax"
	const DefaultVersion = 0
	version := ctx.GetVersion(changeID, DefaultVersion, 1)
	if version == DefaultVersion {
		log.Printf("Using default version logic")
		// Original logic
	} else {
		log.Printf("Using new version logic")
		// New logic
	}

	var shouldDouble bool
	if err := ctx.SideEffect("side-effect-step", func() bool {
		log.Printf("Side effect called")
		return rand.Float32() < 0.5
	}, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        1,
			InitialInterval:    0,
			BackoffCoefficient: 1.0,
			MaxInterval:        0,
		},
	}).Get(&shouldDouble); err != nil {
		log.Printf("Workflow encountered error from side effect: %v", err)
		return -1, err
	}

	if !continueAsNewCalled.Load() {
		continueAsNewCalled.Store(true)
		log.Printf("Using ContinueAsNew to restart workflow")
		err := ctx.ContinueAsNew(Workflow, nil, data*2)
		if err != nil {
			return -1, err
		}
		return 0, nil // This line won't be executed
	}

	if err := ctx.Workflow("subworkflow-step", SubWorkflow, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, data).Get(&value); err != nil {
		log.Printf("Workflow encountered error from sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("Workflow received value from sub-workflow: %d", value)

	// Build and execute a saga
	orderData := value
	sagaBuilder := tempolite.NewSaga()
	sagaBuilder.AddStep(ReserveInventorySaga{Data: orderData})
	sagaBuilder.AddStep(ProcessPaymentSaga{Data: orderData})
	sagaBuilder.AddStep(UpdateLedgerSaga{Data: orderData})

	saga, err := sagaBuilder.Build()
	if err != nil {
		return -1, fmt.Errorf("failed to build saga: %w", err)
	}

	err = ctx.Saga("process-order", saga).Get()
	if err != nil {
		return -1, fmt.Errorf("saga execution failed: %w", err)
	}

	result := value + data

	var a int
	var b int
	if err := ctx.Workflow("a-b", SubManyWorkflow, nil).Get(&a, &b); err != nil {
		log.Printf("Workflow encountered error from sub-many-workflow: %v", err)
		return -1, err
	}

	log.Printf("Workflow returning result: %d - %v %v", result, a, b)
	return result, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("main started")

	database := tempolite.NewDefaultDatabase()
	registry := tempolite.NewRegistry()
	registry.RegisterWorkflow(Workflow) // Only register the root workflow

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a new orchestrator
	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	subWorkflowFailed.Store(true)
	continueAsNewCalled.Store(false)

	future := orchestrator.Workflow(Workflow, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, 40)

	// Simulate pause after 2 seconds
	<-time.After(2 * time.Second)
	log.Printf("\tPausing orchestrator")
	orchestrator.Pause()

	// Wait for some time to simulate the orchestrator being paused
	fmt.Println("\t\tWAITING 10 SECONDS")
	time.Sleep(10 * time.Second)

	log.Printf("\tResuming orchestrator")
	// Create a new orchestrator and resume
	newOrchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	future = newOrchestrator.Resume(future.WorkflowID())

	var result int
	if err := future.Get(&result); err != nil {
		log.Printf("Workflow failed with error: %v", err)
	} else {
		log.Printf("Workflow completed with result: %d", result)
	}

	newOrchestrator.Wait()

	log.Printf("Retrying workflow")
	retryFuture := newOrchestrator.Retry(future.WorkflowID())

	if err := retryFuture.Get(&result); err != nil {
		log.Printf("Retried workflow failed with error: %v", err)
	} else {
		log.Printf("Retried workflow completed with result: %d", result)
	}

	// After execution, clear completed Runs to free up memory
	database.Clear()

	log.Printf("main finished")
}
