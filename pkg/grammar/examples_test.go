package grammar_test

import (
	"os"
	"testing"

	"github.com/shapestone/shape/pkg/grammar"
)

// TestExampleBooleanGrammar verifies the boolean.ebnf example works
func TestExampleBooleanGrammar(t *testing.T) {
	content, err := os.ReadFile("examples/boolean.ebnf")
	if err != nil {
		t.Fatalf("failed to read example: %v", err)
	}

	g, err := grammar.ParseEBNF(string(content))
	if err != nil {
		t.Fatalf("failed to parse grammar: %v", err)
	}

	if len(g.Rules) == 0 {
		t.Fatal("expected grammar to have rules")
	}

	// Generate test cases
	tests := g.GenerateTests(grammar.DefaultOptions())
	if len(tests) == 0 {
		t.Fatal("expected test generation to produce tests")
	}
}

// TestExampleArithmeticGrammar verifies the arithmetic.ebnf example works
func TestExampleArithmeticGrammar(t *testing.T) {
	content, err := os.ReadFile("examples/arithmetic.ebnf")
	if err != nil {
		t.Fatalf("failed to read example: %v", err)
	}

	g, err := grammar.ParseEBNF(string(content))
	if err != nil {
		t.Fatalf("failed to parse grammar: %v", err)
	}

	if len(g.Rules) != 4 {
		t.Errorf("expected 4 rules (Expr, Term, Factor, Number), got %d", len(g.Rules))
	}
}

// TestGrammarVerificationPattern demonstrates the pattern from PARSER_IMPLEMENTATION_GUIDE
func TestGrammarVerificationPattern(t *testing.T) {
	// Use our simple boolean grammar as an example
	grammarText := `BoolExpr = "true" | "false" ;`

	spec, err := grammar.ParseEBNF(grammarText)
	if err != nil {
		t.Fatalf("failed to parse grammar: %v", err)
	}

	// Generate tests from grammar (as shown in guide)
	tests := spec.GenerateTests(grammar.TestOptions{
		MaxDepth:      5,
		CoverAllRules: true,
		EdgeCases:     true,
		InvalidCases:  true,
	})

	// Verify we got tests
	if len(tests) == 0 {
		t.Fatal("expected test generation to produce tests")
	}

	foundValid := false
	foundInvalid := false
	for _, test := range tests {
		if test.ShouldSucceed {
			foundValid = true
		} else {
			foundInvalid = true
		}
	}

	if !foundValid {
		t.Error("expected at least one valid test case")
	}
	if !foundInvalid {
		t.Error("expected at least one invalid test case (InvalidCases was enabled)")
	}
}

// TestCoveragePattern demonstrates the coverage pattern from PARSER_IMPLEMENTATION_GUIDE
func TestCoveragePattern(t *testing.T) {
	grammarText := `
		Value = Type | Literal ;
		Type = "UUID" | "Email" ;
		Literal = "null" | "true" | "false" ;
	`

	spec, err := grammar.ParseEBNF(grammarText)
	if err != nil {
		t.Fatalf("failed to parse grammar: %v", err)
	}

	tracker := grammar.NewCoverageTracker(spec)

	// Simulate parsing that exercises some rules
	tracker.RecordRule("Value")
	tracker.RecordRule("Type")

	// Check coverage
	report := tracker.Report()

	if report.TotalRules != 3 {
		t.Errorf("expected 3 total rules, got %d", report.TotalRules)
	}

	if report.CoveredRules != 2 {
		t.Errorf("expected 2 covered rules, got %d", report.CoveredRules)
	}

	if len(report.UncoveredRules) != 1 {
		t.Errorf("expected 1 uncovered rule, got %d", len(report.UncoveredRules))
	}

	// Test report formatting
	formatted := report.FormatReport()
	if formatted == "" {
		t.Error("expected non-empty formatted report")
	}
}
