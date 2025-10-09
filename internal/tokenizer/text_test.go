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
