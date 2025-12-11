package tokenizer

import (
	"strings"
	"testing"
)

// TestBufferedStreamCloneAcrossMultipleRefills tests that clones work correctly
// even when the original stream advances through multiple buffer refills.
func TestBufferedStreamCloneAcrossMultipleRefills(t *testing.T) {
	// Given - Create data larger than multiple buffer refills
	var sb strings.Builder
	line := "0123456789abcdefghijklmnop" // 26 bytes per line
	// Create 20KB of data to ensure multiple refills (readChunkSize is 8KB)
	for sb.Len() < 20*1024 {
		sb.WriteString(line)
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Create clone at beginning
	clone1 := stream.Clone()

	// Advance original past multiple buffer refills (20KB)
	for i := 0; i < len(expectedRunes); i++ {
		r, ok := stream.NextChar()
		if !ok {
			t.Fatalf("Original unexpectedly hit EOF at position %d", i)
		}
		if r != expectedRunes[i] {
			t.Fatalf("Original mismatch at position %d: expected '%c', got '%c'", i, expectedRunes[i], r)
		}
	}

	// Then - Clone should still be able to read from the beginning
	for i := 0; i < 100; i++ {
		r, ok := clone1.NextChar()
		if !ok {
			t.Fatalf("Clone1 unexpectedly hit EOF at position %d", i)
		}
		if r != expectedRunes[i] {
			t.Fatalf("Clone1 mismatch at position %d: expected '%c' (%U), got '%c' (%U)",
				i, expectedRunes[i], expectedRunes[i], r, r)
		}
	}
}

// TestBufferedStreamCloneAfterRefill tests that a clone created after a buffer
// refill can still access the data.
func TestBufferedStreamCloneAfterRefill(t *testing.T) {
	// Given - Create data that forces buffer refill
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("Line number ")
		sb.WriteString(string(rune('0' + (i % 10))))
		sb.WriteString("\n")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Advance past first buffer refill (9KB)
	for i := 0; i < 9000 && i < len(expectedRunes); i++ {
		stream.NextChar()
	}

	// Create clone after refill
	clone := stream.Clone()
	clonePos := stream.GetOffset()

	// Advance original further
	for i := 0; i < 1000 && !stream.IsEos(); i++ {
		stream.NextChar()
	}

	// Then - Clone should work from its position
	for i := 0; i < 100 && !clone.IsEos(); i++ {
		r, ok := clone.NextChar()
		if !ok {
			t.Fatalf("Clone unexpectedly hit EOF at position %d", i)
		}
		expectedPos := clonePos + i
		if expectedPos < len(expectedRunes) && r != expectedRunes[expectedPos] {
			t.Fatalf("Clone mismatch at position %d: expected '%c', got '%c'",
				expectedPos, expectedRunes[expectedPos], r)
		}
	}
}

// TestBufferedStreamClonesShareBufferUpdates verifies that when the buffer
// is modified (refilled or discarded), all clones see the changes.
func TestBufferedStreamClonesShareBufferUpdates(t *testing.T) {
	// Given
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString("TestData")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)
	clone1 := stream.Clone()

	// Advance original to trigger refill
	for i := 0; i < 10000; i++ {
		stream.NextChar()
	}

	// Create another clone after refill
	clone2 := stream.Clone()

	// Then - Both clones should be able to read their respective data
	// Clone1 from beginning
	for i := 0; i < 50; i++ {
		r, ok := clone1.NextChar()
		if !ok {
			t.Fatalf("Clone1 unexpectedly hit EOF at position %d", i)
		}
		if r != expectedRunes[i] {
			t.Fatalf("Clone1 mismatch at position %d: expected '%c', got '%c'",
				i, expectedRunes[i], r)
		}
	}

	// Clone2 from position 10000
	for i := 0; i < 50; i++ {
		r, ok := clone2.NextChar()
		if !ok {
			t.Fatalf("Clone2 unexpectedly hit EOF at position %d", 10000+i)
		}
		if r != expectedRunes[10000+i] {
			t.Fatalf("Clone2 mismatch at position %d: expected '%c', got '%c'",
				10000+i, expectedRunes[10000+i], r)
		}
	}
}

// TestBufferedStreamMatchAfterBufferManagement tests that Match() works
// correctly after buffer has been refilled and potentially discarded.
func TestBufferedStreamMatchAfterBufferManagement(t *testing.T) {
	// Given
	var sb strings.Builder
	for i := 0; i < 1500; i++ {
		sb.WriteString("ContentLine")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When
	stream := NewStreamFromReader(reader)

	// Advance to position 8000
	for i := 0; i < 8000; i++ {
		stream.NextChar()
	}

	// Create clone at position 8000
	clone := stream.Clone()

	// Advance clone to position 8500
	for i := 0; i < 500; i++ {
		clone.NextChar()
	}

	// Match original to clone (should move original to 8500)
	stream.Match(clone)

	// Then - Original should now be at position 8500
	if stream.GetOffset() != 8500 {
		t.Fatalf("Expected offset 8500, got %d", stream.GetOffset())
	}

	// And should read correct data from that position
	for i := 0; i < 50; i++ {
		r, ok := stream.NextChar()
		if !ok {
			t.Fatalf("Stream unexpectedly hit EOF at position %d", 8500+i)
		}
		if r != expectedRunes[8500+i] {
			t.Fatalf("Stream mismatch at position %d: expected '%c', got '%c'",
				8500+i, expectedRunes[8500+i], r)
		}
	}
}
