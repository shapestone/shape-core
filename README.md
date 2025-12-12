# Shape - Parser Infrastructure for Structured Data

![Build Status](https://github.com/shapestone/shape/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/shapestone/shape)](https://goreportcard.com/report/github.com/shapestone/shape)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codecov](https://codecov.io/gh/shapestone/shape/branch/main/graph/badge.svg)](https://codecov.io/gh/shapestone/shape)
![Go Version](https://img.shields.io/github/go-mod/go-version/shapestone/shape)
![Latest Release](https://img.shields.io/github/v/release/shapestone/shape)
[![GoDoc](https://pkg.go.dev/badge/github.com/shapestone/shape.svg)](https://pkg.go.dev/github.com/shapestone/shape)

[![CodeQL](https://github.com/shapestone/shape/actions/workflows/codeql.yml/badge.svg)](https://github.com/shapestone/shape/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/shapestone/shape/badge)](https://securityscorecards.dev/viewer/?uri=github.com/shapestone/shape)
[![Security Policy](https://img.shields.io/badge/Security-Policy-brightgreen)](SECURITY.md)

**Repository:** github.com/shapestone/shape

Shape is a reusable parser infrastructure library that provides:
- **AST framework** for representing validation schemas
- **Tokenizer API** for building custom parsers
- **Parser interface** for format implementations
- **Schema validator framework** for semantic validation
- **Grammar verification tools** for enforced documentation and parser correctness

## What is Shape?

Shape provides the foundational components for parsing structured data and building custom domain-specific languages (DSLs). It's designed to be a generic, reusable infrastructure that other projects build upon.

**Shape is infrastructure only** - actual parser implementations are in separate projects (see [Parser Projects](#parser-projects) below).

## Features

### AST Framework
- **Unified AST:** Type-safe node definitions for schemas
- **Visitor Pattern:** Traverse and manipulate ASTs
- **Position Tracking:** Line/column numbers for all nodes
- **Serialization:** JSON import/export for ASTs

### Tokenizer API
- **Reusable Tokenization:** Build custom parsers with matchers
- **Stream Processing:** Efficient character stream handling (in-memory and buffered streaming)
- **Streaming Support:** Parse large files with constant memory using `NewStreamFromReader(io.Reader)`
- **Position Tracking:** Detailed error locations
- **Extensible:** Create custom matchers for your DSL

### Parser Interface
- **API patterns:** Recommended patterns for parser implementations (Parse, ParseReader, Format detection, Rendering)
- **Error types:** Standardized error handling for parsers
- **Flexibility:** Parser projects choose their parsing technique (LL(1), Pratt, PEG, etc.)

### Schema Validator Framework
- **Type Registry:** Register custom types (UUID, Email, etc.)
- **Function Registry:** Register validation functions
- **Multi-Error Collection:** Collect all validation errors
- **Rich Error Formatting:** Colored terminal, plain text, JSON output
- **Smart Hints:** "Did you mean" suggestions

### Grammar Verification Framework
- **EBNF Parser:** Parse custom EBNF variant grammars
- **Test Generation:** Auto-generate verification tests from grammars
- **Coverage Tracking:** Track grammar rule coverage in tests
- **Enforced Documentation:** Grammars verify parser correctness in CI
- **LLM-Friendly:** Grammars guide LLM-assisted parser development

## Installation

```bash
go get github.com/shapestone/shape
```

## Quick Start

### Using the Tokenizer to Build a Custom Parser

```go
package main

import (
    "fmt"
    "github.com/shapestone/shape/pkg/tokenizer"
)

func main() {
    // Define custom matchers for your DSL
    tok := tokenizer.NewTokenizer(
        tokenizer.StringMatcherFunc("LBrace", "{"),
        tokenizer.StringMatcherFunc("RBrace", "}"),
        tokenizer.RegexMatcherFunc("Identifier", `[a-zA-Z_][a-zA-Z0-9_]*`),
    )

    // Tokenize input
    tok.Initialize("{ myIdentifier }")
    tokens, err := tok.Tokenize()
    if err != nil {
        panic(err)
    }

    // Process tokens
    for _, token := range tokens {
        fmt.Printf("%s: %s\n", token.Type(), token.Value())
    }
}
```

### Using Parser Projects

Shape provides the infrastructure; actual parsers are in separate projects:

**Data Format Parsers**:
- [shape-json](https://github.com/shapestone/shape-json) - JSON validation
- [shape-yaml](https://github.com/shapestone/shape-yaml) - YAML validation
- [shape-csv](https://github.com/shapestone/shape-csv) - CSV parsing
- [shape-xml](https://github.com/shapestone/shape-xml) - XML validation
- [shape-props](https://github.com/shapestone/shape-props) - Properties file parsing

### Using the AST Framework

```go
package main

import (
    "fmt"
    "github.com/shapestone/shape/pkg/ast"
)

func main() {
    // Create AST nodes
    obj := ast.NewObjectNode(map[string]ast.SchemaNode{
        "id": ast.NewTypeNode("UUID", ast.Position{}),
        "name": ast.NewTypeNode("String", ast.Position{}),
    }, ast.Position{})

    // Traverse with visitor pattern
    visitor := &MyVisitor{}
    obj.Accept(visitor)
}

type MyVisitor struct{}

func (v *MyVisitor) VisitObject(n *ast.ObjectNode) error {
    fmt.Printf("Found object with %d properties\n", len(n.Properties))
    return nil
}

func (v *MyVisitor) VisitType(n *ast.TypeNode) error {
    fmt.Printf("Found type: %s\n", n.TypeName)
    return nil
}

// Implement other visitor methods...
```

### Using the Schema Validator Framework

```go
package main

import (
    "fmt"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/pkg/validator"
)

func main() {
    // Create validator with custom types
    v := validator.NewSchemaValidator()
    v.RegisterType("SSN", validator.TypeDescriptor{
        Name:        "SSN",
        Description: "Social Security Number",
    })

    // Create an AST (normally from parsing)
    schema := ast.NewObjectNode(map[string]ast.SchemaNode{
        "ssn": ast.NewTypeNode("SSN", ast.Position{}),
        "age": ast.NewFunctionNode("Integer", []ast.SchemaNode{
            ast.NewLiteralNode(18, ast.Position{}),
            ast.NewLiteralNode(120, ast.Position{}),
        }, ast.Position{}),
    }, ast.Position{})

    // Validate schema
    result := v.ValidateAll(schema, "")
    if !result.Valid {
        fmt.Println(result.FormatColored())
    }
}
```

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│         Parser Projects (json, yaml, csv, xml, etc.)     │
└──────────────────────┬───────────────────────────────────┘
                       │ depends on
                       ▼
       ┌───────────────────────────────────────┐
       │       Shape Infrastructure            │
       ├───────────────────────────────────────┤
       │  ┌───────┐  ┌──────────┐  ┌────────┐  │
       │  │  AST  │  │Tokenizer │  │ Parser │  │
       │  │  API  │  │   API    │  │  API   │  │
       │  └───────┘  └──────────┘  └────────┘  │
       │  ┌────────────────────────────────┐   │
       │  │    Validator Framework         │   │
       │  └────────────────────────────────┘   │
       │  ┌────────────────────────────────┐   │
       │  │  Grammar Verification Tools    │   │
       │  └────────────────────────────────┘   │
       └───────────────────────────────────────┘
```

## Components

### pkg/ast
Abstract Syntax Tree node definitions:
- `LiteralNode` - Literal values (string, number, boolean, null)
- `TypeNode` - Type identifiers (UUID, Email, etc.)
- `FunctionNode` - Function calls with arguments
- `ObjectNode` - Objects with properties
- `ArrayNode` - Arrays with element schemas
- Visitor pattern for AST traversal

### pkg/tokenizer
Reusable tokenization framework:
- `Tokenizer` - Main tokenizer with matcher support
- `Stream` - Character stream with position tracking (in-memory and buffered)
  - `NewStream(string)` - In-memory stream for small to medium data
  - `NewStreamFromReader(io.Reader)` - Buffered stream for large files and streaming data
- `Token` - Token representation with type and value
- Built-in matchers: String, Regex, Whitespace, Number
- Custom matcher creation

**Streaming capabilities:**
- Parse files of any size with constant memory (64KB buffer)
- Works with any `io.Reader` (files, network, pipes)
- Supports backtracking within buffer window
- Note: Current implementation has known limitations with buffer boundary crossing (see [docs/architecture/buffered-stream-implementation.md](docs/architecture/buffered-stream-implementation.md))

### pkg/validator
Schema validation framework:
- `SchemaValidator` - Main validator with registries
- `TypeRegistry` - Register and validate types
- `FunctionRegistry` - Register and validate functions
- `ValidationResult` - Multi-error collection
- `ValidationError` - Rich error formatting

### pkg/parser
Parser error types and API patterns:
- `ParseError` - Standardized error type with position information
- `NewSyntaxError`, `NewUnexpectedTokenError`, `NewUnexpectedEOFError` - Error constructors
- Recommended API patterns documented in package docs

### pkg/grammar
Grammar verification infrastructure:
- `EBNFParser` - Parse custom EBNF variant grammars
- `TestGenerator` - Generate verification tests from grammars
- `CoverageTracker` - Track grammar rule coverage
- `ASTComparator` - Compare ASTs for equivalence
- Used by parser projects to verify correctness

## Examples

See the [examples/](examples/) directory:
- `examples/custom-dsl/` - Building a custom DSL parser
- `examples/tokenizer-usage/` - Using the tokenizer API
- `examples/ast-manipulation/` - Working with AST nodes

## Use Cases

### Build Custom Parsers
Use Shape's tokenizer to create domain-specific language parsers:
- Configuration languages
- Query languages
- Diagram definitions
- Schema formats

**Example:** [Inkling](https://github.com/shapestone/inkling) uses Shape's tokenizer to parse diagram definitions.

### Build Parser Projects
Use Shape's infrastructure to create format parsers:
- Data format parsers (JSON, YAML, CSV, XML, Properties)
- Custom format parsers for domain-specific languages

**To implement a new parser:** See [Parser Implementation Guide](docs/PARSER_IMPLEMENTATION_GUIDE.md) for complete step-by-step instructions. This guide is optimized for both human developers and LLM-assisted development.

### Build Validation Tools
Use Shape's validator framework and AST:
- Custom schema validators
- Schema transformation tools
- Schema analysis tools

## Parser Projects

Shape provides the infrastructure for these parser projects:

**Data Format Parsers:**
- [shape-json](https://github.com/shapestone/shape-json) - JSON validation
- [shape-yaml](https://github.com/shapestone/shape-yaml) - YAML validation
- [shape-csv](https://github.com/shapestone/shape-csv) - CSV parsing
- [shape-xml](https://github.com/shapestone/shape-xml) - XML validation
- [shape-props](https://github.com/shapestone/shape-props) - Properties file parsing

## Related Projects

Shape is a standalone parser infrastructure that can be used independently or as part of the Shapestone ecosystem. See [ECOSYSTEM.md](ECOSYSTEM.md) for details on related projects.

## Performance

Shape infrastructure is designed for high performance:
- AST operations: Sub-microsecond node creation
- Tokenization: Sub-microsecond for simple inputs
- Validation framework: Minimal overhead

For parser-specific performance metrics, see individual parser project documentation.

## Documentation

- [Parser Implementation Guide](docs/PARSER_IMPLEMENTATION_GUIDE.md) - **Complete guide for implementing new parsers**
- [Architecture](docs/architecture/ARCHITECTURE.md) - System design and components
- [ADR 0004: Parser Strategy](docs/architecture/decisions/0004-parser-strategy.md) - LL(1) recursive descent parsing
- [ADR 0005: Grammar-as-Verification](docs/architecture/decisions/0005-grammar-as-verification.md) - Enforced documentation and parser correctness
- [Contributing](CONTRIBUTING.md) - Contribution guidelines
- [Security](SECURITY.md) - Security policy
- [Ecosystem](ECOSYSTEM.md) - Related projects

## Development

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint

# Run benchmarks
make bench
```

## Requirements

- Go 1.25 or later
- Minimal external dependencies (only `github.com/google/uuid` for validation framework)

## License

Shape is licensed under the [Apache License 2.0](LICENSE).

Copyright © 2020-2025 Shapestone

## Status

Shape is production-ready infrastructure:
- ✅ AST framework with comprehensive node types
- ✅ Public tokenizer API for custom parsers
- ✅ Parser error types and API patterns
- ✅ Schema validator framework
- ✅ Comprehensive test coverage
- ✅ Comprehensive documentation
- ✅ Production-tested in multiple projects
- ✅ 10 parser projects built on Shape infrastructure

## Support

- **Issues:** https://github.com/shapestone/shape/issues
- **Discussions:** https://github.com/shapestone/shape/discussions
- **Security:** See [SECURITY.md](SECURITY.md)
