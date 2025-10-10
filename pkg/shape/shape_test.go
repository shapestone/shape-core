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
		// JSONV detection
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

		// XMLV detection
		{
			name:           "detect XMLV",
			input:          `<user><id>UUID</id></user>`,
			expectedFormat: parser.FormatXMLV,
			wantErr:        false,
			check: func(t *testing.T, node ast.SchemaNode) {
				if node.Type() != ast.NodeTypeObject {
					t.Errorf("node.Type() = %v, want %v", node.Type(), ast.NodeTypeObject)
				}
			},
		},

		// PropsV detection
		{
			name:           "detect PropsV simple",
			input:          `id=UUID`,
			expectedFormat: parser.FormatPropsV,
			wantErr:        false,
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
			name:           "detect PropsV with dot notation",
			input:          `user.id=UUID`,
			expectedFormat: parser.FormatPropsV,
			wantErr:        false,
		},

		// CSVV detection
		{
			name:           "detect CSVV",
			input:          "id,name,email\nUUID,String,Email",
			expectedFormat: parser.FormatCSVV,
			wantErr:        false,
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

		// YAMLV detection
		{
			name:           "detect YAMLV simple",
			input:          "id: UUID\nname: String",
			expectedFormat: parser.FormatYAMLV,
			wantErr:        false,
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
			name:           "detect YAMLV nested",
			input:          "user:\n  id: UUID\n  name: String",
			expectedFormat: parser.FormatYAMLV,
			wantErr:        false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("user"); !ok {
					t.Error("property 'user' not found")
				}
			},
		},

		// TEXTV detection
		{
			name:           "detect TEXTV with dot notation",
			input:          "user.id: UUID",
			expectedFormat: parser.FormatTEXTV,
			wantErr:        false,
			check: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode, got %T", node)
				}
				if _, ok := obj.GetProperty("user"); !ok {
					t.Error("property 'user' not found")
				}
			},
		},
		{
			name:           "detect TEXTV multiple properties",
			input:          "user.id: UUID\nuser.name: String(1, 100)",
			expectedFormat: parser.FormatTEXTV,
			wantErr:        false,
		},

		// Error cases
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		format  parser.Format
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple schema",
			format:  parser.FormatJSONV,
			input:   `{"id": UUID, "name": String}`,
			wantErr: false,
		},
		{
			name:    "valid function with args",
			format:  parser.FormatJSONV,
			input:   `{"name": String(1, 100)}`,
			wantErr: false,
		},
		{
			name:    "valid complex schema",
			format:  parser.FormatJSONV,
			input:   `{"user": {"id": UUID, "age": Integer(18, 120)}}`,
			wantErr: false,
		},
		{
			name:    "valid array",
			format:  parser.FormatJSONV,
			input:   `[UUID]`,
			wantErr: false,
		},
		{
			name:    "unknown type",
			format:  parser.FormatJSONV,
			input:   `{"id": UnknownType}`,
			wantErr: true,
			errMsg:  "unknown type",
		},
		{
			name:    "unknown function",
			format:  parser.FormatJSONV,
			input:   `{"name": UnknownFunc(1, 2)}`,
			wantErr: true,
			errMsg:  "unknown function",
		},
		{
			name:    "invalid function args - min > max",
			format:  parser.FormatJSONV,
			input:   `{"name": String(100, 1)}`,
			wantErr: true,
			errMsg:  "min (100) must be less than or equal to max (1)",
		},
		{
			name:    "invalid function args - too few",
			format:  parser.FormatJSONV,
			input:   `{"status": Enum()}`,
			wantErr: true,
			errMsg:  "requires at least 1 arguments",
		},
		{
			name:    "invalid nested property",
			format:  parser.FormatJSONV,
			input:   `{"user": {"id": InvalidType}}`,
			wantErr: true,
			errMsg:  "unknown type",
		},
		{
			name:    "invalid array element",
			format:  parser.FormatJSONV,
			input:   `[InvalidType]`,
			wantErr: true,
			errMsg:  "unknown type",
		},
		{
			name:    "valid YAMLV schema",
			format:  parser.FormatYAMLV,
			input:   "id: UUID\nname: String(1, 100)",
			wantErr: false,
		},
		{
			name:    "invalid YAMLV schema",
			format:  parser.FormatYAMLV,
			input:   "id: BadType",
			wantErr: true,
			errMsg:  "unknown type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Parse(tt.format, tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			err = Validate(node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidate_ComplexSchema(t *testing.T) {
	input := `{
		"user": {
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"status": Enum("active", "inactive", "banned"),
			"roles": [String(1, 50)]
		}
	}`

	node, err := Parse(parser.FormatJSONV, input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	err = Validate(node)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
