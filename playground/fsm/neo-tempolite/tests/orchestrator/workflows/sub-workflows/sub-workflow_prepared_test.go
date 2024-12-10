package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

/// TestSubWorkflowPrepared 		- Sub-Workflow prepared execution
/// TestSubWorkflowPreparedFailure 	- Sub-Workflow prepared execution with error
/// TestSubWorkflowPreparedPanic 	- Sub-Workflow prepared execution

func TestSubWorkflowPrepared(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered bool
	var flagSubTriggered bool

	subworkflow := func(ctx tempolite.WorkflowContext) error {
		flagSubTriggered = true
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered = true
		if err := ctx.Workflow("sub", subworkflow, nil).Get(); err != nil {
			return err
		}
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

	database.SaveAsJSON("./jsons/workflows_subworkflow_prepared.json")

	if !flagTriggered {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered {
		t.Fatal("subworkflow was not triggered")
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
		t.Fatalf("expected 1 executions, got %d", len(execs))
	}

	if execs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusCompleted, execs[0].Status)
	}

	if execs[0].Error != "" {
		t.Fatal("expected no error")
	}
}

func TestSubWorkflowPreparedFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()

	orchestrator := tempolite.NewOrchestrator(context.Background(), database, registry)

	var flagTriggered bool
	var flagSubTriggered bool

	subworkflow := func(ctx tempolite.WorkflowContext) error {
		flagSubTriggered = true
		return fmt.Errorf("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered = true
		if err := ctx.Workflow("sub", subworkflow, nil).Get(); err != nil {
			return err
		}
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

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_subworkflow_prepared_failure.json")

	if !flagTriggered {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered {
		t.Fatal("subworkflow was not triggered")
	}

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

	if !strings.Contains(execs[0].Error, "fail workflow instance") {
		t.Fatal("expected error contains 'fail workflow instance' got", execs[0].Error)
	}

	ids, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	subwe, err := database.GetWorkflowEntity(ids[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if subwe.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, subwe.Status)
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

	if subexecs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, subexecs[0].Status)
	}

	if subexecs[0].Error != "on purpose" {
		t.Fatalf("expected error %s, got %s", "on purpose", subexecs[0].Error)
	}

}

func TestSubWorkflowPreparedPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered bool
	var flagSubTriggered bool

	subworkflow := func(ctx tempolite.WorkflowContext) error {
		flagSubTriggered = true
		panic("on purpose")
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered = true
		if err := ctx.Workflow("sub", subworkflow, nil).Get(); err != nil {
			return err
		}
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

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	database.SaveAsJSON("./jsons/workflows_subworkflow_prepared_panic.json")

	if !flagTriggered {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered {
		t.Fatal("subworkflow was not triggered")
	}

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

	ids, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 1 {
		t.Fatalf("expected 1 subworkflow, got %d", len(ids))
	}

	subwe, err := database.GetWorkflowEntity(ids[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if subwe.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, subwe.Status)
	}

	subexecs, err := database.GetWorkflowExecutions(subwe.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subexecs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(subexecs))
	}

	if subexecs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, subexecs[0].Status)
	}

	if subexecs[0].Error == "" {
		t.Fatal("expected error")
	}

	if subexecs[0].StackTrace == nil {
		t.Fatal("expected stack trace")
	}
}
