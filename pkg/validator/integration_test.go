package validator

import (
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

// TestValidation_Integration_ValidSchema tests full validation flow with valid schema
func TestValidation_Integration_ValidSchema(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":    ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),
			"email": ast.NewTypeNode("Email", ast.Position{Line: 3, Column: 12}),
			"age":   ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(120)}, ast.Position{Line: 4, Column: 10}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Valid schema should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Valid schema should have 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_UnknownType tests full flow with unknown type error
func TestValidation_Integration_UnknownType(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 5, Column: 15}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema with unknown type should fail validation")
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Should have 1 error, got %d", result.ErrorCount())
	}

	firstErr := result.FirstError()
	if firstErr == nil {
		t.Fatal("FirstError() should not be nil")
	}

	if firstErr.Code != ErrCodeUnknownType {
		t.Errorf("Error code = %q, want %q", firstErr.Code, ErrCodeUnknownType)
	}
}

// TestValidation_Integration_InvalidArgCount tests invalid argument count error
func TestValidation_Integration_InvalidArgCount(t *testing.T) {
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
		t.Error("Schema with invalid arg count should fail validation")
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Should have 1 error, got %d", result.ErrorCount())
	}

	firstErr := result.FirstError()
	if firstErr.Code != ErrCodeInvalidArgCount {
		t.Errorf("Error code = %q, want %q", firstErr.Code, ErrCodeInvalidArgCount)
	}
}

// TestValidation_Integration_MultipleErrors tests collecting multiple errors
func TestValidation_Integration_MultipleErrors(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":      ast.NewTypeNode("UIID", ast.Position{Line: 2, Column: 10}),                                                         // Unknown type (typo)
			"age":     ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 3, Column: 10}), // Invalid arg count
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 15}),                                                  // Unknown type
			"data":    ast.NewFunctionNode("ArrayOf", []interface{}{"String"}, ast.Position{Line: 5, Column: 12}),                         // Unknown function
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema with multiple errors should fail validation")
	}

	if result.ErrorCount() != 4 {
		t.Errorf("Should collect 4 errors, got %d", result.ErrorCount())
	}

	// Verify error types
	errorsByCode := make(map[ErrorCode]int)
	for _, err := range result.Errors {
		errorsByCode[err.Code]++
	}

	if errorsByCode[ErrCodeUnknownType] != 2 {
		t.Errorf("Should have 2 UNKNOWN_TYPE errors, got %d", errorsByCode[ErrCodeUnknownType])
	}

	if errorsByCode[ErrCodeInvalidArgCount] != 1 {
		t.Errorf("Should have 1 INVALID_ARG_COUNT error, got %d", errorsByCode[ErrCodeInvalidArgCount])
	}

	if errorsByCode[ErrCodeUnknownFunction] != 1 {
		t.Errorf("Should have 1 UNKNOWN_FUNCTION error, got %d", errorsByCode[ErrCodeUnknownFunction])
	}
}

// TestValidation_Integration_NestedObjects tests validation of nested object structures
func TestValidation_Integration_NestedObjects(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"user": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"id": ast.NewTypeNode("UUID", ast.Position{Line: 3, Column: 12}),
					"profile": ast.NewObjectNode(
						map[string]ast.SchemaNode{
							"email": ast.NewTypeNode("Email", ast.Position{Line: 5, Column: 18}),
							"bio":   ast.NewFunctionNode("String", []interface{}{int64(0), int64(500)}, ast.Position{Line: 6, Column: 16}),
						},
						ast.Position{Line: 4, Column: 15},
					),
				},
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Valid nested schema should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Should have 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_NestedObjectsWithErrors tests nested objects with errors at different levels
func TestValidation_Integration_NestedObjectsWithErrors(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"user": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"id":      ast.NewTypeNode("UIID", ast.Position{Line: 3, Column: 12}),        // Error at level 2
					"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 15}), // Error at level 2
					"profile": ast.NewObjectNode(
						map[string]ast.SchemaNode{
							"language": ast.NewTypeNode("Language", ast.Position{Line: 6, Column: 20}), // Error at level 3
						},
						ast.Position{Line: 5, Column: 15},
					),
				},
				ast.Position{Line: 2, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema with nested errors should fail validation")
	}

	if result.ErrorCount() != 3 {
		t.Errorf("Should collect 3 errors from nested levels, got %d", result.ErrorCount())
	}

	// Verify JSONPath tracking
	expectedPaths := []string{"$.user.id", "$.user.country", "$.user.profile.language"}
	actualPaths := make(map[string]bool)
	for _, err := range result.Errors {
		actualPaths[err.Path] = true
	}

	for _, path := range expectedPaths {
		if !actualPaths[path] {
			t.Errorf("Expected error at path %q not found", path)
		}
	}
}

// TestValidation_Integration_Arrays tests validation of array schemas
func TestValidation_Integration_Arrays(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"tags": ast.NewArrayNode(
				ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 2, Column: 13}),
				ast.Position{Line: 2, Column: 10},
			),
			"ids": ast.NewArrayNode(
				ast.NewTypeNode("UUID", ast.Position{Line: 3, Column: 12}),
				ast.Position{Line: 3, Column: 10},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Valid array schema should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Should have 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_ArraysWithErrors tests arrays with invalid element schemas
func TestValidation_Integration_ArraysWithErrors(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"tags": ast.NewArrayNode(
				ast.NewTypeNode("Tag", ast.Position{Line: 2, Column: 13}), // Unknown type
				ast.Position{Line: 2, Column: 10},
			),
			"scores": ast.NewArrayNode(
				ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 3, Column: 15}), // Invalid arg count
				ast.Position{Line: 3, Column: 12},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema with array errors should fail validation")
	}

	if result.ErrorCount() != 2 {
		t.Errorf("Should collect 2 errors from arrays, got %d", result.ErrorCount())
	}

	// Verify JSONPath for array elements
	hasTagsArrayError := false
	hasScoresArrayError := false
	for _, err := range result.Errors {
		if err.Path == "$.tags[]" {
			hasTagsArrayError = true
		}
		if err.Path == "$.scores[]" {
			hasScoresArrayError = true
		}
	}

	if !hasTagsArrayError {
		t.Error("Expected error at $.tags[]")
	}
	if !hasScoresArrayError {
		t.Error("Expected error at $.scores[]")
	}
}

// TestValidation_Integration_CustomTypes tests registering and using custom types
func TestValidation_Integration_CustomTypes(t *testing.T) {
	validator := NewSchemaValidator()

	// Register custom types
	validator.RegisterType("SSN", TypeDescriptor{
		Name:        "SSN",
		Description: "Social Security Number",
	})
	validator.RegisterType("CreditCard", TypeDescriptor{
		Name:        "CreditCard",
		Description: "Credit card number",
	})

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":   ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),       // Built-in
			"ssn":  ast.NewTypeNode("SSN", ast.Position{Line: 3, Column: 10}),        // Custom
			"card": ast.NewTypeNode("CreditCard", ast.Position{Line: 4, Column: 12}), // Custom
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Schema with registered custom types should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Should have 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_CustomFunctions tests registering and using custom functions
func TestValidation_Integration_CustomFunctions(t *testing.T) {
	validator := NewSchemaValidator()

	// Register custom function
	validator.RegisterFunction("PhoneNumber", FunctionDescriptor{
		Name:        "PhoneNumber",
		Description: "Phone number validation",
		MinArgs:     0,
		MaxArgs:     1,
	})

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"age":   ast.NewFunctionNode("Integer", []interface{}{int64(18), int64(120)}, ast.Position{Line: 2, Column: 10}), // Built-in
			"phone": ast.NewFunctionNode("PhoneNumber", []interface{}{}, ast.Position{Line: 3, Column: 12}),                  // Custom
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Schema with registered custom function should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Should have 0 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_MixedValidAndInvalid tests schema with both valid and invalid elements
func TestValidation_Integration_MixedValidAndInvalid(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":       ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),                                                         // Valid
			"email":    ast.NewTypeNode("Email", ast.Position{Line: 3, Column: 12}),                                                        // Valid
			"country":  ast.NewTypeNode("CountryCode", ast.Position{Line: 4, Column: 15}),                                                  // Invalid - unknown type
			"age":      ast.NewFunctionNode("Integer", []interface{}{int64(18), int64(120)}, ast.Position{Line: 5, Column: 10}),            // Valid
			"score":    ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 6, Column: 12}), // Invalid - too many args
			"username": ast.NewFunctionNode("String", []interface{}{int64(3), int64(20)}, ast.Position{Line: 7, Column: 15}),               // Valid
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema with invalid elements should fail validation")
	}

	// Should only report the 2 invalid elements
	if result.ErrorCount() != 2 {
		t.Errorf("Should have exactly 2 errors, got %d", result.ErrorCount())
	}
}

// TestValidation_Integration_ErrorFormatting tests that errors are well-formatted
func TestValidation_Integration_ErrorFormatting(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 5, Column: 15}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	if result.Valid {
		t.Error("Schema should fail validation")
	}

	// Test String() formatting
	resultStr := result.String()
	if resultStr == "" {
		t.Error("result.String() should not be empty")
	}

	// Should contain error count
	// Should contain error details
	t.Logf("Result string: %s", resultStr)

	// Test individual error formatting
	if result.ErrorCount() > 0 {
		firstErr := result.FirstError()
		errStr := firstErr.Error()
		if errStr == "" {
			t.Error("error.Error() should not be empty")
		}
		t.Logf("Error string: %s", errStr)
	}
}

// TestValidation_Integration_ErrorsByCode tests filtering errors by code
func TestValidation_Integration_ErrorsByCode(t *testing.T) {
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":      ast.NewTypeNode("UIID", ast.Position{Line: 2, Column: 10}),                                                         // Unknown type
			"country": ast.NewTypeNode("CountryCode", ast.Position{Line: 3, Column: 15}),                                                  // Unknown type
			"age":     ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100), int64(200)}, ast.Position{Line: 4, Column: 10}), // Invalid arg count
			"data":    ast.NewFunctionNode("ArrayOf", []interface{}{"String"}, ast.Position{Line: 5, Column: 12}),                         // Unknown function
		},
		ast.Position{Line: 1, Column: 1},
	)

	validator := NewSchemaValidator()
	result := validator.ValidateAll(schema)

	// Get errors by code
	typeErrors := result.ErrorsByCode(ErrCodeUnknownType)
	if len(typeErrors) != 2 {
		t.Errorf("Should have 2 UNKNOWN_TYPE errors, got %d", len(typeErrors))
	}

	argCountErrors := result.ErrorsByCode(ErrCodeInvalidArgCount)
	if len(argCountErrors) != 1 {
		t.Errorf("Should have 1 INVALID_ARG_COUNT error, got %d", len(argCountErrors))
	}

	functionErrors := result.ErrorsByCode(ErrCodeUnknownFunction)
	if len(functionErrors) != 1 {
		t.Errorf("Should have 1 UNKNOWN_FUNCTION error, got %d", len(functionErrors))
	}
}

// TestValidation_Integration_ComplexRealWorld tests a real-world complex schema
func TestValidation_Integration_ComplexRealWorld(t *testing.T) {
	validator := NewSchemaValidator()

	// Real-world user registration schema
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":       ast.NewTypeNode("UUID", ast.Position{Line: 2, Column: 10}),
			"username": ast.NewFunctionNode("String", []interface{}{int64(3), int64(20)}, ast.Position{Line: 3, Column: 15}),
			"email":    ast.NewTypeNode("Email", ast.Position{Line: 4, Column: 12}),
			"password": ast.NewFunctionNode("String", []interface{}{int64(8), int64(128)}, ast.Position{Line: 5, Column: 15}),
			"age":      ast.NewFunctionNode("Integer", []interface{}{int64(18), int64(120)}, ast.Position{Line: 6, Column: 10}),
			"status":   ast.NewFunctionNode("Enum", []interface{}{"active", "inactive", "suspended"}, ast.Position{Line: 7, Column: 13}),
			"tags": ast.NewArrayNode(
				ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 8, Column: 13}),
				ast.Position{Line: 8, Column: 10},
			),
			"profile": ast.NewObjectNode(
				map[string]ast.SchemaNode{
					"firstName": ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 10, Column: 18}),
					"lastName":  ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 11, Column: 17}),
					"bio":       ast.NewFunctionNode("String", []interface{}{int64(0), int64(500)}, ast.Position{Line: 12, Column: 13}),
					"avatar":    ast.NewTypeNode("URL", ast.Position{Line: 13, Column: 15}),
					"birthdate": ast.NewTypeNode("Date", ast.Position{Line: 14, Column: 17}),
					"address": ast.NewObjectNode(
						map[string]ast.SchemaNode{
							"street":  ast.NewFunctionNode("String", []interface{}{int64(1), int64(100)}, ast.Position{Line: 16, Column: 18}),
							"city":    ast.NewFunctionNode("String", []interface{}{int64(1), int64(50)}, ast.Position{Line: 17, Column: 16}),
							"zipCode": ast.NewFunctionNode("Pattern", []interface{}{"^[0-9]{5}$"}, ast.Position{Line: 18, Column: 18}),
						},
						ast.Position{Line: 15, Column: 17},
					),
				},
				ast.Position{Line: 9, Column: 15},
			),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := validator.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Complex real-world schema should pass validation. Errors: %v", result.Errors)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Should have 0 errors, got %d", result.ErrorCount())
	}
}
