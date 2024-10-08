package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// SagaTask holds the schema definition for the SagaTask entity.
type SagaTask struct {
	ent.Schema
}

// Fields of the SagaTask.
func (SagaTask) Fields() []ent.Field {
	return nil
}

// Edges of the SagaTask.
func (SagaTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("node", Node.Type).
			Ref("saga_step_task").
			Unique(),
	}
}
