package grammar

import (
	"fmt"

	"github.com/shapestone/shape-core/pkg/ast"
)

// ASTEqual performs deep comparison of two AST nodes for structural equality.
//
// Returns true if the nodes have the same type, structure, and values.
// Position information is ignored in the comparison.
//
// This is useful for dual parser verification where a reference parser's
// output is compared against a production parser's output.
//
// Example:
//
//	refAST, _ := referenceParser.Parse(input)
//	prodAST, _ := productionParser.Parse(input)
//	if !grammar.ASTEqual(refAST, prodAST) {
//	    t.Error("Parsers produce different ASTs")
//	}
func ASTEqual(a, b ast.SchemaNode) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Must be same type
	if a.Type() != b.Type() {
		return false
	}

	// Compare based on node type
	switch aNode := a.(type) {
	case *ast.LiteralNode:
		return literalEqual(aNode, b.(*ast.LiteralNode))
	case *ast.TypeNode:
		return typeEqual(aNode, b.(*ast.TypeNode))
	case *ast.FunctionNode:
		return functionEqual(aNode, b.(*ast.FunctionNode))
	case *ast.ObjectNode:
		return objectEqual(aNode, b.(*ast.ObjectNode))
	case *ast.ArrayNode:
		return arrayEqual(aNode, b.(*ast.ArrayNode))
	default:
		return false
	}
}

func literalEqual(a, b *ast.LiteralNode) bool {
	// Compare values (position ignored)
	aVal := a.Value()
	bVal := b.Value()

	// Handle nil values
	if aVal == nil && bVal == nil {
		return true
	}
	if aVal == nil || bVal == nil {
		return false
	}

	// Compare values using type assertion
	return aVal == bVal
}

func typeEqual(a, b *ast.TypeNode) bool {
	// Compare type names (position ignored)
	return a.TypeName() == b.TypeName()
}

func functionEqual(a, b *ast.FunctionNode) bool {
	// Compare function names
	if a.Name() != b.Name() {
		return false
	}

	// Compare argument counts
	aArgs := a.Arguments()
	bArgs := b.Arguments()
	if len(aArgs) != len(bArgs) {
		return false
	}

	// Compare each argument
	for i := range aArgs {
		if !interfaceEqual(aArgs[i], bArgs[i]) {
			return false
		}
	}

	return true
}

func objectEqual(a, b *ast.ObjectNode) bool {
	// Compare property counts
	aProps := a.Properties()
	bProps := b.Properties()
	if len(aProps) != len(bProps) {
		return false
	}

	// Compare each property (recursively)
	for key, aValue := range aProps {
		bValue, exists := bProps[key]
		if !exists {
			return false
		}
		if !ASTEqual(aValue, bValue) {
			return false
		}
	}

	return true
}

func arrayEqual(a, b *ast.ArrayNode) bool {
	// Compare element schemas (recursively)
	return ASTEqual(a.ElementSchema(), b.ElementSchema())
}

// interfaceEqual compares two interface{} values (for function arguments)
func interfaceEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a == b
}

// ASTDiff returns a human-readable description of differences between two ASTs.
//
// If the ASTs are equal, returns an empty string.
// Otherwise, returns a description of the first difference found.
//
// Example:
//
//	diff := grammar.ASTDiff(refAST, prodAST)
//	if diff != "" {
//	    t.Errorf("AST difference: %s", diff)
//	}
func ASTDiff(a, b ast.SchemaNode) string {
	if a == nil && b == nil {
		return ""
	}
	if a == nil {
		return "first AST is nil, second is not"
	}
	if b == nil {
		return "second AST is nil, first is not"
	}

	if a.Type() != b.Type() {
		return fmt.Sprintf("node types differ: %s vs %s", a.Type(), b.Type())
	}

	switch aNode := a.(type) {
	case *ast.LiteralNode:
		bNode := b.(*ast.LiteralNode)
		if !interfaceEqual(aNode.Value(), bNode.Value()) {
			return "literal values differ"
		}

	case *ast.TypeNode:
		bNode := b.(*ast.TypeNode)
		if aNode.TypeName() != bNode.TypeName() {
			return "type names differ: " + aNode.TypeName() + " vs " + bNode.TypeName()
		}

	case *ast.FunctionNode:
		bNode := b.(*ast.FunctionNode)
		if aNode.Name() != bNode.Name() {
			return "function names differ: " + aNode.Name() + " vs " + bNode.Name()
		}
		aArgs := aNode.Arguments()
		bArgs := bNode.Arguments()
		if len(aArgs) != len(bArgs) {
			return "function argument counts differ"
		}
		for i := range aArgs {
			if !interfaceEqual(aArgs[i], bArgs[i]) {
				return "in function argument: arguments differ"
			}
		}

	case *ast.ObjectNode:
		bNode := b.(*ast.ObjectNode)
		aProps := aNode.Properties()
		bProps := bNode.Properties()
		if len(aProps) != len(bProps) {
			return "object property counts differ"
		}
		for key, aValue := range aProps {
			bValue, exists := bProps[key]
			if !exists {
				return "property missing in second object: " + key
			}
			if diff := ASTDiff(aValue, bValue); diff != "" {
				return "in property " + key + ": " + diff
			}
		}

	case *ast.ArrayNode:
		bNode := b.(*ast.ArrayNode)
		if diff := ASTDiff(aNode.ElementSchema(), bNode.ElementSchema()); diff != "" {
			return "in array element schema: " + diff
		}
	}

	return ""
}
