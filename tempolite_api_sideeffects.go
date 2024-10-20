package tempolite

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) GetSideEffect(id SideEffectID) *SideEffectInfo[T] {
	tp.logger.Debug(tp.ctx, "GetSideEffect", "sideEffectID", id)
	return &SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
	}
}

func (tp *Tempolite[T]) getSideEffect(ctx TempoliteContext, id SideEffectID, err error) *SideEffectInfo[T] {
	tp.logger.Debug(tp.ctx, "getSideEffect", "sideEffectID", id, "error", err)
	info := SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
		err:      err,
	}
	return &info
}

// That's how we should use it
//
// ```go
// var value int
//
//	err := ctx.SideEffect("eventual switch", func(ctx SideEffectContext[testIdentifier]) int {
//		return 420
//	}).Get(&value)
//
// ```
func (tp *Tempolite[T]) enqueueSideEffect(ctx TempoliteContext, stepID T, sideEffectHandler interface{}) (SideEffectID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with side effect creation
	default:
		tp.logger.Error(tp.ctx, "creating side effect context entity type not supported", "entityType", ctx.EntityType())
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error
	var tx *ent.Tx

	// Generate a unique identifier for the side effect function
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffectHandler).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	tp.logger.Debug(tp.ctx, "EnqueueSideEffect", "stepID", stepID, "handlerIdentity", handlerIdentity)

	tp.logger.Debug(tp.ctx, "check for existing side effect", "stepID", stepID)
	// Check for existing side effect with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSideEffect),
		).
		First(tp.ctx)
	if err == nil {
		sideEffectEntity, err := tp.client.SideEffect.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting side effect", "error", err)
			return "", err
		}
		if sideEffectEntity.Status == sideeffect.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Side effect already completed", "sideEffectID", exists.ChildEntityID)
			return SideEffectID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	tp.logger.Debug(tp.ctx, "Creating transaction to create side effect", "handlerIdentity", handlerIdentity)
	if tx, err = tp.client.Tx(tp.ctx); err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Creating side effect entity", "handlerIdentity", handlerIdentity)
	sideEffectEntity, err := tx.SideEffect.
		Create().
		SetID(uuid.NewString()).
		SetStepID(fmt.Sprint(stepID)).
		SetIdentity(string(handlerIdentity)).
		SetHandlerName(funcName).
		SetStatus(sideeffect.StatusPending).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect entity", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating side effect entity", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "Creating side effect execution")
	sideEffectExecution, err := tx.SideEffectExecution.
		Create().
		SetID(uuid.NewString()).
		// SetRunID(ctx.RunID()).
		SetSideEffect(sideEffectEntity).
		SetStatus(sideeffectexecution.StatusPending).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect execution", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating side effect execution", "error", err)
		return "", fmt.Errorf("failed to create side effect execution: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(sideEffectEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(sideEffectExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSideEffect).
		Save(tp.ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
			return "", err
		}
		tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
		return "", err
	}

	tp.logger.Debug(tp.ctx, "analyze side effect handler", "handlerIdentity", handlerIdentity)

	// Analyze the side effect handler
	handlerType := reflect.TypeOf(sideEffectHandler)
	if handlerType.Kind() != reflect.Func {
		tp.logger.Error(tp.ctx, "Side effect must be a function", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect must be a function")
	}

	if handlerType.NumIn() != 1 || handlerType.In(0) != reflect.TypeOf(SideEffectContext[T]{}) {
		tp.logger.Error(tp.ctx, "Side effect function must have exactly one input parameter of type SideEffectContext[T]", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect function must have exactly one input parameter of type SideEffectContext[T]")
	}

	// Collect all return types
	numOut := handlerType.NumOut()
	if numOut == 0 {
		tp.logger.Error(tp.ctx, "Side effect function must return at least one value", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("side effect function must return at least one value")
	}

	returnTypes := make([]reflect.Type, numOut)
	returnKinds := make([]reflect.Kind, numOut)
	for i := 0; i < numOut; i++ {
		returnTypes[i] = handlerType.Out(i)
		returnKinds[i] = handlerType.Out(i).Kind()
	}

	tp.logger.Debug(tp.ctx, "Caching side effect info", "handlerIdentity", handlerIdentity)
	// Cache the side effect info
	tp.sideEffects.Store(sideEffectEntity.ID, SideEffect{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         sideEffectHandler,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumOut:          numOut,
	})

	tp.logger.Debug(tp.ctx, "Committing transaction creating side effect", "handlerIdentity", handlerIdentity)
	if err = tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			tp.sideEffects.Delete(sideEffectEntity.ID)
			tp.logger.Error(tp.ctx, "Error rolling back transaction creating side effect", "error", err)
			return "", err
		}
		tp.sideEffects.Delete(sideEffectEntity.ID)
		tp.logger.Error(tp.ctx, "Error committing transaction creating side effect", "error", err)
		return "", err
	}
	tp.logger.Debug(tp.ctx, "Side effect created", "sideEffectID", sideEffectEntity.ID, "handlerIdentity", handlerIdentity)

	log.Printf("Enqueued side effect %s with ID %s", funcName, sideEffectEntity.ID)
	return SideEffectID(sideEffectEntity.ID), nil
}

func (tp *Tempolite[T]) enqueueSideEffectFunc(ctx TempoliteContext, stepID T, sideEffect interface{}) (SideEffectID, error) {
	tp.logger.Debug(tp.ctx, "EnqueueSideEffectFunc", "stepID", stepID)
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffect).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSideEffect(ctx, stepID, handlerIdentity)
}
