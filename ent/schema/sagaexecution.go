package schema

import (
	"time"

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
		field.String("run_id").
			Unique(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Compensating", "Compensated").
			Default("Pending"),
		field.Int("attempt").
			Default(1),
		field.JSON("output", []interface{}{}).
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the SagaExecution.
func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga", Saga.Type).
			Ref("executions").
			Unique().
			Required(),
		edge.To("steps", SagaStepExecution.Type),
	}
}
