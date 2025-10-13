package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shapestone/shape/pkg/ast"
)

// ErrorCode represents a machine-readable error type.
type ErrorCode string

const (
	ErrCodeUnknownType       ErrorCode = "UNKNOWN_TYPE"
	ErrCodeUnknownFunction   ErrorCode = "UNKNOWN_FUNCTION"
	ErrCodeInvalidArgCount   ErrorCode = "INVALID_ARG_COUNT"
	ErrCodeInvalidArgType    ErrorCode = "INVALID_ARG_TYPE"
	ErrCodeInvalidArgValue   ErrorCode = "INVALID_ARG_VALUE"
	ErrCodeCircularReference ErrorCode = "CIRCULAR_REFERENCE"
)

// ValidationError represents a semantic validation error with position, path, code, message, and hint.
type ValidationError struct {
	Position ast.Position // Source position (line, column)
	Path     string       // JSONPath (e.g., "$.user.age")
	Code     ErrorCode    // Machine-readable error code
	Message  string       // Human-readable error message
	Hint     string       // Helpful suggestion for fixing the error

	// Source context for better error display
	Source      string   // Original schema text (optional)
	SourceLines []string // Lines around the error for context display (optional)
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	var parts []string

	// Position info
	parts = append(parts, fmt.Sprintf("line %d, column %d", e.Position.Line, e.Position.Column))

	// Path if present
	if e.Path != "" {
		parts = append(parts, fmt.Sprintf("path %s", e.Path))
	}

	// Message
	parts = append(parts, e.Message)

	// Hint if present
	if e.Hint != "" {
		parts = append(parts, fmt.Sprintf("hint: %s", e.Hint))
	}

	return strings.Join(parts, ": ")
}

// FormatPlain formats the error as plain text (no colors).
// This is suitable for log files, non-terminal output, or when colors are disabled.
func (e *ValidationError) FormatPlain() string {
	var buf strings.Builder

	// Header: Line/Column + Path
	if e.Position.Line > 0 {
		buf.WriteString(fmt.Sprintf("Line %d, Column %d", e.Position.Line, e.Position.Column))
		if e.Path != "" {
			buf.WriteString(fmt.Sprintf(" (%s)", e.Path))
		}
		buf.WriteString("\n")
	}

	// Error code and message
	if e.Code != "" {
		buf.WriteString(fmt.Sprintf("ERROR [%s]: %s\n", e.Code, e.Message))
	} else {
		buf.WriteString(fmt.Sprintf("ERROR: %s\n", e.Message))
	}

	// Source context (if available)
	if len(e.SourceLines) > 0 {
		buf.WriteString("\n")
		// Calculate starting line number
		// SourceLines should start from max(1, errorLine-2), but we'll calculate it dynamically
		// by finding which index in SourceLines corresponds to the error line
		// We expect the error line to be at index min(2, errorLine-1) in SourceLines
		errorLineIndex := minInt(2, e.Position.Line-1)
		startLineNum := e.Position.Line - errorLineIndex

		for i, line := range e.SourceLines {
			lineNum := startLineNum + i
			if lineNum == e.Position.Line {
				// This is the error line - mark it with >
				buf.WriteString(fmt.Sprintf("  > %2d | %s\n", lineNum, line))
				// Add arrow pointing to column
				if e.Position.Column > 0 {
					buf.WriteString(fmt.Sprintf("      | %s^\n", strings.Repeat(" ", e.Position.Column-1)))
				}
			} else {
				buf.WriteString(fmt.Sprintf("    %2d | %s\n", lineNum, line))
			}
		}
		buf.WriteString("\n")
	}

	// Hint
	if e.Hint != "" {
		buf.WriteString(fmt.Sprintf("HINT: %s\n", e.Hint))
	}

	return buf.String()
}

// FormatColored formats the error with ANSI colors for terminal display.
// Automatically falls back to plain formatting if NO_COLOR is set.
func (e *ValidationError) FormatColored() string {
	if !shouldUseColor() {
		return e.FormatPlain()
	}

	var buf strings.Builder

	// Header in cyan
	if e.Position.Line > 0 {
		buf.WriteString(cyan(fmt.Sprintf("Line %d, Column %d", e.Position.Line, e.Position.Column)))
		if e.Path != "" {
			buf.WriteString(gray(fmt.Sprintf(" (%s)", e.Path)))
		}
		buf.WriteString("\n")
	}

	// Error in red
	if e.Code != "" {
		buf.WriteString(red(fmt.Sprintf("ERROR [%s]: ", e.Code)))
		buf.WriteString(fmt.Sprintf("%s\n", e.Message))
	} else {
		buf.WriteString(red("ERROR: "))
		buf.WriteString(fmt.Sprintf("%s\n", e.Message))
	}

	// Source context with colors
	if len(e.SourceLines) > 0 {
		buf.WriteString("\n")
		// Calculate starting line number (same logic as FormatPlain)
		errorLineIndex := minInt(2, e.Position.Line-1)
		startLineNum := e.Position.Line - errorLineIndex

		for i, line := range e.SourceLines {
			lineNum := startLineNum + i
			if lineNum == e.Position.Line {
				// Error line in red
				buf.WriteString(red(fmt.Sprintf("  > %2d | ", lineNum)))
				buf.WriteString(fmt.Sprintf("%s\n", line))
				if e.Position.Column > 0 {
					buf.WriteString(red(fmt.Sprintf("      | %s^\n", strings.Repeat(" ", e.Position.Column-1))))
				}
			} else {
				// Context lines in gray
				buf.WriteString(gray(fmt.Sprintf("    %2d | %s\n", lineNum, line)))
			}
		}
		buf.WriteString("\n")
	}

	// Hint in blue
	if e.Hint != "" {
		buf.WriteString(blue("HINT: "))
		buf.WriteString(fmt.Sprintf("%s\n", e.Hint))
	}

	return buf.String()
}

// ToJSON returns a JSON representation of the validation error.
func (e *ValidationError) ToJSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}
