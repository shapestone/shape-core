package validator_test

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/shape"
	"github.com/shapestone/shape/pkg/validator"
)

// BenchmarkValidateAll_SimpleSchema benchmarks validation of a simple schema
func BenchmarkValidateAll_SimpleSchema(b *testing.B) {
	schema := `{"id": UUID, "name": String(1, 100)}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_ComplexSchema benchmarks validation of a complex nested schema
func BenchmarkValidateAll_ComplexSchema(b *testing.B) {
	schema := `{
		"user": {
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"profile": {
				"bio": String(0, 500),
				"avatar": URL,
				"website": URL,
				"location": {
					"city": String(1, 100),
					"country": String(2, 100)
				}
			},
			"tags": [String(1, 50)],
			"settings": {
				"notifications": Boolean,
				"privacy": String(1, 20)
			}
		}
	}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_WithErrors benchmarks validation with multiple errors
func BenchmarkValidateAll_WithErrors(b *testing.B) {
	schema := `{
		"unknown1": UnknownType,
		"unknown2": AnotherUnknown,
		"badArgs": Integer(1, 100, 200),
		"badFunc": NotAFunction(1, 2)
	}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if result.ErrorCount() != 4 {
			b.Fatalf("expected 4 errors, got %d", result.ErrorCount())
		}
	}
}

// BenchmarkValidateAll_NoSourceText benchmarks validation without source context
func BenchmarkValidateAll_NoSourceText(b *testing.B) {
	schema := `{"id": UUID, "name": String(1, 100)}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast) // No source text
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_WithSourceText benchmarks validation with source context
func BenchmarkValidateAll_WithSourceText(b *testing.B) {
	schema := `{"id": UUID, "name": String(1, 100)}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_XMLV benchmarks validation of XMLV format
func BenchmarkValidateAll_XMLV(b *testing.B) {
	schema := `<schema><id>UUID</id><name>String(1, 100)</name></schema>`
	ast, err := shape.Parse(parser.FormatXMLV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_YAMLV benchmarks validation of YAMLV format
func BenchmarkValidateAll_YAMLV(b *testing.B) {
	schema := `id: UUID
name: String(1, 100)`
	ast, err := shape.Parse(parser.FormatYAMLV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_DeepNesting benchmarks validation of deeply nested schemas
func BenchmarkValidateAll_DeepNesting(b *testing.B) {
	schema := `{
		"level1": {
			"level2": {
				"level3": {
					"level4": {
						"level5": {
							"id": UUID,
							"name": String(1, 100)
						}
					}
				}
			}
		}
	}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_Arrays benchmarks validation of schemas with arrays
func BenchmarkValidateAll_Arrays(b *testing.B) {
	schema := `{
		"tags": [String(1, 50)],
		"numbers": [Integer(1, 100)],
		"nested": [{
			"id": UUID,
			"name": String(1, 50)
		}]
	}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := shape.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkValidateAll_CustomTypes benchmarks validation with custom types
func BenchmarkValidateAll_CustomTypes(b *testing.B) {
	schema := `{"ssn": SSN, "phone": PhoneNumber}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	v := validator.NewSchemaValidator()
	v.RegisterType("SSN", validator.TypeDescriptor{
		Name:        "SSN",
		Description: "Social Security Number",
	})
	v.RegisterType("PhoneNumber", validator.TypeDescriptor{
		Name:        "PhoneNumber",
		Description: "Phone Number",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := v.ValidateAll(ast, schema)
		if !result.Valid {
			b.Fatal("expected valid schema")
		}
	}
}

// BenchmarkFormatColored benchmarks colored output formatting
func BenchmarkFormatColored(b *testing.B) {
	schema := `{"unknown": UnknownType}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	result := shape.ValidateAll(ast, schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.FormatColored()
	}
}

// BenchmarkFormatPlain benchmarks plain output formatting
func BenchmarkFormatPlain(b *testing.B) {
	schema := `{"unknown": UnknownType}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	result := shape.ValidateAll(ast, schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.FormatPlain()
	}
}

// BenchmarkToJSON benchmarks JSON output formatting
func BenchmarkToJSON(b *testing.B) {
	schema := `{"unknown": UnknownType}`
	ast, err := shape.Parse(parser.FormatJSONV, schema)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	result := shape.ValidateAll(ast, schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = result.ToJSON()
	}
}
