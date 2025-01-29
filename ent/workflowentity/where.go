// Code generated by ent, DO NOT EDIT.

package workflowentity

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/schema"
)

// ID filters vertices based on their ID field.
func ID(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id schema.WorkflowEntityID) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLTE(FieldID, id))
}

// HandlerName applies equality check predicate on the "handler_name" field. It's identical to HandlerNameEQ.
func HandlerName(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldHandlerName, v))
}

// Type applies equality check predicate on the "type" field. It's identical to TypeEQ.
func Type(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldType, vc))
}

// Status applies equality check predicate on the "status" field. It's identical to StatusEQ.
func Status(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldStatus, vc))
}

// StepID applies equality check predicate on the "step_id" field. It's identical to StepIDEQ.
func StepID(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldStepID, vc))
}

// RunID applies equality check predicate on the "run_id" field. It's identical to RunIDEQ.
func RunID(v schema.RunID) predicate.WorkflowEntity {
	vc := int(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldRunID, vc))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldUpdatedAt, v))
}

// HandlerNameEQ applies the EQ predicate on the "handler_name" field.
func HandlerNameEQ(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldHandlerName, v))
}

// HandlerNameNEQ applies the NEQ predicate on the "handler_name" field.
func HandlerNameNEQ(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldHandlerName, v))
}

// HandlerNameIn applies the In predicate on the "handler_name" field.
func HandlerNameIn(vs ...string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldIn(FieldHandlerName, vs...))
}

// HandlerNameNotIn applies the NotIn predicate on the "handler_name" field.
func HandlerNameNotIn(vs ...string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldHandlerName, vs...))
}

// HandlerNameGT applies the GT predicate on the "handler_name" field.
func HandlerNameGT(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGT(FieldHandlerName, v))
}

// HandlerNameGTE applies the GTE predicate on the "handler_name" field.
func HandlerNameGTE(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGTE(FieldHandlerName, v))
}

// HandlerNameLT applies the LT predicate on the "handler_name" field.
func HandlerNameLT(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLT(FieldHandlerName, v))
}

// HandlerNameLTE applies the LTE predicate on the "handler_name" field.
func HandlerNameLTE(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLTE(FieldHandlerName, v))
}

// HandlerNameContains applies the Contains predicate on the "handler_name" field.
func HandlerNameContains(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldContains(FieldHandlerName, v))
}

// HandlerNameHasPrefix applies the HasPrefix predicate on the "handler_name" field.
func HandlerNameHasPrefix(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldHasPrefix(FieldHandlerName, v))
}

// HandlerNameHasSuffix applies the HasSuffix predicate on the "handler_name" field.
func HandlerNameHasSuffix(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldHasSuffix(FieldHandlerName, v))
}

// HandlerNameEqualFold applies the EqualFold predicate on the "handler_name" field.
func HandlerNameEqualFold(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEqualFold(FieldHandlerName, v))
}

// HandlerNameContainsFold applies the ContainsFold predicate on the "handler_name" field.
func HandlerNameContainsFold(v string) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldContainsFold(FieldHandlerName, v))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldType, vc))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldType, vc))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...schema.EntityType) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldIn(FieldType, v...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...schema.EntityType) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldType, v...))
}

// TypeGT applies the GT predicate on the "type" field.
func TypeGT(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGT(FieldType, vc))
}

// TypeGTE applies the GTE predicate on the "type" field.
func TypeGTE(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGTE(FieldType, vc))
}

// TypeLT applies the LT predicate on the "type" field.
func TypeLT(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLT(FieldType, vc))
}

// TypeLTE applies the LTE predicate on the "type" field.
func TypeLTE(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLTE(FieldType, vc))
}

// TypeContains applies the Contains predicate on the "type" field.
func TypeContains(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContains(FieldType, vc))
}

// TypeHasPrefix applies the HasPrefix predicate on the "type" field.
func TypeHasPrefix(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasPrefix(FieldType, vc))
}

// TypeHasSuffix applies the HasSuffix predicate on the "type" field.
func TypeHasSuffix(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasSuffix(FieldType, vc))
}

// TypeEqualFold applies the EqualFold predicate on the "type" field.
func TypeEqualFold(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEqualFold(FieldType, vc))
}

// TypeContainsFold applies the ContainsFold predicate on the "type" field.
func TypeContainsFold(v schema.EntityType) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContainsFold(FieldType, vc))
}

// StatusEQ applies the EQ predicate on the "status" field.
func StatusEQ(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldStatus, vc))
}

// StatusNEQ applies the NEQ predicate on the "status" field.
func StatusNEQ(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldStatus, vc))
}

// StatusIn applies the In predicate on the "status" field.
func StatusIn(vs ...schema.EntityStatus) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldIn(FieldStatus, v...))
}

// StatusNotIn applies the NotIn predicate on the "status" field.
func StatusNotIn(vs ...schema.EntityStatus) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldStatus, v...))
}

// StatusGT applies the GT predicate on the "status" field.
func StatusGT(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGT(FieldStatus, vc))
}

// StatusGTE applies the GTE predicate on the "status" field.
func StatusGTE(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGTE(FieldStatus, vc))
}

// StatusLT applies the LT predicate on the "status" field.
func StatusLT(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLT(FieldStatus, vc))
}

// StatusLTE applies the LTE predicate on the "status" field.
func StatusLTE(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLTE(FieldStatus, vc))
}

// StatusContains applies the Contains predicate on the "status" field.
func StatusContains(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContains(FieldStatus, vc))
}

// StatusHasPrefix applies the HasPrefix predicate on the "status" field.
func StatusHasPrefix(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasPrefix(FieldStatus, vc))
}

// StatusHasSuffix applies the HasSuffix predicate on the "status" field.
func StatusHasSuffix(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasSuffix(FieldStatus, vc))
}

// StatusEqualFold applies the EqualFold predicate on the "status" field.
func StatusEqualFold(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEqualFold(FieldStatus, vc))
}

// StatusContainsFold applies the ContainsFold predicate on the "status" field.
func StatusContainsFold(v schema.EntityStatus) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContainsFold(FieldStatus, vc))
}

// StepIDEQ applies the EQ predicate on the "step_id" field.
func StepIDEQ(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldStepID, vc))
}

// StepIDNEQ applies the NEQ predicate on the "step_id" field.
func StepIDNEQ(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldStepID, vc))
}

// StepIDIn applies the In predicate on the "step_id" field.
func StepIDIn(vs ...schema.WorkflowStepID) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldIn(FieldStepID, v...))
}

// StepIDNotIn applies the NotIn predicate on the "step_id" field.
func StepIDNotIn(vs ...schema.WorkflowStepID) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldStepID, v...))
}

// StepIDGT applies the GT predicate on the "step_id" field.
func StepIDGT(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGT(FieldStepID, vc))
}

// StepIDGTE applies the GTE predicate on the "step_id" field.
func StepIDGTE(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldGTE(FieldStepID, vc))
}

// StepIDLT applies the LT predicate on the "step_id" field.
func StepIDLT(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLT(FieldStepID, vc))
}

// StepIDLTE applies the LTE predicate on the "step_id" field.
func StepIDLTE(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldLTE(FieldStepID, vc))
}

// StepIDContains applies the Contains predicate on the "step_id" field.
func StepIDContains(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContains(FieldStepID, vc))
}

// StepIDHasPrefix applies the HasPrefix predicate on the "step_id" field.
func StepIDHasPrefix(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasPrefix(FieldStepID, vc))
}

// StepIDHasSuffix applies the HasSuffix predicate on the "step_id" field.
func StepIDHasSuffix(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldHasSuffix(FieldStepID, vc))
}

// StepIDEqualFold applies the EqualFold predicate on the "step_id" field.
func StepIDEqualFold(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldEqualFold(FieldStepID, vc))
}

// StepIDContainsFold applies the ContainsFold predicate on the "step_id" field.
func StepIDContainsFold(v schema.WorkflowStepID) predicate.WorkflowEntity {
	vc := string(v)
	return predicate.WorkflowEntity(sql.FieldContainsFold(FieldStepID, vc))
}

// RunIDEQ applies the EQ predicate on the "run_id" field.
func RunIDEQ(v schema.RunID) predicate.WorkflowEntity {
	vc := int(v)
	return predicate.WorkflowEntity(sql.FieldEQ(FieldRunID, vc))
}

// RunIDNEQ applies the NEQ predicate on the "run_id" field.
func RunIDNEQ(v schema.RunID) predicate.WorkflowEntity {
	vc := int(v)
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldRunID, vc))
}

// RunIDIn applies the In predicate on the "run_id" field.
func RunIDIn(vs ...schema.RunID) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldIn(FieldRunID, v...))
}

// RunIDNotIn applies the NotIn predicate on the "run_id" field.
func RunIDNotIn(vs ...schema.RunID) predicate.WorkflowEntity {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = int(vs[i])
	}
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldRunID, v...))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasQueue applies the HasEdge predicate on the "queue" edge.
func HasQueue() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, QueueTable, QueueColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasQueueWith applies the HasEdge predicate on the "queue" edge with a given conditions (other predicates).
func HasQueueWith(preds ...predicate.Queue) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newQueueStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRun applies the HasEdge predicate on the "run" edge.
func HasRun() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, RunTable, RunColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRunWith applies the HasEdge predicate on the "run" edge with a given conditions (other predicates).
func HasRunWith(preds ...predicate.Run) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newRunStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasVersions applies the HasEdge predicate on the "versions" edge.
func HasVersions() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, VersionsTable, VersionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasVersionsWith applies the HasEdge predicate on the "versions" edge with a given conditions (other predicates).
func HasVersionsWith(preds ...predicate.Version) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newVersionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasWorkflowData applies the HasEdge predicate on the "workflow_data" edge.
func HasWorkflowData() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, WorkflowDataTable, WorkflowDataColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasWorkflowDataWith applies the HasEdge predicate on the "workflow_data" edge with a given conditions (other predicates).
func HasWorkflowDataWith(preds ...predicate.WorkflowData) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newWorkflowDataStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasActivityChildren applies the HasEdge predicate on the "activity_children" edge.
func HasActivityChildren() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, ActivityChildrenTable, ActivityChildrenColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasActivityChildrenWith applies the HasEdge predicate on the "activity_children" edge with a given conditions (other predicates).
func HasActivityChildrenWith(preds ...predicate.ActivityEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newActivityChildrenStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSagaChildren applies the HasEdge predicate on the "saga_children" edge.
func HasSagaChildren() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, SagaChildrenTable, SagaChildrenColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSagaChildrenWith applies the HasEdge predicate on the "saga_children" edge with a given conditions (other predicates).
func HasSagaChildrenWith(preds ...predicate.SagaEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newSagaChildrenStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSideEffectChildren applies the HasEdge predicate on the "side_effect_children" edge.
func HasSideEffectChildren() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, SideEffectChildrenTable, SideEffectChildrenColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSideEffectChildrenWith applies the HasEdge predicate on the "side_effect_children" edge with a given conditions (other predicates).
func HasSideEffectChildrenWith(preds ...predicate.SideEffectEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newSideEffectChildrenStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasExecutions applies the HasEdge predicate on the "executions" edge.
func HasExecutions() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, ExecutionsTable, ExecutionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasExecutionsWith applies the HasEdge predicate on the "executions" edge with a given conditions (other predicates).
func HasExecutionsWith(preds ...predicate.WorkflowExecution) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newExecutionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasEvents applies the HasEdge predicate on the "events" edge.
func HasEvents() predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, EventsTable, EventsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventsWith applies the HasEdge predicate on the "events" edge with a given conditions (other predicates).
func HasEventsWith(preds ...predicate.EventLog) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(func(s *sql.Selector) {
		step := newEventsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.WorkflowEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.WorkflowEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.WorkflowEntity) predicate.WorkflowEntity {
	return predicate.WorkflowEntity(sql.NotPredicates(p))
}
