// Code generated by ent, DO NOT EDIT.

package sideeffectexecution

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.FieldLTE(FieldID, id))
}

// HasExecution applies the HasEdge predicate on the "execution" edge.
func HasExecution() predicate.SideEffectExecution {
	return predicate.SideEffectExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, ExecutionTable, ExecutionColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionWith applies the HasEdge predicate on the "execution" edge with a given conditions (other predicates).
func HasExecutionWith(preds ...predicate.Execution) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(func(s *sql.Selector) {
		step := newExecutionStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasExecutionData applies the HasEdge predicate on the "execution_data" edge.
func HasExecutionData() predicate.SideEffectExecution {
	return predicate.SideEffectExecution(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, ExecutionDataTable, ExecutionDataColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionDataWith applies the HasEdge predicate on the "execution_data" edge with a given conditions (other predicates).
func HasExecutionDataWith(preds ...predicate.SideEffectExecutionData) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(func(s *sql.Selector) {
		step := newExecutionDataStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SideEffectExecution) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SideEffectExecution) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SideEffectExecution) predicate.SideEffectExecution {
	return predicate.SideEffectExecution(sql.NotPredicates(p))
}
