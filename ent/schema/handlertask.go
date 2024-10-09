package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// HandlerTask holds the schema definition for the HandlerTask entity.
type HandlerTask struct {
	ent.Schema
}

// Fields of the HandlerTask.
func (HandlerTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("handler_name"),
		field.Bytes("payload"),
		field.Bytes("result").
			Optional(),
		field.Bytes("error").
			Optional(),
		field.Enum("status").
			Values("pending", "in_progress", "completed", "failed"),
		field.Time("created_at"),
		field.Time("completed_at").
			Optional(),
	}
}

// Edges of the HandlerTask.
func (HandlerTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("handler_execution", HandlerExecution.Type).
			Ref("tasks").
			Unique(),
	}
}
