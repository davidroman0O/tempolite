package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// SideEffectResult holds the schema definition for the SideEffectResult entity.
type SideEffectResult struct {
	ent.Schema
}

// Fields of the SideEffectResult.
func (SideEffectResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("execution_context_id"),
		field.String("name"),
		field.Bytes("result"),
	}
}

// Edges of the SideEffectResult.
func (SideEffectResult) Edges() []ent.Edge {
	return nil
}
