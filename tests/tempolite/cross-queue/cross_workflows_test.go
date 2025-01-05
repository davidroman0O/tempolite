package tests

import (
	"context"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite"
)

func TestTempoliteCrossWorkflowsExecute(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_cross_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:    "side",
			MaxRuns: 1,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	var flagSideTriggered atomic.Bool
	sideWorkflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagSideTriggered.Store(true)
		return nil
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		if err := ctx.Workflow("sideflow", sideWorkflowFunc, &tempolite.WorkflowOptions{
			Queue: "side",
		}).Get(); err != nil {
			return err
		}
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

	if !flagSideTriggered.Load() {
		t.Fatal("sideWorkflowFunc was not triggered")
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Get parent workflow details
	parentID := future.WorkflowID()
	parentExecID := future.WorkflowExecutionID()

	parentWorkflow, err := database.GetWorkflowEntity(parentID)
	if err != nil {
		t.Fatal(err)
	}

	parentExecution, err := database.GetWorkflowExecution(parentExecID)
	if err != nil {
		t.Fatal(err)
	}

	// Get child workflow through hierarchy
	hierarchies, err := database.GetHierarchiesByParentEntity(int(parentID))
	if err != nil {
		t.Fatal(err)
	}
	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	hierarchy := hierarchies[0]
	childWorkflow, err := database.GetWorkflowEntity(tempolite.WorkflowEntityID(hierarchy.ChildEntityID))
	if err != nil {
		t.Fatal(err)
	}

	childExecution, err := database.GetWorkflowExecution(tempolite.WorkflowExecutionID(hierarchy.ChildExecutionID))
	if err != nil {
		t.Fatal(err)
	}

	// Validate Parent Workflow
	if parentWorkflow.Status != tempolite.StatusCompleted {
		t.Errorf("parent workflow status: expected %s, got %s", tempolite.StatusCompleted, parentWorkflow.Status)
	}
	if parentWorkflow.StepID != "root" {
		t.Errorf("parent workflow step ID: expected root, got %s", parentWorkflow.StepID)
	}
	if !parentWorkflow.WorkflowData.IsRoot {
		t.Error("parent workflow should be marked as root")
	}
	if parentWorkflow.QueueID != 1 { // Default queue
		t.Errorf("parent workflow queue ID: expected 1, got %d", parentWorkflow.QueueID)
	}

	// Validate Child Workflow
	if childWorkflow.Status != tempolite.StatusCompleted {
		t.Errorf("child workflow status: expected %s, got %s", tempolite.StatusCompleted, childWorkflow.Status)
	}
	if childWorkflow.StepID != "sideflow" {
		t.Errorf("child workflow step ID: expected sideflow, got %s", childWorkflow.StepID)
	}
	if childWorkflow.QueueID != 2 { // Side queue
		t.Errorf("child workflow queue ID: expected 2, got %d", childWorkflow.QueueID)
	}

	// Validate Child Workflow Data Relationships
	childData := childWorkflow.WorkflowData
	if childData.IsRoot {
		t.Error("child workflow should not be marked as root")
	}
	if childData.WorkflowFrom == nil || *childData.WorkflowFrom != parentID {
		t.Errorf("child workflow_from: expected %d, got %v", parentID, childData.WorkflowFrom)
	}
	if childData.WorkflowExecutionFrom == nil || *childData.WorkflowExecutionFrom != parentExecID {
		t.Errorf("child workflow_execution_from: expected %d, got %v", parentExecID, childData.WorkflowExecutionFrom)
	}
	if childData.WorkflowStepID == nil || *childData.WorkflowStepID != "root" {
		t.Error("child workflow_step_id should be 'root'")
	}

	// Validate Hierarchy Relationships
	if hierarchy.ParentEntityID != int(parentID) {
		t.Errorf("hierarchy parent entity ID: expected %d, got %d", parentID, hierarchy.ParentEntityID)
	}
	if hierarchy.ParentExecutionID != int(parentExecID) {
		t.Errorf("hierarchy parent execution ID: expected %d, got %d", parentExecID, hierarchy.ParentExecutionID)
	}
	if hierarchy.ParentStepID != "root" {
		t.Errorf("hierarchy parent step ID: expected root, got %s", hierarchy.ParentStepID)
	}
	if hierarchy.ChildStepID != "sideflow" {
		t.Errorf("hierarchy child step ID: expected sideflow, got %s", hierarchy.ChildStepID)
	}
	if hierarchy.ParentType != tempolite.EntityWorkflow {
		t.Errorf("hierarchy parent type: expected %s, got %s", tempolite.EntityWorkflow, hierarchy.ParentType)
	}
	if hierarchy.ChildType != tempolite.EntityWorkflow {
		t.Errorf("hierarchy child type: expected %s, got %s", tempolite.EntityWorkflow, hierarchy.ChildType)
	}

	// Validate Executions
	if parentExecution.Status != tempolite.ExecutionStatusCompleted {
		t.Errorf("parent execution status: expected %s, got %s", tempolite.ExecutionStatusCompleted, parentExecution.Status)
	}
	if childExecution.Status != tempolite.ExecutionStatusCompleted {
		t.Errorf("child execution status: expected %s, got %s", tempolite.ExecutionStatusCompleted, childExecution.Status)
	}

	// Validate runs share same runID
	if parentWorkflow.RunID != childWorkflow.RunID {
		t.Errorf("workflows should share same run ID: parent %d, child %d", parentWorkflow.RunID, childWorkflow.RunID)
	}

	tp.Close()
}
