package tokenizer

import (
	"testing"
)

func TestStringMatcherShouldReturnAToken(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(StringMatcherFunc(`List`, `list`))
	stream := "list"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := StripMargin(`
		|[List: "list"]
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestCharMatcherFunc(t *testing.T) {
	tests := []struct {
		name      string
		char      rune
		input     string
		tokenName string
		wantToken bool
	}{
		{
			name:      "match single char",
			char:      '{',
			input:     "{",
			tokenName: "LBRACE",
			wantToken: true,
		},
		{
			name:      "match char in sequence",
			char:      '[',
			input:     "[abc",
			tokenName: "LBRACKET",
			wantToken: true,
		},
		{
			name:      "no match",
			char:      '}',
			input:     "{",
			tokenName: "RBRACE",
			wantToken: false,
		},
		{
			name:      "empty input",
			char:      'x',
			input:     "",
			tokenName: "CHAR",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := CharMatcherFunc(tt.tokenName, tt.char)
			stream := NewStream(tt.input)

			token := matcher(stream)

			if tt.wantToken {
				if token == nil {
					t.Error("expected token, got nil")
					return
				}
				if token.Kind() != tt.tokenName {
					t.Errorf("token.Kind() = %q, want %q", token.Kind(), tt.tokenName)
				}
				value := token.Value()
				if len(value) != 1 || value[0] != tt.char {
					t.Errorf("token.Value() = %v, want [%c]", value, tt.char)
				}
			} else {
				if token != nil {
					t.Errorf("expected nil token, got %v", token)
				}
			}
		})
	}
}
