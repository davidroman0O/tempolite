package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
)

// Repository errors
var (
	ErrNotFound         = fmt.Errorf("not found")
	ErrInvalidOperation = fmt.Errorf("invalid operation")
	ErrInvalidInput     = fmt.Errorf("invalid input")
	ErrTransactionScope = fmt.Errorf("operation must be executed within a transaction")
	ErrAlreadyExists    = fmt.Errorf("entity already exists")
	ErrInvalidState     = fmt.Errorf("invalid state")
)

// Common type definitions
type Status string

const (
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Repository represents the main repository interface
type Repository interface {
	Queues() QueueRepository
	Runs() RunRepository
	Hierarchies() HierarchyRepository
	Entities() EntityRepository
	Executions() ExecutionRepository
	Workflows() WorkflowRepository
	Activities() ActivityRepository
	Sagas() SagaRepository
	SideEffects() SideEffectRepository

	Tx() (*ent.Tx, error)
}

// repository implements Repository
type repository struct {
	client      *ent.Client
	queues      QueueRepository
	runs        RunRepository
	hierarchies HierarchyRepository
	entities    EntityRepository
	executions  ExecutionRepository
	workflows   WorkflowRepository
	activities  ActivityRepository
	sagas       SagaRepository
	sideeffects SideEffectRepository
}

// NewRepository creates a new repository instance with initialized sub-repositories
func NewRepository(ctx context.Context, client *ent.Client) Repository {
	r := &repository{
		client: client,
	}

	r.entities = NewEntityRepository(ctx, client)
	r.queues = NewQueueRepository(ctx, client)
	r.runs = NewRunRepository(ctx, client)
	r.hierarchies = NewHierarchyRepository(ctx, client)
	r.executions = NewExecutionRepository(ctx, client)
	r.workflows = NewWorkflowRepository(ctx, client)
	r.activities = NewActivityRepository(ctx, client)
	r.sagas = NewSagaRepository(ctx, client)
	r.sideeffects = NewSideEffectRepository(ctx, client)

	return r
}

// Sub-repository accessors
func (r *repository) Queues() QueueRepository {
	return r.queues
}

func (r *repository) Runs() RunRepository {
	return r.runs
}

func (r *repository) Hierarchies() HierarchyRepository {
	return r.hierarchies
}

func (r *repository) Entities() EntityRepository {
	return r.entities
}

func (r *repository) Executions() ExecutionRepository {
	return r.executions
}

func (r *repository) Workflows() WorkflowRepository {
	return r.workflows
}

func (r *repository) Activities() ActivityRepository {
	return r.activities
}

func (r *repository) Sagas() SagaRepository {
	return r.sagas
}

func (r *repository) SideEffects() SideEffectRepository {
	return r.sideeffects
}

func (r *repository) Tx() (*ent.Tx, error) {
	return r.client.Tx(context.Background())
}
