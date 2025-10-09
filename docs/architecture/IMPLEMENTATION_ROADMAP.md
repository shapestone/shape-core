# Shape Parser - 4-Week Implementation Roadmap

**Document Version:** 1.0  
**Date:** 2025-10-09  
**Target Release:** shape v0.1.0

## Overview

This roadmap outlines a 4-week implementation plan for the shape parser library. The plan is structured to deliver incremental value, with each phase producing working, tested code.

**Total Estimated Effort:** 120-160 hours (3-4 weeks full-time)

## Phase Summary

| Phase | Duration | Focus | Key Deliverables |
|-------|----------|-------|------------------|
| Phase 1 | Week 1 | Foundation | AST model, project structure |
| Phase 2 | Week 2 | JSONV Parser | First complete format parser |
| Phase 3 | Week 3 | Core Formats | XMLV, PropsV, CSVV parsers |
| Phase 4 | Week 4 | Advanced + Polish | YAMLV, TEXTV, documentation |

## Phase 1: Foundation (Week 1)

**Goal:** Establish core architecture and AST model  
**Duration:** 30-40 hours  
**Team:** 1 full-stack engineer or 1 system architect

### Deliverables

#### 1.1 Project Structure (4 hours)

**Tasks:**
- Create repository structure following standard Go layout
- Initialize Go module (`go.mod`)
- Set up Makefile (build, test, lint, coverage targets)
- Configure golangci-lint
- Set up GitHub Actions CI/CD (test, lint)

**Files Created:**
```
shape/
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml
├── .github/workflows/ci.yml
├── README.md (basic)
└── LICENSE
```

**Success Criteria:**
- `make test` runs (even with no tests)
- `make lint` passes
- CI pipeline runs on GitHub

#### 1.2 AST Node Interfaces (4 hours)

**Tasks:**
- Define `SchemaNode` interface
- Define `NodeType` enum
- Define `Position` struct
- Create `Visitor` interface

**Files Created:**
```
pkg/ast/
├── node.go       # SchemaNode interface
├── types.go      # NodeType enum
└── position.go   # Position struct
```

**Success Criteria:**
- All interfaces compile
- Godoc comments complete
- Clear API design

#### 1.3 AST Node Implementations (10-12 hours)

**Tasks:**
- Implement `LiteralNode`
- Implement `TypeNode`
- Implement `FunctionNode`
- Implement `ObjectNode`
- Implement `ArrayNode`
- Implement all required methods (Type, Accept, String, Position)

**Files Created:**
```
pkg/ast/
├── literal.go      # LiteralNode
├── type.go         # TypeNode
├── function.go     # FunctionNode
├── object.go       # ObjectNode
├── array.go        # ArrayNode
└── constructors.go # Node constructors
```

**Success Criteria:**
- All node types implement SchemaNode interface
- Constructors for all node types
- String() method returns readable output
- Position tracking works

#### 1.4 AST Utilities (8-10 hours)

**Tasks:**
- Implement visitor pattern
- Implement JSON serialization (MarshalJSON/UnmarshalJSON)
- Implement pretty-printer
- Add AST traversal utilities

**Files Created:**
```
pkg/ast/
├── visitor.go        # Visitor interface and base implementation
├── serialization.go  # JSON marshaling/unmarshaling
└── printer.go        # Pretty-print utilities
```

**Success Criteria:**
- Visitor can traverse entire AST
- AST can be serialized to JSON
- AST can be deserialized from JSON
- Pretty-printer produces readable output

#### 1.5 AST Tests (4-6 hours)

**Tasks:**
- Unit tests for all node types
- Tests for visitor pattern
- Tests for serialization (round-trip)
- Tests for pretty-printer

**Files Created:**
```
pkg/ast/
├── literal_test.go
├── type_test.go
├── function_test.go
├── object_test.go
├── array_test.go
├── visitor_test.go
├── serialization_test.go
└── printer_test.go
```

**Success Criteria:**
- 100% test coverage for AST package
- All tests pass
- Tests cover edge cases

### Phase 1 Exit Criteria

- [ ] Project structure complete
- [ ] All AST node types implemented
- [ ] Visitor pattern works
- [ ] Serialization round-trips correctly
- [ ] 100% test coverage for AST
- [ ] CI pipeline passes
- [ ] Documentation complete (godoc)

**Deliverable:** Working AST model that can be created programmatically

---

## Phase 2: JSONV Parser (Week 2)

**Goal:** Implement first complete format parser  
**Duration:** 35-45 hours  
**Team:** 1 full-stack engineer

### Deliverables

#### 2.1 Tokenizer Framework Setup (1-2 hours)

**Tasks:**
- Review integrated tokenizer framework (internal/streams, internal/tokens)
- Study existing JSON tokenizer patterns
- Document tokenizer framework usage patterns
- Create format-specific tokenizer templates

**Files to Review:**
```
internal/streams/  # Character stream abstraction
internal/tokens/   # Tokenizer framework
internal/text/     # Text utilities
```

**Success Criteria:**
- Team familiar with tokenizer framework API
- Pattern templates created for new format parsers

#### 2.2 JSONV Tokenizer (8-12 hours)

**Tasks:**
- Implement custom matchers:
  - `identifierMatcher` (UUID, Email, etc.)
  - `functionMatcher` (Integer(1,100))
  - `plusMatcher` (unbounded symbol)
- Reuse built-in matchers:
  - Object/array delimiters (CharMatcher, StringMatcher)
  - String/number literals
  - Keywords (true, false, null)
- Assemble matcher list
- Write tokenizer tests

**Files Created:**
```
internal/parser/jsonv/
├── tokenizer.go
└── tokenizer_test.go
```

**Success Criteria:**
- All JSONV tokens recognized
- Tokenizer handles all examples from JSONV spec
- Position tracking works
- 95%+ test coverage

#### 2.3 JSONV Parser (12-16 hours)

**Tasks:**
- Implement recursive descent parser:
  - `parseValue()` - Dispatch to specific parsers
  - `parseObject()` - Parse objects
  - `parseArray()` - Parse arrays
  - `parseFunction()` - Parse function calls
  - `parseIdentifier()` - Parse type identifiers
  - `parseLiteral()` - Parse literals
- Build AST from tokens
- Error handling with position tracking
- Write parser tests

**Files Created:**
```
internal/parser/jsonv/
├── parser.go
└── parser_test.go
```

**Success Criteria:**
- All valid JSONV inputs parse to correct AST
- Invalid inputs produce clear error messages
- Error messages include line/column
- 95%+ test coverage

#### 2.4 Parser Abstraction Layer (6-8 hours)

**Tasks:**
- Define `Parser` interface
- Define `Format` enum
- Implement `ParserFactory`
- Implement format detection
- Write abstraction tests

**Files Created:**
```
internal/parser/
├── parser.go      # Parser interface
├── factory.go     # ParserFactory
├── format.go      # Format enum and detection
├── errors.go      # ParseError types
└── parser_test.go # Tests
```

**Success Criteria:**
- Parser interface is clean and simple
- Factory can create JSONV parser
- Format detection identifies JSONV

#### 2.5 Public API (4-6 hours)

**Tasks:**
- Implement `Parse(format, input)`
- Implement `ParseAuto(input)`
- Implement `MustParse(format, input)`
- Write public API tests
- Write examples

**Files Created:**
```
pkg/shape/
├── shape.go
├── shape_test.go
└── examples_test.go  # Godoc examples
```

**Success Criteria:**
- Public API is ergonomic
- Examples compile and run
- Godoc is complete

#### 2.6 Test Data (2-3 hours)

**Tasks:**
- Create valid JSONV test files
- Create invalid JSONV test files
- Document test data organization

**Files Created:**
```
internal/testdata/jsonv/
├── valid/
│   ├── simple-object.jsonv
│   ├── nested-object.jsonv
│   ├── array-elements.jsonv
│   └── all-validators.jsonv
└── invalid/
    ├── unclosed-object.jsonv
    ├── missing-colon.jsonv
    └── invalid-function.jsonv
```

**Success Criteria:**
- Comprehensive test coverage
- Both positive and negative test cases

### Phase 2 Exit Criteria

- [ ] JSONV tokenizer complete and tested
- [ ] JSONV parser complete and tested
- [ ] Parser abstraction layer works
- [ ] Public API is ergonomic
- [ ] Error messages include line/column
- [ ] 95%+ test coverage
- [ ] Performance < 1ms for typical schemas
- [ ] Documentation complete

**Deliverable:** Working JSONV parser with public API

---

## Phase 3: Core Formats (Week 3)

**Goal:** Implement XMLV, PropsV, CSVV parsers  
**Duration:** 35-45 hours  
**Team:** 1 full-stack engineer

### 3.1 PropsV Parser (10-12 hours)

**Why PropsV First?** Simpler than XMLV, good second format.

**Tasks:**
- Implement PropsV tokenizer
- Implement PropsV parser
- Create test data
- Write tests

**Files Created:**
```
internal/parser/propsv/
├── tokenizer.go
├── parser.go
├── tokenizer_test.go
└── parser_test.go

internal/testdata/propsv/
├── valid/
└── invalid/
```

**Success Criteria:**
- PropsV parses correctly
- Produces same AST as equivalent JSONV
- 90%+ test coverage

### 3.2 XMLV Parser (14-18 hours)

**Tasks:**
- Implement XMLV tokenizer (XML tags, attributes, content)
- Implement XMLV parser
- Handle validation expressions in XML content
- Create test data
- Write tests

**Files Created:**
```
internal/parser/xmlv/
├── tokenizer.go
├── parser.go
├── tokenizer_test.go
└── parser_test.go

internal/testdata/xmlv/
├── valid/
└── invalid/
```

**Success Criteria:**
- XMLV parses correctly
- Produces same AST as equivalent JSONV
- 90%+ test coverage

### 3.3 CSVV Parser (10-12 hours)

**Tasks:**
- Implement CSVV tokenizer (line-oriented)
- Implement CSVV parser (header + validation row)
- Create test data
- Write tests

**Files Created:**
```
internal/parser/csvv/
├── tokenizer.go
├── parser.go
├── tokenizer_test.go
└── parser_test.go

internal/testdata/csvv/
├── valid/
└── invalid/
```

**Success Criteria:**
- CSVV parses correctly
- Handles CSV-specific features (quoting, escaping)
- 90%+ test coverage

### 3.4 Format Detection Enhancement (2-3 hours)

**Tasks:**
- Update format detection to handle all 4 formats
- Write format detection tests

**Files Modified:**
```
internal/parser/format.go       # Enhanced detection
internal/parser/format_test.go  # More tests
```

**Success Criteria:**
- Auto-detection works for JSONV, XMLV, PropsV, CSVV
- Clear error when format cannot be detected

### Phase 3 Exit Criteria

- [ ] PropsV parser complete
- [ ] XMLV parser complete
- [ ] CSVV parser complete
- [ ] Format detection works for all 4 formats
- [ ] 90%+ test coverage per format
- [ ] Cross-format tests (same schema in different formats → same AST)
- [ ] Performance acceptable (< 2ms for typical schemas)

**Deliverable:** 4 working format parsers (JSONV, XMLV, PropsV, CSVV)

---

## Phase 4: Advanced Formats + Polish (Week 4)

**Goal:** Complete remaining formats and production polish  
**Duration:** 30-40 hours  
**Team:** 1 full-stack engineer or system architect

### 4.1 YAMLV Parser (12-16 hours)

**Approach Decision:** Evaluate using `gopkg.in/yaml.v3` vs custom parser

**Tasks:**
- Evaluate yaml.v3 library
- Implement YAMLV parser (likely hybrid: yaml.v3 + custom validation overlay)
- Create test data
- Write tests

**Files Created:**
```
internal/parser/yamlv/
├── parser.go
├── parser_test.go
└── README.md  # Document YAML-specific approach

internal/testdata/yamlv/
├── valid/
└── invalid/
```

**Success Criteria:**
- YAMLV parses correctly
- Handles YAML-specific features (indentation, anchors)
- 85%+ test coverage

### 4.2 TEXTV Parser (10-12 hours)

**Tasks:**
- Implement TEXTV tokenizer (pattern-based)
- Implement TEXTV parser
- Create test data
- Write tests

**Files Created:**
```
internal/parser/textv/
├── tokenizer.go
├── parser.go
├── tokenizer_test.go
└── parser_test.go

internal/testdata/textv/
├── valid/
└── invalid/
```

**Success Criteria:**
- TEXTV parses correctly
- Pattern matching works
- 85%+ test coverage

### 4.3 Performance Benchmarks (4-6 hours)

**Tasks:**
- Write benchmarks for all formats
- Benchmark simple, medium, large schemas
- Profile memory usage
- Document performance characteristics

**Files Created:**
```
pkg/shape/bench_test.go
docs/architecture/PERFORMANCE.md
```

**Success Criteria:**
- Benchmarks for all formats
- Performance targets met:
  - Simple schema (< 10 nodes): < 100μs
  - Medium schema (10-50 nodes): < 500μs
  - Large schema (50-200 nodes): < 2ms
- Memory usage documented

### 4.4 Documentation (6-8 hours)

**Tasks:**
- Write comprehensive README.md
- Write format specifications
- Write contributor guide
- Write migration guide for data-validator
- Write examples

**Files Created/Updated:**
```
README.md                                    # Complete
docs/architecture/specifications/
├── jsonv-spec.md
├── xmlv-spec.md
├── propsv-spec.md
├── csvv-spec.md
├── yamlv-spec.md
└── textv-spec.md

docs/contributor/
├── local-setup.md
├── contributing.md
└── testing-guide.md

examples/
├── basic/main.go
├── advanced/main.go
└── multi-format/main.go
```

**Success Criteria:**
- README explains what shape does
- Quick start example works
- All format specs complete
- Examples compile and run

### 4.5 Release Preparation (2-4 hours)

**Tasks:**
- Final linting pass
- Final test pass
- Create CHANGELOG.md
- Tag release v0.1.0
- Create GitHub release with notes

**Files Created:**
```
CHANGELOG.md
```

**Success Criteria:**
- All tests pass
- All linters pass
- 95%+ test coverage overall
- Release tagged and published

### Phase 4 Exit Criteria

- [ ] YAMLV parser complete
- [ ] TEXTV parser complete
- [ ] All 6 formats working
- [ ] Performance benchmarks complete
- [ ] Documentation complete
- [ ] Examples work
- [ ] v0.1.0 release published

**Deliverable:** Production-ready shape v0.1.0

---

## Testing Strategy

### Test Coverage Goals

| Component | Target Coverage | Priority |
|-----------|----------------|----------|
| AST (pkg/ast) | 100% | Critical |
| JSONV Parser | 95% | Critical |
| Other Parsers | 90% | High |
| Public API | 100% | Critical |
| Utilities | 85% | Medium |
| **Overall** | **95%+** | **Critical** |

### Test Categories

**Unit Tests (90% of tests):**
- Each AST node type
- Each tokenizer matcher
- Each parser method
- Error handling

**Integration Tests (10% of tests):**
- Multi-format (same schema → same AST)
- Round-trip (Parse → Serialize → Parse)
- data-validator integration

### Continuous Testing

- Run tests on every commit (CI)
- Measure coverage on every PR
- Fail CI if coverage drops below 95%

---

## Risk Management

### High-Risk Items

#### Risk: Tokenizer Framework Performance Issues

**Mitigation:**
- Benchmark early (Phase 2)
- Profile memory usage
- Optimize hot paths if needed
- Framework is battle-tested with 39 passing tests

**Contingency:** Can optimize framework directly since it's integrated

#### Risk: YAMLV Complexity

**Mitigation:**
- Evaluate yaml.v3 library early (Phase 4)
- Consider MVP YAML support (no anchors/aliases)
- Document limitations if needed

**Contingency:** Ship YAMLV in v0.2.0 if too complex

#### Risk: Scope Creep

**Mitigation:**
- Stick to 6 formats only
- No custom validators in v0.1.0
- No schema validation in v0.1.0
- Clear phase boundaries

**Contingency:** Move features to v0.2.0

### Medium-Risk Items

#### Risk: API Design Changes

**Mitigation:**
- Public API review in Phase 2
- Get feedback from data-validator team
- Iterate before other formats

**Contingency:** v0.x.x allows breaking changes

#### Risk: Test Data Gaps

**Mitigation:**
- Create test data incrementally
- Review format specs for edge cases
- Copy valid examples from format specs

**Contingency:** Add tests as issues are found

---

## Dependencies

### External Dependencies

| Dependency | Version | Purpose | Risk |
|------------|---------|---------|------|
| google/uuid | v1.6.0 | UUID generation | Low (stable) |
| gopkg.in/yaml.v3 | v3.0.1 | YAMLV parsing | Low (stable) |

### Internal Dependencies

| Dependency | Timeline | Impact |
|------------|----------|--------|
| Tokenizer framework (integrated) | ✓ Ready | Enables all parsers |
| data-validator integration | Phase 2 exit | Validates API design |

**Note:** The tokenizer framework is now integrated into shape at `internal/streams/`, `internal/tokens/`, `internal/text/`, and `internal/numbers/`. No external tokenization dependencies required.

---

## Success Metrics

### Technical Metrics

- **Parsing Performance:** < 2ms for 200-node schema
- **Memory Usage:** < 10MB for large schema
- **Test Coverage:** 95%+ overall
- **Code Quality:** All linters pass, no panics

### Functional Metrics

- **Format Support:** All 6 formats parse correctly
- **Error Quality:** All errors include line/column
- **API Ergonomics:** < 10 lines of code for basic usage

### Project Metrics

- **On-Time Delivery:** v0.1.0 ships in 4 weeks
- **Documentation:** Complete README, format specs, examples
- **Integration:** Successfully used by data-validator

---

## Post-v0.1.0 Roadmap

### v0.2.0 (Future)
- Schema validation (validate schemas for correctness)
- AST optimization (simplify/optimize AST structure)
- Additional validators (custom validator registration)

### v1.0.0 (Future)
- Stable API guarantee
- Production battle-testing
- Performance optimizations
- Comprehensive documentation

---

## Daily Standup Template

**What I did yesterday:**
- [Task from roadmap]

**What I'm doing today:**
- [Task from roadmap]

**Blockers:**
- [Any blockers]

**Progress:**
- Phase X: Y% complete

---

## Phase Completion Checklist

Use this checklist at the end of each phase:

### Phase 1 Checklist
- [ ] Project structure complete
- [ ] All AST node types implemented
- [ ] Visitor pattern works
- [ ] Serialization round-trips
- [ ] 100% AST test coverage
- [ ] CI pipeline passes
- [ ] Godoc complete
- [ ] Phase 1 git tag created

### Phase 2 Checklist
- [ ] Tokenizer framework setup complete
- [ ] JSONV tokenizer complete
- [ ] JSONV parser complete
- [ ] Parser abstraction layer complete
- [ ] Public API complete
- [ ] 95%+ test coverage
- [ ] Performance < 1ms
- [ ] Godoc complete
- [ ] Phase 2 git tag created

### Phase 3 Checklist
- [ ] PropsV parser complete
- [ ] XMLV parser complete
- [ ] CSVV parser complete
- [ ] Format detection works
- [ ] 90%+ test coverage per format
- [ ] Cross-format tests pass
- [ ] Phase 3 git tag created

### Phase 4 Checklist
- [ ] YAMLV parser complete
- [ ] TEXTV parser complete
- [ ] Benchmarks complete
- [ ] Documentation complete
- [ ] Examples work
- [ ] v0.1.0 tagged and released
- [ ] GitHub release notes published

---

**Document Status:** Complete  
**Next Steps:** Begin Phase 1 implementation
