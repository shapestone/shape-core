# Parser Implementation Guide

**Audience:** Developers and LLMs implementing new data format parsers for the Shape ecosystem

**Purpose:** This guide ensures all parser projects follow Shape's architecture, patterns, and quality standards.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Critical Understanding: AST vs Data](#critical-understanding-ast-vs-data)
3. [What Shape Provides](#what-shape-provides)
4. [What You Build](#what-you-build)
5. [Parser Project Structure](#parser-project-structure)
6. [Step-by-Step Implementation](#step-by-step-implementation)
7. [Testing Strategy](#testing-strategy)
8. [Documentation Requirements](#documentation-requirements)
9. [CI/CD Setup](#cicd-setup)
10. [Complete Working Example](#complete-working-example)

---

## Architecture Overview

### Shape Ecosystem Architecture

```
┌──────────────────────────────────────────────────────┐
│  Parser Projects (Separate Repositories)             │
│  - shape-json, shape-yaml, shape-xml (data formats)  │
│  - shape-csv, shape-props                            │
│                                                      │
│  Each project:                                       │
│  - Uses Shape's tokenizer infrastructure             │
│  - Implements format-specific lexing/parsing         │
│  - Returns Go types (NOT AST nodes)                  │
│  - Tests itself using Shape's grammar tools          │
│  - Is self-contained and independently versioned     │
└────────────────┬─────────────────────────────────────┘
                 │
                 │ depends on (import)
                 ▼
┌──────────────────────────────────────────────────────┐
│  Shape (Infrastructure Only)                         │
│  github.com/shapestone/shape                         │
│                                                      │
│  Provides:                                           │
│  - pkg/tokenizer/* - Tokenization framework          │
│  - pkg/ast/* - AST for VALIDATION SCHEMAS            │
│  - pkg/validator/* - Schema validation framework     │
│  - pkg/grammar/* - Grammar verification tools        │
│  - pkg/parser/* - Parser interface definitions       │
│                                                      │
│  Does NOT provide:                                   │
│  - Generic data AST (use Go types instead)           │
│  - Format-specific parsers                           │
│  - Data parsing implementations                      │
└──────────────────────────────────────────────────────┘
```

### Key Principles

1. **Shape is Infrastructure Only**
   - Provides reusable tokenization and validation components
   - No format-specific implementations
   - No dependencies on parser projects

2. **Parser Projects are Self-Contained**
   - Own repository, versioning, releases
   - Own EBNF grammar in `docs/grammar/`
   - Import Shape as a dependency
   - Return Go types for parsed data

3. **Grammar-Driven Development**
   - EBNF grammar is source of truth
   - Grammar generates verification tests
   - Grammar guides LLM-assisted development
   - Grammar fragments document code

4. **Hand-Coded Parsers with Technology Freedom**
   - NOT using parser generators or meta-grammars
   - Freedom to choose parsing technique per format
   - LL(1) recursive descent recommended (default)
   - Optimized for performance and error quality

---

## Critical Understanding: AST vs Data

**IMPORTANT:** This is the most critical concept to understand before implementing a parser.

### Shape's AST is for VALIDATION SCHEMAS, NOT DATA

Shape's AST nodes (`ArrayNode`, `ObjectNode`, etc.) are designed to represent **validation rules**, not parsed data.

#### WRONG (Do NOT Do This):

```go
// WRONG: Trying to use AST nodes to hold parsed JSON data
func parseArray() (*ast.ArrayNode, error) {
    var elements []ast.Node  // WRONG: collecting data elements

    for /* parse elements */ {
        elem, err := parseValue()
        elements = append(elements, elem)  // WRONG: treating data as AST
    }

    // WRONG: This API doesn't exist!
    // NewArrayNode expects (elementSchema SchemaNode, pos Position)
    return ast.NewArrayNode(elements, startPos), nil  // COMPILE ERROR!
}
```

**Why this is wrong:**
- `ast.NewArrayNode(elementSchema SchemaNode, pos Position)` expects a **schema**, not data
- The first parameter should describe what type of elements are ALLOWED, not contain actual data
- Example: `ast.NewArrayNode(ast.NewTypeNode("string"), pos)` means "array of strings are valid"

#### CORRECT (Do This Instead):

```go
// CORRECT: Return Go types for parsed data
func (p *Parser) parseArray() ([]interface{}, error) {
    var elements []interface{}  // CORRECT: Go slice for data

    for /* parse elements */ {
        elem, err := p.parseValue()  // Returns interface{}
        if err != nil {
            return nil, err
        }
        elements = append(elements, elem)  // CORRECT: collecting data
    }

    return elements, nil  // CORRECT: Return Go slice
}

func (p *Parser) parseObject() (map[string]interface{}, error) {
    properties := make(map[string]interface{})  // CORRECT: Go map for data

    for /* parse properties */ {
        key, value, err := p.parseProperty()
        if err != nil {
            return nil, err
        }
        properties[key] = value  // CORRECT: collecting data
    }

    return properties, nil  // CORRECT: Return Go map
}

func (p *Parser) parseString() (string, error) {
    // CORRECT: Return primitive type
    return unquotedString, nil
}

func (p *Parser) parseNumber() (float64, error) {
    // CORRECT: Return primitive type
    return parsedNumber, nil
}
```

### When DO You Use Shape's AST?

You use Shape's AST when implementing **validation**, not parsing:

```go
// Validation schema: "array of strings"
schema := ast.NewArrayNode(
    ast.NewTypeNode("string", pos),  // Element schema: must be string
    pos,
)

// Validation schema: "object with required 'name' property"
schema := ast.NewObjectNode(
    map[string]ast.SchemaNode{
        "name": ast.NewTypeNode("string", pos),  // name must be string
        "age":  ast.NewTypeNode("number", pos),  // age must be number
    },
    pos,
)

// Then validate parsed data against schema
err := validator.Validate(parsedData, schema)
```

### Summary: Two Separate Concerns

| Concern | What | Returns | Uses |
|---------|------|---------|------|
| **Parsing** | Convert format → data | Go types (`interface{}`, `map`, `slice`) | Parser projects |
| **Validation** | Verify data → schema | AST nodes (`ArrayNode`, `ObjectNode`) | Shape's validator |

---

## What Shape Provides

Shape provides infrastructure components that parser projects build upon:

### 1. Tokenization Framework (`pkg/tokenizer/`)

**Stream abstraction:**
```go
type Stream interface {
    Clone() Stream                    // Backtracking checkpoint
    Match(cs Stream)                  // Restore checkpoint
    PeekChar() (rune, bool)          // Look ahead
    NextChar() (rune, bool)          // Read and advance
    MatchChars([]rune) bool          // Match sequence
    IsEos() bool                     // End of stream
    GetRow() int                     // Position tracking
    GetColumn() int
    GetOffset() int
    Reset()                          // Reset to beginning
}

// Create from string
stream := tokenizer.NewStream(input)

// Create from io.Reader (for large files)
stream := tokenizer.NewStreamFromReader(file)
```

**Built-in matchers:**
```go
// Character/String matchers
tokenizer.CharMatcherFunc(kind, char)
tokenizer.StringMatcherFunc(kind, str)
tokenizer.RegexMatcherFunc(kind, pattern)

// Combinators
tokenizer.OneOf(matcher1, matcher2, ...)
tokenizer.Sequence(matcher1, matcher2, ...)

// Utilities
tokenizer.WhitespaceMatcherFunc()
tokenizer.NumberMatcherFunc()
```

**Tokenizer:**
```go
type Tokenizer interface {
    Initialize(input string)
    NextToken() (Token, error)
    Tokenize() ([]Token, bool)
}

// Create custom tokenizer
tok := tokenizer.NewTokenizer(
    tokenizer.WhitespaceMatcherFunc(),  // Skip whitespace
    tokenizer.StringMatcherFunc("LBrace", "{"),
    tokenizer.StringMatcherFunc("RBrace", "}"),
    tokenizer.RegexMatcherFunc("String", `"[^"]*"`),
    // ... more matchers
)
```

### 2. Grammar Verification Tools (`pkg/grammar/`)

**EBNF parsing and test generation:**
```go
// Parse EBNF grammar
spec, err := grammar.ParseEBNF("docs/grammar/json.ebnf")

// Generate verification tests
tests := spec.GenerateTests(grammar.TestOptions{
    MaxDepth:      5,
    CoverAllRules: true,
    EdgeCases:     true,
    InvalidCases:  true,
})

// Track grammar coverage
tracker := grammar.NewCoverageTracker(spec)
coverage := tracker.Report()
```

### 3. Parser Interface (`pkg/parser/`)

**Format enumeration and parser contract:**
```go
type Format int

const (
    FormatJSON Format = iota
    FormatYAML
    FormatXML
    // ... etc
)

type Parser interface {
    Parse(input string) (interface{}, error)
    Format() Format
}
```

### 4. Validation Framework (`pkg/validator/`)

**Schema validation (separate from parsing):**
```go
// After parsing data, validate against schema
err := validator.Validate(parsedData, schema)
```

### 5. AST for Validation Schemas (`pkg/ast/`)

**Node types for describing validation rules:**
- `TypeNode` - Type constraints (string, number, bool, etc.)
- `ArrayNode` - Array validation with element schema
- `ObjectNode` - Object validation with property schemas
- `LiteralNode` - Exact value match
- `UnionNode` - Multiple allowed types
- And more...

**Remember:** These are for validation schemas, NOT for holding parsed data!

---

## What You Build

When creating a parser project (e.g., `shape-json`), you build:

### 1. Format-Specific Tokenizer

Using Shape's tokenizer framework:

```go
package tokenizer

import "github.com/shapestone/shape/pkg/tokenizer"

const (
    TokenLBrace   = "LBrace"
    TokenRBrace   = "RBrace"
    TokenString   = "String"
    TokenNumber   = "Number"
    TokenTrue     = "True"
    TokenFalse    = "False"
    TokenNull     = "Null"
    // ... more token types
)

func NewTokenizer() tokenizer.Tokenizer {
    return tokenizer.NewTokenizer(
        tokenizer.WhitespaceMatcherFunc(),
        tokenizer.StringMatcherFunc(TokenTrue, "true"),
        tokenizer.StringMatcherFunc(TokenFalse, "false"),
        tokenizer.StringMatcherFunc(TokenNull, "null"),
        tokenizer.StringMatcherFunc(TokenLBrace, "{"),
        tokenizer.StringMatcherFunc(TokenRBrace, "}"),
        tokenizer.RegexMatcherFunc(TokenString, `"(?:[^"\\]|\\.)*"`),
        tokenizer.RegexMatcherFunc(TokenNumber, `-?[0-9]+(\.[0-9]+)?`),
        // ... more matchers
    )
}
```

### 2. Format-Specific Parser

Returning Go types:

```go
package parser

type Parser struct {
    tokenizer tokenizer.Tokenizer
    current   tokenizer.Token
}

// Parse returns Go types (NOT AST nodes)
func (p *Parser) Parse() (interface{}, error) {
    return p.parseValue()
}

// parseValue returns interface{} (could be map, slice, string, number, bool, nil)
func (p *Parser) parseValue() (interface{}, error) {
    switch p.peek().Kind() {
    case TokenLBrace:
        return p.parseObject()  // Returns map[string]interface{}
    case TokenLBracket:
        return p.parseArray()   // Returns []interface{}
    case TokenString:
        return p.parseString()  // Returns string
    case TokenNumber:
        return p.parseNumber()  // Returns float64
    case TokenTrue, TokenFalse:
        return p.parseBoolean() // Returns bool
    case TokenNull:
        return nil, nil         // Returns nil
    default:
        return nil, fmt.Errorf("unexpected token: %s", p.peek().Kind())
    }
}

// parseObject returns map[string]interface{} (NOT *ast.ObjectNode)
func (p *Parser) parseObject() (map[string]interface{}, error) {
    properties := make(map[string]interface{})

    // ... parse properties ...

    return properties, nil  // Return Go map
}

// parseArray returns []interface{} (NOT *ast.ArrayNode)
func (p *Parser) parseArray() ([]interface{}, error) {
    var elements []interface{}

    // ... parse elements ...

    return elements, nil  // Return Go slice
}

// parseString returns string (NOT *ast.LiteralNode)
func (p *Parser) parseString() (string, error) {
    // ... parse string ...
    return unquotedString, nil
}
```

### 3. Public API

Simple, clean interface:

```go
package json

import "github.com/shapestone/shape-json/internal/parser"

// Parse parses JSON input and returns Go types
//
// Returns:
//   - map[string]interface{} for objects
//   - []interface{} for arrays
//   - string, float64, bool, or nil for primitives
//
// Example:
//   data, err := json.Parse(`{"name": "Alice", "age": 30}`)
//   if err != nil {
//       // handle error
//   }
//   obj := data.(map[string]interface{})
//   name := obj["name"].(string)
func Parse(input string) (interface{}, error) {
    p := parser.NewParser(input)
    return p.Parse()
}
```

### 4. EBNF Grammar

Source of truth in `docs/grammar/{format}.ebnf`:

```ebnf
// JSON Grammar Specification

Value = Object | Array | String | Number | Boolean | Null ;

Object = "{" [ Property { "," Property } ] "}" ;
Property = String ":" Value ;

Array = "[" [ Value { "," Value } ] "]" ;

String = '"' [^"]* '"' ;
Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ;
Boolean = "true" | "false" ;
Null = "null" ;
```

### 5. Grammar-Based Tests

Verify parser correctness:

```go
func TestGrammarVerification(t *testing.T) {
    spec, err := grammar.ParseEBNF("../../docs/grammar/json.ebnf")
    if err != nil {
        t.Fatalf("failed to parse grammar: %v", err)
    }

    tests := spec.GenerateTests(grammar.TestOptions{
        MaxDepth:      5,
        CoverAllRules: true,
    })

    for _, test := range tests {
        t.Run(test.Name, func(t *testing.T) {
            result, err := Parse(test.Input)

            if test.ShouldSucceed {
                if err != nil {
                    t.Errorf("Valid input rejected: %v", err)
                }
                if result == nil {
                    t.Error("Valid input returned nil")
                }
            } else {
                if err == nil {
                    t.Error("Invalid input accepted")
                }
            }
        })
    }
}
```

---

## Parser Project Structure

### Directory Layout

```
shape-{format}/
├── README.md                      # Project overview, usage
├── LICENSE                        # Apache 2.0
├── go.mod                         # Module definition
├── go.sum
│
├── docs/
│   ├── grammar/
│   │   └── {format}.ebnf          # Canonical EBNF specification
│   └── examples/
│       └── {format}_examples.md   # Usage examples
│
├── pkg/
│   └── {format}/
│       ├── parser.go              # Public API (returns Go types)
│       └── parser_test.go         # Public API tests
│
├── internal/
│   ├── tokenizer/
│   │   ├── tokenizer.go           # Format-specific tokenizer
│   │   ├── tokenizer_test.go
│   │   └── tokens.go              # Token type definitions
│   │
│   └── parser/
│       ├── parser.go              # Parser implementation (returns Go types)
│       ├── parser_test.go         # Manual tests
│       └── grammar_test.go        # Auto-generated grammar tests
│
├── examples/
│   └── main.go                    # Runnable examples
│
└── .github/
    └── workflows/
        └── ci.yml                 # CI: tests, coverage, linting
```

### Naming Conventions

- **Repository:** `shape-{format}` (e.g., `shape-json`, `shape-xml`)
- **Module:** `github.com/shapestone/shape-{format}`
- **Package:** `{format}` (e.g., `import "github.com/shapestone/shape-json/pkg/json"`)
- **Grammar file:** `docs/grammar/{format}.ebnf`

---

## Step-by-Step Implementation

### Step 1: Create Repository and Module

```bash
# Create repository
mkdir shape-json
cd shape-json

# Initialize Go module
go mod init github.com/shapestone/shape-json

# Add Shape dependency
go get github.com/shapestone/shape@latest

# Create directory structure
mkdir -p docs/grammar docs/examples
mkdir -p pkg/json
mkdir -p internal/tokenizer internal/parser
mkdir -p examples
mkdir -p .github/workflows
```

### Step 2: Define EBNF Grammar

Create `docs/grammar/json.ebnf`:

```ebnf
// JSON Grammar Specification
// Based on RFC 8259
//
// Implementation Notes:
// - Parser uses LL(1) recursive descent (see Shape ADR 0004)
// - Each production rule maps to a parser function
// - Parser returns Go types (map, slice, primitives), NOT AST nodes
// - AST nodes are for validation schemas only

// Top-level value (any JSON type)
// Parser function: parseValue() -> interface{}
// Returns: map[string]interface{}, []interface{}, string, float64, bool, or nil
Value = Object | Array | String | Number | Boolean | Null ;

// Object with properties
// Parser function: parseObject() -> map[string]interface{}
// Example valid: { "id": "abc123", "name": "Alice" }
// Example valid: {} (empty object)
// Example invalid: { id: "value" } (missing quotes on key)
// Returns: Go map with string keys and interface{} values
Object = "{" [ Property { "," Property } ] "}" ;

// Property key-value pair
// Parser function: parseProperty() -> (string, interface{})
// Returns: (key string, value interface{})
Property = String ":" Value ;

// Array of values
// Parser function: parseArray() -> []interface{}
// Example valid: [1, 2, 3]
// Example valid: [] (empty array)
// Returns: Go slice of interface{}
Array = "[" [ Value { "," Value } ] "]" ;

// String literal
// Parser function: parseString() -> string
// Returns: unquoted string content
String = '"' [^"]* '"' ;

// Number literal
// Parser function: parseNumber() -> float64
// Returns: parsed number as float64
Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;

// Boolean literal
// Parser function: parseBoolean() -> bool
// Returns: true or false
Boolean = "true" | "false" ;

// Null literal
// Parser function: parseNull() -> nil
// Returns: nil
Null = "null" ;
```

### Step 3: Implement Tokenizer

Create `internal/tokenizer/tokens.go`:

```go
package tokenizer

// Token types for JSON format
const (
    TokenLBrace   = "LBrace"    // {
    TokenRBrace   = "RBrace"    // }
    TokenLBracket = "LBracket"  // [
    TokenRBracket = "RBracket"  // ]
    TokenColon    = "Colon"     // :
    TokenComma    = "Comma"     // ,
    TokenString   = "String"    // "..."
    TokenNumber   = "Number"    // 123, 45.67, -8.9e10
    TokenTrue     = "True"      // true
    TokenFalse    = "False"     // false
    TokenNull     = "Null"      // null
    TokenEOF      = "EOF"
)
```

Create `internal/tokenizer/tokenizer.go`:

```go
package tokenizer

import (
    "github.com/shapestone/shape/pkg/tokenizer"
)

// NewTokenizer creates a tokenizer for JSON format.
func NewTokenizer() tokenizer.Tokenizer {
    return tokenizer.NewTokenizer(
        // Whitespace (automatically skipped)
        tokenizer.WhitespaceMatcherFunc(),

        // Keywords (must come before identifiers)
        tokenizer.StringMatcherFunc(TokenTrue, "true"),
        tokenizer.StringMatcherFunc(TokenFalse, "false"),
        tokenizer.StringMatcherFunc(TokenNull, "null"),

        // Structural tokens
        tokenizer.StringMatcherFunc(TokenLBrace, "{"),
        tokenizer.StringMatcherFunc(TokenRBrace, "}"),
        tokenizer.StringMatcherFunc(TokenLBracket, "["),
        tokenizer.StringMatcherFunc(TokenRBracket, "]"),
        tokenizer.StringMatcherFunc(TokenColon, ":"),
        tokenizer.StringMatcherFunc(TokenComma, ","),

        // String literals (with escape sequences)
        tokenizer.RegexMatcherFunc(TokenString, `"(?:[^"\\]|\\.)*"`),

        // Numbers (integers, floats, scientific notation)
        tokenizer.RegexMatcherFunc(TokenNumber, `-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?`),
    )
}
```

### Step 4: Implement Parser (Returns Go Types!)

Create `internal/parser/parser.go`:

```go
package parser

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/shapestone/shape-json/internal/tokenizer"
    shapeTokenizer "github.com/shapestone/shape/pkg/tokenizer"
)

// Parser implements recursive descent parsing for JSON.
// Returns Go types (map, slice, primitives), NOT AST nodes.
type Parser struct {
    tokenizer shapeTokenizer.Tokenizer
    current   shapeTokenizer.Token
    hasToken  bool
}

// NewParser creates a new JSON parser.
func NewParser(input string) *Parser {
    tok := tokenizer.NewTokenizer()
    tok.Initialize(input)

    p := &Parser{
        tokenizer: tok,
    }
    p.advance() // Load first token
    return p
}

// Parse parses the input and returns Go types.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
//
// Returns:
//   - map[string]interface{} for objects
//   - []interface{} for arrays
//   - string, float64, bool, or nil for primitives
func (p *Parser) Parse() (interface{}, error) {
    return p.parseValue()
}

// parseValue dispatches to specific parse functions.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
//
// Returns: interface{} (actual type depends on input)
func (p *Parser) parseValue() (interface{}, error) {
    switch p.peek().Kind() {
    case tokenizer.TokenLBrace:     // Object
        return p.parseObject()
    case tokenizer.TokenLBracket:   // Array
        return p.parseArray()
    case tokenizer.TokenString:     // String literal
        return p.parseString()
    case tokenizer.TokenNumber:     // Number literal
        return p.parseNumber()
    case tokenizer.TokenTrue, tokenizer.TokenFalse:  // Boolean literal
        return p.parseBoolean()
    case tokenizer.TokenNull:       // Null literal
        return p.parseNull()
    default:
        return nil, fmt.Errorf("expected value at line %d, column %d, got %s",
            p.peek().Row(), p.peek().Column(), p.peek().Kind())
    }
}

// parseObject parses an object and returns a Go map.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns: map[string]interface{} (NOT *ast.ObjectNode)
// Example: {"id": "abc", "age": 30} -> map[string]interface{}{"id": "abc", "age": 30.0}
func (p *Parser) parseObject() (map[string]interface{}, error) {
    // "{"
    if _, err := p.expect(tokenizer.TokenLBrace); err != nil {
        return nil, err
    }

    properties := make(map[string]interface{})

    // [ Property { "," Property } ] - Optional property list
    if p.peek().Kind() != tokenizer.TokenRBrace {
        // First property
        key, value, err := p.parseProperty()
        if err != nil {
            return nil, err
        }
        properties[key] = value

        // Additional properties
        for p.peek().Kind() == tokenizer.TokenComma {
            p.advance() // consume ","

            key, value, err := p.parseProperty()
            if err != nil {
                return nil, fmt.Errorf("in object property after comma: %w", err)
            }
            properties[key] = value
        }
    }

    // "}"
    if _, err := p.expect(tokenizer.TokenRBrace); err != nil {
        return nil, err
    }

    return properties, nil
}

// parseProperty parses a property key-value pair.
//
// Grammar:
//   Property = String ":" Value ;
//
// Returns: (key string, value interface{})
func (p *Parser) parseProperty() (string, interface{}, error) {
    // String (key)
    keyToken, err := p.expect(tokenizer.TokenString)
    if err != nil {
        return "", nil, fmt.Errorf("property key must be string literal: %w", err)
    }
    key := p.unquoteString(keyToken.ValueString())

    // ":"
    if _, err := p.expect(tokenizer.TokenColon); err != nil {
        return "", nil, err
    }

    // Value
    value, err := p.parseValue()
    if err != nil {
        return "", nil, fmt.Errorf("in property value for %q: %w", key, err)
    }

    return key, value, nil
}

// parseArray parses an array and returns a Go slice.
//
// Grammar:
//   Array = "[" [ Value { "," Value } ] "]" ;
//
// Returns: []interface{} (NOT *ast.ArrayNode)
// Example: [1, 2, 3] -> []interface{}{1.0, 2.0, 3.0}
func (p *Parser) parseArray() ([]interface{}, error) {
    // "["
    if _, err := p.expect(tokenizer.TokenLBracket); err != nil {
        return nil, err
    }

    var elements []interface{}

    // [ Value { "," Value } ] - Optional value list
    if p.peek().Kind() != tokenizer.TokenRBracket {
        // First element
        elem, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        elements = append(elements, elem)

        // Additional elements
        for p.peek().Kind() == tokenizer.TokenComma {
            p.advance() // consume ","

            elem, err := p.parseValue()
            if err != nil {
                return nil, fmt.Errorf("in array element after comma: %w", err)
            }
            elements = append(elements, elem)
        }
    }

    // "]"
    if _, err := p.expect(tokenizer.TokenRBracket); err != nil {
        return nil, err
    }

    return elements, nil
}

// parseString parses a string literal and returns unquoted string.
//
// Grammar:
//   String = '"' [^"]* '"' ;
//
// Returns: string (NOT *ast.LiteralNode)
func (p *Parser) parseString() (string, error) {
    token, err := p.expect(tokenizer.TokenString)
    if err != nil {
        return "", err
    }
    value := p.unquoteString(token.ValueString())
    return value, nil
}

// parseNumber parses a number literal and returns float64.
//
// Grammar:
//   Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;
//
// Returns: float64 (NOT *ast.LiteralNode)
func (p *Parser) parseNumber() (float64, error) {
    token, err := p.expect(tokenizer.TokenNumber)
    if err != nil {
        return 0, err
    }
    num, err := strconv.ParseFloat(token.ValueString(), 64)
    if err != nil {
        return 0, fmt.Errorf("invalid number: %w", err)
    }
    return num, nil
}

// parseBoolean parses a boolean literal and returns bool.
//
// Grammar:
//   Boolean = "true" | "false" ;
//
// Returns: bool (NOT *ast.LiteralNode)
func (p *Parser) parseBoolean() (bool, error) {
    token := p.peek()
    var value bool
    if token.Kind() == tokenizer.TokenTrue {
        value = true
    } else if token.Kind() == tokenizer.TokenFalse {
        value = false
    } else {
        return false, fmt.Errorf("expected boolean, got %s", token.Kind())
    }
    p.advance()
    return value, nil
}

// parseNull parses a null literal and returns nil.
//
// Grammar:
//   Null = "null" ;
//
// Returns: nil (NOT *ast.LiteralNode)
func (p *Parser) parseNull() (interface{}, error) {
    if _, err := p.expect(tokenizer.TokenNull); err != nil {
        return nil, err
    }
    return nil, nil
}

// Helper methods

// peek returns current token without advancing.
func (p *Parser) peek() shapeTokenizer.Token {
    if p.hasToken {
        return p.current
    }
    return nil
}

// advance moves to next token.
func (p *Parser) advance() error {
    token, err := p.tokenizer.NextToken()
    if err != nil {
        return err
    }
    p.current = token
    p.hasToken = true
    return nil
}

// expect consumes token of expected kind or returns error.
func (p *Parser) expect(kind string) (shapeTokenizer.Token, error) {
    if p.peek().Kind() != kind {
        return nil, fmt.Errorf("expected %s, got %s at line %d, column %d",
            kind, p.peek().Kind(), p.peek().Row(), p.peek().Column())
    }
    token := p.current
    p.advance()
    return token, nil
}

// unquoteString removes surrounding quotes from a string token.
func (p *Parser) unquoteString(s string) string {
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
        return s[1 : len(s)-1]
    }
    return s
}
```

**Key patterns:**
- Each function returns Go types, NOT AST nodes
- `parseObject()` returns `map[string]interface{}`
- `parseArray()` returns `[]interface{}`
- `parseString()` returns `string`
- `parseNumber()` returns `float64`
- `parseBoolean()` returns `bool`
- `parseNull()` returns `nil`

### Step 5: Implement Public API

Create `pkg/json/parser.go`:

```go
package json

import (
    "github.com/shapestone/shape-json/internal/parser"
)

// Parse parses JSON input and returns Go types.
//
// Returns:
//   - map[string]interface{} for objects
//   - []interface{} for arrays
//   - string for strings
//   - float64 for numbers
//   - bool for booleans
//   - nil for null
//
// Example:
//   data, err := json.Parse(`{"id": "abc123", "age": 30}`)
//   if err != nil {
//       // handle error
//   }
//   obj := data.(map[string]interface{})
//   id := obj["id"].(string)
//   age := obj["age"].(float64)
func Parse(input string) (interface{}, error) {
    p := parser.NewParser(input)
    return p.Parse()
}

// Format returns the format identifier for this parser.
func Format() string {
    return "JSON"
}
```

### Step 6: Implement Tests

Create `internal/parser/parser_test.go`:

```go
package parser

import (
    "testing"
)

func TestParseObject_Empty(t *testing.T) {
    input := `{}`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj, ok := result.(map[string]interface{})
    if !ok {
        t.Fatalf("expected map[string]interface{}, got %T", result)
    }

    if len(obj) != 0 {
        t.Errorf("expected empty object, got %d properties", len(obj))
    }
}

func TestParseObject_SingleProperty(t *testing.T) {
    input := `{ "id": "abc123" }`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj := result.(map[string]interface{})
    if len(obj) != 1 {
        t.Fatalf("expected 1 property, got %d", len(obj))
    }

    idValue, ok := obj["id"].(string)
    if !ok {
        t.Errorf("expected string for 'id', got %T", obj["id"])
    }
    if idValue != "abc123" {
        t.Errorf("expected 'abc123', got %v", idValue)
    }
}

func TestParseObject_MultipleProperties(t *testing.T) {
    input := `{
        "id": "abc123",
        "name": "Alice",
        "age": 30,
        "active": true
    }`

    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj := result.(map[string]interface{})
    if len(obj) != 4 {
        t.Errorf("expected 4 properties, got %d", len(obj))
    }

    if id, ok := obj["id"].(string); !ok || id != "abc123" {
        t.Errorf("expected id='abc123', got %v", obj["id"])
    }
    if name, ok := obj["name"].(string); !ok || name != "Alice" {
        t.Errorf("expected name='Alice', got %v", obj["name"])
    }
    if age, ok := obj["age"].(float64); !ok || age != 30 {
        t.Errorf("expected age=30, got %v", obj["age"])
    }
    if active, ok := obj["active"].(bool); !ok || !active {
        t.Errorf("expected active=true, got %v", obj["active"])
    }
}

func TestParseArray_Empty(t *testing.T) {
    input := `[]`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr, ok := result.([]interface{})
    if !ok {
        t.Fatalf("expected []interface{}, got %T", result)
    }

    if len(arr) != 0 {
        t.Errorf("expected empty array, got %d elements", len(arr))
    }
}

func TestParseArray_Numbers(t *testing.T) {
    input := `[1, 2, 3]`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr := result.([]interface{})
    if len(arr) != 3 {
        t.Errorf("expected 3 elements, got %d", len(arr))
    }

    expected := []float64{1, 2, 3}
    for i, exp := range expected {
        if num, ok := arr[i].(float64); !ok || num != exp {
            t.Errorf("element %d: expected %v, got %v", i, exp, arr[i])
        }
    }
}

func TestParseArray_Mixed(t *testing.T) {
    input := `["hello", 42, true, null]`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr := result.([]interface{})
    if len(arr) != 4 {
        t.Errorf("expected 4 elements, got %d", len(arr))
    }

    if str, ok := arr[0].(string); !ok || str != "hello" {
        t.Errorf("element 0: expected 'hello', got %v", arr[0])
    }
    if num, ok := arr[1].(float64); !ok || num != 42 {
        t.Errorf("element 1: expected 42, got %v", arr[1])
    }
    if b, ok := arr[2].(bool); !ok || !b {
        t.Errorf("element 2: expected true, got %v", arr[2])
    }
    if arr[3] != nil {
        t.Errorf("element 3: expected nil, got %v", arr[3])
    }
}

func TestParsePrimitives(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
    }{
        {"string", `"hello"`, "hello"},
        {"number_int", `42`, 42.0},
        {"number_float", `3.14`, 3.14},
        {"number_negative", `-17`, -17.0},
        {"boolean_true", `true`, true},
        {"boolean_false", `false`, false},
        {"null", `null`, nil},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser(tt.input)
            result, err := p.Parse()

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if result != tt.expected {
                t.Errorf("expected %v (%T), got %v (%T)",
                    tt.expected, tt.expected, result, result)
            }
        })
    }
}

func TestParseNested(t *testing.T) {
    input := `{
        "user": {
            "id": "abc123",
            "name": "Alice"
        },
        "tags": ["admin", "user"],
        "count": 42
    }`

    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj := result.(map[string]interface{})

    // Check nested object
    user, ok := obj["user"].(map[string]interface{})
    if !ok {
        t.Fatalf("expected user to be map[string]interface{}, got %T", obj["user"])
    }
    if user["id"].(string) != "abc123" {
        t.Errorf("expected user.id='abc123', got %v", user["id"])
    }

    // Check array
    tags, ok := obj["tags"].([]interface{})
    if !ok {
        t.Fatalf("expected tags to be []interface{}, got %T", obj["tags"])
    }
    if len(tags) != 2 {
        t.Errorf("expected 2 tags, got %d", len(tags))
    }

    // Check number
    if count, ok := obj["count"].(float64); !ok || count != 42 {
        t.Errorf("expected count=42, got %v", obj["count"])
    }
}
```

Create `internal/parser/grammar_test.go`:

```go
package parser

import (
    "testing"

    "github.com/shapestone/shape/pkg/grammar"
)

func TestGrammarVerification(t *testing.T) {
    // Load our grammar
    spec, err := grammar.ParseEBNF("../../docs/grammar/json.ebnf")
    if err != nil {
        t.Fatalf("failed to parse grammar: %v", err)
    }

    // Generate tests from grammar
    tests := spec.GenerateTests(grammar.TestOptions{
        MaxDepth:      5,
        CoverAllRules: true,
        EdgeCases:     true,
        InvalidCases:  true,
    })

    // Verify parser against grammar
    for _, test := range tests {
        t.Run(test.Name, func(t *testing.T) {
            p := NewParser(test.Input)
            result, err := p.Parse()

            if test.ShouldSucceed {
                if err != nil {
                    t.Errorf("grammar says valid, parser rejected: %v\nInput: %s",
                        err, test.Input)
                }
                if result == nil {
                    t.Errorf("grammar says valid, parser returned nil\nInput: %s",
                        test.Input)
                }
            } else {
                if err == nil {
                    t.Errorf("grammar says invalid, parser accepted\nInput: %s",
                        test.Input)
                }
            }
        })
    }
}

func TestGrammarCoverage(t *testing.T) {
    spec, err := grammar.ParseEBNF("../../docs/grammar/json.ebnf")
    if err != nil {
        t.Fatalf("failed to parse grammar: %v", err)
    }

    tracker := grammar.NewCoverageTracker(spec)

    // Run all tests and track coverage
    // (In real implementation, parser would register rule invocations)

    // Verify 100% coverage
    coverage := tracker.Report()
    if coverage.Percentage < 100.0 {
        t.Errorf("grammar coverage: %.1f%% (expected 100%%)\nMissing rules: %v",
            coverage.Percentage, coverage.UncoveredRules)
    }
}
```

---

## Testing Strategy

### Test Pyramid

```
           ┌─────────────────┐
           │  Grammar Tests  │  (Auto-generated, comprehensive)
           │   100% Coverage │
           └────────┬────────┘
                    │
         ┌──────────▼──────────┐
         │   Manual Tests      │  (Specific scenarios, error cases)
         │   Edge Cases        │
         └──────────┬──────────┘
                    │
    ┌───────────────▼───────────────┐
    │     Unit Tests                │  (Tokenizer, helpers)
    └───────────────────────────────┘
```

### 1. Unit Tests (Tokenizer, Utilities)

Test tokenization:

```go
func TestTokenizer_Basic(t *testing.T) {
    tok := tokenizer.NewTokenizer()
    tok.Initialize(`{ "id": "abc123" }`)

    expected := []string{
        tokenizer.TokenLBrace,
        tokenizer.TokenString,
        tokenizer.TokenColon,
        tokenizer.TokenString,
        tokenizer.TokenRBrace,
        tokenizer.TokenEOF,
    }

    tokens, _ := tok.Tokenize()

    for i, exp := range expected {
        if i >= len(tokens) {
            t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
        }
        if tokens[i].Kind() != exp {
            t.Errorf("token %d: expected %s, got %s",
                i, exp, tokens[i].Kind())
        }
    }
}
```

### 2. Manual Parser Tests

Test specific scenarios and edge cases (see Step 6 above).

### 3. Grammar-Based Tests

Auto-generate comprehensive tests from grammar (see Step 6 above).

### 4. Coverage Requirements

- **Unit tests:** 90%+ coverage
- **Parser tests:** 95%+ coverage
- **Grammar coverage:** 100% (all rules exercised)

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage threshold
go tool cover -func=coverage.out | grep total
```

---

## Documentation Requirements

### 1. README.md

```markdown
# shape-json - JSON Parser

JSON data format parser for the Shape ecosystem.

## Installation

\`\`\`bash
go get github.com/shapestone/shape-json
\`\`\`

## Usage

\`\`\`go
import "github.com/shapestone/shape-json/pkg/json"

data := \`{ "id": "abc123", "age": 30 }\`
result, err := json.Parse(data)
if err != nil {
    // handle error
}

// Result is interface{} - type assert to access
obj := result.(map[string]interface{})
id := obj["id"].(string)
age := obj["age"].(float64)
\`\`\`

## Return Types

Parse() returns Go types based on JSON structure:

- Objects → `map[string]interface{}`
- Arrays → `[]interface{}`
- Strings → `string`
- Numbers → `float64`
- Booleans → `bool`
- Null → `nil`

## Grammar

See [docs/grammar/json.ebnf](docs/grammar/json.ebnf) for the complete EBNF specification.

## Documentation

- [Grammar Specification](docs/grammar/json.ebnf)
- [Examples](docs/examples/)
- [Shape Infrastructure](https://github.com/shapestone/shape)

## License

Apache 2.0
```

### 2. Package Documentation

```go
// Package json provides parsing for JSON data format.
//
// JSON is a lightweight data-interchange format (RFC 8259).
//
// This parser returns Go types (map, slice, primitives), NOT AST nodes.
// Shape's AST nodes are for validation schemas only.
//
// Grammar: See docs/grammar/json.ebnf for complete specification.
//
// This parser uses LL(1) recursive descent parsing (see Shape ADR 0004).
// Grammar-based tests verify parser correctness (see Shape ADR 0005).
//
// Example:
//   data := `{ "id": "abc123", "age": 30 }`
//   result, err := json.Parse(data)
//   obj := result.(map[string]interface{})
//   id := obj["id"].(string)
package json
```

### 3. Function Documentation with Grammar

```go
// parseObject parses an object and returns a Go map.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns: map[string]interface{} with parsed properties
// Accepts empty objects: {}
// Requires quoted property keys: { "id": "value" }
// Rejects unquoted keys: { id: "value" }
func (p *Parser) parseObject() (map[string]interface{}, error)
```

---

## CI/CD Setup

### GitHub Actions Workflow

`.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 95.0" | bc -l) )); then
            echo "Coverage below 95%"
            exit 1
          fi

      - name: Grammar verification tests
        run: go test -v ./internal/parser -run TestGrammar

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Build
        run: go build ./...

  grammar:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Verify grammar file exists
        run: |
          if [ ! -f docs/grammar/json.ebnf ]; then
            echo "Grammar file missing: docs/grammar/json.ebnf"
            exit 1
          fi

      - name: Verify grammar tests
        run: |
          if ! grep -q "TestGrammarVerification" internal/parser/grammar_test.go; then
            echo "Grammar verification tests missing"
            exit 1
          fi
```

### Makefile

```makefile
.PHONY: test lint build coverage grammar-tests

test:
	go test -v -race ./...

lint:
	golangci-lint run

build:
	go build ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

grammar-tests:
	go test -v ./internal/parser -run TestGrammar
	@echo "Grammar verification: PASSED"

all: test lint build coverage grammar-tests
```

---

## Complete Working Example

### Minimal Working Parser

Here's a complete minimal parser for a simple key-value format:

**Grammar (`docs/grammar/simple.ebnf`):**

```ebnf
// Simple key-value format
// Example: name = "Alice"

Pair = Key "=" Value ;
Key = [a-zA-Z_]+ ;
Value = String | Number ;
String = '"' [^"]* '"' ;
Number = [0-9]+ ;
```

**Parser (`pkg/simple/parser.go`):**

```go
package simple

import (
    "fmt"
    "strconv"

    "github.com/shapestone/shape/pkg/tokenizer"
)

// Token types
const (
    TokenKey    = "Key"
    TokenEquals = "Equals"
    TokenString = "String"
    TokenNumber = "Number"
    TokenEOF    = "EOF"
)

// Parse parses simple key=value format
// Returns: map[string]interface{} with key → value
//
// Example: name = "Alice" → {"name": "Alice"}
// Example: age = 30 → {"age": 30.0}
func Parse(input string) (map[string]interface{}, error) {
    // Create tokenizer
    tok := tokenizer.NewTokenizer(
        tokenizer.WhitespaceMatcherFunc(),
        tokenizer.StringMatcherFunc(TokenEquals, "="),
        tokenizer.RegexMatcherFunc(TokenString, `"[^"]*"`),
        tokenizer.RegexMatcherFunc(TokenNumber, `[0-9]+`),
        tokenizer.RegexMatcherFunc(TokenKey, `[a-zA-Z_]+`),
    )
    tok.Initialize(input)

    // Parse: Key "=" Value
    result := make(map[string]interface{})

    // Key
    keyToken, err := tok.NextToken()
    if err != nil {
        return nil, err
    }
    if keyToken.Kind() != TokenKey {
        return nil, fmt.Errorf("expected key, got %s", keyToken.Kind())
    }
    key := keyToken.ValueString()

    // "="
    eqToken, err := tok.NextToken()
    if err != nil {
        return nil, err
    }
    if eqToken.Kind() != TokenEquals {
        return nil, fmt.Errorf("expected '=', got %s", eqToken.Kind())
    }

    // Value (String or Number)
    valueToken, err := tok.NextToken()
    if err != nil {
        return nil, err
    }

    var value interface{}
    switch valueToken.Kind() {
    case TokenString:
        // Remove quotes
        s := valueToken.ValueString()
        value = s[1 : len(s)-1]
    case TokenNumber:
        num, err := strconv.ParseFloat(valueToken.ValueString(), 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number: %w", err)
        }
        value = num
    default:
        return nil, fmt.Errorf("expected string or number, got %s", valueToken.Kind())
    }

    result[key] = value
    return result, nil
}
```

**Test (`pkg/simple/parser_test.go`):**

```go
package simple

import "testing"

func TestParse_String(t *testing.T) {
    input := `name = "Alice"`
    result, err := Parse(input)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(result) != 1 {
        t.Fatalf("expected 1 entry, got %d", len(result))
    }

    if result["name"] != "Alice" {
        t.Errorf("expected name='Alice', got %v", result["name"])
    }
}

func TestParse_Number(t *testing.T) {
    input := `age = 30`
    result, err := Parse(input)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result["age"] != 30.0 {
        t.Errorf("expected age=30.0, got %v", result["age"])
    }
}
```

---

## Checklist for New Parser Project

### Project Setup
- [ ] Create repository `shape-{format}`
- [ ] Initialize Go module
- [ ] Add Shape dependency
- [ ] Create directory structure
- [ ] Add LICENSE (Apache 2.0)
- [ ] Create README.md

### Grammar
- [ ] Write EBNF grammar in `docs/grammar/{format}.ebnf`
- [ ] Add implementation hints specifying Go return types
- [ ] Include examples of valid/invalid syntax
- [ ] Specify that parser returns Go types, NOT AST nodes

### Implementation
- [ ] Implement tokenizer with token definitions
- [ ] Implement parser returning Go types (map, slice, primitives)
- [ ] Add grammar fragments to function comments
- [ ] Implement public API in `pkg/{format}/`
- [ ] Add position tracking for error messages
- [ ] Implement context-aware error messages
- [ ] VERIFY: Parser returns `interface{}`, `map[string]interface{}`, `[]interface{}`
- [ ] VERIFY: Parser does NOT return `*ast.ArrayNode`, `*ast.ObjectNode`, etc.

### Testing
- [ ] Write unit tests for tokenizer
- [ ] Write manual parser tests for specific scenarios
- [ ] Write error handling tests
- [ ] Implement grammar-based verification tests
- [ ] Verify 95%+ test coverage
- [ ] Verify 100% grammar coverage
- [ ] Test that returned types are Go maps/slices/primitives

### Documentation
- [ ] Complete README.md with usage examples
- [ ] Document grammar specification
- [ ] Add godoc comments to all public APIs
- [ ] Add runnable examples in `examples/`
- [ ] Document return types (map, slice, primitives)
- [ ] Reference ADR 0004 and ADR 0005

### CI/CD
- [ ] Set up GitHub Actions for CI
- [ ] Add coverage checking (95%+ threshold)
- [ ] Add linting (golangci-lint)
- [ ] Add grammar verification tests in CI
- [ ] Create Makefile for local development

### Release
- [ ] Tag v0.1.0 (initial release)
- [ ] Publish to pkg.go.dev
- [ ] Update Shape ecosystem documentation

---

## Additional Resources

- **Shape Infrastructure:** https://github.com/shapestone/shape
- **ADR 0004:** LL(1) Recursive Descent Parser Strategy
- **ADR 0005:** Grammar-as-Verification for Parser Correctness
- **Example Parser:** https://github.com/shapestone/shape-json (when available)

---

## Common Pitfalls to Avoid

### DO NOT:

1. **Use AST nodes for parsed data**
   ```go
   // WRONG!
   return ast.NewArrayNode(elements, pos)  // COMPILE ERROR
   return ast.NewObjectNode(properties, pos)  // COMPILE ERROR
   ```

2. **Try to create a "generic data AST"**
   - Shape's AST is for validation schemas only
   - Use Go's built-in types for data (map, slice, primitives)

3. **Mix validation and parsing**
   - Parsing: format → Go types
   - Validation: Go types → schema check
   - Keep these separate

### DO:

1. **Return Go types from parser**
   ```go
   // CORRECT!
   return []interface{}{1, 2, 3}, nil
   return map[string]interface{}{"id": "abc"}, nil
   return "hello", nil
   return 42.0, nil
   ```

2. **Use AST nodes for validation schemas**
   ```go
   // CORRECT! (but in validator, not parser)
   schema := ast.NewArrayNode(ast.NewTypeNode("string", pos), pos)
   err := validator.Validate(parsedData, schema)
   ```

3. **Keep parser projects focused**
   - One format per project
   - Returns Go types
   - Independent versioning

---

## Support

For questions or issues:
- **Shape Issues:** https://github.com/shapestone/shape/issues
- **Format-specific Issues:** https://github.com/shapestone/shape-{format}/issues

---

**This guide ensures all Shape parser projects follow correct architecture patterns, distinguishing between data parsing (returns Go types) and validation schemas (uses AST nodes).**
