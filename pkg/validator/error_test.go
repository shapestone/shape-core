package validator

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

// TestValidationError_Error tests the Error() method formatting
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		contains []string
	}{
		{
			name: "basic error with position",
			err: &ValidationError{
				Position: ast.Position{Line: 5, Column: 10},
				Message:  "unknown type: CountryCode",
			},
			contains: []string{"line 5", "column 10", "unknown type: CountryCode"},
		},
		{
			name: "error with path",
			err: &ValidationError{
				Position: ast.Position{Line: 8, Column: 15},
				Message:  "invalid argument count",
				Path:     "$.user.age",
			},
			contains: []string{"line 8", "column 15", "$.user.age", "invalid argument count"},
		},
		{
			name: "error with hint",
			err: &ValidationError{
				Position: ast.Position{Line: 3, Column: 5},
				Message:  "unknown type: CountryCode",
				Hint:     "Available types: UUID, Email, String, Integer",
			},
			contains: []string{"line 3", "column 5", "unknown type", "Available types"},
		},
		{
			name: "error with code",
			err: &ValidationError{
				Position: ast.Position{Line: 1, Column: 1},
				Message:  "unknown function: ArrayOf",
				Code:     ErrCodeUnknownFunction,
			},
			contains: []string{"line 1", "column 1", "unknown function"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errStr, substr) {
					t.Errorf("Error() = %q, want to contain %q", errStr, substr)
				}
			}
		})
	}
}

// TestErrorCodes_Defined tests that all error codes are defined
func TestErrorCodes_Defined(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeUnknownType,
		ErrCodeUnknownFunction,
		ErrCodeInvalidArgCount,
		ErrCodeInvalidArgType,
		ErrCodeInvalidArgValue,
		ErrCodeCircularReference,
	}

	for _, code := range codes {
		if code == "" {
			t.Error("Error code is empty string")
		}
	}
}

// TestErrorCode_UnknownType tests UNKNOWN_TYPE error code
func TestErrorCode_UnknownType(t *testing.T) {
	if ErrCodeUnknownType != "UNKNOWN_TYPE" {
		t.Errorf("ErrCodeUnknownType = %q, want %q", ErrCodeUnknownType, "UNKNOWN_TYPE")
	}
}

// TestErrorCode_UnknownFunction tests UNKNOWN_FUNCTION error code
func TestErrorCode_UnknownFunction(t *testing.T) {
	if ErrCodeUnknownFunction != "UNKNOWN_FUNCTION" {
		t.Errorf("ErrCodeUnknownFunction = %q, want %q", ErrCodeUnknownFunction, "UNKNOWN_FUNCTION")
	}
}

// TestErrorCode_InvalidArgCount tests INVALID_ARG_COUNT error code
func TestErrorCode_InvalidArgCount(t *testing.T) {
	if ErrCodeInvalidArgCount != "INVALID_ARG_COUNT" {
		t.Errorf("ErrCodeInvalidArgCount = %q, want %q", ErrCodeInvalidArgCount, "INVALID_ARG_COUNT")
	}
}

// TestErrorCode_InvalidArgType tests INVALID_ARG_TYPE error code
func TestErrorCode_InvalidArgType(t *testing.T) {
	if ErrCodeInvalidArgType != "INVALID_ARG_TYPE" {
		t.Errorf("ErrCodeInvalidArgType = %q, want %q", ErrCodeInvalidArgType, "INVALID_ARG_TYPE")
	}
}

// TestErrorCode_InvalidArgValue tests INVALID_ARG_VALUE error code
func TestErrorCode_InvalidArgValue(t *testing.T) {
	if ErrCodeInvalidArgValue != "INVALID_ARG_VALUE" {
		t.Errorf("ErrCodeInvalidArgValue = %q, want %q", ErrCodeInvalidArgValue, "INVALID_ARG_VALUE")
	}
}

// TestErrorCode_CircularReference tests CIRCULAR_REFERENCE error code
func TestErrorCode_CircularReference(t *testing.T) {
	if ErrCodeCircularReference != "CIRCULAR_REFERENCE" {
		t.Errorf("ErrCodeCircularReference = %q, want %q", ErrCodeCircularReference, "CIRCULAR_REFERENCE")
	}
}

// TestValidationError_WithPath tests error with JSONPath
func TestValidationError_WithPath(t *testing.T) {
	err := &ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.user.profile.email",
		Message:  "invalid type",
		Code:     ErrCodeUnknownType,
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "$.user.profile.email") {
		t.Errorf("Error() should contain path, got: %s", errStr)
	}
}

// TestValidationError_WithHint tests error with helpful hint
func TestValidationError_WithHint(t *testing.T) {
	err := &ValidationError{
		Position: ast.Position{Line: 3, Column: 8},
		Message:  "unknown type: Emaill",
		Code:     ErrCodeUnknownType,
		Hint:     "Did you mean 'Email'?",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "Did you mean 'Email'?") {
		t.Errorf("Error() should contain hint, got: %s", errStr)
	}
}

// TestValidationError_AllFields tests error with all fields populated
func TestValidationError_AllFields(t *testing.T) {
	err := &ValidationError{
		Position: ast.Position{Line: 10, Column: 25},
		Path:     "$.data.items[].id",
		Message:  "unknown type: UIID",
		Code:     ErrCodeUnknownType,
		Hint:     "Did you mean 'UUID'?",
	}

	errStr := err.Error()
	required := []string{
		"line 10",
		"column 25",
		"$.data.items[].id",
		"unknown type: UIID",
		"Did you mean 'UUID'?",
	}

	for _, req := range required {
		if !strings.Contains(errStr, req) {
			t.Errorf("Error() missing %q, got: %s", req, errStr)
		}
	}
}
