// Code generated by ent, DO NOT EDIT.

package workflowdata

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldLTE(FieldID, id))
}

// Duration applies equality check predicate on the "duration" field. It's identical to DurationEQ.
func Duration(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldDuration, v))
}

// Paused applies equality check predicate on the "paused" field. It's identical to PausedEQ.
func Paused(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldPaused, v))
}

// Resumable applies equality check predicate on the "resumable" field. It's identical to ResumableEQ.
func Resumable(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldResumable, v))
}

// DurationEQ applies the EQ predicate on the "duration" field.
func DurationEQ(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldDuration, v))
}

// DurationNEQ applies the NEQ predicate on the "duration" field.
func DurationNEQ(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNEQ(FieldDuration, v))
}

// DurationIn applies the In predicate on the "duration" field.
func DurationIn(vs ...string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldIn(FieldDuration, vs...))
}

// DurationNotIn applies the NotIn predicate on the "duration" field.
func DurationNotIn(vs ...string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNotIn(FieldDuration, vs...))
}

// DurationGT applies the GT predicate on the "duration" field.
func DurationGT(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldGT(FieldDuration, v))
}

// DurationGTE applies the GTE predicate on the "duration" field.
func DurationGTE(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldGTE(FieldDuration, v))
}

// DurationLT applies the LT predicate on the "duration" field.
func DurationLT(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldLT(FieldDuration, v))
}

// DurationLTE applies the LTE predicate on the "duration" field.
func DurationLTE(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldLTE(FieldDuration, v))
}

// DurationContains applies the Contains predicate on the "duration" field.
func DurationContains(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldContains(FieldDuration, v))
}

// DurationHasPrefix applies the HasPrefix predicate on the "duration" field.
func DurationHasPrefix(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldHasPrefix(FieldDuration, v))
}

// DurationHasSuffix applies the HasSuffix predicate on the "duration" field.
func DurationHasSuffix(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldHasSuffix(FieldDuration, v))
}

// DurationIsNil applies the IsNil predicate on the "duration" field.
func DurationIsNil() predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldIsNull(FieldDuration))
}

// DurationNotNil applies the NotNil predicate on the "duration" field.
func DurationNotNil() predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNotNull(FieldDuration))
}

// DurationEqualFold applies the EqualFold predicate on the "duration" field.
func DurationEqualFold(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEqualFold(FieldDuration, v))
}

// DurationContainsFold applies the ContainsFold predicate on the "duration" field.
func DurationContainsFold(v string) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldContainsFold(FieldDuration, v))
}

// PausedEQ applies the EQ predicate on the "paused" field.
func PausedEQ(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldPaused, v))
}

// PausedNEQ applies the NEQ predicate on the "paused" field.
func PausedNEQ(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNEQ(FieldPaused, v))
}

// ResumableEQ applies the EQ predicate on the "resumable" field.
func ResumableEQ(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldEQ(FieldResumable, v))
}

// ResumableNEQ applies the NEQ predicate on the "resumable" field.
func ResumableNEQ(v bool) predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNEQ(FieldResumable, v))
}

// InputIsNil applies the IsNil predicate on the "input" field.
func InputIsNil() predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldIsNull(FieldInput))
}

// InputNotNil applies the NotNil predicate on the "input" field.
func InputNotNil() predicate.WorkflowData {
	return predicate.WorkflowData(sql.FieldNotNull(FieldInput))
}

// HasEntity applies the HasEdge predicate on the "entity" edge.
func HasEntity() predicate.WorkflowData {
	return predicate.WorkflowData(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, EntityTable, EntityColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEntityWith applies the HasEdge predicate on the "entity" edge with a given conditions (other predicates).
func HasEntityWith(preds ...predicate.Entity) predicate.WorkflowData {
	return predicate.WorkflowData(func(s *sql.Selector) {
		step := newEntityStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.WorkflowData) predicate.WorkflowData {
	return predicate.WorkflowData(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.WorkflowData) predicate.WorkflowData {
	return predicate.WorkflowData(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.WorkflowData) predicate.WorkflowData {
	return predicate.WorkflowData(sql.NotPredicates(p))
}
