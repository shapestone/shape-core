package yamlv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

// NativeParser implements a native YAMLV parser without external dependencies.
type NativeParser struct{}

// NewNativeParser creates a new native YAMLV parser.
func NewNativeParser() *NativeParser {
	return &NativeParser{}
}

// Format returns FormatYAMLV.
func (p *NativeParser) Format() parser.Format {
	return parser.FormatYAMLV
}

// Parse parses YAMLV input using native implementation.
func (p *NativeParser) Parse(input string) (ast.SchemaNode, error) {
	lines := ParseLines(input)
	if len(lines) == 0 {
		return nil, parser.NewSyntaxError(ast.Position{Line: 1, Column: 1}, "empty input")
	}

	node, _, err := p.parseBlock(lines, 0, 0)
	return node, err
}

// parseBlock parses a block of lines at a given indentation level.
// Returns: (node, next_line_index, error)
func (p *NativeParser) parseBlock(lines []YAMLLine, startIdx int, baseIndent int) (ast.SchemaNode, int, error) {
	if startIdx >= len(lines) {
		return nil, startIdx, nil
	}

	firstLine := lines[startIdx]

	// Detect if this is an array or object
	if strings.HasPrefix(firstLine.Content, "-") {
		return p.parseArray(lines, startIdx, baseIndent)
	}

	return p.parseObject(lines, startIdx, baseIndent)
}

// parseObject parses an object from lines.
func (p *NativeParser) parseObject(lines []YAMLLine, startIdx int, baseIndent int) (ast.SchemaNode, int, error) {
	properties := make(map[string]ast.SchemaNode)
	pos := ast.Position{Line: lines[startIdx].LineNum, Column: 1}

	i := startIdx
	for i < len(lines) {
		line := lines[i]

		// Stop if indentation decreases
		if line.Indent < baseIndent {
			break
		}

		// Skip if indentation is too deep (belongs to a nested structure)
		if line.Indent > baseIndent {
			i++
			continue
		}

		// Check for mixed array/object syntax (invalid)
		if strings.HasPrefix(line.Content, "-") {
			return nil, i, parser.NewSyntaxError(
				ast.Position{Line: line.LineNum, Column: 1},
				"invalid YAML structure: unexpected array syntax in object context",
			)
		}

		// Parse key-value pair
		colonIdx := strings.Index(line.Content, ":")
		if colonIdx == -1 {
			return nil, i, parser.NewSyntaxError(
				ast.Position{Line: line.LineNum, Column: 1},
				"expected ':' in key-value pair",
			)
		}

		key := strings.TrimSpace(line.Content[:colonIdx])
		valueStr := strings.TrimSpace(line.Content[colonIdx+1:])

		var value ast.SchemaNode
		var err error

		if valueStr == "" {
			// Value is on next line(s) - check for nested structure
			if i+1 < len(lines) && lines[i+1].Indent > line.Indent {
				// Check if this is part of an inconsistent structure
				// (e.g., array followed by object key at wrong indentation)
				nextLine := lines[i+1]
				if strings.HasPrefix(lines[i].Content, "-") && strings.Contains(nextLine.Content, ":") && !strings.HasPrefix(nextLine.Content, "-") {
					// Malformed: array item followed by object key
					return nil, i, parser.NewSyntaxError(
						ast.Position{Line: nextLine.LineNum, Column: 1},
						"invalid YAML structure: unexpected key after array item",
					)
				}

				value, i, err = p.parseBlock(lines, i+1, lines[i+1].Indent)
				if err != nil {
					return nil, i, err
				}
			} else {
				return nil, i, parser.NewSyntaxError(
					ast.Position{Line: line.LineNum, Column: colonIdx + 2},
					"expected value after ':'",
				)
			}
		} else {
			// Inline value
			value, err = p.parseValue(valueStr, ast.Position{Line: line.LineNum, Column: colonIdx + 2})
			if err != nil {
				return nil, i, err
			}
			i++
		}

		properties[key] = value
	}

	return ast.NewObjectNode(properties, pos), i, nil
}

// parseArray parses an array from lines.
func (p *NativeParser) parseArray(lines []YAMLLine, startIdx int, baseIndent int) (ast.SchemaNode, int, error) {
	pos := ast.Position{Line: lines[startIdx].LineNum, Column: 1}

	// YAMLV arrays must contain exactly one element (the schema)
	var elementSchema ast.SchemaNode
	elementCount := 0

	i := startIdx
	for i < len(lines) {
		line := lines[i]

		// Stop if indentation decreases
		if line.Indent < baseIndent {
			break
		}

		// Skip if not at base indent
		if line.Indent != baseIndent {
			i++
			continue
		}

		// Check for mixed array/object syntax (invalid)
		if !strings.HasPrefix(line.Content, "-") {
			// Object key in array context
			if strings.Contains(line.Content, ":") && !strings.HasPrefix(line.Content, "-") {
				return nil, i, parser.NewSyntaxError(
					ast.Position{Line: line.LineNum, Column: 1},
					"invalid YAML structure: unexpected object key in array context",
				)
			}
			break
		}

		elementCount++
		if elementCount > 1 {
			return nil, i, parser.NewSyntaxError(
				ast.Position{Line: line.LineNum, Column: 1},
				fmt.Sprintf("array must contain exactly one element (the schema), got %d", elementCount),
			)
		}

		// Get value after dash
		valueStr := strings.TrimSpace(line.Content[1:])

		var err error
		if valueStr == "" {
			// Multi-line array element
			if i+1 < len(lines) && lines[i+1].Indent > line.Indent {
				elementSchema, i, err = p.parseBlock(lines, i+1, lines[i+1].Indent)
				if err != nil {
					return nil, i, err
				}
			} else {
				return nil, i, parser.NewSyntaxError(
					ast.Position{Line: line.LineNum, Column: 3},
					"expected value after '-'",
				)
			}
		} else {
			// Inline array element
			elementSchema, err = p.parseValue(valueStr, ast.Position{Line: line.LineNum, Column: 3})
			if err != nil {
				return nil, i, err
			}
			i++
		}
	}

	if elementCount == 0 {
		return nil, i, parser.NewSyntaxError(pos, "array must contain exactly one element (the schema)")
	}

	return ast.NewArrayNode(elementSchema, pos), i, nil
}

// parseValue parses a string value into a validation expression.
func (p *NativeParser) parseValue(s string, pos ast.Position) (ast.SchemaNode, error) {
	s = strings.TrimSpace(s)

	// Empty array literal
	if s == "[]" {
		return nil, parser.NewSyntaxError(pos, "array must contain exactly one element (the schema)")
	}

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

	// Quoted string literal
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		value := s[1 : len(s)-1]
		value = strings.ReplaceAll(value, `\"`, `"`)
		value = strings.ReplaceAll(value, `\\`, `\`)
		return ast.NewLiteralNode(value, pos), nil
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
func (p *NativeParser) parseArguments(argsStr string, pos ast.Position) ([]interface{}, error) {
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
func (p *NativeParser) splitArguments(s string) []string {
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
