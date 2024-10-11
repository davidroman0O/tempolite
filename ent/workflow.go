// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/go-tempolite/ent/schema"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
)

// Workflow is the model entity for the Workflow schema.
type Workflow struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Status holds the value of the "status" field.
	Status workflow.Status `json:"status,omitempty"`
	// Identity holds the value of the "identity" field.
	Identity string `json:"identity,omitempty"`
	// HandlerName holds the value of the "handler_name" field.
	HandlerName string `json:"handler_name,omitempty"`
	// Input holds the value of the "input" field.
	Input []interface{} `json:"input,omitempty"`
	// RetryPolicy holds the value of the "retry_policy" field.
	RetryPolicy schema.RetryPolicy `json:"retry_policy,omitempty"`
	// Timeout holds the value of the "timeout" field.
	Timeout time.Time `json:"timeout,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the WorkflowQuery when eager-loading is set.
	Edges        WorkflowEdges `json:"edges"`
	selectValues sql.SelectValues
}

// WorkflowEdges holds the relations/edges for other nodes in the graph.
type WorkflowEdges struct {
	// Executions holds the value of the executions edge.
	Executions []*WorkflowExecution `json:"executions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// ExecutionsOrErr returns the Executions value or an error if the edge
// was not loaded in eager-loading.
func (e WorkflowEdges) ExecutionsOrErr() ([]*WorkflowExecution, error) {
	if e.loadedTypes[0] {
		return e.Executions, nil
	}
	return nil, &NotLoadedError{edge: "executions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Workflow) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case workflow.FieldInput, workflow.FieldRetryPolicy:
			values[i] = new([]byte)
		case workflow.FieldID, workflow.FieldStatus, workflow.FieldIdentity, workflow.FieldHandlerName:
			values[i] = new(sql.NullString)
		case workflow.FieldTimeout, workflow.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Workflow fields.
func (w *Workflow) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case workflow.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				w.ID = value.String
			}
		case workflow.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				w.Status = workflow.Status(value.String)
			}
		case workflow.FieldIdentity:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field identity", values[i])
			} else if value.Valid {
				w.Identity = value.String
			}
		case workflow.FieldHandlerName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field handler_name", values[i])
			} else if value.Valid {
				w.HandlerName = value.String
			}
		case workflow.FieldInput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field input", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &w.Input); err != nil {
					return fmt.Errorf("unmarshal field input: %w", err)
				}
			}
		case workflow.FieldRetryPolicy:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_policy", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &w.RetryPolicy); err != nil {
					return fmt.Errorf("unmarshal field retry_policy: %w", err)
				}
			}
		case workflow.FieldTimeout:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field timeout", values[i])
			} else if value.Valid {
				w.Timeout = value.Time
			}
		case workflow.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				w.CreatedAt = value.Time
			}
		default:
			w.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Workflow.
// This includes values selected through modifiers, order, etc.
func (w *Workflow) Value(name string) (ent.Value, error) {
	return w.selectValues.Get(name)
}

// QueryExecutions queries the "executions" edge of the Workflow entity.
func (w *Workflow) QueryExecutions() *WorkflowExecutionQuery {
	return NewWorkflowClient(w.config).QueryExecutions(w)
}

// Update returns a builder for updating this Workflow.
// Note that you need to call Workflow.Unwrap() before calling this method if this Workflow
// was returned from a transaction, and the transaction was committed or rolled back.
func (w *Workflow) Update() *WorkflowUpdateOne {
	return NewWorkflowClient(w.config).UpdateOne(w)
}

// Unwrap unwraps the Workflow entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (w *Workflow) Unwrap() *Workflow {
	_tx, ok := w.config.driver.(*txDriver)
	if !ok {
		panic("ent: Workflow is not a transactional entity")
	}
	w.config.driver = _tx.drv
	return w
}

// String implements the fmt.Stringer.
func (w *Workflow) String() string {
	var builder strings.Builder
	builder.WriteString("Workflow(")
	builder.WriteString(fmt.Sprintf("id=%v, ", w.ID))
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", w.Status))
	builder.WriteString(", ")
	builder.WriteString("identity=")
	builder.WriteString(w.Identity)
	builder.WriteString(", ")
	builder.WriteString("handler_name=")
	builder.WriteString(w.HandlerName)
	builder.WriteString(", ")
	builder.WriteString("input=")
	builder.WriteString(fmt.Sprintf("%v", w.Input))
	builder.WriteString(", ")
	builder.WriteString("retry_policy=")
	builder.WriteString(fmt.Sprintf("%v", w.RetryPolicy))
	builder.WriteString(", ")
	builder.WriteString("timeout=")
	builder.WriteString(w.Timeout.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(w.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Workflows is a parsable slice of Workflow.
type Workflows []*Workflow
