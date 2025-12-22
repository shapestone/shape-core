package tokenizer

import (
	"testing"
)

func TestCursorSync(t *testing.T) {
	// Test ASCII-only content
	input := `{"name":"value"}`
	stream := NewStream(input).(*streamImpl)

	// Read first character '{'
	r, ok := stream.NextChar()
	if !ok || r != '{' {
		t.Fatalf("Expected '{', got %q", r)
	}

	// Both cursors should be at position 1
	if stream.location.Cursor != 1 {
		t.Errorf("After NextChar, location.Cursor = %d, want 1", stream.location.Cursor)
	}
	if stream.bytePos != 1 {
		t.Errorf("After NextChar, bytePos = %d, want 1", stream.bytePos)
	}

	// Peek next byte (should be '"')
	b, ok := stream.PeekByte()
	if !ok {
		t.Fatal("PeekByte failed")
	}
	if b != '"' {
		t.Errorf("Expected '\"', got %q (0x%02x)", b, b)
	}

	// Read via NextByte
	b2, ok := stream.NextByte()
	if !ok || b2 != '"' {
		t.Fatalf("Expected '\"' from NextByte, got %q", b2)
	}

	// Both cursors should now be at position 2
	if stream.location.Cursor != 2 {
		t.Errorf("After NextByte, location.Cursor = %d, want 2", stream.location.Cursor)
	}
	if stream.bytePos != 2 {
		t.Errorf("After NextByte, bytePos = %d, want 2", stream.bytePos)
	}
}

func TestCursorSyncUTF8(t *testing.T) {
	// Test UTF-8 content: "α" is 2 bytes (0xCE 0xB1)
	input := `"α"`
	stream := NewStream(input).(*streamImpl)

	// Check mapping
	t.Logf("Rune->Byte mapping: %v", stream.runeToBytePos)
	t.Logf("Total runes: %d, total bytes: %d", stream.length, stream.totalSize)

	// Read '"' via NextChar
	r, ok := stream.NextChar()
	if !ok || r != '"' {
		t.Fatalf("Expected '\"', got %q", r)
	}

	// Cursor should be at rune 1, byte 1
	if stream.location.Cursor != 1 {
		t.Errorf("After reading '\"', location.Cursor = %d, want 1", stream.location.Cursor)
	}
	if stream.bytePos != 1 {
		t.Errorf("After reading '\"', bytePos = %d, want 1", stream.bytePos)
	}

	// Read 'α' via NextChar (should advance cursor by 1, bytePos by 2)
	r, ok = stream.NextChar()
	if !ok || r != 'α' {
		t.Fatalf("Expected 'α', got %q", r)
	}

	// Cursor should be at rune 2, byte 3
	if stream.location.Cursor != 2 {
		t.Errorf("After reading 'α', location.Cursor = %d, want 2", stream.location.Cursor)
	}
	if stream.bytePos != 3 {
		t.Errorf("After reading 'α', bytePos = %d, want 3 (byte 1 + 2-byte UTF-8)", stream.bytePos)
	}
}
