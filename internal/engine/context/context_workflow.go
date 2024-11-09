package context

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/cq"
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/types"
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

func (w WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var id types.WorkflowID
	var err error
	if id, err = w.cq.CommandSubWorkflow(types.WorkflowID(w.workflowID), types.WorkflowExecutionID(w.executionID), stepID, workflowFunc, options, params...); err != nil {
		return w.cq.QueryNoWorkflow(err)
	}
	return w.cq.QueryWorfklow(workflowFunc, id)
}
