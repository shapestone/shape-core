package textv

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

func TestTEXTVParser_SimpleProperties(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "single property with UUID",
			input: `id: UUID`,
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
					t.Errorf("type name = %q, want %q", typeNode.TypeName(), "UUID")
				}
			},
		},
		{
			name: "single property with function",
			input: `username: String(3, 20)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				username, ok := obj.GetProperty("username")
				if !ok {
					t.Fatal("property 'username' not found")
				}
				fn, ok := username.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", username)
				}
				if fn.Name() != "String" {
					t.Errorf("function name = %q, want %q", fn.Name(), "String")
				}
			},
		},
		{
			name: "multiple properties",
			input: `id: UUID
name: String
email: Email`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if len(obj.Properties()) != 3 {
					t.Errorf("properties count = %d, want 3", len(obj.Properties()))
				}
				expectedProps := map[string]string{
					"id":    "UUID",
					"name":  "String",
					"email": "Email",
				}
				for prop, expectedType := range expectedProps {
					val, ok := obj.GetProperty(prop)
					if !ok {
						t.Errorf("property %q not found", prop)
						continue
					}
					typeNode, ok := val.(*ast.TypeNode)
					if !ok {
						t.Errorf("expected TypeNode for %q, got %T", prop, val)
						continue
					}
					if typeNode.TypeName() != expectedType {
						t.Errorf("type for %q = %q, want %q", prop, typeNode.TypeName(), expectedType)
					}
				}
			},
		},
		{
			name: "with comments",
			input: `# User schema
id: UUID
name: String # User's full name`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if len(obj.Properties()) != 2 {
					t.Errorf("properties count = %d, want 2", len(obj.Properties()))
				}
			},
		},
		{
			name: "boolean literal",
			input: `active: true`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				active, ok := obj.GetProperty("active")
				if !ok {
					t.Fatal("property 'active' not found")
				}
				lit, ok := active.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode, got %T", active)
				}
				if lit.Value() != true {
					t.Errorf("value = %v, want true", lit.Value())
				}
			},
		},
		{
			name: "number literal",
			input: `count: 42`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				count, ok := obj.GetProperty("count")
				if !ok {
					t.Fatal("property 'count' not found")
				}
				lit, ok := count.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode, got %T", count)
				}
				if lit.Value() != int64(42) {
					t.Errorf("value = %v, want 42", lit.Value())
				}
			},
		},
		{
			name: "null literal",
			input: `optional: null`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				optional, ok := obj.GetProperty("optional")
				if !ok {
					t.Fatal("property 'optional' not found")
				}
				lit, ok := optional.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode, got %T", optional)
				}
				if lit.Value() != nil {
					t.Errorf("value = %v, want nil", lit.Value())
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

func TestTEXTVParser_NestedProperties(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple nested property",
			input: `user.id: UUID
user.name: String`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				user, ok := obj.GetProperty("user")
				if !ok {
					t.Fatal("property 'user' not found")
				}

				userObj, ok := user.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for 'user', got %T", user)
				}

				if _, ok := userObj.GetProperty("id"); !ok {
					t.Error("property 'user.id' not found")
				}
				if _, ok := userObj.GetProperty("name"); !ok {
					t.Error("property 'user.name' not found")
				}
			},
		},
		{
			name: "deeply nested property",
			input: `user.profile.name: String(1, 100)
user.profile.avatar: URL`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				user, ok := obj.GetProperty("user")
				if !ok {
					t.Fatal("property 'user' not found")
				}

				userObj, ok := user.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for 'user', got %T", user)
				}

				profile, ok := userObj.GetProperty("profile")
				if !ok {
					t.Fatal("property 'user.profile' not found")
				}

				profileObj, ok := profile.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for 'profile', got %T", profile)
				}

				if _, ok := profileObj.GetProperty("name"); !ok {
					t.Error("property 'profile.name' not found")
				}
				if _, ok := profileObj.GetProperty("avatar"); !ok {
					t.Error("property 'profile.avatar' not found")
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

func TestTEXTVParser_Arrays(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple array",
			input: `tags[]: String`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				tags, ok := obj.GetProperty("tags")
				if !ok {
					t.Fatal("property 'tags' not found")
				}

				arr, ok := tags.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode for 'tags', got %T", tags)
				}

				elem := arr.ElementSchema()
				if elem.Type() != ast.NodeTypeType {
					t.Errorf("element type = %v, want %v", elem.Type(), ast.NodeTypeType)
				}
			},
		},
		{
			name: "nested array",
			input: `user.roles[]: String(1, 50)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				user, ok := obj.GetProperty("user")
				if !ok {
					t.Fatal("property 'user' not found")
				}

				userObj, ok := user.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for 'user', got %T", user)
				}

				roles, ok := userObj.GetProperty("roles")
				if !ok {
					t.Fatal("property 'user.roles' not found")
				}

				arr, ok := roles.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode for 'roles', got %T", roles)
				}

				elem := arr.ElementSchema()
				if elem.Type() != ast.NodeTypeFunction {
					t.Errorf("element type = %v, want %v", elem.Type(), ast.NodeTypeFunction)
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

func TestTEXTVParser_Functions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "function with numeric args",
			input: `age: Integer(18, 120)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				age, ok := obj.GetProperty("age")
				if !ok {
					t.Fatal("property 'age' not found")
				}
				fn, ok := age.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", age)
				}
				if fn.Name() != "Integer" {
					t.Errorf("function name = %q, want %q", fn.Name(), "Integer")
				}
			},
		},
		{
			name: "function with unbounded arg",
			input: `description: String(1, +)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				desc, ok := obj.GetProperty("description")
				if !ok {
					t.Fatal("property 'description' not found")
				}
				fn, ok := desc.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", desc)
				}
				if fn.Name() != "String" {
					t.Errorf("function name = %q, want %q", fn.Name(), "String")
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

func TestTEXTVParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "missing colon",
			input:   "id UUID",
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   "id:",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTEXTVParser_ComplexExample(t *testing.T) {
	input := `# User schema
user.id: UUID
user.username: String(3, 20)
user.email: Email
user.age: Integer(18, 120)
user.active: true
user.roles[]: String(1, 50)
metadata.created: ISO-8601
metadata.tags[]: String(1, 30)`

	p := NewParser()
	node, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	root, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode, got %T", node)
	}

	// Check user
	user, ok := root.GetProperty("user")
	if !ok {
		t.Fatal("property 'user' not found")
	}

	userObj, ok := user.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode for 'user', got %T", user)
	}

	// Verify user has expected properties
	expectedUserProps := []string{"id", "username", "email", "age", "active", "roles"}
	for _, prop := range expectedUserProps {
		if _, ok := userObj.GetProperty(prop); !ok {
			t.Errorf("user property %q not found", prop)
		}
	}

	// Check metadata
	metadata, ok := root.GetProperty("metadata")
	if !ok {
		t.Fatal("property 'metadata' not found")
	}

	metadataObj, ok := metadata.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode for 'metadata', got %T", metadata)
	}

	// Verify metadata has expected properties
	if len(metadataObj.Properties()) != 2 {
		t.Errorf("metadata has %d properties, want 2", len(metadataObj.Properties()))
	}
}

func TestTEXTVParser_Format(t *testing.T) {
	p := NewParser()
	if p.Format() != parser.FormatTEXTV {
		t.Errorf("Format() = %v, want %v", p.Format(), parser.FormatTEXTV)
	}
}
