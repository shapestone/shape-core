package validator

import (
	"strings"
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

// TestValidationResult_Valid tests that new results are valid by default
func TestValidationResult_Valid(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if !result.Valid {
		t.Error("New ValidationResult should be valid by default")
	}

	if len(result.Errors) != 0 {
		t.Errorf("New ValidationResult should have 0 errors, got %d", len(result.Errors))
	}
}

// TestValidationResult_AddError tests adding errors marks result invalid
func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	err := ValidationError{
		Position: ast.Position{Line: 1, Column: 1},
		Message:  "test error",
		Code:     ErrCodeUnknownType,
	}

	result.AddError(err)

	if result.Valid {
		t.Error("ValidationResult should be invalid after adding error")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidationResult should have 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Message != "test error" {
		t.Errorf("Error message = %q, want %q", result.Errors[0].Message, "test error")
	}
}

// TestValidationResult_AddMultipleErrors tests adding multiple errors
func TestValidationResult_AddMultipleErrors(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	err1 := ValidationError{
		Position: ast.Position{Line: 1, Column: 1},
		Message:  "error 1",
		Code:     ErrCodeUnknownType,
	}

	err2 := ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Message:  "error 2",
		Code:     ErrCodeInvalidArgCount,
	}

	err3 := ValidationError{
		Position: ast.Position{Line: 8, Column: 15},
		Message:  "error 3",
		Code:     ErrCodeUnknownFunction,
	}

	result.AddError(err1)
	result.AddError(err2)
	result.AddError(err3)

	if result.Valid {
		t.Error("ValidationResult should be invalid after adding errors")
	}

	if len(result.Errors) != 3 {
		t.Errorf("ValidationResult should have 3 errors, got %d", len(result.Errors))
	}

	// Verify order is preserved
	if result.Errors[0].Message != "error 1" {
		t.Errorf("First error message = %q, want %q", result.Errors[0].Message, "error 1")
	}
	if result.Errors[1].Message != "error 2" {
		t.Errorf("Second error message = %q, want %q", result.Errors[1].Message, "error 2")
	}
	if result.Errors[2].Message != "error 3" {
		t.Errorf("Third error message = %q, want %q", result.Errors[2].Message, "error 3")
	}
}

// TestValidationResult_ErrorCount tests ErrorCount() method
func TestValidationResult_ErrorCount(t *testing.T) {
	tests := []struct {
		name       string
		errorCount int
	}{
		{"no errors", 0},
		{"one error", 1},
		{"multiple errors", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:  tt.errorCount == 0,
				Errors: make([]ValidationError, tt.errorCount),
			}

			count := result.ErrorCount()
			if count != tt.errorCount {
				t.Errorf("ErrorCount() = %d, want %d", count, tt.errorCount)
			}
		})
	}
}

// TestValidationResult_String_SingleError tests String() for single error
func TestValidationResult_String_SingleError(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Position: ast.Position{Line: 5, Column: 10},
				Message:  "unknown type: CountryCode",
				Code:     ErrCodeUnknownType,
			},
		},
	}

	str := result.String()
	if !strings.Contains(str, "1 validation error") {
		t.Errorf("String() should contain '1 validation error', got: %s", str)
	}
	if !strings.Contains(str, "unknown type: CountryCode") {
		t.Errorf("String() should contain error message, got: %s", str)
	}
}

// TestValidationResult_String_MultipleErrors tests String() for multiple errors
func TestValidationResult_String_MultipleErrors(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Position: ast.Position{Line: 3, Column: 5},
				Message:  "unknown type: CountryCode",
				Code:     ErrCodeUnknownType,
			},
			{
				Position: ast.Position{Line: 8, Column: 12},
				Message:  "invalid argument count",
				Code:     ErrCodeInvalidArgCount,
			},
			{
				Position: ast.Position{Line: 15, Column: 20},
				Message:  "unknown function: ArrayOf",
				Code:     ErrCodeUnknownFunction,
			},
		},
	}

	str := result.String()
	if !strings.Contains(str, "3 validation errors") {
		t.Errorf("String() should contain '3 validation errors', got: %s", str)
	}

	// Should contain all error messages
	if !strings.Contains(str, "unknown type: CountryCode") {
		t.Errorf("String() should contain first error, got: %s", str)
	}
	if !strings.Contains(str, "invalid argument count") {
		t.Errorf("String() should contain second error, got: %s", str)
	}
	if !strings.Contains(str, "unknown function: ArrayOf") {
		t.Errorf("String() should contain third error, got: %s", str)
	}
}

// TestValidationResult_String_Valid tests String() for valid result
func TestValidationResult_String_Valid(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	str := result.String()
	if !strings.Contains(str, "valid") || !strings.Contains(str, "0 errors") {
		t.Errorf("String() should indicate validation passed, got: %s", str)
	}
}

// TestValidationResult_HasErrors tests HasErrors() method
func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name      string
		result    *ValidationResult
		hasErrors bool
	}{
		{
			name: "no errors",
			result: &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			},
			hasErrors: false,
		},
		{
			name: "has errors",
			result: &ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{Message: "test error"},
				},
			},
			hasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErrors := tt.result.HasErrors()
			if hasErrors != tt.hasErrors {
				t.Errorf("HasErrors() = %v, want %v", hasErrors, tt.hasErrors)
			}
		})
	}
}

// TestValidationResult_GetErrors tests GetErrors() method
func TestValidationResult_GetErrors(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Message: "error 1"},
			{Message: "error 2"},
		},
	}

	errors := result.GetErrors()
	if len(errors) != 2 {
		t.Errorf("GetErrors() returned %d errors, want 2", len(errors))
	}

	// Verify it returns a copy, not the original
	errors[0].Message = "modified"
	if result.Errors[0].Message == "modified" {
		t.Error("GetErrors() should return a copy, not the original slice")
	}
}

// TestValidationResult_FirstError tests getting the first error
func TestValidationResult_FirstError(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Message: "first error", Code: ErrCodeUnknownType},
			{Message: "second error", Code: ErrCodeInvalidArgCount},
		},
	}

	firstErr := result.FirstError()
	if firstErr == nil {
		t.Fatal("FirstError() returned nil, want error")
	}

	if firstErr.Message != "first error" {
		t.Errorf("FirstError() message = %q, want %q", firstErr.Message, "first error")
	}
	if firstErr.Code != ErrCodeUnknownType {
		t.Errorf("FirstError() code = %q, want %q", firstErr.Code, ErrCodeUnknownType)
	}
}

// TestValidationResult_FirstError_NoErrors tests FirstError() with no errors
func TestValidationResult_FirstError_NoErrors(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	firstErr := result.FirstError()
	if firstErr != nil {
		t.Errorf("FirstError() = %v, want nil", firstErr)
	}
}

// TestValidationResult_ErrorsByCode tests filtering errors by code
func TestValidationResult_ErrorsByCode(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Message: "unknown type 1", Code: ErrCodeUnknownType},
			{Message: "invalid arg", Code: ErrCodeInvalidArgCount},
			{Message: "unknown type 2", Code: ErrCodeUnknownType},
			{Message: "unknown function", Code: ErrCodeUnknownFunction},
		},
	}

	typeErrors := result.ErrorsByCode(ErrCodeUnknownType)
	if len(typeErrors) != 2 {
		t.Errorf("ErrorsByCode(UNKNOWN_TYPE) returned %d errors, want 2", len(typeErrors))
	}

	for _, err := range typeErrors {
		if err.Code != ErrCodeUnknownType {
			t.Errorf("ErrorsByCode returned error with code %q, want %q", err.Code, ErrCodeUnknownType)
		}
	}
}

// TestValidationResult_Clear tests clearing errors
func TestValidationResult_Clear(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Message: "error 1"},
			{Message: "error 2"},
		},
	}

	result.Clear()

	if !result.Valid {
		t.Error("After Clear(), Valid should be true")
	}

	if len(result.Errors) != 0 {
		t.Errorf("After Clear(), should have 0 errors, got %d", len(result.Errors))
	}
}
