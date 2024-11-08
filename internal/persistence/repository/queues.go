package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
)

type QueueInfo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrQueues = fmt.Errorf("queues error")
)

type QueueRepository interface {
	Create(tx *ent.Tx, name string) (*QueueInfo, error)
	Get(tx *ent.Tx, id int) (*QueueInfo, error)
	GetByName(tx *ent.Tx, name string) (*QueueInfo, error)
	List(tx *ent.Tx) ([]*QueueInfo, error)
	Delete(tx *ent.Tx, id int) error
	Update(tx *ent.Tx, id int, name string) (*QueueInfo, error)
}

type queueRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewQueueRepository(ctx context.Context, client *ent.Client) QueueRepository {
	return &queueRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *queueRepository) Create(tx *ent.Tx, name string) (*QueueInfo, error) {
	exists, err := tx.Queue.Query().
		Where(queue.NameEQ(name)).
		Exist(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("checking queue existence"))
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	queueObj, err := tx.Queue.Create().
		SetName(name).
		Save(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating queue"))
	}

	return &QueueInfo{
		ID:        queueObj.ID,
		Name:      queueObj.Name,
		CreatedAt: queueObj.CreatedAt,
		UpdatedAt: queueObj.UpdatedAt,
	}, nil
}

func (r *queueRepository) Get(tx *ent.Tx, id int) (*QueueInfo, error) {
	queueObj, err := tx.Queue.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.Join(ErrQueues, ErrNotFound)
		}
		return nil, errors.Join(err, fmt.Errorf("getting queue"))
	}

	return &QueueInfo{
		ID:        queueObj.ID,
		Name:      queueObj.Name,
		CreatedAt: queueObj.CreatedAt,
		UpdatedAt: queueObj.UpdatedAt,
	}, nil
}

func (r *queueRepository) GetByName(tx *ent.Tx, name string) (*QueueInfo, error) {
	queueObj, err := tx.Queue.Query().
		Where(queue.NameEQ(name)).
		Only(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.Join(ErrQueues, ErrNotFound)
		}
		return nil, errors.Join(ErrQueues, fmt.Errorf("getting queue by name: %w", err))
	}

	return &QueueInfo{
		ID:        queueObj.ID,
		Name:      queueObj.Name,
		CreatedAt: queueObj.CreatedAt,
		UpdatedAt: queueObj.UpdatedAt,
	}, nil
}

func (r *queueRepository) List(tx *ent.Tx) ([]*QueueInfo, error) {
	queueObjs, err := tx.Queue.Query().All(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("listing queues"))
	}

	result := make([]*QueueInfo, len(queueObjs))
	for i, queueObj := range queueObjs {
		result[i] = &QueueInfo{
			ID:        queueObj.ID,
			Name:      queueObj.Name,
			CreatedAt: queueObj.CreatedAt,
			UpdatedAt: queueObj.UpdatedAt,
		}
	}
	return result, nil
}

func (r *queueRepository) Delete(tx *ent.Tx, id int) error {
	exists, err := tx.Queue.Query().
		Where(queue.IDEQ(id)).
		Exist(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("checking queue existence"))
	}
	if !exists {
		return errors.Join(ErrQueues, ErrNotFound)
	}

	inUse, err := tx.Queue.Query().
		Where(queue.IDEQ(id)).
		QueryEntities().
		Exist(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("checking queue usage"))
	}
	if inUse {
		return errors.Join(err, fmt.Errorf("queue is in use by entities"))
	}

	err = tx.Queue.DeleteOneID(id).Exec(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("deleting queue"))
	}

	return nil
}

func (r *queueRepository) Update(tx *ent.Tx, id int, name string) (*QueueInfo, error) {
	exists, err := tx.Queue.Query().
		Where(queue.And(
			queue.NameEQ(name),
			queue.IDNEQ(id),
		)).
		Exist(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("checking queue name existence"))
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	queueObj, err := tx.Queue.UpdateOneID(id).
		SetName(name).
		Save(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.Join(ErrQueues, ErrNotFound)
		}
		return nil, errors.Join(err, fmt.Errorf("updating queue"))
	}

	return &QueueInfo{
		ID:        queueObj.ID,
		Name:      queueObj.Name,
		CreatedAt: queueObj.CreatedAt,
		UpdatedAt: queueObj.UpdatedAt,
	}, nil
}
