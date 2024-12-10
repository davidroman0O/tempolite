package tests

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

func TestWorkflowExecutePauseResumeSubWorkflow(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	subWorkflowFunc := func(ctx tempolite.WorkflowContext) error {
		var life int
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	orchestrator.Pause()
	time.Sleep(time.Second / 2)
	future = orchestrator.Resume(future.WorkflowID())
	time.Sleep(time.Second / 2)

	// Verify workflow completes successfully
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	// Verify flag was triggered
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	// Verify workflow states
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

}

func TestWorkflowExecutePauseResumeSubWorkflowFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	subWorkflowFunc := func(ctx tempolite.WorkflowContext) error {
		var life int
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		return fmt.Errorf("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	orchestrator.Pause()
	time.Sleep(time.Second / 2)
	future = orchestrator.Resume(future.WorkflowID())
	time.Sleep(time.Second / 2)

	// Verify workflow fails with expected error
	err := future.Get()
	if err == nil {
		t.Fatal("expected workflow to fail")
	}
	if !errors.Is(err, tempolite.ErrWorkflowFailed) {
		t.Fatalf("expected error to be ErrWorkflowFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "on purpose") {
		t.Fatalf("expected error to contain 'on purpose', got %v", err)
	}

	// Verify workflow failed state
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	// Verify Run status
	run, err := database.GetRun(1)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != tempolite.RunStatusFailed {
		t.Fatalf("expected run status %s, got %s", tempolite.RunStatusFailed, run.Status)
	}

	// Verify workflow executions
	rootExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(rootExecs) != 2 {
		t.Fatalf("expected 2 workflow executions, got %d", len(rootExecs))
	}

	sort.Slice(rootExecs, func(i, j int) bool {
		return rootExecs[i].ID < rootExecs[j].ID
	})

	// First execution should be paused
	if rootExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, rootExecs[0].Status)
	}

	// Second execution should be failed with error
	if rootExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, rootExecs[1].Status)
	}
	if !strings.Contains(rootExecs[1].Error, "fail workflow instance") {
		t.Fatalf("expected execution error 'fail workflow instance', got %s", rootExecs[1].Error)
	}

	// Verify workflow entity state
	entity, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if entity.Status != tempolite.StatusFailed {
		t.Fatalf("expected workflow status %s, got %s",
			tempolite.StatusFailed, entity.Status)
	}
	if entity.RetryState.Attempts != 1 {
		t.Fatalf("expected retry attempts 1, got %d", entity.RetryState.Attempts)
	}

	subWorkflows, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(subWorkflows) != 1 {
		t.Fatalf("expected 1 sub workflow, got %d", len(subWorkflows))
	}

	fmt.Println(subWorkflows[0].ID)

	// Verify side effect completed successfully despite workflow failure
	sideEffects, err := database.GetSideEffectEntities(subWorkflows[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	sort.Slice(sideEffects, func(i, j int) bool {
		return sideEffects[i].ID < sideEffects[j].ID
	})

	sideEffect := sideEffects[0]
	if sideEffect.Status != tempolite.StatusCompleted {
		t.Fatalf("expected side effect status %s, got %s",
			tempolite.StatusCompleted, sideEffect.Status)
	}

	// Verify side effect execution
	sideEffectExecs, err := database.GetSideEffectExecutions(sideEffect.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectExecs) != 1 {
		t.Fatalf("expected 1 side effect execution, got %d", len(sideEffectExecs))
	}

	sort.Slice(sideEffectExecs, func(i, j int) bool {
		return sideEffectExecs[i].ID < sideEffectExecs[j].ID
	})

	sideEffectExec := sideEffectExecs[0]
	if sideEffectExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected side effect execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, sideEffectExec.Status)
	}

	// Verify side effect data
	sideEffectData, err := database.GetSideEffectExecutionDataByExecutionID(sideEffectExec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sideEffectData.Outputs) != 1 {
		t.Fatalf("expected 1 side effect output, got %d", len(sideEffectData.Outputs))
	}

	// Verify hierarchies
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	// Sort hierarchies
	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	// First hierarchy: Root workflow -> Sub workflow
	h := hierarchies[0]
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}

	// Get the sub workflow's ID from the hierarchy
	subWorkflowID := tempolite.WorkflowEntityID(h.ChildEntityID)

	// Verify sub workflow -> side effect hierarchy
	subHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(subHierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for sub workflow, got %d", len(subHierarchies))
	}

	// Sort sub hierarchies
	sort.Slice(subHierarchies, func(i, j int) bool {
		return subHierarchies[i].ID < subHierarchies[j].ID
	})

	// Second hierarchy: Sub workflow -> Side effect
	subH := subHierarchies[0]
	if subH.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, subH.ParentEntityID)
	}
	if subH.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, subH.ChildEntityID)
	}

	// Verify execution IDs
	if subH.ParentExecutionID != 3 { // Sub workflow execution ID
		t.Fatalf("expected parent execution ID %d, got %d",
			3, subH.ParentExecutionID)
	}
	if subH.ChildExecutionID != int(sideEffectExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d",
			sideEffectExec.ID, subH.ChildExecutionID)
	}

	// Verify entity types
	if subH.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, subH.ParentType)
	}
	if subH.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, subH.ChildType)
	}

	// Verify step IDs
	if subH.ParentStepID != "sub" {
		t.Fatalf("expected parent step ID 'sub', got %s", subH.ParentStepID)
	}
	if subH.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", subH.ChildStepID)
	}
}

func TestWorkflowExecutePauseResumeSubWorkflowPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow_panic.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	subWorkflowFunc := func(ctx tempolite.WorkflowContext) error {
		var life int
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		panic("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	orchestrator.Pause()
	time.Sleep(time.Second / 2)
	future = orchestrator.Resume(future.WorkflowID())
	time.Sleep(time.Second / 2)

	// Verify workflow fails with expected error
	err := future.Get()
	if err == nil {
		t.Fatal("expected workflow to fail")
	}
	if !errors.Is(err, tempolite.ErrWorkflowFailed) {
		t.Fatalf("expected error to be ErrWorkflowFailed, got %v", err)
	}
	if !errors.Is(err, tempolite.ErrWorkflowPanicked) {
		t.Fatalf("expected error to contain ErrWorkflowPanicked, got %v", err)
	}

	// Verify workflow failed state
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	// Verify Run status
	run, err := database.GetRun(1)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != tempolite.RunStatusFailed {
		t.Fatalf("expected run status %s, got %s", tempolite.RunStatusFailed, run.Status)
	}

	// Get sub workflows first
	subWorkflows, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(subWorkflows) != 1 {
		t.Fatalf("expected 1 sub workflow, got %d", len(subWorkflows))
	}

	sort.Slice(subWorkflows, func(i, j int) bool {
		return subWorkflows[i].ID < subWorkflows[j].ID
	})

	subWorkflow := subWorkflows[0]

	// Verify workflow executions
	rootExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(rootExecs) != 2 {
		t.Fatalf("expected 2 workflow executions, got %d", len(rootExecs))
	}

	sort.Slice(rootExecs, func(i, j int) bool {
		return rootExecs[i].ID < rootExecs[j].ID
	})

	// First execution should be paused
	if rootExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, rootExecs[0].Status)
	}

	// Second execution should be failed with error message
	if rootExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, rootExecs[1].Status)
	}
	if !strings.Contains(rootExecs[1].Error, "workflow panicked") {
		t.Fatalf("expected execution error to contain 'workflow panicked', got %s", rootExecs[1].Error)
	}

	// Get sub workflow executions to check stack trace
	subExecs, err := database.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	sort.Slice(subExecs, func(i, j int) bool {
		return subExecs[i].ID < subExecs[j].ID
	})

	if len(subExecs) != 1 {
		t.Fatalf("expected 1 sub workflow execution, got %d", len(subExecs))
	}

	// Stack trace should be on the sub workflow execution
	if subExecs[0].StackTrace == nil {
		t.Fatal("expected stack trace to be present on sub workflow execution")
	}

	// Verify workflow entity state
	entity, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if entity.Status != tempolite.StatusFailed {
		t.Fatalf("expected workflow status %s, got %s",
			tempolite.StatusFailed, entity.Status)
	}
	if entity.RetryState.Attempts != 1 {
		t.Fatalf("expected retry attempts 1, got %d", entity.RetryState.Attempts)
	}

	// Verify side effect completed successfully despite workflow panic
	sideEffects, err := database.GetSideEffectEntities(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	sort.Slice(sideEffects, func(i, j int) bool {
		return sideEffects[i].ID < sideEffects[j].ID
	})

	sideEffect := sideEffects[0]
	if sideEffect.Status != tempolite.StatusCompleted {
		t.Fatalf("expected side effect status %s, got %s",
			tempolite.StatusCompleted, sideEffect.Status)
	}

	// Verify side effect execution
	sideEffectExecs, err := database.GetSideEffectExecutions(sideEffect.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectExecs) != 1 {
		t.Fatalf("expected 1 side effect execution, got %d", len(sideEffectExecs))
	}

	sort.Slice(sideEffectExecs, func(i, j int) bool {
		return sideEffectExecs[i].ID < sideEffectExecs[j].ID
	})

	sideEffectExec := sideEffectExecs[0]
	if sideEffectExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected side effect execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, sideEffectExec.Status)
	}

	// Verify hierarchy
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	// First hierarchy: Root workflow -> Sub workflow
	h := hierarchies[0]

	// Get the sub workflow's ID from the hierarchy
	subWorkflowID := tempolite.WorkflowEntityID(h.ChildEntityID)

	// Verify sub workflow -> side effect hierarchy
	subHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(subHierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for sub workflow, got %d", len(subHierarchies))
	}

	// Sort sub hierarchies
	sort.Slice(subHierarchies, func(i, j int) bool {
		return subHierarchies[i].ID < subHierarchies[j].ID
	})

	// Second hierarchy: Sub workflow -> Side effect
	subH := subHierarchies[0]

	if subH.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, subH.ParentEntityID)
	}
	if subH.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, subH.ChildEntityID)
	}
	if subH.ParentExecutionID != 3 {
		t.Fatalf("expected parent execution ID %d, got %d",
			3, subH.ParentExecutionID)
	}
	if subH.ChildExecutionID != int(sideEffectExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d",
			sideEffectExec.ID, subH.ChildExecutionID)
	}
}
