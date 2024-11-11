package cq

import (
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/types"
)

type CommandsQueries interface {
	CommandSubWorkflow(workflowEntityID types.WorkflowID, workflowExecutionID types.WorkflowExecutionID, stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error)
	QueryWorfklow(workflowFunc interface{}, id types.WorkflowID) *info.WorkflowInfo
	QueryNoWorkflow(err error) *info.WorkflowInfo
	HasSubWorkflowSuccessfullyCompleted(ctx types.SharedContext, stepID string) (types.WorkflowID, bool, error)
}
