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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecutiondata"
)

// WorkflowExecutionDataUpdate is the builder for updating WorkflowExecutionData entities.
type WorkflowExecutionDataUpdate struct {
	config
	hooks    []Hook
	mutation *WorkflowExecutionDataMutation
}

// Where appends a list predicates to the WorkflowExecutionDataUpdate builder.
func (wedu *WorkflowExecutionDataUpdate) Where(ps ...predicate.WorkflowExecutionData) *WorkflowExecutionDataUpdate {
	wedu.mutation.Where(ps...)
	return wedu
}

// SetError sets the "error" field.
func (wedu *WorkflowExecutionDataUpdate) SetError(s string) *WorkflowExecutionDataUpdate {
	wedu.mutation.SetError(s)
	return wedu
}

// SetNillableError sets the "error" field if the given value is not nil.
func (wedu *WorkflowExecutionDataUpdate) SetNillableError(s *string) *WorkflowExecutionDataUpdate {
	if s != nil {
		wedu.SetError(*s)
	}
	return wedu
}

// ClearError clears the value of the "error" field.
func (wedu *WorkflowExecutionDataUpdate) ClearError() *WorkflowExecutionDataUpdate {
	wedu.mutation.ClearError()
	return wedu
}

// SetOutput sets the "output" field.
func (wedu *WorkflowExecutionDataUpdate) SetOutput(u [][]uint8) *WorkflowExecutionDataUpdate {
	wedu.mutation.SetOutput(u)
	return wedu
}

// AppendOutput appends u to the "output" field.
func (wedu *WorkflowExecutionDataUpdate) AppendOutput(u [][]uint8) *WorkflowExecutionDataUpdate {
	wedu.mutation.AppendOutput(u)
	return wedu
}

// ClearOutput clears the value of the "output" field.
func (wedu *WorkflowExecutionDataUpdate) ClearOutput() *WorkflowExecutionDataUpdate {
	wedu.mutation.ClearOutput()
	return wedu
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (wedu *WorkflowExecutionDataUpdate) SetWorkflowExecutionID(id int) *WorkflowExecutionDataUpdate {
	wedu.mutation.SetWorkflowExecutionID(id)
	return wedu
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (wedu *WorkflowExecutionDataUpdate) SetWorkflowExecution(w *WorkflowExecution) *WorkflowExecutionDataUpdate {
	return wedu.SetWorkflowExecutionID(w.ID)
}

// Mutation returns the WorkflowExecutionDataMutation object of the builder.
func (wedu *WorkflowExecutionDataUpdate) Mutation() *WorkflowExecutionDataMutation {
	return wedu.mutation
}

// ClearWorkflowExecution clears the "workflow_execution" edge to the WorkflowExecution entity.
func (wedu *WorkflowExecutionDataUpdate) ClearWorkflowExecution() *WorkflowExecutionDataUpdate {
	wedu.mutation.ClearWorkflowExecution()
	return wedu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (wedu *WorkflowExecutionDataUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, wedu.sqlSave, wedu.mutation, wedu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (wedu *WorkflowExecutionDataUpdate) SaveX(ctx context.Context) int {
	affected, err := wedu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (wedu *WorkflowExecutionDataUpdate) Exec(ctx context.Context) error {
	_, err := wedu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wedu *WorkflowExecutionDataUpdate) ExecX(ctx context.Context) {
	if err := wedu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (wedu *WorkflowExecutionDataUpdate) check() error {
	if wedu.mutation.WorkflowExecutionCleared() && len(wedu.mutation.WorkflowExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "WorkflowExecutionData.workflow_execution"`)
	}
	return nil
}

func (wedu *WorkflowExecutionDataUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := wedu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(workflowexecutiondata.Table, workflowexecutiondata.Columns, sqlgraph.NewFieldSpec(workflowexecutiondata.FieldID, field.TypeInt))
	if ps := wedu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := wedu.mutation.Error(); ok {
		_spec.SetField(workflowexecutiondata.FieldError, field.TypeString, value)
	}
	if wedu.mutation.ErrorCleared() {
		_spec.ClearField(workflowexecutiondata.FieldError, field.TypeString)
	}
	if value, ok := wedu.mutation.Output(); ok {
		_spec.SetField(workflowexecutiondata.FieldOutput, field.TypeJSON, value)
	}
	if value, ok := wedu.mutation.AppendedOutput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, workflowexecutiondata.FieldOutput, value)
		})
	}
	if wedu.mutation.OutputCleared() {
		_spec.ClearField(workflowexecutiondata.FieldOutput, field.TypeJSON)
	}
	if wedu.mutation.WorkflowExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   workflowexecutiondata.WorkflowExecutionTable,
			Columns: []string{workflowexecutiondata.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := wedu.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   workflowexecutiondata.WorkflowExecutionTable,
			Columns: []string{workflowexecutiondata.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, wedu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{workflowexecutiondata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	wedu.mutation.done = true
	return n, nil
}

// WorkflowExecutionDataUpdateOne is the builder for updating a single WorkflowExecutionData entity.
type WorkflowExecutionDataUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *WorkflowExecutionDataMutation
}

// SetError sets the "error" field.
func (weduo *WorkflowExecutionDataUpdateOne) SetError(s string) *WorkflowExecutionDataUpdateOne {
	weduo.mutation.SetError(s)
	return weduo
}

// SetNillableError sets the "error" field if the given value is not nil.
func (weduo *WorkflowExecutionDataUpdateOne) SetNillableError(s *string) *WorkflowExecutionDataUpdateOne {
	if s != nil {
		weduo.SetError(*s)
	}
	return weduo
}

// ClearError clears the value of the "error" field.
func (weduo *WorkflowExecutionDataUpdateOne) ClearError() *WorkflowExecutionDataUpdateOne {
	weduo.mutation.ClearError()
	return weduo
}

// SetOutput sets the "output" field.
func (weduo *WorkflowExecutionDataUpdateOne) SetOutput(u [][]uint8) *WorkflowExecutionDataUpdateOne {
	weduo.mutation.SetOutput(u)
	return weduo
}

// AppendOutput appends u to the "output" field.
func (weduo *WorkflowExecutionDataUpdateOne) AppendOutput(u [][]uint8) *WorkflowExecutionDataUpdateOne {
	weduo.mutation.AppendOutput(u)
	return weduo
}

// ClearOutput clears the value of the "output" field.
func (weduo *WorkflowExecutionDataUpdateOne) ClearOutput() *WorkflowExecutionDataUpdateOne {
	weduo.mutation.ClearOutput()
	return weduo
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (weduo *WorkflowExecutionDataUpdateOne) SetWorkflowExecutionID(id int) *WorkflowExecutionDataUpdateOne {
	weduo.mutation.SetWorkflowExecutionID(id)
	return weduo
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (weduo *WorkflowExecutionDataUpdateOne) SetWorkflowExecution(w *WorkflowExecution) *WorkflowExecutionDataUpdateOne {
	return weduo.SetWorkflowExecutionID(w.ID)
}

// Mutation returns the WorkflowExecutionDataMutation object of the builder.
func (weduo *WorkflowExecutionDataUpdateOne) Mutation() *WorkflowExecutionDataMutation {
	return weduo.mutation
}

// ClearWorkflowExecution clears the "workflow_execution" edge to the WorkflowExecution entity.
func (weduo *WorkflowExecutionDataUpdateOne) ClearWorkflowExecution() *WorkflowExecutionDataUpdateOne {
	weduo.mutation.ClearWorkflowExecution()
	return weduo
}

// Where appends a list predicates to the WorkflowExecutionDataUpdate builder.
func (weduo *WorkflowExecutionDataUpdateOne) Where(ps ...predicate.WorkflowExecutionData) *WorkflowExecutionDataUpdateOne {
	weduo.mutation.Where(ps...)
	return weduo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (weduo *WorkflowExecutionDataUpdateOne) Select(field string, fields ...string) *WorkflowExecutionDataUpdateOne {
	weduo.fields = append([]string{field}, fields...)
	return weduo
}

// Save executes the query and returns the updated WorkflowExecutionData entity.
func (weduo *WorkflowExecutionDataUpdateOne) Save(ctx context.Context) (*WorkflowExecutionData, error) {
	return withHooks(ctx, weduo.sqlSave, weduo.mutation, weduo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (weduo *WorkflowExecutionDataUpdateOne) SaveX(ctx context.Context) *WorkflowExecutionData {
	node, err := weduo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (weduo *WorkflowExecutionDataUpdateOne) Exec(ctx context.Context) error {
	_, err := weduo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (weduo *WorkflowExecutionDataUpdateOne) ExecX(ctx context.Context) {
	if err := weduo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (weduo *WorkflowExecutionDataUpdateOne) check() error {
	if weduo.mutation.WorkflowExecutionCleared() && len(weduo.mutation.WorkflowExecutionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "WorkflowExecutionData.workflow_execution"`)
	}
	return nil
}

func (weduo *WorkflowExecutionDataUpdateOne) sqlSave(ctx context.Context) (_node *WorkflowExecutionData, err error) {
	if err := weduo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(workflowexecutiondata.Table, workflowexecutiondata.Columns, sqlgraph.NewFieldSpec(workflowexecutiondata.FieldID, field.TypeInt))
	id, ok := weduo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "WorkflowExecutionData.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := weduo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, workflowexecutiondata.FieldID)
		for _, f := range fields {
			if !workflowexecutiondata.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != workflowexecutiondata.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := weduo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := weduo.mutation.Error(); ok {
		_spec.SetField(workflowexecutiondata.FieldError, field.TypeString, value)
	}
	if weduo.mutation.ErrorCleared() {
		_spec.ClearField(workflowexecutiondata.FieldError, field.TypeString)
	}
	if value, ok := weduo.mutation.Output(); ok {
		_spec.SetField(workflowexecutiondata.FieldOutput, field.TypeJSON, value)
	}
	if value, ok := weduo.mutation.AppendedOutput(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, workflowexecutiondata.FieldOutput, value)
		})
	}
	if weduo.mutation.OutputCleared() {
		_spec.ClearField(workflowexecutiondata.FieldOutput, field.TypeJSON)
	}
	if weduo.mutation.WorkflowExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   workflowexecutiondata.WorkflowExecutionTable,
			Columns: []string{workflowexecutiondata.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := weduo.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   workflowexecutiondata.WorkflowExecutionTable,
			Columns: []string{workflowexecutiondata.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &WorkflowExecutionData{config: weduo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, weduo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{workflowexecutiondata.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	weduo.mutation.done = true
	return _node, nil
}
