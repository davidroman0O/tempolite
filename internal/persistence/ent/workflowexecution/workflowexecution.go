// Code generated by ent, DO NOT EDIT.

package workflowexecution

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the workflowexecution type in the database.
	Label = "workflow_execution"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// EdgeExecution holds the string denoting the execution edge name in mutations.
	EdgeExecution = "execution"
	// EdgeExecutionData holds the string denoting the execution_data edge name in mutations.
	EdgeExecutionData = "execution_data"
	// Table holds the table name of the workflowexecution in the database.
	Table = "workflow_executions"
	// ExecutionTable is the table that holds the execution relation/edge.
	ExecutionTable = "workflow_executions"
	// ExecutionInverseTable is the table name for the Execution entity.
	// It exists in this package in order to avoid circular dependency with the "execution" package.
	ExecutionInverseTable = "executions"
	// ExecutionColumn is the table column denoting the execution relation/edge.
	ExecutionColumn = "execution_workflow_execution"
	// ExecutionDataTable is the table that holds the execution_data relation/edge.
	ExecutionDataTable = "workflow_execution_data"
	// ExecutionDataInverseTable is the table name for the WorkflowExecutionData entity.
	// It exists in this package in order to avoid circular dependency with the "workflowexecutiondata" package.
	ExecutionDataInverseTable = "workflow_execution_data"
	// ExecutionDataColumn is the table column denoting the execution_data relation/edge.
	ExecutionDataColumn = "workflow_execution_execution_data"
)

// Columns holds all SQL columns for workflowexecution fields.
var Columns = []string{
	FieldID,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "workflow_executions"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"execution_workflow_execution",
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

// OrderOption defines the ordering options for the WorkflowExecution queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByExecutionField orders the results by execution field.
func ByExecutionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExecutionStep(), sql.OrderByField(field, opts...))
	}
}

// ByExecutionDataField orders the results by execution_data field.
func ByExecutionDataField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExecutionDataStep(), sql.OrderByField(field, opts...))
	}
}
func newExecutionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExecutionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, ExecutionTable, ExecutionColumn),
	)
}
func newExecutionDataStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExecutionDataInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ExecutionDataTable, ExecutionDataColumn),
	)
}
