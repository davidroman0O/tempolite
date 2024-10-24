// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

// SideEffectExecution is the model entity for the SideEffectExecution schema.
type SideEffectExecution struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Status holds the value of the "status" field.
	Status sideeffectexecution.Status `json:"status,omitempty"`
	// Attempt holds the value of the "attempt" field.
	Attempt int `json:"attempt,omitempty"`
	// Output holds the value of the "output" field.
	Output []interface{} `json:"output,omitempty"`
	// Error holds the value of the "error" field.
	Error string `json:"error,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SideEffectExecutionQuery when eager-loading is set.
	Edges                  SideEffectExecutionEdges `json:"edges"`
	side_effect_executions *string
	selectValues           sql.SelectValues
}

// SideEffectExecutionEdges holds the relations/edges for other nodes in the graph.
type SideEffectExecutionEdges struct {
	// SideEffect holds the value of the side_effect edge.
	SideEffect *SideEffect `json:"side_effect,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// SideEffectOrErr returns the SideEffect value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SideEffectExecutionEdges) SideEffectOrErr() (*SideEffect, error) {
	if e.SideEffect != nil {
		return e.SideEffect, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: sideeffect.Label}
	}
	return nil, &NotLoadedError{edge: "side_effect"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SideEffectExecution) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case sideeffectexecution.FieldOutput:
			values[i] = new([]byte)
		case sideeffectexecution.FieldAttempt:
			values[i] = new(sql.NullInt64)
		case sideeffectexecution.FieldID, sideeffectexecution.FieldStatus, sideeffectexecution.FieldError:
			values[i] = new(sql.NullString)
		case sideeffectexecution.FieldStartedAt, sideeffectexecution.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case sideeffectexecution.ForeignKeys[0]: // side_effect_executions
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SideEffectExecution fields.
func (see *SideEffectExecution) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case sideeffectexecution.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				see.ID = value.String
			}
		case sideeffectexecution.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				see.Status = sideeffectexecution.Status(value.String)
			}
		case sideeffectexecution.FieldAttempt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field attempt", values[i])
			} else if value.Valid {
				see.Attempt = int(value.Int64)
			}
		case sideeffectexecution.FieldOutput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field output", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &see.Output); err != nil {
					return fmt.Errorf("unmarshal field output: %w", err)
				}
			}
		case sideeffectexecution.FieldError:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field error", values[i])
			} else if value.Valid {
				see.Error = value.String
			}
		case sideeffectexecution.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				see.StartedAt = value.Time
			}
		case sideeffectexecution.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				see.UpdatedAt = value.Time
			}
		case sideeffectexecution.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field side_effect_executions", values[i])
			} else if value.Valid {
				see.side_effect_executions = new(string)
				*see.side_effect_executions = value.String
			}
		default:
			see.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SideEffectExecution.
// This includes values selected through modifiers, order, etc.
func (see *SideEffectExecution) Value(name string) (ent.Value, error) {
	return see.selectValues.Get(name)
}

// QuerySideEffect queries the "side_effect" edge of the SideEffectExecution entity.
func (see *SideEffectExecution) QuerySideEffect() *SideEffectQuery {
	return NewSideEffectExecutionClient(see.config).QuerySideEffect(see)
}

// Update returns a builder for updating this SideEffectExecution.
// Note that you need to call SideEffectExecution.Unwrap() before calling this method if this SideEffectExecution
// was returned from a transaction, and the transaction was committed or rolled back.
func (see *SideEffectExecution) Update() *SideEffectExecutionUpdateOne {
	return NewSideEffectExecutionClient(see.config).UpdateOne(see)
}

// Unwrap unwraps the SideEffectExecution entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (see *SideEffectExecution) Unwrap() *SideEffectExecution {
	_tx, ok := see.config.driver.(*txDriver)
	if !ok {
		panic("ent: SideEffectExecution is not a transactional entity")
	}
	see.config.driver = _tx.drv
	return see
}

// String implements the fmt.Stringer.
func (see *SideEffectExecution) String() string {
	var builder strings.Builder
	builder.WriteString("SideEffectExecution(")
	builder.WriteString(fmt.Sprintf("id=%v, ", see.ID))
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", see.Status))
	builder.WriteString(", ")
	builder.WriteString("attempt=")
	builder.WriteString(fmt.Sprintf("%v", see.Attempt))
	builder.WriteString(", ")
	builder.WriteString("output=")
	builder.WriteString(fmt.Sprintf("%v", see.Output))
	builder.WriteString(", ")
	builder.WriteString("error=")
	builder.WriteString(see.Error)
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(see.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(see.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SideEffectExecutions is a parsable slice of SideEffectExecution.
type SideEffectExecutions []*SideEffectExecution
