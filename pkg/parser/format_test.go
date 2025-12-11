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
		{Format(99), "Unknown"}, // Test invalid format
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
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
			name:     "JSONV - object",
			input:    `{"name": "String(1,100)"}`,
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "JSONV - array",
			input:    `[{"id": "UUID"}]`,
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "JSONV - with whitespace",
			input:    "  \n  {\"name\": \"String\"}",
			expected: FormatJSONV,
			wantErr:  false,
		},
		{
			name:     "XMLV - simple",
			input:    `<schema><field type="String"/></schema>`,
			expected: FormatXMLV,
			wantErr:  false,
		},
		{
			name:     "XMLV - with whitespace",
			input:    "  \t\n<root></root>",
			expected: FormatXMLV,
			wantErr:  false,
		},
		{
			name:     "CSVV - simple header",
			input:    "name,age,email",
			expected: FormatCSVV,
			wantErr:  false,
		},
		{
			name:     "CSVV - header without validation",
			input:    "id,name,address\n1,John,NYC",
			expected: FormatCSVV,
			wantErr:  false,
		},
		{
			name:     "PropsV - simple",
			input:    "name=String(1,100)\nage=Integer(0,150)",
			expected: FormatPropsV,
			wantErr:  false,
		},
		{
			name:     "PropsV - with whitespace",
			input:    "  key = value",
			expected: FormatPropsV,
			wantErr:  false,
		},
		{
			name:     "YAMLV - nested",
			input:    "person:\n  name: String\n  age: Integer",
			expected: FormatYAMLV,
			wantErr:  false,
		},
		{
			name:     "YAMLV - simple",
			input:    "name: String\nage: Integer",
			expected: FormatYAMLV,
			wantErr:  false,
		},
		{
			name:     "TEXTV - dot notation",
			input:    "person.name: String(1,100)\nperson.age: Integer(0,150)",
			expected: FormatTEXTV,
			wantErr:  false,
		},
		{
			name:     "TEXTV - complex dot notation",
			input:    "user.profile.email: Email\nuser.profile.name: String",
			expected: FormatTEXTV,
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
			input:    "   \n\t  ",
			expected: FormatUnknown,
			wantErr:  true,
		},
		{
			name:     "unknown format",
			input:    "random text without structure",
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

func TestGetFirstLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "multiple lines",
			input:    "first line\nsecond line\nthird line",
			expected: "first line",
		},
		{
			name:     "with empty lines",
			input:    "\n\nfirst non-empty\nsecond",
			expected: "first non-empty",
		},
		{
			name:     "with comments",
			input:    "# comment\n# another comment\nactual line",
			expected: "actual line",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFirstLine(tt.input)
			if got != tt.expected {
				t.Errorf("getFirstLine() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHasDotNotation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple dot notation",
			input:    "person.name: String",
			expected: true,
		},
		{
			name:     "nested dot notation",
			input:    "user.profile.email: Email",
			expected: true,
		},
		{
			name:     "no dot notation",
			input:    "name: String",
			expected: false,
		},
		{
			name:     "multiple lines with dots",
			input:    "foo: bar\nbaz.qux: value",
			expected: true,
		},
		{
			name:     "dots in value not key",
			input:    "name: user.name",
			expected: false,
		},
		{
			name:     "with comments",
			input:    "# comment\nperson.name: String",
			expected: true,
		},
		{
			name:     "empty input",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDotNotation(tt.input)
			if got != tt.expected {
				t.Errorf("hasDotNotation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasValidationPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "UUID pattern",
			input:    "id: UUID",
			expected: true,
		},
		{
			name:     "String pattern",
			input:    "name: String(1,100)",
			expected: true,
		},
		{
			name:     "Integer pattern",
			input:    "age: Integer(0,150)",
			expected: true,
		},
		{
			name:     "Email pattern",
			input:    "contact: Email",
			expected: true,
		},
		{
			name:     "function call",
			input:    "value: Custom(param)",
			expected: true,
		},
		{
			name:     "no validation pattern",
			input:    "name,age,email",
			expected: false,
		},
		{
			name:     "plain text",
			input:    "hello world",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasValidationPattern(tt.input)
			if got != tt.expected {
				t.Errorf("hasValidationPattern() = %v, want %v", got, tt.expected)
			}
		})
	}
}
