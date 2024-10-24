package tempolite

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/activity"
	"github.com/davidroman0O/tempolite/ent/executionrelationship"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) GetActivity(id ActivityID) (*ActivityInfo[T], error) {
	tp.logger.Debug(tp.ctx, "GetActivity", "activityID", id)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
	}
	return &info, nil
}

func (tp *Tempolite[T]) getActivity(ctx TempoliteContext, id ActivityID, err error) *ActivityInfo[T] {
	tp.logger.Debug(tp.ctx, "getActivity", "activityID", id, "error", err)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getActivityExecution(ctx TempoliteContext, id ActivityExecutionID, err error) *ActivityExecutionInfo[T] {
	tp.logger.Debug(tp.ctx, "getActivityExecution", "activityExecutionID", id, "error", err)
	info := ActivityExecutionInfo[T]{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite[T]) enqueueActivity(ctx WorkflowContext[T], stepID T, longName HandlerIdentity, params ...interface{}) (ActivityID, error) {

	tp.logger.Debug(tp.ctx, "EnqueueActivity", "stepID", stepID, "longName", longName)
	switch ctx.EntityType() {
	case "workflow":
		// nothing
	default:
		tp.logger.Error(tp.ctx, "Context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error

	// Check for existing activity with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeActivity),
		).
		First(tp.ctx)
	if err == nil {
		// todo: is there a way to just get the status?
		act, err := tp.client.Activity.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting activity", "error", err)
			return "", err
		}
		if act.Status == activity.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Activity already completed", "activityID", exists.ChildEntityID)
			return ActivityID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating activity", "longName", longName)

	var value any
	var ok bool
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching activity handler", "longName", longName)
	if value, ok = tp.activities.Load(longName); ok {
		var activityHandlerInfo Activity
		if activityHandlerInfo, ok = value.(Activity); !ok {
			// could be development bug
			tp.logger.Error(tp.ctx, "Activity is not handler info", "longName", longName)
			return "", fmt.Errorf("activity %s is not handler info", longName)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "activityHandlerInfo", activityHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(activityHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create activity", "longName", longName)
		// Proceed to create a new activity and activity execution
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating activity entity", "longName", longName, "stepID", stepID)
		var activityEntity *ent.Activity
		if activityEntity, err = tx.
			Activity.
			Create().
			SetID(uuid.NewString()).
			SetStepID(fmt.Sprint(stepID)).
			SetIdentity(string(longName)).
			SetHandlerName(activityHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating activity entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating activity entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Created activity execution", "activityID", activityEntity.ID, "longName", longName, "stepID", stepID)
		// Create activity execution with the deterministic ID
		var activityExecution *ent.ActivityExecution
		if activityExecution, err = tx.ActivityExecution.
			Create().
			SetID(activityEntity.ID). // Use the deterministic activity ID
			SetRunID(ctx.RunID()).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating activity execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating activity execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Created activity relationship", "activityID", activityEntity.ID, "actvityExecution", activityExecution.ID, "longName", longName, "stepID", stepID)
		// Add execution relationship
		if _, err := tx.ExecutionRelationship.Create().
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(activityEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(activityEntity.ID).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(fmt.Sprint(stepID)).
			// types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeActivity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Committing transaction creating activity", "longName", longName)
		if err = tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Error committing transaction creating activity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Activity created", "activityID", activityEntity.ID, "longName", longName, "stepID", stepID)
		return ActivityID(activityEntity.ID), nil

	} else {
		tp.logger.Error(tp.ctx, "Activity not found", "longName", longName)
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite[T]) enqueueActivityFunc(ctx WorkflowContext[T], stepID T, activityFunc interface{}, params ...interface{}) (ActivityID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	tp.logger.Debug(tp.ctx, "Enqueue ActivityFunc", "stepID", stepID, "handlerIdentity", handlerIdentity)
	return tp.enqueueActivity(ctx, stepID, handlerIdentity, params...)
}
