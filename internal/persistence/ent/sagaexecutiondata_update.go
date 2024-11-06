// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutiondata"
)

// SagaExecutionDataUpdate is the builder for updating SagaExecutionData entities.
type SagaExecutionDataUpdate struct {
	config
	hooks    []Hook
	mutation *SagaExecutionDataMutation
}

// Where appends a list predicates to the SagaExecutionDataUpdate builder.
func (sedu *SagaExecutionDataUpdate) Where(ps ...predicate.SagaExecutionData) *SagaExecutionDataUpdate {
	sedu.mutation.Where(ps...)
	return sedu
}

// SetTransactionHistory sets the "transaction_history" field.
func (sedu *SagaExecutionDataUpdate) SetTransactionHistory(u [][]uint8) *SagaExecutionDataUpdate {
	sedu.mutation.SetTransactionHistory(u)
	return sedu
}

// AppendTransactionHistory appends u to the "transaction_history" field.
func (sedu *SagaExecutionDataUpdate) AppendTransactionHistory(u [][]uint8) *SagaExecutionDataUpdate {
	sedu.mutation.AppendTransactionHistory(u)
	return sedu
}

// ClearTransactionHistory clears the value of the "transaction_history" field.
func (sedu *SagaExecutionDataUpdate) ClearTransactionHistory() *SagaExecutionDataUpdate {
	sedu.mutation.ClearTransactionHistory()
	return sedu
}

// SetCompensationHistory sets the "compensation_history" field.
func (sedu *SagaExecutionDataUpdate) SetCompensationHistory(u [][]uint8) *SagaExecutionDataUpdate {
	sedu.mutation.SetCompensationHistory(u)
	return sedu
}

// AppendCompensationHistory appends u to the "compensation_history" field.
func (sedu *SagaExecutionDataUpdate) AppendCompensationHistory(u [][]uint8) *SagaExecutionDataUpdate {
	sedu.mutation.AppendCompensationHistory(u)
	return sedu
}

// ClearCompensationHistory clears the value of the "compensation_history" field.
func (sedu *SagaExecutionDataUpdate) ClearCompensationHistory() *SagaExecutionDataUpdate {
	sedu.mutation.ClearCompensationHistory()
	return sedu
}

// SetLastTransaction sets the "last_transaction" field.
func (sedu *SagaExecutionDataUpdate) SetLastTransaction(t time.Time) *SagaExecutionDataUpdate {
	sedu.mutation.SetLastTransaction(t)
	return sedu
}

// SetNillableLastTransaction sets the "last_transaction" field if the given value is not nil.
func (sedu *SagaExecutionDataUpdate) SetNillableLastTransaction(t *time.Time) *SagaExecutionDataUpdate {
	if t != nil {
		sedu.SetLastTransaction(*t)
	}
	return sedu
}

// ClearLastTransaction clears the value of the "last_transaction" field.
func (sedu *SagaExecutionDataUpdate) ClearLastTransaction() *SagaExecutionDataUpdate {
	sedu.mutation.ClearLastTransaction()
	return sedu
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (sedu *SagaExecutionDataUpdate) SetSagaExecutionID(id int) *SagaExecutionDataUpdate {
	sedu.mutation.SetSagaExecutionID(id)
	return sedu
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (sedu *SagaExecutionDataUpdate) SetSagaExecution(s *SagaExecution) *SagaExecutionDataUpdate {
	return sedu.SetSagaExecutionID(s.ID)
}

// Mutation returns the SagaExecutionDataMutation object of the builder.
func (sedu *SagaExecutionDataUpdate) Mutation() *SagaExecutionDataMutation {
	return sedu.mutation
}

// ClearSagaExecution clears the "saga_execution" edge to the SagaExecution entity.
func (sedu *SagaExecutionDataUpdate) ClearSagaExecution() *SagaExecutionDataUpdate {
	sedu.mutation.ClearSagaExecution()
	return sedu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sedu *SagaExecutionDataUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, sedu.sqlSave, sedu.mutation, sedu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sedu *SagaExecutionDataUpdate) SaveX(ctx context.Context) int {
	affected, err := sedu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sedu *SagaExecutionDataUpdate) Exec(ctx context.Context) error {
	_, err := sedu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sedu *SagaExecutionDataUpdate) ExecX(ctx context.Context) {
	if err := sedu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sedu *SagaExecutionDataUpdate) check() error {
	if sedu.mutation.SagaExecutionCleared() && len(sedu.mutation.SagaExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecutionData.saga_execution"`)
	}
	return nil
}

func (sedu *SagaExecutionDataUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sedu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecutiondata.Table, sagaexecutiondata.Columns, sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt))
	if ps := sedu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sedu.mutation.TransactionHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldTransactionHistory, field.TypeJSON, value)
	}
	if value, ok := sedu.mutation.AppendedTransactionHistory(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecutiondata.FieldTransactionHistory, value)
		})
	}
	if sedu.mutation.TransactionHistoryCleared() {
		_spec.ClearField(sagaexecutiondata.FieldTransactionHistory, field.TypeJSON)
	}
	if value, ok := sedu.mutation.CompensationHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldCompensationHistory, field.TypeJSON, value)
	}
	if value, ok := sedu.mutation.AppendedCompensationHistory(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecutiondata.FieldCompensationHistory, value)
		})
	}
	if sedu.mutation.CompensationHistoryCleared() {
		_spec.ClearField(sagaexecutiondata.FieldCompensationHistory, field.TypeJSON)
	}
	if value, ok := sedu.mutation.LastTransaction(); ok {
		_spec.SetField(sagaexecutiondata.FieldLastTransaction, field.TypeTime, value)
	}
	if sedu.mutation.LastTransactionCleared() {
		_spec.ClearField(sagaexecutiondata.FieldLastTransaction, field.TypeTime)
	}
	if sedu.mutation.SagaExecutionCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sedu.mutation.SagaExecutionIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, sedu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sagaexecutiondata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sedu.mutation.done = true
	return n, nil
}

// SagaExecutionDataUpdateOne is the builder for updating a single SagaExecutionData entity.
type SagaExecutionDataUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *SagaExecutionDataMutation
}

// SetTransactionHistory sets the "transaction_history" field.
func (seduo *SagaExecutionDataUpdateOne) SetTransactionHistory(u [][]uint8) *SagaExecutionDataUpdateOne {
	seduo.mutation.SetTransactionHistory(u)
	return seduo
}

// AppendTransactionHistory appends u to the "transaction_history" field.
func (seduo *SagaExecutionDataUpdateOne) AppendTransactionHistory(u [][]uint8) *SagaExecutionDataUpdateOne {
	seduo.mutation.AppendTransactionHistory(u)
	return seduo
}

// ClearTransactionHistory clears the value of the "transaction_history" field.
func (seduo *SagaExecutionDataUpdateOne) ClearTransactionHistory() *SagaExecutionDataUpdateOne {
	seduo.mutation.ClearTransactionHistory()
	return seduo
}

// SetCompensationHistory sets the "compensation_history" field.
func (seduo *SagaExecutionDataUpdateOne) SetCompensationHistory(u [][]uint8) *SagaExecutionDataUpdateOne {
	seduo.mutation.SetCompensationHistory(u)
	return seduo
}

// AppendCompensationHistory appends u to the "compensation_history" field.
func (seduo *SagaExecutionDataUpdateOne) AppendCompensationHistory(u [][]uint8) *SagaExecutionDataUpdateOne {
	seduo.mutation.AppendCompensationHistory(u)
	return seduo
}

// ClearCompensationHistory clears the value of the "compensation_history" field.
func (seduo *SagaExecutionDataUpdateOne) ClearCompensationHistory() *SagaExecutionDataUpdateOne {
	seduo.mutation.ClearCompensationHistory()
	return seduo
}

// SetLastTransaction sets the "last_transaction" field.
func (seduo *SagaExecutionDataUpdateOne) SetLastTransaction(t time.Time) *SagaExecutionDataUpdateOne {
	seduo.mutation.SetLastTransaction(t)
	return seduo
}

// SetNillableLastTransaction sets the "last_transaction" field if the given value is not nil.
func (seduo *SagaExecutionDataUpdateOne) SetNillableLastTransaction(t *time.Time) *SagaExecutionDataUpdateOne {
	if t != nil {
		seduo.SetLastTransaction(*t)
	}
	return seduo
}

// ClearLastTransaction clears the value of the "last_transaction" field.
func (seduo *SagaExecutionDataUpdateOne) ClearLastTransaction() *SagaExecutionDataUpdateOne {
	seduo.mutation.ClearLastTransaction()
	return seduo
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (seduo *SagaExecutionDataUpdateOne) SetSagaExecutionID(id int) *SagaExecutionDataUpdateOne {
	seduo.mutation.SetSagaExecutionID(id)
	return seduo
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (seduo *SagaExecutionDataUpdateOne) SetSagaExecution(s *SagaExecution) *SagaExecutionDataUpdateOne {
	return seduo.SetSagaExecutionID(s.ID)
}

// Mutation returns the SagaExecutionDataMutation object of the builder.
func (seduo *SagaExecutionDataUpdateOne) Mutation() *SagaExecutionDataMutation {
	return seduo.mutation
}

// ClearSagaExecution clears the "saga_execution" edge to the SagaExecution entity.
func (seduo *SagaExecutionDataUpdateOne) ClearSagaExecution() *SagaExecutionDataUpdateOne {
	seduo.mutation.ClearSagaExecution()
	return seduo
}

// Where appends a list predicates to the SagaExecutionDataUpdate builder.
func (seduo *SagaExecutionDataUpdateOne) Where(ps ...predicate.SagaExecutionData) *SagaExecutionDataUpdateOne {
	seduo.mutation.Where(ps...)
	return seduo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (seduo *SagaExecutionDataUpdateOne) Select(field string, fields ...string) *SagaExecutionDataUpdateOne {
	seduo.fields = append([]string{field}, fields...)
	return seduo
}

// Save executes the query and returns the updated SagaExecutionData entity.
func (seduo *SagaExecutionDataUpdateOne) Save(ctx context.Context) (*SagaExecutionData, error) {
	return withHooks(ctx, seduo.sqlSave, seduo.mutation, seduo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seduo *SagaExecutionDataUpdateOne) SaveX(ctx context.Context) *SagaExecutionData {
	node, err := seduo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (seduo *SagaExecutionDataUpdateOne) Exec(ctx context.Context) error {
	_, err := seduo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seduo *SagaExecutionDataUpdateOne) ExecX(ctx context.Context) {
	if err := seduo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seduo *SagaExecutionDataUpdateOne) check() error {
	if seduo.mutation.SagaExecutionCleared() && len(seduo.mutation.SagaExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecutionData.saga_execution"`)
	}
	return nil
}

func (seduo *SagaExecutionDataUpdateOne) sqlSave(ctx context.Context) (_node *SagaExecutionData, err error) {
	if err := seduo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecutiondata.Table, sagaexecutiondata.Columns, sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt))
	id, ok := seduo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SagaExecutionData.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := seduo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sagaexecutiondata.FieldID)
		for _, f := range fields {
			if !sagaexecutiondata.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != sagaexecutiondata.FieldID {
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
	if value, ok := seduo.mutation.TransactionHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldTransactionHistory, field.TypeJSON, value)
	}
	if value, ok := seduo.mutation.AppendedTransactionHistory(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecutiondata.FieldTransactionHistory, value)
		})
	}
	if seduo.mutation.TransactionHistoryCleared() {
		_spec.ClearField(sagaexecutiondata.FieldTransactionHistory, field.TypeJSON)
	}
	if value, ok := seduo.mutation.CompensationHistory(); ok {
		_spec.SetField(sagaexecutiondata.FieldCompensationHistory, field.TypeJSON, value)
	}
	if value, ok := seduo.mutation.AppendedCompensationHistory(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecutiondata.FieldCompensationHistory, value)
		})
	}
	if seduo.mutation.CompensationHistoryCleared() {
		_spec.ClearField(sagaexecutiondata.FieldCompensationHistory, field.TypeJSON)
	}
	if value, ok := seduo.mutation.LastTransaction(); ok {
		_spec.SetField(sagaexecutiondata.FieldLastTransaction, field.TypeTime, value)
	}
	if seduo.mutation.LastTransactionCleared() {
		_spec.ClearField(sagaexecutiondata.FieldLastTransaction, field.TypeTime)
	}
	if seduo.mutation.SagaExecutionCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seduo.mutation.SagaExecutionIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &SagaExecutionData{config: seduo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, seduo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sagaexecutiondata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	seduo.mutation.done = true
	return _node, nil
}
