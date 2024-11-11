package queries

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type Queries struct {
	ctx      context.Context
	db       repository.Repository
	registry *registry.Registry
	info     *info.InfoClock
}

func New(ctx context.Context, db repository.Repository, registry *registry.Registry) *Queries {
	return &Queries{
		ctx:      ctx,
		db:       db,
		registry: registry,
		info:     info.New(ctx),
	}
}

func (e *Queries) Stop() {
	e.info.Stop()
}

func (e *Queries) QueryWorfklow(workflowFunc interface{}, id types.WorkflowID) *info.WorkflowInfo {
	var err error
	var identity types.HandlerIdentity
	logs.Debug(e.ctx, "Query Workflow getting identity", "id", id)
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		logs.Error(e.ctx, "Query Workflow error getting identity", "error", err, "id", id)
		return e.QueryNoWorkflow(err)
	}
	logs.Debug(e.ctx, "Query Workflow getting workflow", "id", id, "identity", identity)
	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		logs.Error(e.ctx, "Query Workflow error getting workflow", "error", err, "id", id)
		return e.QueryNoWorkflow(err)
	}
	logs.Debug(e.ctx, "Query Workflow", "id", id, "workflow", workflow)
	return info.NewWorkflowInfo(e.ctx, id, types.HandlerInfo(workflow), e.db, e.info)
}

func (e *Queries) QueryNoWorkflow(err error) *info.WorkflowInfo {
	return info.NewWorkflowInfoWithError(e.ctx, err)
}

func (e *Queries) HasSubWorkflowSuccessfullyCompleted(ctx types.SharedContext, stepID string) (types.WorkflowID, bool, error) {

	tx, err := e.db.Tx()
	if err != nil {
		logs.Error(e.ctx, "HasSubWorkflowSuccessfullyCompleted error getting tx", "error", err)
		return types.NoWorkflowID, false, err
	}

	info, err := e.db.
		Hierarchies().
		GetExisting(tx, ctx.RunID(), ctx.EntityID(), ctx.StepID(), stepID, hierarchy.ChildTypeWorkflow)
	if err != nil {
		logs.Error(e.ctx, "HasSubWorkflowSuccessfullyCompleted error getting hierarchy", "error", err)
		return types.NoWorkflowID, false, err
	}

	childInfo, err := e.db.Workflows().Get(tx, info.ChildEntityID)
	if err != nil {
		logs.Error(e.ctx, "HasSubWorkflowSuccessfullyCompleted error getting child info", "error", err)
		return types.NoWorkflowID, false, err
	}

	if childInfo.Status != "Completed" {
		logs.Debug(e.ctx, "HasSubWorkflowSuccessfullyCompleted workflow is NOT completed", "status", childInfo.Status, "id", childInfo.ID)
		// no error, no id, just not completed
		return types.NoWorkflowID, false, nil
	}

	logs.Debug(e.ctx, "HasSubWorkflowSuccessfullyCompleted workflow is completed", "status", childInfo.Status, "id", childInfo.ID)
	// everything is good, you can try to get the output
	return types.WorkflowID(childInfo.ID), true, nil
}
