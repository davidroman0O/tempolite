package schema

import "entgo.io/ent"

// Node holds the schema definition for the Node entity.
type Node struct {
	ent.Schema
}

// Fields of the Node.
func (Node) Fields() []ent.Field {
	return nil
}

// Edges of the Node.
func (Node) Edges() []ent.Edge {
	return nil
}
