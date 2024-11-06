package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SagaExecution represents execution of a single saga step
type SagaExecution struct {
	ent.Schema
}

func (SagaExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("handler_name").NotEmpty(),
		field.Enum("step_type").Values("transaction", "compensation"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.Int("sequence"),
		field.String("error").Optional(),
		field.Time("started_at").Default(time.Now),
		field.Time("completed_at").Optional(),
	}
}

func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga", Saga.Type).
			Ref("executions").
			Unique().
			Required(),
		// Links from ExecutionRelationship
		edge.From("parent_exec_relationships", ExecutionRelationship.Type).
			Ref("parent_saga_execution"),
		edge.From("child_exec_relationships", ExecutionRelationship.Type).
			Ref("child_saga_execution"),
	}
}
