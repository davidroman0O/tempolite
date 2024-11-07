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
	"github.com/davidroman0O/tempolite/ent/schema"
)

// SagaCreate is the builder for creating a Saga entity.
type SagaCreate struct {
	config
	mutation *SagaMutation
	hooks    []Hook
}

// SetRunID sets the "run_id" field.
func (sc *SagaCreate) SetRunID(s string) *SagaCreate {
	sc.mutation.SetRunID(s)
	return sc
}

// SetStepID sets the "step_id" field.
func (sc *SagaCreate) SetStepID(s string) *SagaCreate {
	sc.mutation.SetStepID(s)
	return sc
}

// SetStatus sets the "status" field.
func (sc *SagaCreate) SetStatus(s saga.Status) *SagaCreate {
	sc.mutation.SetStatus(s)
	return sc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (sc *SagaCreate) SetNillableStatus(s *saga.Status) *SagaCreate {
	if s != nil {
		sc.SetStatus(*s)
	}
	return sc
}

// SetQueueName sets the "queue_name" field.
func (sc *SagaCreate) SetQueueName(s string) *SagaCreate {
	sc.mutation.SetQueueName(s)
	return sc
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (sc *SagaCreate) SetNillableQueueName(s *string) *SagaCreate {
	if s != nil {
		sc.SetQueueName(*s)
	}
	return sc
}

// SetSagaDefinition sets the "saga_definition" field.
func (sc *SagaCreate) SetSagaDefinition(sdd schema.SagaDefinitionData) *SagaCreate {
	sc.mutation.SetSagaDefinition(sdd)
	return sc
}

// SetError sets the "error" field.
func (sc *SagaCreate) SetError(s string) *SagaCreate {
	sc.mutation.SetError(s)
	return sc
}

// SetNillableError sets the "error" field if the given value is not nil.
func (sc *SagaCreate) SetNillableError(s *string) *SagaCreate {
	if s != nil {
		sc.SetError(*s)
	}
	return sc
}

// SetCreatedAt sets the "created_at" field.
func (sc *SagaCreate) SetCreatedAt(t time.Time) *SagaCreate {
	sc.mutation.SetCreatedAt(t)
	return sc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sc *SagaCreate) SetNillableCreatedAt(t *time.Time) *SagaCreate {
	if t != nil {
		sc.SetCreatedAt(*t)
	}
	return sc
}

// SetUpdatedAt sets the "updated_at" field.
func (sc *SagaCreate) SetUpdatedAt(t time.Time) *SagaCreate {
	sc.mutation.SetUpdatedAt(t)
	return sc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (sc *SagaCreate) SetNillableUpdatedAt(t *time.Time) *SagaCreate {
	if t != nil {
		sc.SetUpdatedAt(*t)
	}
	return sc
}

// SetID sets the "id" field.
func (sc *SagaCreate) SetID(s string) *SagaCreate {
	sc.mutation.SetID(s)
	return sc
}

// AddStepIDs adds the "steps" edge to the SagaExecution entity by IDs.
func (sc *SagaCreate) AddStepIDs(ids ...string) *SagaCreate {
	sc.mutation.AddStepIDs(ids...)
	return sc
}

// AddSteps adds the "steps" edges to the SagaExecution entity.
func (sc *SagaCreate) AddSteps(s ...*SagaExecution) *SagaCreate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sc.AddStepIDs(ids...)
}

// Mutation returns the SagaMutation object of the builder.
func (sc *SagaCreate) Mutation() *SagaMutation {
	return sc.mutation
}

// Save creates the Saga in the database.
func (sc *SagaCreate) Save(ctx context.Context) (*Saga, error) {
	sc.defaults()
	return withHooks(ctx, sc.sqlSave, sc.mutation, sc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sc *SagaCreate) SaveX(ctx context.Context) *Saga {
	v, err := sc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sc *SagaCreate) Exec(ctx context.Context) error {
	_, err := sc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sc *SagaCreate) ExecX(ctx context.Context) {
	if err := sc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sc *SagaCreate) defaults() {
	if _, ok := sc.mutation.Status(); !ok {
		v := saga.DefaultStatus
		sc.mutation.SetStatus(v)
	}
	if _, ok := sc.mutation.QueueName(); !ok {
		v := saga.DefaultQueueName
		sc.mutation.SetQueueName(v)
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		v := saga.DefaultCreatedAt()
		sc.mutation.SetCreatedAt(v)
	}
	if _, ok := sc.mutation.UpdatedAt(); !ok {
		v := saga.DefaultUpdatedAt()
		sc.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sc *SagaCreate) check() error {
	if _, ok := sc.mutation.RunID(); !ok {
		return &ValidationError{Name: "run_id", err: errors.New(`ent: missing required field "Saga.run_id"`)}
	}
	if _, ok := sc.mutation.StepID(); !ok {
		return &ValidationError{Name: "step_id", err: errors.New(`ent: missing required field "Saga.step_id"`)}
	}
	if v, ok := sc.mutation.StepID(); ok {
		if err := saga.StepIDValidator(v); err != nil {
			return &ValidationError{Name: "step_id", err: fmt.Errorf(`ent: validator failed for field "Saga.step_id": %w`, err)}
		}
	}
	if _, ok := sc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Saga.status"`)}
	}
	if v, ok := sc.mutation.Status(); ok {
		if err := saga.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Saga.status": %w`, err)}
		}
	}
	if _, ok := sc.mutation.QueueName(); !ok {
		return &ValidationError{Name: "queue_name", err: errors.New(`ent: missing required field "Saga.queue_name"`)}
	}
	if v, ok := sc.mutation.QueueName(); ok {
		if err := saga.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "Saga.queue_name": %w`, err)}
		}
	}
	if _, ok := sc.mutation.SagaDefinition(); !ok {
		return &ValidationError{Name: "saga_definition", err: errors.New(`ent: missing required field "Saga.saga_definition"`)}
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Saga.created_at"`)}
	}
	if _, ok := sc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Saga.updated_at"`)}
	}
	return nil
}

func (sc *SagaCreate) sqlSave(ctx context.Context) (*Saga, error) {
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
			return nil, fmt.Errorf("unexpected Saga.ID type: %T", _spec.ID.Value)
		}
	}
	sc.mutation.id = &_node.ID
	sc.mutation.done = true
	return _node, nil
}

func (sc *SagaCreate) createSpec() (*Saga, *sqlgraph.CreateSpec) {
	var (
		_node = &Saga{config: sc.config}
		_spec = sqlgraph.NewCreateSpec(saga.Table, sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString))
	)
	if id, ok := sc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := sc.mutation.RunID(); ok {
		_spec.SetField(saga.FieldRunID, field.TypeString, value)
		_node.RunID = value
	}
	if value, ok := sc.mutation.StepID(); ok {
		_spec.SetField(saga.FieldStepID, field.TypeString, value)
		_node.StepID = value
	}
	if value, ok := sc.mutation.Status(); ok {
		_spec.SetField(saga.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := sc.mutation.QueueName(); ok {
		_spec.SetField(saga.FieldQueueName, field.TypeString, value)
		_node.QueueName = value
	}
	if value, ok := sc.mutation.SagaDefinition(); ok {
		_spec.SetField(saga.FieldSagaDefinition, field.TypeJSON, value)
		_node.SagaDefinition = value
	}
	if value, ok := sc.mutation.Error(); ok {
		_spec.SetField(saga.FieldError, field.TypeString, value)
		_node.Error = value
	}
	if value, ok := sc.mutation.CreatedAt(); ok {
		_spec.SetField(saga.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := sc.mutation.UpdatedAt(); ok {
		_spec.SetField(saga.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := sc.mutation.StepsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   saga.StepsTable,
			Columns: []string{saga.StepsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SagaCreateBulk is the builder for creating many Saga entities in bulk.
type SagaCreateBulk struct {
	config
	err      error
	builders []*SagaCreate
}

// Save creates the Saga entities in the database.
func (scb *SagaCreateBulk) Save(ctx context.Context) ([]*Saga, error) {
	if scb.err != nil {
		return nil, scb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scb.builders))
	nodes := make([]*Saga, len(scb.builders))
	mutators := make([]Mutator, len(scb.builders))
	for i := range scb.builders {
		func(i int, root context.Context) {
			builder := scb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SagaMutation)
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
func (scb *SagaCreateBulk) SaveX(ctx context.Context) []*Saga {
	v, err := scb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scb *SagaCreateBulk) Exec(ctx context.Context) error {
	_, err := scb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scb *SagaCreateBulk) ExecX(ctx context.Context) {
	if err := scb.Exec(ctx); err != nil {
		panic(err)
	}
}