package tests

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

/// TestWorkflowExecute 			- Basic Workflow direct execution
/// TestWorkflowExecuteFailure 		- Basic Workflow direct execution with error
/// TestWorkflowExecutePanic 		- Basic Workflow direct execution with panic

func TestWorkflowExecute(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	database.SaveAsJSON("./jsons/workflows_basic_execute.json")

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	if execs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusCompleted, execs[0].Status)
	}

	if execs[0].Error != "" {
		t.Fatal("expected no error")
	}
}

func TestWorkflowExecuteFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		return fmt.Errorf("on purpose")
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_basic_execute_failure.json")

	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
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
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	if execs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
	}

	if execs[0].Error == "" {
		t.Fatal("expected error")
	}
}

func TestWorkflowExecutePanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		panic("on purpose")
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_basic_execute_panic.json")

	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
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
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	if execs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
	}

	if execs[0].Error == "" {
		t.Fatal("expected error")
	}
	if execs[0].StackTrace == nil {
		t.Fatal("expected error")
	}
}
