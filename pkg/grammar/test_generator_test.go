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
