package schema

import (
	"entgo.io/ent"
)

// CompensationTask holds the schema definition for the CompensationTask entity.
type CompensationTask struct {
	ent.Schema
}

// Fields of the CompensationTask.
func (CompensationTask) Fields() []ent.Field {
	return nil
}

// Edges of the CompensationTask.
func (CompensationTask) Edges() []ent.Edge {
	return []ent.Edge{}
}
