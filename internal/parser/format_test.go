package parser

import (
	"testing"
)

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatJSONV, "JSONV"},
		{FormatXMLV, "XMLV"},
		{FormatPropsV, "PropsV"},
		{FormatCSVV, "CSVV"},
		{FormatYAMLV, "YAMLV"},
		{FormatTEXTV, "TEXTV"},
		{FormatUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.expected {
				t.Errorf("Format.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Format
		wantErr  bool
	}{
		{
			name:     "JSONV object",
			input:    `{"id": UUID}`,
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "JSONV array",
			input:    `[String(1, 100)]`,
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "JSONV with whitespace",
			input:    "  \n  {\n  \"id\": UUID\n}",
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "XMLV",
			input:    `<user><id>UUID</id></user>`,
			expected: FormatXMLV,
			wantErr:  false,
		},
		{
			name:     "PropsV",
			input:    `id=UUID`,
			expected: FormatPropsV,
			wantErr:  false,
		},
		{
			name:     "CSVV",
			input:    `id,name,age`,
			expected: FormatCSVV,
			wantErr:  false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: FormatUnknown,
			wantErr:  true,
		},
		{
			name:     "whitespace only",
			input:    "   \n  \t  ",
			expected: FormatUnknown,
			wantErr:  true,
		},
		{
			name:     "unknown format",
			input:    "something else",
			expected: FormatUnknown,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}
