// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/sideeffectentity"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecutiondata"
)

// SideEffectExecution is the model entity for the SideEffectExecution schema.
type SideEffectExecution struct {
	config `json:"-"`
	// ID of the ent.
	ID schema.SideEffectExecutionID `json:"id,omitempty"`
	// SideEffectEntityID holds the value of the "side_effect_entity_id" field.
	SideEffectEntityID schema.SideEffectEntityID `json:"side_effect_entity_id,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// CompletedAt holds the value of the "completed_at" field.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Status holds the value of the "status" field.
	Status schema.ExecutionStatus `json:"status,omitempty"`
	// Error holds the value of the "error" field.
	Error string `json:"error,omitempty"`
	// StackTrace holds the value of the "stack_trace" field.
	StackTrace string `json:"stack_trace,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SideEffectExecutionQuery when eager-loading is set.
	Edges        SideEffectExecutionEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SideEffectExecutionEdges holds the relations/edges for other nodes in the graph.
type SideEffectExecutionEdges struct {
	// SideEffect holds the value of the side_effect edge.
	SideEffect *SideEffectEntity `json:"side_effect,omitempty"`
	// ExecutionData holds the value of the execution_data edge.
	ExecutionData *SideEffectExecutionData `json:"execution_data,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// SideEffectOrErr returns the SideEffect value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SideEffectExecutionEdges) SideEffectOrErr() (*SideEffectEntity, error) {
	if e.SideEffect != nil {
		return e.SideEffect, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: sideeffectentity.Label}
	}
	return nil, &NotLoadedError{edge: "side_effect"}
}

// ExecutionDataOrErr returns the ExecutionData value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SideEffectExecutionEdges) ExecutionDataOrErr() (*SideEffectExecutionData, error) {
	if e.ExecutionData != nil {
		return e.ExecutionData, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: sideeffectexecutiondata.Label}
	}
	return nil, &NotLoadedError{edge: "execution_data"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SideEffectExecution) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case sideeffectexecution.FieldID, sideeffectexecution.FieldSideEffectEntityID:
			values[i] = new(sql.NullInt64)
		case sideeffectexecution.FieldStatus, sideeffectexecution.FieldError, sideeffectexecution.FieldStackTrace:
			values[i] = new(sql.NullString)
		case sideeffectexecution.FieldStartedAt, sideeffectexecution.FieldCompletedAt, sideeffectexecution.FieldCreatedAt, sideeffectexecution.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
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
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				see.ID = schema.SideEffectExecutionID(value.Int64)
			}
		case sideeffectexecution.FieldSideEffectEntityID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field side_effect_entity_id", values[i])
			} else if value.Valid {
				see.SideEffectEntityID = schema.SideEffectEntityID(value.Int64)
			}
		case sideeffectexecution.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				see.StartedAt = value.Time
			}
		case sideeffectexecution.FieldCompletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field completed_at", values[i])
			} else if value.Valid {
				see.CompletedAt = new(time.Time)
				*see.CompletedAt = value.Time
			}
		case sideeffectexecution.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				see.Status = schema.ExecutionStatus(value.String)
			}
		case sideeffectexecution.FieldError:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field error", values[i])
			} else if value.Valid {
				see.Error = value.String
			}
		case sideeffectexecution.FieldStackTrace:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field stack_trace", values[i])
			} else if value.Valid {
				see.StackTrace = value.String
			}
		case sideeffectexecution.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				see.CreatedAt = value.Time
			}
		case sideeffectexecution.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				see.UpdatedAt = value.Time
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
func (see *SideEffectExecution) QuerySideEffect() *SideEffectEntityQuery {
	return NewSideEffectExecutionClient(see.config).QuerySideEffect(see)
}

// QueryExecutionData queries the "execution_data" edge of the SideEffectExecution entity.
func (see *SideEffectExecution) QueryExecutionData() *SideEffectExecutionDataQuery {
	return NewSideEffectExecutionClient(see.config).QueryExecutionData(see)
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
	builder.WriteString("side_effect_entity_id=")
	builder.WriteString(fmt.Sprintf("%v", see.SideEffectEntityID))
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(see.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	if v := see.CompletedAt; v != nil {
		builder.WriteString("completed_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", see.Status))
	builder.WriteString(", ")
	builder.WriteString("error=")
	builder.WriteString(see.Error)
	builder.WriteString(", ")
	builder.WriteString("stack_trace=")
	builder.WriteString(see.StackTrace)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(see.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(see.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SideEffectExecutions is a parsable slice of SideEffectExecution.
type SideEffectExecutions []*SideEffectExecution
