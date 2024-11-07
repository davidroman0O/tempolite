// Code generated by ent, DO NOT EDIT.

package run

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the run type in the database.
	Label = "run"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// EdgeEntities holds the string denoting the entities edge name in mutations.
	EdgeEntities = "entities"
	// EdgeHierarchies holds the string denoting the hierarchies edge name in mutations.
	EdgeHierarchies = "hierarchies"
	// Table holds the table name of the run in the database.
	Table = "runs"
	// EntitiesTable is the table that holds the entities relation/edge.
	EntitiesTable = "entities"
	// EntitiesInverseTable is the table name for the Entity entity.
	// It exists in this package in order to avoid circular dependency with the "entity" package.
	EntitiesInverseTable = "entities"
	// EntitiesColumn is the table column denoting the entities relation/edge.
	EntitiesColumn = "run_entities"
	// HierarchiesTable is the table that holds the hierarchies relation/edge.
	HierarchiesTable = "hierarchies"
	// HierarchiesInverseTable is the table name for the Hierarchy entity.
	// It exists in this package in order to avoid circular dependency with the "hierarchy" package.
	HierarchiesInverseTable = "hierarchies"
	// HierarchiesColumn is the table column denoting the hierarchies relation/edge.
	HierarchiesColumn = "run_id"
)

// Columns holds all SQL columns for run fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldStatus,
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
	StatusCancelled Status = "Cancelled"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled:
		return nil
	default:
		return fmt.Errorf("run: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the Run queries.
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

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByEntitiesCount orders the results by entities count.
func ByEntitiesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEntitiesStep(), opts...)
	}
}

// ByEntities orders the results by entities terms.
func ByEntities(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEntitiesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByHierarchiesCount orders the results by hierarchies count.
func ByHierarchiesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newHierarchiesStep(), opts...)
	}
}

// ByHierarchies orders the results by hierarchies terms.
func ByHierarchies(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newHierarchiesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newEntitiesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EntitiesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, EntitiesTable, EntitiesColumn),
	)
}
func newHierarchiesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(HierarchiesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, HierarchiesTable, HierarchiesColumn),
	)
}
