package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activitydata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagadata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectdata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecutiondata"
)

type RunInfo struct {
	ID        int       `json:"id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RunRepository interface {
	Create(tx *ent.Tx) (*RunInfo, error)
	Get(tx *ent.Tx, id int) (*RunInfo, error)
	List(tx *ent.Tx) ([]*RunInfo, error)
	UpdateStatus(tx *ent.Tx, id int, status Status) (*RunInfo, error)
	Delete(tx *ent.Tx, id int) error
}

type runRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewRunRepository(ctx context.Context, client *ent.Client) RunRepository {
	return &runRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *runRepository) Create(tx *ent.Tx) (*RunInfo, error) {
	runObj, err := tx.Run.Create().
		SetStatus(run.StatusPending).
		Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}

	return &RunInfo{
		ID:        runObj.ID,
		Status:    Status(runObj.Status),
		CreatedAt: runObj.CreatedAt,
		UpdatedAt: runObj.UpdatedAt,
	}, nil
}

func (r *runRepository) Get(tx *ent.Tx, id int) (*RunInfo, error) {
	runObj, err := tx.Run.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting run: %w", err)
	}

	return &RunInfo{
		ID:        runObj.ID,
		Status:    Status(runObj.Status),
		CreatedAt: runObj.CreatedAt,
		UpdatedAt: runObj.UpdatedAt,
	}, nil
}

func (r *runRepository) List(tx *ent.Tx) ([]*RunInfo, error) {
	runObjs, err := tx.Run.Query().All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing runs: %w", err)
	}

	result := make([]*RunInfo, len(runObjs))
	for i, runObj := range runObjs {
		result[i] = &RunInfo{
			ID:        runObj.ID,
			Status:    Status(runObj.Status),
			CreatedAt: runObj.CreatedAt,
			UpdatedAt: runObj.UpdatedAt,
		}
	}
	return result, nil
}

func (r *runRepository) UpdateStatus(tx *ent.Tx, id int, status Status) (*RunInfo, error) {
	currentRun, err := tx.Run.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting run: %w", err)
	}

	if !isValidStatusTransition(Status(currentRun.Status), status) {
		return nil, fmt.Errorf("%w: invalid status transition from %s to %s",
			ErrInvalidState, currentRun.Status, status)
	}

	realStatus := run.Status(status)

	runObj, err := tx.Run.UpdateOneID(id).
		SetStatus(realStatus).
		Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating run status: %w", err)
	}

	return &RunInfo{
		ID:        runObj.ID,
		Status:    Status(runObj.Status),
		CreatedAt: runObj.CreatedAt,
		UpdatedAt: runObj.UpdatedAt,
	}, nil
}

func (r *runRepository) Delete(tx *ent.Tx, id int) error {
	exists, err := tx.Run.Query().
		Where(run.IDEQ(id)).
		Exist(r.ctx)
	if err != nil {
		return fmt.Errorf("checking run existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.Hierarchy.Delete().
		Where(hierarchy.HasRunWith(run.IDEQ(id))).
		Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting run hierarchies: %w", err)
	}

	entObjs, err := tx.Entity.Query().
		Where(entity.HasRunWith(run.IDEQ(id))).
		All(r.ctx)
	if err != nil {
		return fmt.Errorf("querying run entities: %w", err)
	}

	for _, entObj := range entObjs {
		switch entObj.Type {
		case "Workflow":
			_, err = r.deleteWorkflowData(tx, entObj.ID)
		case "Activity":
			_, err = r.deleteActivityData(tx, entObj.ID)
		case "Saga":
			_, err = r.deleteSagaData(tx, entObj.ID)
		case "SideEffect":
			_, err = r.deleteSideEffectData(tx, entObj.ID)
		}
		if err != nil {
			return fmt.Errorf("deleting entity data: %w", err)
		}

		_, err = r.deleteEntityExecutions(tx, entObj.ID)
		if err != nil {
			return fmt.Errorf("deleting entity executions: %w", err)
		}

		err = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		if err != nil {
			return fmt.Errorf("deleting entity: %w", err)
		}
	}

	err = tx.Run.DeleteOneID(id).Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting run: %w", err)
	}

	return nil
}

func (r *runRepository) deleteWorkflowData(tx *ent.Tx, entityID int) (int, error) {
	return tx.WorkflowData.Delete().
		Where(workflowdata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *runRepository) deleteActivityData(tx *ent.Tx, entityID int) (int, error) {
	return tx.ActivityData.Delete().
		Where(activitydata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *runRepository) deleteSagaData(tx *ent.Tx, entityID int) (int, error) {
	return tx.SagaData.Delete().
		Where(sagadata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *runRepository) deleteSideEffectData(tx *ent.Tx, entityID int) (int, error) {
	return tx.SideEffectData.Delete().
		Where(sideeffectdata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *runRepository) deleteEntityExecutions(tx *ent.Tx, entityID int) (int, error) {
	execObjs, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(entityID))).
		All(r.ctx)
	if err != nil {
		return -1, fmt.Errorf("querying executions: %w", err)
	}

	for _, execObj := range execObjs {
		if err := r.deleteExecution(tx, execObj); err != nil {
			return -1, err
		}
	}

	return len(execObjs), nil
}

func (r *runRepository) deleteExecution(tx *ent.Tx, execObj *ent.Execution) error {
	if _, err := tx.WorkflowExecutionData.Delete().
		Where(workflowexecutiondata.HasWorkflowExecutionWith(
			workflowexecution.HasExecutionWith(execution.IDEQ(execObj.ID)))).
		Exec(r.ctx); err != nil {
		return fmt.Errorf("deleting workflow execution data: %w", err)
	}

	if _, err := tx.ActivityExecutionData.Delete().
		Where(activityexecutiondata.HasActivityExecutionWith(
			activityexecution.HasExecutionWith(execution.IDEQ(execObj.ID)))).
		Exec(r.ctx); err != nil {
		return fmt.Errorf("deleting activity execution data: %w", err)
	}

	if _, err := tx.SagaExecutionData.Delete().
		Where(sagaexecutiondata.HasSagaExecutionWith(
			sagaexecution.HasExecutionWith(execution.IDEQ(execObj.ID)))).
		Exec(r.ctx); err != nil {
		return fmt.Errorf("deleting saga execution data: %w", err)
	}

	if _, err := tx.SideEffectExecutionData.Delete().
		Where(sideeffectexecutiondata.HasSideEffectExecutionWith(
			sideeffectexecution.HasExecutionWith(execution.IDEQ(execObj.ID)))).
		Exec(r.ctx); err != nil {
		return fmt.Errorf("deleting side effect execution data: %w", err)
	}

	err := tx.Execution.DeleteOne(execObj).Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting execution: %w", err)
	}
	return nil
}

func isValidStatusTransition(current, target Status) bool {
	switch current {
	case StatusRunning:
		return target == StatusCompleted || target == StatusFailed || target == StatusCancelled
	case StatusCompleted, StatusFailed, StatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}
