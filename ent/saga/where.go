// Code generated by ent, DO NOT EDIT.

package saga

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/go-tempolite/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Saga {
	return predicate.Saga(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Saga {
	return predicate.Saga(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Saga {
	return predicate.Saga(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Saga {
	return predicate.Saga(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Saga {
	return predicate.Saga(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Saga {
	return predicate.Saga(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Saga {
	return predicate.Saga(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.Saga {
	return predicate.Saga(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.Saga {
	return predicate.Saga(sql.FieldContainsFold(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldName, v))
}

// StepID applies equality check predicate on the "step_id" field. It's identical to StepIDEQ.
func StepID(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldStepID, v))
}

// Timeout applies equality check predicate on the "timeout" field. It's identical to TimeoutEQ.
func Timeout(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldTimeout, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldCreatedAt, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Saga {
	return predicate.Saga(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Saga {
	return predicate.Saga(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Saga {
	return predicate.Saga(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Saga {
	return predicate.Saga(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Saga {
	return predicate.Saga(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Saga {
	return predicate.Saga(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Saga {
	return predicate.Saga(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Saga {
	return predicate.Saga(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Saga {
	return predicate.Saga(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Saga {
	return predicate.Saga(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Saga {
	return predicate.Saga(sql.FieldContainsFold(FieldName, v))
}

// StepIDEQ applies the EQ predicate on the "step_id" field.
func StepIDEQ(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldStepID, v))
}

// StepIDNEQ applies the NEQ predicate on the "step_id" field.
func StepIDNEQ(v string) predicate.Saga {
	return predicate.Saga(sql.FieldNEQ(FieldStepID, v))
}

// StepIDIn applies the In predicate on the "step_id" field.
func StepIDIn(vs ...string) predicate.Saga {
	return predicate.Saga(sql.FieldIn(FieldStepID, vs...))
}

// StepIDNotIn applies the NotIn predicate on the "step_id" field.
func StepIDNotIn(vs ...string) predicate.Saga {
	return predicate.Saga(sql.FieldNotIn(FieldStepID, vs...))
}

// StepIDGT applies the GT predicate on the "step_id" field.
func StepIDGT(v string) predicate.Saga {
	return predicate.Saga(sql.FieldGT(FieldStepID, v))
}

// StepIDGTE applies the GTE predicate on the "step_id" field.
func StepIDGTE(v string) predicate.Saga {
	return predicate.Saga(sql.FieldGTE(FieldStepID, v))
}

// StepIDLT applies the LT predicate on the "step_id" field.
func StepIDLT(v string) predicate.Saga {
	return predicate.Saga(sql.FieldLT(FieldStepID, v))
}

// StepIDLTE applies the LTE predicate on the "step_id" field.
func StepIDLTE(v string) predicate.Saga {
	return predicate.Saga(sql.FieldLTE(FieldStepID, v))
}

// StepIDContains applies the Contains predicate on the "step_id" field.
func StepIDContains(v string) predicate.Saga {
	return predicate.Saga(sql.FieldContains(FieldStepID, v))
}

// StepIDHasPrefix applies the HasPrefix predicate on the "step_id" field.
func StepIDHasPrefix(v string) predicate.Saga {
	return predicate.Saga(sql.FieldHasPrefix(FieldStepID, v))
}

// StepIDHasSuffix applies the HasSuffix predicate on the "step_id" field.
func StepIDHasSuffix(v string) predicate.Saga {
	return predicate.Saga(sql.FieldHasSuffix(FieldStepID, v))
}

// StepIDEqualFold applies the EqualFold predicate on the "step_id" field.
func StepIDEqualFold(v string) predicate.Saga {
	return predicate.Saga(sql.FieldEqualFold(FieldStepID, v))
}

// StepIDContainsFold applies the ContainsFold predicate on the "step_id" field.
func StepIDContainsFold(v string) predicate.Saga {
	return predicate.Saga(sql.FieldContainsFold(FieldStepID, v))
}

// RetryPolicyIsNil applies the IsNil predicate on the "retry_policy" field.
func RetryPolicyIsNil() predicate.Saga {
	return predicate.Saga(sql.FieldIsNull(FieldRetryPolicy))
}

// RetryPolicyNotNil applies the NotNil predicate on the "retry_policy" field.
func RetryPolicyNotNil() predicate.Saga {
	return predicate.Saga(sql.FieldNotNull(FieldRetryPolicy))
}

// TimeoutEQ applies the EQ predicate on the "timeout" field.
func TimeoutEQ(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldTimeout, v))
}

// TimeoutNEQ applies the NEQ predicate on the "timeout" field.
func TimeoutNEQ(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldNEQ(FieldTimeout, v))
}

// TimeoutIn applies the In predicate on the "timeout" field.
func TimeoutIn(vs ...time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldIn(FieldTimeout, vs...))
}

// TimeoutNotIn applies the NotIn predicate on the "timeout" field.
func TimeoutNotIn(vs ...time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldNotIn(FieldTimeout, vs...))
}

// TimeoutGT applies the GT predicate on the "timeout" field.
func TimeoutGT(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldGT(FieldTimeout, v))
}

// TimeoutGTE applies the GTE predicate on the "timeout" field.
func TimeoutGTE(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldGTE(FieldTimeout, v))
}

// TimeoutLT applies the LT predicate on the "timeout" field.
func TimeoutLT(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldLT(FieldTimeout, v))
}

// TimeoutLTE applies the LTE predicate on the "timeout" field.
func TimeoutLTE(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldLTE(FieldTimeout, v))
}

// TimeoutIsNil applies the IsNil predicate on the "timeout" field.
func TimeoutIsNil() predicate.Saga {
	return predicate.Saga(sql.FieldIsNull(FieldTimeout))
}

// TimeoutNotNil applies the NotNil predicate on the "timeout" field.
func TimeoutNotNil() predicate.Saga {
	return predicate.Saga(sql.FieldNotNull(FieldTimeout))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.Saga {
	return predicate.Saga(sql.FieldLTE(FieldCreatedAt, v))
}

// HasExecutions applies the HasEdge predicate on the "executions" edge.
func HasExecutions() predicate.Saga {
	return predicate.Saga(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, ExecutionsTable, ExecutionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionsWith applies the HasEdge predicate on the "executions" edge with a given conditions (other predicates).
func HasExecutionsWith(preds ...predicate.SagaExecution) predicate.Saga {
	return predicate.Saga(func(s *sql.Selector) {
		step := newExecutionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Saga) predicate.Saga {
	return predicate.Saga(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Saga) predicate.Saga {
	return predicate.Saga(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Saga) predicate.Saga {
	return predicate.Saga(sql.NotPredicates(p))
}
