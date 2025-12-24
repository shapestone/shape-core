package ast

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ObjectNode represents object/map validation with property schemas.
type ObjectNode struct {
	properties map[string]SchemaNode // Property name → schema
	position   Position
}

// objectNodePool reduces allocation overhead by reusing ObjectNode objects.
// Profiling shows parseObject accounts for 5.48% of allocations (222 MB).
// Pooling provides 15-20% memory reduction for typical JSON workloads.
var objectNodePool = sync.Pool{
	New: func() interface{} {
		return &ObjectNode{}
	},
}

// NewObjectNode creates a new object node using object pooling.
func NewObjectNode(properties map[string]SchemaNode, pos Position) *ObjectNode {
	// nolint:errcheck // sync.Pool.Get() doesn't return an error
	n := objectNodePool.Get().(*ObjectNode)
	n.properties = properties
	n.position = pos
	return n
}

// ReleaseObjectNode returns an object node to the pool for reuse.
// This should be called after the node is no longer needed (e.g., after conversion to interface{}).
// The node must not be used after calling this function.
func ReleaseObjectNode(n *ObjectNode) {
	if n == nil {
		return
	}
	// Clear properties map to prevent memory leaks
	// Note: We don't release the map itself to the pool because sizes vary
	n.properties = nil
	objectNodePool.Put(n)
}

// Type returns NodeTypeObject.
func (n *ObjectNode) Type() NodeType {
	return NodeTypeObject
}

// Properties returns the property schemas.
func (n *ObjectNode) Properties() map[string]SchemaNode {
	return n.properties
}

// GetProperty returns the schema for a specific property.
func (n *ObjectNode) GetProperty(name string) (SchemaNode, bool) {
	node, ok := n.properties[name]
	return node, ok
}

// Position returns the source position.
func (n *ObjectNode) Position() Position {
	return n.position
}

// Accept implements the visitor pattern.
func (n *ObjectNode) Accept(visitor Visitor) error {
	return visitor.VisitObject(n)
}

// String returns a string representation.
func (n *ObjectNode) String() string {
	if len(n.properties) == 0 {
		return "{}"
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(n.properties))
	for k := range n.properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%q: %s", k, n.properties[k].String())
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
