package ast

import (
	"testing"
)

func TestStringInterner_Get(t *testing.T) {
	si := newStringInterner()

	// Test that common types return the expected value
	commonTypes := []string{
		"UUID", "Email", "String", "Integer",
	}

	for _, typeName := range commonTypes {
		s1 := si.Get(typeName)
		s2 := si.Get(typeName)

		// Same value
		if s1 != typeName {
			t.Errorf("Expected %q, got %q", typeName, s1)
		}
		if s2 != typeName {
			t.Errorf("Expected %q, got %q", typeName, s2)
		}
		if s1 != s2 {
			t.Errorf("Expected same value for %q, got %q and %q", typeName, s1, s2)
		}
	}
}

func TestStringInterner_NewStrings(t *testing.T) {
	si := newStringInterner()

	// Test that new strings are interned
	s1 := si.Get("CustomType")
	s2 := si.Get("CustomType")

	// Same value
	if s1 != "CustomType" {
		t.Errorf("Expected 'CustomType', got %q", s1)
	}
	if s1 != s2 {
		t.Error("Expected same value for new string, got different values")
	}
}

func TestInternString(t *testing.T) {
	// Test that global InternString function works
	s1 := InternString("UUID")
	s2 := InternString("UUID")

	if s1 != "UUID" || s2 != "UUID" {
		t.Errorf("Expected 'UUID', got %q and %q", s1, s2)
	}
	if s1 != s2 {
		t.Error("Expected same value from InternString, got different values")
	}
}

func TestTypeNode_StringInterning(t *testing.T) {
	// Create multiple type nodes with the same type name
	node1 := NewTypeNode("UUID", Position{Line: 1, Column: 1})
	node2 := NewTypeNode("UUID", Position{Line: 2, Column: 1})

	// Type names should be the same value
	name1 := node1.TypeName()
	name2 := node2.TypeName()

	if name1 != "UUID" || name2 != "UUID" {
		t.Errorf("Expected 'UUID', got %q and %q", name1, name2)
	}
	if name1 != name2 {
		t.Error("Expected same value for type names, got different values")
	}
}

func TestFunctionNode_StringInterning(t *testing.T) {
	// Create multiple function nodes with the same function name
	node1 := NewFunctionNode("String", []interface{}{int64(1), int64(100)}, Position{Line: 1, Column: 1})
	node2 := NewFunctionNode("String", []interface{}{int64(5), int64(50)}, Position{Line: 2, Column: 1})

	// Function names should be the same value
	name1 := node1.Name()
	name2 := node2.Name()

	if name1 != "String" || name2 != "String" {
		t.Errorf("Expected 'String', got %q and %q", name1, name2)
	}
	if name1 != name2 {
		t.Error("Expected same value for function names, got different values")
	}
}

func BenchmarkStringInterner_Get_Common(b *testing.B) {
	si := newStringInterner()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = si.Get("UUID")
	}
}

func BenchmarkStringInterner_Get_New(b *testing.B) {
	si := newStringInterner()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = si.Get("CustomType")
	}
}

func BenchmarkTypeNode_WithInterning(b *testing.B) {
	pos := Position{Line: 1, Column: 1}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewTypeNode("UUID", pos)
	}
}

func BenchmarkFunctionNode_WithInterning(b *testing.B) {
	pos := Position{Line: 1, Column: 1}
	args := []interface{}{int64(1), int64(100)}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewFunctionNode("String", args, pos)
	}
}
