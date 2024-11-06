package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Saga represents a saga definition with its transaction/compensation pairs
type Saga struct {
	ent.Schema
}

func (Saga) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("step_id").NotEmpty(),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "compensating", "compensated").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.JSON("saga_definition", SagaDefinitionData{}).
			Comment("Stores the full saga definition including transactions and compensations"),
		field.String("error").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Saga) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", SagaExecution.Type),
		// Links from ExecutionRelationship
		edge.From("parent_relationships", ExecutionRelationship.Type).
			Ref("parent_saga"),
		edge.From("child_relationships", ExecutionRelationship.Type).
			Ref("child_saga"),
	}
}
