package validator

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

func TestValidator_ValidTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantErr  bool
	}{
		{"UUID type", "UUID", false},
		{"Email type", "Email", false},
		{"String type", "String", false},
		{"Integer type", "Integer", false},
		{"ISO-8601 type", "ISO-8601", false},
		{"URL type", "URL", false},
		{"unknown type", "UnknownType", true},
		{"lowercase type", "uuid", true},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ast.NewTypeNode(tt.typeName, ast.Position{Line: 1, Column: 1})
			err := v.Validate(node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidFunctions(t *testing.T) {
	tests := []struct {
		name    string
		fn      string
		args    []interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "String with valid range",
			fn:      "String",
			args:    []interface{}{int64(1), int64(100)},
			wantErr: false,
		},
		{
			name:    "String with unbounded range",
			fn:      "String",
			args:    []interface{}{int64(1), "+"},
			wantErr: false,
		},
		{
			name:    "String with invalid range (min > max)",
			fn:      "String",
			args:    []interface{}{int64(100), int64(1)},
			wantErr: true,
			errMsg:  "min (100) must be less than or equal to max (1)",
		},
		{
			name:    "Integer with valid range",
			fn:      "Integer",
			args:    []interface{}{int64(0), int64(120)},
			wantErr: false,
		},
		{
			name:    "Integer with single arg",
			fn:      "Integer",
			args:    []interface{}{int64(18)},
			wantErr: false,
		},
		{
			name:    "Enum with valid values",
			fn:      "Enum",
			args:    []interface{}{"active", "inactive", "pending"},
			wantErr: false,
		},
		{
			name:    "Enum with no args",
			fn:      "Enum",
			args:    []interface{}{},
			wantErr: true,
			errMsg:  "requires at least 1 arguments",
		},
		{
			name:    "Pattern with valid regex",
			fn:      "Pattern",
			args:    []interface{}{"^[a-z]+$"},
			wantErr: false,
		},
		{
			name:    "Pattern with non-string",
			fn:      "Pattern",
			args:    []interface{}{int64(123)},
			wantErr: true,
			errMsg:  "must be a string pattern",
		},
		{
			name:    "Unknown function",
			fn:      "UnknownFunc",
			args:    []interface{}{},
			wantErr: true,
			errMsg:  "unknown function",
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ast.NewFunctionNode(tt.fn, tt.args, ast.Position{Line: 1, Column: 1})
			err := v.Validate(node)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidator_ValidObject(t *testing.T) {
	tests := []struct {
		name    string
		props   map[string]ast.SchemaNode
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid object",
			props: map[string]ast.SchemaNode{
				"id":   ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1}),
				"name": ast.NewFunctionNode("String", []interface{}{int64(1), int64(100)}, ast.Position{Line: 2, Column: 1}),
			},
			wantErr: false,
		},
		{
			name: "object with unknown type",
			props: map[string]ast.SchemaNode{
				"id": ast.NewTypeNode("UnknownType", ast.Position{Line: 1, Column: 1}),
			},
			wantErr: true,
			errMsg:  "unknown type",
		},
		{
			name: "object with invalid function",
			props: map[string]ast.SchemaNode{
				"name": ast.NewFunctionNode("String", []interface{}{int64(100), int64(1)}, ast.Position{Line: 1, Column: 1}),
			},
			wantErr: true,
			errMsg:  "min (100) must be less than or equal to max (1)",
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ast.NewObjectNode(tt.props, ast.Position{Line: 1, Column: 1})
			err := v.Validate(node)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidator_ValidArray(t *testing.T) {
	tests := []struct {
		name    string
		elem    ast.SchemaNode
		wantErr bool
		errMsg  string
	}{
		{
			name:    "array with valid type",
			elem:    ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1}),
			wantErr: false,
		},
		{
			name:    "array with valid function",
			elem:    ast.NewFunctionNode("String", []interface{}{int64(1), int64(100)}, ast.Position{Line: 1, Column: 1}),
			wantErr: false,
		},
		{
			name:    "array with unknown type",
			elem:    ast.NewTypeNode("UnknownType", ast.Position{Line: 1, Column: 1}),
			wantErr: true,
			errMsg:  "unknown type",
		},
		{
			name:    "array with invalid function",
			elem:    ast.NewFunctionNode("String", []interface{}{int64(100), int64(1)}, ast.Position{Line: 1, Column: 1}),
			wantErr: true,
			errMsg:  "min (100) must be less than or equal to max (1)",
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ast.NewArrayNode(tt.elem, ast.Position{Line: 1, Column: 1})
			err := v.Validate(node)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidator_ComplexSchema(t *testing.T) {
	// Build complex schema: user object with nested properties and arrays
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":       ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1}),
			"username": ast.NewFunctionNode("String", []interface{}{int64(3), int64(20)}, ast.Position{Line: 2, Column: 1}),
			"email":    ast.NewTypeNode("Email", ast.Position{Line: 3, Column: 1}),
			"age":      ast.NewFunctionNode("Integer", []interface{}{int64(18), int64(120)}, ast.Position{Line: 4, Column: 1}),
			"status":   ast.NewFunctionNode("Enum", []interface{}{"active", "inactive", "banned"}, ast.Position{Line: 5, Column: 1}),
			"tags":     ast.NewArrayNode(ast.NewFunctionNode("String", []interface{}{int64(1), int64(30)}, ast.Position{Line: 6, Column: 1}), ast.Position{Line: 6, Column: 1}),
		},
		ast.Position{Line: 1, Column: 1},
	)

	v := NewValidator()
	err := v.Validate(schema)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_LiteralAlwaysValid(t *testing.T) {
	tests := []interface{}{
		"string literal",
		int64(42),
		3.14,
		true,
		false,
		nil,
	}

	v := NewValidator()
	for _, val := range tests {
		node := ast.NewLiteralNode(val, ast.Position{Line: 1, Column: 1})
		err := v.Validate(node)
		if err != nil {
			t.Errorf("Validate() literal %v error = %v, want nil", val, err)
		}
	}
}

func TestValidator_RegisterType(t *testing.T) {
	v := NewValidator()

	// Custom type should initially fail validation
	node := ast.NewTypeNode("SSN", ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err == nil {
		t.Error("Expected error for unregistered type 'SSN', got nil")
	}

	// Register the custom type
	v.RegisterType("SSN")

	// Now it should pass validation
	err = v.Validate(node)
	if err != nil {
		t.Errorf("After RegisterType('SSN'), expected nil error, got %v", err)
	}

	// Check if type is registered
	if !v.IsTypeRegistered("SSN") {
		t.Error("IsTypeRegistered('SSN') = false, want true")
	}
}

func TestValidator_RegisterType_Chaining(t *testing.T) {
	v := NewValidator()

	// Test method chaining
	v.RegisterType("SSN").RegisterType("PhoneNumber").RegisterType("ZipCode")

	// All should be registered
	types := []string{"SSN", "PhoneNumber", "ZipCode"}
	for _, typeName := range types {
		if !v.IsTypeRegistered(typeName) {
			t.Errorf("IsTypeRegistered(%q) = false, want true", typeName)
		}

		node := ast.NewTypeNode(typeName, ast.Position{Line: 1, Column: 1})
		err := v.Validate(node)
		if err != nil {
			t.Errorf("Validate(%q) error = %v, want nil", typeName, err)
		}
	}
}

func TestValidator_RegisterFunction(t *testing.T) {
	v := NewValidator()

	// Custom function should initially fail validation
	node := ast.NewFunctionNode("CreditCard", []interface{}{}, ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err == nil {
		t.Error("Expected error for unregistered function 'CreditCard', got nil")
	}

	// Register the custom function
	v.RegisterFunction("CreditCard", FunctionRule{
		MinArgs: 0,
		MaxArgs: 1,
	})

	// Now it should pass validation
	err = v.Validate(node)
	if err != nil {
		t.Errorf("After RegisterFunction('CreditCard'), expected nil error, got %v", err)
	}

	// Check if function is registered
	if !v.IsFunctionRegistered("CreditCard") {
		t.Error("IsFunctionRegistered('CreditCard') = false, want true")
	}
}

func TestValidator_RegisterFunction_WithValidator(t *testing.T) {
	v := NewValidator()

	// Register function with custom argument validator
	v.RegisterFunction("Even", FunctionRule{
		MinArgs: 0,
		MaxArgs: 0,
	})

	// Valid: no arguments
	node := ast.NewFunctionNode("Even", []interface{}{}, ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}

	// Invalid: too many arguments
	node = ast.NewFunctionNode("Even", []interface{}{int64(1)}, ast.Position{Line: 1, Column: 1})
	err = v.Validate(node)
	if err == nil {
		t.Error("Expected error for too many arguments, got nil")
	}
}

func TestValidator_RegisterFunction_Chaining(t *testing.T) {
	v := NewValidator()

	// Test method chaining
	v.RegisterFunction("SSN", FunctionRule{MinArgs: 0, MaxArgs: 0}).
		RegisterFunction("CreditCard", FunctionRule{MinArgs: 0, MaxArgs: 1}).
		RegisterFunction("PhoneNumber", FunctionRule{MinArgs: 1, MaxArgs: 1})

	// All should be registered
	if !v.IsFunctionRegistered("SSN") {
		t.Error("IsFunctionRegistered('SSN') = false, want true")
	}
	if !v.IsFunctionRegistered("CreditCard") {
		t.Error("IsFunctionRegistered('CreditCard') = false, want true")
	}
	if !v.IsFunctionRegistered("PhoneNumber") {
		t.Error("IsFunctionRegistered('PhoneNumber') = false, want true")
	}
}

func TestValidator_UnregisterType(t *testing.T) {
	v := NewValidator()

	// Register and then unregister a custom type
	v.RegisterType("CustomType")
	if !v.IsTypeRegistered("CustomType") {
		t.Error("CustomType should be registered")
	}

	v.UnregisterType("CustomType")
	if v.IsTypeRegistered("CustomType") {
		t.Error("CustomType should be unregistered")
	}

	// Should fail validation after unregistering
	node := ast.NewTypeNode("CustomType", ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err == nil {
		t.Error("Expected error for unregistered type, got nil")
	}
}

func TestValidator_UnregisterType_BuiltIn(t *testing.T) {
	v := NewValidator()

	// Attempting to unregister a built-in type should be silently ignored
	v.UnregisterType("UUID")
	if !v.IsTypeRegistered("UUID") {
		t.Error("Built-in type UUID should still be registered")
	}

	// Should still pass validation
	node := ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err != nil {
		t.Errorf("UUID should still be valid, got error: %v", err)
	}
}

func TestValidator_UnregisterFunction(t *testing.T) {
	v := NewValidator()

	// Register and then unregister a custom function
	v.RegisterFunction("CustomFunc", FunctionRule{MinArgs: 0, MaxArgs: 0})
	if !v.IsFunctionRegistered("CustomFunc") {
		t.Error("CustomFunc should be registered")
	}

	v.UnregisterFunction("CustomFunc")
	if v.IsFunctionRegistered("CustomFunc") {
		t.Error("CustomFunc should be unregistered")
	}

	// Should fail validation after unregistering
	node := ast.NewFunctionNode("CustomFunc", []interface{}{}, ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err == nil {
		t.Error("Expected error for unregistered function, got nil")
	}
}

func TestValidator_UnregisterFunction_BuiltIn(t *testing.T) {
	v := NewValidator()

	// Attempting to unregister a built-in function should be silently ignored
	v.UnregisterFunction("String")
	if !v.IsFunctionRegistered("String") {
		t.Error("Built-in function String should still be registered")
	}

	// Should still pass validation
	node := ast.NewFunctionNode("String", []interface{}{int64(1), int64(100)}, ast.Position{Line: 1, Column: 1})
	err := v.Validate(node)
	if err != nil {
		t.Errorf("String function should still be valid, got error: %v", err)
	}
}

func TestValidator_CustomTypesAndFunctions(t *testing.T) {
	v := NewValidator()

	// Register custom types and functions
	v.RegisterType("SSN").
		RegisterType("CreditCard").
		RegisterFunction("ValidateSSN", FunctionRule{MinArgs: 0, MaxArgs: 1}).
		RegisterFunction("ValidateCreditCard", FunctionRule{MinArgs: 0, MaxArgs: 0})

	// Create a schema using custom types and functions
	schema := ast.NewObjectNode(
		map[string]ast.SchemaNode{
			"id":            ast.NewTypeNode("UUID", ast.Position{Line: 1, Column: 1}),                                    // Built-in
			"ssn":           ast.NewTypeNode("SSN", ast.Position{Line: 2, Column: 1}),                                     // Custom
			"card":          ast.NewTypeNode("CreditCard", ast.Position{Line: 3, Column: 1}),                              // Custom
			"validatedSSN":  ast.NewFunctionNode("ValidateSSN", []interface{}{}, ast.Position{Line: 4, Column: 1}),        // Custom
			"validatedCard": ast.NewFunctionNode("ValidateCreditCard", []interface{}{}, ast.Position{Line: 5, Column: 1}), // Custom
		},
		ast.Position{Line: 1, Column: 1},
	)

	err := v.Validate(schema)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}
