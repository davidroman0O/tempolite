package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davidroman0O/tempolite"
)

type Order struct {
	ID              string
	UserID          string
	Items           []string
	Total           float64
	Status          string
	ShippingAddress string
}

// Activity as a struct
type InventoryChecker struct{}

func (ic InventoryChecker) Run(ctx tempolite.ActivityContext, items []string) (bool, error) {
	log.Printf("Checking inventory for items: %v", items)
	// Simulate inventory check
	time.Sleep(1 * time.Second)
	return true, nil
}

// Activity as a function
func ProcessPayment(ctx tempolite.ActivityContext, orderID string, amount float64) error {
	log.Printf("Processing payment for order %s: $%.2f", orderID, amount)
	// Simulate payment processing
	time.Sleep(2 * time.Second)
	return nil
}

// Saga step: Reserve Inventory
type ReserveInventorySaga struct {
	Items []string
}

func (s ReserveInventorySaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("Reserving inventory for items: %v", s.Items)
	return "Inventory reserved", nil
}

func (s ReserveInventorySaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("Compensating: Releasing reserved inventory for items: %v", s.Items)
	return "Inventory released", nil
}

// Saga step: Update Order Status
type UpdateOrderStatusSaga struct {
	OrderID string
	Status  string
}

func (s UpdateOrderStatusSaga) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	log.Printf("Updating order status: OrderID=%s, Status=%s", s.OrderID, s.Status)
	return "Order status updated", nil
}

func (s UpdateOrderStatusSaga) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	log.Printf("Compensating: Reverting order status for OrderID=%s", s.OrderID)
	return "Order status reverted", nil
}

// Main workflow
func OrderProcessingWorkflow(ctx tempolite.WorkflowContext, order Order) error {
	log.Printf("Starting order processing for Order ID: %s", order.ID)

	// Version check for new shipping address feature
	version := ctx.GetVersion("ShippingAddressFeature", tempolite.DefaultVersion, 1)
	if version == tempolite.DefaultVersion {
		log.Println("Using old order processing logic")
	} else {
		log.Println("Using new order processing logic with shipping address")
		// In this version, we would handle the shipping address
		log.Printf("Shipping address: %s", order.ShippingAddress)
	}

	// Check inventory using struct-based activity
	var inStock bool
	inventoryChecker := InventoryChecker{}
	if err := ctx.Activity("check-inventory", inventoryChecker.Run, nil, order.Items).Get(&inStock); err != nil {
		return fmt.Errorf("inventory check failed: %w", err)
	}

	if !inStock {
		return fmt.Errorf("items out of stock")
	}

	// Process payment using function-based activity
	if err := ctx.Activity("process-payment", ProcessPayment, nil, order.ID, order.Total).Get(); err != nil {
		return fmt.Errorf("payment processing failed: %w", err)
	}

	// Use a side effect to generate a unique tracking number
	var trackingNumber string
	if err := ctx.SideEffect("generate-tracking", func(ctx tempolite.SideEffectContext) string {
		return fmt.Sprintf("TRK-%d", rand.Intn(1000000))
	}).Get(&trackingNumber); err != nil {
		return fmt.Errorf("failed to generate tracking number: %w", err)
	}

	log.Printf("Generated tracking number: %s", trackingNumber)

	// Use a saga to handle the final steps
	sagaBuilder := tempolite.NewSaga()
	sagaBuilder.AddStep(ReserveInventorySaga{Items: order.Items})
	sagaBuilder.AddStep(UpdateOrderStatusSaga{OrderID: order.ID, Status: "Processing"})
	saga, _ := sagaBuilder.Build()

	if err := ctx.Saga("finalize-order", saga).Get(); err != nil {
		return fmt.Errorf("order finalization failed: %w", err)
	}

	// Wait for a signal to confirm shipment
	shipmentSignal := ctx.Signal("shipment-confirmation")
	var shipmentConfirmed bool
	if err := shipmentSignal.Receive(ctx, &shipmentConfirmed); err != nil {
		return fmt.Errorf("failed to receive shipment confirmation: %w", err)
	}

	if shipmentConfirmed {
		log.Printf("Order %s has been shipped with tracking number %s", order.ID, trackingNumber)
	} else {
		return fmt.Errorf("shipment confirmation failed for order %s", order.ID)
	}

	return nil
}

func main() {
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(OrderProcessingWorkflow).
			Activity(InventoryChecker{}.Run).
			Activity(ProcessPayment).
			Build(),
		tempolite.WithPath("./db/templite-advanced.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	order := Order{
		ID:              "ORD-001",
		UserID:          "USER-123",
		Items:           []string{"Item1", "Item2"},
		Total:           99.99,
		Status:          "Pending",
		ShippingAddress: "123 Main St, Anytown, USA",
	}

	workflowInfo := tp.Workflow(OrderProcessingWorkflow, nil, order)

	// Simulate shipment confirmation signal
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("Sending shipment confirmation signal...")
		if err := tp.PublishSignal(workflowInfo.WorkflowID, "shipment-confirmation", true); err != nil {
			log.Printf("Failed to publish shipment confirmation signal: %v", err)
		}
	}()

	if err := workflowInfo.Get(); err != nil {
		log.Printf("Workflow execution failed: %v", err)
	} else {
		log.Println("Workflow completed successfully on first attempt")
	}

	log.Println("Verifying the good behaviour of the workflow...")
	replayInfo := tp.ReplayWorkflow(workflowInfo.WorkflowID)
	if err := replayInfo.Get(); err != nil {
		log.Fatalf("Workflow replay failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflow to complete: %v", err)
	}

	log.Println("E-commerce order processing completed!")
}
