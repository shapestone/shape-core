package shape_test

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/shape"
	"github.com/shapestone/shape/pkg/validator"
)

// TestValidateAll_JSONV tests ValidateAll with JSONV format
func TestValidateAll_JSONV(t *testing.T) {
	tests := []struct {
		name      string
		schema    string
		wantErr   bool
		wantCount int // expected error count
	}{
		{
			name:      "valid schema",
			schema:    `{"id": UUID, "age": Integer(1, 120)}`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "valid nested schema",
			schema:    `{"user": {"id": UUID, "name": String(1, 100)}}`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "valid array schema",
			schema:    `{"tags": [String(1, 50)]}`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "unknown type",
			schema:    `{"country": CountryCode}`,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "invalid arg count - too many",
			schema:    `{"age": Integer(1, 100, 200)}`,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "invalid arg count - too few",
			schema:    `{"name": String()}`,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "unknown function",
			schema:    `{"id": NotAFunction(1, 2)}`,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "multiple errors",
			schema:    `{"unknown": UnknownType, "bad": BadFunction(1, 2, 3)}`,
			wantErr:   true,
			wantCount: 2,
		},
		{
			name:      "complex valid schema",
			schema:    `{"user": {"id": UUID, "profile": {"bio": String(0, 500), "email": Email}, "age": Integer(18, 120)}}`,
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := shape.Parse(parser.FormatJSONV, tt.schema)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result := shape.ValidateAll(ast, tt.schema)

			if (result.ErrorCount() > 0) != tt.wantErr {
				t.Errorf("ValidateAll() errors = %d, wantErr %v", result.ErrorCount(), tt.wantErr)
			}

			if result.ErrorCount() != tt.wantCount {
				t.Errorf("ValidateAll() error count = %d, want %d", result.ErrorCount(), tt.wantCount)
				if result.ErrorCount() > 0 {
					t.Logf("Errors: %v", result.FormatPlain())
				}
			}
		})
	}
}

// TestValidateAll_XMLV tests ValidateAll with XMLV format
func TestValidateAll_XMLV(t *testing.T) {
	tests := []struct {
		name      string
		schema    string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid schema",
			schema:    `<schema><id>UUID</id><age>Integer(1, 120)</age></schema>`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "unknown type",
			schema:    `<schema><country>CountryCode</country></schema>`,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "invalid arg count",
			schema:    `<schema><age>Integer(1, 100, 200)</age></schema>`,
			wantErr:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := shape.Parse(parser.FormatXMLV, tt.schema)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result := shape.ValidateAll(ast, tt.schema)

			if (result.ErrorCount() > 0) != tt.wantErr {
				t.Errorf("ValidateAll() errors = %d, wantErr %v", result.ErrorCount(), tt.wantErr)
			}

			if result.ErrorCount() != tt.wantCount {
				t.Errorf("ValidateAll() error count = %d, want %d", result.ErrorCount(), tt.wantCount)
			}
		})
	}
}

// TestValidateAll_YAMLV tests ValidateAll with YAMLV format
func TestValidateAll_YAMLV(t *testing.T) {
	tests := []struct {
		name      string
		schema    string
		wantErr   bool
		wantCount int
	}{
		{
			name: "valid schema",
			schema: `id: UUID
age: Integer(1, 120)`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "nested valid schema",
			schema: `user:
  id: UUID
  name: String(1, 100)`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "unknown type",
			schema:    `country: CountryCode`,
			wantErr:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := shape.Parse(parser.FormatYAMLV, tt.schema)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result := shape.ValidateAll(ast, tt.schema)

			if (result.ErrorCount() > 0) != tt.wantErr {
				t.Errorf("ValidateAll() errors = %d, wantErr %v", result.ErrorCount(), tt.wantErr)
			}

			if result.ErrorCount() != tt.wantCount {
				t.Errorf("ValidateAll() error count = %d, want %d", result.ErrorCount(), tt.wantCount)
			}
		})
	}
}

// TestValidateAll_AllFormats tests validation works across all formats
func TestValidateAll_AllFormats(t *testing.T) {
	formats := []struct {
		name   string
		format parser.Format
		schema string
	}{
		{
			name:   "JSONV",
			format: parser.FormatJSONV,
			schema: `{"id": UUID}`,
		},
		{
			name:   "XMLV",
			format: parser.FormatXMLV,
			schema: `<schema><id>UUID</id></schema>`,
		},
		{
			name:   "YAMLV",
			format: parser.FormatYAMLV,
			schema: `id: UUID`,
		},
		{
			name:   "PropsV",
			format: parser.FormatPropsV,
			schema: `id=UUID`,
		},
		{
			name:   "TEXTV",
			format: parser.FormatTEXTV,
			schema: `id: UUID`,
		},
	}

	for _, tt := range formats {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := shape.Parse(tt.format, tt.schema)
			if err != nil {
				t.Fatalf("Parse(%s) error = %v", tt.name, err)
			}

			result := shape.ValidateAll(ast, tt.schema)

			if !result.Valid {
				t.Errorf("ValidateAll(%s) failed: %v", tt.name, result.FormatPlain())
			}

			if result.ErrorCount() != 0 {
				t.Errorf("ValidateAll(%s) error count = %d, want 0", tt.name, result.ErrorCount())
			}
		})
	}
}

// TestValidateAll_CustomTypes tests custom type registration
func TestValidateAll_CustomTypes(t *testing.T) {
	schema := `{"ssn": SSN, "phone": PhoneNumber}`

	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Without custom types - should fail
	result := shape.ValidateAll(ast, schema)
	if result.Valid {
		t.Error("Expected validation to fail without custom types")
	}
	if result.ErrorCount() != 2 {
		t.Errorf("Expected 2 errors, got %d", result.ErrorCount())
	}

	// With custom types - should pass
	v := validator.NewSchemaValidator()
	v.RegisterType("SSN", validator.TypeDescriptor{
		Name:        "SSN",
		Description: "Social Security Number",
	})
	v.RegisterType("PhoneNumber", validator.TypeDescriptor{
		Name:        "PhoneNumber",
		Description: "Phone Number",
	})

	result = v.ValidateAll(ast, schema)
	if !result.Valid {
		t.Errorf("Expected validation to pass with custom types, got: %v", result.FormatPlain())
	}
	if result.ErrorCount() != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidateAll_SourceContext tests that source context is included in errors
func TestValidateAll_SourceContext(t *testing.T) {
	schema := `{
  "id": UUID,
  "unknown": UnknownType,
  "age": Integer(1, 120)
}`

	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := shape.ValidateAll(ast, schema)

	if result.Valid {
		t.Error("Expected validation to fail")
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}

	// Check that source context is included
	if result.ErrorCount() > 0 {
		err := result.GetErrors()[0]
		if len(err.SourceLines) == 0 {
			t.Error("Expected source context to be included in error")
		}
		if err.Source == "" {
			t.Error("Expected source text to be included in error")
		}
	}
}

// TestValidateAll_FormattedOutput tests different output formats
func TestValidateAll_FormattedOutput(t *testing.T) {
	schema := `{"unknown": UnknownType}`

	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := shape.ValidateAll(ast, schema)

	if result.Valid {
		t.Error("Expected validation to fail")
	}

	// Test plain formatting
	plain := result.FormatPlain()
	if plain == "" {
		t.Error("Expected plain output to be non-empty")
	}
	if !contains(plain, "UnknownType") {
		t.Error("Expected plain output to contain 'UnknownType'")
	}

	// Test colored formatting (should work even if NO_COLOR is set)
	colored := result.FormatColored()
	if colored == "" {
		t.Error("Expected colored output to be non-empty")
	}

	// Test JSON output
	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("Expected JSON output to be non-empty")
	}
}

// TestValidateAll_NoSourceText tests validation without source text
func TestValidateAll_NoSourceText(t *testing.T) {
	schema := `{"unknown": UnknownType}`

	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Validate without source text
	result := shape.ValidateAll(ast)

	if result.Valid {
		t.Error("Expected validation to fail")
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}

	// Source context should be empty
	if result.ErrorCount() > 0 {
		err := result.GetErrors()[0]
		if len(err.SourceLines) != 0 {
			t.Error("Expected no source context without source text")
		}
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
