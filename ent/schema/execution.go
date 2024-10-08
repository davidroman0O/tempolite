package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Execution holds the schema definition for the Execution entity.
// One execution is one whole execution of a workflow. It could starts with an Handler or a Saga.
// We can have multiple executions for the same ExecutionContext.
type Execution struct {
	ent.Schema
}

// Fields of the Execution.
func (Execution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.Bytes("dag"),
	}
}

// Edges of the Execution.
func (Execution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("execution_context", ExecutionContext.Type).
			Unique(),
	}
}
