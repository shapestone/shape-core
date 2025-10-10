package yamlv

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for YAMLV format.
// NOTE: For v0.1.0, this uses gopkg.in/yaml.v3 to parse YAML structure.
// Future versions (v0.2.0+) should replace this with a native parser using
// the tokenizer framework for consistency and zero external dependencies.
type Parser struct{}

// NewParser creates a new YAMLV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatYAMLV.
func (p *Parser) Format() parser.Format {
	return parser.FormatYAMLV
}

// Parse parses YAMLV input and returns the AST.
// Uses yaml.v3 to parse YAML structure, then converts to our AST.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	var data interface{}

	if err := yaml.Unmarshal([]byte(input), &data); err != nil {
		return nil, parser.NewSyntaxError(ast.Position{Line: 1, Column: 1}, err.Error())
	}

	return p.convertToAST(data, ast.Position{Line: 1, Column: 1})
}

// convertToAST converts a yaml.v3 parsed value to our AST.
func (p *Parser) convertToAST(data interface{}, pos ast.Position) (ast.SchemaNode, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return p.parseObject(v, pos)
	case []interface{}:
		return p.parseArray(v, pos)
	case string:
		return p.parseString(v, pos)
	case int:
		return ast.NewLiteralNode(int64(v), pos), nil
	case int64:
		return ast.NewLiteralNode(v, pos), nil
	case float64:
		return ast.NewLiteralNode(v, pos), nil
	case bool:
		return ast.NewLiteralNode(v, pos), nil
	case nil:
		return ast.NewLiteralNode(nil, pos), nil
	default:
		return nil, parser.NewSyntaxError(pos, fmt.Sprintf("unsupported type: %T", v))
	}
}

// parseObject converts a YAML map to ObjectNode.
func (p *Parser) parseObject(data map[string]interface{}, pos ast.Position) (ast.SchemaNode, error) {
	properties := make(map[string]ast.SchemaNode)

	for key, value := range data {
		node, err := p.convertToAST(value, pos)
		if err != nil {
			return nil, fmt.Errorf("parsing property %q: %w", key, err)
		}
		properties[key] = node
	}

	return ast.NewObjectNode(properties, pos), nil
}

// parseArray converts a YAML array to ArrayNode.
// In YAMLV, arrays contain a single element that represents the schema for all items.
func (p *Parser) parseArray(data []interface{}, pos ast.Position) (ast.SchemaNode, error) {
	if len(data) == 0 {
		return nil, parser.NewSyntaxError(pos, "array must contain exactly one element (the schema)")
	}

	if len(data) > 1 {
		return nil, parser.NewSyntaxError(pos, fmt.Sprintf("array must contain exactly one element (the schema), got %d", len(data)))
	}

	elementSchema, err := p.convertToAST(data[0], pos)
	if err != nil {
		return nil, fmt.Errorf("parsing array element schema: %w", err)
	}

	return ast.NewArrayNode(elementSchema, pos), nil
}

// parseString parses a YAML string value into a validation expression.
// Handles:
// - Type identifiers: "UUID", "Email", "ISO-8601"
// - Functions: "String(1, 100)", "Integer(18, 120)"
// - Literals: "true", "false", "null", numbers
// - String literals: everything else
func (p *Parser) parseString(s string, pos ast.Position) (ast.SchemaNode, error) {
	s = strings.TrimSpace(s)

	// Boolean literals
	if s == "true" {
		return ast.NewLiteralNode(true, pos), nil
	}
	if s == "false" {
		return ast.NewLiteralNode(false, pos), nil
	}

	// Null literal
	if s == "null" {
		return ast.NewLiteralNode(nil, pos), nil
	}

	// Number literal
	if num, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}

	// Function call: FunctionName(args)
	if strings.Contains(s, "(") && strings.HasSuffix(s, ")") {
		openParen := strings.Index(s, "(")
		name := s[:openParen]
		argsStr := s[openParen+1 : len(s)-1]

		args, err := p.parseArguments(argsStr, pos)
		if err != nil {
			return nil, err
		}

		return ast.NewFunctionNode(name, args, pos), nil
	}

	// Type identifier (uppercase start)
	if len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' {
		return ast.NewTypeNode(s, pos), nil
	}

	// Default: treat as string literal
	return ast.NewLiteralNode(s, pos), nil
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
