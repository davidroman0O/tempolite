package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/go-tempolite/types"
)

// HandlerTask holds the schema definition for the HandlerTask entity.
type HandlerTask struct {
	ent.Schema
}

// Fields of the HandlerTask.
func (HandlerTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("handlerName"), // todo: maybe later change it
		field.Enum("status").
			Values(types.TaskStatusValues()...),
		field.Bytes("payload").
			Optional(),
		field.Bytes("result").
			Optional(),
		field.Bytes("error").
			Optional(),
		field.Int("numIn"),
		field.Int("numOut"),
	}
}

// Edges of the HandlerTask.
func (HandlerTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("task_context", TaskContext.Type).
			Unique(),
		edge.To("execution_context", ExecutionContext.Type).
			Unique(),
	}
}
