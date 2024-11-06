package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
)

type WorkflowInfo struct {
	EntityInfo
	Data      *WorkflowDataInfo      `json:"data,omitempty"`
	Execution *WorkflowExecutionInfo `json:"execution,omitempty"`
}

type WorkflowDataInfo struct {
	ID          int                `json:"id"`
	Paused      bool               `json:"paused"`
	Resumable   bool               `json:"resumable"`
	RetryPolicy schema.RetryPolicy `json:"retry_policy"`
	Input       [][]byte           `json:"input,omitempty"`
}

type WorkflowExecutionInfo struct {
	ExecutionInfo
	Outputs [][]byte `json:"outputs,omitempty"`
}

type CreateWorkflowInput struct {
	RunID       int
	HandlerName string
	StepID      string
	QueueIDs    []int
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
}

type UpdateWorkflowDataInput struct {
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
	Output      [][]byte
	Paused      *bool
	Resumable   *bool
}

type WorkflowRepository interface {
	Create(tx *ent.Tx, input CreateWorkflowInput) (*WorkflowInfo, error)
	Get(tx *ent.Tx, id int) (*WorkflowInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*WorkflowInfo, error)
	List(tx *ent.Tx, runID int) ([]*WorkflowInfo, error)
	UpdateData(tx *ent.Tx, id int, input UpdateWorkflowDataInput) (*WorkflowInfo, error)
	Pause(tx *ent.Tx, id int) error
	Resume(tx *ent.Tx, id int) error
}

type workflowRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewWorkflowRepository(ctx context.Context, client *ent.Client) WorkflowRepository {
	return &workflowRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *workflowRepository) Create(tx *ent.Tx, input CreateWorkflowInput) (*WorkflowInfo, error) {
	runObj, err := tx.Run.Get(r.ctx, input.RunID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("run %d not found", input.RunID)
		}
		return nil, fmt.Errorf("getting run: %w", err)
	}

	exists, err := tx.Entity.Query().
		Where(entity.StepIDEQ(input.StepID)).
		Exist(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("checking step ID existence: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	realType := entity.Type(ComponentWorkflow)

	builder := tx.Entity.Create().
		SetHandlerName(input.HandlerName).
		SetType(realType).
		SetStepID(input.StepID).
		SetRun(runObj)

	if len(input.QueueIDs) > 0 {
		queueObjs, err := tx.Queue.Query().
			Where(queue.IDIn(input.QueueIDs...)).
			All(r.ctx)
		if err != nil {
			return nil, fmt.Errorf("getting queues: %w", err)
		}
		if len(queueObjs) != len(input.QueueIDs) {
			return nil, fmt.Errorf("some queues not found")
		}
		builder.AddQueues(queueObjs...)
	}

	entObj, err := builder.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating workflow entity: %w", err)
	}

	retryPolicy := defaultRetryPolicy()
	if input.RetryPolicy != nil {
		retryPolicy = *input.RetryPolicy
	}

	workflowData, err := tx.WorkflowData.Create().
		SetEntity(entObj).
		SetRetryPolicy(&retryPolicy).
		SetInput(input.Input).
		Save(r.ctx)
	if err != nil {
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating workflow data: %w", err)
	}

	realStatus := execution.Status(StatusRunning)

	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(realStatus).
		Save(r.ctx)
	if err != nil {
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating workflow execution: %w", err)
	}

	workflowExec, err := tx.WorkflowExecution.Create().
		SetExecution(execObj).
		Save(r.ctx)
	if err != nil {
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating workflow execution: %w", err)
	}

	_, err = tx.WorkflowExecutionData.Create().
		SetWorkflowExecution(workflowExec).
		Save(r.ctx)
	if err != nil {
		_ = tx.WorkflowExecution.DeleteOne(workflowExec).Exec(r.ctx)
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating workflow execution data: %w", err)
	}

	assignedQueueIDs, err := entObj.QueryQueues().IDs(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue IDs: %w", err)
	}

	return &WorkflowInfo{
		EntityInfo: EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       input.RunID,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
			QueueIDs:    assignedQueueIDs,
		},
		Data: &WorkflowDataInfo{
			ID:          workflowData.ID,
			Paused:      workflowData.Paused,
			Resumable:   workflowData.Resumable,
			RetryPolicy: retryPolicy,
			Input:       workflowData.Input,
		},
		Execution: &WorkflowExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
		},
	}, nil
}

func (r *workflowRepository) Get(tx *ent.Tx, id int) (*WorkflowInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	if entObj.Type != entity.Type(ComponentWorkflow) {
		return nil, ErrInvalidOperation
	}

	runID, err := entObj.QueryRun().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	assignedQueueIDs, err := entObj.QueryQueues().IDs(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue IDs: %w", err)
	}

	workflowData, err := entObj.QueryWorkflowData().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting workflow data: %w", err)
	}

	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(entObj.ID))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	workflowExec, err := execObj.QueryWorkflowExecution().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting workflow execution: %w", err)
	}

	workflowExecData, err := workflowExec.QueryExecutionData().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting workflow execution data: %w", err)
	}

	return &WorkflowInfo{
		EntityInfo: EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runID,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
			QueueIDs:    assignedQueueIDs,
		},
		Data: &WorkflowDataInfo{
			ID:          workflowData.ID,
			Paused:      workflowData.Paused,
			Resumable:   workflowData.Resumable,
			RetryPolicy: *workflowData.RetryPolicy,
			Input:       workflowData.Input,
		},
		Execution: &WorkflowExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			Outputs: workflowExecData.Output,
		},
	}, nil
}

func (r *workflowRepository) GetByStepID(tx *ent.Tx, stepID string) (*WorkflowInfo, error) {
	entObj, err := tx.Entity.Query().
		Where(entity.StepIDEQ(stepID)).
		Only(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity by step ID: %w", err)
	}

	return r.Get(tx, entObj.ID)
}

func (r *workflowRepository) List(tx *ent.Tx, runID int) ([]*WorkflowInfo, error) {
	entObjs, err := tx.Entity.Query().
		Where(
			entity.TypeEQ(entity.Type(ComponentWorkflow)),
			entity.HasRunWith(run.IDEQ(runID)),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("querying entities: %w", err)
	}

	result := make([]*WorkflowInfo, 0, len(entObjs))
	for _, entObj := range entObjs {
		info, err := r.Get(tx, entObj.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

func (r *workflowRepository) UpdateData(tx *ent.Tx, id int, input UpdateWorkflowDataInput) (*WorkflowInfo, error) {
	dataUpdate := tx.WorkflowData.Update().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id)))

	if input.RetryPolicy != nil {
		dataUpdate.SetRetryPolicy(input.RetryPolicy)
	}
	if input.Input != nil {
		dataUpdate.SetInput(input.Input)
	}
	if input.Paused != nil {
		dataUpdate.SetPaused(*input.Paused)
	}
	if input.Resumable != nil {
		dataUpdate.SetResumable(*input.Resumable)
	}

	_, err := dataUpdate.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating workflow data: %w", err)
	}

	return r.Get(tx, id)
}

func (r *workflowRepository) Pause(tx *ent.Tx, id int) error {
	_, err := tx.WorkflowData.Update().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		SetPaused(true).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("pausing workflow: %w", err)
	}
	return nil
}

func (r *workflowRepository) Resume(tx *ent.Tx, id int) error {
	workflowData, err := tx.WorkflowData.Query().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return fmt.Errorf("getting workflow data: %w", err)
	}

	if !workflowData.Resumable {
		return fmt.Errorf("%w: workflow is not resumable", ErrInvalidOperation)
	}

	_, err = tx.WorkflowData.Update().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		SetPaused(false).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("resuming workflow: %w", err)
	}

	return nil
}
