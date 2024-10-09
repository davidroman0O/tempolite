package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.Enum("type").
			Values("handler", "side_effect", "saga_transaction", "saga_compensation"),
		field.Enum("status").
			Values("pending", "in_progress", "completed", "failed"),
		field.String("handler_name"),
		field.Bytes("payload"),
		field.Bytes("result").Optional(),
		field.Bytes("error").Optional(),
		field.Time("created_at"),
		field.Time("completed_at").Optional(),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_unit", ExecutionUnit.Type).
			Ref("tasks").
			Unique(),
	}
}
