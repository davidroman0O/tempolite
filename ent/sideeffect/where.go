// Code generated by ent, DO NOT EDIT.

package sideeffect

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContainsFold(FieldID, id))
}

// Identity applies equality check predicate on the "identity" field. It's identical to IdentityEQ.
func Identity(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldIdentity, v))
}

// StepID applies equality check predicate on the "step_id" field. It's identical to StepIDEQ.
func StepID(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldStepID, v))
}

// HandlerName applies equality check predicate on the "handler_name" field. It's identical to HandlerNameEQ.
func HandlerName(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldHandlerName, v))
}

// Timeout applies equality check predicate on the "timeout" field. It's identical to TimeoutEQ.
func Timeout(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldTimeout, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldCreatedAt, v))
}

// IdentityEQ applies the EQ predicate on the "identity" field.
func IdentityEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldIdentity, v))
}

// IdentityNEQ applies the NEQ predicate on the "identity" field.
func IdentityNEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldIdentity, v))
}

// IdentityIn applies the In predicate on the "identity" field.
func IdentityIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldIdentity, vs...))
}

// IdentityNotIn applies the NotIn predicate on the "identity" field.
func IdentityNotIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldIdentity, vs...))
}

// IdentityGT applies the GT predicate on the "identity" field.
func IdentityGT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldIdentity, v))
}

// IdentityGTE applies the GTE predicate on the "identity" field.
func IdentityGTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldIdentity, v))
}

// IdentityLT applies the LT predicate on the "identity" field.
func IdentityLT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldIdentity, v))
}

// IdentityLTE applies the LTE predicate on the "identity" field.
func IdentityLTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldIdentity, v))
}

// IdentityContains applies the Contains predicate on the "identity" field.
func IdentityContains(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContains(FieldIdentity, v))
}

// IdentityHasPrefix applies the HasPrefix predicate on the "identity" field.
func IdentityHasPrefix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasPrefix(FieldIdentity, v))
}

// IdentityHasSuffix applies the HasSuffix predicate on the "identity" field.
func IdentityHasSuffix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasSuffix(FieldIdentity, v))
}

// IdentityEqualFold applies the EqualFold predicate on the "identity" field.
func IdentityEqualFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEqualFold(FieldIdentity, v))
}

// IdentityContainsFold applies the ContainsFold predicate on the "identity" field.
func IdentityContainsFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContainsFold(FieldIdentity, v))
}

// StepIDEQ applies the EQ predicate on the "step_id" field.
func StepIDEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldStepID, v))
}

// StepIDNEQ applies the NEQ predicate on the "step_id" field.
func StepIDNEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldStepID, v))
}

// StepIDIn applies the In predicate on the "step_id" field.
func StepIDIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldStepID, vs...))
}

// StepIDNotIn applies the NotIn predicate on the "step_id" field.
func StepIDNotIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldStepID, vs...))
}

// StepIDGT applies the GT predicate on the "step_id" field.
func StepIDGT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldStepID, v))
}

// StepIDGTE applies the GTE predicate on the "step_id" field.
func StepIDGTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldStepID, v))
}

// StepIDLT applies the LT predicate on the "step_id" field.
func StepIDLT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldStepID, v))
}

// StepIDLTE applies the LTE predicate on the "step_id" field.
func StepIDLTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldStepID, v))
}

// StepIDContains applies the Contains predicate on the "step_id" field.
func StepIDContains(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContains(FieldStepID, v))
}

// StepIDHasPrefix applies the HasPrefix predicate on the "step_id" field.
func StepIDHasPrefix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasPrefix(FieldStepID, v))
}

// StepIDHasSuffix applies the HasSuffix predicate on the "step_id" field.
func StepIDHasSuffix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasSuffix(FieldStepID, v))
}

// StepIDEqualFold applies the EqualFold predicate on the "step_id" field.
func StepIDEqualFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEqualFold(FieldStepID, v))
}

// StepIDContainsFold applies the ContainsFold predicate on the "step_id" field.
func StepIDContainsFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContainsFold(FieldStepID, v))
}

// HandlerNameEQ applies the EQ predicate on the "handler_name" field.
func HandlerNameEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldHandlerName, v))
}

// HandlerNameNEQ applies the NEQ predicate on the "handler_name" field.
func HandlerNameNEQ(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldHandlerName, v))
}

// HandlerNameIn applies the In predicate on the "handler_name" field.
func HandlerNameIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldHandlerName, vs...))
}

// HandlerNameNotIn applies the NotIn predicate on the "handler_name" field.
func HandlerNameNotIn(vs ...string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldHandlerName, vs...))
}

// HandlerNameGT applies the GT predicate on the "handler_name" field.
func HandlerNameGT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldHandlerName, v))
}

// HandlerNameGTE applies the GTE predicate on the "handler_name" field.
func HandlerNameGTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldHandlerName, v))
}

// HandlerNameLT applies the LT predicate on the "handler_name" field.
func HandlerNameLT(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldHandlerName, v))
}

// HandlerNameLTE applies the LTE predicate on the "handler_name" field.
func HandlerNameLTE(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldHandlerName, v))
}

// HandlerNameContains applies the Contains predicate on the "handler_name" field.
func HandlerNameContains(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContains(FieldHandlerName, v))
}

// HandlerNameHasPrefix applies the HasPrefix predicate on the "handler_name" field.
func HandlerNameHasPrefix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasPrefix(FieldHandlerName, v))
}

// HandlerNameHasSuffix applies the HasSuffix predicate on the "handler_name" field.
func HandlerNameHasSuffix(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldHasSuffix(FieldHandlerName, v))
}

// HandlerNameEqualFold applies the EqualFold predicate on the "handler_name" field.
func HandlerNameEqualFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEqualFold(FieldHandlerName, v))
}

// HandlerNameContainsFold applies the ContainsFold predicate on the "handler_name" field.
func HandlerNameContainsFold(v string) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldContainsFold(FieldHandlerName, v))
}

// StatusEQ applies the EQ predicate on the "status" field.
func StatusEQ(v Status) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldStatus, v))
}

// StatusNEQ applies the NEQ predicate on the "status" field.
func StatusNEQ(v Status) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldStatus, v))
}

// StatusIn applies the In predicate on the "status" field.
func StatusIn(vs ...Status) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldStatus, vs...))
}

// StatusNotIn applies the NotIn predicate on the "status" field.
func StatusNotIn(vs ...Status) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldStatus, vs...))
}

// RetryPolicyIsNil applies the IsNil predicate on the "retry_policy" field.
func RetryPolicyIsNil() predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIsNull(FieldRetryPolicy))
}

// RetryPolicyNotNil applies the NotNil predicate on the "retry_policy" field.
func RetryPolicyNotNil() predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotNull(FieldRetryPolicy))
}

// TimeoutEQ applies the EQ predicate on the "timeout" field.
func TimeoutEQ(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldTimeout, v))
}

// TimeoutNEQ applies the NEQ predicate on the "timeout" field.
func TimeoutNEQ(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldTimeout, v))
}

// TimeoutIn applies the In predicate on the "timeout" field.
func TimeoutIn(vs ...time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldTimeout, vs...))
}

// TimeoutNotIn applies the NotIn predicate on the "timeout" field.
func TimeoutNotIn(vs ...time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldTimeout, vs...))
}

// TimeoutGT applies the GT predicate on the "timeout" field.
func TimeoutGT(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldTimeout, v))
}

// TimeoutGTE applies the GTE predicate on the "timeout" field.
func TimeoutGTE(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldTimeout, v))
}

// TimeoutLT applies the LT predicate on the "timeout" field.
func TimeoutLT(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldTimeout, v))
}

// TimeoutLTE applies the LTE predicate on the "timeout" field.
func TimeoutLTE(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldTimeout, v))
}

// TimeoutIsNil applies the IsNil predicate on the "timeout" field.
func TimeoutIsNil() predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIsNull(FieldTimeout))
}

// TimeoutNotNil applies the NotNil predicate on the "timeout" field.
func TimeoutNotNil() predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotNull(FieldTimeout))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SideEffect {
	return predicate.SideEffect(sql.FieldLTE(FieldCreatedAt, v))
}

// HasExecutions applies the HasEdge predicate on the "executions" edge.
func HasExecutions() predicate.SideEffect {
	return predicate.SideEffect(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, ExecutionsTable, ExecutionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionsWith applies the HasEdge predicate on the "executions" edge with a given conditions (other predicates).
func HasExecutionsWith(preds ...predicate.SideEffectExecution) predicate.SideEffect {
	return predicate.SideEffect(func(s *sql.Selector) {
		step := newExecutionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SideEffect) predicate.SideEffect {
	return predicate.SideEffect(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SideEffect) predicate.SideEffect {
	return predicate.SideEffect(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SideEffect) predicate.SideEffect {
	return predicate.SideEffect(sql.NotPredicates(p))
}
