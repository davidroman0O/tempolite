package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectdata"
)

type SideEffectInfo struct {
	EntityInfo
	Data      *SideEffectDataInfo      `json:"data,omitempty"`
	Execution *SideEffectExecutionInfo `json:"execution,omitempty"`
}

type SideEffectDataInfo struct {
	ID     int      `json:"id"`
	Input  [][]byte `json:"input,omitempty"`
	Output [][]byte `json:"output,omitempty"`
}

type SideEffectExecutionInfo struct {
	ExecutionInfo
	Result         []byte    `json:"result,omitempty"`
	EffectTime     time.Time `json:"effect_time"`
	EffectMetadata []byte    `json:"effect_metadata,omitempty"`
}

type CreateSideEffectInput struct {
	RunID       int
	HandlerName string
	StepID      string
	QueueID     int
	Input       [][]byte
	Metadata    []byte
}

type UpdateSideEffectDataInput struct {
	Input  [][]byte
	Output [][]byte
}

type SideEffectRepository interface {
	Create(tx *ent.Tx, input CreateSideEffectInput) (*SideEffectInfo, error)
	Get(tx *ent.Tx, id int) (*SideEffectInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*SideEffectInfo, error)
	List(tx *ent.Tx, runID int) ([]*SideEffectInfo, error)
	UpdateData(tx *ent.Tx, id int, input UpdateSideEffectDataInput) (*SideEffectInfo, error)
	SetOutput(tx *ent.Tx, id int, output [][]byte) error
	SetResult(tx *ent.Tx, id int, result []byte) error
}

type sideEffectRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewSideEffectRepository(ctx context.Context, client *ent.Client) SideEffectRepository {
	return &sideEffectRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *sideEffectRepository) Create(tx *ent.Tx, input CreateSideEffectInput) (*SideEffectInfo, error) {
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

	realType := entity.Type(ComponentSideEffect)

	builder := tx.Entity.Create().
		SetHandlerName(input.HandlerName).
		SetType(realType).
		SetStepID(input.StepID).
		SetRun(runObj)

	queueObj, err := tx.Queue.Get(r.ctx, input.QueueID)
	if err != nil {
		return nil, fmt.Errorf("getting queue: %w", err)
	}
	builder.SetQueue(queueObj)

	entObj, err := builder.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("creating side effect entity: %w", err)
	}

	sideEffectData, err := tx.SideEffectData.Create().
		SetEntity(entObj).
		SetInput(input.Input).
		Save(r.ctx)
	if err != nil {
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating side effect data: %w", err)
	}

	realStatus := execution.Status(StatusRunning)

	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(realStatus).
		Save(r.ctx)
	if err != nil {
		_ = tx.SideEffectData.DeleteOne(sideEffectData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating side effect execution: %w", err)
	}

	sideEffectExec, err := tx.SideEffectExecution.Create().
		SetExecution(execObj).
		Save(r.ctx)
	if err != nil {
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.SideEffectData.DeleteOne(sideEffectData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating side effect execution: %w", err)
	}

	now := time.Now()
	_, err = tx.SideEffectExecutionData.Create().
		SetSideEffectExecution(sideEffectExec).
		SetEffectTime(now).
		SetEffectMetadata(input.Metadata).
		Save(r.ctx)
	if err != nil {
		_ = tx.SideEffectExecution.DeleteOne(sideEffectExec).Exec(r.ctx)
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.SideEffectData.DeleteOne(sideEffectData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating side effect execution data: %w", err)
	}

	queueID, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	return &SideEffectInfo{
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
		Data: &SideEffectDataInfo{
			ID:    sideEffectData.ID,
			Input: sideEffectData.Input,
		},
		Execution: &SideEffectExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			EffectTime: now,
		},
	}, nil
}

func (r *sideEffectRepository) Get(tx *ent.Tx, id int) (*SideEffectInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	if entObj.Type != entity.Type(ComponentSideEffect) {
		return nil, fmt.Errorf("%w: entity is not a side effect", ErrInvalidOperation)
	}

	runID, err := entObj.QueryRun().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	queueID, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	sideEffectData, err := tx.SideEffectData.Query().
		Where(sideeffectdata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting side effect data: %w", err)
	}

	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(id))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	sideEffectExec, err := execObj.QuerySideEffectExecution().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting side effect execution: %w", err)
	}

	sideEffectExecData, err := sideEffectExec.QueryExecutionData().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting side effect execution data: %w", err)
	}

	return &SideEffectInfo{
		EntityInfo: EntityInfo{
			ID:          entObj.ID,
			HandlerName: entObj.HandlerName,
			Type:        ComponentType(entObj.Type),
			StepID:      entObj.StepID,
			RunID:       runID,
			CreatedAt:   entObj.CreatedAt,
			UpdatedAt:   entObj.UpdatedAt,
			QueueID:     queueID,
		},
		Data: &SideEffectDataInfo{
			ID:     sideEffectData.ID,
			Input:  sideEffectData.Input,
			Output: sideEffectData.Output,
		},
		Execution: &SideEffectExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			Result: sideEffectExec.Result,
			// EffectTime:     sideEffectExecData.EffectTime,
			EffectMetadata: sideEffectExecData.EffectMetadata,
		},
	}, nil
}

func (r *sideEffectRepository) GetByStepID(tx *ent.Tx, stepID string) (*SideEffectInfo, error) {
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

func (r *sideEffectRepository) List(tx *ent.Tx, runID int) ([]*SideEffectInfo, error) {
	entObjs, err := tx.Entity.Query().
		Where(
			entity.TypeEQ(entity.Type(ComponentSideEffect)),
			entity.HasRunWith(run.IDEQ(runID)),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing side effects: %w", err)
	}

	result := make([]*SideEffectInfo, 0, len(entObjs))
	for _, entObj := range entObjs {
		info, err := r.Get(tx, entObj.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

func (r *sideEffectRepository) UpdateData(tx *ent.Tx, id int, input UpdateSideEffectDataInput) (*SideEffectInfo, error) {
	dataUpdate := tx.SideEffectData.Update().
		Where(sideeffectdata.HasEntityWith(entity.IDEQ(id)))

	if input.Input != nil {
		dataUpdate.SetInput(input.Input)
	}
	if input.Output != nil {
		dataUpdate.SetOutput(input.Output)
	}

	_, err := dataUpdate.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating side effect data: %w", err)
	}

	return r.Get(tx, id)
}

func (r *sideEffectRepository) SetOutput(tx *ent.Tx, id int, output [][]byte) error {
	_, err := tx.SideEffectData.Update().
		Where(sideeffectdata.HasEntityWith(entity.IDEQ(id))).
		SetOutput(output).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("setting side effect output: %w", err)
	}
	return nil
}

func (r *sideEffectRepository) SetResult(tx *ent.Tx, id int, result []byte) error {
	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(id))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		return fmt.Errorf("getting execution: %w", err)
	}

	sideEffectExec, err := execObj.QuerySideEffectExecution().Only(r.ctx)
	if err != nil {
		return fmt.Errorf("getting side effect execution: %w", err)
	}

	_, err = tx.SideEffectExecution.UpdateOne(sideEffectExec).
		SetResult(result).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("setting side effect result: %w", err)
	}

	return nil
}
