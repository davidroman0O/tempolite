package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowExecution represents a single execution instance of a workflow
type WorkflowExecution struct {
	ent.Schema
}

func (WorkflowExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "paused", "retried", "cancelled").
			Default("pending"),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").Optional(),
		field.Bool("is_replay").Default(false),
		field.Time("started_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (WorkflowExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("executions").
			Unique().
			Required(),
		edge.From("run", Run.Type).
			Ref("root_workflow").
			Unique(),
		// Links from ExecutionRelationship
		edge.From("parent_exec_relationships", ExecutionRelationship.Type).
			Ref("parent_workflow_execution"),
		edge.From("child_exec_relationships", ExecutionRelationship.Type).
			Ref("child_workflow_execution"),
	}
}
