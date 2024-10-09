package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// HandlerExecution holds the schema definition for the HandlerExecution entity.
type HandlerExecution struct {
	ent.Schema
}

// Fields of the HandlerExecution.
func (HandlerExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"),
		field.String("handler_name"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed"),
		field.Time("start_time"),
		field.Time("end_time").
			Optional(),
		field.Int("retry_count").
			Default(0),
		field.Int("max_retries").
			Default(3),
	}
}

// Edges of the HandlerExecution.
func (HandlerExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_context", ExecutionContext.Type).
			Ref("handler_executions").
			Unique(),
		edge.To("children", HandlerExecution.Type).
			From("parent").
			Unique(),
		edge.To("tasks", HandlerTask.Type),
	}
}
