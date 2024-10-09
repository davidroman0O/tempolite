package schema

import (
	"entgo.io/ent"
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
	return []ent.Edge{}
}
