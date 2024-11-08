package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
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
	QueueID     int
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
	Duration    string
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

	CreateRetry(tx *ent.Tx, id int) (*WorkflowExecutionInfo, error)

	Get(tx *ent.Tx, id int) (*WorkflowInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*WorkflowInfo, error)
	List(tx *ent.Tx, runID int) ([]*WorkflowInfo, error)

	ListPending(tx *ent.Tx, queue string) ([]*WorkflowInfo, error)
	ListExecutionsPending(tx *ent.Tx, queue string) ([]*WorkflowInfo, error)

	UpdateExecutionPendingToRunning(tx *ent.Tx, id int) error
	UpdateExecutionDataError(tx *ent.Tx, id int, errormsg string) error
	UpdateExecutionDataOuput(tx *ent.Tx, id int, output [][]byte) error

	UpdateExecutionSuccess(tx *ent.Tx, id int) error
	UpdateExecutionFailed(tx *ent.Tx, id int) error
	UpdateExecutionPaused(tx *ent.Tx, id int) error
	UpdateExecutionRetried(tx *ent.Tx, id int) error

	IncrementRetryAttempt(tx *ent.Tx, id int) error
	GetRetryState(tx *ent.Tx, id int) (*schema.RetryState, error)
	GetRetryPolicy(tx *ent.Tx, id int) (*schema.RetryPolicy, error)

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
			return nil, errors.Join(err, fmt.Errorf("run %d not found", input.RunID))
		}
		return nil, errors.Join(err, fmt.Errorf("getting run"))
	}

	exists, err := tx.Entity.Query().
		Where(entity.StepIDEQ(input.StepID)).
		Exist(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("checking step ID existence"))
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	realType := entity.Type(ComponentWorkflow)

	builder := tx.Entity.Create().
		SetHandlerName(input.HandlerName).
		SetType(realType).
		SetStepID(input.StepID).
		SetStatus(entity.StatusPending).
		SetRun(runObj)

	if input.QueueID == 0 {
		return nil, errors.Join(err, fmt.Errorf("queue ID is required"))
	}

	queueObj, err := tx.Queue.Get(r.ctx, input.QueueID)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting queue"))
	}

	builder.SetQueue(queueObj)

	entObj, err := builder.Save(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating workflow entity"))
	}

	retryPolicy := defaultRetryPolicy()
	if input.RetryPolicy != nil {
		retryPolicy = *input.RetryPolicy
	}

	builderWorkflow := tx.WorkflowData.Create().
		SetEntity(entObj).
		SetRetryPolicy(&retryPolicy).
		SetInput(input.Input)

	if input.Duration != "" {
		builderWorkflow.SetDuration(input.Duration)
	}

	workflowData, err := builderWorkflow.
		Save(r.ctx)
	if err != nil {
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, errors.Join(err, fmt.Errorf("creating workflow data"))
	}

	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(execution.StatusPending).
		Save(r.ctx)
	if err != nil {
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, errors.Join(err, fmt.Errorf("creating workflow execution"))
	}

	workflowExec, err := tx.WorkflowExecution.Create().
		SetExecution(execObj).
		Save(r.ctx)
	if err != nil {
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, errors.Join(err, fmt.Errorf("creating workflow execution"))
	}

	_, err = tx.WorkflowExecutionData.Create().
		SetWorkflowExecution(workflowExec).
		Save(r.ctx)
	if err != nil {
		_ = tx.WorkflowExecution.DeleteOne(workflowExec).Exec(r.ctx)
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.WorkflowData.DeleteOne(workflowData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, errors.Join(err, fmt.Errorf("creating workflow execution data"))
	}

	queueID, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting queue ID"))
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
			QueueID:     queueID,
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

func (r *workflowRepository) GetFromExecution(tx *ent.Tx, executionID int) (*WorkflowInfo, error) {
	execObj, err := tx.Execution.Get(r.ctx, executionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(err, fmt.Errorf("getting execution"))
	}

	entityID, err := execObj.QueryEntity().OnlyID(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting entity ID"))
	}

	entityObj, err := tx.Entity.Get(r.ctx, entityID)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting entity"))
	}

	workflowExecData, err := execObj.QueryWorkflowExecution().QueryExecutionData().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow execution data"))
	}

	workflowData, err := entityObj.QueryWorkflowData().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	queueObj, err := entityObj.QueryQueue().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting queue"))
	}

	runID := entityObj.QueryRun().OnlyX(r.ctx)

	return &WorkflowInfo{
		EntityInfo: EntityInfo{
			ID:          entityObj.ID,
			HandlerName: entityObj.HandlerName,
			Type:        ComponentType(entityObj.Type),
			StepID:      entityObj.StepID,
			RunID:       runID.ID,
			CreatedAt:   entityObj.CreatedAt,
			UpdatedAt:   entityObj.UpdatedAt,
			QueueID:     queueObj.ID,
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
				EntityID:    entityObj.ID,
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

func (r *workflowRepository) Get(tx *ent.Tx, id int) (*WorkflowInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(err, fmt.Errorf("getting entity"))
	}

	if entObj.Type != entity.Type(ComponentWorkflow) {
		return nil, ErrInvalidOperation
	}

	runID, err := entObj.QueryRun().OnlyID(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting run ID"))
	}

	assignedQueueIDs, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting queue IDs"))
	}

	workflowData, err := entObj.QueryWorkflowData().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(entObj.ID))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(err, fmt.Errorf("getting execution"))
	}

	workflowExec, err := execObj.QueryWorkflowExecution().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow execution"))
	}

	workflowExecData, err := workflowExec.QueryExecutionData().Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow execution data"))
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
			QueueID:     assignedQueueIDs,
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

// Create a new execution for the entity
func (r *workflowRepository) CreateRetry(tx *ent.Tx, id int) (*WorkflowExecutionInfo, error) {

	entityObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.Join(err, ErrNotFound, fmt.Errorf("getting entity:%v", id))
		}
		return nil, errors.Join(err, fmt.Errorf("getting entity: %v", id))
	}

	execObj, err := tx.Execution.Create().
		SetEntity(entityObj).
		SetStatus(execution.StatusPending).
		Save(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating execution"))
	}

	workflowExec, err := tx.WorkflowExecution.Create().
		SetExecution(execObj).
		Save(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating workflow execution"))
	}

	_, err = tx.WorkflowExecutionData.Create().
		SetWorkflowExecution(workflowExec).
		Save(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating workflow execution data"))
	}

	return &WorkflowExecutionInfo{
		ExecutionInfo: ExecutionInfo{
			ID:          execObj.ID,
			EntityID:    id,
			Status:      Status(execObj.Status),
			StartedAt:   execObj.StartedAt,
			CreatedAt:   execObj.CreatedAt,
			UpdatedAt:   execObj.UpdatedAt,
			CompletedAt: execObj.CompletedAt,
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
		return nil, errors.Join(err, fmt.Errorf("getting entity by step ID"))
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
		return nil, errors.Join(err, fmt.Errorf("querying entities"))
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
		return nil, errors.Join(err, fmt.Errorf("updating workflow data"))
	}

	return r.Get(tx, id)
}

func (r *workflowRepository) Pause(tx *ent.Tx, id int) error {
	_, err := tx.WorkflowData.Update().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		SetPaused(true).
		Save(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("pausing workflow"))
	}
	return nil
}

func (r *workflowRepository) Resume(tx *ent.Tx, id int) error {
	workflowData, err := tx.WorkflowData.Query().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	if !workflowData.Resumable {
		return errors.Join(err, fmt.Errorf("workflow is not resumable"), ErrInvalidOperation)
	}

	_, err = tx.WorkflowData.Update().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		SetPaused(false).
		Save(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("resuming workflow"))
	}

	return nil
}

func (r *workflowRepository) ListPending(tx *ent.Tx, queueName string) ([]*WorkflowInfo, error) {
	return nil, nil
}

func (r *workflowRepository) ListExecutionsPending(tx *ent.Tx, queueName string) ([]*WorkflowInfo, error) {
	// First get all pending executions for workflows in the specified queue
	execObjs, err := tx.Execution.Query().
		Where(
			execution.StatusEQ(execution.StatusPending), // Using StatusPending from the Status type
			execution.HasEntityWith(
				entity.And(
					entity.TypeEQ(entity.Type(ComponentWorkflow)),
					entity.HasQueueWith(queue.NameEQ(queueName)),
				),
			),
		).
		Order(ent.Asc(execution.FieldCreatedAt)).
		All(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("querying pending executions"))
	}

	// For each execution, get the full workflow info
	result := make([]*WorkflowInfo, 0, len(execObjs))
	for _, execObj := range execObjs {

		// Get the full workflow info using the existing Get method
		workflowInfo, err := r.GetFromExecution(tx, execObj.ID)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("getting workflow info for entity %d", execObj.ID))
		}

		result = append(result, workflowInfo)
	}

	return result, nil
}

func (r *workflowRepository) UpdateExecutionDataOuput(tx *ent.Tx, executionID int, output [][]byte) error {
	workflowExec, err := tx.WorkflowExecution.Query().
		Where(workflowexecution.HasExecutionWith(execution.IDEQ(executionID))).
		WithExecutionData().
		Only(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting workflow execution"))
	}

	_, err = tx.WorkflowExecutionData.UpdateOneID(workflowExec.Edges.ExecutionData.ID).
		SetOutput(output).
		Save(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("updating workflow execution data"))
	}

	return nil
}

func (r *workflowRepository) UpdateExecutionDataError(tx *ent.Tx, executionID int, errormsg string) error {
	workflowExec, err := tx.WorkflowExecution.Query().
		Where(workflowexecution.HasExecutionWith(execution.IDEQ(executionID))).
		WithExecutionData().
		Only(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting workflow execution"))
	}

	_, err = tx.WorkflowExecutionData.UpdateOneID(workflowExec.Edges.ExecutionData.ID).
		SetError(errormsg).
		Save(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("updating workflow execution data"))
	}

	return nil
}

func (r *workflowRepository) UpdateExecutionPendingToRunning(tx *ent.Tx, executionID int) error {

	if err := tx.Execution.UpdateOneID(executionID).SetStatus(execution.StatusRunning).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating execution status"))
	}

	execObj, err := tx.Execution.Get(r.ctx, executionID)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting execution"))
	}

	entityID, err := execObj.QueryEntity().OnlyID(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting entity ID"))
	}

	if err := tx.Entity.UpdateOneID(entityID).SetStatus(entity.StatusRunning).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating entity status"))
	}

	return nil
}

func (r *workflowRepository) UpdateExecutionPaused(tx *ent.Tx, executionID int) error {

	if err := tx.Execution.UpdateOneID(executionID).SetStatus(execution.StatusPaused).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating execution status"))
	}

	execObj, err := tx.Execution.Get(r.ctx, executionID)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting execution"))
	}

	entityID, err := execObj.QueryEntity().OnlyID(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting entity ID"))
	}

	if err := tx.WorkflowData.UpdateOneID(entityID).SetPaused(true).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating workflow data"))
	}

	return nil
}

// When successful then everything is updated
func (r *workflowRepository) UpdateExecutionSuccess(tx *ent.Tx, executionID int) error {
	if err := tx.Execution.UpdateOneID(executionID).SetStatus(execution.StatusCompleted).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating execution status"))
	}

	execObj, err := tx.Execution.Get(r.ctx, executionID)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting execution"))
	}

	entityID, err := execObj.QueryEntity().OnlyID(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting entity ID"))
	}

	if err := tx.Entity.UpdateOneID(entityID).SetStatus(entity.StatusCompleted).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating entity status"))
	}

	return nil
}

// When retrying only the execution status is updated
func (r *workflowRepository) UpdateExecutionRetried(tx *ent.Tx, executionID int) error {

	fmt.Println("set Retried status to", executionID)
	if err := tx.Execution.UpdateOneID(executionID).SetStatus(execution.StatusRetried).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating execution status"))
	}

	return nil
}

// When a workflow fails, the entity and execution status are updated
func (r *workflowRepository) UpdateExecutionFailed(tx *ent.Tx, executionID int) error {

	if err := tx.Execution.UpdateOneID(executionID).SetStatus(execution.StatusFailed).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating execution status"))
	}

	execObj, err := tx.Execution.Get(r.ctx, executionID)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting execution"))
	}

	entityID, err := execObj.QueryEntity().OnlyID(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting entity ID"))
	}

	if err := tx.Entity.UpdateOneID(entityID).SetStatus(entity.StatusFailed).Exec(r.ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return errors.Join(err, fmt.Errorf("updating entity status"))
	}

	return nil
}

func (r *workflowRepository) IncrementRetryAttempt(tx *ent.Tx, id int) error {
	workflowData, err := tx.WorkflowData.Query().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	workflowData.RetryState.Attempts++

	_, err = tx.WorkflowData.UpdateOneID(workflowData.ID).
		SetRetryState(workflowData.RetryState).
		Save(r.ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("incrementing retry attempt"))
	}

	return nil
}

func (r *workflowRepository) GetRetryState(tx *ent.Tx, id int) (*schema.RetryState, error) {
	workflowData, err := tx.WorkflowData.Query().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	return workflowData.RetryState, nil
}

func (r *workflowRepository) GetRetryPolicy(tx *ent.Tx, id int) (*schema.RetryPolicy, error) {
	workflowData, err := tx.WorkflowData.Query().
		Where(workflowdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("getting workflow data"))
	}

	return workflowData.RetryPolicy, nil
}
