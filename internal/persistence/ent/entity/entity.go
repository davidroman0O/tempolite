// Code generated by ent, DO NOT EDIT.

package entity

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the entity type in the database.
	Label = "entity"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldHandlerName holds the string denoting the handler_name field in the database.
	FieldHandlerName = "handler_name"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldStepID holds the string denoting the step_id field in the database.
	FieldStepID = "step_id"
	// EdgeRun holds the string denoting the run edge name in mutations.
	EdgeRun = "run"
	// EdgeExecutions holds the string denoting the executions edge name in mutations.
	EdgeExecutions = "executions"
	// EdgeQueue holds the string denoting the queue edge name in mutations.
	EdgeQueue = "queue"
	// EdgeVersions holds the string denoting the versions edge name in mutations.
	EdgeVersions = "versions"
	// EdgeWorkflowData holds the string denoting the workflow_data edge name in mutations.
	EdgeWorkflowData = "workflow_data"
	// EdgeActivityData holds the string denoting the activity_data edge name in mutations.
	EdgeActivityData = "activity_data"
	// EdgeSagaData holds the string denoting the saga_data edge name in mutations.
	EdgeSagaData = "saga_data"
	// EdgeSideEffectData holds the string denoting the side_effect_data edge name in mutations.
	EdgeSideEffectData = "side_effect_data"
	// Table holds the table name of the entity in the database.
	Table = "entities"
	// RunTable is the table that holds the run relation/edge.
	RunTable = "entities"
	// RunInverseTable is the table name for the Run entity.
	// It exists in this package in order to avoid circular dependency with the "run" package.
	RunInverseTable = "runs"
	// RunColumn is the table column denoting the run relation/edge.
	RunColumn = "run_entities"
	// ExecutionsTable is the table that holds the executions relation/edge.
	ExecutionsTable = "executions"
	// ExecutionsInverseTable is the table name for the Execution entity.
	// It exists in this package in order to avoid circular dependency with the "execution" package.
	ExecutionsInverseTable = "executions"
	// ExecutionsColumn is the table column denoting the executions relation/edge.
	ExecutionsColumn = "entity_executions"
	// QueueTable is the table that holds the queue relation/edge.
	QueueTable = "entities"
	// QueueInverseTable is the table name for the Queue entity.
	// It exists in this package in order to avoid circular dependency with the "queue" package.
	QueueInverseTable = "queues"
	// QueueColumn is the table column denoting the queue relation/edge.
	QueueColumn = "queue_entities"
	// VersionsTable is the table that holds the versions relation/edge.
	VersionsTable = "versions"
	// VersionsInverseTable is the table name for the Version entity.
	// It exists in this package in order to avoid circular dependency with the "version" package.
	VersionsInverseTable = "versions"
	// VersionsColumn is the table column denoting the versions relation/edge.
	VersionsColumn = "entity_versions"
	// WorkflowDataTable is the table that holds the workflow_data relation/edge.
	WorkflowDataTable = "workflow_data"
	// WorkflowDataInverseTable is the table name for the WorkflowData entity.
	// It exists in this package in order to avoid circular dependency with the "workflowdata" package.
	WorkflowDataInverseTable = "workflow_data"
	// WorkflowDataColumn is the table column denoting the workflow_data relation/edge.
	WorkflowDataColumn = "entity_workflow_data"
	// ActivityDataTable is the table that holds the activity_data relation/edge.
	ActivityDataTable = "activity_data"
	// ActivityDataInverseTable is the table name for the ActivityData entity.
	// It exists in this package in order to avoid circular dependency with the "activitydata" package.
	ActivityDataInverseTable = "activity_data"
	// ActivityDataColumn is the table column denoting the activity_data relation/edge.
	ActivityDataColumn = "entity_activity_data"
	// SagaDataTable is the table that holds the saga_data relation/edge.
	SagaDataTable = "saga_data"
	// SagaDataInverseTable is the table name for the SagaData entity.
	// It exists in this package in order to avoid circular dependency with the "sagadata" package.
	SagaDataInverseTable = "saga_data"
	// SagaDataColumn is the table column denoting the saga_data relation/edge.
	SagaDataColumn = "entity_saga_data"
	// SideEffectDataTable is the table that holds the side_effect_data relation/edge.
	SideEffectDataTable = "side_effect_data"
	// SideEffectDataInverseTable is the table name for the SideEffectData entity.
	// It exists in this package in order to avoid circular dependency with the "sideeffectdata" package.
	SideEffectDataInverseTable = "side_effect_data"
	// SideEffectDataColumn is the table column denoting the side_effect_data relation/edge.
	SideEffectDataColumn = "entity_side_effect_data"
)

// Columns holds all SQL columns for entity fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldHandlerName,
	FieldType,
	FieldStatus,
	FieldStepID,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "entities"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"queue_entities",
	"run_entities",
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
	// HandlerNameValidator is a validator for the "handler_name" field. It is called by the builders before save.
	HandlerNameValidator func(string) error
)

// Type defines the type for the "type" enum field.
type Type string

// Type values.
const (
	TypeWorkflow   Type = "Workflow"
	TypeActivity   Type = "Activity"
	TypeSaga       Type = "Saga"
	TypeSideEffect Type = "SideEffect"
)

func (_type Type) String() string {
	return string(_type)
}

// TypeValidator is a validator for the "type" field enum values. It is called by the builders before save.
func TypeValidator(_type Type) error {
	switch _type {
	case TypeWorkflow, TypeActivity, TypeSaga, TypeSideEffect:
		return nil
	default:
		return fmt.Errorf("entity: invalid enum value for type field: %q", _type)
	}
}

// Status defines the type for the "status" enum field.
type Status string

// StatusPending is the default value of the Status enum.
const DefaultStatus = StatusPending

// Status values.
const (
	StatusPending   Status = "Pending"
	StatusRunning   Status = "Running"
	StatusCompleted Status = "Completed"
	StatusFailed    Status = "Failed"
	StatusRetried   Status = "Retried"
	StatusCancelled Status = "Cancelled"
	StatusPaused    Status = "Paused"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusRetried, StatusCancelled, StatusPaused:
		return nil
	default:
		return fmt.Errorf("entity: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the Entity queries.
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

// ByHandlerName orders the results by the handler_name field.
func ByHandlerName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHandlerName, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByStepID orders the results by the step_id field.
func ByStepID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStepID, opts...).ToFunc()
}

// ByRunField orders the results by run field.
func ByRunField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRunStep(), sql.OrderByField(field, opts...))
	}
}

// ByExecutionsCount orders the results by executions count.
func ByExecutionsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newExecutionsStep(), opts...)
	}
}

// ByExecutions orders the results by executions terms.
func ByExecutions(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExecutionsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByQueueField orders the results by queue field.
func ByQueueField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newQueueStep(), sql.OrderByField(field, opts...))
	}
}

// ByVersionsCount orders the results by versions count.
func ByVersionsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newVersionsStep(), opts...)
	}
}

// ByVersions orders the results by versions terms.
func ByVersions(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newVersionsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByWorkflowDataField orders the results by workflow_data field.
func ByWorkflowDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkflowDataStep(), sql.OrderByField(field, opts...))
	}
}

// ByActivityDataField orders the results by activity_data field.
func ByActivityDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newActivityDataStep(), sql.OrderByField(field, opts...))
	}
}

// BySagaDataField orders the results by saga_data field.
func BySagaDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSagaDataStep(), sql.OrderByField(field, opts...))
	}
}

// BySideEffectDataField orders the results by side_effect_data field.
func BySideEffectDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSideEffectDataStep(), sql.OrderByField(field, opts...))
	}
}
func newRunStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RunInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, RunTable, RunColumn),
	)
}
func newExecutionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExecutionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ExecutionsTable, ExecutionsColumn),
	)
}
func newQueueStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(QueueInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, QueueTable, QueueColumn),
	)
}
func newVersionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(VersionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, VersionsTable, VersionsColumn),
	)
}
func newWorkflowDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkflowDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, WorkflowDataTable, WorkflowDataColumn),
	)
}
func newActivityDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ActivityDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ActivityDataTable, ActivityDataColumn),
	)
}
func newSagaDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SagaDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, SagaDataTable, SagaDataColumn),
	)
}
func newSideEffectDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SideEffectDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, SideEffectDataTable, SideEffectDataColumn),
	)
}