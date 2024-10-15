// Code generated by ent, DO NOT EDIT.

package hook

import (
	"context"
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent"
)

// The ActivityFunc type is an adapter to allow the use of ordinary
// function as Activity mutator.
type ActivityFunc func(context.Context, *ent.ActivityMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f ActivityFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.ActivityMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.ActivityMutation", m)
}

// The ActivityExecutionFunc type is an adapter to allow the use of ordinary
// function as ActivityExecution mutator.
type ActivityExecutionFunc func(context.Context, *ent.ActivityExecutionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f ActivityExecutionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.ActivityExecutionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.ActivityExecutionMutation", m)
}

// The ExecutionRelationshipFunc type is an adapter to allow the use of ordinary
// function as ExecutionRelationship mutator.
type ExecutionRelationshipFunc func(context.Context, *ent.ExecutionRelationshipMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f ExecutionRelationshipFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.ExecutionRelationshipMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.ExecutionRelationshipMutation", m)
}

// The FeatureFlagVersionFunc type is an adapter to allow the use of ordinary
// function as FeatureFlagVersion mutator.
type FeatureFlagVersionFunc func(context.Context, *ent.FeatureFlagVersionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f FeatureFlagVersionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.FeatureFlagVersionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.FeatureFlagVersionMutation", m)
}

// The RunFunc type is an adapter to allow the use of ordinary
// function as Run mutator.
type RunFunc func(context.Context, *ent.RunMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f RunFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.RunMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.RunMutation", m)
}

// The SagaFunc type is an adapter to allow the use of ordinary
// function as Saga mutator.
type SagaFunc func(context.Context, *ent.SagaMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SagaFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SagaMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SagaMutation", m)
}

// The SagaExecutionFunc type is an adapter to allow the use of ordinary
// function as SagaExecution mutator.
type SagaExecutionFunc func(context.Context, *ent.SagaExecutionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SagaExecutionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SagaExecutionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SagaExecutionMutation", m)
}

// The SideEffectFunc type is an adapter to allow the use of ordinary
// function as SideEffect mutator.
type SideEffectFunc func(context.Context, *ent.SideEffectMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SideEffectFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SideEffectMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SideEffectMutation", m)
}

// The SideEffectExecutionFunc type is an adapter to allow the use of ordinary
// function as SideEffectExecution mutator.
type SideEffectExecutionFunc func(context.Context, *ent.SideEffectExecutionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SideEffectExecutionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SideEffectExecutionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SideEffectExecutionMutation", m)
}

// The SignalFunc type is an adapter to allow the use of ordinary
// function as Signal mutator.
type SignalFunc func(context.Context, *ent.SignalMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f SignalFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.SignalMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.SignalMutation", m)
}

// The WorkflowFunc type is an adapter to allow the use of ordinary
// function as Workflow mutator.
type WorkflowFunc func(context.Context, *ent.WorkflowMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f WorkflowFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.WorkflowMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.WorkflowMutation", m)
}

// The WorkflowExecutionFunc type is an adapter to allow the use of ordinary
// function as WorkflowExecution mutator.
type WorkflowExecutionFunc func(context.Context, *ent.WorkflowExecutionMutation) (ent.Value, error)

// Mutate calls f(ctx, m).
func (f WorkflowExecutionFunc) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	if mv, ok := m.(*ent.WorkflowExecutionMutation); ok {
		return f(ctx, mv)
	}
	return nil, fmt.Errorf("unexpected mutation type %T. expect *ent.WorkflowExecutionMutation", m)
}

// Condition is a hook condition function.
type Condition func(context.Context, ent.Mutation) bool

// And groups conditions with the AND operator.
func And(first, second Condition, rest ...Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		if !first(ctx, m) || !second(ctx, m) {
			return false
		}
		for _, cond := range rest {
			if !cond(ctx, m) {
				return false
			}
		}
		return true
	}
}

// Or groups conditions with the OR operator.
func Or(first, second Condition, rest ...Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		if first(ctx, m) || second(ctx, m) {
			return true
		}
		for _, cond := range rest {
			if cond(ctx, m) {
				return true
			}
		}
		return false
	}
}

// Not negates a given condition.
func Not(cond Condition) Condition {
	return func(ctx context.Context, m ent.Mutation) bool {
		return !cond(ctx, m)
	}
}

// HasOp is a condition testing mutation operation.
func HasOp(op ent.Op) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		return m.Op().Is(op)
	}
}

// HasAddedFields is a condition validating `.AddedField` on fields.
func HasAddedFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if _, exists := m.AddedField(field); !exists {
			return false
		}
		for _, field := range fields {
			if _, exists := m.AddedField(field); !exists {
				return false
			}
		}
		return true
	}
}

// HasClearedFields is a condition validating `.FieldCleared` on fields.
func HasClearedFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if exists := m.FieldCleared(field); !exists {
			return false
		}
		for _, field := range fields {
			if exists := m.FieldCleared(field); !exists {
				return false
			}
		}
		return true
	}
}

// HasFields is a condition validating `.Field` on fields.
func HasFields(field string, fields ...string) Condition {
	return func(_ context.Context, m ent.Mutation) bool {
		if _, exists := m.Field(field); !exists {
			return false
		}
		for _, field := range fields {
			if _, exists := m.Field(field); !exists {
				return false
			}
		}
		return true
	}
}

// If executes the given hook under condition.
//
//	hook.If(ComputeAverage, And(HasFields(...), HasAddedFields(...)))
func If(hk ent.Hook, cond Condition) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if cond(ctx, m) {
				return hk(next).Mutate(ctx, m)
			}
			return next.Mutate(ctx, m)
		})
	}
}

// On executes the given hook only for the given operation.
//
//	hook.On(Log, ent.Delete|ent.Create)
func On(hk ent.Hook, op ent.Op) ent.Hook {
	return If(hk, HasOp(op))
}

// Unless skips the given hook only for the given operation.
//
//	hook.Unless(Log, ent.Update|ent.UpdateOne)
func Unless(hk ent.Hook, op ent.Op) ent.Hook {
	return If(hk, Not(HasOp(op)))
}

// FixedError is a hook returning a fixed error.
func FixedError(err error) ent.Hook {
	return func(ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(context.Context, ent.Mutation) (ent.Value, error) {
			return nil, err
		})
	}
}

// Reject returns a hook that rejects all operations that match op.
//
//	func (T) Hooks() []ent.Hook {
//		return []ent.Hook{
//			Reject(ent.Delete|ent.Update),
//		}
//	}
func Reject(op ent.Op) ent.Hook {
	hk := FixedError(fmt.Errorf("%s operation is not allowed", op))
	return On(hk, op)
}

// Chain acts as a list of hooks and is effectively immutable.
// Once created, it will always hold the same set of hooks in the same order.
type Chain struct {
	hooks []ent.Hook
}

// NewChain creates a new chain of hooks.
func NewChain(hooks ...ent.Hook) Chain {
	return Chain{append([]ent.Hook(nil), hooks...)}
}

// Hook chains the list of hooks and returns the final hook.
func (c Chain) Hook() ent.Hook {
	return func(mutator ent.Mutator) ent.Mutator {
		for i := len(c.hooks) - 1; i >= 0; i-- {
			mutator = c.hooks[i](mutator)
		}
		return mutator
	}
}

// Append extends a chain, adding the specified hook
// as the last ones in the mutation flow.
func (c Chain) Append(hooks ...ent.Hook) Chain {
	newHooks := make([]ent.Hook, 0, len(c.hooks)+len(hooks))
	newHooks = append(newHooks, c.hooks...)
	newHooks = append(newHooks, hooks...)
	return Chain{newHooks}
}

// Extend extends a chain, adding the specified chain
// as the last ones in the mutation flow.
func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.hooks...)
}
