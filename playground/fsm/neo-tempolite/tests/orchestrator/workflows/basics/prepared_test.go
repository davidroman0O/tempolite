package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

///	TestWorkflowPrepared	- Basic Workflow prepared execution
///	TestWorkflowPreparedFailure	- Basic Workflow prepared execution with error
///	TestWorkflowPreparedPanic	- Basic Workflow prepared execution

func TestWorkflowPrepared(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return nil
	}

	we, err := tempolite.PrepareWorkflow(registry, database, workflowFunc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	future, err := orchestrator.ExecuteWithEntity(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	database.SaveAsJSON("./jsons/workflows_basic_prepared.json")

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	we, err = database.GetWorkflowEntity(future.WorkflowID())
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

	if execs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusCompleted, execs[0].Status)
	}

	if execs[0].Error != "" {
		t.Fatal("expected no error")
	}
}

func TestWorkflowPreparedFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		return fmt.Errorf("on purpose")
	}

	we, err := tempolite.PrepareWorkflow(registry, database, workflowFunc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	future, err := orchestrator.ExecuteWithEntity(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_basic_prepared_failure.json")

	we, err = database.GetWorkflowEntity(future.WorkflowID())
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

	if execs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
	}

	if execs[0].Error == "" {
		t.Fatal("expected error")
	}
}

func TestWorkflowPreparedPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		panic("on purpose")
	}

	we, err := tempolite.PrepareWorkflow(registry, database, workflowFunc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	future, err := orchestrator.ExecuteWithEntity(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_basic_prepared_panic.json")

	we, err = database.GetWorkflowEntity(future.WorkflowID())
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

	if execs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
	}

	if execs[0].Error == "" {
		t.Fatal("expected error")
	}
}
