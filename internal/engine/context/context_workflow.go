package context

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/types"
)

type WorkflowContext struct {
	context.Context
	workflowID      string
	executionID     string
	runID           string
	workflowType    string
	stepID          string
	handlerIdentity types.HandlerIdentity
	queueName       string
}
