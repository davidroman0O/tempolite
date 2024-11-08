package context

import (
	"context"

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
}

func NewWorkflowContext(
	ctx context.Context,
	workflowID int,
	executionID int,
	runID int,
	stepID string,
	queueName string,
	handlerIdentity types.HandlerIdentity,
) WorkflowContext {
	return WorkflowContext{
		Context:         ctx,
		workflowID:      workflowID,
		executionID:     executionID,
		runID:           runID,
		stepID:          stepID,
		handlerIdentity: handlerIdentity,
		queueName:       queueName,
	}
}
