package yamlv

import (
	"testing"

	"github.com/shapestone/shape/internal/tokenizer"
)

// TestKeyMatcher tests the keyMatcher function
func TestKeyMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "simple key",
			input:      "name:",
			wantToken:  true,
			wantValue:  "name",
			wantRemain: ":",
		},
		{
			name:       "key with underscore",
			input:      "user_name:",
			wantToken:  true,
			wantValue:  "user_name",
			wantRemain: ":",
		},
		{
			name:       "key with hyphen",
			input:      "user-name:",
			wantToken:  true,
			wantValue:  "user-name",
			wantRemain: ":",
		},
		{
			name:       "key with numbers",
			input:      "field123:",
			wantToken:  true,
			wantValue:  "field123",
			wantRemain: ":",
		},
		{
			name:       "underscore start",
			input:      "_private:",
			wantToken:  true,
			wantValue:  "_private",
			wantRemain: ":",
		},
		{
			name:      "digit start (invalid)",
			input:     "123field:",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := keyMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenKey {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenKey)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestFunctionMatcher tests the functionMatcher function
func TestFunctionMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "simple function",
			input:      "String(1, 100)",
			wantToken:  true,
			wantValue:  "String(1, 100)",
			wantRemain: "",
		},
		{
			name:       "function no args",
			input:      "UUID()",
			wantToken:  true,
			wantValue:  "UUID()",
			wantRemain: "",
		},
		{
			name:       "function with hyphen",
			input:      "ISO-8601()",
			wantToken:  true,
			wantValue:  "ISO-8601()",
			wantRemain: "",
		},
		{
			name:       "nested parentheses",
			input:      "Pattern(\"(a|b)\")",
			wantToken:  true,
			wantValue:  "Pattern(\"(a|b)\")",
			wantRemain: "",
		},
		{
			name:       "function followed by text",
			input:      "Integer(1, 100) more",
			wantToken:  true,
			wantValue:  "Integer(1, 100)",
			wantRemain: " more",
		},
		{
			name:      "lowercase start (invalid)",
			input:     "string(1, 100)",
			wantToken: false,
		},
		{
			name:      "no opening paren (not a function)",
			input:     "UUID",
			wantToken: false,
		},
		{
			name:      "unterminated function",
			input:     "String(1, 100",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := functionMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenFunction {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenFunction)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestIdentifierMatcher tests the identifierMatcher function
func TestIdentifierMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "UUID identifier",
			input:      "UUID",
			wantToken:  true,
			wantValue:  "UUID",
			wantRemain: "",
		},
		{
			name:       "Email identifier",
			input:      "Email",
			wantToken:  true,
			wantValue:  "Email",
			wantRemain: "",
		},
		{
			name:       "identifier with hyphen",
			input:      "ISO-8601",
			wantToken:  true,
			wantValue:  "ISO-8601",
			wantRemain: "",
		},
		{
			name:       "identifier with numbers",
			input:      "Type123",
			wantToken:  true,
			wantValue:  "Type123",
			wantRemain: "",
		},
		{
			name:       "identifier followed by delimiter",
			input:      "String ",
			wantToken:  true,
			wantValue:  "String",
			wantRemain: " ",
		},
		{
			name:      "lowercase start (invalid)",
			input:     "string",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := identifierMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenIdentifier {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenIdentifier)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestNumberMatcher tests the numberMatcher function
func TestNumberMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "positive integer",
			input:      "42",
			wantToken:  true,
			wantValue:  "42",
			wantRemain: "",
		},
		{
			name:       "negative integer",
			input:      "-5",
			wantToken:  true,
			wantValue:  "-5",
			wantRemain: "",
		},
		{
			name:       "decimal number",
			input:      "3.14",
			wantToken:  true,
			wantValue:  "3.14",
			wantRemain: "",
		},
		{
			name:       "negative decimal",
			input:      "-2.5",
			wantToken:  true,
			wantValue:  "-2.5",
			wantRemain: "",
		},
		{
			name:       "number followed by delimiter",
			input:      "100,",
			wantToken:  true,
			wantValue:  "100",
			wantRemain: ",",
		},
		{
			name:       "zero",
			input:      "0",
			wantToken:  true,
			wantValue:  "0",
			wantRemain: "",
		},
		{
			name:      "letter start (invalid)",
			input:     "abc",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := numberMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenNumber {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenNumber)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestStringMatcher tests the stringMatcher function
func TestStringMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "simple string",
			input:      `"hello"`,
			wantToken:  true,
			wantValue:  `"hello"`,
			wantRemain: "",
		},
		{
			name:       "empty string",
			input:      `""`,
			wantToken:  true,
			wantValue:  `""`,
			wantRemain: "",
		},
		{
			name:       "string with spaces",
			input:      `"hello world"`,
			wantToken:  true,
			wantValue:  `"hello world"`,
			wantRemain: "",
		},
		{
			name:       "string with escaped quote",
			input:      `"quote \"inside\" string"`,
			wantToken:  true,
			wantValue:  `"quote \"inside\" string"`,
			wantRemain: "",
		},
		{
			name:       "string with backslash",
			input:      `"path\\to\\file"`,
			wantToken:  true,
			wantValue:  `"path\\to\\file"`,
			wantRemain: "",
		},
		{
			name:       "string followed by delimiter",
			input:      `"test",`,
			wantToken:  true,
			wantValue:  `"test"`,
			wantRemain: ",",
		},
		{
			name:      "unterminated string",
			input:     `"hello`,
			wantToken: false,
		},
		{
			name:      "not a string",
			input:     "hello",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := stringMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenString {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenString)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestCommentMatcher tests the commentMatcher function
func TestCommentMatcher(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantToken  bool
		wantValue  string
		wantRemain string
	}{
		{
			name:       "simple comment",
			input:      "# this is a comment",
			wantToken:  true,
			wantValue:  "# this is a comment",
			wantRemain: "",
		},
		{
			name:       "comment with newline",
			input:      "# comment\n",
			wantToken:  true,
			wantValue:  "# comment",
			wantRemain: "\n",
		},
		{
			name:       "empty comment",
			input:      "#",
			wantToken:  true,
			wantValue:  "#",
			wantRemain: "",
		},
		{
			name:       "comment with special chars",
			input:      "# @$%^&*()",
			wantToken:  true,
			wantValue:  "# @$%^&*()",
			wantRemain: "",
		},
		{
			name:      "not a comment",
			input:     "not a comment",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := commentMatcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Fatal("expected token, got nil")
				}
				if token.Kind() != TokenComment {
					t.Errorf("token kind = %q, want %q", token.Kind(), TokenComment)
				}
				if string(token.Value()) != tt.wantValue {
					t.Errorf("token value = %q, want %q", string(token.Value()), tt.wantValue)
				}

				// Check remaining content
				remaining := ""
				for {
					r, ok := stream.PeekChar()
					if !ok {
						break
					}
					stream.NextChar()
					remaining += string(r)
				}
				if remaining != tt.wantRemain {
					t.Errorf("remaining = %q, want %q", remaining, tt.wantRemain)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}

// TestTokenizerEdgeCases tests additional tokenizer edge cases
func TestTokenizerEdgeCases(t *testing.T) {
	t.Run("keyMatcher with empty stream", func(t *testing.T) {
		stream := tokenizer.NewStream("")
		token := keyMatcher(stream)
		if token != nil {
			t.Errorf("expected nil for empty stream, got %v", token)
		}
	})

	t.Run("keyMatcher with special character start", func(t *testing.T) {
		stream := tokenizer.NewStream("@invalid")
		token := keyMatcher(stream)
		if token != nil {
			t.Errorf("expected nil for special char start, got %v", token)
		}
	})

	t.Run("functionMatcher with EOF before paren", func(t *testing.T) {
		stream := tokenizer.NewStream("UUID")
		token := functionMatcher(stream)
		if token != nil {
			t.Errorf("expected nil when no paren found, got %v", token)
		}
	})

	t.Run("functionMatcher with special char before paren", func(t *testing.T) {
		stream := tokenizer.NewStream("UUID@")
		token := functionMatcher(stream)
		if token != nil {
			t.Errorf("expected nil for special char before paren, got %v", token)
		}
	})
}

// TestGetMatchers tests the GetMatchers function
func TestGetMatchers(t *testing.T) {
	matchers := GetMatchers()

	if len(matchers) != 12 {
		t.Errorf("GetMatchers() returned %d matchers, want 12", len(matchers))
	}

	// Verify all matchers work in order
	tests := []struct {
		name         string
		input        string
		wantKind     string
		matcherIndex int
	}{
		{
			name:         "comment first",
			input:        "# comment",
			wantKind:     TokenComment,
			matcherIndex: 0,
		},
		{
			name:         "true keyword",
			input:        "true",
			wantKind:     TokenTrue,
			matcherIndex: 1,
		},
		{
			name:         "false keyword",
			input:        "false",
			wantKind:     TokenFalse,
			matcherIndex: 2,
		},
		{
			name:         "null keyword",
			input:        "null",
			wantKind:     TokenNull,
			matcherIndex: 3,
		},
		{
			name:         "colon",
			input:        ":",
			wantKind:     TokenColon,
			matcherIndex: 4,
		},
		{
			name:         "dash",
			input:        "-",
			wantKind:     TokenDash,
			matcherIndex: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := matchers[tt.matcherIndex](stream)

			if token == nil {
				t.Fatal("expected token, got nil")
			}
			if token.Kind() != tt.wantKind {
				t.Errorf("token kind = %q, want %q", token.Kind(), tt.wantKind)
			}
		})
	}
}
