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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecutiondata"
)

// SagaExecutionUpdate is the builder for updating SagaExecution entities.
type SagaExecutionUpdate struct {
	config
	hooks    []Hook
	mutation *SagaExecutionMutation
}

// Where appends a list predicates to the SagaExecutionUpdate builder.
func (seu *SagaExecutionUpdate) Where(ps ...predicate.SagaExecution) *SagaExecutionUpdate {
	seu.mutation.Where(ps...)
	return seu
}

// SetStepType sets the "step_type" field.
func (seu *SagaExecutionUpdate) SetStepType(st sagaexecution.StepType) *SagaExecutionUpdate {
	seu.mutation.SetStepType(st)
	return seu
}

// SetNillableStepType sets the "step_type" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableStepType(st *sagaexecution.StepType) *SagaExecutionUpdate {
	if st != nil {
		seu.SetStepType(*st)
	}
	return seu
}

// SetCompensationData sets the "compensation_data" field.
func (seu *SagaExecutionUpdate) SetCompensationData(u []uint8) *SagaExecutionUpdate {
	seu.mutation.SetCompensationData(u)
	return seu
}

// AppendCompensationData appends u to the "compensation_data" field.
func (seu *SagaExecutionUpdate) AppendCompensationData(u []uint8) *SagaExecutionUpdate {
	seu.mutation.AppendCompensationData(u)
	return seu
}

// ClearCompensationData clears the value of the "compensation_data" field.
func (seu *SagaExecutionUpdate) ClearCompensationData() *SagaExecutionUpdate {
	seu.mutation.ClearCompensationData()
	return seu
}

// SetExecutionID sets the "execution" edge to the Execution entity by ID.
func (seu *SagaExecutionUpdate) SetExecutionID(id int) *SagaExecutionUpdate {
	seu.mutation.SetExecutionID(id)
	return seu
}

// SetExecution sets the "execution" edge to the Execution entity.
func (seu *SagaExecutionUpdate) SetExecution(e *Execution) *SagaExecutionUpdate {
	return seu.SetExecutionID(e.ID)
}

// SetExecutionDataID sets the "execution_data" edge to the SagaExecutionData entity by ID.
func (seu *SagaExecutionUpdate) SetExecutionDataID(id int) *SagaExecutionUpdate {
	seu.mutation.SetExecutionDataID(id)
	return seu
}

// SetNillableExecutionDataID sets the "execution_data" edge to the SagaExecutionData entity by ID if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableExecutionDataID(id *int) *SagaExecutionUpdate {
	if id != nil {
		seu = seu.SetExecutionDataID(*id)
	}
	return seu
}

// SetExecutionData sets the "execution_data" edge to the SagaExecutionData entity.
func (seu *SagaExecutionUpdate) SetExecutionData(s *SagaExecutionData) *SagaExecutionUpdate {
	return seu.SetExecutionDataID(s.ID)
}

// Mutation returns the SagaExecutionMutation object of the builder.
func (seu *SagaExecutionUpdate) Mutation() *SagaExecutionMutation {
	return seu.mutation
}

// ClearExecution clears the "execution" edge to the Execution entity.
func (seu *SagaExecutionUpdate) ClearExecution() *SagaExecutionUpdate {
	seu.mutation.ClearExecution()
	return seu
}

// ClearExecutionData clears the "execution_data" edge to the SagaExecutionData entity.
func (seu *SagaExecutionUpdate) ClearExecutionData() *SagaExecutionUpdate {
	seu.mutation.ClearExecutionData()
	return seu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (seu *SagaExecutionUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, seu.sqlSave, seu.mutation, seu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seu *SagaExecutionUpdate) SaveX(ctx context.Context) int {
	affected, err := seu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (seu *SagaExecutionUpdate) Exec(ctx context.Context) error {
	_, err := seu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seu *SagaExecutionUpdate) ExecX(ctx context.Context) {
	if err := seu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seu *SagaExecutionUpdate) check() error {
	if v, ok := seu.mutation.StepType(); ok {
		if err := sagaexecution.StepTypeValidator(v); err != nil {
			return &ValidationError{Name: "step_type", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.step_type": %w`, err)}
		}
	}
	if seu.mutation.ExecutionCleared() && len(seu.mutation.ExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecution.execution"`)
	}
	return nil
}

func (seu *SagaExecutionUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := seu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecution.Table, sagaexecution.Columns, sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt))
	if ps := seu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seu.mutation.StepType(); ok {
		_spec.SetField(sagaexecution.FieldStepType, field.TypeEnum, value)
	}
	if value, ok := seu.mutation.CompensationData(); ok {
		_spec.SetField(sagaexecution.FieldCompensationData, field.TypeJSON, value)
	}
	if value, ok := seu.mutation.AppendedCompensationData(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecution.FieldCompensationData, value)
		})
	}
	if seu.mutation.CompensationDataCleared() {
		_spec.ClearField(sagaexecution.FieldCompensationData, field.TypeJSON)
	}
	if seu.mutation.ExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagaexecution.ExecutionTable,
			Columns: []string{sagaexecution.ExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seu.mutation.ExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagaexecution.ExecutionTable,
			Columns: []string{sagaexecution.ExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seu.mutation.ExecutionDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sagaexecution.ExecutionDataTable,
			Columns: []string{sagaexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seu.mutation.ExecutionDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sagaexecution.ExecutionDataTable,
			Columns: []string{sagaexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, seu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sagaexecution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	seu.mutation.done = true
	return n, nil
}

// SagaExecutionUpdateOne is the builder for updating a single SagaExecution entity.
type SagaExecutionUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *SagaExecutionMutation
}

// SetStepType sets the "step_type" field.
func (seuo *SagaExecutionUpdateOne) SetStepType(st sagaexecution.StepType) *SagaExecutionUpdateOne {
	seuo.mutation.SetStepType(st)
	return seuo
}

// SetNillableStepType sets the "step_type" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableStepType(st *sagaexecution.StepType) *SagaExecutionUpdateOne {
	if st != nil {
		seuo.SetStepType(*st)
	}
	return seuo
}

// SetCompensationData sets the "compensation_data" field.
func (seuo *SagaExecutionUpdateOne) SetCompensationData(u []uint8) *SagaExecutionUpdateOne {
	seuo.mutation.SetCompensationData(u)
	return seuo
}

// AppendCompensationData appends u to the "compensation_data" field.
func (seuo *SagaExecutionUpdateOne) AppendCompensationData(u []uint8) *SagaExecutionUpdateOne {
	seuo.mutation.AppendCompensationData(u)
	return seuo
}

// ClearCompensationData clears the value of the "compensation_data" field.
func (seuo *SagaExecutionUpdateOne) ClearCompensationData() *SagaExecutionUpdateOne {
	seuo.mutation.ClearCompensationData()
	return seuo
}

// SetExecutionID sets the "execution" edge to the Execution entity by ID.
func (seuo *SagaExecutionUpdateOne) SetExecutionID(id int) *SagaExecutionUpdateOne {
	seuo.mutation.SetExecutionID(id)
	return seuo
}

// SetExecution sets the "execution" edge to the Execution entity.
func (seuo *SagaExecutionUpdateOne) SetExecution(e *Execution) *SagaExecutionUpdateOne {
	return seuo.SetExecutionID(e.ID)
}

// SetExecutionDataID sets the "execution_data" edge to the SagaExecutionData entity by ID.
func (seuo *SagaExecutionUpdateOne) SetExecutionDataID(id int) *SagaExecutionUpdateOne {
	seuo.mutation.SetExecutionDataID(id)
	return seuo
}

// SetNillableExecutionDataID sets the "execution_data" edge to the SagaExecutionData entity by ID if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableExecutionDataID(id *int) *SagaExecutionUpdateOne {
	if id != nil {
		seuo = seuo.SetExecutionDataID(*id)
	}
	return seuo
}

// SetExecutionData sets the "execution_data" edge to the SagaExecutionData entity.
func (seuo *SagaExecutionUpdateOne) SetExecutionData(s *SagaExecutionData) *SagaExecutionUpdateOne {
	return seuo.SetExecutionDataID(s.ID)
}

// Mutation returns the SagaExecutionMutation object of the builder.
func (seuo *SagaExecutionUpdateOne) Mutation() *SagaExecutionMutation {
	return seuo.mutation
}

// ClearExecution clears the "execution" edge to the Execution entity.
func (seuo *SagaExecutionUpdateOne) ClearExecution() *SagaExecutionUpdateOne {
	seuo.mutation.ClearExecution()
	return seuo
}

// ClearExecutionData clears the "execution_data" edge to the SagaExecutionData entity.
func (seuo *SagaExecutionUpdateOne) ClearExecutionData() *SagaExecutionUpdateOne {
	seuo.mutation.ClearExecutionData()
	return seuo
}

// Where appends a list predicates to the SagaExecutionUpdate builder.
func (seuo *SagaExecutionUpdateOne) Where(ps ...predicate.SagaExecution) *SagaExecutionUpdateOne {
	seuo.mutation.Where(ps...)
	return seuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (seuo *SagaExecutionUpdateOne) Select(field string, fields ...string) *SagaExecutionUpdateOne {
	seuo.fields = append([]string{field}, fields...)
	return seuo
}

// Save executes the query and returns the updated SagaExecution entity.
func (seuo *SagaExecutionUpdateOne) Save(ctx context.Context) (*SagaExecution, error) {
	return withHooks(ctx, seuo.sqlSave, seuo.mutation, seuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seuo *SagaExecutionUpdateOne) SaveX(ctx context.Context) *SagaExecution {
	node, err := seuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (seuo *SagaExecutionUpdateOne) Exec(ctx context.Context) error {
	_, err := seuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seuo *SagaExecutionUpdateOne) ExecX(ctx context.Context) {
	if err := seuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seuo *SagaExecutionUpdateOne) check() error {
	if v, ok := seuo.mutation.StepType(); ok {
		if err := sagaexecution.StepTypeValidator(v); err != nil {
			return &ValidationError{Name: "step_type", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.step_type": %w`, err)}
		}
	}
	if seuo.mutation.ExecutionCleared() && len(seuo.mutation.ExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecution.execution"`)
	}
	return nil
}

func (seuo *SagaExecutionUpdateOne) sqlSave(ctx context.Context) (_node *SagaExecution, err error) {
	if err := seuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecution.Table, sagaexecution.Columns, sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt))
	id, ok := seuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SagaExecution.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := seuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sagaexecution.FieldID)
		for _, f := range fields {
			if !sagaexecution.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != sagaexecution.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := seuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seuo.mutation.StepType(); ok {
		_spec.SetField(sagaexecution.FieldStepType, field.TypeEnum, value)
	}
	if value, ok := seuo.mutation.CompensationData(); ok {
		_spec.SetField(sagaexecution.FieldCompensationData, field.TypeJSON, value)
	}
	if value, ok := seuo.mutation.AppendedCompensationData(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, sagaexecution.FieldCompensationData, value)
		})
	}
	if seuo.mutation.CompensationDataCleared() {
		_spec.ClearField(sagaexecution.FieldCompensationData, field.TypeJSON)
	}
	if seuo.mutation.ExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagaexecution.ExecutionTable,
			Columns: []string{sagaexecution.ExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seuo.mutation.ExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   sagaexecution.ExecutionTable,
			Columns: []string{sagaexecution.ExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seuo.mutation.ExecutionDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sagaexecution.ExecutionDataTable,
			Columns: []string{sagaexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seuo.mutation.ExecutionDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sagaexecution.ExecutionDataTable,
			Columns: []string{sagaexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecutiondata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &SagaExecution{config: seuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, seuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sagaexecution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	seuo.mutation.done = true
	return _node, nil
}