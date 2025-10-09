# Shape Parser Architecture

**Version:** 1.0  
**Date:** 2025-10-09  
**Status:** Architectural Design  
**Repository:** github.com/shapestone/shape

## Executive Summary

Shape is a multi-format validation schema parser library that converts validation schema formats (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV) into a unified Abstract Syntax Tree (AST) representation. Shape serves as the foundational parsing layer for the data-validator ecosystem, providing format-agnostic schema representation that can be consumed by validation engines.

**Key Design Principles:**
- Single Responsibility: Shape parses schemas, doesn't validate data
- Format Agnostic: All formats produce the same AST structure
- Extensible: Easy to add new formats or custom validators
- Production Ready: Comprehensive error handling and testing
- Zero Breaking Changes: Semantic versioning with stable API

## 1. System Overview

### 1.1 Purpose and Scope

**What Shape Does:**
- Parses 6 validation schema formats into unified AST
- Provides format detection and parser selection
- Reports detailed parse errors with position information
- Serializes/deserializes AST for storage or transmission

**What Shape Does NOT Do:**
- Validate actual data (that's data-validator's job)
- Execute validation rules (that's wire's job)
- Manage validation state or results
- Provide CLI or server interfaces

### 1.2 Ecosystem Position

```
┌─────────────────────────────────────────────────────────┐
│                   Data Validator                        │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         Validation Traverser                      │  │
│  │  (Walks AST + validates data + calls wire)       │  │
│  └──────────────────────────────────────────────────┘  │
│                         ▲                                │
│                         │                                │
│                         │ uses AST                       │
│                         │                                │
└─────────────────────────┼────────────────────────────────┘
                          │
                          │
┌─────────────────────────┼────────────────────────────────┐
│                      Shape Parser                        │
│                         │                                │
│  ┌──────────────────────▼──────────────────────────┐    │
│  │              Schema AST Model                    │    │
│  │  (LiteralNode, TypeNode, FunctionNode, etc.)    │    │
│  └──────────────▲───────────────────────────────────┘   │
│                 │                                        │
│  ┌──────────────┴──────────────────────────────────┐    │
│  │           Format Parsers                         │    │
│  │  JSONV | XMLV | PropsV | CSVV | YAMLV | TEXTV  │    │
│  └──────────────┬──────────────────────────────────┘    │
│                 │                                        │
│  ┌──────────────▼──────────────────────────────────┐    │
│  │     Embedded Tokenization Framework             │    │
│  │     (internal/tokenizer/)                       │    │
│  │  Stream, Matchers, Position Tracking           │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
                          │
                          │
┌─────────────────────────┼────────────────────────────────┐
│                      Wire Engine                         │
│                         │                                │
│  (Expression evaluation for dynamic validation)          │
│  Integer(min, max) → validates with wire expressions     │
└──────────────────────────────────────────────────────────┘
```

### 1.3 Format Support

| Format | Description | Priority | Implementation Phase |
|--------|-------------|----------|---------------------|
| JSONV  | JSON with validation expressions | Critical | Phase 2 |
| XMLV   | XML with validation expressions | High | Phase 3 |
| PropsV | Properties (key=value) with validation | High | Phase 3 |
| CSVV   | CSV with validation headers | Medium | Phase 3 |
| YAMLV  | YAML with validation expressions | Medium | Phase 4 |
| TEXTV  | Text patterns with validation | Low | Phase 4 |

## 2. Core Architecture

### 2.1 Layered Architecture

```
┌───────────────────────────────────────────────────────┐
│              Public API Layer                         │
│  - shape.Parse(format, input)                        │
│  - shape.ParseAuto(input)                            │
│  - shape.NewParser(format)                           │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│           Parser Abstraction Layer                    │
│  - Parser interface                                   │
│  - Parser factory/registry                           │
│  - Format detection                                  │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│         Format-Specific Parsers                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │    JSONV    │  │    XMLV     │  │   PropsV    │  │
│  │   Parser    │  │   Parser    │  │   Parser    │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │
│         │                │                │          │
│  ┌──────▼────────────────▼────────────────▼──────┐  │
│  │      Format-Specific Tokenizers               │  │
│  │      (Uses integrated matcher framework)      │  │
│  └───────────────────────────────────────────────┘  │
└───────────────────┬───────────────────────────────────┘
                    │
┌───────────────────▼───────────────────────────────────┐
│            Schema AST Model                           │
│  - SchemaNode interface                              │
│  - Node implementations (Object, Array, Function...)  │
│  - AST traversal utilities                           │
│  - Serialization/deserialization                     │
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
- **Responsibility:** Parse specific format into AST
- **Pattern:** Tokenizer (integrated framework) + Recursive Descent Parser
- **Components per format:**
  - Tokenizer with format-specific matchers
  - Recursive descent parser
  - Error handling with position tracking
- **Location:** `internal/parser/jsonv/`, `internal/parser/xmlv/`, etc.

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

import "github.com/shapestone/shape/pkg/ast"

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
    FormatJSONV  Format = iota
    FormatXMLV
    FormatPropsV
    FormatCSVV
    FormatYAMLV
    FormatTEXTV
)

func (f Format) String() string {
    // Returns "JSONV", "XMLV", etc.
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
    // Heuristic detection based on first non-whitespace character
    // { → JSONV, < → XMLV, etc.
}
```

### 4.3 Public API

```go
package shape

import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
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
        return nil, parser.FormatJSONV, err
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
│       ├── node.go               # SchemaNode interface, NodeType
│       ├── literal.go            # LiteralNode implementation
│       ├── type.go               # TypeNode implementation
│       ├── function.go           # FunctionNode implementation
│       ├── object.go             # ObjectNode implementation
│       ├── array.go              # ArrayNode implementation
│       ├── position.go           # Position struct
│       ├── visitor.go            # Visitor interface for traversal
│       ├── serialization.go      # JSON marshaling/unmarshaling
│       ├── printer.go            # Pretty-print utilities
│       └── ast_test.go           # AST tests
│
├── internal/                      # Private implementation
│   ├── parser/                   # Parser abstraction
│   │   ├── parser.go             # Parser interface
│   │   ├── factory.go            # Parser factory
│   │   ├── format.go             # Format enum and detection
│   │   └── errors.go             # ParseError types
│   │
│   ├── parser/jsonv/             # JSONV parser
│   │   ├── parser.go             # JSONV parser implementation
│   │   ├── tokenizer.go          # JSONV tokenizer (built-in matchers)
│   │   ├── parser_test.go        # JSONV parser tests
│   │   └── tokenizer_test.go     # JSONV tokenizer tests
│   │
│   ├── parser/xmlv/              # XMLV parser
│   │   ├── parser.go
│   │   ├── tokenizer.go
│   │   ├── parser_test.go
│   │   └── tokenizer_test.go
│   │
│   ├── parser/propsv/            # PropsV parser
│   │   ├── parser.go
│   │   ├── tokenizer.go
│   │   ├── parser_test.go
│   │   └── tokenizer_test.go
│   │
│   ├── parser/csvv/              # CSVV parser
│   │   ├── parser.go
│   │   ├── tokenizer.go
│   │   ├── parser_test.go
│   │   └── tokenizer_test.go
│   │
│   ├── parser/yamlv/             # YAMLV parser
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── README.md             # YAML-specific notes
│   │
│   ├── parser/textv/             # TEXTV parser
│   │   ├── parser.go
│   │   ├── tokenizer.go
│   │   ├── parser_test.go
│   │   └── tokenizer_test.go
│   │
│   ├── tokenizer/                # Embedded tokenization framework
│   │   ├── stream.go            # Character stream abstraction
│   │   ├── stream_test.go
│   │   ├── tokens.go            # Token struct and tokenizer
│   │   ├── tokens_test.go
│   │   ├── matchers.go          # Matcher interface + built-ins
│   │   ├── matchers_test.go
│   │   ├── position.go          # Position tracking
│   │   ├── text.go              # Text/rune utilities
│   │   ├── text_test.go
│   │   ├── numbers.go           # Number parsing utilities
│   │   ├── numbers_test.go
│   │   └── README.md            # Tokenizer framework documentation
│   │
│   └── testdata/                 # Shared test fixtures
│       ├── jsonv/
│       │   ├── valid/            # Valid JSONV files
│       │   └── invalid/          # Invalid JSONV files (for error tests)
│       ├── xmlv/
│       │   ├── valid/
│       │   └── invalid/
│       └── ...                   # Other format test data
│
├── docs/                          # Internal documentation
│   ├── architecture/
│   │   ├── ARCHITECTURE.md       # This file
│   │   ├── decisions/            # Architecture Decision Records
│   │   │   ├── 0001-ast-design.md
│   │   │   ├── 0002-use-df2-go.md
│   │   │   ├── 0003-parser-strategy.md
│   │   │   └── 0004-error-handling.md
│   │   ├── diagrams/
│   │   │   ├── component-diagram.md
│   │   │   ├── parser-flow.md
│   │   │   └── ast-structure.md
│   │   └── specifications/
│   │       ├── jsonv-spec.md     # JSONV format specification
│   │       ├── xmlv-spec.md      # XMLV format specification
│   │       ├── propsv-spec.md    # PropsV format specification
│   │       ├── csvv-spec.md      # CSVV format specification
│   │       ├── yamlv-spec.md     # YAMLV format specification
│   │       └── textv-spec.md     # TEXTV format specification
│   └── contributor/
│       ├── local-setup.md
│       ├── contributing.md
│       └── testing-guide.md
│
├── examples/                      # Usage examples
│   ├── basic/
│   │   └── main.go               # Basic parsing example
│   ├── advanced/
│   │   └── main.go               # Advanced usage (visitor, serialization)
│   └── multi-format/
│       └── main.go               # Parsing multiple formats
│
└── tools/                         # Development tools
    ├── ast-visualizer/           # Tool to visualize AST (optional)
    └── format-converter/         # Tool to convert between formats (optional)
```

## 7. Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal:** Establish core architecture and AST model

**Deliverables:**
- Project structure and Go module setup
- AST node interfaces and implementations
- Position tracking
- AST serialization (JSON)
- AST pretty-printer
- Comprehensive AST tests

**Success Criteria:**
- All AST node types implemented
- AST can be created programmatically
- AST can be serialized/deserialized
- AST can be pretty-printed
- 100% test coverage for AST

### Phase 2: JSONV Parser (Week 2)
**Goal:** Implement first complete parser (most important format)

**Deliverables:**
- JSONV tokenizer using integrated framework
- JSONV recursive descent parser
- Parser interface and factory
- Public API (Parse, ParseAuto)
- Error handling with position tracking
- Comprehensive JSONV tests

**Success Criteria:**
- JSONV parser handles all syntax from spec
- Error messages include line/column
- 95%+ test coverage
- Performance < 1ms for typical schemas
- Integration tests with data-validator

### Phase 3: Additional Core Formats (Week 3)
**Goal:** Implement XMLV, PropsV, CSVV parsers

**Deliverables:**
- XMLV parser (similar to JSONV)
- PropsV parser (simpler than JSONV)
- CSVV parser (line-oriented)
- Format detection logic
- Multi-format tests

**Success Criteria:**
- All 3 formats parse correctly
- Format auto-detection works
- Consistent error handling across formats
- 90%+ test coverage per format

### Phase 4: Advanced Formats + Polish (Week 4)
**Goal:** Complete remaining formats and production polish

**Deliverables:**
- YAMLV parser (evaluate 3rd-party lib vs custom)
- TEXTV parser (pattern-based)
- Performance benchmarks
- Documentation (README, examples, godoc)
- Migration guide for data-validator

**Success Criteria:**
- All 6 formats implemented
- Performance benchmarks documented
- Complete API documentation
- Ready for v0.1.0 release

## 8. Embedded Tokenization Framework

### 8.1 Tokenizer Architecture

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

### 8.2 Tokenizer Pattern

Each format implements custom matchers using the integrated framework:

```go
package jsonv

import (
    "github.com/shapestone/shape/internal/tokenizer"
)

// Custom JSONV matchers
func identifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
    // Match type identifiers: UUID, Email, etc.
}

func functionMatcher(stream tokenizer.Stream) *tokenizer.Token {
    // Match function calls: Integer(1, 100)
}

// Matcher list (framework + custom)
var Matchers = []tokenizer.Matcher{
    tokenizer.CharMatcher("ObjectStart", '{'),
    tokenizer.CharMatcher("ObjectEnd", '}'),
    // ... built-in matchers
    functionMatcher,      // Custom
    identifierMatcher,    // Custom
    // ... more matchers
}
```

### 8.3 Parser Pattern

Each format implements Parser interface:

```go
package jsonv

import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/internal/tokenizer"
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
    // Recursive descent parsing
}
```

## 9. Data-Validator Integration

### 9.1 Updated Architecture

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
│   │   ├── function.go       # Validates FunctionNode (uses wire)
│   │   ├── object.go         # Validates ObjectNode
│   │   └── array.go          # Validates ArrayNode
│   │
│   └── wire_integration/     # Wire expression evaluation
│       └── evaluator.go      # Calls wire for function validation
│
└── go.mod
    require (
        github.com/shapestone/shape v0.1.0    // Shape parser
        github.com/shapestone/wire v0.9.0     // Wire engine
    )
```

### 9.2 Usage Pattern

```go
package validator

import (
    "github.com/shapestone/shape"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
)

// Validate validates data against a schema
func Validate(schemaInput string, data interface{}) error {
    // Parse schema using shape
    schemaAST, err := shape.Parse(parser.FormatJSONV, schemaInput)
    if err != nil {
        return fmt.Errorf("schema parse error: %w", err)
    }
    
    // Traverse AST and validate data
    traverser := NewTraverser(schemaAST, data)
    return traverser.Validate()
}
```

### 9.3 Separation of Concerns

| Component | Responsibility | Repository |
|-----------|----------------|------------|
| **shape** | Parse schema → AST | github.com/shapestone/shape |
| **data-validator** | Traverse AST + validate data | github.com/shapestone/data-validator |
| **wire** | Evaluate validation expressions | github.com/shapestone/wire |

## 10. Testing Strategy

### 10.1 Test Pyramid

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

### 10.2 Test Categories

#### Unit Tests (95% of tests)
- **AST Tests:** Node creation, traversal, serialization
- **Tokenizer Tests:** Each matcher in isolation
- **Parser Tests:** Parse valid/invalid inputs
- **Error Tests:** Verify error messages and positions

#### Integration Tests (5% of tests)
- **Multi-Format Tests:** Same schema in different formats produces same AST
- **Round-Trip Tests:** Parse → Serialize → Parse
- **data-validator Integration:** Shape AST used by validator

### 10.3 Test Data Organization

```
internal/testdata/
├── jsonv/
│   ├── valid/
│   │   ├── simple-object.jsonv
│   │   ├── nested-object.jsonv
│   │   ├── array-elements.jsonv
│   │   ├── mixed-literals-validators.jsonv
│   │   └── all-validators.jsonv
│   └── invalid/
│       ├── unclosed-object.jsonv
│       ├── invalid-function-args.jsonv
│       ├── missing-colon.jsonv
│       └── unexpected-token.jsonv
└── ...
```

### 10.4 Benchmarks

```go
package shape

import "testing"

func BenchmarkParse_JSONV_Simple(b *testing.B) {
    input := `{"id": UUID, "name": String(1,100)}`
    for i := 0; i < b.N; i++ {
        Parse(parser.FormatJSONV, input)
    }
}

func BenchmarkParse_JSONV_Complex(b *testing.B) {
    // Large nested schema
}
```

**Performance Goals:**
- Simple schema (< 10 nodes): < 100μs
- Medium schema (10-50 nodes): < 500μs
- Large schema (50-200 nodes): < 2ms

## 11. Versioning and Compatibility

### 11.1 Semantic Versioning

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

### 11.2 Stability Guarantees

**v0.x.x (Development):**
- API may change
- AST structure may change
- No backward compatibility guarantee

**v1.x.x (Stable):**
- Public API is stable (pkg/shape, pkg/ast)
- AST structure is stable
- Breaking changes only in major versions
- 2-version deprecation policy

## 12. Documentation Plan

### 12.1 Required Documentation

#### Repository Documentation
- **README.md:** Quick start, features, installation
- **docs/architecture/ARCHITECTURE.md:** This document
- **docs/contributor/:** Setup, contributing, testing guides

#### Format Specifications
- **docs/architecture/specifications/jsonv-spec.md**
- **docs/architecture/specifications/xmlv-spec.md**
- **docs/architecture/specifications/propsv-spec.md**
- **docs/architecture/specifications/csvv-spec.md**
- **docs/architecture/specifications/yamlv-spec.md**
- **docs/architecture/specifications/textv-spec.md**

#### API Documentation
- **Godoc:** Complete documentation for all public APIs
- **examples/:** Working code examples

#### Architecture Decision Records
- **0001-ast-design.md:** Why this AST structure?
- **0002-use-df2-go.md:** Why integrated tokenizer framework?
- **0003-parser-strategy.md:** Why recursive descent?
- **0004-error-handling.md:** Error handling philosophy

### 12.2 Example README.md Outline

```markdown
# Shape - Multi-Format Validation Schema Parser

Parse validation schemas into a unified AST.

## Features
- 6 format support (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV)
- Format auto-detection
- Detailed error messages
- Production-ready

## Installation
```bash
go get github.com/shapestone/shape
```

## Quick Start
```go
import "github.com/shapestone/shape"

schema, err := shape.Parse(parser.FormatJSONV, `{"id": UUID}`)
```

## Documentation
- [Format Specifications](docs/architecture/specifications/)
- [Architecture](docs/architecture/ARCHITECTURE.md)
- [API Reference](https://pkg.go.dev/github.com/shapestone/shape)

## Contributing
See [CONTRIBUTING.md](docs/contributor/contributing.md)
```

## 13. Production Readiness Checklist

### 13.1 Code Quality
- [ ] 95%+ test coverage
- [ ] All linters pass (golangci-lint)
- [ ] No panics in production code
- [ ] All public APIs documented
- [ ] Examples for all public APIs

### 13.2 Performance
- [ ] Benchmarks for all formats
- [ ] Performance targets met (< 2ms for large schemas)
- [ ] Memory profiling done
- [ ] No memory leaks

### 13.3 Error Handling
- [ ] All errors include position information
- [ ] Error messages are clear and actionable
- [ ] Error types are well-documented

### 13.4 Documentation
- [ ] README with quick start
- [ ] All format specifications written
- [ ] API documentation (godoc)
- [ ] Architecture documentation complete
- [ ] Examples for common use cases

### 13.5 Testing
- [ ] Unit tests for all components
- [ ] Integration tests with data-validator
- [ ] Error case tests
- [ ] Performance benchmarks

### 13.6 Release Preparation
- [ ] CHANGELOG.md
- [ ] Semantic versioning
- [ ] Tagged release (v0.1.0)
- [ ] GitHub release notes

## 14. Future Enhancements (Post v1.0)

### 14.1 Potential Features
- **Schema Validation:** Validate schemas for correctness
- **Schema Transformation:** Convert between formats
- **AST Optimization:** Simplify/optimize AST structure
- **Streaming Parser:** Parse large schemas incrementally
- **Custom Validators:** Plugin system for custom validators
- **IDE Support:** Language server protocol (LSP) for editors

### 14.2 Performance Optimizations
- **Parser Caching:** Cache parsed schemas
- **Parallel Parsing:** Parse multiple schemas concurrently
- **Zero-Copy Parsing:** Reduce allocations

### 14.3 Developer Experience
- **Better Error Messages:** Suggestions for common mistakes
- **Schema Linting:** Validate schema best practices
- **Visual Tools:** AST visualizer, format converter

## 15. Success Metrics

### 15.1 Technical Metrics
- **Parse Performance:** < 2ms for 200-node schema
- **Memory Usage:** < 10MB for large schema
- **Test Coverage:** 95%+ across all packages
- **API Stability:** Zero breaking changes in minor versions

### 15.2 Adoption Metrics
- **data-validator Integration:** Successfully used by data-validator
- **Community Feedback:** Positive feedback on API design
- **Bug Reports:** < 5 critical bugs in first 6 months

## Appendix A: AST Examples

### A.1 Simple Object

**JSONV Input:**
```jsonv
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

**JSONV Input:**
```jsonv
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

**JSONV Input:**
```jsonv
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

## Appendix B: Format Comparison

| Feature | JSONV | XMLV | PropsV | CSVV | YAMLV | TEXTV |
|---------|-------|------|--------|------|-------|-------|
| Objects | ✓ | ✓ | ✓ | ✗ | ✓ | Partial |
| Arrays | ✓ | ✓ | ✗ | ✓ | ✓ | Partial |
| Nesting | ✓ | ✓ | ✓ (dots) | ✗ | ✓ | Limited |
| Comments | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Complexity | Medium | High | Low | Low | High | Medium |
| Framework fit | Excellent | Good | Excellent | Good | Poor | Good |

## Appendix C: References

- **Embedded Tokenization:** Tokenization code embedded from df2-go project at `internal/tokenizer/` (see ADR 0003)
- **df2-go (original):** github.com/shapestone/df2-go
- **wire:** github.com/shapestone/wire
- **data-validator:** github.com/shapestone/data-validator
- **Go Project Layout:** Standard Go project structure
- **Recursive Descent Parsing:** https://en.wikipedia.org/wiki/Recursive_descent_parser

---

**Document Status:** Complete  
**Next Steps:**
1. Create ADRs for key decisions
2. Create format specifications
3. Begin Phase 1 implementation
4. Set up CI/CD pipeline
