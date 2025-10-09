package streams

import (
	"github.com/google/uuid"
)

//
// Streams
//

/*
 * The Stream interface provides the following capabilities
 * - UTF8 support
 * - Peek and fetching the next character
 * - Character sequence matching capabilities
 */

// The Stream interface defines public functions of a data stream instance
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

// The NewStream function creates a new stream instance from the provided string argument
//
// Stream: an instance of a data stream
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

// The streamImpl struct contains a reference to the data stream
//
// data: rune sequence representing the data
// length: length of the data rune sequence
// cursor: cursor location within the data stream
// row: row location of the cursor within the data stream
// column: column location of the cursor within the data stream
type streamImpl struct {
	uuid     uuid.UUID
	data     []rune
	length   int
	location location
}

// The location type holds information about
type location struct {
	cursor int
	row    int
	column int
}

// The Clone function creates a copy of the data stream. This is useful to implement, e.g., backtracking.
//
// stream: a copy of the initial data stream
func (s *streamImpl) Clone() Stream {
	return &streamImpl{
		uuid:     s.uuid,
		data:     s.data,
		length:   s.length,
		location: s.location,
	}
}

// The Match function updates the location object with the location from the passed in stream argument
func (s *streamImpl) Match(other Stream) {
	otherImpl := other.(*streamImpl)
	if s.uuid != otherImpl.uuid {
		panic("trying to match two different streams")
	}
	s.location = otherImpl.location
}

// The PeekChar function peeks at the next rune in the data stream
//
// rune: the next character
// bool: true if a character could be read
func (s *streamImpl) PeekChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}
	r := s.data[s.location.cursor]
	return r, true
}

// The NextChar function reads the next rune from the data stream
//
// rune: the next character
// bool: true if a character could be read
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

// The MatchChars function matches the match rune sequence against the data stream and advances the stream if the rune
// sequence matches
//
// match: rune sequence used to advance the data stream
// bool: true if the match rune sequence could be matched
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

// The IsEos function returns true if the cursor has reached the end of the data stream
//
// bool: true if the cursor is at the end of the data stream
func (s *streamImpl) IsEos() bool {
	return s.location.cursor >= s.length
}

// The GetOffset function returns the cursor offset within the data stream
//
// int: cursor offset within the data stream
func (s *streamImpl) GetOffset() int {
	return s.location.cursor
}

// The GetRow function returns the cursor row location within the data stream
//
// int: cursor row location within the data stream
func (s *streamImpl) GetRow() int {
	return s.location.row
}

// The GetColumn function returns the cursor column location within the data stream
//
// int: cursor column location within the data stream
func (s *streamImpl) GetColumn() int {
	return s.location.column
}

// The Reset function resets the cursor to the beginning of the data stream. The row and column values are reset as well
func (s *streamImpl) Reset() {
	s.location.cursor = 0
	s.location.row = 1
	s.location.column = 1
}

//
// Private functions
//

// The getUuid function returns a unique uuid for this data stream
//
// uuid.UUID the unique id
func (s *streamImpl) getUuid() uuid.UUID {
	return s.uuid
}

// The getLocation function returns the location structure for this data stream
//
// location: stream location data
func (s *streamImpl) getLocation() location {
	return s.location
}
