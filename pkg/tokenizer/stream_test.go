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

func TestStreamReset(t *testing.T) {
	// Given
	stream := NewStream("abc123")

	// Advance the stream by reading chars
	stream.NextChar()
	stream.NextChar()
	stream.NextChar()

	// Verify position before reset
	if stream.GetOffset() != 3 {
		t.Fatalf("Expected offset 3 before reset, got %d", stream.GetOffset())
	}

	// When
	stream.Reset()

	// Then
	if stream.GetOffset() != 0 {
		t.Errorf("Reset() offset = %d, want 0", stream.GetOffset())
	}
	if stream.GetRow() != 1 {
		t.Errorf("Reset() row = %d, want 1", stream.GetRow())
	}
	if stream.GetColumn() != 1 {
		t.Errorf("Reset() column = %d, want 1", stream.GetColumn())
	}

	// Verify we can read from start again
	ch, ok := stream.PeekChar()
	if !ok {
		t.Error("Expected to be able to peek after reset")
	}
	if ch != 'a' {
		t.Errorf("PeekChar() after reset = %c, want 'a'", ch)
	}
}

func TestStream_PeekBytes(t *testing.T) {
	stream := NewStream("hello world").(ByteStream)

	// Peek first 5 bytes
	bytes := stream.PeekBytes(5)
	if string(bytes) != "hello" {
		t.Errorf("PeekBytes(5) = %q, want 'hello'", string(bytes))
	}

	// Stream position shouldn't change
	if stream.GetOffset() != 0 {
		t.Errorf("PeekBytes shouldn't advance stream, offset = %d", stream.GetOffset())
	}

	// Peek more bytes than available
	bytes = stream.PeekBytes(100)
	if string(bytes) != "hello world" {
		t.Errorf("PeekBytes(100) = %q, want 'hello world'", string(bytes))
	}

	// Advance and peek again
	stream.NextChar() // 'h'
	stream.NextChar() // 'e'
	bytes = stream.PeekBytes(3)
	if string(bytes) != "llo" {
		t.Errorf("PeekBytes(3) after advance = %q, want 'llo'", string(bytes))
	}
}

func TestStream_SkipWhitespace(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantOffset     int
		wantRow        int
		wantColumn     int
		wantNextChar   rune
	}{
		{
			name:         "no whitespace",
			input:        "hello",
			wantOffset:   0,
			wantRow:      1,
			wantColumn:   1,
			wantNextChar: 'h',
		},
		{
			name:         "spaces",
			input:        "   hello",
			wantOffset:   3,
			wantRow:      1,
			wantColumn:   4,
			wantNextChar: 'h',
		},
		{
			name:         "tabs",
			input:        "\t\thello",
			wantOffset:   2,
			wantRow:      1,
			wantColumn:   3,
			wantNextChar: 'h',
		},
		{
			name:         "newlines",
			input:        "\n\nhello",
			wantOffset:   2,
			wantRow:      3,
			wantColumn:   1,
			wantNextChar: 'h',
		},
		{
			name:         "mixed whitespace",
			input:        "  \t\n  hello",
			wantOffset:   6,
			wantRow:      2,
			wantColumn:   3,
			wantNextChar: 'h',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := NewStream(tt.input).(ByteStream)
			stream.SkipWhitespace()

			if stream.GetOffset() != tt.wantOffset {
				t.Errorf("SkipWhitespace() offset = %d, want %d", stream.GetOffset(), tt.wantOffset)
			}
			if stream.GetRow() != tt.wantRow {
				t.Errorf("SkipWhitespace() row = %d, want %d", stream.GetRow(), tt.wantRow)
			}
			if stream.GetColumn() != tt.wantColumn {
				t.Errorf("SkipWhitespace() column = %d, want %d", stream.GetColumn(), tt.wantColumn)
			}

			ch, ok := stream.PeekChar()
			if !ok {
				t.Error("Expected to peek character after SkipWhitespace")
			}
			if ch != tt.wantNextChar {
				t.Errorf("PeekChar() after SkipWhitespace = %c, want %c", ch, tt.wantNextChar)
			}
		})
	}
}

func TestStream_SkipUntil(t *testing.T) {
	stream := NewStream("hello,world").(ByteStream)

	// Skip until comma
	skipped := stream.SkipUntil(',')
	if skipped != 5 {
		t.Errorf("SkipUntil(',') = %d, want 5", skipped)
	}

	// Verify stream is at comma (not past it) - use byte-level peek
	b, ok := stream.PeekByte()
	if !ok || b != ',' {
		t.Errorf("PeekByte() after SkipUntil = %c, want ','", b)
	}

	// Skip past comma
	stream.NextByte()

	// Skip until non-existent character
	skipped = stream.SkipUntil('z')
	if skipped != 5 { // "world" = 5 chars
		t.Errorf("SkipUntil('z') when not found = %d, want 5", skipped)
	}
}

func TestStream_FindByte(t *testing.T) {
	stream := NewStream("hello world").(ByteStream)

	// Find space
	offset := stream.FindByte(' ')
	if offset != 5 {
		t.Errorf("FindByte(' ') = %d, want 5", offset)
	}

	// Stream position shouldn't change
	if stream.GetOffset() != 0 {
		t.Errorf("FindByte shouldn't advance stream, offset = %d", stream.GetOffset())
	}

	// Find character that doesn't exist
	offset = stream.FindByte('z')
	if offset != -1 {
		t.Errorf("FindByte('z') = %d, want -1", offset)
	}

	// Advance and find again
	stream.NextChar() // 'h'
	stream.NextChar() // 'e'
	stream.NextChar() // 'l'
	offset = stream.FindByte('o')
	if offset != 1 { // 'o' is 1 position away from current 'l'
		t.Errorf("FindByte('o') after advance = %d, want 1", offset)
	}
}

func TestStream_FindAny(t *testing.T) {
	stream := NewStream("hello world").(ByteStream)

	// Find first of multiple characters
	offset := stream.FindAny([]byte(" ,;"))
	if offset != 5 {
		t.Errorf("FindAny(\" ,;\") = %d, want 5", offset)
	}

	// Stream position shouldn't change
	if stream.GetOffset() != 0 {
		t.Errorf("FindAny shouldn't advance stream, offset = %d", stream.GetOffset())
	}

	// Find none of the characters
	offset = stream.FindAny([]byte("xyz"))
	if offset != -1 {
		t.Errorf("FindAny(\"xyz\") = %d, want -1", offset)
	}

	// Empty chars
	offset = stream.FindAny([]byte{})
	if offset != -1 {
		t.Errorf("FindAny([]) = %d, want -1", offset)
	}
}

func TestStream_GetSetLocation(t *testing.T) {
	stream := NewStream("hello\nworld")

	// Get initial location
	loc := stream.GetLocation()
	if loc.Row != 1 || loc.Column != 1 {
		t.Errorf("Initial location = (%d, %d), want (1, 1)", loc.Row, loc.Column)
	}

	// Advance stream
	stream.NextChar() // 'h'
	stream.NextChar() // 'e'
	stream.NextChar() // 'l'

	// Set custom location
	stream.SetLocation(Location{Row: 10, Column: 20})

	// Verify location changed
	loc = stream.GetLocation()
	if loc.Row != 10 || loc.Column != 20 {
		t.Errorf("After SetLocation = (%d, %d), want (10, 20)", loc.Row, loc.Column)
	}

	// Advance should continue from custom location
	stream.NextChar() // 'l'
	loc = stream.GetLocation()
	if loc.Row != 10 || loc.Column != 21 {
		t.Errorf("After NextChar from custom location = (%d, %d), want (10, 21)", loc.Row, loc.Column)
	}
}
