# ADR 0001: Schema AST Design

**Status:** Accepted  
**Date:** 2025-10-09

## Context

Shape needs a unified Abstract Syntax Tree (AST) representation for schemas across multiple data formats. The AST must be:

1. **Format Agnostic:** Same AST structure regardless of input format
2. **Complete:** Represent all schema constructs
3. **Traversable:** Easy to walk and process
4. **Serializable:** Can be saved/loaded for caching
5. **Extensible:** Support future node types

## Decision

We will use an **interface-based AST** with 5 core node types:

```go
type SchemaNode interface {
    Type() NodeType
    Accept(visitor Visitor) error
    String() string
    Position() Position
}
```

**Node Types:**
1. **LiteralNode** - Exact value match (`"active"`, `42`, `true`, `null`)
2. **TypeNode** - Type identifier (`UUID`, `Email`, `ISO-8601`)
3. **FunctionNode** - Function call (`Integer(1, 100)`, `String(1+)`)
4. **ObjectNode** - Object with properties (`{"id": UUID}`)
5. **ArrayNode** - Array with element schema (`[String(1,50)]`)

## Rationale

### Why Interface-Based?

**Polymorphism:**
- Enables visitor pattern for traversal
- Common operations across all node types
- Easy to add new node types without breaking existing code

**Type Safety:**
- Go's type system ensures correctness
- Compiler catches type mismatches
- Clear type assertions when needed

### Why These 5 Node Types?

**Complete Coverage:**
- LiteralNode: Covers all JSON/XML/etc. literals
- TypeNode: Built-in type validators
- FunctionNode: Parameterized validators
- ObjectNode: Structured data
- ArrayNode: Collections

**Minimal Set:**
- No redundancy (each node type has unique purpose)
- Composable (complex schemas built from simple nodes)
- Easy to understand and implement

### Why Immutable Nodes?

**Correctness:**
- No accidental mutation during traversal
- Thread-safe (can be shared across goroutines)
- Easier to reason about

**Caching:**
- Parsed schemas can be safely cached
- No need for deep copies

### Why Position Tracking?

**Error Messages:**
- "Error at line 5, column 12" is much better than "parse error"
- Essential for schema authors to fix issues quickly

**Tooling:**
- IDEs can highlight errors at exact position
- Language servers can provide diagnostics

## Alternatives Considered

### 1. JSON-Based AST (Dynamic)

```go
type Node map[string]interface{}
```

**Pros:**
- Very flexible
- Easy to serialize (already JSON)

**Cons:**
- No type safety
- Runtime errors instead of compile-time
- Hard to document API
- No visitor pattern support

**Rejected:** Type safety is critical for library quality

### 2. Struct-Based AST (No Interface)

```go
type Node struct {
    Type       NodeType
    Literal    *LiteralValue
    Function   *FunctionCall
    // ... all possibilities
}
```

**Pros:**
- Simple to implement
- Single type to pass around

**Cons:**
- Memory waste (most fields nil)
- Confusing API (which fields are valid?)
- Hard to extend with new types
- No polymorphism benefits

**Rejected:** Interface-based design is cleaner

### 3. Sum Type (Tagged Union)

```go
type Node interface{ isNode() }
type Literal struct{ ... }
func (Literal) isNode() {}
// ... etc
```

**Pros:**
- Type-safe union type pattern
- Exhaustive pattern matching with type switches

**Cons:**
- More boilerplate
- Same benefits as interface-based approach
- Less idiomatic in Go

**Rejected:** Interface approach is more Go-idiomatic

## Implementation Notes

### Node Construction

Provide constructors for all node types:

```go
func NewLiteralNode(value interface{}) *LiteralNode
func NewTypeNode(typeName string) *TypeNode
func NewFunctionNode(name string, args []interface{}) *FunctionNode
func NewObjectNode(properties map[string]SchemaNode) *ObjectNode
func NewArrayNode(elementSchema SchemaNode) *ArrayNode
```

### Visitor Pattern

Enable AST traversal:

```go
type Visitor interface {
    VisitLiteral(*LiteralNode) error
    VisitType(*TypeNode) error
    VisitFunction(*FunctionNode) error
    VisitObject(*ObjectNode) error
    VisitArray(*ArrayNode) error
}

func (n *ObjectNode) Accept(v Visitor) error {
    if err := v.VisitObject(n); err != nil {
        return err
    }
    for _, child := range n.Properties {
        if err := child.Accept(v); err != nil {
            return err
        }
    }
    return nil
}
```

### Serialization

Implement MarshalJSON/UnmarshalJSON:

```go
func (n *ObjectNode) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        Type       string                  `json:"type"`
        Properties map[string]SchemaNode   `json:"properties"`
    }{
        Type:       "object",
        Properties: n.properties,
    })
}
```

## Consequences

### Positive

- **Type Safety:** Compile-time guarantees for correctness
- **Clear API:** Easy to understand node types and usage
- **Extensible:** Can add new node types without breaking changes
- **Traversable:** Visitor pattern enables flexible AST walking
- **Serializable:** Can save/load parsed schemas
- **Testable:** Each node type can be tested in isolation

### Negative

- **Boilerplate:** Need implementations for each node type
- **Verbosity:** More code than dynamic approach
- **Learning Curve:** Need to understand visitor pattern

### Neutral

- **Memory Usage:** Reasonable overhead for node metadata
- **Performance:** Visitor pattern has small overhead vs direct access

## Success Metrics

- AST can represent all schema constructs from multiple formats
- Visitor pattern enables AST traversal and processing
- Serialization round-trips correctly (Parse → JSON → Parse → same AST)
- Position tracking enables helpful error messages
- New node types can be added in minor versions (backward compatible)

## References

- Visitor Pattern: https://en.wikipedia.org/wiki/Visitor_pattern
- Go interfaces: https://go.dev/tour/methods/9
- AST design: https://en.wikipedia.org/wiki/Abstract_syntax_tree

## Date

2025-10-09
