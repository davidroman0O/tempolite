package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
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
	return []ent.Edge{
		edge.From("node", Node.Type).
			Ref("compensation_task").
			Unique(),
	}
}
