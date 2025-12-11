# ADR 0004: LL(1) Recursive Descent Parser Strategy

**Status:** Accepted

**Date:** 2025-12-07

## Context

Shape provides parser infrastructure for data format parsers (JSON, XML, YAML, CSV, and custom formats). We needed to choose a parsing strategy that would:

1. Be simple and maintainable
2. Produce clear, actionable error messages
3. Support multiple data formats
4. Be fast enough for production use
5. Work well with our tokenizer-based architecture

## Decision

We use **LL(1) recursive descent parsing** for all schema parsers.

**LL(1)** means:
- **L**eft-to-right scan of tokens
- **L**eftmost derivation (builds parse tree from left)
- **1** token lookahead

**Recursive Descent** means:
- Each grammar rule becomes a function
- Functions call each other recursively
- Parser structure mirrors grammar structure

## Freedom to Choose Parser Technology

**Critical insight:** By maintaining hand-coded parsers (rather than using parser generators or meta-grammars), Shape parser projects retain **complete freedom** to choose any parsing technique.

While we currently use LL(1) recursive descent for most parsers, individual parser projects can choose different techniques if beneficial:

- **LL(1) Recursive Descent** (current default)
  - O(n) linear time complexity
  - Zero backtracking overhead
  - Best error messages (full context available)
  - Most debuggable (call stack = parse tree)

- **Pratt Parsing** (for operator precedence)
  - If a format needs complex operator precedence rules
  - More elegant than precedence climbing
  - Still hand-coded with full error control

- **Packrat/PEG** (for memoization)
  - If a format has expensive backtracking patterns
  - Memoization provides O(n) guarantee
  - More complex but handles harder grammars

- **Parser Combinators** (for functional composition)
  - Higher-order functions compose parsers
  - Grammar-like code structure
  - Trade-off: harder error messages, but more compositional

- **Hand-optimized Hybrid**
  - Mix techniques per format's needs
  - LL(1) for most rules, Pratt for expressions
  - Optimize hot paths independently

**This flexibility is a key advantage of hand-coded parsers.** With a parser generator or meta-grammar system, all formats would be locked into one technique. Hand-coding lets each format choose the best approach.

**Current status:** All parser projects currently use LL(1) recursive descent because it provides the best combination of simplicity, performance, and error quality for schemas. But the door remains open for different techniques if needed.

## Implementation Characteristics

### Single Token Lookahead

```go
type Parser struct {
    tokenizer *tokenizer.Tokenizer
    current   *tokenizer.Token  // Single current token
    hasToken  bool
}

// peek returns the current token without advancing
func (p *Parser) peek() *tokenizer.Token {
    if p.hasToken {
        return p.current
    }
    return nil
}
```

### Predictive Parsing

Decision made by examining current token only:

```go
func (p *Parser) parseValue() (ast.SchemaNode, error) {
    // Predictive dispatch based on current token
    switch p.current.Kind() {
    case TokenObjectStart:
        return p.parseObject()
    case TokenArrayStart:
        return p.parseArray()
    case TokenString:
        return p.parseLiteralString()
    case TokenNumber:
        return p.parseLiteralNumber()
    case TokenFunction:
        return p.parseFunction()
    case TokenIdentifier:
        return p.parseType()
    default:
        return nil, parser.NewUnexpectedTokenError(
            p.position(),
            "value",
            p.current.Kind(),
        )
    }
}
```

### No Backtracking

Parser commits to a decision after examining current token. No speculative parsing or trying multiple alternatives:

```go
func (p *Parser) parseObject() (ast.SchemaNode, error) {
    // Commit: we know we're parsing an object
    if _, err := p.expect(TokenObjectStart); err != nil {
        return nil, err  // Error, don't try alternatives
    }
    // Continue with object parsing...
}
```

## Advantages

### 1. Simplicity
- Parser structure directly mirrors grammar rules
- Easy to understand and maintain
- Each production rule is a function
- Clear code-to-grammar correspondence

### 2. Predictability
- Deterministic parsing decisions
- No hidden complexity from backtracking
- Linear time complexity O(n)
- Consistent performance

### 3. Clear Error Messages
- Errors detected immediately
- Know exact position and expected token
- Can provide helpful hints ("Did you mean X?")
- Error recovery is straightforward

```go
return nil, parser.NewUnexpectedTokenError(
    p.position(),           // Exact position
    "} or ,",               // What was expected
    p.current.Kind(),       // What was found
)
```

### 4. Fast Execution
- No backtracking overhead
- Single pass through tokens
- Minimal state management
- Cache-friendly (linear access pattern)

### 5. Debuggability
- Call stack reflects parse tree structure
- Easy to add logging at any production rule
- Breakpoints map directly to grammar rules
- Clear execution flow

### 6. Sufficient for Schemas
- Schema grammars are naturally LL(1)
- No complex ambiguities
- Token types are distinct
- Grammar design can be LL(1)-friendly

## Caveats and Limitations

### 1. Grammar Must Be LL(1)
**Limitation:** Can only parse LL(1) grammars

**Impact:**
- Grammar must be deterministic
- No left recursion allowed
- First(α) ∩ First(β) = ∅ for alternatives
- Follow sets must not overlap

**Example of what we CAN'T parse:**
```
// Left-recursive grammar (not LL(1))
expr → expr '+' term | term

// Ambiguous alternatives (not LL(1))
value → 'x' 'y' | 'x' 'z'  // Can't decide after seeing 'x'
```

**Mitigation:** Design schema grammars to be LL(1)-compatible from the start.

### 2. Single Token Lookahead
**Limitation:** Can only look 1 token ahead

**Impact:**
- Can't disambiguate if decision requires looking at 2+ tokens
- Some natural language grammars need LL(k) with k>1

**Example:**
```
// Requires 2 token lookahead (not LL(1))
statement → 'if' '(' expr ')' | 'if' ident

// Must see both 'if' AND next token to decide
```

**Mitigation:** Schema formats don't need multi-token lookahead. Tokenizer can combine tokens if needed (e.g., recognize "if (" as single token).

### 3. No Ambiguity Resolution
**Limitation:** Can't handle ambiguous grammars

**Impact:**
- Grammar must be unambiguous
- Can't use precedence to resolve conflicts
- Can't prefer one parse over another

**Mitigation:** Schemas are naturally unambiguous (types, functions, literals are distinct).

### 4. Grammar Structure Constraints
**Limitation:** Grammar must avoid certain patterns

**Constraints:**
- No left recursion (direct or indirect)
- Alternatives must have disjoint first sets
- Optional/repeated productions need careful design

**Mitigation:** Our grammars are simple enough that these constraints are easy to satisfy.

## Alternatives Considered

### PEG (Parsing Expression Grammars)
**Pros:**
- Ordered choice handles ambiguity
- More expressive than LL(1)
- Packrat parsing can be efficient

**Cons:**
- More complex implementation
- Hidden backtracking complexity
- Less predictable performance
- Error messages can be confusing
- Overkill for schemas

**Decision:** Rejected. LL(1) is simpler and sufficient.

### Parser Generators (yacc/bison/ANTLR)
**Pros:**
- Declarative grammar specification
- Handles more grammar classes (LALR, LL(*))
- Battle-tested tools

**Cons:**
- External dependencies
- Generated code is hard to debug
- Less control over error messages
- Build complexity
- Learning curve for grammar syntax

**Decision:** Rejected. Hand-written parsers provide better control and clarity.

### LL(k) with k>1
**Pros:**
- More expressive than LL(1)
- Can handle more grammar patterns
- Still deterministic and predictable

**Cons:**
- More complex implementation
- Requires k-token lookahead buffer
- Unnecessary for schemas
- Diminishing returns (LL(2) rarely needed)

**Decision:** Rejected. LL(1) is sufficient for all our formats.

### LR/LALR Parsing
**Pros:**
- More powerful (can handle left-recursive grammars)
- Bottom-up parsing
- Handles more grammars than LL(k)

**Cons:**
- Complex implementation (state machines, shift/reduce)
- Error messages are harder to generate
- Less intuitive (parser structure doesn't mirror grammar)
- Overkill for our use case

**Decision:** Rejected. Too complex for schemas.

## Consequences

### For Grammar Design
- **Must design LL(1)-compatible grammars**
  - Avoid left recursion
  - Ensure disjoint first sets for alternatives
  - Use tokenizer to disambiguate when possible

- **Example LL(1) grammar:**
  ```
  value    → object | array | function | type | literal
  object   → '{' (prop (',' prop)*)? '}'
  array    → '[' value ']'
  function → IDENTIFIER '(' args ')'
  type     → IDENTIFIER
  literal  → STRING | NUMBER | TRUE | FALSE | NULL
  ```

### For Tokenizer Design
- **Tokenizer must produce unambiguous tokens**
  - Distinguish keywords from identifiers
  - Handle multi-character operators
  - Combine characters that form single syntactic unit

### For Error Handling
- **Errors are deterministic and immediate**
  - Know exactly what was expected
  - Can provide context and suggestions
  - Recovery strategies are straightforward

### For Performance
- **Linear time complexity**
  - O(n) where n is number of tokens
  - No backtracking overhead
  - Predictable performance

### For Testing
- **Easy to test individual production rules**
  - Each function can be tested independently
  - Grammar coverage maps to function coverage
  - Error cases are deterministic

### For Documentation
- **Grammar documentation maps directly to code**
  - Function names correspond to grammar rules
  - Comments can reference grammar productions
  - Easy to understand parser structure

## References

- [Recursive Descent Parsing](https://en.wikipedia.org/wiki/Recursive_descent_parser)
- [LL Parser](https://en.wikipedia.org/wiki/LL_parser)
- [Crafting Interpreters - Recursive Descent Parsing](https://craftinginterpreters.com/parsing-expressions.html)
- Dragon Book (Compilers: Principles, Techniques, and Tools) - Chapter 4

## Examples in Shape

All format parsers follow this pattern in their respective repositories (shape-json, shape-yaml, shape-xml, shape-csv, etc.):

- Parser implementations at: `internal/parser/{format}/parser.go`

All implement the same LL(1) recursive descent pattern with:
- Single token lookahead (`peek()`)
- Predictive parsing (`switch` on token type)
- No backtracking
- Clear error reporting
