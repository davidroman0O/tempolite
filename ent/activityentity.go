// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/ent/activitydata"
	"github.com/davidroman0O/tempolite/ent/activityentity"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
)

// ActivityEntity is the model entity for the ActivityEntity schema.
type ActivityEntity struct {
	config `json:"-"`
	// ID of the ent.
	ID schema.ActivityEntityID `json:"id,omitempty"`
	// HandlerName holds the value of the "handler_name" field.
	HandlerName string `json:"handler_name,omitempty"`
	// Type holds the value of the "type" field.
	Type schema.EntityType `json:"type,omitempty"`
	// Status holds the value of the "status" field.
	Status schema.EntityStatus `json:"status,omitempty"`
	// StepID holds the value of the "step_id" field.
	StepID string `json:"step_id,omitempty"`
	// RunID holds the value of the "run_id" field.
	RunID schema.RunID `json:"run_id,omitempty"`
	// RetryPolicy holds the value of the "retry_policy" field.
	RetryPolicy schema.RetryPolicy `json:"retry_policy,omitempty"`
	// RetryState holds the value of the "retry_state" field.
	RetryState schema.RetryState `json:"retry_state,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the ActivityEntityQuery when eager-loading is set.
	Edges                             ActivityEntityEdges `json:"edges"`
	workflow_entity_activity_children *schema.WorkflowEntityID
	selectValues                      sql.SelectValues
}

// ActivityEntityEdges holds the relations/edges for other nodes in the graph.
type ActivityEntityEdges struct {
	// Workflow holds the value of the workflow edge.
	Workflow *WorkflowEntity `json:"workflow,omitempty"`
	// ActivityData holds the value of the activity_data edge.
	ActivityData *ActivityData `json:"activity_data,omitempty"`
	// Executions holds the value of the executions edge.
	Executions []*ActivityExecution `json:"executions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// WorkflowOrErr returns the Workflow value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActivityEntityEdges) WorkflowOrErr() (*WorkflowEntity, error) {
	if e.Workflow != nil {
		return e.Workflow, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: workflowentity.Label}
	}
	return nil, &NotLoadedError{edge: "workflow"}
}

// ActivityDataOrErr returns the ActivityData value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActivityEntityEdges) ActivityDataOrErr() (*ActivityData, error) {
	if e.ActivityData != nil {
		return e.ActivityData, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: activitydata.Label}
	}
	return nil, &NotLoadedError{edge: "activity_data"}
}

// ExecutionsOrErr returns the Executions value or an error if the edge
// was not loaded in eager-loading.
func (e ActivityEntityEdges) ExecutionsOrErr() ([]*ActivityExecution, error) {
	if e.loadedTypes[2] {
		return e.Executions, nil
	}
	return nil, &NotLoadedError{edge: "executions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*ActivityEntity) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case activityentity.FieldRetryPolicy, activityentity.FieldRetryState:
			values[i] = new([]byte)
		case activityentity.FieldID, activityentity.FieldRunID:
			values[i] = new(sql.NullInt64)
		case activityentity.FieldHandlerName, activityentity.FieldType, activityentity.FieldStatus, activityentity.FieldStepID:
			values[i] = new(sql.NullString)
		case activityentity.FieldCreatedAt, activityentity.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case activityentity.ForeignKeys[0]: // workflow_entity_activity_children
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the ActivityEntity fields.
func (ae *ActivityEntity) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case activityentity.FieldID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				ae.ID = schema.ActivityEntityID(value.Int64)
			}
		case activityentity.FieldHandlerName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field handler_name", values[i])
			} else if value.Valid {
				ae.HandlerName = value.String
			}
		case activityentity.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				ae.Type = schema.EntityType(value.String)
			}
		case activityentity.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				ae.Status = schema.EntityStatus(value.String)
			}
		case activityentity.FieldStepID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field step_id", values[i])
			} else if value.Valid {
				ae.StepID = value.String
			}
		case activityentity.FieldRunID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field run_id", values[i])
			} else if value.Valid {
				ae.RunID = schema.RunID(value.Int64)
			}
		case activityentity.FieldRetryPolicy:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_policy", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ae.RetryPolicy); err != nil {
					return fmt.Errorf("unmarshal field retry_policy: %w", err)
				}
			}
		case activityentity.FieldRetryState:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_state", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ae.RetryState); err != nil {
					return fmt.Errorf("unmarshal field retry_state: %w", err)
				}
			}
		case activityentity.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ae.CreatedAt = value.Time
			}
		case activityentity.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ae.UpdatedAt = value.Time
			}
		case activityentity.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field workflow_entity_activity_children", values[i])
			} else if value.Valid {
				ae.workflow_entity_activity_children = new(schema.WorkflowEntityID)
				*ae.workflow_entity_activity_children = schema.WorkflowEntityID(value.Int64)
			}
		default:
			ae.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the ActivityEntity.
// This includes values selected through modifiers, order, etc.
func (ae *ActivityEntity) Value(name string) (ent.Value, error) {
	return ae.selectValues.Get(name)
}

// QueryWorkflow queries the "workflow" edge of the ActivityEntity entity.
func (ae *ActivityEntity) QueryWorkflow() *WorkflowEntityQuery {
	return NewActivityEntityClient(ae.config).QueryWorkflow(ae)
}

// QueryActivityData queries the "activity_data" edge of the ActivityEntity entity.
func (ae *ActivityEntity) QueryActivityData() *ActivityDataQuery {
	return NewActivityEntityClient(ae.config).QueryActivityData(ae)
}

// QueryExecutions queries the "executions" edge of the ActivityEntity entity.
func (ae *ActivityEntity) QueryExecutions() *ActivityExecutionQuery {
	return NewActivityEntityClient(ae.config).QueryExecutions(ae)
}

// Update returns a builder for updating this ActivityEntity.
// Note that you need to call ActivityEntity.Unwrap() before calling this method if this ActivityEntity
// was returned from a transaction, and the transaction was committed or rolled back.
func (ae *ActivityEntity) Update() *ActivityEntityUpdateOne {
	return NewActivityEntityClient(ae.config).UpdateOne(ae)
}

// Unwrap unwraps the ActivityEntity entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ae *ActivityEntity) Unwrap() *ActivityEntity {
	_tx, ok := ae.config.driver.(*txDriver)
	if !ok {
		panic("ent: ActivityEntity is not a transactional entity")
	}
	ae.config.driver = _tx.drv
	return ae
}

// String implements the fmt.Stringer.
func (ae *ActivityEntity) String() string {
	var builder strings.Builder
	builder.WriteString("ActivityEntity(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ae.ID))
	builder.WriteString("handler_name=")
	builder.WriteString(ae.HandlerName)
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(fmt.Sprintf("%v", ae.Type))
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", ae.Status))
	builder.WriteString(", ")
	builder.WriteString("step_id=")
	builder.WriteString(ae.StepID)
	builder.WriteString(", ")
	builder.WriteString("run_id=")
	builder.WriteString(fmt.Sprintf("%v", ae.RunID))
	builder.WriteString(", ")
	builder.WriteString("retry_policy=")
	builder.WriteString(fmt.Sprintf("%v", ae.RetryPolicy))
	builder.WriteString(", ")
	builder.WriteString("retry_state=")
	builder.WriteString(fmt.Sprintf("%v", ae.RetryState))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(ae.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ae.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// ActivityEntities is a parsable slice of ActivityEntity.
type ActivityEntities []*ActivityEntity
