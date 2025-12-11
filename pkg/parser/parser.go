package parser

import "github.com/shapestone/shape/pkg/ast"

// Parser interface for all format parsers.
// Each format parser (JSONV, XMLV, etc.) implements this interface.
type Parser interface {
	// Parse converts input string to AST
	Parse(input string) (ast.SchemaNode, error)

	// Format returns the format this parser handles
	Format() Format
}
