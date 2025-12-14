# EBNF Grammar Examples

This directory contains example EBNF grammars demonstrating various features of the Shape grammar package.

## Examples

### `boolean.ebnf`
Simple boolean expression grammar showing basic alternation and terminals.

**Features:**
- Terminal strings
- Simple alternation (`|`)

### `arithmetic.ebnf`
Arithmetic expression grammar with operator precedence.

**Features:**
- Optional expressions `[ ... ]`
- Repetition `{ ... }`
- Grouping `( ... )`
- Left-recursion elimination

### `json-schema.ebnf`
JSON-like schema grammar with objects, arrays, and multiple value types.

**Features:**
- Complex alternations
- Optional sequences
- Recursive rules
- Character classes `[a-z]`

### `shape-validation.ebnf`
Shape validation format grammar showing the actual DSL syntax.

**Features:**
- Complete validation DSL
- Function calls with arguments
- Object and array schemas
- Type identifiers and literals

## Usage

```go
package main

import (
    "fmt"
    "os"
    "github.com/shapestone/shape-core/pkg/grammar"
)

func main() {
    // Read grammar file
    content, _ := os.ReadFile("examples/boolean.ebnf")

    // Parse grammar
    g, err := grammar.ParseEBNF(string(content))
    if err != nil {
        panic(err)
    }

    // Generate test cases
    tests := g.GenerateTests(grammar.DefaultOptions())

    for _, test := range tests {
        fmt.Printf("%s: %s (expect %v)\n",
            test.Name, test.Input, test.ShouldSucceed)
    }
}
```

## EBNF Syntax Reference

Shape uses a custom EBNF variant:

- `"literal"` - Terminal string (exact match)
- `[a-z]` - Character class (regex pattern)
- `Identifier` - Non-terminal (rule reference)
- `a | b` - Alternation (choose one)
- `a b` - Sequence (one after another)
- `[ a ]` - Optional (zero or one)
- `{ a }` - Repetition (zero or more)
- `a+` - One or more
- `a*` - Zero or more (same as `{ a }`)
- `( a )` - Grouping
- `// comment` - Single-line comment
