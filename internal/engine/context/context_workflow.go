package context

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/cq"
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type WorkflowContext struct {
	context.Context
	workflowID      int
	executionID     int
	runID           int
	stepID          string
	handlerIdentity types.HandlerIdentity
	queueName       string
	cq              cq.CommandsQueries
}

func (w WorkflowContext) RunID() int {
	return w.runID
}

func (w WorkflowContext) EntityID() int {
	return w.workflowID
}

func (w WorkflowContext) ExecutionID() int {
	return w.executionID
}

func (w WorkflowContext) EntityType() string {
	return "Workflow"
}

func (w WorkflowContext) StepID() string {
	return w.stepID
}

func (w WorkflowContext) QueueName() string {
	return w.queueName
}

func NewWorkflowContext(
	ctx context.Context,
	workflowID int,
	executionID int,
	runID int,
	stepID string,
	queueName string,
	handlerIdentity types.HandlerIdentity,
	cq cq.CommandsQueries,
) WorkflowContext {
	return WorkflowContext{
		Context:         ctx,
		workflowID:      workflowID,
		executionID:     executionID,
		runID:           runID,
		stepID:          stepID,
		handlerIdentity: handlerIdentity,
		queueName:       queueName,
		cq:              cq,
	}
}

// Create a sub-workflow for a current workflow
func (w WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var id types.WorkflowID
	var completed bool
	var wID types.WorkflowID
	var err error

	logs.Debug(w.Context, "Creating sub workflow", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
	// Before creating a sub-workflow, we need to check if the current workflow has been completed successfully
	if wID, completed, err = w.cq.HasSubWorkflowSuccessfullyCompleted(w, stepID); err != nil {
		logs.Error(w.Context, "Error checking if sub workflow has been completed", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
		return w.cq.QueryNoWorkflow(err)
	}

	// Allowing a parent failure to avoid creating unnecessary sub-workflows since it was already completed
	if completed {
		logs.Debug(w.Context, "Sub workflow has been completed", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
		return w.cq.QueryWorfklow(workflowFunc, wID)
	}

	logs.Debug(w.Context, "Creating sub workflow", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
	if id, err = w.cq.CommandSubWorkflow(types.WorkflowID(w.workflowID), types.WorkflowExecutionID(w.executionID), stepID, workflowFunc, options, params...); err != nil {
		logs.Error(w.Context, "Error creating sub workflow", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
		return w.cq.QueryNoWorkflow(err)
	}

	logs.Debug(w.Context, "Sub workflow created", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID, "subWorkflowID", id)
	return w.cq.QueryWorfklow(workflowFunc, id)
}
