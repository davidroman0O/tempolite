// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ExecutionDelete is the builder for deleting a Execution entity.
type ExecutionDelete struct {
	config
	hooks    []Hook
	mutation *ExecutionMutation
}

// Where appends a list predicates to the ExecutionDelete builder.
func (ed *ExecutionDelete) Where(ps ...predicate.Execution) *ExecutionDelete {
	ed.mutation.Where(ps...)
	return ed
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ed *ExecutionDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ed.sqlExec, ed.mutation, ed.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ed *ExecutionDelete) ExecX(ctx context.Context) int {
	n, err := ed.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ed *ExecutionDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(execution.Table, sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt))
	if ps := ed.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ed.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ed.mutation.done = true
	return affected, err
}

// ExecutionDeleteOne is the builder for deleting a single Execution entity.
type ExecutionDeleteOne struct {
	ed *ExecutionDelete
}

// Where appends a list predicates to the ExecutionDelete builder.
func (edo *ExecutionDeleteOne) Where(ps ...predicate.Execution) *ExecutionDeleteOne {
	edo.ed.mutation.Where(ps...)
	return edo
}

// Exec executes the deletion query.
func (edo *ExecutionDeleteOne) Exec(ctx context.Context) error {
	n, err := edo.ed.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{execution.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (edo *ExecutionDeleteOne) ExecX(ctx context.Context) {
	if err := edo.Exec(ctx); err != nil {
		panic(err)
	}
}
