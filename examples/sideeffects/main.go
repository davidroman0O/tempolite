package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/davidroman0O/tempolite"
)

func RandomNumberActivity(ctx tempolite.ActivityContext, max int) (int, error) {
	return rand.Intn(max), nil
}

func ProcessNumberActivity(ctx tempolite.ActivityContext, num int) (string, error) {
	if num%2 == 0 {
		return fmt.Sprintf("%d is even", num), nil
	}
	return fmt.Sprintf("%d is odd", num), nil
}

func ComplexWorkflow(ctx tempolite.WorkflowContext, maxNumber int) (string, error) {
	var randomNumber int
	err := ctx.Activity("random-number", RandomNumberActivity, maxNumber).Get(&randomNumber)
	if err != nil {
		return "", err
	}

	var shouldDouble bool
	err = ctx.SideEffect("should-double", func(ctx tempolite.SideEffectContext) bool {
		return rand.Float32() < 0.5
	}).Get(&shouldDouble)
	if err != nil {
		return "", err
	}

	if shouldDouble {
		randomNumber *= 2
	}

	var result string
	err = ctx.Activity("process-number", ProcessNumberActivity, randomNumber).Get(&result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func main() {
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(ComplexWorkflow).
			Activity(RandomNumberActivity).
			Activity(ProcessNumberActivity).
			Build(),
		tempolite.WithPath("./db/tempolite-side-effects.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	var result string
	err = tp.Workflow(ComplexWorkflow, nil, 100).Get(&result)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Println("Workflow result:", result)

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for tasks to complete: %v", err)
	}
}
