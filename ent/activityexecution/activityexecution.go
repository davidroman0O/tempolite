// Code generated by ent, DO NOT EDIT.

package activityexecution

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/davidroman0O/tempolite/ent/schema"
)

const (
	// Label holds the string label denoting the activityexecution type in the database.
	Label = "activity_execution"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldActivityEntityID holds the string denoting the activity_entity_id field in the database.
	FieldActivityEntityID = "activity_entity_id"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldCompletedAt holds the string denoting the completed_at field in the database.
	FieldCompletedAt = "completed_at"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldError holds the string denoting the error field in the database.
	FieldError = "error"
	// FieldStackTrace holds the string denoting the stack_trace field in the database.
	FieldStackTrace = "stack_trace"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeActivity holds the string denoting the activity edge name in mutations.
	EdgeActivity = "activity"
	// EdgeExecutionData holds the string denoting the execution_data edge name in mutations.
	EdgeExecutionData = "execution_data"
	// Table holds the table name of the activityexecution in the database.
	Table = "activity_executions"
	// ActivityTable is the table that holds the activity relation/edge.
	ActivityTable = "activity_executions"
	// ActivityInverseTable is the table name for the ActivityEntity entity.
	// It exists in this package in order to avoid circular dependency with the "activityentity" package.
	ActivityInverseTable = "activity_entities"
	// ActivityColumn is the table column denoting the activity relation/edge.
	ActivityColumn = "activity_entity_id"
	// ExecutionDataTable is the table that holds the execution_data relation/edge.
	ExecutionDataTable = "activity_execution_data"
	// ExecutionDataInverseTable is the table name for the ActivityExecutionData entity.
	// It exists in this package in order to avoid circular dependency with the "activityexecutiondata" package.
	ExecutionDataInverseTable = "activity_execution_data"
	// ExecutionDataColumn is the table column denoting the execution_data relation/edge.
	ExecutionDataColumn = "execution_id"
)

// Columns holds all SQL columns for activityexecution fields.
var Columns = []string{
	FieldID,
	FieldActivityEntityID,
	FieldStartedAt,
	FieldCompletedAt,
	FieldStatus,
	FieldError,
	FieldStackTrace,
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
	// DefaultStatus holds the default value on creation for the "status" field.
	DefaultStatus schema.ExecutionStatus
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
)

// OrderOption defines the ordering options for the ActivityExecution queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByActivityEntityID orders the results by the activity_entity_id field.
func ByActivityEntityID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActivityEntityID, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByCompletedAt orders the results by the completed_at field.
func ByCompletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCompletedAt, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByError orders the results by the error field.
func ByError(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldError, opts...).ToFunc()
}

// ByStackTrace orders the results by the stack_trace field.
func ByStackTrace(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStackTrace, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByActivityField orders the results by activity field.
func ByActivityField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newActivityStep(), sql.OrderByField(field, opts...))
	}
}

// ByExecutionDataField orders the results by execution_data field.
func ByExecutionDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExecutionDataStep(), sql.OrderByField(field, opts...))
	}
}
func newActivityStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ActivityInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, ActivityTable, ActivityColumn),
	)
}
func newExecutionDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExecutionDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ExecutionDataTable, ExecutionDataColumn),
	)
}
