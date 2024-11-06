// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagadata"
)

// SagaDataCreate is the builder for creating a SagaData entity.
type SagaDataCreate struct {
	config
	mutation *SagaDataMutation
	hooks    []Hook
}

// SetCompensating sets the "compensating" field.
func (sdc *SagaDataCreate) SetCompensating(b bool) *SagaDataCreate {
	sdc.mutation.SetCompensating(b)
	return sdc
}

// SetNillableCompensating sets the "compensating" field if the given value is not nil.
func (sdc *SagaDataCreate) SetNillableCompensating(b *bool) *SagaDataCreate {
	if b != nil {
		sdc.SetCompensating(*b)
	}
	return sdc
}

// SetCompensationData sets the "compensation_data" field.
func (sdc *SagaDataCreate) SetCompensationData(u [][]uint8) *SagaDataCreate {
	sdc.mutation.SetCompensationData(u)
	return sdc
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (sdc *SagaDataCreate) SetEntityID(id int) *SagaDataCreate {
	sdc.mutation.SetEntityID(id)
	return sdc
}

// SetEntity sets the "entity" edge to the Entity entity.
func (sdc *SagaDataCreate) SetEntity(e *Entity) *SagaDataCreate {
	return sdc.SetEntityID(e.ID)
}

// Mutation returns the SagaDataMutation object of the builder.
func (sdc *SagaDataCreate) Mutation() *SagaDataMutation {
	return sdc.mutation
}

// Save creates the SagaData in the database.
func (sdc *SagaDataCreate) Save(ctx context.Context) (*SagaData, error) {
	sdc.defaults()
	return withHooks(ctx, sdc.sqlSave, sdc.mutation, sdc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sdc *SagaDataCreate) SaveX(ctx context.Context) *SagaData {
	v, err := sdc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sdc *SagaDataCreate) Exec(ctx context.Context) error {
	_, err := sdc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sdc *SagaDataCreate) ExecX(ctx context.Context) {
	if err := sdc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sdc *SagaDataCreate) defaults() {
	if _, ok := sdc.mutation.Compensating(); !ok {
		v := sagadata.DefaultCompensating
		sdc.mutation.SetCompensating(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sdc *SagaDataCreate) check() error {
	if _, ok := sdc.mutation.Compensating(); !ok {
		return &ValidationError{Name: "compensating", err: errors.New(`ent: missing required field "SagaData.compensating"`)}
	}
	if len(sdc.mutation.EntityIDs()) == 0 {
		return &ValidationError{Name: "entity", err: errors.New(`ent: missing required edge "SagaData.entity"`)}
	}
	return nil
}

func (sdc *SagaDataCreate) sqlSave(ctx context.Context) (*SagaData, error) {
	if err := sdc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sdc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sdc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	sdc.mutation.id = &_node.ID
	sdc.mutation.done = true
	return _node, nil
}

func (sdc *SagaDataCreate) createSpec() (*SagaData, *sqlgraph.CreateSpec) {
	var (
		_node = &SagaData{config: sdc.config}
		_spec = sqlgraph.NewCreateSpec(sagadata.Table, sqlgraph.NewFieldSpec(sagadata.FieldID, field.TypeInt))
	)
	if value, ok := sdc.mutation.Compensating(); ok {
		_spec.SetField(sagadata.FieldCompensating, field.TypeBool, value)
		_node.Compensating = value
	}
	if value, ok := sdc.mutation.CompensationData(); ok {
		_spec.SetField(sagadata.FieldCompensationData, field.TypeJSON, value)
		_node.CompensationData = value
	}
	if nodes := sdc.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagadata.EntityTable,
			Columns: []string{sagadata.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entity_saga_data = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SagaDataCreateBulk is the builder for creating many SagaData entities in bulk.
type SagaDataCreateBulk struct {
	config
	err      error
	builders []*SagaDataCreate
}

// Save creates the SagaData entities in the database.
func (sdcb *SagaDataCreateBulk) Save(ctx context.Context) ([]*SagaData, error) {
	if sdcb.err != nil {
		return nil, sdcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(sdcb.builders))
	nodes := make([]*SagaData, len(sdcb.builders))
	mutators := make([]Mutator, len(sdcb.builders))
	for i := range sdcb.builders {
		func(i int, root context.Context) {
			builder := sdcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SagaDataMutation)
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
					_, err = mutators[i+1].Mutate(root, sdcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, sdcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, sdcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (sdcb *SagaDataCreateBulk) SaveX(ctx context.Context) []*SagaData {
	v, err := sdcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sdcb *SagaDataCreateBulk) Exec(ctx context.Context) error {
	_, err := sdcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sdcb *SagaDataCreateBulk) ExecX(ctx context.Context) {
	if err := sdcb.Exec(ctx); err != nil {
		panic(err)
	}
}
