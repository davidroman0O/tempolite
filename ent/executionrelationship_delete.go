// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/executionrelationship"
	"github.com/davidroman0O/tempolite/ent/predicate"
)

// ExecutionRelationshipDelete is the builder for deleting a ExecutionRelationship entity.
type ExecutionRelationshipDelete struct {
	config
	hooks    []Hook
	mutation *ExecutionRelationshipMutation
}

// Where appends a list predicates to the ExecutionRelationshipDelete builder.
func (erd *ExecutionRelationshipDelete) Where(ps ...predicate.ExecutionRelationship) *ExecutionRelationshipDelete {
	erd.mutation.Where(ps...)
	return erd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (erd *ExecutionRelationshipDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, erd.sqlExec, erd.mutation, erd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (erd *ExecutionRelationshipDelete) ExecX(ctx context.Context) int {
	n, err := erd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (erd *ExecutionRelationshipDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(executionrelationship.Table, sqlgraph.NewFieldSpec(executionrelationship.FieldID, field.TypeInt))
	if ps := erd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, erd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	erd.mutation.done = true
	return affected, err
}

// ExecutionRelationshipDeleteOne is the builder for deleting a single ExecutionRelationship entity.
type ExecutionRelationshipDeleteOne struct {
	erd *ExecutionRelationshipDelete
}

// Where appends a list predicates to the ExecutionRelationshipDelete builder.
func (erdo *ExecutionRelationshipDeleteOne) Where(ps ...predicate.ExecutionRelationship) *ExecutionRelationshipDeleteOne {
	erdo.erd.mutation.Where(ps...)
	return erdo
}

// Exec executes the deletion query.
func (erdo *ExecutionRelationshipDeleteOne) Exec(ctx context.Context) error {
	n, err := erdo.erd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{executionrelationship.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (erdo *ExecutionRelationshipDeleteOne) ExecX(ctx context.Context) {
	if err := erdo.Exec(ctx); err != nil {
		panic(err)
	}
}
