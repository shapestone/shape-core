package tokenizer

import (
	"io"
	"unicode/utf8"

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
	GetLocation() Location
	SetLocation(Location)
}

// ByteStream extends Stream with high-performance byte-level operations.
// This interface provides optimized primitives for parsing ASCII-heavy formats like JSON.
// Use byte-level methods for structural tokens and ASCII scanning, falling back to
// rune-level methods only when processing UTF-8 content.
type ByteStream interface {
	Stream // Embed all rune-based methods for backward compatibility

	// Byte-level access (fast path for ASCII)
	PeekByte() (byte, bool)
	NextByte() (byte, bool)
	PeekBytes(n int) []byte

	// Optimized scanning primitives
	SkipWhitespace()
	SkipUntil(delim byte) int
	FindByte(b byte) int
	FindAny(chars []byte) int

	// Zero-copy extraction
	SliceFrom(start int) []byte
	BytePosition() int

	// Access to underlying byte data
	RemainingBytes() []byte
}

// NewStream creates a new stream instance from the provided string.
// The stream supports UTF-8 encoding and tracks position (offset, line, column).
// Returns a ByteStream for access to both rune and byte-level operations.
func NewStream(str string) Stream {
	bytes := []byte(str)
	runes := []rune(str)

	// Check if stream is pure ASCII (fast path optimization)
	isASCIIOnly := len(runes) == len(bytes)

	// Build rune->byte position mapping for synchronization
	var runeToBytePos []int
	if !isASCIIOnly {
		runeToBytePos = make([]int, len(runes)+1) // +1 for EOF position
		byteIdx := 0
		for runeIdx := range runes {
			runeToBytePos[runeIdx] = byteIdx
			// Calculate byte width of this rune
			r := runes[runeIdx]
			if r < 0x80 {
				byteIdx += 1
			} else if r < 0x800 {
				byteIdx += 2
			} else if r < 0x10000 {
				byteIdx += 3
			} else {
				byteIdx += 4
			}
		}
		runeToBytePos[len(runes)] = byteIdx // EOF position
	}

	return &streamImpl{
		uuid:         uuid.New(),
		bytes:        bytes,
		data:         runes,
		length:       len(runes),
		bytePos:      0,
		totalSize:    len(bytes),
		runeToBytePos: runeToBytePos,
		isASCIIOnly:   isASCIIOnly,
		location: Location{
			Cursor: 0,
			Row:    1,
			Column: 1,
		},
	}
}

// streamImpl is the internal implementation of the Stream and ByteStream interfaces.
// It maintains both byte and rune representations for optimal performance:
//   - bytes: Original byte data for fast ASCII scanning
//   - data: Decoded runes for UTF-8 character processing
//   - bytePos: Current position in bytes array
//   - location.Cursor: Current position in runes array
//   - runeToBytePos: Maps rune index to byte position for synchronization
//   - isASCIIOnly: True if all content is ASCII (len(runes) == len(bytes))
type streamImpl struct {
	uuid         uuid.UUID
	bytes        []byte // Original byte data
	data         []rune // Decoded rune data
	length       int    // Number of runes
	totalSize    int    // Number of bytes
	bytePos      int    // Current byte position
	location     Location
	runeToBytePos []int  // Maps rune index -> byte offset for sync
	isASCIIOnly   bool   // True if stream is pure ASCII
}

// Location holds position information within the stream.
type Location struct {
	Cursor int // byte offset
	Row    int // line number (1-indexed)
	Column int // column number (1-indexed)
}

// Clone creates a copy of the stream for backtracking support.
func (s *streamImpl) Clone() Stream {
	return &streamImpl{
		uuid:         s.uuid,
		bytes:        s.bytes,
		data:         s.data,
		length:       s.length,
		totalSize:    s.totalSize,
		bytePos:      s.bytePos,
		location:     s.location,
		runeToBytePos: s.runeToBytePos, // Shared mapping (read-only)
		isASCIIOnly:   s.isASCIIOnly,
	}
}

// Match updates this stream's location to match another stream's location.
// Both streams must be clones of each other (same UUID).
func (s *streamImpl) Match(other Stream) {
	otherImpl, ok := other.(*streamImpl)
	if !ok {
		panic("type assertion failed: expected *streamImpl")
	}
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
	r := s.data[s.location.Cursor]
	return r, true
}

// NextChar reads and returns the next rune, advancing the stream position.
// Automatically tracks newlines for row/column position.
// Synchronizes bytePos using the rune->byte mapping.
func (s *streamImpl) NextChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}
	r := s.data[s.location.Cursor]
	s.location.Cursor += 1
	s.location.Column += 1

	// Synchronize byte position
	if s.isASCIIOnly {
		// Fast path: 1 rune = 1 byte
		s.bytePos = s.location.Cursor
	} else if s.location.Cursor < len(s.runeToBytePos) {
		// UTF-8: Use mapping
		s.bytePos = s.runeToBytePos[s.location.Cursor]
	}

	if r == '\n' {
		s.location.Row += 1
		s.location.Column = 1
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
	return s.location.Cursor >= s.length
}

// GetOffset returns the current byte offset within the stream.
func (s *streamImpl) GetOffset() int {
	return s.location.Cursor
}

// GetRow returns the current line number (1-indexed).
func (s *streamImpl) GetRow() int {
	return s.location.Row
}

// GetColumn returns the current column number (1-indexed).
func (s *streamImpl) GetColumn() int {
	return s.location.Column
}

// Reset resets the stream to the beginning (offset 0, row 1, column 1).
func (s *streamImpl) Reset() {
	s.location.Cursor = 0
	s.location.Row = 1
	s.location.Column = 1
	s.bytePos = 0  // Already at position 0
}

// GetLocation returns the current position in the stream.
func (s *streamImpl) GetLocation() Location {
	return s.location
}

// SetLocation sets the stream position to the specified location.
// Also synchronizes bytePos to match the rune cursor.
func (s *streamImpl) SetLocation(loc Location) {
	s.location = loc
	// Sync bytePos from rune cursor
	if s.isASCIIOnly {
		s.bytePos = loc.Cursor
	} else if loc.Cursor < len(s.runeToBytePos) {
		s.bytePos = s.runeToBytePos[loc.Cursor]
	}
}

//
// ByteStream Implementation - High-performance byte-level operations
//

// PeekByte returns the next byte without advancing (for ASCII fast path).
func (s *streamImpl) PeekByte() (byte, bool) {
	if s.bytePos >= s.totalSize {
		return 0, false
	}
	return s.bytes[s.bytePos], true
}

// syncRuneCursorFromBytePos updates location.Cursor to match current bytePos.
// For ASCII-only streams, this is trivial (cursor = bytePos).
// For UTF-8 streams, uses binary search on the rune->byte mapping.
func (s *streamImpl) syncRuneCursorFromBytePos() {
	// Fast path: ASCII-only content (1 byte = 1 rune)
	if s.isASCIIOnly {
		s.location.Cursor = s.bytePos
		return
	}

	// UTF-8: Check if we're already in sync (common after NextChar)
	if s.location.Cursor < len(s.runeToBytePos) &&
	   s.runeToBytePos[s.location.Cursor] == s.bytePos {
		return
	}

	// Binary search to find rune index for current bytePos
	left, right := 0, len(s.runeToBytePos)-1
	for left < right {
		mid := (left + right + 1) / 2
		if s.runeToBytePos[mid] <= s.bytePos {
			left = mid
		} else {
			right = mid - 1
		}
	}
	s.location.Cursor = left
}

// NextByte reads and returns the next byte, advancing position.
// Synchronizes the rune cursor to maintain compatibility with mixed byte/rune usage.
func (s *streamImpl) NextByte() (byte, bool) {
	if s.bytePos >= s.totalSize {
		return 0, false
	}
	b := s.bytes[s.bytePos]
	s.bytePos++

	// Synchronize rune cursor with byte position
	s.syncRuneCursorFromBytePos()

	// Update location tracking for newlines
	if b == '\n' {
		s.location.Row++
		s.location.Column = 1
	} else {
		s.location.Column++
	}

	return b, true
}

// PeekBytes returns the next n bytes without advancing (zero-copy slice).
func (s *streamImpl) PeekBytes(n int) []byte {
	if s.bytePos+n > s.totalSize {
		n = s.totalSize - s.bytePos
	}
	return s.bytes[s.bytePos : s.bytePos+n]
}

// SkipWhitespace advances past ASCII whitespace characters (space, tab, LF, CR).
func (s *streamImpl) SkipWhitespace() {
	for s.bytePos < s.totalSize {
		b := s.bytes[s.bytePos]
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			// Sync rune cursor before returning
			s.syncRuneCursorFromBytePos()
			return
		}
		s.bytePos++

		if b == '\n' {
			s.location.Row++
			s.location.Column = 1
		} else {
			s.location.Column++
		}
	}
	// Sync at end
	s.syncRuneCursorFromBytePos()
}

// SkipUntil advances until finding the delimiter byte, returning bytes skipped.
// Does not consume the delimiter.
func (s *streamImpl) SkipUntil(delim byte) int {
	start := s.bytePos
	for s.bytePos < s.totalSize {
		if s.bytes[s.bytePos] == delim {
			return s.bytePos - start
		}
		s.bytePos++
	}
	return s.bytePos - start
}

// FindByte searches for a byte from current position, returning offset from current pos.
// Returns -1 if not found. Does not advance stream.
func (s *streamImpl) FindByte(b byte) int {
	for i := s.bytePos; i < s.totalSize; i++ {
		if s.bytes[i] == b {
			return i - s.bytePos
		}
	}
	return -1
}

// FindAny searches for any byte in chars, returning offset to first match.
// Returns -1 if none found. Does not advance stream.
func (s *streamImpl) FindAny(chars []byte) int {
	for i := s.bytePos; i < s.totalSize; i++ {
		for _, c := range chars {
			if s.bytes[i] == c {
				return i - s.bytePos
			}
		}
	}
	return -1
}

// SliceFrom returns a zero-copy byte slice from given start position to current position.
func (s *streamImpl) SliceFrom(start int) []byte {
	if start < 0 || start > s.bytePos {
		return nil
	}
	return s.bytes[start:s.bytePos]
}

// BytePosition returns the current byte offset in the stream.
func (s *streamImpl) BytePosition() int {
	return s.bytePos
}

// RemainingBytes returns the unread portion of the byte stream (zero-copy).
func (s *streamImpl) RemainingBytes() []byte {
	return s.bytes[s.bytePos:]
}

//
// Buffered Stream Implementation - For large files and streaming data
//

const (
	// bufferSize is the maximum number of runes to keep in the sliding window buffer.
	// This is set to 64KB of runes to allow reasonable backtracking while maintaining
	// constant memory usage for large files.
	bufferSize = 64 * 1024

	// readChunkSize is the number of bytes to read from the io.Reader at a time.
	// This is set to 8KB to balance between read performance and memory overhead.
	readChunkSize = 8 * 1024
)

// sharedBuffer holds the buffer state that must be shared across all clones.
// This ensures that when any clone or the original stream modifies the buffer
// (through refilling or discarding), all instances see the updated state.
type sharedBuffer struct {
	data  []rune // The actual sliding window buffer
	start int64  // Global offset where buffer starts
	eof   bool   // True when reader has reached EOF
	err   error  // Error from reader, if any
}

// NewStreamFromReader creates a new buffered stream instance from an io.Reader.
// This implementation is designed for large files and streaming data, using a sliding
// window buffer to maintain constant memory usage regardless of input size.
//
// The buffered stream:
//   - Reads data in chunks from the io.Reader as needed
//   - Maintains a sliding window buffer of runes for backtracking support
//   - Supports Clone() for backtracking within the buffer window
//   - Tracks position (row, column, offset) across buffer boundaries
//   - Handles UTF-8 encoding properly
//
// Limitations:
//   - Backtracking is limited to the buffer window size (64KB of runes)
//   - Reset() requires re-reading from the beginning (only works with seekable readers)
//   - Not safe for concurrent use from multiple goroutines
//
// For small strings that fit entirely in memory, use NewStream() instead.
func NewStreamFromReader(reader io.Reader) Stream {
	shared := &sharedBuffer{
		data:  make([]rune, 0, bufferSize),
		start: 0,
		eof:   false,
		err:   nil,
	}

	s := &bufferedStreamImpl{
		uuid:    uuid.New(),
		reader:  reader,
		shared:  shared,
		readBuf: make([]byte, readChunkSize),
		location: Location{
			Cursor: 0,
			Row:    1,
			Column: 1,
		},
	}

	// Pre-fill the buffer
	s.refillBuffer()

	return s
}

// bufferedStreamImpl is a buffered implementation of the Stream interface
// that works with io.Reader for large files and streaming data.
//
// The stream uses a shared buffer that is referenced by all clones. This ensures
// that buffer modifications (refilling, discarding) are visible to all instances.
type bufferedStreamImpl struct {
	uuid     uuid.UUID
	reader   io.Reader
	shared   *sharedBuffer // Shared buffer state (pointer ensures all clones see updates)
	readBuf  []byte        // Temporary buffer for reading bytes (not shared)
	location Location      // Current position in stream (unique per instance)
}

// refillBuffer reads more data from the reader and appends it to the shared buffer.
// It handles UTF-8 decoding and stops when the buffer is full or EOF is reached.
func (s *bufferedStreamImpl) refillBuffer() {
	if s.shared.eof {
		return
	}

	// Read bytes from the reader
	n, err := s.reader.Read(s.readBuf)
	if err != nil {
		if err == io.EOF {
			s.shared.eof = true
		} else {
			s.shared.err = err
			s.shared.eof = true
			return
		}
	}

	if n == 0 {
		return
	}

	// Decode bytes to runes and append to shared buffer
	data := s.readBuf[:n]
	offset := 0

	for offset < len(data) {
		r, size := utf8.DecodeRune(data[offset:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8, skip this byte
			offset++
			continue
		}
		s.shared.data = append(s.shared.data, r)
		offset += size

		// Check if we've filled the buffer
		if len(s.shared.data) >= bufferSize {
			break
		}
	}

	// If we didn't consume all bytes and not at EOF, we need to handle
	// partial UTF-8 sequences at the boundary. For now, we'll just continue
	// on the next read.
}

// ensureBufferHasData ensures the shared buffer has data available at the current position.
// It refills the buffer if needed and possible.
func (s *bufferedStreamImpl) ensureBufferHasData() {
	// Calculate position within buffer
	posInBuffer := s.location.Cursor - int(s.shared.start)

	// If buffer is full and we can discard old data safely
	// We use a quarter of the buffer as the threshold for more aggressive discarding
	if len(s.shared.data) >= bufferSize && posInBuffer > bufferSize/4 {
		// Discard data up to the current position minus a safety margin
		// Keep 1/8 of buffer before current position for any short-range clones
		safetyMargin := bufferSize / 8
		discardCount := posInBuffer - safetyMargin
		if discardCount > 0 && discardCount < len(s.shared.data) {
			s.shared.data = s.shared.data[discardCount:]
			s.shared.start += int64(discardCount)
		}
	}

	// Now refill the buffer
	s.refillBuffer()
}

// Clone creates a copy of the stream for backtracking support.
// The clone shares the same buffer (via pointer) and reader but has independent position.
// This ensures that buffer modifications by either the clone or original are visible to both.
func (s *bufferedStreamImpl) Clone() Stream {
	return &bufferedStreamImpl{
		uuid:     s.uuid,
		reader:   s.reader,
		shared:   s.shared,                    // Share the pointer to buffer state
		readBuf:  make([]byte, readChunkSize), // Each clone needs its own read buffer
		location: s.location,                  // Clone gets its own copy of position
	}
}

// Match updates this stream's location to match another stream's location.
// Both streams must be clones of each other (same UUID).
// Since the buffer is shared via pointer, no buffer state needs to be copied.
func (s *bufferedStreamImpl) Match(other Stream) {
	otherImpl, ok := other.(*bufferedStreamImpl)
	if !ok {
		panic("type assertion failed: expected *bufferedStreamImpl")
	}
	if s.uuid != otherImpl.uuid {
		panic("trying to match two different streams")
	}

	// Update location to match (buffer is already shared via pointer)
	s.location = otherImpl.location
}

// PeekChar returns the next rune without advancing the stream.
func (s *bufferedStreamImpl) PeekChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}

	posInBuffer := s.location.Cursor - int(s.shared.start)

	// If we're at the end of the buffer but not EOF, try to refill
	if posInBuffer >= len(s.shared.data) && !s.shared.eof {
		s.ensureBufferHasData()
	}

	// Recalculate position after potential buffer management
	posInBuffer = s.location.Cursor - int(s.shared.start)

	// Check again after refill
	if posInBuffer >= len(s.shared.data) {
		return 0, false
	}

	return s.shared.data[posInBuffer], true
}

// NextChar reads and returns the next rune, advancing the stream position.
// Automatically tracks newlines for row/column position.
func (s *bufferedStreamImpl) NextChar() (rune, bool) {
	if s.IsEos() {
		return 0, false
	}

	posInBuffer := s.location.Cursor - int(s.shared.start)

	// If we're at the end of the buffer but not EOF, try to refill
	if posInBuffer >= len(s.shared.data) && !s.shared.eof {
		s.ensureBufferHasData()
	}

	// Recalculate position after potential buffer management
	posInBuffer = s.location.Cursor - int(s.shared.start)

	// Check again after refill
	if posInBuffer >= len(s.shared.data) {
		return 0, false
	}

	if posInBuffer < 0 {
		// This should never happen, but check just in case
		return 0, false
	}

	r := s.shared.data[posInBuffer]
	s.location.Cursor += 1
	s.location.Column += 1

	if r == '\n' {
		s.location.Row += 1
		s.location.Column = 1
	}

	return r, true
}

// MatchChars attempts to match a rune sequence against the stream.
// If successful, the stream is advanced. If not, the stream position is unchanged.
func (s *bufferedStreamImpl) MatchChars(match []rune) bool {
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
func (s *bufferedStreamImpl) IsEos() bool {
	posInBuffer := s.location.Cursor - int(s.shared.start)

	// If we're within the buffer, we're not at EOS
	if posInBuffer < len(s.shared.data) {
		return false
	}

	// If we're at the end of buffer but not EOF, try to refill
	if !s.shared.eof {
		s.ensureBufferHasData()
		// Recalculate position after potential buffer management
		posInBuffer = s.location.Cursor - int(s.shared.start)
		// Check again after refill
		if posInBuffer < len(s.shared.data) {
			return false
		}
	}

	// We're at the end of buffer and EOF
	return true
}

// GetOffset returns the current byte offset within the stream.
func (s *bufferedStreamImpl) GetOffset() int {
	return s.location.Cursor
}

// GetRow returns the current line number (1-indexed).
func (s *bufferedStreamImpl) GetRow() int {
	return s.location.Row
}

// GetColumn returns the current column number (1-indexed).
func (s *bufferedStreamImpl) GetColumn() int {
	return s.location.Column
}

// Reset resets the stream to the beginning (offset 0, row 1, column 1).
// Note: This only works properly with seekable readers. For non-seekable readers,
// this will reset the position tracking but won't actually re-read from the beginning.
func (s *bufferedStreamImpl) Reset() {
	s.location.Cursor = 0
	s.location.Row = 1
	s.location.Column = 1

	// If the reader is seekable, try to seek back to the beginning
	if seeker, ok := s.reader.(io.Seeker); ok {
		// nolint:errcheck // Ignore seek errors - best effort reset
		seeker.Seek(0, io.SeekStart)
		s.shared.data = s.shared.data[:0]
		s.shared.start = 0
		s.shared.eof = false
		s.shared.err = nil
		s.refillBuffer()
	}
	// For non-seekable readers, we can only reset position tracking
	// The buffer still contains data from where it was read
}

// GetLocation returns the current position in the stream.
func (s *bufferedStreamImpl) GetLocation() Location {
	return s.location
}

// SetLocation sets the stream position to the specified location.
func (s *bufferedStreamImpl) SetLocation(loc Location) {
	s.location = loc
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
