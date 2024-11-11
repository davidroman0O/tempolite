package cq

import (
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
)

type CommandsQueries interface {
	CommandSubWorkflow(workflowEntityID types.WorkflowID, workflowExecutionID types.WorkflowExecutionID, stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error)
	QueryWorfklow(workflowFunc interface{}, id types.WorkflowID) *info.WorkflowInfo
	QueryNoWorkflow(err error) *info.WorkflowInfo
	HasParentChildStepID(ctx types.SharedContext, stepID string) (bool, error)
	GetHierarchy(ctx types.SharedContext, stepID string) (*repository.HierarchyInfo, error)
	GetWorkflow(ctx types.SharedContext, id types.WorkflowID) (*repository.WorkflowInfo, error)
	CommandSubWorkflowExecution(workflowEntityID types.WorkflowID) (types.WorkflowID, error)
	GetWorkflowInfo(id types.WorkflowID) *info.WorkflowInfo
}
