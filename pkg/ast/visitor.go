package ast

// Visitor is the interface for traversing the AST using the visitor pattern.
type Visitor interface {
	VisitLiteral(node *LiteralNode) error
	VisitType(node *TypeNode) error
	VisitFunction(node *FunctionNode) error
	VisitObject(node *ObjectNode) error
	VisitArray(node *ArrayNode) error
}

// BaseVisitor provides default implementations for the Visitor interface.
// Embed this in your visitor to only override the methods you need.
type BaseVisitor struct{}

// VisitLiteral is the default implementation for visiting literal nodes.
func (v *BaseVisitor) VisitLiteral(node *LiteralNode) error {
	return nil
}

// VisitType is the default implementation for visiting type nodes.
func (v *BaseVisitor) VisitType(node *TypeNode) error {
	return nil
}

// VisitFunction is the default implementation for visiting function nodes.
func (v *BaseVisitor) VisitFunction(node *FunctionNode) error {
	return nil
}

// VisitObject is the default implementation for visiting object nodes.
func (v *BaseVisitor) VisitObject(node *ObjectNode) error {
	return nil
}

// VisitArray is the default implementation for visiting array nodes.
func (v *BaseVisitor) VisitArray(node *ArrayNode) error {
	return nil
}

// Walk traverses the AST starting from the given node using the provided visitor.
func Walk(node SchemaNode, visitor Visitor) error {
	return node.Accept(visitor)
}
