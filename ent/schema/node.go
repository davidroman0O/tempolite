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

		edge.To("children", Node.Type),

		edge.From("parent", Node.Type).
			Ref("children").
			Unique(),

		edge.To("handler_task", HandlerTask.Type).
			Unique(),
		edge.To("saga_step_task", SagaTask.Type).
			Unique(),
		edge.To("side_effect_task", SideEffectTask.Type).
			Unique(),
		edge.To("compensation_task", CompensationTask.Type).
			Unique(),
	}
}
