// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/go-tempolite/ent/compensationtask"
	"github.com/davidroman0O/go-tempolite/ent/entry"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/sagatask"
	"github.com/davidroman0O/go-tempolite/ent/sideeffecttask"
)

// EntryCreate is the builder for creating a Entry entity.
type EntryCreate struct {
	config
	mutation *EntryMutation
	hooks    []Hook
}

// SetTaskID sets the "taskID" field.
func (ec *EntryCreate) SetTaskID(s string) *EntryCreate {
	ec.mutation.SetTaskID(s)
	return ec
}

// SetType sets the "type" field.
func (ec *EntryCreate) SetType(e entry.Type) *EntryCreate {
	ec.mutation.SetType(e)
	return ec
}

// SetExecutionContextID sets the "execution_context" edge to the ExecutionContext entity by ID.
func (ec *EntryCreate) SetExecutionContextID(id string) *EntryCreate {
	ec.mutation.SetExecutionContextID(id)
	return ec
}

// SetNillableExecutionContextID sets the "execution_context" edge to the ExecutionContext entity by ID if the given value is not nil.
func (ec *EntryCreate) SetNillableExecutionContextID(id *string) *EntryCreate {
	if id != nil {
		ec = ec.SetExecutionContextID(*id)
	}
	return ec
}

// SetExecutionContext sets the "execution_context" edge to the ExecutionContext entity.
func (ec *EntryCreate) SetExecutionContext(e *ExecutionContext) *EntryCreate {
	return ec.SetExecutionContextID(e.ID)
}

// SetHandlerTaskID sets the "handler_task" edge to the HandlerTask entity by ID.
func (ec *EntryCreate) SetHandlerTaskID(id string) *EntryCreate {
	ec.mutation.SetHandlerTaskID(id)
	return ec
}

// SetNillableHandlerTaskID sets the "handler_task" edge to the HandlerTask entity by ID if the given value is not nil.
func (ec *EntryCreate) SetNillableHandlerTaskID(id *string) *EntryCreate {
	if id != nil {
		ec = ec.SetHandlerTaskID(*id)
	}
	return ec
}

// SetHandlerTask sets the "handler_task" edge to the HandlerTask entity.
func (ec *EntryCreate) SetHandlerTask(h *HandlerTask) *EntryCreate {
	return ec.SetHandlerTaskID(h.ID)
}

// SetSagaStepTaskID sets the "saga_step_task" edge to the SagaTask entity by ID.
func (ec *EntryCreate) SetSagaStepTaskID(id int) *EntryCreate {
	ec.mutation.SetSagaStepTaskID(id)
	return ec
}

// SetNillableSagaStepTaskID sets the "saga_step_task" edge to the SagaTask entity by ID if the given value is not nil.
func (ec *EntryCreate) SetNillableSagaStepTaskID(id *int) *EntryCreate {
	if id != nil {
		ec = ec.SetSagaStepTaskID(*id)
	}
	return ec
}

// SetSagaStepTask sets the "saga_step_task" edge to the SagaTask entity.
func (ec *EntryCreate) SetSagaStepTask(s *SagaTask) *EntryCreate {
	return ec.SetSagaStepTaskID(s.ID)
}

// SetSideEffectTaskID sets the "side_effect_task" edge to the SideEffectTask entity by ID.
func (ec *EntryCreate) SetSideEffectTaskID(id int) *EntryCreate {
	ec.mutation.SetSideEffectTaskID(id)
	return ec
}

// SetNillableSideEffectTaskID sets the "side_effect_task" edge to the SideEffectTask entity by ID if the given value is not nil.
func (ec *EntryCreate) SetNillableSideEffectTaskID(id *int) *EntryCreate {
	if id != nil {
		ec = ec.SetSideEffectTaskID(*id)
	}
	return ec
}

// SetSideEffectTask sets the "side_effect_task" edge to the SideEffectTask entity.
func (ec *EntryCreate) SetSideEffectTask(s *SideEffectTask) *EntryCreate {
	return ec.SetSideEffectTaskID(s.ID)
}

// SetCompensationTaskID sets the "compensation_task" edge to the CompensationTask entity by ID.
func (ec *EntryCreate) SetCompensationTaskID(id int) *EntryCreate {
	ec.mutation.SetCompensationTaskID(id)
	return ec
}

// SetNillableCompensationTaskID sets the "compensation_task" edge to the CompensationTask entity by ID if the given value is not nil.
func (ec *EntryCreate) SetNillableCompensationTaskID(id *int) *EntryCreate {
	if id != nil {
		ec = ec.SetCompensationTaskID(*id)
	}
	return ec
}

// SetCompensationTask sets the "compensation_task" edge to the CompensationTask entity.
func (ec *EntryCreate) SetCompensationTask(c *CompensationTask) *EntryCreate {
	return ec.SetCompensationTaskID(c.ID)
}

// Mutation returns the EntryMutation object of the builder.
func (ec *EntryCreate) Mutation() *EntryMutation {
	return ec.mutation
}

// Save creates the Entry in the database.
func (ec *EntryCreate) Save(ctx context.Context) (*Entry, error) {
	return withHooks(ctx, ec.sqlSave, ec.mutation, ec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ec *EntryCreate) SaveX(ctx context.Context) *Entry {
	v, err := ec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ec *EntryCreate) Exec(ctx context.Context) error {
	_, err := ec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ec *EntryCreate) ExecX(ctx context.Context) {
	if err := ec.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ec *EntryCreate) check() error {
	if _, ok := ec.mutation.TaskID(); !ok {
		return &ValidationError{Name: "taskID", err: errors.New(`ent: missing required field "Entry.taskID"`)}
	}
	if _, ok := ec.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "Entry.type"`)}
	}
	if v, ok := ec.mutation.GetType(); ok {
		if err := entry.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Entry.type": %w`, err)}
		}
	}
	return nil
}

func (ec *EntryCreate) sqlSave(ctx context.Context) (*Entry, error) {
	if err := ec.check(); err != nil {
		return nil, err
	}
	_node, _spec := ec.createSpec()
	if err := sqlgraph.CreateNode(ctx, ec.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	ec.mutation.id = &_node.ID
	ec.mutation.done = true
	return _node, nil
}

func (ec *EntryCreate) createSpec() (*Entry, *sqlgraph.CreateSpec) {
	var (
		_node = &Entry{config: ec.config}
		_spec = sqlgraph.NewCreateSpec(entry.Table, sqlgraph.NewFieldSpec(entry.FieldID, field.TypeInt))
	)
	if value, ok := ec.mutation.TaskID(); ok {
		_spec.SetField(entry.FieldTaskID, field.TypeString, value)
		_node.TaskID = value
	}
	if value, ok := ec.mutation.GetType(); ok {
		_spec.SetField(entry.FieldType, field.TypeEnum, value)
		_node.Type = value
	}
	if nodes := ec.mutation.ExecutionContextIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   entry.ExecutionContextTable,
			Columns: []string{entry.ExecutionContextColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(executioncontext.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entry_execution_context = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.HandlerTaskIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   entry.HandlerTaskTable,
			Columns: []string{entry.HandlerTaskColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(handlertask.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entry_handler_task = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.SagaStepTaskIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   entry.SagaStepTaskTable,
			Columns: []string{entry.SagaStepTaskColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagatask.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entry_saga_step_task = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.SideEffectTaskIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   entry.SideEffectTaskTable,
			Columns: []string{entry.SideEffectTaskColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffecttask.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entry_side_effect_task = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.CompensationTaskIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   entry.CompensationTaskTable,
			Columns: []string{entry.CompensationTaskColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(compensationtask.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entry_compensation_task = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// EntryCreateBulk is the builder for creating many Entry entities in bulk.
type EntryCreateBulk struct {
	config
	err      error
	builders []*EntryCreate
}

// Save creates the Entry entities in the database.
func (ecb *EntryCreateBulk) Save(ctx context.Context) ([]*Entry, error) {
	if ecb.err != nil {
		return nil, ecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ecb.builders))
	nodes := make([]*Entry, len(ecb.builders))
	mutators := make([]Mutator, len(ecb.builders))
	for i := range ecb.builders {
		func(i int, root context.Context) {
			builder := ecb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*EntryMutation)
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
					_, err = mutators[i+1].Mutate(root, ecb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ecb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
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
		if _, err := mutators[0].Mutate(ctx, ecb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ecb *EntryCreateBulk) SaveX(ctx context.Context) []*Entry {
	v, err := ecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ecb *EntryCreateBulk) Exec(ctx context.Context) error {
	_, err := ecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ecb *EntryCreateBulk) ExecX(ctx context.Context) {
	if err := ecb.Exec(ctx); err != nil {
		panic(err)
	}
}
