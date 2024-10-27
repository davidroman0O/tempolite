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
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/sagaexecution"
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

// SetHandlerName sets the "handler_name" field.
func (seu *SagaExecutionUpdate) SetHandlerName(s string) *SagaExecutionUpdate {
	seu.mutation.SetHandlerName(s)
	return seu
}

// SetNillableHandlerName sets the "handler_name" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableHandlerName(s *string) *SagaExecutionUpdate {
	if s != nil {
		seu.SetHandlerName(*s)
	}
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

// SetStatus sets the "status" field.
func (seu *SagaExecutionUpdate) SetStatus(s sagaexecution.Status) *SagaExecutionUpdate {
	seu.mutation.SetStatus(s)
	return seu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableStatus(s *sagaexecution.Status) *SagaExecutionUpdate {
	if s != nil {
		seu.SetStatus(*s)
	}
	return seu
}

// SetQueueName sets the "queue_name" field.
func (seu *SagaExecutionUpdate) SetQueueName(s string) *SagaExecutionUpdate {
	seu.mutation.SetQueueName(s)
	return seu
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableQueueName(s *string) *SagaExecutionUpdate {
	if s != nil {
		seu.SetQueueName(*s)
	}
	return seu
}

// SetSequence sets the "sequence" field.
func (seu *SagaExecutionUpdate) SetSequence(i int) *SagaExecutionUpdate {
	seu.mutation.ResetSequence()
	seu.mutation.SetSequence(i)
	return seu
}

// SetNillableSequence sets the "sequence" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableSequence(i *int) *SagaExecutionUpdate {
	if i != nil {
		seu.SetSequence(*i)
	}
	return seu
}

// AddSequence adds i to the "sequence" field.
func (seu *SagaExecutionUpdate) AddSequence(i int) *SagaExecutionUpdate {
	seu.mutation.AddSequence(i)
	return seu
}

// SetError sets the "error" field.
func (seu *SagaExecutionUpdate) SetError(s string) *SagaExecutionUpdate {
	seu.mutation.SetError(s)
	return seu
}

// SetNillableError sets the "error" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableError(s *string) *SagaExecutionUpdate {
	if s != nil {
		seu.SetError(*s)
	}
	return seu
}

// ClearError clears the value of the "error" field.
func (seu *SagaExecutionUpdate) ClearError() *SagaExecutionUpdate {
	seu.mutation.ClearError()
	return seu
}

// SetStartedAt sets the "started_at" field.
func (seu *SagaExecutionUpdate) SetStartedAt(t time.Time) *SagaExecutionUpdate {
	seu.mutation.SetStartedAt(t)
	return seu
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableStartedAt(t *time.Time) *SagaExecutionUpdate {
	if t != nil {
		seu.SetStartedAt(*t)
	}
	return seu
}

// SetCompletedAt sets the "completed_at" field.
func (seu *SagaExecutionUpdate) SetCompletedAt(t time.Time) *SagaExecutionUpdate {
	seu.mutation.SetCompletedAt(t)
	return seu
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (seu *SagaExecutionUpdate) SetNillableCompletedAt(t *time.Time) *SagaExecutionUpdate {
	if t != nil {
		seu.SetCompletedAt(*t)
	}
	return seu
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (seu *SagaExecutionUpdate) ClearCompletedAt() *SagaExecutionUpdate {
	seu.mutation.ClearCompletedAt()
	return seu
}

// SetSagaID sets the "saga" edge to the Saga entity by ID.
func (seu *SagaExecutionUpdate) SetSagaID(id string) *SagaExecutionUpdate {
	seu.mutation.SetSagaID(id)
	return seu
}

// SetSaga sets the "saga" edge to the Saga entity.
func (seu *SagaExecutionUpdate) SetSaga(s *Saga) *SagaExecutionUpdate {
	return seu.SetSagaID(s.ID)
}

// Mutation returns the SagaExecutionMutation object of the builder.
func (seu *SagaExecutionUpdate) Mutation() *SagaExecutionMutation {
	return seu.mutation
}

// ClearSaga clears the "saga" edge to the Saga entity.
func (seu *SagaExecutionUpdate) ClearSaga() *SagaExecutionUpdate {
	seu.mutation.ClearSaga()
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
	if v, ok := seu.mutation.HandlerName(); ok {
		if err := sagaexecution.HandlerNameValidator(v); err != nil {
			return &ValidationError{Name: "handler_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.handler_name": %w`, err)}
		}
	}
	if v, ok := seu.mutation.StepType(); ok {
		if err := sagaexecution.StepTypeValidator(v); err != nil {
			return &ValidationError{Name: "step_type", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.step_type": %w`, err)}
		}
	}
	if v, ok := seu.mutation.Status(); ok {
		if err := sagaexecution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.status": %w`, err)}
		}
	}
	if v, ok := seu.mutation.QueueName(); ok {
		if err := sagaexecution.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.queue_name": %w`, err)}
		}
	}
	if v, ok := seu.mutation.Sequence(); ok {
		if err := sagaexecution.SequenceValidator(v); err != nil {
			return &ValidationError{Name: "sequence", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.sequence": %w`, err)}
		}
	}
	if seu.mutation.SagaCleared() && len(seu.mutation.SagaIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecution.saga"`)
	}
	return nil
}

func (seu *SagaExecutionUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := seu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecution.Table, sagaexecution.Columns, sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeString))
	if ps := seu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seu.mutation.HandlerName(); ok {
		_spec.SetField(sagaexecution.FieldHandlerName, field.TypeString, value)
	}
	if value, ok := seu.mutation.StepType(); ok {
		_spec.SetField(sagaexecution.FieldStepType, field.TypeEnum, value)
	}
	if value, ok := seu.mutation.Status(); ok {
		_spec.SetField(sagaexecution.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := seu.mutation.QueueName(); ok {
		_spec.SetField(sagaexecution.FieldQueueName, field.TypeString, value)
	}
	if value, ok := seu.mutation.Sequence(); ok {
		_spec.SetField(sagaexecution.FieldSequence, field.TypeInt, value)
	}
	if value, ok := seu.mutation.AddedSequence(); ok {
		_spec.AddField(sagaexecution.FieldSequence, field.TypeInt, value)
	}
	if value, ok := seu.mutation.Error(); ok {
		_spec.SetField(sagaexecution.FieldError, field.TypeString, value)
	}
	if seu.mutation.ErrorCleared() {
		_spec.ClearField(sagaexecution.FieldError, field.TypeString)
	}
	if value, ok := seu.mutation.StartedAt(); ok {
		_spec.SetField(sagaexecution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := seu.mutation.CompletedAt(); ok {
		_spec.SetField(sagaexecution.FieldCompletedAt, field.TypeTime, value)
	}
	if seu.mutation.CompletedAtCleared() {
		_spec.ClearField(sagaexecution.FieldCompletedAt, field.TypeTime)
	}
	if seu.mutation.SagaCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sagaexecution.SagaTable,
			Columns: []string{sagaexecution.SagaColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seu.mutation.SagaIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sagaexecution.SagaTable,
			Columns: []string{sagaexecution.SagaColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString),
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

// SetHandlerName sets the "handler_name" field.
func (seuo *SagaExecutionUpdateOne) SetHandlerName(s string) *SagaExecutionUpdateOne {
	seuo.mutation.SetHandlerName(s)
	return seuo
}

// SetNillableHandlerName sets the "handler_name" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableHandlerName(s *string) *SagaExecutionUpdateOne {
	if s != nil {
		seuo.SetHandlerName(*s)
	}
	return seuo
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

// SetStatus sets the "status" field.
func (seuo *SagaExecutionUpdateOne) SetStatus(s sagaexecution.Status) *SagaExecutionUpdateOne {
	seuo.mutation.SetStatus(s)
	return seuo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableStatus(s *sagaexecution.Status) *SagaExecutionUpdateOne {
	if s != nil {
		seuo.SetStatus(*s)
	}
	return seuo
}

// SetQueueName sets the "queue_name" field.
func (seuo *SagaExecutionUpdateOne) SetQueueName(s string) *SagaExecutionUpdateOne {
	seuo.mutation.SetQueueName(s)
	return seuo
}

// SetNillableQueueName sets the "queue_name" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableQueueName(s *string) *SagaExecutionUpdateOne {
	if s != nil {
		seuo.SetQueueName(*s)
	}
	return seuo
}

// SetSequence sets the "sequence" field.
func (seuo *SagaExecutionUpdateOne) SetSequence(i int) *SagaExecutionUpdateOne {
	seuo.mutation.ResetSequence()
	seuo.mutation.SetSequence(i)
	return seuo
}

// SetNillableSequence sets the "sequence" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableSequence(i *int) *SagaExecutionUpdateOne {
	if i != nil {
		seuo.SetSequence(*i)
	}
	return seuo
}

// AddSequence adds i to the "sequence" field.
func (seuo *SagaExecutionUpdateOne) AddSequence(i int) *SagaExecutionUpdateOne {
	seuo.mutation.AddSequence(i)
	return seuo
}

// SetError sets the "error" field.
func (seuo *SagaExecutionUpdateOne) SetError(s string) *SagaExecutionUpdateOne {
	seuo.mutation.SetError(s)
	return seuo
}

// SetNillableError sets the "error" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableError(s *string) *SagaExecutionUpdateOne {
	if s != nil {
		seuo.SetError(*s)
	}
	return seuo
}

// ClearError clears the value of the "error" field.
func (seuo *SagaExecutionUpdateOne) ClearError() *SagaExecutionUpdateOne {
	seuo.mutation.ClearError()
	return seuo
}

// SetStartedAt sets the "started_at" field.
func (seuo *SagaExecutionUpdateOne) SetStartedAt(t time.Time) *SagaExecutionUpdateOne {
	seuo.mutation.SetStartedAt(t)
	return seuo
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableStartedAt(t *time.Time) *SagaExecutionUpdateOne {
	if t != nil {
		seuo.SetStartedAt(*t)
	}
	return seuo
}

// SetCompletedAt sets the "completed_at" field.
func (seuo *SagaExecutionUpdateOne) SetCompletedAt(t time.Time) *SagaExecutionUpdateOne {
	seuo.mutation.SetCompletedAt(t)
	return seuo
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (seuo *SagaExecutionUpdateOne) SetNillableCompletedAt(t *time.Time) *SagaExecutionUpdateOne {
	if t != nil {
		seuo.SetCompletedAt(*t)
	}
	return seuo
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (seuo *SagaExecutionUpdateOne) ClearCompletedAt() *SagaExecutionUpdateOne {
	seuo.mutation.ClearCompletedAt()
	return seuo
}

// SetSagaID sets the "saga" edge to the Saga entity by ID.
func (seuo *SagaExecutionUpdateOne) SetSagaID(id string) *SagaExecutionUpdateOne {
	seuo.mutation.SetSagaID(id)
	return seuo
}

// SetSaga sets the "saga" edge to the Saga entity.
func (seuo *SagaExecutionUpdateOne) SetSaga(s *Saga) *SagaExecutionUpdateOne {
	return seuo.SetSagaID(s.ID)
}

// Mutation returns the SagaExecutionMutation object of the builder.
func (seuo *SagaExecutionUpdateOne) Mutation() *SagaExecutionMutation {
	return seuo.mutation
}

// ClearSaga clears the "saga" edge to the Saga entity.
func (seuo *SagaExecutionUpdateOne) ClearSaga() *SagaExecutionUpdateOne {
	seuo.mutation.ClearSaga()
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
	if v, ok := seuo.mutation.HandlerName(); ok {
		if err := sagaexecution.HandlerNameValidator(v); err != nil {
			return &ValidationError{Name: "handler_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.handler_name": %w`, err)}
		}
	}
	if v, ok := seuo.mutation.StepType(); ok {
		if err := sagaexecution.StepTypeValidator(v); err != nil {
			return &ValidationError{Name: "step_type", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.step_type": %w`, err)}
		}
	}
	if v, ok := seuo.mutation.Status(); ok {
		if err := sagaexecution.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.status": %w`, err)}
		}
	}
	if v, ok := seuo.mutation.QueueName(); ok {
		if err := sagaexecution.QueueNameValidator(v); err != nil {
			return &ValidationError{Name: "queue_name", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.queue_name": %w`, err)}
		}
	}
	if v, ok := seuo.mutation.Sequence(); ok {
		if err := sagaexecution.SequenceValidator(v); err != nil {
			return &ValidationError{Name: "sequence", err: fmt.Errorf(`ent: validator failed for field "SagaExecution.sequence": %w`, err)}
		}
	}
	if seuo.mutation.SagaCleared() && len(seuo.mutation.SagaIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SagaExecution.saga"`)
	}
	return nil
}

func (seuo *SagaExecutionUpdateOne) sqlSave(ctx context.Context) (_node *SagaExecution, err error) {
	if err := seuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(sagaexecution.Table, sagaexecution.Columns, sqlgraph.NewFieldSpec(sagaexecution.FieldID, field.TypeString))
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
	if value, ok := seuo.mutation.HandlerName(); ok {
		_spec.SetField(sagaexecution.FieldHandlerName, field.TypeString, value)
	}
	if value, ok := seuo.mutation.StepType(); ok {
		_spec.SetField(sagaexecution.FieldStepType, field.TypeEnum, value)
	}
	if value, ok := seuo.mutation.Status(); ok {
		_spec.SetField(sagaexecution.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := seuo.mutation.QueueName(); ok {
		_spec.SetField(sagaexecution.FieldQueueName, field.TypeString, value)
	}
	if value, ok := seuo.mutation.Sequence(); ok {
		_spec.SetField(sagaexecution.FieldSequence, field.TypeInt, value)
	}
	if value, ok := seuo.mutation.AddedSequence(); ok {
		_spec.AddField(sagaexecution.FieldSequence, field.TypeInt, value)
	}
	if value, ok := seuo.mutation.Error(); ok {
		_spec.SetField(sagaexecution.FieldError, field.TypeString, value)
	}
	if seuo.mutation.ErrorCleared() {
		_spec.ClearField(sagaexecution.FieldError, field.TypeString)
	}
	if value, ok := seuo.mutation.StartedAt(); ok {
		_spec.SetField(sagaexecution.FieldStartedAt, field.TypeTime, value)
	}
	if value, ok := seuo.mutation.CompletedAt(); ok {
		_spec.SetField(sagaexecution.FieldCompletedAt, field.TypeTime, value)
	}
	if seuo.mutation.CompletedAtCleared() {
		_spec.ClearField(sagaexecution.FieldCompletedAt, field.TypeTime)
	}
	if seuo.mutation.SagaCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sagaexecution.SagaTable,
			Columns: []string{sagaexecution.SagaColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seuo.mutation.SagaIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sagaexecution.SagaTable,
			Columns: []string{sagaexecution.SagaColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString),
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
