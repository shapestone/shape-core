package validator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
	"github.com/shapestone/shape/pkg/shape"
	"github.com/shapestone/shape/pkg/validator"
)

// TestE2E_CLI_ValidSchema tests the CLI tool with valid schemas
func TestE2E_CLI_ValidSchema(t *testing.T) {
	// Build the CLI tool first
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	// Create a temporary valid schema file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "valid.jsonv")
	schemaContent := `{
  "id": UUID,
  "email": Email,
  "age": Integer(18, 120)
}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Run CLI tool
	cmd := exec.Command(cliPath, schemaFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("CLI failed for valid schema: %v\nStderr: %s", err, stderr.String())
	}

	// Verify exit code is 0 (implicit in err == nil)
	if exitErr, ok := err.(*exec.ExitError); ok {
		t.Errorf("Expected exit code 0, got %d", exitErr.ExitCode())
	}

	// Output should indicate success or be empty for valid schema
	output := stdout.String()
	t.Logf("CLI output: %s", output)
}

// TestE2E_CLI_InvalidSchema tests the CLI tool with invalid schemas
func TestE2E_CLI_InvalidSchema(t *testing.T) {
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	tests := []struct {
		name           string
		schema         string
		expectedError  string
		expectedExitCode int
	}{
		{
			name: "unknown type",
			schema: `{
  "country": CountryCode
}`,
			expectedError: "UNKNOWN_TYPE",
			expectedExitCode: 1,
		},
		{
			name: "invalid argument count",
			schema: `{
  "age": Integer(1, 100, 200)
}`,
			expectedError: "INVALID_ARG_COUNT",
			expectedExitCode: 1,
		},
		{
			name: "unknown function",
			schema: `{
  "data": ArrayOf("String")
}`,
			expectedError: "UNKNOWN_FUNCTION",
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			schemaFile := filepath.Join(tmpDir, "invalid.jsonv")
			if err := os.WriteFile(schemaFile, []byte(tt.schema), 0644); err != nil {
				t.Fatalf("Failed to write schema file: %v", err)
			}

			cmd := exec.Command(cliPath, schemaFile)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Error("Expected CLI to fail for invalid schema")
			}

			// Check exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != tt.expectedExitCode {
					t.Errorf("Expected exit code %d, got %d", tt.expectedExitCode, exitErr.ExitCode())
				}
			}

			// Verify error message contains expected error code
			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.expectedError) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectedError, output)
			}
		})
	}
}

// TestE2E_CLI_MultipleFiles tests validating multiple files at once
func TestE2E_CLI_MultipleFiles(t *testing.T) {
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	tmpDir := t.TempDir()

	// Create multiple schema files
	validFile1 := filepath.Join(tmpDir, "valid1.jsonv")
	validFile2 := filepath.Join(tmpDir, "valid2.jsonv")
	invalidFile := filepath.Join(tmpDir, "invalid.jsonv")

	os.WriteFile(validFile1, []byte(`{"id": UUID}`), 0644)
	os.WriteFile(validFile2, []byte(`{"email": Email}`), 0644)
	os.WriteFile(invalidFile, []byte(`{"country": CountryCode}`), 0644)

	// Test with all files (should fail due to invalid file)
	cmd := exec.Command(cliPath, validFile1, validFile2, invalidFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("Expected CLI to fail when one file is invalid")
	}

	// Should process all files
	output := stdout.String() + stderr.String()
	t.Logf("Output: %s", output)
}

// TestE2E_CLI_CustomTypes tests the --register-type flag
func TestE2E_CLI_CustomTypes(t *testing.T) {
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "custom.jsonv")
	schemaContent := `{
  "ssn": SSN,
  "card": CreditCard
}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Without custom type registration (should fail)
	cmd := exec.Command(cliPath, schemaFile)
	if err := cmd.Run(); err == nil {
		t.Error("Expected failure without custom type registration")
	}

	// With custom type registration (should succeed)
	cmd = exec.Command(cliPath, "--register-type", "SSN,CreditCard", schemaFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Errorf("Expected success with custom types: %v\nStderr: %s", err, stderr.String())
	}
}

// TestE2E_CLI_OutputFormats tests different output formats
func TestE2E_CLI_OutputFormats(t *testing.T) {
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.jsonv")
	schemaContent := `{"country": CountryCode}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	tests := []struct {
		name         string
		outputFormat string
		checkJSON    bool
	}{
		{"text output", "text", false},
		{"json output", "json", true},
		{"quiet output", "quiet", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(cliPath, "-o", tt.outputFormat, schemaFile)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			cmd.Run() // Ignore error, we expect failure

			output := stdout.String()
			if tt.checkJSON {
				// Verify JSON is valid
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Invalid JSON output: %v\nOutput: %s", err, output)
				}
			}

			if tt.outputFormat == "quiet" && output != "" {
				t.Errorf("Quiet mode should produce no output, got: %s", output)
			}
		})
	}
}

// TestE2E_CLI_NoColorFlag tests the --no-color flag
func TestE2E_CLI_NoColorFlag(t *testing.T) {
	cliPath := buildCLI(t)
	defer os.Remove(cliPath)

	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.jsonv")
	schemaContent := `{"country": CountryCode}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	cmd := exec.Command(cliPath, "--no-color", schemaFile)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	cmd.Run() // Ignore error

	output := stdout.String()
	// Check that ANSI color codes are not present
	if strings.Contains(output, "\x1b[") {
		t.Error("Output contains ANSI color codes despite --no-color flag")
	}
}

// TestE2E_PublicAPI_ParseValidateWorkflow tests the complete public API workflow
func TestE2E_PublicAPI_ParseValidateWorkflow(t *testing.T) {
	tests := []struct {
		name       string
		format     parser.Format
		input      string
		shouldPass bool
		errorCount int
	}{
		{
			name:       "JSONV valid",
			format:     parser.FormatJSONV,
			input:      `{"id": UUID, "email": Email}`,
			shouldPass: true,
			errorCount: 0,
		},
		{
			name:       "JSONV invalid",
			format:     parser.FormatJSONV,
			input:      `{"country": CountryCode}`,
			shouldPass: false,
			errorCount: 1,
		},
		{
			name:       "XMLV valid",
			format:     parser.FormatXMLV,
			input:      `<root><id>UUID</id><email>Email</email></root>`,
			shouldPass: true,
			errorCount: 0,
		},
		{
			name:       "YAMLV valid",
			format:     parser.FormatYAMLV,
			input:      "id: UUID\nemail: Email\n",
			shouldPass: true,
			errorCount: 0,
		},
		{
			name:       "PropsV valid",
			format:     parser.FormatPropsV,
			input:      "id=UUID\nemail=Email\n",
			shouldPass: true,
			errorCount: 0,
		},
		{
			name:       "CSVV valid",
			format:     parser.FormatCSVV,
			input:      "id,UUID\nemail,Email\n",
			shouldPass: true,
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse
			node, err := shape.Parse(tt.format, tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Step 2: Validate
			result := shape.ValidateAll(node, tt.input)

			// Step 3: Verify results
			if result.Valid != tt.shouldPass {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.shouldPass)
			}

			if result.ErrorCount() != tt.errorCount {
				t.Errorf("ErrorCount = %d, want %d", result.ErrorCount(), tt.errorCount)
			}

			// Step 4: Format results (should not panic)
			_ = result.String()
			_ = result.FormatColored()
			_, _ = result.ToJSON()
		})
	}
}

// TestE2E_PublicAPI_MultiErrorCollection tests collecting multiple errors
func TestE2E_PublicAPI_MultiErrorCollection(t *testing.T) {
	input := `{
  "id": UIID,
  "country": CountryCode,
  "age": Integer(1, 100, 200),
  "data": ArrayOf("String")
}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := shape.ValidateAll(node, input)

	if result.Valid {
		t.Error("Schema with multiple errors should be invalid")
	}

	if result.ErrorCount() != 4 {
		t.Errorf("Expected 4 errors, got %d", result.ErrorCount())
	}

	// Verify all error types are collected
	errorCodes := make(map[validator.ErrorCode]int)
	for _, err := range result.Errors {
		errorCodes[err.Code]++
	}

	expectedCodes := map[validator.ErrorCode]int{
		validator.ErrCodeUnknownType:     2, // UIID, CountryCode
		validator.ErrCodeInvalidArgCount: 1, // Integer with 3 args
		validator.ErrCodeUnknownFunction: 1, // ArrayOf
	}

	for code, expectedCount := range expectedCodes {
		if errorCodes[code] != expectedCount {
			t.Errorf("Error code %s: expected %d, got %d", code, expectedCount, errorCodes[code])
		}
	}
}

// TestE2E_PublicAPI_SourceContext tests source context extraction
func TestE2E_PublicAPI_SourceContext(t *testing.T) {
	input := `{
  "id": UUID,
  "country": CountryCode,
  "email": Email
}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := shape.ValidateAll(node, input)

	if result.Valid {
		t.Error("Schema should be invalid")
	}

	firstErr := result.FirstError()
	if firstErr == nil {
		t.Fatal("FirstError() should not be nil")
	}

	// Verify source context is present
	if firstErr.Position.Line == 0 {
		t.Error("Error should have line number")
	}

	if firstErr.Position.Column == 0 {
		t.Error("Error should have column number")
	}

	if len(firstErr.SourceLines) == 0 {
		t.Error("Error should have source lines")
	}

	// Verify formatted output includes source context
	formatted := result.FormatColored()
	if !strings.Contains(formatted, "CountryCode") {
		t.Error("Formatted output should contain source context")
	}
}

// TestE2E_RealWorld_LargeSchema tests a large schema (100+ fields)
func TestE2E_RealWorld_LargeSchema(t *testing.T) {
	// Build a large schema programmatically
	fields := make(map[string]ast.SchemaNode)
	for i := 0; i < 100; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		fields[fieldName] = ast.NewTypeNode("UUID", ast.Position{Line: i + 2, Column: 10})
	}

	schema := ast.NewObjectNode(fields, ast.Position{Line: 1, Column: 1})

	v := validator.NewSchemaValidator()
	result := v.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Large valid schema should pass validation. Errors: %v", result.Errors)
	}
}

// TestE2E_RealWorld_DeeplyNested tests deeply nested objects (5+ levels)
func TestE2E_RealWorld_DeeplyNested(t *testing.T) {
	// Build a deeply nested schema: level1.level2.level3.level4.level5.value
	level5 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"value": ast.NewTypeNode("UUID", ast.Position{Line: 6, Column: 15}),
		},
		ast.Position{Line: 5, Column: 12},
	)

	level4 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"level5": level5,
		},
		ast.Position{Line: 4, Column: 12},
	)

	level3 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"level4": level4,
		},
		ast.Position{Line: 3, Column: 12},
	)

	level2 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"level3": level3,
		},
		ast.Position{Line: 2, Column: 12},
	)

	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"level2": level2,
		},
		ast.Position{Line: 1, Column: 1},
	)

	v := validator.NewSchemaValidator()
	result := v.ValidateAll(schema)

	if !result.Valid {
		t.Errorf("Deeply nested valid schema should pass validation. Errors: %v", result.Errors)
	}
}

// TestE2E_RealWorld_ComplexErrors tests complex validation errors
func TestE2E_RealWorld_ComplexErrors(t *testing.T) {
	// Schema with errors at multiple nesting levels and different types
	input := `{
  "user": {
    "id": UIID,
    "profile": {
      "country": CountryCode,
      "age": Integer(1, 100, 200)
    },
    "settings": {
      "theme": Enum("light", "dark", "auto"),
      "data": ArrayOf("String")
    }
  },
  "admin": {
    "permissions": [
      CustomType
    ]
  }
}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := shape.ValidateAll(node, input)

	if result.Valid {
		t.Error("Complex schema with multiple errors should be invalid")
	}

	// Should collect all errors from different levels
	if result.ErrorCount() < 4 {
		t.Errorf("Expected at least 4 errors, got %d", result.ErrorCount())
	}

	// Verify errors are properly tracked with JSONPath
	for _, err := range result.Errors {
		if err.Path == "" {
			t.Error("All errors should have JSONPath")
		}
		t.Logf("Error at %s: %s", err.Path, err.Message)
	}
}

// TestE2E_Concurrency_ThreadSafety tests concurrent validation
func TestE2E_Concurrency_ThreadSafety(t *testing.T) {
	schemas := []string{
		`{"id": UUID}`,
		`{"email": Email}`,
		`{"age": Integer(18, 120)}`,
		`{"name": String(1, 50)}`,
		`{"status": Enum("active", "inactive")}`,
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(schemas)*10)

	// Run validation concurrently multiple times
	for i := 0; i < 10; i++ {
		for _, schemaStr := range schemas {
			wg.Add(1)
			go func(input string) {
				defer wg.Done()

				node, err := shape.Parse(parser.FormatJSONV, input)
				if err != nil {
					errors <- fmt.Errorf("parse error: %v", err)
					return
				}

				result := shape.ValidateAll(node, input)
				if !result.Valid {
					errors <- fmt.Errorf("validation failed: %v", result.Errors)
				}
			}(schemaStr)
		}
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent validation error: %v", err)
	}
}

// TestE2E_BackwardCompatibility_SimpleValidate tests that old Validate() API still works
func TestE2E_BackwardCompatibility_SimpleValidate(t *testing.T) {
	input := `{"id": UUID, "email": Email}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Old API: simple Validate()
	err = shape.Validate(node)
	if err != nil {
		t.Errorf("Backward compatibility: Validate() failed: %v", err)
	}

	// Test with invalid schema
	invalidInput := `{"country": CountryCode}`
	invalidNode, err := shape.Parse(parser.FormatJSONV, invalidInput)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	err = shape.Validate(invalidNode)
	if err == nil {
		t.Error("Backward compatibility: Validate() should fail for invalid schema")
	}
}

// TestE2E_BackwardCompatibility_ExistingTests ensures existing tests still pass
func TestE2E_BackwardCompatibility_ExistingTests(t *testing.T) {
	// This test verifies that changes to ValidateAll don't break existing functionality
	// by running some basic scenarios that existing code might rely on

	v := validator.NewSchemaValidator()

	// Test 1: Basic type validation
	schema1 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id": ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result := v.ValidateAll(schema1)
	if !result.Valid {
		t.Error("Basic type validation should work")
	}

	// Test 2: Custom type registration
	v.RegisterType("CustomType", validator.TypeDescriptor{
		Name:        "CustomType",
		Description: "Test custom type",
	})

	schema2 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"custom": ast.NewTypeNode("CustomType", ast.Position{Line: 1, Column: 1}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result = v.ValidateAll(schema2)
	if !result.Valid {
		t.Error("Custom type registration should work")
	}

	// Test 3: Function validation
	schema3 := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"age": ast.NewFunctionNode("Integer", []interface{}{int64(1), int64(100)}, ast.Position{Line: 1, Column: 1}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	result = v.ValidateAll(schema3)
	if !result.Valid {
		t.Error("Function validation should work")
	}
}

// buildCLI compiles the CLI tool and returns the path to the binary
func buildCLI(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "shape-validate")

	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/shape-validate")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v\nStderr: %s", err, stderr.String())
	}

	return binaryPath
}
