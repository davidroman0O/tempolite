// Code generated by ent, DO NOT EDIT.

package activityexecutiondata

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldLTE(FieldID, id))
}

// LastHeartbeat applies equality check predicate on the "last_heartbeat" field. It's identical to LastHeartbeatEQ.
func LastHeartbeat(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldEQ(FieldLastHeartbeat, v))
}

// HeartbeatsIsNil applies the IsNil predicate on the "heartbeats" field.
func HeartbeatsIsNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIsNull(FieldHeartbeats))
}

// HeartbeatsNotNil applies the NotNil predicate on the "heartbeats" field.
func HeartbeatsNotNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotNull(FieldHeartbeats))
}

// LastHeartbeatEQ applies the EQ predicate on the "last_heartbeat" field.
func LastHeartbeatEQ(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldEQ(FieldLastHeartbeat, v))
}

// LastHeartbeatNEQ applies the NEQ predicate on the "last_heartbeat" field.
func LastHeartbeatNEQ(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNEQ(FieldLastHeartbeat, v))
}

// LastHeartbeatIn applies the In predicate on the "last_heartbeat" field.
func LastHeartbeatIn(vs ...time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIn(FieldLastHeartbeat, vs...))
}

// LastHeartbeatNotIn applies the NotIn predicate on the "last_heartbeat" field.
func LastHeartbeatNotIn(vs ...time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotIn(FieldLastHeartbeat, vs...))
}

// LastHeartbeatGT applies the GT predicate on the "last_heartbeat" field.
func LastHeartbeatGT(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldGT(FieldLastHeartbeat, v))
}

// LastHeartbeatGTE applies the GTE predicate on the "last_heartbeat" field.
func LastHeartbeatGTE(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldGTE(FieldLastHeartbeat, v))
}

// LastHeartbeatLT applies the LT predicate on the "last_heartbeat" field.
func LastHeartbeatLT(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldLT(FieldLastHeartbeat, v))
}

// LastHeartbeatLTE applies the LTE predicate on the "last_heartbeat" field.
func LastHeartbeatLTE(v time.Time) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldLTE(FieldLastHeartbeat, v))
}

// LastHeartbeatIsNil applies the IsNil predicate on the "last_heartbeat" field.
func LastHeartbeatIsNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIsNull(FieldLastHeartbeat))
}

// LastHeartbeatNotNil applies the NotNil predicate on the "last_heartbeat" field.
func LastHeartbeatNotNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotNull(FieldLastHeartbeat))
}

// ProgressIsNil applies the IsNil predicate on the "progress" field.
func ProgressIsNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIsNull(FieldProgress))
}

// ProgressNotNil applies the NotNil predicate on the "progress" field.
func ProgressNotNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotNull(FieldProgress))
}

// ExecutionDetailsIsNil applies the IsNil predicate on the "execution_details" field.
func ExecutionDetailsIsNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldIsNull(FieldExecutionDetails))
}

// ExecutionDetailsNotNil applies the NotNil predicate on the "execution_details" field.
func ExecutionDetailsNotNil() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.FieldNotNull(FieldExecutionDetails))
}

// HasActivityExecution applies the HasEdge predicate on the "activity_execution" edge.
func HasActivityExecution() predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, ActivityExecutionTable, ActivityExecutionColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasActivityExecutionWith applies the HasEdge predicate on the "activity_execution" edge with a given conditions (other predicates).
func HasActivityExecutionWith(preds ...predicate.ActivityExecution) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(func(s *sql.Selector) {
		step := newActivityExecutionStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.ActivityExecutionData) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.ActivityExecutionData) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.ActivityExecutionData) predicate.ActivityExecutionData {
	return predicate.ActivityExecutionData(sql.NotPredicates(p))
}
