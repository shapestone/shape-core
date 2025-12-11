package validator

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

// TestValidationError_FormatPlain tests plain text formatting without colors
func TestValidationError_FormatPlain(t *testing.T) {
	err := ValidationError{
		Position: ast.Position{Line: 8, Column: 20},
		Path:     "$.user.country",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: CountryCode",
		Hint:     "Did you mean 'String'?",
		SourceLines: []string{
			`    "age": Integer(1, 120),`,
			`    "country": CountryCode`,
			`}`,
		},
	}

	result := err.FormatPlain()

	// Check that result contains expected components
	if !strings.Contains(result, "Line 8, Column 20") {
		t.Errorf("Expected position info, got: %s", result)
	}
	if !strings.Contains(result, "$.user.country") {
		t.Errorf("Expected path in output, got: %s", result)
	}
	if !strings.Contains(result, "ERROR [UNKNOWN_TYPE]") {
		t.Errorf("Expected error code in output, got: %s", result)
	}
	if !strings.Contains(result, "unknown type: CountryCode") {
		t.Errorf("Expected error message in output, got: %s", result)
	}
	if !strings.Contains(result, "HINT: Did you mean 'String'?") {
		t.Errorf("Expected hint in output, got: %s", result)
	}

	// Check source context formatting
	if !strings.Contains(result, ">  8 |") {
		t.Errorf("Expected error line marker, got: %s", result)
	}
	if !strings.Contains(result, "^") {
		t.Errorf("Expected column pointer, got: %s", result)
	}

	// Ensure no ANSI color codes
	if strings.Contains(result, "\033[") {
		t.Errorf("Plain format should not contain ANSI codes, got: %s", result)
	}
}

// TestValidationError_FormatPlain_NoSourceContext tests plain formatting without source
func TestValidationError_FormatPlain_NoSourceContext(t *testing.T) {
	err := ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.id",
		Code:     ErrCodeUnknownFunction,
		Message:  "unknown function: validate",
		Hint:     "Did you mean 'Validate'?",
	}

	result := err.FormatPlain()

	if !strings.Contains(result, "Line 5, Column 10") {
		t.Errorf("Expected position info")
	}
	if !strings.Contains(result, "ERROR [UNKNOWN_FUNCTION]") {
		t.Errorf("Expected error code")
	}
	if !strings.Contains(result, "HINT: Did you mean 'Validate'?") {
		t.Errorf("Expected hint")
	}
}

// TestValidationError_FormatColored tests colored formatting
func TestValidationError_FormatColored(t *testing.T) {
	// Ensure colors are enabled for this test
	os.Unsetenv("NO_COLOR")

	err := ValidationError{
		Position: ast.Position{Line: 8, Column: 20},
		Path:     "$.user.country",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: CountryCode",
		Hint:     "Did you mean 'String'?",
		SourceLines: []string{
			`    "age": Integer(1, 120),`,
			`    "country": CountryCode`,
			`}`,
		},
	}

	result := err.FormatColored()

	// Should contain ANSI color codes
	if !strings.Contains(result, "\033[") {
		t.Errorf("Colored format should contain ANSI codes, got: %s", result)
	}

	// Should contain cyan for position
	if !strings.Contains(result, colorCyan) {
		t.Errorf("Expected cyan color for position")
	}

	// Should contain red for error
	if !strings.Contains(result, colorRed) {
		t.Errorf("Expected red color for error")
	}

	// Should contain blue for hint
	if !strings.Contains(result, colorBlue) {
		t.Errorf("Expected blue color for hint")
	}

	// Should contain gray for context lines
	if !strings.Contains(result, colorGray) {
		t.Errorf("Expected gray color for context lines")
	}
}

// TestValidationError_FormatColored_NO_COLOR tests NO_COLOR support
func TestValidationError_FormatColored_NO_COLOR(t *testing.T) {
	// Set NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	err := ValidationError{
		Position: ast.Position{Line: 8, Column: 20},
		Path:     "$.user.country",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: CountryCode",
		Hint:     "Did you mean 'String'?",
	}

	result := err.FormatColored()

	// Should NOT contain ANSI color codes when NO_COLOR is set
	if strings.Contains(result, "\033[") {
		t.Errorf("Colored format should respect NO_COLOR, got: %s", result)
	}

	// Should still contain content
	if !strings.Contains(result, "Line 8, Column 20") {
		t.Errorf("Expected position info")
	}
}

// TestValidationError_ToJSON tests JSON serialization
func TestValidationError_ToJSON(t *testing.T) {
	err := ValidationError{
		Position: ast.Position{Line: 8, Column: 20},
		Path:     "$.user.country",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: CountryCode",
		Hint:     "Did you mean 'String'?",
	}

	jsonBytes, jsonErr := err.ToJSON()
	if jsonErr != nil {
		t.Fatalf("ToJSON failed: %v", jsonErr)
	}

	// Parse back to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify fields
	if parsed["Code"] != string(ErrCodeUnknownType) {
		t.Errorf("Expected Code=%s, got %v", ErrCodeUnknownType, parsed["Code"])
	}
	if parsed["Message"] != "unknown type: CountryCode" {
		t.Errorf("Expected Message")
	}
	if parsed["Path"] != "$.user.country" {
		t.Errorf("Expected Path")
	}
	if parsed["Hint"] != "Did you mean 'String'?" {
		t.Errorf("Expected Hint")
	}

	// Check position is nested object
	pos, ok := parsed["Position"].(map[string]interface{})
	if !ok {
		t.Fatalf("Position should be object")
	}
	if pos["Line"] != float64(8) {
		t.Errorf("Expected Line=8")
	}
	if pos["Column"] != float64(20) {
		t.Errorf("Expected Column=20")
	}
}

// TestValidationResult_FormatPlain tests plain formatting of validation results
func TestValidationResult_FormatPlain(t *testing.T) {
	result := NewValidationResult()
	result.AddError(ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.user",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
		Hint:     "Did you mean 'String'?",
	})

	output := result.FormatPlain()

	if !strings.Contains(output, "Line 5, Column 10") {
		t.Errorf("Expected error details in output")
	}
	if !strings.Contains(output, "unknown type: User") {
		t.Errorf("Expected error message")
	}
}

// TestValidationResult_FormatPlain_Valid tests formatting of valid result
func TestValidationResult_FormatPlain_Valid(t *testing.T) {
	result := NewValidationResult()

	output := result.FormatPlain()

	if !strings.Contains(output, "Schema is valid") {
		t.Errorf("Expected valid message, got: %s", output)
	}
}

// TestValidationResult_FormatPlain_MultipleErrors tests multi-error formatting
func TestValidationResult_FormatPlain_MultipleErrors(t *testing.T) {
	result := NewValidationResult()
	result.AddError(ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.user",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
	})
	result.AddError(ValidationError{
		Position: ast.Position{Line: 8, Column: 15},
		Path:     "$.email",
		Code:     ErrCodeUnknownFunction,
		Message:  "unknown function: emailValidator",
	})
	result.AddError(ValidationError{
		Position: ast.Position{Line: 10, Column: 5},
		Path:     "$.age",
		Code:     ErrCodeInvalidArgCount,
		Message:  "Integer requires at least 2 arguments",
	})

	output := result.FormatPlain()

	// Check header
	if !strings.Contains(output, "Found 3 validation errors") {
		t.Errorf("Expected error count header, got: %s", output)
	}

	// Check all errors are present
	if !strings.Contains(output, "Error 1:") {
		t.Errorf("Expected Error 1 label")
	}
	if !strings.Contains(output, "Error 2:") {
		t.Errorf("Expected Error 2 label")
	}
	if !strings.Contains(output, "Error 3:") {
		t.Errorf("Expected Error 3 label")
	}

	// Check separator
	if !strings.Contains(output, "---") {
		t.Errorf("Expected separator between errors")
	}

	// Check each error message
	if !strings.Contains(output, "unknown type: User") {
		t.Errorf("Expected first error message")
	}
	if !strings.Contains(output, "unknown function: emailValidator") {
		t.Errorf("Expected second error message")
	}
	if !strings.Contains(output, "Integer requires at least 2 arguments") {
		t.Errorf("Expected third error message")
	}
}

// TestValidationResult_FormatColored tests colored formatting
func TestValidationResult_FormatColored(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := NewValidationResult()
	result.AddError(ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.user",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
	})

	output := result.FormatColored()

	// Should contain ANSI codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("Colored output should contain ANSI codes")
	}
}

// TestValidationResult_FormatColored_Valid tests colored valid result
func TestValidationResult_FormatColored_Valid(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := NewValidationResult()
	output := result.FormatColored()

	if !strings.Contains(output, "Schema is valid") {
		t.Errorf("Expected valid message")
	}
	// Should have green color for valid result
	if !strings.Contains(output, colorGreen) {
		t.Errorf("Expected green color for valid result")
	}
}

// TestValidationResult_FormatColored_MultipleErrors tests multi-error colored output
func TestValidationResult_FormatColored_MultipleErrors(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := NewValidationResult()
	result.AddError(ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
	})
	result.AddError(ValidationError{
		Position: ast.Position{Line: 8, Column: 15},
		Code:     ErrCodeUnknownFunction,
		Message:  "unknown function: emailValidator",
	})

	output := result.FormatColored()

	// Should have red for error count
	if !strings.Contains(output, colorRed) {
		t.Errorf("Expected red color for error header")
	}

	// Should have cyan for error numbers
	if !strings.Contains(output, colorCyan) {
		t.Errorf("Expected cyan color for error labels")
	}

	// Should have gray for separators
	if !strings.Contains(output, colorGray) {
		t.Errorf("Expected gray color for separators")
	}
}

// TestValidationResult_ToJSON tests JSON serialization of result
func TestValidationResult_ToJSON(t *testing.T) {
	result := NewValidationResult()
	result.AddError(ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Path:     "$.user",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
	})
	result.AddError(ValidationError{
		Position: ast.Position{Line: 8, Column: 15},
		Path:     "$.email",
		Code:     ErrCodeUnknownFunction,
		Message:  "unknown function: emailValidator",
	})

	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse back
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check structure
	if parsed["valid"] != false {
		t.Errorf("Expected valid=false")
	}
	if parsed["errorCount"] != float64(2) {
		t.Errorf("Expected errorCount=2, got %v", parsed["errorCount"])
	}

	// Check errors array
	errors, ok := parsed["errors"].([]interface{})
	if !ok {
		t.Fatalf("Expected errors array")
	}
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

// TestValidationResult_ToJSON_Valid tests JSON for valid result
func TestValidationResult_ToJSON_Valid(t *testing.T) {
	result := NewValidationResult()

	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["valid"] != true {
		t.Errorf("Expected valid=true")
	}
	if parsed["errorCount"] != float64(0) {
		t.Errorf("Expected errorCount=0")
	}
}

// TestSourceContext_Display tests source context extraction and display
func TestSourceContext_Display(t *testing.T) {
	sourceText := `{
    "name": String,
    "age": Integer(1, 120),
    "country": CountryCode,
    "email": Email
}`

	err := ValidationError{
		Position: ast.Position{Line: 4, Column: 16},
		Path:     "$.country",
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: CountryCode",
		Source:   sourceText,
		SourceLines: []string{
			`    "name": String,`,
			`    "age": Integer(1, 120),`,
			`    "country": CountryCode,`,
			`    "email": Email`,
			`}`,
		},
	}

	output := err.FormatPlain()

	// Check that context lines are shown
	if !strings.Contains(output, `"age": Integer(1, 120)`) {
		t.Errorf("Expected context line before error")
	}
	if !strings.Contains(output, `"country": CountryCode`) {
		t.Errorf("Expected error line")
	}
	if !strings.Contains(output, `"email": Email`) {
		t.Errorf("Expected context line after error")
	}

	// Check line numbers
	if !strings.Contains(output, "  2 |") {
		t.Errorf("Expected line 2 number")
	}
	if !strings.Contains(output, "  3 |") {
		t.Errorf("Expected line 3 number")
	}
	if !strings.Contains(output, ">  4 |") {
		t.Errorf("Expected line 4 with error marker")
	}
	if !strings.Contains(output, "  5 |") {
		t.Errorf("Expected line 5 number")
	}

	// Check column pointer
	if !strings.Contains(output, "^") {
		t.Errorf("Expected column pointer")
	}
}

// TestSourceContext_EdgeCases tests edge cases in source context
func TestSourceContext_EdgeCases(t *testing.T) {
	t.Run("FirstLine", func(t *testing.T) {
		// Error on first line - addSourceContext will create SourceLines starting from line 1
		// because max(1, 1-2) = 1
		err := ValidationError{
			Position: ast.Position{Line: 1, Column: 5},
			SourceLines: []string{
				`{ "error": BadType }`, // Line 1 (error line)
				`  "valid": String`,    // Line 2
				`}`,                    // Line 3
			},
		}

		output := err.FormatPlain()
		if !strings.Contains(output, ">  1 |") {
			t.Errorf("Expected error marker on line 1, got:\n%s", output)
		}
		if !strings.Contains(output, `{ "error": BadType }`) {
			t.Errorf("Expected error line content")
		}
	})

	t.Run("LastLine", func(t *testing.T) {
		// Error on line 3 - SourceLines starts at line 1 (max(1, 3-2) = 1)
		err := ValidationError{
			Position: ast.Position{Line: 3, Column: 5},
			SourceLines: []string{
				`  "name": String,`, // Line 1
				`  "age": Integer`,  // Line 2
				`}`,                 // Line 3 (error line)
			},
		}

		output := err.FormatPlain()
		if !strings.Contains(output, ">  3 |") {
			t.Errorf("Expected error marker on line 3, got:\n%s", output)
		}
	})
}

// TestNO_COLOR_RespectedEverywhere tests NO_COLOR is respected throughout
func TestNO_COLOR_RespectedEverywhere(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	// Test single error
	err := ValidationError{
		Position: ast.Position{Line: 5, Column: 10},
		Code:     ErrCodeUnknownType,
		Message:  "unknown type: User",
		Hint:     "Did you mean 'String'?",
	}

	if strings.Contains(err.FormatColored(), "\033[") {
		t.Errorf("FormatColored should respect NO_COLOR for errors")
	}

	// Test result
	result := NewValidationResult()
	result.AddError(err)

	if strings.Contains(result.FormatColored(), "\033[") {
		t.Errorf("FormatColored should respect NO_COLOR for results")
	}

	// Test valid result
	validResult := NewValidationResult()
	if strings.Contains(validResult.FormatColored(), "\033[") {
		t.Errorf("FormatColored should respect NO_COLOR for valid results")
	}
}

// TestColor_Functions tests individual color helper functions
func TestColor_Functions(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name      string
		fn        func(string) string
		colorCode string
	}{
		{"red", red, colorRed},
		{"blue", blue, colorBlue},
		{"yellow", yellow, colorYellow},
		{"cyan", cyan, colorCyan},
		{"gray", gray, colorGray},
		{"green", green, colorGreen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("test")
			if !strings.Contains(result, tt.colorCode) {
				t.Errorf("Expected color code %s in result", tt.colorCode)
			}
			if !strings.Contains(result, colorReset) {
				t.Errorf("Expected reset code in result")
			}
			if !strings.Contains(result, "test") {
				t.Errorf("Expected original text in result")
			}
		})
	}

	// Test with NO_COLOR
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	for _, tt := range tests {
		t.Run(tt.name+"_NO_COLOR", func(t *testing.T) {
			result := tt.fn("test")
			if strings.Contains(result, "\033[") {
				t.Errorf("Should not contain ANSI codes with NO_COLOR set")
			}
			if result != "test" {
				t.Errorf("Expected plain text with NO_COLOR, got: %s", result)
			}
		})
	}
}
