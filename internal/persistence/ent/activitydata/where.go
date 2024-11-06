// Code generated by ent, DO NOT EDIT.

package activitydata

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLTE(FieldID, id))
}

// Timeout applies equality check predicate on the "timeout" field. It's identical to TimeoutEQ.
func Timeout(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldTimeout, v))
}

// MaxAttempts applies equality check predicate on the "max_attempts" field. It's identical to MaxAttemptsEQ.
func MaxAttempts(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldMaxAttempts, v))
}

// ScheduledFor applies equality check predicate on the "scheduled_for" field. It's identical to ScheduledForEQ.
func ScheduledFor(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldScheduledFor, v))
}

// TimeoutEQ applies the EQ predicate on the "timeout" field.
func TimeoutEQ(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldTimeout, v))
}

// TimeoutNEQ applies the NEQ predicate on the "timeout" field.
func TimeoutNEQ(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNEQ(FieldTimeout, v))
}

// TimeoutIn applies the In predicate on the "timeout" field.
func TimeoutIn(vs ...int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIn(FieldTimeout, vs...))
}

// TimeoutNotIn applies the NotIn predicate on the "timeout" field.
func TimeoutNotIn(vs ...int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotIn(FieldTimeout, vs...))
}

// TimeoutGT applies the GT predicate on the "timeout" field.
func TimeoutGT(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGT(FieldTimeout, v))
}

// TimeoutGTE applies the GTE predicate on the "timeout" field.
func TimeoutGTE(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGTE(FieldTimeout, v))
}

// TimeoutLT applies the LT predicate on the "timeout" field.
func TimeoutLT(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLT(FieldTimeout, v))
}

// TimeoutLTE applies the LTE predicate on the "timeout" field.
func TimeoutLTE(v int64) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLTE(FieldTimeout, v))
}

// TimeoutIsNil applies the IsNil predicate on the "timeout" field.
func TimeoutIsNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIsNull(FieldTimeout))
}

// TimeoutNotNil applies the NotNil predicate on the "timeout" field.
func TimeoutNotNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotNull(FieldTimeout))
}

// MaxAttemptsEQ applies the EQ predicate on the "max_attempts" field.
func MaxAttemptsEQ(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldMaxAttempts, v))
}

// MaxAttemptsNEQ applies the NEQ predicate on the "max_attempts" field.
func MaxAttemptsNEQ(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNEQ(FieldMaxAttempts, v))
}

// MaxAttemptsIn applies the In predicate on the "max_attempts" field.
func MaxAttemptsIn(vs ...int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIn(FieldMaxAttempts, vs...))
}

// MaxAttemptsNotIn applies the NotIn predicate on the "max_attempts" field.
func MaxAttemptsNotIn(vs ...int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotIn(FieldMaxAttempts, vs...))
}

// MaxAttemptsGT applies the GT predicate on the "max_attempts" field.
func MaxAttemptsGT(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGT(FieldMaxAttempts, v))
}

// MaxAttemptsGTE applies the GTE predicate on the "max_attempts" field.
func MaxAttemptsGTE(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGTE(FieldMaxAttempts, v))
}

// MaxAttemptsLT applies the LT predicate on the "max_attempts" field.
func MaxAttemptsLT(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLT(FieldMaxAttempts, v))
}

// MaxAttemptsLTE applies the LTE predicate on the "max_attempts" field.
func MaxAttemptsLTE(v int) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLTE(FieldMaxAttempts, v))
}

// ScheduledForEQ applies the EQ predicate on the "scheduled_for" field.
func ScheduledForEQ(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldEQ(FieldScheduledFor, v))
}

// ScheduledForNEQ applies the NEQ predicate on the "scheduled_for" field.
func ScheduledForNEQ(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNEQ(FieldScheduledFor, v))
}

// ScheduledForIn applies the In predicate on the "scheduled_for" field.
func ScheduledForIn(vs ...time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIn(FieldScheduledFor, vs...))
}

// ScheduledForNotIn applies the NotIn predicate on the "scheduled_for" field.
func ScheduledForNotIn(vs ...time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotIn(FieldScheduledFor, vs...))
}

// ScheduledForGT applies the GT predicate on the "scheduled_for" field.
func ScheduledForGT(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGT(FieldScheduledFor, v))
}

// ScheduledForGTE applies the GTE predicate on the "scheduled_for" field.
func ScheduledForGTE(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldGTE(FieldScheduledFor, v))
}

// ScheduledForLT applies the LT predicate on the "scheduled_for" field.
func ScheduledForLT(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLT(FieldScheduledFor, v))
}

// ScheduledForLTE applies the LTE predicate on the "scheduled_for" field.
func ScheduledForLTE(v time.Time) predicate.ActivityData {
	return predicate.ActivityData(sql.FieldLTE(FieldScheduledFor, v))
}

// ScheduledForIsNil applies the IsNil predicate on the "scheduled_for" field.
func ScheduledForIsNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIsNull(FieldScheduledFor))
}

// ScheduledForNotNil applies the NotNil predicate on the "scheduled_for" field.
func ScheduledForNotNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotNull(FieldScheduledFor))
}

// InputIsNil applies the IsNil predicate on the "input" field.
func InputIsNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIsNull(FieldInput))
}

// InputNotNil applies the NotNil predicate on the "input" field.
func InputNotNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotNull(FieldInput))
}

// OutputIsNil applies the IsNil predicate on the "output" field.
func OutputIsNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldIsNull(FieldOutput))
}

// OutputNotNil applies the NotNil predicate on the "output" field.
func OutputNotNil() predicate.ActivityData {
	return predicate.ActivityData(sql.FieldNotNull(FieldOutput))
}

// HasEntity applies the HasEdge predicate on the "entity" edge.
func HasEntity() predicate.ActivityData {
	return predicate.ActivityData(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, EntityTable, EntityColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEntityWith applies the HasEdge predicate on the "entity" edge with a given conditions (other predicates).
func HasEntityWith(preds ...predicate.Entity) predicate.ActivityData {
	return predicate.ActivityData(func(s *sql.Selector) {
		step := newEntityStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.ActivityData) predicate.ActivityData {
	return predicate.ActivityData(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.ActivityData) predicate.ActivityData {
	return predicate.ActivityData(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.ActivityData) predicate.ActivityData {
	return predicate.ActivityData(sql.NotPredicates(p))
}
