# Changelog

All notable changes to the Shape project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.10.0] - 2026-10-31

### Added
- `pkg/data`: `Guard.Grow(oldCap, newCap, elemBytes)` checks the transient cost of growing a backing array (the current estimate plus the whole new array, while the old one is still alive) against `MaxMemory` and credits it back, so `MemoryUsed` is unchanged; `Guard.PeakMemory` reports the highest estimate reached; `Map.Cap`. A parser that calls `Grow` before each reallocating append peaked at 0.9x (list of small integers) and 1.1x (large object) `MaxMemory` instead of 2.1x and 1.8x. The reference JSON parser in `pkg/hostiletest` uses it, with tests that measure the peak
- `pkg/data`: shared value model for format parsers: `Kind`, `Value`, `List` and an insertion-ordered `Map` with primitives for per-format duplicate-key policy (#10)
- `pkg/data`: `Number` keeps the exact source text, validates the JSON number grammar, and offers `Int64` (errors on overflow), `Float64` and `Big` (#10)
- `pkg/data`: `Limits` with secure defaults (depth 10,000 and YAML alias expansion on, all other limits opt-in, `NoLimit` to disable) (#11)
- `pkg/data`: `Guard` enforcing depth, nodes, bytes, string/key size, keys per map, alias expansion, estimated memory and context cancellation (#12)
- `pkg/data`: typed errors `SyntaxError` and `LimitError` with `ErrSyntax` and `ErrLimit` sentinels; context errors pass through unwrapped (#13)
- `pkg/data`: `Guard.CheckStringLen` (non-charging length check for streaming parsers), `Guard.Number`, `ErrNegativeAmount`, `Map.At`, `Value.Equal`, `Value.String`, and `Number.BigWithMaxDigits`/`BigIntWithMaxDigits` with `DefaultMaxBigDigits`, and `NewMapTrusted` (uncapped preallocation)
- `pkg/hostiletest`: shared hostile-input corpus and test harness for the format modules (#15). Cases are per format, built by generators (no large files), and state the expected verdict (valid, syntax error, or a named limit) under the Default, Wire (64 MiB, 1 GiB, depth 10,000, context) and Tight (every limit low, so each limit name can be provoked by a small input) profiles. Covers deep nesting, 100-digit integers, huge exponents, alias bombs, huge strings and fields, many keys, wide CSV rows, byte and memory budgets, BOM, invalid UTF-8, duplicate keys, and XML entity expansion and DOCTYPE
- `pkg/hostiletest`: `Run` runs a format's `ParseFunc` (one `Request` with the profile, `Limits` and a `Policy` such as `RejectDuplicateKeys`) over the corpus as subtests, with an optional differential comparison against a reference (`WithReference`, used for `encoding/json` in tests only). A panicking or hanging parser fails its case instead of the test binary
- `pkg/hostiletest`: cases that may crash the process run in a re-executed test binary that stops itself at a heap cap. Cases expected to fail until a format adopts `pkg/data` carry `Known` reasons: skipped while they fail with a wrong verdict, resource overrun or crash, reported once they pass, and dropped per limit (`Adopted`, as soon as any limit a case allows is adopted), per profile (`WithoutKnownFor`), per case (`WithoutKnown`) or all at once (`Strict`). Memory is checked as a peak (3 times `MaxMemory`, from a collected baseline) and as live heap after a collection. `-short`, `HOSTILETEST_STRICT`, `HOSTILETEST_SKIPKNOWN` and `HOSTILETEST_MAXSIZE` keep CI fast
- Documentation: renamed `shape-props` references to `shape-properties`

### Changed
- Go 1.27: `go.mod` now declares `go 1.27` (was `go 1.23` in v0.9.3), and the CI, CodeQL, benchmark and release workflows, the `custom-dsl` example module and the docs (README, CONTRIBUTING, local setup, Go version policy) name Go 1.27. Tests, `go vet` and golangci-lint pass with Go 1.27.3; no code changes were needed
- `pkg/data`: documented that the peak heap during a parse can reach about 2 to 2.5 times `MaxMemory` without `Guard.Grow` and about 1 to 1.1 times with it (see Added) (append growth plus uncollected garbage), so a client with a hard memory ceiling should set `MaxMemory` to at most a third of it and that the result estimate is approximate (81 bytes charged where a small integer really takes about 56), replacing the "under-counts by 10-30%" note (`Limits.MaxMemory` and `Guard` docs only; no code change)
- `pkg/tokenizer`: stream identity check in `Match` now uses a stdlib pointer token instead of `github.com/google/uuid`; behaviour, including the "trying to match two different streams" panic, is unchanged (#14)
- `pkg/data`: `Limits` integer fields are all `int64`; `Guard` counters saturate and reject negative amounts instead of wrapping or crediting the budget; `Enter` leaves the depth unchanged on a context error; `MapSize` and `Memory` poll the context; `Map` read methods are nil-safe, `Map.Entries` returns a copy and `NewMap` clamps and caps its capacity; `Number.Big`/`BigInt` cap mantissa digits at 4096 and report size refusals as `ErrNumberRange`; number errors quote at most 32 bytes; `Guard.String`/`Key` charge memory atomically; `Big` ignores leading zeros when counting digits but bounds the effective exponent; `Value.Equal` no longer copies map entries
- CI: migrated `.golangci.yml` to golangci-lint v2 format
- CI: upgraded `golangci-lint-action` to v9 for Go 1.25 support
- CI: updated Go version to 1.25 in test and lint jobs
- CI: allowed `golangci-lint-action@v9` in dependency review (license not yet indexed)

### Removed
- Dependency on `github.com/google/uuid`; the module is now standard-library only at runtime (#14)

### Fixed
- Removed local `replace` directive in `custom-dsl` example, pinned to v0.9.3
- Suppressed pre-existing lint issues after golangci-lint v2 migration
- Removed linters merged into staticcheck in golangci-lint v2
- Moved `gofmt` and `goimports` to `formatters` section in golangci-lint v2 config
- Corrected `exclusions.rules` placement in golangci-lint v2 config
- Moved exclusions to top-level in golangci-lint v2 config
- Fixed invalid keys in golangci-lint v2 config schema
- Updated golangci-lint Go version to 1.25 and excluded flaky link checker URL

### Documentation
- Repository marketing improvements for discoverability

---

## [0.9.3] - 2025-12-26

### Added
- **`ArrayDataNode`** (`pkg/ast/arraydata.go`): new AST node type that properly distinguishes arrays from objects, resolving an invariant where empty arrays were indistinguishable from empty objects
- **`ByteStream` interface** (`pkg/tokenizer/stream.go`): byte-level tokenization API extending `Stream` with `PeekByte`, `NextByte`, `SkipWhitespace`, `FindByte`, `FindAny`, `SliceFrom`, and `RemainingBytes` for 2–8× speedup on ASCII-heavy formats
- **SWAR primitives** (`pkg/tokenizer/swar.go`): SIMD Within A Register techniques (pure Go, no assembly) processing 8 bytes in parallel — `FindByte`, `SkipWhitespace`, `NeedsEscaping`, `FindEscapeOrQuote`, `FindAnyByte`
- **AST node pooling** (`pkg/ast`): `sync.Pool`-backed `LiteralNode`, `ObjectNode`, and `ArrayDataNode` pools with opt-in `Release*` functions, reducing allocations by 3–6% when used
- **`NewTokenizerWithoutWhitespace()`** constructor for future parser optimizations
- **ASCII fast path** in stream cursor synchronization via `isASCIIOnly` flag
- **SWAR acceleration** added to `WhiteSpaceMatcher` for 8-byte parallel whitespace scanning
- `docs/AST_CONVENTIONS.md`: comprehensive AST usage conventions reference
- Go version policy documentation (`GO_VERSION.md`, `docs/`)
- Dual-path performance architecture section added to parser guide
- Comprehensive test coverage: `pkg/ast` 69.9% → 83.8%, `pkg/tokenizer` 79.6% → 86.2%

### Changed
- `Visitor` interface extended with `VisitArrayData()` method (implement in all visitors)
- `pkg/ast/serialization.go`: AST marshal/unmarshal updated to handle `ArrayDataNode`
- `pkg/ast/types.go`: added `NodeTypeArrayData` constant
- CI: bumped `softprops/action-gh-release` from 1 to 2
- CI: bumped `ossf/scorecard-action` from 2.4.0 to 2.4.3
- CI: bumped `actions/setup-go` from 5 to 6
- CI: bumped `actions/upload-artifact` from 4 to 6
- CI: bumped `codecov/codecov-action` from 4 to 5
- CI: replaced deprecated `deny-licenses` with `allow-licenses` in dependency review

### Fixed
- Corrected codecov repository slug in CI workflow
- Removed broken FAQ.md link from `GO_VERSION.md`
- Fixed contact email and standardized to Apache 2.0 `LICENSE` file
- Fixed 27+ broken documentation links (removed references to non-existent repos)
- Added `VisitArrayData()` to both `Validator` and `SchemaValidator` to satisfy the updated `Visitor` interface
- Applied `gofmt` formatting across tokenizer and test files

---

## [0.9.2] - 2025-12-12

### Fixed
- **CRITICAL**: `PARSER_IMPLEMENTATION_GUIDE.md` contained incorrect AST usage that blocked parser development
  - Guide incorrectly showed `ast.NewArrayNode(elements, startPos)` for data values (API does not exist)
  - Corrected to show parsers returning Go types (`[]interface{}`, `map[string]interface{}`)
  - Added "Critical Understanding: AST vs Data" section clarifying Shape's AST is for validation schemas only
  - Complete guide rewrite with working, correct examples
- Added clarification to ADR-0005 distinguishing data parsers from schema parsers

---

## [0.9.1] - 2025-12-12

### Security
- Updated `lychee-action` from v1 to v2.0.2 to fix **CVE-2024-48908** (arbitrary code injection vulnerability in link-checker action)

### Fixed
- Fixed failing CI workflows after initial public release
- Completely removed broken `gh-pages` benchmark storage step from workflow
- Fixed broken documentation links throughout the repository
- Simplified benchmark workflow

### Documentation
- Fixed formatting in grammar-as-verification ADR (ADR-0005)

---

## [0.9.0] - 2025-12-11

Initial public Apache 2.0 open source release of Shape as a parser infrastructure library.

Shape provides foundational, reusable components for building parsers and custom DSLs. **Shape is infrastructure only** — actual format parsers live in separate projects.

### Added

**AST Framework (`pkg/ast/`):**
- Unified, type-safe AST node definitions for validation schemas
- Visitor pattern for AST traversal and transformation
- Position tracking (line/column) for all AST nodes
- JSON serialization/deserialization for AST import/export
- Node types: `ObjectNode`, `ArrayNode`, `LiteralNode`, `FunctionNode`, `ReferenceNode`
- AST printer for human-readable output

**Tokenizer API (`pkg/tokenizer/`):**
- Reusable tokenization framework for building custom parsers
- `Stream` interface with UTF-8 character processing
- In-memory streams via `NewStream(string)` for standard parsing
- Buffered streams via `NewStreamFromReader(io.Reader)` for large file support (~20KB constant memory)
- Sliding window buffer (64KB) with backtracking support
- Position tracking across buffer boundaries
- Composable matcher functions and pattern combinators (Sequence, OneOf, Optional)
- Number parsing utilities

**Parser Interface (`pkg/parser/`):**
- Standard `Parser` interface for format implementations
- Format enumeration and standardized error handling
- Supports LL(1), Pratt, PEG, and other parsing techniques

**Schema Validator Framework (`pkg/validator/`):**
- Type registry for custom types (UUID, Email, SSN, etc.)
- Function registry for validation functions
- Multi-error collection for comprehensive validation reporting
- Rich error formatting: colored terminal output, plain text, JSON
- Smart "did you mean" hints for type and function names
- Semantic validation rules for schema correctness

**Grammar Verification Framework (`pkg/grammar/`):**
- EBNF parser for a custom grammar variant
- Automatic test case generation from grammar rules
- Grammar rule coverage tracking during testing
- AST comparison utilities for dual-parser verification
- Grammar-as-documentation enforced correctness

**CI/CD and Project Infrastructure:**
- Full GitHub Actions CI pipeline (test, lint, benchmark, security)
- CodeQL security analysis
- OSSF Scorecard integration
- Dependabot configuration
- Release automation with `scripts/release.sh`
- PR labeler, issue templates, and pull request template

### Performance
- In-memory parsing: O(n) time, O(n) memory
- Buffered streaming: O(n) time, O(1) memory (~20KB)
- 138MB file: 4m31s parse time, 20.65KB memory (99.985% memory reduction vs full load)

### Documentation
- `docs/architecture/ARCHITECTURE.md` — system architecture overview
- `docs/PARSER_IMPLEMENTATION_GUIDE.md` — step-by-step parser development guide
- ADR-0001: AST design decisions
- ADR-0003: Embedded tokenizer rationale
- ADR-0004: Parser strategy framework
- ADR-0005: Grammar-as-verification approach
- `docs/architecture/buffered-stream-implementation.md` — streaming internals
- `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `SECURITY.md`, `ECOSYSTEM.md`
- `docs/contributor/` — branching workflow, local setup, testing guide
- `docs/validation/` — error codes reference
- `examples/custom-dsl/` — working custom DSL example

### Test Coverage at Release
- `pkg/ast`: 94.2%
- `pkg/tokenizer`: 98.8%
- `pkg/parser`: 100%
- `pkg/validator`: 91.7%
- `pkg/grammar`: 85%
- **Overall: 92.7%**

### Dependencies
- Go 1.21 or higher
- `github.com/google/uuid v1.6.0` (UUID generation for stream tracking)
- All other functionality uses only the Go standard library

---

## [0.6.0] - 2025-10-27

> Pre-public release. This version represents the cumulative internal development history
> of the Shape core library before the Apache 2.0 public release.

### Added
- **JSONPath query engine** with full filter expression support (`$..store.book[?(@.price < 10)]`)
- **Data format parsers**: JSON, YAML, CSV native parsers
- **Public Tokenizer API**: exposed tokenizer as a stable public API for external use
- **Semantic schema validation** with multi-error collection (#11)
- **Custom validator registration**: extensible type and function registries
- **AST string interning optimization**: reduced memory usage for repeated string values
- **Schema validation** with built-in type and function rules
- **Native YAMLV parser**: replaced `gopkg.in/yaml.v3` dependency with a custom native parser (#8)
- **Complete format detection** for all 6 supported formats (JSONV, XMLV, YAMLV, CSVV, TEXTV, PROPSV)
- **YAMLV and TEXTV parsers** (Phase 4)
- **CSVV parser** and cross-format integration tests (Phase 3)
- **PropsV and XMLV parsers** (Phase 3)
- **JSONV parser** — first working format parser (Phase 2)
- **Foundation**: AST model and tokenization framework (Phase 1)
- **Release automation** infrastructure (`scripts/`, CI release workflow)
- **Branching workflow** documentation

### Changed
- Architecture updated to embed the tokenizer framework directly
- YAMLV test coverage improved to 95.9%

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes to the public API
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes and documentation, backward compatible

---

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.

Copyright © 2020–2025 Shapestone

---

[Unreleased]: https://github.com/shapestone/shape-core/compare/v0.10.0...HEAD
[0.10.0]: https://github.com/shapestone/shape-core/compare/v0.9.3...v0.10.0
[0.9.3]: https://github.com/shapestone/shape-core/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/shapestone/shape-core/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/shapestone/shape-core/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/shapestone/shape-core/compare/v0.6.0...v0.9.0
[0.6.0]: https://github.com/shapestone/shape-core/releases/tag/v0.6.0
