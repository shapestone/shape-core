package propsv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/tokenizer"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for PropsV format.
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// NewParser creates a new PropsV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatPropsV.
func (p *Parser) Format() parser.Format {
	return parser.FormatPropsV
}

// Parse parses PropsV input and returns the AST.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	// Initialize tokenizer
	tok := tokenizer.NewTokenizer(GetMatchers()...)
	tok.Initialize(input)
	p.tokenizer = &tok

	// Load first token
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse all properties into flat map
	properties := make(map[string]ast.SchemaNode)
	pos := ast.Position{Line: 1, Column: 1}

	for p.hasToken {
		// Skip comments (newlines are handled by whitespace matcher)
		if p.current.Kind() == TokenComment {
			p.advance()
			continue
		}

		// Parse property: key=value
		if err := p.parseProperty(properties); err != nil {
			return nil, err
		}
	}

	// Build nested object from dot notation
	root := p.buildNestedObject(properties)
	return ast.NewObjectNode(root, pos), nil
}

// parseProperty parses a single property line: key=value
func (p *Parser) parseProperty(properties map[string]ast.SchemaNode) error {
	if !p.hasToken {
		return nil
	}

	// Expect property name
	keyToken, err := p.expect(TokenPropertyName)
	if err != nil {
		return err
	}

	key := keyToken.ValueString()

	// Expect equals
	if _, err := p.expect(TokenEquals); err != nil {
		return err
	}

	// Parse value
	value, err := p.parseValue()
	if err != nil {
		return err
	}

	properties[key] = value

	// Skip optional trailing comment (newlines are consumed by whitespace matcher)
	if p.hasToken && p.current.Kind() == TokenComment {
		p.advance()
	}

	return nil
}

// parseValue parses a property value (identifier, function, or literal)
func (p *Parser) parseValue() (ast.SchemaNode, error) {
	if !p.hasToken {
		return nil, parser.NewUnexpectedEOFError(p.position(), "value")
	}

	pos := p.position()

	switch p.current.Kind() {
	case TokenFunction:
		return p.parseFunction()
	case TokenIdentifier:
		return p.parseType()
	case TokenString:
		return p.parseLiteralString()
	case TokenNumber:
		return p.parseLiteralNumber()
	case TokenTrue, TokenFalse:
		return p.parseLiteralBool()
	case TokenNull:
		return p.parseLiteralNull()
	default:
		return nil, parser.NewUnexpectedTokenError(pos, "value", p.current.Kind())
	}
}

// parseFunction parses: FunctionName(args)
func (p *Parser) parseFunction() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenFunction)
	if err != nil {
		return nil, err
	}

	str := token.ValueString()
	openParen := strings.Index(str, "(")
	if openParen == -1 {
		return nil, parser.NewSyntaxError(pos, "malformed function call")
	}

	name := str[:openParen]
	argsStr := str[openParen+1 : len(str)-1]

	args, err := p.parseArguments(argsStr, pos)
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionNode(name, args, pos), nil
}

// parseType parses a type identifier
func (p *Parser) parseType() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	return ast.NewTypeNode(token.ValueString(), pos), nil
}

// parseLiteralString parses a string literal
func (p *Parser) parseLiteralString() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}
	value := p.unquoteString(token.ValueString())
	return ast.NewLiteralNode(value, pos), nil
}

// parseLiteralNumber parses a number literal
func (p *Parser) parseLiteralNumber() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenNumber)
	if err != nil {
		return nil, err
	}

	str := token.ValueString()
	if strings.Contains(str, ".") {
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid number: %s", str))
		}
		return ast.NewLiteralNode(value, pos), nil
	}

	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid number: %s", str))
	}
	return ast.NewLiteralNode(value, pos), nil
}

// parseLiteralBool parses a boolean literal
func (p *Parser) parseLiteralBool() (ast.SchemaNode, error) {
	pos := p.position()
	if p.current.Kind() == TokenTrue {
		p.advance()
		return ast.NewLiteralNode(true, pos), nil
	}
	p.advance()
	return ast.NewLiteralNode(false, pos), nil
}

// parseLiteralNull parses a null literal
func (p *Parser) parseLiteralNull() (ast.SchemaNode, error) {
	pos := p.position()
	_, err := p.expect(TokenNull)
	if err != nil {
		return nil, err
	}
	return ast.NewLiteralNode(nil, pos), nil
}

// parseArguments parses function arguments
func (p *Parser) parseArguments(argsStr string, pos ast.Position) ([]interface{}, error) {
	if strings.TrimSpace(argsStr) == "" {
		return []interface{}{}, nil
	}

	parts := p.splitArguments(argsStr)
	args := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Handle number with + suffix
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

		// String literal
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			args = append(args, p.unquoteString(part))
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

// splitArguments splits comma-separated arguments
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

// buildNestedObject converts flat properties with dot notation to nested objects
func (p *Parser) buildNestedObject(properties map[string]ast.SchemaNode) map[string]ast.SchemaNode {
	result := make(map[string]ast.SchemaNode)

	for key, value := range properties {
		// Check for array notation: tags[]
		if strings.HasSuffix(key, "[]") {
			key = strings.TrimSuffix(key, "[]")
			value = ast.NewArrayNode(value, value.Position())
		}

		// Split by dots for nesting
		parts := strings.Split(key, ".")

		if len(parts) == 1 {
			// Simple property
			result[key] = value
		} else {
			// Nested property - build from innermost
			p.setNested(result, parts, value)
		}
	}

	return result
}

// setNested sets a value in nested map structure
func (p *Parser) setNested(root map[string]ast.SchemaNode, parts []string, value ast.SchemaNode) {
	if len(parts) == 1 {
		root[parts[0]] = value
		return
	}

	// Get or create nested object
	key := parts[0]
	if existing, ok := root[key]; ok {
		if obj, ok := existing.(*ast.ObjectNode); ok {
			p.setNested(obj.Properties(), parts[1:], value)
			return
		}
	}

	// Create new nested object
	nested := make(map[string]ast.SchemaNode)
	p.setNested(nested, parts[1:], value)
	root[key] = ast.NewObjectNode(nested, value.Position())
}

// advance moves to the next token, skipping whitespace
func (p *Parser) advance() error {
	for {
		token, ok := p.tokenizer.NextToken()
		if !ok {
			p.hasToken = false
			p.current = nil
			return nil
		}

		// Skip whitespace (includes newlines, handled by WhiteSpaceMatcher)
		if token.Kind() == "Whitespace" {
			continue
		}

		p.current = token
		p.hasToken = true
		return nil
	}
}

// expect checks if the current token matches and advances
func (p *Parser) expect(tokenKind string) (*tokenizer.Token, error) {
	if !p.hasToken {
		return nil, parser.NewUnexpectedEOFError(p.position(), tokenKind)
	}

	if p.current.Kind() != tokenKind {
		return nil, parser.NewUnexpectedTokenError(
			p.position(),
			tokenKind,
			p.current.Kind(),
		)
	}

	token := p.current
	p.advance()
	return token, nil
}

// position returns the current position
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

// unquoteString removes quotes and handles escapes
func (p *Parser) unquoteString(s string) string {
	if len(s) < 2 {
		return s
	}
	s = s[1 : len(s)-1]
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}
