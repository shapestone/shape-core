# Changelog

All notable changes to the Shape project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2025-12-09 - Initial Public Release

### Overview

Shape v0.9.0 is the first public Apache 2.0 release of Shape as a parser infrastructure library. Shape provides foundational components for parsing structured data and building custom domain-specific languages (DSLs), designed to be reusable infrastructure that other projects build upon.

**Shape is infrastructure only** - actual parser implementations are maintained in separate projects (see [ECOSYSTEM.md](ECOSYSTEM.md) for the complete list).

### Added

**AST Framework (`pkg/ast/`):**
- Unified type-safe AST node definitions for validation schemas
- Visitor pattern for traversing and manipulating ASTs
- Position tracking (line/column numbers) for all nodes
- JSON serialization for AST import/export
- Node types: ObjectNode, ArrayNode, LiteralNode, FunctionNode, ReferenceNode

**Tokenizer API (`pkg/tokenizer/`):**
- Reusable tokenization framework for building custom parsers
- Stream interface for character processing with UTF-8 support
- In-memory streams via `NewStream(string)` for standard parsing
- Buffered streams via `NewStreamFromReader(io.Reader)` for large file support
- Constant memory parsing (~20KB) regardless of input size
- Sliding window buffer (64KB) with backtracking support
- Position tracking across buffer boundaries
- Composable matcher functions for token recognition
- Pattern combinators (Sequence, OneOf, Optional)

**Parser Interface (`pkg/parser/`):**
- Standard parser interface for format implementations
- Format enumeration (values defined by parser projects)
- Standardized error handling and reporting
- Allows parser projects to choose their parsing technique (LL(1), Pratt, PEG, etc.)

**Schema Validator Framework (`pkg/validator/`):**
- Type registry for custom types (UUID, Email, SSN, etc.)
- Function registry for validation functions
- Multi-error collection for comprehensive validation
- Rich error formatting: colored terminal, plain text, JSON output
- Smart hints with "did you mean" suggestions
- Semantic validation rules for schema correctness

**Grammar Verification Framework (`pkg/grammar/`):**
- EBNF parser for custom grammar variant
- Automatic test case generation from grammar rules
- Grammar rule coverage tracking during testing
- AST comparison utilities for dual parser verification
- Enforces parser correctness through grammar-as-documentation
- Supports grammar-driven development workflow

### Architecture

**Design Principles:**
- Generic, reusable infrastructure for parser implementations
- Zero dependencies beyond Go standard library (except github.com/google/uuid)
- Clean separation between infrastructure (Shape) and implementations (parser projects)
- Streaming-first design for memory-efficient parsing
- Comprehensive test coverage (>90% across all packages)

**Parser Ecosystem:**
Shape provides infrastructure for these parser projects:
- [shape-json](https://github.com/shapestone/shape-json) - JSON parser with JSONPath queries
- [shape-yaml](https://github.com/shapestone/shape-yaml) - YAML validation parser
- [shape-csv](https://github.com/shapestone/shape-csv) - CSV parsing and validation
- [shape-xml](https://github.com/shapestone/shape-xml) - XML validation parser
- [shape-props](https://github.com/shapestone/shape-props) - Java properties format

### Performance

**Streaming Parser Performance:**
- In-memory parsing (`NewStream`): O(n) time, O(n) memory
- Buffered streaming (`NewStreamFromReader`): O(n) time, O(1) memory (~20KB)
- 138MB file: 4m31s parse time, 20.65KB memory (99.985% memory reduction vs loading entire file)
- Suitable for files of any size without memory constraints

**Benchmark Results:**
- 66KB file: 69.7% memory savings (66KB → 20KB)
- 135KB file: 85.2% memory savings (135KB → 20KB)
- 138MB file: 99.985% memory savings (138.60MB → 20.65KB)

### Documentation

**Architecture Documentation:**
- [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) - System architecture overview
- [Parser Implementation Guide](docs/PARSER_IMPLEMENTATION_GUIDE.md) - Step-by-step parser development guide
- [ADR 0001](docs/architecture/decisions/0001-ast-design.md) - AST design decisions
- [ADR 0003](docs/architecture/decisions/0003-embed-tokenizer.md) - Embedded tokenizer rationale
- [ADR 0004](docs/architecture/decisions/0004-parser-strategy.md) - Parser strategy framework
- [ADR 0005](docs/architecture/decisions/0005-grammar-as-verification.md) - Grammar-as-verification approach
- [Buffered Stream Implementation](docs/architecture/buffered-stream-implementation.md) - Streaming internals

**Community Documentation:**
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) - Community standards
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [SECURITY.md](SECURITY.md) - Security policy and vulnerability reporting
- [ECOSYSTEM.md](ECOSYSTEM.md) - Related projects and parser implementations

### Testing

**Test Coverage:**
- pkg/ast: 94.2%
- pkg/tokenizer: 98.8%
- pkg/parser: 100%
- pkg/validator: 91.7%
- pkg/grammar: 85%
- **Overall: 92.7%**

**Test Suite:**
- 180+ unit tests across all packages
- 37 buffered stream tests
- 15 grammar verification tests
- Comprehensive edge case coverage
- UTF-8 boundary testing
- Error handling validation

### Dependencies

**Go Version:** 1.21 or higher

**External Dependencies:**
- `github.com/google/uuid v1.6.0` - UUID generation for stream tracking

**Standard Library Only:** All core functionality uses only Go standard library

### Migration Guide

**For New Users:**
Shape is infrastructure-only. To parse specific formats, use one of the parser projects:

```go
import "github.com/shapestone/shape-json/pkg/json"

// Parse JSON
node, err := json.Parse(`{"key": "value"}`)

// Parse with streaming for large files
file, _ := os.Open("large.json")
defer file.Close()
node, err := json.ParseReader(file)
```

**Building Custom Parsers:**
See [Parser Implementation Guide](docs/PARSER_IMPLEMENTATION_GUIDE.md) for step-by-step instructions on building your own parser using Shape infrastructure.

### Breaking Changes

**From Internal Development:**
- First public release, no breaking changes applicable
- All APIs are stable and production-ready

### Contributors

Implementation by the Shapestone team.

### Notes

**Development Timeline:**
- Core infrastructure developed: 2020-2025
- Public release preparation: December 2025
- Apache 2.0 open source release: December 9, 2025

**Production Ready:**
All features in this release are production-ready and battle-tested through the parser ecosystem projects.

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes to public API
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

---

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

Copyright © 2020-2025 Shapestone

---

[0.9.0]: https://github.com/shapestone/shape/releases/tag/v0.9.0
