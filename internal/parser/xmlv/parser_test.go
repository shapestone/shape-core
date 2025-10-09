package xmlv

import (
	"testing"

	"github.com/shapestone/shape/pkg/ast"
)

func TestXMLVParser_SimpleElement(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "type identifier",
			input:   `<id>UUID</id>`,
			wantErr: false,
		},
		{
			name:    "function",
			input:   `<name>String(1, 100)</name>`,
			wantErr: false,
		},
		{
			name:    "boolean",
			input:   `<active>true</active>`,
			wantErr: false,
		},
		{
			name:    "number",
			input:   `<count>42</count>`,
			wantErr: false,
		},
		{
			name:    "null",
			input:   `<value>null</value>`,
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

			if !tt.wantErr && node == nil {
				t.Error("Parse() returned nil node")
			}
		})
	}
}

func TestXMLVParser_NestedElements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		checkProps func(t *testing.T, node ast.SchemaNode)
	}{
		{
			name: "simple nested object",
			input: `<user>
	<id>UUID</id>
	<name>String(1, 100)</name>
</user>`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj, ok := node.(*ast.ObjectNode)
				if !ok {
					t.Fatal("expected ObjectNode")
				}

				if _, ok := obj.Properties()["id"]; !ok {
					t.Error("missing 'id' property")
				}
				if _, ok := obj.Properties()["name"]; !ok {
					t.Error("missing 'name' property")
				}
			},
		},
		{
			name: "deeply nested object",
			input: `<root>
	<user>
		<profile>
			<firstName>String(1, 50)</firstName>
		</profile>
	</user>
</root>`,
			wantErr: false,
			checkProps: func(t *testing.T, node ast.SchemaNode) {
				obj := node.(*ast.ObjectNode)
				userObj := obj.Properties()["user"].(*ast.ObjectNode)
				profileObj := userObj.Properties()["profile"].(*ast.ObjectNode)

				if _, ok := profileObj.Properties()["firstName"]; !ok {
					t.Error("missing 'firstName'")
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

func TestXMLVParser_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "unclosed tag",
			input:   `<user><id>UUID</id>`,
			wantErr: true,
		},
		{
			name:    "mismatched tags",
			input:   `<user><id>UUID</name></user>`,
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

func TestXMLVParser_ComplexExample(t *testing.T) {
	input := `<user>
	<id>UUID</id>
	<name>String(1, 100)</name>
	<age>Integer(1, 120)</age>
	<email>Email</email>
	<active>true</active>
</user>`

	parser := NewParser()
	node, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatal("expected ObjectNode")
	}

	// Check all properties exist
	expectedProps := []string{"id", "name", "age", "email", "active"}
	for _, prop := range expectedProps {
		if _, ok := obj.Properties()[prop]; !ok {
			t.Errorf("missing property: %s", prop)
		}
	}

	// Check types
	if obj.Properties()["id"].Type() != ast.NodeTypeType {
		t.Error("'id' should be TypeNode")
	}
	if obj.Properties()["name"].Type() != ast.NodeTypeFunction {
		t.Error("'name' should be FunctionNode")
	}
	if obj.Properties()["active"].Type() != ast.NodeTypeLiteral {
		t.Error("'active' should be LiteralNode")
	}
}
