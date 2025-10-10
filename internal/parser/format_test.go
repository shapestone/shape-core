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
		// JSONV tests
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
			name:     "JSONV nested object",
			input:    `{"user": {"id": UUID, "name": String}}`,
			expected: FormatJSONV,
			wantErr:  false,
		},

		// XMLV tests
		{
			name:     "XMLV simple",
			input:    `<user><id>UUID</id></user>`,
			expected: FormatXMLV,
			wantErr:  false,
		},
		{
			name:     "XMLV with whitespace",
			input:    "  \n <user>\n  <id>UUID</id>\n</user>",
			expected: FormatXMLV,
			wantErr:  false,
		},

		// PropsV tests
		{
			name:     "PropsV simple",
			input:    `id=UUID`,
			expected: FormatPropsV,
			wantErr:  false,
		},
		{
			name:     "PropsV with dot notation",
			input:    `user.id=UUID`,
			expected: FormatPropsV,
			wantErr:  false,
		},
		{
			name:     "PropsV multiple lines",
			input:    "id=UUID\nname=String",
			expected: FormatPropsV,
			wantErr:  false,
		},

		// CSVV tests
		{
			name:     "CSVV simple header",
			input:    `id,name,age`,
			expected: FormatCSVV,
			wantErr:  false,
		},
		{
			name:     "CSVV with header and validation",
			input:    "id,name\nUUID,String",
			expected: FormatCSVV,
			wantErr:  false,
		},
		{
			name:     "CSVV with comment",
			input:    "# Schema\nid,name,email",
			expected: FormatCSVV,
			wantErr:  false,
		},

		// YAMLV tests
		{
			name:     "YAMLV simple",
			input:    `id: UUID`,
			expected: FormatYAMLV,
			wantErr:  false,
		},
		{
			name:     "YAMLV nested",
			input:    "user:\n  id: UUID\n  name: String",
			expected: FormatYAMLV,
			wantErr:  false,
		},
		{
			name:     "YAMLV with array",
			input:    "tags:\n  - String(1, 30)",
			expected: FormatYAMLV,
			wantErr:  false,
		},
		{
			name:     "YAMLV multiple properties",
			input:    "id: UUID\nname: String\nemail: Email",
			expected: FormatYAMLV,
			wantErr:  false,
		},

		// TEXTV tests
		{
			name:     "TEXTV with dot notation",
			input:    `user.id: UUID`,
			expected: FormatTEXTV,
			wantErr:  false,
		},
		{
			name:     "TEXTV multiple properties",
			input:    "user.id: UUID\nuser.name: String",
			expected: FormatTEXTV,
			wantErr:  false,
		},
		{
			name:     "TEXTV with array syntax",
			input:    "user.tags[]: String",
			expected: FormatTEXTV,
			wantErr:  false,
		},
		{
			name:     "TEXTV deeply nested",
			input:    "user.profile.name: String(1, 100)",
			expected: FormatTEXTV,
			wantErr:  false,
		},

		// Error cases
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
