# Shape Tokenization Migration Plan

**Date:** 2025-10-09  
**Version:** 1.0  
**Status:** Planning

## Executive Summary

This document outlines the plan to migrate tokenization code from the df2-go repository into shape as an embedded tokenization layer. This migration transforms shape from a library with external tokenization dependency to a fully self-contained parser library.

## Architectural Change Overview

### Before: External Dependency
```
┌─────────────────────────────────────┐
│         data-validator              │
│    ┌─────────────┐                  │
│    │   validator │                  │
│    └──────┬──────┘                  │
└───────────┼─────────────────────────┘
            │ depends on
            ▼
┌─────────────────────────────────────┐
│            shape                    │
│    ┌─────────────┐                  │
│    │   parsers   │                  │
│    └──────┬──────┘                  │
└───────────┼─────────────────────────┘
            │ depends on
            ▼
┌─────────────────────────────────────┐
│           df2-go                    │
│    ┌─────────────┐                  │
│    │  tokenizer  │                  │
│    └──────┬──────┘                  │
└───────────┼─────────────────────────┘
            │ depends on
            ▼
     google/uuid v1.6.0
```

**Dependency Chain:**
data-validator → shape → df2-go → google/uuid

**Issues:**
- Transitive dependencies
- Version coordination required
- Multiple repositories to manage
- Unclear ownership of tokenization

### After: Embedded Tokenization
```
┌─────────────────────────────────────┐
│         data-validator              │
│    ┌─────────────┐                  │
│    │   validator │                  │
│    └──────┬──────┘                  │
└───────────┼─────────────────────────┘
            │ depends on
            ▼
┌─────────────────────────────────────┐
│            shape                    │
│                                     │
│    ┌─────────────┐                  │
│    │   parsers   │                  │
│    └──────┬──────┘                  │
│           │ uses                    │
│           ▼                          │
│    ┌─────────────┐                  │
│    │  tokenizer  │ (internal)       │
│    │  (embedded) │                  │
│    └─────────────┘                  │
└─────────────────────────────────────┘
            │ depends on
            ▼
     google/uuid v1.6.0
     gopkg.in/yaml.v3
```

**Dependency Chain:**
data-validator → shape

**Benefits:**
- Zero transitive dependencies
- Single repository
- Clear ownership
- Self-contained library

## Components to Migrate

### From df2-go Repository

**Source Location:** `/Users/michaelsundell/Projects/shapestone/df2-go/`

**Components:**

| df2-go Package | Lines of Code | Purpose | Destination |
|----------------|---------------|---------|-------------|
| `streams/` | ~400 | Stream abstraction, position tracking | `internal/tokenizer/stream.go` |
| `tokens/` | ~600 | Token struct, Tokenizer, Matchers | `internal/tokenizer/tokens.go` + `matchers.go` |
| `text/` | ~200 | Rune utilities, text helpers | `internal/tokenizer/text.go` |
| `numbers/` | ~100 | Number parsing utilities | `internal/tokenizer/numbers.go` |
| **Tests** | ~800 | Comprehensive test coverage | `internal/tokenizer/*_test.go` |

**Total:** ~2100 lines of production code + ~800 lines of tests

### What NOT to Migrate

| Component | Reason |
|-----------|--------|
| `df/json/` | Shape implements its own JSONV tokenizer |
| `df/scenario/` | Not relevant to validation schemas |
| `parsers/` | Shape has its own parser abstractions |
| `main.go` | df2-go executable, not needed |
| `docs/` | Shape has its own documentation |

## Migration Phases

### Phase 1: Setup and Copy (Day 1, 4-6 hours)

**Goal:** Create structure and copy files

**Tasks:**

1. **Create Directory Structure** (30 min)
   ```bash
   cd /Users/michaelsundell/Projects/shapestone/shape
   mkdir -p internal/tokenizer
   ```

2. **Copy Core Files** (2-3 hours)
   ```bash
   # Copy streams
   cp df2-go/streams/stream.go shape/internal/tokenizer/stream.go
   cp df2-go/streams/patterns.go shape/internal/tokenizer/patterns.go
   
   # Copy tokens
   cp df2-go/tokens/tokens.go shape/internal/tokenizer/tokens.go
   cp df2-go/tokens/tokenizer.go shape/internal/tokenizer/tokenizer.go
   cp df2-go/tokens/tokenizer_lib.go shape/internal/tokenizer/matchers.go
   
   # Copy text utilities
   cp df2-go/text/text.go shape/internal/tokenizer/text.go
   cp df2-go/text/runes.go shape/internal/tokenizer/runes.go
   
   # Copy number utilities
   cp df2-go/numbers/numbers.go shape/internal/tokenizer/numbers.go
   ```

3. **Update Package Declarations** (1-2 hours)
   - Change `package df2` → `package tokenizer`
   - Change `package streams` → `package tokenizer`
   - Change `package tokens` → `package tokenizer`
   - Change `package text` → `package tokenizer`
   - Change `package numbers` → `package tokenizer`

4. **Fix Import Paths** (30-60 min)
   - Update internal imports from df2 packages to tokenizer package
   - Remove unused imports
   - Add shape-specific imports if needed

**Deliverable:** Files copied with correct package names

### Phase 2: Refactor and Consolidate (Day 1-2, 4-6 hours)

**Goal:** Adapt code for shape's needs

**Tasks:**

1. **Consolidate Files** (2-3 hours)
   ```
   Before:
   - stream.go
   - patterns.go
   - tokens.go
   - tokenizer.go
   - tokenizer_lib.go
   - text.go
   - runes.go
   
   After (consolidated):
   - stream.go (stream + patterns)
   - tokens.go (tokens + tokenizer core)
   - matchers.go (matcher interface + built-ins)
   - text.go (text + runes utilities)
   - numbers.go (unchanged)
   - position.go (extracted position tracking)
   ```

2. **Simplify APIs** (1-2 hours)
   - Remove df2-specific features not needed by shape
   - Simplify interfaces for validation schema use case
   - Extract position tracking into separate file
   - Add shape-specific conveniences

3. **Update Documentation** (1 hour)
   - Add godoc comments for shape context
   - Update examples to use shape types
   - Document matcher composition patterns

**Deliverable:** Refactored, shape-optimized tokenizer

### Phase 3: Migrate Tests (Day 2, 6-8 hours)

**Goal:** Ensure correctness with comprehensive tests

**Tasks:**

1. **Copy Test Files** (1-2 hours)
   ```bash
   cp df2-go/streams/stream_test.go shape/internal/tokenizer/stream_test.go
   cp df2-go/streams/patterns_test.go shape/internal/tokenizer/patterns_test.go
   cp df2-go/tokens/tokens_test.go shape/internal/tokenizer/tokens_test.go
   cp df2-go/tokens/tokenizer_test.go shape/internal/tokenizer/tokenizer_test.go
   cp df2-go/tokens/tokenizer_lib_test.go shape/internal/tokenizer/matchers_test.go
   cp df2-go/text/text_test.go shape/internal/tokenizer/text_test.go
   cp df2-go/text/runes_test.go shape/internal/tokenizer/runes_test.go
   cp df2-go/numbers/numbers_test.go shape/internal/tokenizer/numbers_test.go
   ```

2. **Update Test Package Names** (30-60 min)
   - Change package names to `tokenizer`
   - Update import paths
   - Fix test references to match refactored code

3. **Run and Fix Tests** (3-4 hours)
   ```bash
   cd shape
   go test ./internal/tokenizer/...
   ```
   - Fix any broken tests due to refactoring
   - Add missing test coverage
   - Ensure 95%+ coverage

4. **Add Shape-Specific Tests** (1-2 hours)
   - Test validation schema patterns
   - Test function call tokenization: `Integer(1,100)`
   - Test type identifier tokenization: `UUID`, `Email`
   - Integration tests with parsers

**Deliverable:** 95%+ test coverage, all tests passing

### Phase 4: Integration and Documentation (Day 2-3, 2-3 hours)

**Goal:** Integrate with format parsers and document

**Tasks:**

1. **Update Parser Imports** (30 min)
   - Update JSONV parser to import `internal/tokenizer`
   - Verify parser still works
   - No functional changes to parsers

2. **Create Tokenizer Documentation** (1 hour)
   ```
   internal/tokenizer/README.md:
   - Overview of embedded tokenizer framework
   - API reference for format parser authors
   - Matcher composition examples
   - Position tracking usage
   - Performance characteristics
   ```

3. **Update Architecture Docs** (30-60 min)
   - Update references in ARCHITECTURE.md
   - Update IMPLEMENTATION_ROADMAP.md
   - Update DATA_VALIDATOR_INTEGRATION.md
   - Update SUMMARY.md

4. **Update Public Documentation** (30 min)
   - Update README.md to emphasize self-contained nature
   - Update DELIVERABLES.md
   - Update ADR 0002 with superseded notice

**Deliverable:** Complete documentation

### Phase 5: Cleanup and Verification (Day 3, 1-2 hours)

**Goal:** Final cleanup and verification

**Tasks:**

1. **Remove df2-go Dependency** (15 min)
   ```bash
   # Update go.mod
   # Remove: require github.com/shapestone/df2-go vX.X.X
   go mod tidy
   ```

2. **Final Testing** (30-45 min)
   ```bash
   make test
   make lint
   make coverage
   ```
   - Verify all tests pass
   - Verify coverage >= 95%
   - Verify no linter errors

3. **Performance Benchmarks** (30 min)
   ```bash
   go test -bench=. ./internal/tokenizer/
   ```
   - Verify performance is equivalent to df2-go
   - Document any performance changes

**Deliverable:** Clean, tested, documented codebase

## Directory Structure After Migration

```
shape/
├── go.mod                          # No df2-go dependency
├── go.sum                          
├── README.md                       # Updated: emphasizes self-contained
├── docs/
│   ├── architecture/
│   │   ├── ARCHITECTURE.md         # Updated: embedded tokenization section
│   │   ├── IMPLEMENTATION_ROADMAP.md  # Updated: migration tasks added
│   │   ├── DATA_VALIDATOR_INTEGRATION.md  # Updated: simplified dependencies
│   │   ├── SUMMARY.md              # Updated: self-contained emphasis
│   │   ├── MIGRATION_PLAN.md       # This document
│   │   └── decisions/
│   │       ├── 0001-ast-design.md
│   │       ├── 0002-use-df2-go.md  # Marked as superseded
│   │       └── 0003-embed-tokenizer.md  # New ADR
│   └── ...
├── internal/
│   ├── tokenizer/                  # NEW: Embedded tokenization framework
│   │   ├── README.md              # Tokenizer framework documentation
│   │   ├── stream.go              # Stream abstraction + patterns
│   │   ├── stream_test.go
│   │   ├── tokens.go              # Token + Tokenizer
│   │   ├── tokens_test.go
│   │   ├── matchers.go            # Matcher interface + built-ins
│   │   ├── matchers_test.go
│   │   ├── position.go            # Position tracking
│   │   ├── text.go                # Text + rune utilities
│   │   ├── text_test.go
│   │   ├── numbers.go             # Number parsing
│   │   └── numbers_test.go
│   │
│   └── parser/
│       ├── jsonv/
│       │   ├── tokenizer.go       # Uses internal/tokenizer
│       │   └── parser.go
│       └── ...
└── pkg/
    ├── shape/
    └── ast/
```

## Effort Estimates

| Phase | Duration | Description |
|-------|----------|-------------|
| Phase 1: Setup and Copy | 4-6 hours | Create structure, copy files, update packages |
| Phase 2: Refactor | 4-6 hours | Consolidate, simplify, optimize for shape |
| Phase 3: Tests | 6-8 hours | Migrate tests, fix, achieve 95%+ coverage |
| Phase 4: Documentation | 2-3 hours | Document tokenizer, update architecture docs |
| Phase 5: Cleanup | 1-2 hours | Remove dependency, final verification |
| **Total** | **17-25 hours** | ~2-3 days |

**Recommendation:** Allocate 3 full days for migration to allow buffer.

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Migration introduces bugs | High | Low | Comprehensive tests, copy df2-go tests |
| Performance regression | Medium | Low | Benchmark before/after, profile |
| Breaking parser integration | High | Low | Integration tests, minimal API changes |
| Incomplete test coverage | Medium | Low | Target 95%+, measure with coverage tools |
| Documentation gaps | Low | Medium | Checklist for all doc updates |

**Overall Risk:** **LOW**

## Success Criteria

- [ ] All df2-go tokenization code embedded in `internal/tokenizer/`
- [ ] Zero external dependencies except google/uuid and yaml.v3
- [ ] All df2-go tests migrated and passing
- [ ] 95%+ test coverage in tokenizer package
- [ ] Format parsers work with embedded tokenizer
- [ ] Performance equivalent to or better than df2-go
- [ ] All architecture documentation updated
- [ ] ADR 0002 marked as superseded
- [ ] ADR 0003 created and accepted
- [ ] README.md emphasizes self-contained nature
- [ ] go.mod has no df2-go dependency
- [ ] CI/CD passes all checks

## Impact on Roadmap

### Phase 1 (Week 1) Changes

**Original Phase 1:** 30-40 hours
- Project structure
- AST model
- CI/CD setup

**Updated Phase 1:** 47-65 hours
- Project structure
- AST model
- CI/CD setup
- **Tokenization migration** (17-25 hours)

**Recommendation:** Extend Phase 1 to 1.5-2 weeks or split migration into separate phase.

### Alternative: Pre-Phase 0 (Week 0.5)

**Option:** Do migration as "Phase 0" before official Phase 1 kickoff.

**Phase 0: Tokenization Migration (3 days)**
- Migrate df2-go code
- Test and document
- No AST work yet

**Phase 1: Foundation (Week 1)**
- Proceed as originally planned
- Use embedded tokenizer from day 1

**Benefit:** Clean separation, Phase 1 starts with tokenizer ready

## Post-Migration

### Verify Integration

1. **Test with Format Parsers**
   - Ensure JSONV parser still works
   - Verify position tracking works
   - Check error messages include line/column

2. **Performance Validation**
   - Run benchmarks
   - Compare with df2-go baseline
   - Profile memory usage

3. **Documentation Review**
   - Verify all df2-go references removed
   - Check consistency across docs
   - Ensure examples still work

### Future Considerations

**If other projects need tokenization:**
1. They can depend on shape (use shape's tokenizer)
2. They can copy tokenizer code (internal/ not importable)
3. We can extract tokenizer to separate library (if demand exists)

**For now:** Keep it simple, embed in shape.

## Checklist

### Pre-Migration
- [ ] Read this migration plan
- [ ] Review ADR 0003
- [ ] Allocate 3 days for migration
- [ ] Backup current shape codebase
- [ ] Note df2-go commit SHA for reference

### During Migration
- [ ] Complete Phase 1: Setup and Copy
- [ ] Complete Phase 2: Refactor
- [ ] Complete Phase 3: Tests (95%+ coverage)
- [ ] Complete Phase 4: Documentation
- [ ] Complete Phase 5: Cleanup

### Post-Migration
- [ ] All tests passing
- [ ] All linters passing
- [ ] Coverage >= 95%
- [ ] Documentation updated
- [ ] ADR 0003 approved
- [ ] go.mod clean (no df2-go)
- [ ] CI/CD green
- [ ] Migration tagged in git

## References

- ADR 0002: Use df2-go (superseded)
- ADR 0003: Embed Tokenizer (new)
- df2-go repository: /Users/michaelsundell/Projects/shapestone/df2-go
- ARCHITECTURE.md
- IMPLEMENTATION_ROADMAP.md

## Contacts

**Migration Lead:** System Architect  
**Implementation:** Full-Stack Engineer  
**Review:** Technical Product Owner

---

**Status:** Ready for Execution  
**Next Step:** Begin Phase 1 - Setup and Copy
