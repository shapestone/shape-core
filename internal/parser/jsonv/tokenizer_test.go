package jsonv

import (
	"testing"

	"github.com/shapestone/shape/internal/tokenizer"
)

func TestIdentifierMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantOk   bool
	}{
		{
			name:     "UUID identifier",
			input:    "UUID",
			expected: "UUID",
			wantOk:   true,
		},
		{
			name:     "Email identifier",
			input:    "Email",
			expected: "Email",
			wantOk:   true,
		},
		{
			name:     "ISO-8601 with hyphen",
			input:    "ISO-8601",
			expected: "ISO-8601",
			wantOk:   true,
		},
		{
			name:     "Mixed case identifier",
			input:    "MyCustomType",
			expected: "MyCustomType",
			wantOk:   true,
		},
		{
			name:     "Identifier with numbers",
			input:    "Type123",
			expected: "Type123",
			wantOk:   true,
		},
		{
			name:     "lowercase start (should fail)",
			input:    "uuid",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "digit start (should fail)",
			input:    "123Type",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Identifier followed by delimiter",
			input:    "UUID}",
			expected: "UUID",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := identifierMatcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Fatalf("identifierMatcher() returned nil, want token")
				}
				if token.Kind() != TokenIdentifier {
					t.Errorf("token.Kind() = %q, want %q", token.Kind(), TokenIdentifier)
				}
				if token.ValueString() != tt.expected {
					t.Errorf("token.ValueString() = %q, want %q", token.ValueString(), tt.expected)
				}
			} else {
				if token != nil {
					t.Errorf("identifierMatcher() = %v, want nil", token)
				}
			}
		})
	}
}

func TestFunctionMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantOk   bool
	}{
		{
			name:     "Integer with args",
			input:    "Integer(1, 100)",
			expected: "Integer(1, 100)",
			wantOk:   true,
		},
		{
			name:     "String with unbounded",
			input:    "String(5+)",
			expected: "String(5+)",
			wantOk:   true,
		},
		{
			name:     "Enum with string args",
			input:    `Enum("M", "F", "O")`,
			expected: `Enum("M", "F", "O")`,
			wantOk:   true,
		},
		{
			name:     "No args",
			input:    "Function()",
			expected: "Function()",
			wantOk:   true,
		},
		{
			name:     "Nested parens",
			input:    "Func((1, 2))",
			expected: "Func((1, 2))",
			wantOk:   true,
		},
		{
			name:     "lowercase start (should fail)",
			input:    "integer(1, 100)",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "No opening paren (should fail)",
			input:    "Integer 1, 100)",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Unclosed paren (should fail)",
			input:    "Integer(1, 100",
			expected: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := functionMatcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Fatalf("functionMatcher() returned nil, want token")
				}
				if token.Kind() != TokenFunction {
					t.Errorf("token.Kind() = %q, want %q", token.Kind(), TokenFunction)
				}
				if token.ValueString() != tt.expected {
					t.Errorf("token.ValueString() = %q, want %q", token.ValueString(), tt.expected)
				}
			} else {
				if token != nil {
					t.Errorf("functionMatcher() = %v, want nil", token)
				}
			}
		})
	}
}

func TestStringMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantOk   bool
	}{
		{
			name:     "simple string",
			input:    `"hello"`,
			expected: `"hello"`,
			wantOk:   true,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: `""`,
			wantOk:   true,
		},
		{
			name:     "string with spaces",
			input:    `"hello world"`,
			expected: `"hello world"`,
			wantOk:   true,
		},
		{
			name:     "string with escaped quote",
			input:    `"he said \"hi\""`,
			expected: `"he said \"hi\""`,
			wantOk:   true,
		},
		{
			name:     "string with backslash",
			input:    `"path\\to\\file"`,
			expected: `"path\\to\\file"`,
			wantOk:   true,
		},
		{
			name:     "unterminated string (should fail)",
			input:    `"hello`,
			expected: "",
			wantOk:   false,
		},
		{
			name:     "single quote (should fail)",
			input:    `'hello'`,
			expected: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := stringMatcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Fatalf("stringMatcher() returned nil, want token")
				}
				if token.Kind() != TokenString {
					t.Errorf("token.Kind() = %q, want %q", token.Kind(), TokenString)
				}
				if token.ValueString() != tt.expected {
					t.Errorf("token.ValueString() = %q, want %q", token.ValueString(), tt.expected)
				}
			} else {
				if token != nil {
					t.Errorf("stringMatcher() = %v, want nil", token)
				}
			}
		})
	}
}

func TestNumberMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantOk   bool
	}{
		{
			name:     "integer",
			input:    "42",
			expected: "42",
			wantOk:   true,
		},
		{
			name:     "negative integer",
			input:    "-123",
			expected: "-123",
			wantOk:   true,
		},
		{
			name:     "zero",
			input:    "0",
			expected: "0",
			wantOk:   true,
		},
		{
			name:     "decimal",
			input:    "3.14",
			expected: "3.14",
			wantOk:   true,
		},
		{
			name:     "negative decimal",
			input:    "-2.5",
			expected: "-2.5",
			wantOk:   true,
		},
		{
			name:     "exponent",
			input:    "1e10",
			expected: "1e10",
			wantOk:   true,
		},
		{
			name:     "exponent with sign",
			input:    "1.5e-10",
			expected: "1.5e-10",
			wantOk:   true,
		},
		{
			name:     "capital E",
			input:    "2.5E+3",
			expected: "2.5E+3",
			wantOk:   true,
		},
		{
			name:     "leading zero (should match single zero)",
			input:    "0123",
			expected: "0",
			wantOk:   true,
		},
		{
			name:     "letter start (should fail)",
			input:    "abc",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "just minus (should fail)",
			input:    "-",
			expected: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tokenizer.NewStream(tt.input)
			token := numberMatcher(stream)

			if tt.wantOk {
				if token == nil {
					t.Fatalf("numberMatcher() returned nil, want token")
				}
				if token.Kind() != TokenNumber {
					t.Errorf("token.Kind() = %q, want %q", token.Kind(), TokenNumber)
				}
				if token.ValueString() != tt.expected {
					t.Errorf("token.ValueString() = %q, want %q", token.ValueString(), tt.expected)
				}
			} else {
				if token != nil {
					t.Errorf("numberMatcher() = %v, want nil", token)
				}
			}
		})
	}
}

func TestGetMatchers_Tokenization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "simple object",
			input: `{"id": UUID}`,
			expected: []string{
				TokenObjectStart,
				TokenString,
				TokenColon,
				TokenIdentifier,
				TokenObjectEnd,
			},
		},
		{
			name:  "object with function",
			input: `{"name": String(1, 100)}`,
			expected: []string{
				TokenObjectStart,
				TokenString,
				TokenColon,
				TokenFunction,
				TokenObjectEnd,
			},
		},
		{
			name:  "array",
			input: `[String(1, 30)]`,
			expected: []string{
				TokenArrayStart,
				TokenFunction,
				TokenArrayEnd,
			},
		},
		{
			name:  "mixed literals",
			input: `{"active": true, "count": 42, "value": null}`,
			expected: []string{
				TokenObjectStart,
				TokenString,
				TokenColon,
				TokenTrue,
				TokenComma,
				TokenString,
				TokenColon,
				TokenNumber,
				TokenComma,
				TokenString,
				TokenColon,
				TokenNull,
				TokenObjectEnd,
			},
		},
		{
			name: "with whitespace",
			input: `{
				"id": UUID,
				"name": String(1, 100)
			}`,
			expected: []string{
				TokenObjectStart,
				TokenString,
				TokenColon,
				TokenIdentifier,
				TokenComma,
				TokenString,
				TokenColon,
				TokenFunction,
				TokenObjectEnd,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := tokenizer.NewTokenizer(GetMatchers()...)
			tok.Initialize(tt.input)

			var kinds []string
			for {
				token, ok := tok.NextToken()
				if !ok {
					break
				}
				// Skip whitespace tokens
				if token.Kind() != TokenWhitespace {
					kinds = append(kinds, token.Kind())
				}
			}

			if len(kinds) != len(tt.expected) {
				t.Errorf("got %d tokens, want %d\nGot: %v\nWant: %v", len(kinds), len(tt.expected), kinds, tt.expected)
				return
			}

			for i := range kinds {
				if kinds[i] != tt.expected[i] {
					t.Errorf("token[%d] = %q, want %q", i, kinds[i], tt.expected[i])
				}
			}
		})
	}
}
