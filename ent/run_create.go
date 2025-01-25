// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/hierarchy"
	"github.com/davidroman0O/tempolite/ent/run"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
)

// RunCreate is the builder for creating a Run entity.
type RunCreate struct {
	config
	mutation *RunMutation
	hooks    []Hook
}

// SetStatus sets the "status" field.
func (rc *RunCreate) SetStatus(ss schema.RunStatus) *RunCreate {
	rc.mutation.SetStatus(ss)
	return rc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (rc *RunCreate) SetNillableStatus(ss *schema.RunStatus) *RunCreate {
	if ss != nil {
		rc.SetStatus(*ss)
	}
	return rc
}

// SetCreatedAt sets the "created_at" field.
func (rc *RunCreate) SetCreatedAt(t time.Time) *RunCreate {
	rc.mutation.SetCreatedAt(t)
	return rc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (rc *RunCreate) SetNillableCreatedAt(t *time.Time) *RunCreate {
	if t != nil {
		rc.SetCreatedAt(*t)
	}
	return rc
}

// SetUpdatedAt sets the "updated_at" field.
func (rc *RunCreate) SetUpdatedAt(t time.Time) *RunCreate {
	rc.mutation.SetUpdatedAt(t)
	return rc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (rc *RunCreate) SetNillableUpdatedAt(t *time.Time) *RunCreate {
	if t != nil {
		rc.SetUpdatedAt(*t)
	}
	return rc
}

// SetID sets the "id" field.
func (rc *RunCreate) SetID(si schema.RunID) *RunCreate {
	rc.mutation.SetID(si)
	return rc
}

// AddWorkflowIDs adds the "workflows" edge to the WorkflowEntity entity by IDs.
func (rc *RunCreate) AddWorkflowIDs(ids ...schema.WorkflowEntityID) *RunCreate {
	rc.mutation.AddWorkflowIDs(ids...)
	return rc
}

// AddWorkflows adds the "workflows" edges to the WorkflowEntity entity.
func (rc *RunCreate) AddWorkflows(w ...*WorkflowEntity) *RunCreate {
	ids := make([]schema.WorkflowEntityID, len(w))
	for i := range w {
		ids[i] = w[i].ID
	}
	return rc.AddWorkflowIDs(ids...)
}

// AddHierarchyIDs adds the "hierarchies" edge to the Hierarchy entity by IDs.
func (rc *RunCreate) AddHierarchyIDs(ids ...schema.HierarchyID) *RunCreate {
	rc.mutation.AddHierarchyIDs(ids...)
	return rc
}

// AddHierarchies adds the "hierarchies" edges to the Hierarchy entity.
func (rc *RunCreate) AddHierarchies(h ...*Hierarchy) *RunCreate {
	ids := make([]schema.HierarchyID, len(h))
	for i := range h {
		ids[i] = h[i].ID
	}
	return rc.AddHierarchyIDs(ids...)
}

// Mutation returns the RunMutation object of the builder.
func (rc *RunCreate) Mutation() *RunMutation {
	return rc.mutation
}

// Save creates the Run in the database.
func (rc *RunCreate) Save(ctx context.Context) (*Run, error) {
	rc.defaults()
	return withHooks(ctx, rc.sqlSave, rc.mutation, rc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (rc *RunCreate) SaveX(ctx context.Context) *Run {
	v, err := rc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rc *RunCreate) Exec(ctx context.Context) error {
	_, err := rc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rc *RunCreate) ExecX(ctx context.Context) {
	if err := rc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (rc *RunCreate) defaults() {
	if _, ok := rc.mutation.Status(); !ok {
		v := run.DefaultStatus
		rc.mutation.SetStatus(v)
	}
	if _, ok := rc.mutation.CreatedAt(); !ok {
		v := run.DefaultCreatedAt()
		rc.mutation.SetCreatedAt(v)
	}
	if _, ok := rc.mutation.UpdatedAt(); !ok {
		v := run.DefaultUpdatedAt()
		rc.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (rc *RunCreate) check() error {
	if _, ok := rc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Run.status"`)}
	}
	if _, ok := rc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Run.created_at"`)}
	}
	if _, ok := rc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Run.updated_at"`)}
	}
	return nil
}

func (rc *RunCreate) sqlSave(ctx context.Context) (*Run, error) {
	if err := rc.check(); err != nil {
		return nil, err
	}
	_node, _spec := rc.createSpec()
	if err := sqlgraph.CreateNode(ctx, rc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = schema.RunID(id)
	}
	rc.mutation.id = &_node.ID
	rc.mutation.done = true
	return _node, nil
}

func (rc *RunCreate) createSpec() (*Run, *sqlgraph.CreateSpec) {
	var (
		_node = &Run{config: rc.config}
		_spec = sqlgraph.NewCreateSpec(run.Table, sqlgraph.NewFieldSpec(run.FieldID, field.TypeInt))
	)
	if id, ok := rc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := rc.mutation.Status(); ok {
		_spec.SetField(run.FieldStatus, field.TypeString, value)
		_node.Status = value
	}
	if value, ok := rc.mutation.CreatedAt(); ok {
		_spec.SetField(run.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := rc.mutation.UpdatedAt(); ok {
		_spec.SetField(run.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := rc.mutation.WorkflowsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   run.WorkflowsTable,
			Columns: []string{run.WorkflowsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := rc.mutation.HierarchiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   run.HierarchiesTable,
			Columns: []string{run.HierarchiesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hierarchy.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// RunCreateBulk is the builder for creating many Run entities in bulk.
type RunCreateBulk struct {
	config
	err      error
	builders []*RunCreate
}

// Save creates the Run entities in the database.
func (rcb *RunCreateBulk) Save(ctx context.Context) ([]*Run, error) {
	if rcb.err != nil {
		return nil, rcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(rcb.builders))
	nodes := make([]*Run, len(rcb.builders))
	mutators := make([]Mutator, len(rcb.builders))
	for i := range rcb.builders {
		func(i int, root context.Context) {
			builder := rcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*RunMutation)
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
					_, err = mutators[i+1].Mutate(root, rcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, rcb.driver, spec); err != nil {
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
					nodes[i].ID = schema.RunID(id)
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
		if _, err := mutators[0].Mutate(ctx, rcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (rcb *RunCreateBulk) SaveX(ctx context.Context) []*Run {
	v, err := rcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rcb *RunCreateBulk) Exec(ctx context.Context) error {
	_, err := rcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcb *RunCreateBulk) ExecX(ctx context.Context) {
	if err := rcb.Exec(ctx); err != nil {
		panic(err)
	}
}
