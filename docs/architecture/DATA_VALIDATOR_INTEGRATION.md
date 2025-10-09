# Data-Validator Integration Guide

**Version:** 1.0  
**Date:** 2025-10-09  
**Audience:** data-validator developers

## Overview

This document describes how data-validator integrates with the shape parser library. Shape handles schema parsing (text → AST), while data-validator handles validation (AST + data → results).

## Architecture Changes

### Before shape (Monolithic)

```
data-validator/
├── internal/
│   ├── parser/           # Schema parsing (JSONV, XMLV, etc.)
│   │   ├── jsonv/
│   │   ├── xmlv/
│   │   └── schema_model/ # Internal AST
│   └── traverser/        # Validation logic
```

**Problems:**
- Tight coupling between parsing and validation
- Cannot reuse parsers in other projects
- Difficult to version parsers independently
- Large codebase

### After shape (Decoupled)

```
data-validator/ (depends on shape only)
└── Traverse AST + validate data

shape/ (self-contained)
├── Parse schemas → AST
└── Embedded tokenization (internal/tokenizer/)
```

**Benefits:**
- Clear separation of concerns
- Shape can be used by other projects
- Independent versioning
- Smaller, focused codebases
- Easier to test

## Dependency Management

### data-validator go.mod

```go
module github.com/shapestone/data-validator

go 1.25

require (
    github.com/shapestone/shape v0.1.0  // Self-contained parser (no df2-go dependency)
    github.com/shapestone/wire v0.9.0   // Wire expression engine
)
```

**Note:** Shape v0.1.0+ includes embedded tokenization. No df2-go dependency required.

### Version Pinning

**Recommendation:** Pin shape to specific version

```go
require (
    github.com/shapestone/shape v0.1.0  // Exact version
)
```

**Rationale:**
- shape v0.x.x may have breaking changes
- Controlled upgrades
- Predictable behavior
- Shape is self-contained with no transitive dependencies (except google/uuid)

**Future (v1.0.0+):** Can use version ranges

```go
require (
    github.com/shapestone/shape v1.2.0  // Semantic versioning
)
```

## Updated data-validator Architecture

### New Directory Structure

```
data-validator/
├── pkg/
│   └── validator/
│       ├── validator.go       # Public validation API
│       └── validator_test.go
│
├── internal/
│   ├── traverser/             # AST traversal + validation
│   │   ├── traverser.go       # Main traverser (uses shape AST)
│   │   ├── literal.go         # Validate LiteralNode
│   │   ├── type.go            # Validate TypeNode
│   │   ├── function.go        # Validate FunctionNode (uses wire)
│   │   ├── object.go          # Validate ObjectNode
│   │   ├── array.go           # Validate ArrayNode
│   │   └── traverser_test.go
│   │
│   └── wire_integration/      # Wire expression evaluation
│       ├── evaluator.go       # Calls wire for function validation
│       └── evaluator_test.go
│
└── go.mod
```

### Removed Components

**Deleted (moved to shape):**
- `internal/parser/` - All format parsers
- `internal/schema/` - Schema AST model

**Result:** Significantly smaller codebase

## Integration Patterns

### Pattern 1: Parse Then Validate

```go
package validator

import (
    "github.com/shapestone/shape/pkg/shape"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
)

// Validate parses schema and validates data
func Validate(schemaInput string, data interface{}) error {
    // Step 1: Parse schema using shape
    schemaAST, err := shape.Parse(parser.FormatJSONV, schemaInput)
    if err != nil {
        return fmt.Errorf("schema parse error: %w", err)
    }
    
    // Step 2: Validate data against AST
    return ValidateWithAST(schemaAST, data)
}

// ValidateWithAST validates data against pre-parsed AST
func ValidateWithAST(schemaAST ast.SchemaNode, data interface{}) error {
    traverser := NewTraverser(schemaAST, data)
    return traverser.Validate()
}
```

### Pattern 2: Cached Schema

```go
package validator

// Validator caches parsed schema
type Validator struct {
    schemaAST ast.SchemaNode
}

// NewValidator creates validator with cached schema
func NewValidator(schemaInput string) (*Validator, error) {
    schemaAST, err := shape.Parse(parser.FormatJSONV, schemaInput)
    if err != nil {
        return nil, err
    }
    
    return &Validator{schemaAST: schemaAST}, nil
}

// Validate validates data (reuses cached AST)
func (v *Validator) Validate(data interface{}) error {
    return ValidateWithAST(v.schemaAST, data)
}
```

**Benefits:**
- Parse schema once, validate many times
- Performance optimization for repeated validation
- thread-safe (AST is immutable)

### Pattern 3: Auto-Detect Format

```go
func ValidateAuto(schemaInput string, data interface{}) error {
    // Auto-detect and parse
    schemaAST, format, err := shape.ParseAuto(schemaInput)
    if err != nil {
        return fmt.Errorf("schema parse error: %w", err)
    }
    
    log.Printf("Detected format: %s", format)
    
    // Validate
    return ValidateWithAST(schemaAST, data)
}
```

## Traverser Implementation

### Traverser Structure

```go
package traverser

import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/wire/engine"
)

type Traverser struct {
    schemaAST ast.SchemaNode
    data      interface{}
    errors    []ValidationError
    path      []string  // Current path in data (for error messages)
}

func NewTraverser(schemaAST ast.SchemaNode, data interface{}) *Traverser {
    return &Traverser{
        schemaAST: schemaAST,
        data:      data,
        errors:    []ValidationError{},
        path:      []string{},
    }
}

func (t *Traverser) Validate() error {
    // Traverse AST using visitor pattern
    visitor := &ValidationVisitor{
        traverser: t,
    }
    
    if err := t.schemaAST.Accept(visitor); err != nil {
        return err
    }
    
    if len(t.errors) > 0 {
        return &ValidationErrors{Errors: t.errors}
    }
    
    return nil
}
```

### Validation Visitor

```go
package traverser

import "github.com/shapestone/shape/pkg/ast"

type ValidationVisitor struct {
    traverser *Traverser
}

func (v *ValidationVisitor) VisitLiteral(n *ast.LiteralNode) error {
    // Validate exact match
    return validateLiteral(v.traverser, n)
}

func (v *ValidationVisitor) VisitType(n *ast.TypeNode) error {
    // Validate type (UUID, Email, etc.)
    return validateType(v.traverser, n)
}

func (v *ValidationVisitor) VisitFunction(n *ast.FunctionNode) error {
    // Validate with function constraints (uses wire)
    return validateFunction(v.traverser, n)
}

func (v *ValidationVisitor) VisitObject(n *ast.ObjectNode) error {
    // Validate object properties
    return validateObject(v.traverser, n)
}

func (v *ValidationVisitor) VisitArray(n *ast.ArrayNode) error {
    // Validate array elements
    return validateArray(v.traverser, n)
}
```

## Node-Specific Validation

### LiteralNode Validation

```go
func validateLiteral(t *Traverser, node *ast.LiteralNode) error {
    expected := node.Value()
    actual := t.data
    
    if !deepEqual(expected, actual) {
        return t.addError(fmt.Sprintf(
            "expected literal %v, got %v",
            expected, actual,
        ))
    }
    
    return nil
}
```

### TypeNode Validation

```go
func validateType(t *Traverser, node *ast.TypeNode) error {
    typeName := node.TypeName()
    
    validator, ok := typeValidators[typeName]
    if !ok {
        return fmt.Errorf("unknown type: %s", typeName)
    }
    
    if !validator(t.data) {
        return t.addError(fmt.Sprintf(
            "value does not match type %s",
            typeName,
        ))
    }
    
    return nil
}

// Built-in type validators
var typeValidators = map[string]func(interface{}) bool{
    "UUID":     validateUUID,
    "Email":    validateEmail,
    "ISO-8601": validateISO8601,
    // ... more types
}
```

### FunctionNode Validation (Wire Integration)

```go
func validateFunction(t *Traverser, node *ast.FunctionNode) error {
    funcName := node.Name()
    args := node.Arguments()
    
    // Get function validator
    validator, ok := functionValidators[funcName]
    if !ok {
        return fmt.Errorf("unknown function: %s", funcName)
    }
    
    // Call validator with wire if needed
    return validator(t.data, args, t.wireEngine)
}

// Example: Integer(min, max) validation
func validateInteger(value interface{}, args []interface{}, wire *engine.Engine) error {
    // Extract constraints
    min := args[0].(int64)
    max := args[1].(int64)
    
    // Convert value to int64
    intVal, ok := value.(int64)
    if !ok {
        return fmt.Errorf("expected integer, got %T", value)
    }
    
    // Validate range
    if intVal < min || intVal > max {
        return fmt.Errorf("integer %d outside range [%d, %d]", intVal, min, max)
    }
    
    return nil
}
```

### ObjectNode Validation

```go
func validateObject(t *Traverser, node *ast.ObjectNode) error {
    // Ensure data is object/map
    dataMap, ok := t.data.(map[string]interface{})
    if !ok {
        return t.addError("expected object")
    }
    
    // Validate all schema properties exist in data
    for propName, propSchema := range node.Properties() {
        propData, ok := dataMap[propName]
        if !ok {
            return t.addError(fmt.Sprintf("missing property: %s", propName))
        }
        
        // Recursively validate property
        t.pushPath(propName)
        subTraverser := NewTraverser(propSchema, propData)
        err := subTraverser.Validate()
        t.popPath()
        
        if err != nil {
            return err
        }
    }
    
    // Reject extra properties (strict mode)
    for dataKey := range dataMap {
        if _, ok := node.Properties()[dataKey]; !ok {
            return t.addError(fmt.Sprintf("unexpected property: %s", dataKey))
        }
    }
    
    return nil
}
```

### ArrayNode Validation

```go
func validateArray(t *Traverser, node *ast.ArrayNode) error {
    // Ensure data is array/slice
    dataArray, ok := t.data.([]interface{})
    if !ok {
        return t.addError("expected array")
    }
    
    // Get element schema
    elementSchema := node.ElementSchema()
    if elementSchema == nil {
        // Empty array schema [] matches any array
        return nil
    }
    
    // Validate each element against element schema
    for i, elem := range dataArray {
        t.pushPath(fmt.Sprintf("[%d]", i))
        subTraverser := NewTraverser(elementSchema, elem)
        err := subTraverser.Validate()
        t.popPath()
        
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## Migration Steps

### Step 1: Add shape Dependency (1 hour)

```bash
cd data-validator
go get github.com/shapestone/shape@v0.1.0
go mod tidy
```

### Step 2: Update Imports (1 hour)

Replace internal schema types with shape types:

```go
// Old
import "github.com/shapestone/data-validator/internal/schema"

// New
import "github.com/shapestone/shape/pkg/ast"
```

### Step 3: Remove Old Parser Code (1 hour)

```bash
# Delete old parsers (moved to shape)
rm -rf internal/parser/jsonv
rm -rf internal/parser/xmlv
rm -rf internal/parser/schema

# Keep only traverser and wire integration
```

### Step 4: Update Traverser (4-6 hours)

Update traverser to use shape's AST:

```go
// Old
import "github.com/shapestone/data-validator/internal/schema"

func (t *Traverser) visitLiteral(n *schema.LiteralNode) error {
    // ...
}

// New
import "github.com/shapestone/shape/pkg/ast"

func (t *Traverser) VisitLiteral(n *ast.LiteralNode) error {
    // ...
}
```

**Key Changes:**
- Update node type imports
- Update visitor method signatures
- Use shape's node accessors (`.Value()`, `.TypeName()`, etc.)

### Step 5: Update Public API (2-3 hours)

```go
package validator

import (
    "github.com/shapestone/shape"
    "github.com/shapestone/shape/internal/parser"
)

// Old
func Validate(schemaInput string, data interface{}) error {
    ast, err := parseJSONV(schemaInput)  // Internal parser
    // ...
}

// New
func Validate(schemaInput string, data interface{}) error {
    ast, err := shape.Parse(parser.FormatJSONV, schemaInput)  // Shape parser
    // ...
}
```

### Step 6: Update Tests (4-6 hours)

Update all tests to use shape:

```go
// Old
ast := schema.NewObjectNode(...)

// New
import "github.com/shapestone/shape/pkg/ast"
astNode := ast.NewObjectNode(...)
```

### Step 7: Update Documentation (2-3 hours)

- Update README to mention shape dependency
- Update architecture docs
- Update examples

### Step 8: Integration Testing (2-4 hours)

```go
func TestShapeIntegration(t *testing.T) {
    // Parse schema with shape
    schemaInput := `{"id": UUID, "age": Integer(1, 120)}`
    ast, err := shape.Parse(parser.FormatJSONV, schemaInput)
    require.NoError(t, err)
    
    // Validate data with data-validator
    data := map[string]interface{}{
        "id":  "550e8400-e29b-41d4-a716-446655440000",
        "age": 25,
    }
    
    err = ValidateWithAST(ast, data)
    require.NoError(t, err)
}
```

### Total Migration Effort: 18-28 hours

## Benefits After Migration

### For data-validator

**Smaller Codebase:**
- Remove ~5000-8000 lines of parser code
- Focus only on validation logic
- Easier to understand and maintain

**Clear Separation:**
- Shape: Parse schemas
- data-validator: Validate data
- wire: Evaluate expressions

**Easier Testing:**
- Mock AST for validation tests
- No need to test parsing in validation tests

### For shape

**Reusable Parsers:**
- Other projects can use shape
- Schema parsers become shared infrastructure

**Independent Evolution:**
- Shape can add formats without affecting data-validator
- data-validator can improve validation without touching parsers

**Focused Development:**
- Each repo has single responsibility
- Easier to reason about changes

## Compatibility Matrix

| data-validator | shape | wire | Status |
|----------------|-------|------|--------|
| v0.1.0 | v0.1.0 | v0.9.0 | Compatible |
| v0.2.0 | v0.1.0 - v0.2.0 | v0.9.0 | Compatible |
| v1.0.0 | v1.x.x | v1.0.0 | Stable |

## API Examples

### Basic Validation

```go
import (
    "github.com/shapestone/shape"
    "github.com/shapestone/data-validator/pkg/validator"
)

func main() {
    schema := `{"id": UUID, "name": String(1, 100)}`
    data := map[string]interface{}{
        "id":   "550e8400-e29b-41d4-a716-446655440000",
        "name": "John Doe",
    }
    
    err := validator.Validate(schema, data)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Cached Validator

```go
// Parse schema once
v, err := validator.NewValidator(schemaInput)
if err != nil {
    log.Fatal(err)
}

// Validate many times
for _, data := range dataSet {
    if err := v.Validate(data); err != nil {
        log.Printf("Validation failed: %v", err)
    }
}
```

### Multi-Format Support

```go
// Auto-detect format
err := validator.ValidateAuto(schemaInput, data)
if err != nil {
    log.Fatal(err)
}

// Or explicit format
ast, _ := shape.Parse(parser.FormatXMLV, xmlSchema)
err = validator.ValidateWithAST(ast, data)
```

## Testing Strategy

### Unit Tests

Test validation logic independently:

```go
func TestValidateLiteral(t *testing.T) {
    // Create AST programmatically (no parsing)
    astNode := ast.NewLiteralNode("active")
    
    // Validate
    err := validator.ValidateWithAST(astNode, "active")
    assert.NoError(t, err)
    
    err = validator.ValidateWithAST(astNode, "inactive")
    assert.Error(t, err)
}
```

### Integration Tests

Test shape + data-validator together:

```go
func TestEndToEnd(t *testing.T) {
    schema := `{"status": Enum("active", "inactive")}`
    
    // Valid data
    err := validator.Validate(schema, map[string]interface{}{
        "status": "active",
    })
    assert.NoError(t, err)
    
    // Invalid data
    err = validator.Validate(schema, map[string]interface{}{
        "status": "pending",
    })
    assert.Error(t, err)
}
```

## Performance Considerations

### Parse Once, Validate Many

```go
// Slow (parses every time)
for _, data := range largeDataset {
    validator.Validate(schemaInput, data)  // Parses schema N times
}

// Fast (parse once)
v, _ := validator.NewValidator(schemaInput)  // Parse once
for _, data := range largeDataset {
    v.Validate(data)  // Reuse parsed AST
}
```

**Performance Gain:** 10-100x faster for large datasets

### Concurrent Validation

Shape's AST is immutable and thread-safe:

```go
v, _ := validator.NewValidator(schemaInput)

var wg sync.WaitGroup
for _, data := range largeDataset {
    wg.Add(1)
    go func(d interface{}) {
        defer wg.Done()
        v.Validate(d)  // Safe: AST is immutable
    }(data)
}
wg.Wait()
```

## Troubleshooting

### Error: "cannot find package shape"

```bash
go get github.com/shapestone/shape@v0.1.0
go mod tidy
```

### Error: "SchemaNode does not implement interface"

Update to shape's AST types:

```go
// Old
import "github.com/shapestone/data-validator/internal/schema"

// New
import "github.com/shapestone/shape/pkg/ast"
```

### Error: "version conflict"

Pin shape version:

```go
// go.mod
require github.com/shapestone/shape v0.1.0
```

## Support

- **shape Issues:** https://github.com/shapestone/shape/issues
- **data-validator Issues:** https://github.com/shapestone/data-validator/issues
- **Migration Help:** Create issue in data-validator repo

---

**Document Status:** Complete  
**Next Steps:** Begin migration in data-validator Phase 3
