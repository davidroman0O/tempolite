package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// HandlerTask holds the schema definition for the HandlerTask entity.
type HandlerTask struct {
	ent.Schema
}

// Fields of the HandlerTask.
func (HandlerTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
	}
}

// Edges of the HandlerTask.
func (HandlerTask) Edges() []ent.Edge {
	return nil
}
