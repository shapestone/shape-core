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
│  - Returns universal AST nodes                       │
│  - Tests itself using Shape's grammar tools          │
│  - Is self-contained and independently versioned     │
└────────────────┬─────────────────────────────────────┘
                 │
                 │ depends on (import)
                 ▼
┌──────────────────────────────────────────────────────┐
│  Shape (Infrastructure Only)                         │
│  github.com/shapestone/shape-core                         │
│                                                      │
│  Provides:                                           │
│  - pkg/tokenizer/* - Tokenization framework          │
│  - pkg/ast/* - Universal AST (data + schemas)        │
│  - pkg/validator/* - Schema validation framework     │
│  - pkg/grammar/* - Grammar verification tools        │
│  - pkg/parser/* - Parser interface definitions       │
│                                                      │
│  Universal AST serves dual purposes:                 │
│  - Data representation (LiteralNode, ObjectNode)     │
│  - Validation schemas (TypeNode, FunctionNode)       │
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
   - Return universal AST nodes for parsed data

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

## Critical Understanding: The Universal AST

**IMPORTANT:** This is the most critical concept to understand before implementing a parser.

### Shape's AST is Universal and Format-Agnostic

Shape's AST provides a **universal, format-agnostic representation** that serves dual purposes:

#### 1. Validation Schemas
Define rules and constraints:
- `TypeNode` - Type validators (UUID, Email, Integer)
- `FunctionNode` - Parameterized validators (Integer(1, 100), String(min, max))

#### 2. Data Representation
Parsed data from any format (JSON, XML, YAML, CSV):
- `LiteralNode` - Primitives (string, number, boolean, null)
- `ObjectNode` - Structured data (JSON objects, XML elements, YAML maps)
- `ArrayNode` - Collections (JSON arrays, YAML sequences)

### Why Universal AST Instead of Go Types?

**Use Cases Requiring AST:**

1. ✅ **Error Reporting** - Source positions for precise error messages ("Error at line 5, column 12")
2. ✅ **Querying** - JSONPath/XPath navigate structure (after conversion to Go types for query execution)
3. ✅ **Diffing** - Compare two documents structurally with position information
4. ✅ **Programmatic Construction** - Build JSON/XML programmatically with proper type system
5. ✅ **Transformation** - Convert between formats (JSON ↔ XML) with intermediate representation
6. ✅ **Validation** - Validate parsed data against schemas

**Why Go Types Are Insufficient:**
- ❌ No source position tracking
- ❌ Can't distinguish null vs missing vs empty
- ❌ Can't represent format-specific features (XML attributes vs elements)
- ❌ Type information lost during parsing
- ❌ Structural validation requires rebuilding AST

### CORRECT: Parsers Return Universal AST

```go
// CORRECT: parseObject returns AST
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    properties := make(map[string]ast.SchemaNode)

    for p.hasMoreProperties() {
        key, value, err := p.parseMember()
        if err != nil {
            return nil, err
        }
        properties[key] = value  // value is ast.SchemaNode
    }

    return ast.NewObjectNode(properties, startPos), nil
}

// CORRECT: parseArray returns AST (as ObjectNode with numeric keys)
func (p *Parser) parseArray() (*ast.ObjectNode, error) {
    properties := make(map[string]ast.SchemaNode)
    index := 0

    for p.hasMoreElements() {
        elem, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        properties[strconv.Itoa(index)] = elem
        index++
    }

    return ast.NewObjectNode(properties, startPos), nil
}

// CORRECT: parsePrimitives return AST
func (p *Parser) parseString() (*ast.LiteralNode, error) {
    value, pos := p.consumeString()
    return ast.NewLiteralNode(value, pos), nil
}

func (p *Parser) parseNumber() (*ast.LiteralNode, error) {
    value, pos := p.consumeNumber()
    return ast.NewLiteralNode(value, pos), nil
}
```

### When to Use Go Types: The Dual API Pattern

While parsers return AST, you should **also provide Go type convenience APIs**:

```go
// Primary API: Parse → Universal AST
func Parse(input string) (ast.SchemaNode, error) {
    p := parser.NewParser(input)
    return p.Parse()  // Returns AST
}

// Secondary API: Unmarshal → Go types (convenience)
func Unmarshal(data []byte, v interface{}) error {
    // 1. Parse to AST
    node, err := Parse(string(data))
    if err != nil {
        return err
    }

    // 2. Convert AST → Go types and populate struct
    return unmarshalNode(node, v)
}

// Helper: Convert AST → Go types for queries
func ToGoValue(node ast.SchemaNode) interface{} {
    switch n := node.(type) {
    case *ast.LiteralNode:
        return n.Value()
    case *ast.ObjectNode:
        result := make(map[string]interface{})
        for key, val := range n.Properties() {
            result[key] = ToGoValue(val)
        }
        return result
    // ... handle other types
    }
}
```

### Summary: Universal AST for All Parsers

| Aspect | Primary API (Parse) | Secondary API (Unmarshal) |
|--------|---------------------|---------------------------|
| **Returns** | Universal AST (`ast.SchemaNode`) | Go types (via struct tags) |
| **Use for** | Queries, diffing, transformations | Simple data consumption |
| **Position info** | ✅ Preserved | ❌ Lost |
| **Format features** | ✅ Fully represented | ⚠️ May lose fidelity |
| **Example** | `node, _ := json.Parse(data)` | `json.Unmarshal(data, &struct{})` |

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

### 5. Universal AST (`pkg/ast/`)

**Node types for both data representation and validation schemas:**
- `LiteralNode` - Primitive values (string, number, bool, null) or exact value constraints
- `ObjectNode` - Structured data (maps/objects) or object validation with property schemas
- `ArrayNode` - Collections (arrays/sequences) or array validation with element schema
- `TypeNode` - Type constraints for validation (UUID, Email, etc.)
- `FunctionNode` - Parameterized validators (Integer(1, 100), etc.)
- And more...

**Remember:** The same AST represents both parsed data AND validation schemas!

---

## What You Build

When creating a parser project (e.g., `shape-json`), you build:

### 1. Format-Specific Tokenizer

Using Shape's tokenizer framework:

```go
package tokenizer

import "github.com/shapestone/shape-core/pkg/tokenizer"

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

Returning universal AST nodes:

```go
package parser

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-json/internal/tokenizer"
)

type Parser struct {
    tokenizer tokenizer.Tokenizer
    current   tokenizer.Token
}

// Parse returns universal AST (SchemaNode interface)
func (p *Parser) Parse() (ast.SchemaNode, error) {
    return p.parseValue()
}

// parseValue returns ast.SchemaNode (ObjectNode, LiteralNode, etc.)
func (p *Parser) parseValue() (ast.SchemaNode, error) {
    switch p.peek().Kind() {
    case TokenLBrace:
        return p.parseObject()  // Returns *ast.ObjectNode
    case TokenLBracket:
        return p.parseArray()   // Returns *ast.ObjectNode (numeric keys)
    case TokenString:
        return p.parseString()  // Returns *ast.LiteralNode
    case TokenNumber:
        return p.parseNumber()  // Returns *ast.LiteralNode
    case TokenTrue, TokenFalse:
        return p.parseBoolean() // Returns *ast.LiteralNode
    case TokenNull:
        return p.parseNull()    // Returns *ast.LiteralNode (value: nil)
    default:
        return nil, fmt.Errorf("unexpected token: %s", p.peek().Kind())
    }
}

// parseObject returns *ast.ObjectNode
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    startPos := p.peek().Position()
    properties := make(map[string]ast.SchemaNode)

    // ... parse properties into AST nodes...

    return ast.NewObjectNode(properties, startPos), nil
}

// parseArray returns *ast.ObjectNode with numeric keys
func (p *Parser) parseArray() (*ast.ObjectNode, error) {
    startPos := p.peek().Position()
    properties := make(map[string]ast.SchemaNode)

    // ... parse elements as numeric keys "0", "1", "2"...

    return ast.NewObjectNode(properties, startPos), nil
}

// parseString returns *ast.LiteralNode
func (p *Parser) parseString() (*ast.LiteralNode, error) {
    pos := p.peek().Position()
    value := p.unquoteString(p.peek().ValueString())
    return ast.NewLiteralNode(value, pos), nil
}
```

### 3. Public API

Dual API pattern - AST for advanced use, convenience functions for simple use:

```go
package json

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-json/internal/parser"
)

// Parse parses JSON input and returns universal AST
//
// Returns: ast.SchemaNode representing the parsed data
//   - *ast.ObjectNode for objects
//   - *ast.ObjectNode with numeric keys for arrays
//   - *ast.LiteralNode for primitives (string, number, bool, null)
//
// Example:
//   node, err := json.Parse(`{"name": "Alice", "age": 30}`)
//   if err != nil {
//       // handle error
//   }
//   // Use AST directly for queries, transformations, etc.
//   obj := node.(*ast.ObjectNode)
//   nameNode := obj.Properties()["name"]
func Parse(input string) (ast.SchemaNode, error) {
    p := parser.NewParser(input)
    return p.Parse()
}

// Unmarshal parses JSON and populates a Go struct (convenience API)
//
// This is a secondary API that converts AST -> Go types
//
// Example:
//   type User struct {
//       Name string `json:"name"`
//       Age  int    `json:"age"`
//   }
//   var user User
//   err := json.Unmarshal([]byte(`{"name": "Alice", "age": 30}`), &user)
func Unmarshal(data []byte, v interface{}) error {
    // 1. Parse to AST
    node, err := Parse(string(data))
    if err != nil {
        return err
    }
    // 2. Convert AST to Go types and populate struct
    return unmarshalNode(node, v)
}

// ToGoValue converts AST to Go types (helper for queries)
func ToGoValue(node ast.SchemaNode) interface{} {
    switch n := node.(type) {
    case *ast.LiteralNode:
        return n.Value()
    case *ast.ObjectNode:
        result := make(map[string]interface{})
        for key, val := range n.Properties() {
            result[key] = ToGoValue(val)
        }
        return result
    }
    return nil
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

## Format-Specific Feature Mapping

The universal AST is format-agnostic, but each format has unique features. Here's how to map them:

### JSON → AST

| JSON Feature | AST Node | Example |
|--------------|----------|---------|
| Object | `*ast.ObjectNode` | `{"name": "Alice"}` → `ObjectNode{properties: {"name": LiteralNode("Alice")}}` |
| Array | `*ast.ObjectNode` with numeric keys | `[1,2,3]` → `ObjectNode{properties: {"0": Literal(1), "1": Literal(2), "2": Literal(3)}}` |
| String | `*ast.LiteralNode` | `"hello"` → `LiteralNode{value: "hello"}` |
| Number | `*ast.LiteralNode` | `42` → `LiteralNode{value: int64(42)}` or `LiteralNode{value: float64(42.5)}` |
| Boolean | `*ast.LiteralNode` | `true` → `LiteralNode{value: true}` |
| Null | `*ast.LiteralNode` | `null` → `LiteralNode{value: nil}` |

**Example JSON → AST:**
```json
{
  "name": "Alice",
  "age": 30,
  "tags": ["admin", "user"]
}
```

Maps to:
```go
*ast.ObjectNode{
    properties: {
        "name": *ast.LiteralNode{value: "Alice"},
        "age":  *ast.LiteralNode{value: int64(30)},
        "tags": *ast.ObjectNode{
            properties: {
                "0": *ast.LiteralNode{value: "admin"},
                "1": *ast.LiteralNode{value: "user"},
            },
        },
    },
}
```

### XML → AST

| XML Feature | AST Representation | Convention |
|-------------|-------------------|------------|
| Element | `*ast.ObjectNode` | `<user>` → `ObjectNode` |
| Attribute | Property with `@` prefix | `<user id="123">` → `{"@id": LiteralNode("123")}` |
| Text content | Property with `#text` key | `<name>Alice</name>` → `{"#text": LiteralNode("Alice")}` |
| Mixed content | Combine text and element properties | `<p>Hello <b>world</b></p>` |
| Namespace | Include prefix in property name | `<ns:element>` → `{"ns:element": ...}` |
| Multiple children with same name | Use numeric suffixes | `<item>`, `<item>` → `{"item": ..., "item#1": ...}` or array representation |

**Example XML → AST:**
```xml
<user id="123" active="true">
    <name>Alice</name>
    <email>alice@example.com</email>
</user>
```

Maps to:
```go
*ast.ObjectNode{
    properties: {
        "@id":     *ast.LiteralNode{value: "123"},
        "@active": *ast.LiteralNode{value: "true"},
        "name":    *ast.LiteralNode{value: "Alice"},
        "email":   *ast.LiteralNode{value: "alice@example.com"},
    },
}
```

### YAML → AST

| YAML Feature | AST Representation | Strategy |
|--------------|-------------------|----------|
| Map | `*ast.ObjectNode` | Direct mapping - keys become properties |
| Sequence | `*ast.ObjectNode` with numeric keys | `[1, 2, 3]` → `{"0": 1, "1": 2, "2": 3}` |
| Scalar (string) | `*ast.LiteralNode` | Direct value mapping |
| Scalar (number) | `*ast.LiteralNode` | Parse to int64 or float64 |
| Scalar (boolean) | `*ast.LiteralNode` | Parse to bool |
| Scalar (null) | `*ast.LiteralNode` | value: nil |
| Anchor (`&anchor`) | Resolve during parsing - expand references |
| Alias (`*anchor`) | Resolve during parsing - copy referenced node |
| Multi-document | Return array of `ast.SchemaNode` (multiple roots) |

**Example YAML → AST:**
```yaml
name: Alice
age: 30
tags:
  - admin
  - user
```

Maps to:
```go
*ast.ObjectNode{
    properties: {
        "name": *ast.LiteralNode{value: "Alice"},
        "age":  *ast.LiteralNode{value: int64(30)},
        "tags": *ast.ObjectNode{
            properties: {
                "0": *ast.LiteralNode{value: "admin"},
                "1": *ast.LiteralNode{value: "user"},
            },
        },
    },
}
```

### CSV → AST

Represent as an object with numeric keys (rows), where each row is an object with column names:

**Example CSV:**
```csv
name,age,active
Alice,30,true
Bob,25,false
```

**Maps to AST:**
```go
*ast.ObjectNode{
    properties: {
        "0": *ast.ObjectNode{  // Row 1 (header row can be skipped)
            properties: {
                "name":   *ast.LiteralNode{value: "Alice"},
                "age":    *ast.LiteralNode{value: "30"},
                "active": *ast.LiteralNode{value: "true"},
            },
        },
        "1": *ast.ObjectNode{  // Row 2
            properties: {
                "name":   *ast.LiteralNode{value: "Bob"},
                "age":    *ast.LiteralNode{value: "25"},
                "active": *ast.LiteralNode{value: "false"},
            },
        },
    },
}
```

**Alternative:** If CSV has no headers, use numeric keys for columns too:
```go
*ast.ObjectNode{
    properties: {
        "0": *ast.ObjectNode{  // Row 0
            properties: {
                "0": *ast.LiteralNode{value: "Alice"},
                "1": *ast.LiteralNode{value: "30"},
                "2": *ast.LiteralNode{value: "true"},
            },
        },
    },
}
```

### Properties Files → AST

Simple key-value format maps directly to ObjectNode:

**Example `.properties`:**
```properties
app.name=MyApp
app.version=1.0.0
database.host=localhost
database.port=5432
```

**Maps to AST:**
```go
*ast.ObjectNode{
    properties: {
        "app.name":       *ast.LiteralNode{value: "MyApp"},
        "app.version":    *ast.LiteralNode{value: "1.0.0"},
        "database.host":  *ast.LiteralNode{value: "localhost"},
        "database.port":  *ast.LiteralNode{value: "5432"},
    },
}
```

**Alternative (hierarchical):** Split on `.` to create nested objects:
```go
*ast.ObjectNode{
    properties: {
        "app": *ast.ObjectNode{
            properties: {
                "name":    *ast.LiteralNode{value: "MyApp"},
                "version": *ast.LiteralNode{value: "1.0.0"},
            },
        },
        "database": *ast.ObjectNode{
            properties: {
                "host": *ast.LiteralNode{value: "localhost"},
                "port": *ast.LiteralNode{value: "5432"},
            },
        },
    },
}
```

---

## Building Data Programmatically

The universal AST provides a proper type system for constructing documents programmatically, which Go's `map[string]interface{}` cannot provide.

### Why Programmatic Construction Matters

Use cases:
- **Code generation:** Generate configuration files from templates
- **Data transformation:** Convert between formats (JSON ↔ XML ↔ YAML)
- **API responses:** Build structured responses dynamically
- **Testing:** Create test fixtures programmatically
- **Document assembly:** Combine multiple data sources into one document

### Building JSON Programmatically

```go
import "github.com/shapestone/shape-core/pkg/ast"

// Build {"name": "Alice", "age": 30, "active": true}
pos := ast.Position{Line: 1, Column: 1}  // Can be synthetic

user := ast.NewObjectNode(map[string]ast.SchemaNode{
    "name":   ast.NewLiteralNode("Alice", pos),
    "age":    ast.NewLiteralNode(int64(30), pos),
    "active": ast.NewLiteralNode(true, pos),
}, pos)

// Render to JSON string
output := json.Render(user)
// Output: {"name":"Alice","age":30,"active":true}
```

**Building nested structures:**
```go
// Build {"user": {"name": "Alice"}, "tags": ["admin", "user"]}
document := ast.NewObjectNode(map[string]ast.SchemaNode{
    "user": ast.NewObjectNode(map[string]ast.SchemaNode{
        "name": ast.NewLiteralNode("Alice", pos),
    }, pos),
    "tags": ast.NewObjectNode(map[string]ast.SchemaNode{
        "0": ast.NewLiteralNode("admin", pos),
        "1": ast.NewLiteralNode("user", pos),
    }, pos),
}, pos)

output := json.Render(document)
// Output: {"user":{"name":"Alice"},"tags":["admin","user"]}
```

### Building XML Programmatically

```go
// Build <user id="123" active="true"><name>Alice</name></user>
user := ast.NewObjectNode(map[string]ast.SchemaNode{
    "@id":     ast.NewLiteralNode("123", pos),      // Attribute
    "@active": ast.NewLiteralNode("true", pos),     // Attribute
    "name":    ast.NewLiteralNode("Alice", pos),    // Child element
}, pos)

output := xml.Render(user)
// Output: <user id="123" active="true"><name>Alice</name></user>
```

**Building XML with text content:**
```go
// Build <message priority="high">Hello World</message>
message := ast.NewObjectNode(map[string]ast.SchemaNode{
    "@priority": ast.NewLiteralNode("high", pos),
    "#text":     ast.NewLiteralNode("Hello World", pos),
}, pos)

output := xml.Render(message)
// Output: <message priority="high">Hello World</message>
```

### Building YAML Programmatically

```go
// Build YAML document
config := ast.NewObjectNode(map[string]ast.SchemaNode{
    "database": ast.NewObjectNode(map[string]ast.SchemaNode{
        "host": ast.NewLiteralNode("localhost", pos),
        "port": ast.NewLiteralNode(int64(5432), pos),
    }, pos),
    "features": ast.NewObjectNode(map[string]ast.SchemaNode{
        "0": ast.NewLiteralNode("caching", pos),
        "1": ast.NewLiteralNode("logging", pos),
    }, pos),
}, pos)

output := yaml.Render(config)
// Output:
// database:
//   host: localhost
//   port: 5432
// features:
//   - caching
//   - logging
```

### Format Conversion Using AST

The universal AST enables lossless conversion between formats:

```go
// JSON → XML conversion
jsonInput := `{"user": {"name": "Alice", "id": "123"}}`

// 1. Parse JSON to AST
node, _ := json.Parse(jsonInput)

// 2. Render AST as XML (with attribute mapping)
xmlOutput := xml.Render(node)
// Could produce: <user id="123"><name>Alice</name></user>
// (with custom rendering rules for attributes)
```

```go
// XML → JSON conversion
xmlInput := `<user id="123"><name>Alice</name></user>`

// 1. Parse XML to AST
node, _ := xml.Parse(xmlInput)

// 2. Render AST as JSON
jsonOutput := json.Render(node)
// Output: {"@id":"123","name":"Alice"}
```

### Why Go Types Are Insufficient for Programmatic Construction

**Problems with `map[string]interface{}`:**

1. **No position tracking** - Can't generate source maps or error locations
   ```go
   // ❌ Go map - no position info
   data := map[string]interface{}{"name": "Alice"}

   // ✅ AST - has position for every node
   node := ast.NewObjectNode(map[string]ast.SchemaNode{
       "name": ast.NewLiteralNode("Alice", pos),
   }, pos)
   ```

2. **Can't distinguish null vs missing vs empty**
   ```go
   // ❌ Go map - all these look the same
   map[string]interface{}{"value": nil}  // explicit null
   map[string]interface{}{}              // missing key

   // ✅ AST - can distinguish
   ast.NewObjectNode(map[string]ast.SchemaNode{
       "value": ast.NewLiteralNode(nil, pos),  // explicit null
   }, pos)
   ast.NewObjectNode(map[string]ast.SchemaNode{}, pos)  // no properties
   ```

3. **Can't represent format-specific features**
   ```go
   // ❌ Go map - can't distinguish XML attributes from elements
   map[string]interface{}{
       "id": "123",    // Is this <id>123</id> or id="123"?
       "name": "Alice",
   }

   // ✅ AST - explicit representation
   ast.NewObjectNode(map[string]ast.SchemaNode{
       "@id": ast.NewLiteralNode("123", pos),    // Attribute
       "name": ast.NewLiteralNode("Alice", pos), // Element
   }, pos)
   ```

4. **Type information is lost**
   ```go
   // ❌ Go map - was 123 an int or a string "123"?
   map[string]interface{}{"count": 123}

   // ✅ AST - preserves original type
   ast.NewLiteralNode(int64(123), pos)  // Explicit integer
   ast.NewLiteralNode("123", pos)       // Explicit string
   ```

### Summary: Universal AST Enables Rich Programmatic Construction

| Capability | Go Types (`map[string]interface{}`) | Universal AST |
|------------|-------------------------------------|---------------|
| **Build documents** | ✅ Basic | ✅ Full featured |
| **Position tracking** | ❌ Not possible | ✅ Every node has position |
| **Format features** | ❌ Generic only | ✅ XML attributes, etc. |
| **Type preservation** | ❌ Types lost | ✅ Types explicit |
| **Null vs missing** | ❌ Can't distinguish | ✅ Explicit representation |
| **Cross-format conversion** | ⚠️ Lossy | ✅ Lossless (with conventions) |
| **Error reporting** | ❌ No source info | ✅ Precise error locations |

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
│       ├── parser.go              # Public API (returns universal AST)
│       └── parser_test.go         # Public API tests
│
├── internal/
│   ├── tokenizer/
│   │   ├── tokenizer.go           # Format-specific tokenizer
│   │   ├── tokenizer_test.go
│   │   └── tokens.go              # Token type definitions
│   │
│   └── parser/
│       ├── parser.go              # Parser implementation (returns AST nodes)
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
go get github.com/shapestone/shape-core@latest

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
// - Parser returns universal AST nodes (*ast.ObjectNode, *ast.LiteralNode)
// - AST serves dual purposes: data representation AND validation schemas

// Top-level value (any JSON type)
// Parser function: parseValue() -> ast.SchemaNode
// Returns: *ast.ObjectNode, *ast.LiteralNode (depends on input type)
Value = Object | Array | String | Number | Boolean | Null ;

// Object with properties
// Parser function: parseObject() -> *ast.ObjectNode
// Example valid: { "id": "abc123", "name": "Alice" }
// Example valid: {} (empty object)
// Example invalid: { id: "value" } (missing quotes on key)
// Returns: *ast.ObjectNode with properties map[string]ast.SchemaNode
Object = "{" [ Property { "," Property } ] "}" ;

// Property key-value pair
// Parser function: parseProperty() -> (string, ast.SchemaNode)
// Returns: (key string, value ast.SchemaNode)
Property = String ":" Value ;

// Array of values
// Parser function: parseArray() -> *ast.ObjectNode
// Example valid: [1, 2, 3]
// Example valid: [] (empty array)
// Returns: *ast.ObjectNode with numeric keys "0", "1", "2"...
Array = "[" [ Value { "," Value } ] "]" ;

// String literal
// Parser function: parseString() -> *ast.LiteralNode
// Returns: *ast.LiteralNode with string value
String = '"' [^"]* '"' ;

// Number literal
// Parser function: parseNumber() -> *ast.LiteralNode
// Returns: *ast.LiteralNode with int64 or float64 value
Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;

// Boolean literal
// Parser function: parseBoolean() -> *ast.LiteralNode
// Returns: *ast.LiteralNode with bool value (true or false)
Boolean = "true" | "false" ;

// Null literal
// Parser function: parseNull() -> *ast.LiteralNode
// Returns: *ast.LiteralNode with nil value
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
    "github.com/shapestone/shape-core/pkg/tokenizer"
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

### Step 4: Implement Parser (Returns Universal AST!)

Create `internal/parser/parser.go`:

```go
package parser

import (
    "fmt"
    "strconv"

    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-json/internal/tokenizer"
    shapeTokenizer "github.com/shapestone/shape-core/pkg/tokenizer"
)

// Parser implements recursive descent parsing for JSON.
// Returns universal AST nodes (*ast.ObjectNode, *ast.LiteralNode).
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

// Parse parses the input and returns universal AST.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
//
// Returns: ast.SchemaNode (either *ast.ObjectNode or *ast.LiteralNode)
func (p *Parser) Parse() (ast.SchemaNode, error) {
    return p.parseValue()
}

// parseValue dispatches to specific parse functions.
//
// Grammar:
//   Value = Object | Array | String | Number | Boolean | Null ;
//
// Returns: ast.SchemaNode (actual type depends on input)
func (p *Parser) parseValue() (ast.SchemaNode, error) {
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

// parseObject parses an object and returns *ast.ObjectNode.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns: *ast.ObjectNode
// Example: {"id": "abc", "age": 30} -> ObjectNode{properties: {"id": LiteralNode("abc"), "age": LiteralNode(30)}}
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    startPos := p.currentPosition()

    // "{"
    if _, err := p.expect(tokenizer.TokenLBrace); err != nil {
        return nil, err
    }

    properties := make(map[string]ast.SchemaNode)

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

    return ast.NewObjectNode(properties, startPos), nil
}

// parseProperty parses a property key-value pair.
//
// Grammar:
//   Property = String ":" Value ;
//
// Returns: (key string, value ast.SchemaNode)
func (p *Parser) parseProperty() (string, ast.SchemaNode, error) {
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

// parseArray parses an array and returns *ast.ObjectNode with numeric keys.
//
// Grammar:
//   Array = "[" [ Value { "," Value } ] "]" ;
//
// Returns: *ast.ObjectNode with properties "0", "1", "2"...
// Example: [1, 2, 3] -> ObjectNode{properties: {"0": LiteralNode(1), "1": LiteralNode(2), "2": LiteralNode(3)}}
func (p *Parser) parseArray() (*ast.ObjectNode, error) {
    startPos := p.currentPosition()

    // "["
    if _, err := p.expect(tokenizer.TokenLBracket); err != nil {
        return nil, err
    }

    properties := make(map[string]ast.SchemaNode)
    index := 0

    // [ Value { "," Value } ] - Optional value list
    if p.peek().Kind() != tokenizer.TokenRBracket {
        // First element
        elem, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        properties[strconv.Itoa(index)] = elem
        index++

        // Additional elements
        for p.peek().Kind() == tokenizer.TokenComma {
            p.advance() // consume ","

            elem, err := p.parseValue()
            if err != nil {
                return nil, fmt.Errorf("in array element after comma: %w", err)
            }
            properties[strconv.Itoa(index)] = elem
            index++
        }
    }

    // "]"
    if _, err := p.expect(tokenizer.TokenRBracket); err != nil {
        return nil, err
    }

    return ast.NewObjectNode(properties, startPos), nil
}

// parseString parses a string literal and returns *ast.LiteralNode.
//
// Grammar:
//   String = '"' [^"]* '"' ;
//
// Returns: *ast.LiteralNode with string value
func (p *Parser) parseString() (*ast.LiteralNode, error) {
    pos := p.currentPosition()
    token, err := p.expect(tokenizer.TokenString)
    if err != nil {
        return nil, err
    }
    value := p.unquoteString(token.ValueString())
    return ast.NewLiteralNode(value, pos), nil
}

// parseNumber parses a number literal and returns *ast.LiteralNode.
//
// Grammar:
//   Number = "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE][+-]?[0-9]+)? ;
//
// Returns: *ast.LiteralNode with int64 or float64 value
func (p *Parser) parseNumber() (*ast.LiteralNode, error) {
    pos := p.currentPosition()
    token, err := p.expect(tokenizer.TokenNumber)
    if err != nil {
        return nil, err
    }

    numStr := token.ValueString()

    // Try parsing as integer first
    if !strings.Contains(numStr, ".") && !strings.Contains(numStr, "e") && !strings.Contains(numStr, "E") {
        num, err := strconv.ParseInt(numStr, 10, 64)
        if err == nil {
            return ast.NewLiteralNode(num, pos), nil
        }
    }

    // Parse as float
    num, err := strconv.ParseFloat(numStr, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid number: %w", err)
    }
    return ast.NewLiteralNode(num, pos), nil
}

// parseBoolean parses a boolean literal and returns *ast.LiteralNode.
//
// Grammar:
//   Boolean = "true" | "false" ;
//
// Returns: *ast.LiteralNode with bool value
func (p *Parser) parseBoolean() (*ast.LiteralNode, error) {
    pos := p.currentPosition()
    token := p.peek()
    var value bool
    if token.Kind() == tokenizer.TokenTrue {
        value = true
    } else if token.Kind() == tokenizer.TokenFalse {
        value = false
    } else {
        return nil, fmt.Errorf("expected boolean, got %s", token.Kind())
    }
    p.advance()
    return ast.NewLiteralNode(value, pos), nil
}

// parseNull parses a null literal and returns *ast.LiteralNode.
//
// Grammar:
//   Null = "null" ;
//
// Returns: *ast.LiteralNode with nil value
func (p *Parser) parseNull() (*ast.LiteralNode, error) {
    pos := p.currentPosition()
    if _, err := p.expect(tokenizer.TokenNull); err != nil {
        return nil, err
    }
    return ast.NewLiteralNode(nil, pos), nil
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

// currentPosition returns position from current token.
func (p *Parser) currentPosition() ast.Position {
    if p.hasToken {
        return ast.Position{
            Line:   p.current.Row(),
            Column: p.current.Column(),
            Offset: p.current.Offset(),
        }
    }
    return ast.Position{Line: 1, Column: 1, Offset: 0}
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
- Each function returns universal AST nodes
- `parseObject()` returns `*ast.ObjectNode`
- `parseArray()` returns `*ast.ObjectNode` (with numeric keys)
- `parseString()` returns `*ast.LiteralNode`
- `parseNumber()` returns `*ast.LiteralNode`
- `parseBoolean()` returns `*ast.LiteralNode`
- `parseNull()` returns `*ast.LiteralNode`

### Step 5: Implement Public API

Create `pkg/json/parser.go`:

```go
package json

import (
    "github.com/shapestone/shape-core/pkg/ast"
    "github.com/shapestone/shape-json/internal/parser"
)

// Parse parses JSON input and returns universal AST.
//
// Returns: ast.SchemaNode representing the parsed data
//   - *ast.ObjectNode for JSON objects
//   - *ast.ObjectNode with numeric keys for JSON arrays
//   - *ast.LiteralNode for primitives (string, number, bool, null)
//
// Example:
//   node, err := json.Parse(`{"id": "abc123", "age": 30}`)
//   if err != nil {
//       // handle error
//   }
//   obj := node.(*ast.ObjectNode)
//   idNode := obj.Properties()["id"].(*ast.LiteralNode)
//   id := idNode.Value().(string)  // "abc123"
func Parse(input string) (ast.SchemaNode, error) {
    p := parser.NewParser(input)
    return p.Parse()
}

// Unmarshal parses JSON and populates a Go struct (convenience API).
//
// This is a secondary API that converts AST -> Go types.
//
// Example:
//   type User struct {
//       ID   string `json:"id"`
//       Age  int    `json:"age"`
//   }
//   var user User
//   err := json.Unmarshal([]byte(`{"id": "abc123", "age": 30}`), &user)
func Unmarshal(data []byte, v interface{}) error {
    // 1. Parse to AST
    node, err := Parse(string(data))
    if err != nil {
        return err
    }
    // 2. Convert AST to Go types and populate struct
    return unmarshalNode(node, v)
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

    "github.com/shapestone/shape-core/pkg/ast"
)

func TestParseObject_Empty(t *testing.T) {
    input := `{}`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode, got %T", result)
    }

    if len(obj.Properties()) != 0 {
        t.Errorf("expected empty object, got %d properties", len(obj.Properties()))
    }
}

func TestParseObject_SingleProperty(t *testing.T) {
    input := `{ "id": "abc123" }`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    obj, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode, got %T", result)
    }

    if len(obj.Properties()) != 1 {
        t.Fatalf("expected 1 property, got %d", len(obj.Properties()))
    }

    idNode, ok := obj.Properties()["id"].(*ast.LiteralNode)
    if !ok {
        t.Errorf("expected *ast.LiteralNode for 'id', got %T", obj.Properties()["id"])
    }
    if idNode.Value().(string) != "abc123" {
        t.Errorf("expected 'abc123', got %v", idNode.Value())
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

    obj, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode, got %T", result)
    }

    if len(obj.Properties()) != 4 {
        t.Errorf("expected 4 properties, got %d", len(obj.Properties()))
    }

    idNode := obj.Properties()["id"].(*ast.LiteralNode)
    if idNode.Value().(string) != "abc123" {
        t.Errorf("expected id='abc123', got %v", idNode.Value())
    }

    nameNode := obj.Properties()["name"].(*ast.LiteralNode)
    if nameNode.Value().(string) != "Alice" {
        t.Errorf("expected name='Alice', got %v", nameNode.Value())
    }

    ageNode := obj.Properties()["age"].(*ast.LiteralNode)
    if ageNode.Value().(int64) != 30 {
        t.Errorf("expected age=30, got %v", ageNode.Value())
    }

    activeNode := obj.Properties()["active"].(*ast.LiteralNode)
    if activeNode.Value().(bool) != true {
        t.Errorf("expected active=true, got %v", activeNode.Value())
    }
}

func TestParseArray_Empty(t *testing.T) {
    input := `[]`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode (array), got %T", result)
    }

    if len(arr.Properties()) != 0 {
        t.Errorf("expected empty array, got %d elements", len(arr.Properties()))
    }
}

func TestParseArray_Numbers(t *testing.T) {
    input := `[1, 2, 3]`
    p := NewParser(input)
    result, err := p.Parse()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    arr, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode (array), got %T", result)
    }

    if len(arr.Properties()) != 3 {
        t.Errorf("expected 3 elements, got %d", len(arr.Properties()))
    }

    expected := []int64{1, 2, 3}
    for i, exp := range expected {
        key := strconv.Itoa(i)
        node := arr.Properties()[key].(*ast.LiteralNode)
        if node.Value().(int64) != exp {
            t.Errorf("element %d: expected %v, got %v", i, exp, node.Value())
        }
    }
}

func TestParsePrimitives(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
    }{
        {"string", `"hello"`, "hello"},
        {"number_int", `42`, int64(42)},
        {"number_float", `3.14`, 3.14},
        {"number_negative", `-17`, int64(-17)},
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

            literal, ok := result.(*ast.LiteralNode)
            if !ok {
                t.Fatalf("expected *ast.LiteralNode, got %T", result)
            }

            if literal.Value() != tt.expected {
                t.Errorf("expected %v (%T), got %v (%T)",
                    tt.expected, tt.expected, literal.Value(), literal.Value())
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

    obj, ok := result.(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected *ast.ObjectNode, got %T", result)
    }

    // Check nested object
    user, ok := obj.Properties()["user"].(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected user to be *ast.ObjectNode, got %T", obj.Properties()["user"])
    }
    idNode := user.Properties()["id"].(*ast.LiteralNode)
    if idNode.Value().(string) != "abc123" {
        t.Errorf("expected user.id='abc123', got %v", idNode.Value())
    }

    // Check array
    tags, ok := obj.Properties()["tags"].(*ast.ObjectNode)
    if !ok {
        t.Fatalf("expected tags to be *ast.ObjectNode (array), got %T", obj.Properties()["tags"])
    }
    if len(tags.Properties()) != 2 {
        t.Errorf("expected 2 tags, got %d", len(tags.Properties()))
    }

    // Check number
    countNode := obj.Properties()["count"].(*ast.LiteralNode)
    if countNode.Value().(int64) != 42 {
        t.Errorf("expected count=42, got %v", countNode.Value())
    }
}
```

Create `internal/parser/grammar_test.go`:

```go
package parser

import (
    "testing"

    "github.com/shapestone/shape-core/pkg/grammar"
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

Parse() returns universal AST based on JSON structure:

- Objects → `*ast.ObjectNode`
- Arrays → `*ast.ObjectNode` (with numeric keys)
- Strings → `*ast.LiteralNode` (string value)
- Numbers → `*ast.LiteralNode` (int64 or float64 value)
- Booleans → `*ast.LiteralNode` (bool value)
- Null → `*ast.LiteralNode` (nil value)

## Grammar

See [docs/grammar/json.ebnf](docs/grammar/json.ebnf) for the complete EBNF specification.

## Documentation

- [Grammar Specification](docs/grammar/json.ebnf)
- [Examples](docs/examples/)
- [Shape Infrastructure](https://github.com/shapestone/shape-core)

## License

Apache 2.0
```

### 2. Package Documentation

```go
// Package json provides parsing for JSON data format.
//
// JSON is a lightweight data-interchange format (RFC 8259).
//
// This parser returns universal AST nodes (*ast.ObjectNode, *ast.LiteralNode).
// The AST serves dual purposes: data representation AND validation schemas.
//
// Grammar: See docs/grammar/json.ebnf for complete specification.
//
// This parser uses LL(1) recursive descent parsing (see Shape ADR 0004).
// Grammar-based tests verify parser correctness (see Shape ADR 0005).
//
// Example:
//   data := `{ "id": "abc123", "age": 30 }`
//   node, err := json.Parse(data)
//   obj := node.(*ast.ObjectNode)
//   idNode := obj.Properties()["id"].(*ast.LiteralNode)
//   id := idNode.Value().(string)  // "abc123"
package json
```

### 3. Function Documentation with Grammar

```go
// parseObject parses an object and returns *ast.ObjectNode.
//
// Grammar:
//   Object = "{" [ Property { "," Property } ] "}" ;
//
// Returns: *ast.ObjectNode with properties map[string]ast.SchemaNode
// Accepts empty objects: {}
// Requires quoted property keys: { "id": "value" }
// Rejects unquoted keys: { id: "value" }
func (p *Parser) parseObject() (*ast.ObjectNode, error)
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

    "github.com/shapestone/shape-core/pkg/tokenizer"
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
// Returns: *ast.ObjectNode with single property
//
// Example: name = "Alice" → ObjectNode{properties: {"name": LiteralNode("Alice")}}
// Example: age = 30 → ObjectNode{properties: {"age": LiteralNode(30)}}
func Parse(input string) (*ast.ObjectNode, error) {
    // Create tokenizer
    tok := tokenizer.NewTokenizer(
        tokenizer.WhitespaceMatcherFunc(),
        tokenizer.StringMatcherFunc(TokenEquals, "="),
        tokenizer.RegexMatcherFunc(TokenString, `"[^"]*"`),
        tokenizer.RegexMatcherFunc(TokenNumber, `[0-9]+`),
        tokenizer.RegexMatcherFunc(TokenKey, `[a-zA-Z_]+`),
    )
    tok.Initialize(input)

    pos := ast.Position{Line: 1, Column: 1}

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

    var valueNode *ast.LiteralNode
    switch valueToken.Kind() {
    case TokenString:
        // Remove quotes
        s := valueToken.ValueString()
        valueNode = ast.NewLiteralNode(s[1:len(s)-1], pos)
    case TokenNumber:
        num, err := strconv.ParseInt(valueToken.ValueString(), 10, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number: %w", err)
        }
        valueNode = ast.NewLiteralNode(num, pos)
    default:
        return nil, fmt.Errorf("expected string or number, got %s", valueToken.Kind())
    }

    properties := map[string]ast.SchemaNode{
        key: valueNode,
    }
    return ast.NewObjectNode(properties, pos), nil
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
- [ ] Add implementation hints specifying AST return types
- [ ] Include examples of valid/invalid syntax
- [ ] Specify that parser returns universal AST nodes

### Implementation
- [ ] Implement tokenizer with token definitions
- [ ] Implement parser returning universal AST nodes (*ast.ObjectNode, *ast.LiteralNode)
- [ ] Add grammar fragments to function comments
- [ ] Implement public API in `pkg/{format}/`
- [ ] Add position tracking for error messages
- [ ] Implement context-aware error messages
- [ ] VERIFY: Parser returns `ast.SchemaNode`, `*ast.ObjectNode`, `*ast.LiteralNode`
- [ ] VERIFY: Parser does NOT return Go types (`map[string]interface{}`, `[]interface{}`)

### Testing
- [ ] Write unit tests for tokenizer
- [ ] Write manual parser tests for specific scenarios
- [ ] Write error handling tests
- [ ] Implement grammar-based verification tests
- [ ] Verify 95%+ test coverage
- [ ] Verify 100% grammar coverage
- [ ] Test that returned types are AST nodes (*ast.ObjectNode, *ast.LiteralNode)

### Documentation
- [ ] Complete README.md with usage examples
- [ ] Document grammar specification
- [ ] Add godoc comments to all public APIs
- [ ] Add runnable examples in `examples/`
- [ ] Document return types (universal AST nodes)
- [ ] Document dual API pattern (Parse → AST, Unmarshal → Go types)
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

1. **Return Go types from parser**
   ```go
   // WRONG!
   return map[string]interface{}{"id": "abc"}, nil  // NO!
   return []interface{}{1, 2, 3}, nil               // NO!
   return "hello", nil                              // NO!
   ```

2. **Think AST is only for validation schemas**
   - The universal AST serves DUAL purposes:
     - Data representation (what parsers return)
     - Validation schemas (type constraints)
   - Parsers MUST return AST nodes

3. **Mix parsing concerns**
   - Parsing: format → AST nodes
   - Querying: AST → Go types (for query execution)
   - Unmarshaling: AST → Go structs (convenience API)
   - Keep these operations separate

### DO:

1. **Return AST nodes from parser**
   ```go
   // CORRECT!
   return ast.NewObjectNode(properties, pos), nil
   return ast.NewLiteralNode("hello", pos), nil
   return ast.NewLiteralNode(int64(42), pos), nil
   ```

2. **Use the dual API pattern**
   ```go
   // CORRECT! Primary API returns AST
   func Parse(input string) (ast.SchemaNode, error) {
       p := parser.NewParser(input)
       return p.Parse()  // Returns AST
   }

   // CORRECT! Secondary API for convenience
   func Unmarshal(data []byte, v interface{}) error {
       node, err := Parse(string(data))
       if err != nil {
           return err
       }
       return unmarshalNode(node, v)  // AST → Go types
   }
   ```

3. **Keep parser projects focused**
   - One format per project
   - Returns universal AST nodes
   - Provides both Parse (→ AST) and Unmarshal (→ structs) APIs
   - Independent versioning

---

## Support

For questions or issues:
- **Shape Issues:** https://github.com/shapestone/shape-core/issues
- **Format-specific Issues:** File issues in your parser repository (e.g., https://github.com/shapestone/shape-json/issues)

---

**This guide ensures all Shape parser projects follow correct architecture patterns using the universal AST for both data representation and validation schemas.**
