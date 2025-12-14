# Shape Parser Architecture

**Version:** 1.0
**Date:** 2025-10-09
**Status:** Architectural Design
**Repository:** github.com/shapestone/shape

## Executive Summary

Shape is a parser infrastructure library that provides foundational components for parsing structured data formats into a unified Abstract Syntax Tree (AST) representation. Shape is a general parser infrastructure that serves as the foundational parsing layer providing format-agnostic schema representations.

**Key Design Principles:**
- Single Responsibility: Shape parses schemas, doesn't validate data
- Format Agnostic: All formats produce the same AST structure
- Extensible: Easy to add new formats or custom validators
- Production Ready: Comprehensive error handling and testing
- Zero Breaking Changes: Semantic versioning with stable API

## 1. System Overview

### 1.1 Purpose and Scope

**What Shape Does:**
- Provides infrastructure for parsing validation schemas into unified AST
- Offers format detection and parser selection interfaces
- Reports detailed parse errors with position information
- Serializes/deserializes AST for storage or transmission

**What Shape Does NOT Do:**
- Validate actual data against schemas
- Execute validation expressions
- Manage validation state or results
- Provide web or application server interfaces

### 1.2 Ecosystem Position

For a comprehensive overview of how Shape fits into the Shapestone ecosystem, see [ECOSYSTEM.md](../../ECOSYSTEM.md).

```
┌──────────────────────────────────────────────────────────┐
│                 Downstream Projects                      │
│         (data-validator, custom validators, etc.)        │
│                                                          │
│  ┌──────────────────────────────────────────────────┐    │
│  │         Validation/Processing Logic              │    │
│  │  (Walks AST + performs domain-specific work)     │    │
│  └──────────────────────────────────────────────────┘    │
│                         ▲                                │
│                         │                                │
│                         │ uses AST                       │
│                         │                                │
└─────────────────────────┼────────────────────────────────┘
                          │
                          │
┌─────────────────────────┼────────────────────────────────┐
│                   Shape Parser (This Project)            │
│                         │                                │
│  ┌──────────────────────▼──────────────────────────┐     │
│  │              Schema AST Model                   │     │
│  │  (LiteralNode, TypeNode, FunctionNode, etc.)    │     │
│  └──────────────▲──────────────────────────────────┘     │
│                 │                                        │
│  ┌──────────────┴──────────────────────────────────┐     │
│  │           Format Parsers                        │     │
│  │  (Implemented in separate parser projects)      │     │
│  └──────────────┬──────────────────────────────────┘     │
│                 │                                        │
│  ┌──────────────▼──────────────────────────────────┐     │
│  │     Embedded Tokenization Framework             │     │
│  │     (internal/tokenizer/)                       │     │
│  │  Stream, Matchers, Position Tracking            │     │
│  └─────────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────────┘
```
## 2. Core Architecture

### 2.1 Layered Architecture

```
┌───────────────────────────────────────────────────────┐
│              Public API Layer                         │
│  - shape.Parse(format, input)                         │
│  - shape.ParseAuto(input)                             │
│  - shape.NewParser(format)                            │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│           Parser Abstraction Layer                    │
│  - Parser interface                                   │
│  - Parser factory/registry                            │
│  - Format detection                                   │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│         Format-Specific Parsers                       │
│         (Implemented in separate projects)            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   Parser    │  │   Parser    │  │   Parser    │    │
│  │  Project A  │  │  Project B  │  │  Project C  │    │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘    │
│         │                │                │           │
│  ┌──────▼────────────────▼────────────────▼──────┐    │
│  │      Format-Specific Tokenizers               │    │
│  │      (Uses integrated matcher framework)      │    │
│  └───────────────────────────────────────────────┘    │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│            Schema AST Model                           │
│  - SchemaNode interface                               │
│  - Node implementations (Object, Array, Function...)  │
│  - AST traversal utilities                            │
│  - Serialization/deserialization                      │
└───────────────────────────────────────────────────────┘
```

### 2.2 Component Responsibilities

#### Public API
- **Responsibility:** Simple, ergonomic interface for library consumers
- **Key Functions:**
  - `Parse(format, input) (SchemaNode, error)` - Parse with explicit format
  - `ParseAuto(input) (SchemaNode, Format, error)` - Auto-detect format
  - `NewParser(format) Parser` - Get format-specific parser instance
- **Location:** `pkg/shape/`

#### Parser Abstraction
- **Responsibility:** Common interface and factory for all parsers
- **Components:**
  - `Parser` interface - Common parse contract
  - `ParserFactory` - Creates format-specific parsers
  - Format detection logic
- **Location:** `internal/parser/`

#### Format-Specific Parsers
- **Responsibility:** Parse specific format into AST (implemented in parser projects)
- **Pattern:** Tokenizer (integrated framework) + Parser (various techniques)
- **Components per format:**
  - Tokenizer with format-specific matchers
  - Parser (LL(1), Pratt, PEG, etc. - technique chosen by parser project)
  - Error handling with position tracking
- **Location:** Separate parser project repositories (e.g., shape-json, shape-yaml)

#### Schema AST Model
- **Responsibility:** Format-agnostic representation of validation rules
- **Components:**
  - `SchemaNode` interface
  - Node implementations (LiteralNode, TypeNode, FunctionNode, etc.)
  - AST utilities (traversal, pretty-print, serialization)
- **Location:** `pkg/ast/`

## 3. Schema AST Design

### 3.1 AST Philosophy

The AST represents validation rules in a format-agnostic way:
- **Immutable:** Once created, nodes cannot be modified
- **Typed:** Strong type system with interface-based polymorphism
- **Serializable:** Can be marshaled to JSON/binary for storage
- **Traversable:** Visitor pattern for walking the tree
- **Printable:** Human-readable string representation for debugging

### 3.2 Node Hierarchy

```go
// SchemaNode is the root interface for all AST nodes
type SchemaNode interface {
    // Type returns the node type (literal, type, function, object, array)
    Type() NodeType
    
    // Accept allows visitor pattern traversal
    Accept(visitor Visitor) error
    
    // String returns human-readable representation
    String() string
    
    // Position returns source location (for error messages)
    Position() Position
}

// NodeType enum
type NodeType int

const (
    NodeTypeLiteral  NodeType = iota  // Literal value (string, number, bool, null)
    NodeTypeType                      // Type identifier (UUID, Email)
    NodeTypeFunction                  // Function call (Integer(1, 100))
    NodeTypeObject                    // Object with properties
    NodeTypeArray                     // Array with element schema
)
```

### 3.3 Node Types

#### LiteralNode
Represents exact match validation (literals from JSON/XML/etc.)

```go
type LiteralNode struct {
    value    interface{}  // string, int64, float64, bool, or nil
    position Position
}

// Examples:
// "active" → LiteralNode{value: "active"}
// 42       → LiteralNode{value: int64(42)}
// true     → LiteralNode{value: true}
// null     → LiteralNode{value: nil}
```

#### TypeNode
Represents type validation (built-in type identifiers)

```go
type TypeNode struct {
    typeName string      // "UUID", "Email", "ISO-8601", etc.
    position Position
}

// Examples:
// UUID     → TypeNode{typeName: "UUID"}
// Email    → TypeNode{typeName: "Email"}
// ISO-8601 → TypeNode{typeName: "ISO-8601"}
```

#### FunctionNode
Represents function-based validation with arguments

```go
type FunctionNode struct {
    name      string        // Function name (Integer, String, Enum)
    arguments []interface{} // Arguments (literals or special symbols like "+")
    position  Position
}

// Examples:
// Integer(1, 100)           → FunctionNode{name: "Integer", arguments: [1, 100]}
// String(1+)                → FunctionNode{name: "String", arguments: [1, "+"]}
// Enum("M", "F", "O")       → FunctionNode{name: "Enum", arguments: ["M", "F", "O"]}
```

#### ObjectNode
Represents object/map validation with property schemas

```go
type ObjectNode struct {
    properties map[string]SchemaNode  // Property name → schema
    position   Position
}

// Example:
// {"id": UUID, "name": String(1,100)}
// → ObjectNode{
//     properties: {
//       "id":   TypeNode{typeName: "UUID"},
//       "name": FunctionNode{name: "String", arguments: [1, 100]}
//     }
//   }
```

#### ArrayNode
Represents array validation with element schema

```go
type ArrayNode struct {
    elementSchema SchemaNode  // Schema for all array elements
    position      Position
}

// Example:
// [String(1, 50)]
// → ArrayNode{
//     elementSchema: FunctionNode{name: "String", arguments: [1, 50]}
//   }
```

### 3.4 Position Tracking

```go
type Position struct {
    Offset int  // Byte offset in source
    Line   int  // Line number (1-indexed)
    Column int  // Column number (1-indexed)
}

// Used for error messages:
// "Error at line 5, column 12: expected '}'"
```

## 4. Parser Interface Design

### 4.1 Core Parser Interface

```go
package parser

import "github.com/shapestone/shape-core/pkg/ast"

// Parser interface implemented by all format parsers
type Parser interface {
    // Parse converts input string to AST
    Parse(input string) (ast.SchemaNode, error)
    
    // Format returns the format this parser handles
    Format() Format
}

// Format enum
type Format int

const (
    FormatJSON  Format = iota
    FormatYAML
    FormatXML
    FormatCSV
    FormatProps
    // ... additional formats defined by parser projects
)

func (f Format) String() string {
    // Returns format name string
}
```

### 4.2 Parser Factory

```go
package parser

// ParserFactory creates parsers for specific formats
type ParserFactory struct {
    // private registry
}

// NewParser creates a parser for the specified format
func NewParser(format Format) (Parser, error) {
    // Returns format-specific parser or error if unsupported
}

// DetectFormat attempts to detect format from input
func DetectFormat(input string) (Format, error) {
    // Heuristic detection based on input characteristics
    // Implemented by parser projects
}
```

### 4.3 Public API

```go
package shape

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-core/internal/parser"
)

// Parse parses input with explicit format
func Parse(format parser.Format, input string) (ast.SchemaNode, error) {
    p, err := parser.NewParser(format)
    if err != nil {
        return nil, err
    }
    return p.Parse(input)
}

// ParseAuto auto-detects format and parses
func ParseAuto(input string) (ast.SchemaNode, parser.Format, error) {
    format, err := parser.DetectFormat(input)
    if err != nil {
        return nil, parser.FormatUnknown, err
    }
    node, err := Parse(format, input)
    return node, format, err
}

// MustParse parses or panics (for tests/initialization)
func MustParse(format parser.Format, input string) ast.SchemaNode {
    node, err := Parse(format, input)
    if err != nil {
        panic(err)
    }
    return node
}
```

## 5. Error Handling Strategy

### 5.1 Error Types

```go
package errors

// ParseError represents a parsing error with position information
type ParseError struct {
    Message  string
    Position Position
    Format   Format
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s at line %d, column %d",
        e.Message, e.Position.Line, e.Position.Column)
}

// Common error constructors
func NewSyntaxError(pos Position, message string) *ParseError
func NewUnexpectedTokenError(pos Position, expected, got string) *ParseError
func NewUnexpectedEOFError(pos Position, expected string) *ParseError
```

### 5.2 Error Recovery

**Philosophy:** Fail fast with detailed error message
- No error recovery or "best effort" parsing
- First error stops parsing
- Error message includes:
  - Exact position (line, column)
  - What was expected
  - What was found
  - Surrounding context (if available)

**Example Error Messages:**
```
Error at line 5, column 12: expected '}', got 'EOF'
Error at line 10, column 5: invalid function argument: expected number, got string
Error at line 15, column 20: unclosed string literal
```

## 6. Directory Structure

```
shape/
├── README.md                      # Project overview, quick start
├── LICENSE                        # License file
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build, test, lint targets
│
├── pkg/                           # Public API (importable by consumers)
│   ├── shape/                    # Main public API
│   │   ├── shape.go              # Parse(), ParseAuto(), MustParse()
│   │   └── shape_test.go         # Public API tests
│   │
│   └── ast/                       # AST model (public)
│       ├── node.go                # SchemaNode interface, NodeType
│       ├── literal.go             # LiteralNode implementation
│       ├── type.go                # TypeNode implementation
│       ├── function.go            # FunctionNode implementation
│       ├── object.go              # ObjectNode implementation
│       ├── array.go               # ArrayNode implementation
│       ├── position.go            # Position struct
│       ├── visitor.go             # Visitor interface for traversal
│       ├── serialization.go       # JSON marshaling/unmarshaling
│       ├── printer.go             # Pretty-print utilities
│       └── ast_test.go            # AST tests
│
├── internal/                      # Private implementation
│   ├── parser/                    # Parser abstraction
│   │   ├── parser.go              # Parser interface
│   │   ├── factory.go             # Parser factory
│   │   ├── format.go              # Format enum and detection
│   │   └── errors.go              # ParseError types
│   │
│   ├── parser/                    # Parser infrastructure
│   │   ├── parser.go              # Parser interface
│   │   ├── factory.go             # Parser factory
│   │   ├── format.go              # Format enum
│   │   └── errors.go              # Error types
│   │
│   │   # Note: Format-specific parsers are implemented in
│   │   # separate parser project repositories, not in shape/
│   │
│   ├── tokenizer/                 # Embedded tokenization framework
│   │   ├── stream.go              # Character stream abstraction
│   │   ├── stream_test.go
│   │   ├── tokens.go              # Token struct and tokenizer
│   │   ├── tokens_test.go
│   │   ├── matchers.go            # Matcher interface + built-ins
│   │   ├── matchers_test.go
│   │   ├── position.go            # Position tracking
│   │   ├── text.go                # Text/rune utilities
│   │   ├── text_test.go
│   │   ├── numbers.go             # Number parsing utilities
│   │   ├── numbers_test.go
│   │   └── README.md              # Tokenizer framework documentation
│   │
│   └── testdata/                  # Shared test fixtures
│       # Note: Format-specific test data lives in parser projects
│
├── docs/                          # Internal documentation
│   ├── architecture/
│   │   ├── ARCHITECTURE.md        # This file
│   │   ├── decisions/             # Architecture Decision Records
│   │   │   ├── 0001-ast-design.md
│   │   │   ├── 0002-use-df2-go.md
│   │   │   ├── 0003-embed-tokenizer.md
│   │   │   ├── 0004-parser-strategy.md
│   │   │   └── 0005-grammar-as-verification.md
│   │   ├── diagrams/
│   │   │   ├── component-diagram.md
│   │   │   ├── parser-flow.md
│   │   │   └── ast-structure.md
│   │   └── specifications/
│   │       # Note: Format specifications are maintained in
│   │       # respective parser project repositories
│   └── contributor/
│       ├── local-setup.md
│       ├── contributing.md
│       └── testing-guide.md
│
├── examples/                      # Usage examples
│   ├── basic/
│   │   └── main.go                # Basic parsing example
│   ├── advanced/
│   │   └── main.go                # Advanced usage (visitor, serialization)
│   └── multi-format/
│       └── main.go                # Parsing multiple formats
│
└── tools/                         # Development tools
    ├── ast-visualizer/            # Tool to visualize AST (optional)
    └── format-converter/          # Tool to convert between formats (optional)
```

## 7. Embedded Tokenization Framework

### 7.1 Tokenizer Architecture

Shape includes an embedded tokenization framework in `internal/tokenizer/` that provides:

**Architecture Decision:** Originally developed as the df2-go project, the tokenization code has been embedded directly into shape to create a fully self-contained parser library with zero external tokenization dependencies (see ADR 0003).

- **UTF-8 Support**: Native rune-based character stream processing
- **Backtracking**: Stream cloning for speculative matching
- **Position Tracking**: Automatic line/column tracking for error messages
- **Pattern Composition**: Functional approach to building matchers

**Embedded Structure:**
- `internal/tokenizer/stream.go` - Stream abstraction with position tracking
- `internal/tokenizer/tokens.go` - Token struct and Tokenizer implementation
- `internal/tokenizer/matchers.go` - Matcher interface and built-in matchers
- `internal/tokenizer/text.go` - Text and rune manipulation utilities
- `internal/tokenizer/numbers.go` - Number parsing utilities

### 7.2 Tokenizer Pattern

Each format implements custom matchers using the integrated framework:

```go
package customformat

import (
    "github.com/shapestone/shape-core/pkg/tokenizer"
)

// Custom matchers for your format
func identifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
    // Match identifiers specific to your format
}

func functionMatcher(stream tokenizer.Stream) *tokenizer.Token {
    // Match function calls if applicable
}

// Matcher list (framework + custom)
var Matchers = []tokenizer.Matcher{
    tokenizer.CharMatcher("BraceOpen", '{'),
    tokenizer.CharMatcher("BraceClose", '}'),
    // ... built-in matchers
    functionMatcher,      // Custom
    identifierMatcher,    // Custom
    // ... more matchers
}
```

### 7.3 Parser Pattern

Each format implements Parser interface:

```go
package customformat

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-core/pkg/parser"
    "github.com/shapestone/shape-core/pkg/tokenizer"
)

type Parser struct {
    tokenizer *tokenizer.Tokenizer
    current   *tokenizer.Token
    hasToken  bool
}

func NewParser() *Parser {
    return &Parser{}
}

func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
    // Initialize tokenizer
    t := tokenizer.NewTokenizer(Matchers...)
    t.Initialize(input)
    p.tokenizer = &t

    // Load first token
    p.advance()

    // Parse recursively
    return p.parseValue()
}

func (p *Parser) parseValue() (ast.SchemaNode, error) {
    // Parse based on your chosen technique (LL(1), Pratt, etc.)
    switch p.current.Kind() {
    case TokenBraceOpen:
        return p.parseObject()
    case TokenBracketOpen:
        return p.parseArray()
    // ... more cases
    }
}
```

## 8. Data-Validator Integration

### 8.1 Updated Architecture

```
data-validator/
├── pkg/validator/
│   ├── validator.go          # Public validation API
│   └── validator_test.go
│
├── internal/
│   ├── traverser/            # AST traversal + validation logic
│   │   ├── traverser.go      # Walks shape AST
│   │   ├── literal.go        # Validates LiteralNode
│   │   ├── type.go           # Validates TypeNode
│   │   ├── function.go       # Validates FunctionNode
│   │   ├── object.go         # Validates ObjectNode
│   │   └── array.go          # Validates ArrayNode
│   │
│   └── expression_eval/      # Expression evaluation (project-specific)
│       └── evaluator.go      # Expression evaluation logic
│
└── go.mod
    require (
        github.com/shapestone/shape-core    // Shape parser
    )
```

### 8.2 Usage Pattern

```go
package validator

import (
    "github.com/shapestone/shape-core"
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-core/internal/parser"
)

// Validate validates data against a schema
func Validate(schemaInput string, format parser.Format, data interface{}) error {
    // Parse schema using shape infrastructure
    schemaAST, err := shape.Parse(format, schemaInput)
    if err != nil {
        return fmt.Errorf("schema parse error: %w", err)
    }

    // Traverse AST and validate data
    traverser := NewTraverser(schemaAST, data)
    return traverser.Validate()
}
```

### 8.3 Separation of Concerns

| Component | Responsibility | Repository |
|-----------|----------------|------------|
| **shape** | Parse schemas → AST | github.com/shapestone/shape-core |
| **downstream projects** | Use AST for validation, transformation, etc. | Various (data-validator, custom tools) |

## 9. Testing Strategy

### 9.1 Test Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E ╲          Integration tests (5%)
                 ╱────────╲        - Multi-format tests
                ╱          ╲       - data-validator integration
               ╱  Integration╲      
              ╱──────────────╲    
             ╱                ╲    Unit tests (95%)
            ╱   Unit Tests     ╲   - AST creation
           ╱                    ╲  - Tokenizer matchers
          ╱──────────────────────╲ - Parser logic
         ╱________________________╲
```

### 9.2 Test Categories

#### Unit Tests (95% of tests)
- **AST Tests:** Node creation, traversal, serialization
- **Tokenizer Tests:** Each matcher in isolation
- **Parser Tests:** Parse valid/invalid inputs
- **Error Tests:** Verify error messages and positions

#### Integration Tests (5% of tests)
- **Multi-Format Tests:** Same schema in different formats produces same AST
- **Round-Trip Tests:** Parse → Serialize → Parse
- **data-validator Integration:** Shape AST used by validator

### 9.3 Test Data Organization

```
internal/testdata/
├── custom-format/
│   ├── valid/
│   │   ├── simple-object.txt
│   │   ├── nested-object.txt
│   │   ├── array-elements.txt
│   │   └── complex-example.txt
│   └── invalid/
│       ├── unclosed-object.txt
│       ├── invalid-syntax.txt
│       └── unexpected-token.txt
└── ...
```

### 9.4 Benchmarks

```go
package shape

import "testing"

func BenchmarkParse_Simple(b *testing.B) {
    input := `simple schema example`
    format := parser.FormatCustom
    for i := 0; i < b.N; i++ {
        Parse(format, input)
    }
}

func BenchmarkParse_Complex(b *testing.B) {
    // Large nested schema
}
```

**Performance Goals:**
- Simple schema (< 10 nodes): < 100μs
- Medium schema (10-50 nodes): < 500μs
- Large schema (50-200 nodes): < 2ms

## 10. Versioning and Compatibility

### 10.1 Semantic Versioning

**Version Format:** `v{major}.{minor}.{patch}`

**Breaking Changes (major):**
- AST structure changes that break traversal
- Parser interface changes
- Public API changes

**Non-Breaking Changes (minor):**
- New format support
- New AST node types (with interface compatibility)
- New public API functions (additive)

**Bug Fixes (patch):**
- Parser bug fixes
- Error message improvements
- Performance improvements

### 10.2 Stability Guarantees

**v0.x.x (Development):**
- API may change
- AST structure may change
- No backward compatibility guarantee

**v1.x.x (Stable):**
- Public API is stable (pkg/shape, pkg/ast)
- AST structure is stable
- Breaking changes only in major versions
- 2-version deprecation policy

## 11. Documentation Plan

### 11.1 Required Documentation

#### Repository Documentation
- **README.md:** Quick start, features, installation
- **docs/architecture/ARCHITECTURE.md:** This document
- **docs/contributor/:** Setup, contributing, testing guides

#### Format Specifications
- Format specifications are maintained in their respective parser project repositories
- Each parser project documents its own format specification

#### API Documentation
- **Godoc:** Complete documentation for all public APIs
- **examples/:** Working code examples

#### Architecture Decision Records
- **0001-ast-design.md:** Why this AST structure?
- **0002-use-df2-go.md:** Why df2-go tokenizer integration?
- **0003-embed-tokenizer.md:** Why embed tokenizer instead of external dependency?
- **0004-parser-strategy.md:** LL(1) recursive descent parsing strategy
- **0005-grammar-as-verification.md:** Grammar-based verification for parser correctness

### 11.2 Example README.md Outline

```markdown
# Shape - Multi-Format Validation Schema Parser

Parse validation schemas into a unified AST.

## Features
- Infrastructure for multiple format parsers
- Format detection interface
- Detailed error messages
- Production-ready

## Installation
```bash
go get github.com/shapestone/shape
```

## Quick Start
```go
import "github.com/shapestone/shape-core"
import "github.com/shapestone/shape-customparser"

schema, err := customparser.Parse(`schema input`)
```

## Documentation
- [Format Specifications](../../docs/architecture/specifications/)
- [Architecture](../../docs/architecture/ARCHITECTURE.md)
- [API Reference](https://pkg.go.dev/github.com/shapestone/shape-core)

## Contributing
See [CONTRIBUTING.md](../../CONTRIBUTING.md)
```

## 12. Production Readiness Checklist

### 12.1 Code Quality
- [x] 95%+ test coverage (currently 95%+)
- [x] All linters pass (golangci-lint)
- [x] No panics in production code
- [x] All public APIs documented
- [x] Examples for all public APIs

### 12.2 Performance
- [x] Benchmarks for all formats
- [x] Performance targets met (< 2ms for large schemas - actual: 0.7-70µs)
- [x] Memory profiling done
- [x] No memory leaks

### 12.3 Error Handling
- [x] All errors include position information
- [x] Error messages are clear and actionable
- [x] Error types are well-documented

### 12.4 Documentation
- [x] README with quick start
- [x] All format specifications written
- [x] API documentation (godoc)
- [x] Architecture documentation complete
- [x] Examples for common use cases (basic, advanced, multi-format)
- [x] Ecosystem documentation (ECOSYSTEM.md)
- [x] Contributor guides (local-setup.md, testing-guide.md)

### 12.5 Testing
- [x] Unit tests for all components
- [x] Integration tests (cross-format validation)
- [x] Error case tests
- [x] Performance benchmarks

### 12.6 Release Preparation
- [x] CHANGELOG.md
- [x] Semantic versioning (following semver)
- [x] Tagged releases with GitHub release notes
- [x] Automated release workflow

## 13. Future Enhancements (Post v1.0)

### 13.1 Potential Features
- **Schema Validation:** Validate schemas for correctness
- **Schema Transformation:** Convert between formats
- **AST Optimization:** Simplify/optimize AST structure
- **Streaming Parser:** Parse large schemas incrementally
- **Custom Validators:** Plugin system for custom validators
- **IDE Support:** Language server protocol (LSP) for editors

### 13.2 Performance Optimizations
- **Parser Caching:** Cache parsed schemas
- **Parallel Parsing:** Parse multiple schemas concurrently
- **Zero-Copy Parsing:** Reduce allocations

### 13.3 Developer Experience
- **Better Error Messages:** Suggestions for common mistakes
- **Schema Linting:** Validate schema best practices
- **Visual Tools:** AST visualizer, format converter

## 14. Success Metrics

### 14.1 Technical Metrics
- **Parse Performance:** < 2ms for 200-node schema
- **Memory Usage:** < 10MB for large schema
- **Test Coverage:** 95%+ across all packages
- **API Stability:** Zero breaking changes in minor versions

### 14.2 Adoption Metrics
- **data-validator Integration:** Successfully used by data-validator
- **Community Feedback:** Positive feedback on API design
- **Bug Reports:** < 5 critical bugs in first 6 months

## Appendix A: AST Examples

### A.1 Simple Object

**Schema Input (format-agnostic):**
```
{"id": UUID, "name": String(1, 100)}
```

**AST:**
```go
ObjectNode{
    Properties: {
        "id":   TypeNode{TypeName: "UUID"},
        "name": FunctionNode{Name: "String", Arguments: []interface{}{1, 100}},
    },
}
```

### A.2 Nested Object

**Schema Input (format-agnostic):**
```
{
    "user": {
        "id": UUID,
        "email": Email
    }
}
```

**AST:**
```go
ObjectNode{
    Properties: {
        "user": ObjectNode{
            Properties: {
                "id":    TypeNode{TypeName: "UUID"},
                "email": TypeNode{TypeName: "Email"},
            },
        },
    },
}
```

### A.3 Array

**Schema Input (format-agnostic):**
```
{"tags": [String(1, 30)]}
```

**AST:**
```go
ObjectNode{
    Properties: {
        "tags": ArrayNode{
            ElementSchema: FunctionNode{Name: "String", Arguments: []interface{}{1, 30}},
        },
    },
}
```

## Appendix B: Parser Projects

Shape provides infrastructure for parser projects. Each parser project:
- Maintains its own repository
- Implements its own format specification
- Uses Shape's AST, tokenizer, and validation infrastructure
- Chooses its own parsing technique (LL(1), Pratt, PEG, etc.)

See [ECOSYSTEM.md](../../ECOSYSTEM.md) for a complete list of parser projects.

## Appendix C: References

- **Embedded Tokenization:** Tokenization code embedded from df2-go project (see ADR 0003)
- **LL(1) Parsing Strategy:** See ADR 0004 for detailed parser design decisions
- **df2-go (original):** github.com/shapestone/df2-go
- **Shapestone Ecosystem:** See [ECOSYSTEM.md](../../ECOSYSTEM.md) for related projects
- **Go Project Layout:** Standard Go project structure
- **Recursive Descent Parsing:** https://en.wikipedia.org/wiki/Recursive_descent_parser
- **LL Parser:** https://en.wikipedia.org/wiki/LL_parser

---

**Document Status:** Complete
**Next Steps:**
1. Create ADRs for key decisions
2. Create format specifications
3. Set up CI/CD pipeline
