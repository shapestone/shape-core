package tokenizer

import (
	"github.com/google/uuid"
)

//
// Stream - Stream abstraction with UTF-8 support and position tracking
//

// Stream defines the interface for a character stream with position tracking.
// It provides UTF-8 support, character peeking/fetching, and pattern matching capabilities.
type Stream interface {
	Clone() Stream
	Match(cs Stream)
	PeekChar() (rune, bool)
	NextChar() (rune, bool)
	MatchChars([]rune) bool
	IsEos() bool
	GetRow() int
	GetOffset() int
	GetColumn() int
	Reset()
}

// NewStream creates a new stream instance from the provided string.
// The stream supports UTF-8 encoding and tracks position (offset, line, column).
func NewStream(str string) Stream {
	runes := []rune(str)
	return &streamImpl{
		uuid:   uuid.New(),
		data:   runes,
		length: len(runes),
		location: location{
			cursor: 0,
			row:    1,
			column: 1,
		},
	}
}

// streamImpl is the internal implementation of the Stream interface.
type streamImpl struct {
	uuid     uuid.UUID
	data     []rune
	length   int
	location location
}

// location holds position information within the stream.
type location struct {
	cursor int // byte offset
	row    int // line number (1-indexed)
	column int // column number (1-indexed)
}

// Clone creates a copy of the stream for backtracking support.
func (s *streamImpl) Clone() Stream {
	return &streamImpl{
		uuid:     s.uuid,
		data:     s.data,
		length:   s.length,
		location: s.location,
	}
}

// Match updates this stream's location to match another stream's location.
// Both streams must be clones of each other (same UUID).
func (s *streamImpl) Match(other Stream) {
	otherImpl := other.(*streamImpl)
	if s.uuid != otherImpl.uuid {
		panic("trying to match two different streams")
	}
	s.location = otherImpl.location
}

// PeekChar returns the next rune without advancing the stream.
func (s *streamImpl) PeekChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}
	r := s.data[s.location.cursor]
	return r, true
}

// NextChar reads and returns the next rune, advancing the stream position.
// Automatically tracks newlines for row/column position.
func (s *streamImpl) NextChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}
	r := s.data[s.location.cursor]
	s.location.cursor += 1
	s.location.column += 1
	if r == '\n' {
		s.location.row += 1
		s.location.column = 1
	}
	return r, true
}

// MatchChars attempts to match a rune sequence against the stream.
// If successful, the stream is advanced. If not, the stream position is unchanged.
func (s *streamImpl) MatchChars(match []rune) bool {
	origLocation := s.location
	for _, mr := range match {
		sr, ok := s.NextChar()
		if !ok || mr != sr {
			s.location = origLocation
			return false
		}
	}
	return true
}

// IsEos returns true if the cursor has reached the end of stream.
func (s *streamImpl) IsEos() bool {
	return s.location.cursor >= s.length
}

// GetOffset returns the current byte offset within the stream.
func (s *streamImpl) GetOffset() int {
	return s.location.cursor
}

// GetRow returns the current line number (1-indexed).
func (s *streamImpl) GetRow() int {
	return s.location.row
}

// GetColumn returns the current column number (1-indexed).
func (s *streamImpl) GetColumn() int {
	return s.location.column
}

// Reset resets the stream to the beginning (offset 0, row 1, column 1).
func (s *streamImpl) Reset() {
	s.location.cursor = 0
	s.location.row = 1
	s.location.column = 1
}

//
// Pattern Matching - Higher-order functions for composing stream matchers
//

// Pattern is a function type that matches patterns in a stream.
// It returns the matched runes and a success flag.
// The stream is mutated only if the match succeeds.
//
// Pattern matchers can be composed using higher-order functions like:
// - Sequence: matches patterns in order
// - OneOf: matches the first successful pattern
// - Optional: matches if possible, but always succeeds
type Pattern func(stream Stream) ([]rune, bool)

// CharMatcher creates a pattern that matches a single character.
func CharMatcher(char rune) Pattern {
	return func(stream Stream) ([]rune, bool) {
		if r, ok := stream.NextChar(); ok && r == char {
			return []rune{char}, true
		}
		return nil, false
	}
}

// StringMatcher creates a pattern that matches a literal string.
func StringMatcher(literal string) Pattern {
	var rLiteral = []rune(literal)
	return func(stream Stream) ([]rune, bool) {
		var value []rune

		for _, ch := range rLiteral {
			if r, ok := stream.NextChar(); ok && r == ch {
				value = append(value, r)
				continue
			}
			break
		}

		if len(value) != len(rLiteral) {
			return nil, false
		}
		return value, true
	}
}

// Sequence applies patterns sequentially. All must succeed for success.
func Sequence(patterns ...Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		var value []rune
		for _, pattern := range patterns {
			ra, ok := pattern(stream)
			if !ok {
				return nil, false
			}
			value = append(value, ra...)
		}
		return value, true
	}
}

// OneOf tries patterns in order and returns the first match.
// Uses backtracking (stream cloning) to try each pattern.
func OneOf(patterns ...Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		for _, pattern := range patterns {
			cs := stream.Clone() // enable backtracking
			ra, ok := pattern(cs)
			if ok {
				stream.Match(cs) // update parent stream
				return ra, true
			}
		}
		return nil, false
	}
}

// Optional tries to match a pattern but always succeeds.
// Uses backtracking if the pattern doesn't match.
func Optional(pattern Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		cs := stream.Clone()
		ra, ok := pattern(cs)
		if ok {
			stream.Match(cs)
			return ra, true
		}
		return nil, true // always succeeds
	}
}
