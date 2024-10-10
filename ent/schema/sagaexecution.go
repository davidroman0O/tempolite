package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaExecution holds the schema definition for the SagaExecution entity.
type SagaExecution struct {
	ent.Schema
}

// Fields of the SagaExecution.
func (SagaExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("execution_context_id"),
		field.Enum("status").
			Values("running", "completed", "failed"),
		field.Time("start_time"),
		field.Time("end_time").
			Optional(),
	}
}

// Edges of the SagaExecution.
func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_context", ExecutionContext.Type).
			Ref("saga_executions").
			Field("execution_context_id").
			Unique().
			Required(), // Added Required() to match the non-optional foreign key field
		edge.To("steps", SagaStepExecution.Type),
	}
}
