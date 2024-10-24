// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/signal"
	"github.com/davidroman0O/tempolite/ent/signalexecution"
)

// SignalUpdate is the builder for updating Signal entities.
type SignalUpdate struct {
	config
	hooks    []Hook
	mutation *SignalMutation
}

// Where appends a list predicates to the SignalUpdate builder.
func (su *SignalUpdate) Where(ps ...predicate.Signal) *SignalUpdate {
	su.mutation.Where(ps...)
	return su
}

// SetStepID sets the "step_id" field.
func (su *SignalUpdate) SetStepID(s string) *SignalUpdate {
	su.mutation.SetStepID(s)
	return su
}

// SetNillableStepID sets the "step_id" field if the given value is not nil.
func (su *SignalUpdate) SetNillableStepID(s *string) *SignalUpdate {
	if s != nil {
		su.SetStepID(*s)
	}
	return su
}

// SetStatus sets the "status" field.
func (su *SignalUpdate) SetStatus(s signal.Status) *SignalUpdate {
	su.mutation.SetStatus(s)
	return su
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (su *SignalUpdate) SetNillableStatus(s *signal.Status) *SignalUpdate {
	if s != nil {
		su.SetStatus(*s)
	}
	return su
}

// SetCreatedAt sets the "created_at" field.
func (su *SignalUpdate) SetCreatedAt(t time.Time) *SignalUpdate {
	su.mutation.SetCreatedAt(t)
	return su
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (su *SignalUpdate) SetNillableCreatedAt(t *time.Time) *SignalUpdate {
	if t != nil {
		su.SetCreatedAt(*t)
	}
	return su
}

// SetConsumed sets the "consumed" field.
func (su *SignalUpdate) SetConsumed(b bool) *SignalUpdate {
	su.mutation.SetConsumed(b)
	return su
}

// SetNillableConsumed sets the "consumed" field if the given value is not nil.
func (su *SignalUpdate) SetNillableConsumed(b *bool) *SignalUpdate {
	if b != nil {
		su.SetConsumed(*b)
	}
	return su
}

// AddExecutionIDs adds the "executions" edge to the SignalExecution entity by IDs.
func (su *SignalUpdate) AddExecutionIDs(ids ...string) *SignalUpdate {
	su.mutation.AddExecutionIDs(ids...)
	return su
}

// AddExecutions adds the "executions" edges to the SignalExecution entity.
func (su *SignalUpdate) AddExecutions(s ...*SignalExecution) *SignalUpdate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return su.AddExecutionIDs(ids...)
}

// Mutation returns the SignalMutation object of the builder.
func (su *SignalUpdate) Mutation() *SignalMutation {
	return su.mutation
}

// ClearExecutions clears all "executions" edges to the SignalExecution entity.
func (su *SignalUpdate) ClearExecutions() *SignalUpdate {
	su.mutation.ClearExecutions()
	return su
}

// RemoveExecutionIDs removes the "executions" edge to SignalExecution entities by IDs.
func (su *SignalUpdate) RemoveExecutionIDs(ids ...string) *SignalUpdate {
	su.mutation.RemoveExecutionIDs(ids...)
	return su
}

// RemoveExecutions removes "executions" edges to SignalExecution entities.
func (su *SignalUpdate) RemoveExecutions(s ...*SignalExecution) *SignalUpdate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return su.RemoveExecutionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (su *SignalUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, su.sqlSave, su.mutation, su.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (su *SignalUpdate) SaveX(ctx context.Context) int {
	affected, err := su.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (su *SignalUpdate) Exec(ctx context.Context) error {
	_, err := su.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (su *SignalUpdate) ExecX(ctx context.Context) {
	if err := su.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (su *SignalUpdate) check() error {
	if v, ok := su.mutation.StepID(); ok {
		if err := signal.StepIDValidator(v); err != nil {
			return &ValidationError{Name: "step_id", err: fmt.Errorf(`ent: validator failed for field "Signal.step_id": %w`, err)}
		}
	}
	if v, ok := su.mutation.Status(); ok {
		if err := signal.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Signal.status": %w`, err)}
		}
	}
	return nil
}

func (su *SignalUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := su.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(signal.Table, signal.Columns, sqlgraph.NewFieldSpec(signal.FieldID, field.TypeString))
	if ps := su.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := su.mutation.StepID(); ok {
		_spec.SetField(signal.FieldStepID, field.TypeString, value)
	}
	if value, ok := su.mutation.Status(); ok {
		_spec.SetField(signal.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := su.mutation.CreatedAt(); ok {
		_spec.SetField(signal.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := su.mutation.Consumed(); ok {
		_spec.SetField(signal.FieldConsumed, field.TypeBool, value)
	}
	if su.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := su.mutation.RemovedExecutionsIDs(); len(nodes) > 0 && !su.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := su.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, su.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{signal.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	su.mutation.done = true
	return n, nil
}

// SignalUpdateOne is the builder for updating a single Signal entity.
type SignalUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *SignalMutation
}

// SetStepID sets the "step_id" field.
func (suo *SignalUpdateOne) SetStepID(s string) *SignalUpdateOne {
	suo.mutation.SetStepID(s)
	return suo
}

// SetNillableStepID sets the "step_id" field if the given value is not nil.
func (suo *SignalUpdateOne) SetNillableStepID(s *string) *SignalUpdateOne {
	if s != nil {
		suo.SetStepID(*s)
	}
	return suo
}

// SetStatus sets the "status" field.
func (suo *SignalUpdateOne) SetStatus(s signal.Status) *SignalUpdateOne {
	suo.mutation.SetStatus(s)
	return suo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (suo *SignalUpdateOne) SetNillableStatus(s *signal.Status) *SignalUpdateOne {
	if s != nil {
		suo.SetStatus(*s)
	}
	return suo
}

// SetCreatedAt sets the "created_at" field.
func (suo *SignalUpdateOne) SetCreatedAt(t time.Time) *SignalUpdateOne {
	suo.mutation.SetCreatedAt(t)
	return suo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (suo *SignalUpdateOne) SetNillableCreatedAt(t *time.Time) *SignalUpdateOne {
	if t != nil {
		suo.SetCreatedAt(*t)
	}
	return suo
}

// SetConsumed sets the "consumed" field.
func (suo *SignalUpdateOne) SetConsumed(b bool) *SignalUpdateOne {
	suo.mutation.SetConsumed(b)
	return suo
}

// SetNillableConsumed sets the "consumed" field if the given value is not nil.
func (suo *SignalUpdateOne) SetNillableConsumed(b *bool) *SignalUpdateOne {
	if b != nil {
		suo.SetConsumed(*b)
	}
	return suo
}

// AddExecutionIDs adds the "executions" edge to the SignalExecution entity by IDs.
func (suo *SignalUpdateOne) AddExecutionIDs(ids ...string) *SignalUpdateOne {
	suo.mutation.AddExecutionIDs(ids...)
	return suo
}

// AddExecutions adds the "executions" edges to the SignalExecution entity.
func (suo *SignalUpdateOne) AddExecutions(s ...*SignalExecution) *SignalUpdateOne {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return suo.AddExecutionIDs(ids...)
}

// Mutation returns the SignalMutation object of the builder.
func (suo *SignalUpdateOne) Mutation() *SignalMutation {
	return suo.mutation
}

// ClearExecutions clears all "executions" edges to the SignalExecution entity.
func (suo *SignalUpdateOne) ClearExecutions() *SignalUpdateOne {
	suo.mutation.ClearExecutions()
	return suo
}

// RemoveExecutionIDs removes the "executions" edge to SignalExecution entities by IDs.
func (suo *SignalUpdateOne) RemoveExecutionIDs(ids ...string) *SignalUpdateOne {
	suo.mutation.RemoveExecutionIDs(ids...)
	return suo
}

// RemoveExecutions removes "executions" edges to SignalExecution entities.
func (suo *SignalUpdateOne) RemoveExecutions(s ...*SignalExecution) *SignalUpdateOne {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return suo.RemoveExecutionIDs(ids...)
}

// Where appends a list predicates to the SignalUpdate builder.
func (suo *SignalUpdateOne) Where(ps ...predicate.Signal) *SignalUpdateOne {
	suo.mutation.Where(ps...)
	return suo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (suo *SignalUpdateOne) Select(field string, fields ...string) *SignalUpdateOne {
	suo.fields = append([]string{field}, fields...)
	return suo
}

// Save executes the query and returns the updated Signal entity.
func (suo *SignalUpdateOne) Save(ctx context.Context) (*Signal, error) {
	return withHooks(ctx, suo.sqlSave, suo.mutation, suo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (suo *SignalUpdateOne) SaveX(ctx context.Context) *Signal {
	node, err := suo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (suo *SignalUpdateOne) Exec(ctx context.Context) error {
	_, err := suo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (suo *SignalUpdateOne) ExecX(ctx context.Context) {
	if err := suo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (suo *SignalUpdateOne) check() error {
	if v, ok := suo.mutation.StepID(); ok {
		if err := signal.StepIDValidator(v); err != nil {
			return &ValidationError{Name: "step_id", err: fmt.Errorf(`ent: validator failed for field "Signal.step_id": %w`, err)}
		}
	}
	if v, ok := suo.mutation.Status(); ok {
		if err := signal.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Signal.status": %w`, err)}
		}
	}
	return nil
}

func (suo *SignalUpdateOne) sqlSave(ctx context.Context) (_node *Signal, err error) {
	if err := suo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(signal.Table, signal.Columns, sqlgraph.NewFieldSpec(signal.FieldID, field.TypeString))
	id, ok := suo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Signal.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := suo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, signal.FieldID)
		for _, f := range fields {
			if !signal.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != signal.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := suo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := suo.mutation.StepID(); ok {
		_spec.SetField(signal.FieldStepID, field.TypeString, value)
	}
	if value, ok := suo.mutation.Status(); ok {
		_spec.SetField(signal.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := suo.mutation.CreatedAt(); ok {
		_spec.SetField(signal.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := suo.mutation.Consumed(); ok {
		_spec.SetField(signal.FieldConsumed, field.TypeBool, value)
	}
	if suo.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := suo.mutation.RemovedExecutionsIDs(); len(nodes) > 0 && !suo.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := suo.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   signal.ExecutionsTable,
			Columns: []string{signal.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Signal{config: suo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, suo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{signal.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	suo.mutation.done = true
	return _node, nil
}
