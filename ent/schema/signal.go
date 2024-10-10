package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Signal holds the schema definition for the Signal entity.
type Signal struct {
	ent.Schema
}

// Fields of the Signal.
func (Signal) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("name").
			NotEmpty(),
		field.JSON("data", []interface{}{}).
			Optional(),
		field.Enum("status").
			Values("Pending", "Received", "Processed").
			Default("Pending"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Signal.
func (Signal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow_execution", WorkflowExecution.Type).
			Ref("signals").
			Unique(),
	}
}
