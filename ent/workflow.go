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
	"github.com/davidroman0O/tempolite/ent/workflow"
)

// Workflow is the model entity for the Workflow schema.
type Workflow struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// StepID holds the value of the "step_id" field.
	StepID string `json:"step_id,omitempty"`
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
	// IsPaused holds the value of the "is_paused" field.
	IsPaused bool `json:"is_paused,omitempty"`
	// IsReady holds the value of the "is_ready" field.
	IsReady bool `json:"is_ready,omitempty"`
	// Timeout holds the value of the "timeout" field.
	Timeout time.Time `json:"timeout,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// ID of the workflow this one was continued from
	ContinuedFromID string `json:"continued_from_id,omitempty"`
	// ID of the workflow this one was retried from
	RetriedFromID *string `json:"retried_from_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the WorkflowQuery when eager-loading is set.
	Edges        WorkflowEdges `json:"edges"`
	selectValues sql.SelectValues
}

// WorkflowEdges holds the relations/edges for other nodes in the graph.
type WorkflowEdges struct {
	// Executions holds the value of the executions edge.
	Executions []*WorkflowExecution `json:"executions,omitempty"`
	// ContinuedFrom holds the value of the continued_from edge.
	ContinuedFrom *Workflow `json:"continued_from,omitempty"`
	// ContinuedTo holds the value of the continued_to edge.
	ContinuedTo *Workflow `json:"continued_to,omitempty"`
	// RetriedFrom holds the value of the retried_from edge.
	RetriedFrom *Workflow `json:"retried_from,omitempty"`
	// RetriedTo holds the value of the retried_to edge.
	RetriedTo []*Workflow `json:"retried_to,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [5]bool
}

// ExecutionsOrErr returns the Executions value or an error if the edge
// was not loaded in eager-loading.
func (e WorkflowEdges) ExecutionsOrErr() ([]*WorkflowExecution, error) {
	if e.loadedTypes[0] {
		return e.Executions, nil
	}
	return nil, &NotLoadedError{edge: "executions"}
}

// ContinuedFromOrErr returns the ContinuedFrom value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e WorkflowEdges) ContinuedFromOrErr() (*Workflow, error) {
	if e.ContinuedFrom != nil {
		return e.ContinuedFrom, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: workflow.Label}
	}
	return nil, &NotLoadedError{edge: "continued_from"}
}

// ContinuedToOrErr returns the ContinuedTo value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e WorkflowEdges) ContinuedToOrErr() (*Workflow, error) {
	if e.ContinuedTo != nil {
		return e.ContinuedTo, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: workflow.Label}
	}
	return nil, &NotLoadedError{edge: "continued_to"}
}

// RetriedFromOrErr returns the RetriedFrom value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e WorkflowEdges) RetriedFromOrErr() (*Workflow, error) {
	if e.RetriedFrom != nil {
		return e.RetriedFrom, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: workflow.Label}
	}
	return nil, &NotLoadedError{edge: "retried_from"}
}

// RetriedToOrErr returns the RetriedTo value or an error if the edge
// was not loaded in eager-loading.
func (e WorkflowEdges) RetriedToOrErr() ([]*Workflow, error) {
	if e.loadedTypes[4] {
		return e.RetriedTo, nil
	}
	return nil, &NotLoadedError{edge: "retried_to"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Workflow) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case workflow.FieldInput, workflow.FieldRetryPolicy:
			values[i] = new([]byte)
		case workflow.FieldIsPaused, workflow.FieldIsReady:
			values[i] = new(sql.NullBool)
		case workflow.FieldID, workflow.FieldStepID, workflow.FieldStatus, workflow.FieldIdentity, workflow.FieldHandlerName, workflow.FieldContinuedFromID, workflow.FieldRetriedFromID:
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
		case workflow.FieldStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field step_id", values[i])
			} else if value.Valid {
				w.StepID = value.String
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
		case workflow.FieldIsPaused:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_paused", values[i])
			} else if value.Valid {
				w.IsPaused = value.Bool
			}
		case workflow.FieldIsReady:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_ready", values[i])
			} else if value.Valid {
				w.IsReady = value.Bool
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
		case workflow.FieldContinuedFromID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field continued_from_id", values[i])
			} else if value.Valid {
				w.ContinuedFromID = value.String
			}
		case workflow.FieldRetriedFromID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field retried_from_id", values[i])
			} else if value.Valid {
				w.RetriedFromID = new(string)
				*w.RetriedFromID = value.String
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

// QueryContinuedFrom queries the "continued_from" edge of the Workflow entity.
func (w *Workflow) QueryContinuedFrom() *WorkflowQuery {
	return NewWorkflowClient(w.config).QueryContinuedFrom(w)
}

// QueryContinuedTo queries the "continued_to" edge of the Workflow entity.
func (w *Workflow) QueryContinuedTo() *WorkflowQuery {
	return NewWorkflowClient(w.config).QueryContinuedTo(w)
}

// QueryRetriedFrom queries the "retried_from" edge of the Workflow entity.
func (w *Workflow) QueryRetriedFrom() *WorkflowQuery {
	return NewWorkflowClient(w.config).QueryRetriedFrom(w)
}

// QueryRetriedTo queries the "retried_to" edge of the Workflow entity.
func (w *Workflow) QueryRetriedTo() *WorkflowQuery {
	return NewWorkflowClient(w.config).QueryRetriedTo(w)
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
	builder.WriteString("step_id=")
	builder.WriteString(w.StepID)
	builder.WriteString(", ")
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
	builder.WriteString("is_paused=")
	builder.WriteString(fmt.Sprintf("%v", w.IsPaused))
	builder.WriteString(", ")
	builder.WriteString("is_ready=")
	builder.WriteString(fmt.Sprintf("%v", w.IsReady))
	builder.WriteString(", ")
	builder.WriteString("timeout=")
	builder.WriteString(w.Timeout.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(w.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("continued_from_id=")
	builder.WriteString(w.ContinuedFromID)
	builder.WriteString(", ")
	if v := w.RetriedFromID; v != nil {
		builder.WriteString("retried_from_id=")
		builder.WriteString(*v)
	}
	builder.WriteByte(')')
	return builder.String()
}

// Workflows is a parsable slice of Workflow.
type Workflows []*Workflow
