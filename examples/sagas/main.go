package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/davidroman0O/go-tempolite"
)

type CustomIdentifier string

// OrderData represents the data for an order
type OrderData struct {
	OrderID string
	Amount  float64
}

// ReserveInventorySaga represents the first step in the saga
type ReserveInventorySaga struct {
	Data OrderData
}

func (s ReserveInventorySaga) Transaction(ctx tempolite.TransactionContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Reserving inventory for order %s", s.Data.OrderID)
	// Simulate inventory reservation
	if rand.Float32() < 0.3 { // 30% chance of failure
		return nil, errors.New("inventory reservation failed: out of stock")
	}
	return "Inventory reserved", nil
}

func (s ReserveInventorySaga) Compensation(ctx tempolite.CompensationContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Compensating: Releasing reserved inventory for order %s", s.Data.OrderID)
	return "Inventory released", nil
}

// ProcessPaymentSaga represents the second step in the saga
type ProcessPaymentSaga struct {
	Data OrderData
}

func (s ProcessPaymentSaga) Transaction(ctx tempolite.TransactionContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Processing payment for order %s, amount %.2f", s.Data.OrderID, s.Data.Amount)
	if s.Data.Amount > 1000 {
		return nil, errors.New("payment declined: amount too high")
	}
	return "Payment processed", nil
}

func (s ProcessPaymentSaga) Compensation(ctx tempolite.CompensationContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Compensating: Refunding payment for order %s", s.Data.OrderID)
	return "Payment refunded", nil
}

// UpdateLedgerSaga represents the third step in the saga
type UpdateLedgerSaga struct {
	Data OrderData
}

func (s UpdateLedgerSaga) Transaction(ctx tempolite.TransactionContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Updating ledger for order %s", s.Data.OrderID)
	// Simulate ledger update
	if rand.Float32() < 0.2 { // 20% chance of failure
		return nil, errors.New("ledger update failed: database error")
	}
	return "Ledger updated", nil
}

func (s UpdateLedgerSaga) Compensation(ctx tempolite.CompensationContext[CustomIdentifier]) (interface{}, error) {
	log.Printf("Compensating: Reverting ledger update for order %s", s.Data.OrderID)
	return "Ledger update reverted", nil
}

// OrderWorkflow is the main workflow function
func OrderWorkflow(ctx tempolite.WorkflowContext[CustomIdentifier], orderData OrderData) (string, error) {
	log.Printf("Starting OrderWorkflow for order %s", orderData.OrderID)

	sagaBuilder := tempolite.NewSaga[CustomIdentifier]()
	sagaBuilder.AddStep(ReserveInventorySaga{Data: orderData})
	sagaBuilder.AddStep(ProcessPaymentSaga{Data: orderData})
	sagaBuilder.AddStep(UpdateLedgerSaga{Data: orderData})

	saga, err := sagaBuilder.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build saga: %w", err)
	}

	err = ctx.Saga("process-order", saga).Get()
	if err != nil {
		return "", fmt.Errorf("saga execution failed: %w", err)
	}

	return fmt.Sprintf("Order %s processed successfully", orderData.OrderID), nil
}

func main() {
	tp, err := tempolite.New[CustomIdentifier](
		context.Background(),
		tempolite.NewRegistry[CustomIdentifier]().
			Workflow(OrderWorkflow).
			Build(),
		tempolite.WithPath("./db/tempolite-example-saga.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Run multiple orders to demonstrate compensation
	for i := 1; i <= 5; i++ {
		orderData := OrderData{
			OrderID: fmt.Sprintf("ORD-%03d", i),
			Amount:  float64(500 + i*100), // Varying amounts
		}

		var result string
		err = tp.Workflow(CustomIdentifier(fmt.Sprintf("order-workflow-%d", i)), OrderWorkflow, orderData).Get(&result)
		if err != nil {
			log.Printf("Workflow execution failed for order %s: %v", orderData.OrderID, err)
		} else {
			log.Printf("Workflow result for order %s: %s", orderData.OrderID, result)
		}
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for tasks to complete: %v", err)
	}
}
