package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaExecution holds the schema definition for the SagaExecution entity (representing steps).
type SagaExecution struct {
	ent.Schema
}

// Fields of the SagaExecution.
func (SagaExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("handler_name").
			NotEmpty(),
		field.Enum("step_type").
			Values("Transaction", "Compensation"),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.Int("sequence").
			NonNegative(),
		field.String("error").
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("completed_at").
			Optional(),
	}
}

// Edges of the SagaExecution.
func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga", Saga.Type).
			Ref("steps").
			Unique().
			Required(),
	}
}
