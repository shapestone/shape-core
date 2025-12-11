package parser

import (
	"fmt"

	"github.com/shapestone/shape/pkg/ast"
)

// ParseError represents a parsing error with position information.
type ParseError struct {
	Message  string
	Position ast.Position
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Position.Line > 0 && e.Position.Column > 0 {
		return fmt.Sprintf("error at line %d, column %d: %s",
			e.Position.Line, e.Position.Column, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// NewSyntaxError creates a syntax error with position.
func NewSyntaxError(pos ast.Position, message string) *ParseError {
	return &ParseError{
		Message:  message,
		Position: pos,
	}
}

// NewUnexpectedTokenError creates an error for unexpected tokens.
func NewUnexpectedTokenError(pos ast.Position, expected, got string) *ParseError {
	message := fmt.Sprintf("expected %s, got %s", expected, got)
	return &ParseError{
		Message:  message,
		Position: pos,
	}
}

// NewUnexpectedEOFError creates an error for unexpected end of file.
func NewUnexpectedEOFError(pos ast.Position, expected string) *ParseError {
	message := fmt.Sprintf("unexpected EOF, expected %s", expected)
	return &ParseError{
		Message:  message,
		Position: pos,
	}
}
