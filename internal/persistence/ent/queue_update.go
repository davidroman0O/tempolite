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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
)

// QueueUpdate is the builder for updating Queue entities.
type QueueUpdate struct {
	config
	hooks    []Hook
	mutation *QueueMutation
}

// Where appends a list predicates to the QueueUpdate builder.
func (qu *QueueUpdate) Where(ps ...predicate.Queue) *QueueUpdate {
	qu.mutation.Where(ps...)
	return qu
}

// SetUpdatedAt sets the "updated_at" field.
func (qu *QueueUpdate) SetUpdatedAt(t time.Time) *QueueUpdate {
	qu.mutation.SetUpdatedAt(t)
	return qu
}

// SetName sets the "name" field.
func (qu *QueueUpdate) SetName(s string) *QueueUpdate {
	qu.mutation.SetName(s)
	return qu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (qu *QueueUpdate) SetNillableName(s *string) *QueueUpdate {
	if s != nil {
		qu.SetName(*s)
	}
	return qu
}

// AddEntityIDs adds the "entities" edge to the Entity entity by IDs.
func (qu *QueueUpdate) AddEntityIDs(ids ...int) *QueueUpdate {
	qu.mutation.AddEntityIDs(ids...)
	return qu
}

// AddEntities adds the "entities" edges to the Entity entity.
func (qu *QueueUpdate) AddEntities(e ...*Entity) *QueueUpdate {
	ids := make([]int, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return qu.AddEntityIDs(ids...)
}

// Mutation returns the QueueMutation object of the builder.
func (qu *QueueUpdate) Mutation() *QueueMutation {
	return qu.mutation
}

// ClearEntities clears all "entities" edges to the Entity entity.
func (qu *QueueUpdate) ClearEntities() *QueueUpdate {
	qu.mutation.ClearEntities()
	return qu
}

// RemoveEntityIDs removes the "entities" edge to Entity entities by IDs.
func (qu *QueueUpdate) RemoveEntityIDs(ids ...int) *QueueUpdate {
	qu.mutation.RemoveEntityIDs(ids...)
	return qu
}

// RemoveEntities removes "entities" edges to Entity entities.
func (qu *QueueUpdate) RemoveEntities(e ...*Entity) *QueueUpdate {
	ids := make([]int, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return qu.RemoveEntityIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (qu *QueueUpdate) Save(ctx context.Context) (int, error) {
	qu.defaults()
	return withHooks(ctx, qu.sqlSave, qu.mutation, qu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (qu *QueueUpdate) SaveX(ctx context.Context) int {
	affected, err := qu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (qu *QueueUpdate) Exec(ctx context.Context) error {
	_, err := qu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (qu *QueueUpdate) ExecX(ctx context.Context) {
	if err := qu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (qu *QueueUpdate) defaults() {
	if _, ok := qu.mutation.UpdatedAt(); !ok {
		v := queue.UpdateDefaultUpdatedAt()
		qu.mutation.SetUpdatedAt(v)
	}
}

func (qu *QueueUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(queue.Table, queue.Columns, sqlgraph.NewFieldSpec(queue.FieldID, field.TypeInt))
	if ps := qu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := qu.mutation.UpdatedAt(); ok {
		_spec.SetField(queue.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := qu.mutation.Name(); ok {
		_spec.SetField(queue.FieldName, field.TypeString, value)
	}
	if qu.mutation.EntitiesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := qu.mutation.RemovedEntitiesIDs(); len(nodes) > 0 && !qu.mutation.EntitiesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := qu.mutation.EntitiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
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
	if n, err = sqlgraph.UpdateNodes(ctx, qu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{queue.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	qu.mutation.done = true
	return n, nil
}

// QueueUpdateOne is the builder for updating a single Queue entity.
type QueueUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *QueueMutation
}

// SetUpdatedAt sets the "updated_at" field.
func (quo *QueueUpdateOne) SetUpdatedAt(t time.Time) *QueueUpdateOne {
	quo.mutation.SetUpdatedAt(t)
	return quo
}

// SetName sets the "name" field.
func (quo *QueueUpdateOne) SetName(s string) *QueueUpdateOne {
	quo.mutation.SetName(s)
	return quo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (quo *QueueUpdateOne) SetNillableName(s *string) *QueueUpdateOne {
	if s != nil {
		quo.SetName(*s)
	}
	return quo
}

// AddEntityIDs adds the "entities" edge to the Entity entity by IDs.
func (quo *QueueUpdateOne) AddEntityIDs(ids ...int) *QueueUpdateOne {
	quo.mutation.AddEntityIDs(ids...)
	return quo
}

// AddEntities adds the "entities" edges to the Entity entity.
func (quo *QueueUpdateOne) AddEntities(e ...*Entity) *QueueUpdateOne {
	ids := make([]int, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return quo.AddEntityIDs(ids...)
}

// Mutation returns the QueueMutation object of the builder.
func (quo *QueueUpdateOne) Mutation() *QueueMutation {
	return quo.mutation
}

// ClearEntities clears all "entities" edges to the Entity entity.
func (quo *QueueUpdateOne) ClearEntities() *QueueUpdateOne {
	quo.mutation.ClearEntities()
	return quo
}

// RemoveEntityIDs removes the "entities" edge to Entity entities by IDs.
func (quo *QueueUpdateOne) RemoveEntityIDs(ids ...int) *QueueUpdateOne {
	quo.mutation.RemoveEntityIDs(ids...)
	return quo
}

// RemoveEntities removes "entities" edges to Entity entities.
func (quo *QueueUpdateOne) RemoveEntities(e ...*Entity) *QueueUpdateOne {
	ids := make([]int, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return quo.RemoveEntityIDs(ids...)
}

// Where appends a list predicates to the QueueUpdate builder.
func (quo *QueueUpdateOne) Where(ps ...predicate.Queue) *QueueUpdateOne {
	quo.mutation.Where(ps...)
	return quo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (quo *QueueUpdateOne) Select(field string, fields ...string) *QueueUpdateOne {
	quo.fields = append([]string{field}, fields...)
	return quo
}

// Save executes the query and returns the updated Queue entity.
func (quo *QueueUpdateOne) Save(ctx context.Context) (*Queue, error) {
	quo.defaults()
	return withHooks(ctx, quo.sqlSave, quo.mutation, quo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (quo *QueueUpdateOne) SaveX(ctx context.Context) *Queue {
	node, err := quo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (quo *QueueUpdateOne) Exec(ctx context.Context) error {
	_, err := quo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (quo *QueueUpdateOne) ExecX(ctx context.Context) {
	if err := quo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (quo *QueueUpdateOne) defaults() {
	if _, ok := quo.mutation.UpdatedAt(); !ok {
		v := queue.UpdateDefaultUpdatedAt()
		quo.mutation.SetUpdatedAt(v)
	}
}

func (quo *QueueUpdateOne) sqlSave(ctx context.Context) (_node *Queue, err error) {
	_spec := sqlgraph.NewUpdateSpec(queue.Table, queue.Columns, sqlgraph.NewFieldSpec(queue.FieldID, field.TypeInt))
	id, ok := quo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Queue.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := quo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, queue.FieldID)
		for _, f := range fields {
			if !queue.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != queue.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := quo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := quo.mutation.UpdatedAt(); ok {
		_spec.SetField(queue.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := quo.mutation.Name(); ok {
		_spec.SetField(queue.FieldName, field.TypeString, value)
	}
	if quo.mutation.EntitiesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := quo.mutation.RemovedEntitiesIDs(); len(nodes) > 0 && !quo.mutation.EntitiesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := quo.mutation.EntitiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   queue.EntitiesTable,
			Columns: queue.EntitiesPrimaryKey,
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
	_node = &Queue{config: quo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, quo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{queue.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	quo.mutation.done = true
	return _node, nil
}
