package tokenizer

import (
	"testing"
)

//
// Tokenizer Tests
//

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
	expected := StripMargin(`
		|[Whitespace: " \r\n\t\v\f"]
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
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
	expected := StripMargin(`
		|[Whitespace: "   "]
		|[Alpha: "abc"]
		|[Whitespace: "  "]
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
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
	expected := StripMargin(`
		|[Alpha: "abc"]
		|[Numeric: "123"]
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
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
	expected := StripMargin(`
		|[Alpha: "abc"]
		|[Stream...]
	`)

	diff, tdOk := Diff(expected, actual)
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
	expected := StripMargin(`
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
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
	expected := StripMargin(`
		|[Alpha: "abc"]
		|[Numeric: "123"]
		|[EOS]
	`)

	diff, tdOk := Diff(expected, actual)
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
// Custom Matchers for testing
//

func alphaMatcher(stream Stream) *Token {
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

func numericMatcher(stream Stream) *Token {
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

//
// Token Accessor Tests
//

func TestTokenAccessors(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher)
	tokenizer.Initialize("abc")

	// When
	token, ok := tokenizer.NextToken()

	// Then
	if !ok {
		t.Fatalf("Expected token to be found")
	}

	if token.Kind() != "Alpha" {
		t.Fatalf("Expected kind to be 'Alpha', got %s", token.Kind())
	}

	if string(token.Value()) != "abc" {
		t.Fatalf("Expected value to be 'abc', got %s", string(token.Value()))
	}

	if token.ValueString() != "abc" {
		t.Fatalf("Expected value string to be 'abc', got %s", token.ValueString())
	}

	if token.Offset() != 0 {
		t.Fatalf("Expected offset to be 0, got %d", token.Offset())
	}

	if token.Row() != 1 {
		t.Fatalf("Expected row to be 1, got %d", token.Row())
	}

	if token.Column() != 1 {
		t.Fatalf("Expected column to be 1, got %d", token.Column())
	}
}

//
// Mark/Rewind Tests
//

func TestMarkAndRewind(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher, numericMatcher)
	tokenizer.Initialize("abc123")

	// When - Mark position, consume token, then rewind
	tokenizer.Mark()
	token1, _ := tokenizer.NextToken()
	rewound := tokenizer.Rewind()

	// Then
	if !rewound {
		t.Fatalf("Expected rewind to succeed")
	}

	// When - Read token again after rewind
	token2, ok := tokenizer.NextToken()

	// Then
	if !ok {
		t.Fatalf("Expected to read token after rewind")
	}

	if token1.Kind() != token2.Kind() {
		t.Fatalf("Expected same token after rewind, got %s and %s", token1.Kind(), token2.Kind())
	}
}

func TestRewindWithoutMark(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher)
	tokenizer.Initialize("abc")

	// When
	rewound := tokenizer.Rewind()

	// Then
	if rewound {
		t.Fatalf("Expected rewind to fail when no mark exists")
	}
}

func TestGetRowAndColumn(t *testing.T) {
	// Given
	tokenizer := NewTokenizer(alphaMatcher)
	tokenizer.Initialize("abc\nde")

	// When
	tokenizer.NextToken() // consume "abc"
	tokenizer.NextToken() // consume whitespace (newline)

	// Then
	if tokenizer.GetRow() != 2 {
		t.Fatalf("Expected row to be 2, got %d", tokenizer.GetRow())
	}

	if tokenizer.GetColumn() != 1 {
		t.Fatalf("Expected column to be 1, got %d", tokenizer.GetColumn())
	}
}
