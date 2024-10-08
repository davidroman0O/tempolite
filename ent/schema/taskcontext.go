package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// TaskContext holds the schema definition for the TaskContext entity.
type TaskContext struct {
	ent.Schema
}

// Fields of the TaskContext.
func (TaskContext) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.Int("RetryCount").Default(0),
		field.Int("MaxRetry").Default(1),
	}
}

// Edges of the TaskContext.
func (TaskContext) Edges() []ent.Edge {
	return nil
}
