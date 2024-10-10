package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Activity holds the schema definition for the Activity entity.
type Activity struct {
	ent.Schema
}

// Fields of the Activity.
func (Activity) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("identity").
			NotEmpty(),
		field.String("handler_name").
			NotEmpty(),
		field.JSON("input", []interface{}{}),
		field.JSON("retry_policy", RetryPolicy{}).
			Optional(),
		field.Time("timeout").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Activity.
func (Activity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", ActivityExecution.Type),
		edge.From("workflow", Workflow.Type).
			Ref("activities").
			Unique(),
		edge.To("sagas", Saga.Type),
		edge.To("side_effects", SideEffect.Type),
	}
}
