package text

import "testing"

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
