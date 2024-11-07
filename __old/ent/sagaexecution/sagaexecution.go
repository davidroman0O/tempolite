// Code generated by ent, DO NOT EDIT.

package sagaexecution

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the sagaexecution type in the database.
	Label = "saga_execution"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldHandlerName holds the string denoting the handler_name field in the database.
	FieldHandlerName = "handler_name"
	// FieldStepType holds the string denoting the step_type field in the database.
	FieldStepType = "step_type"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldQueueName holds the string denoting the queue_name field in the database.
	FieldQueueName = "queue_name"
	// FieldSequence holds the string denoting the sequence field in the database.
	FieldSequence = "sequence"
	// FieldError holds the string denoting the error field in the database.
	FieldError = "error"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldCompletedAt holds the string denoting the completed_at field in the database.
	FieldCompletedAt = "completed_at"
	// EdgeSaga holds the string denoting the saga edge name in mutations.
	EdgeSaga = "saga"
	// Table holds the table name of the sagaexecution in the database.
	Table = "saga_executions"
	// SagaTable is the table that holds the saga relation/edge.
	SagaTable = "saga_executions"
	// SagaInverseTable is the table name for the Saga entity.
	// It exists in this package in order to avoid circular dependency with the "saga" package.
	SagaInverseTable = "sagas"
	// SagaColumn is the table column denoting the saga relation/edge.
	SagaColumn = "saga_steps"
)

// Columns holds all SQL columns for sagaexecution fields.
var Columns = []string{
	FieldID,
	FieldHandlerName,
	FieldStepType,
	FieldStatus,
	FieldQueueName,
	FieldSequence,
	FieldError,
	FieldStartedAt,
	FieldCompletedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "saga_executions"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"saga_steps",
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
	// HandlerNameValidator is a validator for the "handler_name" field. It is called by the builders before save.
	HandlerNameValidator func(string) error
	// DefaultQueueName holds the default value on creation for the "queue_name" field.
	DefaultQueueName string
	// QueueNameValidator is a validator for the "queue_name" field. It is called by the builders before save.
	QueueNameValidator func(string) error
	// SequenceValidator is a validator for the "sequence" field. It is called by the builders before save.
	SequenceValidator func(int) error
	// DefaultStartedAt holds the default value on creation for the "started_at" field.
	DefaultStartedAt func() time.Time
)

// StepType defines the type for the "step_type" enum field.
type StepType string

// StepType values.
const (
	StepTypeTransaction  StepType = "Transaction"
	StepTypeCompensation StepType = "Compensation"
)

func (st StepType) String() string {
	return string(st)
}

// StepTypeValidator is a validator for the "step_type" field enum values. It is called by the builders before save.
func StepTypeValidator(st StepType) error {
	switch st {
	case StepTypeTransaction, StepTypeCompensation:
		return nil
	default:
		return fmt.Errorf("sagaexecution: invalid enum value for step_type field: %q", st)
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
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed:
		return nil
	default:
		return fmt.Errorf("sagaexecution: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the SagaExecution queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByHandlerName orders the results by the handler_name field.
func ByHandlerName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHandlerName, opts...).ToFunc()
}

// ByStepType orders the results by the step_type field.
func ByStepType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStepType, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByQueueName orders the results by the queue_name field.
func ByQueueName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldQueueName, opts...).ToFunc()
}

// BySequence orders the results by the sequence field.
func BySequence(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSequence, opts...).ToFunc()
}

// ByError orders the results by the error field.
func ByError(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldError, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByCompletedAt orders the results by the completed_at field.
func ByCompletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCompletedAt, opts...).ToFunc()
}

// BySagaField orders the results by saga field.
func BySagaField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSagaStep(), sql.OrderByField(field, opts...))
	}
}
func newSagaStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SagaInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, SagaTable, SagaColumn),
	)
}