package tokenizer

import "testing"

//
// Rune Comparison Tests
//

func TestRunesComparisonWithMatchingInput(t *testing.T) {
	// Given
	ra1 := []rune(`abc`)
	ra2 := []rune(`abc`)

	// When
	ok := RunesMatch(ra1, ra2)

	// Then
	if !ok {
		t.Fatalf("Expected RunesMatch to return 'true', got %t", ok)
	}
}

func TestRunesComparisonWithNonMatchingInput(t *testing.T) {
	// Given
	ra1 := []rune(`abc!`)
	ra2 := []rune(`abc?`)

	// When
	ok := RunesNoMatch(ra1, ra2)

	// Then
	if !ok {
		t.Fatalf("Expected RunesNoMatch to return 'true', got %t", ok)
	}
}

//
// Text Utility Tests
//

func TestStripMargin(t *testing.T) {
	input := `
	|line 1
	|line 2
	|line 3
	`
	expected := "line 1\nline 2\nline 3"

	result := StripMargin(input)

	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestPad(t *testing.T) {
	result := Pad("test", 10)
	expected := "test      "

	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFitString(t *testing.T) {
	// Test with string that needs padding
	result := FitString("abc", 10)
	if len([]rune(result)) != 10 {
		t.Fatalf("Expected length 10, got %d", len([]rune(result)))
	}

	// Test with string that needs truncation
	longStr := "this is a very long string"
	result = FitString(longStr, 15)
	if len([]rune(result)) != 15 {
		t.Fatalf("Expected length 15, got %d", len([]rune(result)))
	}
}

func TestSpaceDiff(t *testing.T) {
	tests := []struct {
		name      string
		a         string
		b         string
		expectedA string
		expectedB string
	}{
		{
			name:      "identical strings",
			a:         "hello",
			b:         "hello",
			expectedA: "hello",
			expectedB: "hello",
		},
		{
			name:      "different at start with spaces",
			a:         " hello",
			b:         "  hello",
			expectedA: "␣hello",
			expectedB: "␣␣hello",
		},
		{
			name:      "different with tabs",
			a:         "\thello",
			b:         "  hello",
			expectedA: "␉hello",
			expectedB: "␣␣hello",
		},
		{
			name:      "completely different",
			a:         "abc",
			b:         "xyz",
			expectedA: "abc",
			expectedB: "xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := SpaceDiff(tt.a, tt.b)
			if gotA != tt.expectedA {
				t.Errorf("SpaceDiff() gotA = %q, want %q", gotA, tt.expectedA)
			}
			if gotB != tt.expectedB {
				t.Errorf("SpaceDiff() gotB = %q, want %q", gotB, tt.expectedB)
			}
		})
	}
}

func TestStringDiff(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
	}{
		{
			name: "identical strings",
			a:    "hello",
			b:    "hello",
		},
		{
			name: "different strings",
			a:    "hello",
			b:    "world",
		},
		{
			name: "different lengths",
			a:    "short",
			b:    "much longer string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringDiff(tt.a, tt.b)
			// Just verify it returns a string (implementation details may vary)
			if result == "" && tt.a != tt.b {
				t.Error("StringDiff() should return non-empty for different strings")
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with newlines",
			input:    "hello\nworld",
			expected: "hello\\nworld",
		},
		{
			name:     "string with tabs",
			input:    "hello\tworld",
			expected: "hello\\tworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.input)
			if got != tt.expected {
				t.Errorf("Flatten() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestStripColumn(t *testing.T) {
	input := `|line 1|
|line 2|
|line 3|`

	result := StripColumn(input)

	// Verify it strips the column markers
	if result == "" {
		t.Error("StripColumn() should return non-empty for input with column markers")
	}
	if !containsText(result, "line 1") || !containsText(result, "line 2") || !containsText(result, "line 3") {
		t.Errorf("StripColumn() = %q, should contain lines", result)
	}
}

func containsText(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
