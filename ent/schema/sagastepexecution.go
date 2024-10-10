package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// SagaStepExecution holds the schema definition for the SagaStepExecution entity.
type SagaStepExecution struct {
	ent.Schema
}

// Fields of the SagaStepExecution.
func (SagaStepExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("saga_execution_id"),
		field.Int("step_number"),
		field.Enum("status").
			Values("pending", "in_progress", "completed", "failed", "compensated"),
		field.Time("start_time").
			Optional(),
		field.Time("end_time").
			Optional(),
	}
}

// Edges of the SagaStepExecution.
func (SagaStepExecution) Edges() []ent.Edge {
	return nil
}
