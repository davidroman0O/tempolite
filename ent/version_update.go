// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/version"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
)

// VersionUpdate is the builder for updating Version entities.
type VersionUpdate struct {
	config
	hooks    []Hook
	mutation *VersionMutation
}

// Where appends a list predicates to the VersionUpdate builder.
func (vu *VersionUpdate) Where(ps ...predicate.Version) *VersionUpdate {
	vu.mutation.Where(ps...)
	return vu
}

// SetEntityID sets the "entity_id" field.
func (vu *VersionUpdate) SetEntityID(sei schema.WorkflowEntityID) *VersionUpdate {
	vu.mutation.SetEntityID(sei)
	return vu
}

// SetNillableEntityID sets the "entity_id" field if the given value is not nil.
func (vu *VersionUpdate) SetNillableEntityID(sei *schema.WorkflowEntityID) *VersionUpdate {
	if sei != nil {
		vu.SetEntityID(*sei)
	}
	return vu
}

// SetChangeID sets the "change_id" field.
func (vu *VersionUpdate) SetChangeID(sc schema.VersionChange) *VersionUpdate {
	vu.mutation.SetChangeID(sc)
	return vu
}

// SetNillableChangeID sets the "change_id" field if the given value is not nil.
func (vu *VersionUpdate) SetNillableChangeID(sc *schema.VersionChange) *VersionUpdate {
	if sc != nil {
		vu.SetChangeID(*sc)
	}
	return vu
}

// SetVersion sets the "version" field.
func (vu *VersionUpdate) SetVersion(sn schema.VersionNumber) *VersionUpdate {
	vu.mutation.ResetVersion()
	vu.mutation.SetVersion(sn)
	return vu
}

// SetNillableVersion sets the "version" field if the given value is not nil.
func (vu *VersionUpdate) SetNillableVersion(sn *schema.VersionNumber) *VersionUpdate {
	if sn != nil {
		vu.SetVersion(*sn)
	}
	return vu
}

// AddVersion adds sn to the "version" field.
func (vu *VersionUpdate) AddVersion(sn schema.VersionNumber) *VersionUpdate {
	vu.mutation.AddVersion(sn)
	return vu
}

// SetData sets the "data" field.
func (vu *VersionUpdate) SetData(m map[string]interface{}) *VersionUpdate {
	vu.mutation.SetData(m)
	return vu
}

// SetCreatedAt sets the "created_at" field.
func (vu *VersionUpdate) SetCreatedAt(t time.Time) *VersionUpdate {
	vu.mutation.SetCreatedAt(t)
	return vu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (vu *VersionUpdate) SetNillableCreatedAt(t *time.Time) *VersionUpdate {
	if t != nil {
		vu.SetCreatedAt(*t)
	}
	return vu
}

// SetUpdatedAt sets the "updated_at" field.
func (vu *VersionUpdate) SetUpdatedAt(t time.Time) *VersionUpdate {
	vu.mutation.SetUpdatedAt(t)
	return vu
}

// SetWorkflowID sets the "workflow" edge to the WorkflowEntity entity by ID.
func (vu *VersionUpdate) SetWorkflowID(id schema.WorkflowEntityID) *VersionUpdate {
	vu.mutation.SetWorkflowID(id)
	return vu
}

// SetWorkflow sets the "workflow" edge to the WorkflowEntity entity.
func (vu *VersionUpdate) SetWorkflow(w *WorkflowEntity) *VersionUpdate {
	return vu.SetWorkflowID(w.ID)
}

// Mutation returns the VersionMutation object of the builder.
func (vu *VersionUpdate) Mutation() *VersionMutation {
	return vu.mutation
}

// ClearWorkflow clears the "workflow" edge to the WorkflowEntity entity.
func (vu *VersionUpdate) ClearWorkflow() *VersionUpdate {
	vu.mutation.ClearWorkflow()
	return vu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (vu *VersionUpdate) Save(ctx context.Context) (int, error) {
	vu.defaults()
	return withHooks(ctx, vu.sqlSave, vu.mutation, vu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (vu *VersionUpdate) SaveX(ctx context.Context) int {
	affected, err := vu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (vu *VersionUpdate) Exec(ctx context.Context) error {
	_, err := vu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vu *VersionUpdate) ExecX(ctx context.Context) {
	if err := vu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (vu *VersionUpdate) defaults() {
	if _, ok := vu.mutation.UpdatedAt(); !ok {
		v := version.UpdateDefaultUpdatedAt()
		vu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (vu *VersionUpdate) check() error {
	if vu.mutation.WorkflowCleared() && len(vu.mutation.WorkflowIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Version.workflow"`)
	}
	return nil
}

func (vu *VersionUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := vu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(version.Table, version.Columns, sqlgraph.NewFieldSpec(version.FieldID, field.TypeInt))
	if ps := vu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := vu.mutation.ChangeID(); ok {
		_spec.SetField(version.FieldChangeID, field.TypeString, value)
	}
	if value, ok := vu.mutation.Version(); ok {
		_spec.SetField(version.FieldVersion, field.TypeUint, value)
	}
	if value, ok := vu.mutation.AddedVersion(); ok {
		_spec.AddField(version.FieldVersion, field.TypeUint, value)
	}
	if value, ok := vu.mutation.Data(); ok {
		_spec.SetField(version.FieldData, field.TypeJSON, value)
	}
	if value, ok := vu.mutation.CreatedAt(); ok {
		_spec.SetField(version.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := vu.mutation.UpdatedAt(); ok {
		_spec.SetField(version.FieldUpdatedAt, field.TypeTime, value)
	}
	if vu.mutation.WorkflowCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   version.WorkflowTable,
			Columns: []string{version.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := vu.mutation.WorkflowIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   version.WorkflowTable,
			Columns: []string{version.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, vu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{version.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	vu.mutation.done = true
	return n, nil
}

// VersionUpdateOne is the builder for updating a single Version entity.
type VersionUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *VersionMutation
}

// SetEntityID sets the "entity_id" field.
func (vuo *VersionUpdateOne) SetEntityID(sei schema.WorkflowEntityID) *VersionUpdateOne {
	vuo.mutation.SetEntityID(sei)
	return vuo
}

// SetNillableEntityID sets the "entity_id" field if the given value is not nil.
func (vuo *VersionUpdateOne) SetNillableEntityID(sei *schema.WorkflowEntityID) *VersionUpdateOne {
	if sei != nil {
		vuo.SetEntityID(*sei)
	}
	return vuo
}

// SetChangeID sets the "change_id" field.
func (vuo *VersionUpdateOne) SetChangeID(sc schema.VersionChange) *VersionUpdateOne {
	vuo.mutation.SetChangeID(sc)
	return vuo
}

// SetNillableChangeID sets the "change_id" field if the given value is not nil.
func (vuo *VersionUpdateOne) SetNillableChangeID(sc *schema.VersionChange) *VersionUpdateOne {
	if sc != nil {
		vuo.SetChangeID(*sc)
	}
	return vuo
}

// SetVersion sets the "version" field.
func (vuo *VersionUpdateOne) SetVersion(sn schema.VersionNumber) *VersionUpdateOne {
	vuo.mutation.ResetVersion()
	vuo.mutation.SetVersion(sn)
	return vuo
}

// SetNillableVersion sets the "version" field if the given value is not nil.
func (vuo *VersionUpdateOne) SetNillableVersion(sn *schema.VersionNumber) *VersionUpdateOne {
	if sn != nil {
		vuo.SetVersion(*sn)
	}
	return vuo
}

// AddVersion adds sn to the "version" field.
func (vuo *VersionUpdateOne) AddVersion(sn schema.VersionNumber) *VersionUpdateOne {
	vuo.mutation.AddVersion(sn)
	return vuo
}

// SetData sets the "data" field.
func (vuo *VersionUpdateOne) SetData(m map[string]interface{}) *VersionUpdateOne {
	vuo.mutation.SetData(m)
	return vuo
}

// SetCreatedAt sets the "created_at" field.
func (vuo *VersionUpdateOne) SetCreatedAt(t time.Time) *VersionUpdateOne {
	vuo.mutation.SetCreatedAt(t)
	return vuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (vuo *VersionUpdateOne) SetNillableCreatedAt(t *time.Time) *VersionUpdateOne {
	if t != nil {
		vuo.SetCreatedAt(*t)
	}
	return vuo
}

// SetUpdatedAt sets the "updated_at" field.
func (vuo *VersionUpdateOne) SetUpdatedAt(t time.Time) *VersionUpdateOne {
	vuo.mutation.SetUpdatedAt(t)
	return vuo
}

// SetWorkflowID sets the "workflow" edge to the WorkflowEntity entity by ID.
func (vuo *VersionUpdateOne) SetWorkflowID(id schema.WorkflowEntityID) *VersionUpdateOne {
	vuo.mutation.SetWorkflowID(id)
	return vuo
}

// SetWorkflow sets the "workflow" edge to the WorkflowEntity entity.
func (vuo *VersionUpdateOne) SetWorkflow(w *WorkflowEntity) *VersionUpdateOne {
	return vuo.SetWorkflowID(w.ID)
}

// Mutation returns the VersionMutation object of the builder.
func (vuo *VersionUpdateOne) Mutation() *VersionMutation {
	return vuo.mutation
}

// ClearWorkflow clears the "workflow" edge to the WorkflowEntity entity.
func (vuo *VersionUpdateOne) ClearWorkflow() *VersionUpdateOne {
	vuo.mutation.ClearWorkflow()
	return vuo
}

// Where appends a list predicates to the VersionUpdate builder.
func (vuo *VersionUpdateOne) Where(ps ...predicate.Version) *VersionUpdateOne {
	vuo.mutation.Where(ps...)
	return vuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (vuo *VersionUpdateOne) Select(field string, fields ...string) *VersionUpdateOne {
	vuo.fields = append([]string{field}, fields...)
	return vuo
}

// Save executes the query and returns the updated Version entity.
func (vuo *VersionUpdateOne) Save(ctx context.Context) (*Version, error) {
	vuo.defaults()
	return withHooks(ctx, vuo.sqlSave, vuo.mutation, vuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (vuo *VersionUpdateOne) SaveX(ctx context.Context) *Version {
	node, err := vuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (vuo *VersionUpdateOne) Exec(ctx context.Context) error {
	_, err := vuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vuo *VersionUpdateOne) ExecX(ctx context.Context) {
	if err := vuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (vuo *VersionUpdateOne) defaults() {
	if _, ok := vuo.mutation.UpdatedAt(); !ok {
		v := version.UpdateDefaultUpdatedAt()
		vuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (vuo *VersionUpdateOne) check() error {
	if vuo.mutation.WorkflowCleared() && len(vuo.mutation.WorkflowIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Version.workflow"`)
	}
	return nil
}

func (vuo *VersionUpdateOne) sqlSave(ctx context.Context) (_node *Version, err error) {
	if err := vuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(version.Table, version.Columns, sqlgraph.NewFieldSpec(version.FieldID, field.TypeInt))
	id, ok := vuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Version.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := vuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, version.FieldID)
		for _, f := range fields {
			if !version.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != version.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := vuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := vuo.mutation.ChangeID(); ok {
		_spec.SetField(version.FieldChangeID, field.TypeString, value)
	}
	if value, ok := vuo.mutation.Version(); ok {
		_spec.SetField(version.FieldVersion, field.TypeUint, value)
	}
	if value, ok := vuo.mutation.AddedVersion(); ok {
		_spec.AddField(version.FieldVersion, field.TypeUint, value)
	}
	if value, ok := vuo.mutation.Data(); ok {
		_spec.SetField(version.FieldData, field.TypeJSON, value)
	}
	if value, ok := vuo.mutation.CreatedAt(); ok {
		_spec.SetField(version.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := vuo.mutation.UpdatedAt(); ok {
		_spec.SetField(version.FieldUpdatedAt, field.TypeTime, value)
	}
	if vuo.mutation.WorkflowCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   version.WorkflowTable,
			Columns: []string{version.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := vuo.mutation.WorkflowIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   version.WorkflowTable,
			Columns: []string{version.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Version{config: vuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, vuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{version.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	vuo.mutation.done = true
	return _node, nil
}
