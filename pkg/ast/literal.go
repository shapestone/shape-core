package ast

import "fmt"

// LiteralNode represents an exact match validation (literal values from JSON/XML/etc.).
type LiteralNode struct {
	value    interface{} // string, int64, float64, bool, or nil
	position Position
}

// NewLiteralNode creates a new literal node.
func NewLiteralNode(value interface{}, pos Position) *LiteralNode {
	return &LiteralNode{
		value:    value,
		position: pos,
	}
}

// Type returns NodeTypeLiteral.
func (n *LiteralNode) Type() NodeType {
	return NodeTypeLiteral
}

// Value returns the literal value.
func (n *LiteralNode) Value() interface{} {
	return n.value
}

// Position returns the source position.
func (n *LiteralNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *LiteralNode) Accept(visitor Visitor) error {
	return visitor.VisitLiteral(n)
}

// String returns a string representation.
func (n *LiteralNode) String() string {
	switch v := n.value.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("%q", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
