package jsonv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/tokenizer"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for JSONV format.
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// NewParser creates a new JSONV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatJSONV.
func (p *Parser) Format() parser.Format {
	return parser.FormatJSONV
}

// Parse parses JSONV input and returns the AST.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	// Initialize tokenizer with JSONV matchers
	tok := tokenizer.NewTokenizer(GetMatchers()...)
	tok.Initialize(input)
	p.tokenizer = &tok

	// Load first token
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse root value
	node, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// Ensure we consumed all input
	if p.hasToken {
		return nil, parser.NewUnexpectedTokenError(
			p.position(),
			"end of input",
			p.current.Kind(),
		)
	}

	return node, nil
}

// advance moves to the next token, skipping whitespace.
func (p *Parser) advance() error {
	for {
		token, ok := p.tokenizer.NextToken()
		if !ok {
			p.hasToken = false
			p.current = nil
			return nil
		}

		// Skip whitespace
		if token.Kind() == TokenWhitespace {
			continue
		}

		p.current = token
		p.hasToken = true
		return nil
	}
}

// expect checks if the current token matches the expected kind and advances.
func (p *Parser) expect(tokenKind string) (*tokenizer.Token, error) {
	if !p.hasToken {
		return nil, parser.NewUnexpectedEOFError(
			p.position(),
			tokenKind,
		)
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

// peek returns the current token without advancing.
func (p *Parser) peek() *tokenizer.Token {
	if p.hasToken {
		return p.current
	}
	return nil
}

// position returns the current position for error messages.
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

// parseValue dispatches to specific parsers based on token type.
func (p *Parser) parseValue() (ast.SchemaNode, error) {
	if !p.hasToken {
		return nil, parser.NewUnexpectedEOFError(p.position(), "value")
	}

	switch p.current.Kind() {
	case TokenObjectStart:
		return p.parseObject()
	case TokenArrayStart:
		return p.parseArray()
	case TokenString:
		return p.parseLiteralString()
	case TokenNumber:
		return p.parseLiteralNumber()
	case TokenTrue, TokenFalse:
		return p.parseLiteralBool()
	case TokenNull:
		return p.parseLiteralNull()
	case TokenFunction:
		return p.parseFunction()
	case TokenIdentifier:
		return p.parseType()
	default:
		return nil, parser.NewUnexpectedTokenError(
			p.position(),
			"value",
			p.current.Kind(),
		)
	}
}

// parseObject parses: { "key": value, ... }
func (p *Parser) parseObject() (ast.SchemaNode, error) {
	pos := p.position()

	// Expect '{'
	if _, err := p.expect(TokenObjectStart); err != nil {
		return nil, err
	}

	properties := make(map[string]ast.SchemaNode)

	// Check for empty object
	if p.hasToken && p.current.Kind() == TokenObjectEnd {
		p.advance()
		return ast.NewObjectNode(properties, pos), nil
	}

	// Parse properties
	for {
		// Expect property name (string)
		keyToken, err := p.expect(TokenString)
		if err != nil {
			return nil, err
		}

		// Extract key from quoted string
		key := p.unquoteString(keyToken.ValueString())

		// Expect ':'
		if _, err := p.expect(TokenColon); err != nil {
			return nil, err
		}

		// Parse property value
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		properties[key] = value

		// Check for comma or end
		if !p.hasToken {
			return nil, parser.NewUnexpectedEOFError(p.position(), "} or ,")
		}

		if p.current.Kind() == TokenObjectEnd {
			p.advance()
			break
		}

		if p.current.Kind() == TokenComma {
			p.advance()
			// After comma, expect another property
			continue
		}

		return nil, parser.NewUnexpectedTokenError(
			p.position(),
			"} or ,",
			p.current.Kind(),
		)
	}

	return ast.NewObjectNode(properties, pos), nil
}

// parseArray parses: [ elementSchema ]
func (p *Parser) parseArray() (ast.SchemaNode, error) {
	pos := p.position()

	// Expect '['
	if _, err := p.expect(TokenArrayStart); err != nil {
		return nil, err
	}

	// Parse element schema
	elementSchema, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// Expect ']'
	if _, err := p.expect(TokenArrayEnd); err != nil {
		return nil, err
	}

	return ast.NewArrayNode(elementSchema, pos), nil
}

// parseLiteralString parses a string literal.
func (p *Parser) parseLiteralString() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}

	// Unquote the string
	value := p.unquoteString(token.ValueString())
	return ast.NewLiteralNode(value, pos), nil
}

// parseLiteralNumber parses a number literal.
func (p *Parser) parseLiteralNumber() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenNumber)
	if err != nil {
		return nil, err
	}

	// Parse number
	str := token.ValueString()
	if strings.Contains(str, ".") || strings.ContainsAny(str, "eE") {
		// Float
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid number: %s", str))
		}
		return ast.NewLiteralNode(value, pos), nil
	} else {
		// Integer
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid number: %s", str))
		}
		return ast.NewLiteralNode(value, pos), nil
	}
}

// parseLiteralBool parses a boolean literal.
func (p *Parser) parseLiteralBool() (ast.SchemaNode, error) {
	pos := p.position()

	if p.current.Kind() == TokenTrue {
		p.advance()
		return ast.NewLiteralNode(true, pos), nil
	} else if p.current.Kind() == TokenFalse {
		p.advance()
		return ast.NewLiteralNode(false, pos), nil
	}

	return nil, parser.NewUnexpectedTokenError(pos, "true or false", p.current.Kind())
}

// parseLiteralNull parses a null literal.
func (p *Parser) parseLiteralNull() (ast.SchemaNode, error) {
	pos := p.position()
	_, err := p.expect(TokenNull)
	if err != nil {
		return nil, err
	}
	return ast.NewLiteralNode(nil, pos), nil
}

// parseFunction parses: FunctionName(arg1, arg2, ...)
func (p *Parser) parseFunction() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenFunction)
	if err != nil {
		return nil, err
	}

	// Parse function call: "FunctionName(args)"
	str := token.ValueString()

	// Find opening paren
	openParen := strings.Index(str, "(")
	if openParen == -1 {
		return nil, parser.NewSyntaxError(pos, "malformed function call")
	}

	// Extract function name
	name := str[:openParen]

	// Extract arguments (between parens)
	argsStr := str[openParen+1 : len(str)-1]

	// Parse arguments
	args, err := p.parseArguments(argsStr, pos)
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionNode(name, args, pos), nil
}

// parseArguments parses function arguments from a string.
func (p *Parser) parseArguments(argsStr string, pos ast.Position) ([]interface{}, error) {
	if strings.TrimSpace(argsStr) == "" {
		return []interface{}{}, nil
	}

	// Split by comma (simple split, doesn't handle nested strings with commas)
	parts := p.splitArguments(argsStr)
	args := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for number with + suffix (e.g., "5+")
		if strings.HasSuffix(part, "+") && len(part) > 1 {
			numPart := part[:len(part)-1]
			if num, err := strconv.ParseInt(numPart, 10, 64); err == nil {
				args = append(args, num)
				args = append(args, "+")
				continue
			}
		}

		// Check for special symbols
		if part == "+" {
			args = append(args, "+")
			continue
		}

		// Check for string literal
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			value := p.unquoteString(part)
			args = append(args, value)
			continue
		}

		// Check for number
		if num, err := strconv.ParseInt(part, 10, 64); err == nil {
			args = append(args, num)
			continue
		}

		if num, err := strconv.ParseFloat(part, 64); err == nil {
			args = append(args, num)
			continue
		}

		// Check for boolean
		if part == "true" {
			args = append(args, true)
			continue
		}
		if part == "false" {
			args = append(args, false)
			continue
		}

		// Check for null
		if part == "null" {
			args = append(args, nil)
			continue
		}

		return nil, parser.NewSyntaxError(pos, fmt.Sprintf("invalid function argument: %s", part))
	}

	return args, nil
}

// splitArguments splits comma-separated arguments, respecting quoted strings.
func (p *Parser) splitArguments(s string) []string {
	var parts []string
	var current strings.Builder
	inString := false
	escaped := false

	for _, r := range s {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			current.WriteRune(r)
			escaped = true
			continue
		}

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

// parseType parses a type identifier (e.g., UUID, Email).
func (p *Parser) parseType() (ast.SchemaNode, error) {
	pos := p.position()
	token, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}

	return ast.NewTypeNode(token.ValueString(), pos), nil
}

// unquoteString removes surrounding quotes and handles escape sequences.
func (p *Parser) unquoteString(s string) string {
	if len(s) < 2 {
		return s
	}

	// Remove surrounding quotes
	s = s[1 : len(s)-1]

	// Handle escape sequences
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")

	return s
}
