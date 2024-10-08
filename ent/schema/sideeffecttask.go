package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// SideEffectTask holds the schema definition for the SideEffectTask entity.
type SideEffectTask struct {
	ent.Schema
}

// Fields of the SideEffectTask.
func (SideEffectTask) Fields() []ent.Field {
	return nil
}

// Edges of the SideEffectTask.
func (SideEffectTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("node", Node.Type).
			Ref("side_effect_task").
			Unique(),
	}
}
