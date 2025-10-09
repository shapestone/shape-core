package tokenizer

import (
	"testing"
)

//
// Stream Tests
//

func TestInitializerStream(t *testing.T) {
	// Given
	data := "abc123"

	// When
	stream := NewStream(data)

	// Then
	if stream.IsEos() != false {
		t.Fatalf("Expected IsEos() to be false, got %t", stream.IsEos())
	}

	if stream.GetRow() != 1 {
		t.Fatalf("Expected GetRow() to be 1, got %d", stream.GetRow())
	}

	if stream.GetColumn() != 1 {
		t.Fatalf("Expected GetColumn() to be 1, got %d", stream.GetColumn())
	}
}

func TestStreamWithPeekChar(t *testing.T) {
	// Given
	data := "a1"

	// When
	stream := NewStream(data)
	ch, ok := stream.PeekChar()

	// Then
	if !ok {
		t.Fatalf("Expected peekChar() to be true, got %t", ok)
	}

	if ch != 'a' {
		t.Fatalf("Expected 'a', got '%c'", ch)
	}

	// When - a repeated PeekChar() should read the same data again
	ch2, ok2 := stream.PeekChar()

	// Then
	if !ok2 {
		t.Fatalf("Expected peekChar() to be true, got %t", ok2)
	}

	if ch2 != 'a' {
		t.Fatalf("Expected 'a', got '%c'", ch2)
	}
}

func TestStreamWithPeekCharWithEndOfStream(t *testing.T) {
	// Given
	data := ""

	// When
	stream := NewStream(data)
	eos := stream.IsEos()
	ch, ok := stream.PeekChar()

	// Then
	if !eos {
		t.Fatalf("Expected end-of-stream to be true, got %t", eos)
	}

	if ok {
		t.Fatalf("Expected ok to be false, got %t", ok)
	}

	if ch != 0 {
		t.Fatalf("Expected rune to be 0, got %d", ch)
	}
}

func TestStreamWithNextChar(t *testing.T) {
	// Given
	data := "a"

	// When
	stream := NewStream(data)
	eos := stream.IsEos()
	ch, ok := stream.NextChar()

	// Then
	if eos {
		t.Fatalf("Expected end-of-stream to be false, got %t", eos)
	}

	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if ch != 'a' {
		t.Fatalf("Expected 'a', got '%c'", ch)
	}

	// When
	eos2 := stream.IsEos()
	ch2, ok2 := stream.NextChar()

	// Then
	if !eos2 {
		t.Fatalf("Expected end-of-stream to be true, got %t", eos2)
	}

	if ok2 {
		t.Fatalf("Expected ok to be false, got %t", ok2)
	}

	if ch2 != 0 {
		t.Fatalf("Expected rune to be 0, got %d", ch2)
	}
}

func TestStreamWithNextCharWithEndOfStream(t *testing.T) {
	// Given
	data := ""

	// When
	stream := NewStream(data)
	eos := stream.IsEos()

	// Then
	if !eos {
		t.Fatalf("Expected end-of-stream to be true, got %t", eos)
	}
}

func TestStreamRowAndColumnTrackingAtMiddleOfContent(t *testing.T) {
	// Given
	data := "abc\n123\n!@#"

	// When
	var r rune
	var ok bool
	stream := NewStream(data)
	for i := 0; i < len([]rune("abc\n12")); i++ {
		r, ok = stream.NextChar()
	}

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if r != '2' {
		t.Fatalf("Expected '2', got '%c'", r)
	}

	if stream.GetColumn() != 3 {
		t.Fatalf("Expected GetColumn() to be 3, got %d", stream.GetColumn())
	}

	if stream.GetRow() != 2 {
		t.Fatalf("Expected GetRow() to be 2, got %d", stream.GetRow())
	}
}

func TestStreamRowAndColumnTrackingAtEndOfContent(t *testing.T) {
	// Given
	data := "abc\n123\n!"

	// When
	var r rune
	var ok bool
	stream := NewStream(data)
	for i := 0; i < len([]rune(data)); i++ {
		r, ok = stream.NextChar()
	}

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if r != '!' {
		t.Fatalf("Expected '!', got '%c'", r)
	}

	if stream.GetColumn() != 2 {
		t.Fatalf("Expected GetColumn() to be 2, got %d", stream.GetColumn())
	}

	if stream.GetRow() != 3 {
		t.Fatalf("Expected GetRow() to be 3, got %d", stream.GetRow())
	}
}

func TestStreamConfirmWithMatchingRuneSequence(t *testing.T) {
	// Given
	data := "abc123"

	// When
	stream := NewStream(data)
	ok := stream.MatchChars([]rune("abc"))
	offset := stream.GetOffset()
	next, _ := stream.PeekChar()

	// Then
	if !ok {
		t.Fatalf("Expected a match, but got: %t", ok)
	}

	if offset != 3 {
		t.Fatalf("Expected offset to be 3, got %d", offset)
	}

	if stream.GetOffset() != 3 {
		t.Fatalf("Expected GetOffset() to be 3, got %d", stream.GetOffset())
	}

	if stream.GetColumn() != 4 {
		t.Fatalf("Expected GetColumn() to be 4, got %d", stream.GetColumn())
	}

	if stream.GetRow() != 1 {
		t.Fatalf("Expected GetRow() to be 1, got %d", stream.GetRow())
	}

	if next != '1' {
		t.Fatalf("Expected next to be '1', got %d", next)
	}
}

func TestStreamConfirmWithNonMatchingRuneSequence(t *testing.T) {
	// Given
	data := "abc123"

	// When
	stream := NewStream(data)
	ok := stream.MatchChars([]rune("ab12"))
	offset := stream.GetOffset()
	next, _ := stream.PeekChar()

	// Then
	if ok {
		t.Fatalf("Expected a non-match, but got: %t", ok)
	}

	if offset != 0 {
		t.Fatalf("Expected offset to be 0, got %d", offset)
	}

	if stream.GetOffset() != 0 {
		t.Fatalf("Expected GetOffset() to be 0, got %d", stream.GetOffset())
	}

	if stream.GetColumn() != 1 {
		t.Fatalf("Expected GetColumn() to be 1, got %d", stream.GetColumn())
	}

	if stream.GetRow() != 1 {
		t.Fatalf("Expected GetRow() to be 1, got %d", stream.GetRow())
	}

	if next != 'a' {
		t.Fatalf("Expected next to be 'a', got %d", next)
	}
}

//
// Pattern Matching Tests
//

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

	diff, tdOk := Diff(data, string(match))
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
