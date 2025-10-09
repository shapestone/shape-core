package jsonv

import (
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

func TestParser_ParseLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "string literal",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
		{
			name:     "integer",
			input:    `42`,
			expected: int64(42),
		},
		{
			name:     "negative integer",
			input:    `-123`,
			expected: int64(-123),
		},
		{
			name:     "float",
			input:    `3.14`,
			expected: float64(3.14),
		},
		{
			name:     "true",
			input:    `true`,
			expected: true,
		},
		{
			name:     "false",
			input:    `false`,
			expected: false,
		},
		{
			name:     "null",
			input:    `null`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			literal, ok := node.(*ast.LiteralNode)
			if !ok {
				t.Fatalf("expected LiteralNode, got %T", node)
			}

			if literal.Value() != tt.expected {
				t.Errorf("literal.Value() = %v, want %v", literal.Value(), tt.expected)
			}
		})
	}
}

func TestParser_ParseType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "UUID type",
			input:    `UUID`,
			expected: "UUID",
		},
		{
			name:     "Email type",
			input:    `Email`,
			expected: "Email",
		},
		{
			name:     "ISO-8601 type",
			input:    `ISO-8601`,
			expected: "ISO-8601",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			typeNode, ok := node.(*ast.TypeNode)
			if !ok {
				t.Fatalf("expected TypeNode, got %T", node)
			}

			if typeNode.TypeName() != tt.expected {
				t.Errorf("typeNode.TypeName() = %q, want %q", typeNode.TypeName(), tt.expected)
			}
		})
	}
}

func TestParser_ParseFunction(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectName string
		expectArgs []interface{}
	}{
		{
			name:       "Integer with range",
			input:      `Integer(1, 100)`,
			expectName: "Integer",
			expectArgs: []interface{}{int64(1), int64(100)},
		},
		{
			name:       "String with unbounded",
			input:      `String(5+)`,
			expectName: "String",
			expectArgs: []interface{}{int64(5), "+"},
		},
		{
			name:       "Enum with strings",
			input:      `Enum("M", "F", "O")`,
			expectName: "Enum",
			expectArgs: []interface{}{"M", "F", "O"},
		},
		{
			name:       "Function with no args",
			input:      `NoArgs()`,
			expectName: "NoArgs",
			expectArgs: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			fn, ok := node.(*ast.FunctionNode)
			if !ok {
				t.Fatalf("expected FunctionNode, got %T", node)
			}

			if fn.Name() != tt.expectName {
				t.Errorf("fn.Name() = %q, want %q", fn.Name(), tt.expectName)
			}

			args := fn.Arguments()
			if len(args) != len(tt.expectArgs) {
				t.Errorf("len(args) = %d, want %d", len(args), len(tt.expectArgs))
				return
			}

			for i, expected := range tt.expectArgs {
				if args[i] != expected {
					t.Errorf("args[%d] = %v, want %v", i, args[i], expected)
				}
			}
		})
	}
}

func TestParser_ParseObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:  "empty object",
			input: `{}`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if len(obj.Properties()) != 0 {
					t.Errorf("expected empty object, got %d properties", len(obj.Properties()))
				}
			},
		},
		{
			name:  "object with literal",
			input: `{"active": true}`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				active, ok := obj.GetProperty("active")
				if !ok {
					t.Fatal("property 'active' not found")
				}

				literal, ok := active.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode, got %T", active)
				}

				if literal.Value() != true {
					t.Errorf("literal.Value() = %v, want true", literal.Value())
				}
			},
		},
		{
			name:  "object with type",
			input: `{"id": UUID}`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				id, ok := obj.GetProperty("id")
				if !ok {
					t.Fatal("property 'id' not found")
				}

				typeNode, ok := id.(*ast.TypeNode)
				if !ok {
					t.Fatalf("expected TypeNode, got %T", id)
				}

				if typeNode.TypeName() != "UUID" {
					t.Errorf("typeNode.TypeName() = %q, want %q", typeNode.TypeName(), "UUID")
				}
			},
		},
		{
			name:  "object with function",
			input: `{"name": String(1, 100)}`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				name, ok := obj.GetProperty("name")
				if !ok {
					t.Fatal("property 'name' not found")
				}

				fn, ok := name.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", name)
				}

				if fn.Name() != "String" {
					t.Errorf("fn.Name() = %q, want %q", fn.Name(), "String")
				}
			},
		},
		{
			name: "object with multiple properties",
			input: `{
				"id": UUID,
				"name": String(1, 100),
				"active": true
			}`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				if len(obj.Properties()) != 3 {
					t.Errorf("expected 3 properties, got %d", len(obj.Properties()))
				}

				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
				if _, ok := obj.GetProperty("name"); !ok {
					t.Error("property 'name' not found")
				}
				if _, ok := obj.GetProperty("active"); !ok {
					t.Error("property 'active' not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			tt.check(t, node)
		})
	}
}

func TestParser_ParseArray(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:  "array with type",
			input: `[UUID]`,
			check: func(t *testing.T, node ast.SchemaNode) {
				arr, ok := node.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode, got %T", node)
				}

				elem := arr.ElementSchema()
				typeNode, ok := elem.(*ast.TypeNode)
				if !ok {
					t.Fatalf("expected TypeNode, got %T", elem)
				}

				if typeNode.TypeName() != "UUID" {
					t.Errorf("typeNode.TypeName() = %q, want %q", typeNode.TypeName(), "UUID")
				}
			},
		},
		{
			name:  "array with function",
			input: `[String(1, 30)]`,
			check: func(t *testing.T, node ast.SchemaNode) {
				arr, ok := node.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode, got %T", node)
				}

				elem := arr.ElementSchema()
				fn, ok := elem.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", elem)
				}

				if fn.Name() != "String" {
					t.Errorf("fn.Name() = %q, want %q", fn.Name(), "String")
				}
			},
		},
		{
			name:  "array with literal",
			input: `["fixed"]`,
			check: func(t *testing.T, node ast.SchemaNode) {
				arr, ok := node.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode, got %T", node)
				}

				elem := arr.ElementSchema()
				literal, ok := elem.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode, got %T", elem)
				}

				if literal.Value() != "fixed" {
					t.Errorf("literal.Value() = %v, want %q", literal.Value(), "fixed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			tt.check(t, node)
		})
	}
}

func TestParser_ParseNested(t *testing.T) {
	input := `{
		"user": {
			"id": UUID,
			"profile": {
				"name": String(1, 100),
				"email": Email
			}
		},
		"tags": [String(1, 30)]
	}`

	p := NewParser()
	node, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check root is object
	root, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode, got %T", node)
	}

	// Check user property
	user, ok := root.GetProperty("user")
	if !ok {
		t.Fatal("property 'user' not found")
	}

	userObj, ok := user.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode for 'user', got %T", user)
	}

	// Check nested profile
	profile, ok := userObj.GetProperty("profile")
	if !ok {
		t.Fatal("property 'profile' not found")
	}

	profileObj, ok := profile.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode for 'profile', got %T", profile)
	}

	if len(profileObj.Properties()) != 2 {
		t.Errorf("expected 2 properties in profile, got %d", len(profileObj.Properties()))
	}

	// Check tags array
	tags, ok := root.GetProperty("tags")
	if !ok {
		t.Fatal("property 'tags' not found")
	}

	tagsArr, ok := tags.(*ast.ArrayNode)
	if !ok {
		t.Fatalf("expected ArrayNode for 'tags', got %T", tags)
	}

	elem := tagsArr.ElementSchema()
	fn, ok := elem.(*ast.FunctionNode)
	if !ok {
		t.Fatalf("expected FunctionNode for array element, got %T", elem)
	}

	if fn.Name() != "String" {
		t.Errorf("fn.Name() = %q, want %q", fn.Name(), "String")
	}
}

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "unclosed object",
			input:   `{"id": UUID`,
			wantErr: "EOF",
		},
		{
			name:    "missing colon",
			input:   `{"id" UUID}`,
			wantErr: "expected Colon",
		},
		{
			name:    "missing value",
			input:   `{"id":}`,
			wantErr: "expected value",
		},
		{
			name:    "unclosed array",
			input:   `[UUID`,
			wantErr: "expected ArrayEnd",
		},
		{
			name:    "trailing comma in object",
			input:   `{"id": UUID,}`,
			wantErr: "expected String",
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Parse(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
