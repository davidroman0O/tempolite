package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SignalExecution holds the schema definition for the SignalExecution entity.
type SignalExecution struct {
	ent.Schema
}

// Fields of the SignalExecution.
func (SignalExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the SignalExecution.
func (SignalExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("signal", Signal.Type).
			Ref("executions").
			Unique().
			Required(),
	}
}
