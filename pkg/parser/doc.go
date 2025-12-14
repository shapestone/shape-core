// Package parser provides error types and patterns for building parsers
// using the Shape infrastructure.
//
// # Parser API Patterns
//
// Shape provides infrastructure for parser implementations but does not enforce
// a rigid interface. Instead, it recommends common patterns that parser projects
// should follow for consistency across the ecosystem.
//
// # Core Parsing Functions
//
// All parser implementations should provide two core parsing functions:
//
//	func Parse(input string) (ast.SchemaNode, error)
//	func ParseReader(reader io.Reader) (ast.SchemaNode, error)
//
// Parse() is for in-memory parsing where the entire input is loaded into memory.
// ParseReader() is for streaming parsing of large files with constant memory usage.
//
// Example implementation (from shape-json):
//
//	import (
//	    "io"
//	    "github.com/shapestone/shape-core/pkg/ast"
//	    "github.com/shapestone/shape-core/pkg/tokenizer"
//	)
//
//	func Parse(input string) (ast.SchemaNode, error) {
//	    stream := tokenizer.NewStream(input)
//	    return parseFromStream(stream)
//	}
//
//	func ParseReader(reader io.Reader) (ast.SchemaNode, error) {
//	    stream := tokenizer.NewStreamFromReader(reader)
//	    return parseFromStream(stream)
//	}
//
// # Format Detection
//
// Parser implementations should provide a Format() function that returns
// the format name as a string:
//
//	func Format() string
//
// Example:
//
//	func Format() string {
//	    return "JSON"
//	}
//
// For multi-format parsers, consider providing a DetectFormat() function:
//
//	func DetectFormat(input string) (string, error)
//	func DetectFormatFromReader(reader io.Reader) (string, error)
//
// # Output Rendering
//
// Parser implementations that support multiple output styles should provide
// rendering functions. For formats like JSON and XML, this typically means
// compact (minified) and pretty-printed (indented) output:
//
//	func RenderCompact(node ast.SchemaNode) (string, error)
//	func RenderPretty(node ast.SchemaNode) (string, error)
//
// Example:
//
//	// Compact: {"name":"Alice","age":30}
//	compact, err := json.RenderCompact(node)
//
//	// Pretty-printed:
//	// {
//	//   "name": "Alice",
//	//   "age": 30
//	// }
//	pretty, err := json.RenderPretty(node)
//
// For streaming output to files or network:
//
//	func RenderCompactWriter(node ast.SchemaNode, w io.Writer) error
//	func RenderPrettyWriter(node ast.SchemaNode, w io.Writer) error
//
// # Diffing and Comparison
//
// Parser implementations that support diffing should provide comparison functions:
//
//	func Diff(a, b ast.SchemaNode) ([]Change, error)
//	func DiffStrings(s1, s2 string) ([]Change, error)
//	func DiffReaders(r1, r2 io.Reader) ([]Change, error)
//
// Where Change is a type representing a difference between two ASTs.
//
// # Error Handling
//
// This package provides ParseError for consistent error reporting across
// parser implementations. Use NewSyntaxError, NewUnexpectedTokenError, and
// NewUnexpectedEOFError for common error cases:
//
//	if unexpected {
//	    return nil, parser.NewUnexpectedTokenError(pos, "}", got)
//	}
//
//	if atEOF {
//	    return nil, parser.NewUnexpectedEOFError(pos, "closing brace")
//	}
//
// Custom errors can be created using NewSyntaxError:
//
//	return nil, parser.NewSyntaxError(pos, "duplicate key 'id'")
//
// # Complete Example
//
// A complete parser implementation following these patterns:
//
//	package myformat
//
//	import (
//	    "io"
//	    "github.com/shapestone/shape-core/pkg/ast"
//	    "github.com/shapestone/shape-core/pkg/parser"
//	    "github.com/shapestone/shape-core/pkg/tokenizer"
//	)
//
//	// Core parsing
//	func Parse(input string) (ast.SchemaNode, error) {
//	    stream := tokenizer.NewStream(input)
//	    return parseFromStream(stream)
//	}
//
//	func ParseReader(reader io.Reader) (ast.SchemaNode, error) {
//	    stream := tokenizer.NewStreamFromReader(reader)
//	    return parseFromStream(stream)
//	}
//
//	// Format detection
//	func Format() string {
//	    return "MyFormat"
//	}
//
//	// Rendering
//	func RenderCompact(node ast.SchemaNode) (string, error) {
//	    // Implementation
//	}
//
//	func RenderPretty(node ast.SchemaNode) (string, error) {
//	    // Implementation
//	}
//
//	// Internal parsing logic
//	func parseFromStream(stream tokenizer.Stream) (ast.SchemaNode, error) {
//	    // Implementation using tokenizer and returning parser.ParseError on errors
//	}
//
// # Reference Implementations
//
// See the shape-json project for a complete reference implementation:
// https://github.com/shapestone/shape-json
//
// Other parser implementations in the Shape ecosystem:
//   - shape-yaml: YAML data format parser
//   - shape-xml: XML data format parser
//   - shape-csv: CSV data format parser
//   - shape-props: Java properties format parser
//
// See ECOSYSTEM.md for the complete list of parser projects.
package parser
