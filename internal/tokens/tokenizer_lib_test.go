package tokens

import (
	"github.com/shapestone/shape/internal/text"
	"testing"
)

func TestStringMatcherShouldReturnAToken(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(StringMatcher(`List`, `list`))
	stream := "list"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[List: "list"]
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}
