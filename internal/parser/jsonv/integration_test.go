package jsonv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_ValidFiles(t *testing.T) {
	validDir := filepath.Join("..", "..", "..", "internal", "testdata", "jsonv", "valid")

	entries, err := os.ReadDir(validDir)
	if err != nil {
		t.Fatalf("failed to read valid test directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no valid test files found")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonv") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			filePath := filepath.Join(validDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			parser := NewParser()
			node, err := parser.Parse(string(content))
			if err != nil {
				t.Errorf("failed to parse valid file %s: %v", entry.Name(), err)
				return
			}

			if node == nil {
				t.Errorf("parser returned nil node for valid file %s", entry.Name())
			}
		})
	}
}

func TestIntegration_InvalidFiles(t *testing.T) {
	invalidDir := filepath.Join("..", "..", "..", "internal", "testdata", "jsonv", "invalid")

	entries, err := os.ReadDir(invalidDir)
	if err != nil {
		t.Fatalf("failed to read invalid test directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no invalid test files found")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonv") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			filePath := filepath.Join(invalidDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			parser := NewParser()
			_, err = parser.Parse(string(content))
			if err == nil {
				t.Errorf("expected error for invalid file %s, but parsing succeeded", entry.Name())
			}
		})
	}
}

func TestIntegration_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		wantErrPart string
	}{
		{
			name:        "unclosed object",
			file:        "unclosed-object.jsonv",
			wantErrPart: "line",
		},
		{
			name:        "missing colon",
			file:        "missing-colon.jsonv",
			wantErrPart: "Colon",
		},
		{
			name:        "invalid function",
			file:        "invalid-function.jsonv",
			wantErrPart: "argument",
		},
		{
			name:        "trailing comma",
			file:        "trailing-comma.jsonv",
			wantErrPart: "String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join("..", "..", "..", "internal", "testdata", "jsonv", "invalid", tt.file)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			parser := NewParser()
			_, err = parser.Parse(string(content))
			if err == nil {
				t.Fatal("expected error but got nil")
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.wantErrPart) {
				t.Errorf("error message %q should contain %q", errMsg, tt.wantErrPart)
			}

			// Verify error includes position information
			if !strings.Contains(errMsg, "line") && !strings.Contains(errMsg, "column") {
				t.Errorf("error message should include position information: %q", errMsg)
			}
		})
	}
}
