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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
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

type ComponentType string

const (
	ComponentWorkflow   ComponentType = "Workflow"
	ComponentActivity   ComponentType = "Activity"
	ComponentSaga       ComponentType = "Saga"
	ComponentSideEffect ComponentType = "SideEffect"
)

type EntityInfo struct {
	ID          int           `json:"id"`
	HandlerName string        `json:"handler_name"`
	Type        ComponentType `json:"type"`
	Status      string        `json:"status"`
	StepID      string        `json:"step_id"`
	RunID       int           `json:"run_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	QueueID     int           `json:"queue_id"`
	QueueName   string        `json:"queue_name"`
}

type EntityRepository interface {
	Create(tx *ent.Tx, runID int, handlerName string, componentType ComponentType,
		stepID string, queueID int) (*EntityInfo, error)
	Get(tx *ent.Tx, id int) (*EntityInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*EntityInfo, error)
	List(tx *ent.Tx, runID int) ([]*EntityInfo, error)
	ListByType(tx *ent.Tx, runID int, componentType ComponentType) ([]*EntityInfo, error)
	AddToQueue(tx *ent.Tx, entityID int, queueID int) error
	RemoveFromQueue(tx *ent.Tx, entityID int, queueID int) error
	Delete(tx *ent.Tx, id int) error
}

type entityRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewEntityRepository(ctx context.Context, client *ent.Client) EntityRepository {
	return &entityRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *entityRepository) Create(tx *ent.Tx, runID int, handlerName string,
	componentType ComponentType, stepID string, queueID int) (*EntityInfo, error) {

	runObj, err := tx.Run.Get(r.ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("run %d not found", runID)
		}
		return nil, fmt.Errorf("getting run: %w", err)
	}

	exists, err := tx.Entity.Query().
		Where(entity.StepIDEQ(stepID)).
		Exist(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("checking step ID existence: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	realType := entity.Type(componentType)

	builder := tx.Entity.Create().
		SetHandlerName(handlerName).
		SetType(realType).
		SetStepID(stepID).
		SetRun(runObj)

	if queueID == 0 {
		return nil, fmt.Errorf("queue ID is required")
	}

	queueObj, err := tx.Queue.Get(r.ctx, queueID)
	if err != nil {
		return nil, fmt.Errorf("getting queue: %w", err)
	}

	builder.SetQueue(queueObj)

	entObj, err := builder.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating entity: %w", err)
	}

	assignedQueueIDs, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue IDs: %w", err)
	}

	return &EntityInfo{
		ID:          entObj.ID,
		HandlerName: entObj.HandlerName,
		Type:        ComponentType(entObj.Type),
		StepID:      entObj.StepID,
		RunID:       runID,
		CreatedAt:   entObj.CreatedAt,
		UpdatedAt:   entObj.UpdatedAt,
		QueueID:     assignedQueueIDs,
		QueueName:   queueObj.Name,
	}, nil
}

func (r *entityRepository) Get(tx *ent.Tx, id int) (*EntityInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	runID, err := entObj.QueryRun().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	queue, err := entObj.QueryQueue().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	return &EntityInfo{
		ID:          entObj.ID,
		HandlerName: entObj.HandlerName,
		Type:        ComponentType(entObj.Type),
		StepID:      entObj.StepID,
		RunID:       runID,
		Status:      string(entObj.Status),
		CreatedAt:   entObj.CreatedAt,
		UpdatedAt:   entObj.UpdatedAt,
		QueueID:     queue.ID,
		QueueName:   queue.Name,
	}, nil
}

func (r *entityRepository) GetByStepID(tx *ent.Tx, stepID string) (*EntityInfo, error) {
	entObj, err := tx.Entity.Query().
		Where(entity.StepIDEQ(stepID)).
		Only(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity by step ID: %w", err)
	}

	runID, err := entObj.QueryRun().FirstID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	queue, err := entObj.QueryQueue().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	return &EntityInfo{
		ID:          entObj.ID,
		HandlerName: entObj.HandlerName,
		Type:        ComponentType(entObj.Type),
		StepID:      entObj.StepID,
		RunID:       runID,
		CreatedAt:   entObj.CreatedAt,
		UpdatedAt:   entObj.UpdatedAt,
		QueueID:     queue.ID,
		QueueName:   queue.Name,
	}, nil
}

func (r *entityRepository) List(tx *ent.Tx, runID int) ([]*EntityInfo, error) {
	entObjs, err := tx.Entity.Query().
		Where(entity.HasRunWith(run.IDEQ(runID))).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing entities: %w", err)
	}

	result := make([]*EntityInfo, len(entObjs))
	for i, entObj := range entObjs {

		queue, err := entObj.QueryQueue().Only(r.ctx)
		if err != nil {
			return nil, fmt.Errorf("getting queue ID: %w", err)
		}

		result[i] = &EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runID,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
			QueueID:     queue.ID,
			QueueName:   queue.Name,
		}
	}

	return result, nil
}

func (r *entityRepository) ListByType(tx *ent.Tx, runID int, componentType ComponentType) ([]*EntityInfo, error) {
	realType := entity.Type(componentType)

	entObjs, err := tx.Entity.Query().
		Where(
			entity.HasRunWith(run.IDEQ(runID)),
			entity.TypeEQ(realType),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing entities by type: %w", err)
	}

	result := make([]*EntityInfo, len(entObjs))
	for i, entObj := range entObjs {

		queue, err := entObj.QueryQueue().Only(r.ctx)
		if err != nil {
			return nil, fmt.Errorf("getting queue ID: %w", err)
		}

		result[i] = &EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runID,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
			QueueID:     queue.ID,
			QueueName:   queue.Name,
		}
	}

	return result, nil
}

func (r *entityRepository) AddToQueue(tx *ent.Tx, entityID int, queueID int) error {
	entObj, err := tx.Entity.Get(r.ctx, entityID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("getting entity: %w", err)
	}

	queueObj, err := tx.Queue.Get(r.ctx, queueID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("queue %d not found", queueID)
		}
		return fmt.Errorf("getting queue: %w", err)
	}

	exists, err := entObj.QueryQueue().
		Where(queue.IDEQ(queueID)).
		Exist(r.ctx)
	if err != nil {
		return fmt.Errorf("checking queue membership: %w", err)
	}
	if exists {
		return nil
	}

	err = entObj.Update().
		SetQueue(queueObj).
		Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("adding to queue: %w", err)
	}

	return nil
}

func (r *entityRepository) RemoveFromQueue(tx *ent.Tx, entityID int, queueID int) error {
	entObj, err := tx.Entity.Get(r.ctx, entityID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("getting entity: %w", err)
	}

	exists, err := entObj.QueryQueue().
		Where(queue.IDEQ(queueID)).
		Exist(r.ctx)
	if err != nil {
		return fmt.Errorf("checking queue membership: %w", err)
	}
	if !exists {
		return nil
	}

	err = entObj.Update().
		SetNillableQueueID(nil).
		Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("removing from queue: %w", err)
	}

	return nil
}

func (r *entityRepository) Delete(tx *ent.Tx, id int) error {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("getting entity: %w", err)
	}

	switch ComponentType(entObj.Type) {
	case ComponentWorkflow:
		_, err = r.deleteWorkflowData(tx, id)
	case ComponentActivity:
		_, err = r.deleteActivityData(tx, id)
	case ComponentSaga:
		_, err = r.deleteSagaData(tx, id)
	case ComponentSideEffect:
		_, err = r.deleteSideEffectData(tx, id)
	}
	if err != nil {
		return err
	}

	_, err = r.deleteEntityExecutions(tx, id)
	if err != nil {
		return err
	}

	_, err = tx.Hierarchy.Delete().
		Where(
			hierarchy.Or(
				hierarchy.ParentEntityIDEQ(id),
				hierarchy.ChildEntityIDEQ(id),
			),
		).
		Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting hierarchies: %w", err)
	}

	err = tx.Entity.DeleteOneID(id).Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("deleting entity: %w", err)
	}

	return nil
}

func (r *entityRepository) deleteWorkflowData(tx *ent.Tx, entityID int) (int, error) {
	return tx.WorkflowData.Delete().
		Where(workflowdata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *entityRepository) deleteActivityData(tx *ent.Tx, entityID int) (int, error) {
	return tx.ActivityData.Delete().
		Where(activitydata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *entityRepository) deleteSagaData(tx *ent.Tx, entityID int) (int, error) {
	return tx.SagaData.Delete().
		Where(sagadata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *entityRepository) deleteSideEffectData(tx *ent.Tx, entityID int) (int, error) {
	return tx.SideEffectData.Delete().
		Where(sideeffectdata.HasEntityWith(entity.IDEQ(entityID))).
		Exec(r.ctx)
}

func (r *entityRepository) deleteEntityExecutions(tx *ent.Tx, entityID int) (int, error) {
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

func (r *entityRepository) deleteExecution(tx *ent.Tx, execObj *ent.Execution) error {
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
