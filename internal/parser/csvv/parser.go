package csvv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/tokenizer"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for CSVV format.
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// NewParser creates a new CSVV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatCSVV.
func (p *Parser) Format() parser.Format {
	return parser.FormatCSVV
}

// Parse parses CSVV input and returns the AST.
// Expected format:
//   # Optional comments
//   header1,header2,header3
//   Schema1,Schema2,Schema3
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	// Initialize tokenizer
	tok := tokenizer.NewTokenizer(GetMatchers()...)
	tok.Initialize(input)
	p.tokenizer = &tok

	// Load first token
	if err := p.advance(); err != nil {
		return nil, err
	}

	pos := p.position()

	// Skip comment lines
	for p.hasToken && p.current.Kind() == TokenComment {
		p.advance()
		// Skip newline after comment
		if p.hasToken && p.current.Kind() == TokenNewline {
			p.advance()
		}
	}

	// Parse header row
	headers, err := p.parseRow()
	if err != nil {
		return nil, fmt.Errorf("parsing header row: %w", err)
	}

	if len(headers) == 0 {
		return nil, parser.NewSyntaxError(p.position(), "empty header row")
	}

	// Expect newline after header
	if p.hasToken && p.current.Kind() == TokenNewline {
		p.advance()
	}

	// Parse validation row
	schemas, err := p.parseRow()
	if err != nil {
		return nil, fmt.Errorf("parsing validation row: %w", err)
	}

	if len(schemas) != len(headers) {
		return nil, parser.NewSyntaxError(
			p.position(),
			fmt.Sprintf("validation row has %d columns, header has %d", len(schemas), len(headers)),
		)
	}

	// Build object from headers and schemas
	properties := make(map[string]ast.SchemaNode)
	for i, header := range headers {
		schemaText := schemas[i]
		schemaNode, err := p.parseSchema(schemaText, pos)
		if err != nil {
			return nil, fmt.Errorf("parsing schema for column %q: %w", header, err)
		}
		properties[header] = schemaNode
	}

	return ast.NewObjectNode(properties, pos), nil
}

// parseRow parses a CSV row and returns cell values
func (p *Parser) parseRow() ([]string, error) {
	var cells []string

	for p.hasToken {
		// End of row
		if p.current.Kind() == TokenNewline {
			break
		}

		// Expect cell
		if p.current.Kind() != TokenCell {
			return nil, parser.NewUnexpectedTokenError(
				p.position(),
				"cell value",
				p.current.Kind(),
			)
		}

		cellValue := p.unquoteCell(p.current.ValueString())
		cells = append(cells, cellValue)
		p.advance()

		// Check for comma (more cells) or newline/EOF (end of row)
		if p.hasToken && p.current.Kind() == TokenComma {
			p.advance() // consume comma
			continue
		}

		// End of row
		break
	}

	return cells, nil
}

// parseSchema parses a schema string into an AST node
func (p *Parser) parseSchema(schemaText string, pos ast.Position) (ast.SchemaNode, error) {
	schemaText = strings.TrimSpace(schemaText)

	// Boolean literals
	if schemaText == "true" {
		return ast.NewLiteralNode(true, pos), nil
	}
	if schemaText == "false" {
		return ast.NewLiteralNode(false, pos), nil
	}

	// Null literal
	if schemaText == "null" {
		return ast.NewLiteralNode(nil, pos), nil
	}

	// Number literal
	if num, err := strconv.ParseInt(schemaText, 10, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}
	if num, err := strconv.ParseFloat(schemaText, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}

	// Function call: FunctionName(args)
	if strings.Contains(schemaText, "(") && strings.HasSuffix(schemaText, ")") {
		openParen := strings.Index(schemaText, "(")
		name := schemaText[:openParen]
		argsStr := schemaText[openParen+1 : len(schemaText)-1]

		args, err := p.parseArguments(argsStr, pos)
		if err != nil {
			return nil, err
		}

		return ast.NewFunctionNode(name, args, pos), nil
	}

	// Type identifier (uppercase start)
	if len(schemaText) > 0 && schemaText[0] >= 'A' && schemaText[0] <= 'Z' {
		return ast.NewTypeNode(schemaText, pos), nil
	}

	// String literal (quoted)
	if strings.HasPrefix(schemaText, `"`) && strings.HasSuffix(schemaText, `"`) {
		value := schemaText[1 : len(schemaText)-1]
		value = strings.ReplaceAll(value, `\"`, `"`)
		value = strings.ReplaceAll(value, `\\`, `\`)
		return ast.NewLiteralNode(value, pos), nil
	}

	// Default: treat as string literal
	return ast.NewLiteralNode(schemaText, pos), nil
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

// unquoteCell removes quotes from CSV cell if quoted
func (p *Parser) unquoteCell(cell string) string {
	cell = strings.TrimSpace(cell)

	// Handle quoted cells
	if len(cell) >= 2 && cell[0] == '"' && cell[len(cell)-1] == '"' {
		cell = cell[1 : len(cell)-1]
		// Handle escaped quotes (doubled quotes in CSV)
		cell = strings.ReplaceAll(cell, `""`, `"`)
	}

	return cell
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

		// Skip whitespace (but not newlines - they're structural in CSV)
		if token.Kind() == "Whitespace" {
			continue
		}

		p.current = token
		p.hasToken = true
		return nil
	}
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
