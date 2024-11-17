// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowdata"
)

// WorkflowData is the model entity for the WorkflowData schema.
type WorkflowData struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Duration holds the value of the "duration" field.
	Duration string `json:"duration,omitempty"`
	// Paused holds the value of the "paused" field.
	Paused bool `json:"paused,omitempty"`
	// Resumable holds the value of the "resumable" field.
	Resumable bool `json:"resumable,omitempty"`
	// RetryState holds the value of the "retry_state" field.
	RetryState *schema.RetryState `json:"retry_state,omitempty"`
	// RetryPolicy holds the value of the "retry_policy" field.
	RetryPolicy *schema.RetryPolicy `json:"retry_policy,omitempty"`
	// Input holds the value of the "input" field.
	Input [][]uint8 `json:"input,omitempty"`
	// Attempt holds the value of the "attempt" field.
	Attempt int `json:"attempt,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the WorkflowDataQuery when eager-loading is set.
	Edges                WorkflowDataEdges `json:"edges"`
	entity_workflow_data *int
	selectValues         sql.SelectValues
}

// WorkflowDataEdges holds the relations/edges for other nodes in the graph.
type WorkflowDataEdges struct {
	// Entity holds the value of the entity edge.
	Entity *Entity `json:"entity,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// EntityOrErr returns the Entity value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e WorkflowDataEdges) EntityOrErr() (*Entity, error) {
	if e.Entity != nil {
		return e.Entity, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: entity.Label}
	}
	return nil, &NotLoadedError{edge: "entity"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*WorkflowData) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case workflowdata.FieldRetryState, workflowdata.FieldRetryPolicy, workflowdata.FieldInput:
			values[i] = new([]byte)
		case workflowdata.FieldPaused, workflowdata.FieldResumable:
			values[i] = new(sql.NullBool)
		case workflowdata.FieldID, workflowdata.FieldAttempt:
			values[i] = new(sql.NullInt64)
		case workflowdata.FieldDuration:
			values[i] = new(sql.NullString)
		case workflowdata.ForeignKeys[0]: // entity_workflow_data
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the WorkflowData fields.
func (wd *WorkflowData) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case workflowdata.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			wd.ID = int(value.Int64)
		case workflowdata.FieldDuration:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field duration", values[i])
			} else if value.Valid {
				wd.Duration = value.String
			}
		case workflowdata.FieldPaused:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field paused", values[i])
			} else if value.Valid {
				wd.Paused = value.Bool
			}
		case workflowdata.FieldResumable:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field resumable", values[i])
			} else if value.Valid {
				wd.Resumable = value.Bool
			}
		case workflowdata.FieldRetryState:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_state", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &wd.RetryState); err != nil {
					return fmt.Errorf("unmarshal field retry_state: %w", err)
				}
			}
		case workflowdata.FieldRetryPolicy:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_policy", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &wd.RetryPolicy); err != nil {
					return fmt.Errorf("unmarshal field retry_policy: %w", err)
				}
			}
		case workflowdata.FieldInput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field input", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &wd.Input); err != nil {
					return fmt.Errorf("unmarshal field input: %w", err)
				}
			}
		case workflowdata.FieldAttempt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field attempt", values[i])
			} else if value.Valid {
				wd.Attempt = int(value.Int64)
			}
		case workflowdata.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field entity_workflow_data", value)
			} else if value.Valid {
				wd.entity_workflow_data = new(int)
				*wd.entity_workflow_data = int(value.Int64)
			}
		default:
			wd.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the WorkflowData.
// This includes values selected through modifiers, order, etc.
func (wd *WorkflowData) Value(name string) (ent.Value, error) {
	return wd.selectValues.Get(name)
}

// QueryEntity queries the "entity" edge of the WorkflowData entity.
func (wd *WorkflowData) QueryEntity() *EntityQuery {
	return NewWorkflowDataClient(wd.config).QueryEntity(wd)
}

// Update returns a builder for updating this WorkflowData.
// Note that you need to call WorkflowData.Unwrap() before calling this method if this WorkflowData
// was returned from a transaction, and the transaction was committed or rolled back.
func (wd *WorkflowData) Update() *WorkflowDataUpdateOne {
	return NewWorkflowDataClient(wd.config).UpdateOne(wd)
}

// Unwrap unwraps the WorkflowData entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (wd *WorkflowData) Unwrap() *WorkflowData {
	_tx, ok := wd.config.driver.(*txDriver)
	if !ok {
		panic("ent: WorkflowData is not a transactional entity")
	}
	wd.config.driver = _tx.drv
	return wd
}

// String implements the fmt.Stringer.
func (wd *WorkflowData) String() string {
	var builder strings.Builder
	builder.WriteString("WorkflowData(")
	builder.WriteString(fmt.Sprintf("id=%v, ", wd.ID))
	builder.WriteString("duration=")
	builder.WriteString(wd.Duration)
	builder.WriteString(", ")
	builder.WriteString("paused=")
	builder.WriteString(fmt.Sprintf("%v", wd.Paused))
	builder.WriteString(", ")
	builder.WriteString("resumable=")
	builder.WriteString(fmt.Sprintf("%v", wd.Resumable))
	builder.WriteString(", ")
	builder.WriteString("retry_state=")
	builder.WriteString(fmt.Sprintf("%v", wd.RetryState))
	builder.WriteString(", ")
	builder.WriteString("retry_policy=")
	builder.WriteString(fmt.Sprintf("%v", wd.RetryPolicy))
	builder.WriteString(", ")
	builder.WriteString("input=")
	builder.WriteString(fmt.Sprintf("%v", wd.Input))
	builder.WriteString(", ")
	builder.WriteString("attempt=")
	builder.WriteString(fmt.Sprintf("%v", wd.Attempt))
	builder.WriteByte(')')
	return builder.String()
}

// WorkflowDataSlice is a parsable slice of WorkflowData.
type WorkflowDataSlice []*WorkflowData
