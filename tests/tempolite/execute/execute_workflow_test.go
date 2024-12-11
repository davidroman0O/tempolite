package tests

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite"
)

func TestTempoliteWorkflowsExecute(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Get workflow entity directly using ID from future
	workflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Get workflow execution directly using ID from future
	execution, err := database.GetWorkflowExecution(future.WorkflowExecutionID())
	if err != nil {
		t.Fatal(err)
	}

	// Get the run directly
	run, err := database.GetRun(workflow.RunID)
	if err != nil {
		t.Fatal(err)
	}

	// Get the queue directly
	queue, err := database.GetQueue(workflow.QueueID)
	if err != nil {
		t.Fatal(err)
	}

	// Validate Run
	if run.Status != tempolite.RunStatusCompleted {
		t.Errorf("expected run status %s, got %s", tempolite.RunStatusCompleted, run.Status)
	}

	// Validate Queue
	if queue.Name != tempolite.DefaultQueue {
		t.Errorf("expected default queue name %s, got %s", tempolite.DefaultQueue, queue.Name)
	}

	// Check workflow entity properties
	if workflow.Status != tempolite.StatusCompleted {
		t.Errorf("expected workflow status %s, got %s", tempolite.StatusCompleted, workflow.Status)
	}
	if workflow.Type != tempolite.EntityWorkflow {
		t.Errorf("expected entity type %s, got %s", tempolite.EntityWorkflow, workflow.Type)
	}
	if workflow.StepID != "root" {
		t.Errorf("expected step ID 'root', got %s", workflow.StepID)
	}
	if workflow.QueueID != queue.ID {
		t.Errorf("expected queue ID %d, got %d", queue.ID, workflow.QueueID)
	}
	if workflow.RunID != run.ID {
		t.Errorf("expected run ID %d, got %d", run.ID, workflow.RunID)
	}

	// Validate Workflow Data
	if workflow.WorkflowData == nil {
		t.Fatal("workflow data should not be nil")
	}
	if !workflow.WorkflowData.IsRoot {
		t.Error("workflow should be marked as root")
	}
	if workflow.WorkflowData.Paused {
		t.Error("workflow should not be paused")
	}
	if workflow.WorkflowData.Resumable {
		t.Error("workflow should not be resumable")
	}

	// Validate Workflow Execution
	if execution.Status != tempolite.ExecutionStatusCompleted {
		t.Errorf("expected execution status %s, got %s", tempolite.ExecutionStatusCompleted, execution.Status)
	}
	if execution.Error != "" {
		t.Errorf("expected no execution error, got: %s", execution.Error)
	}
	if execution.WorkflowEntityID != workflow.ID {
		t.Errorf("expected workflow entity ID %d, got %d", workflow.ID, execution.WorkflowEntityID)
	}
	if execution.CompletedAt == nil {
		t.Error("execution completed timestamp should be set")
	}

	// Verify no hierarchies exist for this workflow by checking directly
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil && !errors.Is(err, tempolite.ErrHierarchyNotFound) {
		t.Fatal(err)
	}
	if len(hierarchies) != 0 {
		t.Errorf("expected 0 hierarchies, got %d", len(hierarchies))
	}

	// Validate RetryPolicy defaults
	defaultPolicy := tempolite.DefaultRetryPolicyInternal()
	if workflow.RetryPolicy.MaxAttempts != defaultPolicy.MaxAttempts {
		t.Errorf("expected max attempts %d, got %d", defaultPolicy.MaxAttempts, workflow.RetryPolicy.MaxAttempts)
	}
	if workflow.RetryPolicy.MaxInterval != defaultPolicy.MaxInterval {
		t.Errorf("expected max interval %d, got %d", defaultPolicy.MaxInterval, workflow.RetryPolicy.MaxInterval)
	}

	// Validate no retry attempts were made
	if workflow.RetryState.Attempts != 0 {
		t.Errorf("expected 0 retry attempts, got %d", workflow.RetryState.Attempts)
	}

	tp.Close()
}
