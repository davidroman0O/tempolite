package tempolite

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/signal"
	"github.com/davidroman0O/go-tempolite/ent/signalexecution"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) PublishSignal(workflowID WorkflowID, stepID T, value interface{}) error {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			return tp.ctx.Err()
		case <-ticker.C:
			// Check if there is a workflowID and stepId in the execution relationship
			relationship, err := tp.client.ExecutionRelationship.Query().
				Where(
					executionrelationship.ParentEntityID(workflowID.String()),
					executionrelationship.ChildStepID(fmt.Sprint(stepID)),
					executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSignal),
				).
				First(tp.ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					continue // Signal not created yet, wait for next tick
				}
				return fmt.Errorf("error querying execution relationship: %w", err)
			}

			// Find the latest completed signal execution
			latestExecution, err := tp.client.SignalExecution.Query().
				Where(
					signalexecution.HasSignalWith(signal.ID(relationship.ChildEntityID)),
					signalexecution.StatusEQ(signalexecution.StatusPending),
				).
				Order(ent.Desc(signalexecution.FieldUpdatedAt)).
				First(tp.ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					// fmt.Println("Signal not completed yet, waiting for next tick")
					continue // No completed execution yet, wait for next tick
				}
				return fmt.Errorf("error querying signal execution: %w", err)
			}

			// Encode the value to JSON
			jsonValue, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("error encoding signal value: %w", err)
			}

			// Update the signal execution with the JSON-encoded value
			_, err = tp.client.SignalExecution.UpdateOne(latestExecution).
				SetOutput([]interface{}{string(jsonValue)}).
				SetStatus(signalexecution.StatusCompleted).
				Save(tp.ctx)
			if err != nil {
				return fmt.Errorf("error updating signal execution: %w", err)
			}

			return nil
		}
	}
}

func (tp *Tempolite[T]) enqueueSignal(ctx WorkflowContext[T], stepID T) (SignalID, error) {
	// Check for existing completed signal execution
	relationship, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.ParentEntityID(ctx.EntityID()),
			executionrelationship.ChildStepID(fmt.Sprint(stepID)),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSignal),
		).
		First(tp.ctx)
	if err == nil {
		// Found existing relationship, check for completed execution
		latestExecution, err := tp.client.SignalExecution.Query().
			Where(
				signalexecution.HasSignalWith(signal.ID(relationship.ChildEntityID)),
				signalexecution.StatusEQ(signalexecution.StatusCompleted),
			).
			Order(ent.Desc(signalexecution.FieldUpdatedAt)).
			WithSignal().
			First(tp.ctx)
		if err == nil {
			// Found completed execution, return its ID
			return SignalID(latestExecution.Edges.Signal.ID), nil
		}
	} else if !ent.IsNotFound(err) {
		return "", fmt.Errorf("error checking for existing signal: %w", err)
	}

	// No existing completed signal, create a new one
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	signalEntity, err := tx.Signal.Create().
		SetID(uuid.New().String()).
		SetStepID(fmt.Sprint(stepID)).
		SetStatus(signal.StatusPending).
		Save(tp.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create signal entity: %w", err)
	}

	signalExecution, err := tx.SignalExecution.Create().
		SetID(uuid.New().String()).
		SetRunID(ctx.RunID()).
		SetSignal(signalEntity).
		SetStatus(signalexecution.StatusPending).
		Save(tp.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create signal execution: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(signalEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(signalExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSignal).
		Save(tp.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create execution relationship: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return SignalID(signalEntity.ID), nil
}

type SignalInfo[T Identifier] struct {
	tp       *Tempolite[T]
	EntityID SignalID
	err      error
}

func (tp *Tempolite[T]) getSignalInfo(id SignalID, err error) *SignalInfo[T] {
	return &SignalInfo[T]{
		tp:       tp,
		EntityID: id,
		err:      err,
	}
}

func (s *SignalInfo[T]) Receive(ctx WorkflowContext[T], value interface{}) error {
	if s.err != nil {
		return s.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.tp.ctx.Done():
			return ctx.tp.ctx.Err()
		case <-ticker.C:
			signalExecution, err := s.tp.client.SignalExecution.Query().
				Where(
					signalexecution.HasSignalWith(signal.ID(s.EntityID.String())),
					signalexecution.StatusEQ(signalexecution.StatusCompleted),
				).
				Order(ent.Desc(signalexecution.FieldUpdatedAt)).
				First(s.tp.ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					continue // No completed execution yet, wait for next tick
				}
				return fmt.Errorf("error querying signal execution: %w", err)
			}

			// Check if value is a pointer
			valueType := reflect.TypeOf(value)
			if valueType.Kind() != reflect.Ptr {
				return fmt.Errorf("value must be a pointer")
			}

			// Check if the output slice is empty
			if len(signalExecution.Output) == 0 {
				return fmt.Errorf("signal execution output is empty")
			}

			// Get the JSON-encoded output value
			jsonValue, ok := signalExecution.Output[0].(string)
			if !ok {
				return fmt.Errorf("unexpected output type: expected string, got %T", signalExecution.Output[0])
			}

			// Decode the JSON value into the provided pointer
			if err := json.Unmarshal([]byte(jsonValue), value); err != nil {
				return fmt.Errorf("error decoding signal value: %w", err)
			}

			return nil
		}
	}
}

func (w *WorkflowContext[T]) Signal(stepID T) *SignalInfo[T] {
	id, err := w.tp.enqueueSignal(*w, stepID)
	return w.tp.getSignalInfo(id, err)
}
