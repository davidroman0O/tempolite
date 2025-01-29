// Code generated by ent, DO NOT EDIT.

package workflowexecution

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/schema"
)

// ID filters vertices based on their ID field.
func ID(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id schema.WorkflowExecutionID) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldID, id))
}

// WorkflowEntityID applies equality check predicate on the "workflow_entity_id" field. It's identical to WorkflowEntityIDEQ.
func WorkflowEntityID(v schema.WorkflowEntityID) predicate.WorkflowExecution {
	vc := int(v)
	return predicate.WorkflowExecution(sql.FieldEQ(FieldWorkflowEntityID, vc))
}

// StartedAt applies equality check predicate on the "started_at" field. It's identical to StartedAtEQ.
func StartedAt(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStartedAt, v))
}

// CompletedAt applies equality check predicate on the "completed_at" field. It's identical to CompletedAtEQ.
func CompletedAt(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldCompletedAt, v))
}

// Status applies equality check predicate on the "status" field. It's identical to StatusEQ.
func Status(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStatus, vc))
}

// Error applies equality check predicate on the "error" field. It's identical to ErrorEQ.
func Error(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldError, v))
}

// StackTrace applies equality check predicate on the "stack_trace" field. It's identical to StackTraceEQ.
func StackTrace(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStackTrace, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldUpdatedAt, v))
}

// WorkflowEntityIDEQ applies the EQ predicate on the "workflow_entity_id" field.
func WorkflowEntityIDEQ(v schema.WorkflowEntityID) predicate.WorkflowExecution {
	vc := int(v)
	return predicate.WorkflowExecution(sql.FieldEQ(FieldWorkflowEntityID, vc))
}

// WorkflowEntityIDNEQ applies the NEQ predicate on the "workflow_entity_id" field.
func WorkflowEntityIDNEQ(v schema.WorkflowEntityID) predicate.WorkflowExecution {
	vc := int(v)
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldWorkflowEntityID, vc))
}

// WorkflowEntityIDIn applies the In predicate on the "workflow_entity_id" field.
func WorkflowEntityIDIn(vs ...schema.WorkflowEntityID) predicate.WorkflowExecution {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.WorkflowExecution(sql.FieldIn(FieldWorkflowEntityID, v...))
}

// WorkflowEntityIDNotIn applies the NotIn predicate on the "workflow_entity_id" field.
func WorkflowEntityIDNotIn(vs ...schema.WorkflowEntityID) predicate.WorkflowExecution {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldWorkflowEntityID, v...))
}

// StartedAtEQ applies the EQ predicate on the "started_at" field.
func StartedAtEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStartedAt, v))
}

// StartedAtNEQ applies the NEQ predicate on the "started_at" field.
func StartedAtNEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldStartedAt, v))
}

// StartedAtIn applies the In predicate on the "started_at" field.
func StartedAtIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldStartedAt, vs...))
}

// StartedAtNotIn applies the NotIn predicate on the "started_at" field.
func StartedAtNotIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldStartedAt, vs...))
}

// StartedAtGT applies the GT predicate on the "started_at" field.
func StartedAtGT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldStartedAt, v))
}

// StartedAtGTE applies the GTE predicate on the "started_at" field.
func StartedAtGTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldStartedAt, v))
}

// StartedAtLT applies the LT predicate on the "started_at" field.
func StartedAtLT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldStartedAt, v))
}

// StartedAtLTE applies the LTE predicate on the "started_at" field.
func StartedAtLTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldStartedAt, v))
}

// CompletedAtEQ applies the EQ predicate on the "completed_at" field.
func CompletedAtEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldCompletedAt, v))
}

// CompletedAtNEQ applies the NEQ predicate on the "completed_at" field.
func CompletedAtNEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldCompletedAt, v))
}

// CompletedAtIn applies the In predicate on the "completed_at" field.
func CompletedAtIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldCompletedAt, vs...))
}

// CompletedAtNotIn applies the NotIn predicate on the "completed_at" field.
func CompletedAtNotIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldCompletedAt, vs...))
}

// CompletedAtGT applies the GT predicate on the "completed_at" field.
func CompletedAtGT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldCompletedAt, v))
}

// CompletedAtGTE applies the GTE predicate on the "completed_at" field.
func CompletedAtGTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldCompletedAt, v))
}

// CompletedAtLT applies the LT predicate on the "completed_at" field.
func CompletedAtLT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldCompletedAt, v))
}

// CompletedAtLTE applies the LTE predicate on the "completed_at" field.
func CompletedAtLTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldCompletedAt, v))
}

// CompletedAtIsNil applies the IsNil predicate on the "completed_at" field.
func CompletedAtIsNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIsNull(FieldCompletedAt))
}

// CompletedAtNotNil applies the NotNil predicate on the "completed_at" field.
func CompletedAtNotNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotNull(FieldCompletedAt))
}

// StatusEQ applies the EQ predicate on the "status" field.
func StatusEQ(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStatus, vc))
}

// StatusNEQ applies the NEQ predicate on the "status" field.
func StatusNEQ(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldStatus, vc))
}

// StatusIn applies the In predicate on the "status" field.
func StatusIn(vs ...schema.ExecutionStatus) predicate.WorkflowExecution {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowExecution(sql.FieldIn(FieldStatus, v...))
}

// StatusNotIn applies the NotIn predicate on the "status" field.
func StatusNotIn(vs ...schema.ExecutionStatus) predicate.WorkflowExecution {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldStatus, v...))
}

// StatusGT applies the GT predicate on the "status" field.
func StatusGT(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldGT(FieldStatus, vc))
}

// StatusGTE applies the GTE predicate on the "status" field.
func StatusGTE(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldGTE(FieldStatus, vc))
}

// StatusLT applies the LT predicate on the "status" field.
func StatusLT(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldLT(FieldStatus, vc))
}

// StatusLTE applies the LTE predicate on the "status" field.
func StatusLTE(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldLTE(FieldStatus, vc))
}

// StatusContains applies the Contains predicate on the "status" field.
func StatusContains(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldContains(FieldStatus, vc))
}

// StatusHasPrefix applies the HasPrefix predicate on the "status" field.
func StatusHasPrefix(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldHasPrefix(FieldStatus, vc))
}

// StatusHasSuffix applies the HasSuffix predicate on the "status" field.
func StatusHasSuffix(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldHasSuffix(FieldStatus, vc))
}

// StatusEqualFold applies the EqualFold predicate on the "status" field.
func StatusEqualFold(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldEqualFold(FieldStatus, vc))
}

// StatusContainsFold applies the ContainsFold predicate on the "status" field.
func StatusContainsFold(v schema.ExecutionStatus) predicate.WorkflowExecution {
	vc := string(v)
	return predicate.WorkflowExecution(sql.FieldContainsFold(FieldStatus, vc))
}

// ErrorEQ applies the EQ predicate on the "error" field.
func ErrorEQ(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldError, v))
}

// ErrorNEQ applies the NEQ predicate on the "error" field.
func ErrorNEQ(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldError, v))
}

// ErrorIn applies the In predicate on the "error" field.
func ErrorIn(vs ...string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldError, vs...))
}

// ErrorNotIn applies the NotIn predicate on the "error" field.
func ErrorNotIn(vs ...string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldError, vs...))
}

// ErrorGT applies the GT predicate on the "error" field.
func ErrorGT(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldError, v))
}

// ErrorGTE applies the GTE predicate on the "error" field.
func ErrorGTE(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldError, v))
}

// ErrorLT applies the LT predicate on the "error" field.
func ErrorLT(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldError, v))
}

// ErrorLTE applies the LTE predicate on the "error" field.
func ErrorLTE(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldError, v))
}

// ErrorContains applies the Contains predicate on the "error" field.
func ErrorContains(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldContains(FieldError, v))
}

// ErrorHasPrefix applies the HasPrefix predicate on the "error" field.
func ErrorHasPrefix(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldHasPrefix(FieldError, v))
}

// ErrorHasSuffix applies the HasSuffix predicate on the "error" field.
func ErrorHasSuffix(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldHasSuffix(FieldError, v))
}

// ErrorIsNil applies the IsNil predicate on the "error" field.
func ErrorIsNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIsNull(FieldError))
}

// ErrorNotNil applies the NotNil predicate on the "error" field.
func ErrorNotNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotNull(FieldError))
}

// ErrorEqualFold applies the EqualFold predicate on the "error" field.
func ErrorEqualFold(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEqualFold(FieldError, v))
}

// ErrorContainsFold applies the ContainsFold predicate on the "error" field.
func ErrorContainsFold(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldContainsFold(FieldError, v))
}

// StackTraceEQ applies the EQ predicate on the "stack_trace" field.
func StackTraceEQ(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldStackTrace, v))
}

// StackTraceNEQ applies the NEQ predicate on the "stack_trace" field.
func StackTraceNEQ(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldStackTrace, v))
}

// StackTraceIn applies the In predicate on the "stack_trace" field.
func StackTraceIn(vs ...string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldStackTrace, vs...))
}

// StackTraceNotIn applies the NotIn predicate on the "stack_trace" field.
func StackTraceNotIn(vs ...string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldStackTrace, vs...))
}

// StackTraceGT applies the GT predicate on the "stack_trace" field.
func StackTraceGT(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldStackTrace, v))
}

// StackTraceGTE applies the GTE predicate on the "stack_trace" field.
func StackTraceGTE(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldStackTrace, v))
}

// StackTraceLT applies the LT predicate on the "stack_trace" field.
func StackTraceLT(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldStackTrace, v))
}

// StackTraceLTE applies the LTE predicate on the "stack_trace" field.
func StackTraceLTE(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldStackTrace, v))
}

// StackTraceContains applies the Contains predicate on the "stack_trace" field.
func StackTraceContains(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldContains(FieldStackTrace, v))
}

// StackTraceHasPrefix applies the HasPrefix predicate on the "stack_trace" field.
func StackTraceHasPrefix(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldHasPrefix(FieldStackTrace, v))
}

// StackTraceHasSuffix applies the HasSuffix predicate on the "stack_trace" field.
func StackTraceHasSuffix(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldHasSuffix(FieldStackTrace, v))
}

// StackTraceIsNil applies the IsNil predicate on the "stack_trace" field.
func StackTraceIsNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIsNull(FieldStackTrace))
}

// StackTraceNotNil applies the NotNil predicate on the "stack_trace" field.
func StackTraceNotNil() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotNull(FieldStackTrace))
}

// StackTraceEqualFold applies the EqualFold predicate on the "stack_trace" field.
func StackTraceEqualFold(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEqualFold(FieldStackTrace, v))
}

// StackTraceContainsFold applies the ContainsFold predicate on the "stack_trace" field.
func StackTraceContainsFold(v string) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldContainsFold(FieldStackTrace, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasWorkflow applies the HasEdge predicate on the "workflow" edge.
func HasWorkflow() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, WorkflowTable, WorkflowColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasWorkflowWith applies the HasEdge predicate on the "workflow" edge with a given conditions (other predicates).
func HasWorkflowWith(preds ...predicate.WorkflowEntity) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := newWorkflowStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasExecutionData applies the HasEdge predicate on the "execution_data" edge.
func HasExecutionData() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, ExecutionDataTable, ExecutionDataColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionDataWith applies the HasEdge predicate on the "execution_data" edge with a given conditions (other predicates).
func HasExecutionDataWith(preds ...predicate.WorkflowExecutionData) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := newExecutionDataStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasEvents applies the HasEdge predicate on the "events" edge.
func HasEvents() predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, EventsTable, EventsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventsWith applies the HasEdge predicate on the "events" edge with a given conditions (other predicates).
func HasEventsWith(preds ...predicate.EventLog) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(func(s *sql.Selector) {
		step := newEventsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.WorkflowExecution) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.WorkflowExecution) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.WorkflowExecution) predicate.WorkflowExecution {
	return predicate.WorkflowExecution(sql.NotPredicates(p))
}
