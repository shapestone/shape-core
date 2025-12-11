package ast

import (
	"testing"
)

func TestBaseVisitor_VisitLiteral(t *testing.T) {
	visitor := &BaseVisitor{}
	node := NewLiteralNode("test", Position{})

	err := visitor.VisitLiteral(node)
	if err != nil {
		t.Errorf("VisitLiteral() error = %v, want nil", err)
	}
}

func TestBaseVisitor_VisitType(t *testing.T) {
	visitor := &BaseVisitor{}
	node := NewTypeNode("String", Position{})

	err := visitor.VisitType(node)
	if err != nil {
		t.Errorf("VisitType() error = %v, want nil", err)
	}
}

func TestBaseVisitor_VisitFunction(t *testing.T) {
	visitor := &BaseVisitor{}
	node := NewFunctionNode("TestFunc", []interface{}{}, Position{})

	err := visitor.VisitFunction(node)
	if err != nil {
		t.Errorf("VisitFunction() error = %v, want nil", err)
	}
}

func TestBaseVisitor_VisitObject(t *testing.T) {
	visitor := &BaseVisitor{}
	node := NewObjectNode(make(map[string]SchemaNode), Position{})

	err := visitor.VisitObject(node)
	if err != nil {
		t.Errorf("VisitObject() error = %v, want nil", err)
	}
}

func TestBaseVisitor_VisitArray(t *testing.T) {
	visitor := &BaseVisitor{}
	node := NewArrayNode(NewTypeNode("String", Position{}), Position{})

	err := visitor.VisitArray(node)
	if err != nil {
		t.Errorf("VisitArray() error = %v, want nil", err)
	}
}

// TestVisitor is a test visitor that tracks which nodes were visited
type TestVisitor struct {
	BaseVisitor
	visitedLiteral  bool
	visitedType     bool
	visitedFunction bool
	visitedObject   bool
	visitedArray    bool
}

func (v *TestVisitor) VisitLiteral(node *LiteralNode) error {
	v.visitedLiteral = true
	return nil
}

func (v *TestVisitor) VisitType(node *TypeNode) error {
	v.visitedType = true
	return nil
}

func (v *TestVisitor) VisitFunction(node *FunctionNode) error {
	v.visitedFunction = true
	return nil
}

func (v *TestVisitor) VisitObject(node *ObjectNode) error {
	v.visitedObject = true
	return nil
}

func (v *TestVisitor) VisitArray(node *ArrayNode) error {
	v.visitedArray = true
	return nil
}

func TestWalk_Literal(t *testing.T) {
	visitor := &TestVisitor{}
	node := NewLiteralNode("test", Position{})

	err := Walk(node, visitor)
	if err != nil {
		t.Errorf("Walk() error = %v, want nil", err)
	}
	if !visitor.visitedLiteral {
		t.Error("Walk() did not visit LiteralNode")
	}
}

func TestWalk_Type(t *testing.T) {
	visitor := &TestVisitor{}
	node := NewTypeNode("String", Position{})

	err := Walk(node, visitor)
	if err != nil {
		t.Errorf("Walk() error = %v, want nil", err)
	}
	if !visitor.visitedType {
		t.Error("Walk() did not visit TypeNode")
	}
}

func TestWalk_Function(t *testing.T) {
	visitor := &TestVisitor{}
	node := NewFunctionNode("TestFunc", []interface{}{}, Position{})

	err := Walk(node, visitor)
	if err != nil {
		t.Errorf("Walk() error = %v, want nil", err)
	}
	if !visitor.visitedFunction {
		t.Error("Walk() did not visit FunctionNode")
	}
}

func TestWalk_Object(t *testing.T) {
	visitor := &TestVisitor{}
	node := NewObjectNode(make(map[string]SchemaNode), Position{})

	err := Walk(node, visitor)
	if err != nil {
		t.Errorf("Walk() error = %v, want nil", err)
	}
	if !visitor.visitedObject {
		t.Error("Walk() did not visit ObjectNode")
	}
}

func TestWalk_Array(t *testing.T) {
	visitor := &TestVisitor{}
	node := NewArrayNode(NewTypeNode("String", Position{}), Position{})

	err := Walk(node, visitor)
	if err != nil {
		t.Errorf("Walk() error = %v, want nil", err)
	}
	if !visitor.visitedArray {
		t.Error("Walk() did not visit ArrayNode")
	}
}
