package tokenizer

import (
	"fmt"
	"os"
	"strings"
)

// ExampleNewStream demonstrates using the basic stream implementation
// for small strings that fit entirely in memory.
func ExampleNewStream() {
	data := "Hello, World!"
	stream := NewStream(data)

	// Read characters one by one
	for !stream.IsEos() {
		ch, ok := stream.NextChar()
		if ok {
			fmt.Printf("%c", ch)
		}
	}
	fmt.Println()

	// Output: Hello, World!
}

// ExampleNewStreamFromReader demonstrates using the buffered stream
// implementation for large files and streaming data.
func ExampleNewStreamFromReader() {
	data := "Line 1\nLine 2\nLine 3\n"
	reader := strings.NewReader(data)
	stream := NewStreamFromReader(reader)

	lineCount := 0
	for !stream.IsEos() {
		ch, ok := stream.NextChar()
		if ok && ch == '\n' {
			lineCount++
		}
	}

	fmt.Printf("Total lines: %d\n", lineCount)
	// Output: Total lines: 3
}

// ExampleNewStreamFromReader_largeFile demonstrates processing a large file
// with constant memory usage.
func ExampleNewStreamFromReader_largeFile() {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "example_*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some data
	tmpFile.WriteString("Line 1\nLine 2\nLine 3\n")
	tmpFile.Close()

	// Open for reading
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create buffered stream
	stream := NewStreamFromReader(file)

	// Process the file with constant memory
	lineCount := 0
	for !stream.IsEos() {
		ch, ok := stream.NextChar()
		if ok && ch == '\n' {
			lineCount++
		}
	}

	fmt.Printf("Processed %d lines\n", lineCount)
	// Output: Processed 3 lines
}

// ExampleStream_Clone demonstrates backtracking with Clone() and Match().
func ExampleStream_Clone() {
	stream := NewStream("abc123")

	// Save current position
	checkpoint := stream.Clone()

	// Try to match something
	stream.NextChar() // a
	stream.NextChar() // b
	stream.NextChar() // c

	// Backtrack to checkpoint
	stream.Match(checkpoint)

	// We're back at the beginning
	ch, _ := stream.PeekChar()
	fmt.Printf("After backtrack: %c\n", ch)
	// Output: After backtrack: a
}

// ExampleStream_PeekChar demonstrates peeking at the next character
// without advancing the stream.
func ExampleStream_PeekChar() {
	stream := NewStream("test")

	// Peek multiple times - should return the same character
	ch1, _ := stream.PeekChar()
	ch2, _ := stream.PeekChar()

	fmt.Printf("First peek: %c\n", ch1)
	fmt.Printf("Second peek: %c\n", ch2)

	// Now advance
	ch3, _ := stream.NextChar()
	fmt.Printf("After NextChar: %c\n", ch3)

	// Output:
	// First peek: t
	// Second peek: t
	// After NextChar: t
}

// ExampleStream_MatchChars demonstrates matching a sequence of characters.
func ExampleStream_MatchChars() {
	stream := NewStream("hello world")

	// Try to match "hello"
	matched := stream.MatchChars([]rune("hello"))
	fmt.Printf("Matched 'hello': %t\n", matched)

	// Stream is now positioned after "hello"
	ch, _ := stream.PeekChar()
	fmt.Printf("Next character: %c\n", ch)

	// Output:
	// Matched 'hello': true
	// Next character:
}

// ExampleStream_GetRow demonstrates tracking line numbers.
func ExampleStream_GetRow() {
	stream := NewStream("line 1\nline 2\nline 3")

	// Read until second newline
	for {
		ch, ok := stream.NextChar()
		if !ok {
			break
		}
		if ch == '\n' && stream.GetRow() == 2 {
			break
		}
	}

	fmt.Printf("Current row: %d\n", stream.GetRow())
	fmt.Printf("Current column: %d\n", stream.GetColumn())

	// Output:
	// Current row: 2
	// Current column: 1
}

// ExamplePattern demonstrates using pattern matchers with buffered streams.
func ExamplePattern() {
	reader := strings.NewReader("abc123xyz")
	stream := NewStreamFromReader(reader)

	// Try to match either "abc" or "xyz"
	pattern := OneOf(
		StringMatcher("abc"),
		StringMatcher("xyz"),
	)

	matched, ok := pattern(stream)
	if ok {
		fmt.Printf("Matched: %s\n", string(matched))
	}

	// Output: Matched: abc
}

// ExampleSequence demonstrates sequential pattern matching.
func ExampleSequence() {
	stream := NewStream("abc123")

	// Match "abc" followed by "123"
	pattern := Sequence(
		StringMatcher("abc"),
		StringMatcher("123"),
	)

	matched, ok := pattern(stream)
	if ok {
		fmt.Printf("Matched: %s\n", string(matched))
	}

	// Output: Matched: abc123
}

// ExampleOptional demonstrates optional pattern matching.
func ExampleOptional() {
	stream := NewStream("hello")

	// Try to match optional "+"
	pattern := Sequence(
		StringMatcher("hello"),
		Optional(CharMatcher('+')),
	)

	matched, ok := pattern(stream)
	if ok {
		fmt.Printf("Matched: %s\n", string(matched))
	}

	// Output: Matched: hello
}

// ExampleBufferedStreamImpl_UTF8 demonstrates UTF-8 handling.
func ExampleNewStreamFromReader_uTF8() {
	data := "Hello 世界 🌍"
	reader := strings.NewReader(data)
	stream := NewStreamFromReader(reader)

	runeCount := 0
	for !stream.IsEos() {
		_, ok := stream.NextChar()
		if ok {
			runeCount++
		}
	}

	fmt.Printf("Total runes: %d\n", runeCount)
	// Output: Total runes: 10
}
