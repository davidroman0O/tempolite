package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Signal represents an asynchronous input mechanism
type Signal struct {
	ent.Schema
}

func (Signal) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("step_id").NotEmpty(),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "paused", "retried", "cancelled").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.Time("created_at").Default(time.Now),
		field.Bool("consumed").Default(false),
	}
}

func (Signal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SignalExecution.Type),
		// Links from ExecutionRelationship
		edge.From("parent_relationships", ExecutionRelationship.Type).
			Ref("parent_signal"),
		edge.From("child_relationships", ExecutionRelationship.Type).
			Ref("child_signal"),
	}
}
