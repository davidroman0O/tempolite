package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/executionunit"
	"github.com/davidroman0O/go-tempolite/ent/sagacompensation"
	"github.com/davidroman0O/go-tempolite/ent/sagatransaction"
	"github.com/google/uuid"
)

type SagaStep struct {
	TransactionHandler  interface{}
	CompensationHandler interface{}
}

func (t *Tempolite) CreateSaga(ctx context.Context, steps []SagaStep) (*ent.ExecutionUnit, error) {
	tx, err := t.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	execCtx, err := tx.ExecutionContext.
		Create().
		SetID(uuid.New().String()).
		SetCurrentRunID(uuid.New().String()).
		SetStatus(executioncontext.StatusRunning).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution context: %w", err)
	}

	sagaExecUnit, err := tx.ExecutionUnit.
		Create().
		SetID(uuid.New().String()).
		SetType(executionunit.TypeSaga).
		SetStatus(executionunit.StatusPending).
		SetExecutionContext(execCtx).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create saga execution unit: %w", err)
	}

	for i, step := range steps {
		transactionName := getFunctionName(step.TransactionHandler)
		compensationName := getFunctionName(step.CompensationHandler)

		sagaTransaction, err := tx.SagaTransaction.
			Create().
			SetID(uuid.New().String()).
			SetOrder(i).
			SetNextTransactionName(transactionName).
			SetFailureCompensationName(compensationName).
			SetExecutionUnit(sagaExecUnit).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create saga transaction: %w", err)
		}

		_, err = tx.SagaCompensation.
			Create().
			SetID(uuid.New().String()).
			SetOrder(i).
			SetNextCompensationName(compensationName).
			SetExecutionUnit(sagaExecUnit).
			SetTransaction(sagaTransaction).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create saga compensation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sagaExecUnit, nil
}

func (t *Tempolite) ExecuteSaga(ctx context.Context, sagaID string) error {
	saga, err := t.client.ExecutionUnit.Get(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to get saga: %w", err)
	}

	transactions, err := saga.QuerySagaTransactions().Order(ent.Asc(sagatransaction.FieldOrder)).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get saga transactions: %w", err)
	}

	for _, transaction := range transactions {
		taskID, err := t.Enqueue(ctx, transaction.NextTransactionName, nil, WithExecutionContextID(saga.ID))
		if err != nil {
			return fmt.Errorf("failed to enqueue transaction task: %w", err)
		}

		_, err = t.WaitFor(ctx, taskID)
		if err != nil {
			// Transaction failed, start compensation
			return t.compensateSaga(ctx, saga.ID, transaction.Order)
		}
	}

	// All transactions completed successfully
	_, err = t.client.ExecutionUnit.UpdateOne(saga).SetStatus(executionunit.StatusCompleted).Save(ctx)
	return err
}

func (t *Tempolite) compensateSaga(ctx context.Context, sagaID string, failedTransactionOrder int) error {
	saga, err := t.client.ExecutionUnit.Get(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to get saga: %w", err)
	}

	compensations, err := saga.QuerySagaCompensations().
		Where(sagacompensation.OrderLTE(failedTransactionOrder)).
		Order(ent.Desc(sagacompensation.FieldOrder)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get saga compensations: %w", err)
	}

	for _, compensation := range compensations {
		taskID, err := t.Enqueue(ctx, compensation.NextCompensationName, nil, WithExecutionContextID(saga.ID))
		if err != nil {
			return fmt.Errorf("failed to enqueue compensation task: %w", err)
		}

		_, err = t.WaitFor(ctx, taskID)
		if err != nil {
			// Compensation failed, log error and continue with other compensations
			log.Printf("Compensation failed for saga %s, step %d: %v", sagaID, compensation.Order, err)
		}
	}

	// Mark saga as failed after compensations
	_, err = t.client.ExecutionUnit.UpdateOne(saga).SetStatus(executionunit.StatusFailed).Save(ctx)
	return err
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
