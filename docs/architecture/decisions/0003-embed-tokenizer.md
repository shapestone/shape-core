# ADR 0003: Embed Tokenization Layer (Supersedes df2-go Dependency)

**Status:** Accepted  
**Date:** 2025-10-09  
**Supersedes:** ADR 0002 (use df2-go dependency)

## Context

ADR 0002 initially proposed using df2-go as an external dependency for tokenization. However, after architectural review, we've identified that shape should be a self-contained library that owns its tokenization responsibilities entirely.

### Previous Architecture (ADR 0002)
```
shape/
└── depends on df2-go (external dependency)
    └── tokenization framework

data-validator/
└── depends on shape
    └── depends on df2-go (transitive)
```

**Problems with External Dependency:**
- Transitive dependency burden on consumers
- Version coupling between shape and df2-go
- Less flexibility to evolve tokenization for shape's needs
- Deployment complexity (multiple repos to manage)
- Unclear ownership of tokenization layer

## Decision

We will **embed the df2-go tokenization code directly into shape** at `internal/tokenizer/`. Shape will become a self-contained parser library with no external dependencies (except google/uuid and gopkg.in/yaml.v3).

### New Architecture
```
shape/
├── internal/
│   ├── tokenizer/          # Tokenization layer (embedded from df2-go)
│   │   ├── stream.go      # Character stream handling
│   │   ├── matchers.go    # Token matchers
│   │   ├── position.go    # Position tracking
│   │   └── tokens.go      # Token types and tokenizer
│   └── parser/
│       ├── jsonv/          # JSONV parser (uses internal tokenizer)
│       ├── xmlv/
│       └── ...

data-validator/
└── depends on shape only (self-contained)
```

## Rationale

### 1. Single Responsibility, Clear Ownership

**Shape owns parsing entirely:**
- Tokenization is core to shape's parsing mission
- Shape has full control over tokenization evolution
- No ambiguity about who owns the tokenization layer
- Simpler mental model: "shape parses schemas, period"

### 2. Simplified Dependency Chain

**Before (with df2-go dependency):**
```
data-validator → shape → df2-go → google/uuid
```

**After (embedded tokenization):**
```
data-validator → shape
shape → google/uuid (only for AST UUIDs)
```

**Benefits:**
- One fewer repository to manage
- No version coordination between shape and df2-go
- Consumers only depend on shape
- Cleaner go.mod files
- Easier deployment

### 3. Full Control and Flexibility

**Evolution without coordination:**
- Shape can evolve tokenization for its specific needs
- No need to coordinate changes with df2-go maintainers
- Can optimize tokenization for schema patterns
- No risk of breaking changes from df2-go updates

**Example: Shape-specific optimizations:**
- Tokenize schema expressions (Integer(1,100)) efficiently
- Optimize for schema patterns (not general text processing)
- Add shape-specific token types without df2-go changes

### 4. Self-Contained Library Philosophy

**Shape as standalone library:**
- "Zero dependencies" is a strong selling point
- Easier to adopt (fewer deps to audit)
- Smaller attack surface
- Predictable behavior (no external changes)

**Compare:**
- **With df2-go:** "Shape depends on df2-go for tokenization"
- **Embedded:** "Shape is self-contained with built-in tokenization"

Much clearer value proposition.

### 5. Maintenance Simplicity

**Single codebase:**
- All shape code in one repository
- Unified CI/CD pipeline
- Single release process
- One issue tracker
- Simpler for contributors

**No transitive maintenance:**
- No need to track df2-go releases
- No need to coordinate breaking changes
- No version compatibility matrix

## Migration Strategy

### What to Migrate from df2-go

**Core Components to Embed:**
1. **streams/** → `internal/tokenizer/stream.go`
   - Stream interface
   - Character stream implementation
   - Position tracking
   - Clone/backtracking support

2. **tokens/** → `internal/tokenizer/tokens.go`, `internal/tokenizer/matchers.go`
   - Token struct
   - Tokenizer implementation
   - Matcher interface
   - Built-in matchers (CharMatcher, StringMatcher, etc.)

3. **text/** → `internal/tokenizer/text.go`
   - Rune utilities
   - Text manipulation helpers

4. **numbers/** → `internal/tokenizer/numbers.go`
   - Number parsing utilities

**What NOT to Migrate:**
- `df/json/` - Shape will implement its own JSONV tokenizer
- `df/scenario/` - Not relevant to shape
- `parsers/` - Shape implements its own parsers
- `main.go` - df2-go executable, not needed

### Migration Steps

**Phase 1: Copy Core Framework (4-6 hours)**

1. Create `internal/tokenizer/` directory structure
2. Copy df2-go core files:
   ```
   df2-go/streams/   → shape/internal/tokenizer/stream.go
   df2-go/tokens/    → shape/internal/tokenizer/tokens.go + matchers.go
   df2-go/text/      → shape/internal/tokenizer/text.go
   df2-go/numbers/   → shape/internal/tokenizer/numbers.go
   ```
3. Update package declarations from `df2` to `tokenizer`
4. Update import paths within embedded code

**Phase 2: Refactor for Shape (4-6 hours)**

1. Simplify APIs for shape's use cases
2. Remove df2-specific features not needed by shape
3. Consolidate files (e.g., merge stream.go + patterns.go)
4. Add shape-specific optimizations
5. Update godoc comments for shape context

**Phase 3: Migrate Tests (6-8 hours)**

1. Copy relevant test files
2. Update test package names
3. Ensure all tests pass
4. Add shape-specific test cases
5. Achieve 95%+ coverage

**Phase 4: Update Documentation (2-3 hours)**

1. Update all architecture docs (this ADR, ARCHITECTURE.md, etc.)
2. Document tokenizer API for format parser authors
3. Create internal/tokenizer/README.md
4. Remove df2-go references

**Total Migration Effort: 16-23 hours**

### Directory Structure After Migration

```
shape/
├── internal/
│   ├── tokenizer/              # Embedded tokenization framework
│   │   ├── stream.go          # Stream interface + implementation
│   │   ├── tokens.go          # Token struct + Tokenizer
│   │   ├── matchers.go        # Matcher interface + built-ins
│   │   ├── text.go            # Text utilities
│   │   ├── numbers.go         # Number utilities
│   │   ├── position.go        # Position tracking
│   │   ├── stream_test.go     # Tests
│   │   ├── tokens_test.go     # Tests
│   │   ├── matchers_test.go   # Tests
│   │   └── README.md          # Tokenizer framework docs
│   │
│   └── parser/
│       ├── jsonv/
│       │   ├── tokenizer.go   # JSONV-specific matchers (uses internal/tokenizer)
│       │   └── parser.go      # JSONV parser
│       └── ...
```

## Alternatives Considered

### 1. Keep df2-go as External Dependency (ADR 0002)

**Pros:**
- Separate evolution of tokenization framework
- Could be reused by other projects

**Cons:**
- Transitive dependency for all consumers
- Version coordination overhead
- Less flexibility for shape-specific needs
- More complex deployment
- Unclear ownership

**Rejected:** Embedding provides more benefits than separation

### 2. Use Go Standard Library (text/scanner)

**Pros:**
- No custom code
- Standard library support

**Cons:**
- Insufficient for validation schema tokenization
- No backtracking support
- Limited pattern matching
- Still need significant custom code

**Rejected:** Doesn't meet shape's needs

### 3. Build Tokenization from Scratch

**Pros:**
- Perfectly optimized for shape
- No migration needed

**Cons:**
- 40-60 hours of additional work
- Need to implement UTF-8, backtracking, position tracking
- Reinventing proven code
- More testing burden

**Rejected:** df2-go code is proven and ready

### 4. Create Shared Tokenization Library

**Pros:**
- Could be reused by multiple projects

**Cons:**
- Same issues as df2-go dependency
- Premature optimization (no other users yet)
- Maintenance overhead

**Rejected:** YAGNI - embed now, extract later if needed

## Implementation Notes

### Package Naming

**Internal package:** `internal/tokenizer`

This makes it clear:
- Tokenization is internal implementation detail
- Not part of shape's public API
- Cannot be imported by external packages

### API Surface

**Keep it minimal:**
- Expose only what format parsers need
- Hide implementation details
- Document with godoc
- Provide examples

### Testing

**Comprehensive coverage:**
- Unit tests for all tokenizer components
- Integration tests with format parsers
- Performance benchmarks
- 95%+ coverage target

## Consequences

### Positive

- **Self-Contained:** Shape has zero external dependencies (except uuid, yaml)
- **Full Control:** Can evolve tokenization without coordination
- **Simpler Dependencies:** Consumers only depend on shape
- **Easier Maintenance:** Single repository, unified process
- **Better Ownership:** Clear that shape owns all parsing
- **Flexibility:** Can optimize tokenization for shape's needs
- **Faster Development:** No cross-repo coordination needed

### Negative

- **Code Size:** Adds ~2000-3000 lines to shape codebase
- **Migration Effort:** 16-23 hours to migrate and test
- **Maintenance Burden:** Shape owns tokenizer bug fixes
- **Cannot Share:** Other projects can't easily reuse tokenization

### Neutral

- **df2-go Future:** df2-go can remain as separate project if needed elsewhere
- **Code Duplication:** If another project needs tokenization, they can copy or depend on shape

### Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Migration bugs | Medium | Comprehensive testing, copy tests from df2-go |
| Performance regression | Low | Benchmark before/after, profile |
| Increased complexity | Low | Good documentation, clear structure |
| Maintenance burden | Medium | Code is stable, minimal changes expected |

**Overall Risk:** LOW

## Impact on Roadmap

**Phase 1 (Week 1) Changes:**

Add migration tasks:
1. Copy df2-go code to internal/tokenizer/ (4-6 hours)
2. Refactor for shape (4-6 hours)
3. Migrate tests (6-8 hours)
4. Update documentation (2-3 hours)

**Total Additional Effort:** 16-23 hours

**Recommendation:** Add 1-2 days to Phase 1 for migration work.

**Updated Phase 1 Duration:** 40-50 hours (was 30-40 hours)

## Success Metrics

- All df2-go tokenization code successfully embedded in shape
- Zero external dependencies (except google/uuid, yaml.v3)
- All existing df2-go tests pass in new location
- 95%+ test coverage maintained
- Documentation updated to reflect embedded tokenization
- Format parsers work with embedded tokenizer
- Performance equivalent to or better than df2-go

## Migration Checklist

- [ ] Create internal/tokenizer/ directory structure
- [ ] Copy streams, tokens, text, numbers packages
- [ ] Update package names and import paths
- [ ] Refactor APIs for shape's needs
- [ ] Migrate and adapt tests
- [ ] Achieve 95%+ coverage
- [ ] Update all architecture documentation
- [ ] Update ADR 0002 with superseded notice
- [ ] Create tokenizer API documentation
- [ ] Remove df2-go from go.mod
- [ ] Update README.md to highlight self-contained nature
- [ ] Verify format parsers work with embedded tokenizer

## References

- ADR 0002: Use df2-go (superseded by this ADR)
- df2-go repository: github.com/shapestone/df2-go
- ARCHITECTURE.md: Updated to reflect embedded tokenization
- IMPLEMENTATION_ROADMAP.md: Updated with migration tasks

## Date

2025-10-09

---

**Note:** This ADR supersedes ADR 0002. The decision to use df2-go as an external dependency has been replaced with the decision to embed tokenization code directly into shape for better ownership, simplicity, and flexibility.
