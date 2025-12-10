package validator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidationResult represents the result of validating a schema.
// It collects ALL errors found during validation (not just the first one).
type ValidationResult struct {
	Valid  bool              // True if no errors were found
	Errors []ValidationError // All errors found during validation
}

// NewValidationResult creates a new validation result with no errors.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}
}

// AddError adds a validation error to the result.
// Marks the result as invalid.
func (r *ValidationResult) AddError(err ValidationError) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

// ErrorCount returns the number of errors in the result.
func (r *ValidationResult) ErrorCount() int {
	return len(r.Errors)
}

// HasErrors returns true if there are any errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// GetErrors returns a copy of the errors slice.
func (r *ValidationResult) GetErrors() []ValidationError {
	// Return a copy to prevent modification of internal state
	errors := make([]ValidationError, len(r.Errors))
	copy(errors, r.Errors)
	return errors
}

// FirstError returns the first error, or nil if there are no errors.
func (r *ValidationResult) FirstError() *ValidationError {
	if len(r.Errors) == 0 {
		return nil
	}
	return &r.Errors[0]
}

// ErrorsByCode returns all errors matching the given error code.
func (r *ValidationResult) ErrorsByCode(code ErrorCode) []ValidationError {
	var filtered []ValidationError
	for _, err := range r.Errors {
		if err.Code == code {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// Clear clears all errors and marks the result as valid.
func (r *ValidationResult) Clear() {
	r.Valid = true
	r.Errors = []ValidationError{}
}

// String returns a human-readable string representation of the result.
func (r *ValidationResult) String() string {
	if r.Valid {
		return "Validation passed: valid schema (0 errors)"
	}

	errorCount := len(r.Errors)
	var sb strings.Builder

	// Header
	if errorCount == 1 {
		sb.WriteString("Validation failed: 1 validation error found\n")
	} else {
		sb.WriteString(fmt.Sprintf("Validation failed: %d validation errors found\n", errorCount))
	}

	// List all errors
	for i, err := range r.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}

	return sb.String()
}

// FormatPlain formats all errors as plain text (no colors).
// This is suitable for log files, non-terminal output, or when colors are disabled.
func (r *ValidationResult) FormatPlain() string {
	if r.Valid {
		return "Schema is valid\n"
	}

	// Single error - format directly
	if len(r.Errors) == 1 {
		return r.Errors[0].FormatPlain()
	}

	// Multiple errors - format with separators
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Found %d validation errors:\n\n", len(r.Errors)))

	for i, err := range r.Errors {
		buf.WriteString(fmt.Sprintf("Error %d:\n", i+1))
		buf.WriteString(err.FormatPlain())
		if i < len(r.Errors)-1 {
			buf.WriteString("\n---\n\n")
		}
	}

	return buf.String()
}

// FormatColored formats all errors with ANSI colors for terminal display.
// Automatically falls back to plain formatting if NO_COLOR is set.
func (r *ValidationResult) FormatColored() string {
	if r.Valid {
		return green("Schema is valid\n")
	}

	// Single error - format directly
	if len(r.Errors) == 1 {
		return r.Errors[0].FormatColored()
	}

	// Multiple errors - format with colored separators
	var buf strings.Builder
	buf.WriteString(red(fmt.Sprintf("Found %d validation errors:\n\n", len(r.Errors))))

	for i, err := range r.Errors {
		buf.WriteString(cyan(fmt.Sprintf("Error %d:\n", i+1)))
		buf.WriteString(err.FormatColored())
		if i < len(r.Errors)-1 {
			buf.WriteString("\n" + gray("---") + "\n\n")
		}
	}

	return buf.String()
}

// ToJSON returns a JSON representation of the validation result.
func (r *ValidationResult) ToJSON() ([]byte, error) {
	type jsonResult struct {
		Valid      bool              `json:"valid"`
		ErrorCount int               `json:"errorCount"`
		Errors     []ValidationError `json:"errors,omitempty"`
	}

	jr := jsonResult{
		Valid:      r.Valid,
		ErrorCount: len(r.Errors),
		Errors:     r.Errors,
	}

	return json.MarshalIndent(jr, "", "  ")
}
