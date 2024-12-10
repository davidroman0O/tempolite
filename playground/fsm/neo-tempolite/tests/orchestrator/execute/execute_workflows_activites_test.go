package tests

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

/// TestWorkflowExecuteActivities 			- Basic Workflow direct execution
/// TestWorkflowExecuteActivitiesFailure 		- Basic Workflow direct execution with error
/// TestWorkflowExecuteActivitiesPanic 		- Basic Workflow direct execution with panic

func TestWorkflowExecuteActivities(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_activities_execute.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagActivityTriggered atomic.Bool
	var flagTriggered atomic.Bool

	activity := func(ctx tempolite.ActivityContext) error {
		flagActivityTriggered.Store(true)
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Activity("sub", activity, nil).Get()
	}

	future := orchestrator.Execute(workflowFunc, nil)
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagActivityTriggered.Load() {
		t.Fatal("activity was not triggered")
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

	// After the flagTriggered check, replace with:
	activities, err := database.GetActivityEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activityEntity := activities[0]
	if activityEntity.Status != tempolite.StatusCompleted {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusCompleted, activityEntity.Status)
	}

	activityExecs, err := database.GetActivityExecutions(activityEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	activityExec := activityExecs[0]
	if activityExec.Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, activityExec.Status)
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
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d", future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(activityEntity.ID) {
		t.Fatalf("expected child entity ID %d, got %d", activityEntity.ID, h.ChildEntityID)
	}
	if h.ParentExecutionID != int(execs[0].ID) {
		t.Fatalf("expected parent execution ID %d, got %d", execs[0].ID, h.ParentExecutionID)
	}
	if h.ChildExecutionID != int(activityExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d", activityExec.ID, h.ChildExecutionID)
	}
}

func TestWorkflowExecuteActivitiesFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_activities_execute_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagActivityTriggered atomic.Bool
	var flagTriggered atomic.Bool

	activity := func(ctx tempolite.ActivityContext) error {
		flagActivityTriggered.Store(true)
		return fmt.Errorf("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Activity("sub", activity, nil).Get()
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagActivityTriggered.Load() {
		t.Fatal("activity was not triggered")
	}

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

	activities, err := database.GetActivityEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activityEntity := activities[0]
	if activityEntity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activityEntity.Status)
	}

	activityExecs, err := database.GetActivityExecutions(activityEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	activityExec := activityExecs[0]
	if activityExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusFailed, activityExec.Status)
	}

	if !strings.Contains(activityExec.Error, "on purpose") {
		t.Fatalf("expected activity execution error to contain 'on purpose', got %s", activityExec.Error)
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

	if !strings.Contains(execs[0].Error, "on purpose") {
		t.Fatalf("expected error to contain 'on purpose'")
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
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d", future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(activityEntity.ID) {
		t.Fatalf("expected child entity ID %d, got %d", activityEntity.ID, h.ChildEntityID)
	}
	if h.ParentExecutionID != int(execs[0].ID) {
		t.Fatalf("expected parent execution ID %d, got %d", execs[0].ID, h.ParentExecutionID)
	}
	if h.ChildExecutionID != int(activityExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d", activityExec.ID, h.ChildExecutionID)
	}
}

func TestWorkflowExecuteActivitiesPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_activities_execute_panic.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagActivityTriggered atomic.Bool
	var flagTriggered atomic.Bool

	activity := func(ctx tempolite.ActivityContext) error {
		flagActivityTriggered.Store(true)
		panic("on purpose")
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Activity("sub", activity, nil).Get()
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagActivityTriggered.Load() {
		t.Fatal("activity was not triggered")
	}

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

	activities, err := database.GetActivityEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	activityEntity := activities[0]
	if activityEntity.Status != tempolite.StatusFailed {
		t.Fatalf("expected activity status %s, got %s",
			tempolite.StatusFailed, activityEntity.Status)
	}

	activityExecs, err := database.GetActivityExecutions(activityEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(activityExecs) != 1 {
		t.Fatalf("expected 1 activity execution, got %d", len(activityExecs))
	}

	activityExec := activityExecs[0]
	if activityExec.Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected activity execution status %s, got %s",
			tempolite.ExecutionStatusFailed, activityExec.Status)
	}

	if !strings.Contains(activityExec.Error, "activity panicked") {
		t.Fatalf("expected activity execution error to contain 'activity panicked', got %s",
			activityExec.Error)
	}

	if activityExec.StackTrace == nil {
		t.Fatal("expected stack trace to be present")
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

	if !strings.Contains(execs[0].Error, "activity panicked") {
		t.Fatalf("expected error to contain 'activity panicked'")
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
	if h.ParentEntityID != int(future.WorkflowID()) {
		t.Fatalf("expected parent entity ID %d, got %d", future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(activityEntity.ID) {
		t.Fatalf("expected child entity ID %d, got %d", activityEntity.ID, h.ChildEntityID)
	}
	if h.ParentExecutionID != int(execs[0].ID) {
		t.Fatalf("expected parent execution ID %d, got %d", execs[0].ID, h.ParentExecutionID)
	}
	if h.ChildExecutionID != int(activityExec.ID) {
		t.Fatalf("expected child execution ID %d, got %d", activityExec.ID, h.ChildExecutionID)
	}
}
