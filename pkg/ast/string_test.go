package ast

import (
	"testing"
)

func TestObjectNode_String_Empty(t *testing.T) {
	obj := NewObjectNode(make(map[string]SchemaNode), Position{})
	expected := "{}"
	got := obj.String()
	if got != expected {
		t.Errorf("ObjectNode.String() = %q, want %q", got, expected)
	}
}

func TestObjectNode_String_WithProperties(t *testing.T) {
	props := map[string]SchemaNode{
		"name": NewTypeNode("String", Position{}),
		"age":  NewTypeNode("Integer", Position{}),
	}
	obj := NewObjectNode(props, Position{})
	got := obj.String()

	// Should have both properties
	if !contains(got, `"age"`) || !contains(got, `"name"`) {
		t.Errorf("ObjectNode.String() = %q, should contain both properties", got)
	}
	if !contains(got, "String") || !contains(got, "Integer") {
		t.Errorf("ObjectNode.String() = %q, should contain type names", got)
	}
}

func TestLiteralNode_String_Types(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "null",
			value:    nil,
			expected: "null",
		},
		{
			name:     "string",
			value:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
		},
		{
			name:     "int64",
			value:    int64(42),
			expected: "42",
		},
		{
			name:     "float64",
			value:    float64(3.14),
			expected: "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewLiteralNode(tt.value, Position{})
			got := node.String()
			if got != tt.expected {
				t.Errorf("LiteralNode.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFunctionNode_String(t *testing.T) {
	tests := []struct {
		name      string
		funcName  string
		args      []interface{}
		wantStart string
	}{
		{
			name:      "no arguments",
			funcName:  "UUID",
			args:      []interface{}{},
			wantStart: "UUID(",
		},
		{
			name:      "with integer arguments",
			funcName:  "String",
			args:      []interface{}{int64(1), int64(100)},
			wantStart: "String(1",
		},
		{
			name:      "with float arguments",
			funcName:  "Number",
			args:      []interface{}{float64(0.0), float64(1.0)},
			wantStart: "Number(0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewFunctionNode(tt.funcName, tt.args, Position{})
			got := node.String()
			if !contains(got, tt.wantStart) {
				t.Errorf("FunctionNode.String() = %q, want to contain %q", got, tt.wantStart)
			}
			if !contains(got, tt.funcName) {
				t.Errorf("FunctionNode.String() = %q, want to contain function name %q", got, tt.funcName)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
