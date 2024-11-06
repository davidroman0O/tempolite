package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SignalExecution represents a single signal processing instance
type SignalExecution struct {
	ent.Schema
}

func (SignalExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed", "paused", "retried", "cancelled").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").Optional(),
		field.Time("started_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (SignalExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("signal", Signal.Type).
			Ref("executions").
			Unique().
			Required(),
		// Links from ExecutionRelationship
		edge.From("parent_exec_relationships", ExecutionRelationship.Type).
			Ref("parent_signal_execution"),
		edge.From("child_exec_relationships", ExecutionRelationship.Type).
			Ref("child_signal_execution"),
	}
}
