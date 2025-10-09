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
