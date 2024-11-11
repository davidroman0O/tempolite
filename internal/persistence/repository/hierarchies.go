package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
)

type HierarchyInfo struct {
	ID                int    `json:"id"`
	RunID             int    `json:"run_id"`
	ParentEntityID    int    `json:"parent_entity_id"`
	ChildEntityID     int    `json:"child_entity_id"`
	ParentStepID      string `json:"parent_step_id"`
	ChildStepID       string `json:"child_step_id"`
	ParentExecutionID int    `json:"parent_execution_id"`
	ChildExecutionID  int    `json:"child_execution_id"`
	ParentType        string `json:"parent_type"`
	ChildType         string `json:"child_type"`
}

// TODO: make a pre-check that RunID + ParentID + StepID has to be unique, we need to be super strict about it

type HierarchyRepository interface {
	Create(tx *ent.Tx, runID int, parentEntityID int, childEntityID int,
		parentStepID string, childStepID string, parentExecID int, childExecID int, childType hierarchy.ChildType, parentType hierarchy.ParentType) (*HierarchyInfo, error)
	Get(tx *ent.Tx, id int) (*HierarchyInfo, error)
	GetChildren(tx *ent.Tx, parentEntityID int) ([]*HierarchyInfo, error)
	GetParent(tx *ent.Tx, childEntityID int) (*HierarchyInfo, error)
	Delete(tx *ent.Tx, id int) error
	DeleteByChild(tx *ent.Tx, childEntityID int) error

	HasHierarchy(tx *ent.Tx, runID int, parentEntityID int, childStepID string) (bool, error)
	GetExisting(tx *ent.Tx, runID int, parentEntityID int, childStepID string) (*HierarchyInfo, error)
}

type hierarchyRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewHierarchyRepository(ctx context.Context, client *ent.Client) HierarchyRepository {
	return &hierarchyRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *hierarchyRepository) HasHierarchy(tx *ent.Tx, runID int, parentEntityID int, childStepID string) (bool, error) {
	exists, err := tx.Hierarchy.Query().
		Where(
			hierarchy.RunIDEQ(runID),
			hierarchy.ParentEntityIDEQ(parentEntityID),
			hierarchy.ChildStepIDEQ(childStepID),
		).
		Exist(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("checking hierarchy existence: %w", err)
	}
	return exists, nil
}

func (r *hierarchyRepository) GetExisting(tx *ent.Tx, runID int, parentEntityID int, childStepID string) (*HierarchyInfo, error) {
	hierarchyObj, err := tx.Hierarchy.Query().
		Where(
			hierarchy.RunIDEQ(runID),
			hierarchy.ParentEntityIDEQ(parentEntityID),
			hierarchy.ChildStepIDEQ(childStepID),
		).
		Only(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting hierarchy: %w", err)
	}

	return &HierarchyInfo{
		ID:                hierarchyObj.ID,
		RunID:             hierarchyObj.RunID,
		ParentEntityID:    hierarchyObj.ParentEntityID,
		ChildEntityID:     hierarchyObj.ChildEntityID,
		ParentStepID:      hierarchyObj.ParentStepID,
		ChildStepID:       hierarchyObj.ChildStepID,
		ParentExecutionID: hierarchyObj.ParentExecutionID,
		ChildExecutionID:  hierarchyObj.ChildExecutionID,
		ParentType:        hierarchyObj.ParentType.String(),
		ChildType:         hierarchyObj.ChildType.String(),
	}, nil
}

func (r *hierarchyRepository) Create(tx *ent.Tx, runID int, parentEntityID int, childEntityID int,
	parentStepID string, childStepID string, parentExecID int, childExecID int, childType hierarchy.ChildType, parentType hierarchy.ParentType) (*HierarchyInfo, error) {

	exists, err := tx.Run.Query().
		Where(run.IDEQ(runID)).
		Exist(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("checking run existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("run %d not found", runID)
	}

	exists, err = tx.Hierarchy.Query().
		Where(hierarchy.ChildEntityIDEQ(childEntityID)).
		Exist(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("checking existing hierarchy: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: child entity already has a parent", ErrInvalidOperation)
	}

	hierarchyObj, err := tx.Hierarchy.Create().
		SetRunID(runID).
		SetParentEntityID(parentEntityID).
		SetChildEntityID(childEntityID).
		SetParentStepID(parentStepID).
		SetChildStepID(childStepID).
		SetParentExecutionID(parentExecID).
		SetChildExecutionID(childExecID).
		Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating hierarchy: %w", err)
	}

	return &HierarchyInfo{
		ID:                hierarchyObj.ID,
		RunID:             hierarchyObj.RunID,
		ParentEntityID:    hierarchyObj.ParentEntityID,
		ChildEntityID:     hierarchyObj.ChildEntityID,
		ParentStepID:      hierarchyObj.ParentStepID,
		ChildStepID:       hierarchyObj.ChildStepID,
		ParentExecutionID: hierarchyObj.ParentExecutionID,
		ChildExecutionID:  hierarchyObj.ChildExecutionID,
		ParentType:        hierarchyObj.ParentType.String(),
		ChildType:         hierarchyObj.ChildType.String(),
	}, nil
}

func (r *hierarchyRepository) Get(tx *ent.Tx, id int) (*HierarchyInfo, error) {
	hierarchyObj, err := tx.Hierarchy.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting hierarchy: %w", err)
	}

	return &HierarchyInfo{
		ID:                hierarchyObj.ID,
		RunID:             hierarchyObj.RunID,
		ParentEntityID:    hierarchyObj.ParentEntityID,
		ChildEntityID:     hierarchyObj.ChildEntityID,
		ParentStepID:      hierarchyObj.ParentStepID,
		ChildStepID:       hierarchyObj.ChildStepID,
		ParentExecutionID: hierarchyObj.ParentExecutionID,
		ChildExecutionID:  hierarchyObj.ChildExecutionID,
		ParentType:        hierarchyObj.ParentType.String(),
		ChildType:         hierarchyObj.ChildType.String(),
	}, nil
}

func (r *hierarchyRepository) GetChildren(tx *ent.Tx, parentEntityID int) ([]*HierarchyInfo, error) {
	hierarchyObjs, err := tx.Hierarchy.Query().
		Where(hierarchy.ParentEntityIDEQ(parentEntityID)).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting children: %w", err)
	}

	result := make([]*HierarchyInfo, len(hierarchyObjs))
	for i, hierarchyObj := range hierarchyObjs {
		result[i] = &HierarchyInfo{
			ID:                hierarchyObj.ID,
			RunID:             hierarchyObj.RunID,
			ParentEntityID:    hierarchyObj.ParentEntityID,
			ChildEntityID:     hierarchyObj.ChildEntityID,
			ParentStepID:      hierarchyObj.ParentStepID,
			ChildStepID:       hierarchyObj.ChildStepID,
			ParentExecutionID: hierarchyObj.ParentExecutionID,
			ChildExecutionID:  hierarchyObj.ChildExecutionID,
			ParentType:        hierarchyObj.ParentType.String(),
			ChildType:         hierarchyObj.ChildType.String(),
		}
	}
	return result, nil
}

func (r *hierarchyRepository) GetParent(tx *ent.Tx, childEntityID int) (*HierarchyInfo, error) {
	hierarchyObj, err := tx.Hierarchy.Query().
		Where(hierarchy.ChildEntityIDEQ(childEntityID)).
		Only(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting parent: %w", err)
	}

	return &HierarchyInfo{
		ID:                hierarchyObj.ID,
		RunID:             hierarchyObj.RunID,
		ParentEntityID:    hierarchyObj.ParentEntityID,
		ChildEntityID:     hierarchyObj.ChildEntityID,
		ParentStepID:      hierarchyObj.ParentStepID,
		ChildStepID:       hierarchyObj.ChildStepID,
		ParentExecutionID: hierarchyObj.ParentExecutionID,
		ChildExecutionID:  hierarchyObj.ChildExecutionID,
		ParentType:        hierarchyObj.ParentType.String(),
		ChildType:         hierarchyObj.ChildType.String(),
	}, nil
}

func (r *hierarchyRepository) Delete(tx *ent.Tx, id int) error {
	exists, err := tx.Hierarchy.Query().
		Where(hierarchy.IDEQ(id)).
		Exist(r.ctx)
	if err != nil {
		return fmt.Errorf("checking hierarchy existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	err = tx.Hierarchy.DeleteOneID(id).Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting hierarchy: %w", err)
	}

	return nil
}

func (r *hierarchyRepository) DeleteByChild(tx *ent.Tx, childEntityID int) error {
	exists, err := tx.Hierarchy.Query().
		Where(hierarchy.ChildEntityIDEQ(childEntityID)).
		Exist(r.ctx)
	if err != nil {
		return fmt.Errorf("checking hierarchy existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.Hierarchy.Delete().
		Where(hierarchy.ChildEntityIDEQ(childEntityID)).
		Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting hierarchies for child: %w", err)
	}

	return nil
}
