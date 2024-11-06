// repository/run.go
package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/repository/base"
)

type RunRepository struct {
	client        *ent.Client
	queueRepo     *QueueRepository
	executionRepo *base.ExecutionRepository
}

type RunOptions struct {
	Name   string
	Limit  *int
	Offset *int
}

type RunOption func(*RunOptions)

func WithRunName(name string) RunOption {
	return func(o *RunOptions) {
		o.Name = name
	}
}

func WithRunPagination(limit, offset int) RunOption {
	return func(o *RunOptions) {
		o.Limit = &limit
		o.Offset = &offset
	}
}

func NewRunRepository(client *ent.Client) *RunRepository {
	return &RunRepository{
		client:        client,
		queueRepo:     NewQueueRepository(client),
		executionRepo: base.NewExecutionRepository(client),
	}
}

// Create creates a new run and initializes the default queue
func (r *RunRepository) Create(ctx context.Context, tx *ent.Tx, opts ...RunOption) (*ent.Run, error) {
	options := &RunOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return tx.Run.Create().
		SetName(options.Name).
		Save(ctx)
}

func (r *RunRepository) Get(ctx context.Context, tx *ent.Tx, id int) (*ent.Run, error) {
	return tx.Run.Get(ctx, id)
}

func (r *RunRepository) List(ctx context.Context, tx *ent.Tx, opts ...RunOption) ([]*ent.Run, error) {
	options := &RunOptions{}
	for _, opt := range opts {
		opt(options)
	}

	query := tx.Run.Query()

	if options.Name != "" {
		query.Where(run.NameEQ(options.Name))
	}
	if options.Limit != nil {
		query.Limit(*options.Limit)
	}
	if options.Offset != nil {
		query.Offset(*options.Offset)
	}

	return query.All(ctx)
}

func (r *RunRepository) Delete(ctx context.Context, tx *ent.Tx, id int) error {
	return tx.Run.DeleteOneID(id).Exec(ctx)
}

// Hierarchy Management
func (r *RunRepository) CreateHierarchy(ctx context.Context, tx *ent.Tx, parentEntity, childEntity *ent.Entity) (*ent.Hierarchy, error) {
	// Create execution for parent if not exists
	parentExec, err := r.executionRepo.Create(tx, base.WithExecutionEntityID(parentEntity.ID))
	if err != nil {
		return nil, fmt.Errorf("creating parent execution: %w", err)
	}

	// Create execution for child
	childExec, err := r.executionRepo.Create(tx, base.WithExecutionEntityID(childEntity.ID))
	if err != nil {
		return nil, fmt.Errorf("creating child execution: %w", err)
	}

	// Create hierarchy relationship
	return tx.Hierarchy.Create().
		SetRunID(parentEntity.QueryRun().OnlyX(ctx).ID).
		SetParentEntityID(parentEntity.ID).
		SetChildEntityID(childEntity.ID).
		SetParentExecutionID(parentExec.ID).
		SetChildExecutionID(childExec.ID).
		SetParentType(hierarchy.ParentType(parentEntity.Type)).
		SetChildType(hierarchy.ChildType(childEntity.Type)).
		SetParentStepID(parentEntity.StepID).
		SetChildStepID(childEntity.StepID).
		Save(ctx)
}

func (r *RunRepository) GetHierarchies(ctx context.Context, tx *ent.Tx, runID int) ([]*ent.Hierarchy, error) {
	return tx.Hierarchy.Query().
		Where(hierarchy.RunID(runID)).
		All(ctx)
}

func (r *RunRepository) GetChildHierarchies(ctx context.Context, tx *ent.Tx, runID int, parentEntityID int) ([]*ent.Hierarchy, error) {
	return tx.Hierarchy.Query().
		Where(
			hierarchy.RunID(runID),
			hierarchy.ParentEntityID(parentEntityID),
		).
		All(ctx)
}

func (r *RunRepository) GetParentHierarchy(ctx context.Context, tx *ent.Tx, runID int, childEntityID int) (*ent.Hierarchy, error) {
	return tx.Hierarchy.Query().
		Where(
			hierarchy.RunID(runID),
			hierarchy.ChildEntityID(childEntityID),
		).
		Only(ctx)
}

// Analysis and Utilities
func (r *RunRepository) GetHierarchyDepth(ctx context.Context, tx *ent.Tx, runID int, entityID int) (int, error) {
	depth := 0
	currentEntityID := entityID

	for {
		hierarchy, err := r.GetParentHierarchy(ctx, tx, runID, currentEntityID)
		if err != nil {
			if ent.IsNotFound(err) {
				break
			}
			return 0, err
		}
		depth++
		currentEntityID = hierarchy.ParentEntityID
	}

	return depth, nil
}

func (r *RunRepository) GetRootEntities(ctx context.Context, tx *ent.Tx, runID int) ([]*ent.Entity, error) {
	// Get all child IDs from hierarchies
	childIDs, err := tx.Hierarchy.Query().
		Where(hierarchy.RunID(runID)).
		Select(hierarchy.FieldChildEntityID).
		Ints(ctx)
	if err != nil {
		return nil, err
	}

	// Get entities that aren't children (root entities)
	return tx.Entity.Query().
		Where(
			entity.HasRunWith(run.ID(runID)),
			entity.IDNotIn(childIDs...),
		).
		All(ctx)
}

func (r *RunRepository) BuildEntityTree(ctx context.Context, tx *ent.Tx, runID int, rootEntityID int) (map[string]interface{}, error) {
	tree := make(map[string]interface{})

	entity, err := tx.Entity.Get(ctx, rootEntityID)
	if err != nil {
		return nil, err
	}

	tree["id"] = entity.ID
	tree["type"] = entity.Type
	tree["step_id"] = entity.StepID
	tree["handler_name"] = entity.HandlerName

	// Get executions
	executions, err := r.executionRepo.List(tx, base.WithExecutionEntityID(entity.ID))
	if err != nil {
		return nil, err
	}
	tree["executions"] = len(executions)

	children, err := r.GetChildHierarchies(ctx, tx, runID, rootEntityID)
	if err != nil {
		return nil, err
	}

	if len(children) > 0 {
		childTrees := make([]map[string]interface{}, 0)
		for _, child := range children {
			childTree, err := r.BuildEntityTree(ctx, tx, runID, child.ChildEntityID)
			if err != nil {
				return nil, err
			}
			childTrees = append(childTrees, childTree)
		}
		tree["children"] = childTrees
	}

	return tree, nil
}
