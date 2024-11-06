package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
)

type QueueRepository struct {
	client *ent.Client
}

type QueueOptions struct {
	Name   string
	Limit  *int
	Offset *int
}

type QueueOption func(*QueueOptions)

func WithQueueName(name string) QueueOption {
	return func(o *QueueOptions) {
		o.Name = name
	}
}

func WithQueuePagination(limit, offset int) QueueOption {
	return func(o *QueueOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func NewQueueRepository(client *ent.Client) *QueueRepository {
	return &QueueRepository{
		client: client,
	}
}

func (r *QueueRepository) Create(ctx context.Context, tx *ent.Tx, opts ...QueueOption) (*ent.Queue, error) {
	options := &QueueOptions{}
	for _, opt := range opts {
		opt(options)
	}

	queue, err := tx.Queue.Create().
		SetName(options.Name).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating queue: %w", err)
	}

	return queue, nil
}

func (r *QueueRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Queue, error) {
	queue, err := tx.Queue.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed getting queue: %w", err)
	}
	return queue, nil
}

func (r *QueueRepository) List(ctx context.Context, tx *ent.Tx, opts ...QueueOption) ([]*ent.Queue, error) {
	options := &QueueOptions{}
	for _, opt := range opts {
		opt(options)
	}

	query := tx.Queue.Query()

	if options.Name != "" {
		query.Where(queue.NameEQ(options.Name))
	}
	if options.Limit != nil {
		query.Limit(*options.Limit)
	}
	if options.Offset != nil {
		query.Offset(*options.Offset)
	}

	return query.All(ctx)
}

func (r *QueueRepository) Update(ctx context.Context, tx *ent.Tx, id int, opts ...QueueOption) (*ent.Queue, error) {
	options := &QueueOptions{}
	for _, opt := range opts {
		opt(options)
	}

	update := tx.Queue.UpdateOneID(id)
	if options.Name != "" {
		update.SetName(options.Name)
	}

	return update.Save(ctx)
}

func (r *QueueRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	return tx.Queue.DeleteOneID(id).Exec(ctx)
}

func (r *QueueRepository) AddEntity(ctx context.Context, tx *ent.Tx, queueID, entityID int) error {
	entity, err := tx.Entity.Get(ctx, entityID)
	if err != nil {
		return fmt.Errorf("failed getting entity: %w", err)
	}

	_, err = entity.Update().
		AddQueueIDs(queueID).
		Save(ctx)
	return err
}

func (r *QueueRepository) RemoveEntity(ctx context.Context, tx *ent.Tx, queueID, entityID int) error {
	entity, err := tx.Entity.Get(ctx, entityID)
	if err != nil {
		return fmt.Errorf("failed getting entity: %w", err)
	}

	_, err = entity.Update().
		RemoveQueueIDs(queueID).
		Save(ctx)
	return err
}

func (r *QueueRepository) ListEntities(ctx context.Context, tx *ent.Tx, queueID int) ([]*ent.Entity, error) {
	return tx.Entity.Query().
		Where(entity.HasQueuesWith(queue.ID(queueID))).
		All(ctx)
}

func (r *QueueRepository) ListEntitiesByType(ctx context.Context, tx *ent.Tx, queueID int, entityType string) ([]*ent.Entity, error) {
	return tx.Entity.Query().
		Where(
			entity.HasQueuesWith(queue.ID(queueID)),
			entity.TypeEQ(entity.Type(entityType)),
		).
		All(ctx)
}
