# Documentation Update Guide for Embedded Tokenization

**Date:** 2025-10-09  
**Version:** 1.0  
**Purpose:** Reference guide for updating all shape documentation to reflect embedded tokenization

## Overview

This document provides specific updates needed across all shape documentation to reflect the architectural change from external df2-go dependency to embedded tokenization layer.

## Global Changes

**Find and Replace:**
- "df2-go dependency" → "embedded tokenization"
- "uses df2-go" → "includes embedded tokenization framework"
- "depends on df2-go" → "includes tokenization layer"
- "github.com/shapestone/df2-go" → "internal/tokenizer (embedded)"

## Document-Specific Updates

### 1. README.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/README.md`

#### Line 15-16: Features Section
**OLD:**
```markdown
- **Zero Dependencies:** (Except google/uuid)
```

**NEW:**
```markdown
- **Self-Contained Library:** Zero external dependencies except google/uuid and gopkg.in/yaml.v3
- **Embedded Tokenization:** Built-in tokenization framework, no external tokenizer dependencies
- **Production-Ready:** Comprehensive error handling, battle-tested tokenization, 95%+ test coverage
```

#### Line 342-343: Related Projects Section
**OLD:**
```markdown
- **[df2-go](https://github.com/shapestone/df2-go):** Original tokenization framework (now integrated into shape)
```

**NEW:**
```markdown
- **[df2-go](https://github.com/shapestone/df2-go):** Tokenization framework code embedded in shape's `internal/tokenizer/` for self-contained operation
```

---

### 2. ARCHITECTURE.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/ARCHITECTURE.md`

#### Section 1.2: Ecosystem Position (Lines 36-78)
**UPDATE:** Diagram should show shape without df2-go dependency

**NEW Diagram:**
```
┌─────────────────────────────────────────────────────────┐
│                   Data Validator                        │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         Validation Traverser                      │  │
│  │  (Walks AST + validates data + calls wire)       │  │
│  └──────────────────────────────────────────────────┘  │
│                         ▲                                │
│                         │                                │
│                         │ uses AST                       │
│                         │                                │
└─────────────────────────┼────────────────────────────────┘
                          │
                          │
┌─────────────────────────┼────────────────────────────────┐
│                      Shape Parser                        │
│                         │                                │
│  ┌──────────────────────▼──────────────────────────┐    │
│  │              Schema AST Model                    │    │
│  │  (LiteralNode, TypeNode, FunctionNode, etc.)    │    │
│  └──────────────▲───────────────────────────────────┘   │
│                 │                                        │
│  ┌──────────────┴──────────────────────────────────┐    │
│  │           Format Parsers                         │    │
│  │  JSONV | XMLV | PropsV | CSVV | YAMLV | TEXTV  │    │
│  └──────────────┬──────────────────────────────────┘    │
│                 │                                        │
│  ┌──────────────▼──────────────────────────────────┐    │
│  │     Embedded Tokenization Framework             │    │
│  │     (internal/tokenizer/)                       │    │
│  │  Stream, Matchers, Position Tracking           │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
                          │
                          │
┌─────────────────────────┼────────────────────────────────┐
│                      Wire Engine                         │
│                         │                                │
│  (Expression evaluation for dynamic validation)          │
│  Integer(min, max) → validates with wire expressions     │
└──────────────────────────────────────────────────────────┘
```

#### Section 6: Directory Structure (Lines 446-558)
**ADD:** New section for internal/tokenizer/

**INSERT AFTER Line 512 (internal/parser/textv/):**
```markdown
│   │
│   ├── tokenizer/                # Embedded tokenization framework
│   │   ├── stream.go            # Character stream abstraction
│   │   ├── stream_test.go
│   │   ├── tokens.go            # Token struct and tokenizer
│   │   ├── tokens_test.go
│   │   ├── matchers.go          # Matcher interface + built-ins
│   │   ├── matchers_test.go
│   │   ├── position.go          # Position tracking
│   │   ├── text.go              # Text/rune utilities
│   │   ├── text_test.go
│   │   ├── numbers.go           # Number parsing utilities
│   │   ├── numbers_test.go
│   │   └── README.md            # Tokenizer framework documentation
│   │
```

#### Section 8: Tokenizer Framework (Lines 630-728)
**REPLACE Title:**
**OLD:** "8. Integrated Tokenizer Framework"
**NEW:** "8. Embedded Tokenization Framework"

**UPDATE First Paragraph (Lines 632-639):**
**OLD:**
```markdown
Shape includes an integrated tokenizer framework (originally from the df2-go project) that provides:
```

**NEW:**
```markdown
Shape includes an embedded tokenization framework in `internal/tokenizer/` that provides:

**Architecture Decision:** Originally developed as the df2-go project, the tokenization code has been embedded directly into shape to create a fully self-contained parser library with zero external tokenization dependencies (see ADR 0003).
```

**UPDATE Section 8.1 (Lines 641-649):**
**ADD After Line 649:**
```markdown

**Embedded Structure:**
- `internal/tokenizer/stream.go` - Stream abstraction with position tracking
- `internal/tokenizer/tokens.go` - Token struct and Tokenizer implementation
- `internal/tokenizer/matchers.go` - Matcher interface and built-in matchers
- `internal/tokenizer/text.go` - Text and rune manipulation utilities
- `internal/tokenizer/numbers.go` - Number parsing utilities
```

#### Section 8.2: Tokenizer Pattern (Lines 650-677)
**UPDATE Import Paths:**
**OLD:**
```go
import (
    "github.com/shapestone/shape/internal/streams"
    "github.com/shapestone/shape/internal/tokens"
)
```

**NEW:**
```go
import (
    "github.com/shapestone/shape/internal/tokenizer"
)
```

**UPDATE Function Signatures:**
**OLD:**
```go
func identifierMatcher(stream streams.Stream) *tokens.Token {
```

**NEW:**
```go
func identifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
```

**UPDATE Matcher List:**
**OLD:**
```go
var Matchers = []tokens.Matcher{
    tokens.CharMatcher("ObjectStart", '{'),
```

**NEW:**
```go
var Matchers = []tokenizer.Matcher{
    tokenizer.CharMatcher("ObjectStart", '{'),
```

#### Section 8.3: Parser Pattern (Lines 679-718)
**UPDATE Import:**
**OLD:**
```go
import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/internal/tokens"
)
```

**NEW:**
```go
import (
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/internal/tokenizer"
)
```

**UPDATE Struct:**
**OLD:**
```go
type Parser struct {
    tokenizer *tokens.Tokenizer
    current   *tokens.Token
    hasToken  bool
}
```

**NEW:**
```go
type Parser struct {
    tokenizer *tokenizer.Tokenizer
    current   *tokenizer.Token
    hasToken  bool
}
```

#### Appendix C: References (Lines 1097-1105)
**UPDATE:**
**OLD:**
```markdown
- **df2-go:** github.com/shapestone/df2-go (original tokenizer framework, now integrated into shape)
```

**NEW:**
```markdown
- **Embedded Tokenization:** Tokenization code embedded from df2-go project at `internal/tokenizer/` (see ADR 0003)
- **df2-go (original):** github.com/shapestone/df2-go
```

---

### 3. IMPLEMENTATION_ROADMAP.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/IMPLEMENTATION_ROADMAP.md`

#### Phase 1: Foundation (Lines 22-163)
**UPDATE Duration:**
**OLD:** 30-40 hours
**NEW:** 47-65 hours (includes tokenization migration)

**ADD New Section 1.6: Tokenization Migration (After Section 1.5, around line 163)**

```markdown
#### 1.6 Tokenization Migration (17-25 hours)

**Tasks:**
- Migrate tokenization code from df2-go to internal/tokenizer/
- Copy streams, tokens, text, numbers packages
- Update package names and imports
- Refactor and consolidate for shape's needs
- Migrate all tests
- Achieve 95%+ test coverage

**Files Created:**
```
internal/tokenizer/
├── stream.go             # Stream abstraction + patterns
├── stream_test.go
├── tokens.go             # Token + Tokenizer
├── tokens_test.go
├── matchers.go           # Matcher interface + built-ins
├── matchers_test.go
├── position.go           # Position tracking
├── text.go               # Text + rune utilities
├── text_test.go
├── numbers.go            # Number parsing
├── numbers_test.go
└── README.md             # Tokenizer documentation
```

**Success Criteria:**
- All df2-go tokenization code embedded
- All df2-go tests migrated and passing
- 95%+ test coverage
- Zero external tokenization dependencies
- Documentation complete

**Reference:** See MIGRATION_PLAN.md for detailed migration steps
```

#### Phase 1 Exit Criteria (Lines 152-163)
**ADD:**
```markdown
- [ ] Tokenization framework embedded and tested
- [ ] Zero df2-go dependency in go.mod
```

#### Phase 2: Section 2.1 (Lines 174-191)
**RENAME:**
**OLD:** "2.1 Tokenizer Framework Setup (1-2 hours)"
**NEW:** "2.1 JSONV Tokenizer Implementation (8-12 hours)"

**UPDATE Content:**
**OLD:**
```markdown
**Tasks:**
- Review integrated tokenizer framework (internal/streams, internal/tokens)
- Study existing JSON tokenizer patterns
```

**NEW:**
```markdown
**Tasks:**
- Use embedded tokenizer framework (internal/tokenizer)
- Implement JSONV-specific matchers
```

#### Dependencies Section (Lines 691-707)
**UPDATE:**
**OLD:**
```markdown
### Internal Dependencies

| Dependency | Timeline | Impact |
|------------|----------|--------|
| Tokenizer framework (integrated) | ✓ Ready | Enables all parsers |
```

**NEW:**
```markdown
### Internal Dependencies

| Dependency | Timeline | Impact |
|------------|----------|--------|
| Tokenization framework (embedded) | Phase 1 | Enables all parsers |
```

**UPDATE Note:**
**OLD:**
```markdown
**Note:** The tokenizer framework is now integrated into shape at `internal/streams/`, `internal/tokens/`, `internal/text/`, and `internal/numbers/`. No external tokenization dependencies required.
```

**NEW:**
```markdown
**Note:** The tokenization framework is embedded in shape at `internal/tokenizer/` (migrated from df2-go). See MIGRATION_PLAN.md and ADR 0003 for details. Zero external tokenization dependencies.
```

---

### 4. DATA_VALIDATOR_INTEGRATION.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/DATA_VALIDATOR_INTEGRATION.md`

#### Section: Architecture Changes (Lines 11-47)
**UPDATE After Section:**
**OLD:**
```
data-validator/ (depends on shape)
└── Traverse AST + validate data
```

**NEW:**
```
data-validator/ (depends on shape only)
└── Traverse AST + validate data

shape/ (self-contained)
├── Parse schemas → AST
└── Embedded tokenization (internal/tokenizer/)
```

#### Dependencies Section (Lines 50-79)
**UPDATE go.mod example (Lines 52-62):**
**OLD:**
```go
require (
    github.com/shapestone/shape v0.1.0  // Shape parser (self-contained)
    github.com/shapestone/wire v0.9.0   // Wire expression engine
)
```

**NEW:**
```go
require (
    github.com/shapestone/shape v0.1.0  // Self-contained parser (no df2-go dependency)
    github.com/shapestone/wire v0.9.0   // Wire expression engine
)
```

**ADD Note After Line 62:**
```markdown
**Note:** Shape v0.1.0+ includes embedded tokenization. No df2-go dependency required.
```

#### Section 9.1: Updated Architecture (Lines 722-747)
**UPDATE Note at end:**
**ADD:**
```markdown

**Note:** Shape's dependency chain is simplified:
- Before: data-validator → shape → df2-go
- After: data-validator → shape (self-contained)
```

---

### 5. SUMMARY.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/SUMMARY.md`

#### Key Decisions Section (Lines 11-60)
**UPDATE Decision 3:**
**OLD:**
```markdown
### 3. Integrated Tokenizer Framework

**Decision:** Integrate tokenizer framework for most formats

**Rationale:**
- 60-70% code reuse (40-60 hours vs 100-150 hours)
- UTF-8 and backtracking built-in
- Battle-tested, production-ready
- Self-contained framework (no external dependencies)

**Reference:** ADR 0002
```

**NEW:**
```markdown
### 3. Embedded Tokenization Framework

**Decision:** Embed tokenization code directly in shape (internal/tokenizer/)

**Rationale:**
- Self-contained library with zero tokenization dependencies
- Full control over tokenization evolution
- Simpler dependency chain for consumers
- Battle-tested code migrated from df2-go

**Reference:** ADR 0003 (supersedes ADR 0002)
```

#### Component Architecture Section (Lines 62-74)
**UPDATE:**
**OLD:**
```markdown
└── internal/parser/    # Format parsers (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV)
    ├── parser.go       # Parser interface
    ├── factory.go      # Parser factory
    └── {format}/       # Format-specific parsers
        ├── tokenizer.go
        └── parser.go
```

**NEW:**
```markdown
└── internal/
    ├── tokenizer/      # Embedded tokenization framework
    │   ├── stream.go
    │   ├── tokens.go
    │   └── matchers.go
    └── parser/         # Format parsers (JSONV, XMLV, PropsV, CSVV, YAMLV, TEXTV)
        ├── parser.go   # Parser interface
        ├── factory.go  # Parser factory
        └── {format}/   # Format-specific parsers
            ├── tokenizer.go
            └── parser.go
```

#### Dependencies Section (Lines 208-217)
**UPDATE:**
**OLD:**
```markdown
### Integrated Components
- **Tokenizer framework** - Built into shape at internal/streams, internal/tokens, internal/text (ready)
```

**NEW:**
```markdown
### Embedded Components
- **Tokenization framework** - Embedded at internal/tokenizer/ (migrated from df2-go, see ADR 0003)
```

#### Document Index (Lines 249-260)
**ADD:**
```markdown
| [ADR 0003](decisions/0003-embed-tokenizer.md) | Embed tokenization | Architects |
| [MIGRATION_PLAN.md](MIGRATION_PLAN.md) | Tokenization migration | Implementation team |
| [IMPACT_ANALYSIS.md](IMPACT_ANALYSIS.md) | Migration impact | Stakeholders |
```

---

### 6. DELIVERABLES.md

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/DELIVERABLES.md`

#### ADRs Section (Lines 199-238)
**UPDATE ADR 0002:**
**OLD:**
```markdown
### ADR 0002: Integrated Tokenizer Framework
**Status:** ✓ Complete
```

**NEW:**
```markdown
### ADR 0002: Use df2-go for Tokenization
**Status:** ⚠ Superseded by ADR 0003
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/decisions/0002-use-df2-go.md`

**Note:** This ADR is superseded by ADR 0003 (Embed Tokenization). The decision to use df2-go as an external dependency has been replaced with embedding tokenization directly in shape.
```

**ADD ADR 0003:**
```markdown
### ADR 0003: Embed Tokenization (Supersedes ADR 0002)
**Status:** ✓ Complete
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/decisions/0003-embed-tokenizer.md`

**Contents:**
- Context: Why embed instead of external df2-go dependency
- Decision: Embed tokenization code in internal/tokenizer/
- Rationale: Self-contained library, simpler dependencies, full control
- Migration strategy: Copy df2-go code, refactor, test
- Consequences: Zero external tokenization dependencies
- Impact on roadmap: Phase 1 extended by 17-25 hours

**Key Decision:** Embed tokenization directly in shape for self-contained operation
```

#### Architectural Decisions Summary (Lines 260-272)
**UPDATE:**
**OLD:**
```markdown
| Integrated tokenizer framework | Self-contained, proven, direct control | High |
```

**NEW:**
```markdown
| Embedded tokenization | Self-contained, zero deps, full control | High |
```

#### Document Structure (Lines 9-45)
**ADD:**
```markdown
│   │   │   ├── 0003-embed-tokenizer.md           ✓ Complete (Migration from df2-go)
```

**ADD:**
```markdown
│   │   ├── MIGRATION_PLAN.md                      ✓ Complete (Tokenization migration)
│   │   ├── IMPACT_ANALYSIS.md                     ✓ Complete (Migration impact)
```

---

### 7. ADR 0002 (Mark as Superseded)

**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/decisions/0002-use-df2-go.md`

**ADD at top of file (after line 1):**
```markdown
**SUPERSEDED:** This ADR is superseded by ADR 0003 (Embed Tokenization)
```

**UPDATE Status Line:**
**OLD:**
```markdown
**Status:** Accepted
```

**NEW:**
```markdown
**Status:** Superseded by ADR 0003
```

**ADD Note at end:**
```markdown
---

## Superseded Notice

**Date:** 2025-10-09  
**Superseded By:** ADR 0003 (Embed Tokenization)

**Reason:** The decision to use df2-go as an external dependency has been replaced with embedding tokenization code directly into shape at `internal/tokenizer/`. This provides:
- Zero external tokenization dependencies
- Simpler dependency chain for consumers
- Full control over tokenization evolution
- Self-contained library architecture

See ADR 0003 and MIGRATION_PLAN.md for details.
```

---

## New Documents Created

### 1. ADR 0003: Embed Tokenization
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/decisions/0003-embed-tokenizer.md`
**Status:** ✓ Created
**Size:** ~10,000 words

### 2. Migration Plan
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/MIGRATION_PLAN.md`
**Status:** ✓ Created
**Size:** ~8,000 words

### 3. Impact Analysis
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/IMPACT_ANALYSIS.md`
**Status:** ✓ Created
**Size:** ~5,000 words

### 4. Documentation Updates Guide
**Location:** `/Users/michaelsundell/Projects/shapestone/shape/docs/architecture/DOCUMENTATION_UPDATES.md`
**Status:** ✓ Created (this document)

---

## Update Checklist

- [ ] README.md - Update features and related projects
- [ ] ARCHITECTURE.md - Update diagrams, directory structure, tokenizer sections
- [ ] IMPLEMENTATION_ROADMAP.md - Add migration phase, update dependencies
- [ ] DATA_VALIDATOR_INTEGRATION.md - Update dependency chain, add notes
- [ ] SUMMARY.md - Update key decisions, component architecture
- [ ] DELIVERABLES.md - Add ADR 0003, mark ADR 0002 as superseded
- [ ] ADR 0002 - Add superseded notice
- [ ] ADR 0003 - Created ✓
- [ ] MIGRATION_PLAN.md - Created ✓
- [ ] IMPACT_ANALYSIS.md - Created ✓

---

**Document Status:** Complete Reference Guide  
**Next Step:** Apply updates to shape documentation  
**Prepared by:** System Architect  
**Date:** 2025-10-09
