package propsv

import (
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

func TestPropsVParser_SimpleProperties(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "single property with UUID",
			input: `id=UUID`,
			wantErr: false,
		},
		{
			name: "single property with function",
			input: `name=String(1, 100)`,
			wantErr: false,
		},
		{
			name: "multiple properties",
			input: `id=UUID
name=String(1, 100)
email=Email`,
			wantErr: false,
		},
		{
			name: "with comments",
			input: `# User schema
id=UUID
name=String(1, 100)`,
			wantErr: false,
		},
		{
			name: "boolean literal",
			input: `active=true`,
			wantErr: false,
		},
		{
			name: "number literal",
			input: `age=42`,
			wantErr: false,
		},
		{
			name: "null literal",
			input: `optional=null`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			node, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if node == nil {
					t.Error("Parse() returned nil node")
					return
				}

				// Should return an ObjectNode
				if node.Type() != ast.NodeTypeObject {
					t.Errorf("Parse() returned %v, want ObjectNode", node.Type())
				}
			}
		})
	}
}

func TestPropsVParser_NestedProperties(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		checkProps func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple nested property",
			input: `user.id=UUID
user.name=String(1, 100)`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatal("expected ObjectNode")
				}

				// Should have "user" property
				userProp, ok := obj.Properties()["user"]
				if !ok {
					t.Error("missing 'user' property")
					return
				}

				// User should be an object
				userObj, ok := userProp.(*ast.ObjectNode)
				if !ok {
					t.Error("'user' should be ObjectNode")
					return
				}

				// Check nested properties
				if _, ok := userObj.Properties()["id"]; !ok {
					t.Error("missing 'user.id' property")
				}
				if _, ok := userObj.Properties()["name"]; !ok {
					t.Error("missing 'user.name' property")
				}
			},
		},
		{
			name: "deeply nested property",
			input: `user.profile.firstName=String(1, 50)
user.profile.lastName=String(1, 50)`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				userObj := obj.Properties()["user"].(*ast.ObjectNode)
				profileObj := userObj.Properties()["profile"].(*ast.ObjectNode)

				if _, ok := profileObj.Properties()["firstName"]; !ok {
					t.Error("missing 'firstName'")
				}
				if _, ok := profileObj.Properties()["lastName"]; !ok {
					t.Error("missing 'lastName'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			node, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkProps != nil {
				tt.checkProps(t, node)
			}
		})
	}
}

func TestPropsVParser_Arrays(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		checkProps func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:    "simple array",
			input:   `tags[]=String(1, 30)`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				tagsProp, ok := obj.Properties()["tags"]
				if !ok {
					t.Error("missing 'tags' property")
					return
				}

				// Should be ArrayNode
				if tagsProp.Type() != ast.NodeTypeArray {
					t.Errorf("'tags' should be ArrayNode, got %v", tagsProp.Type())
				}
			},
		},
		{
			name: "nested array",
			input: `user.roles[]=String(1, 50)`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				userObj := obj.Properties()["user"].(*ast.ObjectNode)
				rolesProp := userObj.Properties()["roles"]

				if rolesProp.Type() != ast.NodeTypeArray {
					t.Errorf("'roles' should be ArrayNode, got %v", rolesProp.Type())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			node, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkProps != nil {
				tt.checkProps(t, node)
			}
		})
	}
}

func TestPropsVParser_Functions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name:    "function with numeric args",
			input:   `name=String(1, 100)`,
			wantErr: false,
			checkFn: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				nameProp := obj.Properties()["name"]
				
				fnNode, ok := nameProp.(*ast.FunctionNode)
				if !ok {
					t.Fatalf("expected FunctionNode, got %T", nameProp)
				}

				if fnNode.Name() != "String" {
					t.Errorf("function name = %s, want String", fnNode.Name())
				}

				args := fnNode.Arguments()
				if len(args) != 2 {
					t.Errorf("len(args) = %d, want 2", len(args))
				}
			},
		},
		{
			name:    "function with unbounded arg",
			input:   `name=String(5+)`,
			wantErr: false,
			checkFn: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				nameProp := obj.Properties()["name"]
				
				fnNode := nameProp.(*ast.FunctionNode)
				args := fnNode.Arguments()
				
				// Should have [5, "+"]
				if len(args) != 2 {
					t.Errorf("len(args) = %d, want 2", len(args))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			node, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, node)
			}
		})
	}
}

func TestPropsVParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "missing equals",
			input:   `id UUID`,
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   `id=`,
			wantErr: true,
		},
		{
			name:    "invalid property name",
			input:   `123invalid=UUID`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPropsVParser_ComplexExample(t *testing.T) {
	input := `# User schema with comments
id=UUID
email=Email
name=String(1, 100)
age=Integer(18, 120)

# Nested profile
user.profile.firstName=String(1, 50)
user.profile.lastName=String(1, 50)

# Arrays
tags[]=String(1, 30)
user.roles[]=String(1, 50)

# Literals
active=true
count=42`

	parser := NewParser()
	node, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatal("expected ObjectNode")
	}

	// Check top-level properties
	expectedProps := []string{"id", "email", "name", "age", "user", "tags", "active", "count"}
	for _, prop := range expectedProps {
		if _, ok := obj.Properties()[prop]; !ok {
			t.Errorf("missing property: %s", prop)
		}
	}

	// Check nested user.profile
	userObj, ok := obj.Properties()["user"].(*ast.ObjectNode)
	if !ok {
		t.Fatal("'user' should be ObjectNode")
	}

	profileObj, ok := userObj.Properties()["profile"].(*ast.ObjectNode)
	if !ok {
		t.Fatal("'user.profile' should be ObjectNode")
	}

	if _, ok := profileObj.Properties()["firstName"]; !ok {
		t.Error("missing 'firstName'")
	}

	// Check array
	if obj.Properties()["tags"].Type() != ast.NodeTypeArray {
		t.Error("'tags' should be ArrayNode")
	}
}
