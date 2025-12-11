package grammar

import "testing"

func TestGenerateTests_SimpleAlternation(t *testing.T) {
	input := `Value = "true" | "false" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tests := grammar.GenerateTests(DefaultOptions())

	if len(tests) == 0 {
		t.Fatal("expected test cases to be generated")
	}

	// Should have tests for "true" and "false"
	foundTrue := false
	foundFalse := false
	for _, test := range tests {
		if test.ShouldSucceed {
			if stringContains(test.Input, "true") {
				foundTrue = true
			}
			if stringContains(test.Input, "false") {
				foundFalse = true
			}
		}
	}

	if !foundTrue {
		t.Error("expected test for 'true'")
	}
	if !foundFalse {
		t.Error("expected test for 'false'")
	}
}

func TestGenerateTests_OptionalExpression(t *testing.T) {
	input := `Value = [ "optional" ] ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	opts := DefaultOptions()
	opts.EdgeCases = true

	tests := grammar.GenerateTests(opts)

	// Should generate empty case (omit optional) and case with optional
	foundEmpty := false
	foundOptional := false

	for _, test := range tests {
		if test.ShouldSucceed {
			if test.Input == "" {
				foundEmpty = true
			}
			if stringContains(test.Input, "optional") {
				foundOptional = true
			}
		}
	}

	if !foundEmpty {
		t.Error("expected empty test case for optional")
	}
	if !foundOptional {
		t.Error("expected test case with optional value")
	}
}

func TestGenerateTests_WithInvalidCases(t *testing.T) {
	input := `Value = "expected" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	opts := DefaultOptions()
	opts.InvalidCases = true

	tests := grammar.GenerateTests(opts)

	foundInvalid := false
	for _, test := range tests {
		if !test.ShouldSucceed {
			foundInvalid = true
			break
		}
	}

	if !foundInvalid {
		t.Error("expected invalid test cases to be generated")
	}
}

func TestGenerateTests_Sequence(t *testing.T) {
	input := `Value = "hello" "world" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tests := grammar.GenerateTests(DefaultOptions())

	if len(tests) == 0 {
		t.Fatal("expected test cases to be generated")
	}

	// Should have a test with both terminals in sequence
	foundSequence := false
	for _, test := range tests {
		if stringContains(test.Input, "hello") && stringContains(test.Input, "world") {
			foundSequence = true
			break
		}
	}

	if !foundSequence {
		t.Error("expected test case with sequence 'hello world'")
	}
}

func TestGenerateTests_Repetition(t *testing.T) {
	input := `Value = { "item" } ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	opts := DefaultOptions()
	opts.EdgeCases = true
	tests := grammar.GenerateTests(opts)

	if len(tests) == 0 {
		t.Fatal("expected test cases to be generated")
	}

	// Should have at least one test (edge cases include empty repetition)
	foundEmpty := false
	foundItem := false
	for _, test := range tests {
		if test.Input == "" {
			foundEmpty = true
		}
		if stringContains(test.Input, "item") {
			foundItem = true
		}
	}

	if !foundEmpty && !foundItem {
		t.Error("expected test cases for repetition (empty or with items)")
	}
}

func TestGenerateTests_OneOrMore(t *testing.T) {
	input := `Value = Digit+ ; Digit = "0" | "1" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tests := grammar.GenerateTests(DefaultOptions())

	if len(tests) == 0 {
		t.Fatal("expected test cases to be generated")
	}

	// Should have tests with at least one digit
	foundDigit := false
	for _, test := range tests {
		if stringContains(test.Input, "0") || stringContains(test.Input, "1") {
			foundDigit = true
			break
		}
	}

	if !foundDigit {
		t.Error("expected test cases with digits")
	}
}

func TestGenerateTests_CharClass(t *testing.T) {
	input := `Letter = [a-z] ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tests := grammar.GenerateTests(DefaultOptions())

	if len(tests) == 0 {
		t.Fatal("expected test cases to be generated")
	}

	// Should have at least one test
	if len(tests[0].Input) == 0 {
		t.Error("expected non-empty test input for character class")
	}
}

func TestGenerateTests_CoverAllRules(t *testing.T) {
	input := `
		Value = Type | Literal ;
		Type = "string" ;
		Literal = "123" ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	opts := DefaultOptions()
	opts.CoverAllRules = true
	tests := grammar.GenerateTests(opts)

	// Should have tests for all three rules
	rules := make(map[string]bool)
	for _, test := range tests {
		if len(test.RulePath) > 0 {
			rules[test.RulePath[0]] = true
		}
	}

	expectedRules := []string{"Value", "Type", "Literal"}
	for _, ruleName := range expectedRules {
		if !rules[ruleName] {
			t.Errorf("expected test for rule %q", ruleName)
		}
	}
}
