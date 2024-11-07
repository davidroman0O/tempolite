// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activitydata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
)

// ActivityData is the model entity for the ActivityData schema.
type ActivityData struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Timeout holds the value of the "timeout" field.
	Timeout int64 `json:"timeout,omitempty"`
	// MaxAttempts holds the value of the "max_attempts" field.
	MaxAttempts int `json:"max_attempts,omitempty"`
	// ScheduledFor holds the value of the "scheduled_for" field.
	ScheduledFor time.Time `json:"scheduled_for,omitempty"`
	// RetryPolicy holds the value of the "retry_policy" field.
	RetryPolicy *schema.RetryPolicy `json:"retry_policy,omitempty"`
	// Input holds the value of the "input" field.
	Input [][]uint8 `json:"input,omitempty"`
	// Output holds the value of the "output" field.
	Output [][]uint8 `json:"output,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the ActivityDataQuery when eager-loading is set.
	Edges                ActivityDataEdges `json:"edges"`
	entity_activity_data *int
	selectValues         sql.SelectValues
}

// ActivityDataEdges holds the relations/edges for other nodes in the graph.
type ActivityDataEdges struct {
	// Entity holds the value of the entity edge.
	Entity *Entity `json:"entity,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// EntityOrErr returns the Entity value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActivityDataEdges) EntityOrErr() (*Entity, error) {
	if e.Entity != nil {
		return e.Entity, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: entity.Label}
	}
	return nil, &NotLoadedError{edge: "entity"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*ActivityData) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case activitydata.FieldRetryPolicy, activitydata.FieldInput, activitydata.FieldOutput:
			values[i] = new([]byte)
		case activitydata.FieldID, activitydata.FieldTimeout, activitydata.FieldMaxAttempts:
			values[i] = new(sql.NullInt64)
		case activitydata.FieldScheduledFor:
			values[i] = new(sql.NullTime)
		case activitydata.ForeignKeys[0]: // entity_activity_data
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the ActivityData fields.
func (ad *ActivityData) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case activitydata.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			ad.ID = int(value.Int64)
		case activitydata.FieldTimeout:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field timeout", values[i])
			} else if value.Valid {
				ad.Timeout = value.Int64
			}
		case activitydata.FieldMaxAttempts:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field max_attempts", values[i])
			} else if value.Valid {
				ad.MaxAttempts = int(value.Int64)
			}
		case activitydata.FieldScheduledFor:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field scheduled_for", values[i])
			} else if value.Valid {
				ad.ScheduledFor = value.Time
			}
		case activitydata.FieldRetryPolicy:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field retry_policy", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ad.RetryPolicy); err != nil {
					return fmt.Errorf("unmarshal field retry_policy: %w", err)
				}
			}
		case activitydata.FieldInput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field input", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ad.Input); err != nil {
					return fmt.Errorf("unmarshal field input: %w", err)
				}
			}
		case activitydata.FieldOutput:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field output", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ad.Output); err != nil {
					return fmt.Errorf("unmarshal field output: %w", err)
				}
			}
		case activitydata.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field entity_activity_data", value)
			} else if value.Valid {
				ad.entity_activity_data = new(int)
				*ad.entity_activity_data = int(value.Int64)
			}
		default:
			ad.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the ActivityData.
// This includes values selected through modifiers, order, etc.
func (ad *ActivityData) Value(name string) (ent.Value, error) {
	return ad.selectValues.Get(name)
}

// QueryEntity queries the "entity" edge of the ActivityData entity.
func (ad *ActivityData) QueryEntity() *EntityQuery {
	return NewActivityDataClient(ad.config).QueryEntity(ad)
}

// Update returns a builder for updating this ActivityData.
// Note that you need to call ActivityData.Unwrap() before calling this method if this ActivityData
// was returned from a transaction, and the transaction was committed or rolled back.
func (ad *ActivityData) Update() *ActivityDataUpdateOne {
	return NewActivityDataClient(ad.config).UpdateOne(ad)
}

// Unwrap unwraps the ActivityData entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ad *ActivityData) Unwrap() *ActivityData {
	_tx, ok := ad.config.driver.(*txDriver)
	if !ok {
		panic("ent: ActivityData is not a transactional entity")
	}
	ad.config.driver = _tx.drv
	return ad
}

// String implements the fmt.Stringer.
func (ad *ActivityData) String() string {
	var builder strings.Builder
	builder.WriteString("ActivityData(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ad.ID))
	builder.WriteString("timeout=")
	builder.WriteString(fmt.Sprintf("%v", ad.Timeout))
	builder.WriteString(", ")
	builder.WriteString("max_attempts=")
	builder.WriteString(fmt.Sprintf("%v", ad.MaxAttempts))
	builder.WriteString(", ")
	builder.WriteString("scheduled_for=")
	builder.WriteString(ad.ScheduledFor.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("retry_policy=")
	builder.WriteString(fmt.Sprintf("%v", ad.RetryPolicy))
	builder.WriteString(", ")
	builder.WriteString("input=")
	builder.WriteString(fmt.Sprintf("%v", ad.Input))
	builder.WriteString(", ")
	builder.WriteString("output=")
	builder.WriteString(fmt.Sprintf("%v", ad.Output))
	builder.WriteByte(')')
	return builder.String()
}

// ActivityDataSlice is a parsable slice of ActivityData.
type ActivityDataSlice []*ActivityData