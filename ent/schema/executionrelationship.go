package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExecutionRelationship schema
type ExecutionRelationship struct {
	ent.Schema
}

func (ExecutionRelationship) Fields() []ent.Field {
	return []ent.Field{
		field.String("run_id").NotEmpty(),
		field.String("parent_entity_id").NotEmpty(),
		field.String("child_entity_id").NotEmpty(),
		field.String("parent_id"),
		field.String("child_id"),
		field.Enum("parent_type").
			Values("workflow", "activity", "saga", "side_effect", "yield"),
		field.Enum("child_type").
			Values("workflow", "activity", "saga", "side_effect", "yield"),
		field.String("parent_step_id").NotEmpty(),
		field.String("child_step_id").NotEmpty(),
	}
}

func (ExecutionRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id", "child_id").
			Unique(),
	}
}
