// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sideeffectexecutiondata"
)

// SideEffectExecutionData is the model entity for the SideEffectExecutionData schema.
type SideEffectExecutionData struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// EffectTime holds the value of the "effect_time" field.
	EffectTime *time.Time `json:"effect_time,omitempty"`
	// EffectMetadata holds the value of the "effect_metadata" field.
	EffectMetadata []uint8 `json:"effect_metadata,omitempty"`
	// ExecutionContext holds the value of the "execution_context" field.
	ExecutionContext []uint8 `json:"execution_context,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SideEffectExecutionDataQuery when eager-loading is set.
	Edges                                SideEffectExecutionDataEdges `json:"edges"`
	side_effect_execution_execution_data *int
	selectValues                         sql.SelectValues
}

// SideEffectExecutionDataEdges holds the relations/edges for other nodes in the graph.
type SideEffectExecutionDataEdges struct {
	// SideEffectExecution holds the value of the side_effect_execution edge.
	SideEffectExecution *SideEffectExecution `json:"side_effect_execution,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// SideEffectExecutionOrErr returns the SideEffectExecution value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SideEffectExecutionDataEdges) SideEffectExecutionOrErr() (*SideEffectExecution, error) {
	if e.SideEffectExecution != nil {
		return e.SideEffectExecution, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: sideeffectexecution.Label}
	}
	return nil, &NotLoadedError{edge: "side_effect_execution"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SideEffectExecutionData) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case sideeffectexecutiondata.FieldEffectMetadata, sideeffectexecutiondata.FieldExecutionContext:
			values[i] = new([]byte)
		case sideeffectexecutiondata.FieldID:
			values[i] = new(sql.NullInt64)
		case sideeffectexecutiondata.FieldEffectTime:
			values[i] = new(sql.NullTime)
		case sideeffectexecutiondata.ForeignKeys[0]: // side_effect_execution_execution_data
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SideEffectExecutionData fields.
func (seed *SideEffectExecutionData) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case sideeffectexecutiondata.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			seed.ID = int(value.Int64)
		case sideeffectexecutiondata.FieldEffectTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field effect_time", values[i])
			} else if value.Valid {
				seed.EffectTime = new(time.Time)
				*seed.EffectTime = value.Time
			}
		case sideeffectexecutiondata.FieldEffectMetadata:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field effect_metadata", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &seed.EffectMetadata); err != nil {
					return fmt.Errorf("unmarshal field effect_metadata: %w", err)
				}
			}
		case sideeffectexecutiondata.FieldExecutionContext:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field execution_context", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &seed.ExecutionContext); err != nil {
					return fmt.Errorf("unmarshal field execution_context: %w", err)
				}
			}
		case sideeffectexecutiondata.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field side_effect_execution_execution_data", value)
			} else if value.Valid {
				seed.side_effect_execution_execution_data = new(int)
				*seed.side_effect_execution_execution_data = int(value.Int64)
			}
		default:
			seed.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SideEffectExecutionData.
// This includes values selected through modifiers, order, etc.
func (seed *SideEffectExecutionData) Value(name string) (ent.Value, error) {
	return seed.selectValues.Get(name)
}

// QuerySideEffectExecution queries the "side_effect_execution" edge of the SideEffectExecutionData entity.
func (seed *SideEffectExecutionData) QuerySideEffectExecution() *SideEffectExecutionQuery {
	return NewSideEffectExecutionDataClient(seed.config).QuerySideEffectExecution(seed)
}

// Update returns a builder for updating this SideEffectExecutionData.
// Note that you need to call SideEffectExecutionData.Unwrap() before calling this method if this SideEffectExecutionData
// was returned from a transaction, and the transaction was committed or rolled back.
func (seed *SideEffectExecutionData) Update() *SideEffectExecutionDataUpdateOne {
	return NewSideEffectExecutionDataClient(seed.config).UpdateOne(seed)
}

// Unwrap unwraps the SideEffectExecutionData entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (seed *SideEffectExecutionData) Unwrap() *SideEffectExecutionData {
	_tx, ok := seed.config.driver.(*txDriver)
	if !ok {
		panic("ent: SideEffectExecutionData is not a transactional entity")
	}
	seed.config.driver = _tx.drv
	return seed
}

// String implements the fmt.Stringer.
func (seed *SideEffectExecutionData) String() string {
	var builder strings.Builder
	builder.WriteString("SideEffectExecutionData(")
	builder.WriteString(fmt.Sprintf("id=%v, ", seed.ID))
	if v := seed.EffectTime; v != nil {
		builder.WriteString("effect_time=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("effect_metadata=")
	builder.WriteString(fmt.Sprintf("%v", seed.EffectMetadata))
	builder.WriteString(", ")
	builder.WriteString("execution_context=")
	builder.WriteString(fmt.Sprintf("%v", seed.ExecutionContext))
	builder.WriteByte(')')
	return builder.String()
}

// SideEffectExecutionDataSlice is a parsable slice of SideEffectExecutionData.
type SideEffectExecutionDataSlice []*SideEffectExecutionData