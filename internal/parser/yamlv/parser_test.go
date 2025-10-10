package yamlv

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

func TestYAMLVParser_SimpleObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple type identifiers",
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

				// Check each property is a TypeNode
				for propName, expectedType := range map[string]string{
					"id":    "UUID",
					"name":  "String",
					"email": "Email",
				} {
					prop, ok := obj.GetProperty(propName)
					if !ok {
						t.Errorf("property %q not found", propName)
						continue
					}
					typeNode, ok := prop.(*ast.TypeNode)
					if !ok {
						t.Errorf("expected TypeNode for %q, got %T", propName, prop)
						continue
					}
					if typeNode.TypeName() != expectedType {
						t.Errorf("type for %q = %q, want %q", propName, typeNode.TypeName(), expectedType)
					}
				}
			},
		},
		{
			name: "with functions",
			input: `username: String(3, 20)
age: Integer(18, 120)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				// Check username
				username, ok := obj.GetProperty("username")
				if !ok {
					t.Fatal("property 'username' not found")
				}
				fn, ok := username.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode for 'username', got %T", username)
				}
				if fn.Name() != "String" {
					t.Errorf("username function name = %q, want %q", fn.Name(), "String")
				}

				// Check age
				age, ok := obj.GetProperty("age")
				if !ok {
					t.Fatal("property 'age' not found")
				}
				fn2, ok := age.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode for 'age', got %T", age)
				}
				if fn2.Name() != "Integer" {
					t.Errorf("age function name = %q, want %q", fn2.Name(), "Integer")
				}
			},
		},
		{
			name: "with literals",
			input: `active: true
count: 42
rate: 3.14`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				// Check boolean
				active, ok := obj.GetProperty("active")
				if !ok {
					t.Fatal("property 'active' not found")
				}
				lit, ok := active.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode for 'active', got %T", active)
				}
				if lit.Value() != true {
					t.Errorf("active value = %v, want true", lit.Value())
				}

				// Check integer
				count, ok := obj.GetProperty("count")
				if !ok {
					t.Fatal("property 'count' not found")
				}
				lit2, ok := count.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode for 'count', got %T", count)
				}
				// YAML unmarshals to int which becomes int64
				if lit2.Value() != int64(42) {
					t.Errorf("count value = %v, want 42", lit2.Value())
				}

				// Check float
				rate, ok := obj.GetProperty("rate")
				if !ok {
					t.Fatal("property 'rate' not found")
				}
				lit3, ok := rate.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode for 'rate', got %T", rate)
				}
				if lit3.Value() != 3.14 {
					t.Errorf("rate value = %v, want 3.14", lit3.Value())
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

func TestYAMLVParser_NestedObjects(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple nested object",
			input: `user:
  id: UUID
  name: String`,
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
			name: "deeply nested object",
			input: `user:
  profile:
    name: String(1, 100)
    avatar: URL`,
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

func TestYAMLVParser_Arrays(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "array with type",
			input: `tags:
  - String`,
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
				typeNode, ok := elem.(*ast.TypeNode)
				if !ok {
					t.Fatalf("expected TypeNode for array element, got %T", elem)
				}
				if typeNode.TypeName() != "String" {
					t.Errorf("element type = %q, want %q", typeNode.TypeName(), "String")
				}
			},
		},
		{
			name: "array with function",
			input: `scores:
  - Integer(0, 100)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}

				scores, ok := obj.GetProperty("scores")
				if !ok {
					t.Fatal("property 'scores' not found")
				}

				arr, ok := scores.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode for 'scores', got %T", scores)
				}

				elem := arr.ElementSchema()
				fn, ok := elem.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode for array element, got %T", elem)
				}
				if fn.Name() != "Integer" {
					t.Errorf("function name = %q, want %q", fn.Name(), "Integer")
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

func TestYAMLVParser_ComplexExample(t *testing.T) {
	input := `user:
  id: UUID
  username: String(3, 20)
  email: Email
  age: Integer(18, 120)
  active: true
  roles:
    - String(1, 50)
metadata:
  created: ISO-8601
  tags:
    - String(1, 30)`

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

	// Verify user properties
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

func TestYAMLVParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid YAML syntax",
			input:   "invalid:\n  - yaml\n  syntax:",
			wantErr: true,
		},
		{
			name:    "array with multiple elements",
			input:   "tags:\n  - String\n  - Integer",
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   "tags: []",
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

func TestYAMLVParser_Format(t *testing.T) {
	p := NewParser()
	if p.Format() != parser.FormatYAMLV {
		t.Errorf("Format() = %v, want %v", p.Format(), parser.FormatYAMLV)
	}
}
