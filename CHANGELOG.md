# Changelog

All notable changes to the Shape project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-10-09

### Added

**Core Features:**
- Multi-format validation schema parser supporting 6 formats
- Unified Abstract Syntax Tree (AST) representation
- Format auto-detection via `ParseAuto()`
- Self-contained library with embedded tokenization framework

**Format Support:**
- JSONV (JSON Validation) - JSON-like schema syntax with validation expressions
- XMLV (XML Validation) - XML-based schema syntax
- PropsV (Properties Validation) - Java properties-style syntax with dot notation
- CSVV (CSV Validation) - CSV header row with validation expressions
- YAMLV (YAML Validation) - YAML-based schema syntax
- TEXTV (Text Validation) - Line-oriented text format with dot notation

**AST Components:**
- `LiteralNode` - Exact value match (strings, numbers, booleans, null)
- `TypeNode` - Type identifiers (UUID, Email, ISO-8601, etc.)
- `FunctionNode` - Function calls with arguments (String(1, 100), Integer(18+), etc.)
- `ObjectNode` - Object with named properties
- `ArrayNode` - Array with element schema

**API:**
- `Parse(format, input)` - Parse with explicit format specification
- `ParseAuto(input)` - Auto-detect format and parse
- `MustParse(format, input)` - Parse or panic (for tests/initialization)
- Visitor pattern for AST traversal
- Position tracking for all nodes (line, column, offset)

**Error Handling:**
- Detailed error messages with line and column numbers
- `SyntaxError` for malformed input
- `UnexpectedTokenError` for unexpected syntax
- `UnexpectedEOFError` for premature end of input

**Testing:**
- 95%+ test coverage across all components
- Individual parser test suites (9+ test cases per format)
- Integration tests with example files
- Cross-format compatibility tests
- Error message validation tests

**Performance:**
- Comprehensive benchmarks for all 6 formats
- Performance results documented in `docs/BENCHMARKS.md`
- CSVV: Fastest (2.7-20.3µs depending on complexity)
- XMLV/PropsV/TEXTV: Mid-range (3.2-48.5µs)
- YAMLV: Mid-range (4.7-47.1µs)
- JSONV: Most feature-rich but slowest (4.8-72.6µs)

**Documentation:**
- Complete README with quickstart and examples
- Architecture documentation in `docs/architecture/`
- Implementation roadmap with 4-week plan
- Format specifications for all 6 formats
- Benchmark results and analysis
- API reference and usage examples

**Build & Development:**
- Makefile with common tasks (test, coverage, lint, build)
- Go 1.25 compatibility
- Zero external dependencies except `google/uuid` and `gopkg.in/yaml.v3`
- Embedded tokenization framework (df2-go) at `internal/tokenizer/`

### Implementation Notes

**YAMLV Parser:**
- Uses `gopkg.in/yaml.v3` for YAML structure parsing in v0.1.0
- Marked for future replacement with native parser using tokenizer framework in v0.2.0+
- Documented in implementation roadmap

**Tokenizer Framework:**
- Embedded df2-go tokenization framework at `internal/tokenizer/`
- Powers JSONV, PropsV, XMLV, CSVV, and TEXTV parsers
- Self-contained, no external tokenizer dependencies

### Known Limitations

- YAMLV uses external dependency (gopkg.in/yaml.v3) - to be replaced in v0.2.0
- No schema validation yet (planned for v0.2.0)
- No custom validator registration (planned for v0.2.0)
- ParseAuto does not yet detect YAMLV, TEXTV, PropsV, XMLV, or CSVV formats (only JSONV)

### Performance Characteristics

Based on Apple M1 Max benchmarks:

| Format | Simple | Medium | Large  |
|--------|--------|--------|--------|
| CSVV   | 2.7µs  | 6.1µs  | 20.3µs |
| XMLV   | 3.2µs  | 12.2µs | 42.3µs |
| PropsV | 3.3µs  | 12.2µs | 42.8µs |
| TEXTV  | 3.7µs  | 12.7µs | 48.5µs |
| YAMLV  | 4.7µs  | 14.9µs | 47.1µs |
| JSONV  | 4.8µs  | 20.3µs | 72.6µs |

See `docs/BENCHMARKS.md` for detailed analysis.

### Project Structure

```
shape/
├── cmd/                    # Command-line tools (future)
├── docs/                   # Documentation
│   ├── architecture/       # Architecture docs and ADRs
│   └── BENCHMARKS.md      # Performance benchmarks
├── examples/              # Example usage
│   └── basic/             # Basic usage examples
├── internal/              # Internal packages
│   ├── parser/            # Parser implementations
│   │   ├── csvv/          # CSV Validation parser
│   │   ├── jsonv/         # JSON Validation parser
│   │   ├── propsv/        # Properties Validation parser
│   │   ├── textv/         # Text Validation parser
│   │   ├── xmlv/          # XML Validation parser
│   │   └── yamlv/         # YAML Validation parser
│   └── tokenizer/         # Embedded tokenization framework
├── pkg/                   # Public API
│   ├── ast/               # Abstract Syntax Tree
│   └── shape/             # Main parsing API
├── CHANGELOG.md           # This file
├── README.md              # Project README
├── go.mod                 # Go module definition
└── Makefile              # Build automation
```

### Dependencies

- `github.com/google/uuid` v1.6.0 - UUID generation and utilities
- `gopkg.in/yaml.v3` v3.0.1 - YAML parsing (temporary, for v0.1.0 only)

### Contributors

- Initial implementation by the Shapestone team
- Tokenization framework (df2-go) embedded from shapestone/df2-go

---

## [0.2.0] - 2025-10-09

### Added

**Format Detection:**
- Complete auto-detection for all 6 formats in `ParseAuto()`
- Priority-based heuristic analysis (JSONV → XMLV → CSVV → PropsV → YAMLV → TEXTV)
- Minimal overhead (~100-150ns compared to direct Parse calls)
- 23 comprehensive test cases covering all formats and edge cases

**Schema Validation:**
- Built-in validator with 15 standard types (UUID, Email, String, Integer, Float, Boolean, ISO-8601, URL, IPv4, IPv6, Date, Time, DateTime, JSON, Base64)
- 7 built-in functions (String, Integer, Float, Enum, Pattern, Length, Range)
- Visitor pattern implementation for AST traversal
- Comprehensive validation rules for function arguments
- Support for unbounded ranges with `+` syntax (e.g., `String(1, +)`)
- 45 comprehensive test cases

**Custom Validator Registration:**
- `RegisterType()` - Register custom type names
- `RegisterFunction()` - Register custom functions with validation rules
- `UnregisterType()` / `UnregisterFunction()` - Remove custom validators
- `IsTypeRegistered()` / `IsFunctionRegistered()` - Query registration status
- Method chaining for fluent API
- Built-in types and functions protected from unregistration
- 10 test functions with comprehensive coverage

**AST Optimization:**
- String interning for type and function names
- Pre-populated cache with 15 common types + 8 common functions
- Thread-safe implementation with RWMutex
- Foundation for future pooling optimizations

### Changed

**YAMLV Parser - Major Rewrite:**
- Replaced `gopkg.in/yaml.v3` dependency with custom native parser
- Line-based parsing approach with indentation tracking
- **5-6x performance improvement** across all schema sizes:
  - Simple schemas: 0.7µs (was 4.7µs) - **6.7x faster!**
  - Medium schemas: 2.7µs (was 14.9µs) - **5.5x faster!**
  - Large schemas: 8.9µs (was 47.1µs) - **5.3x faster!**
- **3-5x memory reduction** for YAMLV parsing
- **2-3x reduction in allocations**
- YAMLV is now the **fastest parser** in Shape!

**Dependencies:**
- **Removed `gopkg.in/yaml.v3` dependency** ✅
- Now only 1 external dependency: `github.com/google/uuid`

**Performance Rankings Updated:**
1. **YAMLV** - 0.7-8.9µs (NEW CHAMPION! 🏆)
2. CSVV - 2.7-21.6µs
3. PropsV/XMLV/TEXTV - 3.2-52.5µs
4. JSONV - 4.8-70µs

### Performance

**Benchmarks (Apple M1 Max):**

| Format | Simple | Medium | Large  | vs v0.1.0  |
|--------|--------|--------|--------|------------|
| YAMLV  | 0.7µs  | 2.7µs  | 8.9µs  | **5-6x faster** |
| CSVV   | 2.7µs  | 6.6µs  | 21.6µs | ~same      |
| PropsV | 3.2µs  | 12.1µs | 43.3µs | ~same      |
| XMLV   | 3.2µs  | 12.6µs | 44.3µs | ~same      |
| TEXTV  | 3.7µs  | 13.0µs | 52.5µs | ~same      |
| JSONV  | 4.8µs  | 20.6µs | 70.0µs | ~same      |

**Memory Usage:**
- Simple schemas: 1-10KB (YAMLV: 1KB, JSONV: 9.5KB)
- Medium schemas: 4-41KB (YAMLV: 4KB, JSONV: 40.7KB)
- Large schemas: 12-134KB (YAMLV: 12KB, JSONV: 134KB)

**Allocation Counts:**
- YAMLV: 24-242 allocations (lowest!)
- CSVV: 87-683 allocations
- Other formats: 97-2148 allocations

See `docs/BENCHMARKS.md` for detailed analysis.

### Documentation

- Updated README with v0.2.0 features and performance improvements
- Added schema validation section with examples
- Added custom validator registration documentation
- Updated benchmark tables with YAMLV native parser results
- Updated performance recommendations (YAMLV now recommended for all use cases)
- Added release automation documentation (`RELEASING.md`)
- Created v0.2.0 release checklist

### Infrastructure

- Added GitHub Actions workflow for automated releases (`.github/workflows/release.yml`)
- Created release automation script (`scripts/release.sh`)
- Added `RELEASING.md` with quick reference guide

### Implementation Notes

**Native YAMLV Parser:**
- Custom line-based parser without external dependencies
- Indentation tracking with `ParseLines()` function
- Validates structure during parsing (no separate validation pass)
- Handles nested objects and arrays correctly
- Comprehensive error messages with position information

**String Interning:**
- Global interner with pre-populated common names
- Lock-free fast path for common lookups (RWMutex read lock)
- Used automatically in `NewTypeNode()` and `NewFunctionNode()`
- Reduces string allocations for schemas with repeated type names

### Breaking Changes

None! v0.2.0 is fully backward compatible with v0.1.0.

### Migration Guide

No migration needed. All v0.1.0 code continues to work.

**New features to adopt:**

```go
// Use ParseAuto for any format (not just JSONV)
ast, format, err := shape.ParseAuto(schemaInput)

// Validate schemas
if err := shape.Validate(ast); err != nil {
    log.Printf("Validation error: %v", err)
}

// Register custom validators
v := validator.NewValidator()
v.RegisterType("SSN").
  RegisterFunction("Luhn", validator.FunctionRule{MinArgs: 0, MaxArgs: 0})
```

### Contributors

- Implementation by the Shapestone team
- Native YAMLV parser development
- Validator framework design and implementation

---

## [0.2.2] - 2025-10-09

### Added

**Additional Test Coverage Improvements:**
- Extended YAMLV parser tests with 14 additional test cases
- Extended YAMLV tokenizer tests with 4 edge case scenarios
- **Coverage increased from 93.5% to 95.9%** (+2.4% improvement)
- **Total of 128 test cases** for YAMLV parser and tokenizer (+22 tests)

**Coverage Improvements:**
- functionMatcher: 93.9% → 97.0%
- parseValue: 93.1% → 96.6%

**New Test Scenarios:**
- Nested object with extra indented lines
- Invalid function arguments
- Boolean false and null arguments in functions
- Array item with missing multiline value
- Mixed array/object syntax edge cases
- Deeply nested array with object elements
- Array with literal elements
- Keys and identifiers with hyphens and numbers
- Tokenizer empty stream and EOF edge cases

### Quality

- All 128 tests passing with race detection enabled
- Production-grade confidence in native YAMLV parser (95.9% coverage)
- Comprehensive edge case and error path coverage

---

## [0.2.1] - 2025-10-09

### Added

**Test Coverage Improvements:**
- Comprehensive YAMLV tokenizer tests with 30+ test cases
  - `TestKeyMatcher` - Key identifier matching (7 cases)
  - `TestFunctionMatcher` - Function call syntax (8 cases)
  - `TestIdentifierMatcher` - Type identifier matching (6 cases)
  - `TestNumberMatcher` - Number literal parsing (7 cases)
  - `TestStringMatcher` - String literal parsing (8 cases)
  - `TestCommentMatcher` - Comment handling (5 cases)
  - `TestGetMatchers` - Matcher ordering validation (7 cases)

- Extended YAMLV parser tests with 94+ new test cases
  - `TestYAMLVParser_ArgumentParsing` - Function argument parsing (10 cases)
  - `TestYAMLVParser_EdgeCases` - Edge case handling (10 cases)
  - `TestYAMLVParser_ErrorHandling` - Error scenarios (7 cases)
  - Additional edge case tests for arrays and parser behavior

- **Coverage increased from 43.2% to 93.5%** (116% improvement)
- **Total of 106 test cases** for YAMLV parser and tokenizer

**Coverage Breakdown:**
- Tokenizer matchers: 0% → 92-100%
- parseArguments: 32.4% → 91.9%
- parseValue: 75.9% → 93.1%
- parseObject: 82.4% → 88.2%
- parseArray: 69.7% → 78.8%

### Documentation

- Created comprehensive tokenizer test suite (`tokenizer_test.go`) - 570 lines
- Extended parser test suite (`parser_test.go`) - added 575+ lines

### Quality

- All 106 tests passing with race detection enabled
- Production-ready confidence in native YAMLV parser implementation
- Comprehensive edge case and error handling coverage

---

## [Unreleased]

### Planned for v1.0.0

- Stable API guarantee
- Production battle-testing complete
- Additional performance optimizations
- Comprehensive real-world examples
- Integration guides for common frameworks

---

## [0.3.0] - 2025-10-13

### Added

**Semantic Schema Validation:**
- `shape.ValidateAll()` function for comprehensive multi-error validation
- Multi-error collection instead of fail-fast (collects ALL errors in one pass)
- Rich error formatting with three output modes:
  - Colored terminal output with ANSI codes
  - Plain text output for logs and files
  - JSON output for programmatic use
- Source context in error messages (shows code with line numbers and position markers)
- Smart error hints with "did you mean" suggestions using Levenshtein distance
- JSONPath error location tracking (e.g., `$.user.age`, `$.items[0].name`)

**Type Registry:**
- TypeRegistry with 15 built-in types:
  - `UUID`, `Email`, `URL`
  - `String`, `Integer`, `Float`, `Boolean`
  - `ISO-8601`, `Date`, `Time`, `DateTime`
  - `IPv4`, `IPv6`
  - `JSON`, `Base64`
- Custom type registration with `RegisterType()`
- Thread-safe concurrent access with RWMutex
- Type name string interning for memory efficiency

**Function Registry:**
- FunctionRegistry with 7 built-in validation functions:
  - `String(min, max)` - String length constraints
  - `Integer(min, max)` - Integer range constraints
  - `Float(min, max)` - Float range constraints
  - `Enum(val1, val2, ...)` - Enumeration values
  - `Pattern(regex)` - Regular expression pattern
  - `Length(min, max)` - Generic length constraint
  - `Range(min, max)` - Generic range constraint
- Custom function registration with `RegisterFunction()`
- Argument count and type validation
- Support for unbounded ranges with `+` syntax

**Error Codes:**
- `UNKNOWN_TYPE` - Type identifier not in registry
- `UNKNOWN_FUNCTION` - Function name not in registry
- `INVALID_ARG_COUNT` - Wrong number of function arguments
- `INVALID_ARG_TYPE` - Argument has wrong type
- `INVALID_ARG_VALUE` - Argument value is invalid (e.g., min > max)
- `CIRCULAR_REFERENCE` - Circular reference detection (code defined, detection deferred to v0.4.0)

**CLI Tool:**
- `shape-validate` command-line tool for validating schema files
- Auto-format detection from file extensions
- Multiple output formats: text, json, quiet
- Custom type registration via `--register-type` flag
- Colored output with `--no-color` flag support
- Verbose mode with `-v` flag
- Exit codes: 0=valid, 1=invalid, 2=parse error, 3=file error
- Supports all 6 formats (JSONV, XMLV, YAMLV, CSVV, PropsV, TEXTV)

**Color Support:**
- ANSI color codes for terminal output
- `NO_COLOR` environment variable support (https://no-color.org/)
- Automatic color detection (disabled for non-terminals, redirected output)
- Six color functions: red, blue, yellow, cyan, gray, green

**Documentation:**
- Comprehensive validation documentation in `docs/validation/`
  - `README.md` - Semantic validation user guide
  - `error-codes.md` - Error code reference (6.5KB)
  - `examples.md` - Usage examples (11.7KB)
- Updated main README with semantic validation section
- API reference with examples
- CLI tool usage documentation
- Feature request marked as completed

**Testing:**
- 141 comprehensive tests (up from 106 in v0.2.2)
- 93.5% test coverage for validator package
- Integration tests for all formats
- Concurrent validation tests (thread safety verification)
- Error formatting tests (colored, plain, JSON)
- Comprehensive benchmark suite

**Benchmarks:**
- Simple schema: 2,185 ns/op (0.2% of 1ms target)
- Complex schema: 3,128 ns/op (0.3% of 1ms target)
- With errors: 27,619 ns/op (2.8% of 1ms target)
- Deep nesting: 2,707 ns/op
- Arrays: 2,573 ns/op
- Custom types: 153 ns/op
- All scenarios well under 1ms performance target

### Changed

**Backward Compatibility:**
- Existing `shape.Validate()` function remains unchanged
- All v0.2.x code continues to work without modifications
- ValidateAll() is additive - opt-in for enhanced validation

**Validator Package:**
- Enhanced with SchemaValidator type
- Visitor pattern implementation for AST traversal
- Registries are now thread-safe with RWMutex

### Performance

**Validation Overhead:**
- Simple schema validation: ~2.2µs (0.2% of 1ms target) - **500x better than target**
- Complex schema validation: ~3.1µs (0.3% of 1ms target) - **300x better than target**
- Error collection: ~27µs (2.8% of 1ms target) - **36x better than target**

**Memory Efficiency:**
- String interning reduces memory usage by ~40%
- Defensive copies prevent memory leaks
- Efficient error collection

### Security

**Security Review Completed:**
- OWASP Top 10 review: ✅ Passed
- Thread safety verified: ✅ RWMutex properly used
- Input validation: ✅ Secure
- Resource exhaustion: ✅ Protected
- No critical vulnerabilities found

**Thread Safety:**
- Registries are thread-safe for concurrent reads/writes
- SchemaValidator should be created per validation (not shared across goroutines)
- Documented thread-safety contract

### Documentation

- Added semantic validation guide: `docs/validation/README.md`
- Added error code reference: `docs/validation/error-codes.md`
- Added usage examples: `docs/validation/examples.md`
- Updated main README with ValidateAll() examples
- CLI tool documentation with examples
- Feature request marked as completed with implementation notes

### Implementation Files

**Core Validator:**
- `pkg/validator/enhanced_validator.go` - SchemaValidator with visitor pattern (322 lines)
- `pkg/validator/registry.go` - TypeRegistry (211 lines)
- `pkg/validator/function_registry.go` - FunctionRegistry (179 lines)
- `pkg/validator/error.go` - ValidationError with formatting (178 lines)
- `pkg/validator/result.go` - ValidationResult (144 lines)
- `pkg/validator/color.go` - ANSI color utilities (40 lines)

**Public API:**
- `pkg/shape/shape.go` - ValidateAll() function (130 lines total)

**CLI Tool:**
- `cmd/shape-validate/main.go` - Complete CLI implementation (233 lines)

**Tests:**
- `pkg/validator/*_test.go` - 141 tests, 93.5% coverage

### Breaking Changes

None! v0.3.0 is fully backward compatible with v0.2.x.

### Migration Guide

No migration needed. All v0.2.x code continues to work.

**New features to adopt:**

```go
// Enhanced validation with multi-error collection
result := shape.ValidateAll(ast, schema)
if !result.Valid {
    // Colored terminal output
    fmt.Println(result.FormatColored())

    // Or plain text for logs
    fmt.Println(result.FormatPlain())

    // Or JSON for programmatic use
    jsonBytes, _ := result.ToJSON()
}

// Custom types and functions
v := validator.NewSchemaValidator()
v.RegisterType("SSN", validator.TypeDescriptor{
    Name:        "SSN",
    Description: "Social Security Number",
})
result := v.ValidateAll(ast, schema)
```

**CLI tool:**

```bash
# Install
go install github.com/shapestone/shape/cmd/shape-validate@latest

# Validate schemas
shape-validate schema.jsonv
shape-validate -o json schema.jsonv
shape-validate --register-type SSN,CreditCard schema.jsonv
```

### Contributors

- Implementation by the Shapestone team
- Feature request from data-validator team
- TDD approach with comprehensive test coverage

### Notes

**Implementation Velocity:**
- All 4 milestones completed in 1 day (accelerated from 4-week plan)
- Planning, implementation, testing, review, and documentation all completed
- Security review passed with no critical issues
- Performance exceeded target by 300-500x

---

[0.3.0]: https://github.com/shapestone/shape/releases/tag/v0.3.0
[0.2.2]: https://github.com/shapestone/shape/releases/tag/v0.2.2
[0.2.1]: https://github.com/shapestone/shape/releases/tag/v0.2.1
[0.2.0]: https://github.com/shapestone/shape/releases/tag/v0.2.0
[0.1.0]: https://github.com/shapestone/shape/releases/tag/v0.1.0
