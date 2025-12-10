# pkg/grammar

Grammar-based verification tools for Shape parsers. Provides EBNF parsing, test generation, coverage tracking, and AST comparison for verifying parser correctness.

## Overview

The `pkg/grammar` package implements the grammar-as-verification infrastructure described in [ADR 0005](../../docs/architecture/decisions/0005-grammar-as-verification.md). It allows you to:

1. **Parse EBNF grammars** into structured representations
2. **Generate test cases** automatically from grammar rules
3. **Track coverage** of grammar rules during testing
4. **Compare ASTs** for dual parser verification

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/shapestone/shape/pkg/grammar"
)

func main() {
    // Parse an EBNF grammar
    g, _ := grammar.ParseEBNF(`
        Value = "true" | "false" | "null" ;
    `)

    // Generate test cases
    tests := g.GenerateTests(grammar.DefaultOptions())

    // Run tests against your parser
    for _, test := range tests {
        result, err := YourParser(test.Input)

        if test.ShouldSucceed && err != nil {
            fmt.Printf("FAIL: %s\n", test.Name)
        }
    }
}
```

## Components

### EBNF Parser

Parse EBNF grammar specifications into structured Grammar objects.

```go
grammar, err := grammar.ParseEBNF(`
    // Comments are supported
    Expression = Term { ("+" | "-") Term } ;
    Term = Factor { ("*" | "/") Factor } ;
    Factor = Number | "(" Expression ")" ;
    Number = [0-9]+ ;
`)
```

**Supported EBNF Syntax:**
- `"literal"` - Terminal string (exact match)
- `[a-z]` - Character class (regex pattern)
- `Identifier` - Non-terminal (rule reference)
- `a | b` - Alternation (choose one)
- `a b` - Sequence (one after another)
- `[ a ]` - Optional (zero or one)
- `{ a }` - Repetition (zero or more)
- `a+` - One or more
- `a*` - Zero or more (equivalent to `{ a }`)
- `( a )` - Grouping
- `// comment` - Single-line comments

**Note:** Grammar fragments with undefined rule references are allowed (useful for documentation examples). Call `grammar.Validate()` explicitly if you need validation.

### Test Generator

Automatically generate test cases from grammar rules.

```go
tests := grammar.GenerateTests(grammar.TestOptions{
    MaxDepth:        5,      // Limit recursion depth
    CoverAllRules:   true,   // Ensure all rules are tested
    EdgeCases:       true,   // Generate empty/boundary cases
    InvalidCases:    true,   // Generate invalid inputs
    MaxAlternatives: 3,      // Limit alternation choices
})

for _, test := range tests {
    fmt.Printf("%s: %s (expect %v)\n",
        test.Name, test.Input, test.ShouldSucceed)
}
```

**Generated Test Types:**
- **Valid cases**: Cover all grammar paths
- **Invalid cases**: Violate grammar rules (if `InvalidCases: true`)
- **Edge cases**: Empty inputs, single elements, boundaries (if `EdgeCases: true`)

**Default Options:**
```go
func DefaultOptions() TestOptions {
    return TestOptions{
        MaxDepth:        3,
        CoverAllRules:   true,
        EdgeCases:       false,
        InvalidCases:    false,
        MaxAlternatives: 2,
    }
}
```

### Coverage Tracker

Track which grammar rules are exercised during parsing/testing.

```go
// Create tracker
tracker := grammar.NewCoverageTracker(g)

// Instrument your parser to record rule invocations
func (p *Parser) parseValue() {
    if p.tracker != nil {
        p.tracker.RecordRule("Value")
    }
    // ... parsing logic
}

// Generate report
report := tracker.Report()

fmt.Printf("Coverage: %.1f%%\n", report.Percentage)
fmt.Printf("Total Rules: %d\n", report.TotalRules)
fmt.Printf("Covered Rules: %d\n", report.CoveredRules)

if len(report.UncoveredRules) > 0 {
    fmt.Println("Uncovered rules:")
    for _, rule := range report.UncoveredRules {
        fmt.Printf("  - %s\n", rule)
    }
}

// Or use formatted output
fmt.Println(report.FormatReport())
```

**CoverageReport Structure:**
```go
type CoverageReport struct {
    TotalRules      int            // Total rules in grammar
    CoveredRules    int            // Rules that were invoked
    Percentage      float64        // Coverage percentage
    UncoveredRules  []string       // Names of uncovered rules
    RuleInvocations map[string]int // Invocation count per rule
}
```

### AST Comparator

Deep structural comparison of AST nodes (for dual parser verification).

```go
// Compare two ASTs for equality (ignores positions)
if !grammar.ASTEqual(refAST, prodAST) {
    // Get human-readable diff
    diff := grammar.ASTDiff(refAST, prodAST)
    t.Errorf("AST mismatch: %s", diff)
}
```

**Use Cases:**
- **Dual parser verification**: Compare reference parser vs production parser
- **Refactoring verification**: Ensure parser changes don't alter behavior
- **Cross-format consistency**: Verify JSONV and YAMLV produce same AST

**Comparison Behavior:**
- Position information is ignored
- Only structure and values are compared
- Recursive comparison for nested nodes
- Works with all AST node types: Literal, Type, Function, Object, Array

## Examples

See the [`examples/`](./examples/) directory for complete EBNF grammar examples:

- [`boolean.ebnf`](./examples/boolean.ebnf) - Simple boolean expressions
- [`arithmetic.ebnf`](./examples/arithmetic.ebnf) - Arithmetic with operator precedence
- [`json-schema.ebnf`](./examples/json-schema.ebnf) - JSON-like schema grammar
- [`shape-validation.ebnf`](./examples/shape-validation.ebnf) - Shape validation DSL

## Testing Patterns

### Pattern 1: Grammar-Generated Tests

```go
func TestParserAgainstGrammar(t *testing.T) {
    spec, _ := grammar.ParseEBNF(grammarSource)

    tests := spec.GenerateTests(grammar.TestOptions{
        MaxDepth:      5,
        CoverAllRules: true,
        InvalidCases:  true,
    })

    for _, test := range tests {
        t.Run(test.Name, func(t *testing.T) {
            result, err := YourParser(test.Input)

            if test.ShouldSucceed {
                if err != nil {
                    t.Errorf("expected valid, got error: %v", err)
                }
            } else {
                if err == nil {
                    t.Errorf("expected invalid, got success")
                }
            }
        })
    }
}
```

### Pattern 2: Coverage Tracking

```go
func TestGrammarCoverage(t *testing.T) {
    spec, _ := grammar.ParseEBNF(grammarSource)
    tracker := grammar.NewCoverageTracker(spec)

    // Attach tracker to parser
    parser := NewParser()
    parser.SetTracker(tracker)

    // Run existing tests
    RunAllTests(parser)

    // Check coverage
    report := tracker.Report()
    if report.Percentage < 90.0 {
        t.Errorf("insufficient grammar coverage: %.1f%%", report.Percentage)
        t.Logf("Uncovered rules: %v", report.UncoveredRules)
    }
}
```

### Pattern 3: Dual Parser Verification

```go
func TestDualParserEquivalence(t *testing.T) {
    inputs := []string{"example1", "example2"}

    for _, input := range inputs {
        refAST, _ := ReferenceParser(input)
        prodAST, _ := ProductionParser(input)

        if !grammar.ASTEqual(refAST, prodAST) {
            diff := grammar.ASTDiff(refAST, prodAST)
            t.Errorf("parsers disagree on %q: %s", input, diff)
        }
    }
}
```

## API Reference

### Grammar Types

```go
type Grammar struct {
    Rules   []*Rule
    RuleMap map[string]*Rule
}

type Rule struct {
    Name       string
    Expression Expression
    Comment    string
}

type Expression interface{}
// Implementations: Terminal, NonTerminal, Sequence, Alternation,
// Optional, Repetition, OneOrMore, Grouping
```

### Functions

```go
// Parse EBNF grammar string
func ParseEBNF(input string) (*Grammar, error)

// Validate grammar (check for undefined rule references)
func (g *Grammar) Validate() error

// Generate test cases from grammar
func (g *Grammar) GenerateTests(options TestOptions) []TestCase

// Create coverage tracker
func NewCoverageTracker(g *Grammar) *CoverageTracker

// Compare ASTs for equality
func ASTEqual(a, b ast.SchemaNode) bool

// Get diff description
func ASTDiff(a, b ast.SchemaNode) string

// Default test generation options
func DefaultOptions() TestOptions
```

## Design Philosophy

The grammar package follows these principles:

1. **Grammar-as-Verification**: Grammars are executable specifications that verify parser correctness
2. **Parser Technology Freedom**: Works with any parsing technique (LL, Pratt, PEG, combinators)
3. **Pragmatic EBNF**: Custom variant optimized for Shape's use cases, not ISO 14977
4. **Optional Validation**: Allows grammar fragments for documentation/examples
5. **Test Generation**: Automatic test generation reduces manual test writing
6. **Coverage Guidance**: Coverage metrics guide test completeness

## Integration with Shape

The grammar package integrates with Shape's architecture:

- **Tokenizer Integration**: EBNF parser uses `pkg/tokenizer` for lexing
- **AST Comparison**: Compares `pkg/ast` nodes across parsers
- **Parser Verification**: Validates format parsers (shape-jsonv, shape-yamlv, etc.)
- **Documentation**: Grammars serve as formal specifications in docs

## Related Documentation

- [ADR 0005: Grammar-as-Verification](../../docs/architecture/decisions/0005-grammar-as-verification.md)
- [ADR 0004: Parser Strategy](../../docs/architecture/decisions/0004-parser-strategy.md)
- [Parser Implementation Guide](../../docs/PARSER_IMPLEMENTATION_GUIDE.md)
- [Examples Directory](./examples/README.md)

## Test Coverage

- **36 of 37 tests passing (97.3%)**
- All core functionality tested
- Example grammars verified
- Guide patterns validated

## License

Copyright 2025 Shapestone. All rights reserved.
