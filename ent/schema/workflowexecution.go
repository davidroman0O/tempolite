package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowExecution holds the schema definition for the WorkflowExecution entity.
type WorkflowExecution struct {
	ent.Schema
}

// Fields of the WorkflowExecution.
func (WorkflowExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled").
			Default("Pending"),
		field.JSON("output", []interface{}{}).
			Optional(),
		field.String("error").
			Optional(),
		field.Bool("is_replay").
			Default(false),
		field.Time("started_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the WorkflowExecution.
func (WorkflowExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("executions").
			Unique().
			Required(),
	}
}
