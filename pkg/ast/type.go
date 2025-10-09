package ast

// TypeNode represents type validation (built-in type identifiers like UUID, Email, etc.).
type TypeNode struct {
	typeName string // "UUID", "Email", "ISO-8601", etc.
	position Position
}

// NewTypeNode creates a new type node.
func NewTypeNode(typeName string, pos Position) *TypeNode {
	return &TypeNode{
		typeName: typeName,
		position: pos,
	}
}

// Type returns NodeTypeType.
func (n *TypeNode) Type() NodeType {
	return NodeTypeType
}

// TypeName returns the type identifier name.
func (n *TypeNode) TypeName() string {
	return n.typeName
}

// Position returns the source position.
func (n *TypeNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *TypeNode) Accept(visitor Visitor) error {
	return visitor.VisitType(n)
}

// String returns a string representation.
func (n *TypeNode) String() string {
	return n.typeName
}
