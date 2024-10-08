package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Node holds the schema definition for the Node entity.
type Node struct {
	ent.Schema
}

// Fields of the Node.
func (Node) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
	}
}

// Edges of the Node.
func (Node) Edges() []ent.Edge {
	return []ent.Edge{
		// Self-referencing edge for parent node relationship
		edge.To("children", Node.Type).
			From("parent").
			Unique(), // A node has only one parent

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
