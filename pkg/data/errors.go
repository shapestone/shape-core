package data

import (
	"errors"
	"fmt"
)

// ErrSyntax is matched by errors.Is for every *SyntaxError.
var ErrSyntax = errors.New("syntax error")

// ErrLimit is matched by errors.Is for every *LimitError.
var ErrLimit = errors.New("limit exceeded")

// Names of the limits reported in LimitError.Limit.
const (
	LimitDepth          = "depth"
	LimitBytes          = "bytes"
	LimitNodes          = "nodes"
	LimitStringBytes    = "string-bytes"
	LimitKeysPerMap     = "keys-per-map"
	LimitAliasExpansion = "alias-expansion"
	LimitMemory         = "memory"
)

// SyntaxError reports malformed input. Offset is a zero-based byte offset;
// Line and Col are one-based, and zero means unknown.
type SyntaxError struct {
	Format string // "json", "yaml", "csv", "xml", ...
	Offset int
	Line   int
	Col    int
	Msg    string
}

// Error implements the error interface.
func (e *SyntaxError) Error() string {
	format := e.Format
	if format == "" {
		format = "data"
	}
	if e.Line > 0 {
		return fmt.Sprintf("%s: syntax error at line %d, column %d (offset %d): %s",
			format, e.Line, e.Col, e.Offset, e.Msg)
	}
	return fmt.Sprintf("%s: syntax error at offset %d: %s", format, e.Offset, e.Msg)
}

// Is reports whether target is ErrSyntax.
func (e *SyntaxError) Is(target error) bool { return target == ErrSyntax }

// LimitError reports that a configured limit was exceeded. Limit is one of
// the Limit* names, Max is the configured ceiling and Got is the value that
// crossed it.
type LimitError struct {
	Limit string
	Max   int64
	Got   int64
}

// Error implements the error interface.
func (e *LimitError) Error() string {
	return fmt.Sprintf("limit exceeded: %s (max %d, got %d)", e.Limit, e.Max, e.Got)
}

// Is reports whether target is ErrLimit.
func (e *LimitError) Is(target error) bool { return target == ErrLimit }
