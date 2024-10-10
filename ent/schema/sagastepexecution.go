package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaStepExecution holds the schema definition for the SagaStepExecution entity.
// We won't care about seeing the complete steps in advance, we will create steps as we go, even the compensation steps. it's like a log.
type SagaStepExecution struct {
	ent.Schema
}

// Fields of the SagaStepExecution.
func (SagaStepExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("handler_name").
			NotEmpty(),
		field.Enum("step_type").
			Values("Transaction", "Compensation"),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Compensated").
			Default("Pending"),
		field.Int("sequence").
			NonNegative(),
		field.Int("attempt").
			Default(1),
		field.JSON("input", []interface{}{}),
		field.JSON("output", []interface{}{}).
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the SagaStepExecution.
func (SagaStepExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga_execution", SagaExecution.Type).
			Ref("steps").
			Unique().
			Required(),
	}
}
