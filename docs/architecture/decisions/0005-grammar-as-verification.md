# ADR 0005: Grammar-as-Verification for Parser Correctness

**Status:** Accepted

**Date:** 2025-12-07

## Context

Shape provides parser infrastructure for format parser projects (JSON, XML, CSV, YAML, and other custom formats). We use hand-coded parsers for maximum performance and error quality (see ADR 0004).

However, hand-coded parsers face challenges:

1. **Documentation Drift:** Grammar documentation can become out-of-sync with parser implementation
2. **Correctness Uncertainty:** No automated way to prove parser correctly implements intended grammar
3. **Breaking Change Detection:** Hard to detect when changes make parser more/less restrictive
4. **Test Completeness:** Manual test cases may miss grammar edge cases
5. **Contributor Confidence:** New contributors uncertain whether parser matches documented grammar

**Traditional approaches:**
- Write grammar in documentation (can drift from implementation)
- Write manual tests (incomplete coverage)
- Hope they stay in sync (they don't)

**Better approach:** Make grammar an executable specification that automatically verifies parser correctness.

## Decision

**We use canonical EBNF grammars as executable specifications** that automatically verify hand-coded parser correctness through grammar-based test generation.

**Key principle:** Grammar is the source of truth. Parser correctness is proven by automated tests generated from grammar.

### Two Separate Artifacts

1. **EBNF Grammar** - Canonical specification (mathematical definition of language)
2. **Hand-coded Parser** - Optimized implementation (performance, error messages)

**Relationship:** Grammar generates tests that verify parser correctness.

### Not Compiling Grammar to Parser

**Important:** We are **NOT** using parser generators or meta-grammars (like DFR). We still write hand-coded parsers for performance and error quality. Grammar serves as **verification**, not **generation**.

## Critical Distinction: Schema Parsers vs Data Parsers

**IMPORTANT:** This ADR applies to ALL parsers, but there are two fundamentally different types:

### 1. Data Format Parsers (shape-json, shape-yaml, shape-xml)

**Purpose:** Parse data files into Go types

**Returns:** Go built-in types (`interface{}`, `map[string]interface{}`, `[]interface{}`, `string`, `float64`, `bool`, `nil`)

**Example:**
```go
// Parse JSON data
result, err := json.Parse(`{"name": "Alice", "age": 30}`)
obj := result.(map[string]interface{})  // Go map
name := obj["name"].(string)            // Go string
age := obj["age"].(float64)             // Go float64
```

**AST usage:** Data parsers do **NOT** use Shape's AST nodes. Shape's AST is for validation schemas only.

**See:** [PARSER_IMPLEMENTATION_GUIDE.md](../../PARSER_IMPLEMENTATION_GUIDE.md) for complete data parser implementation details.

### 2. Schema Parsers (Shape's own schema language)

**Purpose:** Parse validation schemas into AST nodes

**Returns:** AST nodes (`*ast.ObjectNode`, `*ast.ArrayNode`, `*ast.TypeNode`, etc.)

**Example:**
```go
// Parse Shape schema definition
schema, err := schema.Parse(`{ "id": UUID, "name": String }`)
obj := schema.(*ast.ObjectNode)  // AST node representing validation rule
idSchema := obj.Properties()["id"]  // *ast.TypeNode for UUID type
```

**AST usage:** Schema parsers **DO** use Shape's AST nodes because they're parsing validation rules, not data.

### Summary

| Parser Type | Input | Returns | Uses AST? | Example |
|-------------|-------|---------|-----------|---------|
| **Data Parser** | Data files (JSON, XML, YAML) | Go types (`map`, `slice`, primitives) | NO | `shape-json`, `shape-yaml` |
| **Schema Parser** | Validation schemas (Shape DSL) | AST nodes (`*ast.ObjectNode`, etc.) | YES | Shape's internal schema parser |

**The examples in this ADR show SCHEMA PARSERS** (parsing Shape's validation language). If you're implementing a **DATA PARSER** (JSON, XML, YAML), your parser should return Go types, not AST nodes.

## Architectural Boundaries

**Critical:** Shape is **infrastructure only**. Parser projects are **self-contained**.

### Shape's Responsibility (Infrastructure)

**Shape provides:**
- Grammar parsing tools (`pkg/grammar/ebnf_parser.go`)
- Test generation infrastructure (`pkg/grammar/test_generator.go`)
- Coverage tracking utilities (`pkg/grammar/coverage_tracker.go`)
- AST node definitions (`pkg/ast/*`)

**Shape does NOT:**
- Contain parser implementations for specific formats
- Test parser projects (shape-json, shape-xml, etc.)
- Know about specific format grammars
- Have dependencies on parser projects

### Parser Project's Responsibility (Self-Contained)

**Each parser project (shape-json, shape-xml, etc.):**
- Maintains its own EBNF grammar (`docs/grammar/{format}.ebnf`)
- Implements its own parser (`internal/parser/parser.go`)
- Tests itself using Shape's infrastructure (`grammar_test.go`)
- Is responsible for its own correctness

**Example: shape-json tests itself**
```go
// shape-json/grammar_test.go
import "github.com/shapestone/shape/pkg/grammar"  // Use Shape's tools

func TestGrammarVerification(t *testing.T) {
    // Load OUR grammar (in our project)
    spec := grammar.ParseEBNF("docs/grammar/json.ebnf")

    // Generate tests from OUR grammar
    tests := spec.GenerateTests(grammar.DefaultOptions())

    // Test OUR parser
    for _, test := range tests {
        result, err := parser.Parse(test.Input)
        // ... verify
    }
}
```

**Shape provides the tools. Parser projects use the tools to test themselves.**

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Canonical EBNF Grammar                             │
│  (Source of Truth)                                  │
│                                                     │
│  docs/grammar/{format}.ebnf                         │
│  - Mathematical specification                       │
│  - Defines accepted language                        │
│  - Maintained alongside parser                      │
└──────────────────┬──────────────────────────────────┘
                   │
                   │ (generates)
                   ▼
┌─────────────────────────────────────────────────────┐
│  Automated Test Generation                          │
│                                                     │
│  - Parse EBNF grammar                               │
│  - Generate valid inputs (all paths)                │
│  - Generate invalid inputs (violations)             │
│  - Generate edge cases (empty, nested)              │
│  - Generate coverage reports                        │
└──────────────────┬──────────────────────────────────┘
                   │
                   │ (validates)
                   ▼
┌─────────────────────────────────────────────────────┐
│  Hand-Coded Parser Implementation                   │
│  (Optimized for Performance and Error Quality)      │
│                                                     │
│  internal/parser/{format}/parser.go                 │
│  - LL(1) recursive descent (or other technique)     │
│  - Precise error messages                           │
│  - Performance optimized                            │
│  - Full control over behavior                       │
└─────────────────────────────────────────────────────┘
```

## Verification Approaches

### 1. Grammar-Based Test Generation

**Automatically generate test cases from EBNF:**

```go
// Parse canonical grammar
grammar := ParseEBNF("docs/grammar/{format}.ebnf")

// Generate comprehensive test cases
testCases := GenerateTestCases(grammar, TestOptions{
    MaxDepth:       5,      // Nesting depth
    CoverAllRules:  true,   // Exercise every production
    EdgeCases:      true,   // Empty, single, multiple
    InvalidCases:   true,   // Violate each rule
})

// Verify parser against generated tests
for _, tc := range testCases {
    result, err := parser.Parse(tc.Input)

    if tc.ShouldSucceed {
        assert.NoError(t, err, "Grammar says valid: %s", tc.Input)
        assert.NotNil(t, result)
    } else {
        assert.Error(t, err, "Grammar says invalid: %s", tc.Input)
    }
}
```

**Example generated tests for:**
```ebnf
ObjectNode = "{" [ Property { "," Property } ] "}" ;
```

Generated:
- `{}` - Empty object (valid, optional properties)
- `{ "id": UUID }` - Single property (valid)
- `{ "a": String, "b": Int }` - Multiple properties (valid)
- `{` - Missing `}` (invalid)
- `{ "a" }` - Missing `: Value` (invalid)
- `{ "a": UUID "b": Int }` - Missing `,` (invalid)

### 2. Grammar Coverage Analysis

**Track which grammar rules are exercised:**

```go
type CoverageTracker struct {
    rulesInvoked map[string]int
    grammar      *Grammar
}

func TestGrammarCoverage(t *testing.T) {
    grammar := ParseEBNF("docs/grammar/{format}.ebnf")
    allRules := grammar.GetAllRules()

    tracker := NewCoverageTracker(grammar)
    RunAllTests(tracker)

    // Verify 100% grammar coverage
    for _, rule := range allRules {
        assert.True(t, tracker.rulesInvoked[rule] > 0,
            "Grammar rule '%s' never tested", rule)
    }

    // Report
    coverage := float64(len(tracker.rulesInvoked)) / float64(len(allRules))
    fmt.Printf("Grammar coverage: %.1f%%\n", coverage*100)
}
```

### 3. Property-Based Testing from Grammar

**Generate random inputs constrained by grammar:**

```go
func TestParserProperties(t *testing.T) {
    grammar := ParseEBNF("docs/grammar/{format}.ebnf")
    generator := NewGrammarGenerator(grammar)

    // Property: Parser never panics
    rapid.Check(t, func(t *rapid.T) {
        input := generator.Generate(t)
        assert.NotPanics(t, func() {
            parser.Parse(input)
        })
    })

    // Property: Valid inputs always parse
    rapid.Check(t, func(t *rapid.T) {
        validInput := generator.GenerateValid(t)
        _, err := parser.Parse(validInput)
        assert.NoError(t, err)
    })
}
```

### 4. Dual Parser Verification (Optional)

**Compare hand-coded parser against reference parser:**

```go
func TestParserMatchesReference(t *testing.T) {
    // Reference parser generated from EBNF (correct but slow)
    refParser := CompileEBNF(grammar)

    // Hand-coded production parser (fast, good errors)
    prodParser := NewFormatParser()

    for _, input := range testInputs {
        refAST, refErr := refParser.Parse(input)
        prodAST, prodErr := prodParser.Parse(input)

        // Both must agree on success/failure
        assert.Equal(t, refErr != nil, prodErr != nil)

        // If both succeed, ASTs must be equivalent
        if refErr == nil && prodErr == nil {
            assert.True(t, ASTEqual(refAST, prodAST))
        }
    }
}
```

## Grammar Format

**We use a custom EBNF variant** optimized for readability and developer familiarity:

```ebnf
// Production rules
rule_name = expression ;

// Operators
[ ]     Optional (0 or 1)
+       One or more (suffix)
*       Zero or more (suffix)
{ }     Zero or more (alternative)
( )     Grouping
|       Alternation

// Character notation (regex-like)
Digit = [0-9] ;
Letter = [a-zA-Z] ;
Hex = [0-9a-fA-F] ;

// Example: Format Grammar
ObjectNode = "{" [ Property { "," Property } ] "}" ;
Property   = StringLiteral ":" Value ;
Value      = Literal | Type | Function | Object | Array ;
```

**Note:** This is a pragmatic variant, not ISO 14977 compliant. Prioritizes readability and familiarity over standards compliance.

## Grammar as Implementation Guide (LLM Assistance)

**Critical use case:** EBNF grammar should be detailed enough for LLMs to generate correct hand-coded parser implementations.

### Grammar with Implementation Hints

Grammars should include comments that guide implementation:

```ebnf
// Format Grammar Specification
// This grammar defines the Format schema language.
//
// Implementation Guide:
// - Use LL(1) recursive descent parsing (see ADR 0004)
// - Each production rule becomes a parse function
// - Return appropriate ast.SchemaNode types (see Shape pkg/ast)
// - Provide context-aware error messages at each rule

// Top-level value can be any schema type
// Parser function: parseValue() -> ast.SchemaNode
// Dispatch based on current token kind
Value = Literal | Type | Function | Object | Array ;

// Object schema with typed properties
// Parser function: parseObject() -> *ast.ObjectNode
// Returns: ast.NewObjectNode(properties map[string]ast.SchemaNode, position)
// Example valid input: { "id": UUID, "name": String }
// Example invalid: { id: UUID } (missing quotes on key)
ObjectNode = "{" [ Property { "," Property } ] "}" ;

// Property: key-value pair in object
// Parser function: parseProperty() -> (key string, value ast.SchemaNode)
// Note: Key MUST be string literal, not bare identifier
// Error hint: "Property keys must be quoted strings"
Property = StringLiteral ":" Value ;

// Type reference (built-in or custom)
// Parser function: parseType() -> *ast.TypeNode
// Returns: ast.NewTypeNode(typeName string, position)
// Examples: UUID, String, Email, CustomType
// Tokenizer: Matches identifier, no parentheses following
Type = Identifier ;

// Function call with arguments
// Parser function: parseFunction() -> *ast.FunctionNode
// Returns: ast.NewFunctionNode(functionName string, args []ast.SchemaNode, position)
// Examples: Integer(0, 100), Email(allowDomains("example.com"))
// Tokenizer: Identifier followed by '('
Function = Identifier "(" [ Args ] ")" ;

// Terminal: String literal
// Tokenizer: StringMatcherFunc matches quoted strings "..."
// Returns: string value with quotes removed
// Error: Unterminated string if no closing quote
StringLiteral = '"' [^"]* '"' ;

// Terminal: Identifier
// Tokenizer: RegexMatcherFunc matches [a-zA-Z_][a-zA-Z0-9_]*
// Returns: identifier string
// Used for: Type names, Function names, but NOT property keys
Identifier = [a-zA-Z_][a-zA-Z0-9_]* ;
```

### LLM-Assisted Parser Development Workflow

**1. Start with Grammar + AST Definitions**

Provide LLM with:
- EBNF grammar with implementation hints
- Shape's AST node definitions (`pkg/ast/`)
- Parsing strategy (ADR 0004: LL(1) recursive descent)

**2. LLM Generates Parser Skeleton**

LLM can generate:
```go
// parseValue implements: Value = Literal | Type | Function | Object | Array
func (p *Parser) parseValue() (ast.SchemaNode, error) {
    switch p.peek().Kind() {
    case TokenLBrace:
        return p.parseObject()
    case TokenLBracket:
        return p.parseArray()
    case TokenString:
        return p.parseLiteral()
    case TokenIdentifier:
        // Lookahead to distinguish Type vs Function
        if p.peekNext().Kind() == TokenLParen {
            return p.parseFunction()
        }
        return p.parseType()
    default:
        return nil, fmt.Errorf("expected value at %s, got %s",
            p.position(), p.peek().Kind())
    }
}

// parseObject implements: ObjectNode = "{" [ Property { "," Property } ] "}"
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    startPos := p.position()

    if _, err := p.expect(TokenLBrace); err != nil {
        return nil, err
    }

    properties := make(map[string]ast.SchemaNode)

    // Optional properties
    if p.peek().Kind() != TokenRBrace {
        // First property
        key, value, err := p.parseProperty()
        if err != nil {
            return nil, err
        }
        properties[key] = value

        // Additional properties
        for p.peek().Kind() == TokenComma {
            p.advance() // consume comma

            key, value, err := p.parseProperty()
            if err != nil {
                return nil, fmt.Errorf("in object property after comma: %w", err)
            }
            properties[key] = value
        }
    }

    if _, err := p.expect(TokenRBrace); err != nil {
        return nil, err
    }

    return ast.NewObjectNode(properties, startPos), nil
}
```

**3. Human Refines for Quality**

Developer improves:
- Error messages (add context, suggestions)
- Performance (optimize hot paths)
- Edge cases (better handling)
- Tests (add specific scenarios)

### Benefits of LLM-Assisted Development

**For new parser projects:**
1. ✅ **Fast initial implementation** - LLM generates working parser from grammar
2. ✅ **Consistent structure** - All parsers follow same LL(1) pattern
3. ✅ **Correct AST usage** - LLM uses right node types from grammar hints
4. ✅ **Grammar-parser alignment** - Generated code matches specification
5. ✅ **Learning tool** - Developers see how grammar maps to code

**For contributors:**
- Grammar is executable specification
- LLM can help implement missing features
- Faster onboarding (grammar → code is explicit)
- Consistent with existing parser patterns

**Quality assurance:**
- Grammar-based tests verify LLM-generated code
- Human review for error message quality
- Performance profiling for optimization
- Final parser is hand-coded (optimized) but grammar-verified (correct)

### Example LLM Prompt

```
Task: Implement a parser for the Format schema language.

Inputs:
1. EBNF Grammar: docs/grammar/{format}.ebnf (with implementation hints)
2. AST Definitions: github.com/shapestone/shape/pkg/ast
3. Parsing Strategy: LL(1) recursive descent (ADR 0004)
4. Tokenizer: Already implemented, provides Token stream

Generate:
- Parser struct with tokenizer field
- One parse function per grammar production rule
- Functions return appropriate ast.SchemaNode types
- Use single token lookahead (peek())
- Provide context-aware error messages

Example grammar rule:
  ObjectNode = "{" [ Property { "," Property } ] "}" ;

Should generate:
  func (p *Parser) parseObject() (*ast.ObjectNode, error)

Include:
- Token consumption (expect(), advance())
- Error handling with position and context
- AST node construction with correct types
```

**Result:** LLM generates 80% correct parser that humans refine for production quality.

### EBNF Fragments in Code Documentation

**Use grammar fragments directly in parser code comments** to create clear mapping between specification and implementation.

#### Example: Documenting Parser Functions with Grammar Rules

```go
// parseObject parses an object node.
//
// Grammar:
//   ObjectNode = "{" [ Property { "," Property } ] "}" ;
//
// Returns *ast.ObjectNode with properties map.
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    startPos := p.position()

    if _, err := p.expect(TokenLBrace); err != nil {  // "{"
        return nil, err
    }

    properties := make(map[string]ast.SchemaNode)

    // [ Property { "," Property } ]  - Optional property list
    if p.peek().Kind() != TokenRBrace {
        key, value, err := p.parseProperty()  // Property
        if err != nil {
            return nil, err
        }
        properties[key] = value

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
//   Property = StringLiteral ":" Value ;
//
// Returns (key string, value ast.SchemaNode).
func (p *Parser) parseProperty() (string, ast.SchemaNode, error) {
    // StringLiteral
    keyToken, err := p.expect(TokenString)
    if err != nil {
        return "", nil, fmt.Errorf("property key must be string literal: %w", err)
    }
    key := keyToken.Value()

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

// parseValue parses any schema value type.
//
// Grammar:
//   Value = Literal | Type | Function | Object | Array ;
//
// Dispatch based on token type (LL(1) predictive parsing).
func (p *Parser) parseValue() (ast.SchemaNode, error) {
    switch p.peek().Kind() {
    case TokenLBrace:     // Object
        return p.parseObject()
    case TokenLBracket:   // Array
        return p.parseArray()
    case TokenString, TokenNumber, TokenTrue, TokenFalse, TokenNull:  // Literal
        return p.parseLiteral()
    case TokenIdentifier:
        // Type | Function (distinguish with lookahead)
        if p.peekNext().Kind() == TokenLParen {
            return p.parseFunction()  // Function
        }
        return p.parseType()  // Type
    default:
        return nil, fmt.Errorf("expected value at %s, got %s",
            p.position(), p.peek().Kind())
    }
}
```

#### Benefits of EBNF in Code Comments

**1. Direct Specification-Implementation Mapping**
```go
// Grammar fragment shows EXACTLY what this function parses
// Grammar:
//   ObjectNode = "{" [ Property { "," Property } ] "}" ;
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    // Implementation directly follows grammar structure
}
```

**2. Code Reviews Made Easy**

Reviewer can verify:
- ✅ Does implementation match grammar?
- ✅ Are all grammar elements handled?
- ✅ Is error handling complete?

**Example review comment:**
> "Grammar shows `[ Property ... ]` is optional, but code doesn't handle empty object. Need check for `TokenRBrace` before parsing first property."

**3. Maintenance and Refactoring**

When grammar changes:
```diff
  // Grammar:
- ObjectNode = "{" Property { "," Property } "}" ;
+ ObjectNode = "{" [ Property { "," Property } ] "}" ;

  func (p *Parser) parseObject() (*ast.ObjectNode, error) {
      // ...
-     // MUST parse at least one property
+     // Optional properties - check for empty object
  }
```

Grammar comment shows what changed, guiding refactoring.

**4. Self-Documenting Code**

No need to hunt for grammar specification:
```go
// Grammar right here in the code:
//   Property = StringLiteral ":" Value ;
func (p *Parser) parseProperty() (string, ast.SchemaNode, error) {
    // Clear what we're parsing
}
```

**5. LLM Context for Code Changes**

When LLM assists with modifications:
- Grammar fragment provides specification
- LLM understands what function should do
- Changes remain consistent with grammar
- No need to fetch separate grammar file

**6. Inline Documentation of Complex Rules**

For complex grammar rules, annotate the mapping:
```go
// parseDecimalLiteral parses decimal number literals.
//
// Grammar:
//   DecimalLiteral = [ "+" | "-" ] Digit+ [ "." Digit+ ] [ ExponentPart ] ;
//   ExponentPart   = ( "e" | "E" ) [ "+" | "-" ] Digit+ ;
//
// Examples: 123, -456, 3.14, 2.5e10, -1.23E-4
func (p *Parser) parseDecimalLiteral() (*ast.LiteralNode, error) {
    // [ "+" | "-" ]       - Optional sign
    sign := p.parseOptionalSign()

    // Digit+              - Integer part (required)
    intPart := p.parseDigits()

    // [ "." Digit+ ]      - Optional fraction
    fracPart := p.parseOptionalFraction()

    // [ ExponentPart ]    - Optional exponent
    expPart := p.parseOptionalExponent()

    return p.buildDecimalLiteral(sign, intPart, fracPart, expPart), nil
}
```

**7. Testing Coverage Visibility**

Grammar fragments help identify test gaps:
```go
// Grammar:
//   Value = Literal | Type | Function | Object | Array ;
func (p *Parser) parseValue() (ast.SchemaNode, error) {
    // Need tests for: ✓ Literal, ✓ Type, ✓ Function, ✗ Object, ✗ Array
}
```

#### Best Practices

**1. Include Grammar Fragment in Function Comment**
```go
// parseX parses...
//
// Grammar:
//   X = ... ;
```

**2. Inline Comments for Grammar Elements**
```go
if _, err := p.expect(TokenLBrace); err != nil {  // "{"
    return nil, err
}
```

**3. Mark Optional vs Required in Comments**
```go
// [ Property { "," Property } ]  - Optional property list
if p.peek().Kind() != TokenRBrace {
    // ...
}

// Digit+  - Required: one or more digits
digits := p.parseDigits()
```

**4. Reference Grammar File for Full Specification**
```go
// Package parser implements the Format schema parser.
//
// Grammar: See docs/grammar/{format}.ebnf for complete specification.
//
// This parser uses LL(1) recursive descent parsing (see ADR 0004).
// Each production rule in the grammar corresponds to a parse function.
```

**5. Keep Grammar Comments Synchronized**

When grammar changes:
1. Update grammar file (`docs/grammar/{format}.ebnf`)
2. Update code comments to match
3. Grammar-based tests will verify implementation matches

#### Impact on Code Quality

**Before (no grammar fragments):**
```go
// parseObject parses an object
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    // What does this function accept?
    // Need to read docs or reverse-engineer from code
}
```

**After (with grammar fragments):**
```go
// parseObject parses an object node.
//
// Grammar:
//   ObjectNode = "{" [ Property { "," Property } ] "}" ;
//
// Accepts: { "id": UUID, "name": String }
// Accepts: {} (empty object, properties are optional)
// Rejects: { id: UUID } (keys must be quoted)
func (p *Parser) parseObject() (*ast.ObjectNode, error) {
    // Crystal clear what this function parses
}
```

**Result:**
- ✅ Better code comprehension
- ✅ Easier code reviews
- ✅ Faster onboarding
- ✅ Self-documenting code
- ✅ LLM-friendly context
- ✅ Specification-implementation traceability

## Benefits

### 1. Enforced Documentation

**Grammar cannot lie.**

**Traditional documentation:**
```markdown
# Format Parser

Objects have properties separated by commas.
(Last updated: 2023-01-15, may be outdated)
```

**Grammar-as-verification:**
```ebnf
// docs/grammar/{format}.ebnf
// This grammar is verified against parser implementation in CI

ObjectNode = "{" [ Property { "," Property } ] "}" ;
```

**CI verification:**
```bash
$ make grammar-tests
✓ Parser matches grammar specification
✓ 847 test cases generated from grammar
✓ 100% grammar coverage
✓ Documentation proven accurate
```

**If parser changes but grammar doesn't:** Tests fail ❌
**If grammar changes but parser doesn't:** Tests fail ❌
**Only way to pass:** Parser and grammar are synchronized ✅

### 2. Provable Correctness

**Mathematical proof that parser implements grammar:**

```
EBNF Grammar (Mathematical Specification)
    ↓ (generates)
Test Cases (Executable Verification)
    ↓ (validates)
Parser Implementation (Production Code)
```

**Guarantee:** If tests pass, parser accepts exactly the language defined by grammar.

**Not just:** "Parser works on my examples"
**But rather:** "Parser provably implements grammar specification"

### 3. Automatic Breaking Change Detection

**Grammar makes breaking changes explicit:**

```bash
$ git diff docs/grammar/{format}.ebnf
- FunctionNode = Identifier "(" [ Args ] ")" ;
+ FunctionNode = Identifier "(" Args ")" ;  # Args now required!

$ make grammar-tests
⚠ BREAKING CHANGE DETECTED
⚠ Grammar now rejects: Integer()
⚠ Previously valid inputs are now invalid

Update version: 0.9.0 → 1.0.0
```

**CI can detect:**
- Grammar more restrictive → breaking change → major version
- Grammar more permissive → feature addition → minor version
- Grammar unchanged → bug fix → patch version

### 4. Comprehensive Test Coverage

**Manual tests miss edge cases. Grammar-based generation doesn't:**

- ✅ All grammar paths exercised
- ✅ Every production rule tested
- ✅ Edge cases: empty, single, multiple, nested
- ✅ Invalid inputs: violate each rule systematically
- ✅ Coverage reports show gaps

**Example coverage:**
```
Grammar Coverage Report:
  ObjectNode:    ✓ 47 invocations
  Property:      ✓ 89 invocations
  ArrayNode:     ✓ 23 invocations
  FunctionNode:  ✓ 156 invocations

Overall: 100% grammar coverage
All production rules exercised
```

### 5. Contributor Confidence

**New contributor asks: "What does Format accept?"**

**With enforced grammar:**
- "Read `docs/grammar/{format}.ebnf`"
- "That's the canonical spec"
- "Verified in CI every commit"
- **Definitive answer**

**Without:**
- "Um, check the README?"
- "Look at the tests?"
- "Read the parser code?"
- **Uncertain answer**

### 6. Multi-Format Consistency

**Shape has multiple parser projects. Grammar shows consistency:**

```ebnf
// shape-json/docs/grammar/{format}.ebnf
Value = Literal | Type | Function | Object | Array ;

// shape-yaml/docs/grammar/{format}.ebnf
Value = Literal | Type | Function | Object | Array ;  # Same!

// shape-xml/docs/grammar/{format}.ebnf
Value = Literal | Type | Function | Element | Attributes ;  # Different (XML-specific)
```

**Enforced documentation shows:**
- Which formats share identical semantics
- Which formats diverge (and why)
- Cannot accidentally make them inconsistent

## Implementation

### Directory Structure

```
shape-{format}/
├── docs/
│   └── grammar/
│       └── {format}.ebnf           # Canonical specification
├── internal/
│   └── parser/
│       └── {format}/
│           ├── parser.go           # Hand-coded implementation
│           ├── parser_test.go      # Manual tests
│           └── grammar_test.go     # Auto-generated verification tests
└── tools/
    └── grammar-test-gen/           # Test generator (shared across all parsers)
```

### Workflow

**Development:**
1. Define grammar: `docs/grammar/{format}.ebnf`
2. Implement parser: `internal/parser/{format}/parser.go`
3. Generate tests: `go run tools/grammar-test-gen docs/grammar/{format}.ebnf > internal/parser/{format}/grammar_test.go`
4. Run tests: `make test`

**CI Pipeline:**
```bash
# Verify parser matches grammar
make grammar-tests

# Check coverage
make grammar-coverage

# Generate report
make grammar-report
```

**Pull Requests:**
- Grammar change → Must update parser to pass tests
- Parser change → Must update grammar if behavior changes
- CI enforces synchronization

### Test Generation Tool

**Shared infrastructure (lives in Shape):**

```
shape/
└── pkg/
    └── grammar/
        ├── ebnf_parser.go          # Parse EBNF using Shape's tokenizer
        ├── test_generator.go       # Generate test cases from grammar
        ├── coverage_tracker.go     # Track grammar rule coverage
        └── ast_comparator.go       # Compare ASTs for equivalence
```

**Used by all parser projects:**
```go
import "github.com/shapestone/shape/pkg/grammar"

func TestGrammar(t *testing.T) {
    grammar := grammar.ParseEBNF("docs/grammar/{format}.ebnf")
    tests := grammar.GenerateTests(grammar.DefaultOptions())

    for _, test := range tests {
        // Verify parser against generated tests
    }
}
```

## Advantages Over Alternatives

### vs Manual Tests Only
- ✅ Grammar provides mathematical specification
- ✅ Automatic test generation (don't miss edge cases)
- ✅ Documentation enforced (can't drift)
- ✅ Breaking changes detected automatically

### vs Parser Generators (yacc/ANTLR)
- ✅ Keep hand-coded parsers (performance, error quality)
- ✅ Freedom to choose parsing technique per format
- ✅ Full control over error messages
- ✅ Grammar still serves as verification

### vs Meta-Grammars (DFR)
- ✅ Maximum performance (no interpreter overhead)
- ✅ Best error messages (full context available)
- ✅ Most debuggable (transparent execution)
- ✅ Grammar proves correctness without sacrificing quality

## Trade-offs

### What We Gain
✅ **Enforced documentation** - Grammar cannot drift from implementation
✅ **Provable correctness** - Tests prove parser implements grammar
✅ **Automatic coverage** - All grammar paths tested
✅ **Breaking change detection** - CI alerts on incompatible changes
✅ **Contributor confidence** - Grammar is source of truth
✅ **Multi-format consistency** - Compare grammars across formats

### What We Accept
❌ **Maintain two artifacts** - Grammar file + parser code
❌ **Build complexity** - Need test generation tool
❌ **Grammar must stay in sync** - Changes require updating both

**Worth it:** Benefits far outweigh costs. Documentation drift is a serious problem in parser projects. Enforced verification solves it.

## Consequences

### For Parser Projects
- **Must maintain canonical EBNF grammar** in `docs/grammar/`
- **CI verifies parser matches grammar** on every commit
- **Cannot merge if tests fail** - parser must match specification
- **Breaking changes are explicit** - grammar diff shows language changes

### For Documentation
- **Grammar is the documentation** - Most authoritative source
- **README references grammar** - "See docs/grammar/{format}.ebnf for specification"
- **Grammar comments explain rationale** - Why rules exist

### For Testing
- **Grammar-based tests are automatic** - Regenerated when grammar changes
- **Manual tests still valuable** - For specific scenarios, error messages
- **Coverage reports show gaps** - Which grammar rules need more testing

### For Contributors
- **Clear specification** - Grammar defines what parser should accept
- **Confidence in changes** - Tests prove correctness
- **Easy to understand** - Grammar is more readable than parser code

### For Releases
- **Version bumps guided by grammar** - Breaking vs non-breaking changes clear
- **Changelog references grammar** - "Added support for X (see grammar rule Y)"
- **Migration guides easier** - Grammar diff shows what changed

## References

- [Property-Based Testing](https://hypothesis.works/articles/what-is-property-based-testing/)
- [Grammar-Based Fuzzing](https://www.fuzzingbook.org/html/Grammars.html)
- [Test Generation from Grammars](https://cseweb.ucsd.edu/~dstefan/papers/grammar-testing.pdf)
- [Crafting Interpreters](https://craftinginterpreters.com/) - Grammar-driven development

## Examples

### Before Grammar-as-Verification

```go
// parser_test.go - Manual tests (incomplete)
func TestParseObject(t *testing.T) {
    tests := []struct{
        input string
        valid bool
    }{
        {`{}`, true},
        {`{"id": UUID}`, true},
        // Missing: nested objects, multiple properties, edge cases
    }
}
```

**Problems:**
- Incomplete coverage (what about 3+ properties? nested objects?)
- No grammar documentation
- No way to detect breaking changes
- Documentation could be wrong

### After Grammar-as-Verification

```ebnf
// docs/grammar/{format}.ebnf
ObjectNode = "{" [ Property { "," Property } ] "}" ;
Property   = StringLiteral ":" Value ;
Value      = Literal | Type | Function | Object | Array ;
```

```go
// grammar_test.go - Auto-generated
func TestGrammar(t *testing.T) {
    grammar := grammar.ParseEBNF("docs/grammar/{format}.ebnf")
    tests := grammar.GenerateTests()

    // Tests automatically include:
    // - {} (empty)
    // - {"a": UUID} (one property)
    // - {"a": String, "b": Int} (two properties)
    // - {"a": String, "b": Int, "c": Bool} (three properties)
    // - {"obj": {}} (nested object)
    // - Invalid: {, { "a", { "a": }, etc.

    for _, test := range tests {
        result, err := parser.Parse(test.Input)
        // ... verify
    }
}
```

**Benefits:**
- ✅ Comprehensive coverage (all grammar paths)
- ✅ Grammar is executable documentation
- ✅ Breaking changes detected (grammar diff)
- ✅ Documentation cannot drift

---

**Summary:** Grammar-as-verification combines the best of both worlds: hand-coded parsers for performance and error quality, with grammar-based testing for provable correctness and enforced documentation.
