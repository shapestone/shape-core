package streams

import (
	"github.com/shapestone/shape/internal/text"
	"testing"
)

func TestStreamPatternMatchingWithCharMatcherShouldMatch(t *testing.T) {
	// Given
	data := "a"
	stream := NewStream(data)

	// When
	match, ok := CharMatcher('a')(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match == nil {
		t.Fatalf("Expected match, got nil")
	}
}

func TestStreamPatternMatchingWithCharMatcherShouldNotMatch(t *testing.T) {
	// Given
	data := "a"
	stream := NewStream(data)

	// When
	match, ok := CharMatcher('b')(stream)

	// Then
	if ok {
		t.Fatalf("Expected ok to be false, got %t", ok)
	}

	if match != nil {
		t.Fatalf("Expected a non-match, got %v", match)
	}
}

func TestStreamPatternMatchingWithStringMatcherShouldMatch(t *testing.T) {
	// Given
	data := "abc"
	stream := NewStream(data)

	// When
	match, ok := StringMatcher(`abc`)(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match == nil {
		t.Fatalf("Expected match, got nil")
	}
}

func TestStreamPatternMatchingWithStringMatcherShouldNotMatch(t *testing.T) {
	// Given
	data := "abc"
	stream := NewStream(data)

	// When
	match, ok := StringMatcher(`123`)(stream)

	// Then
	if ok {
		t.Fatalf("Expected ok to be false, got %t", ok)
	}

	if match != nil {
		t.Fatalf("Expected a non-match, got %v", match)
	}
}

func TestStreamPatternMatchingWithOptionalCharMatcherShouldMatch(t *testing.T) {
	// Given
	data := "a"
	stream := NewStream(data)

	// When
	match, ok := Optional(CharMatcher('a'))(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match == nil {
		t.Fatalf("Expected match, got nil")
	}
}

func TestStreamPatternMatchingWithOptionalCharMatcherShouldNotMatch(t *testing.T) {
	// Given
	data := "a"
	stream := NewStream(data)

	// When
	match, ok := Optional(CharMatcher('b'))(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match != nil {
		t.Fatalf("Expected a non-match, got %v", match)
	}
}

func TestStreamPatternMatchingWithSequenceShouldMatch(t *testing.T) {
	// Given
	data := "abc123"
	stream := NewStream(data)

	// When
	match, ok := Sequence(StringMatcher(`abc`), StringMatcher(`123`))(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match == nil {
		t.Fatalf("Expected match, got nil")
	}

	diff, tdOk := text.Diff(data, string(match))
	if !tdOk {
		t.Fatalf("Stream pattern matching validation error: \n%v", diff)
	}
}

func TestStreamPatternMatchingWithSequenceShouldNotMatch(t *testing.T) {
	// Given
	data := "abc!123"
	stream := NewStream(data)

	// When
	match, ok := Sequence(StringMatcher(`abc`), StringMatcher(`123`))(stream)

	// Then
	if ok {
		t.Fatalf("Expected ok to be false, got %t", ok)
	}

	if match != nil {
		t.Fatalf("Expected a non-match, got %v", match)
	}
}

func TestStreamPatternMatchingWithOneOfShouldMatchFirst(t *testing.T) {
	// Given
	data := "abc"
	stream := NewStream(data)

	// When
	match, ok := OneOf(
		StringMatcher(`abc`),
		StringMatcher(`123`))(stream)

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if match == nil {
		t.Fatalf("Expected match, got nil")
	}
}

func TestStreamPatternMatchingWithOneOfShouldNotMatch(t *testing.T) {
	// Given
	data := "xyz"
	stream := NewStream(data)

	// When
	match, ok := OneOf(
		StringMatcher(`abc`),
		StringMatcher(`123`))(stream)

	// Then
	if ok {
		t.Fatalf("Expected ok to be false, got %t", ok)
	}

	if match != nil {
		t.Fatalf("Expected a non-match, got %v", match)
	}
}
