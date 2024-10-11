// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
)

// SideEffectExecutionCreate is the builder for creating a SideEffectExecution entity.
type SideEffectExecutionCreate struct {
	config
	mutation *SideEffectExecutionMutation
	hooks    []Hook
}

// SetRunID sets the "run_id" field.
func (seec *SideEffectExecutionCreate) SetRunID(s string) *SideEffectExecutionCreate {
	seec.mutation.SetRunID(s)
	return seec
}

// SetStatus sets the "status" field.
func (seec *SideEffectExecutionCreate) SetStatus(s sideeffectexecution.Status) *SideEffectExecutionCreate {
	seec.mutation.SetStatus(s)
	return seec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (seec *SideEffectExecutionCreate) SetNillableStatus(s *sideeffectexecution.Status) *SideEffectExecutionCreate {
	if s != nil {
		seec.SetStatus(*s)
	}
	return seec
}

// SetAttempt sets the "attempt" field.
func (seec *SideEffectExecutionCreate) SetAttempt(i int) *SideEffectExecutionCreate {
	seec.mutation.SetAttempt(i)
	return seec
}

// SetNillableAttempt sets the "attempt" field if the given value is not nil.
func (seec *SideEffectExecutionCreate) SetNillableAttempt(i *int) *SideEffectExecutionCreate {
	if i != nil {
		seec.SetAttempt(*i)
	}
	return seec
}

// SetOutput sets the "output" field.
func (seec *SideEffectExecutionCreate) SetOutput(i []interface{}) *SideEffectExecutionCreate {
	seec.mutation.SetOutput(i)
	return seec
}

// SetStartedAt sets the "started_at" field.
func (seec *SideEffectExecutionCreate) SetStartedAt(t time.Time) *SideEffectExecutionCreate {
	seec.mutation.SetStartedAt(t)
	return seec
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (seec *SideEffectExecutionCreate) SetNillableStartedAt(t *time.Time) *SideEffectExecutionCreate {
	if t != nil {
		seec.SetStartedAt(*t)
	}
	return seec
}

// SetUpdatedAt sets the "updated_at" field.
func (seec *SideEffectExecutionCreate) SetUpdatedAt(t time.Time) *SideEffectExecutionCreate {
	seec.mutation.SetUpdatedAt(t)
	return seec
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (seec *SideEffectExecutionCreate) SetNillableUpdatedAt(t *time.Time) *SideEffectExecutionCreate {
	if t != nil {
		seec.SetUpdatedAt(*t)
	}
	return seec
}

// SetID sets the "id" field.
func (seec *SideEffectExecutionCreate) SetID(s string) *SideEffectExecutionCreate {
	seec.mutation.SetID(s)
	return seec
}

// SetSideEffectID sets the "side_effect" edge to the SideEffect entity by ID.
func (seec *SideEffectExecutionCreate) SetSideEffectID(id string) *SideEffectExecutionCreate {
	seec.mutation.SetSideEffectID(id)
	return seec
}

// SetSideEffect sets the "side_effect" edge to the SideEffect entity.
func (seec *SideEffectExecutionCreate) SetSideEffect(s *SideEffect) *SideEffectExecutionCreate {
	return seec.SetSideEffectID(s.ID)
}

// Mutation returns the SideEffectExecutionMutation object of the builder.
func (seec *SideEffectExecutionCreate) Mutation() *SideEffectExecutionMutation {
	return seec.mutation
}

// Save creates the SideEffectExecution in the database.
func (seec *SideEffectExecutionCreate) Save(ctx context.Context) (*SideEffectExecution, error) {
	seec.defaults()
	return withHooks(ctx, seec.sqlSave, seec.mutation, seec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (seec *SideEffectExecutionCreate) SaveX(ctx context.Context) *SideEffectExecution {
	v, err := seec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (seec *SideEffectExecutionCreate) Exec(ctx context.Context) error {
	_, err := seec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seec *SideEffectExecutionCreate) ExecX(ctx context.Context) {
	if err := seec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (seec *SideEffectExecutionCreate) defaults() {
	if _, ok := seec.mutation.Status(); !ok {
		v := sideeffectexecution.DefaultStatus
		seec.mutation.SetStatus(v)
	}
	if _, ok := seec.mutation.Attempt(); !ok {
		v := sideeffectexecution.DefaultAttempt
		seec.mutation.SetAttempt(v)
	}
	if _, ok := seec.mutation.StartedAt(); !ok {
		v := sideeffectexecution.DefaultStartedAt()
		seec.mutation.SetStartedAt(v)
	}
	if _, ok := seec.mutation.UpdatedAt(); !ok {
		v := sideeffectexecution.DefaultUpdatedAt()
		seec.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seec *SideEffectExecutionCreate) check() error {
	if _, ok := seec.mutation.RunID(); !ok {
		return &ValidationError{Name: "run_id", err: errors.New(`ent: missing required field "SideEffectExecution.run_id"`)}
	}
	if _, ok := seec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "SideEffectExecution.status"`)}
	}
	if v, ok := seec.mutation.Status(); ok {
		if err := sideeffectexecution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SideEffectExecution.status": %w`, err)}
		}
	}
	if _, ok := seec.mutation.Attempt(); !ok {
		return &ValidationError{Name: "attempt", err: errors.New(`ent: missing required field "SideEffectExecution.attempt"`)}
	}
	if _, ok := seec.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "SideEffectExecution.started_at"`)}
	}
	if _, ok := seec.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "SideEffectExecution.updated_at"`)}
	}
	if len(seec.mutation.SideEffectIDs()) == 0 {
		return &ValidationError{Name: "side_effect", err: errors.New(`ent: missing required edge "SideEffectExecution.side_effect"`)}
	}
	return nil
}

func (seec *SideEffectExecutionCreate) sqlSave(ctx context.Context) (*SideEffectExecution, error) {
	if err := seec.check(); err != nil {
		return nil, err
	}
	_node, _spec := seec.createSpec()
	if err := sqlgraph.CreateNode(ctx, seec.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected SideEffectExecution.ID type: %T", _spec.ID.Value)
		}
	}
	seec.mutation.id = &_node.ID
	seec.mutation.done = true
	return _node, nil
}

func (seec *SideEffectExecutionCreate) createSpec() (*SideEffectExecution, *sqlgraph.CreateSpec) {
	var (
		_node = &SideEffectExecution{config: seec.config}
		_spec = sqlgraph.NewCreateSpec(sideeffectexecution.Table, sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeString))
	)
	if id, ok := seec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := seec.mutation.RunID(); ok {
		_spec.SetField(sideeffectexecution.FieldRunID, field.TypeString, value)
		_node.RunID = value
	}
	if value, ok := seec.mutation.Status(); ok {
		_spec.SetField(sideeffectexecution.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := seec.mutation.Attempt(); ok {
		_spec.SetField(sideeffectexecution.FieldAttempt, field.TypeInt, value)
		_node.Attempt = value
	}
	if value, ok := seec.mutation.Output(); ok {
		_spec.SetField(sideeffectexecution.FieldOutput, field.TypeJSON, value)
		_node.Output = value
	}
	if value, ok := seec.mutation.StartedAt(); ok {
		_spec.SetField(sideeffectexecution.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := seec.mutation.UpdatedAt(); ok {
		_spec.SetField(sideeffectexecution.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := seec.mutation.SideEffectIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sideeffectexecution.SideEffectTable,
			Columns: []string{sideeffectexecution.SideEffectColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffect.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.side_effect_executions = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SideEffectExecutionCreateBulk is the builder for creating many SideEffectExecution entities in bulk.
type SideEffectExecutionCreateBulk struct {
	config
	err      error
	builders []*SideEffectExecutionCreate
}

// Save creates the SideEffectExecution entities in the database.
func (seecb *SideEffectExecutionCreateBulk) Save(ctx context.Context) ([]*SideEffectExecution, error) {
	if seecb.err != nil {
		return nil, seecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(seecb.builders))
	nodes := make([]*SideEffectExecution, len(seecb.builders))
	mutators := make([]Mutator, len(seecb.builders))
	for i := range seecb.builders {
		func(i int, root context.Context) {
			builder := seecb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SideEffectExecutionMutation)
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
					_, err = mutators[i+1].Mutate(root, seecb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, seecb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, seecb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (seecb *SideEffectExecutionCreateBulk) SaveX(ctx context.Context) []*SideEffectExecution {
	v, err := seecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (seecb *SideEffectExecutionCreateBulk) Exec(ctx context.Context) error {
	_, err := seecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seecb *SideEffectExecutionCreateBulk) ExecX(ctx context.Context) {
	if err := seecb.Exec(ctx); err != nil {
		panic(err)
	}
}
