package ent

// Hashcode returns a custom hash code for the Node entity.
func (n *Node) Hashcode() interface{} {
	// Your custom logic here, for example, using the ID field
	return n.ID
}
