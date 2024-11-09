package queries

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
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
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		return e.QueryNoWorkflow(err)
	}
	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		return e.QueryNoWorkflow(err)
	}
	return info.NewWorkflowInfo(id, types.HandlerInfo(workflow), e.db, e.info)
}

func (e *Queries) QueryNoWorkflow(err error) *info.WorkflowInfo {
	return info.NewWorkflowInfoWithError(err)
}
