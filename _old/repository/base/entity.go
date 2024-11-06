// base/entity.go
package base

import (
	"context"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
)

type EntityRepository struct {
	client *ent.Client
}

type EntityOptions struct {
	Ctx         context.Context
	HandlerName string
	EntityType  string
	StepID      string
	RunID       int
	Limit       *int
	Offset      *int
}

type EntityOption func(*EntityOptions)

func WithEntityContext(ctx context.Context) EntityOption {
	return func(o *EntityOptions) {
		o.Ctx = ctx
	}
}

func WithEntityHandlerName(name string) EntityOption {
	return func(o *EntityOptions) {
		o.HandlerName = name
	}
}

func WithEntityType(t string) EntityOption {
	return func(o *EntityOptions) {
		o.EntityType = t
	}
}

func WithEntityStepID(id string) EntityOption {
	return func(o *EntityOptions) {
		o.StepID = id
	}
}

func WithEntityRunID(id int) EntityOption {
	return func(o *EntityOptions) {
		o.RunID = id
	}
}

func WithEntityPagination(limit, offset int) EntityOption {
	return func(o *EntityOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func NewEntityRepository(client *ent.Client) *EntityRepository {
	return &EntityRepository{client: client}
}

func (r *EntityRepository) Create(tx *ent.Tx, opts ...EntityOption) (*ent.Entity, error) {
	options := &EntityOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Entity.Create().
		SetHandlerName(options.HandlerName).
		SetType(entity.Type(options.EntityType)).
		SetStepID(options.StepID).
		SetRunID(options.RunID).
		Save(options.Ctx)
}

func (r *EntityRepository) Get(tx *ent.Tx, id int, opts ...EntityOption) (*ent.Entity, error) {
	options := &EntityOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Entity.Get(options.Ctx, id)
}

func (r *EntityRepository) Update(tx *ent.Tx, id int, opts ...EntityOption) (*ent.Entity, error) {
	options := &EntityOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	update := tx.Entity.UpdateOneID(id)
	if options.HandlerName != "" {
		update.SetHandlerName(options.HandlerName)
	}
	if options.EntityType != "" {
		update.SetType(entity.Type(options.EntityType))
	}
	if options.StepID != "" {
		update.SetStepID(options.StepID)
	}

	return update.Save(options.Ctx)
}

func (r *EntityRepository) Delete(tx *ent.Tx, id int, opts ...EntityOption) error {
	options := &EntityOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Entity.DeleteOneID(id).Exec(options.Ctx)
}

func (r *EntityRepository) List(tx *ent.Tx, opts ...EntityOption) ([]*ent.Entity, error) {
	options := &EntityOptions{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(options)
	}

	query := tx.Entity.Query()

	if options.EntityType != "" {
		query.Where(entity.TypeEQ(entity.Type(options.EntityType)))
	}
	if options.RunID != 0 {
		query.Where(entity.HasRunWith(run.ID(options.RunID)))
	}
	if options.Limit != nil {
		query.Limit(*options.Limit)
	}
	if options.Offset != nil {
		query.Offset(*options.Offset)
	}

	return query.All(options.Ctx)
}
