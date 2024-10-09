package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaTransaction holds the schema definition for the SagaTransaction entity.
type SagaTransaction struct {
	ent.Schema
}

// Fields of the SagaTransaction.
func (SagaTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.Int("order"),
		field.String("next_transaction_name"),
		field.String("failure_compensation_name"),
	}
}

// Edges of the SagaTransaction.
func (SagaTransaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution_unit", ExecutionUnit.Type).
			Ref("saga_transactions").
			Unique(),
		edge.To("task", Task.Type).
			Unique(),
		edge.To("compensation", SagaCompensation.Type).
			Unique(),
	}
}
