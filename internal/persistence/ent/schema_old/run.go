package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Run tracks the root execution of either a workflow or activity
type Run struct {
	ent.Schema
}

func (Run) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("run_id").Unique(),
		field.Enum("type").Values("workflow", "activity"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Run) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("root_workflow", WorkflowExecution.Type).Unique(),
		edge.To("root_activity", ActivityExecution.Type).Unique(),
	}
}
