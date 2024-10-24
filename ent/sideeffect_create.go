// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

// SideEffectCreate is the builder for creating a SideEffect entity.
type SideEffectCreate struct {
	config
	mutation *SideEffectMutation
	hooks    []Hook
}

// SetIdentity sets the "identity" field.
func (sec *SideEffectCreate) SetIdentity(s string) *SideEffectCreate {
	sec.mutation.SetIdentity(s)
	return sec
}

// SetStepID sets the "step_id" field.
func (sec *SideEffectCreate) SetStepID(s string) *SideEffectCreate {
	sec.mutation.SetStepID(s)
	return sec
}

// SetHandlerName sets the "handler_name" field.
func (sec *SideEffectCreate) SetHandlerName(s string) *SideEffectCreate {
	sec.mutation.SetHandlerName(s)
	return sec
}

// SetStatus sets the "status" field.
func (sec *SideEffectCreate) SetStatus(s sideeffect.Status) *SideEffectCreate {
	sec.mutation.SetStatus(s)
	return sec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (sec *SideEffectCreate) SetNillableStatus(s *sideeffect.Status) *SideEffectCreate {
	if s != nil {
		sec.SetStatus(*s)
	}
	return sec
}

// SetRetryPolicy sets the "retry_policy" field.
func (sec *SideEffectCreate) SetRetryPolicy(sp schema.RetryPolicy) *SideEffectCreate {
	sec.mutation.SetRetryPolicy(sp)
	return sec
}

// SetNillableRetryPolicy sets the "retry_policy" field if the given value is not nil.
func (sec *SideEffectCreate) SetNillableRetryPolicy(sp *schema.RetryPolicy) *SideEffectCreate {
	if sp != nil {
		sec.SetRetryPolicy(*sp)
	}
	return sec
}

// SetTimeout sets the "timeout" field.
func (sec *SideEffectCreate) SetTimeout(t time.Time) *SideEffectCreate {
	sec.mutation.SetTimeout(t)
	return sec
}

// SetNillableTimeout sets the "timeout" field if the given value is not nil.
func (sec *SideEffectCreate) SetNillableTimeout(t *time.Time) *SideEffectCreate {
	if t != nil {
		sec.SetTimeout(*t)
	}
	return sec
}

// SetCreatedAt sets the "created_at" field.
func (sec *SideEffectCreate) SetCreatedAt(t time.Time) *SideEffectCreate {
	sec.mutation.SetCreatedAt(t)
	return sec
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sec *SideEffectCreate) SetNillableCreatedAt(t *time.Time) *SideEffectCreate {
	if t != nil {
		sec.SetCreatedAt(*t)
	}
	return sec
}

// SetID sets the "id" field.
func (sec *SideEffectCreate) SetID(s string) *SideEffectCreate {
	sec.mutation.SetID(s)
	return sec
}

// AddExecutionIDs adds the "executions" edge to the SideEffectExecution entity by IDs.
func (sec *SideEffectCreate) AddExecutionIDs(ids ...string) *SideEffectCreate {
	sec.mutation.AddExecutionIDs(ids...)
	return sec
}

// AddExecutions adds the "executions" edges to the SideEffectExecution entity.
func (sec *SideEffectCreate) AddExecutions(s ...*SideEffectExecution) *SideEffectCreate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sec.AddExecutionIDs(ids...)
}

// Mutation returns the SideEffectMutation object of the builder.
func (sec *SideEffectCreate) Mutation() *SideEffectMutation {
	return sec.mutation
}

// Save creates the SideEffect in the database.
func (sec *SideEffectCreate) Save(ctx context.Context) (*SideEffect, error) {
	sec.defaults()
	return withHooks(ctx, sec.sqlSave, sec.mutation, sec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sec *SideEffectCreate) SaveX(ctx context.Context) *SideEffect {
	v, err := sec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sec *SideEffectCreate) Exec(ctx context.Context) error {
	_, err := sec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sec *SideEffectCreate) ExecX(ctx context.Context) {
	if err := sec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sec *SideEffectCreate) defaults() {
	if _, ok := sec.mutation.Status(); !ok {
		v := sideeffect.DefaultStatus
		sec.mutation.SetStatus(v)
	}
	if _, ok := sec.mutation.CreatedAt(); !ok {
		v := sideeffect.DefaultCreatedAt()
		sec.mutation.SetCreatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sec *SideEffectCreate) check() error {
	if _, ok := sec.mutation.Identity(); !ok {
		return &ValidationError{Name: "identity", err: errors.New(`ent: missing required field "SideEffect.identity"`)}
	}
	if v, ok := sec.mutation.Identity(); ok {
		if err := sideeffect.IdentityValidator(v); err != nil {
			return &ValidationError{Name: "identity", err: fmt.Errorf(`ent: validator failed for field "SideEffect.identity": %w`, err)}
		}
	}
	if _, ok := sec.mutation.StepID(); !ok {
		return &ValidationError{Name: "step_id", err: errors.New(`ent: missing required field "SideEffect.step_id"`)}
	}
	if v, ok := sec.mutation.StepID(); ok {
		if err := sideeffect.StepIDValidator(v); err != nil {
			return &ValidationError{Name: "step_id", err: fmt.Errorf(`ent: validator failed for field "SideEffect.step_id": %w`, err)}
		}
	}
	if _, ok := sec.mutation.HandlerName(); !ok {
		return &ValidationError{Name: "handler_name", err: errors.New(`ent: missing required field "SideEffect.handler_name"`)}
	}
	if v, ok := sec.mutation.HandlerName(); ok {
		if err := sideeffect.HandlerNameValidator(v); err != nil {
			return &ValidationError{Name: "handler_name", err: fmt.Errorf(`ent: validator failed for field "SideEffect.handler_name": %w`, err)}
		}
	}
	if _, ok := sec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "SideEffect.status"`)}
	}
	if v, ok := sec.mutation.Status(); ok {
		if err := sideeffect.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SideEffect.status": %w`, err)}
		}
	}
	if _, ok := sec.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SideEffect.created_at"`)}
	}
	return nil
}

func (sec *SideEffectCreate) sqlSave(ctx context.Context) (*SideEffect, error) {
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
			return nil, fmt.Errorf("unexpected SideEffect.ID type: %T", _spec.ID.Value)
		}
	}
	sec.mutation.id = &_node.ID
	sec.mutation.done = true
	return _node, nil
}

func (sec *SideEffectCreate) createSpec() (*SideEffect, *sqlgraph.CreateSpec) {
	var (
		_node = &SideEffect{config: sec.config}
		_spec = sqlgraph.NewCreateSpec(sideeffect.Table, sqlgraph.NewFieldSpec(sideeffect.FieldID, field.TypeString))
	)
	if id, ok := sec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := sec.mutation.Identity(); ok {
		_spec.SetField(sideeffect.FieldIdentity, field.TypeString, value)
		_node.Identity = value
	}
	if value, ok := sec.mutation.StepID(); ok {
		_spec.SetField(sideeffect.FieldStepID, field.TypeString, value)
		_node.StepID = value
	}
	if value, ok := sec.mutation.HandlerName(); ok {
		_spec.SetField(sideeffect.FieldHandlerName, field.TypeString, value)
		_node.HandlerName = value
	}
	if value, ok := sec.mutation.Status(); ok {
		_spec.SetField(sideeffect.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := sec.mutation.RetryPolicy(); ok {
		_spec.SetField(sideeffect.FieldRetryPolicy, field.TypeJSON, value)
		_node.RetryPolicy = value
	}
	if value, ok := sec.mutation.Timeout(); ok {
		_spec.SetField(sideeffect.FieldTimeout, field.TypeTime, value)
		_node.Timeout = value
	}
	if value, ok := sec.mutation.CreatedAt(); ok {
		_spec.SetField(sideeffect.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := sec.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffect.ExecutionsTable,
			Columns: []string{sideeffect.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SideEffectCreateBulk is the builder for creating many SideEffect entities in bulk.
type SideEffectCreateBulk struct {
	config
	err      error
	builders []*SideEffectCreate
}

// Save creates the SideEffect entities in the database.
func (secb *SideEffectCreateBulk) Save(ctx context.Context) ([]*SideEffect, error) {
	if secb.err != nil {
		return nil, secb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(secb.builders))
	nodes := make([]*SideEffect, len(secb.builders))
	mutators := make([]Mutator, len(secb.builders))
	for i := range secb.builders {
		func(i int, root context.Context) {
			builder := secb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SideEffectMutation)
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
func (secb *SideEffectCreateBulk) SaveX(ctx context.Context) []*SideEffect {
	v, err := secb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (secb *SideEffectCreateBulk) Exec(ctx context.Context) error {
	_, err := secb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (secb *SideEffectCreateBulk) ExecX(ctx context.Context) {
	if err := secb.Exec(ctx); err != nil {
		panic(err)
	}
}
