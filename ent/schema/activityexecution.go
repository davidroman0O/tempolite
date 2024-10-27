package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ActivityExecution holds the schema definition for the ActivityExecution entity.
type ActivityExecution struct {
	ent.Schema
}

// Fields of the ActivityExecution.
func (ActivityExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Retried").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.Int("attempt").
			Default(1),
		field.JSON("output", [][]byte{}).Optional(),
		field.String("error").
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ActivityExecution.
func (ActivityExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", Activity.Type).
			Ref("executions").
			Unique().
			Required(),
	}
}
