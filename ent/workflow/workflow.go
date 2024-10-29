// Code generated by ent, DO NOT EDIT.

package workflow

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the workflow type in the database.
	Label = "workflow"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldStepID holds the string denoting the step_id field in the database.
	FieldStepID = "step_id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldIdentity holds the string denoting the identity field in the database.
	FieldIdentity = "identity"
	// FieldHandlerName holds the string denoting the handler_name field in the database.
	FieldHandlerName = "handler_name"
	// FieldInput holds the string denoting the input field in the database.
	FieldInput = "input"
	// FieldQueueName holds the string denoting the queue_name field in the database.
	FieldQueueName = "queue_name"
	// FieldRetryPolicy holds the string denoting the retry_policy field in the database.
	FieldRetryPolicy = "retry_policy"
	// FieldIsPaused holds the string denoting the is_paused field in the database.
	FieldIsPaused = "is_paused"
	// FieldIsReady holds the string denoting the is_ready field in the database.
	FieldIsReady = "is_ready"
	// FieldMaxDuration holds the string denoting the max_duration field in the database.
	FieldMaxDuration = "max_duration"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldContinuedFromID holds the string denoting the continued_from_id field in the database.
	FieldContinuedFromID = "continued_from_id"
	// FieldRetriedFromID holds the string denoting the retried_from_id field in the database.
	FieldRetriedFromID = "retried_from_id"
	// EdgeExecutions holds the string denoting the executions edge name in mutations.
	EdgeExecutions = "executions"
	// EdgeContinuedFrom holds the string denoting the continued_from edge name in mutations.
	EdgeContinuedFrom = "continued_from"
	// EdgeContinuedTo holds the string denoting the continued_to edge name in mutations.
	EdgeContinuedTo = "continued_to"
	// EdgeRetriedFrom holds the string denoting the retried_from edge name in mutations.
	EdgeRetriedFrom = "retried_from"
	// EdgeRetriedTo holds the string denoting the retried_to edge name in mutations.
	EdgeRetriedTo = "retried_to"
	// Table holds the table name of the workflow in the database.
	Table = "workflows"
	// ExecutionsTable is the table that holds the executions relation/edge.
	ExecutionsTable = "workflow_executions"
	// ExecutionsInverseTable is the table name for the WorkflowExecution entity.
	// It exists in this package in order to avoid circular dependency with the "workflowexecution" package.
	ExecutionsInverseTable = "workflow_executions"
	// ExecutionsColumn is the table column denoting the executions relation/edge.
	ExecutionsColumn = "workflow_executions"
	// ContinuedFromTable is the table that holds the continued_from relation/edge.
	ContinuedFromTable = "workflows"
	// ContinuedFromColumn is the table column denoting the continued_from relation/edge.
	ContinuedFromColumn = "continued_from_id"
	// ContinuedToTable is the table that holds the continued_to relation/edge.
	ContinuedToTable = "workflows"
	// ContinuedToColumn is the table column denoting the continued_to relation/edge.
	ContinuedToColumn = "continued_from_id"
	// RetriedFromTable is the table that holds the retried_from relation/edge.
	RetriedFromTable = "workflows"
	// RetriedFromColumn is the table column denoting the retried_from relation/edge.
	RetriedFromColumn = "retried_from_id"
	// RetriedToTable is the table that holds the retried_to relation/edge.
	RetriedToTable = "workflows"
	// RetriedToColumn is the table column denoting the retried_to relation/edge.
	RetriedToColumn = "retried_from_id"
)

// Columns holds all SQL columns for workflow fields.
var Columns = []string{
	FieldID,
	FieldStepID,
	FieldStatus,
	FieldIdentity,
	FieldHandlerName,
	FieldInput,
	FieldQueueName,
	FieldRetryPolicy,
	FieldIsPaused,
	FieldIsReady,
	FieldMaxDuration,
	FieldCreatedAt,
	FieldContinuedFromID,
	FieldRetriedFromID,
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
	// StepIDValidator is a validator for the "step_id" field. It is called by the builders before save.
	StepIDValidator func(string) error
	// IdentityValidator is a validator for the "identity" field. It is called by the builders before save.
	IdentityValidator func(string) error
	// HandlerNameValidator is a validator for the "handler_name" field. It is called by the builders before save.
	HandlerNameValidator func(string) error
	// DefaultQueueName holds the default value on creation for the "queue_name" field.
	DefaultQueueName string
	// QueueNameValidator is a validator for the "queue_name" field. It is called by the builders before save.
	QueueNameValidator func(string) error
	// DefaultIsPaused holds the default value on creation for the "is_paused" field.
	DefaultIsPaused bool
	// DefaultIsReady holds the default value on creation for the "is_ready" field.
	DefaultIsReady bool
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
)

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
		return fmt.Errorf("workflow: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the Workflow queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByStepID orders the results by the step_id field.
func ByStepID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStepID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByIdentity orders the results by the identity field.
func ByIdentity(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIdentity, opts...).ToFunc()
}

// ByHandlerName orders the results by the handler_name field.
func ByHandlerName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHandlerName, opts...).ToFunc()
}

// ByQueueName orders the results by the queue_name field.
func ByQueueName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldQueueName, opts...).ToFunc()
}

// ByIsPaused orders the results by the is_paused field.
func ByIsPaused(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsPaused, opts...).ToFunc()
}

// ByIsReady orders the results by the is_ready field.
func ByIsReady(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsReady, opts...).ToFunc()
}

// ByMaxDuration orders the results by the max_duration field.
func ByMaxDuration(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMaxDuration, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByContinuedFromID orders the results by the continued_from_id field.
func ByContinuedFromID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldContinuedFromID, opts...).ToFunc()
}

// ByRetriedFromID orders the results by the retried_from_id field.
func ByRetriedFromID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRetriedFromID, opts...).ToFunc()
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

// ByContinuedFromField orders the results by continued_from field.
func ByContinuedFromField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newContinuedFromStep(), sql.OrderByField(field, opts...))
	}
}

// ByContinuedToField orders the results by continued_to field.
func ByContinuedToField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newContinuedToStep(), sql.OrderByField(field, opts...))
	}
}

// ByRetriedFromField orders the results by retried_from field.
func ByRetriedFromField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRetriedFromStep(), sql.OrderByField(field, opts...))
	}
}

// ByRetriedToCount orders the results by retried_to count.
func ByRetriedToCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRetriedToStep(), opts...)
	}
}

// ByRetriedTo orders the results by retried_to terms.
func ByRetriedTo(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRetriedToStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newExecutionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExecutionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ExecutionsTable, ExecutionsColumn),
	)
}
func newContinuedFromStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, ContinuedFromTable, ContinuedFromColumn),
	)
}
func newContinuedToStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ContinuedToTable, ContinuedToColumn),
	)
}
func newRetriedFromStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, RetriedFromTable, RetriedFromColumn),
	)
}
func newRetriedToStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, RetriedToTable, RetriedToColumn),
	)
}
