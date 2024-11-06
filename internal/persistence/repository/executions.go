package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecutiondata"
)

type ExecutionInfo struct {
	ID          int        `json:"id"`
	EntityID    int        `json:"entity_id"`
	Status      Status     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ExecutionRepository interface {
	Create(tx *ent.Tx, entityID int) (*ExecutionInfo, error)
	Get(tx *ent.Tx, id int) (*ExecutionInfo, error)
	List(tx *ent.Tx, entityID int) ([]*ExecutionInfo, error)
	ListByStatus(tx *ent.Tx, entityID int, status Status) ([]*ExecutionInfo, error)
	UpdateStatus(tx *ent.Tx, id int, status Status) (*ExecutionInfo, error)
	Complete(tx *ent.Tx, id int) (*ExecutionInfo, error)
	Fail(tx *ent.Tx, id int) (*ExecutionInfo, error)
	Delete(tx *ent.Tx, id int) error
}

type executionRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewExecutionRepository(ctx context.Context, client *ent.Client) ExecutionRepository {
	return &executionRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *executionRepository) Create(tx *ent.Tx, entityID int) (*ExecutionInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, entityID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("entity %d not found", entityID)
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	realStatus := execution.Status(StatusRunning)

	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(realStatus).
		SetStartedAt(time.Now()).
		Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating execution: %w", err)
	}

	return &ExecutionInfo{
		ID:          execObj.ID,
		EntityID:    entityID,
		Status:      Status(execObj.Status),
		StartedAt:   execObj.StartedAt,
		CompletedAt: execObj.CompletedAt,
		CreatedAt:   execObj.CreatedAt,
		UpdatedAt:   execObj.UpdatedAt,
	}, nil
}

func (r *executionRepository) Get(tx *ent.Tx, id int) (*ExecutionInfo, error) {
	execObj, err := tx.Execution.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	entityID, err := execObj.QueryEntity().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting entity ID: %w", err)
	}

	return &ExecutionInfo{
		ID:          execObj.ID,
		EntityID:    entityID,
		Status:      Status(execObj.Status),
		StartedAt:   execObj.StartedAt,
		CompletedAt: execObj.CompletedAt,
		CreatedAt:   execObj.CreatedAt,
		UpdatedAt:   execObj.UpdatedAt,
	}, nil
}

func (r *executionRepository) List(tx *ent.Tx, entityID int) ([]*ExecutionInfo, error) {
	execObjs, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(entityID))).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing executions: %w", err)
	}

	result := make([]*ExecutionInfo, len(execObjs))
	for i, execObj := range execObjs {
		result[i] = &ExecutionInfo{
			ID:          execObj.ID,
			EntityID:    entityID,
			Status:      Status(execObj.Status),
			StartedAt:   execObj.StartedAt,
			CompletedAt: execObj.CompletedAt,
			CreatedAt:   execObj.CreatedAt,
			UpdatedAt:   execObj.UpdatedAt,
		}
	}

	return result, nil
}

func (r *executionRepository) ListByStatus(tx *ent.Tx, entityID int, status Status) ([]*ExecutionInfo, error) {
	realStatus := execution.Status(status)

	execObjs, err := tx.Execution.Query().
		Where(
			execution.HasEntityWith(entity.IDEQ(entityID)),
			execution.StatusEQ(realStatus),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing executions by status: %w", err)
	}

	result := make([]*ExecutionInfo, len(execObjs))
	for i, execObj := range execObjs {
		result[i] = &ExecutionInfo{
			ID:          execObj.ID,
			EntityID:    entityID,
			Status:      Status(execObj.Status),
			StartedAt:   execObj.StartedAt,
			CompletedAt: execObj.CompletedAt,
			CreatedAt:   execObj.CreatedAt,
			UpdatedAt:   execObj.UpdatedAt,
		}
	}

	return result, nil
}

func (r *executionRepository) UpdateStatus(tx *ent.Tx, id int, status Status) (*ExecutionInfo, error) {
	currentExec, err := tx.Execution.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	if !isValidExecutionStatusTransition(Status(currentExec.Status), status) {
		return nil, fmt.Errorf("%w: invalid status transition from %s to %s",
			ErrInvalidState, currentExec.Status, status)
	}

	realStatus := execution.Status(status)

	execObj, err := tx.Execution.UpdateOne(currentExec).
		SetStatus(realStatus).
		Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating execution status: %w", err)
	}

	entityID, err := execObj.QueryEntity().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting entity ID: %w", err)
	}

	return &ExecutionInfo{
		ID:          execObj.ID,
		EntityID:    entityID,
		Status:      Status(execObj.Status),
		StartedAt:   execObj.StartedAt,
		CompletedAt: execObj.CompletedAt,
		CreatedAt:   execObj.CreatedAt,
		UpdatedAt:   execObj.UpdatedAt,
	}, nil
}

func (r *executionRepository) Complete(tx *ent.Tx, id int) (*ExecutionInfo, error) {
	now := time.Now()
	realStatus := execution.Status(StatusCompleted)

	execObj, err := tx.Execution.UpdateOneID(id).
		SetStatus(realStatus).
		SetCompletedAt(now).
		Save(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("completing execution: %w", err)
	}

	entityID, err := execObj.QueryEntity().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting entity ID: %w", err)
	}

	return &ExecutionInfo{
		ID:          execObj.ID,
		EntityID:    entityID,
		Status:      Status(execObj.Status),
		StartedAt:   execObj.StartedAt,
		CompletedAt: execObj.CompletedAt,
		CreatedAt:   execObj.CreatedAt,
		UpdatedAt:   execObj.UpdatedAt,
	}, nil
}

func (r *executionRepository) Fail(tx *ent.Tx, id int) (*ExecutionInfo, error) {
	now := time.Now()
	realStatus := execution.Status(StatusFailed)

	execObj, err := tx.Execution.UpdateOneID(id).
		SetStatus(realStatus).
		SetCompletedAt(now).
		Save(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failing execution: %w", err)
	}

	entityID, err := execObj.QueryEntity().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting entity ID: %w", err)
	}

	return &ExecutionInfo{
		ID:          execObj.ID,
		EntityID:    entityID,
		Status:      Status(execObj.Status),
		StartedAt:   execObj.StartedAt,
		CompletedAt: execObj.CompletedAt,
		CreatedAt:   execObj.CreatedAt,
		UpdatedAt:   execObj.UpdatedAt,
	}, nil
}

func (r *executionRepository) Delete(tx *ent.Tx, id int) error {
	execObj, err := tx.Execution.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("getting execution: %w", err)
	}

	if _, err = r.deleteExecutionData(tx, execObj); err != nil {
		return err
	}

	if err := tx.Execution.DeleteOne(execObj).Exec(r.ctx); err != nil {
		return fmt.Errorf("deleting execution: %w", err)
	}

	return nil
}

func (r *executionRepository) deleteExecutionData(tx *ent.Tx, execObj *ent.Execution) (int, error) {
	entObj, err := execObj.QueryEntity().Only(r.ctx)
	if err != nil {
		return -1, fmt.Errorf("getting entity: %w", err)
	}

	var id int
	switch ComponentType(entObj.Type) {
	case ComponentWorkflow:
		if id, err = r.deleteWorkflowExecutionData(tx, execObj.ID); err != nil {
			return -1, err
		}
	case ComponentActivity:
		if id, err = r.deleteActivityExecutionData(tx, execObj.ID); err != nil {
			return -1, err
		}
	case ComponentSaga:
		if id, err = r.deleteSagaExecutionData(tx, execObj.ID); err != nil {
			return -1, err
		}
	case ComponentSideEffect:
		if id, err = r.deleteSideEffectExecutionData(tx, execObj.ID); err != nil {
			return -1, err
		}
	}

	return id, nil
}

func (r *executionRepository) deleteWorkflowExecutionData(tx *ent.Tx, execID int) (int, error) {
	return tx.WorkflowExecutionData.Delete().
		Where(workflowexecutiondata.HasWorkflowExecutionWith(
			workflowexecution.HasExecutionWith(
				execution.IDEQ(execID)))).
		Exec(r.ctx)
}

func (r *executionRepository) deleteActivityExecutionData(tx *ent.Tx, execID int) (int, error) {
	return tx.ActivityExecutionData.Delete().
		Where(activityexecutiondata.HasActivityExecutionWith(
			activityexecution.HasExecutionWith(
				execution.IDEQ(execID)))).
		Exec(r.ctx)
}

func (r *executionRepository) deleteSagaExecutionData(tx *ent.Tx, execID int) (int, error) {
	return tx.SagaExecutionData.Delete().
		Where(sagaexecutiondata.HasSagaExecutionWith(
			sagaexecution.HasExecutionWith(
				execution.IDEQ(execID)))).
		Exec(r.ctx)
}

func (r *executionRepository) deleteSideEffectExecutionData(tx *ent.Tx, execID int) (int, error) {
	return tx.SideEffectExecutionData.Delete().
		Where(sideeffectexecutiondata.HasSideEffectExecutionWith(
			sideeffectexecution.HasExecutionWith(
				execution.IDEQ(execID)))).
		Exec(r.ctx)
}

func isValidExecutionStatusTransition(current, target Status) bool {
	switch current {
	case StatusRunning:
		return target == StatusCompleted || target == StatusFailed
	case StatusCompleted, StatusFailed:
		return false // Terminal states
	default:
		return false
	}
}
