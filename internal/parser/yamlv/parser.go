package yamlv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for YAMLV format.
// Uses a native YAML parser with no external dependencies.
type Parser struct {
	native *NativeParser
}

// NewParser creates a new YAMLV parser.
func NewParser() *Parser {
	return &Parser{
		native: NewNativeParser(),
	}
}

// Format returns FormatYAMLV.
func (p *Parser) Format() parser.Format {
	return parser.FormatYAMLV
}

// Parse parses YAMLV input and returns the AST.
// Uses native parser implementation.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	return p.native.Parse(input)
}

// The following methods are kept for backwards compatibility but are not used.
// They will be removed in a future version.

// Unused imports kept for compatibility
var _ = fmt.Sprint
var _ = strconv.Itoa
var _ = strings.Contains
