package textv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/tokenizer"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for TEXTV format.
// TEXTV is a simple line-oriented format:
//   property.name: ValidationExpression
//   user.id: UUID
//   user.name: String(1, 100)
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// NewParser creates a new TEXTV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatTEXTV.
func (p *Parser) Format() parser.Format {
	return parser.FormatTEXTV
}

// Parse parses TEXTV input and returns the AST.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	// Initialize tokenizer
	tok := tokenizer.NewTokenizer(GetMatchers()...)
	tok.Initialize(input)
	p.tokenizer = &tok

	// Load first token
	if err := p.advance(); err != nil {
		return nil, err
	}

	properties := make(map[string]ast.SchemaNode)
	pos := ast.Position{Line: 1, Column: 1}

	// Parse property lines
	for p.hasToken {
		// Skip comments and empty lines
		if p.current.Kind() == TokenComment {
			p.advance()
			continue
		}
		if p.current.Kind() == TokenNewline {
			p.advance()
			continue
		}

		// Parse property
		if err := p.parseProperty(properties); err != nil {
			return nil, err
		}
	}

	if len(properties) == 0 {
		return nil, parser.NewSyntaxError(pos, "empty schema")
	}

	// Build nested object from dot notation
	root := p.buildNestedObject(properties)
	return ast.NewObjectNode(root, pos), nil
}

// parseProperty parses a single property line: name: value
func (p *Parser) parseProperty(properties map[string]ast.SchemaNode) error {
	// Expect property name
	if p.current.Kind() != TokenPropertyName {
		return parser.NewUnexpectedTokenError(
			p.position(),
			"property name",
			p.current.Kind(),
		)
	}

	propName := p.current.ValueString()
	p.advance()

	// Expect colon
	if !p.hasToken || p.current.Kind() != TokenColon {
		return parser.NewUnexpectedTokenError(
			p.position(),
			":",
			p.current.Kind(),
		)
	}
	p.advance()

	// Parse value
	value, err := p.parseValue()
	if err != nil {
		return err
	}

	properties[propName] = value

	// Skip optional trailing comment
	if p.hasToken && p.current.Kind() == TokenComment {
		p.advance()
	}

	// Skip newline
	if p.hasToken && p.current.Kind() == TokenNewline {
		p.advance()
	}

	return nil
}

// parseValue parses a validation expression.
func (p *Parser) parseValue() (ast.SchemaNode, error) {
	if !p.hasToken {
		return nil, parser.NewUnexpectedEOFError(p.position(), "value")
	}

	pos := p.position()

	switch p.current.Kind() {
	case TokenIdentifier:
		// Type identifier: UUID, Email
		typeName := p.current.ValueString()
		p.advance()
		return ast.NewTypeNode(typeName, pos), nil

	case TokenFunction:
		// Function call: String(1, 100)
		return p.parseFunction()

	case TokenNumber:
		// Number literal
		return p.parseNumber()

	case TokenTrue:
		p.advance()
		return ast.NewLiteralNode(true, pos), nil

	case TokenFalse:
		p.advance()
		return ast.NewLiteralNode(false, pos), nil

	case TokenNull:
		p.advance()
		return ast.NewLiteralNode(nil, pos), nil

	default:
		return nil, parser.NewUnexpectedTokenError(
			pos,
			"value",
			p.current.Kind(),
		)
	}
}

// parseFunction parses a function call.
func (p *Parser) parseFunction() (ast.SchemaNode, error) {
	pos := p.position()
	funcText := p.current.ValueString()
	p.advance()

	// Extract function name and arguments
	openParen := strings.Index(funcText, "(")
	if openParen == -1 {
		return nil, parser.NewSyntaxError(pos, "invalid function syntax")
	}

	name := funcText[:openParen]
	argsStr := funcText[openParen+1 : len(funcText)-1]

	args, err := p.parseArguments(argsStr, pos)
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionNode(name, args, pos), nil
}

// parseNumber parses a number literal.
func (p *Parser) parseNumber() (ast.SchemaNode, error) {
	pos := p.position()
	numStr := p.current.ValueString()
	p.advance()

	// Try integer first
	if intVal, err := strconv.ParseInt(numStr, 10, 64); err == nil {
		return ast.NewLiteralNode(intVal, pos), nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(numStr, 64); err == nil {
		return ast.NewLiteralNode(floatVal, pos), nil
	}

	return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid number: %s", numStr))
}

// parseArguments parses function arguments from a string.
func (p *Parser) parseArguments(argsStr string, pos ast.Position) ([]interface{}, error) {
	if strings.TrimSpace(argsStr) == "" {
		return []interface{}{}, nil
	}

	parts := p.splitArguments(argsStr)
	args := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Handle number with + suffix (unbounded)
		if strings.HasSuffix(part, "+") && len(part) > 1 {
			numPart := part[:len(part)-1]
			if num, err := strconv.ParseInt(numPart, 10, 64); err == nil {
				args = append(args, num, "+")
				continue
			}
		}

		// Special symbols
		if part == "+" {
			args = append(args, "+")
			continue
		}

		// String literal (quoted)
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			value := part[1 : len(part)-1]
			value = strings.ReplaceAll(value, `\"`, `"`)
			value = strings.ReplaceAll(value, `\\`, `\`)
			args = append(args, value)
			continue
		}

		// Number
		if num, err := strconv.ParseInt(part, 10, 64); err == nil {
			args = append(args, num)
			continue
		}
		if num, err := strconv.ParseFloat(part, 64); err == nil {
			args = append(args, num)
			continue
		}

		// Boolean
		if part == "true" {
			args = append(args, true)
			continue
		}
		if part == "false" {
			args = append(args, false)
			continue
		}

		// Null
		if part == "null" {
			args = append(args, nil)
			continue
		}

		return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid argument: %s", part))
	}

	return args, nil
}

// splitArguments splits comma-separated arguments, respecting quoted strings.
func (p *Parser) splitArguments(s string) []string {
	var parts []string
	var current strings.Builder
	inString := false

	for _, r := range s {
		if r == '"' {
			current.WriteRune(r)
			inString = !inString
			continue
		}
		if r == ',' && !inString {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// buildNestedObject converts flat dot notation to nested objects.
func (p *Parser) buildNestedObject(properties map[string]ast.SchemaNode) map[string]ast.SchemaNode {
	result := make(map[string]ast.SchemaNode)

	for key, value := range properties {
		// Handle array syntax: tags[]
		if strings.HasSuffix(key, "[]") {
			key = strings.TrimSuffix(key, "[]")
			value = ast.NewArrayNode(value, value.Position())
		}

		parts := strings.Split(key, ".")
		if len(parts) == 1 {
			result[key] = value
		} else {
			p.setNested(result, parts, value)
		}
	}

	return result
}

// setNested sets a value in a nested structure.
func (p *Parser) setNested(root map[string]ast.SchemaNode, parts []string, value ast.SchemaNode) {
	if len(parts) == 1 {
		root[parts[0]] = value
		return
	}

	key := parts[0]
	rest := parts[1:]

	if existing, ok := root[key]; ok {
		if obj, ok := existing.(*ast.ObjectNode); ok {
			props := obj.Properties()
			p.setNested(props, rest, value)
			return
		}
	}

	// Create new nested object
	nested := make(map[string]ast.SchemaNode)
	p.setNested(nested, rest, value)
	root[key] = ast.NewObjectNode(nested, value.Position())
}

// advance moves to the next token, skipping whitespace (but not newlines).
func (p *Parser) advance() error {
	for {
		token, ok := p.tokenizer.NextToken()
		if !ok {
			p.hasToken = false
			p.current = nil
			return nil
		}

		// Skip whitespace (spaces, tabs) but not newlines
		if token.Kind() == "Whitespace" {
			continue
		}

		p.current = token
		p.hasToken = true
		return nil
	}
}

// position returns the current position.
func (p *Parser) position() ast.Position {
	if p.hasToken {
		return ast.Position{
			Offset: p.current.Offset(),
			Line:   p.current.Row(),
			Column: p.current.Column(),
		}
	}
	return ast.Position{
		Line:   p.tokenizer.GetRow(),
		Column: p.tokenizer.GetColumn(),
	}
}
