# Parser Implementation Guide

**Audience:** Developers and LLMs implementing new data format parsers for the Shape ecosystem

**Purpose:** This guide ensures all parser projects follow Shape's architecture, patterns, and quality standards.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Parser Project Structure](#parser-project-structure)
3. [Implementation Steps](#implementation-steps)
4. [EBNF Grammar Specification](#ebnf-grammar-specification)
5. [Tokenizer Implementation](#tokenizer-implementation)
6. [Parser Implementation](#parser-implementation)
7. [Testing Strategy](#testing-strategy)
8. [Documentation Requirements](#documentation-requirements)
9. [CI/CD Setup](#cicd-setup)
10. [Complete Example](#complete-example)

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
│  - Maintains its own EBNF grammar                    │
│  - Implements its own parser                         │
│  - Tests itself using Shape's infrastructure         │
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
│  - pkg/ast/* - AST node definitions                  │
│  - pkg/tokenizer/* - Tokenization framework          │
│  - pkg/parser/* - Parser interface, Format enum      │
│  - pkg/validator/* - Schema validation framework     │
│  - pkg/grammar/* - Grammar verification tools        │
│                                                      │
│  Does NOT provide:                                   │
│  - Format-specific parsers                           │
│  - Format-specific tests                             │
│  - Convenience Parse() functions                     │
└──────────────────────────────────────────────────────┘
```

### Key Principles

1. **Shape is Infrastructure Only**
   - Provides reusable components
   - No format-specific implementations
   - No dependencies on parser projects

2. **Parser Projects are Self-Contained**
   - Own repository, versioning, releases
   - Own EBNF grammar in `docs/grammar/`
   - Own tests, documentation, CI/CD
   - Import Shape as a dependency

3. **Grammar-Driven Development**
   - EBNF grammar is source of truth
   - Grammar generates verification tests
   - Grammar guides LLM-assisted development
   - Grammar fragments document code

4. **Hand-Coded Parsers**
   - NOT using parser generators or meta-grammars
   - Freedom to choose parsing technique per format
   - LL(1) recursive descent recommended (default)
   - Optimized for performance and error quality

### Parser Technology Freedom

**Critical insight:** Hand-coded parsers give you complete freedom to choose the best parsing technique for your format. You are NOT locked into one approach.

**Available techniques:**

- **LL(1) Recursive Descent** (recommended default)
  - O(n) linear time complexity
  - Zero backtracking overhead
  - Best error messages (full context available)
  - Most debuggable (call stack = parse tree)
  - Simplest implementation
  - **Used by:** All current Shape parser projects

- **Pratt Parsing** (for operator precedence)
  - Elegant handling of operator precedence
  - More concise than precedence climbing
  - Still hand-coded with full error control
  - **Use when:** Format has complex operator precedence rules (e.g., expression languages)

- **Packrat/PEG** (for memoization)
  - Handles expensive backtracking patterns efficiently
  - Memoization provides O(n) guarantee
  - More complex but handles harder grammars
  - **Use when:** Format has ambiguities requiring backtracking

- **Parser Combinators** (functional composition)
  - Higher-order functions compose parsers
  - Grammar-like code structure
  - Type-safe composition
  - **Trade-off:** Harder error messages, but more compositional
  - **Use when:** Prefer functional programming style

- **Hand-Optimized Hybrid** (mix techniques)
  - LL(1) for most rules
  - Pratt for expression parsing
  - Optimize hot paths independently
  - **Use when:** Different parts of grammar have different needs

**This guide focuses on LL(1) recursive descent** because:
- It's the simplest and most maintainable
- All current Shape parsers use it successfully
- Best error messages and debuggability
- Sufficient for most data formats

**However:** If your format has specific needs (operator precedence, ambiguity, etc.), you're free to choose a different technique. Grammar-based verification works with ANY hand-coded parser approach.

**See ADR 0004** for detailed analysis of parser technology options.

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
│       ├── parser.go              # Public API
│       └── parser_test.go         # Public API tests
│
├── internal/
│   ├── tokenizer/
│   │   ├── tokenizer.go           # Format-specific tokenizer
│   │   ├── tokenizer_test.go
│   │   └── tokens.go              # Token type definitions
│   │
│   └── parser/
│       ├── parser.go              # Parser implementation
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

## Implementation Steps

### 1. Create Repository and Module

```bash
# Create repository
mkdir shape-json
cd shape-json

# Initialize Go module
go mod init github.com/shapestone/shape-json

# Add Shape dependency
go get github.com/shapestone/shape@latest
```

### 2. Define EBNF Grammar

Create `docs/grammar/json.ebnf` with your format's grammar:

```ebnf
// JSON Grammar Specification
// This grammar defines the JSON data format (RFC 8259).
//
// Implementation Guide:
// - Use LL(1) recursive descent parsing (see Shape ADR 0004)
// - Each production rule becomes a parse function
// - Return appropriate ast.Node types
// - Provide context-aware error messages

// Top-level value
// Parser function: parseValue() -> ast.Node
Value = Object | Array | String | Number | Boolean | Null ;

// Object with properties
// Parser function: parseObject() -> *ast.ObjectNode
// Example: { "id": "abc123", "name": "Alice" }
Object = "{" [ Property { "," Property } ] "}" ;

// Property key-value pair
// Parser function: parseProperty() -> (string, ast.Node)
Property = String ":" Value ;

// Array of values
// Parser function: parseArray() -> *ast.ArrayNode
// Example: [1, 2, 3]
Array = "[" [ Value { "," Value } ] "]" ;

// Primitive literals
// Parser functions: parseString(), parseNumber(), parseBoolean(), parseNull()
String = '"' [^"]* '"' ;
Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;
Boolean = "true" | "false" ;
Null = "null" ;
```

**Key elements:**
- Comments explain each rule
- Implementation hints for parser functions
- AST node types to return
- Examples of valid syntax

### 3. Implement Tokenizer

Create `internal/tokenizer/tokenizer.go`:

```go
package tokenizer

import (
    "github.com/shapestone/shape/pkg/tokenizer"
)

// Token types
const (
    TokenLBrace   = "LBrace"
    TokenRBrace   = "RBrace"
    TokenLBracket = "LBracket"
    TokenRBracket = "RBracket"
    TokenColon    = "Colon"
    TokenComma    = "Comma"
    TokenString   = "String"
    TokenNumber   = "Number"
    TokenTrue     = "True"
    TokenFalse    = "False"
    TokenNull     = "Null"
    TokenEOF      = "EOF"
)

// NewTokenizer creates a tokenizer for JSON format.
func NewTokenizer() *tokenizer.Tokenizer {
    return tokenizer.NewTokenizer(
        // Whitespace (skip)
        tokenizer.WhitespaceMatcherFunc(),

        // Keywords (before identifiers to match specifically)
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

        // String literals
        tokenizer.RegexMatcherFunc(TokenString, `"(?:[^"\\]|\\.)*"`),

        // Numbers (integers, floats, scientific notation)
        tokenizer.RegexMatcherFunc(TokenNumber, `-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?`),
    )
}
```

**Pattern:**
- Use Shape's tokenizer framework
- Define token type constants
- Create matchers for your grammar's terminals
- Order matters: specific before general (e.g., keywords before numbers)

### 4. Implement Parser

Create `internal/parser/parser.go` using LL(1) recursive descent:

```go
package parser

import (
    "fmt"
    "strconv"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/pkg/parser"
    "github.com/shapestone/shape-json/internal/tokenizer"
)

// Parser implements LL(1) recursive descent parsing for JSON.
type Parser struct {
    tokenizer *tokenizer.Tokenizer
    current   *tokenizer.Token
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

// Parse parses the input and returns an AST.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
func (p *Parser) Parse() (ast.Node, error) {
    return p.parseValue()
}

// parseValue dispatches to specific parse functions.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
//
// Dispatch based on token type (LL(1) predictive parsing).
func (p *Parser) parseValue() (ast.Node, error) {
    switch p.peek().Kind() {
    case TokenLBrace:     // Object
        return p.parseObject()
    case TokenLBracket:   // Array
        return p.parseArray()
    case TokenString:     // String literal
        return p.parseString()
    case TokenNumber:     // Number literal
        return p.parseNumber()
    case TokenTrue, TokenFalse:  // Boolean literal
        return p.parseBoolean()
    case TokenNull:       // Null literal
        return p.parseNull()
    default:
        return nil, fmt.Errorf("expected value at %s, got %s",
            p.position(), p.peek().Kind())
    }
}

// parseObject parses an object node.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns *ast.ObjectNode with properties map.
// Example valid: { "id": "abc123", "name": "Alice" }
// Example valid: {} (empty object)
// Example invalid: { id: "value" } (missing quotes on key)
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    startPos := p.position()

    if _, err := p.expect(TokenLBrace); err != nil {  // "{"
        return nil, err
    }

    properties := make(map[string]ast.Node)

    // [ Property { "," Property } ]  - Optional property list
    if p.peek().Kind() != TokenRBrace {
        // First property
        key, value, err := p.parseProperty()  // Property
        if err != nil {
            return nil, err
        }
        properties[key] = value

        // Additional properties
        for p.peek().Kind() == TokenComma {  // { "," Property }
            p.advance()  // consume ","

            key, value, err := p.parseProperty()  // Property
            if err != nil {
                return nil, fmt.Errorf("in object property after comma: %w", err)
            }
            properties[key] = value
        }
    }

    if _, err := p.expect(TokenRBrace); err != nil {  // "}"
        return nil, err
    }

    return ast.NewObjectNode(properties, startPos), nil
}

// parseProperty parses a property key-value pair.
//
// Grammar:
//   Property = String ":" Value ;
//
// Returns (key string, value ast.Node).
func (p *Parser) parseProperty() (string, ast.Node, error) {
    // String
    keyToken, err := p.expect(TokenString)
    if err != nil {
        return "", nil, fmt.Errorf("property key must be string literal: %w", err)
    }
    key := p.unquoteString(keyToken.Value())

    // ":"
    if _, err := p.expect(TokenColon); err != nil {
        return "", nil, err
    }

    // Value
    value, err := p.parseValue()
    if err != nil {
        return "", nil, fmt.Errorf("in property value for %q: %w", key, err)
    }

    return key, value, nil
}

// parseArray parses an array node.
//
// Grammar:
//   Array = "[" [ Value { "," Value } ] "]" ;
//
// Returns *ast.ArrayNode with elements.
// Example: [1, 2, 3]
func (p *Parser) parseArray() (*ast.ArrayNode, error) {
    startPos := p.position()

    if _, err := p.expect(TokenLBracket); err != nil {  // "["
        return nil, err
    }

    var elements []ast.Node

    // [ Value { "," Value } ]  - Optional value list
    if p.peek().Kind() != TokenRBracket {
        // First element
        elem, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        elements = append(elements, elem)

        // Additional elements
        for p.peek().Kind() == TokenComma {
            p.advance()  // consume ","

            elem, err := p.parseValue()
            if err != nil {
                return nil, fmt.Errorf("in array element after comma: %w", err)
            }
            elements = append(elements, elem)
        }
    }

    if _, err := p.expect(TokenRBracket); err != nil {  // "]"
        return nil, err
    }

    return ast.NewArrayNode(elements, startPos), nil
}

// parseString parses a string literal.
//
// Grammar:
//   String = '"' [^"]* '"' ;
func (p *Parser) parseString() (*ast.LiteralNode, error) {
    token, err := p.expect(TokenString)
    if err != nil {
        return nil, err
    }
    value := p.unquoteString(token.Value())
    return ast.NewLiteralNode(value, token.Position()), nil
}

// parseNumber parses a number literal.
//
// Grammar:
//   Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;
func (p *Parser) parseNumber() (*ast.LiteralNode, error) {
    token, err := p.expect(TokenNumber)
    if err != nil {
        return nil, err
    }
    num, err := strconv.ParseFloat(token.Value(), 64)
    if err != nil {
        return nil, fmt.Errorf("invalid number: %w", err)
    }
    return ast.NewLiteralNode(num, token.Position()), nil
}

// parseBoolean parses a boolean literal.
//
// Grammar:
//   Boolean = "true" | "false" ;
func (p *Parser) parseBoolean() (*ast.LiteralNode, error) {
    token := p.peek()
    var value bool
    if token.Kind() == TokenTrue {
        value = true
    } else if token.Kind() == TokenFalse {
        value = false
    } else {
        return nil, fmt.Errorf("expected boolean, got %s", token.Kind())
    }
    pos := token.Position()
    p.advance()
    return ast.NewLiteralNode(value, pos), nil
}

// parseNull parses a null literal.
//
// Grammar:
//   Null = "null" ;
func (p *Parser) parseNull() (*ast.LiteralNode, error) {
    token, err := p.expect(TokenNull)
    if err != nil {
        return nil, err
    }
    return ast.NewLiteralNode(nil, token.Position()), nil
}

// Helper methods

// peek returns current token without advancing.
func (p *Parser) peek() *tokenizer.Token {
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
func (p *Parser) expect(kind string) (*tokenizer.Token, error) {
    if p.peek().Kind() != kind {
        return nil, fmt.Errorf("expected %s, got %s at %s",
            kind, p.peek().Kind(), p.position())
    }
    token := p.current
    p.advance()
    return token, nil
}

// position returns current position for error reporting.
func (p *Parser) position() ast.Position {
    if p.hasToken {
        return ast.Position{
            Line:   p.current.Position().Line,
            Column: p.current.Position().Column,
        }
    }
    return ast.Position{}
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
- Each grammar rule → parse function
- Grammar fragment in function comment
- Single token lookahead (`peek()`)
- Inline comments mark grammar elements (`// "{"`)
- Context-aware error messages
- Return appropriate AST node types

### 5. Implement Public API

Create `pkg/json/parser.go`:

```go
package json

import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape-json/internal/parser"
)

// Parse parses JSON data format into an AST.
//
// Example:
//   data := `{ "id": "abc123", "age": 30 }`
//   node, err := json.Parse(data)
func Parse(input string) (ast.Node, error) {
    p := parser.NewParser(input)
    return p.Parse()
}

// Format returns the format identifier for this parser.
func Format() string {
    return "JSON"
}
```

### 6. Implement Tests

**Manual tests** (`internal/parser/parser_test.go`):

```go
package parser

import (
    "testing"
    "github.com/shapestone/shape/pkg/ast"
)

func TestParseObject_Empty(t *testing.T) {
    input := `{}`
    p := NewParser(input)
    node, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj, ok := node.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode, got %T", node)
    }

    if len(obj.Properties) != 0 {
        t.Errorf("expected empty object, got %d properties", len(obj.Properties))
    }
}

func TestParseObject_SingleProperty(t *testing.T) {
    input := `{ "id": "abc123" }`
    p := NewParser(input)
    node, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj := node.(*ast.ObjectNode)
    if len(obj.Properties) != 1 {
        t.Fatalf("expected 1 property, got %d", len(obj.Properties))
    }

    idValue, ok := obj.Properties["id"].(*ast.LiteralNode)
    if !ok {
        t.Errorf("expected LiteralNode for 'id', got %T", obj.Properties["id"])
    }
    if str, ok := idValue.Value.(string); !ok || str != "abc123" {
        t.Errorf("expected 'abc123', got %v", idValue.Value)
    }
}

func TestParseArray(t *testing.T) {
    input := `[1, 2, 3]`
    p := NewParser(input)
    node, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr, ok := node.(*ast.ArrayNode)
    if !ok {
        t.Fatalf("expected *ast.ArrayNode, got %T", node)
    }

    if len(arr.Elements) != 3 {
        t.Errorf("expected 3 elements, got %d", len(arr.Elements))
    }
}
```

**Grammar-based tests** (`internal/parser/grammar_test.go`):

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
        MaxDepth:        5,
        CoverAllRules:   true,
        EdgeCases:       true,
        InvalidCases:    true,
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

    // Run all tests (from parser_test.go)
    // ... (tests would register rule invocations with tracker)

    // Verify 100% coverage
    coverage := tracker.Report()
    if coverage.Percentage < 100.0 {
        t.Errorf("grammar coverage: %.1f%% (expected 100%%)\nMissing rules: %v",
            coverage.Percentage, coverage.UncoveredRules)
    }
}
```

---

## EBNF Grammar Specification

### Grammar Format (Custom EBNF Variant)

**Not ISO 14977 compliant.** We use a pragmatic variant optimized for readability.

```ebnf
// Production rules
rule_name = expression ;

// Operators
[ ]     Optional (0 or 1)
+       One or more (suffix, like regex)
*       Zero or more (suffix, like regex)
{ }     Zero or more (alternative notation)
( )     Grouping
|       Alternation (OR)

// Character notation (regex-like)
Digit = [0-9] ;
Letter = [a-zA-Z] ;
Hex = [0-9a-fA-F] ;

// Concatenation (no commas)
Rule = "keyword" Identifier Number ;
```

### Implementation Hints in Grammar

Include comments to guide implementation:

```ebnf
// Object with key-value pairs
// Parser function: parseObject() -> *ast.ObjectNode
// Returns: ast.NewObjectNode(properties map[string]ast.Node, position)
// Example valid: { "id": "abc123", "name": "Alice" }
// Example invalid: { id: "value" } (missing quotes on key)
// Error message: "Property keys must be quoted strings"
Object = "{" [ Property { "," Property } ] "}" ;
```

**Include:**
- What AST node to return
- Examples of valid/invalid input
- Error message suggestions
- Edge cases to handle

---

## Tokenizer Implementation

### Token Definition Pattern

```go
package tokenizer

const (
    // Keywords (if any)
    TokenTrue  = "True"
    TokenFalse = "False"
    TokenNull  = "Null"

    // Structural
    TokenLBrace   = "LBrace"
    TokenRBrace   = "RBrace"
    TokenColon    = "Colon"
    TokenComma    = "Comma"

    // Literals
    TokenString = "String"
    TokenNumber = "Number"

    // Identifiers
    TokenIdentifier = "Identifier"

    // Special
    TokenEOF = "EOF"
)
```

### Matcher Ordering

**Critical:** Order matchers from most specific to most general:

```go
func NewTokenizer() *tokenizer.Tokenizer {
    return tokenizer.NewTokenizer(
        // 1. Whitespace (skip)
        tokenizer.WhitespaceMatcherFunc(),

        // 2. Keywords (before identifiers!)
        tokenizer.StringMatcherFunc(TokenTrue, "true"),
        tokenizer.StringMatcherFunc(TokenFalse, "false"),
        tokenizer.StringMatcherFunc(TokenNull, "null"),

        // 3. Multi-character operators (before single-char)
        tokenizer.StringMatcherFunc(TokenArrow, "=>"),
        tokenizer.StringMatcherFunc(TokenEqualEqual, "=="),

        // 4. Single-character operators
        tokenizer.StringMatcherFunc(TokenLBrace, "{"),
        tokenizer.StringMatcherFunc(TokenColon, ":"),

        // 5. Complex literals (strings, numbers)
        tokenizer.RegexMatcherFunc(TokenString, `"(?:[^"\\]|\\.)*"`),
        tokenizer.RegexMatcherFunc(TokenNumber, `-?[0-9]+(\.[0-9]+)?`),

        // 6. Identifiers (last, most general)
        tokenizer.RegexMatcherFunc(TokenIdentifier, `[a-zA-Z_][a-zA-Z0-9_]*`),
    )
}
```

### Testing Tokenizer

```go
func TestTokenizer_Basic(t *testing.T) {
    tok := NewTokenizer()
    tok.Initialize(`{ "id": "abc123" }`)

    expected := []string{
        TokenLBrace,   // {
        TokenString,   // "id"
        TokenColon,    // :
        TokenString,   // "abc123"
        TokenRBrace,   // }
        TokenEOF,
    }

    for i, exp := range expected {
        token, err := tok.NextToken()
        if err != nil {
            t.Fatalf("token %d: unexpected error: %v", i, err)
        }
        if token.Kind() != exp {
            t.Errorf("token %d: expected %s, got %s",
                i, exp, token.Kind())
        }
    }
}
```

---

## Parser Implementation

**Note:** This section demonstrates **LL(1) recursive descent parsing** as the recommended default technique. However, you have **complete freedom** to use other parsing techniques (Pratt, PEG, combinators, hybrid) if your format requires it. See [Parser Technology Freedom](#parser-technology-freedom) above.

Grammar-based verification works with ANY hand-coded parser approach.

### LL(1) Recursive Descent Pattern

**See ADR 0004 for full details on LL(1) parsing strategy.**

#### Parser Structure

```go
type Parser struct {
    tokenizer *tokenizer.Tokenizer
    current   *tokenizer.Token  // Single token lookahead
    hasToken  bool
}
```

#### Core Methods

```go
// peek returns current token without advancing
func (p *Parser) peek() *tokenizer.Token

// advance moves to next token
func (p *Parser) advance() error

// expect consumes token of expected kind
func (p *Parser) expect(kind string) (*tokenizer.Token, error)

// position returns current position for errors
func (p *Parser) position() ast.Position
```

#### Parse Function Pattern

```go
// parseX parses grammar rule X.
//
// Grammar:
//   X = ... ;
//
// Implementation notes, examples, etc.
func (p *Parser) parseX() (*ast.XNode, error) {
    startPos := p.position()

    // Parse according to grammar structure
    // ...

    return ast.NewXNode(..., startPos), nil
}
```

#### Predictive Dispatch (Alternatives)

```go
// Grammar: Value = Object | Array | String | Number | Boolean | Null
func (p *Parser) parseValue() (ast.Node, error) {
    switch p.peek().Kind() {
    case TokenLBrace:
        return p.parseObject()
    case TokenLBracket:
        return p.parseArray()
    case TokenString:
        return p.parseString()
    case TokenNumber:
        return p.parseNumber()
    case TokenTrue, TokenFalse:
        return p.parseBoolean()
    case TokenNull:
        return p.parseNull()
    default:
        return nil, fmt.Errorf("expected value, got %s", p.peek().Kind())
    }
}
```

#### Repetition Patterns

**Optional: `[ ... ]`**
```go
// Grammar: [ Property ]
if p.peek().Kind() == expectedToken {
    prop, err := p.parseProperty()
    if err != nil {
        return nil, err
    }
}
```

**Zero or more: `{ ... }`**
```go
// Grammar: { Property }
for p.peek().Kind() == expectedToken {
    prop, err := p.parseProperty()
    if err != nil {
        return nil, err
    }
    properties = append(properties, prop)
}
```

**One or more: `Property+`**
```go
// Grammar: Property+
// First occurrence (required)
prop, err := p.parseProperty()
if err != nil {
    return nil, err
}
properties = append(properties, prop)

// Additional occurrences (zero or more)
for p.peek().Kind() == expectedToken {
    prop, err := p.parseProperty()
    if err != nil {
        return nil, err
    }
    properties = append(properties, prop)
}
```

#### Error Handling

**Context-aware messages:**
```go
if err != nil {
    return nil, fmt.Errorf("in object property after comma: %w", err)
}
```

**Hints and suggestions:**
```go
if p.peek().Kind() == TokenIdentifier {
    return nil, fmt.Errorf("property key must be quoted string at %s, got identifier %q (did you forget quotes?)",
        p.position(), p.peek().Value())
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

Test individual components:

```go
func TestTokenizer_String(t *testing.T) {
    tok := NewTokenizer()
    tok.Initialize(`"hello world"`)

    token, err := tok.NextToken()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if token.Kind() != TokenString {
        t.Errorf("expected TokenString, got %s", token.Kind())
    }

    if token.Value() != `"hello world"` {
        t.Errorf("expected quoted string, got %s", token.Value())
    }
}
```

### 2. Manual Parser Tests (Specific Scenarios)

Test specific features and error cases:

```go
func TestParseObject_MultipleProperties(t *testing.T) {
    input := `{
        "id": "abc123",
        "name": "Alice",
        "age": 30
    }`

    node, err := parser.NewParser(input).Parse()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj := node.(*ast.ObjectNode)
    if len(obj.Properties) != 3 {
        t.Errorf("expected 3 properties, got %d", len(obj.Properties))
    }

    // Verify each property value
    if lit, ok := obj.Properties["id"].(*ast.LiteralNode); !ok {
        t.Error("'id' should be LiteralNode")
    } else if str, ok := lit.Value.(string); !ok || str != "abc123" {
        t.Errorf("expected 'abc123', got %v", lit.Value)
    }

    if lit, ok := obj.Properties["age"].(*ast.LiteralNode); !ok {
        t.Error("'age' should be LiteralNode")
    } else if num, ok := lit.Value.(float64); !ok || num != 30 {
        t.Errorf("expected 30, got %v", lit.Value)
    }
}

func TestParseObject_ErrorUnquotedKey(t *testing.T) {
    input := `{ id: "value" }`

    _, err := parser.NewParser(input).Parse()
    if err == nil {
        t.Fatal("expected error for unquoted property key")
    }

    if !strings.Contains(err.Error(), "quoted") {
        t.Errorf("error should mention quoted strings: %v", err)
    }
}
```

### 3. Grammar-Based Tests (Comprehensive)

Auto-generate from EBNF grammar:

```go
func TestGrammarVerification(t *testing.T) {
    spec, err := grammar.ParseEBNF("../../docs/grammar/json.ebnf")
    if err != nil {
        t.Fatalf("failed to parse grammar: %v", err)
    }

    tests := spec.GenerateTests(grammar.TestOptions{
        MaxDepth:        5,      // Nesting depth
        CoverAllRules:   true,   // Exercise every production
        EdgeCases:       true,   // Empty, single, multiple
        InvalidCases:    true,   // Test error handling
    })

    for _, test := range tests {
        t.Run(test.Name, func(t *testing.T) {
            p := parser.NewParser(test.Input)
            result, err := p.Parse()

            if test.ShouldSucceed {
                if err != nil {
                    t.Errorf("Valid input rejected: %v\nInput: %s",
                        err, test.Input)
                }
            } else {
                if err == nil {
                    t.Errorf("Invalid input accepted\nInput: %s",
                        test.Input)
                }
            }
        })
    }
}
```

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
go tool cover -func=coverage.out | grep total | awk '{print $3}'
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

data := `{ "id": "abc123", "age": 30 }`
ast, err := json.Parse(data)
if err != nil {
    // handle error
}
\`\`\`

## Grammar

See [docs/grammar/json.ebnf](docs/grammar/json.ebnf) for the complete EBNF specification.

## Documentation

- [Grammar Specification](docs/grammar/json.ebnf)
- [Examples](docs/examples/)
- [Shape Infrastructure](https://github.com/shapestone/shape)

## License

Apache 2.0
```

### 2. Grammar Documentation

`docs/grammar/{format}.ebnf` - Already covered in earlier sections.

### 3. Code Documentation

**Package-level godoc:**

```go
// Package json provides parsing for JSON data format.
//
// JSON is a lightweight data-interchange format (RFC 8259).
//
// Grammar: See docs/grammar/json.ebnf for complete specification.
//
// This parser uses LL(1) recursive descent parsing (see Shape ADR 0004).
// Each production rule in the grammar corresponds to a parse function.
// Grammar-based tests verify parser correctness (see Shape ADR 0005).
//
// Example:
//   data := `{ "id": "abc123", "age": 30 }`
//   ast, err := json.Parse(data)
package json
```

**Function-level comments with grammar:**

```go
// parseObject parses an object node.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns *ast.ObjectNode with properties map.
// Accepts empty objects: {}
// Requires quoted property keys: { "id": "value" }
// Rejects unquoted keys: { id: "value" }
func (p *Parser) parseObject() (*ast.ObjectNode, error)
```

### 4. Examples

`examples/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "github.com/shapestone/shape-json/pkg/json"
)

func main() {
    // Simple JSON object
    data := `{
        "id": "abc123",
        "email": "alice@example.com",
        "age": 30,
        "active": true,
        "tags": ["admin", "user"]
    }`

    ast, err := json.Parse(data)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }

    fmt.Printf("Parsed AST: %+v\n", ast)
}
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

### Pre-Commit Checks

Add `Makefile`:

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

## Complete Example

### Minimal Working Parser

Here's a complete minimal parser implementation for a simple format:

```go
// File: pkg/simple/parser.go
package simple

import (
    "fmt"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/pkg/tokenizer"
)

// Token types
const (
    TokenString = "String"
    TokenEOF    = "EOF"
)

// Parse parses simple format: just string literals
// Grammar: Value = String ;
func Parse(input string) (ast.Node, error) {
    tok := tokenizer.NewTokenizer(
        tokenizer.WhitespaceMatcherFunc(),
        tokenizer.RegexMatcherFunc(TokenString, `"(?:[^"\\]|\\.)*"`),
    )
    tok.Initialize(input)

    token, err := tok.NextToken()
    if err != nil {
        return nil, err
    }

    if token.Kind() != TokenString {
        return nil, fmt.Errorf("expected string, got %s", token.Kind())
    }

    // Remove quotes
    value := token.Value()
    if len(value) >= 2 {
        value = value[1 : len(value)-1]
    }

    return ast.NewLiteralNode(value, ast.Position{
        Line:   token.Position().Line,
        Column: token.Position().Column,
    }), nil
}
```

```go
// File: pkg/simple/parser_test.go
package simple

import (
    "testing"
    "github.com/shapestone/shape/pkg/ast"
)

func TestParse_SimpleString(t *testing.T) {
    input := `"hello world"`
    node, err := Parse(input)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    litNode, ok := node.(*ast.LiteralNode)
    if !ok {
        t.Fatalf("expected *ast.LiteralNode, got %T", node)
    }

    str, ok := litNode.Value.(string)
    if !ok || str != "hello world" {
        t.Errorf("expected 'hello world', got %v", litNode.Value)
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
- [ ] Add implementation hints in grammar comments
- [ ] Include examples of valid/invalid syntax
- [ ] Specify AST node types to return

### Implementation
- [ ] Implement tokenizer with token definitions
- [ ] Implement parser using LL(1) recursive descent
- [ ] Add grammar fragments to function comments
- [ ] Implement public API in `pkg/{format}/`
- [ ] Add position tracking for error messages
- [ ] Implement context-aware error messages

### Testing
- [ ] Write unit tests for tokenizer
- [ ] Write manual parser tests for specific scenarios
- [ ] Write error handling tests
- [ ] Implement grammar-based verification tests
- [ ] Verify 95%+ test coverage
- [ ] Verify 100% grammar coverage

### Documentation
- [ ] Complete README.md with usage examples
- [ ] Document grammar specification
- [ ] Add godoc comments to all public APIs
- [ ] Add runnable examples in `examples/`
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
- **Example Parser:** https://github.com/shapestone/shape-json

---

## Support

For questions or issues:
- **Shape Issues:** https://github.com/shapestone/shape/issues
- **Format-specific Issues:** https://github.com/shapestone/shape-{format}/issues

---

**This guide ensures all Shape parser projects follow consistent patterns, architecture, and quality standards.**
