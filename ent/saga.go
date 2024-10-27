// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/schema"
)

// Saga is the model entity for the Saga schema.
type Saga struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// RunID holds the value of the "run_id" field.
	RunID string `json:"run_id,omitempty"`
	// StepID holds the value of the "step_id" field.
	StepID string `json:"step_id,omitempty"`
	// Status holds the value of the "status" field.
	Status saga.Status `json:"status,omitempty"`
	// QueueName holds the value of the "queue_name" field.
	QueueName string `json:"queue_name,omitempty"`
	// Stores the full saga definition including transactions and compensations
	SagaDefinition schema.SagaDefinitionData `json:"saga_definition,omitempty"`
	// Error holds the value of the "error" field.
	Error string `json:"error,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SagaQuery when eager-loading is set.
	Edges        SagaEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SagaEdges holds the relations/edges for other nodes in the graph.
type SagaEdges struct {
	// Steps holds the value of the steps edge.
	Steps []*SagaExecution `json:"steps,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// StepsOrErr returns the Steps value or an error if the edge
// was not loaded in eager-loading.
func (e SagaEdges) StepsOrErr() ([]*SagaExecution, error) {
	if e.loadedTypes[0] {
		return e.Steps, nil
	}
	return nil, &NotLoadedError{edge: "steps"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Saga) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case saga.FieldSagaDefinition:
			values[i] = new([]byte)
		case saga.FieldID, saga.FieldRunID, saga.FieldStepID, saga.FieldStatus, saga.FieldQueueName, saga.FieldError:
			values[i] = new(sql.NullString)
		case saga.FieldCreatedAt, saga.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Saga fields.
func (s *Saga) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case saga.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				s.ID = value.String
			}
		case saga.FieldRunID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field run_id", values[i])
			} else if value.Valid {
				s.RunID = value.String
			}
		case saga.FieldStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field step_id", values[i])
			} else if value.Valid {
				s.StepID = value.String
			}
		case saga.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				s.Status = saga.Status(value.String)
			}
		case saga.FieldQueueName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field queue_name", values[i])
			} else if value.Valid {
				s.QueueName = value.String
			}
		case saga.FieldSagaDefinition:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field saga_definition", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &s.SagaDefinition); err != nil {
					return fmt.Errorf("unmarshal field saga_definition: %w", err)
				}
			}
		case saga.FieldError:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field error", values[i])
			} else if value.Valid {
				s.Error = value.String
			}
		case saga.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				s.CreatedAt = value.Time
			}
		case saga.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				s.UpdatedAt = value.Time
			}
		default:
			s.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Saga.
// This includes values selected through modifiers, order, etc.
func (s *Saga) Value(name string) (ent.Value, error) {
	return s.selectValues.Get(name)
}

// QuerySteps queries the "steps" edge of the Saga entity.
func (s *Saga) QuerySteps() *SagaExecutionQuery {
	return NewSagaClient(s.config).QuerySteps(s)
}

// Update returns a builder for updating this Saga.
// Note that you need to call Saga.Unwrap() before calling this method if this Saga
// was returned from a transaction, and the transaction was committed or rolled back.
func (s *Saga) Update() *SagaUpdateOne {
	return NewSagaClient(s.config).UpdateOne(s)
}

// Unwrap unwraps the Saga entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (s *Saga) Unwrap() *Saga {
	_tx, ok := s.config.driver.(*txDriver)
	if !ok {
		panic("ent: Saga is not a transactional entity")
	}
	s.config.driver = _tx.drv
	return s
}

// String implements the fmt.Stringer.
func (s *Saga) String() string {
	var builder strings.Builder
	builder.WriteString("Saga(")
	builder.WriteString(fmt.Sprintf("id=%v, ", s.ID))
	builder.WriteString("run_id=")
	builder.WriteString(s.RunID)
	builder.WriteString(", ")
	builder.WriteString("step_id=")
	builder.WriteString(s.StepID)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", s.Status))
	builder.WriteString(", ")
	builder.WriteString("queue_name=")
	builder.WriteString(s.QueueName)
	builder.WriteString(", ")
	builder.WriteString("saga_definition=")
	builder.WriteString(fmt.Sprintf("%v", s.SagaDefinition))
	builder.WriteString(", ")
	builder.WriteString("error=")
	builder.WriteString(s.Error)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(s.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(s.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Sagas is a parsable slice of Saga.
type Sagas []*Saga
