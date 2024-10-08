package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Entry struct {
	ent.Schema
}

// Fields of the Entry.
func (Entry) Fields() []ent.Field {
	return []ent.Field{
		field.String("taskID"),
		field.Enum("type").
			Values("handler", "saga", "side_effect", "compensation"),
	}
}

// Edges of the Entry.
func (Entry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("execution_context", ExecutionContext.Type).
			Unique(),

		// Define an edge for each possible type of task
		edge.To("handler_task", HandlerTask.Type).
			Unique(), // A node can have only one HandlerTask

		edge.To("saga_step_task", SagaTask.Type).
			Unique(), // A node can have only one SagaStepTask

		edge.To("side_effect_task", SideEffectTask.Type).
			Unique(), // A node can have only one SideEffectTask

		edge.To("compensation_task", CompensationTask.Type).
			Unique(), // A node can have only one CompensationTask
	}
}
