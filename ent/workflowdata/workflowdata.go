// Code generated by ent, DO NOT EDIT.

package workflowdata

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the workflowdata type in the database.
	Label = "workflow_data"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldEntityID holds the string denoting the entity_id field in the database.
	FieldEntityID = "entity_id"
	// FieldDuration holds the string denoting the duration field in the database.
	FieldDuration = "duration"
	// FieldPaused holds the string denoting the paused field in the database.
	FieldPaused = "paused"
	// FieldResumable holds the string denoting the resumable field in the database.
	FieldResumable = "resumable"
	// FieldIsRoot holds the string denoting the is_root field in the database.
	FieldIsRoot = "is_root"
	// FieldInputs holds the string denoting the inputs field in the database.
	FieldInputs = "inputs"
	// FieldContinuedFrom holds the string denoting the continued_from field in the database.
	FieldContinuedFrom = "continued_from"
	// FieldContinuedExecutionFrom holds the string denoting the continued_execution_from field in the database.
	FieldContinuedExecutionFrom = "continued_execution_from"
	// FieldWorkflowStepID holds the string denoting the workflow_step_id field in the database.
	FieldWorkflowStepID = "workflow_step_id"
	// FieldWorkflowFrom holds the string denoting the workflow_from field in the database.
	FieldWorkflowFrom = "workflow_from"
	// FieldWorkflowExecutionFrom holds the string denoting the workflow_execution_from field in the database.
	FieldWorkflowExecutionFrom = "workflow_execution_from"
	// FieldVersions holds the string denoting the versions field in the database.
	FieldVersions = "versions"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeWorkflow holds the string denoting the workflow edge name in mutations.
	EdgeWorkflow = "workflow"
	// Table holds the table name of the workflowdata in the database.
	Table = "workflow_data"
	// WorkflowTable is the table that holds the workflow relation/edge.
	WorkflowTable = "workflow_data"
	// WorkflowInverseTable is the table name for the WorkflowEntity entity.
	// It exists in this package in order to avoid circular dependency with the "workflowentity" package.
	WorkflowInverseTable = "workflow_entities"
	// WorkflowColumn is the table column denoting the workflow relation/edge.
	WorkflowColumn = "entity_id"
)

// Columns holds all SQL columns for workflowdata fields.
var Columns = []string{
	FieldID,
	FieldEntityID,
	FieldDuration,
	FieldPaused,
	FieldResumable,
	FieldIsRoot,
	FieldInputs,
	FieldContinuedFrom,
	FieldContinuedExecutionFrom,
	FieldWorkflowStepID,
	FieldWorkflowFrom,
	FieldWorkflowExecutionFrom,
	FieldVersions,
	FieldCreatedAt,
	FieldUpdatedAt,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
)

// OrderOption defines the ordering options for the WorkflowData queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByEntityID orders the results by the entity_id field.
func ByEntityID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEntityID, opts...).ToFunc()
}

// ByDuration orders the results by the duration field.
func ByDuration(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDuration, opts...).ToFunc()
}

// ByPaused orders the results by the paused field.
func ByPaused(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPaused, opts...).ToFunc()
}

// ByResumable orders the results by the resumable field.
func ByResumable(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldResumable, opts...).ToFunc()
}

// ByIsRoot orders the results by the is_root field.
func ByIsRoot(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsRoot, opts...).ToFunc()
}

// ByContinuedFrom orders the results by the continued_from field.
func ByContinuedFrom(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldContinuedFrom, opts...).ToFunc()
}

// ByContinuedExecutionFrom orders the results by the continued_execution_from field.
func ByContinuedExecutionFrom(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldContinuedExecutionFrom, opts...).ToFunc()
}

// ByWorkflowStepID orders the results by the workflow_step_id field.
func ByWorkflowStepID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWorkflowStepID, opts...).ToFunc()
}

// ByWorkflowFrom orders the results by the workflow_from field.
func ByWorkflowFrom(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWorkflowFrom, opts...).ToFunc()
}

// ByWorkflowExecutionFrom orders the results by the workflow_execution_from field.
func ByWorkflowExecutionFrom(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWorkflowExecutionFrom, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByWorkflowField orders the results by workflow field.
func ByWorkflowField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkflowStep(), sql.OrderByField(field, opts...))
	}
}
func newWorkflowStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkflowInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, WorkflowTable, WorkflowColumn),
	)
}
