package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

/// TestSubWorkflowExecute - Sub-Workflow direct execution

func TestSubWorkflowExecute(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool
	var flagSubTriggered atomic.Bool

	subworkflow := func(ctx tempolite.WorkflowContext) error {
		flagSubTriggered.Store(true)
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subworkflow, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	database.SaveAsJSON("./jsons/workflows_subworkflow_execute.json")

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered.Load() {
		t.Fatal("subworkflow was not triggered")
	}

	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.StatusCompleted, we.Status)
	}

	if we.StepID != "root" {
		t.Fatalf("expected stepID %s, got %s", "root", we.StepID)
	}

	if we.RunID != 1 {
		t.Fatalf("expected runID %d, got %d", 1, we.RunID)
	}

	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 1 {
		t.Fatalf("expected 1 executions, got %d", len(execs))
	}

	// let's search for the sub-workflow
	// TODO: we should provide a function to get all sub-workflows in a workflow
	// TODO: we should provive a function to get all the components of a workflow
	hierarchies, err := database.GetHierarchiesByParentEntityAndStep(int(we.ID), "sub", tempolite.EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	subwe, err := database.GetWorkflowEntity(tempolite.WorkflowEntityID(hierarchies[0].ChildEntityID))
	if err != nil {
		t.Fatal(err)
	}

	if subwe.Status != tempolite.StatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.StatusCompleted, subwe.Status)
	}

	if subwe.StepID != "sub" {
		t.Fatalf("expected stepID %s, got %s", "sub", subwe.StepID)
	}

	if subwe.RunID != 1 {
		t.Fatalf("expected runID %d, got %d", 1, subwe.RunID)
	}

	subexecs, err := database.GetWorkflowExecutions(subwe.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subexecs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(subexecs))
	}

	if subexecs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusCompleted, subexecs[0].Status)
	}

	if subexecs[0].Error != "" {
		t.Fatal("expected no error")
	}
}

func TestSubWorkflowExecuteFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	subworkflow := func(ctx tempolite.WorkflowContext) error {
		return fmt.Errorf("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		if err := ctx.Workflow("sub", subworkflow, nil).Get(); err == nil {
			return err
		}
		return nil
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	database.SaveAsJSON("./jsons/workflows_subworkflow_execute_failure.json")

	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
	}

	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	if execs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
	}

	if execs[0].Error == "" {
		t.Fatal("expected error")
	}
}
