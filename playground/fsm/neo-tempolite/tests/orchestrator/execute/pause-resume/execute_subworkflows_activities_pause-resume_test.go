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

func TestWorkflowExecutePauseResumeSubWorkflowWithActivity(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow_activity.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	var calculateActivity = func(ctx tempolite.ActivityContext) (int, error) {
		return 42, nil
	}

	var subWorkflowFunc = func(ctx tempolite.WorkflowContext) error {
		var activityResult int
		<-time.After(time.Second / 4)
		if err := ctx.Activity("calculate", calculateActivity, nil).Get(&activityResult); err != nil {
			return err
		}
		return nil
	}

	var workflowFunc = func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	<-time.After(time.Second / 6)
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

	// Sort by ID for consistent testing
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

	// Get hierarchies to find sub-workflows
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

	// Should have exactly 2 executions: paused and completed
	if len(subWorkflowExecs) != 2 {
		t.Fatalf("expected 2 sub workflow executions, got %d", len(subWorkflowExecs))
	}

	// Sort executions by ID
	sort.Slice(subWorkflowExecs, func(i, j int) bool {
		return subWorkflowExecs[i].ID < subWorkflowExecs[j].ID
	})

	// First execution should be paused
	if subWorkflowExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusPaused, subWorkflowExecs[0].Status)
	}

	// Second execution should be completed
	if subWorkflowExecs[1].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected second sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, subWorkflowExecs[1].Status)
	}

	// Get sub workflow's activities
	activities, err := database.GetActivityEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	// Sort activities by ID
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]

	// Verify activity status
	if activity.Status != tempolite.StatusCompleted {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusCompleted, activity.Status)
	}

	// Get activity executions
	activityExecs, err := database.GetActivityExecutions(activity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	// Sort activity executions by ID
	sort.Slice(activityExecs, func(i, j int) bool {
		return activityExecs[i].ID < activityExecs[j].ID
	})

	activityExec := activityExecs[0]

	// Verify activity execution status
	if activityExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, activityExec.Status)
	}

	// Verify activity hierarchy
	activityHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(activityHierarchies) != 1 {
		t.Fatalf("expected 1 activity hierarchy, got %d", len(activityHierarchies))
	}

	// Sort activity hierarchies by ID
	sort.Slice(activityHierarchies, func(i, j int) bool {
		return activityHierarchies[i].ID < activityHierarchies[j].ID
	})

	ah := activityHierarchies[0]
	if ah.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, ah.ParentEntityID)
	}
	if ah.ChildEntityID != int(activity.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			activity.ID, ah.ChildEntityID)
	}
	if ah.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, ah.ParentType)
	}
	if ah.ChildType != tempolite.EntityActivity {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityActivity, ah.ChildType)
	}
}

func TestWorkflowExecutePauseResumeSubWorkflowWithActivityFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow_activity_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	var calculateActivity = func(ctx tempolite.ActivityContext) (int, error) {
		return 0, fmt.Errorf("activity failed on purpose")
	}

	var subWorkflowFunc = func(ctx tempolite.WorkflowContext) error {
		var activityResult int
		<-time.After(time.Second / 4)
		if err := ctx.Activity("calculate", calculateActivity, nil).Get(&activityResult); err != nil {
			return err
		}
		return nil
	}

	var workflowFunc = func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	<-time.After(time.Second / 6)
	orchestrator.Pause()
	time.Sleep(time.Second / 2)
	future = orchestrator.Resume(future.WorkflowID())
	time.Sleep(time.Second / 2)

	// Verify workflow fails with expected error
	err := future.Get()
	if err == nil {
		t.Fatal("expected workflow to fail")
	}
	if !strings.Contains(err.Error(), "activity failed on purpose") {
		t.Fatalf("expected error to contain 'activity failed on purpose', got %v", err)
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

	// Sort by ID for consistent testing
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
	if !strings.Contains(rootExecs[1].Error, "activity failed on purpose") {
		t.Fatalf("expected root execution error to contain 'activity failed on purpose', got %s",
			rootExecs[1].Error)
	}

	// Get root workflow entity
	rootWorkflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if rootWorkflow.Status != tempolite.StatusFailed {
		t.Fatalf("expected root workflow status %s, got %s",
			tempolite.StatusFailed, rootWorkflow.Status)
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

	// Should have exactly 2 executions: paused and failed
	if len(subWorkflowExecs) != 2 {
		t.Fatalf("expected 2 sub workflow executions, got %d", len(subWorkflowExecs))
	}

	// Sort executions by ID
	sort.Slice(subWorkflowExecs, func(i, j int) bool {
		return subWorkflowExecs[i].ID < subWorkflowExecs[j].ID
	})

	// First execution should be paused
	if subWorkflowExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusPaused, subWorkflowExecs[0].Status)
	}

	// Second execution should be failed with error
	if subWorkflowExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second sub workflow execution status %s, got %s",
			tempolite.ExecutionStatusFailed, subWorkflowExecs[1].Status)
	}
	if !strings.Contains(subWorkflowExecs[1].Error, "activity failed on purpose") {
		t.Fatalf("expected sub workflow execution error to contain 'activity failed on purpose', got %s",
			subWorkflowExecs[1].Error)
	}

	// Get sub workflow's activities
	activities, err := database.GetActivityEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	// Sort activities by ID
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]

	// Verify activity status
	if activity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activity.Status)
	}

	// Get activity executions
	activityExecs, err := database.GetActivityExecutions(activity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	// Sort activity executions by ID
	sort.Slice(activityExecs, func(i, j int) bool {
		return activityExecs[i].ID < activityExecs[j].ID
	})

	activityExec := activityExecs[0]

	// Verify activity execution status and error
	if activityExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusFailed, activityExec.Status)
	}
	if !strings.Contains(activityExec.Error, "activity failed on purpose") {
		t.Fatalf("expected activity execution error to contain 'activity failed on purpose', got %s",
			activityExec.Error)
	}

	// Verify activity hierarchy
	activityHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(activityHierarchies) != 1 {
		t.Fatalf("expected 1 activity hierarchy, got %d", len(activityHierarchies))
	}

	// Sort activity hierarchies by ID
	sort.Slice(activityHierarchies, func(i, j int) bool {
		return activityHierarchies[i].ID < activityHierarchies[j].ID
	})

	ah := activityHierarchies[0]
	if ah.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, ah.ParentEntityID)
	}
	if ah.ChildEntityID != int(activity.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			activity.ID, ah.ChildEntityID)
	}
	if ah.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, ah.ParentType)
	}
	if ah.ChildType != tempolite.EntityActivity {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityActivity, ah.ChildType)
	}
}

func TestWorkflowExecutePauseResumeSubWorkflowWithActivityPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_pause_resume_subworkflow_activity_panic.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	var calculateActivity = func(ctx tempolite.ActivityContext) (int, error) {
		panic("activity panicked on purpose")
	}

	var subWorkflowFunc = func(ctx tempolite.WorkflowContext) error {
		var activityResult int
		<-time.After(time.Second / 4)
		if err := ctx.Activity("calculate", calculateActivity, nil).Get(&activityResult); err != nil {
			return err
		}
		return nil
	}

	var workflowFunc = func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sub", subWorkflowFunc, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	// Execute and test pause/resume
	future := orchestrator.Execute(workflowFunc, nil)
	<-time.After(time.Second / 6)
	orchestrator.Pause()
	time.Sleep(time.Second / 2)
	future = orchestrator.Resume(future.WorkflowID())
	time.Sleep(time.Second / 2)

	// Verify workflow fails with expected error
	err := future.Get()
	if err == nil {
		t.Fatal("expected workflow to fail")
	}
	if !errors.Is(err, tempolite.ErrActivityPanicked) {
		t.Fatalf("expected error to be ErrActivityPanicked, got %v", err)
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

	// Second execution should be failed with error message
	if rootExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, rootExecs[1].Status)
	}

	// Get hierarchies to find sub-workflows
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil {
		t.Fatal(err)
	}

	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy for root workflow, got %d", len(hierarchies))
	}

	// Sort hierarchies by ID
	sort.Slice(hierarchies, func(i, j int) bool {
		return hierarchies[i].ID < hierarchies[j].ID
	})

	h := hierarchies[0]
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, h.ParentType)
	}
	if h.ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityWorkflow, h.ChildType)
	}

	// Get sub workflow from hierarchy
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

	if len(subWorkflowExecs) != 2 {
		t.Fatalf("expected 2 sub workflow executions, got %d", len(subWorkflowExecs))
	}

	// Verify paused and failed executions
	if subWorkflowExecs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution status %s, got %s",
			tempolite.ExecutionStatusPaused, subWorkflowExecs[0].Status)
	}
	if subWorkflowExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second execution status %s, got %s",
			tempolite.ExecutionStatusFailed, subWorkflowExecs[1].Status)
	}

	// Get activities from sub workflow
	activities, err := database.GetActivityEntities(subWorkflowID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	// Sort activities by ID
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]
	if activity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activity.Status)
	}

	// Get activity executions
	activityExecs, err := database.GetActivityExecutions(activity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	// Sort activity executions by ID
	sort.Slice(activityExecs, func(i, j int) bool {
		return activityExecs[i].ID < activityExecs[j].ID
	})

	activityExec := activityExecs[0]

	// Verify activity execution status and error details
	if activityExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusFailed, activityExec.Status)
	}

	if activityExec.StackTrace == nil {
		t.Fatal("expected activity execution to have stack trace")
	}

	if !strings.Contains(activityExec.Error, "activity panicked on purpose") {
		t.Fatal("expected activity execution error to contain 'activity panicked on purpose'")
	}

	// Get activity hierarchies
	activityHierarchies, err := database.GetHierarchiesByParentEntity(int(subWorkflowID))
	if err != nil {
		t.Fatal(err)
	}

	if len(activityHierarchies) != 1 {
		t.Fatalf("expected 1 activity hierarchy, got %d", len(activityHierarchies))
	}

	// Sort activity hierarchies by ID
	sort.Slice(activityHierarchies, func(i, j int) bool {
		return activityHierarchies[i].ID < activityHierarchies[j].ID
	})

	ah := activityHierarchies[0]
	if ah.ParentEntityID != int(subWorkflowID) {
		t.Fatalf("expected parent entity ID %d, got %d",
			subWorkflowID, ah.ParentEntityID)
	}
	if ah.ChildEntityID != int(activity.ID) {
		t.Fatalf("expected child entity ID %d, got %d",
			activity.ID, ah.ChildEntityID)
	}
	if ah.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s",
			tempolite.EntityWorkflow, ah.ParentType)
	}
	if ah.ChildType != tempolite.EntityActivity {
		t.Fatalf("expected child type %s, got %s",
			tempolite.EntityActivity, ah.ChildType)
	}
}
