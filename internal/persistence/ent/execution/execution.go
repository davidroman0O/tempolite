// Code generated by ent, DO NOT EDIT.

package execution

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the execution type in the database.
	Label = "execution"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldCompletedAt holds the string denoting the completed_at field in the database.
	FieldCompletedAt = "completed_at"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldError holds the string denoting the error field in the database.
	FieldError = "error"
	// EdgeEntity holds the string denoting the entity edge name in mutations.
	EdgeEntity = "entity"
	// EdgeWorkflowExecution holds the string denoting the workflow_execution edge name in mutations.
	EdgeWorkflowExecution = "workflow_execution"
	// EdgeActivityExecution holds the string denoting the activity_execution edge name in mutations.
	EdgeActivityExecution = "activity_execution"
	// EdgeSagaExecution holds the string denoting the saga_execution edge name in mutations.
	EdgeSagaExecution = "saga_execution"
	// EdgeSideEffectExecution holds the string denoting the side_effect_execution edge name in mutations.
	EdgeSideEffectExecution = "side_effect_execution"
	// Table holds the table name of the execution in the database.
	Table = "executions"
	// EntityTable is the table that holds the entity relation/edge.
	EntityTable = "executions"
	// EntityInverseTable is the table name for the Entity entity.
	// It exists in this package in order to avoid circular dependency with the "entity" package.
	EntityInverseTable = "entities"
	// EntityColumn is the table column denoting the entity relation/edge.
	EntityColumn = "entity_executions"
	// WorkflowExecutionTable is the table that holds the workflow_execution relation/edge.
	WorkflowExecutionTable = "workflow_executions"
	// WorkflowExecutionInverseTable is the table name for the WorkflowExecution entity.
	// It exists in this package in order to avoid circular dependency with the "workflowexecution" package.
	WorkflowExecutionInverseTable = "workflow_executions"
	// WorkflowExecutionColumn is the table column denoting the workflow_execution relation/edge.
	WorkflowExecutionColumn = "execution_workflow_execution"
	// ActivityExecutionTable is the table that holds the activity_execution relation/edge.
	ActivityExecutionTable = "activity_executions"
	// ActivityExecutionInverseTable is the table name for the ActivityExecution entity.
	// It exists in this package in order to avoid circular dependency with the "activityexecution" package.
	ActivityExecutionInverseTable = "activity_executions"
	// ActivityExecutionColumn is the table column denoting the activity_execution relation/edge.
	ActivityExecutionColumn = "execution_activity_execution"
	// SagaExecutionTable is the table that holds the saga_execution relation/edge.
	SagaExecutionTable = "saga_executions"
	// SagaExecutionInverseTable is the table name for the SagaExecution entity.
	// It exists in this package in order to avoid circular dependency with the "sagaexecution" package.
	SagaExecutionInverseTable = "saga_executions"
	// SagaExecutionColumn is the table column denoting the saga_execution relation/edge.
	SagaExecutionColumn = "execution_saga_execution"
	// SideEffectExecutionTable is the table that holds the side_effect_execution relation/edge.
	SideEffectExecutionTable = "side_effect_executions"
	// SideEffectExecutionInverseTable is the table name for the SideEffectExecution entity.
	// It exists in this package in order to avoid circular dependency with the "sideeffectexecution" package.
	SideEffectExecutionInverseTable = "side_effect_executions"
	// SideEffectExecutionColumn is the table column denoting the side_effect_execution relation/edge.
	SideEffectExecutionColumn = "execution_side_effect_execution"
)

// Columns holds all SQL columns for execution fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldStartedAt,
	FieldCompletedAt,
	FieldStatus,
	FieldError,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "executions"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"entity_executions",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
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
	// DefaultStartedAt holds the default value on creation for the "started_at" field.
	DefaultStartedAt func() time.Time
)

// Status defines the type for the "status" enum field.
type Status string

// StatusPending is the default value of the Status enum.
const DefaultStatus = StatusPending

// Status values.
const (
	StatusPending   Status = "Pending"
	StatusQueued    Status = "Queued"
	StatusRunning   Status = "Running"
	StatusRetried   Status = "Retried"
	StatusPaused    Status = "Paused"
	StatusCancelled Status = "Cancelled"
	StatusCompleted Status = "Completed"
	StatusFailed    Status = "Failed"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusPending, StatusQueued, StatusRunning, StatusRetried, StatusPaused, StatusCancelled, StatusCompleted, StatusFailed:
		return nil
	default:
		return fmt.Errorf("execution: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the Execution queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
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

// ByEntityField orders the results by entity field.
func ByEntityField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEntityStep(), sql.OrderByField(field, opts...))
	}
}

// ByWorkflowExecutionField orders the results by workflow_execution field.
func ByWorkflowExecutionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkflowExecutionStep(), sql.OrderByField(field, opts...))
	}
}

// ByActivityExecutionField orders the results by activity_execution field.
func ByActivityExecutionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newActivityExecutionStep(), sql.OrderByField(field, opts...))
	}
}

// BySagaExecutionField orders the results by saga_execution field.
func BySagaExecutionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSagaExecutionStep(), sql.OrderByField(field, opts...))
	}
}

// BySideEffectExecutionField orders the results by side_effect_execution field.
func BySideEffectExecutionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSideEffectExecutionStep(), sql.OrderByField(field, opts...))
	}
}
func newEntityStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EntityInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, EntityTable, EntityColumn),
	)
}
func newWorkflowExecutionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkflowExecutionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, WorkflowExecutionTable, WorkflowExecutionColumn),
	)
}
func newActivityExecutionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ActivityExecutionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ActivityExecutionTable, ActivityExecutionColumn),
	)
}
func newSagaExecutionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SagaExecutionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, SagaExecutionTable, SagaExecutionColumn),
	)
}
func newSideEffectExecutionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SideEffectExecutionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, SideEffectExecutionTable, SideEffectExecutionColumn),
	)
}
