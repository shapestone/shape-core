package ast

import (
	"fmt"
	"strings"
)

// FunctionNode represents function-based validation with arguments.
// Examples: Integer(1, 100), String(5+), Enum("M", "F", "O")
type FunctionNode struct {
	name      string        // Function name (Integer, String, Enum, etc.)
	arguments []interface{} // Arguments (literals or special symbols like "+")
	position  Position
}

// NewFunctionNode creates a new function node.
func NewFunctionNode(name string, arguments []interface{}, pos Position) *FunctionNode {
	return &FunctionNode{
		name:      name,
		arguments: arguments,
		position:  pos,
	}
}

// Type returns NodeTypeFunction.
func (n *FunctionNode) Type() NodeType {
	return NodeTypeFunction
}

// Name returns the function name.
func (n *FunctionNode) Name() string {
	return n.name
}

// Arguments returns the function arguments.
func (n *FunctionNode) Arguments() []interface{} {
	return n.arguments
}

// Position returns the source position.
func (n *FunctionNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *FunctionNode) Accept(visitor Visitor) error {
	return visitor.VisitFunction(n)
}

// String returns a string representation.
func (n *FunctionNode) String() string {
	args := make([]string, len(n.arguments))
	for i, arg := range n.arguments {
		switch v := arg.(type) {
		case string:
			// Check if it's a special symbol
			if v == "+" {
				args[i] = v
			} else {
				args[i] = fmt.Sprintf("%q", v)
			}
		case nil:
			args[i] = "null"
		case bool:
			args[i] = fmt.Sprintf("%t", v)
		case int64:
			args[i] = fmt.Sprintf("%d", v)
		case float64:
			args[i] = fmt.Sprintf("%g", v)
		default:
			args[i] = fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%s(%s)", n.name, strings.Join(args, ", "))
}
