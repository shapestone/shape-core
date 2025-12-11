package grammar

import (
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

func TestASTEqual_BothNil(t *testing.T) {
	if !ASTEqual(nil, nil) {
		t.Error("expected nil nodes to be equal")
	}
}

func TestASTEqual_OneNil(t *testing.T) {
	literal := ast.NewLiteralNode("test", ast.Position{})

	if ASTEqual(nil, literal) {
		t.Error("expected nil and non-nil to not be equal")
	}

	if ASTEqual(literal, nil) {
		t.Error("expected non-nil and nil to not be equal")
	}
}

func TestASTEqual_DifferentTypes(t *testing.T) {
	literal := ast.NewLiteralNode("test", ast.Position{})
	typeNode := ast.NewTypeNode("String", ast.Position{})

	if ASTEqual(literal, typeNode) {
		t.Error("expected different node types to not be equal")
	}
}

func TestASTEqual_LiteralNodes(t *testing.T) {
	tests := []struct {
		name     string
		a        *ast.LiteralNode
		b        *ast.LiteralNode
		expected bool
	}{
		{
			name:     "equal literals",
			a:        ast.NewLiteralNode("test", ast.Position{}),
			b:        ast.NewLiteralNode("test", ast.Position{}),
			expected: true,
		},
		{
			name:     "different literals",
			a:        ast.NewLiteralNode("test", ast.Position{}),
			b:        ast.NewLiteralNode("other", ast.Position{}),
			expected: false,
		},
		{
			name:     "different positions ignored",
			a:        ast.NewLiteralNode("test", ast.Position{Line: 1, Column: 1}),
			b:        ast.NewLiteralNode("test", ast.Position{Line: 2, Column: 5}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASTEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTEqual_TypeNodes(t *testing.T) {
	tests := []struct {
		name     string
		a        *ast.TypeNode
		b        *ast.TypeNode
		expected bool
	}{
		{
			name:     "equal types",
			a:        ast.NewTypeNode("String", ast.Position{}),
			b:        ast.NewTypeNode("String", ast.Position{}),
			expected: true,
		},
		{
			name:     "different types",
			a:        ast.NewTypeNode("String", ast.Position{}),
			b:        ast.NewTypeNode("Number", ast.Position{}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASTEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTEqual_FunctionNodes(t *testing.T) {
	tests := []struct {
		name     string
		a        *ast.FunctionNode
		b        *ast.FunctionNode
		expected bool
	}{
		{
			name:     "equal functions",
			a:        ast.NewFunctionNode("maxLength", []interface{}{"10"}, ast.Position{}),
			b:        ast.NewFunctionNode("maxLength", []interface{}{"10"}, ast.Position{}),
			expected: true,
		},
		{
			name:     "different function names",
			a:        ast.NewFunctionNode("maxLength", []interface{}{}, ast.Position{}),
			b:        ast.NewFunctionNode("minLength", []interface{}{}, ast.Position{}),
			expected: false,
		},
		{
			name:     "different argument counts",
			a:        ast.NewFunctionNode("func", []interface{}{"a"}, ast.Position{}),
			b:        ast.NewFunctionNode("func", []interface{}{"a", "b"}, ast.Position{}),
			expected: false,
		},
		{
			name:     "different argument values",
			a:        ast.NewFunctionNode("func", []interface{}{"a"}, ast.Position{}),
			b:        ast.NewFunctionNode("func", []interface{}{"b"}, ast.Position{}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASTEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTEqual_ObjectNodes(t *testing.T) {
	tests := []struct {
		name     string
		a        *ast.ObjectNode
		b        *ast.ObjectNode
		expected bool
	}{
		{
			name: "equal objects",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
				"age":  ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
				"age":  ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			expected: true,
		},
		{
			name: "different property counts",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
				"age":  ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			expected: false,
		},
		{
			name: "missing property",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"other": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			expected: false,
		},
		{
			name: "different property values",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASTEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTEqual_ArrayNodes(t *testing.T) {
	tests := []struct {
		name     string
		a        *ast.ArrayNode
		b        *ast.ArrayNode
		expected bool
	}{
		{
			name: "equal arrays",
			a: ast.NewArrayNode(
				ast.NewTypeNode("String", ast.Position{}),
				ast.Position{},
			),
			b: ast.NewArrayNode(
				ast.NewTypeNode("String", ast.Position{}),
				ast.Position{},
			),
			expected: true,
		},
		{
			name: "different element schemas",
			a: ast.NewArrayNode(
				ast.NewTypeNode("String", ast.Position{}),
				ast.Position{},
			),
			b: ast.NewArrayNode(
				ast.NewTypeNode("Number", ast.Position{}),
				ast.Position{},
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASTEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTDiff_BothNil(t *testing.T) {
	diff := ASTDiff(nil, nil)
	if diff != "" {
		t.Errorf("expected empty diff for nil nodes, got: %s", diff)
	}
}

func TestASTDiff_OneNil(t *testing.T) {
	literal := ast.NewLiteralNode("test", ast.Position{})

	diff := ASTDiff(nil, literal)
	if !stringContains(diff, "first AST is nil") {
		t.Errorf("expected 'first AST is nil', got: %s", diff)
	}

	diff = ASTDiff(literal, nil)
	if !stringContains(diff, "second AST is nil") {
		t.Errorf("expected 'second AST is nil', got: %s", diff)
	}
}

func TestASTDiff_DifferentTypes(t *testing.T) {
	literal := ast.NewLiteralNode("test", ast.Position{})
	typeNode := ast.NewTypeNode("String", ast.Position{})

	diff := ASTDiff(literal, typeNode)
	if !stringContains(diff, "node types differ") {
		t.Errorf("expected 'node types differ', got: %s", diff)
	}
}

func TestASTDiff_LiteralNodes(t *testing.T) {
	a := ast.NewLiteralNode("test", ast.Position{})
	b := ast.NewLiteralNode("other", ast.Position{})

	diff := ASTDiff(a, b)
	if !stringContains(diff, "literal values differ") {
		t.Errorf("expected 'literal values differ', got: %s", diff)
	}
}

func TestASTDiff_TypeNodes(t *testing.T) {
	a := ast.NewTypeNode("String", ast.Position{})
	b := ast.NewTypeNode("Number", ast.Position{})

	diff := ASTDiff(a, b)
	if !stringContains(diff, "type names differ") {
		t.Errorf("expected 'type names differ', got: %s", diff)
	}
	if !stringContains(diff, "String") || !stringContains(diff, "Number") {
		t.Errorf("expected type names in diff, got: %s", diff)
	}
}

func TestASTDiff_FunctionNodes(t *testing.T) {
	tests := []struct {
		name         string
		a            *ast.FunctionNode
		b            *ast.FunctionNode
		expectedText string
	}{
		{
			name:         "different function names",
			a:            ast.NewFunctionNode("maxLength", []interface{}{}, ast.Position{}),
			b:            ast.NewFunctionNode("minLength", []interface{}{}, ast.Position{}),
			expectedText: "function names differ",
		},
		{
			name:         "different argument counts",
			a:            ast.NewFunctionNode("func", []interface{}{"a"}, ast.Position{}),
			b:            ast.NewFunctionNode("func", []interface{}{}, ast.Position{}),
			expectedText: "argument counts differ",
		},
		{
			name:         "different argument values",
			a:            ast.NewFunctionNode("func", []interface{}{"a"}, ast.Position{}),
			b:            ast.NewFunctionNode("func", []interface{}{"b"}, ast.Position{}),
			expectedText: "arguments differ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ASTDiff(tt.a, tt.b)
			if !stringContains(diff, tt.expectedText) {
				t.Errorf("expected diff to contain '%s', got: %s", tt.expectedText, diff)
			}
		})
	}
}

func TestASTDiff_ObjectNodes(t *testing.T) {
	tests := []struct {
		name         string
		a            *ast.ObjectNode
		b            *ast.ObjectNode
		expectedText string
	}{
		{
			name: "different property counts",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
				"age":  ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			expectedText: "property counts differ",
		},
		{
			name: "missing property",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"other": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			expectedText: "missing in second object",
		},
		{
			name: "different property values",
			a: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("String", ast.Position{}),
			}, ast.Position{}),
			b: ast.NewObjectNode(map[string]ast.SchemaNode{
				"name": ast.NewTypeNode("Number", ast.Position{}),
			}, ast.Position{}),
			expectedText: "in property name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ASTDiff(tt.a, tt.b)
			if !stringContains(diff, tt.expectedText) {
				t.Errorf("expected diff to contain '%s', got: %s", tt.expectedText, diff)
			}
		})
	}
}

func TestASTDiff_ArrayNodes(t *testing.T) {
	a := ast.NewArrayNode(
		ast.NewTypeNode("String", ast.Position{}),
		ast.Position{},
	)
	b := ast.NewArrayNode(
		ast.NewTypeNode("Number", ast.Position{}),
		ast.Position{},
	)

	diff := ASTDiff(a, b)
	if !stringContains(diff, "array element schema") {
		t.Errorf("expected 'array element schema', got: %s", diff)
	}
}

func TestASTDiff_NestedStructures(t *testing.T) {
	// Test deeply nested object comparison
	a := ast.NewObjectNode(map[string]ast.SchemaNode{
		"user": ast.NewObjectNode(map[string]ast.SchemaNode{
			"name": ast.NewTypeNode("String", ast.Position{}),
		}, ast.Position{}),
	}, ast.Position{})

	b := ast.NewObjectNode(map[string]ast.SchemaNode{
		"user": ast.NewObjectNode(map[string]ast.SchemaNode{
			"name": ast.NewTypeNode("Number", ast.Position{}),
		}, ast.Position{}),
	}, ast.Position{})

	diff := ASTDiff(a, b)
	if diff == "" {
		t.Error("expected non-empty diff for nested difference")
	}
	if !stringContains(diff, "property user") {
		t.Errorf("expected diff to mention 'property user', got: %s", diff)
	}
}
