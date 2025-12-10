# Tokenizer Framework

The tokenizer framework provides a flexible, composable system for lexical analysis. It was embedded from df2-go and optimized for the Shape parser library.

## Overview

The tokenizer framework consists of three main components:

1. **Stream** - Character stream with UTF-8 support and position tracking
2. **Tokenizer** - Main tokenization engine that applies matchers to produce tokens
3. **Matchers** - Composable functions that recognize and extract tokens from streams

## Core Concepts

### Stream

A `Stream` represents a character sequence with position tracking. The tokenizer framework provides two stream implementations:

#### NewStream - In-Memory Stream

For small strings that fit entirely in memory:

```go
stream := tokenizer.NewStream("Hello World")

// Peek at next character without advancing
ch, ok := stream.PeekChar()

// Read and advance
ch, ok = stream.NextChar()

// Position tracking
offset := stream.GetOffset()  // byte offset
row := stream.GetRow()        // line number (1-indexed)
column := stream.GetColumn()  // column number (1-indexed)
```

#### NewStreamFromReader - Buffered Stream

For large files and streaming data with constant memory usage:

```go
file, _ := os.Open("large_file.json")
defer file.Close()

// Create buffered stream with 64KB sliding window
stream := tokenizer.NewStreamFromReader(file)

// Process file with constant memory (only 64KB buffer + overhead)
for !stream.IsEos() {
    ch, ok := stream.NextChar()
    // Process character
}
```

**Buffered Stream Features:**

- **Constant memory usage**: Maintains a 64KB sliding window regardless of file size
- **Streaming support**: Works with any `io.Reader` (files, network streams, etc.)
- **Backtracking**: Supports `Clone()` and `Match()` within the buffer window
- **UTF-8 handling**: Properly decodes multi-byte UTF-8 characters
- **Position tracking**: Accurate row/column tracking across buffer boundaries
- **Performance**: Tested with 100MB+ files using only ~3MB of memory

**When to use each implementation:**

- Use `NewStream()` for:
  - Small input strings (< 1MB)
  - When you need the entire input in memory
  - Simple parsing tasks

- Use `NewStreamFromReader()` for:
  - Large files (> 1MB)
  - Streaming data sources
  - Memory-constrained environments
  - Processing logs, data files, or large JSON/XML documents

**Limitations:**

- Backtracking is limited to the 64KB buffer window
- `Reset()` only works with seekable readers (files, bytes.Reader, strings.Reader)

### Token

A `Token` represents a recognized lexical element:

```go
type Token struct {
    kind   string    // token type (e.g., "Identifier", "Number")
    value  []rune    // token value
    offset int       // byte offset in source
    row    int       // line number
    column int       // column number
}
```

### Matcher

A `Matcher` is a function that attempts to recognize a token from a stream:

```go
type Matcher func(stream Stream) *Token
```

## Pattern Matching API

The framework provides powerful pattern combinators for building matchers:

### Basic Patterns

```go
// Match a single character
pattern := CharMatcher('a')

// Match a string literal
pattern := StringMatcher("hello")
```

### Pattern Combinators

```go
// Sequence - all patterns must match in order
pattern := Sequence(
    StringMatcher("function"),
    CharMatcher(' '),
    StringMatcher("name"),
)

// OneOf - match first successful pattern
pattern := OneOf(
    StringMatcher("true"),
    StringMatcher("false"),
    StringMatcher("null"),
)

// Optional - pattern matches if possible, but always succeeds
pattern := Sequence(
    StringMatcher("var"),
    Optional(CharMatcher('?')),
)
```

## Building Custom Matchers

### Example: Number Matcher

```go
func numberMatcher(stream Stream) *Token {
    var value []rune

    for {
        r, ok := stream.NextChar()
        if !ok {
            break
        }

        if r >= '0' && r <= '9' {
            value = append(value, r)
            continue
        }

        break
    }

    if len(value) == 0 {
        return nil
    }

    return NewToken("Number", value)
}
```

### Example: Identifier Matcher

```go
func identifierMatcher(stream Stream) *Token {
    var value []rune

    // First character must be letter or underscore
    r, ok := stream.NextChar()
    if !ok || !(unicode.IsLetter(r) || r == '_') {
        return nil
    }
    value = append(value, r)

    // Subsequent characters can be letters, digits, or underscores
    for {
        r, ok := stream.NextChar()
        if !ok {
            break
        }

        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            value = append(value, r)
            continue
        }

        break
    }

    return NewToken("Identifier", value)
}
```

## Using the Tokenizer

### Basic Usage

```go
// Create tokenizer with custom matchers
tokenizer := NewTokenizer(
    identifierMatcher,
    numberMatcher,
    StringMatcherFunc("Equals", "="),
    StringMatcherFunc("Plus", "+"),
)

// Initialize with input
tokenizer.Initialize("x = 42 + 10")

// Tokenize all
tokens, eos := tokenizer.Tokenize()
for _, token := range tokens {
    fmt.Printf("%s at line %d, column %d\n",
        token.Kind(), token.Row(), token.Column())
}

// Or process tokens one at a time
for {
    token, ok := tokenizer.NextToken()
    if !ok {
        break
    }

    // Process token
    fmt.Printf("Token: %s = %q\n", token.Kind(), token.ValueString())
}
```

### Backtracking with Mark/Rewind

```go
tokenizer := NewTokenizer(matchers...)
tokenizer.Initialize(input)

// Mark current position
tokenizer.Mark()

// Try to parse something
token1, ok := tokenizer.NextToken()
if !ok || token1.Kind() != "Expected" {
    // Rewind to marked position
    tokenizer.Rewind()
    // Try alternative parsing
}
```

## Built-in Utilities

### Whitespace Handling

The tokenizer automatically prepends a `WhiteSpaceMatcher` that consumes whitespace:

```go
// Whitespace tokens are automatically generated
tokenizer := NewTokenizer(yourMatchers...)
```

### Token Factory Functions

```go
// Character matcher
CharMatcherFunc("LParen", '(')

// String matcher
StringMatcherFunc("If", "if")
```

### Text Utilities

```go
// Compare rune slices
match := RunesMatch(runes1, runes2)

// String diff for testing
diff, ok := Diff(expected, actual)

// Multiline string support
text := StripMargin(`
    |line 1
    |line 2
`)
```

## Integration with Shape Parsers

Format-specific parsers use the tokenizer framework:

```go
// Example: JSONV parser uses custom matchers
func createJSONVTokenizer() Tokenizer {
    return NewTokenizer(
        // Type identifiers (UUID, Email, etc.)
        typeIdentifierMatcher,

        // Function calls: Integer(1,100)
        functionMatcher,

        // Built-in matchers
        StringMatcherFunc("LBrace", "{"),
        StringMatcherFunc("RBrace", "}"),
        StringMatcherFunc("Colon", ":"),
        // ...
    )
}
```

## Performance Characteristics

### In-Memory Stream (`NewStream`)

- **Stream operations**: O(1) for NextChar, PeekChar
- **Pattern matching**: O(n) where n is pattern length
- **Backtracking**: Uses stream cloning (copy-on-write via shared data)
- **Memory**: Entire input loaded into memory as rune slice

### Buffered Stream (`NewStreamFromReader`)

- **Stream operations**: O(1) amortized (may trigger buffer refill)
- **Pattern matching**: O(n) where n is pattern length
- **Backtracking**: Supported within 64KB buffer window
- **Memory**: Constant (~3MB overhead for 100MB+ files)
  - Buffer: 64KB of runes (~256KB)
  - Read chunks: 8KB bytes
  - Overhead: Position tracking, clone tracking

**Benchmark Results:**

Processing a 100MB file:
- Memory usage: ~3MB (constant regardless of file size)
- Throughput: ~700MB/s on Apple M1 Max
- Lines processed: ~1.4M lines in 3.7 seconds

### Memory Efficiency Comparison

| File Size | NewStream Memory | NewStreamFromReader Memory |
|-----------|-----------------|---------------------------|
| 1 MB      | ~4 MB          | ~3 MB                     |
| 10 MB     | ~40 MB         | ~3 MB                     |
| 100 MB    | ~400 MB        | ~3 MB                     |
| 1 GB      | ~4 GB          | ~3 MB                     |

For files larger than 1MB, `NewStreamFromReader` provides significant memory savings.

## Testing

The framework includes comprehensive tests:

```bash
# Run tests
go test ./internal/tokenizer/...

# Check coverage
go test -cover ./internal/tokenizer/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/tokenizer/...
go tool cover -html=coverage.out
```

## Migration from df2-go

This tokenizer was embedded from df2-go with the following changes:

- **Package consolidation**: Merged `streams`, `tokens`, `text`, `numbers` into single package
- **API refinement**: Simplified interfaces for validation schema use case
- **Position tracking**: Enhanced with Position type
- **Documentation**: Added comprehensive godoc and examples

For the original implementation, see: [df2-go repository](https://github.com/shapestone/df2-go)

## Examples

See the test files for comprehensive examples:
- `stream_test.go` - Stream and pattern matching examples
- `tokens_test.go` - Tokenizer usage examples
- `matchers_test.go` - Custom matcher examples
