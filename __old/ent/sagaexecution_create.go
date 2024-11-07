// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/sagaexecution"
)

// SagaExecutionCreate is the builder for creating a SagaExecution entity.
type SagaExecutionCreate struct {
	config
	mutation *SagaExecutionMutation
	hooks    []Hook
}

// SetHandlerName sets the "handler_name" field.
func (sec *SagaExecutionCreate) SetHandlerName(s string) *SagaExecutionCreate {
	sec.mutation.SetHandlerName(s)
	return sec
}

// SetStepType sets the "step_type" field.
func (sec *SagaExecutionCreate) SetStepType(st sagaexecution.StepType) *SagaExecutionCreate {
	sec.mutation.SetStepType(st)
	return sec
}

// SetStatus sets the "status" field.
func (sec *SagaExecutionCreate) SetStatus(s sagaexecution.Status) *SagaExecutionCreate {
	sec.mutation.SetStatus(s)
	return sec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (sec *SagaExecutionCreate) SetNillableStatus(s *sagaexecution.Status) *SagaExecutionCreate {
	if s != nil {
		sec.SetStatus(*s)
	}
	return sec
}

// SetQueueName sets the "queue_name" field.
func (sec *SagaExecutionCreate) SetQueueName(s string) *SagaExecutionCreate {
	sec.mutation.SetQueueName(s)
	return sec
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (sec *SagaExecutionCreate) SetNillableQueueName(s *string) *SagaExecutionCreate {
	if s != nil {
		sec.SetQueueName(*s)
	}
	return sec
}

// SetSequence sets the "sequence" field.
func (sec *SagaExecutionCreate) SetSequence(i int) *SagaExecutionCreate {
	sec.mutation.SetSequence(i)
	return sec
}

// SetError sets the "error" field.
func (sec *SagaExecutionCreate) SetError(s string) *SagaExecutionCreate {
	sec.mutation.SetError(s)
	return sec
}

// SetNillableError sets the "error" field if the given value is not nil.
func (sec *SagaExecutionCreate) SetNillableError(s *string) *SagaExecutionCreate {
	if s != nil {
		sec.SetError(*s)
	}
	return sec
}

// SetStartedAt sets the "started_at" field.
func (sec *SagaExecutionCreate) SetStartedAt(t time.Time) *SagaExecutionCreate {
	sec.mutation.SetStartedAt(t)
	return sec
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (sec *SagaExecutionCreate) SetNillableStartedAt(t *time.Time) *SagaExecutionCreate {
	if t != nil {
		sec.SetStartedAt(*t)
	}
	return sec
}

// SetCompletedAt sets the "completed_at" field.
func (sec *SagaExecutionCreate) SetCompletedAt(t time.Time) *SagaExecutionCreate {
	sec.mutation.SetCompletedAt(t)
	return sec
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (sec *SagaExecutionCreate) SetNillableCompletedAt(t *time.Time) *SagaExecutionCreate {
	if t != nil {
		sec.SetCompletedAt(*t)
	}
	return sec
}

// SetID sets the "id" field.
func (sec *SagaExecutionCreate) SetID(s string) *SagaExecutionCreate {
	sec.mutation.SetID(s)
	return sec
}

// SetSagaID sets the "saga" edge to the Saga entity by ID.
func (sec *SagaExecutionCreate) SetSagaID(id string) *SagaExecutionCreate {
	sec.mutation.SetSagaID(id)
	return sec
}

// SetSaga sets the "saga" edge to the Saga entity.
func (sec *SagaExecutionCreate) SetSaga(s *Saga) *SagaExecutionCreate {
	return sec.SetSagaID(s.ID)
}

// Mutation returns the SagaExecutionMutation object of the builder.
func (sec *SagaExecutionCreate) Mutation() *SagaExecutionMutation {
	return sec.mutation
}

// Save creates the SagaExecution in the database.
func (sec *SagaExecutionCreate) Save(ctx context.Context) (*SagaExecution, error) {
	sec.defaults()
	return withHooks(ctx, sec.sqlSave, sec.mutation, sec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sec *SagaExecutionCreate) SaveX(ctx context.Context) *SagaExecution {
	v, err := sec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sec *SagaExecutionCreate) Exec(ctx context.Context) error {
	_, err := sec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sec *SagaExecutionCreate) ExecX(ctx context.Context) {
	if err := sec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sec *SagaExecutionCreate) defaults() {
	if _, ok := sec.mutation.Status(); !ok {
		v := sagaexecution.DefaultStatus
		sec.mutation.SetStatus(v)
	}
	if _, ok := sec.mutation.QueueName(); !ok {
		v := sagaexecution.DefaultQueueName
		sec.mutation.SetQueueName(v)
	}
	if _, ok := sec.mutation.StartedAt(); !ok {
		v := sagaexecution.DefaultStartedAt()
		sec.mutation.SetStartedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sec *SagaExecutionCreate) check() error {
	if _, ok := sec.mutation.HandlerName(); !ok {
		return &ValidationError{Name: "handler_name", err: errors.New(`ent: missing required field "SagaExecution.handler_name"`)}
	}
	if v, ok := sec.mutation.HandlerName(); ok {
		if err := sagaexecution.HandlerNameValidator(v); err != nil {
			return &ValidationError{Name: "handler_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.handler_name": %w`, err)}
		}
	}
	if _, ok := sec.mutation.StepType(); !ok {
		return &ValidationError{Name: "step_type", err: errors.New(`ent: missing required field "SagaExecution.step_type"`)}
	}
	if v, ok := sec.mutation.StepType(); ok {
		if err := sagaexecution.StepTypeValidator(v); err != nil {
			return &ValidationError{Name: "step_type", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.step_type": %w`, err)}
		}
	}
	if _, ok := sec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "SagaExecution.status"`)}
	}
	if v, ok := sec.mutation.Status(); ok {
		if err := sagaexecution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.status": %w`, err)}
		}
	}
	if _, ok := sec.mutation.QueueName(); !ok {
		return &ValidationError{Name: "queue_name", err: errors.New(`ent: missing required field "SagaExecution.queue_name"`)}
	}
	if v, ok := sec.mutation.QueueName(); ok {
		if err := sagaexecution.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.queue_name": %w`, err)}
		}
	}
	if _, ok := sec.mutation.Sequence(); !ok {
		return &ValidationError{Name: "sequence", err: errors.New(`ent: missing required field "SagaExecution.sequence"`)}
	}
	if v, ok := sec.mutation.Sequence(); ok {
		if err := sagaexecution.SequenceValidator(v); err != nil {
			return &ValidationError{Name: "sequence", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.sequence": %w`, err)}
		}
	}
	if _, ok := sec.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "SagaExecution.started_at"`)}
	}
	if len(sec.mutation.SagaIDs()) == 0 {
		return &ValidationError{Name: "saga", err: errors.New(`ent: missing required edge "SagaExecution.saga"`)}
	}
	return nil
}

func (sec *SagaExecutionCreate) sqlSave(ctx context.Context) (*SagaExecution, error) {
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
			return nil, fmt.Errorf("unexpected SagaExecution.ID type: %T", _spec.ID.Value)
		}
	}
	sec.mutation.id = &_node.ID
	sec.mutation.done = true
	return _node, nil
}

func (sec *SagaExecutionCreate) createSpec() (*SagaExecution, *sqlgraph.CreateSpec) {
	var (
		_node = &SagaExecution{config: sec.config}
		_spec = sqlgraph.NewCreateSpec(sagaexecution.Table, sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeString))
	)
	if id, ok := sec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := sec.mutation.HandlerName(); ok {
		_spec.SetField(sagaexecution.FieldHandlerName, field.TypeString, value)
		_node.HandlerName = value
	}
	if value, ok := sec.mutation.StepType(); ok {
		_spec.SetField(sagaexecution.FieldStepType, field.TypeEnum, value)
		_node.StepType = value
	}
	if value, ok := sec.mutation.Status(); ok {
		_spec.SetField(sagaexecution.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := sec.mutation.QueueName(); ok {
		_spec.SetField(sagaexecution.FieldQueueName, field.TypeString, value)
		_node.QueueName = value
	}
	if value, ok := sec.mutation.Sequence(); ok {
		_spec.SetField(sagaexecution.FieldSequence, field.TypeInt, value)
		_node.Sequence = value
	}
	if value, ok := sec.mutation.Error(); ok {
		_spec.SetField(sagaexecution.FieldError, field.TypeString, value)
		_node.Error = value
	}
	if value, ok := sec.mutation.StartedAt(); ok {
		_spec.SetField(sagaexecution.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := sec.mutation.CompletedAt(); ok {
		_spec.SetField(sagaexecution.FieldCompletedAt, field.TypeTime, value)
		_node.CompletedAt = value
	}
	if nodes := sec.mutation.SagaIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sagaexecution.SagaTable,
			Columns: []string{sagaexecution.SagaColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.saga_steps = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SagaExecutionCreateBulk is the builder for creating many SagaExecution entities in bulk.
type SagaExecutionCreateBulk struct {
	config
	err      error
	builders []*SagaExecutionCreate
}

// Save creates the SagaExecution entities in the database.
func (secb *SagaExecutionCreateBulk) Save(ctx context.Context) ([]*SagaExecution, error) {
	if secb.err != nil {
		return nil, secb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(secb.builders))
	nodes := make([]*SagaExecution, len(secb.builders))
	mutators := make([]Mutator, len(secb.builders))
	for i := range secb.builders {
		func(i int, root context.Context) {
			builder := secb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SagaExecutionMutation)
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
func (secb *SagaExecutionCreateBulk) SaveX(ctx context.Context) []*SagaExecution {
	v, err := secb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (secb *SagaExecutionCreateBulk) Exec(ctx context.Context) error {
	_, err := secb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (secb *SagaExecutionCreateBulk) ExecX(ctx context.Context) {
	if err := secb.Exec(ctx); err != nil {
		panic(err)
	}
}