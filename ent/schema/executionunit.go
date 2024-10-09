package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ExecutionUnit holds the schema definition for the ExecutionUnit entity.
type ExecutionUnit struct {
	ent.Schema
}

// Fields of the ExecutionUnit.
func (ExecutionUnit) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.Enum("type").
			Values("handler", "side_effect", "saga"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed"),
		field.Time("start_time"),
		field.Time("end_time").Optional(),
		field.Int("retry_count").Default(0),
		field.Int("max_retries").Default(3),
	}
}

// Edges of the ExecutionUnit.
func (ExecutionUnit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_context", ExecutionContext.Type).
			Ref("execution_units").
			Unique(),
		edge.To("children", ExecutionUnit.Type).
			From("parent").
			Unique(),
		edge.To("tasks", Task.Type),
		edge.To("saga_transactions", SagaTransaction.Type),
		edge.To("saga_compensations", SagaCompensation.Type),
	}
}
