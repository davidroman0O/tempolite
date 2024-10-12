package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Saga holds the schema definition for the Saga entity.
type Saga struct {
	ent.Schema
}

// Fields of the Saga.
func (Saga) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("name").
			NotEmpty(),
		field.String("step_id").NotEmpty(),
		field.JSON("input", []interface{}{}),
		field.JSON("retry_policy", RetryPolicy{}).
			Optional(),
		field.Time("timeout").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Saga.
func (Saga) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SagaExecution.Type),
	}
}
