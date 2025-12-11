package tokenizer

import (
	"bytes"
	"fmt"
	"strings"
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

//
// Buffered Stream Tests
//

func TestBufferedStreamFromReader(t *testing.T) {
	// Given
	data := "abc123"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

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

func TestBufferedStreamPeekChar(t *testing.T) {
	// Given
	data := "a1"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
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

func TestBufferedStreamNextChar(t *testing.T) {
	// Given
	data := "a"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
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

func TestBufferedStreamRowAndColumnTracking(t *testing.T) {
	// Given
	data := "abc\n123\n!@#"
	reader := strings.NewReader(data)

	// When
	var r rune
	var ok bool
	stream := NewStreamFromReader(reader)
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

func TestBufferedStreamMatchChars(t *testing.T) {
	// Given
	data := "abc123"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
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

	if next != '1' {
		t.Fatalf("Expected next to be '1', got %c", next)
	}
}

func TestBufferedStreamCloneAndMatch(t *testing.T) {
	// Given
	data := "abc123"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
	clone := stream.Clone()

	// Advance original
	stream.NextChar()
	stream.NextChar()

	// Clone should still be at start
	ch, ok := clone.PeekChar()

	// Then
	if !ok {
		t.Fatalf("Expected ok to be true, got %t", ok)
	}

	if ch != 'a' {
		t.Fatalf("Expected 'a', got '%c'", ch)
	}

	// When - match original to clone
	stream.Match(clone)
	ch2, _ := stream.PeekChar()

	// Then
	if ch2 != 'a' {
		t.Fatalf("Expected stream to be reset to 'a', got '%c'", ch2)
	}
}

func TestBufferedStreamUTF8Support(t *testing.T) {
	// Given - String with multi-byte UTF-8 characters
	data := "Hello 世界 🌍"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
	runes := []rune{}
	for !stream.IsEos() {
		r, ok := stream.NextChar()
		if ok {
			runes = append(runes, r)
		}
	}

	// Then
	expected := []rune(data)
	if len(runes) != len(expected) {
		t.Fatalf("Expected %d runes, got %d", len(expected), len(runes))
	}

	for i, r := range runes {
		if r != expected[i] {
			t.Fatalf("Rune mismatch at position %d: expected '%c' (%U), got '%c' (%U)",
				i, expected[i], expected[i], r, r)
		}
	}
}

func TestBufferedStreamLargeInput(t *testing.T) {
	// Given - Create a large string that exceeds buffer size
	// Using 100KB of data (larger than the 64KB buffer)
	var sb strings.Builder
	line := "This is a test line with some content\n"
	for i := 0; i < 2600; i++ { // ~100KB
		sb.WriteString(line)
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
	count := 0
	for !stream.IsEos() {
		_, ok := stream.NextChar()
		if ok {
			count++
		}
	}

	// Then
	expectedCount := len([]rune(data))
	if count != expectedCount {
		t.Fatalf("Expected %d characters, got %d", expectedCount, count)
	}
}

func TestBufferedStreamBufferRefill(t *testing.T) {
	// Given - Data larger than initial buffer
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("abcdefghij")
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

	// Read past the buffer size
	for i := 0; i < 70000; i++ {
		stream.NextChar()
	}

	// Then - Should still be able to read correctly
	ch, ok := stream.PeekChar()
	if !ok {
		t.Fatalf("Expected to still have data to read")
	}

	expected := []rune(data)
	if ch != expected[70000] {
		t.Fatalf("Expected '%c', got '%c'", expected[70000], ch)
	}
}

func TestBufferedStreamCloneWithBacktracking(t *testing.T) {
	// Given
	data := "abc123xyz"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

	// Advance to position 3
	stream.NextChar() // a
	stream.NextChar() // b
	stream.NextChar() // c

	clone1 := stream.Clone()

	// Advance original further
	stream.NextChar() // 1
	stream.NextChar() // 2
	stream.NextChar() // 3

	clone2 := stream.Clone()

	// Advance original to end
	stream.NextChar() // x
	stream.NextChar() // y
	stream.NextChar() // z

	// Then - clone1 should be at position 3
	ch1, _ := clone1.PeekChar()
	if ch1 != '1' {
		t.Fatalf("Expected clone1 to be at '1', got '%c'", ch1)
	}

	// clone2 should be at position 6
	ch2, _ := clone2.PeekChar()
	if ch2 != 'x' {
		t.Fatalf("Expected clone2 to be at 'x', got '%c'", ch2)
	}

	// original should be at end
	if !stream.IsEos() {
		t.Fatalf("Expected original stream to be at EOS")
	}
}

func TestBufferedStreamPositionTrackingAcrossBufferBoundaries(t *testing.T) {
	// Given - Create data with newlines that will cross buffer boundaries
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString("line")
		sb.WriteString(fmt.Sprintf("%d", i))
		sb.WriteString("\n")
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

	// Read through multiple buffer refills
	lineCount := 0
	for !stream.IsEos() {
		r, ok := stream.NextChar()
		if ok && r == '\n' {
			lineCount++
		}
	}

	// Then
	if lineCount != 2000 {
		t.Fatalf("Expected 2000 lines, got %d", lineCount)
	}

	// Final position should be correct
	if stream.GetRow() != 2001 { // One past the last line
		t.Fatalf("Expected GetRow() to be 2001, got %d", stream.GetRow())
	}
}

func TestBufferedStreamReadingAcrossBufferBoundary(t *testing.T) {
	// Given - Create data that is exactly 2x buffer size to force boundary crossing
	// Buffer size is 64KB of runes, so create ~130KB of data
	var sb strings.Builder
	line := "0123456789abcdef" // 16 bytes per line
	targetSize := 130 * 1024   // 130KB
	for sb.Len() < targetSize {
		sb.WriteString(line)
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When - Read all data character by character
	stream := NewStreamFromReader(reader)
	readCount := 0
	expectedRunes := []rune(data)

	for i := 0; i < len(expectedRunes); i++ {
		r, ok := stream.NextChar()
		if !ok {
			t.Fatalf("Expected to read character at position %d, but got EOF. Read %d/%d characters",
				i, readCount, len(expectedRunes))
		}
		if r != expectedRunes[i] {
			t.Fatalf("Character mismatch at position %d: expected '%c' (%U), got '%c' (%U)",
				i, expectedRunes[i], expectedRunes[i], r, r)
		}
		readCount++
	}

	// Then
	if readCount != len(expectedRunes) {
		t.Fatalf("Expected to read %d characters, got %d", len(expectedRunes), readCount)
	}

	// Verify we're at EOF
	if !stream.IsEos() {
		t.Fatal("Expected to be at end of stream")
	}
}

func TestBufferedStreamReset(t *testing.T) {
	// Given
	data := "abc123"
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)
	stream.NextChar()
	stream.NextChar()
	stream.Reset()

	// Then
	if stream.GetOffset() != 0 {
		t.Fatalf("Expected offset to be 0, got %d", stream.GetOffset())
	}

	if stream.GetRow() != 1 {
		t.Fatalf("Expected row to be 1, got %d", stream.GetRow())
	}

	if stream.GetColumn() != 1 {
		t.Fatalf("Expected column to be 1, got %d", stream.GetColumn())
	}

	ch, _ := stream.PeekChar()
	if ch != 'a' {
		t.Fatalf("Expected to be back at 'a', got '%c'", ch)
	}
}

func TestBufferedStreamWithBytesBuffer(t *testing.T) {
	// Given
	data := []byte("test data with bytes")
	buffer := bytes.NewBuffer(data)

	// When
	stream := NewStreamFromReader(buffer)

	result := []rune{}
	for !stream.IsEos() {
		r, ok := stream.NextChar()
		if ok {
			result = append(result, r)
		}
	}

	// Then
	expected := string(data)
	actual := string(result)

	if actual != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, actual)
	}
}

func TestBufferedStreamPatternMatchingWithOneOf(t *testing.T) {
	// Given
	data := "abc"
	reader := strings.NewReader(data)
	stream := NewStreamFromReader(reader)

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

	if string(match) != "abc" {
		t.Fatalf("Expected 'abc', got '%s'", string(match))
	}
}

func TestBufferedStreamEmptyReader(t *testing.T) {
	// Given
	reader := strings.NewReader("")

	// When
	stream := NewStreamFromReader(reader)

	// Then
	if !stream.IsEos() {
		t.Fatalf("Expected IsEos() to be true for empty reader")
	}

	ch, ok := stream.PeekChar()
	if ok {
		t.Fatalf("Expected ok to be false, got true")
	}

	if ch != 0 {
		t.Fatalf("Expected rune to be 0, got %d", ch)
	}
}
