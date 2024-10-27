package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workflow holds the schema definition for the Workflow entity.
type Workflow struct {
	ent.Schema
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("step_id").NotEmpty(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Retried", "Cancelled", "Paused").
			Default("Pending"),
		field.String("identity").
			NotEmpty(),
		field.String("handler_name").
			NotEmpty(),
		field.JSON("input", [][]byte{}),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.JSON("retry_policy", RetryPolicy{}).
			Optional(),
		field.Bool("is_paused").Default(false),
		field.Bool("is_ready").Default(false),
		field.Time("timeout").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.String("continued_from_id").
			Optional().
			Comment("ID of the workflow this one was continued from"),
		field.String("retried_from_id").
			Optional().
			Comment("ID of the workflow this one was retried from").
			Nillable(),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", WorkflowExecution.Type),
		edge.To("continued_to", Workflow.Type).
			Unique().
			From("continued_from").
			Field("continued_from_id").
			Unique(),
		edge.To("retried_to", Workflow.Type).
			From("retried_from").
			Field("retried_from_id").
			Unique(),
	}
}
