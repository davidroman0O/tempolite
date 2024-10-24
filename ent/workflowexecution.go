// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

// WorkflowExecution is the model entity for the WorkflowExecution schema.
type WorkflowExecution struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// RunID holds the value of the "run_id" field.
	RunID string `json:"run_id,omitempty"`
	// Status holds the value of the "status" field.
	Status workflowexecution.Status `json:"status,omitempty"`
	// Output holds the value of the "output" field.
	Output []interface{} `json:"output,omitempty"`
	// Error holds the value of the "error" field.
	Error string `json:"error,omitempty"`
	// IsReplay holds the value of the "is_replay" field.
	IsReplay bool `json:"is_replay,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the WorkflowExecutionQuery when eager-loading is set.
	Edges               WorkflowExecutionEdges `json:"edges"`
	workflow_executions *string
	selectValues        sql.SelectValues
}

// WorkflowExecutionEdges holds the relations/edges for other nodes in the graph.
type WorkflowExecutionEdges struct {
	// Workflow holds the value of the workflow edge.
	Workflow *Workflow `json:"workflow,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// WorkflowOrErr returns the Workflow value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e WorkflowExecutionEdges) WorkflowOrErr() (*Workflow, error) {
	if e.Workflow != nil {
		return e.Workflow, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: workflow.Label}
	}
	return nil, &NotLoadedError{edge: "workflow"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*WorkflowExecution) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case workflowexecution.FieldOutput:
			values[i] = new([]byte)
		case workflowexecution.FieldIsReplay:
			values[i] = new(sql.NullBool)
		case workflowexecution.FieldID, workflowexecution.FieldRunID, workflowexecution.FieldStatus, workflowexecution.FieldError:
			values[i] = new(sql.NullString)
		case workflowexecution.FieldStartedAt, workflowexecution.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case workflowexecution.ForeignKeys[0]: // workflow_executions
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the WorkflowExecution fields.
func (we *WorkflowExecution) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case workflowexecution.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				we.ID = value.String
			}
		case workflowexecution.FieldRunID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field run_id", values[i])
			} else if value.Valid {
				we.RunID = value.String
			}
		case workflowexecution.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				we.Status = workflowexecution.Status(value.String)
			}
		case workflowexecution.FieldOutput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field output", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &we.Output); err != nil {
					return fmt.Errorf("unmarshal field output: %w", err)
				}
			}
		case workflowexecution.FieldError:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field error", values[i])
			} else if value.Valid {
				we.Error = value.String
			}
		case workflowexecution.FieldIsReplay:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_replay", values[i])
			} else if value.Valid {
				we.IsReplay = value.Bool
			}
		case workflowexecution.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				we.StartedAt = value.Time
			}
		case workflowexecution.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				we.UpdatedAt = value.Time
			}
		case workflowexecution.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field workflow_executions", values[i])
			} else if value.Valid {
				we.workflow_executions = new(string)
				*we.workflow_executions = value.String
			}
		default:
			we.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the WorkflowExecution.
// This includes values selected through modifiers, order, etc.
func (we *WorkflowExecution) Value(name string) (ent.Value, error) {
	return we.selectValues.Get(name)
}

// QueryWorkflow queries the "workflow" edge of the WorkflowExecution entity.
func (we *WorkflowExecution) QueryWorkflow() *WorkflowQuery {
	return NewWorkflowExecutionClient(we.config).QueryWorkflow(we)
}

// Update returns a builder for updating this WorkflowExecution.
// Note that you need to call WorkflowExecution.Unwrap() before calling this method if this WorkflowExecution
// was returned from a transaction, and the transaction was committed or rolled back.
func (we *WorkflowExecution) Update() *WorkflowExecutionUpdateOne {
	return NewWorkflowExecutionClient(we.config).UpdateOne(we)
}

// Unwrap unwraps the WorkflowExecution entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (we *WorkflowExecution) Unwrap() *WorkflowExecution {
	_tx, ok := we.config.driver.(*txDriver)
	if !ok {
		panic("ent: WorkflowExecution is not a transactional entity")
	}
	we.config.driver = _tx.drv
	return we
}

// String implements the fmt.Stringer.
func (we *WorkflowExecution) String() string {
	var builder strings.Builder
	builder.WriteString("WorkflowExecution(")
	builder.WriteString(fmt.Sprintf("id=%v, ", we.ID))
	builder.WriteString("run_id=")
	builder.WriteString(we.RunID)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", we.Status))
	builder.WriteString(", ")
	builder.WriteString("output=")
	builder.WriteString(fmt.Sprintf("%v", we.Output))
	builder.WriteString(", ")
	builder.WriteString("error=")
	builder.WriteString(we.Error)
	builder.WriteString(", ")
	builder.WriteString("is_replay=")
	builder.WriteString(fmt.Sprintf("%v", we.IsReplay))
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(we.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(we.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// WorkflowExecutions is a parsable slice of WorkflowExecution.
type WorkflowExecutions []*WorkflowExecution
