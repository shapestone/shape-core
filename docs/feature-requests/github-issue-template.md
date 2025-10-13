# GitHub Issue Template: Semantic Schema Validation

**Copy this template to create a GitHub issue in the Shape repository**

---

## Title
Add Semantic Schema Validation to Shape Parser

## Labels
`enhancement`, `feature-request`, `validation`, `high-priority`

## Description

### Summary
Shape currently validates syntax but not semantics. We need semantic validation to check if schemas are well-formed (valid types, functions, parameters).

### Problem
Shape successfully parses this schema:
```jsonv
{
    "country": CountryCode,
    "age": Integer(1, 100, 200)
}
```

But it doesn't catch:
- `CountryCode` - unknown type ❌
- `Integer(1, 100, 200)` - wrong argument count (needs 2, got 3) ❌

This forces every Shape consumer to implement their own validation, leading to:
- Duplicated validation logic
- Inconsistent error messages
- Late error detection
- Reduced reusability

### Proposed Solution
Add optional semantic validation to Shape:

```go
// Parse (existing behavior - unchanged)
ast, err := shape.Parse(parser.FormatJSONV, schema)

// Validate AST (new functionality)
result, err := shape.ValidateAST(ast, validator.DefaultValidator)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("%s at %s (line %d)\n", err.Message, err.Path, err.Line)
    }
}
```

**Output:**
```
Integer function expects 2 arguments (min, max), got 3 at $.age (line 3)
Unknown type identifier: 'CountryCode' at $.country (line 2)
```

### Benefits
1. **For Shape**: More complete parser, better error messages, standardization
2. **For Consumers**: Less code, consistent errors, early detection
3. **For End Users**: Better DX, faster feedback, helpful hints

### Requirements

**Type Registry:**
- Built-in types: UUID, Email, ISO-8601, URL, String, Integer, Float, Boolean
- Extension API for custom types

**Function Registry:**
- Built-in functions: Integer(min, max), Float(min, max), String(minLen, maxLen), Enum(...)
- Extension API for custom functions

**Validation:**
- Validate type identifiers
- Validate function names
- Validate parameter counts and types
- Detailed error messages with hints
- JSONPath to errors

**Error Codes:**
- `UNKNOWN_TYPE`
- `UNKNOWN_FUNCTION`
- `INVALID_ARG_COUNT`
- `INVALID_ARG_TYPE`

### API Design Options

**Option 1: Separate Function (Recommended)**
```go
result, err := shape.ValidateAST(ast, validator.DefaultValidator)
```
✅ Backwards compatible, opt-in

**Option 2: Parse + Validate**
```go
ast, result, err := shape.ParseAndValidate(format, schema, validator)
```
✅ Convenient one-step

**Option 3: Configurable Parser**
```go
p := shape.NewParser(shape.WithValidation(validator))
ast, result, err := p.Parse(schema)
```
✅ Most flexible

### Implementation Plan

**Phase 1: Core (Target: v0.3.0)**
- Week 1-2: Type & Function registries with built-ins
- Week 3: SchemaValidator with AST traversal
- Week 4: Public API, tests, docs

**Phase 2: Advanced (v0.4.0)**
- Circular reference detection
- Schema complexity analysis
- Validation caching

**Phase 3: Format-Specific (v0.5.0)**
- Format-specific validation rules
- Cross-format consistency

### Backwards Compatibility
✅ **No breaking changes** - validation is opt-in

Existing code continues to work:
```go
ast, err := shape.Parse(parser.FormatJSONV, schema) // unchanged
```

New validation is optional:
```go
result, err := shape.ValidateAST(ast, validator) // new, opt-in
```

### References
- [Detailed Feature Request](./semantic-schema-validation.md)
- [Use Case: data-validator library](https://github.com/shapestone/data-validator)
- [JSONV Format Spec](../specifications/jsonv-format-spec.md)

### Request for Comments

**Questions for maintainers:**
1. Which API design option do you prefer (1, 2, or 3)?
2. Should default validators be global or per-parser?
3. Should validation be opt-in or opt-out?
4. Any concerns about the proposed registries?

**Alternative Approaches:**
We considered a separate `schema-validator` library but rejected it because validation is tightly coupled to the Shape AST.

### Next Steps
1. ✅ Create feature request document
2. ⏳ Review by Shape maintainers
3. ⏳ Discuss architecture decisions
4. ⏳ Approve Phase 1 scope
5. ⏳ Implement and release v0.3.0

---

**Requested by**: @[your-github-username] (data-validator team)
**Priority**: High (impacts multiple Shape consumers)
**Complexity**: Medium-High (3-4 weeks for Phase 1)

cc: @shapestone-maintainers
