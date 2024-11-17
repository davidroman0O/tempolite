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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagaexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
)

// ExecutionUpdate is the builder for updating Execution entities.
type ExecutionUpdate struct {
	config
	hooks    []Hook
	mutation *ExecutionMutation
}

// Where appends a list predicates to the ExecutionUpdate builder.
func (eu *ExecutionUpdate) Where(ps ...predicate.Execution) *ExecutionUpdate {
	eu.mutation.Where(ps...)
	return eu
}

// SetUpdatedAt sets the "updated_at" field.
func (eu *ExecutionUpdate) SetUpdatedAt(t time.Time) *ExecutionUpdate {
	eu.mutation.SetUpdatedAt(t)
	return eu
}

// SetStartedAt sets the "started_at" field.
func (eu *ExecutionUpdate) SetStartedAt(t time.Time) *ExecutionUpdate {
	eu.mutation.SetStartedAt(t)
	return eu
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableStartedAt(t *time.Time) *ExecutionUpdate {
	if t != nil {
		eu.SetStartedAt(*t)
	}
	return eu
}

// SetCompletedAt sets the "completed_at" field.
func (eu *ExecutionUpdate) SetCompletedAt(t time.Time) *ExecutionUpdate {
	eu.mutation.SetCompletedAt(t)
	return eu
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableCompletedAt(t *time.Time) *ExecutionUpdate {
	if t != nil {
		eu.SetCompletedAt(*t)
	}
	return eu
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (eu *ExecutionUpdate) ClearCompletedAt() *ExecutionUpdate {
	eu.mutation.ClearCompletedAt()
	return eu
}

// SetStatus sets the "status" field.
func (eu *ExecutionUpdate) SetStatus(e execution.Status) *ExecutionUpdate {
	eu.mutation.SetStatus(e)
	return eu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableStatus(e *execution.Status) *ExecutionUpdate {
	if e != nil {
		eu.SetStatus(*e)
	}
	return eu
}

// SetError sets the "error" field.
func (eu *ExecutionUpdate) SetError(s string) *ExecutionUpdate {
	eu.mutation.SetError(s)
	return eu
}

// SetNillableError sets the "error" field if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableError(s *string) *ExecutionUpdate {
	if s != nil {
		eu.SetError(*s)
	}
	return eu
}

// ClearError clears the value of the "error" field.
func (eu *ExecutionUpdate) ClearError() *ExecutionUpdate {
	eu.mutation.ClearError()
	return eu
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (eu *ExecutionUpdate) SetEntityID(id int) *ExecutionUpdate {
	eu.mutation.SetEntityID(id)
	return eu
}

// SetEntity sets the "entity" edge to the Entity entity.
func (eu *ExecutionUpdate) SetEntity(e *Entity) *ExecutionUpdate {
	return eu.SetEntityID(e.ID)
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (eu *ExecutionUpdate) SetWorkflowExecutionID(id int) *ExecutionUpdate {
	eu.mutation.SetWorkflowExecutionID(id)
	return eu
}

// SetNillableWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableWorkflowExecutionID(id *int) *ExecutionUpdate {
	if id != nil {
		eu = eu.SetWorkflowExecutionID(*id)
	}
	return eu
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (eu *ExecutionUpdate) SetWorkflowExecution(w *WorkflowExecution) *ExecutionUpdate {
	return eu.SetWorkflowExecutionID(w.ID)
}

// SetActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID.
func (eu *ExecutionUpdate) SetActivityExecutionID(id int) *ExecutionUpdate {
	eu.mutation.SetActivityExecutionID(id)
	return eu
}

// SetNillableActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableActivityExecutionID(id *int) *ExecutionUpdate {
	if id != nil {
		eu = eu.SetActivityExecutionID(*id)
	}
	return eu
}

// SetActivityExecution sets the "activity_execution" edge to the ActivityExecution entity.
func (eu *ExecutionUpdate) SetActivityExecution(a *ActivityExecution) *ExecutionUpdate {
	return eu.SetActivityExecutionID(a.ID)
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (eu *ExecutionUpdate) SetSagaExecutionID(id int) *ExecutionUpdate {
	eu.mutation.SetSagaExecutionID(id)
	return eu
}

// SetNillableSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableSagaExecutionID(id *int) *ExecutionUpdate {
	if id != nil {
		eu = eu.SetSagaExecutionID(*id)
	}
	return eu
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (eu *ExecutionUpdate) SetSagaExecution(s *SagaExecution) *ExecutionUpdate {
	return eu.SetSagaExecutionID(s.ID)
}

// SetSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID.
func (eu *ExecutionUpdate) SetSideEffectExecutionID(id int) *ExecutionUpdate {
	eu.mutation.SetSideEffectExecutionID(id)
	return eu
}

// SetNillableSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID if the given value is not nil.
func (eu *ExecutionUpdate) SetNillableSideEffectExecutionID(id *int) *ExecutionUpdate {
	if id != nil {
		eu = eu.SetSideEffectExecutionID(*id)
	}
	return eu
}

// SetSideEffectExecution sets the "side_effect_execution" edge to the SideEffectExecution entity.
func (eu *ExecutionUpdate) SetSideEffectExecution(s *SideEffectExecution) *ExecutionUpdate {
	return eu.SetSideEffectExecutionID(s.ID)
}

// Mutation returns the ExecutionMutation object of the builder.
func (eu *ExecutionUpdate) Mutation() *ExecutionMutation {
	return eu.mutation
}

// ClearEntity clears the "entity" edge to the Entity entity.
func (eu *ExecutionUpdate) ClearEntity() *ExecutionUpdate {
	eu.mutation.ClearEntity()
	return eu
}

// ClearWorkflowExecution clears the "workflow_execution" edge to the WorkflowExecution entity.
func (eu *ExecutionUpdate) ClearWorkflowExecution() *ExecutionUpdate {
	eu.mutation.ClearWorkflowExecution()
	return eu
}

// ClearActivityExecution clears the "activity_execution" edge to the ActivityExecution entity.
func (eu *ExecutionUpdate) ClearActivityExecution() *ExecutionUpdate {
	eu.mutation.ClearActivityExecution()
	return eu
}

// ClearSagaExecution clears the "saga_execution" edge to the SagaExecution entity.
func (eu *ExecutionUpdate) ClearSagaExecution() *ExecutionUpdate {
	eu.mutation.ClearSagaExecution()
	return eu
}

// ClearSideEffectExecution clears the "side_effect_execution" edge to the SideEffectExecution entity.
func (eu *ExecutionUpdate) ClearSideEffectExecution() *ExecutionUpdate {
	eu.mutation.ClearSideEffectExecution()
	return eu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (eu *ExecutionUpdate) Save(ctx context.Context) (int, error) {
	eu.defaults()
	return withHooks(ctx, eu.sqlSave, eu.mutation, eu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (eu *ExecutionUpdate) SaveX(ctx context.Context) int {
	affected, err := eu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (eu *ExecutionUpdate) Exec(ctx context.Context) error {
	_, err := eu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (eu *ExecutionUpdate) ExecX(ctx context.Context) {
	if err := eu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (eu *ExecutionUpdate) defaults() {
	if _, ok := eu.mutation.UpdatedAt(); !ok {
		v := execution.UpdateDefaultUpdatedAt()
		eu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (eu *ExecutionUpdate) check() error {
	if v, ok := eu.mutation.Status(); ok {
		if err := execution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Execution.status": %w`, err)}
		}
	}
	if eu.mutation.EntityCleared() && len(eu.mutation.EntityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Execution.entity"`)
	}
	return nil
}

func (eu *ExecutionUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := eu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(execution.Table, execution.Columns, sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt))
	if ps := eu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := eu.mutation.UpdatedAt(); ok {
		_spec.SetField(execution.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := eu.mutation.StartedAt(); ok {
		_spec.SetField(execution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := eu.mutation.CompletedAt(); ok {
		_spec.SetField(execution.FieldCompletedAt, field.TypeTime, value)
	}
	if eu.mutation.CompletedAtCleared() {
		_spec.ClearField(execution.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := eu.mutation.Status(); ok {
		_spec.SetField(execution.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := eu.mutation.Error(); ok {
		_spec.SetField(execution.FieldError, field.TypeString, value)
	}
	if eu.mutation.ErrorCleared() {
		_spec.ClearField(execution.FieldError, field.TypeString)
	}
	if eu.mutation.EntityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   execution.EntityTable,
			Columns: []string{execution.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   execution.EntityTable,
			Columns: []string{execution.EntityColumn},
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
	if eu.mutation.WorkflowExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.WorkflowExecutionTable,
			Columns: []string{execution.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.WorkflowExecutionTable,
			Columns: []string{execution.WorkflowExecutionColumn},
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
	if eu.mutation.ActivityExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.ActivityExecutionTable,
			Columns: []string{execution.ActivityExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.ActivityExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.ActivityExecutionTable,
			Columns: []string{execution.ActivityExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if eu.mutation.SagaExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SagaExecutionTable,
			Columns: []string{execution.SagaExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.SagaExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SagaExecutionTable,
			Columns: []string{execution.SagaExecutionColumn},
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
	if eu.mutation.SideEffectExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SideEffectExecutionTable,
			Columns: []string{execution.SideEffectExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.SideEffectExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SideEffectExecutionTable,
			Columns: []string{execution.SideEffectExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, eu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{execution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	eu.mutation.done = true
	return n, nil
}

// ExecutionUpdateOne is the builder for updating a single Execution entity.
type ExecutionUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ExecutionMutation
}

// SetUpdatedAt sets the "updated_at" field.
func (euo *ExecutionUpdateOne) SetUpdatedAt(t time.Time) *ExecutionUpdateOne {
	euo.mutation.SetUpdatedAt(t)
	return euo
}

// SetStartedAt sets the "started_at" field.
func (euo *ExecutionUpdateOne) SetStartedAt(t time.Time) *ExecutionUpdateOne {
	euo.mutation.SetStartedAt(t)
	return euo
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableStartedAt(t *time.Time) *ExecutionUpdateOne {
	if t != nil {
		euo.SetStartedAt(*t)
	}
	return euo
}

// SetCompletedAt sets the "completed_at" field.
func (euo *ExecutionUpdateOne) SetCompletedAt(t time.Time) *ExecutionUpdateOne {
	euo.mutation.SetCompletedAt(t)
	return euo
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableCompletedAt(t *time.Time) *ExecutionUpdateOne {
	if t != nil {
		euo.SetCompletedAt(*t)
	}
	return euo
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (euo *ExecutionUpdateOne) ClearCompletedAt() *ExecutionUpdateOne {
	euo.mutation.ClearCompletedAt()
	return euo
}

// SetStatus sets the "status" field.
func (euo *ExecutionUpdateOne) SetStatus(e execution.Status) *ExecutionUpdateOne {
	euo.mutation.SetStatus(e)
	return euo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableStatus(e *execution.Status) *ExecutionUpdateOne {
	if e != nil {
		euo.SetStatus(*e)
	}
	return euo
}

// SetError sets the "error" field.
func (euo *ExecutionUpdateOne) SetError(s string) *ExecutionUpdateOne {
	euo.mutation.SetError(s)
	return euo
}

// SetNillableError sets the "error" field if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableError(s *string) *ExecutionUpdateOne {
	if s != nil {
		euo.SetError(*s)
	}
	return euo
}

// ClearError clears the value of the "error" field.
func (euo *ExecutionUpdateOne) ClearError() *ExecutionUpdateOne {
	euo.mutation.ClearError()
	return euo
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (euo *ExecutionUpdateOne) SetEntityID(id int) *ExecutionUpdateOne {
	euo.mutation.SetEntityID(id)
	return euo
}

// SetEntity sets the "entity" edge to the Entity entity.
func (euo *ExecutionUpdateOne) SetEntity(e *Entity) *ExecutionUpdateOne {
	return euo.SetEntityID(e.ID)
}

// SetWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID.
func (euo *ExecutionUpdateOne) SetWorkflowExecutionID(id int) *ExecutionUpdateOne {
	euo.mutation.SetWorkflowExecutionID(id)
	return euo
}

// SetNillableWorkflowExecutionID sets the "workflow_execution" edge to the WorkflowExecution entity by ID if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableWorkflowExecutionID(id *int) *ExecutionUpdateOne {
	if id != nil {
		euo = euo.SetWorkflowExecutionID(*id)
	}
	return euo
}

// SetWorkflowExecution sets the "workflow_execution" edge to the WorkflowExecution entity.
func (euo *ExecutionUpdateOne) SetWorkflowExecution(w *WorkflowExecution) *ExecutionUpdateOne {
	return euo.SetWorkflowExecutionID(w.ID)
}

// SetActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID.
func (euo *ExecutionUpdateOne) SetActivityExecutionID(id int) *ExecutionUpdateOne {
	euo.mutation.SetActivityExecutionID(id)
	return euo
}

// SetNillableActivityExecutionID sets the "activity_execution" edge to the ActivityExecution entity by ID if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableActivityExecutionID(id *int) *ExecutionUpdateOne {
	if id != nil {
		euo = euo.SetActivityExecutionID(*id)
	}
	return euo
}

// SetActivityExecution sets the "activity_execution" edge to the ActivityExecution entity.
func (euo *ExecutionUpdateOne) SetActivityExecution(a *ActivityExecution) *ExecutionUpdateOne {
	return euo.SetActivityExecutionID(a.ID)
}

// SetSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID.
func (euo *ExecutionUpdateOne) SetSagaExecutionID(id int) *ExecutionUpdateOne {
	euo.mutation.SetSagaExecutionID(id)
	return euo
}

// SetNillableSagaExecutionID sets the "saga_execution" edge to the SagaExecution entity by ID if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableSagaExecutionID(id *int) *ExecutionUpdateOne {
	if id != nil {
		euo = euo.SetSagaExecutionID(*id)
	}
	return euo
}

// SetSagaExecution sets the "saga_execution" edge to the SagaExecution entity.
func (euo *ExecutionUpdateOne) SetSagaExecution(s *SagaExecution) *ExecutionUpdateOne {
	return euo.SetSagaExecutionID(s.ID)
}

// SetSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID.
func (euo *ExecutionUpdateOne) SetSideEffectExecutionID(id int) *ExecutionUpdateOne {
	euo.mutation.SetSideEffectExecutionID(id)
	return euo
}

// SetNillableSideEffectExecutionID sets the "side_effect_execution" edge to the SideEffectExecution entity by ID if the given value is not nil.
func (euo *ExecutionUpdateOne) SetNillableSideEffectExecutionID(id *int) *ExecutionUpdateOne {
	if id != nil {
		euo = euo.SetSideEffectExecutionID(*id)
	}
	return euo
}

// SetSideEffectExecution sets the "side_effect_execution" edge to the SideEffectExecution entity.
func (euo *ExecutionUpdateOne) SetSideEffectExecution(s *SideEffectExecution) *ExecutionUpdateOne {
	return euo.SetSideEffectExecutionID(s.ID)
}

// Mutation returns the ExecutionMutation object of the builder.
func (euo *ExecutionUpdateOne) Mutation() *ExecutionMutation {
	return euo.mutation
}

// ClearEntity clears the "entity" edge to the Entity entity.
func (euo *ExecutionUpdateOne) ClearEntity() *ExecutionUpdateOne {
	euo.mutation.ClearEntity()
	return euo
}

// ClearWorkflowExecution clears the "workflow_execution" edge to the WorkflowExecution entity.
func (euo *ExecutionUpdateOne) ClearWorkflowExecution() *ExecutionUpdateOne {
	euo.mutation.ClearWorkflowExecution()
	return euo
}

// ClearActivityExecution clears the "activity_execution" edge to the ActivityExecution entity.
func (euo *ExecutionUpdateOne) ClearActivityExecution() *ExecutionUpdateOne {
	euo.mutation.ClearActivityExecution()
	return euo
}

// ClearSagaExecution clears the "saga_execution" edge to the SagaExecution entity.
func (euo *ExecutionUpdateOne) ClearSagaExecution() *ExecutionUpdateOne {
	euo.mutation.ClearSagaExecution()
	return euo
}

// ClearSideEffectExecution clears the "side_effect_execution" edge to the SideEffectExecution entity.
func (euo *ExecutionUpdateOne) ClearSideEffectExecution() *ExecutionUpdateOne {
	euo.mutation.ClearSideEffectExecution()
	return euo
}

// Where appends a list predicates to the ExecutionUpdate builder.
func (euo *ExecutionUpdateOne) Where(ps ...predicate.Execution) *ExecutionUpdateOne {
	euo.mutation.Where(ps...)
	return euo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (euo *ExecutionUpdateOne) Select(field string, fields ...string) *ExecutionUpdateOne {
	euo.fields = append([]string{field}, fields...)
	return euo
}

// Save executes the query and returns the updated Execution entity.
func (euo *ExecutionUpdateOne) Save(ctx context.Context) (*Execution, error) {
	euo.defaults()
	return withHooks(ctx, euo.sqlSave, euo.mutation, euo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (euo *ExecutionUpdateOne) SaveX(ctx context.Context) *Execution {
	node, err := euo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (euo *ExecutionUpdateOne) Exec(ctx context.Context) error {
	_, err := euo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (euo *ExecutionUpdateOne) ExecX(ctx context.Context) {
	if err := euo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (euo *ExecutionUpdateOne) defaults() {
	if _, ok := euo.mutation.UpdatedAt(); !ok {
		v := execution.UpdateDefaultUpdatedAt()
		euo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (euo *ExecutionUpdateOne) check() error {
	if v, ok := euo.mutation.Status(); ok {
		if err := execution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "Execution.status": %w`, err)}
		}
	}
	if euo.mutation.EntityCleared() && len(euo.mutation.EntityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Execution.entity"`)
	}
	return nil
}

func (euo *ExecutionUpdateOne) sqlSave(ctx context.Context) (_node *Execution, err error) {
	if err := euo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(execution.Table, execution.Columns, sqlgraph.NewFieldSpec(execution.FieldID, field.TypeInt))
	id, ok := euo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Execution.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := euo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, execution.FieldID)
		for _, f := range fields {
			if !execution.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != execution.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := euo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := euo.mutation.UpdatedAt(); ok {
		_spec.SetField(execution.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := euo.mutation.StartedAt(); ok {
		_spec.SetField(execution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := euo.mutation.CompletedAt(); ok {
		_spec.SetField(execution.FieldCompletedAt, field.TypeTime, value)
	}
	if euo.mutation.CompletedAtCleared() {
		_spec.ClearField(execution.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := euo.mutation.Status(); ok {
		_spec.SetField(execution.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := euo.mutation.Error(); ok {
		_spec.SetField(execution.FieldError, field.TypeString, value)
	}
	if euo.mutation.ErrorCleared() {
		_spec.ClearField(execution.FieldError, field.TypeString)
	}
	if euo.mutation.EntityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   execution.EntityTable,
			Columns: []string{execution.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   execution.EntityTable,
			Columns: []string{execution.EntityColumn},
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
	if euo.mutation.WorkflowExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.WorkflowExecutionTable,
			Columns: []string{execution.WorkflowExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.WorkflowExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.WorkflowExecutionTable,
			Columns: []string{execution.WorkflowExecutionColumn},
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
	if euo.mutation.ActivityExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.ActivityExecutionTable,
			Columns: []string{execution.ActivityExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.ActivityExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.ActivityExecutionTable,
			Columns: []string{execution.ActivityExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if euo.mutation.SagaExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SagaExecutionTable,
			Columns: []string{execution.SagaExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.SagaExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SagaExecutionTable,
			Columns: []string{execution.SagaExecutionColumn},
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
	if euo.mutation.SideEffectExecutionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SideEffectExecutionTable,
			Columns: []string{execution.SideEffectExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.SideEffectExecutionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   execution.SideEffectExecutionTable,
			Columns: []string{execution.SideEffectExecutionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Execution{config: euo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, euo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{execution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	euo.mutation.done = true
	return _node, nil
}
