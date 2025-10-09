package shape

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
)

// TestCrossFormat_SimpleObject verifies that the same simple object schema
// produces equivalent ASTs across all four formats.
func TestCrossFormat_SimpleObject(t *testing.T) {
	schemas := map[parser.Format]string{
		parser.FormatJSONV: `{"id": UUID, "name": String}`,
		parser.FormatPropsV: `id=UUID
name=String`,
		parser.FormatXMLV: `<root><id>UUID</id><name>String</name></root>`,
		parser.FormatCSVV: `id,name
UUID,String`,
	}

	var nodes []ast.SchemaNode
	for format, input := range schemas {
		node, err := Parse(format, input)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", format, err)
		}
		nodes = append(nodes, node)
	}

	// All should be ObjectNode
	for i, node := range nodes {
		if node.Type() != ast.NodeTypeObject {
			t.Errorf("node[%d].Type() = %v, want %v", i, node.Type(), ast.NodeTypeObject)
		}
	}

	// Verify all have same properties
	for i := 0; i < len(nodes); i++ {
		obj1 := nodes[i].(*ast.ObjectNode)
		for j := i + 1; j < len(nodes); j++ {
			obj2 := nodes[j].(*ast.ObjectNode)
			if len(obj1.Properties()) != len(obj2.Properties()) {
				t.Errorf("node[%d] has %d properties, node[%d] has %d properties",
					i, len(obj1.Properties()), j, len(obj2.Properties()))
			}

			// Check property existence
			for propName := range obj1.Properties() {
				if _, ok := obj2.GetProperty(propName); !ok {
					t.Errorf("node[%d] has property %q but node[%d] doesn't", i, propName, j)
				}
			}
		}
	}
}

// TestCrossFormat_WithFunctions verifies that function schemas
// produce equivalent ASTs across all formats.
func TestCrossFormat_WithFunctions(t *testing.T) {
	schemas := map[parser.Format]string{
		parser.FormatJSONV: `{"username": String(3, 20), "age": Integer(18, 120)}`,
		parser.FormatPropsV: `username=String(3, 20)
age=Integer(18, 120)`,
		parser.FormatXMLV: `<root><username>String(3, 20)</username><age>Integer(18, 120)</age></root>`,
		parser.FormatCSVV: `username,age
String(3, 20),Integer(18, 120)`,
	}

	var nodes []ast.SchemaNode
	for format, input := range schemas {
		node, err := Parse(format, input)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", format, err)
		}
		nodes = append(nodes, node)
	}

	// Verify all have same structure
	for i, node := range nodes {
		obj, ok := node.(*ast.ObjectNode)
		if !ok {
			t.Fatalf("node[%d] is not ObjectNode", i)
		}

		// Check username
		username, ok := obj.GetProperty("username")
		if !ok {
			t.Errorf("node[%d] missing 'username' property", i)
			continue
		}
		fn, ok := username.(*ast.FunctionNode)
		if !ok {
			t.Errorf("node[%d] 'username' is not FunctionNode, got %T", i, username)
			continue
		}
		if fn.Name() != "String" {
			t.Errorf("node[%d] 'username' function name = %q, want %q", i, fn.Name(), "String")
		}

		// Check age
		age, ok := obj.GetProperty("age")
		if !ok {
			t.Errorf("node[%d] missing 'age' property", i)
			continue
		}
		fn2, ok := age.(*ast.FunctionNode)
		if !ok {
			t.Errorf("node[%d] 'age' is not FunctionNode, got %T", i, age)
			continue
		}
		if fn2.Name() != "Integer" {
			t.Errorf("node[%d] 'age' function name = %q, want %q", i, fn2.Name(), "Integer")
		}
	}
}

// TestCrossFormat_WithLiterals verifies that literal schemas
// produce equivalent ASTs across all formats.
func TestCrossFormat_WithLiterals(t *testing.T) {
	schemas := map[parser.Format]string{
		parser.FormatJSONV:  `{"active": true, "count": 42}`,
		parser.FormatPropsV: `active=true
count=42`,
		parser.FormatXMLV: `<root><active>true</active><count>42</count></root>`,
		parser.FormatCSVV:  `active,count
true,42`,
	}

	var nodes []ast.SchemaNode
	for format, input := range schemas {
		node, err := Parse(format, input)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", format, err)
		}
		nodes = append(nodes, node)
	}

	// Verify all have same literal values
	for i, node := range nodes {
		obj, ok := node.(*ast.ObjectNode)
		if !ok {
			t.Fatalf("node[%d] is not ObjectNode", i)
		}

		// Check active
		active, ok := obj.GetProperty("active")
		if !ok {
			t.Errorf("node[%d] missing 'active' property", i)
			continue
		}
		lit, ok := active.(*ast.LiteralNode)
		if !ok {
			t.Errorf("node[%d] 'active' is not LiteralNode, got %T", i, active)
			continue
		}
		if lit.Value() != true {
			t.Errorf("node[%d] 'active' value = %v, want true", i, lit.Value())
		}

		// Check count
		count, ok := obj.GetProperty("count")
		if !ok {
			t.Errorf("node[%d] missing 'count' property", i)
			continue
		}
		lit2, ok := count.(*ast.LiteralNode)
		if !ok {
			t.Errorf("node[%d] 'count' is not LiteralNode, got %T", i, count)
			continue
		}
		if lit2.Value() != int64(42) {
			t.Errorf("node[%d] 'count' value = %v, want 42", i, lit2.Value())
		}
	}
}

// TestCrossFormat_NestedObjects verifies that nested object schemas
// produce equivalent ASTs across formats that support nesting.
func TestCrossFormat_NestedObjects(t *testing.T) {
	schemas := map[parser.Format]string{
		parser.FormatJSONV: `{"user": {"id": UUID, "name": String}}`,
		parser.FormatPropsV: `user.id=UUID
user.name=String`,
		parser.FormatXMLV: `<root><user><id>UUID</id><name>String</name></user></root>`,
	}

	var nodes []ast.SchemaNode
	for format, input := range schemas {
		node, err := Parse(format, input)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", format, err)
		}
		nodes = append(nodes, node)
	}

	// Verify all have nested structure
	for i, node := range nodes {
		obj, ok := node.(*ast.ObjectNode)
		if !ok {
			t.Fatalf("node[%d] is not ObjectNode", i)
		}

		user, ok := obj.GetProperty("user")
		if !ok {
			t.Errorf("node[%d] missing 'user' property", i)
			continue
		}

		userObj, ok := user.(*ast.ObjectNode)
		if !ok {
			t.Errorf("node[%d] 'user' is not ObjectNode, got %T", i, user)
			continue
		}

		// Check nested properties
		if _, ok := userObj.GetProperty("id"); !ok {
			t.Errorf("node[%d] 'user.id' property not found", i)
		}
		if _, ok := userObj.GetProperty("name"); !ok {
			t.Errorf("node[%d] 'user.name' property not found", i)
		}
	}
}

// TestCrossFormat_ComplexSchema verifies a complex schema across all formats.
func TestCrossFormat_ComplexSchema(t *testing.T) {
	schemas := map[parser.Format]string{
		parser.FormatJSONV: `{
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"active": true
		}`,
		parser.FormatPropsV: `id=UUID
username=String(3, 20)
email=Email
age=Integer(18, 120)
active=true`,
		parser.FormatXMLV: `<root>
			<id>UUID</id>
			<username>String(3, 20)</username>
			<email>Email</email>
			<age>Integer(18, 120)</age>
			<active>true</active>
		</root>`,
		parser.FormatCSVV: `id,username,email,age,active
UUID,String(3, 20),Email,Integer(18, 120),true`,
	}

	expectedProps := []string{"id", "username", "email", "age", "active"}

	for format, input := range schemas {
		t.Run(format.String(), func(t *testing.T) {
			node, err := Parse(format, input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			obj, ok := node.(*ast.ObjectNode)
			if !ok {
				t.Fatalf("expected ObjectNode, got %T", node)
			}

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
			if typeNode, ok := id.(*ast.TypeNode); !ok {
				t.Errorf("'id' should be TypeNode, got %T", id)
			} else if typeNode.TypeName() != "UUID" {
				t.Errorf("'id' type = %q, want %q", typeNode.TypeName(), "UUID")
			}

			username, _ := obj.GetProperty("username")
			if fn, ok := username.(*ast.FunctionNode); !ok {
				t.Errorf("'username' should be FunctionNode, got %T", username)
			} else if fn.Name() != "String" {
				t.Errorf("'username' function = %q, want %q", fn.Name(), "String")
			}

			active, _ := obj.GetProperty("active")
			if lit, ok := active.(*ast.LiteralNode); !ok {
				t.Errorf("'active' should be LiteralNode, got %T", active)
			} else if lit.Value() != true {
				t.Errorf("'active' value = %v, want true", lit.Value())
			}
		})
	}
}
