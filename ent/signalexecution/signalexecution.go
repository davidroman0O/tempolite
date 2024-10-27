// Code generated by ent, DO NOT EDIT.

package signalexecution

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the signalexecution type in the database.
	Label = "signal_execution"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldRunID holds the string denoting the run_id field in the database.
	FieldRunID = "run_id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldQueueName holds the string denoting the queue_name field in the database.
	FieldQueueName = "queue_name"
	// FieldOutput holds the string denoting the output field in the database.
	FieldOutput = "output"
	// FieldError holds the string denoting the error field in the database.
	FieldError = "error"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeSignal holds the string denoting the signal edge name in mutations.
	EdgeSignal = "signal"
	// Table holds the table name of the signalexecution in the database.
	Table = "signal_executions"
	// SignalTable is the table that holds the signal relation/edge.
	SignalTable = "signal_executions"
	// SignalInverseTable is the table name for the Signal entity.
	// It exists in this package in order to avoid circular dependency with the "signal" package.
	SignalInverseTable = "signals"
	// SignalColumn is the table column denoting the signal relation/edge.
	SignalColumn = "signal_executions"
)

// Columns holds all SQL columns for signalexecution fields.
var Columns = []string{
	FieldID,
	FieldRunID,
	FieldStatus,
	FieldQueueName,
	FieldOutput,
	FieldError,
	FieldStartedAt,
	FieldUpdatedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "signal_executions"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"signal_executions",
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
	// DefaultQueueName holds the default value on creation for the "queue_name" field.
	DefaultQueueName string
	// QueueNameValidator is a validator for the "queue_name" field. It is called by the builders before save.
	QueueNameValidator func(string) error
	// DefaultStartedAt holds the default value on creation for the "started_at" field.
	DefaultStartedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
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
	StatusPaused    Status = "Paused"
	StatusRetried   Status = "Retried"
	StatusCancelled Status = "Cancelled"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusPaused, StatusRetried, StatusCancelled:
		return nil
	default:
		return fmt.Errorf("signalexecution: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the SignalExecution queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByRunID orders the results by the run_id field.
func ByRunID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRunID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByQueueName orders the results by the queue_name field.
func ByQueueName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldQueueName, opts...).ToFunc()
}

// ByError orders the results by the error field.
func ByError(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldError, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// BySignalField orders the results by signal field.
func BySignalField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSignalStep(), sql.OrderByField(field, opts...))
	}
}
func newSignalStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SignalInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, SignalTable, SignalColumn),
	)
}
