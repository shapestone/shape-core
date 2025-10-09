package csvv

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

func TestCSVVParser_SimpleSchema(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "single column UUID",
			input: `id
UUID`,
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
					t.Fatalf("expected TypeNode for 'id', got %T", id)
				}
				if typeNode.TypeName() != "UUID" {
					t.Errorf("type name = %q, want %q", typeNode.TypeName(), "UUID")
				}
			},
		},
		{
			name: "multiple columns",
			input: `id,name,email
UUID,String,Email`,
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
			name: "with function calls",
			input: `username,age
String(3,20),Integer(18,120)`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				// Check username function
				username, ok := obj.GetProperty("username")
				if !ok {
					t.Fatal("property 'username' not found")
				}
				fn, ok := username.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode for 'username', got %T", username)
				}
				if fn.Name() != "String" {
					t.Errorf("function name = %q, want %q", fn.Name(), "String")
				}
				// Check age function
				age, ok := obj.GetProperty("age")
				if !ok {
					t.Fatal("property 'age' not found")
				}
				fn2, ok := age.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode for 'age', got %T", age)
				}
				if fn2.Name() != "Integer" {
					t.Errorf("function name = %q, want %q", fn2.Name(), "Integer")
				}
			},
		},
		{
			name: "with literal values",
			input: `active,count,rate
true,42,3.14`,
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

func TestCSVVParser_QuotedValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "quoted header",
			input: `"user name","user id"
String,UUID`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("user name"); !ok {
					t.Error("property 'user name' not found")
				}
				if _, ok := obj.GetProperty("user id"); !ok {
					t.Error("property 'user id' not found")
				}
			},
		},
		{
			name: "quoted validation",
			input: `description
"String(1, 1000)"`,
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
					t.Fatalf("expected FunctionNode for 'description', got %T", desc)
				}
				if fn.Name() != "String" {
					t.Errorf("function name = %q, want %q", fn.Name(), "String")
				}
			},
		},
		{
			name: "escaped quotes in cell",
			input: `name
"String with ""quotes"""`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("name"); !ok {
					t.Error("property 'name' not found")
				}
			},
		},
		{
			name: "mixed quoted and unquoted",
			input: `id,"full name",age
UUID,"String(1, 100)",Integer`,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if len(obj.Properties()) != 3 {
					t.Errorf("properties count = %d, want 3", len(obj.Properties()))
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

func TestCSVVParser_WithComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "comment before header",
			input: `# User schema
id,name
UUID,String`,
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
			name: "multiple comments",
			input: `# User schema
# Generated from API spec
id,email
UUID,Email`,
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

func TestCSVVParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: "empty header row",
		},
		{
			name:    "only header no validation",
			input:   "id,name",
			wantErr: "validation row has 0 columns, header has 2",
		},
		{
			name: "mismatched column count - more in validation",
			input: `id,name
UUID,String,Email`,
			wantErr: "validation row has 3 columns, header has 2",
		},
		{
			name: "mismatched column count - less in validation",
			input: `id,name,email
UUID,String`,
			wantErr: "validation row has 2 columns, header has 3",
		},
		{
			name:    "only comments",
			input:   "# Just a comment",
			wantErr: "empty header row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Parse(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tt.wantErr != "" {
				errMsg := err.Error()
				if !contains(errMsg, tt.wantErr) {
					t.Errorf("error = %q, want substring %q", errMsg, tt.wantErr)
				}
			}
		})
	}
}

func TestCSVVParser_ComplexExample(t *testing.T) {
	input := `# User profile schema
id,username,email,age,active,bio
UUID,String(3,20),Email,Integer(18,120),true,String(0,500)`

	p := NewParser()
	node, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode, got %T", node)
	}

	expectedProps := []string{"id", "username", "email", "age", "active", "bio"}
	if len(obj.Properties()) != len(expectedProps) {
		t.Errorf("properties count = %d, want %d", len(obj.Properties()), len(expectedProps))
	}

	for _, prop := range expectedProps {
		if _, ok := obj.GetProperty(prop); !ok {
			t.Errorf("property %q not found", prop)
		}
	}

	// Verify specific types
	id, _ := obj.GetProperty("id")
	if _, ok := id.(*ast.TypeNode); !ok {
		t.Errorf("expected TypeNode for 'id', got %T", id)
	}

	username, _ := obj.GetProperty("username")
	if fn, ok := username.(*ast.FunctionNode); !ok {
		t.Errorf("expected FunctionNode for 'username', got %T", username)
	} else if fn.Name() != "String" {
		t.Errorf("username function name = %q, want %q", fn.Name(), "String")
	}

	active, _ := obj.GetProperty("active")
	if lit, ok := active.(*ast.LiteralNode); !ok {
		t.Errorf("expected LiteralNode for 'active', got %T", active)
	} else if lit.Value() != true {
		t.Errorf("active value = %v, want true", lit.Value())
	}
}

func TestCSVVParser_Format(t *testing.T) {
	p := NewParser()
	if p.Format() != parser.FormatCSVV {
		t.Errorf("Format() = %v, want %v", p.Format(), parser.FormatCSVV)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
