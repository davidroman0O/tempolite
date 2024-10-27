package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SideEffect holds the schema definition for the SideEffect entity.
type SideEffect struct {
	ent.Schema
}

// Fields of the SideEffect.
func (SideEffect) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("identity").
			NotEmpty(),
		field.String("step_id").NotEmpty(),
		field.String("handler_name").
			NotEmpty(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.JSON("retry_policy", RetryPolicy{}).
			Optional(),
		field.Time("timeout").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the SideEffect.
func (SideEffect) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SideEffectExecution.Type),
	}
}
