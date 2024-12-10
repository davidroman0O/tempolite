package tests

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

func TestWorkflowCrossQueuePauseResume(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_activities_cross_queue_pause-resume.json")
	ctx := context.Background()

	// Create side queue
	if _, err := database.AddQueue(&tempolite.Queue{Name: "side"}); err != nil {
		t.Fatal(err)
	}

	// Create side orchestrator
	orchestratorSide := tempolite.NewOrchestrator(
		ctx, database, registry,
		tempolite.WithQueueName("side"),
	)

	// Create default orchestrator with cross-queue handling
	orchestratorDefault := tempolite.NewOrchestrator(
		ctx, database, registry,
		tempolite.WithCrossWorkflow(func(queueName string, workflowID tempolite.WorkflowEntityID,
			workflowFunc interface{}, options *tempolite.WorkflowOptions, args ...interface{}) tempolite.Future {

			if queueName == "side" {
				future, err := orchestratorSide.ExecuteWithEntity(workflowID)
				if err != nil {
					f := tempolite.NewRuntimeFuture()
					f.SetError(err)
					return f
				}
				return future
			}
			f := tempolite.NewRuntimeFuture()
			f.SetError(fmt.Errorf("unsupported queue: %s", queueName))
			return f
		}),
	)

	// Define test workflows
	workflowSide := func(ctx tempolite.WorkflowContext) error {
		<-time.After(time.Second / 2)
		return nil
	}

	workflowDefault := func(ctx tempolite.WorkflowContext) error {
		<-time.After(time.Second / 2)
		return ctx.Workflow("side", workflowSide, &tempolite.WorkflowOptions{
			Queue: "side",
		}).Get()
	}

	// Execute and test pause/resume
	future := orchestratorDefault.Execute(workflowDefault, nil)
	orchestratorDefault.Pause()
	time.Sleep(time.Second) // Allow pause to propagate

	// Resume and verify completion
	future = orchestratorDefault.Resume(future.WorkflowID())
	time.Sleep(time.Second) // Allow completion

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	// Check root workflow completion
	if err := orchestratorDefault.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatalf("workflow %d: expected status %s, got error: %v", future.WorkflowID(), tempolite.StatusCompleted, err)
	}

	// Get and verify sub-workflows
	subs, err := database.GetWorkflowSubWorkflows(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 sub workflow, got %d", len(subs))
	}

	sort.Slice(subs, func(i, j int) bool {
		return subs[i].ID < subs[j].ID
	})

	// Check side workflow completion
	if err := orchestratorDefault.WaitFor(subs[0].ID, tempolite.StatusCompleted); err != nil {
		t.Fatalf("workflow %d: expected status %s, got error: %v", subs[0].ID, tempolite.StatusCompleted, err)
	}

	// Verify workflow data integrity
	data, err := database.GetWorkflowDataByEntityID(subs[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify no continued workflow references
	if data.ContinuedFrom != nil {
		t.Fatalf("expected nil ContinuedFrom, got %v", data.ContinuedFrom)
	}
	if data.ContinuedExecutionFrom != nil {
		t.Fatalf("expected nil ContinuedExecutionFrom, got %v", data.ContinuedExecutionFrom)
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
		t.Fatalf("expected parent entity ID %d, got %d", future.WorkflowID(), h.ParentEntityID)
	}
	if h.ChildEntityID != int(subs[0].ID) {
		t.Fatalf("expected child entity ID %d, got %d", subs[0].ID, h.ChildEntityID)
	}

	// Verify entity types
	if h.ChildType != tempolite.EntityWorkflow {
		t.Fatalf("expected child type %s, got %s", tempolite.EntityWorkflow, h.ChildType)
	}
	if h.ParentType != tempolite.EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s", tempolite.EntityWorkflow, h.ParentType)
	}

	// Verify step IDs
	if h.ChildStepID != "side" {
		t.Fatalf("expected child step ID 'side', got %s", h.ChildStepID)
	}
	if h.ParentStepID != "root" {
		t.Fatalf("expected parent step ID 'root', got %s", h.ParentStepID)
	}

	// Verify executions
	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// sort all the time
	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	if len(execs) != 2 { // Should have pause and completion executions
		t.Fatalf("expected 2 executions for root workflow, got %d", len(execs))
	}
	if execs[0].Status != tempolite.ExecutionStatusPaused {
		t.Fatalf("expected first execution to be paused, got %s", execs[0].Status)
	}
	if execs[1].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected second execution to be completed, got %s", execs[1].Status)
	}
}
