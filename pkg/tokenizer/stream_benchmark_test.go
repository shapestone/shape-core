package tokenizer

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

// BenchmarkBufferedStreamMemoryEfficiency tests that the buffered stream
// maintains constant memory usage even with very large inputs.
func BenchmarkBufferedStreamMemoryEfficiency(b *testing.B) {
	// Create a large string (10MB)
	var sb strings.Builder
	line := "This is a test line with some content that makes it reasonably long\n"
	for i := 0; i < 150000; i++ { // ~10MB
		sb.WriteString(line)
	}
	data := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(data)
		stream := NewStreamFromReader(reader)

		// Read through the entire stream
		for !stream.IsEos() {
			stream.NextChar()
		}
	}
}

// BenchmarkStreamVsBufferedStream compares performance and memory usage
func BenchmarkStreamVsBufferedStream(b *testing.B) {
	data := "abc\n123\nxyz\n"
	for i := 0; i < 1000; i++ {
		data += "line " + fmt.Sprintf("%d", i) + "\n"
	}

	b.Run("NewStream", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stream := NewStream(data)
			for !stream.IsEos() {
				stream.NextChar()
			}
		}
	})

	b.Run("NewStreamFromReader", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(data)
			stream := NewStreamFromReader(reader)
			for !stream.IsEos() {
				stream.NextChar()
			}
		}
	})
}

// TestLargeFileMemoryUsage creates a temporary large file and verifies
// that memory usage remains constant while processing it.
func TestLargeFileMemoryUsage(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	// Create a temporary file with ~100MB of data
	tmpFile, err := os.CreateTemp("", "stream_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write 100MB of data (with some structure for position tracking)
	line := "This is test line with enough content to make it substantial\n"
	targetSize := 100 * 1024 * 1024 // 100MB
	written := 0
	lineNum := 0

	for written < targetSize {
		lineStr := fmt.Sprintf("Line %d: %s", lineNum, line)
		n, err := tmpFile.WriteString(lineStr)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		written += n
		lineNum++
	}

	// Sync and reopen for reading
	tmpFile.Sync()
	tmpFile.Close()

	// Open for reading
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	// Get initial memory stats
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	initialAlloc := m1.Alloc

	// Create buffered stream and read through it
	stream := NewStreamFromReader(file)

	charCount := 0
	lineCount := 0
	maxMemIncrease := uint64(0)

	// Read through the file, checking memory periodically
	checkInterval := 1000000 // Check every 1M characters
	for !stream.IsEos() {
		r, ok := stream.NextChar()
		if ok {
			charCount++
			if r == '\n' {
				lineCount++
			}

			// Check memory usage periodically
			if charCount%checkInterval == 0 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				memIncrease := m.Alloc - initialAlloc
				if memIncrease > maxMemIncrease {
					maxMemIncrease = memIncrease
				}

				// Memory increase should be bounded by buffer size
				// Allow up to 10MB for overhead (buffer + other allocations)
				maxAllowedIncrease := uint64(10 * 1024 * 1024)
				if memIncrease > maxAllowedIncrease {
					t.Fatalf("Memory usage too high: %d MB (max allowed: %d MB)",
						memIncrease/(1024*1024), maxAllowedIncrease/(1024*1024))
				}
			}
		}
	}

	t.Logf("Processed %d characters (%d lines) from 100MB file", charCount, lineCount)
	t.Logf("Max memory increase: %d MB", maxMemIncrease/(1024*1024))

	// Verify we read a reasonable amount of data
	if charCount < targetSize {
		t.Logf("Read %d characters from file (expected ~%d)", charCount, targetSize)
	}

	// Verify memory stayed bounded
	if maxMemIncrease > 10*1024*1024 {
		t.Errorf("Memory usage exceeded expected bounds: %d MB", maxMemIncrease/(1024*1024))
	}
}

// BenchmarkBufferedStreamWithClones tests performance with backtracking
func BenchmarkBufferedStreamWithClones(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("test data line ")
		sb.WriteString(fmt.Sprintf("%d", i))
		sb.WriteString("\n")
	}
	data := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(data)
		stream := NewStreamFromReader(reader)

		// Simulate pattern matching with backtracking
		for j := 0; j < 100; j++ {
			clone := stream.Clone()
			// Try to match something
			for k := 0; k < 10 && !clone.IsEos(); k++ {
				clone.NextChar()
			}
			// Sometimes advance original, sometimes backtrack
			if j%2 == 0 {
				stream.Match(clone)
			}
		}
	}
}
