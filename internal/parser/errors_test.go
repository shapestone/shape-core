package parser

import (
	"strings"
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "error with position",
			err: &ParseError{
				Message:  "unexpected token",
				Position: ast.Position{Line: 5, Column: 12},
			},
			expected: "error at line 5, column 12: unexpected token",
		},
		{
			name: "error without position",
			err: &ParseError{
				Message:  "syntax error",
				Position: ast.Position{},
			},
			expected: "parse error: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ParseError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewSyntaxError(t *testing.T) {
	pos := ast.Position{Line: 10, Column: 5}
	err := NewSyntaxError(pos, "invalid syntax")

	if err.Message != "invalid syntax" {
		t.Errorf("NewSyntaxError() message = %q, want %q", err.Message, "invalid syntax")
	}
	if err.Position != pos {
		t.Errorf("NewSyntaxError() position = %v, want %v", err.Position, pos)
	}
}

func TestNewUnexpectedTokenError(t *testing.T) {
	pos := ast.Position{Line: 3, Column: 7}
	err := NewUnexpectedTokenError(pos, "}", "EOF")

	if !strings.Contains(err.Message, "expected }") {
		t.Errorf("NewUnexpectedTokenError() message should contain 'expected }', got %q", err.Message)
	}
	if !strings.Contains(err.Message, "got EOF") {
		t.Errorf("NewUnexpectedTokenError() message should contain 'got EOF', got %q", err.Message)
	}
	if err.Position != pos {
		t.Errorf("NewUnexpectedTokenError() position = %v, want %v", err.Position, pos)
	}
}

func TestNewUnexpectedEOFError(t *testing.T) {
	pos := ast.Position{Line: 8, Column: 1}
	err := NewUnexpectedEOFError(pos, "}")

	if !strings.Contains(err.Message, "unexpected EOF") {
		t.Errorf("NewUnexpectedEOFError() message should contain 'unexpected EOF', got %q", err.Message)
	}
	if !strings.Contains(err.Message, "expected }") {
		t.Errorf("NewUnexpectedEOFError() message should contain 'expected }', got %q", err.Message)
	}
	if err.Position != pos {
		t.Errorf("NewUnexpectedEOFError() position = %v, want %v", err.Position, pos)
	}
}
