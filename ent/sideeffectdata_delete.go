// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/sideeffectdata"
)

// SideEffectDataDelete is the builder for deleting a SideEffectData entity.
type SideEffectDataDelete struct {
	config
	hooks    []Hook
	mutation *SideEffectDataMutation
}

// Where appends a list predicates to the SideEffectDataDelete builder.
func (sedd *SideEffectDataDelete) Where(ps ...predicate.SideEffectData) *SideEffectDataDelete {
	sedd.mutation.Where(ps...)
	return sedd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (sedd *SideEffectDataDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, sedd.sqlExec, sedd.mutation, sedd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (sedd *SideEffectDataDelete) ExecX(ctx context.Context) int {
	n, err := sedd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (sedd *SideEffectDataDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(sideeffectdata.Table, sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt))
	if ps := sedd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, sedd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	sedd.mutation.done = true
	return affected, err
}

// SideEffectDataDeleteOne is the builder for deleting a single SideEffectData entity.
type SideEffectDataDeleteOne struct {
	sedd *SideEffectDataDelete
}

// Where appends a list predicates to the SideEffectDataDelete builder.
func (seddo *SideEffectDataDeleteOne) Where(ps ...predicate.SideEffectData) *SideEffectDataDeleteOne {
	seddo.sedd.mutation.Where(ps...)
	return seddo
}

// Exec executes the deletion query.
func (seddo *SideEffectDataDeleteOne) Exec(ctx context.Context) error {
	n, err := seddo.sedd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{sideeffectdata.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (seddo *SideEffectDataDeleteOne) ExecX(ctx context.Context) {
	if err := seddo.Exec(ctx); err != nil {
		panic(err)
	}
}
