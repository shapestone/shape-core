package tokens

import (
	"github.com/shapestone/shape/internal/streams"
	"github.com/shapestone/shape/internal/text"
	"testing"
)

func TestTokenizerInitialization(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := "abc123"

	// When
	tokenizer.Initialize(stream)

	// Then
	if tokenizer.stream.GetOffset() != 0 {
		t.Fatalf("Expected tokenizer stream offset to be 0, got %d", tokenizer.stream.GetOffset())
	}
}

func TestNextTokenOnWhitespaceShouldYieldAToken(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := " \r\n\t\v\f"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[Whitespace: " \r\n\t\v\f"]
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestNextTokenOnAlphaLiteralWithWhitespaceBeforeAndAfterShouldYieldThreeTokens(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher)
	stream := "   abc  "

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[Whitespace: "   "]
		|[Alpha: "abc"]
		|[Whitespace: "  "]
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestRepeatedNextTokenShouldYieldTwoTokens(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := "abc123"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[Alpha: "abc"]
		|[Numeric: "123"]
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestNextTokenShouldNotMatchShouldFail(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher)
	stream := "abc123"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[Alpha: "abc"]
		|[Stream...]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestNextTokenOnEmptyInputStreamShouldFail(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := ""

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestRepeatedPeeksShouldYieldTheSameToken(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := "abc123"

	// When
	tokenizer.Initialize(stream)
	actual := tokenizer.TokenizeToString("\n")

	// Then
	expected := text.StripMargin(`
		|[Alpha: "abc"]
		|[Numeric: "123"]
		|[EOS]
	`)

	diff, tdOk := text.Diff(expected, actual)
	if !tdOk {
		t.Fatalf("Tokenization validation error: \n%v", diff)
	}
}

func TestPeekShouldOnNonMatchShouldFail(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(numericMatcher)
	stream := "abc123"

	// When
	tokenizer.Initialize(stream)
	token, ok := tokenizer.PeekToken()

	// Then
	if ok {
		t.Fatalf("Expected tokenizer stream to not have a token")
	}

	if token != nil {
		t.Fatalf("Expected token to be nil, got %+v", token)
	}
}

func TestPeekOnEmptyInputStreamShouldFail(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	stream := ""

	// When
	tokenizer.Initialize(stream)
	token, ok := tokenizer.PeekToken()

	// Then
	if ok {
		t.Fatalf("Expected tokenizer stream to not have a token")
	}

	if token != nil {
		t.Fatalf("Expected token to be nil, got %+v", token)
	}
}

//
// Custom Matchers
//

func alphaMatcher(stream streams.Stream) *Token {
	var value []rune
	for {
		if r, ok := stream.NextChar(); ok {
			if r >= 'a' && r <= 'z' {
				value = append(value, r)
				continue
			}
		}
		break
	}
	if len(value) == 0 {
		return nil
	}
	return NewToken(`Alpha`, value)
}

func numericMatcher(stream streams.Stream) *Token {
	var value []rune
	for {
		if r, ok := stream.NextChar(); ok {
			if r >= '0' && r <= '9' {
				value = append(value, r)
				continue
			}
		}
		break
	}
	if len(value) == 0 {
		return nil
	}
	return NewToken(`Numeric`, value)
}
