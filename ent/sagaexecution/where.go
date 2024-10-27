// Code generated by ent, DO NOT EDIT.

package sagaexecution

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContainsFold(FieldID, id))
}

// HandlerName applies equality check predicate on the "handler_name" field. It's identical to HandlerNameEQ.
func HandlerName(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldHandlerName, v))
}

// QueueName applies equality check predicate on the "queue_name" field. It's identical to QueueNameEQ.
func QueueName(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldQueueName, v))
}

// Sequence applies equality check predicate on the "sequence" field. It's identical to SequenceEQ.
func Sequence(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldSequence, v))
}

// Error applies equality check predicate on the "error" field. It's identical to ErrorEQ.
func Error(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldError, v))
}

// StartedAt applies equality check predicate on the "started_at" field. It's identical to StartedAtEQ.
func StartedAt(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldStartedAt, v))
}

// CompletedAt applies equality check predicate on the "completed_at" field. It's identical to CompletedAtEQ.
func CompletedAt(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldCompletedAt, v))
}

// HandlerNameEQ applies the EQ predicate on the "handler_name" field.
func HandlerNameEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldHandlerName, v))
}

// HandlerNameNEQ applies the NEQ predicate on the "handler_name" field.
func HandlerNameNEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldHandlerName, v))
}

// HandlerNameIn applies the In predicate on the "handler_name" field.
func HandlerNameIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldHandlerName, vs...))
}

// HandlerNameNotIn applies the NotIn predicate on the "handler_name" field.
func HandlerNameNotIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldHandlerName, vs...))
}

// HandlerNameGT applies the GT predicate on the "handler_name" field.
func HandlerNameGT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldHandlerName, v))
}

// HandlerNameGTE applies the GTE predicate on the "handler_name" field.
func HandlerNameGTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldHandlerName, v))
}

// HandlerNameLT applies the LT predicate on the "handler_name" field.
func HandlerNameLT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldHandlerName, v))
}

// HandlerNameLTE applies the LTE predicate on the "handler_name" field.
func HandlerNameLTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldHandlerName, v))
}

// HandlerNameContains applies the Contains predicate on the "handler_name" field.
func HandlerNameContains(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContains(FieldHandlerName, v))
}

// HandlerNameHasPrefix applies the HasPrefix predicate on the "handler_name" field.
func HandlerNameHasPrefix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasPrefix(FieldHandlerName, v))
}

// HandlerNameHasSuffix applies the HasSuffix predicate on the "handler_name" field.
func HandlerNameHasSuffix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasSuffix(FieldHandlerName, v))
}

// HandlerNameEqualFold applies the EqualFold predicate on the "handler_name" field.
func HandlerNameEqualFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEqualFold(FieldHandlerName, v))
}

// HandlerNameContainsFold applies the ContainsFold predicate on the "handler_name" field.
func HandlerNameContainsFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContainsFold(FieldHandlerName, v))
}

// StepTypeEQ applies the EQ predicate on the "step_type" field.
func StepTypeEQ(v StepType) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldStepType, v))
}

// StepTypeNEQ applies the NEQ predicate on the "step_type" field.
func StepTypeNEQ(v StepType) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldStepType, v))
}

// StepTypeIn applies the In predicate on the "step_type" field.
func StepTypeIn(vs ...StepType) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldStepType, vs...))
}

// StepTypeNotIn applies the NotIn predicate on the "step_type" field.
func StepTypeNotIn(vs ...StepType) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldStepType, vs...))
}

// StatusEQ applies the EQ predicate on the "status" field.
func StatusEQ(v Status) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldStatus, v))
}

// StatusNEQ applies the NEQ predicate on the "status" field.
func StatusNEQ(v Status) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldStatus, v))
}

// StatusIn applies the In predicate on the "status" field.
func StatusIn(vs ...Status) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldStatus, vs...))
}

// StatusNotIn applies the NotIn predicate on the "status" field.
func StatusNotIn(vs ...Status) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldStatus, vs...))
}

// QueueNameEQ applies the EQ predicate on the "queue_name" field.
func QueueNameEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldQueueName, v))
}

// QueueNameNEQ applies the NEQ predicate on the "queue_name" field.
func QueueNameNEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldQueueName, v))
}

// QueueNameIn applies the In predicate on the "queue_name" field.
func QueueNameIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldQueueName, vs...))
}

// QueueNameNotIn applies the NotIn predicate on the "queue_name" field.
func QueueNameNotIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldQueueName, vs...))
}

// QueueNameGT applies the GT predicate on the "queue_name" field.
func QueueNameGT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldQueueName, v))
}

// QueueNameGTE applies the GTE predicate on the "queue_name" field.
func QueueNameGTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldQueueName, v))
}

// QueueNameLT applies the LT predicate on the "queue_name" field.
func QueueNameLT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldQueueName, v))
}

// QueueNameLTE applies the LTE predicate on the "queue_name" field.
func QueueNameLTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldQueueName, v))
}

// QueueNameContains applies the Contains predicate on the "queue_name" field.
func QueueNameContains(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContains(FieldQueueName, v))
}

// QueueNameHasPrefix applies the HasPrefix predicate on the "queue_name" field.
func QueueNameHasPrefix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasPrefix(FieldQueueName, v))
}

// QueueNameHasSuffix applies the HasSuffix predicate on the "queue_name" field.
func QueueNameHasSuffix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasSuffix(FieldQueueName, v))
}

// QueueNameEqualFold applies the EqualFold predicate on the "queue_name" field.
func QueueNameEqualFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEqualFold(FieldQueueName, v))
}

// QueueNameContainsFold applies the ContainsFold predicate on the "queue_name" field.
func QueueNameContainsFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContainsFold(FieldQueueName, v))
}

// SequenceEQ applies the EQ predicate on the "sequence" field.
func SequenceEQ(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldSequence, v))
}

// SequenceNEQ applies the NEQ predicate on the "sequence" field.
func SequenceNEQ(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldSequence, v))
}

// SequenceIn applies the In predicate on the "sequence" field.
func SequenceIn(vs ...int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldSequence, vs...))
}

// SequenceNotIn applies the NotIn predicate on the "sequence" field.
func SequenceNotIn(vs ...int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldSequence, vs...))
}

// SequenceGT applies the GT predicate on the "sequence" field.
func SequenceGT(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldSequence, v))
}

// SequenceGTE applies the GTE predicate on the "sequence" field.
func SequenceGTE(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldSequence, v))
}

// SequenceLT applies the LT predicate on the "sequence" field.
func SequenceLT(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldSequence, v))
}

// SequenceLTE applies the LTE predicate on the "sequence" field.
func SequenceLTE(v int) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldSequence, v))
}

// ErrorEQ applies the EQ predicate on the "error" field.
func ErrorEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldError, v))
}

// ErrorNEQ applies the NEQ predicate on the "error" field.
func ErrorNEQ(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldError, v))
}

// ErrorIn applies the In predicate on the "error" field.
func ErrorIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldError, vs...))
}

// ErrorNotIn applies the NotIn predicate on the "error" field.
func ErrorNotIn(vs ...string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldError, vs...))
}

// ErrorGT applies the GT predicate on the "error" field.
func ErrorGT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldError, v))
}

// ErrorGTE applies the GTE predicate on the "error" field.
func ErrorGTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldError, v))
}

// ErrorLT applies the LT predicate on the "error" field.
func ErrorLT(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldError, v))
}

// ErrorLTE applies the LTE predicate on the "error" field.
func ErrorLTE(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldError, v))
}

// ErrorContains applies the Contains predicate on the "error" field.
func ErrorContains(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContains(FieldError, v))
}

// ErrorHasPrefix applies the HasPrefix predicate on the "error" field.
func ErrorHasPrefix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasPrefix(FieldError, v))
}

// ErrorHasSuffix applies the HasSuffix predicate on the "error" field.
func ErrorHasSuffix(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldHasSuffix(FieldError, v))
}

// ErrorIsNil applies the IsNil predicate on the "error" field.
func ErrorIsNil() predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIsNull(FieldError))
}

// ErrorNotNil applies the NotNil predicate on the "error" field.
func ErrorNotNil() predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotNull(FieldError))
}

// ErrorEqualFold applies the EqualFold predicate on the "error" field.
func ErrorEqualFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEqualFold(FieldError, v))
}

// ErrorContainsFold applies the ContainsFold predicate on the "error" field.
func ErrorContainsFold(v string) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldContainsFold(FieldError, v))
}

// StartedAtEQ applies the EQ predicate on the "started_at" field.
func StartedAtEQ(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldStartedAt, v))
}

// StartedAtNEQ applies the NEQ predicate on the "started_at" field.
func StartedAtNEQ(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldStartedAt, v))
}

// StartedAtIn applies the In predicate on the "started_at" field.
func StartedAtIn(vs ...time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldStartedAt, vs...))
}

// StartedAtNotIn applies the NotIn predicate on the "started_at" field.
func StartedAtNotIn(vs ...time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldStartedAt, vs...))
}

// StartedAtGT applies the GT predicate on the "started_at" field.
func StartedAtGT(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldStartedAt, v))
}

// StartedAtGTE applies the GTE predicate on the "started_at" field.
func StartedAtGTE(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldStartedAt, v))
}

// StartedAtLT applies the LT predicate on the "started_at" field.
func StartedAtLT(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldStartedAt, v))
}

// StartedAtLTE applies the LTE predicate on the "started_at" field.
func StartedAtLTE(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldStartedAt, v))
}

// CompletedAtEQ applies the EQ predicate on the "completed_at" field.
func CompletedAtEQ(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldEQ(FieldCompletedAt, v))
}

// CompletedAtNEQ applies the NEQ predicate on the "completed_at" field.
func CompletedAtNEQ(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNEQ(FieldCompletedAt, v))
}

// CompletedAtIn applies the In predicate on the "completed_at" field.
func CompletedAtIn(vs ...time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIn(FieldCompletedAt, vs...))
}

// CompletedAtNotIn applies the NotIn predicate on the "completed_at" field.
func CompletedAtNotIn(vs ...time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotIn(FieldCompletedAt, vs...))
}

// CompletedAtGT applies the GT predicate on the "completed_at" field.
func CompletedAtGT(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGT(FieldCompletedAt, v))
}

// CompletedAtGTE applies the GTE predicate on the "completed_at" field.
func CompletedAtGTE(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldGTE(FieldCompletedAt, v))
}

// CompletedAtLT applies the LT predicate on the "completed_at" field.
func CompletedAtLT(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLT(FieldCompletedAt, v))
}

// CompletedAtLTE applies the LTE predicate on the "completed_at" field.
func CompletedAtLTE(v time.Time) predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldLTE(FieldCompletedAt, v))
}

// CompletedAtIsNil applies the IsNil predicate on the "completed_at" field.
func CompletedAtIsNil() predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldIsNull(FieldCompletedAt))
}

// CompletedAtNotNil applies the NotNil predicate on the "completed_at" field.
func CompletedAtNotNil() predicate.SagaExecution {
	return predicate.SagaExecution(sql.FieldNotNull(FieldCompletedAt))
}

// HasSaga applies the HasEdge predicate on the "saga" edge.
func HasSaga() predicate.SagaExecution {
	return predicate.SagaExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, SagaTable, SagaColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSagaWith applies the HasEdge predicate on the "saga" edge with a given conditions (other predicates).
func HasSagaWith(preds ...predicate.Saga) predicate.SagaExecution {
	return predicate.SagaExecution(func(s *sql.Selector) {
		step := newSagaStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SagaExecution) predicate.SagaExecution {
	return predicate.SagaExecution(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SagaExecution) predicate.SagaExecution {
	return predicate.SagaExecution(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SagaExecution) predicate.SagaExecution {
	return predicate.SagaExecution(sql.NotPredicates(p))
}
