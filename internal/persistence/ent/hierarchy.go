// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
)

// Hierarchy is the model entity for the Hierarchy schema.
type Hierarchy struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// RunID holds the value of the "run_id" field.
	RunID int `json:"run_id,omitempty"`
	// ParentEntityID holds the value of the "parent_entity_id" field.
	ParentEntityID int `json:"parent_entity_id,omitempty"`
	// ChildEntityID holds the value of the "child_entity_id" field.
	ChildEntityID int `json:"child_entity_id,omitempty"`
	// ParentExecutionID holds the value of the "parent_execution_id" field.
	ParentExecutionID int `json:"parent_execution_id,omitempty"`
	// ChildExecutionID holds the value of the "child_execution_id" field.
	ChildExecutionID int `json:"child_execution_id,omitempty"`
	// ParentStepID holds the value of the "parent_step_id" field.
	ParentStepID string `json:"parent_step_id,omitempty"`
	// ChildStepID holds the value of the "child_step_id" field.
	ChildStepID string `json:"child_step_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the HierarchyQuery when eager-loading is set.
	Edges        HierarchyEdges `json:"edges"`
	selectValues sql.SelectValues
}

// HierarchyEdges holds the relations/edges for other nodes in the graph.
type HierarchyEdges struct {
	// Run holds the value of the run edge.
	Run *Run `json:"run,omitempty"`
	// ParentEntity holds the value of the parent_entity edge.
	ParentEntity *Entity `json:"parent_entity,omitempty"`
	// ChildEntity holds the value of the child_entity edge.
	ChildEntity *Entity `json:"child_entity,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// RunOrErr returns the Run value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e HierarchyEdges) RunOrErr() (*Run, error) {
	if e.Run != nil {
		return e.Run, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: run.Label}
	}
	return nil, &NotLoadedError{edge: "run"}
}

// ParentEntityOrErr returns the ParentEntity value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e HierarchyEdges) ParentEntityOrErr() (*Entity, error) {
	if e.ParentEntity != nil {
		return e.ParentEntity, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: entity.Label}
	}
	return nil, &NotLoadedError{edge: "parent_entity"}
}

// ChildEntityOrErr returns the ChildEntity value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e HierarchyEdges) ChildEntityOrErr() (*Entity, error) {
	if e.ChildEntity != nil {
		return e.ChildEntity, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: entity.Label}
	}
	return nil, &NotLoadedError{edge: "child_entity"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Hierarchy) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case hierarchy.FieldID, hierarchy.FieldRunID, hierarchy.FieldParentEntityID, hierarchy.FieldChildEntityID, hierarchy.FieldParentExecutionID, hierarchy.FieldChildExecutionID:
			values[i] = new(sql.NullInt64)
		case hierarchy.FieldParentStepID, hierarchy.FieldChildStepID:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Hierarchy fields.
func (h *Hierarchy) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case hierarchy.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			h.ID = int(value.Int64)
		case hierarchy.FieldRunID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field run_id", values[i])
			} else if value.Valid {
				h.RunID = int(value.Int64)
			}
		case hierarchy.FieldParentEntityID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field parent_entity_id", values[i])
			} else if value.Valid {
				h.ParentEntityID = int(value.Int64)
			}
		case hierarchy.FieldChildEntityID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field child_entity_id", values[i])
			} else if value.Valid {
				h.ChildEntityID = int(value.Int64)
			}
		case hierarchy.FieldParentExecutionID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field parent_execution_id", values[i])
			} else if value.Valid {
				h.ParentExecutionID = int(value.Int64)
			}
		case hierarchy.FieldChildExecutionID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field child_execution_id", values[i])
			} else if value.Valid {
				h.ChildExecutionID = int(value.Int64)
			}
		case hierarchy.FieldParentStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field parent_step_id", values[i])
			} else if value.Valid {
				h.ParentStepID = value.String
			}
		case hierarchy.FieldChildStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field child_step_id", values[i])
			} else if value.Valid {
				h.ChildStepID = value.String
			}
		default:
			h.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Hierarchy.
// This includes values selected through modifiers, order, etc.
func (h *Hierarchy) Value(name string) (ent.Value, error) {
	return h.selectValues.Get(name)
}

// QueryRun queries the "run" edge of the Hierarchy entity.
func (h *Hierarchy) QueryRun() *RunQuery {
	return NewHierarchyClient(h.config).QueryRun(h)
}

// QueryParentEntity queries the "parent_entity" edge of the Hierarchy entity.
func (h *Hierarchy) QueryParentEntity() *EntityQuery {
	return NewHierarchyClient(h.config).QueryParentEntity(h)
}

// QueryChildEntity queries the "child_entity" edge of the Hierarchy entity.
func (h *Hierarchy) QueryChildEntity() *EntityQuery {
	return NewHierarchyClient(h.config).QueryChildEntity(h)
}

// Update returns a builder for updating this Hierarchy.
// Note that you need to call Hierarchy.Unwrap() before calling this method if this Hierarchy
// was returned from a transaction, and the transaction was committed or rolled back.
func (h *Hierarchy) Update() *HierarchyUpdateOne {
	return NewHierarchyClient(h.config).UpdateOne(h)
}

// Unwrap unwraps the Hierarchy entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (h *Hierarchy) Unwrap() *Hierarchy {
	_tx, ok := h.config.driver.(*txDriver)
	if !ok {
		panic("ent: Hierarchy is not a transactional entity")
	}
	h.config.driver = _tx.drv
	return h
}

// String implements the fmt.Stringer.
func (h *Hierarchy) String() string {
	var builder strings.Builder
	builder.WriteString("Hierarchy(")
	builder.WriteString(fmt.Sprintf("id=%v, ", h.ID))
	builder.WriteString("run_id=")
	builder.WriteString(fmt.Sprintf("%v", h.RunID))
	builder.WriteString(", ")
	builder.WriteString("parent_entity_id=")
	builder.WriteString(fmt.Sprintf("%v", h.ParentEntityID))
	builder.WriteString(", ")
	builder.WriteString("child_entity_id=")
	builder.WriteString(fmt.Sprintf("%v", h.ChildEntityID))
	builder.WriteString(", ")
	builder.WriteString("parent_execution_id=")
	builder.WriteString(fmt.Sprintf("%v", h.ParentExecutionID))
	builder.WriteString(", ")
	builder.WriteString("child_execution_id=")
	builder.WriteString(fmt.Sprintf("%v", h.ChildExecutionID))
	builder.WriteString(", ")
	builder.WriteString("parent_step_id=")
	builder.WriteString(h.ParentStepID)
	builder.WriteString(", ")
	builder.WriteString("child_step_id=")
	builder.WriteString(h.ChildStepID)
	builder.WriteByte(')')
	return builder.String()
}

// Hierarchies is a parsable slice of Hierarchy.
type Hierarchies []*Hierarchy
