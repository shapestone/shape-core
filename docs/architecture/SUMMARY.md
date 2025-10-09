# Shape Parser - Architecture Summary

**Date:** 2025-10-09  
**Version:** 1.0  
**Status:** Complete

## Executive Summary

Shape is a multi-format validation schema parser library that converts 6 validation formats (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV) into a unified Abstract Syntax Tree (AST). This document summarizes the complete architectural design.

## Key Decisions

### 1. Separate Repository Strategy

**Decision:** Shape as standalone library, separate from data-validator

**Rationale:**
- Clear separation of concerns (parsing vs validation)
- Reusable by other projects
- Independent versioning
- Smaller, focused codebases

**Reference:** ARCHITECTURE.md Section 1.2

### 2. Interface-Based AST with 5 Node Types

**Decision:** Use interface-based AST with LiteralNode, TypeNode, FunctionNode, ObjectNode, ArrayNode

**Rationale:**
- Type-safe, compile-time guarantees
- Visitor pattern for traversal
- Format-agnostic representation
- Extensible design

**Reference:** ADR 0001

### 3. Integrated Tokenizer Framework

**Decision:** Integrate tokenizer framework for most formats

**Rationale:**
- 60-70% code reuse (40-60 hours vs 100-150 hours)
- UTF-8 and backtracking built-in
- Battle-tested, production-ready
- Self-contained framework (no external dependencies)

**Reference:** ADR 0002

### 4. Recursive Descent Parsing

**Decision:** Format-specific recursive descent parsers on top of integrated tokenizers

**Rationale:**
- Clear, readable code
- Easy to debug
- Flexible for ambiguous syntax
- Well-suited for validation formats

**Reference:** ARCHITECTURE.md Section 2

## Component Architecture

```
shape/
├── pkg/shape/          # Public API (Parse, ParseAuto, MustParse)
├── pkg/ast/            # AST model (SchemaNode, 5 node types)
└── internal/parser/    # Format parsers (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV)
    ├── parser.go       # Parser interface
    ├── factory.go      # Parser factory
    └── {format}/       # Format-specific parsers
        ├── tokenizer.go
        └── parser.go
```

## AST Node Types

1. **LiteralNode** - Exact value: `"active"`, `42`, `true`, `null`
2. **TypeNode** - Type identifier: `UUID`, `Email`, `ISO-8601`
3. **FunctionNode** - Function call: `Integer(1, 100)`, `String(1+)`
4. **ObjectNode** - Object/map: `{"id": UUID}`
5. **ArrayNode** - Array: `[String(1,50)]`

**Key Feature:** All formats produce the same AST structure

## Format Support

| Format | Description | Priority | Phase |
|--------|-------------|----------|-------|
| JSONV | JSON + validation | Critical | 2 |
| XMLV | XML + validation | High | 3 |
| PropsV | Properties + validation | High | 3 |
| CSVV | CSV + validation | Medium | 3 |
| YAMLV | YAML + validation | Medium | 4 |
| TEXTV | Text patterns + validation | Low | 4 |

## Implementation Timeline

**4-Week Roadmap:**

- **Week 1:** AST model, project structure
- **Week 2:** JSONV parser (first format)
- **Week 3:** XMLV, PropsV, CSVV parsers
- **Week 4:** YAMLV, TEXTV, documentation, v0.1.0 release

**Estimated Effort:** 120-160 hours

**Reference:** IMPLEMENTATION_ROADMAP.md

## Public API

```go
package shape

// Parse with explicit format
func Parse(format Format, input string) (ast.SchemaNode, error)

// Auto-detect format
func ParseAuto(input string) (ast.SchemaNode, Format, error)

// Parse or panic (tests/init)
func MustParse(format Format, input string) ast.SchemaNode
```

**Design Goals:**
- Simple, ergonomic API
- < 10 lines of code for basic usage
- Clear error messages with line/column

## data-validator Integration

### Separation of Concerns

| Component | Responsibility |
|-----------|----------------|
| shape | Parse schemas → AST |
| data-validator | Traverse AST + validate data |
| wire | Evaluate validation expressions |

### Usage Pattern

```go
// Parse schema (shape)
ast, err := shape.Parse(parser.FormatJSONV, schemaInput)

// Validate data (data-validator)
err = validator.ValidateWithAST(ast, data)
```

**Reference:** DATA_VALIDATOR_INTEGRATION.md

## Testing Strategy

**Coverage Goals:**
- AST: 100%
- JSONV Parser: 95%
- Other Parsers: 90%
- Overall: 95%+

**Test Categories:**
- Unit tests (90%): Each component in isolation
- Integration tests (10%): Multi-format, round-trip, data-validator

## Performance Targets

- Simple schema (< 10 nodes): < 100μs
- Medium schema (10-50 nodes): < 500μs  
- Large schema (50-200 nodes): < 2ms

## Success Metrics

### Technical
- All 6 formats parse correctly
- Error messages include line/column
- Performance targets met
- 95%+ test coverage

### Functional
- Successfully used by data-validator
- Produces same AST for equivalent schemas across formats
- Clear, helpful error messages

### Project
- v0.1.0 ships in 4 weeks
- Complete documentation
- Production-ready code quality

## Documentation Deliverables

### Complete ✓

1. **ARCHITECTURE.md** - Complete system design
2. **IMPLEMENTATION_ROADMAP.md** - 4-week implementation plan
3. **DATA_VALIDATOR_INTEGRATION.md** - Integration guide
4. **README.md** - Project overview, quick start
5. **ADR 0001** - AST design rationale
6. **ADR 0002** - Integrated tokenizer framework rationale

### To Be Created (During Implementation)

1. Format specifications (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV)
2. ADR 0003 - Parser strategy
3. ADR 0004 - Error handling
4. Contributor guides (setup, testing, contributing)
5. Code examples (basic, advanced, multi-format)

## Dependencies

### External
- **google/uuid v1.6.0** - UUID generation (stable)
- **gopkg.in/yaml.v3** - YAML parsing (for YAMLV)

### Integrated Components
- **Tokenizer framework** - Built into shape at internal/streams, internal/tokens, internal/text (self-contained)

### Dependents
- **data-validator** - Uses shape for schema parsing
- Future projects needing validation schema parsing

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Tokenizer framework performance | Medium | Low | Early benchmarking, direct optimization |
| YAMLV complexity | Medium | Medium | Use yaml.v3 library, MVP approach |
| Scope creep | High | Medium | Strict phase boundaries, no extras |
| API design changes | Medium | Low | Early review, data-validator feedback |

**Overall Risk:** LOW

## Next Steps

### Immediate (Week 1)
1. Review this architecture with team
2. Get approval for repository structure
3. Set up shape repository
4. Begin Phase 1: AST implementation

### Short-term (Weeks 2-4)
1. Implement JSONV parser
2. Get data-validator feedback on API
3. Implement remaining formats
4. Complete documentation

### Medium-term (Post-v0.1.0)
1. Integrate with data-validator
2. Gather production feedback
3. Plan v0.2.0 features

## Document Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Complete system design | All developers |
| [IMPLEMENTATION_ROADMAP.md](IMPLEMENTATION_ROADMAP.md) | 4-week plan | Implementation team |
| [DATA_VALIDATOR_INTEGRATION.md](DATA_VALIDATOR_INTEGRATION.md) | Integration guide | data-validator team |
| [README.md](../../README.md) | Project overview | External users |
| [ADR 0001](decisions/0001-ast-design.md) | AST design | Architects |
| [ADR 0002](decisions/0002-use-df2-go.md) | Integrated tokenizer framework | Architects |
| SUMMARY.md (this doc) | Architecture summary | Stakeholders |

## Approval

This architecture has been designed by the System Architect and is ready for:
- [ ] Team review
- [ ] Stakeholder approval
- [ ] Implementation kickoff

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-09 | Initial complete architecture |

---

**Status:** Ready for Implementation  
**Contact:** System Architect
