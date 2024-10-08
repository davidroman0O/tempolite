// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/go-tempolite/ent/compensationtask"
	"github.com/davidroman0O/go-tempolite/ent/entry"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/sagatask"
	"github.com/davidroman0O/go-tempolite/ent/sideeffecttask"
)

// Entry is the model entity for the Entry schema.
type Entry struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// TaskID holds the value of the "taskID" field.
	TaskID string `json:"taskID,omitempty"`
	// Type holds the value of the "type" field.
	Type entry.Type `json:"type,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the EntryQuery when eager-loading is set.
	Edges                   EntryEdges `json:"edges"`
	entry_execution_context *string
	entry_handler_task      *string
	entry_saga_step_task    *int
	entry_side_effect_task  *int
	entry_compensation_task *int
	selectValues            sql.SelectValues
}

// EntryEdges holds the relations/edges for other nodes in the graph.
type EntryEdges struct {
	// ExecutionContext holds the value of the execution_context edge.
	ExecutionContext *ExecutionContext `json:"execution_context,omitempty"`
	// HandlerTask holds the value of the handler_task edge.
	HandlerTask *HandlerTask `json:"handler_task,omitempty"`
	// SagaStepTask holds the value of the saga_step_task edge.
	SagaStepTask *SagaTask `json:"saga_step_task,omitempty"`
	// SideEffectTask holds the value of the side_effect_task edge.
	SideEffectTask *SideEffectTask `json:"side_effect_task,omitempty"`
	// CompensationTask holds the value of the compensation_task edge.
	CompensationTask *CompensationTask `json:"compensation_task,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [5]bool
}

// ExecutionContextOrErr returns the ExecutionContext value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntryEdges) ExecutionContextOrErr() (*ExecutionContext, error) {
	if e.ExecutionContext != nil {
		return e.ExecutionContext, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: executioncontext.Label}
	}
	return nil, &NotLoadedError{edge: "execution_context"}
}

// HandlerTaskOrErr returns the HandlerTask value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntryEdges) HandlerTaskOrErr() (*HandlerTask, error) {
	if e.HandlerTask != nil {
		return e.HandlerTask, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: handlertask.Label}
	}
	return nil, &NotLoadedError{edge: "handler_task"}
}

// SagaStepTaskOrErr returns the SagaStepTask value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntryEdges) SagaStepTaskOrErr() (*SagaTask, error) {
	if e.SagaStepTask != nil {
		return e.SagaStepTask, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: sagatask.Label}
	}
	return nil, &NotLoadedError{edge: "saga_step_task"}
}

// SideEffectTaskOrErr returns the SideEffectTask value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntryEdges) SideEffectTaskOrErr() (*SideEffectTask, error) {
	if e.SideEffectTask != nil {
		return e.SideEffectTask, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: sideeffecttask.Label}
	}
	return nil, &NotLoadedError{edge: "side_effect_task"}
}

// CompensationTaskOrErr returns the CompensationTask value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntryEdges) CompensationTaskOrErr() (*CompensationTask, error) {
	if e.CompensationTask != nil {
		return e.CompensationTask, nil
	} else if e.loadedTypes[4] {
		return nil, &NotFoundError{label: compensationtask.Label}
	}
	return nil, &NotLoadedError{edge: "compensation_task"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Entry) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case entry.FieldID:
			values[i] = new(sql.NullInt64)
		case entry.FieldTaskID, entry.FieldType:
			values[i] = new(sql.NullString)
		case entry.ForeignKeys[0]: // entry_execution_context
			values[i] = new(sql.NullString)
		case entry.ForeignKeys[1]: // entry_handler_task
			values[i] = new(sql.NullString)
		case entry.ForeignKeys[2]: // entry_saga_step_task
			values[i] = new(sql.NullInt64)
		case entry.ForeignKeys[3]: // entry_side_effect_task
			values[i] = new(sql.NullInt64)
		case entry.ForeignKeys[4]: // entry_compensation_task
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Entry fields.
func (e *Entry) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case entry.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			e.ID = int(value.Int64)
		case entry.FieldTaskID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field taskID", values[i])
			} else if value.Valid {
				e.TaskID = value.String
			}
		case entry.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				e.Type = entry.Type(value.String)
			}
		case entry.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field entry_execution_context", values[i])
			} else if value.Valid {
				e.entry_execution_context = new(string)
				*e.entry_execution_context = value.String
			}
		case entry.ForeignKeys[1]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field entry_handler_task", values[i])
			} else if value.Valid {
				e.entry_handler_task = new(string)
				*e.entry_handler_task = value.String
			}
		case entry.ForeignKeys[2]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field entry_saga_step_task", value)
			} else if value.Valid {
				e.entry_saga_step_task = new(int)
				*e.entry_saga_step_task = int(value.Int64)
			}
		case entry.ForeignKeys[3]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field entry_side_effect_task", value)
			} else if value.Valid {
				e.entry_side_effect_task = new(int)
				*e.entry_side_effect_task = int(value.Int64)
			}
		case entry.ForeignKeys[4]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field entry_compensation_task", value)
			} else if value.Valid {
				e.entry_compensation_task = new(int)
				*e.entry_compensation_task = int(value.Int64)
			}
		default:
			e.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Entry.
// This includes values selected through modifiers, order, etc.
func (e *Entry) Value(name string) (ent.Value, error) {
	return e.selectValues.Get(name)
}

// QueryExecutionContext queries the "execution_context" edge of the Entry entity.
func (e *Entry) QueryExecutionContext() *ExecutionContextQuery {
	return NewEntryClient(e.config).QueryExecutionContext(e)
}

// QueryHandlerTask queries the "handler_task" edge of the Entry entity.
func (e *Entry) QueryHandlerTask() *HandlerTaskQuery {
	return NewEntryClient(e.config).QueryHandlerTask(e)
}

// QuerySagaStepTask queries the "saga_step_task" edge of the Entry entity.
func (e *Entry) QuerySagaStepTask() *SagaTaskQuery {
	return NewEntryClient(e.config).QuerySagaStepTask(e)
}

// QuerySideEffectTask queries the "side_effect_task" edge of the Entry entity.
func (e *Entry) QuerySideEffectTask() *SideEffectTaskQuery {
	return NewEntryClient(e.config).QuerySideEffectTask(e)
}

// QueryCompensationTask queries the "compensation_task" edge of the Entry entity.
func (e *Entry) QueryCompensationTask() *CompensationTaskQuery {
	return NewEntryClient(e.config).QueryCompensationTask(e)
}

// Update returns a builder for updating this Entry.
// Note that you need to call Entry.Unwrap() before calling this method if this Entry
// was returned from a transaction, and the transaction was committed or rolled back.
func (e *Entry) Update() *EntryUpdateOne {
	return NewEntryClient(e.config).UpdateOne(e)
}

// Unwrap unwraps the Entry entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (e *Entry) Unwrap() *Entry {
	_tx, ok := e.config.driver.(*txDriver)
	if !ok {
		panic("ent: Entry is not a transactional entity")
	}
	e.config.driver = _tx.drv
	return e
}

// String implements the fmt.Stringer.
func (e *Entry) String() string {
	var builder strings.Builder
	builder.WriteString("Entry(")
	builder.WriteString(fmt.Sprintf("id=%v, ", e.ID))
	builder.WriteString("taskID=")
	builder.WriteString(e.TaskID)
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(fmt.Sprintf("%v", e.Type))
	builder.WriteByte(')')
	return builder.String()
}

// Entries is a parsable slice of Entry.
type Entries []*Entry
