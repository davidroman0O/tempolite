package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workflow represents a workflow definition and its current state
type Workflow struct {
	ent.Schema
}

func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("step_id").NotEmpty(),
		field.String("identity").NotEmpty(),
		field.String("handler_name").NotEmpty(),
		field.JSON("input", [][]byte{}),
		field.String("queue_name").Default("default"),
		field.JSON("retry_policy", RetryPolicy{}),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "retried", "cancelled", "paused").
			Default("pending"),
		field.Bool("is_paused").Default(false),
		field.Bool("is_ready").Default(false),
		field.String("max_duration").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", WorkflowExecution.Type),
		edge.To("continued_to", Workflow.Type).Unique(),
		edge.From("continued_from", Workflow.Type).Ref("continued_to").Unique(),
		edge.To("retried_to", Workflow.Type).Unique(),
		edge.From("retried_from", Workflow.Type).Ref("retried_to").Unique(),
		// Links from ExecutionRelationship
		edge.From("parent_relationships", ExecutionRelationship.Type).Ref("parent_workflow"),
		edge.From("child_relationships", ExecutionRelationship.Type).Ref("child_workflow"),
	}
}
