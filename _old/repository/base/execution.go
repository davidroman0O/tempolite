// base/execution.go
package base

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
)

type ExecutionRepository struct {
	client *ent.Client
}

type ExecutionOptions struct {
	Ctx      context.Context
	EntityID int
	Limit    *int
	Offset   *int
}

type ExecutionOption func(*ExecutionOptions)

func WithExecutionContext(ctx context.Context) ExecutionOption {
	return func(o *ExecutionOptions) {
		o.Ctx = ctx
	}
}

func WithExecutionEntityID(id int) ExecutionOption {
	return func(o *ExecutionOptions) {
		o.EntityID = id
	}
}

func WithExecutionPagination(limit, offset int) ExecutionOption {
	return func(o *ExecutionOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func NewExecutionRepository(client *ent.Client) *ExecutionRepository {
	return &ExecutionRepository{client: client}
}

func (r *ExecutionRepository) Create(tx *ent.Tx, opts ...ExecutionOption) (*ent.Execution, error) {
	options := &ExecutionOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Execution.Create().
		SetEntityID(options.EntityID).
		Save(options.Ctx)
}

func (r *ExecutionRepository) Get(tx *ent.Tx, id int, opts ...ExecutionOption) (*ent.Execution, error) {
	options := &ExecutionOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Execution.Get(options.Ctx, id)
}

func (r *ExecutionRepository) Delete(tx *ent.Tx, id int, opts ...ExecutionOption) error {
	options := &ExecutionOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Execution.DeleteOneID(id).Exec(options.Ctx)
}

func (r *ExecutionRepository) List(tx *ent.Tx, opts ...ExecutionOption) ([]*ent.Execution, error) {
	options := &ExecutionOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	query := tx.Execution.Query()

	if options.EntityID != 0 {
		query.Where(execution.HasEntityWith(entity.ID(options.EntityID)))
	}
	if options.Limit != nil {
		query.Limit(*options.Limit)
	}
	if options.Offset != nil {
		query.Offset(*options.Offset)
	}

	return query.All(options.Ctx)
}
