// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/signal"
	"github.com/davidroman0O/tempolite/ent/signalexecution"
)

// SignalCreate is the builder for creating a Signal entity.
type SignalCreate struct {
	config
	mutation *SignalMutation
	hooks    []Hook
}

// SetStepID sets the "step_id" field.
func (sc *SignalCreate) SetStepID(s string) *SignalCreate {
	sc.mutation.SetStepID(s)
	return sc
}

// SetStatus sets the "status" field.
func (sc *SignalCreate) SetStatus(s signal.Status) *SignalCreate {
	sc.mutation.SetStatus(s)
	return sc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (sc *SignalCreate) SetNillableStatus(s *signal.Status) *SignalCreate {
	if s != nil {
		sc.SetStatus(*s)
	}
	return sc
}

// SetQueueName sets the "queue_name" field.
func (sc *SignalCreate) SetQueueName(s string) *SignalCreate {
	sc.mutation.SetQueueName(s)
	return sc
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (sc *SignalCreate) SetNillableQueueName(s *string) *SignalCreate {
	if s != nil {
		sc.SetQueueName(*s)
	}
	return sc
}

// SetCreatedAt sets the "created_at" field.
func (sc *SignalCreate) SetCreatedAt(t time.Time) *SignalCreate {
	sc.mutation.SetCreatedAt(t)
	return sc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sc *SignalCreate) SetNillableCreatedAt(t *time.Time) *SignalCreate {
	if t != nil {
		sc.SetCreatedAt(*t)
	}
	return sc
}

// SetConsumed sets the "consumed" field.
func (sc *SignalCreate) SetConsumed(b bool) *SignalCreate {
	sc.mutation.SetConsumed(b)
	return sc
}

// SetNillableConsumed sets the "consumed" field if the given value is not nil.
func (sc *SignalCreate) SetNillableConsumed(b *bool) *SignalCreate {
	if b != nil {
		sc.SetConsumed(*b)
	}
	return sc
}

// SetID sets the "id" field.
func (sc *SignalCreate) SetID(s string) *SignalCreate {
	sc.mutation.SetID(s)
	return sc
}

// AddExecutionIDs adds the "executions" edge to the SignalExecution entity by IDs.
func (sc *SignalCreate) AddExecutionIDs(ids ...string) *SignalCreate {
	sc.mutation.AddExecutionIDs(ids...)
	return sc
}

// AddExecutions adds the "executions" edges to the SignalExecution entity.
func (sc *SignalCreate) AddExecutions(s ...*SignalExecution) *SignalCreate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sc.AddExecutionIDs(ids...)
}

// Mutation returns the SignalMutation object of the builder.
func (sc *SignalCreate) Mutation() *SignalMutation {
	return sc.mutation
}

// Save creates the Signal in the database.
func (sc *SignalCreate) Save(ctx context.Context) (*Signal, error) {
	sc.defaults()
	return withHooks(ctx, sc.sqlSave, sc.mutation, sc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sc *SignalCreate) SaveX(ctx context.Context) *Signal {
	v, err := sc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sc *SignalCreate) Exec(ctx context.Context) error {
	_, err := sc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sc *SignalCreate) ExecX(ctx context.Context) {
	if err := sc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sc *SignalCreate) defaults() {
	if _, ok := sc.mutation.Status(); !ok {
		v := signal.DefaultStatus
		sc.mutation.SetStatus(v)
	}
	if _, ok := sc.mutation.QueueName(); !ok {
		v := signal.DefaultQueueName
		sc.mutation.SetQueueName(v)
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		v := signal.DefaultCreatedAt()
		sc.mutation.SetCreatedAt(v)
	}
	if _, ok := sc.mutation.Consumed(); !ok {
		v := signal.DefaultConsumed
		sc.mutation.SetConsumed(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sc *SignalCreate) check() error {
	if _, ok := sc.mutation.StepID(); !ok {
		return &ValidationError{Name: "step_id", err: errors.New(`ent: missing required field "Signal.step_id"`)}
	}
	if v, ok := sc.mutation.StepID(); ok {
		if err := signal.StepIDValidator(v); err != nil {
			return &ValidationError{Name: "step_id", err: fmt.Errorf(`ent: validator failed for field "Signal.step_id": %w`, err)}
		}
	}
	if _, ok := sc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Signal.status"`)}
	}
	if v, ok := sc.mutation.Status(); ok {
		if err := signal.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Signal.status": %w`, err)}
		}
	}
	if _, ok := sc.mutation.QueueName(); !ok {
		return &ValidationError{Name: "queue_name", err: errors.New(`ent: missing required field "Signal.queue_name"`)}
	}
	if v, ok := sc.mutation.QueueName(); ok {
		if err := signal.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "Signal.queue_name": %w`, err)}
		}
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Signal.created_at"`)}
	}
	if _, ok := sc.mutation.Consumed(); !ok {
		return &ValidationError{Name: "consumed", err: errors.New(`ent: missing required field "Signal.consumed"`)}
	}
	return nil
}

func (sc *SignalCreate) sqlSave(ctx context.Context) (*Signal, error) {
	if err := sc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Signal.ID type: %T", _spec.ID.Value)
		}
	}
	sc.mutation.id = &_node.ID
	sc.mutation.done = true
	return _node, nil
}

func (sc *SignalCreate) createSpec() (*Signal, *sqlgraph.CreateSpec) {
	var (
		_node = &Signal{config: sc.config}
		_spec = sqlgraph.NewCreateSpec(signal.Table, sqlgraph.NewFieldSpec(signal.FieldID, field.TypeString))
	)
	if id, ok := sc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := sc.mutation.StepID(); ok {
		_spec.SetField(signal.FieldStepID, field.TypeString, value)
		_node.StepID = value
	}
	if value, ok := sc.mutation.Status(); ok {
		_spec.SetField(signal.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := sc.mutation.QueueName(); ok {
		_spec.SetField(signal.FieldQueueName, field.TypeString, value)
		_node.QueueName = value
	}
	if value, ok := sc.mutation.CreatedAt(); ok {
		_spec.SetField(signal.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := sc.mutation.Consumed(); ok {
		_spec.SetField(signal.FieldConsumed, field.TypeBool, value)
		_node.Consumed = value
	}
	if nodes := sc.mutation.ExecutionsIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SignalCreateBulk is the builder for creating many Signal entities in bulk.
type SignalCreateBulk struct {
	config
	err      error
	builders []*SignalCreate
}

// Save creates the Signal entities in the database.
func (scb *SignalCreateBulk) Save(ctx context.Context) ([]*Signal, error) {
	if scb.err != nil {
		return nil, scb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scb.builders))
	nodes := make([]*Signal, len(scb.builders))
	mutators := make([]Mutator, len(scb.builders))
	for i := range scb.builders {
		func(i int, root context.Context) {
			builder := scb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SignalMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, scb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, scb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scb *SignalCreateBulk) SaveX(ctx context.Context) []*Signal {
	v, err := scb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scb *SignalCreateBulk) Exec(ctx context.Context) error {
	_, err := scb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scb *SignalCreateBulk) ExecX(ctx context.Context) {
	if err := scb.Exec(ctx); err != nil {
		panic(err)
	}
}
