# Shape - Multi-Format Validation Schema Parser

**Version:** 0.1.0 (In Development)  
**Repository:** github.com/shapestone/shape

Shape is a production-ready parser library that converts validation schema formats (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV) into a unified Abstract Syntax Tree (AST). Shape serves as the foundational parsing layer for the data-validator ecosystem.

## Features

- **6 Format Support:** JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV
- **Unified AST:** All formats produce the same AST structure
- **Format Auto-Detection:** Automatically detect and parse formats
- **Schema Validation:** Validate schemas for unknown types, invalid functions, and constraint violations
- **Detailed Error Messages:** Line and column numbers for all parse errors
- **Self-Contained Library:** Zero external dependencies except google/uuid
- **Embedded Tokenization:** Built-in tokenization framework, no external tokenizer dependencies
- **Production-Ready:** Comprehensive error handling, battle-tested tokenization, 95%+ test coverage
- **UTF-8 Support:** International schemas supported

## Installation

```bash
go get github.com/shapestone/shape
```

## Quick Start

### Parse a JSONV Schema

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/shapestone/shape/pkg/shape"
    "github.com/shapestone/shape/internal/parser"
)

func main() {
    schemaInput := `{
        "id": UUID,
        "name": String(1, 100),
        "age": Integer(1, 120),
        "email": Email
    }`
    
    // Parse with explicit format
    ast, err := shape.Parse(parser.FormatJSONV, schemaInput)
    if err != nil {
        log.Fatal(err)
    }
    
    // Print AST
    fmt.Println(ast.String())
}
```

### Auto-Detect Format

```go
// Auto-detect and parse
ast, format, err := shape.ParseAuto(schemaInput)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Detected format: %s\n", format)
fmt.Println(ast.String())
```

### Validate Schema

```go
// Parse schema
ast, err := shape.Parse(parser.FormatJSONV, schemaInput)
if err != nil {
    log.Fatal(err)
}

// Validate for unknown types and invalid constraints
if err := shape.Validate(ast); err != nil {
    log.Printf("Validation error: %v", err)
}
```

### Walk the AST

```go
import "github.com/shapestone/shape/pkg/ast"

type MyVisitor struct{}

func (v *MyVisitor) VisitObject(n *ast.ObjectNode) error {
    fmt.Printf("Found object with %d properties\n", len(n.Properties))
    return nil
}

func (v *MyVisitor) VisitType(n *ast.TypeNode) error {
    fmt.Printf("Found type: %s\n", n.TypeName)
    return nil
}

// ... implement other visitor methods

visitor := &MyVisitor{}
ast.Accept(visitor)
```

## Supported Formats

### JSONV (JSON Validation)

```jsonv
{
    "user": {
        "id": UUID,
        "name": String(1, 100),
        "age": Integer(1, 120),
        "tags": [String(1, 30)]
    }
}
```

### XMLV (XML Validation)

```xmlv
<user>
    <id>UUID</id>
    <name>String(1, 100)</name>
    <age>Integer(1, 120)</age>
</user>
```

### PropsV (Properties Validation)

```propsv
user.id=UUID
user.name=String(1, 100)
user.age=Integer(1, 120)
```

### CSVV (CSV Validation)

```csvv
# Header row with validation
id,name,age,email
UUID,String(1,100),Integer(1,120),Email
```

### YAMLV (YAML Validation)

```yamlv
user:
  id: UUID
  name: String(1, 100)
  age: Integer(1, 120)
```

### TEXTV (Text Validation)

```textv
# Simple line-oriented format with dot notation
user.id: UUID
user.name: String(1, 100)
user.age: Integer(1, 120)
user.tags[]: String(1, 30)
```

## AST Structure

Shape produces a format-agnostic AST with 5 node types:

- **LiteralNode:** Exact value match (`"active"`, `42`, `true`, `null`)
- **TypeNode:** Type identifier (`UUID`, `Email`, `ISO-8601`)
- **FunctionNode:** Function call (`Integer(1, 100)`, `String(1+)`)
- **ObjectNode:** Object with properties
- **ArrayNode:** Array with element schema

### Example AST

Input JSONV:
```jsonv
{"id": UUID, "name": String(1, 100)}
```

Resulting AST:
```go
ObjectNode{
    Properties: {
        "id":   TypeNode{TypeName: "UUID"},
        "name": FunctionNode{Name: "String", Arguments: []interface{}{1, 100}},
    },
}
```

## API Reference

### Public API

```go
package shape

// Parse parses input with explicit format
func Parse(format parser.Format, input string) (ast.SchemaNode, error)

// ParseAuto auto-detects format and parses
func ParseAuto(input string) (ast.SchemaNode, parser.Format, error)

// MustParse parses or panics (for tests/initialization)
func MustParse(format parser.Format, input string) ast.SchemaNode

// Validate validates a schema AST against default validation rules
func Validate(node ast.SchemaNode) error
```

### AST Package

```go
package ast

// SchemaNode is the root interface for all AST nodes
type SchemaNode interface {
    Type() NodeType
    Accept(visitor Visitor) error
    String() string
    Position() Position
}

// Node constructors
func NewLiteralNode(value interface{}) *LiteralNode
func NewTypeNode(typeName string) *TypeNode
func NewFunctionNode(name string, args []interface{}) *FunctionNode
func NewObjectNode(properties map[string]SchemaNode) *ObjectNode
func NewArrayNode(elementSchema SchemaNode) *ArrayNode
```

### Visitor Pattern

```go
package ast

type Visitor interface {
    VisitLiteral(*LiteralNode) error
    VisitType(*TypeNode) error
    VisitFunction(*FunctionNode) error
    VisitObject(*ObjectNode) error
    VisitArray(*ArrayNode) error
}
```

## Error Handling

Shape provides detailed error messages with position information:

```go
ast, err := shape.Parse(parser.FormatJSONV, `{"id": UUID`)
if err != nil {
    fmt.Println(err)
    // Output: unexpected end of input at line 1, column 12: expected '}'
}
```

All errors include:
- Exact position (line and column)
- What was expected
- What was found
- Context from the input

## Schema Validation

Shape includes a built-in validator to catch schema errors before runtime:

```go
ast, err := shape.Parse(parser.FormatJSONV, `{"id": UnknownType}`)
if err != nil {
    log.Fatal(err)
}

if err := shape.Validate(ast); err != nil {
    fmt.Println(err)
    // Output: validation error at line 1, column 8: unknown type: UnknownType
}
```

### Supported Types

The validator recognizes these built-in types:
- `UUID`, `Email`, `URL`
- `String`, `Integer`, `Float`, `Boolean`
- `ISO-8601`, `Date`, `Time`, `DateTime`
- `IPv4`, `IPv6`
- `JSON`, `Base64`

### Supported Functions

The validator validates these built-in functions:
- `String(min, max)` - String length constraints
- `Integer(min, max)` - Integer range constraints
- `Float(min, max)` - Float range constraints
- `Enum(val1, val2, ...)` - Enumeration values
- `Pattern(regex)` - Regular expression pattern
- `Length(min, max)` - Generic length constraint
- `Range(min, max)` - Generic range constraint

All range functions support unbounded ranges with `+`:
```go
String(1, +)     // Minimum 1 character, no maximum
Integer(18, +)   // Minimum 18, no maximum
```

### Validation Errors

The validator catches:
- **Unknown types**: Types not in the built-in type list
- **Unknown functions**: Functions not in the built-in function list
- **Invalid arguments**: Wrong number or type of arguments
- **Invalid ranges**: `min > max` in range constraints
- **Nested errors**: Errors in object properties and array elements

## Performance

Shape is designed for speed (benchmarked on Apple M1 Max):

- **Simple schema** (2 properties): 0.7-4.8µs (YAMLV fastest, JSONV slowest)
- **Medium schema** (nested, 7 properties): 2.7-20.6µs (YAMLV fastest, JSONV slowest)
- **Large schema** (deep nesting, 25 properties): 8.9-70µs (YAMLV fastest, JSONV slowest)

**Format Performance Ranking** (fastest to slowest):
1. **YAMLV** (6.8-7.9x faster than JSONV!) - Native parser in v0.2.0
2. CSVV (2.4-3.2x faster than JSONV)
3. XMLV, PropsV, TEXTV (mid-range, similar performance)
4. JSONV (most allocations, slowest)

**v0.2.0 Performance Improvements:**
- YAMLV now 5-6x faster with native line-based parser
- 3-5x memory reduction for YAMLV
- AST string interning reduces allocations

See [docs/BENCHMARKS.md](docs/BENCHMARKS.md) for detailed benchmark results and analysis.

Run benchmarks:
```bash
go test -bench=. -benchmem ./pkg/shape
```

## Testing

Shape has comprehensive test coverage (95%+):

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run linters
make lint
```

## Documentation

- **[Architecture](docs/architecture/ARCHITECTURE.md):** System design and components
- **[Implementation Roadmap](docs/architecture/IMPLEMENTATION_ROADMAP.md):** 4-week implementation plan
- **[Format Specifications](docs/architecture/specifications/):** Detailed format specs
- **[ADRs](docs/architecture/decisions/):** Architecture decision records
- **[API Reference](https://pkg.go.dev/github.com/shapestone/shape):** Complete API documentation

## Examples

See [examples/](examples/) for working code examples:

- [Basic Usage](examples/basic/main.go)
- [Advanced Usage](examples/advanced/main.go)
- [Multi-Format](examples/multi-format/main.go)

## Integration with data-validator

Shape is designed to work seamlessly with data-validator:

```go
import (
    "github.com/shapestone/shape"
    "github.com/shapestone/data-validator/pkg/validator"
)

// Parse schema
schemaAST, err := shape.Parse(parser.FormatJSONV, schemaInput)
if err != nil {
    log.Fatal(err)
}

// Validate data
err = validator.ValidateWithAST(schemaAST, data)
if err != nil {
    log.Fatal(err)
}
```

See [data-validator integration guide](docs/architecture/DATA_VALIDATOR_INTEGRATION.md) for details.

## Contributing

We welcome contributions! See:

- [Local Setup](docs/contributor/local-setup.md)
- [Contributing Guide](docs/contributor/contributing.md)
- [Testing Guide](docs/contributor/testing-guide.md)

## Versioning

Shape follows [Semantic Versioning](https://semver.org/):

- **v0.x.x:** Development, API may change
- **v1.x.x:** Stable, backward compatibility guaranteed

See [CHANGELOG.md](CHANGELOG.md) for version history.

## License

[To Be Determined]

## Related Projects

- **[df2-go](https://github.com/shapestone/df2-go):** Tokenization framework code embedded in shape's `internal/tokenizer/` for self-contained operation
- **[wire](https://github.com/shapestone/wire):** Expression evaluation engine
- **[data-validator](https://github.com/shapestone/data-validator):** Data validation using shape

## Roadmap

### v0.1.0 (Released)
- All 6 format parsers
- Unified AST
- Format auto-detection (JSONV only)
- Comprehensive testing

### v0.2.0 (In Development)
- Format auto-detection for all 6 formats ✅
- Replace YAMLV yaml.v3 with native parser ✅
- Schema validation ✅
- AST optimization (string interning) ✅
- Custom validator registration

### v1.0.0 (Future)
- Stable API
- Production battle-testing
- Performance optimizations

## Support

- **Issues:** [GitHub Issues](https://github.com/shapestone/shape/issues)
- **Documentation:** [docs/](docs/)
- **Examples:** [examples/](examples/)

---

Built with ❤️ by the Shapestone team
