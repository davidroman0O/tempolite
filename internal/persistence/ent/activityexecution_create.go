// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
)

// ActivityExecutionCreate is the builder for creating a ActivityExecution entity.
type ActivityExecutionCreate struct {
	config
	mutation *ActivityExecutionMutation
	hooks    []Hook
}

// SetInputs sets the "inputs" field.
func (aec *ActivityExecutionCreate) SetInputs(u [][]uint8) *ActivityExecutionCreate {
	aec.mutation.SetInputs(u)
	return aec
}

// SetExecutionID sets the "execution" edge to the Execution entity by ID.
func (aec *ActivityExecutionCreate) SetExecutionID(id int) *ActivityExecutionCreate {
	aec.mutation.SetExecutionID(id)
	return aec
}

// SetExecution sets the "execution" edge to the Execution entity.
func (aec *ActivityExecutionCreate) SetExecution(e *Execution) *ActivityExecutionCreate {
	return aec.SetExecutionID(e.ID)
}

// SetExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID.
func (aec *ActivityExecutionCreate) SetExecutionDataID(id int) *ActivityExecutionCreate {
	aec.mutation.SetExecutionDataID(id)
	return aec
}

// SetNillableExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID if the given value is not nil.
func (aec *ActivityExecutionCreate) SetNillableExecutionDataID(id *int) *ActivityExecutionCreate {
	if id != nil {
		aec = aec.SetExecutionDataID(*id)
	}
	return aec
}

// SetExecutionData sets the "execution_data" edge to the ActivityExecutionData entity.
func (aec *ActivityExecutionCreate) SetExecutionData(a *ActivityExecutionData) *ActivityExecutionCreate {
	return aec.SetExecutionDataID(a.ID)
}

// Mutation returns the ActivityExecutionMutation object of the builder.
func (aec *ActivityExecutionCreate) Mutation() *ActivityExecutionMutation {
	return aec.mutation
}

// Save creates the ActivityExecution in the database.
func (aec *ActivityExecutionCreate) Save(ctx context.Context) (*ActivityExecution, error) {
	return withHooks(ctx, aec.sqlSave, aec.mutation, aec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (aec *ActivityExecutionCreate) SaveX(ctx context.Context) *ActivityExecution {
	v, err := aec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (aec *ActivityExecutionCreate) Exec(ctx context.Context) error {
	_, err := aec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aec *ActivityExecutionCreate) ExecX(ctx context.Context) {
	if err := aec.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (aec *ActivityExecutionCreate) check() error {
	if len(aec.mutation.ExecutionIDs()) == 0 {
		return &ValidationError{Name: "execution", err: errors.New(`ent: missing required edge "ActivityExecution.execution"`)}
	}
	return nil
}

func (aec *ActivityExecutionCreate) sqlSave(ctx context.Context) (*ActivityExecution, error) {
	if err := aec.check(); err != nil {
		return nil, err
	}
	_node, _spec := aec.createSpec()
	if err := sqlgraph.CreateNode(ctx, aec.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	aec.mutation.id = &_node.ID
	aec.mutation.done = true
	return _node, nil
}

func (aec *ActivityExecutionCreate) createSpec() (*ActivityExecution, *sqlgraph.CreateSpec) {
	var (
		_node = &ActivityExecution{config: aec.config}
		_spec = sqlgraph.NewCreateSpec(activityexecution.Table, sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt))
	)
	if value, ok := aec.mutation.Inputs(); ok {
		_spec.SetField(activityexecution.FieldInputs, field.TypeJSON, value)
		_node.Inputs = value
	}
	if nodes := aec.mutation.ExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   activityexecution.ExecutionTable,
			Columns: []string{activityexecution.ExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.execution_activity_execution = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := aec.mutation.ExecutionDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityexecution.ExecutionDataTable,
			Columns: []string{activityexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecutiondata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// ActivityExecutionCreateBulk is the builder for creating many ActivityExecution entities in bulk.
type ActivityExecutionCreateBulk struct {
	config
	err      error
	builders []*ActivityExecutionCreate
}

// Save creates the ActivityExecution entities in the database.
func (aecb *ActivityExecutionCreateBulk) Save(ctx context.Context) ([]*ActivityExecution, error) {
	if aecb.err != nil {
		return nil, aecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(aecb.builders))
	nodes := make([]*ActivityExecution, len(aecb.builders))
	mutators := make([]Mutator, len(aecb.builders))
	for i := range aecb.builders {
		func(i int, root context.Context) {
			builder := aecb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ActivityExecutionMutation)
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
					_, err = mutators[i+1].Mutate(root, aecb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, aecb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, aecb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (aecb *ActivityExecutionCreateBulk) SaveX(ctx context.Context) []*ActivityExecution {
	v, err := aecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (aecb *ActivityExecutionCreateBulk) Exec(ctx context.Context) error {
	_, err := aecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aecb *ActivityExecutionCreateBulk) ExecX(ctx context.Context) {
	if err := aecb.Exec(ctx); err != nil {
		panic(err)
	}
}
