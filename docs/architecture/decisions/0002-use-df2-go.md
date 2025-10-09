# ADR 0002: Integrated Tokenizer Framework

**Status:** Accepted
**Date:** 2025-10-09
**Updated:** 2025-10-09
**Context:** Shape needs to tokenize 6 validation formats

## Context

Shape must parse 6 different validation schema formats (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV). Each format requires tokenization before parsing. We need to decide whether to:

1. Build custom tokenizers from scratch for each format
2. Integrate a proven tokenizer framework
3. Use Go's standard library (text/scanner)
4. Use a third-party parser generator

## Decision

We will **integrate a proven tokenizer framework** (originally df2-go) as the foundation for tokenizing most validation formats (JSONV, XMLV, PropsV, CSVV, TEXTV).

The framework is integrated into shape at `internal/streams/`, `internal/tokens/`, `internal/text/`, and `internal/numbers/`. Each format will implement custom matchers using the framework's matcher composition API, then build format-specific recursive descent parsers on top.

## Rationale

### Why an Integrated Framework?

**1. Proven Foundation**
- Battle-tested with 21+ JSON tests
- Production-ready tokenizer framework
- Clear, well-architected codebase

**2. Essential Built-in Features**
- **UTF-8 Support:** Native rune-based processing (critical for international schemas)
- **Backtracking:** Stream cloning enables speculative matching
- **Position Tracking:** Automatic line/column tracking for error messages
- **Matcher Composition:** Functional approach to building complex patterns

**3. Time Efficiency**
- **60-70% effort savings** vs building from scratch
- Integrated framework: ~40-60 hours for all formats
- From scratch: ~100-150 hours
- Reuse proven patterns instead of reinventing

**4. Self-Contained**
- No external dependencies beyond Go standard library and google/uuid
- Full control over the codebase
- Can evolve framework as needed for shape's requirements

**5. Excellent Fit for Validation Formats**

| Format | Framework Compatibility | Notes |
|--------|----------------------|-------|
| JSONV  | Excellent (80% reuse) | Framework has JSON tokenizer |
| XMLV   | Good | Need XML-specific matchers |
| PropsV | Excellent | Simpler than JSON |
| CSVV   | Good | Line-oriented processing |
| TEXTV  | Excellent | Pattern-based matching |
| YAMLV  | Poor | Consider 3rd-party YAML parser |

### Key Framework Features for Shape

**Stream Cloning (Backtracking):**
```go
// Try to match function call, fall back to type identifier
clone := stream.Clone()
if token := functionMatcher(clone); token != nil {
    stream.Match(clone)  // Success, update parent
    return token
}
// Failed, try identifier (parent stream unchanged)
return identifierMatcher(stream)
```

**Matcher Composition:**
```go
// Build complex patterns from simple pieces
identifierPattern := Sequence(
    CharMatcher(isUpper),           // First char uppercase
    ZeroOrMore(CharMatcher(isAlnum))  // Rest alphanumeric
)
```

**Position Tracking:**
```go
// Automatic position in every token
token := tokenizer.NextToken()
fmt.Printf("Error at line %d, column %d", token.Row(), token.Column())
```

## Alternatives Considered

### 1. Build from Scratch

**Pros:**
- Full control, no dependencies
- Optimized for exact use case

**Cons:**
- 2-3x more effort
- Need to implement UTF-8, backtracking, position tracking
- Reinventing battle-tested code
- More testing burden

**Rejected:** Not worth effort when proven solution exists

### 2. Go's text/scanner

**Pros:**
- Standard library (no dependency)
- Basic tokenization support

**Cons:**
- No backtracking support
- Limited composition
- Still need parser logic
- Not much simpler than integrated framework

**Rejected:** Doesn't provide enough value over integrated framework

### 3. Parser Generator (yacc, antlr)

**Pros:**
- Formal grammar definition
- Generated code

**Cons:**
- Overkill for validation formats
- Additional dependency and tooling
- Less flexible for ambiguous syntax
- Harder to debug

**Rejected:** Validation formats are better suited to hand-written parsers

### 4. Third-Party Parser Libraries

**Pros:**
- May have rich features
- Community support

**Cons:**
- External dependency risk
- May not fit validation format needs
- Learning curve for team
- Less control

**Rejected:** Integrated framework is already available and fits perfectly

## Integration Strategy

### Module Configuration

```go
// shape/go.mod
module github.com/shapestone/shape

go 1.25

require github.com/google/uuid v1.6.0
```

**Note:** The tokenizer framework is integrated into shape's codebase, not imported as an external dependency.

### Format-Specific Tokenizer Pattern

```go
package jsonv

import (
    "github.com/shapestone/shape/internal/streams"
    "github.com/shapestone/shape/internal/tokens"
)

// Custom matcher for type identifiers
func identifierMatcher(stream streams.Stream) *tokens.Token {
    // Implementation
}

// Matcher list: built-in + custom
var Matchers = []tokens.Matcher{
    tokens.CharMatcher("ObjectStart", '{'),  // Reuse
    tokens.CharMatcher("ObjectEnd", '}'),    // Reuse
    identifierMatcher,                        // Custom
    functionMatcher,                          // Custom
    // ... more matchers
}

// Initialize tokenizer
tokenizer := tokens.NewTokenizer(Matchers...)
tokenizer.Initialize(input)
```

### Parser Pattern

```go
package jsonv

type Parser struct {
    tokenizer *tokens.Tokenizer
    current   *tokens.Token
}

func (p *Parser) Parse(input string) (ast.SchemaNode, error) {
    // Initialize tokenizer with custom matchers
    t := tokens.NewTokenizer(Matchers...)
    t.Initialize(input)
    p.tokenizer = &t
    
    // Parse using recursive descent
    return p.parseValue()
}
```

## Consequences

### Positive

- **60-70% Code Reuse:** Massive time savings over building from scratch
- **UTF-8 Support Built-in:** Essential for international schemas
- **Backtracking Support:** Handles ambiguous validation syntax elegantly
- **Position Tracking:** Precise error messages out of the box
- **Clean Separation:** Tokenization (framework) vs Parsing (format-specific)
- **Battle-Tested:** Production-ready tokenization framework
- **Self-Contained:** No external dependencies beyond standard library + uuid
- **Full Control:** Can evolve framework as needed for shape's requirements

### Negative

- **Code Size:** Adds ~2000 lines of framework code to shape
- **Learning Curve:** Team needs to learn framework API
- **Still Need Parsers:** Framework only handles tokenization, parsers still needed

### Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Framework bugs | Comprehensive testing, fix directly |
| Team unfamiliarity | Documentation, examples, training |
| Performance issues | Benchmarking, profiling, optimize directly |
| Code maintenance burden | Framework is stable, minimal changes needed |

**Overall Risk:** LOW (self-contained, full control)

## YAML Exception

YAMLV format is an exception to this decision. YAML has complex features (indentation, anchors, aliases) that are difficult to implement with the tokenizer framework. For YAMLV, we will evaluate using an existing YAML parser (e.g., `gopkg.in/yaml.v3`) with a validation expression overlay.

## Background

This integrated tokenizer framework was originally developed as the df2-go project. It has been integrated directly into shape to create a self-contained, dependency-free validation schema parser. This maintains architectural consistency while eliminating external dependencies.

## Success Metrics

- All 6 formats tokenize correctly
- Error messages include line/column numbers
- UTF-8 schemas parse correctly
- Tokenization performance < 500μs for typical schemas
- 95%+ test coverage for tokenizers
- Clear documentation for adding new formats

## References

- Original Framework: df2-go (github.com/shapestone/df2-go) - now integrated
- Backtracking Parsers: https://en.wikipedia.org/wiki/Backtracking
- Go Modules: https://go.dev/doc/modules/

## Date

2025-10-09
