package tokenizer

import "fmt"

//
// Position - Position tracking for tokens and streams
//

// Position represents a location in the source text.
type Position struct {
	Offset int // Byte offset (0-indexed)
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed)
}

// NewPosition creates a new Position with the given offset, line, and column.
func NewPosition(offset, line, column int) Position {
	return Position{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}

// IsValid returns true if the position has been set (not default -1 values).
func (p Position) IsValid() bool {
	return p.Offset >= 0 && p.Line > 0 && p.Column > 0
}

// String returns a string representation of the position.
func (p Position) String() string {
	if !p.IsValid() {
		return "<unknown position>"
	}
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}
