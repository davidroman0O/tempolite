// Code generated by ent, DO NOT EDIT.

package sagaexecutiondata

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldLTE(FieldID, id))
}

// LastTransaction applies equality check predicate on the "last_transaction" field. It's identical to LastTransactionEQ.
func LastTransaction(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldEQ(FieldLastTransaction, v))
}

// TransactionHistoryIsNil applies the IsNil predicate on the "transaction_history" field.
func TransactionHistoryIsNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldIsNull(FieldTransactionHistory))
}

// TransactionHistoryNotNil applies the NotNil predicate on the "transaction_history" field.
func TransactionHistoryNotNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNotNull(FieldTransactionHistory))
}

// CompensationHistoryIsNil applies the IsNil predicate on the "compensation_history" field.
func CompensationHistoryIsNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldIsNull(FieldCompensationHistory))
}

// CompensationHistoryNotNil applies the NotNil predicate on the "compensation_history" field.
func CompensationHistoryNotNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNotNull(FieldCompensationHistory))
}

// LastTransactionEQ applies the EQ predicate on the "last_transaction" field.
func LastTransactionEQ(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldEQ(FieldLastTransaction, v))
}

// LastTransactionNEQ applies the NEQ predicate on the "last_transaction" field.
func LastTransactionNEQ(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNEQ(FieldLastTransaction, v))
}

// LastTransactionIn applies the In predicate on the "last_transaction" field.
func LastTransactionIn(vs ...time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldIn(FieldLastTransaction, vs...))
}

// LastTransactionNotIn applies the NotIn predicate on the "last_transaction" field.
func LastTransactionNotIn(vs ...time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNotIn(FieldLastTransaction, vs...))
}

// LastTransactionGT applies the GT predicate on the "last_transaction" field.
func LastTransactionGT(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldGT(FieldLastTransaction, v))
}

// LastTransactionGTE applies the GTE predicate on the "last_transaction" field.
func LastTransactionGTE(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldGTE(FieldLastTransaction, v))
}

// LastTransactionLT applies the LT predicate on the "last_transaction" field.
func LastTransactionLT(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldLT(FieldLastTransaction, v))
}

// LastTransactionLTE applies the LTE predicate on the "last_transaction" field.
func LastTransactionLTE(v time.Time) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldLTE(FieldLastTransaction, v))
}

// LastTransactionIsNil applies the IsNil predicate on the "last_transaction" field.
func LastTransactionIsNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldIsNull(FieldLastTransaction))
}

// LastTransactionNotNil applies the NotNil predicate on the "last_transaction" field.
func LastTransactionNotNil() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.FieldNotNull(FieldLastTransaction))
}

// HasSagaExecution applies the HasEdge predicate on the "saga_execution" edge.
func HasSagaExecution() predicate.SagaExecutionData {
	return predicate.SagaExecutionData(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, SagaExecutionTable, SagaExecutionColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSagaExecutionWith applies the HasEdge predicate on the "saga_execution" edge with a given conditions (other predicates).
func HasSagaExecutionWith(preds ...predicate.SagaExecution) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(func(s *sql.Selector) {
		step := newSagaExecutionStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SagaExecutionData) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SagaExecutionData) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SagaExecutionData) predicate.SagaExecutionData {
	return predicate.SagaExecutionData(sql.NotPredicates(p))
}
