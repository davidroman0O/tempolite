package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SideEffect represents a non-deterministic operation
type SideEffect struct {
	ent.Schema
}

func (SideEffect) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("identity").NotEmpty(),
		field.String("step_id").NotEmpty(),
		field.String("handler_name").NotEmpty(),
		field.Enum("status").
			Values("pending", "running", "completed", "failed").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.JSON("retry_policy", RetryPolicy{}).Optional(),
		field.Time("timeout").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SideEffect) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SideEffectExecution.Type),
		// Links from ExecutionRelationship
		edge.From("parent_relationships", ExecutionRelationship.Type).
			Ref("parent_side_effect"),
		edge.From("child_relationships", ExecutionRelationship.Type).
			Ref("child_side_effect"),
	}
}
