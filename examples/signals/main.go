package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite"
)

type CoffeeShopID string

type CustomerOrder struct {
	OrderID   string
	DrinkType string
	Size      string
}

func CoffeeShopWorkflow(ctx tempolite.WorkflowContext[CoffeeShopID], order CustomerOrder) error {
	log.Printf("Received order: %+v", order)

	// Prepare the drink
	if err := ctx.Activity("prepare-drink", PrepareDrink, order).Get(); err != nil {
		return fmt.Errorf("failed to prepare drink: %w", err)
	}

	// Wait for the barista to signal that the drink is ready
	drinkReadySignal := ctx.Signal("drink-ready")
	var isReady bool
	if err := drinkReadySignal.Receive(ctx, &isReady); err != nil {
		return fmt.Errorf("failed to receive drink-ready signal: %w", err)
	}

	if !isReady {
		return fmt.Errorf("drink preparation failed")
	}

	log.Printf("Drink for order %s is ready!", order.OrderID)

	// Wait for the customer to pick up the drink
	customerSignal := ctx.Signal("customer-pickup")
	var customerFeedback string
	if err := customerSignal.Receive(ctx, &customerFeedback); err != nil {
		return fmt.Errorf("failed to receive customer-pickup signal: %w", err)
	}

	log.Printf("Customer feedback for order %s: %s", order.OrderID, customerFeedback)

	return nil
}

func PrepareDrink(ctx tempolite.ActivityContext[CoffeeShopID], order CustomerOrder) error {
	log.Printf("Preparing %s %s for order %s", order.Size, order.DrinkType, order.OrderID)
	time.Sleep(3 * time.Second) // Simulating drink preparation time
	return nil
}

func main() {
	tp, err := tempolite.New[CoffeeShopID](
		context.Background(),
		tempolite.NewRegistry[CoffeeShopID]().
			Workflow(CoffeeShopWorkflow).
			Activity(PrepareDrink).
			Build(),
		tempolite.WithPath("./db/tempolite-signals.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	order := CustomerOrder{
		OrderID:   "ORDER-001",
		DrinkType: "Latte",
		Size:      "Grande",
	}

	workflowInfo := tp.Workflow("coffee-order", CoffeeShopWorkflow, nil, order)

	// Simulate barista signaling that the drink is ready
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Barista: The drink is ready!")
		if err := tp.PublishSignal(workflowInfo.WorkflowID, "drink-ready", true); err != nil {
			log.Printf("Failed to publish drink-ready signal: %v", err)
		}
	}()

	// Simulate customer picking up the drink and providing feedback
	go func() {
		time.Sleep(8 * time.Second)
		log.Println("Customer: Picking up the drink...")
		feedback := "This latte is so good, I think I can see the future!"
		if err := tp.PublishSignal(workflowInfo.WorkflowID, "customer-pickup", feedback); err != nil {
			log.Printf("Failed to publish customer-pickup signal: %v", err)
		}
	}()

	if err := workflowInfo.Get(); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for workflow to complete: %v", err)
	}

	log.Println("Coffee shop workflow completed successfully!")
}
