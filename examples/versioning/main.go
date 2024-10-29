package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/davidroman0O/tempolite"
)

// Define the change IDs for versioning
const (
	ChangeIDCalculateTax = "CalculateTotalWithTax"
)

var codeAfterUpdateOneRuntime atomic.Bool
var codeAfterUpdateTwoRuntime atomic.Bool
var errOnPurpose atomic.Bool
var panicOnPurpose atomic.Bool
var panicActivity atomic.Bool

// OrderWorkflow is the main workflow function
func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
	var err error
	var total float64

	if codeAfterUpdateOneRuntime.Load() { // on purpose to create non-deterministic behavior
		version := ctx.GetVersion(ChangeIDCalculateTax, tempolite.DefaultVersion, 1)
		if version == tempolite.DefaultVersion {
			err = ctx.Activity("taxes", ActivityComputeTaxes, nil, orderID).Get(&total)
			fmt.Println("Using original logic after update: total without tax.")
		} else {
			err = ctx.Activity("taxes", ActivityComputeTotalWithTax, nil, orderID, total).Get(&total)
			fmt.Println("Using new logic: total with tax.")
		}
	} else {
		err = ctx.Activity("taxes", ActivityComputeTaxes, nil, orderID).Get(&total)
		fmt.Println("Using original logic before update: total without tax.")
	}

	if err != nil {
		return err
	}

	fmt.Printf("Order %s total: $%.2f\n", orderID, total)
	// Proceed with further processing, e.g., charging the customer
	return nil
}

func ActivityComputeTaxes(ctx tempolite.ActivityContext, orderID string) (float64, error) {
	if errOnPurpose.Load() {
		errOnPurpose.Store(false)
		return 0, fmt.Errorf("error on purpose ActivityComputeTaxes")
	}
	if panicActivity.Load() {
		panicActivity.Store(false)
		panic("panic on purpose ActivityComputeTaxes")
	}
	return getOrderSubtotal(orderID), nil
}

func ActivityComputeTotalWithTax(ctx tempolite.ActivityContext, orderID string, subtotal float64) (float64, error) {
	if errOnPurpose.Load() {
		errOnPurpose.Store(false)
		return 0, fmt.Errorf("error on purpose ActivityComputeTotalWithTax")
	}
	if panicActivity.Load() {
		panicActivity.Store(false)
		panic("panic on purpose ActivityComputeTotalWithTax")
	}
	subtotal = getOrderSubtotal(orderID)
	tax := calculateTax(subtotal)
	return subtotal + tax, nil

}

func getOrderSubtotal(orderID string) float64 {
	if panicOnPurpose.Load() {
		panicOnPurpose.Store(false)
		panic("panic on purpose getOrderSubtotal")
	}
	// Placeholder for fetching the order subtotal
	return 100.0 // Assume a subtotal of $100
}

func calculateTax(subtotal float64) float64 {
	if panicOnPurpose.Load() {
		panicOnPurpose.Store(false)
		panic("panic on purpose calculateTax")
	}
	// Assume a tax rate of 10%
	return subtotal * 0.10
}

func main() {
	// Create a new Tempolite instance with a destructive option to reset the database
	ctx := context.Background()
	tp, err := tempolite.New(
		ctx,
		tempolite.NewRegistry().
			Workflow(OrderWorkflow).
			Activity(ActivityComputeTaxes).
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	errOnPurpose.Store(true) // simulate an error in the activity

	// Enqueue the workflow before updating the version
	orderID1 := "order123"
	if err := tp.Workflow(OrderWorkflow, nil, orderID1).Get(); err != nil {
		log.Fatalf("Failed to enqueue workflow: %v", err)
	}

	// Wait for the workflow to complete
	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}

	orderID1 = "order124"
	tp.Workflow(OrderWorkflow, nil, orderID1) // on purpose, it won't be scheduled, but the next instance will pick it up

	tp.Close()

	codeAfterUpdateOneRuntime.Store(true) // simulate deploying new code

	<-time.After(1 * time.Second)

	// Simulate updating the version in the code
	// For example, change the maxSupported version for the ChangeID
	fmt.Println("\n--- Updating workflow logic to include tax ---\n")

	ctx = context.Background()
	tp, err = tempolite.New(
		ctx,
		tempolite.NewRegistry().
			Workflow(OrderWorkflow).
			Activity(ActivityComputeTaxes).
			Activity(ActivityComputeTotalWithTax).
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	panicActivity.Store(true) // simulate a panic in the activity

	// Enqueue the workflow after updating the version
	orderID2 := "order456"
	if err := tp.Workflow(OrderWorkflow, nil, orderID2).Get(); err != nil {
		log.Fatalf("Failed to enqueue workflow: %v", err)
	}

	// Wait for the new workflow to complete
	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflows to complete: %v", err)
	}
}
