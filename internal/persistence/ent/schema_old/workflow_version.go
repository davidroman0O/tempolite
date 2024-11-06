package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WorkflowVersion manages feature flags and versioning
type WorkflowVersion struct {
	ent.Schema
}

func (WorkflowVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_type"),
		field.String("workflow_id"),
		field.String("change_id"),
		field.Int("version"),
	}
}

func (WorkflowVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("versions").
			Unique().
			Required(),
	}
}

func (WorkflowVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_type", "change_id").
			Unique(),
	}
}
