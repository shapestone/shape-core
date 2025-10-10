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

// TestYAMLVParser_ArgumentParsing tests function argument parsing comprehensively
func TestYAMLVParser_ArgumentParsing(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:  "multiple numeric arguments",
			input: "field: String(1, 100)",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 2 {
					t.Fatalf("expected 2 args, got %d", len(args))
				}
				if args[0] != int64(1) {
					t.Errorf("arg[0] = %v, want 1", args[0])
				}
				if args[1] != int64(100) {
					t.Errorf("arg[1] = %v, want 100", args[1])
				}
			},
		},
		{
			name:  "unbounded with plus separate",
			input: "field: Integer(18, +)",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 2 {
					t.Fatalf("expected 2 args, got %d", len(args))
				}
				if args[0] != int64(18) {
					t.Errorf("arg[0] = %v, want 18", args[0])
				}
				if args[1] != "+" {
					t.Errorf("arg[1] = %v, want +", args[1])
				}
			},
		},
		{
			name:  "unbounded with plus attached",
			input: "field: Integer(18+)",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 2 {
					t.Fatalf("expected 2 args, got %d", len(args))
				}
				if args[0] != int64(18) {
					t.Errorf("arg[0] = %v, want 18", args[0])
				}
				if args[1] != "+" {
					t.Errorf("arg[1] = %v, want +", args[1])
				}
			},
		},
		{
			name:  "quoted string arguments",
			input: `field: Enum("active", "inactive", "pending")`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 3 {
					t.Fatalf("expected 3 args, got %d", len(args))
				}
				if args[0] != "active" {
					t.Errorf("arg[0] = %v, want 'active'", args[0])
				}
				if args[1] != "inactive" {
					t.Errorf("arg[1] = %v, want 'inactive'", args[1])
				}
				if args[2] != "pending" {
					t.Errorf("arg[2] = %v, want 'pending'", args[2])
				}
			},
		},
		{
			name:  "string with comma inside",
			input: `field: Enum("a,b", "c")`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 2 {
					t.Fatalf("expected 2 args, got %d", len(args))
				}
				if args[0] != "a,b" {
					t.Errorf("arg[0] = %v, want 'a,b'", args[0])
				}
			},
		},
		{
			name:  "mixed type arguments",
			input: `field: Function(1, "test", true, null)`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 4 {
					t.Fatalf("expected 4 args, got %d", len(args))
				}
				if args[0] != int64(1) {
					t.Errorf("arg[0] = %v (type %T), want 1", args[0], args[0])
				}
				if args[1] != "test" {
					t.Errorf("arg[1] = %v, want 'test'", args[1])
				}
				if args[2] != true {
					t.Errorf("arg[2] = %v, want true", args[2])
				}
				if args[3] != nil {
					t.Errorf("arg[3] = %v, want nil", args[3])
				}
			},
		},
		{
			name:  "float arguments",
			input: "field: Function(1.5, 2.7)",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 2 {
					t.Fatalf("expected 2 args, got %d", len(args))
				}
				if args[0] != 1.5 {
					t.Errorf("arg[0] = %v, want 1.5", args[0])
				}
				if args[1] != 2.7 {
					t.Errorf("arg[1] = %v, want 2.7", args[1])
				}
			},
		},
		{
			name:  "empty args function",
			input: "field: UUID()",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 0 {
					t.Fatalf("expected 0 args, got %d", len(args))
				}
			},
		},
		{
			name:  "escaped quotes in string",
			input: `field: Pattern("quote \"inside\" string")`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 1 {
					t.Fatalf("expected 1 arg, got %d", len(args))
				}
				expected := `quote "inside" string`
				if args[0] != expected {
					t.Errorf("arg[0] = %q, want %q", args[0], expected)
				}
			},
		},
		{
			name:  "backslash escaping",
			input: `field: Pattern("path\\to\\file")`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 1 {
					t.Fatalf("expected 1 arg, got %d", len(args))
				}
				expected := `path\to\file`
				if args[0] != expected {
					t.Errorf("arg[0] = %q, want %q", args[0], expected)
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
			tt.checkFunc(t, node)
		})
	}
}

// TestYAMLVParser_EdgeCases tests edge cases and special scenarios
func TestYAMLVParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "tab indentation",
			input: "user:\n\tid: UUID\n\tname: String",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				user, ok := obj.GetProperty("user")
				if !ok {
					t.Fatal("user property not found")
				}
				userObj := user.(*ast.ObjectNode)
				if len(userObj.Properties()) != 2 {
					t.Errorf("expected 2 properties, got %d", len(userObj.Properties()))
				}
			},
		},
		{
			name: "inline comment after value",
			input: "name: String # user's name\nage: Integer # user's age",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if len(obj.Properties()) != 2 {
					t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
				}
				name, ok := obj.GetProperty("name")
				if !ok {
					t.Fatal("name property not found")
				}
				typeNode := name.(*ast.TypeNode)
				if typeNode.TypeName() != "String" {
					t.Errorf("name type = %q, want String", typeNode.TypeName())
				}
			},
		},
		{
			name:  "null value",
			input: "optional: null",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				optional, ok := obj.GetProperty("optional")
				if !ok {
					t.Fatal("optional property not found")
				}
				lit := optional.(*ast.LiteralNode)
				if lit.Value() != nil {
					t.Errorf("value = %v, want nil", lit.Value())
				}
			},
		},
		{
			name:  "empty string literal",
			input: `field: ""`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, ok := obj.GetProperty("field")
				if !ok {
					t.Fatal("field property not found")
				}
				lit := field.(*ast.LiteralNode)
				if lit.Value() != "" {
					t.Errorf("value = %q, want empty string", lit.Value())
				}
			},
		},
		{
			name:  "negative integer",
			input: "min: -100",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				min, ok := obj.GetProperty("min")
				if !ok {
					t.Fatal("min property not found")
				}
				lit := min.(*ast.LiteralNode)
				if lit.Value() != int64(-100) {
					t.Errorf("value = %v, want -100", lit.Value())
				}
			},
		},
		{
			name:  "negative float",
			input: "value: -3.14",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				val, ok := obj.GetProperty("value")
				if !ok {
					t.Fatal("value property not found")
				}
				lit := val.(*ast.LiteralNode)
				if lit.Value() != -3.14 {
					t.Errorf("value = %v, want -3.14", lit.Value())
				}
			},
		},
		{
			name: "deeply nested 4 levels",
			input: `level1:
  level2:
    level3:
      level4: UUID`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				l1, _ := obj.GetProperty("level1")
				l2, _ := l1.(*ast.ObjectNode).GetProperty("level2")
				l3, _ := l2.(*ast.ObjectNode).GetProperty("level3")
				l4, ok := l3.(*ast.ObjectNode).GetProperty("level4")
				if !ok {
					t.Fatal("level4 not found")
				}
				typeNode := l4.(*ast.TypeNode)
				if typeNode.TypeName() != "UUID" {
					t.Errorf("level4 type = %q, want UUID", typeNode.TypeName())
				}
			},
		},
		{
			name: "quoted string value with spaces",
			input: `message: "hello world"`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				msg, ok := obj.GetProperty("message")
				if !ok {
					t.Fatal("message property not found")
				}
				lit := msg.(*ast.LiteralNode)
				if lit.Value() != "hello world" {
					t.Errorf("value = %q, want 'hello world'", lit.Value())
				}
			},
		},
		{
			name: "multiple properties with mixed types",
			input: `str: String
num: 42
bool: true
fn: Integer(1, 100)
arr:
  - UUID`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if len(obj.Properties()) != 5 {
					t.Errorf("expected 5 properties, got %d", len(obj.Properties()))
				}
			},
		},
		{
			name: "comment-only lines should be skipped",
			input: `# This is a comment
name: String
# Another comment
age: Integer`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if len(obj.Properties()) != 2 {
					t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
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
			tt.checkFunc(t, node)
		})
	}
}

// TestYAMLVParser_ErrorHandling tests comprehensive error scenarios
func TestYAMLVParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty input",
			input:       "",
			wantErr:     true,
			errContains: "empty input",
		},
		{
			name:        "missing colon",
			input:       "name String",
			wantErr:     true,
			errContains: "expected ':'",
		},
		{
			name: "missing value after colon",
			input: `name:
age: Integer`,
			wantErr:     true,
			errContains: "expected value",
		},
		{
			name:        "array with 3 elements",
			input:       "tags:\n  - String\n  - Integer\n  - UUID",
			wantErr:     true,
			errContains: "exactly one element",
		},
		{
			name:        "empty array bracket syntax",
			input:       "tags: []",
			wantErr:     true,
			errContains: "exactly one element",
		},
		{
			name: "mixed array and object syntax",
			input: `data:
  - item
  key: value`,
			wantErr:     true,
			errContains: "invalid YAML structure",
		},
		{
			name:        "whitespace only",
			input:       "   \n  \n  ",
			wantErr:     true,
			errContains: "empty input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// contains checks if s contains substr
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

// TestYAMLVParser_ArrayEdgeCases tests additional array edge cases
func TestYAMLVParser_ArrayEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, node ast.SchemaNode)
		wantErr   bool
	}{
		{
			name: "array with nested object element",
			input: `items:
  - 
    id: UUID
    name: String`,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				items, ok := obj.GetProperty("items")
				if !ok {
					t.Fatal("items property not found")
				}
				arr := items.(*ast.ArrayNode)
				elem := arr.ElementSchema()
				elemObj, ok := elem.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for array element, got %T", elem)
				}
				if len(elemObj.Properties()) != 2 {
					t.Errorf("expected 2 properties in array element, got %d", len(elemObj.Properties()))
				}
			},
			wantErr: false,
		},
		{
			name:      "object key directly in array context",
			input:     "arr:\n  - item\n  key: value",
			checkFunc: func(t *testing.T, node ast.SchemaNode) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}
				tt.checkFunc(t, node)
			}
		})
	}
}

// TestNativeParser_Format tests the NativeParser Format method
func TestNativeParser_Format(t *testing.T) {
	p := NewNativeParser()
	if p.Format() != parser.FormatYAMLV {
		t.Errorf("Format() = %v, want %v", p.Format(), parser.FormatYAMLV)
	}
}

// TestYAMLVParser_ParserEdgeCases tests remaining parser edge cases
func TestYAMLVParser_ParserEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "lowercase string treated as literal",
			input:   "field: lowercase",
			wantErr: false,
		},
		{
			name: "array with empty dash line",
			input: `tags:
  -`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Parse(tt.input)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestYAMLVParser_UncoveredLines tests specific uncovered code paths
func TestYAMLVParser_UncoveredLines(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "nested object with extra indented lines",
			input: `parent:
  child:
    grandchild: String
      extraIndent: ignored
  sibling: UUID`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				parent, ok := obj.GetProperty("parent")
				if !ok {
					t.Fatal("parent not found")
				}
				parentObj := parent.(*ast.ObjectNode)
				if _, ok := parentObj.GetProperty("sibling"); !ok {
					t.Error("sibling property not found")
				}
			},
		},
		{
			name:    "invalid argument in function",
			input:   "field: Function(@invalid)",
			wantErr: true,
		},
		{
			name: "boolean false argument",
			input: "field: Function(false)",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 1 || args[0] != false {
					t.Errorf("expected args=[false], got %v", args)
				}
			},
		},
		{
			name: "null argument in function",
			input: "field: Function(null)",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 1 || args[0] != nil {
					t.Errorf("expected args=[nil], got %v", args)
				}
			},
		},
		{
			name: "array item with missing value (multiline)",
			input: `items:
  -
other: value`,
			wantErr: true,
		},
		{
			name: "mixed syntax - array then object key",
			input: `data:
  - String
other: value`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if len(obj.Properties()) != 2 {
					t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, node)
				}
			}
		})
	}
}

// TestYAMLVParser_TokenizerEdgeCases tests tokenizer edge cases
func TestYAMLVParser_TokenizerEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "key starting with underscore",
			input: "_private: String",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if _, ok := obj.GetProperty("_private"); !ok {
					t.Error("_private property not found")
				}
			},
		},
		{
			name: "key with numbers",
			input: "field123: String",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if _, ok := obj.GetProperty("field123"); !ok {
					t.Error("field123 property not found")
				}
			},
		},
		{
			name: "function with nested parentheses in string",
			input: `field: Pattern("(a|b)")`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				fn := field.(*ast.FunctionNode)
				args := fn.Arguments()
				if len(args) != 1 || args[0] != "(a|b)" {
					t.Errorf("expected args=[\"(a|b)\"], got %v", args)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, node)
				}
			}
		})
	}
}

// TestYAMLVParser_AdditionalCoverage adds tests for remaining uncovered paths
func TestYAMLVParser_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "deeply nested array with object",
			input: `data:
  items:
    -
      id: UUID
      value: String`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				data, _ := obj.GetProperty("data")
				dataObj := data.(*ast.ObjectNode)
				items, _ := dataObj.GetProperty("items")
				arr := items.(*ast.ArrayNode)
				elem := arr.ElementSchema()
				elemObj, ok := elem.(*ast.ObjectNode)
				if !ok {
					t.Fatalf("expected ObjectNode for array element, got %T", elem)
				}
				if len(elemObj.Properties()) != 2 {
					t.Errorf("expected 2 properties in array element, got %d", len(elemObj.Properties()))
				}
			},
		},
		{
			name: "array with literal element",
			input: `tags:
  - "fixed-value"`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				tags, _ := obj.GetProperty("tags")
				arr := tags.(*ast.ArrayNode)
				elem := arr.ElementSchema()
				lit, ok := elem.(*ast.LiteralNode)
				if !ok {
					t.Fatalf("expected LiteralNode for array element, got %T", elem)
				}
				if lit.Value() != "fixed-value" {
					t.Errorf("expected 'fixed-value', got %v", lit.Value())
				}
			},
		},
		{
			name: "function with no parentheses not a function",
			input: `field: UUID`,
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				field, _ := obj.GetProperty("field")
				typeNode, ok := field.(*ast.TypeNode)
				if !ok {
					t.Fatalf("expected TypeNode, got %T", field)
				}
				if typeNode.TypeName() != "UUID" {
					t.Errorf("expected UUID, got %s", typeNode.TypeName())
				}
			},
		},
		{
			name: "key with hyphens",
			input: "user-name: String",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				if _, ok := obj.GetProperty("user-name"); !ok {
					t.Error("user-name property not found")
				}
			},
		},
		{
			name: "identifier with hyphens",
			input: "date: ISO-8601",
			wantErr: false,
			checkFunc: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				date, _ := obj.GetProperty("date")
				typeNode := date.(*ast.TypeNode)
				if typeNode.TypeName() != "ISO-8601" {
					t.Errorf("expected ISO-8601, got %s", typeNode.TypeName())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			node, err := p.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, node)
				}
			}
		})
	}
}
