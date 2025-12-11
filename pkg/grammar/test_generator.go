package grammar

import (
	"fmt"
	"strings"
)

// GenerateTests generates test cases from a grammar.
//
// This generates comprehensive test cases covering:
//   - Valid inputs (all grammar paths)
//   - Invalid inputs (grammar violations)
//   - Edge cases (empty, single, multiple elements)
//
// Example:
//
//	grammar, _ := ParseEBNF(`Value = "true" | "false" ;`)
//	tests := grammar.GenerateTests(DefaultOptions())
//	// Returns tests for: "true", "false", "invalid", etc.
func (g *Grammar) GenerateTests(options TestOptions) []TestCase {
	gen := &testGenerator{
		grammar: g,
		options: options,
		visited: make(map[string]int),
	}

	return gen.generate()
}

// testGenerator generates test cases from a grammar.
type testGenerator struct {
	grammar *Grammar
	options TestOptions
	visited map[string]int // Tracks recursion depth per rule
	tests   []TestCase
}

func (g *testGenerator) generate() []TestCase {
	g.tests = []TestCase{}

	// Generate tests for each rule if CoverAllRules is enabled
	if g.options.CoverAllRules {
		for _, rule := range g.grammar.Rules {
			g.generateForRule(rule, 0)
		}
	} else {
		// Just generate for first rule (assumed to be start rule)
		if len(g.grammar.Rules) > 0 {
			g.generateForRule(g.grammar.Rules[0], 0)
		}
	}

	return g.tests
}

func (g *testGenerator) generateForRule(rule *Rule, depth int) {
	// Check depth limit
	if depth >= g.options.MaxDepth {
		return
	}

	// Track recursion for this rule
	g.visited[rule.Name]++
	defer func() { g.visited[rule.Name]-- }()

	// Generate valid cases
	validInputs := g.generateValid(rule.Expression, depth)
	for i, input := range validInputs {
		g.tests = append(g.tests, TestCase{
			Name:          fmt.Sprintf("%s_valid_%d", rule.Name, i+1),
			Input:         input,
			ShouldSucceed: true,
			Description:   fmt.Sprintf("Valid input for rule %s", rule.Name),
			RulePath:      []string{rule.Name},
		})
	}

	// Generate invalid cases if requested
	if g.options.InvalidCases {
		invalidInputs := g.generateInvalid(rule.Expression, depth)
		for i, input := range invalidInputs {
			g.tests = append(g.tests, TestCase{
				Name:          fmt.Sprintf("%s_invalid_%d", rule.Name, i+1),
				Input:         input,
				ShouldSucceed: false,
				Description:   fmt.Sprintf("Invalid input for rule %s", rule.Name),
				RulePath:      []string{rule.Name},
			})
		}
	}
}

// generateValid generates valid inputs for an expression.
func (g *testGenerator) generateValid(expr Expression, depth int) []string {
	if depth >= g.options.MaxDepth {
		return []string{""}
	}

	switch e := expr.(type) {
	case *Terminal:
		return []string{g.generateTerminal(e)}

	case *NonTerminal:
		// Avoid infinite recursion
		if g.visited[e.RuleName] > 2 {
			return []string{""}
		}

		rule := g.grammar.RuleMap[e.RuleName]
		if rule == nil {
			return []string{""}
		}
		return g.generateValid(rule.Expression, depth+1)

	case *Sequence:
		return g.generateSequence(e, depth)

	case *Alternation:
		return g.generateAlternation(e, depth)

	case *Optional:
		results := []string{}
		// Edge case: empty (omit the optional part)
		if g.options.EdgeCases {
			results = append(results, "")
		}
		// Include the optional part
		inner := g.generateValid(e.Expression, depth)
		results = append(results, inner...)
		return results

	case *Repetition:
		results := []string{}
		// Edge case: empty (zero occurrences)
		if g.options.EdgeCases {
			results = append(results, "")
		}
		// Single occurrence
		single := g.generateValid(e.Expression, depth)
		results = append(results, single...)
		// Multiple occurrences (if edge cases enabled)
		if g.options.EdgeCases {
			multiple := g.generateRepetition(e.Expression, depth, 3)
			results = append(results, multiple...)
		}
		return results

	case *OneOrMore:
		results := []string{}
		// Single occurrence (minimum)
		single := g.generateValid(e.Expression, depth)
		results = append(results, single...)
		// Multiple occurrences
		if g.options.EdgeCases {
			multiple := g.generateRepetition(e.Expression, depth, 3)
			results = append(results, multiple...)
		}
		return results

	case *Grouping:
		return g.generateValid(e.Expression, depth)

	default:
		return []string{""}
	}
}

// generateInvalid generates invalid inputs for an expression.
func (g *testGenerator) generateInvalid(expr Expression, depth int) []string {
	if depth >= g.options.MaxDepth {
		return []string{}
	}

	switch e := expr.(type) {
	case *Terminal:
		// For terminals, generate wrong values
		if e.IsCharClass {
			return []string{"!"} // Invalid character
		}
		return []string{"WRONG"} // Wrong literal

	case *Sequence:
		// Missing elements in sequence
		if len(e.Elements) > 1 {
			// Generate first element only (missing rest)
			first := g.generateValid(e.Elements[0], depth)
			return first
		}
		return []string{}

	case *Alternation:
		// Something that matches none of the alternatives
		return []string{"INVALID_ALTERNATIVE"}

	case *Optional:
		// Optional can't really be invalid (it's optional)
		return []string{}

	case *Repetition:
		// Repetition can't be invalid (zero is allowed)
		return []string{}

	case *OneOrMore:
		// Empty is invalid for one-or-more
		return []string{""}

	default:
		return []string{}
	}
}

// Helper methods

func (g *testGenerator) generateTerminal(t *Terminal) string {
	if t.IsCharClass {
		// For character classes like [a-z], generate a single matching character
		// Simple heuristic: extract first character from range
		value := strings.Trim(t.Value, "[]")
		if len(value) > 0 {
			// Handle ranges like "a-z"
			if len(value) >= 3 && value[1] == '-' {
				return string(value[0]) // Return first char of range
			}
			return string(value[0])
		}
		return "a" // Fallback
	}
	return t.Value
}

func (g *testGenerator) generateSequence(seq *Sequence, depth int) []string {
	if len(seq.Elements) == 0 {
		return []string{""}
	}

	// Generate all elements and concatenate
	result := []string{""}

	for _, elem := range seq.Elements {
		elemResults := g.generateValid(elem, depth)
		if len(elemResults) == 0 {
			continue
		}

		// Concatenate each result with each element result
		newResults := []string{}
		for _, r := range result {
			for _, e := range elemResults {
				combined := r
				if r != "" && e != "" {
					combined += " "
				}
				combined += e
				newResults = append(newResults, combined)
			}
		}
		result = newResults

		// Limit combinatorial explosion
		if len(result) > 10 {
			result = result[:10]
		}
	}

	return result
}

func (g *testGenerator) generateAlternation(alt *Alternation, depth int) []string {
	results := []string{}

	maxAlts := len(alt.Alternatives)
	if g.options.MaxAlternatives > 0 && g.options.MaxAlternatives < maxAlts {
		maxAlts = g.options.MaxAlternatives
	}

	// Generate for each alternative (up to limit)
	for i := 0; i < maxAlts && i < len(alt.Alternatives); i++ {
		altResults := g.generateValid(alt.Alternatives[i], depth)
		results = append(results, altResults...)
	}

	return results
}

func (g *testGenerator) generateRepetition(expr Expression, depth int, count int) []string {
	// Generate multiple occurrences of an expression
	results := g.generateValid(expr, depth)
	if len(results) == 0 {
		return []string{""}
	}

	// Pick first result and repeat it
	repeated := []string{}
	for i := 0; i < count; i++ {
		parts := make([]string, count)
		for j := 0; j < count; j++ {
			parts[j] = results[0]
		}
		repeated = append(repeated, strings.Join(parts, " "))
	}

	return repeated[:1] // Return one repeated case
}
