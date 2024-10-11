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
		field.String("parent_id"),
		field.String("child_id"),
		field.Enum("parent_type").
			Values("workflow", "activity", "saga", "side_effect"),
		field.Enum("child_type").
			Values("workflow", "activity", "saga", "side_effect"),
	}
}

func (ExecutionRelationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id", "child_id").
			Unique(),
	}
}
