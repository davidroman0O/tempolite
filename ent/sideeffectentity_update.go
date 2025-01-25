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
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/sideeffectdata"
	"github.com/davidroman0O/tempolite/ent/sideeffectentity"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
)

// SideEffectEntityUpdate is the builder for updating SideEffectEntity entities.
type SideEffectEntityUpdate struct {
	config
	hooks    []Hook
	mutation *SideEffectEntityMutation
}

// Where appends a list predicates to the SideEffectEntityUpdate builder.
func (seeu *SideEffectEntityUpdate) Where(ps ...predicate.SideEffectEntity) *SideEffectEntityUpdate {
	seeu.mutation.Where(ps...)
	return seeu
}

// SetHandlerName sets the "handler_name" field.
func (seeu *SideEffectEntityUpdate) SetHandlerName(s string) *SideEffectEntityUpdate {
	seeu.mutation.SetHandlerName(s)
	return seeu
}

// SetNillableHandlerName sets the "handler_name" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableHandlerName(s *string) *SideEffectEntityUpdate {
	if s != nil {
		seeu.SetHandlerName(*s)
	}
	return seeu
}

// SetType sets the "type" field.
func (seeu *SideEffectEntityUpdate) SetType(st schema.EntityType) *SideEffectEntityUpdate {
	seeu.mutation.SetType(st)
	return seeu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableType(st *schema.EntityType) *SideEffectEntityUpdate {
	if st != nil {
		seeu.SetType(*st)
	}
	return seeu
}

// SetStatus sets the "status" field.
func (seeu *SideEffectEntityUpdate) SetStatus(ss schema.EntityStatus) *SideEffectEntityUpdate {
	seeu.mutation.SetStatus(ss)
	return seeu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableStatus(ss *schema.EntityStatus) *SideEffectEntityUpdate {
	if ss != nil {
		seeu.SetStatus(*ss)
	}
	return seeu
}

// SetStepID sets the "step_id" field.
func (seeu *SideEffectEntityUpdate) SetStepID(sesi schema.SideEffectStepID) *SideEffectEntityUpdate {
	seeu.mutation.SetStepID(sesi)
	return seeu
}

// SetNillableStepID sets the "step_id" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableStepID(sesi *schema.SideEffectStepID) *SideEffectEntityUpdate {
	if sesi != nil {
		seeu.SetStepID(*sesi)
	}
	return seeu
}

// SetRunID sets the "run_id" field.
func (seeu *SideEffectEntityUpdate) SetRunID(si schema.RunID) *SideEffectEntityUpdate {
	seeu.mutation.ResetRunID()
	seeu.mutation.SetRunID(si)
	return seeu
}

// SetNillableRunID sets the "run_id" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableRunID(si *schema.RunID) *SideEffectEntityUpdate {
	if si != nil {
		seeu.SetRunID(*si)
	}
	return seeu
}

// AddRunID adds si to the "run_id" field.
func (seeu *SideEffectEntityUpdate) AddRunID(si schema.RunID) *SideEffectEntityUpdate {
	seeu.mutation.AddRunID(si)
	return seeu
}

// SetRetryPolicy sets the "retry_policy" field.
func (seeu *SideEffectEntityUpdate) SetRetryPolicy(sp schema.RetryPolicy) *SideEffectEntityUpdate {
	seeu.mutation.SetRetryPolicy(sp)
	return seeu
}

// SetNillableRetryPolicy sets the "retry_policy" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableRetryPolicy(sp *schema.RetryPolicy) *SideEffectEntityUpdate {
	if sp != nil {
		seeu.SetRetryPolicy(*sp)
	}
	return seeu
}

// SetRetryState sets the "retry_state" field.
func (seeu *SideEffectEntityUpdate) SetRetryState(ss schema.RetryState) *SideEffectEntityUpdate {
	seeu.mutation.SetRetryState(ss)
	return seeu
}

// SetNillableRetryState sets the "retry_state" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableRetryState(ss *schema.RetryState) *SideEffectEntityUpdate {
	if ss != nil {
		seeu.SetRetryState(*ss)
	}
	return seeu
}

// SetCreatedAt sets the "created_at" field.
func (seeu *SideEffectEntityUpdate) SetCreatedAt(t time.Time) *SideEffectEntityUpdate {
	seeu.mutation.SetCreatedAt(t)
	return seeu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableCreatedAt(t *time.Time) *SideEffectEntityUpdate {
	if t != nil {
		seeu.SetCreatedAt(*t)
	}
	return seeu
}

// SetUpdatedAt sets the "updated_at" field.
func (seeu *SideEffectEntityUpdate) SetUpdatedAt(t time.Time) *SideEffectEntityUpdate {
	seeu.mutation.SetUpdatedAt(t)
	return seeu
}

// SetWorkflowID sets the "workflow" edge to the WorkflowEntity entity by ID.
func (seeu *SideEffectEntityUpdate) SetWorkflowID(id schema.WorkflowEntityID) *SideEffectEntityUpdate {
	seeu.mutation.SetWorkflowID(id)
	return seeu
}

// SetWorkflow sets the "workflow" edge to the WorkflowEntity entity.
func (seeu *SideEffectEntityUpdate) SetWorkflow(w *WorkflowEntity) *SideEffectEntityUpdate {
	return seeu.SetWorkflowID(w.ID)
}

// SetSideEffectDataID sets the "side_effect_data" edge to the SideEffectData entity by ID.
func (seeu *SideEffectEntityUpdate) SetSideEffectDataID(id schema.SideEffectDataID) *SideEffectEntityUpdate {
	seeu.mutation.SetSideEffectDataID(id)
	return seeu
}

// SetNillableSideEffectDataID sets the "side_effect_data" edge to the SideEffectData entity by ID if the given value is not nil.
func (seeu *SideEffectEntityUpdate) SetNillableSideEffectDataID(id *schema.SideEffectDataID) *SideEffectEntityUpdate {
	if id != nil {
		seeu = seeu.SetSideEffectDataID(*id)
	}
	return seeu
}

// SetSideEffectData sets the "side_effect_data" edge to the SideEffectData entity.
func (seeu *SideEffectEntityUpdate) SetSideEffectData(s *SideEffectData) *SideEffectEntityUpdate {
	return seeu.SetSideEffectDataID(s.ID)
}

// AddExecutionIDs adds the "executions" edge to the SideEffectExecution entity by IDs.
func (seeu *SideEffectEntityUpdate) AddExecutionIDs(ids ...schema.SideEffectExecutionID) *SideEffectEntityUpdate {
	seeu.mutation.AddExecutionIDs(ids...)
	return seeu
}

// AddExecutions adds the "executions" edges to the SideEffectExecution entity.
func (seeu *SideEffectEntityUpdate) AddExecutions(s ...*SideEffectExecution) *SideEffectEntityUpdate {
	ids := make([]schema.SideEffectExecutionID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return seeu.AddExecutionIDs(ids...)
}

// Mutation returns the SideEffectEntityMutation object of the builder.
func (seeu *SideEffectEntityUpdate) Mutation() *SideEffectEntityMutation {
	return seeu.mutation
}

// ClearWorkflow clears the "workflow" edge to the WorkflowEntity entity.
func (seeu *SideEffectEntityUpdate) ClearWorkflow() *SideEffectEntityUpdate {
	seeu.mutation.ClearWorkflow()
	return seeu
}

// ClearSideEffectData clears the "side_effect_data" edge to the SideEffectData entity.
func (seeu *SideEffectEntityUpdate) ClearSideEffectData() *SideEffectEntityUpdate {
	seeu.mutation.ClearSideEffectData()
	return seeu
}

// ClearExecutions clears all "executions" edges to the SideEffectExecution entity.
func (seeu *SideEffectEntityUpdate) ClearExecutions() *SideEffectEntityUpdate {
	seeu.mutation.ClearExecutions()
	return seeu
}

// RemoveExecutionIDs removes the "executions" edge to SideEffectExecution entities by IDs.
func (seeu *SideEffectEntityUpdate) RemoveExecutionIDs(ids ...schema.SideEffectExecutionID) *SideEffectEntityUpdate {
	seeu.mutation.RemoveExecutionIDs(ids...)
	return seeu
}

// RemoveExecutions removes "executions" edges to SideEffectExecution entities.
func (seeu *SideEffectEntityUpdate) RemoveExecutions(s ...*SideEffectExecution) *SideEffectEntityUpdate {
	ids := make([]schema.SideEffectExecutionID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return seeu.RemoveExecutionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (seeu *SideEffectEntityUpdate) Save(ctx context.Context) (int, error) {
	seeu.defaults()
	return withHooks(ctx, seeu.sqlSave, seeu.mutation, seeu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seeu *SideEffectEntityUpdate) SaveX(ctx context.Context) int {
	affected, err := seeu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (seeu *SideEffectEntityUpdate) Exec(ctx context.Context) error {
	_, err := seeu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seeu *SideEffectEntityUpdate) ExecX(ctx context.Context) {
	if err := seeu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (seeu *SideEffectEntityUpdate) defaults() {
	if _, ok := seeu.mutation.UpdatedAt(); !ok {
		v := sideeffectentity.UpdateDefaultUpdatedAt()
		seeu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seeu *SideEffectEntityUpdate) check() error {
	if seeu.mutation.WorkflowCleared() && len(seeu.mutation.WorkflowIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SideEffectEntity.workflow"`)
	}
	return nil
}

func (seeu *SideEffectEntityUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := seeu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(sideeffectentity.Table, sideeffectentity.Columns, sqlgraph.NewFieldSpec(sideeffectentity.FieldID, field.TypeInt))
	if ps := seeu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seeu.mutation.HandlerName(); ok {
		_spec.SetField(sideeffectentity.FieldHandlerName, field.TypeString, value)
	}
	if value, ok := seeu.mutation.GetType(); ok {
		_spec.SetField(sideeffectentity.FieldType, field.TypeString, value)
	}
	if value, ok := seeu.mutation.Status(); ok {
		_spec.SetField(sideeffectentity.FieldStatus, field.TypeString, value)
	}
	if value, ok := seeu.mutation.StepID(); ok {
		_spec.SetField(sideeffectentity.FieldStepID, field.TypeString, value)
	}
	if value, ok := seeu.mutation.RunID(); ok {
		_spec.SetField(sideeffectentity.FieldRunID, field.TypeInt, value)
	}
	if value, ok := seeu.mutation.AddedRunID(); ok {
		_spec.AddField(sideeffectentity.FieldRunID, field.TypeInt, value)
	}
	if value, ok := seeu.mutation.RetryPolicy(); ok {
		_spec.SetField(sideeffectentity.FieldRetryPolicy, field.TypeJSON, value)
	}
	if value, ok := seeu.mutation.RetryState(); ok {
		_spec.SetField(sideeffectentity.FieldRetryState, field.TypeJSON, value)
	}
	if value, ok := seeu.mutation.CreatedAt(); ok {
		_spec.SetField(sideeffectentity.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := seeu.mutation.UpdatedAt(); ok {
		_spec.SetField(sideeffectentity.FieldUpdatedAt, field.TypeTime, value)
	}
	if seeu.mutation.WorkflowCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sideeffectentity.WorkflowTable,
			Columns: []string{sideeffectentity.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeu.mutation.WorkflowIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sideeffectentity.WorkflowTable,
			Columns: []string{sideeffectentity.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seeu.mutation.SideEffectDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sideeffectentity.SideEffectDataTable,
			Columns: []string{sideeffectentity.SideEffectDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeu.mutation.SideEffectDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sideeffectentity.SideEffectDataTable,
			Columns: []string{sideeffectentity.SideEffectDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seeu.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeu.mutation.RemovedExecutionsIDs(); len(nodes) > 0 && !seeu.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeu.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
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
	if n, err = sqlgraph.UpdateNodes(ctx, seeu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sideeffectentity.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	seeu.mutation.done = true
	return n, nil
}

// SideEffectEntityUpdateOne is the builder for updating a single SideEffectEntity entity.
type SideEffectEntityUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *SideEffectEntityMutation
}

// SetHandlerName sets the "handler_name" field.
func (seeuo *SideEffectEntityUpdateOne) SetHandlerName(s string) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetHandlerName(s)
	return seeuo
}

// SetNillableHandlerName sets the "handler_name" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableHandlerName(s *string) *SideEffectEntityUpdateOne {
	if s != nil {
		seeuo.SetHandlerName(*s)
	}
	return seeuo
}

// SetType sets the "type" field.
func (seeuo *SideEffectEntityUpdateOne) SetType(st schema.EntityType) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetType(st)
	return seeuo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableType(st *schema.EntityType) *SideEffectEntityUpdateOne {
	if st != nil {
		seeuo.SetType(*st)
	}
	return seeuo
}

// SetStatus sets the "status" field.
func (seeuo *SideEffectEntityUpdateOne) SetStatus(ss schema.EntityStatus) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetStatus(ss)
	return seeuo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableStatus(ss *schema.EntityStatus) *SideEffectEntityUpdateOne {
	if ss != nil {
		seeuo.SetStatus(*ss)
	}
	return seeuo
}

// SetStepID sets the "step_id" field.
func (seeuo *SideEffectEntityUpdateOne) SetStepID(sesi schema.SideEffectStepID) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetStepID(sesi)
	return seeuo
}

// SetNillableStepID sets the "step_id" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableStepID(sesi *schema.SideEffectStepID) *SideEffectEntityUpdateOne {
	if sesi != nil {
		seeuo.SetStepID(*sesi)
	}
	return seeuo
}

// SetRunID sets the "run_id" field.
func (seeuo *SideEffectEntityUpdateOne) SetRunID(si schema.RunID) *SideEffectEntityUpdateOne {
	seeuo.mutation.ResetRunID()
	seeuo.mutation.SetRunID(si)
	return seeuo
}

// SetNillableRunID sets the "run_id" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableRunID(si *schema.RunID) *SideEffectEntityUpdateOne {
	if si != nil {
		seeuo.SetRunID(*si)
	}
	return seeuo
}

// AddRunID adds si to the "run_id" field.
func (seeuo *SideEffectEntityUpdateOne) AddRunID(si schema.RunID) *SideEffectEntityUpdateOne {
	seeuo.mutation.AddRunID(si)
	return seeuo
}

// SetRetryPolicy sets the "retry_policy" field.
func (seeuo *SideEffectEntityUpdateOne) SetRetryPolicy(sp schema.RetryPolicy) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetRetryPolicy(sp)
	return seeuo
}

// SetNillableRetryPolicy sets the "retry_policy" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableRetryPolicy(sp *schema.RetryPolicy) *SideEffectEntityUpdateOne {
	if sp != nil {
		seeuo.SetRetryPolicy(*sp)
	}
	return seeuo
}

// SetRetryState sets the "retry_state" field.
func (seeuo *SideEffectEntityUpdateOne) SetRetryState(ss schema.RetryState) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetRetryState(ss)
	return seeuo
}

// SetNillableRetryState sets the "retry_state" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableRetryState(ss *schema.RetryState) *SideEffectEntityUpdateOne {
	if ss != nil {
		seeuo.SetRetryState(*ss)
	}
	return seeuo
}

// SetCreatedAt sets the "created_at" field.
func (seeuo *SideEffectEntityUpdateOne) SetCreatedAt(t time.Time) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetCreatedAt(t)
	return seeuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableCreatedAt(t *time.Time) *SideEffectEntityUpdateOne {
	if t != nil {
		seeuo.SetCreatedAt(*t)
	}
	return seeuo
}

// SetUpdatedAt sets the "updated_at" field.
func (seeuo *SideEffectEntityUpdateOne) SetUpdatedAt(t time.Time) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetUpdatedAt(t)
	return seeuo
}

// SetWorkflowID sets the "workflow" edge to the WorkflowEntity entity by ID.
func (seeuo *SideEffectEntityUpdateOne) SetWorkflowID(id schema.WorkflowEntityID) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetWorkflowID(id)
	return seeuo
}

// SetWorkflow sets the "workflow" edge to the WorkflowEntity entity.
func (seeuo *SideEffectEntityUpdateOne) SetWorkflow(w *WorkflowEntity) *SideEffectEntityUpdateOne {
	return seeuo.SetWorkflowID(w.ID)
}

// SetSideEffectDataID sets the "side_effect_data" edge to the SideEffectData entity by ID.
func (seeuo *SideEffectEntityUpdateOne) SetSideEffectDataID(id schema.SideEffectDataID) *SideEffectEntityUpdateOne {
	seeuo.mutation.SetSideEffectDataID(id)
	return seeuo
}

// SetNillableSideEffectDataID sets the "side_effect_data" edge to the SideEffectData entity by ID if the given value is not nil.
func (seeuo *SideEffectEntityUpdateOne) SetNillableSideEffectDataID(id *schema.SideEffectDataID) *SideEffectEntityUpdateOne {
	if id != nil {
		seeuo = seeuo.SetSideEffectDataID(*id)
	}
	return seeuo
}

// SetSideEffectData sets the "side_effect_data" edge to the SideEffectData entity.
func (seeuo *SideEffectEntityUpdateOne) SetSideEffectData(s *SideEffectData) *SideEffectEntityUpdateOne {
	return seeuo.SetSideEffectDataID(s.ID)
}

// AddExecutionIDs adds the "executions" edge to the SideEffectExecution entity by IDs.
func (seeuo *SideEffectEntityUpdateOne) AddExecutionIDs(ids ...schema.SideEffectExecutionID) *SideEffectEntityUpdateOne {
	seeuo.mutation.AddExecutionIDs(ids...)
	return seeuo
}

// AddExecutions adds the "executions" edges to the SideEffectExecution entity.
func (seeuo *SideEffectEntityUpdateOne) AddExecutions(s ...*SideEffectExecution) *SideEffectEntityUpdateOne {
	ids := make([]schema.SideEffectExecutionID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return seeuo.AddExecutionIDs(ids...)
}

// Mutation returns the SideEffectEntityMutation object of the builder.
func (seeuo *SideEffectEntityUpdateOne) Mutation() *SideEffectEntityMutation {
	return seeuo.mutation
}

// ClearWorkflow clears the "workflow" edge to the WorkflowEntity entity.
func (seeuo *SideEffectEntityUpdateOne) ClearWorkflow() *SideEffectEntityUpdateOne {
	seeuo.mutation.ClearWorkflow()
	return seeuo
}

// ClearSideEffectData clears the "side_effect_data" edge to the SideEffectData entity.
func (seeuo *SideEffectEntityUpdateOne) ClearSideEffectData() *SideEffectEntityUpdateOne {
	seeuo.mutation.ClearSideEffectData()
	return seeuo
}

// ClearExecutions clears all "executions" edges to the SideEffectExecution entity.
func (seeuo *SideEffectEntityUpdateOne) ClearExecutions() *SideEffectEntityUpdateOne {
	seeuo.mutation.ClearExecutions()
	return seeuo
}

// RemoveExecutionIDs removes the "executions" edge to SideEffectExecution entities by IDs.
func (seeuo *SideEffectEntityUpdateOne) RemoveExecutionIDs(ids ...schema.SideEffectExecutionID) *SideEffectEntityUpdateOne {
	seeuo.mutation.RemoveExecutionIDs(ids...)
	return seeuo
}

// RemoveExecutions removes "executions" edges to SideEffectExecution entities.
func (seeuo *SideEffectEntityUpdateOne) RemoveExecutions(s ...*SideEffectExecution) *SideEffectEntityUpdateOne {
	ids := make([]schema.SideEffectExecutionID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return seeuo.RemoveExecutionIDs(ids...)
}

// Where appends a list predicates to the SideEffectEntityUpdate builder.
func (seeuo *SideEffectEntityUpdateOne) Where(ps ...predicate.SideEffectEntity) *SideEffectEntityUpdateOne {
	seeuo.mutation.Where(ps...)
	return seeuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (seeuo *SideEffectEntityUpdateOne) Select(field string, fields ...string) *SideEffectEntityUpdateOne {
	seeuo.fields = append([]string{field}, fields...)
	return seeuo
}

// Save executes the query and returns the updated SideEffectEntity entity.
func (seeuo *SideEffectEntityUpdateOne) Save(ctx context.Context) (*SideEffectEntity, error) {
	seeuo.defaults()
	return withHooks(ctx, seeuo.sqlSave, seeuo.mutation, seeuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (seeuo *SideEffectEntityUpdateOne) SaveX(ctx context.Context) *SideEffectEntity {
	node, err := seeuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (seeuo *SideEffectEntityUpdateOne) Exec(ctx context.Context) error {
	_, err := seeuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (seeuo *SideEffectEntityUpdateOne) ExecX(ctx context.Context) {
	if err := seeuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (seeuo *SideEffectEntityUpdateOne) defaults() {
	if _, ok := seeuo.mutation.UpdatedAt(); !ok {
		v := sideeffectentity.UpdateDefaultUpdatedAt()
		seeuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (seeuo *SideEffectEntityUpdateOne) check() error {
	if seeuo.mutation.WorkflowCleared() && len(seeuo.mutation.WorkflowIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SideEffectEntity.workflow"`)
	}
	return nil
}

func (seeuo *SideEffectEntityUpdateOne) sqlSave(ctx context.Context) (_node *SideEffectEntity, err error) {
	if err := seeuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(sideeffectentity.Table, sideeffectentity.Columns, sqlgraph.NewFieldSpec(sideeffectentity.FieldID, field.TypeInt))
	id, ok := seeuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SideEffectEntity.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := seeuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sideeffectentity.FieldID)
		for _, f := range fields {
			if !sideeffectentity.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != sideeffectentity.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := seeuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := seeuo.mutation.HandlerName(); ok {
		_spec.SetField(sideeffectentity.FieldHandlerName, field.TypeString, value)
	}
	if value, ok := seeuo.mutation.GetType(); ok {
		_spec.SetField(sideeffectentity.FieldType, field.TypeString, value)
	}
	if value, ok := seeuo.mutation.Status(); ok {
		_spec.SetField(sideeffectentity.FieldStatus, field.TypeString, value)
	}
	if value, ok := seeuo.mutation.StepID(); ok {
		_spec.SetField(sideeffectentity.FieldStepID, field.TypeString, value)
	}
	if value, ok := seeuo.mutation.RunID(); ok {
		_spec.SetField(sideeffectentity.FieldRunID, field.TypeInt, value)
	}
	if value, ok := seeuo.mutation.AddedRunID(); ok {
		_spec.AddField(sideeffectentity.FieldRunID, field.TypeInt, value)
	}
	if value, ok := seeuo.mutation.RetryPolicy(); ok {
		_spec.SetField(sideeffectentity.FieldRetryPolicy, field.TypeJSON, value)
	}
	if value, ok := seeuo.mutation.RetryState(); ok {
		_spec.SetField(sideeffectentity.FieldRetryState, field.TypeJSON, value)
	}
	if value, ok := seeuo.mutation.CreatedAt(); ok {
		_spec.SetField(sideeffectentity.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := seeuo.mutation.UpdatedAt(); ok {
		_spec.SetField(sideeffectentity.FieldUpdatedAt, field.TypeTime, value)
	}
	if seeuo.mutation.WorkflowCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sideeffectentity.WorkflowTable,
			Columns: []string{sideeffectentity.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeuo.mutation.WorkflowIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   sideeffectentity.WorkflowTable,
			Columns: []string{sideeffectentity.WorkflowColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(workflowentity.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seeuo.mutation.SideEffectDataCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sideeffectentity.SideEffectDataTable,
			Columns: []string{sideeffectentity.SideEffectDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeuo.mutation.SideEffectDataIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   sideeffectentity.SideEffectDataTable,
			Columns: []string{sideeffectentity.SideEffectDataColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectdata.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if seeuo.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeuo.mutation.RemovedExecutionsIDs(); len(nodes) > 0 && !seeuo.mutation.ExecutionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := seeuo.mutation.ExecutionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   sideeffectentity.ExecutionsTable,
			Columns: []string{sideeffectentity.ExecutionsColumn},
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
	_node = &SideEffectEntity{config: seeuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, seeuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{sideeffectentity.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	seeuo.mutation.done = true
	return _node, nil
}
