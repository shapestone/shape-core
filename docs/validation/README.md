# Schema Validation Framework

Shape includes a comprehensive schema validation framework for semantic validation of AST nodes.

## Overview

The validator framework provides:
- **Type Registry** - Register and validate type names
- **Function Registry** - Register and validate function calls
- **Multi-Error Collection** - Collect all errors in one pass
- **Rich Error Formatting** - Colored terminal, plain text, and JSON output
- **Smart Hints** - "Did you mean" suggestions using Levenshtein distance

## Quick Start

```go
import (
    "fmt"
    "github.com/shapestone/shape/pkg/ast"
    "github.com/shapestone/shape/pkg/validator"
)

func main() {
    // Create a schema AST
    schema := ast.NewObjectNode(map[string]ast.SchemaNode{
        "id": ast.NewTypeNode("UUID", ast.Position{}),
        "age": ast.NewFunctionNode("Integer", []ast.SchemaNode{
            ast.NewLiteralNode(1, ast.Position{}),
            ast.NewLiteralNode(120, ast.Position{}),
        }, ast.Position{}),
    }, ast.Position{})

    // Validate the schema
    v := validator.NewSchemaValidator()
    result := v.ValidateAll(schema, "")

    if !result.Valid {
        fmt.Println(result.FormatColored())
    }
}
```

## What is Validated

The validator checks the following aspects of schema ASTs:

### 1. Type References

All type names must be registered. Built-in types include:

- `UUID` - Universally unique identifier
- `Email` - Email address
- `URL` - Web URL
- `String` - Text string
- `Integer` - Whole number
- `Float` - Floating point number
- `Boolean` - True/false value
- `ISO-8601` - ISO-8601 date/time format
- `Date` - Date value
- `Time` - Time value
- `DateTime` - Date and time value
- `IPv4` - IPv4 address
- `IPv6` - IPv6 address
- `JSON` - JSON data
- `Base64` - Base64 encoded data

**Example Error:**
```go
schema := ast.NewObjectNode(map[string]ast.SchemaNode{
    "country": ast.NewTypeNode("CountryCode", ast.Position{}),  // CountryCode not registered
}, ast.Position{})
```

### 2. Function References

All function names must be registered. Built-in functions include:

- `String(min, max)` - String length constraints
- `Integer(min, max)` - Integer range constraints
- `Float(min, max)` - Float range constraints
- `Enum(val1, val2, ...)` - Enumeration values
- `Pattern(regex)` - Regular expression pattern
- `Length(min, max)` - Generic length constraint
- `Range(min, max)` - Generic range constraint

**Example Error:**
```go
schema := ast.NewObjectNode(map[string]ast.SchemaNode{
    "id": ast.NewFunctionNode("NotAFunction", []ast.SchemaNode{
        ast.NewLiteralNode(1, ast.Position{}),
    }, ast.Position{}),  // NotAFunction not registered
}, ast.Position{})
```

### 3. Function Arguments

Argument counts must match function definitions.

**Example Errors:**
```go
// Too few arguments
ast.NewFunctionNode("String", []ast.SchemaNode{}, ast.Position{})  // min: 1

// Too many arguments
ast.NewFunctionNode("Integer", []ast.SchemaNode{
    ast.NewLiteralNode(1, ast.Position{}),
    ast.NewLiteralNode(100, ast.Position{}),
    ast.NewLiteralNode(200, ast.Position{}),
}, ast.Position{})  // max: 2
```

### 4. Argument Values

For range functions, min must be ≤ max.

**Example Error:**
```go
ast.NewFunctionNode("Integer", []ast.SchemaNode{
    ast.NewLiteralNode(100, ast.Position{}),  // min
    ast.NewLiteralNode(1, ast.Position{}),    // max - invalid!
}, ast.Position{})
```

## Output Formats

The validator provides three output formats:

### Colored Terminal Output (Default)

Best for interactive terminal use with syntax highlighting:

```go
result := v.ValidateAll(schema, sourceText)
fmt.Println(result.FormatColored())
```

Output:
```
Found 2 validation errors:

Error 1:
Line 1, Column 37 ($.badargs)
ERROR [INVALID_ARG_COUNT]: Integer accepts at most 2 arguments, got 3

HINT: Expected Integer to have between 1 and 2 arguments

---

Error 2:
Line 1, Column 13 ($.unknown)
ERROR [UNKNOWN_TYPE]: unknown type: UnknownType

HINT: Available types include: Base64, Boolean, Date, DateTime, Email, Float, ...
```

Colors automatically disabled when:
- `NO_COLOR` environment variable is set
- Output is redirected to a file
- Terminal doesn't support ANSI colors

### Plain Text Output

Best for log files and non-terminal output:

```go
fmt.Println(result.FormatPlain())
```

### JSON Output

Best for programmatic use and integration with other tools:

```go
jsonBytes, _ := result.ToJSON()
fmt.Println(string(jsonBytes))
```

Output:
```json
{
  "valid": false,
  "errorCount": 2,
  "errors": [
    {
      "Position": {"Offset": 12, "Line": 1, "Column": 13},
      "Path": "$.unknown",
      "Code": "UNKNOWN_TYPE",
      "Message": "unknown type: UnknownType",
      "Hint": "Available types include: Base64, Boolean, Date, ..."
    }
  ]
}
```

## Custom Types and Functions

Register custom types and functions for domain-specific validation:

```go
import "github.com/shapestone/shape/pkg/validator"

v := validator.NewSchemaValidator()

// Register custom types
v.RegisterType("SSN", validator.TypeDescriptor{
    Name:        "SSN",
    Description: "Social Security Number",
})

v.RegisterType("PhoneNumber", validator.TypeDescriptor{
    Name:        "PhoneNumber",
    Description: "US Phone Number",
})

// Register custom functions
v.RegisterFunction("Luhn", validator.FunctionDescriptor{
    Name:        "Luhn",
    Description: "Luhn checksum validation",
    MinArgs:     0,
    MaxArgs:     0,
})

// Create schema using custom types
schema := ast.NewObjectNode(map[string]ast.SchemaNode{
    "ssn": ast.NewTypeNode("SSN", ast.Position{}),
    "phone": ast.NewTypeNode("PhoneNumber", ast.Position{}),
    "card": ast.NewFunctionNode("Luhn", []ast.SchemaNode{}, ast.Position{}),
}, ast.Position{})

result := v.ValidateAll(schema, "")
```

## Performance

Validation is designed to be fast with minimal overhead:

**Benchmark Results (Apple M1 Max):**

- Simple schema (2 properties): ~2.2 µs
- Complex schema (nested, 7 properties): ~3.1 µs
- With errors (4 errors): ~27 µs

**All validation operations complete well under 1 millisecond.**

Run benchmarks:
```bash
go test -bench=. -benchmem ./pkg/validator/
```

## Error Codes

See [error-codes.md](error-codes.md) for a complete reference of all error codes.

## API Reference

### Validation Functions

```go
// Create validator
v := validator.NewSchemaValidator()

// Validate AST
result := v.ValidateAll(node ast.SchemaNode, sourceText ...string)
```

The `sourceText` parameter is optional but recommended for better error messages with source context.

### ValidationResult Methods

```go
// Check if validation passed
if result.Valid { ... }

// Get error count
count := result.ErrorCount()

// Get all errors
errors := result.GetErrors()

// Get first error (or nil)
firstErr := result.FirstError()

// Get errors by code
unknownTypeErrors := result.ErrorsByCode(validator.ErrCodeUnknownType)

// Format output
colored := result.FormatColored()
plain := result.FormatPlain()
jsonBytes, _ := result.ToJSON()
```

### SchemaValidator Methods

```go
v := validator.NewSchemaValidator()

// Register types
v.RegisterType("CustomType", validator.TypeDescriptor{...})

// Register functions
v.RegisterFunction("CustomFunc", validator.FunctionDescriptor{...})

// Validate
result := v.ValidateAll(ast, sourceText)
```

## See Also

- [Error Codes Reference](error-codes.md)
- [Main README](../../README.md)
- [Architecture Documentation](../architecture/ARCHITECTURE.md)
