package parser

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
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
				Position: ast.Position{Line: 5, Column: 10},
			},
			expected: "error at line 5, column 10: unexpected token",
		},
		{
			name: "error without position",
			err: &ParseError{
				Message:  "invalid syntax",
				Position: ast.Position{},
			},
			expected: "parse error: invalid syntax",
		},
		{
			name: "error with line only",
			err: &ParseError{
				Message:  "missing bracket",
				Position: ast.Position{Line: 3},
			},
			expected: "parse error: missing bracket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewSyntaxError(t *testing.T) {
	pos := ast.Position{Line: 10, Column: 5}
	message := "invalid syntax"

	err := NewSyntaxError(pos, message)

	if err.Message != message {
		t.Errorf("Message = %q, want %q", err.Message, message)
	}
	if err.Position != pos {
		t.Errorf("Position = %v, want %v", err.Position, pos)
	}
}

func TestNewUnexpectedTokenError(t *testing.T) {
	pos := ast.Position{Line: 15, Column: 8}
	expected := "identifier"
	got := "number"

	err := NewUnexpectedTokenError(pos, expected, got)

	if !strings.Contains(err.Message, expected) {
		t.Errorf("Message should contain expected token %q", expected)
	}
	if !strings.Contains(err.Message, got) {
		t.Errorf("Message should contain got token %q", got)
	}
	if err.Position != pos {
		t.Errorf("Position = %v, want %v", err.Position, pos)
	}

	expectedMsg := "expected identifier, got number"
	if err.Message != expectedMsg {
		t.Errorf("Message = %q, want %q", err.Message, expectedMsg)
	}
}

func TestNewUnexpectedEOFError(t *testing.T) {
	pos := ast.Position{Line: 20, Column: 1}
	expected := "closing brace"

	err := NewUnexpectedEOFError(pos, expected)

	if !strings.Contains(err.Message, "unexpected EOF") {
		t.Errorf("Message should contain 'unexpected EOF'")
	}
	if !strings.Contains(err.Message, expected) {
		t.Errorf("Message should contain expected token %q", expected)
	}
	if err.Position != pos {
		t.Errorf("Position = %v, want %v", err.Position, pos)
	}

	expectedMsg := "unexpected EOF, expected closing brace"
	if err.Message != expectedMsg {
		t.Errorf("Message = %q, want %q", err.Message, expectedMsg)
	}
}
