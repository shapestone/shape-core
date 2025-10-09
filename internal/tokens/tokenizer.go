package tokens

import (
	"fmt"
	"github.com/shapestone/shape/internal/streams"
	"strings"
)

//
// Tokenizer
//

// The Token struct represent a parsed token
//
// - kind: the token type
// - value: the token value
type Token struct {
	kind   string
	value  []rune
	offset int
	row    int
	column int
}

// The NewToken function constructs a new Token instance. Offset, row, and column are initialized to -1.
func NewToken(kind string, value []rune) *Token {
	return &Token{kind, value, -1, -1, -1}
}

// The String function returns a literal representation of a token
func (t *Token) String() string {
	return fmt.Sprintf("[%s: %q]", t.kind, string(t.value))
}

// The Matcher function signature is used to create custom tokenizer functions
type Matcher func(stream streams.Stream) *Token

// The Tokenizer struct contains an array of Matcher's and a reference to a Stream instance
type Tokenizer struct {
	matchers []Matcher
	stream   streams.Stream
	marks    []streams.Stream // Stack of marked positions for rewinding
}

// The NewTokenizer function constructs a Tokenizer instance
func NewTokenizer(matchers ...Matcher) Tokenizer {
	newMatchers := make([]Matcher, 0)
	newMatchers = append(newMatchers, WhiteSpaceMatcher)
	newMatchers = append(newMatchers, matchers...)
	return Tokenizer{
		matchers: newMatchers,
		marks:    make([]streams.Stream, 0),
	}
}

// The Initialize function initializes a Stream instance with the Tokenizer
func (t *Tokenizer) Initialize(input string) {
	t.stream = streams.NewStream(input)
}

// The Mark function pushes the current stream position onto the marks stack
func (t *Tokenizer) Mark() {
	t.marks = append(t.marks, t.stream.Clone())
}

// The Rewind function restores the stream to the most recently marked position
// Returns false if there are no marks to rewind to
func (t *Tokenizer) Rewind() bool {
	if len(t.marks) == 0 {
		return false
	}
	lastIdx := len(t.marks) - 1
	marked := t.marks[lastIdx]
	t.marks = t.marks[:lastIdx] // Pop the used mark
	t.stream.Match(marked)
	return true
}

// The Tokenize function applies the NextToken until the end of the stream or until a token could not be read
// It returns two values:
// - A sequence of Tokens// - A boolean value, true if the end of stream was reached, i.e., the stream was fully consumed and tokenized
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

// The TokenizeToString function applies the NextToken until the end of the stream or until a token could not be read
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

// The NextToken function applies each customer Matcher in order and returns a Token if there is a match
// If there is a match the stream associated with the Tokenizer is advanced with the token size
func (t *Tokenizer) NextToken() (*Token, bool) {
	if !t.hasMoreTokens() {
		return nil, false
	}

	offset := t.stream.GetOffset()
	row := t.stream.GetRow()
	column := t.stream.GetColumn()
	for _, matcher := range t.matchers {
		cs := t.stream.Clone()
		token := matcher(cs)
		if token != nil && t.stream.MatchChars(token.value) {
			token.offset = offset
			token.row = row
			token.column = column
			return token, true
		}
	}

	return nil, false
}

// The PeekToken function applies each customer Matcher in order and returns a Token if there is a match
// If there is a match the stream associated with the Tokenizer is not advanced
func (t *Tokenizer) PeekToken() (*Token, bool) {
	if !t.hasMoreTokens() {
		return nil, false
	}

	for _, matcher := range t.matchers {
		cs := t.stream.Clone()
		token := matcher(cs)
		if token != nil && t.stream.MatchChars(token.value) {
			return token, true
		}
	}

	return nil, false
}

// The GetRow function returns the current stream row position
func (t *Tokenizer) GetRow() int {
	return t.stream.GetRow()
}

// The GetColumn function returns the current stream column position
func (t *Tokenizer) GetColumn() int {
	return t.stream.GetColumn()
}

//
// Helper functions
//

// The hasMoreTokens function returns true if the associated Stream is not at the end
func (t *Tokenizer) hasMoreTokens() bool {
	return !t.stream.IsEos()
}

func (t *Tokenizer) skipWhitespace() {}

func (t *Tokenizer) isEof() {}

func (t *Tokenizer) reset() {
	t.stream.Reset()
	t.marks = t.marks[:0] // Clear the marks stack when resetting
}
