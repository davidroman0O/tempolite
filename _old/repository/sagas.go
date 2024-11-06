// repository/saga.go
package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutions"
	"github.com/davidroman0O/tempolite/internal/persistence/repository/base"
)

type SagaRepository struct {
	entityRepo    *base.EntityRepository
	executionRepo *base.ExecutionRepository
	client        *ent.Client
}

type SagaOptions struct {
	base.EntityOptions
	stepType *string
}

type SagaOption func(*SagaOptions)

func WithSagaHandlerName(name string) SagaOption {
	return func(o *SagaOptions) {
		o.HandlerName = name
	}
}

func WithSagaStepID(id string) SagaOption {
	return func(o *SagaOptions) {
		o.StepID = id
	}
}

func WithSagaRunID(id int) SagaOption {
	return func(o *SagaOptions) {
		o.RunID = id
	}
}

func WithSagaPagination(limit, offset int) SagaOption {
	return func(o *SagaOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func WithSagaStepType(stepType string) SagaOption {
	return func(o *SagaOptions) {
		o.stepType = &stepType
	}
}

func NewSagaRepository(client *ent.Client) *SagaRepository {
	return &SagaRepository{
		client:        client,
		entityRepo:    base.NewEntityRepository(client),
		executionRepo: base.NewExecutionRepository(client),
	}
}

func (r *SagaRepository) Create(ctx context.Context, tx *ent.Tx, opts ...SagaOption) (*ent.Entity, error) {
	options := &SagaOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityHandlerName(options.HandlerName),
		base.WithEntityType("Saga"),
		base.WithEntityStepID(options.StepID),
		base.WithEntityRunID(options.RunID),
	}

	return r.entityRepo.Create(tx, entityOpts...)
}

func (r *SagaRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Entity, error) {
	return r.entityRepo.Get(tx, id, base.WithEntityContext(ctx))
}

func (r *SagaRepository) List(ctx context.Context, tx *ent.Tx, opts ...SagaOption) ([]*ent.Entity, error) {
	options := &SagaOptions{}
	for _, opt := range opts {
		opt(options)
	}

	entityOpts := []base.EntityOption{
		base.WithEntityContext(ctx),
		base.WithEntityType("Saga"),
	}

	if options.Limit != nil {
		entityOpts = append(entityOpts, base.WithEntityPagination(*options.Limit, *options.Offset))
	}

	return r.entityRepo.List(tx, entityOpts...)
}

func (r *SagaRepository) Update(ctx context.Context, tx *ent.Tx, id int, opts ...SagaOption) (*ent.Entity, error) {
	options := &SagaOptions{}
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

func (r *SagaRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	return r.entityRepo.Delete(tx, id, base.WithEntityContext(ctx))
}

// Execution Management with Saga-specific additions
func (r *SagaRepository) CreateExecution(ctx context.Context, tx *ent.Tx, sagaID int, stepType string) (*ent.SagaExecutions, error) {
	if stepType != "Transaction" && stepType != "Compensation" {
		return nil, fmt.Errorf("invalid step type: %s", stepType)
	}

	return tx.SagaExecutions.Create().
		SetStepType(sagaexecutions.StepType(stepType)).
		Save(ctx)
}

func (r *SagaRepository) GetExecution(ctx context.Context, tx *ent.Tx, executionID int) (*ent.SagaExecutions, error) {
	return tx.SagaExecutions.Get(ctx, executionID)
}

func (r *SagaRepository) ListExecutions(ctx context.Context, tx *ent.Tx, sagaID int, limit, offset int) ([]*ent.SagaExecutions, error) {
	query := tx.SagaExecutions.Query()

	if limit > 0 {
		query.Limit(limit)
	}
	if offset > 0 {
		query.Offset(offset)
	}

	return query.All(ctx)
}

func (r *SagaRepository) DeleteExecution(ctx context.Context, tx *ent.Tx, executionID int) error {
	return tx.SagaExecutions.DeleteOneID(executionID).Exec(ctx)
}

// Specific Saga operations
func (r *SagaRepository) ListTransactionExecutions(ctx context.Context, tx *ent.Tx, sagaID int) ([]*ent.SagaExecutions, error) {
	return tx.SagaExecutions.Query().
		Where(sagaexecutions.StepTypeEQ(sagaexecutions.StepTypeTransaction)).
		All(ctx)
}

func (r *SagaRepository) ListCompensationExecutions(ctx context.Context, tx *ent.Tx, sagaID int) ([]*ent.SagaExecutions, error) {
	return tx.SagaExecutions.Query().
		Where(sagaexecutions.StepTypeEQ(sagaexecutions.StepTypeCompensation)).
		All(ctx)
}
