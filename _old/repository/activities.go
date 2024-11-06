// repository/activity.go
package repository

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository/base"
)

type ActivityRepository struct {
	entityRepo    *base.EntityRepository
	executionRepo *base.ExecutionRepository
	client        *ent.Client
}

type ActivityOptions struct {
	base.EntityOptions
}

type ActivityOption func(*ActivityOptions)

func WithActivityHandlerName(name string) ActivityOption {
	return func(o *ActivityOptions) {
		o.HandlerName = name
	}
}

func WithActivityStepID(id string) ActivityOption {
	return func(o *ActivityOptions) {
		o.StepID = id
	}
}

func WithActivityRunID(id int) ActivityOption {
	return func(o *ActivityOptions) {
		o.RunID = id
	}
}

func WithActivityPagination(limit, offset int) ActivityOption {
	return func(o *ActivityOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func NewActivityRepository(client *ent.Client) *ActivityRepository {
	return &ActivityRepository{
		client:        client,
		entityRepo:    base.NewEntityRepository(client),
		executionRepo: base.NewExecutionRepository(client),
	}
}

func (r *ActivityRepository) Create(ctx context.Context, tx *ent.Tx, opts ...ActivityOption) (*ent.Entity, error) {
	options := &ActivityOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityHandlerName(options.HandlerName),
		base.WithEntityType("Activity"),
		base.WithEntityStepID(options.StepID),
		base.WithEntityRunID(options.RunID),
	}

	return r.entityRepo.Create(tx, entityOpts...)
}

func (r *ActivityRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Entity, error) {
	return r.entityRepo.Get(tx, id, base.WithEntityType("Activity"))
}

func (r *ActivityRepository) List(ctx context.Context, tx *ent.Tx, opts ...ActivityOption) ([]*ent.Entity, error) {
	options := &ActivityOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityType("Activity"),
	}

	if options.Limit != nil {
		entityOpts = append(entityOpts, base.WithEntityPagination(*options.Limit, *options.Offset))
	}

	return r.entityRepo.List(tx, entityOpts...)
}

func (r *ActivityRepository) Update(ctx context.Context, tx *ent.Tx, id int, opts ...ActivityOption) (*ent.Entity, error) {
	options := &ActivityOptions{}
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

func (r *ActivityRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	return r.entityRepo.Delete(tx, id, base.WithEntityType("Activity"))
}

func (r *ActivityRepository) CreateExecution(ctx context.Context, tx *ent.Tx, activityID int) (*ent.Execution, error) {
	return r.executionRepo.Create(tx, base.WithExecutionEntityID(activityID))
}

func (r *ActivityRepository) GetExecution(ctx context.Context, tx *ent.Tx, executionID int) (*ent.Execution, error) {
	return r.executionRepo.Get(tx, executionID)
}

func (r *ActivityRepository) DeleteExecution(ctx context.Context, tx *ent.Tx, executionID int) error {
	return r.executionRepo.Delete(tx, executionID)
}

func (r *ActivityRepository) ListExecutions(ctx context.Context, tx *ent.Tx, activityID int, limit, offset int) ([]*ent.Execution, error) {
	return r.executionRepo.List(tx,
		base.WithExecutionEntityID(activityID),
		base.WithExecutionPagination(limit, offset),
	)
}
