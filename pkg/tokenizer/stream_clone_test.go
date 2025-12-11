package tokenizer

import (
	"strings"
	"testing"
)

// TestBufferedStreamCloneBufferSharing tests that clones properly share buffer state
// even when the buffer is modified (sliced or refilled) by either the original or clone.
func TestBufferedStreamCloneBufferSharing(t *testing.T) {
	// Given - Create data larger than readChunkSize (8KB) to force buffer refill
	var sb strings.Builder
	line := "0123456789abcdef" // 16 bytes per line
	targetSize := 10 * 1024    // 10KB - larger than readChunkSize
	for sb.Len() < targetSize {
		sb.WriteString(line)
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When - Create stream and clone at beginning
	stream := NewStreamFromReader(reader)
	clone := stream.Clone()

	// Read past the first buffer refill on the original
	expectedRunes := []rune(data)
	for i := 0; i < 9000; i++ { // Read 9KB, past the 8KB buffer refill point
		r, ok := stream.NextChar()
		if !ok {
			t.Fatalf("Unexpected EOF at position %d", i)
		}
		if r != expectedRunes[i] {
			t.Fatalf("Mismatch at position %d: expected '%c', got '%c'", i, expectedRunes[i], r)
		}
	}

	// Then - Clone should still be able to read from the beginning
	// This tests that the buffer is properly shared and not lost during refill
	for i := 0; i < 100; i++ {
		r, ok := clone.NextChar()
		if !ok {
			t.Fatalf("Clone unexpectedly hit EOF at position %d", i)
		}
		if r != expectedRunes[i] {
			t.Fatalf("Clone mismatch at position %d: expected '%c' (%U), got '%c' (%U)",
				i, expectedRunes[i], expectedRunes[i], r, r)
		}
	}
}

// TestBufferedStreamCloneAcrossBufferBoundary tests cloning behavior when
// the original advances past a buffer refill point.
func TestBufferedStreamCloneAcrossBufferBoundary(t *testing.T) {
	// Given - Create data that will require multiple buffer refills
	var sb strings.Builder
	line := "This is a test line for buffer boundary testing\n"
	for i := 0; i < 1000; i++ {
		sb.WriteString(line)
	}
	data := sb.String()
	reader := strings.NewReader(data)

	// When - Read through first buffer and create clone
	stream := NewStreamFromReader(reader)

	// Read 8KB (past first refill)
	for i := 0; i < 8192; i++ {
		stream.NextChar()
	}

	// Create clone at this position
	clone := stream.Clone()
	cloneStartPos := stream.GetOffset()

	// Read more on original to trigger another buffer refill
	for i := 0; i < 8192; i++ {
		stream.NextChar()
	}

	// Then - Clone should be able to read from its saved position
	expectedRunes := []rune(data)
	for i := 0; i < 100; i++ {
		r, ok := clone.NextChar()
		if !ok {
			t.Fatalf("Clone unexpectedly hit EOF at position %d (offset %d)", i, cloneStartPos+i)
		}
		expectedPos := cloneStartPos + i
		if r != expectedRunes[expectedPos] {
			t.Fatalf("Clone mismatch at position %d: expected '%c', got '%c'",
				expectedPos, expectedRunes[expectedPos], r)
		}
	}
}

// TestBufferedStreamMultipleClonesWithBufferManagement tests that multiple clones
// at different positions all work correctly even as the buffer is managed.
func TestBufferedStreamMultipleClonesWithBufferManagement(t *testing.T) {
	// Given - Create substantial data
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString("ABCDEFGHIJ")
	}
	data := sb.String()
	reader := strings.NewReader(data)
	expectedRunes := []rune(data)

	// When - Create stream and multiple clones at different positions
	stream := NewStreamFromReader(reader)

	// Advance to position 1000 and create clone1
	for i := 0; i < 1000; i++ {
		stream.NextChar()
	}
	clone1 := stream.Clone()
	clone1Pos := 1000

	// Advance to position 5000 and create clone2
	for i := 0; i < 4000; i++ {
		stream.NextChar()
	}
	clone2 := stream.Clone()
	clone2Pos := 5000

	// Advance original to position 10000 (triggering buffer management)
	for i := 0; i < 5000; i++ {
		stream.NextChar()
	}

	// Then - All clones should read correct data from their positions
	// Test clone1
	for i := 0; i < 50; i++ {
		r, ok := clone1.NextChar()
		if !ok {
			t.Fatalf("Clone1 unexpectedly hit EOF at position %d", clone1Pos+i)
		}
		if r != expectedRunes[clone1Pos+i] {
			t.Fatalf("Clone1 mismatch at position %d: expected '%c', got '%c'",
				clone1Pos+i, expectedRunes[clone1Pos+i], r)
		}
	}

	// Test clone2
	for i := 0; i < 50; i++ {
		r, ok := clone2.NextChar()
		if !ok {
			t.Fatalf("Clone2 unexpectedly hit EOF at position %d", clone2Pos+i)
		}
		if r != expectedRunes[clone2Pos+i] {
			t.Fatalf("Clone2 mismatch at position %d: expected '%c', got '%c'",
				clone2Pos+i, expectedRunes[clone2Pos+i], r)
		}
	}

	// Test original stream
	for i := 0; i < 50; i++ {
		r, ok := stream.NextChar()
		if !ok {
			t.Fatalf("Original unexpectedly hit EOF at position %d", 10000+i)
		}
		if r != expectedRunes[10000+i] {
			t.Fatalf("Original mismatch at position %d: expected '%c', got '%c'",
				10000+i, expectedRunes[10000+i], r)
		}
	}
}
