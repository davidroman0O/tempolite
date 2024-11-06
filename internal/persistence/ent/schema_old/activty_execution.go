package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ActivityExecution represents a single execution instance of an activity
type ActivityExecution struct {
	ent.Schema
}

func (ActivityExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "retried", "cancelled").
			Default("pending"),
		field.Int("attempt").Default(1),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").Optional(),
		field.Time("started_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ActivityExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", Activity.Type).
			Ref("executions").
			Unique().
			Required(),
		edge.From("run", Run.Type).
			Ref("root_activity").
			Unique(),
		// Links from ExecutionRelationship
		edge.From("parent_exec_relationships", ExecutionRelationship.Type).
			Ref("parent_activity_execution"),
		edge.From("child_exec_relationships", ExecutionRelationship.Type).
			Ref("child_activity_execution"),
	}
}
