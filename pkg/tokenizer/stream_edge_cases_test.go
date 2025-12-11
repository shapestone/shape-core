package tokenizer

import (
	"strings"
	"testing"
)

// TestBufferedStreamCloneAtExactBufferBoundary tests cloning behavior
// when the clone is created exactly at a buffer boundary (8KB).
func TestBufferedStreamCloneAtExactBufferBoundary(t *testing.T) {
	// Given - Create data with exactly 8KB + some more
	var sb strings.Builder
	targetSize := 8192 // Exactly readChunkSize
	for sb.Len() < targetSize+100 {
		sb.WriteString("x")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Read exactly to the buffer boundary (8192 bytes)
	for i := 0; i < 8192; i++ {
		stream.NextChar()
	}

	// Create clone at boundary
	clone := stream.Clone()

	// Advance original past boundary
	for i := 0; i < 50; i++ {
		stream.NextChar()
	}

	// Then - Clone should work at boundary position
	for i := 0; i < 50; i++ {
		r, ok := clone.NextChar()
		if !ok {
			t.Fatalf("Clone unexpectedly hit EOF at position %d", 8192+i)
		}
		if r != expectedRunes[8192+i] {
			t.Fatalf("Clone mismatch at position %d: expected '%c', got '%c'",
				8192+i, expectedRunes[8192+i], r)
		}
	}
}

// TestBufferedStreamCloneBeforeAndAfterDiscard tests that clones work correctly
// even when buffer discarding happens.
func TestBufferedStreamCloneBeforeAndAfterDiscard(t *testing.T) {
	// Given - Create large enough data to trigger discard
	var sb strings.Builder
	for i := 0; i < 70000; i++ { // More than bufferSize
		sb.WriteString("a")
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

	// Create clone at beginning (will prevent discard initially)
	clone1 := stream.Clone()

	// Advance original far ahead
	for i := 0; i < 60000; i++ {
		stream.NextChar()
	}

	// Create clone at current position
	clone2 := stream.Clone()

	// Then - Both clones should work
	// Clone1 from beginning
	for i := 0; i < 10; i++ {
		r, ok := clone1.NextChar()
		if !ok {
			t.Fatalf("Clone1 unexpectedly hit EOF at position %d", i)
		}
		if r != 'a' {
			t.Fatalf("Clone1 got wrong character at position %d: '%c'", i, r)
		}
	}

	// Clone2 from position 60000
	for i := 0; i < 10; i++ {
		r, ok := clone2.NextChar()
		if !ok {
			t.Fatalf("Clone2 unexpectedly hit EOF at position %d", 60000+i)
		}
		if r != 'a' {
			t.Fatalf("Clone2 got wrong character at position %d: '%c'", 60000+i, r)
		}
	}
}

// TestBufferedStreamMultipleClonesAtSamePosition tests that multiple clones
// at the same position all work correctly.
func TestBufferedStreamMultipleClonesAtSamePosition(t *testing.T) {
	// Given
	data := "0123456789" + strings.Repeat("abcdefghij", 1000)
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Advance to position 1000
	for i := 0; i < 1000; i++ {
		stream.NextChar()
	}

	// Create multiple clones at same position
	clone1 := stream.Clone()
	clone2 := stream.Clone()
	clone3 := stream.Clone()

	// Then - All clones should read the same data
	for i := 0; i < 100; i++ {
		r1, ok1 := clone1.NextChar()
		r2, ok2 := clone2.NextChar()
		r3, ok3 := clone3.NextChar()

		if !ok1 || !ok2 || !ok3 {
			t.Fatalf("One of the clones hit EOF unexpectedly at position %d", 1000+i)
		}

		expected := expectedRunes[1000+i]
		if r1 != expected || r2 != expected || r3 != expected {
			t.Fatalf("Clone mismatch at position %d: expected '%c', got '%c','%c','%c'",
				1000+i, expected, r1, r2, r3)
		}
	}
}

// TestBufferedStreamCloneAfterMatch tests that cloning after a Match() works correctly.
func TestBufferedStreamCloneAfterMatch(t *testing.T) {
	// Given
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("TestLine")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Create clone and advance it
	clone1 := stream.Clone()
	for i := 0; i < 5000; i++ {
		clone1.NextChar()
	}

	// Match original to clone
	stream.Match(clone1)

	// Create new clone after match
	clone2 := stream.Clone()

	// Then - New clone should be at correct position (5000)
	for i := 0; i < 50; i++ {
		r, ok := clone2.NextChar()
		if !ok {
			t.Fatalf("Clone2 unexpectedly hit EOF at position %d", 5000+i)
		}
		if r != expectedRunes[5000+i] {
			t.Fatalf("Clone2 mismatch at position %d: expected '%c', got '%c'",
				5000+i, expectedRunes[5000+i], r)
		}
	}
}

// TestBufferedStreamCloneWithPatternMatching tests that clones work correctly
// with pattern matching functions that use Clone internally.
func TestBufferedStreamCloneWithPatternMatching(t *testing.T) {
	// Given
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		sb.WriteString("abc123xyz")
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When
	stream := NewStreamFromReader(reader)

	// Try pattern matching multiple times (uses Clone internally)
	for i := 0; i < 100; i++ {
		match, ok := OneOf(
			StringMatcher("abc"),
			StringMatcher("xyz"),
		)(stream)

		if !ok {
			t.Fatalf("Pattern match failed at iteration %d", i)
		}

		if string(match) != "abc" {
			t.Fatalf("Expected 'abc', got '%s' at iteration %d", string(match), i)
		}

		// Consume "123xyz" to get to next "abc"
		stream.NextChar() // 1
		stream.NextChar() // 2
		stream.NextChar() // 3
		stream.NextChar() // x
		stream.NextChar() // y
		stream.NextChar() // z
	}

	// Then - Should have processed data correctly
	if stream.GetOffset() != 900 { // 100 * 9 characters
		t.Fatalf("Expected offset 900, got %d", stream.GetOffset())
	}
}

// TestBufferedStreamCloneWithUTF8AcrossBoundary tests that UTF-8 characters
// are handled correctly when they span buffer boundaries.
func TestBufferedStreamCloneWithUTF8AcrossBoundary(t *testing.T) {
	// Given - Create data with UTF-8 characters near buffer boundary
	var sb strings.Builder
	// Fill up to near 8KB boundary
	for sb.Len() < 8180 {
		sb.WriteString("a")
	}
	// Add UTF-8 characters that might span boundary
	sb.WriteString("Hello 世界 🌍 World")

	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Create clone at beginning
	clone := stream.Clone()

	// Advance original to read through entire content
	count := 0
	for !stream.IsEos() {
		stream.NextChar()
		count++
	}

	// Then - Clone should be able to read all content correctly
	for i := 0; i < len(expectedRunes); i++ {
		r, ok := clone.NextChar()
		if !ok {
			t.Fatalf("Clone unexpectedly hit EOF at position %d (expected %d chars)",
				i, len(expectedRunes))
		}
		if r != expectedRunes[i] {
			t.Fatalf("Clone mismatch at position %d: expected '%c' (%U), got '%c' (%U)",
				i, expectedRunes[i], expectedRunes[i], r, r)
		}
	}

	if count != len(expectedRunes) {
		t.Fatalf("Original read %d runes, expected %d", count, len(expectedRunes))
	}
}
