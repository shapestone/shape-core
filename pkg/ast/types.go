package ast

// NodeType represents the type of an AST node.
type NodeType int

const (
	// NodeTypeLiteral represents a literal value (string, number, bool, null)
	NodeTypeLiteral NodeType = iota

	// NodeTypeType represents a type identifier (UUID, Email, ISO-8601, etc.)
	NodeTypeType

	// NodeTypeFunction represents a function call (Integer(1, 100), String(5+), etc.)
	NodeTypeFunction

	// NodeTypeObject represents an object with properties
	NodeTypeObject

	// NodeTypeArray represents an array with element schema
	NodeTypeArray

	// NodeTypeArrayData represents actual array data with elements
	NodeTypeArrayData
)

// String returns the string representation of the node type.
func (nt NodeType) String() string {
	switch nt {
	case NodeTypeLiteral:
		return "Literal"
	case NodeTypeType:
		return "Type"
	case NodeTypeFunction:
		return "Function"
	case NodeTypeObject:
		return "Object"
	case NodeTypeArray:
		return "Array"
	case NodeTypeArrayData:
		return "ArrayData"
	default:
		return "Unknown"
	}
}
