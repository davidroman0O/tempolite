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

/// TestWorkflowExecutePauseResume tests the execution of a workflow that is paused and resumed.
/// TestWorkflowExecutePauseResumeFailure tests the execution of a workflow that is paused and resumed, but fails.
/// TestWorkflowExecutePauseResumePanic tests the execution of a workflow that is paused and resumed, but panics.

func TestWorkflowExecutePauseResume(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		var life int
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
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

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Verify flag was triggered
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	// Verify workflow states
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

	// Verify workflow executions
	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and completed
	if len(execs) != 2 {
		t.Fatalf("expected 2 workflow executions, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	// First execution should be paused
	if execs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, execs[0].Status)
	}

	// Second execution should be completed
	if execs[1].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, execs[1].Status)
	}

	// Verify side effect
	sideEffects, err := database.GetSideEffectEntities(future.WorkflowID())
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

	if sideEffect.StepID != "something" {
		t.Fatalf("expected side effect step ID 'something', got %s",
			sideEffect.StepID)
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

	// Verify hierarchy
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	h := hierarchies[0]

	// Verify parent-child relationship
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}

	if h.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, h.ChildEntityID)
	}

	// Verify step IDs
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}

	if h.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", h.ChildStepID)
	}

	// Verify entity types
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}

	if h.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, h.ChildType)
	}

	// Verify execution IDs in hierarchy
	if h.ParentExecutionID != int(execs[1].ID) {
		t.Fatalf("expected parent execution ID %d, got %d",
			execs[1].ID, h.ParentExecutionID)
	}

	if h.ChildExecutionID != int(sideEffectExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d",
			sideEffectExec.ID, h.ChildExecutionID)
	}
}

func TestWorkflowExecutePauseResumeFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		var life int
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		return fmt.Errorf("on purpose")
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

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
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
	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(execs) != 2 {
		t.Fatalf("expected 2 workflow executions, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	// First execution should be paused
	if execs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, execs[0].Status)
	}

	// Second execution should be failed with error
	if execs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, execs[1].Status)
	}
	if execs[1].Error != "on purpose" {
		t.Fatalf("expected execution error 'on purpose', got %s", execs[1].Error)
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

	// Verify side effect completed successfully despite workflow failure
	sideEffects, err := database.GetSideEffectEntities(future.WorkflowID())
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

	// Verify hierarchy
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	h := hierarchies[0]

	// Verify parent-child relationship
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, h.ChildEntityID)
	}

	// Verify execution IDs in hierarchy
	if h.ParentExecutionID != int(execs[1].ID) {
		t.Fatalf("expected parent execution ID %d, got %d",
			execs[1].ID, h.ParentExecutionID)
	}
	if h.ChildExecutionID != int(sideEffectExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d",
			sideEffectExec.ID, h.ChildExecutionID)
	}

	// Verify entity types
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}
	if h.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, h.ChildType)
	}

	// Verify step IDs
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}
	if h.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", h.ChildStepID)
	}
}

func TestWorkflowExecutePauseResumePanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_panic.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		var life int
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		panic("on purpose")
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

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
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
	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(execs) != 2 {
		t.Fatalf("expected 2 workflow executions, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	// First execution should be paused
	if execs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, execs[0].Status)
	}

	// Second execution should be failed with panic error and stack trace
	if execs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, execs[1].Status)
	}
	if !strings.Contains(execs[1].Error, "workflow panicked") {
		t.Fatalf("expected execution error to contain 'workflow panicked', got %s", execs[1].Error)
	}
	if execs[1].StackTrace == nil {
		t.Fatal("expected stack trace to be present")
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
	sideEffects, err := database.GetSideEffectEntities(future.WorkflowID())
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
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	h := hierarchies[0]

	// Verify relationship IDs
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, h.ChildEntityID)
	}
	if h.ParentExecutionID != int(execs[1].ID) {
		t.Fatalf("expected parent execution ID %d, got %d",
			execs[1].ID, h.ParentExecutionID)
	}
	if h.ChildExecutionID != int(sideEffectExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d",
			sideEffectExec.ID, h.ChildExecutionID)
	}
}
