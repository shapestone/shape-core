// Package ast provides the Abstract Syntax Tree (AST) representation for validation schemas.
// All validation schema formats are parsed into this unified AST structure.
package ast

// SchemaNode is the root interface for all AST nodes.
// All validation schema elements implement this interface.
type SchemaNode interface {
	// Type returns the node type (literal, type, function, object, array)
	Type() NodeType

	// Accept allows visitor pattern traversal
	Accept(visitor Visitor) error

	// String returns a human-readable representation
	String() string

	// Position returns the source location for error messages
	Position() Position
}
