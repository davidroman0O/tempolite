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
	"github.com/davidroman0O/tempolite/ent/activityentity"
	"github.com/davidroman0O/tempolite/ent/activityexecution"
	"github.com/davidroman0O/tempolite/ent/activityexecutiondata"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/schema"
)

// ActivityExecutionUpdate is the builder for updating ActivityExecution entities.
type ActivityExecutionUpdate struct {
	config
	hooks    []Hook
	mutation *ActivityExecutionMutation
}

// Where appends a list predicates to the ActivityExecutionUpdate builder.
func (aeu *ActivityExecutionUpdate) Where(ps ...predicate.ActivityExecution) *ActivityExecutionUpdate {
	aeu.mutation.Where(ps...)
	return aeu
}

// SetActivityEntityID sets the "activity_entity_id" field.
func (aeu *ActivityExecutionUpdate) SetActivityEntityID(sei schema.ActivityEntityID) *ActivityExecutionUpdate {
	aeu.mutation.SetActivityEntityID(sei)
	return aeu
}

// SetNillableActivityEntityID sets the "activity_entity_id" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableActivityEntityID(sei *schema.ActivityEntityID) *ActivityExecutionUpdate {
	if sei != nil {
		aeu.SetActivityEntityID(*sei)
	}
	return aeu
}

// SetStartedAt sets the "started_at" field.
func (aeu *ActivityExecutionUpdate) SetStartedAt(t time.Time) *ActivityExecutionUpdate {
	aeu.mutation.SetStartedAt(t)
	return aeu
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableStartedAt(t *time.Time) *ActivityExecutionUpdate {
	if t != nil {
		aeu.SetStartedAt(*t)
	}
	return aeu
}

// SetCompletedAt sets the "completed_at" field.
func (aeu *ActivityExecutionUpdate) SetCompletedAt(t time.Time) *ActivityExecutionUpdate {
	aeu.mutation.SetCompletedAt(t)
	return aeu
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableCompletedAt(t *time.Time) *ActivityExecutionUpdate {
	if t != nil {
		aeu.SetCompletedAt(*t)
	}
	return aeu
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (aeu *ActivityExecutionUpdate) ClearCompletedAt() *ActivityExecutionUpdate {
	aeu.mutation.ClearCompletedAt()
	return aeu
}

// SetStatus sets the "status" field.
func (aeu *ActivityExecutionUpdate) SetStatus(ss schema.ExecutionStatus) *ActivityExecutionUpdate {
	aeu.mutation.SetStatus(ss)
	return aeu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableStatus(ss *schema.ExecutionStatus) *ActivityExecutionUpdate {
	if ss != nil {
		aeu.SetStatus(*ss)
	}
	return aeu
}

// SetError sets the "error" field.
func (aeu *ActivityExecutionUpdate) SetError(s string) *ActivityExecutionUpdate {
	aeu.mutation.SetError(s)
	return aeu
}

// SetNillableError sets the "error" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableError(s *string) *ActivityExecutionUpdate {
	if s != nil {
		aeu.SetError(*s)
	}
	return aeu
}

// ClearError clears the value of the "error" field.
func (aeu *ActivityExecutionUpdate) ClearError() *ActivityExecutionUpdate {
	aeu.mutation.ClearError()
	return aeu
}

// SetStackTrace sets the "stack_trace" field.
func (aeu *ActivityExecutionUpdate) SetStackTrace(s string) *ActivityExecutionUpdate {
	aeu.mutation.SetStackTrace(s)
	return aeu
}

// SetNillableStackTrace sets the "stack_trace" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableStackTrace(s *string) *ActivityExecutionUpdate {
	if s != nil {
		aeu.SetStackTrace(*s)
	}
	return aeu
}

// ClearStackTrace clears the value of the "stack_trace" field.
func (aeu *ActivityExecutionUpdate) ClearStackTrace() *ActivityExecutionUpdate {
	aeu.mutation.ClearStackTrace()
	return aeu
}

// SetCreatedAt sets the "created_at" field.
func (aeu *ActivityExecutionUpdate) SetCreatedAt(t time.Time) *ActivityExecutionUpdate {
	aeu.mutation.SetCreatedAt(t)
	return aeu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableCreatedAt(t *time.Time) *ActivityExecutionUpdate {
	if t != nil {
		aeu.SetCreatedAt(*t)
	}
	return aeu
}

// SetUpdatedAt sets the "updated_at" field.
func (aeu *ActivityExecutionUpdate) SetUpdatedAt(t time.Time) *ActivityExecutionUpdate {
	aeu.mutation.SetUpdatedAt(t)
	return aeu
}

// SetActivityID sets the "activity" edge to the ActivityEntity entity by ID.
func (aeu *ActivityExecutionUpdate) SetActivityID(id schema.ActivityEntityID) *ActivityExecutionUpdate {
	aeu.mutation.SetActivityID(id)
	return aeu
}

// SetActivity sets the "activity" edge to the ActivityEntity entity.
func (aeu *ActivityExecutionUpdate) SetActivity(a *ActivityEntity) *ActivityExecutionUpdate {
	return aeu.SetActivityID(a.ID)
}

// SetExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID.
func (aeu *ActivityExecutionUpdate) SetExecutionDataID(id schema.ActivityExecutionDataID) *ActivityExecutionUpdate {
	aeu.mutation.SetExecutionDataID(id)
	return aeu
}

// SetNillableExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID if the given value is not nil.
func (aeu *ActivityExecutionUpdate) SetNillableExecutionDataID(id *schema.ActivityExecutionDataID) *ActivityExecutionUpdate {
	if id != nil {
		aeu = aeu.SetExecutionDataID(*id)
	}
	return aeu
}

// SetExecutionData sets the "execution_data" edge to the ActivityExecutionData entity.
func (aeu *ActivityExecutionUpdate) SetExecutionData(a *ActivityExecutionData) *ActivityExecutionUpdate {
	return aeu.SetExecutionDataID(a.ID)
}

// Mutation returns the ActivityExecutionMutation object of the builder.
func (aeu *ActivityExecutionUpdate) Mutation() *ActivityExecutionMutation {
	return aeu.mutation
}

// ClearActivity clears the "activity" edge to the ActivityEntity entity.
func (aeu *ActivityExecutionUpdate) ClearActivity() *ActivityExecutionUpdate {
	aeu.mutation.ClearActivity()
	return aeu
}

// ClearExecutionData clears the "execution_data" edge to the ActivityExecutionData entity.
func (aeu *ActivityExecutionUpdate) ClearExecutionData() *ActivityExecutionUpdate {
	aeu.mutation.ClearExecutionData()
	return aeu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (aeu *ActivityExecutionUpdate) Save(ctx context.Context) (int, error) {
	aeu.defaults()
	return withHooks(ctx, aeu.sqlSave, aeu.mutation, aeu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (aeu *ActivityExecutionUpdate) SaveX(ctx context.Context) int {
	affected, err := aeu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (aeu *ActivityExecutionUpdate) Exec(ctx context.Context) error {
	_, err := aeu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aeu *ActivityExecutionUpdate) ExecX(ctx context.Context) {
	if err := aeu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (aeu *ActivityExecutionUpdate) defaults() {
	if _, ok := aeu.mutation.UpdatedAt(); !ok {
		v := activityexecution.UpdateDefaultUpdatedAt()
		aeu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (aeu *ActivityExecutionUpdate) check() error {
	if aeu.mutation.ActivityCleared() && len(aeu.mutation.ActivityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "ActivityExecution.activity"`)
	}
	return nil
}

func (aeu *ActivityExecutionUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := aeu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(activityexecution.Table, activityexecution.Columns, sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt))
	if ps := aeu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := aeu.mutation.StartedAt(); ok {
		_spec.SetField(activityexecution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := aeu.mutation.CompletedAt(); ok {
		_spec.SetField(activityexecution.FieldCompletedAt, field.TypeTime, value)
	}
	if aeu.mutation.CompletedAtCleared() {
		_spec.ClearField(activityexecution.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := aeu.mutation.Status(); ok {
		_spec.SetField(activityexecution.FieldStatus, field.TypeString, value)
	}
	if value, ok := aeu.mutation.Error(); ok {
		_spec.SetField(activityexecution.FieldError, field.TypeString, value)
	}
	if aeu.mutation.ErrorCleared() {
		_spec.ClearField(activityexecution.FieldError, field.TypeString)
	}
	if value, ok := aeu.mutation.StackTrace(); ok {
		_spec.SetField(activityexecution.FieldStackTrace, field.TypeString, value)
	}
	if aeu.mutation.StackTraceCleared() {
		_spec.ClearField(activityexecution.FieldStackTrace, field.TypeString)
	}
	if value, ok := aeu.mutation.CreatedAt(); ok {
		_spec.SetField(activityexecution.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := aeu.mutation.UpdatedAt(); ok {
		_spec.SetField(activityexecution.FieldUpdatedAt, field.TypeTime, value)
	}
	if aeu.mutation.ActivityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   activityexecution.ActivityTable,
			Columns: []string{activityexecution.ActivityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := aeu.mutation.ActivityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   activityexecution.ActivityTable,
			Columns: []string{activityexecution.ActivityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if aeu.mutation.ExecutionDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityexecution.ExecutionDataTable,
			Columns: []string{activityexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecutiondata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := aeu.mutation.ExecutionDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityexecution.ExecutionDataTable,
			Columns: []string{activityexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecutiondata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, aeu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{activityexecution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	aeu.mutation.done = true
	return n, nil
}

// ActivityExecutionUpdateOne is the builder for updating a single ActivityExecution entity.
type ActivityExecutionUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ActivityExecutionMutation
}

// SetActivityEntityID sets the "activity_entity_id" field.
func (aeuo *ActivityExecutionUpdateOne) SetActivityEntityID(sei schema.ActivityEntityID) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetActivityEntityID(sei)
	return aeuo
}

// SetNillableActivityEntityID sets the "activity_entity_id" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableActivityEntityID(sei *schema.ActivityEntityID) *ActivityExecutionUpdateOne {
	if sei != nil {
		aeuo.SetActivityEntityID(*sei)
	}
	return aeuo
}

// SetStartedAt sets the "started_at" field.
func (aeuo *ActivityExecutionUpdateOne) SetStartedAt(t time.Time) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetStartedAt(t)
	return aeuo
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableStartedAt(t *time.Time) *ActivityExecutionUpdateOne {
	if t != nil {
		aeuo.SetStartedAt(*t)
	}
	return aeuo
}

// SetCompletedAt sets the "completed_at" field.
func (aeuo *ActivityExecutionUpdateOne) SetCompletedAt(t time.Time) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetCompletedAt(t)
	return aeuo
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableCompletedAt(t *time.Time) *ActivityExecutionUpdateOne {
	if t != nil {
		aeuo.SetCompletedAt(*t)
	}
	return aeuo
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (aeuo *ActivityExecutionUpdateOne) ClearCompletedAt() *ActivityExecutionUpdateOne {
	aeuo.mutation.ClearCompletedAt()
	return aeuo
}

// SetStatus sets the "status" field.
func (aeuo *ActivityExecutionUpdateOne) SetStatus(ss schema.ExecutionStatus) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetStatus(ss)
	return aeuo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableStatus(ss *schema.ExecutionStatus) *ActivityExecutionUpdateOne {
	if ss != nil {
		aeuo.SetStatus(*ss)
	}
	return aeuo
}

// SetError sets the "error" field.
func (aeuo *ActivityExecutionUpdateOne) SetError(s string) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetError(s)
	return aeuo
}

// SetNillableError sets the "error" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableError(s *string) *ActivityExecutionUpdateOne {
	if s != nil {
		aeuo.SetError(*s)
	}
	return aeuo
}

// ClearError clears the value of the "error" field.
func (aeuo *ActivityExecutionUpdateOne) ClearError() *ActivityExecutionUpdateOne {
	aeuo.mutation.ClearError()
	return aeuo
}

// SetStackTrace sets the "stack_trace" field.
func (aeuo *ActivityExecutionUpdateOne) SetStackTrace(s string) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetStackTrace(s)
	return aeuo
}

// SetNillableStackTrace sets the "stack_trace" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableStackTrace(s *string) *ActivityExecutionUpdateOne {
	if s != nil {
		aeuo.SetStackTrace(*s)
	}
	return aeuo
}

// ClearStackTrace clears the value of the "stack_trace" field.
func (aeuo *ActivityExecutionUpdateOne) ClearStackTrace() *ActivityExecutionUpdateOne {
	aeuo.mutation.ClearStackTrace()
	return aeuo
}

// SetCreatedAt sets the "created_at" field.
func (aeuo *ActivityExecutionUpdateOne) SetCreatedAt(t time.Time) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetCreatedAt(t)
	return aeuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableCreatedAt(t *time.Time) *ActivityExecutionUpdateOne {
	if t != nil {
		aeuo.SetCreatedAt(*t)
	}
	return aeuo
}

// SetUpdatedAt sets the "updated_at" field.
func (aeuo *ActivityExecutionUpdateOne) SetUpdatedAt(t time.Time) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetUpdatedAt(t)
	return aeuo
}

// SetActivityID sets the "activity" edge to the ActivityEntity entity by ID.
func (aeuo *ActivityExecutionUpdateOne) SetActivityID(id schema.ActivityEntityID) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetActivityID(id)
	return aeuo
}

// SetActivity sets the "activity" edge to the ActivityEntity entity.
func (aeuo *ActivityExecutionUpdateOne) SetActivity(a *ActivityEntity) *ActivityExecutionUpdateOne {
	return aeuo.SetActivityID(a.ID)
}

// SetExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID.
func (aeuo *ActivityExecutionUpdateOne) SetExecutionDataID(id schema.ActivityExecutionDataID) *ActivityExecutionUpdateOne {
	aeuo.mutation.SetExecutionDataID(id)
	return aeuo
}

// SetNillableExecutionDataID sets the "execution_data" edge to the ActivityExecutionData entity by ID if the given value is not nil.
func (aeuo *ActivityExecutionUpdateOne) SetNillableExecutionDataID(id *schema.ActivityExecutionDataID) *ActivityExecutionUpdateOne {
	if id != nil {
		aeuo = aeuo.SetExecutionDataID(*id)
	}
	return aeuo
}

// SetExecutionData sets the "execution_data" edge to the ActivityExecutionData entity.
func (aeuo *ActivityExecutionUpdateOne) SetExecutionData(a *ActivityExecutionData) *ActivityExecutionUpdateOne {
	return aeuo.SetExecutionDataID(a.ID)
}

// Mutation returns the ActivityExecutionMutation object of the builder.
func (aeuo *ActivityExecutionUpdateOne) Mutation() *ActivityExecutionMutation {
	return aeuo.mutation
}

// ClearActivity clears the "activity" edge to the ActivityEntity entity.
func (aeuo *ActivityExecutionUpdateOne) ClearActivity() *ActivityExecutionUpdateOne {
	aeuo.mutation.ClearActivity()
	return aeuo
}

// ClearExecutionData clears the "execution_data" edge to the ActivityExecutionData entity.
func (aeuo *ActivityExecutionUpdateOne) ClearExecutionData() *ActivityExecutionUpdateOne {
	aeuo.mutation.ClearExecutionData()
	return aeuo
}

// Where appends a list predicates to the ActivityExecutionUpdate builder.
func (aeuo *ActivityExecutionUpdateOne) Where(ps ...predicate.ActivityExecution) *ActivityExecutionUpdateOne {
	aeuo.mutation.Where(ps...)
	return aeuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (aeuo *ActivityExecutionUpdateOne) Select(field string, fields ...string) *ActivityExecutionUpdateOne {
	aeuo.fields = append([]string{field}, fields...)
	return aeuo
}

// Save executes the query and returns the updated ActivityExecution entity.
func (aeuo *ActivityExecutionUpdateOne) Save(ctx context.Context) (*ActivityExecution, error) {
	aeuo.defaults()
	return withHooks(ctx, aeuo.sqlSave, aeuo.mutation, aeuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (aeuo *ActivityExecutionUpdateOne) SaveX(ctx context.Context) *ActivityExecution {
	node, err := aeuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (aeuo *ActivityExecutionUpdateOne) Exec(ctx context.Context) error {
	_, err := aeuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aeuo *ActivityExecutionUpdateOne) ExecX(ctx context.Context) {
	if err := aeuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (aeuo *ActivityExecutionUpdateOne) defaults() {
	if _, ok := aeuo.mutation.UpdatedAt(); !ok {
		v := activityexecution.UpdateDefaultUpdatedAt()
		aeuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (aeuo *ActivityExecutionUpdateOne) check() error {
	if aeuo.mutation.ActivityCleared() && len(aeuo.mutation.ActivityIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "ActivityExecution.activity"`)
	}
	return nil
}

func (aeuo *ActivityExecutionUpdateOne) sqlSave(ctx context.Context) (_node *ActivityExecution, err error) {
	if err := aeuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(activityexecution.Table, activityexecution.Columns, sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt))
	id, ok := aeuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "ActivityExecution.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := aeuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, activityexecution.FieldID)
		for _, f := range fields {
			if !activityexecution.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != activityexecution.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := aeuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := aeuo.mutation.StartedAt(); ok {
		_spec.SetField(activityexecution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := aeuo.mutation.CompletedAt(); ok {
		_spec.SetField(activityexecution.FieldCompletedAt, field.TypeTime, value)
	}
	if aeuo.mutation.CompletedAtCleared() {
		_spec.ClearField(activityexecution.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := aeuo.mutation.Status(); ok {
		_spec.SetField(activityexecution.FieldStatus, field.TypeString, value)
	}
	if value, ok := aeuo.mutation.Error(); ok {
		_spec.SetField(activityexecution.FieldError, field.TypeString, value)
	}
	if aeuo.mutation.ErrorCleared() {
		_spec.ClearField(activityexecution.FieldError, field.TypeString)
	}
	if value, ok := aeuo.mutation.StackTrace(); ok {
		_spec.SetField(activityexecution.FieldStackTrace, field.TypeString, value)
	}
	if aeuo.mutation.StackTraceCleared() {
		_spec.ClearField(activityexecution.FieldStackTrace, field.TypeString)
	}
	if value, ok := aeuo.mutation.CreatedAt(); ok {
		_spec.SetField(activityexecution.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := aeuo.mutation.UpdatedAt(); ok {
		_spec.SetField(activityexecution.FieldUpdatedAt, field.TypeTime, value)
	}
	if aeuo.mutation.ActivityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   activityexecution.ActivityTable,
			Columns: []string{activityexecution.ActivityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := aeuo.mutation.ActivityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   activityexecution.ActivityTable,
			Columns: []string{activityexecution.ActivityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if aeuo.mutation.ExecutionDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityexecution.ExecutionDataTable,
			Columns: []string{activityexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecutiondata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := aeuo.mutation.ExecutionDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   activityexecution.ExecutionDataTable,
			Columns: []string{activityexecution.ExecutionDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(activityexecutiondata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &ActivityExecution{config: aeuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, aeuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{activityexecution.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	aeuo.mutation.done = true
	return _node, nil
}
