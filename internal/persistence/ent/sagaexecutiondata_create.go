// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutiondata"
)

// SagaExecutionDataCreate is the builder for creating a SagaExecutionData entity.
type SagaExecutionDataCreate struct {
	config
	mutation *SagaExecutionDataMutation
	hooks    []Hook
}

// SetTransactionHistory sets the "transaction_history" field.
func (sedc *SagaExecutionDataCreate) SetTransactionHistory(u [][]uint8) *SagaExecutionDataCreate {
	sedc.mutation.SetTransactionHistory(u)
	return sedc
}

// SetCompensationHistory sets the "compensation_history" field.
func (sedc *SagaExecutionDataCreate) SetCompensationHistory(u [][]uint8) *SagaExecutionDataCreate {
	sedc.mutation.SetCompensationHistory(u)
	return sedc
}

// SetLastTransaction sets the "last_transaction" field.
func (sedc *SagaExecutionDataCreate) SetLastTransaction(t time.Time) *SagaExecutionDataCreate {
	sedc.mutation.SetLastTransaction(t)
	return sedc
}

// SetNillableLastTransaction sets the "last_transaction" field if the given value is not nil.
func (sedc *SagaExecutionDataCreate) SetNillableLastTransaction(t *time.Time) *SagaExecutionDataCreate {
	if t != nil {
		sedc.SetLastTransaction(*t)
	}
	return sedc
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (sedc *SagaExecutionDataCreate) SetSagaExecutionID(id int) *SagaExecutionDataCreate {
	sedc.mutation.SetSagaExecutionID(id)
	return sedc
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (sedc *SagaExecutionDataCreate) SetSagaExecution(s *SagaExecution) *SagaExecutionDataCreate {
	return sedc.SetSagaExecutionID(s.ID)
}

// Mutation returns the SagaExecutionDataMutation object of the builder.
func (sedc *SagaExecutionDataCreate) Mutation() *SagaExecutionDataMutation {
	return sedc.mutation
}

// Save creates the SagaExecutionData in the database.
func (sedc *SagaExecutionDataCreate) Save(ctx context.Context) (*SagaExecutionData, error) {
	return withHooks(ctx, sedc.sqlSave, sedc.mutation, sedc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sedc *SagaExecutionDataCreate) SaveX(ctx context.Context) *SagaExecutionData {
	v, err := sedc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sedc *SagaExecutionDataCreate) Exec(ctx context.Context) error {
	_, err := sedc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sedc *SagaExecutionDataCreate) ExecX(ctx context.Context) {
	if err := sedc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sedc *SagaExecutionDataCreate) check() error {
	if len(sedc.mutation.SagaExecutionIDs()) == 0 {
		return &ValidationError{Name: "saga_execution", err: errors.New(`ent: missing required edge "SagaExecutionData.saga_execution"`)}
	}
	return nil
}

func (sedc *SagaExecutionDataCreate) sqlSave(ctx context.Context) (*SagaExecutionData, error) {
	if err := sedc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sedc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sedc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	sedc.mutation.id = &_node.ID
	sedc.mutation.done = true
	return _node, nil
}

func (sedc *SagaExecutionDataCreate) createSpec() (*SagaExecutionData, *sqlgraph.CreateSpec) {
	var (
		_node = &SagaExecutionData{config: sedc.config}
		_spec = sqlgraph.NewCreateSpec(sagaexecutiondata.Table, sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt))
	)
	if value, ok := sedc.mutation.TransactionHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldTransactionHistory, field.TypeJSON, value)
		_node.TransactionHistory = value
	}
	if value, ok := sedc.mutation.CompensationHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldCompensationHistory, field.TypeJSON, value)
		_node.CompensationHistory = value
	}
	if value, ok := sedc.mutation.LastTransaction(); ok {
		_spec.SetField(sagaexecutiondata.FieldLastTransaction, field.TypeTime, value)
		_node.LastTransaction = &value
	}
	if nodes := sedc.mutation.SagaExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagaexecutiondata.SagaExecutionTable,
			Columns: []string{sagaexecutiondata.SagaExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.saga_execution_execution_data = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SagaExecutionDataCreateBulk is the builder for creating many SagaExecutionData entities in bulk.
type SagaExecutionDataCreateBulk struct {
	config
	err      error
	builders []*SagaExecutionDataCreate
}

// Save creates the SagaExecutionData entities in the database.
func (sedcb *SagaExecutionDataCreateBulk) Save(ctx context.Context) ([]*SagaExecutionData, error) {
	if sedcb.err != nil {
		return nil, sedcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(sedcb.builders))
	nodes := make([]*SagaExecutionData, len(sedcb.builders))
	mutators := make([]Mutator, len(sedcb.builders))
	for i := range sedcb.builders {
		func(i int, root context.Context) {
			builder := sedcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SagaExecutionDataMutation)
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
					_, err = mutators[i+1].Mutate(root, sedcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, sedcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, sedcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (sedcb *SagaExecutionDataCreateBulk) SaveX(ctx context.Context) []*SagaExecutionData {
	v, err := sedcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sedcb *SagaExecutionDataCreateBulk) Exec(ctx context.Context) error {
	_, err := sedcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sedcb *SagaExecutionDataCreateBulk) ExecX(ctx context.Context) {
	if err := sedcb.Exec(ctx); err != nil {
		panic(err)
	}
}