package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SideEffectExecution represents a single execution of a side effect
type SideEffectExecution struct {
	ent.Schema
}

func (SideEffectExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),
		field.Enum("status").
			Values("pending", "running", "completed", "failed").
			Default("pending"),
		field.String("queue_name").Default("default"),
		field.Int("attempt").Default(1),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").Optional(),
		field.Time("started_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (SideEffectExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("side_effect", SideEffect.Type).
			Ref("executions").
			Unique().
			Required(),
		// Links from ExecutionRelationship
		edge.From("parent_exec_relationships", ExecutionRelationship.Type).
			Ref("parent_side_effect_execution"),
		edge.From("child_exec_relationships", ExecutionRelationship.Type).
			Ref("child_side_effect_execution"),
	}
}
