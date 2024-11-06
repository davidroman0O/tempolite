// repository/sideeffect.go
package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository/base"
)

type SideEffectRepository struct {
	entityRepo    *base.EntityRepository
	executionRepo *base.ExecutionRepository
	client        *ent.Client
}

type SideEffectOptions struct {
	base.EntityOptions
	output [][]byte
}

type SideEffectOption func(*SideEffectOptions)

func WithSideEffectHandlerName(name string) SideEffectOption {
	return func(o *SideEffectOptions) {
		o.HandlerName = name
	}
}

func WithSideEffectStepID(id string) SideEffectOption {
	return func(o *SideEffectOptions) {
		o.StepID = id
	}
}

func WithSideEffectRunID(id int) SideEffectOption {
	return func(o *SideEffectOptions) {
		o.RunID = id
	}
}

func WithSideEffectPagination(limit, offset int) SideEffectOption {
	return func(o *SideEffectOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func WithSideEffectOutput(output [][]byte) SideEffectOption {
	return func(o *SideEffectOptions) {
		o.output = output
	}
}

func NewSideEffectRepository(client *ent.Client) *SideEffectRepository {
	return &SideEffectRepository{
		client:        client,
		entityRepo:    base.NewEntityRepository(client),
		executionRepo: base.NewExecutionRepository(client),
	}
}

func (r *SideEffectRepository) Create(ctx context.Context, tx *ent.Tx, opts ...SideEffectOption) (*ent.Entity, error) {
	options := &SideEffectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityHandlerName(options.HandlerName),
		base.WithEntityType("SideEffect"),
		base.WithEntityStepID(options.StepID),
		base.WithEntityRunID(options.RunID),
	}

	return r.entityRepo.Create(tx, entityOpts...)
}

func (r *SideEffectRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Entity, error) {
	return r.entityRepo.Get(tx, id, base.WithEntityType("SideEffect"))
}

func (r *SideEffectRepository) List(ctx context.Context, tx *ent.Tx, opts ...SideEffectOption) ([]*ent.Entity, error) {
	options := &SideEffectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityType("SideEffect"),
	}

	if options.Limit != nil {
		entityOpts = append(entityOpts, base.WithEntityPagination(*options.Limit, *options.Offset))
	}

	return r.entityRepo.List(tx, entityOpts...)
}

func (r *SideEffectRepository) Update(ctx context.Context, tx *ent.Tx, id int, opts ...SideEffectOption) (*ent.Entity, error) {
	options := &SideEffectOptions{}
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

func (r *SideEffectRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	return r.entityRepo.Delete(tx, id, base.WithEntityType("SideEffect"))
}

func (r *SideEffectRepository) CreateExecution(ctx context.Context, tx *ent.Tx, sideEffectID int) (*ent.SideEffectExecutions, error) {
	return tx.SideEffectExecutions.Create().Save(ctx)
}

func (r *SideEffectRepository) GetExecution(ctx context.Context, tx *ent.Tx, executionID int) (*ent.SideEffectExecutions, error) {
	return tx.SideEffectExecutions.Get(ctx, executionID)
}

func (r *SideEffectRepository) ListExecutions(ctx context.Context, tx *ent.Tx, sideEffectID int, limit, offset int) ([]*ent.SideEffectExecutions, error) {
	query := tx.SideEffectExecutions.Query()

	if limit > 0 {
		query.Limit(limit)
	}
	if offset > 0 {
		query.Offset(offset)
	}

	return query.All(ctx)
}

func (r *SideEffectRepository) DeleteExecution(ctx context.Context, tx *ent.Tx, executionID int) error {
	return tx.SideEffectExecutions.DeleteOneID(executionID).Exec(ctx)
}

func (r *SideEffectRepository) GetLatestExecution(ctx context.Context, tx *ent.Tx, sideEffectID int) (*ent.SideEffectExecutions, error) {
	executions, err := r.ListExecutions(ctx, tx, sideEffectID, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("getting latest side effect execution: %w", err)
	}
	if len(executions) == 0 {
		return nil, fmt.Errorf("no executions found for side effect: %d", sideEffectID)
	}
	return executions[0], nil
}
