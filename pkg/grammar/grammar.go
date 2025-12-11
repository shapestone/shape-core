// Package grammar provides EBNF grammar parsing and verification infrastructure.
//
// This package enables grammar-as-verification for Shape parser projects:
//   - Parse custom EBNF variant grammars
//   - Generate verification tests from grammars
//   - Track grammar rule coverage
//   - Compare ASTs for equivalence
//
// See ADR 0005 for the complete grammar-as-verification strategy.
package grammar

import "fmt"

// Grammar represents a complete EBNF grammar specification.
type Grammar struct {
	Rules []*Rule
	// RuleMap provides fast lookup by rule name
	RuleMap map[string]*Rule
}

// Rule represents a single grammar production rule.
//
// Example EBNF:
//
//	ObjectNode = "{" [ Property { "," Property } ] "}" ;
//
// Would be represented as:
//
//	Rule{
//	    Name: "ObjectNode",
//	    Expression: Sequence{...},
//	}
type Rule struct {
	Name       string
	Expression Expression
	Comment    string // Implementation hints from EBNF comments
}

// Expression represents a grammar expression (right-hand side of a rule).
//
// Grammar expressions form a tree structure representing the EBNF syntax.
type Expression interface {
	// Type returns the expression type
	Type() ExpressionType
	// String returns EBNF representation
	String() string
}

// ExpressionType identifies the type of expression.
type ExpressionType string

const (
	TypeTerminal    ExpressionType = "Terminal"
	TypeNonTerminal ExpressionType = "NonTerminal"
	TypeSequence    ExpressionType = "Sequence"
	TypeAlternation ExpressionType = "Alternation"
	TypeOptional    ExpressionType = "Optional"
	TypeRepetition  ExpressionType = "Repetition"
	TypeOneOrMore   ExpressionType = "OneOrMore"
	TypeGrouping    ExpressionType = "Grouping"
)

// Terminal represents a terminal symbol (literal string or character class).
//
// Examples:
//   - "{"  (literal string)
//   - [a-z]  (character class)
type Terminal struct {
	Value       string
	IsCharClass bool // true for [a-z], false for "literal"
}

func (t *Terminal) Type() ExpressionType { return TypeTerminal }
func (t *Terminal) String() string {
	if t.IsCharClass {
		return t.Value // Already includes brackets
	}
	return fmt.Sprintf(`"%s"`, t.Value)
}

// NonTerminal represents a reference to another rule.
//
// Example: ObjectNode (references the ObjectNode rule)
type NonTerminal struct {
	RuleName string
}

func (n *NonTerminal) Type() ExpressionType { return TypeNonTerminal }
func (n *NonTerminal) String() string       { return n.RuleName }

// Sequence represents concatenation of expressions.
//
// Example: "hello" "world" (two terminals in sequence)
type Sequence struct {
	Elements []Expression
}

func (s *Sequence) Type() ExpressionType { return TypeSequence }
func (s *Sequence) String() string {
	result := ""
	for i, elem := range s.Elements {
		if i > 0 {
			result += " "
		}
		result += elem.String()
	}
	return result
}

// Alternation represents choice between expressions (|).
//
// Example: "true" | "false"
type Alternation struct {
	Alternatives []Expression
}

func (a *Alternation) Type() ExpressionType { return TypeAlternation }
func (a *Alternation) String() string {
	result := ""
	for i, alt := range a.Alternatives {
		if i > 0 {
			result += " | "
		}
		result += alt.String()
	}
	return result
}

// Optional represents an optional expression [ ... ].
//
// Example: [ Property ]
type Optional struct {
	Expression Expression
}

func (o *Optional) Type() ExpressionType { return TypeOptional }
func (o *Optional) String() string {
	return fmt.Sprintf("[ %s ]", o.Expression.String())
}

// Repetition represents zero or more occurrences { ... }.
//
// Example: { Property }
type Repetition struct {
	Expression Expression
}

func (r *Repetition) Type() ExpressionType { return TypeRepetition }
func (r *Repetition) String() string {
	return fmt.Sprintf("{ %s }", r.Expression.String())
}

// OneOrMore represents one or more occurrences (suffix +).
//
// Example: Digit+
type OneOrMore struct {
	Expression Expression
}

func (o *OneOrMore) Type() ExpressionType { return TypeOneOrMore }
func (o *OneOrMore) String() string {
	return fmt.Sprintf("%s+", o.Expression.String())
}

// Grouping represents grouped expression ( ... ).
//
// Example: ( "e" | "E" )
type Grouping struct {
	Expression Expression
}

func (g *Grouping) Type() ExpressionType { return TypeGrouping }
func (g *Grouping) String() string {
	return fmt.Sprintf("( %s )", g.Expression.String())
}

// GetRule returns a rule by name, or nil if not found.
func (g *Grammar) GetRule(name string) *Rule {
	return g.RuleMap[name]
}

// GetAllRules returns all rule names in the grammar.
func (g *Grammar) GetAllRules() []string {
	names := make([]string, 0, len(g.Rules))
	for _, rule := range g.Rules {
		names = append(names, rule.Name)
	}
	return names
}

// Validate checks the grammar for common errors.
func (g *Grammar) Validate() error {
	// Check for undefined rule references
	for _, rule := range g.Rules {
		if err := g.validateExpression(rule.Expression); err != nil {
			return fmt.Errorf("in rule %s: %w", rule.Name, err)
		}
	}
	return nil
}

func (g *Grammar) validateExpression(expr Expression) error {
	switch e := expr.(type) {
	case *NonTerminal:
		if g.RuleMap[e.RuleName] == nil {
			return fmt.Errorf("undefined rule: %s", e.RuleName)
		}
	case *Sequence:
		for _, elem := range e.Elements {
			if err := g.validateExpression(elem); err != nil {
				return err
			}
		}
	case *Alternation:
		for _, alt := range e.Alternatives {
			if err := g.validateExpression(alt); err != nil {
				return err
			}
		}
	case *Optional:
		return g.validateExpression(e.Expression)
	case *Repetition:
		return g.validateExpression(e.Expression)
	case *OneOrMore:
		return g.validateExpression(e.Expression)
	case *Grouping:
		return g.validateExpression(e.Expression)
	}
	return nil
}
