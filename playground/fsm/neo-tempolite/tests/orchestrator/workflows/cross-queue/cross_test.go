package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

/// TestWorkflowCrossQueue 			- Cross Queue Workflow execution, using the ExecuteWithEntity since the orchestrator pre-create the workflow

func TestWorkflowCrossQueue(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	defer database.SaveAsJSON("./jsons/workflows_cross_queue.json")

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

	workflowSide := func(ctx tempolite.WorkflowContext) error {
		fmt.Println("workflowSide")
		return nil
	}

	workflowDefault := func(ctx tempolite.WorkflowContext) error {
		if err := ctx.Workflow("side", workflowSide, &tempolite.WorkflowOptions{
			Queue: "side",
		}).Get(); err != nil {
			return err
		}
		return nil
	}

	future := orchestratorDefault.Execute(workflowDefault, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if err := orchestratorDefault.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

	subs, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(subs) != 1 {
		t.Fatalf("expected 1 sub workflow, got %d", len(subs))
	}

	if err := orchestratorSide.WaitFor(subs[0].ID, tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

	// we have to check that all the time because one day we had a bug of those values being filled
	data, err := database.GetWorkflowDataByEntityID(subs[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if data.ContinuedFrom != nil {
		t.Fatalf("expected nil, got %v", data.ContinuedFrom)
	}

	if data.ContinuedExecutionFrom != nil {
		t.Fatalf("expected nil, got %v", data.ContinuedExecutionFrom)
	}

	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	if hierarchies[0].ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity id %d, got %d", future.WorkflowID(), hierarchies[0].ParentEntityID)
	}

	if hierarchies[0].ChildEntityID != int(subs[0].ID) {
		t.Fatalf("expected child entity id %d, got %d", subs[0].ID, hierarchies[0].ChildEntityID)
	}

	if hierarchies[0].ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected type %s, got %s", tempolite.EntityWorkflow, hierarchies[0].ChildType)
	}

	if hierarchies[0].ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected type %s, got %s", tempolite.EntityWorkflow, hierarchies[0].ParentType)
	}

	if hierarchies[0].ChildStepID != "side" {
		t.Fatalf("expected step id %s, got %s", "side", hierarchies[0].ChildStepID)
	}

	if hierarchies[0].ParentStepID != "root" {
		t.Fatalf("expected step id %s, got %s", "root", hierarchies[0].ParentStepID)
	}

}

func TestWorkflowCrossQueueFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

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

	workflowSide := func(ctx tempolite.WorkflowContext) error {
		fmt.Println("workflowSide")
		return fmt.Errorf("on purpose")
	}

	workflowDefault := func(ctx tempolite.WorkflowContext) error {
		if err := ctx.Workflow("side", workflowSide, &tempolite.WorkflowOptions{
			Queue: "side",
		}).Get(); err != nil {
			return err
		}
		return nil
	}

	future := orchestratorDefault.Execute(workflowDefault, nil)

	if err := future.Get(); err != nil {
		if !errors.Is(err, tempolite.ErrWorkflowFailed) {
			t.Fatal(err)
		}
	}

	if err := orchestratorDefault.WaitFor(future.WorkflowID(), tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	subs, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(subs) != 1 {
		t.Fatalf("expected 1 sub workflow, got %d", len(subs))
	}

	if err := orchestratorSide.WaitFor(subs[0].ID, tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	// we have to check that all the time because one day we had a bug of those values being filled
	data, err := database.GetWorkflowDataByEntityID(subs[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if data.ContinuedFrom != nil {
		t.Fatalf("expected nil, got %v", data.ContinuedFrom)
	}

	if data.ContinuedExecutionFrom != nil {
		t.Fatalf("expected nil, got %v", data.ContinuedExecutionFrom)
	}

	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	if hierarchies[0].ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity id %d, got %d", future.WorkflowID(), hierarchies[0].ParentEntityID)
	}

	if hierarchies[0].ChildEntityID != int(subs[0].ID) {
		t.Fatalf("expected child entity id %d, got %d", subs[0].ID, hierarchies[0].ChildEntityID)
	}

	if hierarchies[0].ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected type %s, got %s", tempolite.EntityWorkflow, hierarchies[0].ChildType)
	}

	if hierarchies[0].ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected type %s, got %s", tempolite.EntityWorkflow, hierarchies[0].ParentType)
	}

	if hierarchies[0].ChildStepID != "root" {
		t.Fatalf("expected step id %s, got %s", "root", hierarchies[0].ChildStepID)
	}

	if hierarchies[0].ParentStepID != "root" {
		t.Fatalf("expected step id %s, got %s", "root", hierarchies[0].ParentStepID)
	}

	database.SaveAsJSON("./jsons/workflows_cross_queue_failure.json")
}
