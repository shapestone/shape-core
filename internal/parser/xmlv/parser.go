package xmlv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/internal/tokenizer"
	"github.com/shapestone/shape/pkg/ast"
)

// Parser implements the Parser interface for XMLV format.
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// NewParser creates a new XMLV parser.
func NewParser() *Parser {
	return &Parser{}
}

// Format returns FormatXMLV.
func (p *Parser) Format() parser.Format {
	return parser.FormatXMLV
}

// Parse parses XMLV input and returns the AST.
func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
	// Initialize tokenizer
	tok := tokenizer.NewTokenizer(GetMatchers()...)
	tok.Initialize(input)
	p.tokenizer = &tok

	// Load first token
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse root element
	return p.parseElement()
}

// parseElement parses an XML element: <tag>content</tag> or <tag/>
func (p *Parser) parseElement() (ast.SchemaNode, error) {
	pos := p.position()

	// Expect opening tag
	openTag, err := p.expect(TokenTagOpen)
	if err != nil {
		return nil, err
	}

	// Extract tag name from <tagname>
	tagName := p.extractTagName(openTag.ValueString())

	// Check for self-closing tag
	if strings.HasSuffix(openTag.ValueString(), "/>") {
		// Self-closing tag represents empty object
		return ast.NewObjectNode(make(map[string]ast.SchemaNode), pos), nil
	}

	// Parse content - could be text or nested elements
	properties := make(map[string]ast.SchemaNode)
	var textContent string

	for p.hasToken {
		switch p.current.Kind() {
		case TokenTagClose:
			// End of this element
			closeTag, err := p.expect(TokenTagClose)
			if err != nil {
				return nil, err
			}

			// Verify tag names match
			closeTagName := p.extractTagName(closeTag.ValueString())
			if closeTagName != tagName {
				return nil, parser.NewSyntaxError(p.position(),
					fmt.Sprintf("mismatched tags: <%s> vs </%s>", tagName, closeTagName))
			}

			// If we have text content, parse it as validation expression
			if textContent != "" {
				return p.parseTextContent(strings.TrimSpace(textContent), pos)
			}

			// If we have properties, return as object
			if len(properties) > 0 {
				return ast.NewObjectNode(properties, pos), nil
			}

			// Empty element
			return ast.NewObjectNode(make(map[string]ast.SchemaNode), pos), nil

		case TokenTagOpen:
			// Nested element - peek at the tag name to use as property name
			childTagName := p.extractTagName(p.current.ValueString())

			// Parse the nested element
			childNode, err := p.parseElement()
			if err != nil {
				return nil, err
			}

			// Add as property
			properties[childTagName] = childNode

		case TokenText:
			// Text content
			textToken, err := p.expect(TokenText)
			if err != nil {
				return nil, err
			}
			textContent += textToken.ValueString()

		case TokenTagSelfClose:
			// Self-closing nested element
			selfCloseTag, err := p.expect(TokenTagSelfClose)
			if err != nil {
				return nil, err
			}
			childTagName := p.extractTagName(selfCloseTag.ValueString())
			properties[childTagName] = ast.NewObjectNode(make(map[string]ast.SchemaNode), p.position())

		default:
			return nil, parser.NewUnexpectedTokenError(
				p.position(),
				"tag or text",
				p.current.Kind(),
			)
		}
	}

	return nil, parser.NewUnexpectedEOFError(p.position(), fmt.Sprintf("closing tag </%s>", tagName))
}

// parseTextContent parses text content as a validation expression
func (p *Parser) parseTextContent(content string, pos ast.Position) (ast.SchemaNode, error) {
	// The content should be a validation expression: UUID, String(1,100), true, 42, etc.
	content = strings.TrimSpace(content)
	if content == "" {
		return ast.NewLiteralNode("", pos), nil
	}

	// Try to parse as different types
	// Boolean
	if content == "true" {
		return ast.NewLiteralNode(true, pos), nil
	}
	if content == "false" {
		return ast.NewLiteralNode(false, pos), nil
	}

	// Null
	if content == "null" {
		return ast.NewLiteralNode(nil, pos), nil
	}

	// Number
	if num, err := strconv.ParseInt(content, 10, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}
	if num, err := strconv.ParseFloat(content, 64); err == nil {
		return ast.NewLiteralNode(num, pos), nil
	}

	// Function call
	if strings.Contains(content, "(") && strings.HasSuffix(content, ")") {
		openParen := strings.Index(content, "(")
		name := content[:openParen]
		argsStr := content[openParen+1 : len(content)-1]

		args, err := p.parseArguments(argsStr, pos)
		if err != nil {
			return nil, err
		}

		return ast.NewFunctionNode(name, args, pos), nil
	}

	// Type identifier (must start with uppercase)
	if len(content) > 0 && content[0] >= 'A' && content[0] <= 'Z' {
		return ast.NewTypeNode(content, pos), nil
	}

	// String literal
	return ast.NewLiteralNode(content, pos), nil
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

// extractTagName extracts tag name from <tagname> or </tagname>
func (p *Parser) extractTagName(tag string) string {
	tag = strings.TrimPrefix(tag, "<")
	tag = strings.TrimPrefix(tag, "/")
	tag = strings.TrimSuffix(tag, ">")
	tag = strings.TrimSuffix(tag, "/")
	tag = strings.TrimSpace(tag)
	return tag
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

		// Skip whitespace
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
