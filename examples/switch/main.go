package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite"
)

// Different workflow types and their data
type OrderData struct {
	OrderID string
	Amount  float64
}

type ShipmentData struct {
	TrackingID string
	Address    string
}

type RefundData struct {
	OrderID string
	Reason  string
}

// Workflow handler functions
func OrderWorkflow(ctx tempolite.WorkflowContext, data OrderData) error {
	log.Printf("Processing order: %s for $%.2f", data.OrderID, data.Amount)
	return ctx.Activity("process-payment", ProcessPayment, data.OrderID, data.Amount).Get()
}

func ShipmentWorkflow(ctx tempolite.WorkflowContext, data ShipmentData) error {
	log.Printf("Processing shipment: %s to %s", data.TrackingID, data.Address)
	return ctx.Activity("ship-order", ShipOrder, data.TrackingID, data.Address).Get()
}

func RefundWorkflow(ctx tempolite.WorkflowContext, data RefundData) error {
	log.Printf("Processing refund for order: %s, reason: %s", data.OrderID, data.Reason)
	return ctx.Activity("process-refund", ProcessRefund, data.OrderID).Get()
}

// Activity handlers
func ProcessPayment(ctx tempolite.ActivityContext, orderID string, amount float64) error {
	log.Printf("Processing payment for order %s: $%.2f", orderID, amount)
	time.Sleep(1 * time.Second) // Simulate payment processing
	return nil
}

func ShipOrder(ctx tempolite.ActivityContext, trackingID, address string) error {
	log.Printf("Shipping order %s to %s", trackingID, address)
	time.Sleep(1 * time.Second) // Simulate shipping process
	return nil
}

func ProcessRefund(ctx tempolite.ActivityContext, orderID string) error {
	log.Printf("Processing refund for order %s", orderID)
	time.Sleep(1 * time.Second) // Simulate refund process
	return nil
}

// Helper function to process workflow based on type
func processWorkflow(info *tempolite.WorkflowInfo) {
	info.Switch().
		Case(OrderWorkflow, func(wi *tempolite.WorkflowInfo) {
			log.Printf("Handling Order workflow %s", wi.WorkflowID)
			var err error
			wi.Get(&err) // Get result
			if err != nil {
				log.Printf("Order workflow failed: %v", err)
			}
		}).
		Case(ShipmentWorkflow, func(wi *tempolite.WorkflowInfo) {
			log.Printf("Handling Shipment workflow %s", wi.WorkflowID)
			var err error
			wi.Get(&err) // Get result
			if err != nil {
				log.Printf("Shipment workflow failed: %v", err)
			}
		}).
		Case(RefundWorkflow, func(wi *tempolite.WorkflowInfo) {
			log.Printf("Handling Refund workflow %s", wi.WorkflowID)
			var err error
			wi.Get(&err) // Get result
			if err != nil {
				log.Printf("Refund workflow failed: %v", err)
			}
		}).
		Default(func(wi *tempolite.WorkflowInfo) {
			log.Printf("Unknown workflow type for %s", wi.WorkflowID)
		}).
		Execute()
}

func main() {
	// Initialize Tempolite
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(OrderWorkflow).
			Workflow(ShipmentWorkflow).
			Workflow(RefundWorkflow).
			Activity(ProcessPayment).
			Activity(ShipOrder).
			Activity(ProcessRefund).
			Build(),
		tempolite.WithPath("./db/tempolite-handler-switch.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Start different types of workflows
	orderWorkflowInfo := tp.Workflow(OrderWorkflow, nil, OrderData{
		OrderID: "ORD-001",
		Amount:  99.99,
	})

	shipmentWorkflowInfo := tp.Workflow(ShipmentWorkflow, nil, ShipmentData{
		TrackingID: "TRK-001",
		Address:    "123 Main St, City, Country",
	})

	refundWorkflowInfo := tp.Workflow(RefundWorkflow, nil, RefundData{
		OrderID: "ORD-001",
		Reason:  "Customer request",
	})

	// Process each workflow
	fmt.Println("\nProcessing Order Workflow:")
	processWorkflow(orderWorkflowInfo)

	fmt.Println("\nProcessing Shipment Workflow:")
	processWorkflow(shipmentWorkflowInfo)

	fmt.Println("\nProcessing Refund Workflow:")
	processWorkflow(refundWorkflowInfo)

	// Try with an unknown workflow (should hit default case)
	unknownWorkflowInfo := &tempolite.WorkflowInfo{
		WorkflowID: tempolite.WorkflowID("unknown-1"),
	}
	fmt.Println("\nProcessing Unknown Workflow:")
	processWorkflow(unknownWorkflowInfo)

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}
}
