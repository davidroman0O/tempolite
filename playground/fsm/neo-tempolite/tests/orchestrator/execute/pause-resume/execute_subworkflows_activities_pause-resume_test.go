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

	// Verify flag was triggered
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	// Verify workflow states
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

	// Get sub workflows
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

	// Get activities from sub workflow
	activities, err := database.GetActivityEntities(subWorkflows[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]
	if activity.Status != tempolite.StatusCompleted {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusCompleted, activity.Status)
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

	// Verify workflow failed state
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	// Get sub workflows
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

	// Get activities from sub workflow
	activities, err := database.GetActivityEntities(subWorkflows[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]
	if activity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activity.Status)
	}

	// Verify activity executions
	activityExecs, err := database.GetActivityExecutions(activity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	sort.Slice(activityExecs, func(i, j int) bool {
		return activityExecs[i].ID < activityExecs[j].ID
	})

	activityExec := activityExecs[0]
	if activityExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusFailed, activityExec.Status)
	}
	if !strings.Contains(activityExec.Error, "activity failed on purpose") {
		t.Fatalf("expected activity execution error to contain 'activity failed on purpose', got %s",
			activityExec.Error)
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

	// Verify workflow failed state
	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusFailed); err != nil {
		t.Fatal(err)
	}

	// Get sub workflows
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

	// Get activities from sub workflow
	activities, err := database.GetActivityEntities(subWorkflows[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activity := activities[0]
	if activity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activity.Status)
	}

	// Verify activity executions
	activityExecs, err := database.GetActivityExecutions(activity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	sort.Slice(activityExecs, func(i, j int) bool {
		return activityExecs[i].ID < activityExecs[j].ID
	})

	activityExec := activityExecs[0]

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
	if activityExec.StackTrace == nil {
		t.Fatal("expected activity execution to have stack trace")
	}
}
