package tokenizer

import (
	"bytes"
	"testing"
)

func TestFindByte(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		find byte
		want int
	}{
		{"empty", []byte{}, 'a', -1},
		{"single match", []byte("hello"), 'h', 0},
		{"middle match", []byte("hello world"), ' ', 5},
		{"end match", []byte("hello"), 'o', 4},
		{"no match", []byte("hello"), 'z', -1},
		{"long string", []byte("this is a longer string for testing SWAR"), 'S', 36},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindByte(tt.data, tt.find)
			if got != tt.want {
				t.Errorf("FindByte() = %d, want %d", got, tt.want)
			}

			// Compare with bytes.IndexByte for correctness
			expected := bytes.IndexByte(tt.data, tt.find)
			if got != expected {
				t.Errorf("FindByte() = %d, bytes.IndexByte() = %d", got, expected)
			}
		})
	}
}

func TestSkipWhitespace(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"empty", []byte{}, 0},
		{"no whitespace", []byte("hello"), 0},
		{"spaces", []byte("   hello"), 3},
		{"tabs", []byte("\t\thello"), 2},
		{"newlines", []byte("\n\nhello"), 2},
		{"mixed", []byte("  \t\n  hello"), 6},
		{"all whitespace", []byte("    "), 4},
		{"long prefix", []byte("                hello"), 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SkipWhitespace(tt.data)
			if got != tt.want {
				t.Errorf("SkipWhitespace() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNeedsEscaping(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"simple", []byte("hello"), false},
		{"with quote", []byte(`hello"world`), true},
		{"with backslash", []byte(`hello\world`), true},
		{"with newline", []byte("hello\nworld"), true},
		{"with tab", []byte("hello\tworld"), true},
		{"non-ASCII", []byte("hello世界"), true},
		{"empty", []byte{}, false},
		{"alphanumeric", []byte("abc123XYZ"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsEscaping(tt.data)
			if got != tt.want {
				t.Errorf("NeedsEscaping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindEscapeOrQuote(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"empty", []byte{}, -1},
		{"no escape", []byte("hello"), -1},
		{"quote", []byte(`hello"world`), 5},
		{"backslash", []byte(`hello\world`), 5},
		{"quote first", []byte(`"\hello`), 0},
		{"backslash first", []byte(`\hello"`), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindEscapeOrQuote(tt.data)
			if got != tt.want {
				t.Errorf("FindEscapeOrQuote() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFindAnyByte(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		chars []byte
		want  int
	}{
		{"empty data", []byte{}, []byte("abc"), -1},
		{"empty chars", []byte("hello"), []byte{}, -1},
		{"single char match", []byte("hello"), []byte("l"), 2},
		{"multiple chars", []byte("hello"), []byte("lo"), 2},
		{"no match", []byte("hello"), []byte("xyz"), -1},
		{"match at start", []byte("hello"), []byte("h"), 0},
		{"long string multiple chars", []byte("this is a very long string for testing SWAR optimization paths"), []byte("WR"), 40},
		{"many target chars", []byte("abcdefghijklmnop"), []byte("xyz123"), -1},
		{"match second char", []byte("hello world"), []byte("ow"), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindAnyByte(tt.data, tt.chars)
			if got != tt.want {
				t.Errorf("FindAnyByte() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSkipWhitespace_LongStrings(t *testing.T) {
	// Test with various long whitespace sequences to hit SWAR paths
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			"long spaces",
			[]byte("                        hello"),
			24,
		},
		{
			"mixed whitespace long",
			[]byte("  \t  \n  \r  \t\t   content"),
			16,
		},
		{
			"16+ spaces",
			[]byte("                    x"),
			20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SkipWhitespace(tt.data)
			if got != tt.want {
				t.Errorf("SkipWhitespace() = %d, want %d", got, tt.want)
			}
		})
	}
}
