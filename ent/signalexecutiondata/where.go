// Code generated by ent, DO NOT EDIT.

package signalexecutiondata

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/schema"
)

// ID filters vertices based on their ID field.
func ID(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id schema.SignalExecutionDataID) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLTE(FieldID, id))
}

// ExecutionID applies equality check predicate on the "execution_id" field. It's identical to ExecutionIDEQ.
func ExecutionID(v schema.SignalExecutionID) predicate.SignalExecutionData {
	vc := int(v)
	return predicate.SignalExecutionData(sql.FieldEQ(FieldExecutionID, vc))
}

// Value applies equality check predicate on the "value" field. It's identical to ValueEQ.
func Value(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldValue, v))
}

// Kind applies equality check predicate on the "kind" field. It's identical to KindEQ.
func Kind(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldKind, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldUpdatedAt, v))
}

// ExecutionIDEQ applies the EQ predicate on the "execution_id" field.
func ExecutionIDEQ(v schema.SignalExecutionID) predicate.SignalExecutionData {
	vc := int(v)
	return predicate.SignalExecutionData(sql.FieldEQ(FieldExecutionID, vc))
}

// ExecutionIDNEQ applies the NEQ predicate on the "execution_id" field.
func ExecutionIDNEQ(v schema.SignalExecutionID) predicate.SignalExecutionData {
	vc := int(v)
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldExecutionID, vc))
}

// ExecutionIDIn applies the In predicate on the "execution_id" field.
func ExecutionIDIn(vs ...schema.SignalExecutionID) predicate.SignalExecutionData {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.SignalExecutionData(sql.FieldIn(FieldExecutionID, v...))
}

// ExecutionIDNotIn applies the NotIn predicate on the "execution_id" field.
func ExecutionIDNotIn(vs ...schema.SignalExecutionID) predicate.SignalExecutionData {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldExecutionID, v...))
}

// ValueEQ applies the EQ predicate on the "value" field.
func ValueEQ(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldValue, v))
}

// ValueNEQ applies the NEQ predicate on the "value" field.
func ValueNEQ(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldValue, v))
}

// ValueIn applies the In predicate on the "value" field.
func ValueIn(vs ...[]byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIn(FieldValue, vs...))
}

// ValueNotIn applies the NotIn predicate on the "value" field.
func ValueNotIn(vs ...[]byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldValue, vs...))
}

// ValueGT applies the GT predicate on the "value" field.
func ValueGT(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGT(FieldValue, v))
}

// ValueGTE applies the GTE predicate on the "value" field.
func ValueGTE(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGTE(FieldValue, v))
}

// ValueLT applies the LT predicate on the "value" field.
func ValueLT(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLT(FieldValue, v))
}

// ValueLTE applies the LTE predicate on the "value" field.
func ValueLTE(v []byte) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLTE(FieldValue, v))
}

// ValueIsNil applies the IsNil predicate on the "value" field.
func ValueIsNil() predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIsNull(FieldValue))
}

// ValueNotNil applies the NotNil predicate on the "value" field.
func ValueNotNil() predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotNull(FieldValue))
}

// KindEQ applies the EQ predicate on the "kind" field.
func KindEQ(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldKind, v))
}

// KindNEQ applies the NEQ predicate on the "kind" field.
func KindNEQ(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldKind, v))
}

// KindIn applies the In predicate on the "kind" field.
func KindIn(vs ...uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIn(FieldKind, vs...))
}

// KindNotIn applies the NotIn predicate on the "kind" field.
func KindNotIn(vs ...uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldKind, vs...))
}

// KindGT applies the GT predicate on the "kind" field.
func KindGT(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGT(FieldKind, v))
}

// KindGTE applies the GTE predicate on the "kind" field.
func KindGTE(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGTE(FieldKind, v))
}

// KindLT applies the LT predicate on the "kind" field.
func KindLT(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLT(FieldKind, v))
}

// KindLTE applies the LTE predicate on the "kind" field.
func KindLTE(v uint) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLTE(FieldKind, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasExecution applies the HasEdge predicate on the "execution" edge.
func HasExecution() predicate.SignalExecutionData {
	return predicate.SignalExecutionData(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, ExecutionTable, ExecutionColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionWith applies the HasEdge predicate on the "execution" edge with a given conditions (other predicates).
func HasExecutionWith(preds ...predicate.SignalExecution) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(func(s *sql.Selector) {
		step := newExecutionStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SignalExecutionData) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SignalExecutionData) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SignalExecutionData) predicate.SignalExecutionData {
	return predicate.SignalExecutionData(sql.NotPredicates(p))
}
