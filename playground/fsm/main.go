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

// Example Activity
func ProcessOrder(ctx tempolite.ActivityContext, orderID int) (int, error) {
	log.Printf("ProcessOrder called with orderID: %d", orderID)
	select {
	case <-time.After(time.Second * 2):
		// Simulate processing
	case <-ctx.Done():
		log.Printf("ProcessOrder context cancelled")
		return -1, ctx.Err()
	}
	result := orderID * 2
	log.Printf("ProcessOrder completed for orderID: %d with result: %d", orderID, result)
	return result, nil
}

// Nested activity for deeper workflow testing
func ComputePrice(ctx tempolite.ActivityContext, basePrice int) (int, error) {
	log.Printf("ComputePrice called with base price: %d", basePrice)
	select {
	case <-time.After(time.Second):
	case <-ctx.Done():
		return -1, ctx.Err()
	}
	return basePrice * 3, nil
}

// Sub-sub-workflow for deep nesting
func ComputeTaxWorkflow(ctx tempolite.WorkflowContext, amount int) (int, error) {
	log.Printf("ComputeTaxWorkflow started with amount: %d", amount)

	var computedPrice int
	if err := ctx.Activity("compute-price", ComputePrice, &tempolite.ActivityOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        time.Minute,
		},
	}, amount).Get(&computedPrice); err != nil {
		return -1, fmt.Errorf("failed to compute price: %w", err)
	}

	return computedPrice + (computedPrice / 10), nil // Add 10% tax
}

// Sub-workflow that can fail
var subWorkflowShouldFail atomic.Bool

func ValidateOrderWorkflow(ctx tempolite.WorkflowContext, orderID int) (int, error) {
	log.Printf("ValidateOrderWorkflow started with orderID: %d", orderID)

	if subWorkflowShouldFail.Load() {
		return -1, fmt.Errorf("validation failed on purpose")
	}

	var price int
	if err := ctx.Workflow("compute-tax", ComputeTaxWorkflow, nil, orderID).Get(&price); err != nil {
		return -1, fmt.Errorf("tax computation failed: %w", err)
	}

	return price, nil
}

// Saga implementations
type ValidateInventorySaga struct {
	OrderID int
	Amount  int
}

func (s ValidateInventorySaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("ValidateInventorySaga Transaction called with OrderID: %d, Amount: %d", s.OrderID, s.Amount)
	return nil, nil
}

func (s ValidateInventorySaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("ValidateInventorySaga Compensation called for OrderID: %d", s.OrderID)
	return nil, nil
}

type ProcessPaymentSaga struct {
	OrderID int
	Amount  int
}

func (s ProcessPaymentSaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Transaction called with OrderID: %d, Amount: %d", s.OrderID, s.Amount)
	return nil, nil
}

func (s ProcessPaymentSaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("ProcessPaymentSaga Compensation called for OrderID: %d", s.OrderID)
	return nil, nil
}

// Main workflow with versioning, side effects, and sagas
var continueAsNewCalled atomic.Bool

func ProcessOrderWorkflow(ctx tempolite.WorkflowContext, orderID int) (int, error) {
	log.Printf("ProcessOrderWorkflow started with orderID: %d", orderID)

	// Version check example
	changeID := "OrderProcessingV2"
	version := ctx.GetVersion(changeID, 0, 1)
	if version > 0 {
		log.Printf("Using new order processing version")
	}

	// Test ContinueAsNew
	if !continueAsNewCalled.Load() {
		continueAsNewCalled.Store(true)
		log.Printf("Using ContinueAsNew to restart workflow")
		err := ctx.ContinueAsNew(nil, orderID*2)
		if err != nil {
			return -1, err
		}
		return 0, nil
	}

	// Side effect example
	var shouldPrioritize bool
	if err := ctx.SideEffect("check-priority", func() bool {
		return rand.Float32() < 0.3
	}).Get(&shouldPrioritize); err != nil {
		return -1, fmt.Errorf("failed to check priority: %w", err)
	}

	log.Printf("Order %d priority status: %v", orderID, shouldPrioritize)

	// Execute sub-workflow that might fail
	var validatedPrice int
	if err := ctx.Workflow("validate-order", ValidateOrderWorkflow, &tempolite.WorkflowOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        2,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, orderID).Get(&validatedPrice); err != nil {
		return -1, fmt.Errorf("order validation failed: %w", err)
	}

	// Activity execution
	var processedOrder int
	if err := ctx.Activity("process-order", ProcessOrder, &tempolite.ActivityOptions{
		RetryPolicy: &tempolite.RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        5 * time.Minute,
		},
	}, orderID).Get(&processedOrder); err != nil {
		return -1, fmt.Errorf("failed to process order: %w", err)
	}

	// Saga execution
	saga := tempolite.NewSaga()
	saga.AddStep(ValidateInventorySaga{OrderID: orderID, Amount: validatedPrice})
	saga.AddStep(ProcessPaymentSaga{OrderID: orderID, Amount: processedOrder})

	sagaDef, err := saga.Build()
	if err != nil {
		return -1, fmt.Errorf("failed to build saga: %w", err)
	}

	if err := ctx.Saga("order-transaction", sagaDef).Get(); err != nil {
		return -1, fmt.Errorf("saga execution failed: %w", err)
	}

	return processedOrder + validatedPrice, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Starting Tempolite example")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Initialize engine with multiple queues
	engine, _ := tempolite.New(
		ctx,
		tempolite.NewDefaultDatabase(),
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:        "high-priority",
			WorkerCount: 3,
		}),
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:        "low-priority",
			WorkerCount: 2,
		}),
		tempolite.WithDefaultQueueWorkers(2),
	)
	defer engine.Close()

	// Register workflows
	if err := engine.RegisterWorkflow(ProcessOrderWorkflow); err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}
	if err := engine.RegisterWorkflow(ValidateOrderWorkflow); err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}
	if err := engine.RegisterWorkflow(ComputeTaxWorkflow); err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}

	// Test scenario 1: Normal execution across queues
	log.Printf("=== Scenario 1: Normal execution ===")
	queues := []string{"default", "high-priority", "low-priority"}
	futures := make([]*tempolite.RuntimeFuture, 0)

	for i := 0; i < 2; i++ {
		for _, queueName := range queues {
			var future *tempolite.RuntimeFuture
			if queueName == "default" {
				future = engine.Workflow(ProcessOrderWorkflow, nil, i+1)
			} else {
				future = engine.ExecuteWorkflow(queueName, ProcessOrderWorkflow, nil, i+1)
			}
			futures = append(futures, future)
			log.Printf("Started workflow on queue %s", queueName)
		}
	}

	// Test scenario 2: Failure and retry
	log.Printf("\n=== Scenario 2: Failure and retry ===")
	subWorkflowShouldFail.Store(true)
	failureFuture := engine.ExecuteWorkflow("high-priority", ProcessOrderWorkflow, nil, 100)
	futures = append(futures, failureFuture)

	// Monitor queue status
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, queueName := range queues {
					info, err := engine.GetQueueInfo(queueName)
					if err != nil {
						log.Printf("Failed to get queue info for %s: %v", queueName, err)
						continue
					}
					log.Printf("Queue %s - Workers: %d, Pending: %d, Processing: %d, Failed: %d",
						info.Name, info.WorkerCount, info.PendingTasks, info.ProcessingTasks, info.FailedTasks)
				}
			}
		}
	}()

	// Wait for results
	for i, future := range futures {
		var result int
		if err := future.Get(&result); err != nil {
			log.Printf("Workflow %d failed: %v", i, err)
		} else {
			log.Printf("Workflow %d completed with result: %d", i, result)
		}
	}

	// Test scenario 3: Dynamic scaling
	log.Printf("\n=== Scenario 3: Dynamic scaling ===")
	log.Printf("Adding workers to high-priority queue")
	if err := engine.AddQueueWorkers("high-priority", 2); err != nil {
		log.Printf("Failed to add workers: %v", err)
	}

	// Run another workflow after scaling
	subWorkflowShouldFail.Store(false) // Allow it to succeed
	scaledFuture := engine.ExecuteWorkflow("high-priority", ProcessOrderWorkflow, nil, 200)

	var finalResult int
	if err := scaledFuture.Get(&finalResult); err != nil {
		log.Printf("Scaled workflow failed: %v", err)
	} else {
		log.Printf("Scaled workflow completed with result: %d", finalResult)
	}

	// Cleanup
	log.Printf("Removing workers from high-priority queue")
	if err := engine.RemoveQueueWorkers("high-priority", 1); err != nil {
		log.Printf("Failed to remove workers: %v", err)
	}

	log.Printf("Example completed")

	if err := engine.Wait(); err != nil {
		log.Fatalf("Failed to wait for engine: %v", err)
	}

}
