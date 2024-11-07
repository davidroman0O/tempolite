// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
)

// ExecutionCreate is the builder for creating a Execution entity.
type ExecutionCreate struct {
	config
	mutation *ExecutionMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (ec *ExecutionCreate) SetCreatedAt(t time.Time) *ExecutionCreate {
	ec.mutation.SetCreatedAt(t)
	return ec
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (ec *ExecutionCreate) SetNillableCreatedAt(t *time.Time) *ExecutionCreate {
	if t != nil {
		ec.SetCreatedAt(*t)
	}
	return ec
}

// SetUpdatedAt sets the "updated_at" field.
func (ec *ExecutionCreate) SetUpdatedAt(t time.Time) *ExecutionCreate {
	ec.mutation.SetUpdatedAt(t)
	return ec
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (ec *ExecutionCreate) SetNillableUpdatedAt(t *time.Time) *ExecutionCreate {
	if t != nil {
		ec.SetUpdatedAt(*t)
	}
	return ec
}

// SetStartedAt sets the "started_at" field.
func (ec *ExecutionCreate) SetStartedAt(t time.Time) *ExecutionCreate {
	ec.mutation.SetStartedAt(t)
	return ec
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (ec *ExecutionCreate) SetNillableStartedAt(t *time.Time) *ExecutionCreate {
	if t != nil {
		ec.SetStartedAt(*t)
	}
	return ec
}

// SetCompletedAt sets the "completed_at" field.
func (ec *ExecutionCreate) SetCompletedAt(t time.Time) *ExecutionCreate {
	ec.mutation.SetCompletedAt(t)
	return ec
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (ec *ExecutionCreate) SetNillableCompletedAt(t *time.Time) *ExecutionCreate {
	if t != nil {
		ec.SetCompletedAt(*t)
	}
	return ec
}

// SetStatus sets the "status" field.
func (ec *ExecutionCreate) SetStatus(e execution.Status) *ExecutionCreate {
	ec.mutation.SetStatus(e)
	return ec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (ec *ExecutionCreate) SetNillableStatus(e *execution.Status) *ExecutionCreate {
	if e != nil {
		ec.SetStatus(*e)
	}
	return ec
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (ec *ExecutionCreate) SetEntityID(id int) *ExecutionCreate {
	ec.mutation.SetEntityID(id)
	return ec
}

// SetEntity sets the "entity" edge to the Entity entity.
func (ec *ExecutionCreate) SetEntity(e *Entity) *ExecutionCreate {
	return ec.SetEntityID(e.ID)
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (ec *ExecutionCreate) SetWorkflowExecutionID(id int) *ExecutionCreate {
	ec.mutation.SetWorkflowExecutionID(id)
	return ec
}

// SetNillableWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID if the given value is not nil.
func (ec *ExecutionCreate) SetNillableWorkflowExecutionID(id *int) *ExecutionCreate {
	if id != nil {
		ec = ec.SetWorkflowExecutionID(*id)
	}
	return ec
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (ec *ExecutionCreate) SetWorkflowExecution(w *WorkflowExecution) *ExecutionCreate {
	return ec.SetWorkflowExecutionID(w.ID)
}

// SetActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID.
func (ec *ExecutionCreate) SetActivityExecutionID(id int) *ExecutionCreate {
	ec.mutation.SetActivityExecutionID(id)
	return ec
}

// SetNillableActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID if the given value is not nil.
func (ec *ExecutionCreate) SetNillableActivityExecutionID(id *int) *ExecutionCreate {
	if id != nil {
		ec = ec.SetActivityExecutionID(*id)
	}
	return ec
}

// SetActivityExecution sets the "activity_execution" edge to the ActivityExecution entity.
func (ec *ExecutionCreate) SetActivityExecution(a *ActivityExecution) *ExecutionCreate {
	return ec.SetActivityExecutionID(a.ID)
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (ec *ExecutionCreate) SetSagaExecutionID(id int) *ExecutionCreate {
	ec.mutation.SetSagaExecutionID(id)
	return ec
}

// SetNillableSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID if the given value is not nil.
func (ec *ExecutionCreate) SetNillableSagaExecutionID(id *int) *ExecutionCreate {
	if id != nil {
		ec = ec.SetSagaExecutionID(*id)
	}
	return ec
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (ec *ExecutionCreate) SetSagaExecution(s *SagaExecution) *ExecutionCreate {
	return ec.SetSagaExecutionID(s.ID)
}

// SetSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID.
func (ec *ExecutionCreate) SetSideEffectExecutionID(id int) *ExecutionCreate {
	ec.mutation.SetSideEffectExecutionID(id)
	return ec
}

// SetNillableSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID if the given value is not nil.
func (ec *ExecutionCreate) SetNillableSideEffectExecutionID(id *int) *ExecutionCreate {
	if id != nil {
		ec = ec.SetSideEffectExecutionID(*id)
	}
	return ec
}

// SetSideEffectExecution sets the "side_effect_execution" edge to the SideEffectExecution entity.
func (ec *ExecutionCreate) SetSideEffectExecution(s *SideEffectExecution) *ExecutionCreate {
	return ec.SetSideEffectExecutionID(s.ID)
}

// Mutation returns the ExecutionMutation object of the builder.
func (ec *ExecutionCreate) Mutation() *ExecutionMutation {
	return ec.mutation
}

// Save creates the Execution in the database.
func (ec *ExecutionCreate) Save(ctx context.Context) (*Execution, error) {
	ec.defaults()
	return withHooks(ctx, ec.sqlSave, ec.mutation, ec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ec *ExecutionCreate) SaveX(ctx context.Context) *Execution {
	v, err := ec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ec *ExecutionCreate) Exec(ctx context.Context) error {
	_, err := ec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ec *ExecutionCreate) ExecX(ctx context.Context) {
	if err := ec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ec *ExecutionCreate) defaults() {
	if _, ok := ec.mutation.CreatedAt(); !ok {
		v := execution.DefaultCreatedAt()
		ec.mutation.SetCreatedAt(v)
	}
	if _, ok := ec.mutation.UpdatedAt(); !ok {
		v := execution.DefaultUpdatedAt()
		ec.mutation.SetUpdatedAt(v)
	}
	if _, ok := ec.mutation.StartedAt(); !ok {
		v := execution.DefaultStartedAt()
		ec.mutation.SetStartedAt(v)
	}
	if _, ok := ec.mutation.Status(); !ok {
		v := execution.DefaultStatus
		ec.mutation.SetStatus(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ec *ExecutionCreate) check() error {
	if _, ok := ec.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Execution.created_at"`)}
	}
	if _, ok := ec.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Execution.updated_at"`)}
	}
	if _, ok := ec.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "Execution.started_at"`)}
	}
	if _, ok := ec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Execution.status"`)}
	}
	if v, ok := ec.mutation.Status(); ok {
		if err := execution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Execution.status": %w`, err)}
		}
	}
	if len(ec.mutation.EntityIDs()) == 0 {
		return &ValidationError{Name: "entity", err: errors.New(`ent: missing required edge "Execution.entity"`)}
	}
	return nil
}

func (ec *ExecutionCreate) sqlSave(ctx context.Context) (*Execution, error) {
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

func (ec *ExecutionCreate) createSpec() (*Execution, *sqlgraph.CreateSpec) {
	var (
		_node = &Execution{config: ec.config}
		_spec = sqlgraph.NewCreateSpec(execution.Table, sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt))
	)
	if value, ok := ec.mutation.CreatedAt(); ok {
		_spec.SetField(execution.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := ec.mutation.UpdatedAt(); ok {
		_spec.SetField(execution.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := ec.mutation.StartedAt(); ok {
		_spec.SetField(execution.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := ec.mutation.CompletedAt(); ok {
		_spec.SetField(execution.FieldCompletedAt, field.TypeTime, value)
		_node.CompletedAt = &value
	}
	if value, ok := ec.mutation.Status(); ok {
		_spec.SetField(execution.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if nodes := ec.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   execution.EntityTable,
			Columns: []string{execution.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entity_executions = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.WorkflowExecutionTable,
			Columns: []string{execution.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.ActivityExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.ActivityExecutionTable,
			Columns: []string{execution.ActivityExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.SagaExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SagaExecutionTable,
			Columns: []string{execution.SagaExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.SideEffectExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SideEffectExecutionTable,
			Columns: []string{execution.SideEffectExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// ExecutionCreateBulk is the builder for creating many Execution entities in bulk.
type ExecutionCreateBulk struct {
	config
	err      error
	builders []*ExecutionCreate
}

// Save creates the Execution entities in the database.
func (ecb *ExecutionCreateBulk) Save(ctx context.Context) ([]*Execution, error) {
	if ecb.err != nil {
		return nil, ecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ecb.builders))
	nodes := make([]*Execution, len(ecb.builders))
	mutators := make([]Mutator, len(ecb.builders))
	for i := range ecb.builders {
		func(i int, root context.Context) {
			builder := ecb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ExecutionMutation)
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
func (ecb *ExecutionCreateBulk) SaveX(ctx context.Context) []*Execution {
	v, err := ecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ecb *ExecutionCreateBulk) Exec(ctx context.Context) error {
	_, err := ecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ecb *ExecutionCreateBulk) ExecX(ctx context.Context) {
	if err := ecb.Exec(ctx); err != nil {
		panic(err)
	}
}