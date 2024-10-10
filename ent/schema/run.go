package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Run will be created ONCE
// Every retry of root Workflow or Activity will set run_id with its id
type Run struct {
	ent.Schema
}

func (Run) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"). // current instance
					Unique(),
		field.Enum("type").
			Values("workflow", "activity"),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (Run) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workflow", Activity.Type),
		edge.To("activities", Activity.Type),
	}
}
