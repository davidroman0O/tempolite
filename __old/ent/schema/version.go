package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FeatureFlagVersion holds the schema definition for the FeatureFlagVersion entity.
type FeatureFlagVersion struct {
	ent.Schema
}

// Fields of the FeatureFlagVersion.
func (FeatureFlagVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_type"),
		field.String("workflow_id"),
		field.String("change_id"),
		field.Int("version"),
	}
}

// Indexes of the FeatureFlagVersion.
func (FeatureFlagVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_type", "change_id").Unique(),
	}
}
