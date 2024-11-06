package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activitydata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
)

type ActivityInfo struct {
	EntityInfo
	Data      *ActivityDataInfo      `json:"data,omitempty"`
	Execution *ActivityExecutionInfo `json:"execution,omitempty"`
}

type ActivityDataInfo struct {
	ID          int                `json:"id"`
	RetryPolicy schema.RetryPolicy `json:"retry_policy"`
	Input       [][]byte           `json:"input,omitempty"`
	Output      [][]byte           `json:"output,omitempty"`
}

type ActivityExecutionInfo struct {
	ExecutionInfo
	Result          []byte     `json:"result,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	LastUpdatedTime time.Time  `json:"last_updated_time"`
}

type CreateActivityInput struct {
	RunID       int
	HandlerName string
	StepID      string
	QueueIDs    []int
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
}

type UpdateActivityDataInput struct {
	RetryPolicy *schema.RetryPolicy
	Input       [][]byte
	Output      [][]byte
}

type ActivityRepository interface {
	Create(tx *ent.Tx, input CreateActivityInput) (*ActivityInfo, error)
	Get(tx *ent.Tx, id int) (*ActivityInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*ActivityInfo, error)
	List(tx *ent.Tx, runID int) ([]*ActivityInfo, error)
	UpdateData(tx *ent.Tx, id int, input UpdateActivityDataInput) (*ActivityInfo, error)
	SetOutput(tx *ent.Tx, id int, output [][]byte) error
}

type activityRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewActivityRepository(ctx context.Context, client *ent.Client) ActivityRepository {
	return &activityRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *activityRepository) Create(tx *ent.Tx, input CreateActivityInput) (*ActivityInfo, error) {
	// Verify run exists
	runObj, err := tx.Run.Get(r.ctx, input.RunID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("run not found: %w", err)
		}
		return nil, fmt.Errorf("getting run: %w", err)
	}

	// Check if step ID already exists
	exists, err := tx.Entity.Query().
		Where(entity.StepIDEQ(input.StepID)).
		Exist(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("checking step ID existence: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	realType := entity.Type(ComponentActivity)

	// Create entity
	builder := tx.Entity.Create().
		SetHandlerName(input.HandlerName).
		SetType(realType).
		SetStepID(input.StepID).
		SetRun(runObj)

	// Add queues if specified
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

	// Save entity
	entObj, err := builder.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating entity: %w", err)
	}

	// Create activity data
	retryPolicy := defaultRetryPolicy()
	if input.RetryPolicy != nil {
		retryPolicy = *input.RetryPolicy
	}

	activityData, err := tx.ActivityData.Create().
		SetEntity(entObj).
		SetRetryPolicy(&retryPolicy).
		SetInput(input.Input).
		Save(r.ctx)
	if err != nil {
		// Cleanup entity on failure
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating activity data: %w", err)
	}

	realStatus := execution.Status(StatusRunning)

	// Create initial execution
	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(realStatus).
		SetStartedAt(time.Now()).
		Save(r.ctx)
	if err != nil {
		// Cleanup on failure
		_ = tx.ActivityData.DeleteOne(activityData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating execution: %w", err)
	}

	// Create activity execution
	activityExec, err := tx.ActivityExecution.Create().
		SetExecution(execObj).
		Save(r.ctx)
	if err != nil {
		// Cleanup on failure
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.ActivityData.DeleteOne(activityData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating activity execution: %w", err)
	}

	// Create activity execution data
	_, err = tx.ActivityExecutionData.Create().
		SetActivityExecution(activityExec).
		Save(r.ctx)
	if err != nil {
		// Cleanup on failure
		_ = tx.ActivityExecution.DeleteOne(activityExec).Exec(r.ctx)
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.ActivityData.DeleteOne(activityData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating activity execution data: %w", err)
	}

	// Get queue IDs for response
	assignedQueueIDs, err := entObj.QueryQueues().IDs(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue IDs: %w", err)
	}

	return &ActivityInfo{
		EntityInfo: EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runObj.ID,
			QueueIDs:    assignedQueueIDs,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
		},
		Data: &ActivityDataInfo{
			ID:          activityData.ID,
			RetryPolicy: *activityData.RetryPolicy,
			Input:       activityData.Input,
			Output:      activityData.Output,
		},
		Execution: &ActivityExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			Result:          nil,
			StartedAt:       execObj.StartedAt,
			CompletedAt:     execObj.CompletedAt,
			LastUpdatedTime: execObj.UpdatedAt,
		},
	}, nil
}

func (r *activityRepository) Get(tx *ent.Tx, id int) (*ActivityInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	runID, err := entObj.QueryRun().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	assignedQueueIDs, err := entObj.QueryQueues().IDs(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue IDs: %w", err)
	}

	activityData, err := tx.ActivityData.Query().
		Where(activitydata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting activity data: %w", err)
	}

	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(id))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	return &ActivityInfo{
		EntityInfo: EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runID,
			QueueIDs:    assignedQueueIDs,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
		},
		Data: &ActivityDataInfo{
			ID:          activityData.ID,
			RetryPolicy: *activityData.RetryPolicy,
			Input:       activityData.Input,
			Output:      activityData.Output,
		},
		Execution: &ActivityExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			Result:          nil,
			StartedAt:       execObj.StartedAt,
			CompletedAt:     execObj.CompletedAt,
			LastUpdatedTime: execObj.UpdatedAt,
		},
	}, nil
}

func (r *activityRepository) GetByStepID(tx *ent.Tx, stepID string) (*ActivityInfo, error) {
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

func (r *activityRepository) List(tx *ent.Tx, runID int) ([]*ActivityInfo, error) {
	entObjs, err := tx.Entity.Query().
		Where(
			entity.TypeEQ(entity.Type(ComponentActivity)),
			entity.HasRunWith(run.IDEQ(runID)),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing activities: %w", err)
	}

	result := make([]*ActivityInfo, 0, len(entObjs))
	for _, entObj := range entObjs {
		info, err := r.Get(tx, entObj.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

func (r *activityRepository) UpdateData(tx *ent.Tx, id int, input UpdateActivityDataInput) (*ActivityInfo, error) {
	dataUpdate := tx.ActivityData.Update().
		Where(activitydata.HasEntityWith(entity.IDEQ(id)))

	if input.RetryPolicy != nil {
		dataUpdate.SetRetryPolicy(input.RetryPolicy)
	}
	if input.Input != nil {
		dataUpdate.SetInput(input.Input)
	}
	if input.Output != nil {
		dataUpdate.SetOutput(input.Output)
	}

	_, err := dataUpdate.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating activity data: %w", err)
	}

	return r.Get(tx, id)
}

func (r *activityRepository) SetOutput(tx *ent.Tx, id int, output [][]byte) error {
	_, err := tx.ActivityData.Update().
		Where(activitydata.HasEntityWith(entity.IDEQ(id))).
		SetOutput(output).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("setting activity output: %w", err)
	}
	return nil
}

func defaultRetryPolicy() schema.RetryPolicy {
	return schema.RetryPolicy{
		MaxAttempts:        1,
		InitialInterval:    1000000000, // 1 second
		BackoffCoefficient: 2.0,
		MaxInterval:        300000000000, // 5 minutes
	}
}
