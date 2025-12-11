package tokenizer

import (
	"testing"
)

func TestNewPosition(t *testing.T) {
	pos := NewPosition(100, 5, 10)

	if pos.Offset != 100 {
		t.Errorf("Offset = %d, want 100", pos.Offset)
	}
	if pos.Line != 5 {
		t.Errorf("Line = %d, want 5", pos.Line)
	}
	if pos.Column != 10 {
		t.Errorf("Column = %d, want 10", pos.Column)
	}
}

func TestPosition_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected bool
	}{
		{
			name:     "valid position",
			pos:      Position{Offset: 0, Line: 1, Column: 1},
			expected: true,
		},
		{
			name:     "valid position with offset",
			pos:      Position{Offset: 100, Line: 5, Column: 20},
			expected: true,
		},
		{
			name:     "invalid - negative offset",
			pos:      Position{Offset: -1, Line: 1, Column: 1},
			expected: false,
		},
		{
			name:     "invalid - zero line",
			pos:      Position{Offset: 0, Line: 0, Column: 1},
			expected: false,
		},
		{
			name:     "invalid - zero column",
			pos:      Position{Offset: 0, Line: 1, Column: 0},
			expected: false,
		},
		{
			name:     "invalid - default zero values",
			pos:      Position{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.IsValid()
			if got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPosition_String(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected string
	}{
		{
			name:     "valid position",
			pos:      Position{Offset: 0, Line: 1, Column: 1},
			expected: "line 1, column 1",
		},
		{
			name:     "valid position with larger values",
			pos:      Position{Offset: 100, Line: 10, Column: 25},
			expected: "line 10, column 25",
		},
		{
			name:     "invalid position",
			pos:      Position{Offset: -1, Line: 0, Column: 0},
			expected: "<unknown position>",
		},
		{
			name:     "default position",
			pos:      Position{},
			expected: "<unknown position>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
