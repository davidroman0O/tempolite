// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
)

// SideEffect is the model entity for the SideEffect schema.
type SideEffect struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Identity holds the value of the "identity" field.
	Identity string `json:"identity,omitempty"`
	// StepID holds the value of the "step_id" field.
	StepID string `json:"step_id,omitempty"`
	// HandlerName holds the value of the "handler_name" field.
	HandlerName string `json:"handler_name,omitempty"`
	// Status holds the value of the "status" field.
	Status sideeffect.Status `json:"status,omitempty"`
	// RetryPolicy holds the value of the "retry_policy" field.
	RetryPolicy schema.RetryPolicy `json:"retry_policy,omitempty"`
	// Timeout holds the value of the "timeout" field.
	Timeout time.Time `json:"timeout,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SideEffectQuery when eager-loading is set.
	Edges        SideEffectEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SideEffectEdges holds the relations/edges for other nodes in the graph.
type SideEffectEdges struct {
	// Executions holds the value of the executions edge.
	Executions []*SideEffectExecution `json:"executions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// ExecutionsOrErr returns the Executions value or an error if the edge
// was not loaded in eager-loading.
func (e SideEffectEdges) ExecutionsOrErr() ([]*SideEffectExecution, error) {
	if e.loadedTypes[0] {
		return e.Executions, nil
	}
	return nil, &NotLoadedError{edge: "executions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SideEffect) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case sideeffect.FieldRetryPolicy:
			values[i] = new([]byte)
		case sideeffect.FieldID, sideeffect.FieldIdentity, sideeffect.FieldStepID, sideeffect.FieldHandlerName, sideeffect.FieldStatus:
			values[i] = new(sql.NullString)
		case sideeffect.FieldTimeout, sideeffect.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SideEffect fields.
func (se *SideEffect) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case sideeffect.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				se.ID = value.String
			}
		case sideeffect.FieldIdentity:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field identity", values[i])
			} else if value.Valid {
				se.Identity = value.String
			}
		case sideeffect.FieldStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field step_id", values[i])
			} else if value.Valid {
				se.StepID = value.String
			}
		case sideeffect.FieldHandlerName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field handler_name", values[i])
			} else if value.Valid {
				se.HandlerName = value.String
			}
		case sideeffect.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				se.Status = sideeffect.Status(value.String)
			}
		case sideeffect.FieldRetryPolicy:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_policy", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &se.RetryPolicy); err != nil {
					return fmt.Errorf("unmarshal field retry_policy: %w", err)
				}
			}
		case sideeffect.FieldTimeout:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field timeout", values[i])
			} else if value.Valid {
				se.Timeout = value.Time
			}
		case sideeffect.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				se.CreatedAt = value.Time
			}
		default:
			se.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SideEffect.
// This includes values selected through modifiers, order, etc.
func (se *SideEffect) Value(name string) (ent.Value, error) {
	return se.selectValues.Get(name)
}

// QueryExecutions queries the "executions" edge of the SideEffect entity.
func (se *SideEffect) QueryExecutions() *SideEffectExecutionQuery {
	return NewSideEffectClient(se.config).QueryExecutions(se)
}

// Update returns a builder for updating this SideEffect.
// Note that you need to call SideEffect.Unwrap() before calling this method if this SideEffect
// was returned from a transaction, and the transaction was committed or rolled back.
func (se *SideEffect) Update() *SideEffectUpdateOne {
	return NewSideEffectClient(se.config).UpdateOne(se)
}

// Unwrap unwraps the SideEffect entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (se *SideEffect) Unwrap() *SideEffect {
	_tx, ok := se.config.driver.(*txDriver)
	if !ok {
		panic("ent: SideEffect is not a transactional entity")
	}
	se.config.driver = _tx.drv
	return se
}

// String implements the fmt.Stringer.
func (se *SideEffect) String() string {
	var builder strings.Builder
	builder.WriteString("SideEffect(")
	builder.WriteString(fmt.Sprintf("id=%v, ", se.ID))
	builder.WriteString("identity=")
	builder.WriteString(se.Identity)
	builder.WriteString(", ")
	builder.WriteString("step_id=")
	builder.WriteString(se.StepID)
	builder.WriteString(", ")
	builder.WriteString("handler_name=")
	builder.WriteString(se.HandlerName)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", se.Status))
	builder.WriteString(", ")
	builder.WriteString("retry_policy=")
	builder.WriteString(fmt.Sprintf("%v", se.RetryPolicy))
	builder.WriteString(", ")
	builder.WriteString("timeout=")
	builder.WriteString(se.Timeout.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(se.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SideEffects is a parsable slice of SideEffect.
type SideEffects []*SideEffect
