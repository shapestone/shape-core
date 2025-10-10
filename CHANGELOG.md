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

## [Unreleased]

### Planned for v0.2.0

- Replace YAMLV yaml.v3 dependency with native parser
- Extend ParseAuto to detect all 6 formats
- Schema validation rules
- AST optimization passes
- Custom validator registration
- Additional format support (potential)

### Planned for v1.0.0

- Stable API guarantee
- Production battle-testing complete
- Performance optimizations
- Comprehensive real-world examples
- Integration guides for common frameworks

---

[0.1.0]: https://github.com/shapestone/shape/releases/tag/v0.1.0
