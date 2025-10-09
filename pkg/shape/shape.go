// Package shape provides a multi-format validation schema parser.
// It converts validation schema formats (JSONV, XMLV, PropsV, etc.) into a unified AST representation.
package shape

import (
	"fmt"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/parser/jsonv"
	"github.com/shapestone/shape/internal/parser/propsv"
	"github.com/shapestone/shape/internal/parser/xmlv"
	"github.com/shapestone/shape/pkg/ast"
)

// Parse parses input with an explicit format.
// Returns the parsed AST or an error if parsing fails.
//
// Example:
//
//	node, err := shape.Parse(parser.FormatJSONV, `{"id": UUID}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Parse(format parser.Format, input string) (ast.SchemaNode, error) {
	p, err := newParser(format)
	if err != nil {
		return nil, err
	}
	return p.Parse(input)
}

// newParser creates a parser for the specified format.
func newParser(format parser.Format) (parser.Parser, error) {
	switch format {
	case parser.FormatJSONV:
		return jsonv.NewParser(), nil
	case parser.FormatPropsV:
		return propsv.NewParser(), nil
	case parser.FormatXMLV:
		return xmlv.NewParser(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}
}

// ParseAuto auto-detects the format and parses the input.
// Returns the parsed AST, the detected format, and any error.
//
// Example:
//
//	node, format, err := shape.ParseAuto(`{"id": UUID}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Detected format: %s\n", format)
func ParseAuto(input string) (ast.SchemaNode, parser.Format, error) {
	format, err := parser.DetectFormat(input)
	if err != nil {
		return nil, parser.FormatUnknown, err
	}

	node, err := Parse(format, input)
	if err != nil {
		return nil, format, err
	}

	return node, format, nil
}

// MustParse parses or panics (useful for tests and initialization).
// Use this only when you're certain the input is valid.
//
// Example:
//
//	node := shape.MustParse(parser.FormatJSONV, `{"id": UUID}`)
func MustParse(format parser.Format, input string) ast.SchemaNode {
	node, err := Parse(format, input)
	if err != nil {
		panic(err)
	}
	return node
}
