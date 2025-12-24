package validator

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

// TestSchemaValidator_ValidateAll_UnknownType tests validation of unknown types
// This test should FAIL initially (TDD red phase) until ValidationResult is implemented
func TestSchemaValidator_ValidateAll_UnknownType(t *testing.T) {
	// Create schema with unknown type
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 5, Column: 15}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	// Should have UNKNOWN_TYPE error
	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for unknown type")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]
		if err.Code != ErrCodeUnknownType {
			t.Errorf("Error code = %q, want %q", err.Code, ErrCodeUnknownType)
		}

		if !strings.Contains(err.Message, "CountryCode") {
			t.Errorf("Error message should contain 'CountryCode', got: %s", err.Message)
		}
	}
}

// TestSchemaValidator_ValidateAll_KnownType tests validation of known types
// This test should PASS even initially since known types should work
func TestSchemaValidator_ValidateAll_KnownType(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id": ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for known type UUID")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_ValidateAll_UnknownFunction tests validation of unknown functions
func TestSchemaValidator_ValidateAll_UnknownFunction(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"tags": ast.NewFunctionNode("ArrayOf", []interface{}{"String"}, ast.Position{Line: 8, Column: 12}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for unknown function")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]
		if err.Code != ErrCodeUnknownFunction {
			t.Errorf("Error code = %q, want %q", err.Code, ErrCodeUnknownFunction)
		}

		if !strings.Contains(err.Message, "ArrayOf") {
			t.Errorf("Error message should contain 'ArrayOf', got: %s", err.Message)
		}
	}
}

// TestSchemaValidator_ValidateAll_InvalidArgCount tests function with wrong argument count
func TestSchemaValidator_ValidateAll_InvalidArgCount(t *testing.T) {
	// Integer(min, max) expects 2 args, giving it 3
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"age": ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)},
				ast.Position{Line: 3, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for invalid arg count")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]
		if err.Code != ErrCodeInvalidArgCount {
			t.Errorf("Error code = %q, want %q", err.Code, ErrCodeInvalidArgCount)
		}
	}
}

// TestSchemaValidator_ValidateAll_ValidArgCount tests function with correct argument count
func TestSchemaValidator_ValidateAll_ValidArgCount(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"age": ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(120)},
				ast.Position{Line: 2, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for valid arg count")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0. Errors: %v", len(result.Errors), result.Errors)
	}
}

// TestSchemaValidator_ValidateAll_MultipleErrors tests collecting BOTH errors (not just first)
func TestSchemaValidator_ValidateAll_MultipleErrors(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"age":     ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 3, Column: 10}),
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 5, Column: 15}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for multiple errors")
	}

	// Should collect BOTH errors
	if len(result.Errors) != 2 {
		t.Errorf("ValidateAll() should collect 2 errors, got %d", len(result.Errors))
	}

	// Verify both error types are present
	hasBadArgCount := false
	hasUnknownType := false
	for _, err := range result.Errors {
		if err.Code == ErrCodeInvalidArgCount {
			hasBadArgCount = true
		}
		if err.Code == ErrCodeUnknownType {
			hasUnknownType = true
		}
	}

	if !hasBadArgCount {
		t.Error("Should have INVALID_ARG_COUNT error")
	}
	if !hasUnknownType {
		t.Error("Should have UNKNOWN_TYPE error")
	}
}

// TestSchemaValidator_ValidateAll_NestedObject tests validation of nested objects
func TestSchemaValidator_ValidateAll_NestedObject(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"user": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"id":   ast.NewTypeNode("UUID", ast.Position{Line: 3, Column: 12}),
					"name": ast.NewFunctionNode("String", []interface{}{int64(1), int64(100)}, ast.Position{Line: 4, Column: 13}),
				},
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for nested object. Errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_ValidateAll_NestedObjectWithError tests nested object with error
func TestSchemaValidator_ValidateAll_NestedObjectWithError(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"user": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"id":      ast.NewTypeNode("UUID", ast.Position{Line: 3, Column: 12}),
					"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 17}),
				},
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for nested error")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}
}

// TestSchemaValidator_ValidateAll_Array tests validation of array schemas
func TestSchemaValidator_ValidateAll_Array(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"tags": ast.NewArrayNode(
				ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 2, Column: 13}),
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for array. Errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_ValidateAll_ArrayWithError tests array with invalid element schema
func TestSchemaValidator_ValidateAll_ArrayWithError(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"tags": ast.NewArrayNode(
				ast.NewTypeNode("Tag", ast.Position{Line: 2, Column: 13}),
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false for array with unknown type")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}
}

// TestSchemaValidator_JSONPath tests JSONPath tracking in errors
func TestSchemaValidator_JSONPath(t *testing.T) {
	tests := []struct {
		name         string
		schema       ast.SchemaNode
		wantPath     string
		wantErrCount int
	}{
		{
			name: "root property",
			schema: ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 2, Column: 15}),
				},
				ast.Position{Line: 1, Column: 1},
			),
			wantPath:     "$.country",
			wantErrCount: 1,
		},
		{
			name: "nested property",
			schema: ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"user": ast.NewObjectNode(
						map[string]ast.SchemaNode{
							"address": ast.NewObjectNode(
								map[string]ast.SchemaNode{
									"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 5, Column: 20}),
								},
								ast.Position{Line: 4, Column: 15},
							),
						},
						ast.Position{Line: 2, Column: 10},
					),
				},
				ast.Position{Line: 1, Column: 1},
			),
			wantPath:     "$.user.address.country",
			wantErrCount: 1,
		},
		{
			name: "array element",
			schema: ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"tags": ast.NewArrayNode(
						ast.NewTypeNode("Tag", ast.Position{Line: 2, Column: 13}),
						ast.Position{Line: 2, Column: 10},
					),
				},
				ast.Position{Line: 1, Column: 1},
			),
			wantPath:     "$.tags[]",
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSchemaValidator()
			result := validator.ValidateAll(tt.schema)

			if result.Valid {
				t.Error("ValidateAll() result.Valid = true, want false")
			}

			if len(result.Errors) != tt.wantErrCount {
				t.Errorf("ValidateAll() error count = %d, want %d", len(result.Errors), tt.wantErrCount)
			}

			if len(result.Errors) > 0 {
				err := result.Errors[0]
				if err.Path != tt.wantPath {
					t.Errorf("Error path = %q, want %q", err.Path, tt.wantPath)
				}
			}
		})
	}
}

// TestSchemaValidator_CustomTypes tests registering custom types
func TestSchemaValidator_CustomTypes(t *testing.T) {
	validator := NewSchemaValidator()

	// Register custom type
	validator.RegisterType("SSN", TypeDescriptor{
		Name:        "SSN",
		Description: "Social Security Number",
	})

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"ssn": ast.NewTypeNode("SSN", ast.Position{Line: 2, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for registered custom type. Errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_CustomFunctions tests registering custom functions
func TestSchemaValidator_CustomFunctions(t *testing.T) {
	validator := NewSchemaValidator()

	// Register custom function
	validator.RegisterFunction("CreditCard", FunctionDescriptor{
		Name:        "CreditCard",
		Description: "Credit card validation",
		MinArgs:     0,
		MaxArgs:     1,
	})

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"card": ast.NewFunctionNode("CreditCard", []interface{}{}, ast.Position{Line: 2, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for registered custom function. Errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_ErrorHints tests that helpful hints are provided
func TestSchemaValidator_ErrorHints(t *testing.T) {
	validator := NewSchemaValidator()

	// Test unknown type with similar name
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"email": ast.NewTypeNode("Emaill", ast.Position{Line: 2, Column: 12}), // Typo: Emaill instead of Email
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false")
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]
		if err.Hint == "" {
			t.Error("Error should have a helpful hint for typo")
		}

		// Hint should suggest "Email"
		if !strings.Contains(err.Hint, "Email") {
			t.Errorf("Hint should suggest 'Email', got: %s", err.Hint)
		}
	}
}

// TestSchemaValidator_ComplexSchema tests complex schema with multiple levels
func TestSchemaValidator_ComplexSchema(t *testing.T) {
	validator := NewSchemaValidator()

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":       ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),
			"username": ast.NewFunctionNode("String", []interface{}{int64(3), int64(20)}, ast.Position{Line: 3, Column: 15}),
			"email":    ast.NewTypeNode("Email", ast.Position{Line: 4, Column: 12}),
			"age":      ast.NewFunctionNode("Integer", []interface{}{int64(18), int64(120)}, ast.Position{Line: 5, Column: 10}),
			"status":   ast.NewFunctionNode("Enum", []interface{}{"active", "inactive", "banned"}, ast.Position{Line: 6, Column: 13}),
			"tags": ast.NewArrayNode(
				ast.NewFunctionNode("String", []interface{}{int64(1), int64(30)}, ast.Position{Line: 7, Column: 13}),
				ast.Position{Line: 7, Column: 10},
			),
			"profile": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"bio":      ast.NewFunctionNode("String", []interface{}{int64(0), int64(500)}, ast.Position{Line: 9, Column: 13}),
					"avatar":   ast.NewTypeNode("URL", ast.Position{Line: 10, Column: 15}),
					"verified": ast.NewTypeNode("Boolean", ast.Position{Line: 11, Column: 17}),
				},
				ast.Position{Line: 8, Column: 15},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("ValidateAll() result.Valid = false, want true for complex valid schema. Errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateAll() error count = %d, want 0", len(result.Errors))
	}
}

// TestSchemaValidator_ComplexSchemaWithMultipleErrors tests complex schema with multiple errors
func TestSchemaValidator_ComplexSchemaWithMultipleErrors(t *testing.T) {
	validator := NewSchemaValidator()

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":       ast.NewTypeNode("UIID", ast.Position{Line: 2, Column: 10}),                                                         // Typo
			"age":      ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 3, Column: 10}), // Too many args
			"country":  ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 15}),                                                  // Unknown type
			"validate": ast.NewFunctionNode("ArrayOf", []interface{}{"String"}, ast.Position{Line: 5, Column: 15}),                         // Unknown function
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false")
	}

	// Should collect ALL 4 errors
	if len(result.Errors) != 4 {
		t.Errorf("ValidateAll() should collect 4 errors, got %d", len(result.Errors))
		for i, err := range result.Errors {
			t.Logf("Error %d: %v", i, err)
		}
	}
}

// TestSchemaValidator_LiteralsAlwaysValid tests that literals are always valid
func TestSchemaValidator_LiteralsAlwaysValid(t *testing.T) {
	validator := NewSchemaValidator()

	literals := []interface{}{
		"string literal",
		int64(42),
		3.14,
		true,
		false,
		nil,
	}

	for _, val := range literals {
		schema := ast.NewObjectNode(
			map[string]ast.SchemaNode{
				"value": ast.NewLiteralNode(val, ast.Position{Line: 1, Column: 1}),
			},
			ast.Position{Line: 1, Column: 1},
		)

		result := validator.ValidateAll(schema)

		if !result.Valid {
			t.Errorf("Literal %v should be valid, got errors: %v", val, result.Errors)
		}
	}
}

// TestSchemaValidator_ValidateAll_WithSourceText tests validation with source text for context
func TestSchemaValidator_ValidateAll_WithSourceText(t *testing.T) {
	sourceText := `{
  "name": String,
  "age": Integer(1, 120),
  "country": CountryCode,
  "email": Email
}`

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"name":    ast.NewTypeNode("String", ast.Position{Line: 2, Column: 11}),
			"age":     ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(120)}, ast.Position{Line: 3, Column: 10}),
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 16}),
			"email":   ast.NewTypeNode("Email", ast.Position{Line: 5, Column: 12}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema, sourceText)

	// Should have error for CountryCode
	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAll() error count = %d, want 1", len(result.Errors))
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]

		// Check that source context was added
		if len(err.SourceLines) == 0 {
			t.Error("Error should have SourceLines when sourceText is provided")
		}

		if err.Source != sourceText {
			t.Error("Error should have Source field set to sourceText")
		}

		// Check that source context contains the error line
		foundErrorLine := false
		for _, line := range err.SourceLines {
			if strings.Contains(line, "CountryCode") {
				foundErrorLine = true
				break
			}
		}
		if !foundErrorLine {
			t.Errorf("SourceLines should contain the error line with 'CountryCode'")
		}
	}
}

// TestSchemaValidator_ValidateAll_SourceContext_EdgeCases tests source context at file boundaries
func TestSchemaValidator_ValidateAll_SourceContext_EdgeCases(t *testing.T) {
	t.Run("ErrorOnFirstLine", func(t *testing.T) {
		sourceText := `{ "bad": BadType, "good": String }`

		schema := ast.NewObjectNode(
			map[string]ast.SchemaNode{
				"bad":  ast.NewTypeNode("BadType", ast.Position{Line: 1, Column: 10}),
				"good": ast.NewTypeNode("String", ast.Position{Line: 1, Column: 28}),
			},
			ast.Position{Line: 1, Column: 1},
		)

		validator := NewSchemaValidator()
		result := validator.ValidateAll(schema, sourceText)

		if result.Valid {
			t.Error("Expected validation error")
		}

		if len(result.Errors) > 0 {
			err := result.Errors[0]
			if len(err.SourceLines) == 0 {
				t.Error("SourceLines should be populated even for first line error")
			}
		}
	})

	t.Run("ErrorOnLastLine", func(t *testing.T) {
		sourceText := `{
  "first": String,
  "last": BadType
}`

		schema := ast.NewObjectNode(
			map[string]ast.SchemaNode{
				"first": ast.NewTypeNode("String", ast.Position{Line: 2, Column: 12}),
				"last":  ast.NewTypeNode("BadType", ast.Position{Line: 3, Column: 11}),
			},
			ast.Position{Line: 1, Column: 1},
		)

		validator := NewSchemaValidator()
		result := validator.ValidateAll(schema, sourceText)

		if result.Valid {
			t.Error("Expected validation error")
		}

		if len(result.Errors) > 0 {
			err := result.Errors[0]
			if len(err.SourceLines) == 0 {
				t.Error("SourceLines should be populated even for last line error")
			}

			// Should include lines before the error
			if len(err.SourceLines) < 2 {
				t.Errorf("SourceLines should include context lines, got %d lines", len(err.SourceLines))
			}
		}
	})
}

// TestSchemaValidator_ValidateAll_WithoutSourceText tests validation without source text
func TestSchemaValidator_ValidateAll_WithoutSourceText(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 16}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema) // No source text

	// Should still have error, just without source context
	if result.Valid {
		t.Error("ValidateAll() result.Valid = true, want false")
	}

	if len(result.Errors) > 0 {
		err := result.Errors[0]

		// Source context should be empty when no source text is provided
		if len(err.SourceLines) != 0 {
			t.Error("SourceLines should be empty when no sourceText is provided")
		}

		if err.Source != "" {
			t.Error("Source should be empty when no sourceText is provided")
		}
	}
}

func TestSchemaValidator_ArrayDataNode(t *testing.T) {
	validator := NewSchemaValidator()

	// ArrayDataNode represents JSON array data values (not schemas)
	// It should always validate successfully as no schema validation is needed
	elements := []ast.SchemaNode{
		ast.NewLiteralNode("value1", ast.Position{Line: 1, Column: 1}),
		ast.NewLiteralNode("value2", ast.Position{Line: 1, Column: 10}),
		ast.NewLiteralNode(int64(42), ast.Position{Line: 1, Column: 20}),
	}

	node := ast.NewArrayDataNode(elements, ast.Position{Line: 1, Column: 1})
	result := validator.ValidateAll(node)

	if !result.Valid {
		t.Errorf("ValidateAll(ArrayDataNode) result.Valid = false, want true. Errors: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("ValidateAll(ArrayDataNode) found %d errors, want 0. Errors: %v", len(result.Errors), result.Errors)
	}
}
