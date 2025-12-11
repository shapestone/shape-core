package grammar

import (
	"testing"
)

func TestTerminal_Type(t *testing.T) {
	term := &Terminal{Value: "test", IsCharClass: false}
	if term.Type() != TypeTerminal {
		t.Errorf("Terminal.Type() = %v, want %v", term.Type(), TypeTerminal)
	}
}

func TestTerminal_String(t *testing.T) {
	tests := []struct {
		name     string
		terminal *Terminal
		expected string
	}{
		{
			name:     "literal string",
			terminal: &Terminal{Value: "hello", IsCharClass: false},
			expected: `"hello"`,
		},
		{
			name:     "character class",
			terminal: &Terminal{Value: "[a-z]", IsCharClass: true},
			expected: "[a-z]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.terminal.String()
			if got != tt.expected {
				t.Errorf("Terminal.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNonTerminal_Type(t *testing.T) {
	nt := &NonTerminal{RuleName: "TestRule"}
	if nt.Type() != TypeNonTerminal {
		t.Errorf("NonTerminal.Type() = %v, want %v", nt.Type(), TypeNonTerminal)
	}
}

func TestNonTerminal_String(t *testing.T) {
	nt := &NonTerminal{RuleName: "ObjectNode"}
	expected := "ObjectNode"
	if nt.String() != expected {
		t.Errorf("NonTerminal.String() = %q, want %q", nt.String(), expected)
	}
}

func TestSequence_Type(t *testing.T) {
	seq := &Sequence{Elements: []Expression{}}
	if seq.Type() != TypeSequence {
		t.Errorf("Sequence.Type() = %v, want %v", seq.Type(), TypeSequence)
	}
}

func TestSequence_String(t *testing.T) {
	seq := &Sequence{
		Elements: []Expression{
			&Terminal{Value: "hello", IsCharClass: false},
			&Terminal{Value: "world", IsCharClass: false},
		},
	}
	expected := `"hello" "world"`
	if seq.String() != expected {
		t.Errorf("Sequence.String() = %q, want %q", seq.String(), expected)
	}
}

func TestAlternation_Type(t *testing.T) {
	alt := &Alternation{Alternatives: []Expression{}}
	if alt.Type() != TypeAlternation {
		t.Errorf("Alternation.Type() = %v, want %v", alt.Type(), TypeAlternation)
	}
}

func TestAlternation_String(t *testing.T) {
	alt := &Alternation{
		Alternatives: []Expression{
			&Terminal{Value: "true", IsCharClass: false},
			&Terminal{Value: "false", IsCharClass: false},
		},
	}
	expected := `"true" | "false"`
	if alt.String() != expected {
		t.Errorf("Alternation.String() = %q, want %q", alt.String(), expected)
	}
}

func TestOptional_Type(t *testing.T) {
	opt := &Optional{Expression: &Terminal{Value: "test", IsCharClass: false}}
	if opt.Type() != TypeOptional {
		t.Errorf("Optional.Type() = %v, want %v", opt.Type(), TypeOptional)
	}
}

func TestOptional_String(t *testing.T) {
	opt := &Optional{
		Expression: &Terminal{Value: "property", IsCharClass: false},
	}
	expected := `[ "property" ]`
	if opt.String() != expected {
		t.Errorf("Optional.String() = %q, want %q", opt.String(), expected)
	}
}

func TestRepetition_Type(t *testing.T) {
	rep := &Repetition{Expression: &Terminal{Value: "test", IsCharClass: false}}
	if rep.Type() != TypeRepetition {
		t.Errorf("Repetition.Type() = %v, want %v", rep.Type(), TypeRepetition)
	}
}

func TestRepetition_String(t *testing.T) {
	rep := &Repetition{
		Expression: &Terminal{Value: "property", IsCharClass: false},
	}
	expected := `{ "property" }`
	if rep.String() != expected {
		t.Errorf("Repetition.String() = %q, want %q", rep.String(), expected)
	}
}

func TestOneOrMore_Type(t *testing.T) {
	oom := &OneOrMore{Expression: &Terminal{Value: "test", IsCharClass: false}}
	if oom.Type() != TypeOneOrMore {
		t.Errorf("OneOrMore.Type() = %v, want %v", oom.Type(), TypeOneOrMore)
	}
}

func TestOneOrMore_String(t *testing.T) {
	oom := &OneOrMore{
		Expression: &NonTerminal{RuleName: "Digit"},
	}
	expected := "Digit+"
	if oom.String() != expected {
		t.Errorf("OneOrMore.String() = %q, want %q", oom.String(), expected)
	}
}

func TestGrouping_Type(t *testing.T) {
	grp := &Grouping{Expression: &Terminal{Value: "test", IsCharClass: false}}
	if grp.Type() != TypeGrouping {
		t.Errorf("Grouping.Type() = %v, want %v", grp.Type(), TypeGrouping)
	}
}

func TestGrouping_String(t *testing.T) {
	grp := &Grouping{
		Expression: &Alternation{
			Alternatives: []Expression{
				&Terminal{Value: "e", IsCharClass: false},
				&Terminal{Value: "E", IsCharClass: false},
			},
		},
	}
	expected := `( "e" | "E" )`
	if grp.String() != expected {
		t.Errorf("Grouping.String() = %q, want %q", grp.String(), expected)
	}
}

func TestGrammar_GetRule(t *testing.T) {
	grammar := &Grammar{
		Rules: []*Rule{
			{Name: "Rule1", Expression: &Terminal{Value: "test", IsCharClass: false}},
			{Name: "Rule2", Expression: &Terminal{Value: "test2", IsCharClass: false}},
		},
		RuleMap: map[string]*Rule{
			"Rule1": {Name: "Rule1", Expression: &Terminal{Value: "test", IsCharClass: false}},
			"Rule2": {Name: "Rule2", Expression: &Terminal{Value: "test2", IsCharClass: false}},
		},
	}

	tests := []struct {
		name     string
		ruleName string
		wantNil  bool
	}{
		{
			name:     "existing rule",
			ruleName: "Rule1",
			wantNil:  false,
		},
		{
			name:     "non-existent rule",
			ruleName: "NonExistent",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grammar.GetRule(tt.ruleName)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetRule(%q) = %v, wantNil = %v", tt.ruleName, got, tt.wantNil)
			}
			if !tt.wantNil && got.Name != tt.ruleName {
				t.Errorf("GetRule(%q).Name = %q, want %q", tt.ruleName, got.Name, tt.ruleName)
			}
		})
	}
}

func TestGrammar_GetAllRules(t *testing.T) {
	grammar := &Grammar{
		Rules: []*Rule{
			{Name: "Rule1", Expression: &Terminal{Value: "test", IsCharClass: false}},
			{Name: "Rule2", Expression: &Terminal{Value: "test2", IsCharClass: false}},
			{Name: "Rule3", Expression: &Terminal{Value: "test3", IsCharClass: false}},
		},
		RuleMap: map[string]*Rule{},
	}

	allRules := grammar.GetAllRules()

	if len(allRules) != 3 {
		t.Errorf("GetAllRules() returned %d rules, want 3", len(allRules))
	}

	// Check that all expected rule names are present
	ruleMap := make(map[string]bool)
	for _, name := range allRules {
		ruleMap[name] = true
	}

	expectedRules := []string{"Rule1", "Rule2", "Rule3"}
	for _, expected := range expectedRules {
		if !ruleMap[expected] {
			t.Errorf("GetAllRules() missing expected rule %q", expected)
		}
	}
}
