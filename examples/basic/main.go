package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/go-tempolite"
)

type CustomIdentifier string

func SimpleActivity(ctx tempolite.ActivityContext[CustomIdentifier], input string) (string, error) {
	return fmt.Sprintf("Processed: %s", input), nil
}

func SimpleWorkflow(ctx tempolite.WorkflowContext[CustomIdentifier], input string) (string, error) {
	var result string
	// Inspired by how Ansible can name their steps, you use the stepID to identify the step
	// It becomes pretty useful when you have tons of activities in your workflow you quickly know what is doing what
	err := ctx.Activity("simple-activity", SimpleActivity, input).Get(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func main() {
	// Create a new Tempolite instance with custom options
	tp, err := tempolite.New[CustomIdentifier]( // in a minimalistic workflow engine while needed consistency without big algorithms, you can use a custom identifier, could be just a `string` or `int`
		context.Background(),
		tempolite.NewRegistry[CustomIdentifier](). // you can create a registry of workflows and activities that you can then re-use on another Tempolite instance
								Workflow(SimpleWorkflow).
								Activity(SimpleActivity).
								Build(),
		tempolite.WithPath("./tempolite.db"),      // once you executed it once, go check the db!
		tempolite.WithDestructive(),               // means it will attempt to destroy the previous path db at starts
		tempolite.WithInitialWorkflowsWorkers(10), // completely overkill
		tempolite.WithInitialActivityWorkers(10),  // completely overkill
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Execute the workflow with retry options
	var output string
	err = tp.Workflow("simple-workflow", SimpleWorkflow, tempolite.WorkflowConfig(
		tempolite.WithRetryMaximumAttempts(3),
	), "Hello, Tempolite!").Get(&output)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Println("Workflow result:", output)

	if err := tp.Wait(); err != nil {
		log.Fatalf("Error waiting for tasks to complete: %v", err)
	}
}
