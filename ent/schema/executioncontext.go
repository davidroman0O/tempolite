package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ExecutionContext holds the schema definition for the ExecutionContext entity.
// Each time you use `tempolite` to enqueue an Handler or Saga, it will create one unique ExecutionContext.
// That whole flow might fail and retry, thus creating many Execution for the same ExecutionContext.
type ExecutionContext struct {
	ent.Schema
}

// Fields of the ExecutionContext.
func (ExecutionContext) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
	}
}

// Edges of the ExecutionContext.
func (ExecutionContext) Edges() []ent.Edge {
	return []ent.Edge{}
}
