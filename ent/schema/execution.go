package schema

import "entgo.io/ent"

// Execution holds the schema definition for the Execution entity.
// One execution is one whole execution of a workflow. It could starts with an Handler or a Saga.
// We can have multiple executions for the same ExecutionContext.
type Execution struct {
	ent.Schema
}

// Fields of the Execution.
func (Execution) Fields() []ent.Field {
	return nil
}

// Edges of the Execution.
func (Execution) Edges() []ent.Edge {
	return nil
}
