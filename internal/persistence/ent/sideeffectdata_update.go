// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectdata"
)

// SideEffectDataUpdate is the builder for updating SideEffectData entities.
type SideEffectDataUpdate struct {
	config
	hooks    []Hook
	mutation *SideEffectDataMutation
}

// Where appends a list predicates to the SideEffectDataUpdate builder.
func (sedu *SideEffectDataUpdate) Where(ps ...predicate.SideEffectData) *SideEffectDataUpdate {
	sedu.mutation.Where(ps...)
	return sedu
}

// SetInput sets the "input" field.
func (sedu *SideEffectDataUpdate) SetInput(u [][]uint8) *SideEffectDataUpdate {
	sedu.mutation.SetInput(u)
	return sedu
}

// AppendInput appends u to the "input" field.
func (sedu *SideEffectDataUpdate) AppendInput(u [][]uint8) *SideEffectDataUpdate {
	sedu.mutation.AppendInput(u)
	return sedu
}

// ClearInput clears the value of the "input" field.
func (sedu *SideEffectDataUpdate) ClearInput() *SideEffectDataUpdate {
	sedu.mutation.ClearInput()
	return sedu
}

// SetOutput sets the "output" field.
func (sedu *SideEffectDataUpdate) SetOutput(u [][]uint8) *SideEffectDataUpdate {
	sedu.mutation.SetOutput(u)
	return sedu
}

// AppendOutput appends u to the "output" field.
func (sedu *SideEffectDataUpdate) AppendOutput(u [][]uint8) *SideEffectDataUpdate {
	sedu.mutation.AppendOutput(u)
	return sedu
}

// ClearOutput clears the value of the "output" field.
func (sedu *SideEffectDataUpdate) ClearOutput() *SideEffectDataUpdate {
	sedu.mutation.ClearOutput()
	return sedu
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (sedu *SideEffectDataUpdate) SetEntityID(id int) *SideEffectDataUpdate {
	sedu.mutation.SetEntityID(id)
	return sedu
}

// SetEntity sets the "entity" edge to the Entity entity.
func (sedu *SideEffectDataUpdate) SetEntity(e *Entity) *SideEffectDataUpdate {
	return sedu.SetEntityID(e.ID)
}

// Mutation returns the SideEffectDataMutation object of the builder.
func (sedu *SideEffectDataUpdate) Mutation() *SideEffectDataMutation {
	return sedu.mutation
}

// ClearEntity clears the "entity" edge to the Entity entity.
func (sedu *SideEffectDataUpdate) ClearEntity() *SideEffectDataUpdate {
	sedu.mutation.ClearEntity()
	return sedu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sedu *SideEffectDataUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, sedu.sqlSave, sedu.mutation, sedu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sedu *SideEffectDataUpdate) SaveX(ctx context.Context) int {
	affected, err := sedu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sedu *SideEffectDataUpdate) Exec(ctx context.Context) error {
	_, err := sedu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sedu *SideEffectDataUpdate) ExecX(ctx context.Context) {
	if err := sedu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sedu *SideEffectDataUpdate) check() error {
	if sedu.mutation.EntityCleared() && len(sedu.mutation.EntityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SideEffectData.entity"`)
	}
	return nil
}

func (sedu *SideEffectDataUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sedu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(sideeffectdata.Table, sideeffectdata.Columns, sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt))
	if ps := sedu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sedu.mutation.Input(); ok {
		_spec.SetField(sideeffectdata.FieldInput, field.TypeJSON, value)
	}
	if value, ok := sedu.mutation.AppendedInput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sideeffectdata.FieldInput, value)
		})
	}
	if sedu.mutation.InputCleared() {
		_spec.ClearField(sideeffectdata.FieldInput, field.TypeJSON)
	}
	if value, ok := sedu.mutation.Output(); ok {
		_spec.SetField(sideeffectdata.FieldOutput, field.TypeJSON, value)
	}
	if value, ok := sedu.mutation.AppendedOutput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sideeffectdata.FieldOutput, value)
		})
	}
	if sedu.mutation.OutputCleared() {
		_spec.ClearField(sideeffectdata.FieldOutput, field.TypeJSON)
	}
	if sedu.mutation.EntityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sideeffectdata.EntityTable,
			Columns: []string{sideeffectdata.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sedu.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sideeffectdata.EntityTable,
			Columns: []string{sideeffectdata.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, sedu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sideeffectdata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sedu.mutation.done = true
	return n, nil
}

// SideEffectDataUpdateOne is the builder for updating a single SideEffectData entity.
type SideEffectDataUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *SideEffectDataMutation
}

// SetInput sets the "input" field.
func (seduo *SideEffectDataUpdateOne) SetInput(u [][]uint8) *SideEffectDataUpdateOne {
	seduo.mutation.SetInput(u)
	return seduo
}

// AppendInput appends u to the "input" field.
func (seduo *SideEffectDataUpdateOne) AppendInput(u [][]uint8) *SideEffectDataUpdateOne {
	seduo.mutation.AppendInput(u)
	return seduo
}

// ClearInput clears the value of the "input" field.
func (seduo *SideEffectDataUpdateOne) ClearInput() *SideEffectDataUpdateOne {
	seduo.mutation.ClearInput()
	return seduo
}

// SetOutput sets the "output" field.
func (seduo *SideEffectDataUpdateOne) SetOutput(u [][]uint8) *SideEffectDataUpdateOne {
	seduo.mutation.SetOutput(u)
	return seduo
}

// AppendOutput appends u to the "output" field.
func (seduo *SideEffectDataUpdateOne) AppendOutput(u [][]uint8) *SideEffectDataUpdateOne {
	seduo.mutation.AppendOutput(u)
	return seduo
}

// ClearOutput clears the value of the "output" field.
func (seduo *SideEffectDataUpdateOne) ClearOutput() *SideEffectDataUpdateOne {
	seduo.mutation.ClearOutput()
	return seduo
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (seduo *SideEffectDataUpdateOne) SetEntityID(id int) *SideEffectDataUpdateOne {
	seduo.mutation.SetEntityID(id)
	return seduo
}

// SetEntity sets the "entity" edge to the Entity entity.
func (seduo *SideEffectDataUpdateOne) SetEntity(e *Entity) *SideEffectDataUpdateOne {
	return seduo.SetEntityID(e.ID)
}

// Mutation returns the SideEffectDataMutation object of the builder.
func (seduo *SideEffectDataUpdateOne) Mutation() *SideEffectDataMutation {
	return seduo.mutation
}

// ClearEntity clears the "entity" edge to the Entity entity.
func (seduo *SideEffectDataUpdateOne) ClearEntity() *SideEffectDataUpdateOne {
	seduo.mutation.ClearEntity()
	return seduo
}

// Where appends a list predicates to the SideEffectDataUpdate builder.
func (seduo *SideEffectDataUpdateOne) Where(ps ...predicate.SideEffectData) *SideEffectDataUpdateOne {
	seduo.mutation.Where(ps...)
	return seduo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (seduo *SideEffectDataUpdateOne) Select(field string, fields ...string) *SideEffectDataUpdateOne {
	seduo.fields = append([]string{field}, fields...)
	return seduo
}

// Save executes the query and returns the updated SideEffectData entity.
func (seduo *SideEffectDataUpdateOne) Save(ctx context.Context) (*SideEffectData, error) {
	return withHooks(ctx, seduo.sqlSave, seduo.mutation, seduo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seduo *SideEffectDataUpdateOne) SaveX(ctx context.Context) *SideEffectData {
	node, err := seduo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (seduo *SideEffectDataUpdateOne) Exec(ctx context.Context) error {
	_, err := seduo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seduo *SideEffectDataUpdateOne) ExecX(ctx context.Context) {
	if err := seduo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seduo *SideEffectDataUpdateOne) check() error {
	if seduo.mutation.EntityCleared() && len(seduo.mutation.EntityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SideEffectData.entity"`)
	}
	return nil
}

func (seduo *SideEffectDataUpdateOne) sqlSave(ctx context.Context) (_node *SideEffectData, err error) {
	if err := seduo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(sideeffectdata.Table, sideeffectdata.Columns, sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt))
	id, ok := seduo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SideEffectData.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := seduo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sideeffectdata.FieldID)
		for _, f := range fields {
			if !sideeffectdata.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != sideeffectdata.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := seduo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seduo.mutation.Input(); ok {
		_spec.SetField(sideeffectdata.FieldInput, field.TypeJSON, value)
	}
	if value, ok := seduo.mutation.AppendedInput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sideeffectdata.FieldInput, value)
		})
	}
	if seduo.mutation.InputCleared() {
		_spec.ClearField(sideeffectdata.FieldInput, field.TypeJSON)
	}
	if value, ok := seduo.mutation.Output(); ok {
		_spec.SetField(sideeffectdata.FieldOutput, field.TypeJSON, value)
	}
	if value, ok := seduo.mutation.AppendedOutput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sideeffectdata.FieldOutput, value)
		})
	}
	if seduo.mutation.OutputCleared() {
		_spec.ClearField(sideeffectdata.FieldOutput, field.TypeJSON)
	}
	if seduo.mutation.EntityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sideeffectdata.EntityTable,
			Columns: []string{sideeffectdata.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seduo.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sideeffectdata.EntityTable,
			Columns: []string{sideeffectdata.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &SideEffectData{config: seduo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, seduo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sideeffectdata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	seduo.mutation.done = true
	return _node, nil
}
