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
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		<-time.After(time.Second / 4)
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

	// Verify Run status
	run, err := database.GetRun(1)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != tempolite.RunStatusCompleted {
		t.Fatalf("expected run status %s, got %s",
			tempolite.RunStatusCompleted, run.Status)
	}

	// Get root workflow executions
	rootExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and completed
	if len(rootExecs) != 2 {
		t.Fatalf("expected 2 root workflow executions, got %d", len(rootExecs))
	}

	// Sort executions by ID
	sort.Slice(rootExecs, func(i, j int) bool {
		return rootExecs[i].ID < rootExecs[j].ID
	})

	// First execution should be paused
	if rootExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, rootExecs[0].Status)
	}

	// Second execution should be completed
	if rootExecs[1].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, rootExecs[1].Status)
	}

	// Get root workflow entity
	rootWorkflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if rootWorkflow.Status != tempolite.StatusCompleted {
		t.Fatalf("expected root workflow status %s, got %s",
			tempolite.StatusCompleted, rootWorkflow.Status)
	}

	// Get hierarchies to find sub-workflow
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	// Sort hierarchies for consistent testing
	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	// Verify hierarchy relationship
	h := hierarchies[0]
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}
	if h.ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityWorkflow, h.ChildType)
	}
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}
	if h.ChildStepID != "sub" {
		t.Fatalf("expected child step ID 'sub', got %s", h.ChildStepID)
	}

	// Get the sub workflow from hierarchy
	subWorkflowID := tempolite.WorkflowEntityID(h.ChildEntityID)
	subWorkflow, err := database.GetWorkflowEntity(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify sub workflow status
	if subWorkflow.Status != tempolite.StatusCompleted {
		t.Fatalf("expected sub workflow status %s, got %s",
			tempolite.StatusCompleted, subWorkflow.Status)
	}

	// Get sub workflow executions
	subWorkflowExecs, err := database.GetWorkflowExecutions(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subWorkflowExecs) != 1 {
		t.Fatalf("expected 1 sub workflow execution, got %d", len(subWorkflowExecs))
	}

	// Sort executions by ID
	sort.Slice(subWorkflowExecs, func(i, j int) bool {
		return subWorkflowExecs[i].ID < subWorkflowExecs[j].ID
	})

	// Verify sub workflow execution status
	if subWorkflowExecs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, subWorkflowExecs[0].Status)
	}

	// Get side effects from sub workflow
	sideEffects, err := database.GetSideEffectEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	// Sort side effects by ID
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

	// Sort side effect executions by ID
	sort.Slice(sideEffectExecs, func(i, j int) bool {
		return sideEffectExecs[i].ID < sideEffectExecs[j].ID
	})

	sideEffectExec := sideEffectExecs[0]
	if sideEffectExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected side effect execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, sideEffectExec.Status)
	}

	// Verify side effect hierarchy
	sideEffectHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectHierarchies) != 1 {
		t.Fatalf("expected 1 side effect hierarchy, got %d", len(sideEffectHierarchies))
	}

	// Sort side effect hierarchies by ID
	sort.Slice(sideEffectHierarchies, func(i, j int) bool {
		return sideEffectHierarchies[i].ID < sideEffectHierarchies[j].ID
	})

	sh := sideEffectHierarchies[0]
	if sh.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, sh.ParentEntityID)
	}
	if sh.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, sh.ChildEntityID)
	}
	if sh.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, sh.ParentType)
	}
	if sh.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, sh.ChildType)
	}
	if sh.ParentStepID != "sub" {
		t.Fatalf("expected parent step ID 'sub', got %s", sh.ParentStepID)
	}
	if sh.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", sh.ChildStepID)
	}

	// Verify side effect data
	sideEffectData, err := database.GetSideEffectExecutionDataByExecutionID(sideEffectExec.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectData.Outputs) != 1 {
		t.Fatalf("expected 1 side effect output, got %d", len(sideEffectData.Outputs))
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
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		return fmt.Errorf("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		<-time.After(time.Second / 4)
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

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Verify Run status
	run, err := database.GetRun(1)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != tempolite.RunStatusFailed {
		t.Fatalf("expected run status %s, got %s",
			tempolite.RunStatusFailed, run.Status)
	}

	// Get root workflow executions
	rootExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(rootExecs) != 2 {
		t.Fatalf("expected 2 root workflow executions, got %d", len(rootExecs))
	}

	// Sort executions by ID
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
	if !strings.Contains(rootExecs[1].Error, "on purpose") {
		t.Fatalf("expected execution error to contain 'on purpose', got %s",
			rootExecs[1].Error)
	}

	// Get root workflow entity
	rootWorkflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Verify root workflow status and retry state
	if rootWorkflow.Status != tempolite.StatusFailed {
		t.Fatalf("expected root workflow status %s, got %s",
			tempolite.StatusFailed, rootWorkflow.Status)
	}
	if rootWorkflow.RetryState.Attempts != 1 {
		t.Fatalf("expected retry attempts 1, got %d", rootWorkflow.RetryState.Attempts)
	}

	// Get hierarchies to find sub-workflow
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	// Sort hierarchies for consistent testing
	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	// Verify hierarchy relationship
	h := hierarchies[0]
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}
	if h.ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityWorkflow, h.ChildType)
	}
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}
	if h.ChildStepID != "sub" {
		t.Fatalf("expected child step ID 'sub', got %s", h.ChildStepID)
	}

	// Get the sub workflow from hierarchy
	subWorkflowID := tempolite.WorkflowEntityID(h.ChildEntityID)
	subWorkflow, err := database.GetWorkflowEntity(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify sub workflow status
	if subWorkflow.Status != tempolite.StatusFailed {
		t.Fatalf("expected sub workflow status %s, got %s",
			tempolite.StatusFailed, subWorkflow.Status)
	}

	// Get sub workflow executions
	subWorkflowExecs, err := database.GetWorkflowExecutions(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	// Sort executions by ID
	sort.Slice(subWorkflowExecs, func(i, j int) bool {
		return subWorkflowExecs[i].ID < subWorkflowExecs[j].ID
	})

	if len(subWorkflowExecs) != 1 {
		t.Fatalf("expected 1 sub workflow execution, got %d", len(subWorkflowExecs))
	}

	// Verify sub workflow execution error
	if subWorkflowExecs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusFailed, subWorkflowExecs[0].Status)
	}
	if !strings.Contains(subWorkflowExecs[0].Error, "on purpose") {
		t.Fatalf("expected sub workflow execution error to contain 'on purpose', got %s",
			subWorkflowExecs[0].Error)
	}

	// Get side effects from sub workflow
	sideEffects, err := database.GetSideEffectEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	// Sort side effects by ID
	sort.Slice(sideEffects, func(i, j int) bool {
		return sideEffects[i].ID < sideEffects[j].ID
	})

	// Verify side effect completed despite workflow failure
	sideEffect := sideEffects[0]
	if sideEffect.Status != tempolite.StatusCompleted {
		t.Fatalf("expected side effect status %s, got %s",
			tempolite.StatusCompleted, sideEffect.Status)
	}

	// Get side effect executions
	sideEffectExecs, err := database.GetSideEffectExecutions(sideEffect.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectExecs) != 1 {
		t.Fatalf("expected 1 side effect execution, got %d", len(sideEffectExecs))
	}

	// Sort side effect executions by ID
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

	// Verify side effect hierarchy
	sideEffectHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectHierarchies) != 1 {
		t.Fatalf("expected 1 side effect hierarchy, got %d", len(sideEffectHierarchies))
	}

	// Sort side effect hierarchies by ID
	sort.Slice(sideEffectHierarchies, func(i, j int) bool {
		return sideEffectHierarchies[i].ID < sideEffectHierarchies[j].ID
	})

	sh := sideEffectHierarchies[0]
	if sh.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, sh.ParentEntityID)
	}
	if sh.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, sh.ChildEntityID)
	}
	if sh.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, sh.ParentType)
	}
	if sh.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, sh.ChildType)
	}
	if sh.ParentStepID != "sub" {
		t.Fatalf("expected parent step ID 'sub', got %s", sh.ParentStepID)
	}
	if sh.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", sh.ChildStepID)
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
		<-time.After(time.Second / 4)
		if err := ctx.SideEffect("something", func() int {
			return 42
		}).Get(&life); err != nil {
			return err
		}
		panic("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		<-time.After(time.Second / 4)
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
		t.Fatalf("expected error to be ErrWorkflowPanicked, got %v", err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Verify Run status
	run, err := database.GetRun(1)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != tempolite.RunStatusFailed {
		t.Fatalf("expected run status %s, got %s",
			tempolite.RunStatusFailed, run.Status)
	}

	// Get root workflow executions
	rootExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 2 executions: paused and failed
	if len(rootExecs) != 2 {
		t.Fatalf("expected 2 root workflow executions, got %d", len(rootExecs))
	}

	// Sort executions by ID
	sort.Slice(rootExecs, func(i, j int) bool {
		return rootExecs[i].ID < rootExecs[j].ID
	})

	// First execution should be paused
	if rootExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, rootExecs[0].Status)
	}

	// Second execution should be failed with panic error
	if rootExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, rootExecs[1].Status)
	}
	if !strings.Contains(rootExecs[1].Error, "workflow panicked") {
		t.Fatalf("expected root execution error to contain 'workflow panicked', got %s",
			rootExecs[1].Error)
	}

	// Get root workflow entity
	rootWorkflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Verify root workflow status and retry state
	if rootWorkflow.Status != tempolite.StatusFailed {
		t.Fatalf("expected root workflow status %s, got %s",
			tempolite.StatusFailed, rootWorkflow.Status)
	}
	if rootWorkflow.RetryState.Attempts != 1 {
		t.Fatalf("expected retry attempts 1, got %d", rootWorkflow.RetryState.Attempts)
	}

	// Get hierarchies to find sub-workflow
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	// Sort hierarchies for consistent testing
	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	// Verify hierarchy relationship
	h := hierarchies[0]
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d",
			future.WorkflowID(), h.ParentEntityID)
	}
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}
	if h.ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityWorkflow, h.ChildType)
	}
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}
	if h.ChildStepID != "sub" {
		t.Fatalf("expected child step ID 'sub', got %s", h.ChildStepID)
	}

	// Get the sub workflow from hierarchy
	subWorkflowID := tempolite.WorkflowEntityID(h.ChildEntityID)
	subWorkflow, err := database.GetWorkflowEntity(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify sub workflow status
	if subWorkflow.Status != tempolite.StatusFailed {
		t.Fatalf("expected sub workflow status %s, got %s",
			tempolite.StatusFailed, subWorkflow.Status)
	}

	// Get sub workflow executions
	subWorkflowExecs, err := database.GetWorkflowExecutions(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subWorkflowExecs) != 1 {
		t.Fatalf("expected 1 sub workflow execution, got %d", len(subWorkflowExecs))
	}

	// Sort executions by ID
	sort.Slice(subWorkflowExecs, func(i, j int) bool {
		return subWorkflowExecs[i].ID < subWorkflowExecs[j].ID
	})

	// Verify sub workflow execution panic details
	subExec := subWorkflowExecs[0]
	if subExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusFailed, subExec.Status)
	}
	if !strings.Contains(subExec.Error, "on purpose") {
		t.Fatalf("expected sub workflow error to contain 'on purpose', got %s",
			subExec.Error)
	}
	if subExec.StackTrace == nil {
		t.Fatal("expected stack trace to be present on sub workflow execution")
	}
	if !strings.Contains(*subExec.StackTrace, "panic") {
		t.Fatalf("expected stack trace to contain 'panic', got %s",
			*subExec.StackTrace)
	}

	// Get side effects from sub workflow
	sideEffects, err := database.GetSideEffectEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	// Sort side effects by ID
	sort.Slice(sideEffects, func(i, j int) bool {
		return sideEffects[i].ID < sideEffects[j].ID
	})

	// Verify side effect completed despite workflow panic
	sideEffect := sideEffects[0]
	if sideEffect.Status != tempolite.StatusCompleted {
		t.Fatalf("expected side effect status %s, got %s",
			tempolite.StatusCompleted, sideEffect.Status)
	}

	// Get side effect executions
	sideEffectExecs, err := database.GetSideEffectExecutions(sideEffect.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectExecs) != 1 {
		t.Fatalf("expected 1 side effect execution, got %d", len(sideEffectExecs))
	}

	// Sort side effect executions by ID
	sort.Slice(sideEffectExecs, func(i, j int) bool {
		return sideEffectExecs[i].ID < sideEffectExecs[j].ID
	})

	sideEffectExec := sideEffectExecs[0]
	if sideEffectExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected side effect execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, sideEffectExec.Status)
	}

	// Verify side effect hierarchy
	sideEffectHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(sideEffectHierarchies) != 1 {
		t.Fatalf("expected 1 side effect hierarchy, got %d", len(sideEffectHierarchies))
	}

	// Sort side effect hierarchies by ID
	sort.Slice(sideEffectHierarchies, func(i, j int) bool {
		return sideEffectHierarchies[i].ID < sideEffectHierarchies[j].ID
	})

	sh := sideEffectHierarchies[0]
	if sh.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, sh.ParentEntityID)
	}
	if sh.ChildEntityID != int(sideEffect.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			sideEffect.ID, sh.ChildEntityID)
	}
	if sh.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, sh.ParentType)
	}
	if sh.ChildType != tempolite.EntitySideEffect {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntitySideEffect, sh.ChildType)
	}
	if sh.ParentStepID != "sub" {
		t.Fatalf("expected parent step ID 'sub', got %s", sh.ParentStepID)
	}
	if sh.ChildStepID != "something" {
		t.Fatalf("expected child step ID 'something', got %s", sh.ChildStepID)
	}
}
