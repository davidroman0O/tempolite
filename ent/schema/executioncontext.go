package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ExecutionContext holds the schema definition for the ExecutionContext entity.
type ExecutionContext struct {
	ent.Schema
}

// Fields of the ExecutionContext.
func (ExecutionContext) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("current_run_id"),
		field.Enum("status").
			Values("running", "completed", "failed"),
		field.Time("start_time"),
		field.Time("end_time").
			Optional(),
	}
}

// Edges of the ExecutionContext.
func (ExecutionContext) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("handler_executions", HandlerExecution.Type),
		edge.To("side_effect_results", SideEffectResult.Type),
		edge.To("saga_executions", SagaExecution.Type),
	}
}
