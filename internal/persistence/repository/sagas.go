package repository

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagadata"
)

type SagaInfo struct {
	EntityInfo
	Data      *SagaDataInfo      `json:"data,omitempty"`
	Execution *SagaExecutionInfo `json:"execution,omitempty"`
}

type SagaDataInfo struct {
	ID               int      `json:"id"`
	Compensating     bool     `json:"compensating"`
	CompensationData [][]byte `json:"compensation_data,omitempty"`
	Input            [][]byte `json:"input,omitempty"`
	Output           [][]byte `json:"output,omitempty"`
}

type SagaExecutionInfo struct {
	ExecutionInfo
	StepType         string `json:"step_type"`
	CompensationData []byte `json:"compensation_data,omitempty"`
}

type CreateSagaInput struct {
	RunID            int
	HandlerName      string
	StepID           string
	QueueID          int
	CompensationData [][]byte
	Input            [][]byte
}

type UpdateSagaDataInput struct {
	CompensationData [][]byte
	Input            [][]byte
	Output           [][]byte
	Compensating     *bool
}

type SagaRepository interface {
	Create(tx *ent.Tx, input CreateSagaInput) (*SagaInfo, error)
	Get(tx *ent.Tx, id int) (*SagaInfo, error)
	GetByStepID(tx *ent.Tx, stepID string) (*SagaInfo, error)
	List(tx *ent.Tx, runID int) ([]*SagaInfo, error)
	UpdateData(tx *ent.Tx, id int, input UpdateSagaDataInput) (*SagaInfo, error)
	SetOutput(tx *ent.Tx, id int, output [][]byte) error
	StartCompensation(tx *ent.Tx, id int) error
}

type sagaRepository struct {
	ctx    context.Context
	client *ent.Client
}

func NewSagaRepository(ctx context.Context, client *ent.Client) SagaRepository {
	return &sagaRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *sagaRepository) Create(tx *ent.Tx, input CreateSagaInput) (*SagaInfo, error) {
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

	realType := entity.Type(ComponentSaga)

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
		return nil, fmt.Errorf("creating saga entity: %w", err)
	}

	sagaData, err := tx.SagaData.Create().
		SetEntity(entObj).
		SetCompensationData(input.CompensationData).
		// SetInput(input.Input).
		Save(r.ctx)
	if err != nil {
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating saga data: %w", err)
	}

	realStatus := execution.Status(StatusRunning)

	execObj, err := tx.Execution.Create().
		SetEntity(entObj).
		SetStatus(realStatus).
		Save(r.ctx)
	if err != nil {
		_ = tx.SagaData.DeleteOne(sagaData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating saga execution: %w", err)
	}

	sagaExec, err := tx.SagaExecution.Create().
		SetExecution(execObj).
		SetStepType("transaction").
		Save(r.ctx)
	if err != nil {
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.SagaData.DeleteOne(sagaData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating saga execution: %w", err)
	}

	_, err = tx.SagaExecutionData.Create().
		SetSagaExecution(sagaExec).
		Save(r.ctx)
	if err != nil {
		_ = tx.SagaExecution.DeleteOne(sagaExec).Exec(r.ctx)
		_ = tx.Execution.DeleteOne(execObj).Exec(r.ctx)
		_ = tx.SagaData.DeleteOne(sagaData).Exec(r.ctx)
		_ = tx.Entity.DeleteOne(entObj).Exec(r.ctx)
		return nil, fmt.Errorf("creating saga execution data: %w", err)
	}

	queueID, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	return &SagaInfo{
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
		Data: &SagaDataInfo{
			ID:               sagaData.ID,
			Compensating:     sagaData.Compensating,
			CompensationData: sagaData.CompensationData,
			// Input:            sagaData.Input,
		},
		Execution: &SagaExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			// StepType: sagaExec.StepType,
		},
	}, nil
}

func (r *sagaRepository) Get(tx *ent.Tx, id int) (*SagaInfo, error) {
	entObj, err := tx.Entity.Get(r.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting entity: %w", err)
	}

	if entObj.Type != entity.Type(ComponentSaga) {
		return nil, fmt.Errorf("%w: entity is not a saga", ErrInvalidOperation)
	}

	runID, err := entObj.QueryRun().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting run ID: %w", err)
	}

	queueID, err := entObj.QueryQueue().OnlyID(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting queue ID: %w", err)
	}

	sagaData, err := tx.SagaData.Query().
		Where(sagadata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting saga data: %w", err)
	}

	execObj, err := tx.Execution.Query().
		Where(execution.HasEntityWith(entity.IDEQ(id))).
		Order(ent.Desc(execution.FieldCreatedAt)).
		First(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting execution: %w", err)
	}

	sagaExec, err := execObj.QuerySagaExecution().Only(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting saga execution: %w", err)
	}

	return &SagaInfo{
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
		Data: &SagaDataInfo{
			ID:               sagaData.ID,
			Compensating:     sagaData.Compensating,
			CompensationData: sagaData.CompensationData,
			// Input:            sagaData.Input,
			// Output:           sagaData.Output,
		},
		Execution: &SagaExecutionInfo{
			ExecutionInfo: ExecutionInfo{
				ID:          execObj.ID,
				EntityID:    entObj.ID,
				Status:      Status(execObj.Status),
				StartedAt:   execObj.StartedAt,
				CompletedAt: execObj.CompletedAt,
				CreatedAt:   execObj.CreatedAt,
				UpdatedAt:   execObj.UpdatedAt,
			},
			// StepType:         sagaExec.StepType,
			CompensationData: sagaExec.CompensationData,
		},
	}, nil
}

func (r *sagaRepository) GetByStepID(tx *ent.Tx, stepID string) (*SagaInfo, error) {
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

func (r *sagaRepository) List(tx *ent.Tx, runID int) ([]*SagaInfo, error) {
	entObjs, err := tx.Entity.Query().
		Where(
			entity.TypeEQ(entity.Type(ComponentSaga)),
			entity.HasRunWith(run.IDEQ(runID)),
		).
		All(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing sagas: %w", err)
	}

	result := make([]*SagaInfo, 0, len(entObjs))
	for _, entObj := range entObjs {
		info, err := r.Get(tx, entObj.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

func (r *sagaRepository) UpdateData(tx *ent.Tx, id int, input UpdateSagaDataInput) (*SagaInfo, error) {
	dataUpdate := tx.SagaData.Update().
		Where(sagadata.HasEntityWith(entity.IDEQ(id)))

	if input.CompensationData != nil {
		dataUpdate.SetCompensationData(input.CompensationData)
	}
	// if input.Input != nil {
	// 	dataUpdate.SetInput(input.Input)
	// }
	// if input.Output != nil {
	// 	dataUpdate.SetOutput(input.Output)
	// }
	if input.Compensating != nil {
		dataUpdate.SetCompensating(*input.Compensating)
	}

	_, err := dataUpdate.Save(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("updating saga data: %w", err)
	}

	return r.Get(tx, id)
}

func (r *sagaRepository) SetOutput(tx *ent.Tx, id int, output [][]byte) error {
	_, err := tx.SagaData.Update().
		Where(sagadata.HasEntityWith(entity.IDEQ(id))).
		// SetOutput(output).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("setting saga output: %w", err)
	}
	return nil
}

func (r *sagaRepository) StartCompensation(tx *ent.Tx, id int) error {
	sagaData, err := tx.SagaData.Query().
		Where(sagadata.HasEntityWith(entity.IDEQ(id))).
		Only(r.ctx)
	if err != nil {
		return fmt.Errorf("getting saga data: %w", err)
	}

	if sagaData.Compensating {
		return fmt.Errorf("%w: saga compensation already started", ErrInvalidOperation)
	}

	_, err = tx.SagaData.Update().
		Where(sagadata.HasEntityWith(entity.IDEQ(id))).
		SetCompensating(true).
		Save(r.ctx)
	if err != nil {
		return fmt.Errorf("starting compensation: %w", err)
	}

	return nil
}
