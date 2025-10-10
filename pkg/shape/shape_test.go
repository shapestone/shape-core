package shape

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		format  parser.Format
		input   string
		wantErr bool
		check   func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:    "JSONV simple object",
			format:  parser.FormatJSONV,
			input:   `{"id": UUID}`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
			},
		},
		{
			name:    "JSONV with function",
			format:  parser.FormatJSONV,
			input:   `{"name": String(1, 100)}`,
			wantErr: false,
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
			name:    "JSONV array",
			format:  parser.FormatJSONV,
			input:   `[UUID]`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				arr, ok := node.(*ast.ArrayNode)
				if !ok {
					t.Fatalf("expected ArrayNode, got %T", node)
				}
				elem := arr.ElementSchema()
				if elem.Type() != ast.NodeTypeType {
					t.Errorf("elem.Type() = %v, want %v", elem.Type(), ast.NodeTypeType)
				}
			},
		},
		{
			name:    "invalid JSONV",
			format:  parser.FormatJSONV,
			input:   `{"id": }`,
			wantErr: true,
		},
		{
			name:    "unsupported format",
			format:  parser.FormatUnknown,
			input:   `user:\n  id: UUID`,
			wantErr: true,
		},
		{
			name:    "PropsV simple",
			format:  parser.FormatPropsV,
			input:   `id=UUID`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
			},
		},
		{
			name:    "XMLV simple",
			format:  parser.FormatXMLV,
			input:   `<user><id>UUID</id></user>`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
			},
		},
		{
			name:   "CSVV simple",
			format: parser.FormatCSVV,
			input: `id,name
UUID,String`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
				if _, ok := obj.GetProperty("name"); !ok {
					t.Error("property 'name' not found")
				}
			},
		},
		{
			name:   "CSVV with function",
			format: parser.FormatCSVV,
			input: `username,age
String(3,20),Integer(18,120)`,
			wantErr: false,
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
					t.Errorf("fn.Name() = %q, want %q", fn.Name(), "String")
				}
			},
		},
		{
			name:   "YAMLV simple",
			format: parser.FormatYAMLV,
			input: `id: UUID
name: String`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
				if _, ok := obj.GetProperty("name"); !ok {
					t.Error("property 'name' not found")
				}
			},
		},
		{
			name:   "YAMLV with nested",
			format: parser.FormatYAMLV,
			input: `user:
  id: UUID
  name: String(1, 100)`,
			wantErr: false,
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
			},
		},
		{
			name:   "TEXTV simple",
			format: parser.FormatTEXTV,
			input: `id: UUID
name: String`,
			wantErr: false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("id"); !ok {
					t.Error("property 'id' not found")
				}
				if _, ok := obj.GetProperty("name"); !ok {
					t.Error("property 'name' not found")
				}
			},
		},
		{
			name:   "TEXTV with nested",
			format: parser.FormatTEXTV,
			input: `user.id: UUID
user.name: String(1, 100)`,
			wantErr: false,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Parse(tt.format, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, node)
			}
		})
	}
}

func TestParseAuto(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedFormat parser.Format
		wantErr        bool
		check          func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:           "detect JSONV object",
			input:          `{"id": UUID}`,
			expectedFormat: parser.FormatJSONV,
			wantErr:        false,
			check: func(t *testing.T, node ast.SchemaNode) {
				if node.Type() != ast.NodeTypeObject {
					t.Errorf("node.Type() = %v, want %v", node.Type(), ast.NodeTypeObject)
				}
			},
		},
		{
			name:           "detect JSONV array",
			input:          `[String(1, 30)]`,
			expectedFormat: parser.FormatJSONV,
			wantErr:        false,
			check: func(t *testing.T, node ast.SchemaNode) {
				if node.Type() != ast.NodeTypeArray {
					t.Errorf("node.Type() = %v, want %v", node.Type(), ast.NodeTypeArray)
				}
			},
		},
		{
			name:           "detect JSONV with whitespace",
			input:          "  \n  {\n  \"id\": UUID\n}",
			expectedFormat: parser.FormatJSONV,
			wantErr:        false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, format, err := ParseAuto(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAuto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if format != tt.expectedFormat {
					t.Errorf("format = %v, want %v", format, tt.expectedFormat)
				}
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		// Should not panic
		node := MustParse(parser.FormatJSONV, `{"id": UUID}`)
		if node == nil {
			t.Error("MustParse() returned nil")
		}
	})

	t.Run("invalid input panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParse() did not panic on invalid input")
			}
		}()

		MustParse(parser.FormatJSONV, `{"id": }`)
	})
}

func TestParse_ComplexExample(t *testing.T) {
	input := `{
		"user": {
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"active": true,
			"roles": [String(1, 50)]
		},
		"metadata": {
			"created": ISO-8601,
			"tags": [String(1, 30)]
		}
	}`

	node, err := Parse(parser.FormatJSONV, input)
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
