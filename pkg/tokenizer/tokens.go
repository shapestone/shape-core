package tokenizer

import (
	"fmt"
	"strings"
)

//
// Token - Token representation and tokenizer implementation
//

// Token represents a parsed token with its type, value, and position information.
type Token struct {
	kind   string
	value  []rune
	offset int
	row    int
	column int
}

// NewToken constructs a new Token with the given kind and value.
// Position fields (offset, row, column) are initialized to -1.
func NewToken(kind string, value []rune) *Token {
	return &Token{kind, value, -1, -1, -1}
}

// Kind returns the token's type/kind.
func (t *Token) Kind() string {
	return t.kind
}

// Value returns the token's value as a slice of runes.
func (t *Token) Value() []rune {
	return t.value
}

// ValueString returns the token's value as a string.
func (t *Token) ValueString() string {
	return string(t.value)
}

// Offset returns the token's byte offset in the source.
func (t *Token) Offset() int {
	return t.offset
}

// Row returns the token's line number (1-indexed).
func (t *Token) Row() int {
	return t.row
}

// Column returns the token's column number (1-indexed).
func (t *Token) Column() int {
	return t.column
}

// String returns a string representation of the token.
func (t *Token) String() string {
	return fmt.Sprintf("[%s: %q]", t.kind, string(t.value))
}

// Matcher is a function type that attempts to match and return a token from a stream.
// Returns nil if no match is found.
type Matcher func(stream Stream) *Token

//
// Tokenizer - Main tokenizer implementation
//

// Tokenizer processes a stream using a set of matchers to produce tokens.
// It automatically handles whitespace and supports backtracking.
type Tokenizer struct {
	matchers []Matcher
	stream   Stream
	marks    []Stream // stack of marked positions for rewinding
}

// NewTokenizer constructs a Tokenizer with the given matchers.
// WhiteSpaceMatcher is automatically prepended to consume whitespace.
func NewTokenizer(matchers ...Matcher) Tokenizer {
	newMatchers := make([]Matcher, 0)
	newMatchers = append(newMatchers, WhiteSpaceMatcher)
	newMatchers = append(newMatchers, matchers...)
	return Tokenizer{
		matchers: newMatchers,
		marks:    make([]Stream, 0),
	}
}

// NewTokenizerWithoutWhitespace constructs a Tokenizer with the given matchers
// WITHOUT automatically prepending WhiteSpaceMatcher.
// Use this for parsers that handle whitespace themselves (e.g., JSON parser
// that skips whitespace inline to avoid creating tokens that are immediately discarded).
func NewTokenizerWithoutWhitespace(matchers ...Matcher) Tokenizer {
	return Tokenizer{
		matchers: matchers,
		marks:    make([]Stream, 0),
	}
}

// Initialize initializes the tokenizer with the given input string.
func (t *Tokenizer) Initialize(input string) {
	t.stream = NewStream(input)
}

// InitializeFromStream initializes the tokenizer with a pre-configured stream.
// This allows using streams created with NewStreamFromReader for parsing large files.
func (t *Tokenizer) InitializeFromStream(stream Stream) {
	t.stream = stream
}

// Mark pushes the current stream position onto the marks stack for later rewinding.
func (t *Tokenizer) Mark() {
	t.marks = append(t.marks, t.stream.Clone())
}

// Rewind restores the stream to the most recently marked position.
// Returns false if there are no marks to rewind to.
func (t *Tokenizer) Rewind() bool {
	if len(t.marks) == 0 {
		return false
	}
	lastIdx := len(t.marks) - 1
	marked := t.marks[lastIdx]
	t.marks = t.marks[:lastIdx] // pop the mark
	t.stream.Match(marked)
	return true
}

// Tokenize applies NextToken until the end of stream or until a token cannot be read.
// Returns:
// - A slice of tokens
// - true if the stream was fully consumed (EOS reached)
func (t *Tokenizer) Tokenize() ([]Token, bool) {
	tokens := make([]Token, 0)
	for {
		token, ok := t.NextToken()
		if !ok {
			break
		}
		tokens = append(tokens, *token)
	}
	return tokens, t.stream.IsEos()
}

// TokenizeToString tokenizes the input and returns a debug string representation.
func (t *Tokenizer) TokenizeToString(separator string) string {
	tokens, eos := t.Tokenize()
	var sb strings.Builder
	for _, token := range tokens {
		sb.WriteString(token.String())
		if len(separator) > 0 {
			sb.WriteString(separator)
		}
	}
	if eos {
		sb.WriteString(`[EOS]`)
	} else {
		sb.WriteString(`[Stream...]`)
	}
	return sb.String()
}

// NextToken applies each matcher in order and returns the first successful token.
// The stream is advanced by the token's length.
// Returns nil, false if no matcher succeeds.
func (t *Tokenizer) NextToken() (*Token, bool) {
	if !t.hasMoreTokens() {
		return nil, false
	}

	// Save the current position for token metadata
	offset := t.stream.GetOffset()
	row := t.stream.GetRow()
	column := t.stream.GetColumn()

	// Save location for rewinding on failed matches
	startLocation := t.stream.GetLocation()

	for _, matcher := range t.matchers {
		// Try the matcher directly on the stream (no cloning!)
		token := matcher(t.stream)
		if token != nil {
			// Match succeeded - but the matcher may have consumed extra characters
			// to determine where the match ends. We need to position the stream
			// exactly at the end of the matched token value.
			// Use MatchChars to correctly position the stream based on token value.
			t.stream.SetLocation(startLocation)
			if t.stream.MatchChars(token.value) {
				// Stream is now correctly positioned
				token.offset = offset
				token.row = row
				token.column = column
				return token, true
			}
			// This shouldn't happen, but if MatchChars fails, try next matcher
		}
		// Match failed - rewind stream to start position for next matcher
		t.stream.SetLocation(startLocation)
	}

	return nil, false
}

// PeekToken applies each matcher in order and returns the first successful token
// without advancing the stream.
// Returns nil, false if no matcher succeeds.
func (t *Tokenizer) PeekToken() (*Token, bool) {
	if !t.hasMoreTokens() {
		return nil, false
	}

	// Save the current location to restore after peeking
	startLocation := t.stream.GetLocation()

	for _, matcher := range t.matchers {
		// Try the matcher directly on the stream (no cloning!)
		token := matcher(t.stream)
		if token != nil {
			// Match succeeded - restore position (peek doesn't advance) and return
			t.stream.SetLocation(startLocation)
			return token, true
		}
		// Match failed - rewind stream to start position for next matcher
		t.stream.SetLocation(startLocation)
	}

	// All matchers failed - position is already restored
	return nil, false
}

// GetRow returns the current stream row position.
func (t *Tokenizer) GetRow() int {
	return t.stream.GetRow()
}

// GetColumn returns the current stream column position.
func (t *Tokenizer) GetColumn() int {
	return t.stream.GetColumn()
}

// hasMoreTokens returns true if the stream is not at end.
func (t *Tokenizer) hasMoreTokens() bool {
	return !t.stream.IsEos()
}
