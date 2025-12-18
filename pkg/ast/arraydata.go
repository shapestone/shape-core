package ast

import (
	"fmt"
	"strings"
)

// ArrayDataNode represents actual array data with elements.
// This is distinct from ArrayNode which represents array schema validation.
type ArrayDataNode struct {
	elements []SchemaNode // Actual array elements
	position Position
}

// NewArrayDataNode creates a new array data node.
func NewArrayDataNode(elements []SchemaNode, pos Position) *ArrayDataNode {
	return &ArrayDataNode{
		elements: elements,
		position: pos,
	}
}

// Type returns NodeTypeArrayData.
func (n *ArrayDataNode) Type() NodeType {
	return NodeTypeArrayData
}

// Elements returns the array elements.
func (n *ArrayDataNode) Elements() []SchemaNode {
	return n.elements
}

// Len returns the number of elements in the array.
func (n *ArrayDataNode) Len() int {
	return len(n.elements)
}

// Get returns the element at the specified index.
// Returns nil if index is out of bounds.
func (n *ArrayDataNode) Get(index int) SchemaNode {
	if index < 0 || index >= len(n.elements) {
		return nil
	}
	return n.elements[index]
}

// Position returns the source position.
func (n *ArrayDataNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *ArrayDataNode) Accept(visitor Visitor) error {
	return visitor.VisitArrayData(n)
}

// String returns a string representation.
func (n *ArrayDataNode) String() string {
	if len(n.elements) == 0 {
		return "[]"
	}

	parts := make([]string, len(n.elements))
	for i, elem := range n.elements {
		parts[i] = elem.String()
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}
