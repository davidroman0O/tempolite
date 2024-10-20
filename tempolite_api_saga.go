package tempolite

import (
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sagaexecution"
	"github.com/davidroman0O/go-tempolite/ent/schema"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) saga(ctx TempoliteContext, stepID T, saga *SagaDefinition[T]) *SagaInfo[T] {
	tp.logger.Debug(tp.ctx, "Saga", "stepID", stepID)
	id, err := tp.enqueueSaga(ctx, stepID, saga)
	return tp.getSaga(id, err)
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite[T]) enqueueSaga(ctx TempoliteContext, stepID T, sagaDef *SagaDefinition[T]) (SagaID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with saga creation
	default:
		tp.logger.Error(tp.ctx, "creating saga context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error
	var tx *ent.Tx
	tp.logger.Debug(tp.ctx, "EnqueueSaga", "stepID", stepID)

	tp.logger.Debug(tp.ctx, "check for existing saga", "stepID", stepID)
	// Check for existing saga with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSaga),
		).
		First(tp.ctx)
	if err == nil {
		sagaEntity, err := tp.client.Saga.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting saga", "error", err)
			return "", err
		}
		if sagaEntity.Status == saga.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Saga already completed", "sagaID", exists.ChildEntityID)
			return SagaID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction to create saga")
	if tx, err = tp.client.Tx(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
		return "", err
	}

	// Convert SagaHandlerInfo to SagaDefinitionData
	sagaDefData := schema.SagaDefinitionData{
		Steps: make([]schema.SagaStepPair, len(sagaDef.HandlerInfo.TransactionInfo)),
	}

	for i, txInfo := range sagaDef.HandlerInfo.TransactionInfo {
		sagaDefData.Steps[i] = schema.SagaStepPair{
			TransactionHandlerName:  txInfo.HandlerName,
			CompensationHandlerName: sagaDef.HandlerInfo.CompensationInfo[i].HandlerName,
		}
	}

	tp.logger.Debug(tp.ctx, "Creating saga entity")
	sagaEntity, err := tx.Saga.
		Create().
		SetID(uuid.NewString()).
		SetRunID(ctx.RunID()).
		SetStepID(fmt.Sprint(stepID)).
		SetStatus(saga.StatusPending).
		SetSagaDefinition(sagaDefData).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating saga entity", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating saga entity", "error", err)
		return "", err
	}

	// Create the first transaction step
	firstStep := sagaDefData.Steps[0]
	sagaExecution, err := tx.SagaExecution.
		Create().
		SetID(uuid.NewString()).
		SetStatus(sagaexecution.StatusPending).
		SetStepType(sagaexecution.StepTypeTransaction).
		SetHandlerName(firstStep.TransactionHandlerName).
		SetSequence(0).
		SetSaga(sagaEntity).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating first saga execution step", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating first saga execution step", "error", err)
		return "", fmt.Errorf("failed to create first saga execution step: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(sagaEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(sagaExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSaga).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Committing transaction creating saga")
	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating saga", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error committing transaction creating saga", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Saga created", "sagaID", sagaEntity.ID)
	// store the cache
	tp.sagas.Store(sagaEntity.ID, sagaDef)

	tp.logger.Debug(tp.ctx, "Created saga", "sagaID", sagaEntity.ID)
	return SagaID(sagaEntity.ID), nil
}

func (tp *Tempolite[T]) getSaga(id SagaID, err error) *SagaInfo[T] {
	tp.logger.Debug(tp.ctx, "getSaga", "sagaID", id, "error", err)
	return &SagaInfo[T]{
		tp:     tp,
		SagaID: id,
		err:    err,
	}
}
