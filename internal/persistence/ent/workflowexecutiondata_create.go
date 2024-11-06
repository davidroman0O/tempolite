// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecutiondata"
)

// WorkflowExecutionDataCreate is the builder for creating a WorkflowExecutionData entity.
type WorkflowExecutionDataCreate struct {
	config
	mutation *WorkflowExecutionDataMutation
	hooks    []Hook
}

// SetCheckpoints sets the "checkpoints" field.
func (wedc *WorkflowExecutionDataCreate) SetCheckpoints(u [][]uint8) *WorkflowExecutionDataCreate {
	wedc.mutation.SetCheckpoints(u)
	return wedc
}

// SetCheckpointTime sets the "checkpoint_time" field.
func (wedc *WorkflowExecutionDataCreate) SetCheckpointTime(t time.Time) *WorkflowExecutionDataCreate {
	wedc.mutation.SetCheckpointTime(t)
	return wedc
}

// SetNillableCheckpointTime sets the "checkpoint_time" field if the given value is not nil.
func (wedc *WorkflowExecutionDataCreate) SetNillableCheckpointTime(t *time.Time) *WorkflowExecutionDataCreate {
	if t != nil {
		wedc.SetCheckpointTime(*t)
	}
	return wedc
}

// SetError sets the "error" field.
func (wedc *WorkflowExecutionDataCreate) SetError(s string) *WorkflowExecutionDataCreate {
	wedc.mutation.SetError(s)
	return wedc
}

// SetNillableError sets the "error" field if the given value is not nil.
func (wedc *WorkflowExecutionDataCreate) SetNillableError(s *string) *WorkflowExecutionDataCreate {
	if s != nil {
		wedc.SetError(*s)
	}
	return wedc
}

// SetOutput sets the "output" field.
func (wedc *WorkflowExecutionDataCreate) SetOutput(u [][]uint8) *WorkflowExecutionDataCreate {
	wedc.mutation.SetOutput(u)
	return wedc
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (wedc *WorkflowExecutionDataCreate) SetWorkflowExecutionID(id int) *WorkflowExecutionDataCreate {
	wedc.mutation.SetWorkflowExecutionID(id)
	return wedc
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (wedc *WorkflowExecutionDataCreate) SetWorkflowExecution(w *WorkflowExecution) *WorkflowExecutionDataCreate {
	return wedc.SetWorkflowExecutionID(w.ID)
}

// Mutation returns the WorkflowExecutionDataMutation object of the builder.
func (wedc *WorkflowExecutionDataCreate) Mutation() *WorkflowExecutionDataMutation {
	return wedc.mutation
}

// Save creates the WorkflowExecutionData in the database.
func (wedc *WorkflowExecutionDataCreate) Save(ctx context.Context) (*WorkflowExecutionData, error) {
	return withHooks(ctx, wedc.sqlSave, wedc.mutation, wedc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (wedc *WorkflowExecutionDataCreate) SaveX(ctx context.Context) *WorkflowExecutionData {
	v, err := wedc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (wedc *WorkflowExecutionDataCreate) Exec(ctx context.Context) error {
	_, err := wedc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wedc *WorkflowExecutionDataCreate) ExecX(ctx context.Context) {
	if err := wedc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (wedc *WorkflowExecutionDataCreate) check() error {
	if len(wedc.mutation.WorkflowExecutionIDs()) == 0 {
		return &ValidationError{Name: "workflow_execution", err: errors.New(`ent: missing required edge "WorkflowExecutionData.workflow_execution"`)}
	}
	return nil
}

func (wedc *WorkflowExecutionDataCreate) sqlSave(ctx context.Context) (*WorkflowExecutionData, error) {
	if err := wedc.check(); err != nil {
		return nil, err
	}
	_node, _spec := wedc.createSpec()
	if err := sqlgraph.CreateNode(ctx, wedc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	wedc.mutation.id = &_node.ID
	wedc.mutation.done = true
	return _node, nil
}

func (wedc *WorkflowExecutionDataCreate) createSpec() (*WorkflowExecutionData, *sqlgraph.CreateSpec) {
	var (
		_node = &WorkflowExecutionData{config: wedc.config}
		_spec = sqlgraph.NewCreateSpec(workflowexecutiondata.Table, sqlgraph.NewFieldSpec(workflowexecutiondata.FieldID, field.TypeInt))
	)
	if value, ok := wedc.mutation.Checkpoints(); ok {
		_spec.SetField(workflowexecutiondata.FieldCheckpoints, field.TypeJSON, value)
		_node.Checkpoints = value
	}
	if value, ok := wedc.mutation.CheckpointTime(); ok {
		_spec.SetField(workflowexecutiondata.FieldCheckpointTime, field.TypeTime, value)
		_node.CheckpointTime = &value
	}
	if value, ok := wedc.mutation.Error(); ok {
		_spec.SetField(workflowexecutiondata.FieldError, field.TypeString, value)
		_node.Error = value
	}
	if value, ok := wedc.mutation.Output(); ok {
		_spec.SetField(workflowexecutiondata.FieldOutput, field.TypeJSON, value)
		_node.Output = value
	}
	if nodes := wedc.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   workflowexecutiondata.WorkflowExecutionTable,
			Columns: []string{workflowexecutiondata.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.workflow_execution_execution_data = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// WorkflowExecutionDataCreateBulk is the builder for creating many WorkflowExecutionData entities in bulk.
type WorkflowExecutionDataCreateBulk struct {
	config
	err      error
	builders []*WorkflowExecutionDataCreate
}

// Save creates the WorkflowExecutionData entities in the database.
func (wedcb *WorkflowExecutionDataCreateBulk) Save(ctx context.Context) ([]*WorkflowExecutionData, error) {
	if wedcb.err != nil {
		return nil, wedcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(wedcb.builders))
	nodes := make([]*WorkflowExecutionData, len(wedcb.builders))
	mutators := make([]Mutator, len(wedcb.builders))
	for i := range wedcb.builders {
		func(i int, root context.Context) {
			builder := wedcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*WorkflowExecutionDataMutation)
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
					_, err = mutators[i+1].Mutate(root, wedcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, wedcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, wedcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (wedcb *WorkflowExecutionDataCreateBulk) SaveX(ctx context.Context) []*WorkflowExecutionData {
	v, err := wedcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (wedcb *WorkflowExecutionDataCreateBulk) Exec(ctx context.Context) error {
	_, err := wedcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wedcb *WorkflowExecutionDataCreateBulk) ExecX(ctx context.Context) {
	if err := wedcb.Exec(ctx); err != nil {
		panic(err)
	}
}
