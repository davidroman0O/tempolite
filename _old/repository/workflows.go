// repository/workflow.go
package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
	"github.com/davidroman0O/tempolite/internal/persistence/repository/base"
)

type WorkflowRepository struct {
	entityRepo    *base.EntityRepository
	executionRepo *base.ExecutionRepository
	client        *ent.Client
}

type WorkflowOptions struct {
	base.EntityOptions
	paused    *bool
	resumable *bool
}

type WorkflowOption func(*WorkflowOptions)

func WithWorkflowPaused(paused bool) WorkflowOption {
	return func(o *WorkflowOptions) {
		o.paused = &paused
	}
}

func WithWorkflowResumable(resumable bool) WorkflowOption {
	return func(o *WorkflowOptions) {
		o.resumable = &resumable
	}
}

func WithWorkflowHandlerName(name string) WorkflowOption {
	return func(o *WorkflowOptions) {
		o.HandlerName = name
	}
}

func WithWorkflowStepID(id string) WorkflowOption {
	return func(o *WorkflowOptions) {
		o.StepID = id
	}
}

func WithWorkflowRunID(id int) WorkflowOption {
	return func(o *WorkflowOptions) {
		o.RunID = id
	}
}

func NewWorkflowRepository(client *ent.Client) *WorkflowRepository {
	return &WorkflowRepository{
		client:        client,
		entityRepo:    base.NewEntityRepository(client),
		executionRepo: base.NewExecutionRepository(client),
	}
}

func (r *WorkflowRepository) Create(ctx context.Context, tx *ent.Tx, opts ...WorkflowOption) (*ent.Entity, error) {
	options := &WorkflowOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Create the entity first
	entity, err := tx.Entity.Create().
		SetHandlerName(options.HandlerName).
		SetType("Workflow").
		SetStepID(options.StepID).
		SetRunID(options.RunID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating workflow entity: %w", err)
	}

	// Create workflow data with the new entity ID
	_, err = tx.WorkflowData.Create().
		SetWorkflowID(entity.ID).
		SetInput([][]byte{}).
		SetStatus(workflowdata.StatusPending).
		SetRetryPolicy(schema.RetryPolicy{
			MaximumAttempts:    1,
			InitialInterval:    0,
			BackoffCoefficient: 0,
			MaximumInterval:    0,
		}).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating workflow data: %w", err)
	}

	return entity, nil
}

func (r *WorkflowRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Entity, error) {
	return r.entityRepo.Get(tx, id, base.WithEntityType("Workflow"))
}

func (r *WorkflowRepository) List(ctx context.Context, tx *ent.Tx, opts ...WorkflowOption) ([]*ent.Entity, error) {
	options := &WorkflowOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityType("Workflow"),
	}

	if options.Limit != nil {
		entityOpts = append(entityOpts, base.WithEntityPagination(*options.Limit, *options.Offset))
	}

	return r.entityRepo.List(tx, entityOpts...)
}

func (r *WorkflowRepository) Update(ctx context.Context, tx *ent.Tx, id int, opts ...WorkflowOption) (*ent.Entity, error) {
	options := &WorkflowOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityHandlerName(options.HandlerName),
		base.WithEntityStepID(options.StepID),
	}

	return r.entityRepo.Update(tx, id, entityOpts...)
}

func (r *WorkflowRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	rowsAffected, err := tx.WorkflowData.Delete().
		Where(workflowdata.WorkflowID(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("deleting workflow data: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no workflow data found for ID: %d", id)
	}

	return r.entityRepo.Delete(tx, id, base.WithEntityType("Workflow"))
}

func (r *WorkflowRepository) CreateExecution(ctx context.Context, tx *ent.Tx, workflowID int) (*ent.Execution, error) {
	return r.executionRepo.Create(tx, base.WithExecutionEntityID(workflowID))
}

func (r *WorkflowRepository) GetExecution(ctx context.Context, tx *ent.Tx, executionID int) (*ent.Execution, error) {
	return r.executionRepo.Get(tx, executionID)
}

func (r *WorkflowRepository) DeleteExecution(ctx context.Context, tx *ent.Tx, executionID int) error {
	return r.executionRepo.Delete(tx, executionID)
}

func (r *WorkflowRepository) ListExecutions(ctx context.Context, tx *ent.Tx, workflowID int, limit, offset int) ([]*ent.Execution, error) {
	return r.executionRepo.List(tx,
		base.WithExecutionEntityID(workflowID),
		base.WithExecutionPagination(limit, offset),
	)
}
