package ast

import (
	"fmt"
	"sort"
	"strings"
)

// ObjectNode represents object/map validation with property schemas.
type ObjectNode struct {
	properties map[string]SchemaNode // Property name → schema
	position   Position
}

// NewObjectNode creates a new object node.
func NewObjectNode(properties map[string]SchemaNode, pos Position) *ObjectNode {
	return &ObjectNode{
		properties: properties,
		position:   pos,
	}
}

// Type returns NodeTypeObject.
func (n *ObjectNode) Type() NodeType {
	return NodeTypeObject
}

// Properties returns the property schemas.
func (n *ObjectNode) Properties() map[string]SchemaNode {
	return n.properties
}

// GetProperty returns the schema for a specific property.
func (n *ObjectNode) GetProperty(name string) (SchemaNode, bool) {
	node, ok := n.properties[name]
	return node, ok
}

// Position returns the source position.
func (n *ObjectNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *ObjectNode) Accept(visitor Visitor) error {
	return visitor.VisitObject(n)
}

// String returns a string representation.
func (n *ObjectNode) String() string {
	if len(n.properties) == 0 {
		return "{}"
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(n.properties))
	for k := range n.properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%q: %s", k, n.properties[k].String())
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
