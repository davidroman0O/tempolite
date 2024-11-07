// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/signal"
	"github.com/davidroman0O/tempolite/ent/signalexecution"
)

// SignalExecution is the model entity for the SignalExecution schema.
type SignalExecution struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// RunID holds the value of the "run_id" field.
	RunID string `json:"run_id,omitempty"`
	// Status holds the value of the "status" field.
	Status signalexecution.Status `json:"status,omitempty"`
	// QueueName holds the value of the "queue_name" field.
	QueueName string `json:"queue_name,omitempty"`
	// Output holds the value of the "output" field.
	Output [][]uint8 `json:"output,omitempty"`
	// Error holds the value of the "error" field.
	Error string `json:"error,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SignalExecutionQuery when eager-loading is set.
	Edges             SignalExecutionEdges `json:"edges"`
	signal_executions *string
	selectValues      sql.SelectValues
}

// SignalExecutionEdges holds the relations/edges for other nodes in the graph.
type SignalExecutionEdges struct {
	// Signal holds the value of the signal edge.
	Signal *Signal `json:"signal,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// SignalOrErr returns the Signal value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SignalExecutionEdges) SignalOrErr() (*Signal, error) {
	if e.Signal != nil {
		return e.Signal, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: signal.Label}
	}
	return nil, &NotLoadedError{edge: "signal"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SignalExecution) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case signalexecution.FieldOutput:
			values[i] = new([]byte)
		case signalexecution.FieldID, signalexecution.FieldRunID, signalexecution.FieldStatus, signalexecution.FieldQueueName, signalexecution.FieldError:
			values[i] = new(sql.NullString)
		case signalexecution.FieldStartedAt, signalexecution.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case signalexecution.ForeignKeys[0]: // signal_executions
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SignalExecution fields.
func (se *SignalExecution) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case signalexecution.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				se.ID = value.String
			}
		case signalexecution.FieldRunID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field run_id", values[i])
			} else if value.Valid {
				se.RunID = value.String
			}
		case signalexecution.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				se.Status = signalexecution.Status(value.String)
			}
		case signalexecution.FieldQueueName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field queue_name", values[i])
			} else if value.Valid {
				se.QueueName = value.String
			}
		case signalexecution.FieldOutput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field output", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &se.Output); err != nil {
					return fmt.Errorf("unmarshal field output: %w", err)
				}
			}
		case signalexecution.FieldError:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field error", values[i])
			} else if value.Valid {
				se.Error = value.String
			}
		case signalexecution.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				se.StartedAt = value.Time
			}
		case signalexecution.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				se.UpdatedAt = value.Time
			}
		case signalexecution.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field signal_executions", values[i])
			} else if value.Valid {
				se.signal_executions = new(string)
				*se.signal_executions = value.String
			}
		default:
			se.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SignalExecution.
// This includes values selected through modifiers, order, etc.
func (se *SignalExecution) Value(name string) (ent.Value, error) {
	return se.selectValues.Get(name)
}

// QuerySignal queries the "signal" edge of the SignalExecution entity.
func (se *SignalExecution) QuerySignal() *SignalQuery {
	return NewSignalExecutionClient(se.config).QuerySignal(se)
}

// Update returns a builder for updating this SignalExecution.
// Note that you need to call SignalExecution.Unwrap() before calling this method if this SignalExecution
// was returned from a transaction, and the transaction was committed or rolled back.
func (se *SignalExecution) Update() *SignalExecutionUpdateOne {
	return NewSignalExecutionClient(se.config).UpdateOne(se)
}

// Unwrap unwraps the SignalExecution entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (se *SignalExecution) Unwrap() *SignalExecution {
	_tx, ok := se.config.driver.(*txDriver)
	if !ok {
		panic("ent: SignalExecution is not a transactional entity")
	}
	se.config.driver = _tx.drv
	return se
}

// String implements the fmt.Stringer.
func (se *SignalExecution) String() string {
	var builder strings.Builder
	builder.WriteString("SignalExecution(")
	builder.WriteString(fmt.Sprintf("id=%v, ", se.ID))
	builder.WriteString("run_id=")
	builder.WriteString(se.RunID)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", se.Status))
	builder.WriteString(", ")
	builder.WriteString("queue_name=")
	builder.WriteString(se.QueueName)
	builder.WriteString(", ")
	builder.WriteString("output=")
	builder.WriteString(fmt.Sprintf("%v", se.Output))
	builder.WriteString(", ")
	builder.WriteString("error=")
	builder.WriteString(se.Error)
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(se.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(se.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SignalExecutions is a parsable slice of SignalExecution.
type SignalExecutions []*SignalExecution