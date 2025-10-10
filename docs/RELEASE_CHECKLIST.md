# Release Checklist for v0.1.0

This document outlines the steps to prepare and release version 0.1.0 of the Shape library.

## Pre-Release Verification

### Code Quality

- [x] All 6 format parsers implemented and tested
  - [x] JSONV parser with comprehensive tests
  - [x] PropsV parser with comprehensive tests
  - [x] XMLV parser with comprehensive tests
  - [x] CSVV parser with comprehensive tests
  - [x] YAMLV parser with comprehensive tests
  - [x] TEXTV parser with comprehensive tests

- [x] All tests passing
  ```bash
  make test
  ```

- [x] Test coverage ≥ 95%
  ```bash
  make coverage
  ```

- [x] No lint errors
  ```bash
  make lint
  ```

- [x] All benchmarks running successfully
  ```bash
  go test -bench=. -benchmem ./pkg/shape/
  ```

### Documentation

- [x] README.md updated with:
  - [x] All 6 formats documented
  - [x] Accurate performance numbers from benchmarks
  - [x] Correct TEXTV examples
  - [x] Installation instructions
  - [x] Quick start guide
  - [x] API reference
  - [x] Examples

- [x] CHANGELOG.md created with v0.1.0 entry
  - [x] All features listed
  - [x] Implementation notes
  - [x] Known limitations
  - [x] Performance characteristics
  - [x] Dependencies documented

- [x] BENCHMARKS.md created with performance analysis

- [x] Architecture documentation up to date
  - [x] ARCHITECTURE.md
  - [x] IMPLEMENTATION_ROADMAP.md
  - [x] Format specifications
  - [x] ADRs

### Dependencies

- [x] go.mod file correct
  - [x] Module name: `github.com/shapestone/shape`
  - [x] Go version: 1.25
  - [x] Dependencies listed:
    - [x] `github.com/google/uuid v1.6.0`
    - [x] `gopkg.in/yaml.v3 v3.0.1`

- [x] go.sum file up to date
  ```bash
  go mod tidy
  go mod verify
  ```

### Code Organization

- [x] All files properly organized
  - [x] Public API in `pkg/`
  - [x] Internal implementations in `internal/`
  - [x] Documentation in `docs/`
  - [x] Examples in `examples/`

- [x] No TODO comments in production code
- [x] All exported functions have documentation
- [x] All public types have examples

## Release Process

### 1. Final Verification

```bash
# Clean build
make clean
make build

# Run all tests with race detection
make test

# Run benchmarks
go test -bench=. -benchmem ./pkg/shape/

# Check coverage
make coverage

# Run linters
make lint

# Verify dependencies
go mod tidy
go mod verify
```

### 2. Version Tagging

```bash
# Ensure on main branch and up to date
git checkout main
git pull origin main

# Verify clean working directory
git status

# Create and push tag
git tag -a v0.1.0 -m "Release v0.1.0: Multi-format validation schema parser"
git push origin v0.1.0
```

### 3. GitHub Release

Create a GitHub release at https://github.com/shapestone/shape/releases/new

**Tag:** v0.1.0
**Title:** Shape v0.1.0 - Multi-Format Validation Schema Parser

**Description:**
```markdown
# Shape v0.1.0 - Multi-Format Validation Schema Parser

First official release of Shape, a production-ready parser library that converts validation schema formats into a unified Abstract Syntax Tree (AST).

## Features

- **6 Format Support:** JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV
- **Unified AST:** All formats produce the same AST structure
- **Format Auto-Detection:** Automatically detect and parse JSONV format
- **Detailed Error Messages:** Line and column numbers for all parse errors
- **Self-Contained:** Embedded tokenization framework, minimal external dependencies
- **Production-Ready:** 95%+ test coverage, comprehensive error handling

## Installation

```bash
go get github.com/shapestone/shape@v0.1.0
```

## Quick Start

```go
import "github.com/shapestone/shape/pkg/shape"
import "github.com/shapestone/shape/internal/parser"

// Parse with explicit format
ast, err := shape.Parse(parser.FormatJSONV, `{"id": UUID}`)

// Auto-detect format
ast, format, err := shape.ParseAuto(`{"id": UUID}`)
```

## Performance

Benchmarked on Apple M1 Max:
- Simple schemas: 2.7-4.8µs
- Medium schemas: 6.1-20.3µs
- Large schemas: 20.3-72.6µs

See [BENCHMARKS.md](https://github.com/shapestone/shape/blob/main/docs/BENCHMARKS.md) for details.

## Documentation

- [README](https://github.com/shapestone/shape/blob/main/README.md)
- [Architecture](https://github.com/shapestone/shape/blob/main/docs/architecture/ARCHITECTURE.md)
- [Benchmarks](https://github.com/shapestone/shape/blob/main/docs/BENCHMARKS.md)
- [Changelog](https://github.com/shapestone/shape/blob/main/CHANGELOG.md)

## Known Limitations

- YAMLV uses external dependency (`gopkg.in/yaml.v3`) - to be replaced in v0.2.0
- ParseAuto only detects JSONV format - other formats planned for v0.2.0
- No schema validation yet - planned for v0.2.0

## What's Next (v0.2.0)

- Replace YAMLV yaml.v3 dependency with native parser
- Extend ParseAuto to detect all 6 formats
- Schema validation rules
- Custom validator registration

---

Full changelog: [CHANGELOG.md](https://github.com/shapestone/shape/blob/main/CHANGELOG.md)
```

### 4. Post-Release

- [ ] Announce release on project channels
- [ ] Update integration guide for data-validator project
- [ ] Monitor for issues and feedback
- [ ] Begin planning v0.2.0 features

## Rollback Procedure

If critical issues are found after release:

```bash
# Delete remote tag
git push --delete origin v0.1.0

# Delete local tag
git tag -d v0.1.0

# Delete GitHub release
# (Use GitHub UI to delete the release)

# Fix issues, then re-release as v0.1.1
```

## Success Criteria

Release is successful when:

- [x] All tests passing on main branch
- [x] Code coverage ≥ 95%
- [x] All 6 format parsers working correctly
- [x] Documentation complete and accurate
- [x] Benchmarks documented
- [x] Dependencies properly declared
- [ ] GitHub release published
- [ ] Tag pushed to repository

## Notes

- This is a v0.x release, so API changes are allowed before v1.0.0
- Focus is on correctness and test coverage over optimization
- YAMLV dependency on yaml.v3 is temporary and documented
- Performance benchmarks establish baseline for future optimization

---

**Prepared by:** Shapestone Team
**Date:** 2025-10-09
**Target Release Date:** 2025-10-09
