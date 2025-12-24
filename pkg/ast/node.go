// Package ast provides a universal, format-agnostic Abstract Syntax Tree (AST)
// representation for both validation schemas and parsed data structures.
//
// The AST serves dual purposes:
//  1. Schema Validation: TypeNode, FunctionNode define constraints
//  2. Data Representation: LiteralNode, ObjectNode hold parsed values from any format
//
// This unified representation enables:
//   - Source position tracking for precise error messages
//   - Structural queries (JSONPath, XPath) via conversion to Go types
//   - Document diffing and comparison
//   - Programmatic construction of JSON/XML/YAML
//   - Format transformations and conversions
//   - Cross-format validation with consistent semantics
//
// Supported formats: JSON, XML, YAML, CSV, and custom formats.
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
