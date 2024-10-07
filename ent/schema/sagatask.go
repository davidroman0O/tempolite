package schema

import "entgo.io/ent"

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
	return nil
}
