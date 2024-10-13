// File: ./examples/versioning/main.go

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/go-tempolite"
)

// Define the change IDs for versioning
const (
	ChangeIDCalculateTax = "CalculateTotalWithTax"
)

type identifier string

const (
	ComputeTaxes = "ComputeTaxes"
)

// OrderWorkflow is the main workflow function
func OrderWorkflow(ctx tempolite.WorkflowContext[identifier], orderID string) error {
	total, err := calculateTotal(ctx, orderID)
	if err != nil {
		return err
	}

	fmt.Printf("Order %s total: $%.2f\n", orderID, total)
	// Proceed with further processing, e.g., charging the customer
	return nil
}

func calculateTotal(ctx tempolite.WorkflowContext[identifier], orderID string) (float64, error) {
	// Use GetVersion to manage changes in the workflow logic
	version := ctx.GetVersion(ChangeIDCalculateTax, tempolite.DefaultVersion, getMaxSupportedVersion())
	var total float64

	if version == tempolite.DefaultVersion {
		// Original logic: total without tax
		total = getOrderSubtotal(orderID)
		fmt.Println("Using original logic: total without tax.")
	} else {
		// New logic: total with tax
		subtotal := getOrderSubtotal(orderID)
		tax := calculateTax(subtotal)
		total = subtotal + tax
		fmt.Println("Using new logic: total with tax.")
	}

	return total, nil
}

func getOrderSubtotal(orderID string) float64 {
	// Placeholder for fetching the order subtotal
	return 100.0 // Assume a subtotal of $100
}

func calculateTax(subtotal float64) float64 {
	// Assume a tax rate of 10%
	return subtotal * 0.10
}

// getMaxSupportedVersion simulates updating the version in code
var currentMaxSupportedVersion int = tempolite.DefaultVersion

func getMaxSupportedVersion() int {
	return currentMaxSupportedVersion
}

func main() {
	// Create a new Tempolite instance with a destructive option to reset the database
	ctx := context.Background()
	tp, err := tempolite.New[identifier](ctx)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Register the workflow
	if err := tp.RegisterWorkflow(OrderWorkflow); err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}

	// Enqueue the workflow before updating the version
	orderID1 := "order123"
	if err := tp.Workflow(ComputeTaxes, OrderWorkflow, orderID1).Get(); err != nil {
		log.Fatalf("Failed to enqueue workflow: %v", err)
	}

	// Wait for the workflow to complete
	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}

	// Simulate updating the version in the code
	// For example, change the maxSupported version for the ChangeID
	fmt.Println("\n--- Updating workflow logic to include tax ---\n")
	currentMaxSupportedVersion = 1

	// Enqueue the workflow after updating the version
	orderID2 := "order456"
	if err := tp.Workflow(ComputeTaxes, OrderWorkflow, orderID2).Get(); err != nil {
		log.Fatalf("Failed to enqueue workflow: %v", err)
	}

	// Wait for the new workflow to complete
	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}
}
