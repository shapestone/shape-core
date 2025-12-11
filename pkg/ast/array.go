package ast

import "fmt"

// ArrayNode represents array validation with element schema.
type ArrayNode struct {
	elementSchema SchemaNode // Schema for all array elements
	position      Position
}

// NewArrayNode creates a new array node.
func NewArrayNode(elementSchema SchemaNode, pos Position) *ArrayNode {
	return &ArrayNode{
		elementSchema: elementSchema,
		position:      pos,
	}
}

// Type returns NodeTypeArray.
func (n *ArrayNode) Type() NodeType {
	return NodeTypeArray
}

// ElementSchema returns the element schema.
func (n *ArrayNode) ElementSchema() SchemaNode {
	return n.elementSchema
}

// Position returns the source position.
func (n *ArrayNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *ArrayNode) Accept(visitor Visitor) error {
	return visitor.VisitArray(n)
}

// String returns a string representation.
func (n *ArrayNode) String() string {
	return fmt.Sprintf("[%s]", n.elementSchema.String())
}
