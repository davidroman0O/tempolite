package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Signal holds the schema definition for the Signal entity.
type Signal struct {
	ent.Schema
}

// Fields of the Signal.
func (Signal) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("step_id").
			NotEmpty(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now),
		field.Bool("consumed").
			Default(false),
	}
}

// Edges of the Signal.
func (Signal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SignalExecution.Type),
	}
}
