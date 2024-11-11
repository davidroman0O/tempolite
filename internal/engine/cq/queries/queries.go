package queries

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
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

// HasParentChildStepID checks if the parent has a child with the given stepID
// We only need the parent context and the child stepID
func (e *Queries) HasParentChildStepID(ctx types.SharedContext, stepID string) (bool, error) {

	logs.Debug(e.ctx, "HasParentChildStepID", "runID", ctx.RunID(), "stepID", stepID, "parentWorkflowID", ctx.EntityID(), "parentExecutionID", ctx.ExecutionID(), "parentType", ctx.EntityType(), "parentQueue", ctx.QueueName(), "parentStepID", ctx.StepID())

	tx, err := e.db.Tx()
	if err != nil {
		logs.Error(e.ctx, "HasParentChildStepID error getting tx", "error", err)
		return false, err
	}

	var hasHierarchy bool

	if hasHierarchy, err = e.db.Hierarchies().HasHierarchy(tx, ctx.RunID(), ctx.EntityID(), stepID); err != nil {
		logs.Error(e.ctx, "HasParentChildStepID error checking hierarchy", "error", err)
		return false, err
	}

	if err = tx.Commit(); err != nil {
		logs.Error(e.ctx, "HasParentChildStepID error committing tx", "error", err)
		return false, err
	}

	return hasHierarchy, nil
}

func (e *Queries) GetHierarchy(ctx types.SharedContext, stepID string) (*repository.HierarchyInfo, error) {
	tx, err := e.db.Tx()
	if err != nil {
		logs.Error(e.ctx, "GetHierarchy error getting tx", "error", err)
		return nil, err
	}

	var hierarchy *repository.HierarchyInfo

	if hierarchy, err = e.db.Hierarchies().GetExisting(tx, ctx.RunID(), ctx.EntityID(), stepID); err != nil {
		logs.Error(e.ctx, "GetHierarchy error getting hierarchy", "error", err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		logs.Error(e.ctx, "GetHierarchy error committing tx", "error", err)
		return nil, err
	}

	return hierarchy, nil
}

func (e *Queries) GetWorkflow(ctx types.SharedContext, id types.WorkflowID) (*repository.WorkflowInfo, error) {
	tx, err := e.db.Tx()
	if err != nil {
		logs.Error(e.ctx, "GetWorkflow error getting tx", "error", err)
		return nil, err
	}

	var workflow *repository.WorkflowInfo

	if workflow, err = e.db.Workflows().Get(tx, id.ID()); err != nil {
		logs.Error(e.ctx, "GetWorkflow error getting workflow", "error", err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		logs.Error(e.ctx, "GetWorkflow error committing tx", "error", err)
		return nil, err
	}

	return workflow, nil
}
