package tokenizer

import (
	"testing"
)

func TestTokenizer_InitializeFromStream(t *testing.T) {
	// Create a stream
	stream := NewStream("hello world")

	// Create a tokenizer
	tokenizer := NewTokenizer()

	// Initialize from stream
	tokenizer.InitializeFromStream(stream)

	// Verify it works by peeking
	if tokenizer.hasMoreTokens() == false {
		t.Error("Tokenizer should have tokens after InitializeFromStream")
	}
}
