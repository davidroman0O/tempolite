package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SideEffectExecution holds the schema definition for the SideEffectExecution entity.
type SideEffectExecution struct {
	ent.Schema
}

// Fields of the SideEffectExecution.
func (SideEffectExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.Int("attempt").
			Default(1),
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

// Edges of the SideEffectExecution.
func (SideEffectExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("side_effect", SideEffect.Type).
			Ref("executions").
			Unique().
			Required(),
	}
}
