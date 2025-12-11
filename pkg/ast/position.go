package ast

import "fmt"

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

// IsValid returns true if the position has been set (not default zero values).
func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}

// String returns a string representation of the position.
func (p Position) String() string {
	if !p.IsValid() {
		return "<unknown position>"
	}
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// ZeroPosition returns a zero/invalid position.
func ZeroPosition() Position {
	return Position{Offset: 0, Line: 0, Column: 0}
}
