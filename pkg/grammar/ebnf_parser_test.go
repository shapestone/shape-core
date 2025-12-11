package grammar

import (
	"testing"
)

func TestParseEBNF_SimpleRule(t *testing.T) {
	input := `Value = "true" | "false" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grammar.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(grammar.Rules))
	}

	rule := grammar.Rules[0]
	if rule.Name != "Value" {
		t.Errorf("expected rule name 'Value', got %s", rule.Name)
	}

	// Check that it's an alternation
	alt, ok := rule.Expression.(*Alternation)
	if !ok {
		t.Fatalf("expected Alternation, got %T", rule.Expression)
	}

	if len(alt.Alternatives) != 2 {
		t.Errorf("expected 2 alternatives, got %d", len(alt.Alternatives))
	}
}

func TestParseEBNF_SequenceRule(t *testing.T) {
	input := `ObjectNode = "{" Property "}" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	seq, ok := rule.Expression.(*Sequence)
	if !ok {
		t.Fatalf("expected Sequence, got %T", rule.Expression)
	}

	if len(seq.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(seq.Elements))
	}
}

func TestParseEBNF_OptionalExpression(t *testing.T) {
	input := `Value = [ "optional" ] ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	opt, ok := rule.Expression.(*Optional)
	if !ok {
		t.Fatalf("expected Optional, got %T", rule.Expression)
	}

	terminal, ok := opt.Expression.(*Terminal)
	if !ok {
		t.Fatalf("expected Terminal inside Optional, got %T", opt.Expression)
	}

	if terminal.Value != "optional" {
		t.Errorf("expected terminal value 'optional', got %s", terminal.Value)
	}
}

func TestParseEBNF_RepetitionExpression(t *testing.T) {
	input := `List = { Item } ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	rep, ok := rule.Expression.(*Repetition)
	if !ok {
		t.Fatalf("expected Repetition, got %T", rule.Expression)
	}

	nonTerm, ok := rep.Expression.(*NonTerminal)
	if !ok {
		t.Fatalf("expected NonTerminal inside Repetition, got %T", rep.Expression)
	}

	if nonTerm.RuleName != "Item" {
		t.Errorf("expected rule name 'Item', got %s", nonTerm.RuleName)
	}
}

func TestParseEBNF_OneOrMoreSuffix(t *testing.T) {
	input := `Digits = Digit+ ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	oneOrMore, ok := rule.Expression.(*OneOrMore)
	if !ok {
		t.Fatalf("expected OneOrMore, got %T", rule.Expression)
	}

	nonTerm, ok := oneOrMore.Expression.(*NonTerminal)
	if !ok {
		t.Fatalf("expected NonTerminal, got %T", oneOrMore.Expression)
	}

	if nonTerm.RuleName != "Digit" {
		t.Errorf("expected 'Digit', got %s", nonTerm.RuleName)
	}
}

func TestParseEBNF_RepetitionSuffix(t *testing.T) {
	input := `Digits = Digit* ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	rep, ok := rule.Expression.(*Repetition)
	if !ok {
		t.Fatalf("expected Repetition, got %T", rule.Expression)
	}

	nonTerm, ok := rep.Expression.(*NonTerminal)
	if !ok {
		t.Fatalf("expected NonTerminal, got %T", rep.Expression)
	}

	if nonTerm.RuleName != "Digit" {
		t.Errorf("expected 'Digit', got %s", nonTerm.RuleName)
	}
}

func TestParseEBNF_CharacterClass(t *testing.T) {
	input := `Digit = [0-9] ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	terminal, ok := rule.Expression.(*Terminal)
	if !ok {
		t.Fatalf("expected Terminal, got %T", rule.Expression)
	}

	if !terminal.IsCharClass {
		t.Error("expected IsCharClass to be true")
	}

	if terminal.Value != "[0-9]" {
		t.Errorf("expected '[0-9]', got %s", terminal.Value)
	}
}

func TestParseEBNF_MultipleRules(t *testing.T) {
	input := `
		Value = Type | Literal ;
		Type = Identifier ;
		Literal = String | Number ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grammar.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(grammar.Rules))
	}

	expectedNames := []string{"Value", "Type", "Literal"}
	for i, expected := range expectedNames {
		if grammar.Rules[i].Name != expected {
			t.Errorf("rule %d: expected name %s, got %s",
				i, expected, grammar.Rules[i].Name)
		}
	}
}

func TestParseEBNF_Comments(t *testing.T) {
	input := `
		// This is a comment about Value
		// Another comment line
		Value = "true" | "false" ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := grammar.Rules[0]
	if rule.Comment == "" {
		t.Error("expected comment to be captured")
	}

	if !contains(rule.Comment, "This is a comment") {
		t.Errorf("expected comment text, got: %s", rule.Comment)
	}
}

func TestParseEBNF_ComplexExample(t *testing.T) {
	input := `
		// Object with properties
		ObjectNode = "{" [ Property { "," Property } ] "}" ;
		Property = StringLiteral ":" Value ;
		Value = Literal | Type | Function ;
		StringLiteral = "\"" "string" "\"" ;
		Literal = "true" | "false" | "null" ;
		Type = "String" | "Number" | "Boolean" ;
		Function = "func" "(" [ Value { "," Value } ] ")" ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grammar.Rules) != 7 {
		t.Fatalf("expected 7 rules, got %d", len(grammar.Rules))
	}

	// Verify grammar is valid (no undefined references)
	if err := grammar.Validate(); err != nil {
		t.Errorf("grammar validation failed: %v", err)
	}
}

func TestParseEBNF_Error_MissingEquals(t *testing.T) {
	input := `Value "true" ;`

	_, err := ParseEBNF(input)
	if err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestParseEBNF_Error_MissingSemicolon(t *testing.T) {
	input := `Value = "true"`

	_, err := ParseEBNF(input)
	if err == nil {
		t.Fatal("expected error for missing ';'")
	}
}

func TestParseEBNF_Error_EmptyGrammar(t *testing.T) {
	input := ``

	_, err := ParseEBNF(input)
	if err == nil {
		t.Fatal("expected error for empty grammar")
	}
}

func TestGrammar_Validate_UndefinedRule(t *testing.T) {
	// Create grammar with undefined rule reference
	grammar := &Grammar{
		Rules: []*Rule{
			{
				Name:       "Value",
				Expression: &NonTerminal{RuleName: "UndefinedRule"},
			},
		},
		RuleMap: map[string]*Rule{
			"Value": {
				Name:       "Value",
				Expression: &NonTerminal{RuleName: "UndefinedRule"},
			},
		},
	}

	err := grammar.Validate()
	if err == nil {
		t.Fatal("expected validation error for undefined rule")
	}

	if !contains(err.Error(), "undefined rule") {
		t.Errorf("expected 'undefined rule' error, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
