package tests

import (
	"context"
	"fmt"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

func TestWorkflowCrossWorkflowsActivitiesQueue(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	defer database.SaveAsJSON("./jsons/workflows_activities_cross_queue.json")

	var orchestratorDefault *tempolite.Orchestrator
	var orchestratorSide *tempolite.Orchestrator

	if _, err := database.AddQueue(&tempolite.Queue{
		Name: "side",
	}); err != nil {
		t.Fatal(err)
	}

	// let's create the other queue
	orchestratorSide = tempolite.NewOrchestrator(
		ctx,
		database,
		registry,
		tempolite.WithQueueName("side"),
	)

	// thta will be our main one
	orchestratorDefault = tempolite.NewOrchestrator(
		ctx,
		database,
		registry,
		// CrossWorkflow feature require a callback to be used on that orchestrator - means unusable on the side orchestrator
		tempolite.WithCrossWorkflow(
			func(
				queueName string,
				workflowID tempolite.WorkflowEntityID,
				workflowFunc interface{},
				options *tempolite.WorkflowOptions,
				args ...interface{},
			) tempolite.Future {
				fmt.Println("cross queue", queueName, workflowID)
				if queueName == "side" {
					// since the orchestrator already created the workflow entity, we just need to execute it when we want
					future, err := orchestratorSide.ExecuteWithEntity(workflowID)
					if err != nil {
						f := tempolite.NewRuntimeFuture()
						f.SetError(err)
						return f
					}
					return future
				} else {
					f := tempolite.NewRuntimeFuture()
					f.SetError(fmt.Errorf("queue %s not found", queueName))
					return f
				}
			}),
	)

	act1 := func(ctx tempolite.ActivityContext) error {
		fmt.Println("act1")
		return nil
	}

	act2 := func(ctx tempolite.ActivityContext) error {
		fmt.Println("act2")
		return nil
	}

	act3 := func(ctx tempolite.ActivityContext) error {
		fmt.Println("act3")
		return nil
	}

	workflowSide := func(ctx tempolite.WorkflowContext) error {
		fmt.Println("workflowSide")
		if err := ctx.Activity("act2", act2, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	workflowDefault := func(ctx tempolite.WorkflowContext) error {

		if err := ctx.Activity("act1", act1, nil).Get(); err != nil {
			return err
		}
		if err := ctx.Workflow("side", workflowSide, &tempolite.WorkflowOptions{
			Queue: "side",
		}).Get(); err != nil {
			return err
		}

		if err := ctx.Activity("act3", act3, nil).Get(); err != nil {
			return err
		}

		return nil
	}

	future := orchestratorDefault.Execute(workflowDefault, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

}
