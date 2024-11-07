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

// SignalExecutionCreate is the builder for creating a SignalExecution entity.
type SignalExecutionCreate struct {
	config
	mutation *SignalExecutionMutation
	hooks    []Hook
}

// SetRunID sets the "run_id" field.
func (sec *SignalExecutionCreate) SetRunID(s string) *SignalExecutionCreate {
	sec.mutation.SetRunID(s)
	return sec
}

// SetStatus sets the "status" field.
func (sec *SignalExecutionCreate) SetStatus(s signalexecution.Status) *SignalExecutionCreate {
	sec.mutation.SetStatus(s)
	return sec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (sec *SignalExecutionCreate) SetNillableStatus(s *signalexecution.Status) *SignalExecutionCreate {
	if s != nil {
		sec.SetStatus(*s)
	}
	return sec
}

// SetQueueName sets the "queue_name" field.
func (sec *SignalExecutionCreate) SetQueueName(s string) *SignalExecutionCreate {
	sec.mutation.SetQueueName(s)
	return sec
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (sec *SignalExecutionCreate) SetNillableQueueName(s *string) *SignalExecutionCreate {
	if s != nil {
		sec.SetQueueName(*s)
	}
	return sec
}

// SetOutput sets the "output" field.
func (sec *SignalExecutionCreate) SetOutput(u [][]uint8) *SignalExecutionCreate {
	sec.mutation.SetOutput(u)
	return sec
}

// SetError sets the "error" field.
func (sec *SignalExecutionCreate) SetError(s string) *SignalExecutionCreate {
	sec.mutation.SetError(s)
	return sec
}

// SetNillableError sets the "error" field if the given value is not nil.
func (sec *SignalExecutionCreate) SetNillableError(s *string) *SignalExecutionCreate {
	if s != nil {
		sec.SetError(*s)
	}
	return sec
}

// SetStartedAt sets the "started_at" field.
func (sec *SignalExecutionCreate) SetStartedAt(t time.Time) *SignalExecutionCreate {
	sec.mutation.SetStartedAt(t)
	return sec
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (sec *SignalExecutionCreate) SetNillableStartedAt(t *time.Time) *SignalExecutionCreate {
	if t != nil {
		sec.SetStartedAt(*t)
	}
	return sec
}

// SetUpdatedAt sets the "updated_at" field.
func (sec *SignalExecutionCreate) SetUpdatedAt(t time.Time) *SignalExecutionCreate {
	sec.mutation.SetUpdatedAt(t)
	return sec
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (sec *SignalExecutionCreate) SetNillableUpdatedAt(t *time.Time) *SignalExecutionCreate {
	if t != nil {
		sec.SetUpdatedAt(*t)
	}
	return sec
}

// SetID sets the "id" field.
func (sec *SignalExecutionCreate) SetID(s string) *SignalExecutionCreate {
	sec.mutation.SetID(s)
	return sec
}

// SetSignalID sets the "signal" edge to the Signal entity by ID.
func (sec *SignalExecutionCreate) SetSignalID(id string) *SignalExecutionCreate {
	sec.mutation.SetSignalID(id)
	return sec
}

// SetSignal sets the "signal" edge to the Signal entity.
func (sec *SignalExecutionCreate) SetSignal(s *Signal) *SignalExecutionCreate {
	return sec.SetSignalID(s.ID)
}

// Mutation returns the SignalExecutionMutation object of the builder.
func (sec *SignalExecutionCreate) Mutation() *SignalExecutionMutation {
	return sec.mutation
}

// Save creates the SignalExecution in the database.
func (sec *SignalExecutionCreate) Save(ctx context.Context) (*SignalExecution, error) {
	sec.defaults()
	return withHooks(ctx, sec.sqlSave, sec.mutation, sec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sec *SignalExecutionCreate) SaveX(ctx context.Context) *SignalExecution {
	v, err := sec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sec *SignalExecutionCreate) Exec(ctx context.Context) error {
	_, err := sec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sec *SignalExecutionCreate) ExecX(ctx context.Context) {
	if err := sec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sec *SignalExecutionCreate) defaults() {
	if _, ok := sec.mutation.Status(); !ok {
		v := signalexecution.DefaultStatus
		sec.mutation.SetStatus(v)
	}
	if _, ok := sec.mutation.QueueName(); !ok {
		v := signalexecution.DefaultQueueName
		sec.mutation.SetQueueName(v)
	}
	if _, ok := sec.mutation.StartedAt(); !ok {
		v := signalexecution.DefaultStartedAt()
		sec.mutation.SetStartedAt(v)
	}
	if _, ok := sec.mutation.UpdatedAt(); !ok {
		v := signalexecution.DefaultUpdatedAt()
		sec.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sec *SignalExecutionCreate) check() error {
	if _, ok := sec.mutation.RunID(); !ok {
		return &ValidationError{Name: "run_id", err: errors.New(`ent: missing required field "SignalExecution.run_id"`)}
	}
	if _, ok := sec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "SignalExecution.status"`)}
	}
	if v, ok := sec.mutation.Status(); ok {
		if err := signalexecution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SignalExecution.status": %w`, err)}
		}
	}
	if _, ok := sec.mutation.QueueName(); !ok {
		return &ValidationError{Name: "queue_name", err: errors.New(`ent: missing required field "SignalExecution.queue_name"`)}
	}
	if v, ok := sec.mutation.QueueName(); ok {
		if err := signalexecution.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "SignalExecution.queue_name": %w`, err)}
		}
	}
	if _, ok := sec.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "SignalExecution.started_at"`)}
	}
	if _, ok := sec.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "SignalExecution.updated_at"`)}
	}
	if len(sec.mutation.SignalIDs()) == 0 {
		return &ValidationError{Name: "signal", err: errors.New(`ent: missing required edge "SignalExecution.signal"`)}
	}
	return nil
}

func (sec *SignalExecutionCreate) sqlSave(ctx context.Context) (*SignalExecution, error) {
	if err := sec.check(); err != nil {
		return nil, err
	}
	_node, _spec := sec.createSpec()
	if err := sqlgraph.CreateNode(ctx, sec.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected SignalExecution.ID type: %T", _spec.ID.Value)
		}
	}
	sec.mutation.id = &_node.ID
	sec.mutation.done = true
	return _node, nil
}

func (sec *SignalExecutionCreate) createSpec() (*SignalExecution, *sqlgraph.CreateSpec) {
	var (
		_node = &SignalExecution{config: sec.config}
		_spec = sqlgraph.NewCreateSpec(signalexecution.Table, sqlgraph.NewFieldSpec(signalexecution.FieldID, field.TypeString))
	)
	if id, ok := sec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := sec.mutation.RunID(); ok {
		_spec.SetField(signalexecution.FieldRunID, field.TypeString, value)
		_node.RunID = value
	}
	if value, ok := sec.mutation.Status(); ok {
		_spec.SetField(signalexecution.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := sec.mutation.QueueName(); ok {
		_spec.SetField(signalexecution.FieldQueueName, field.TypeString, value)
		_node.QueueName = value
	}
	if value, ok := sec.mutation.Output(); ok {
		_spec.SetField(signalexecution.FieldOutput, field.TypeJSON, value)
		_node.Output = value
	}
	if value, ok := sec.mutation.Error(); ok {
		_spec.SetField(signalexecution.FieldError, field.TypeString, value)
		_node.Error = value
	}
	if value, ok := sec.mutation.StartedAt(); ok {
		_spec.SetField(signalexecution.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := sec.mutation.UpdatedAt(); ok {
		_spec.SetField(signalexecution.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := sec.mutation.SignalIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   signalexecution.SignalTable,
			Columns: []string{signalexecution.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(signal.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.signal_executions = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SignalExecutionCreateBulk is the builder for creating many SignalExecution entities in bulk.
type SignalExecutionCreateBulk struct {
	config
	err      error
	builders []*SignalExecutionCreate
}

// Save creates the SignalExecution entities in the database.
func (secb *SignalExecutionCreateBulk) Save(ctx context.Context) ([]*SignalExecution, error) {
	if secb.err != nil {
		return nil, secb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(secb.builders))
	nodes := make([]*SignalExecution, len(secb.builders))
	mutators := make([]Mutator, len(secb.builders))
	for i := range secb.builders {
		func(i int, root context.Context) {
			builder := secb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SignalExecutionMutation)
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
					_, err = mutators[i+1].Mutate(root, secb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, secb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, secb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (secb *SignalExecutionCreateBulk) SaveX(ctx context.Context) []*SignalExecution {
	v, err := secb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (secb *SignalExecutionCreateBulk) Exec(ctx context.Context) error {
	_, err := secb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (secb *SignalExecutionCreateBulk) ExecX(ctx context.Context) {
	if err := secb.Exec(ctx); err != nil {
		panic(err)
	}
}