package grammar

import (
	"fmt"
	"strings"

	"github.com/shapestone/shape-core/pkg/tokenizer"
)

// Token types for EBNF parsing
const (
	TokenIdentifier = "Identifier"
	TokenString     = "String"
	TokenCharClass  = "CharClass"
	TokenEquals     = "Equals"
	TokenSemicolon  = "Semicolon"
	TokenPipe       = "Pipe"
	TokenLBracket   = "LBracket"
	TokenRBracket   = "RBracket"
	TokenLBrace     = "LBrace"
	TokenRBrace     = "RBrace"
	TokenLParen     = "LParen"
	TokenRParen     = "RParen"
	TokenPlus       = "Plus"
	TokenStar       = "Star"
	TokenComment    = "Comment"
	TokenEOF        = "EOF"
)

// ParseEBNF parses an EBNF grammar string and returns a Grammar.
//
// Supports the custom EBNF variant defined in ADR 0005:
//   - rule_name = expression ;
//   - [ ] optional
//   - { } zero or more
//   - + one or more (suffix)
//   - * zero or more (suffix)
//   - | alternation
//   - ( ) grouping
//   - "literal" string literals
//   - [a-z] character classes
//   - // comments
//
// Example:
//
//	grammar := `
//	  // Object with properties
//	  ObjectNode = "{" [ Property { "," Property } ] "}" ;
//	  Property = StringLiteral ":" Value ;
//	`
//	g, err := grammar.ParseEBNF(grammar)
func ParseEBNF(input string) (*Grammar, error) {
	tok := newEBNFTokenizer()
	tok.Initialize(input)

	p := &ebnfParser{
		tokenizer: tok,
	}

	if !p.advance() {
		return nil, fmt.Errorf("failed to read first token")
	}

	return p.parseGrammar()
}

// ebnfParser implements LL(1) recursive descent parsing for EBNF.
type ebnfParser struct {
	tokenizer tokenizer.Tokenizer
	current   *tokenizer.Token
	hasToken  bool
}

// parseGrammar parses the entire grammar.
//
// Grammar:
//
//	Grammar = { Rule } ;
func (p *ebnfParser) parseGrammar() (*Grammar, error) {
	rules := []*Rule{}
	ruleMap := make(map[string]*Rule)

	// Parse rules until EOF
	for p.peek() != nil {
		// parseRule will handle comments before each rule
		rule, err := p.parseRule()
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
		ruleMap[rule.Name] = rule
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("grammar is empty: no rules found")
	}

	g := &Grammar{
		Rules:   rules,
		RuleMap: ruleMap,
	}

	// Note: Validation is NOT performed automatically to allow grammar fragments
	// (common in examples/docs). Call grammar.Validate() explicitly if needed.
	return g, nil
}

// parseRule parses a single grammar rule.
//
// Grammar:
//
//	Rule = Identifier "=" Expression ";" ;
func (p *ebnfParser) parseRule() (*Rule, error) {
	// Collect comments before rule
	comment := ""
	for p.peek() != nil && p.peek().Kind() == TokenComment {
		commentText := string(p.peek().Value())
		// Strip // prefix
		commentText = strings.TrimPrefix(commentText, "//")
		commentText = strings.TrimSpace(commentText)
		if comment != "" {
			comment += "\n"
		}
		comment += commentText
		p.advance()
	}

	// Identifier
	nameToken, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, fmt.Errorf("expected rule name: %w", err)
	}
	name := string(nameToken.Value())

	// "="
	if _, err := p.expect(TokenEquals); err != nil {
		return nil, fmt.Errorf("in rule %s: %w", name, err)
	}

	// Expression
	expr, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("in rule %s: %w", name, err)
	}

	// ";"
	if _, err := p.expect(TokenSemicolon); err != nil {
		return nil, fmt.Errorf("in rule %s: expected ';' at end of rule: %w", name, err)
	}

	return &Rule{
		Name:       name,
		Expression: expr,
		Comment:    comment,
	}, nil
}

// parseExpression parses an expression (right-hand side of rule).
//
// Grammar:
//
//	Expression = Sequence { "|" Sequence } ;
func (p *ebnfParser) parseExpression() (Expression, error) {
	// First sequence
	first, err := p.parseSequence()
	if err != nil {
		return nil, err
	}

	// Check for alternation (|)
	if p.peek() == nil || p.peek().Kind() != TokenPipe {
		return first, nil
	}

	// Multiple alternatives
	alternatives := []Expression{first}
	for p.peek() != nil && p.peek().Kind() == TokenPipe {
		p.advance() // consume |

		alt, err := p.parseSequence()
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return &Alternation{Alternatives: alternatives}, nil
}

// parseSequence parses a sequence of terms.
//
// Grammar:
//
//	Sequence = Term { Term } ;
func (p *ebnfParser) parseSequence() (Expression, error) {
	elements := []Expression{}

	// Parse first term
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	elements = append(elements, term)

	// Parse additional terms
	for p.isTermStart() {
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		elements = append(elements, term)
	}

	if len(elements) == 1 {
		return elements[0], nil
	}

	return &Sequence{Elements: elements}, nil
}

// parseTerm parses a single term with optional suffix.
//
// Grammar:
//
//	Term = Factor [ "+" | "*" ] ;
func (p *ebnfParser) parseTerm() (Expression, error) {
	factor, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	// Check for suffix operators
	if p.peek() != nil {
		switch p.peek().Kind() {
		case TokenPlus: // +
			p.advance()
			return &OneOrMore{Expression: factor}, nil
		case TokenStar: // *
			p.advance()
			return &Repetition{Expression: factor}, nil
		}
	}

	return factor, nil
}

// parseFactor parses a factor (primary expression).
//
// Grammar:
//
//	Factor = Terminal | NonTerminal | Optional | Repetition | Grouping ;
//	Optional   = "[" Expression "]" ;
//	Repetition = "{" Expression "}" ;
//	Grouping   = "(" Expression ")" ;
func (p *ebnfParser) parseFactor() (Expression, error) {
	if p.peek() == nil {
		return nil, fmt.Errorf("unexpected end of input: expected terminal, identifier, or grouping")
	}

	switch p.peek().Kind() {
	case TokenString:
		// Terminal string literal
		token := p.current
		p.advance()
		// Remove quotes
		value := string(token.Value())
		value = strings.Trim(value, `"`)
		return &Terminal{Value: value, IsCharClass: false}, nil

	case TokenCharClass:
		// Terminal character class [a-z]
		token := p.current
		p.advance()
		return &Terminal{Value: string(token.Value()), IsCharClass: true}, nil

	case TokenIdentifier:
		// Non-terminal (rule reference)
		token := p.current
		p.advance()
		return &NonTerminal{RuleName: string(token.Value())}, nil

	case TokenLBracket:
		// Optional [ ... ]
		p.advance() // consume [
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRBracket); err != nil {
			return nil, err
		}
		return &Optional{Expression: expr}, nil

	case TokenLBrace:
		// Repetition { ... }
		p.advance() // consume {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRBrace); err != nil {
			return nil, err
		}
		return &Repetition{Expression: expr}, nil

	case TokenLParen:
		// Grouping ( ... )
		p.advance() // consume (
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		return &Grouping{Expression: expr}, nil

	default:
		return nil, fmt.Errorf("unexpected token %s at position (%d:%d): expected terminal, identifier, or grouping",
			p.peek().Kind(), p.peek().Row(), p.peek().Column())
	}
}

// isTermStart returns true if current token can start a term.
func (p *ebnfParser) isTermStart() bool {
	if p.peek() == nil {
		return false
	}
	kind := p.peek().Kind()
	return kind == TokenString ||
		kind == TokenCharClass ||
		kind == TokenIdentifier ||
		kind == TokenLBracket ||
		kind == TokenLBrace ||
		kind == TokenLParen
}

// Helper methods

func (p *ebnfParser) peek() *tokenizer.Token {
	if p.hasToken {
		return p.current
	}
	return nil
}

func (p *ebnfParser) advance() bool {
	for {
		token, ok := p.tokenizer.NextToken()
		if !ok {
			p.hasToken = false
			return false
		}
		// Skip whitespace tokens only (not comments - those are handled separately)
		if token.Kind() == "Whitespace" {
			continue
		}
		p.current = token
		p.hasToken = true
		return true
	}
}

func (p *ebnfParser) expect(kind string) (*tokenizer.Token, error) {
	if p.peek() == nil {
		return nil, fmt.Errorf("expected %s, got EOF", kind)
	}
	if p.peek().Kind() != kind {
		return nil, fmt.Errorf("expected %s, got %s at position (%d:%d)",
			kind, p.peek().Kind(), p.peek().Row(), p.peek().Column())
	}
	token := p.current
	p.advance()
	return token, nil
}

// newEBNFTokenizer creates a tokenizer for EBNF syntax.
func newEBNFTokenizer() tokenizer.Tokenizer {
	return tokenizer.NewTokenizer(
		// Comments
		commentMatcher(),

		// Character classes [a-z]
		charClassMatcher(),

		// String literals "..."
		stringLiteralMatcher(),

		// Structural tokens (order matters: longer strings first to avoid ambiguity)
		tokenizer.StringMatcherFunc(TokenEquals, "="),
		tokenizer.StringMatcherFunc(TokenSemicolon, ";"),
		tokenizer.StringMatcherFunc(TokenPipe, "|"),
		tokenizer.StringMatcherFunc(TokenLBracket, "["),
		tokenizer.StringMatcherFunc(TokenRBracket, "]"),
		tokenizer.StringMatcherFunc(TokenLBrace, "{"),
		tokenizer.StringMatcherFunc(TokenRBrace, "}"),
		tokenizer.StringMatcherFunc(TokenLParen, "("),
		tokenizer.StringMatcherFunc(TokenRParen, ")"),
		tokenizer.StringMatcherFunc(TokenPlus, "+"),
		tokenizer.StringMatcherFunc(TokenStar, "*"),

		// Identifiers (rule names and non-terminals)
		identifierMatcher(),
	)
}

// commentMatcher matches comments starting with //
func commentMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		// Check for //
		r1, ok1 := stream.NextChar()
		if !ok1 || r1 != '/' {
			return nil
		}
		r2, ok2 := stream.NextChar()
		if !ok2 || r2 != '/' {
			return nil
		}

		value := []rune{'/', '/'}
		// Read until newline
		for {
			r, ok := stream.NextChar()
			if !ok || r == '\n' {
				break
			}
			value = append(value, r)
		}
		return tokenizer.NewToken(TokenComment, value)
	}
}

// charClassMatcher matches character classes like [a-z] or [0-9]
// Must not match optional syntax like [ "expression" ]
func charClassMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.NextChar()
		if !ok || r != '[' {
			return nil
		}

		// Peek ahead to distinguish [a-z] from [ expr ]
		// Character classes don't have spaces after [
		r2, ok2 := stream.NextChar()
		if !ok2 {
			return nil
		}

		// If there's whitespace after [, it's optional syntax, not a char class
		if r2 == ' ' || r2 == '\t' || r2 == '\n' {
			return nil
		}

		value := []rune{'[', r2}
		// Read until ]
		for {
			r, ok := stream.NextChar()
			if !ok {
				return nil // Unclosed bracket
			}
			value = append(value, r)
			if r == ']' {
				// Validate it looks like a character class (contains - or is short)
				// This prevents matching things like ["string"]
				str := string(value)
				if len(str) <= 20 && !stringContainsQuote(str) {
					return tokenizer.NewToken(TokenCharClass, value)
				}
				return nil
			}
		}
	}
}

func stringContainsQuote(s string) bool {
	return strings.Contains(s, "\"") || strings.Contains(s, "'")
}

// stringLiteralMatcher matches string literals "..."
func stringLiteralMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.NextChar()
		if !ok || r != '"' {
			return nil
		}

		value := []rune{'"'}
		escaped := false
		// Read until closing "
		for {
			r, ok := stream.NextChar()
			if !ok {
				return nil // Unclosed string
			}
			value = append(value, r)
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				return tokenizer.NewToken(TokenString, value)
			}
		}
	}
}

// identifierMatcher matches identifiers [a-zA-Z_][a-zA-Z0-9_]*
func identifierMatcher() tokenizer.Matcher {
	return func(stream tokenizer.Stream) *tokenizer.Token {
		r, ok := stream.NextChar()
		if !ok {
			return nil
		}

		// First character must be letter or underscore
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
			return nil
		}

		value := []rune{r}
		// Subsequent characters can be letter, digit, or underscore
		for {
			r, ok := stream.NextChar()
			if !ok {
				break
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				value = append(value, r)
			} else {
				break
			}
		}
		return tokenizer.NewToken(TokenIdentifier, value)
	}
}
