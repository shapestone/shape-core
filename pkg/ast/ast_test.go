package ast

import (
	"encoding/json"
	"strings"
	"testing"
)

//
// Position Tests
//

func TestPosition(t *testing.T) {
	pos := NewPosition(10, 2, 5)

	if pos.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", pos.Offset)
	}

	if pos.Line != 2 {
		t.Errorf("Expected line 2, got %d", pos.Line)
	}

	if pos.Column != 5 {
		t.Errorf("Expected column 5, got %d", pos.Column)
	}

	if !pos.IsValid() {
		t.Error("Expected position to be valid")
	}

	if pos.String() != "line 2, column 5" {
		t.Errorf("Unexpected position string: %s", pos.String())
	}
}

func TestZeroPosition(t *testing.T) {
	pos := ZeroPosition()

	if pos.IsValid() {
		t.Error("Expected zero position to be invalid")
	}

	if pos.String() != "<unknown position>" {
		t.Errorf("Unexpected zero position string: %s", pos.String())
	}
}

//
// NodeType Tests
//

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{NodeTypeLiteral, "Literal"},
		{NodeTypeType, "Type"},
		{NodeTypeFunction, "Function"},
		{NodeTypeObject, "Object"},
		{NodeTypeArray, "Array"},
		{NodeType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.nodeType.String(); got != tt.expected {
			t.Errorf("NodeType(%d).String() = %s, want %s", tt.nodeType, got, tt.expected)
		}
	}
}

//
// LiteralNode Tests
//

func TestLiteralNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	tests := []struct {
		value    interface{}
		expected string
	}{
		{nil, "null"},
		{"hello", `"hello"`},
		{true, "true"},
		{false, "false"},
		{int64(42), "42"},
		{float64(3.14), "3.14"},
	}

	for _, tt := range tests {
		node := NewLiteralNode(tt.value, pos)

		if node.Type() != NodeTypeLiteral {
			t.Errorf("Expected NodeTypeLiteral, got %v", node.Type())
		}

		if node.Value() != tt.value {
			t.Errorf("Expected value %v, got %v", tt.value, node.Value())
		}

		if node.String() != tt.expected {
			t.Errorf("Expected string %q, got %q", tt.expected, node.String())
		}

		if node.Position() != pos {
			t.Errorf("Position mismatch")
		}
	}
}

func TestLiteralNode_Pooling(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Create and release a node
	node1 := NewLiteralNode("test1", pos)
	ReleaseLiteralNode(node1)

	// Create another node - should reuse from pool
	node2 := NewLiteralNode("test2", pos)

	// Verify the new node works correctly
	if node2.Value() != "test2" {
		t.Errorf("Expected 'test2', got %v", node2.Value())
	}

	// Test releasing nil node - should not panic
	ReleaseLiteralNode(nil)
}

func TestLiteralNode_String_DefaultCase(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Test with a non-standard type (uint32) to hit the default case
	node := NewLiteralNode(uint32(42), pos)
	result := node.String()

	// The default case uses fmt.Sprintf("%v", v)
	if result != "42" {
		t.Errorf("Expected '42', got %s", result)
	}
}

//
// TypeNode Tests
//

func TestTypeNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)
	node := NewTypeNode("UUID", pos)

	if node.Type() != NodeTypeType {
		t.Errorf("Expected NodeTypeType, got %v", node.Type())
	}

	if node.TypeName() != "UUID" {
		t.Errorf("Expected type name UUID, got %s", node.TypeName())
	}

	if node.String() != "UUID" {
		t.Errorf("Expected string UUID, got %s", node.String())
	}

	if node.Position() != pos {
		t.Errorf("Position mismatch")
	}
}

//
// FunctionNode Tests
//

func TestFunctionNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)
	args := []interface{}{int64(1), int64(100)}
	node := NewFunctionNode("Integer", args, pos)

	if node.Type() != NodeTypeFunction {
		t.Errorf("Expected NodeTypeFunction, got %v", node.Type())
	}

	if node.Name() != "Integer" {
		t.Errorf("Expected function name Integer, got %s", node.Name())
	}

	if len(node.Arguments()) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(node.Arguments()))
	}

	if node.String() != "Integer(1, 100)" {
		t.Errorf("Expected string 'Integer(1, 100)', got %s", node.String())
	}

	// Test with unbounded symbol
	node2 := NewFunctionNode("String", []interface{}{int64(5), "+"}, pos)
	if node2.String() != "String(5, +)" {
		t.Errorf("Expected string 'String(5, +)', got %s", node2.String())
	}

	// Test with nil argument
	node3 := NewFunctionNode("Test", []interface{}{nil}, pos)
	if node3.String() != "Test(null)" {
		t.Errorf("Expected string 'Test(null)', got %s", node3.String())
	}

	// Test with boolean arguments
	node4 := NewFunctionNode("Bool", []interface{}{true, false}, pos)
	if node4.String() != "Bool(true, false)" {
		t.Errorf("Expected string 'Bool(true, false)', got %s", node4.String())
	}
}

//
// ObjectNode Tests
//

func TestObjectNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)
	props := map[string]SchemaNode{
		"id":   NewTypeNode("UUID", pos),
		"name": NewFunctionNode("String", []interface{}{int64(1), int64(100)}, pos),
	}
	node := NewObjectNode(props, pos)

	if node.Type() != NodeTypeObject {
		t.Errorf("Expected NodeTypeObject, got %v", node.Type())
	}

	if len(node.Properties()) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(node.Properties()))
	}

	idNode, ok := node.GetProperty("id")
	if !ok {
		t.Error("Expected to find 'id' property")
	}

	if idNode.Type() != NodeTypeType {
		t.Error("Expected 'id' to be a TypeNode")
	}

	// Test empty object
	emptyNode := NewObjectNode(map[string]SchemaNode{}, pos)
	if emptyNode.String() != "{}" {
		t.Errorf("Expected empty object string '{}', got %s", emptyNode.String())
	}
}

//
// ArrayNode Tests
//

func TestArrayNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)
	elemSchema := NewFunctionNode("String", []interface{}{int64(1), int64(50)}, pos)
	node := NewArrayNode(elemSchema, pos)

	if node.Type() != NodeTypeArray {
		t.Errorf("Expected NodeTypeArray, got %v", node.Type())
	}

	if node.ElementSchema() == nil {
		t.Error("Expected element schema to be set")
	}

	if node.ElementSchema().Type() != NodeTypeFunction {
		t.Error("Expected element schema to be a FunctionNode")
	}

	if node.String() != "[String(1, 50)]" {
		t.Errorf("Expected string '[String(1, 50)]', got %s", node.String())
	}

	if node.Position() != pos {
		t.Errorf("Position mismatch")
	}
}

func TestObjectNode_Pooling(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Create and release a node
	props1 := map[string]SchemaNode{
		"test": NewLiteralNode("value1", pos),
	}
	node1 := NewObjectNode(props1, pos)
	ReleaseObjectNode(node1)

	// Create another node - should reuse from pool
	props2 := map[string]SchemaNode{
		"test": NewLiteralNode("value2", pos),
	}
	node2 := NewObjectNode(props2, pos)

	// Verify the new node works correctly
	prop, ok := node2.GetProperty("test")
	if !ok {
		t.Error("Expected to find 'test' property")
	}
	if lit, ok := prop.(*LiteralNode); ok {
		if lit.Value() != "value2" {
			t.Errorf("Expected 'value2', got %v", lit.Value())
		}
	}

	// Test releasing nil node - should not panic
	ReleaseObjectNode(nil)
}

//
// Visitor Pattern Tests
//

type countVisitor struct {
	BaseVisitor
	literalCount  int
	typeCount     int
	functionCount int
	objectCount   int
	arrayCount    int
}

func (v *countVisitor) VisitLiteral(node *LiteralNode) error {
	v.literalCount++
	return nil
}

func (v *countVisitor) VisitType(node *TypeNode) error {
	v.typeCount++
	return nil
}

func (v *countVisitor) VisitFunction(node *FunctionNode) error {
	v.functionCount++
	return nil
}

func (v *countVisitor) VisitObject(node *ObjectNode) error {
	v.objectCount++
	// Traverse all properties
	for _, propNode := range node.Properties() {
		if err := propNode.Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v *countVisitor) VisitArray(node *ArrayNode) error {
	v.arrayCount++
	// Traverse element schema
	return node.ElementSchema().Accept(v)
}

func TestVisitorPattern(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Build complex AST
	obj := NewObjectNode(map[string]SchemaNode{
		"id":    NewTypeNode("UUID", pos),
		"name":  NewFunctionNode("String", []interface{}{int64(1), int64(100)}, pos),
		"tags":  NewArrayNode(NewLiteralNode("tag", pos), pos),
		"count": NewLiteralNode(int64(5), pos),
	}, pos)

	// Visit AST
	visitor := &countVisitor{}
	if err := Walk(obj, visitor); err != nil {
		t.Errorf("Walk failed: %v", err)
	}

	if visitor.objectCount != 1 {
		t.Errorf("Expected 1 object, got %d", visitor.objectCount)
	}

	if visitor.typeCount != 1 {
		t.Errorf("Expected 1 type, got %d", visitor.typeCount)
	}

	if visitor.functionCount != 1 {
		t.Errorf("Expected 1 function, got %d", visitor.functionCount)
	}

	if visitor.arrayCount != 1 {
		t.Errorf("Expected 1 array, got %d", visitor.arrayCount)
	}

	if visitor.literalCount != 2 {
		t.Errorf("Expected 2 literals, got %d", visitor.literalCount)
	}
}

//
// Serialization Tests
//

func TestJSONSerialization(t *testing.T) {
	pos := NewPosition(10, 2, 5)

	tests := []struct {
		name string
		node SchemaNode
	}{
		{
			name: "LiteralNode",
			node: NewLiteralNode("hello", pos),
		},
		{
			name: "TypeNode",
			node: NewTypeNode("UUID", pos),
		},
		{
			name: "FunctionNode",
			node: NewFunctionNode("Integer", []interface{}{int64(1), int64(100)}, pos),
		},
		{
			name: "ObjectNode",
			node: NewObjectNode(map[string]SchemaNode{
				"id": NewTypeNode("UUID", pos),
			}, pos),
		},
		{
			name: "ArrayNode",
			node: NewArrayNode(NewTypeNode("String", pos), pos),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.node)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			node, err := UnmarshalSchemaNode(data)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Verify type
			if node.Type() != tt.node.Type() {
				t.Errorf("Type mismatch: got %v, want %v", node.Type(), tt.node.Type())
			}

			// Verify position
			if node.Position() != tt.node.Position() {
				t.Errorf("Position mismatch")
			}
		})
	}
}

func TestComplexSerialization(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Build complex nested structure
	original := NewObjectNode(map[string]SchemaNode{
		"user": NewObjectNode(map[string]SchemaNode{
			"id":    NewTypeNode("UUID", pos),
			"email": NewTypeNode("Email", pos),
			"age":   NewFunctionNode("Integer", []interface{}{int64(0), int64(120)}, pos),
		}, pos),
		"tags": NewArrayNode(NewFunctionNode("String", []interface{}{int64(1), int64(30)}, pos), pos),
	}, pos)

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	restored, err := UnmarshalSchemaNode(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify structure
	objNode, ok := restored.(*ObjectNode)
	if !ok {
		t.Fatal("Expected ObjectNode")
	}

	if len(objNode.Properties()) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(objNode.Properties()))
	}

	userNode, ok := objNode.GetProperty("user")
	if !ok {
		t.Fatal("Expected to find 'user' property")
	}

	if userNode.Type() != NodeTypeObject {
		t.Error("Expected 'user' to be an ObjectNode")
	}
}

//
// Pretty Printer Tests
//

func TestPrettyPrint(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	tests := []struct {
		name     string
		node     SchemaNode
		contains []string
	}{
		{
			name:     "Literal",
			node:     NewLiteralNode("test", pos),
			contains: []string{"Literal:", `"test"`},
		},
		{
			name:     "Type",
			node:     NewTypeNode("UUID", pos),
			contains: []string{"Type:", "UUID"},
		},
		{
			name:     "Function",
			node:     NewFunctionNode("Integer", []interface{}{int64(1), int64(100)}, pos),
			contains: []string{"Function:", "Integer(1, 100)"},
		},
		{
			name: "Object",
			node: NewObjectNode(map[string]SchemaNode{
				"id": NewTypeNode("UUID", pos),
			}, pos),
			contains: []string{"Object:", `"id":`},
		},
		{
			name:     "Array",
			node:     NewArrayNode(NewTypeNode("String", pos), pos),
			contains: []string{"Array:", "element:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrettyPrint(tt.node)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("PrettyPrint output missing %q:\n%s", substr, result)
				}
			}
		})
	}
}

func TestTreePrint(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	obj := NewObjectNode(map[string]SchemaNode{
		"id": NewTypeNode("UUID", pos),
	}, pos)

	result := TreePrint(obj)

	expected := []string{"└──", "Object", `"id":`, "Type:", "UUID"}

	for _, substr := range expected {
		if !strings.Contains(result, substr) {
			t.Errorf("TreePrint output missing %q:\n%s", substr, result)
		}
	}
}

//
// ArrayDataNode Tests
//

func TestArrayDataNode(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	elements := []SchemaNode{
		NewLiteralNode("item1", pos),
		NewLiteralNode("item2", pos),
		NewLiteralNode(int64(42), pos),
	}

	node := NewArrayDataNode(elements, pos)

	if node.Type() != NodeTypeArrayData {
		t.Errorf("Expected NodeTypeArrayData, got %v", node.Type())
	}

	if node.Len() != 3 {
		t.Errorf("Expected length 3, got %d", node.Len())
	}

	if node.Position() != pos {
		t.Errorf("Expected position %v, got %v", pos, node.Position())
	}

	// Test Elements()
	retrievedElements := node.Elements()
	if len(retrievedElements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(retrievedElements))
	}

	// Test Get()
	elem := node.Get(0)
	if elem == nil {
		t.Error("Expected element at index 0, got nil")
	}
	if lit, ok := elem.(*LiteralNode); ok {
		if lit.Value() != "item1" {
			t.Errorf("Expected 'item1', got %v", lit.Value())
		}
	} else {
		t.Error("Expected LiteralNode at index 0")
	}

	// Test out of bounds
	if node.Get(10) != nil {
		t.Error("Expected nil for out of bounds index")
	}

	if node.Get(-1) != nil {
		t.Error("Expected nil for negative index")
	}

	// Test String()
	str := node.String()
	if !strings.Contains(str, "[") || !strings.Contains(str, "]") {
		t.Errorf("Expected array string representation, got: %s", str)
	}

	// Test String() with empty array
	emptyNode := NewArrayDataNode([]SchemaNode{}, pos)
	if emptyNode.String() != "[]" {
		t.Errorf("Empty array String() = %q, want '[]'", emptyNode.String())
	}
}

func TestArrayDataNode_Pooling(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	// Create and release a node
	elements1 := []SchemaNode{
		NewLiteralNode("test", pos),
	}
	node1 := NewArrayDataNode(elements1, pos)

	// Release it back to the pool
	ReleaseArrayDataNode(node1)

	// Create another node - should reuse from pool
	elements2 := []SchemaNode{
		NewLiteralNode("test2", pos),
	}
	node2 := NewArrayDataNode(elements2, pos)

	// Verify the new node works correctly
	if node2.Len() != 1 {
		t.Errorf("Expected length 1, got %d", node2.Len())
	}

	if elem := node2.Get(0); elem != nil {
		if lit, ok := elem.(*LiteralNode); ok {
			if lit.Value() != "test2" {
				t.Errorf("Expected 'test2', got %v", lit.Value())
			}
		}
	}

	// Test releasing nil node - should not panic
	ReleaseArrayDataNode(nil)
}

func TestArrayDataNode_Visitor(t *testing.T) {
	pos := NewPosition(0, 1, 1)

	elements := []SchemaNode{
		NewLiteralNode("test", pos),
	}
	node := NewArrayDataNode(elements, pos)

	// Test visitor pattern
	visitor := &testVisitor{}
	err := node.Accept(visitor)
	if err != nil {
		t.Errorf("Accept() error = %v", err)
	}

	if !visitor.visitedArrayData {
		t.Error("Visitor did not visit ArrayDataNode")
	}
}

// Helper visitor for testing
type testVisitor struct {
	BaseVisitor
	visitedArrayData bool
}

func (v *testVisitor) VisitArrayData(node *ArrayDataNode) error {
	v.visitedArrayData = true
	return nil
}

// TestBaseVisitor_VisitArrayData tests the default BaseVisitor implementation
func TestBaseVisitor_VisitArrayData(t *testing.T) {
	pos := NewPosition(0, 1, 1)
	elements := []SchemaNode{
		NewLiteralNode("value", pos),
	}
	node := NewArrayDataNode(elements, pos)

	// Use BaseVisitor directly (not overridden)
	visitor := &BaseVisitor{}
	err := visitor.VisitArrayData(node)
	if err != nil {
		t.Errorf("BaseVisitor.VisitArrayData() error = %v, want nil", err)
	}
}
