package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Activity represents a single unit of work definition and state
type Activity struct {
	ent.Schema
}

func (Activity) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("identity").NotEmpty(),
		field.String("step_id").NotEmpty(),
		field.String("handler_name").NotEmpty(),
		field.JSON("input", [][]byte{}),
		field.String("queue_name").Default("default"),
		field.JSON("retry_policy", RetryPolicy{}),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "paused", "retried", "cancelled").
			Default("pending"),
		field.String("max_duration").Optional(),
		field.Time("timeout").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (Activity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", ActivityExecution.Type),
		// Links from ExecutionRelationship
		edge.From("parent_relationships", ExecutionRelationship.Type).
			Ref("parent_entity_id"),
		edge.From("child_relationships", ExecutionRelationship.Type).
			Ref("child_entity_id"),
	}
}
