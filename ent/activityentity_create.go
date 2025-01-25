// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/activitydata"
	"github.com/davidroman0O/tempolite/ent/activityentity"
	"github.com/davidroman0O/tempolite/ent/activityexecution"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
)

// ActivityEntityCreate is the builder for creating a ActivityEntity entity.
type ActivityEntityCreate struct {
	config
	mutation *ActivityEntityMutation
	hooks    []Hook
}

// SetHandlerName sets the "handler_name" field.
func (aec *ActivityEntityCreate) SetHandlerName(s string) *ActivityEntityCreate {
	aec.mutation.SetHandlerName(s)
	return aec
}

// SetType sets the "type" field.
func (aec *ActivityEntityCreate) SetType(st schema.EntityType) *ActivityEntityCreate {
	aec.mutation.SetType(st)
	return aec
}

// SetNillableType sets the "type" field if the given value is not nil.
func (aec *ActivityEntityCreate) SetNillableType(st *schema.EntityType) *ActivityEntityCreate {
	if st != nil {
		aec.SetType(*st)
	}
	return aec
}

// SetStatus sets the "status" field.
func (aec *ActivityEntityCreate) SetStatus(ss schema.EntityStatus) *ActivityEntityCreate {
	aec.mutation.SetStatus(ss)
	return aec
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (aec *ActivityEntityCreate) SetNillableStatus(ss *schema.EntityStatus) *ActivityEntityCreate {
	if ss != nil {
		aec.SetStatus(*ss)
	}
	return aec
}

// SetStepID sets the "step_id" field.
func (aec *ActivityEntityCreate) SetStepID(ssi schema.ActivityStepID) *ActivityEntityCreate {
	aec.mutation.SetStepID(ssi)
	return aec
}

// SetRunID sets the "run_id" field.
func (aec *ActivityEntityCreate) SetRunID(si schema.RunID) *ActivityEntityCreate {
	aec.mutation.SetRunID(si)
	return aec
}

// SetRetryPolicy sets the "retry_policy" field.
func (aec *ActivityEntityCreate) SetRetryPolicy(sp schema.RetryPolicy) *ActivityEntityCreate {
	aec.mutation.SetRetryPolicy(sp)
	return aec
}

// SetRetryState sets the "retry_state" field.
func (aec *ActivityEntityCreate) SetRetryState(ss schema.RetryState) *ActivityEntityCreate {
	aec.mutation.SetRetryState(ss)
	return aec
}

// SetCreatedAt sets the "created_at" field.
func (aec *ActivityEntityCreate) SetCreatedAt(t time.Time) *ActivityEntityCreate {
	aec.mutation.SetCreatedAt(t)
	return aec
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (aec *ActivityEntityCreate) SetNillableCreatedAt(t *time.Time) *ActivityEntityCreate {
	if t != nil {
		aec.SetCreatedAt(*t)
	}
	return aec
}

// SetUpdatedAt sets the "updated_at" field.
func (aec *ActivityEntityCreate) SetUpdatedAt(t time.Time) *ActivityEntityCreate {
	aec.mutation.SetUpdatedAt(t)
	return aec
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (aec *ActivityEntityCreate) SetNillableUpdatedAt(t *time.Time) *ActivityEntityCreate {
	if t != nil {
		aec.SetUpdatedAt(*t)
	}
	return aec
}

// SetID sets the "id" field.
func (aec *ActivityEntityCreate) SetID(sei schema.ActivityEntityID) *ActivityEntityCreate {
	aec.mutation.SetID(sei)
	return aec
}

// SetWorkflowID sets the "workflow" edge to the WorkflowEntity entity by ID.
func (aec *ActivityEntityCreate) SetWorkflowID(id schema.WorkflowEntityID) *ActivityEntityCreate {
	aec.mutation.SetWorkflowID(id)
	return aec
}

// SetWorkflow sets the "workflow" edge to the WorkflowEntity entity.
func (aec *ActivityEntityCreate) SetWorkflow(w *WorkflowEntity) *ActivityEntityCreate {
	return aec.SetWorkflowID(w.ID)
}

// SetActivityDataID sets the "activity_data" edge to the ActivityData entity by ID.
func (aec *ActivityEntityCreate) SetActivityDataID(id schema.ActivityDataID) *ActivityEntityCreate {
	aec.mutation.SetActivityDataID(id)
	return aec
}

// SetNillableActivityDataID sets the "activity_data" edge to the ActivityData entity by ID if the given value is not nil.
func (aec *ActivityEntityCreate) SetNillableActivityDataID(id *schema.ActivityDataID) *ActivityEntityCreate {
	if id != nil {
		aec = aec.SetActivityDataID(*id)
	}
	return aec
}

// SetActivityData sets the "activity_data" edge to the ActivityData entity.
func (aec *ActivityEntityCreate) SetActivityData(a *ActivityData) *ActivityEntityCreate {
	return aec.SetActivityDataID(a.ID)
}

// AddExecutionIDs adds the "executions" edge to the ActivityExecution entity by IDs.
func (aec *ActivityEntityCreate) AddExecutionIDs(ids ...schema.ActivityExecutionID) *ActivityEntityCreate {
	aec.mutation.AddExecutionIDs(ids...)
	return aec
}

// AddExecutions adds the "executions" edges to the ActivityExecution entity.
func (aec *ActivityEntityCreate) AddExecutions(a ...*ActivityExecution) *ActivityEntityCreate {
	ids := make([]schema.ActivityExecutionID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return aec.AddExecutionIDs(ids...)
}

// Mutation returns the ActivityEntityMutation object of the builder.
func (aec *ActivityEntityCreate) Mutation() *ActivityEntityMutation {
	return aec.mutation
}

// Save creates the ActivityEntity in the database.
func (aec *ActivityEntityCreate) Save(ctx context.Context) (*ActivityEntity, error) {
	aec.defaults()
	return withHooks(ctx, aec.sqlSave, aec.mutation, aec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (aec *ActivityEntityCreate) SaveX(ctx context.Context) *ActivityEntity {
	v, err := aec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (aec *ActivityEntityCreate) Exec(ctx context.Context) error {
	_, err := aec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aec *ActivityEntityCreate) ExecX(ctx context.Context) {
	if err := aec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (aec *ActivityEntityCreate) defaults() {
	if _, ok := aec.mutation.GetType(); !ok {
		v := activityentity.DefaultType
		aec.mutation.SetType(v)
	}
	if _, ok := aec.mutation.Status(); !ok {
		v := activityentity.DefaultStatus
		aec.mutation.SetStatus(v)
	}
	if _, ok := aec.mutation.CreatedAt(); !ok {
		v := activityentity.DefaultCreatedAt()
		aec.mutation.SetCreatedAt(v)
	}
	if _, ok := aec.mutation.UpdatedAt(); !ok {
		v := activityentity.DefaultUpdatedAt()
		aec.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (aec *ActivityEntityCreate) check() error {
	if _, ok := aec.mutation.HandlerName(); !ok {
		return &ValidationError{Name: "handler_name", err: errors.New(`ent: missing required field "ActivityEntity.handler_name"`)}
	}
	if _, ok := aec.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "ActivityEntity.type"`)}
	}
	if _, ok := aec.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "ActivityEntity.status"`)}
	}
	if _, ok := aec.mutation.StepID(); !ok {
		return &ValidationError{Name: "step_id", err: errors.New(`ent: missing required field "ActivityEntity.step_id"`)}
	}
	if _, ok := aec.mutation.RunID(); !ok {
		return &ValidationError{Name: "run_id", err: errors.New(`ent: missing required field "ActivityEntity.run_id"`)}
	}
	if _, ok := aec.mutation.RetryPolicy(); !ok {
		return &ValidationError{Name: "retry_policy", err: errors.New(`ent: missing required field "ActivityEntity.retry_policy"`)}
	}
	if _, ok := aec.mutation.RetryState(); !ok {
		return &ValidationError{Name: "retry_state", err: errors.New(`ent: missing required field "ActivityEntity.retry_state"`)}
	}
	if _, ok := aec.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "ActivityEntity.created_at"`)}
	}
	if _, ok := aec.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "ActivityEntity.updated_at"`)}
	}
	if len(aec.mutation.WorkflowIDs()) == 0 {
		return &ValidationError{Name: "workflow", err: errors.New(`ent: missing required edge "ActivityEntity.workflow"`)}
	}
	return nil
}

func (aec *ActivityEntityCreate) sqlSave(ctx context.Context) (*ActivityEntity, error) {
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
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = schema.ActivityEntityID(id)
	}
	aec.mutation.id = &_node.ID
	aec.mutation.done = true
	return _node, nil
}

func (aec *ActivityEntityCreate) createSpec() (*ActivityEntity, *sqlgraph.CreateSpec) {
	var (
		_node = &ActivityEntity{config: aec.config}
		_spec = sqlgraph.NewCreateSpec(activityentity.Table, sqlgraph.NewFieldSpec(activityentity.FieldID, field.TypeInt))
	)
	if id, ok := aec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := aec.mutation.HandlerName(); ok {
		_spec.SetField(activityentity.FieldHandlerName, field.TypeString, value)
		_node.HandlerName = value
	}
	if value, ok := aec.mutation.GetType(); ok {
		_spec.SetField(activityentity.FieldType, field.TypeString, value)
		_node.Type = value
	}
	if value, ok := aec.mutation.Status(); ok {
		_spec.SetField(activityentity.FieldStatus, field.TypeString, value)
		_node.Status = value
	}
	if value, ok := aec.mutation.StepID(); ok {
		_spec.SetField(activityentity.FieldStepID, field.TypeString, value)
		_node.StepID = value
	}
	if value, ok := aec.mutation.RunID(); ok {
		_spec.SetField(activityentity.FieldRunID, field.TypeInt, value)
		_node.RunID = value
	}
	if value, ok := aec.mutation.RetryPolicy(); ok {
		_spec.SetField(activityentity.FieldRetryPolicy, field.TypeJSON, value)
		_node.RetryPolicy = value
	}
	if value, ok := aec.mutation.RetryState(); ok {
		_spec.SetField(activityentity.FieldRetryState, field.TypeJSON, value)
		_node.RetryState = value
	}
	if value, ok := aec.mutation.CreatedAt(); ok {
		_spec.SetField(activityentity.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := aec.mutation.UpdatedAt(); ok {
		_spec.SetField(activityentity.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := aec.mutation.WorkflowIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   activityentity.WorkflowTable,
			Columns: []string{activityentity.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.workflow_entity_activity_children = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := aec.mutation.ActivityDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityentity.ActivityDataTable,
			Columns: []string{activityentity.ActivityDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activitydata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := aec.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   activityentity.ExecutionsTable,
			Columns: []string{activityentity.ExecutionsColumn},
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
	return _node, _spec
}

// ActivityEntityCreateBulk is the builder for creating many ActivityEntity entities in bulk.
type ActivityEntityCreateBulk struct {
	config
	err      error
	builders []*ActivityEntityCreate
}

// Save creates the ActivityEntity entities in the database.
func (aecb *ActivityEntityCreateBulk) Save(ctx context.Context) ([]*ActivityEntity, error) {
	if aecb.err != nil {
		return nil, aecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(aecb.builders))
	nodes := make([]*ActivityEntity, len(aecb.builders))
	mutators := make([]Mutator, len(aecb.builders))
	for i := range aecb.builders {
		func(i int, root context.Context) {
			builder := aecb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ActivityEntityMutation)
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
				if specs[i].ID.Value != nil && nodes[i].ID == 0 {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = schema.ActivityEntityID(id)
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
func (aecb *ActivityEntityCreateBulk) SaveX(ctx context.Context) []*ActivityEntity {
	v, err := aecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (aecb *ActivityEntityCreateBulk) Exec(ctx context.Context) error {
	_, err := aecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aecb *ActivityEntityCreateBulk) ExecX(ctx context.Context) {
	if err := aecb.Exec(ctx); err != nil {
		panic(err)
	}
}
