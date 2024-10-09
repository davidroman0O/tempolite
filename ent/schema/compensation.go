package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaCompensation holds the schema definition for the SagaCompensation entity.
type SagaCompensation struct {
	ent.Schema
}

// Fields of the SagaCompensation.
func (SagaCompensation) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.Int("order"),
		field.String("next_compensation_name"),
	}
}

// Edges of the SagaCompensation.
func (SagaCompensation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_unit", ExecutionUnit.Type).
			Ref("saga_compensations").
			Unique(),
		edge.To("task", Task.Type).
			Unique(),
		edge.From("transaction", SagaTransaction.Type).
			Ref("compensation").
			Unique(),
	}
}
