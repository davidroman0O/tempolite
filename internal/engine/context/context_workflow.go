package context

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/cq"
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type WorkflowContext struct {
	context.Context
	workflowID  int
	executionID int
	runID       int
	stepID      string
	handler     types.HandlerInfo
	queueName   string
	cq          cq.CommandsQueries
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

func (w WorkflowContext) Handler() types.HandlerInfo {
	return w.handler
}

func NewWorkflowContext(
	ctx context.Context,
	workflowID int,
	executionID int,
	runID int,
	stepID string,
	queueName string,
	handler types.HandlerInfo,
	cq cq.CommandsQueries,
) WorkflowContext {
	return WorkflowContext{
		Context:     ctx,
		workflowID:  workflowID,
		executionID: executionID,
		runID:       runID,
		stepID:      stepID,
		handler:     handler,
		queueName:   queueName,
		cq:          cq,
	}
}

// Create a sub-workflow for a current workflow
//
// Here the pseudo code for the workflow creation:
//
// ```
// => Sub Workflow stepID
// -	- has hierarchy parent RunID + ParentID + StepID?
// -   		true:
// - 			- was entity completed? was execution completed?
// -   			true:
// - 				- return output (return workflow info)
// -   			false:
// -  				- create workflow execution
// -   		false:
// - 			- create workflow entity
// - 			- create workflow execution
// - 	- return workflow info (and wait)
// ```
func (w WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var err error
	var id types.WorkflowID
	var hasHierarchy bool

	logs.Debug(w.Context, "Creating sub workflow", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)

	// Before creating a sub-workflow, we need to check if the current workflow has been completed successfully
	if hasHierarchy, err = w.cq.HasParentChildStepID(w, stepID); err != nil {
		logs.Error(w.Context, "Error checking if sub workflow has been completed", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
		return w.cq.QueryNoWorkflow(err)
	}

	// Meaning it's not the first time
	if hasHierarchy {
		var hInfo *repository.HierarchyInfo
		if hInfo, err = w.cq.GetHierarchy(w, stepID); err != nil {
			logs.Error(w.Context, "Error getting hierarchy", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
			return w.cq.QueryNoWorkflow(err)
		}

		// if we have a hierarchy, then we need to kow if the workflow entity is completed
		var wInfo *repository.WorkflowInfo
		if wInfo, err = w.cq.GetWorkflow(w, types.WorkflowID(hInfo.ChildEntityID)); err != nil {
			logs.Error(w.Context, "Error getting workflow", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
			return w.cq.QueryNoWorkflow(err)
		}

		// if it is completed, then we can just return the output
		// but for that, we just need to return a WorkflowInfo
		if wInfo.Status == string(entity.StatusCompleted) {
			logs.Debug(w.Context, "Sub workflow has been completed", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
			return w.cq.QueryWorfklow(workflowFunc, types.WorkflowID(hInfo.ChildEntityID))
		}

		// if it is not completed, then we need to create a new workflow execution
		if id, err = w.cq.CommandSubWorkflowExecution(types.WorkflowID(hInfo.ChildEntityID)); err != nil {
			logs.Error(w.Context, "Error creating sub workflow execution", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
			return w.cq.QueryNoWorkflow(err)
		}

	} else {
		// If we don't have a hierarchy, that mean we need to create a whole new entity with execution
		logs.Debug(w.Context, "Creating sub workflow", "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
		if id, err = w.cq.CommandSubWorkflow(types.WorkflowID(w.workflowID), types.WorkflowExecutionID(w.executionID), stepID, workflowFunc, options, params...); err != nil {
			logs.Error(w.Context, "Error creating sub workflow", "error", err, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID)
			return w.cq.QueryNoWorkflow(err)
		}
	}

	logs.Debug(w.Context, "Sub workflow created", "hasHierarchy", hasHierarchy, "workflowID", w.EntityID(), "executionID", w.ExecutionID(), "stepID", stepID, "subWorkflowID", id)
	return w.cq.QueryWorfklow(workflowFunc, id)
}
