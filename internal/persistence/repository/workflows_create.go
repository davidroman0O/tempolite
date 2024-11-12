package repository

import (
	"errors"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type CreateWorkflowInput struct {
	RunID       int
	HandlerName string
	StepID      string
	QueueID     int
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
	Duration    string
}

func (r *workflowRepository) Create(tx *ent.Tx, input CreateWorkflowInput) (*WorkflowInfo, error) {

	logs.Debug(r.ctx, "Creating workflow",
		"RunID", input.RunID,
		"HandlerName", input.HandlerName,
		"StepID", input.StepID,
		"QueueID", input.QueueID)

	runObj, err := tx.Run.Get(r.ctx, input.RunID)
	if err != nil {
		if ent.IsNotFound(err) {
			logs.Error(r.ctx, "Run not found", "error", err, "RunID", input.RunID, "HandlerName", input.HandlerName, "StepID", input.StepID, "QueueID", input.QueueID)
			return nil, errors.Join(err, fmt.Errorf("run %d not found", input.RunID))
		}
		logs.Error(r.ctx, "Error getting run", "error", err, "RunID", input.RunID, "HandlerName", input.HandlerName, "StepID", input.StepID, "QueueID", input.QueueID)
		return nil, errors.Join(err, fmt.Errorf("getting run"))
	}

	builder := tx.Entity.Create().
		SetHandlerName(input.HandlerName).
		SetType(entity.TypeWorkflow).
		SetStepID(input.StepID).
		SetStatus(entity.StatusPending).
		SetRun(runObj)

	if input.QueueID == 0 {
		logs.Error(r.ctx, "Queue ID is required", "RunID", input.RunID, "HandlerName", input.HandlerName, "StepID", input.StepID, "QueueID", input.QueueID)
		return nil, errors.Join(err, fmt.Errorf("queue ID is required"))
	}

	queueObj, err := tx.Queue.Get(r.ctx, input.QueueID)
	if err != nil {
		logs.Error(r.ctx, "Error getting queue", "error", err, "RunID", input.RunID, "HandlerName", input.HandlerName, "StepID", input.StepID, "QueueID", input.QueueID)
		return nil, errors.Join(err, fmt.Errorf("getting queue"))
	}

	logs.Debug(r.ctx, "Create workflow Queue found", "QueueID", queueObj.ID)
	builder.SetQueue(queueObj)

	entObj, err := builder.Save(r.ctx)
	if err != nil {
		logs.Error(r.ctx, "Error creating workflow entity", "error", err, "RunID", input.RunID, "HandlerName", input.HandlerName, "StepID", input.StepID, "QueueID", input.QueueID)
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

	workflowExecData, err := tx.WorkflowExecutionData.Create().
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
			WorkflowExecutionData: WorkflowExecutionData{
				ExecutionDataID: workflowExecData.ID,
				ErrMsg:          workflowExecData.Error,
				Output:          workflowExecData.Output,
			},
		},
	}, nil
}
