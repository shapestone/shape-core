# Semantic Validation

Shape v0.3.0+ includes comprehensive semantic validation to catch schema errors early.

## Quick Start

```go
import "github.com/shapestone/shape/pkg/shape"
import "github.com/shapestone/shape/internal/parser"

schema := `{"id": UUID, "age": Integer(1, 120)}`
ast, err := shape.Parse(parser.FormatJSONV, schema)
if err != nil {
    log.Fatal(err)
}

result := shape.ValidateAll(ast, schema)
if !result.Valid {
    fmt.Println(result.FormatColored())
}
```

## What is Validated

Shape validates the following aspects of your schemas:

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
schema := `{"country": CountryCode}`  // CountryCode is not registered
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
schema := `{"id": NotAFunction(1, 2)}`  // NotAFunction is not registered
```

### 3. Function Arguments

Argument counts must match function definitions.

**Example Errors:**
```go
schema := `{"name": String()}`           // Too few args (min: 1)
schema := `{"age": Integer(1, 100, 200)}`  // Too many args (max: 2)
```

### 4. Argument Values

For range functions, min must be ≤ max.

**Example Error:**
```go
schema := `{"age": Integer(100, 1)}`  // min > max
```

## Output Formats

Shape provides three output formats for validation results:

### Colored Terminal Output (Default)

Best for interactive terminal use with syntax highlighting:

```go
fmt.Println(result.FormatColored())
```

Output:
```
Found 2 validation errors:

Error 1:
Line 1, Column 37 ($.badargs)
ERROR [INVALID_ARG_COUNT]: Integer accepts at most 2 arguments, got 3

  >  1 | {"unknown": UnknownType, "badargs": Integer(1, 100, 200)}
      |                                     ^

HINT: Expected Integer to have between 1 and 2 arguments

---

Error 2:
Line 1, Column 13 ($.unknown)
ERROR [UNKNOWN_TYPE]: unknown type: UnknownType

  >  1 | {"unknown": UnknownType, "badargs": Integer(1, 100, 200)}
      |             ^

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

Output:
```
Found 2 validation errors:

Error 1:
Line 1, Column 37 ($.badargs)
ERROR [INVALID_ARG_COUNT]: Integer accepts at most 2 arguments, got 3

  >  1 | {"unknown": UnknownType, "badargs": Integer(1, 100, 200)}
      |                                     ^

HINT: Expected Integer to have between 1 and 2 arguments

---

Error 2:
Line 1, Column 13 ($.unknown)
ERROR [UNKNOWN_TYPE]: unknown type: UnknownType

  >  1 | {"unknown": UnknownType, "badargs": Integer(1, 100, 200)}
      |             ^

HINT: Available types include: Base64, Boolean, Date, DateTime, Email, Float, ...
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
      "Position": {
        "Offset": 12,
        "Line": 1,
        "Column": 13
      },
      "Path": "$.unknown",
      "Code": "UNKNOWN_TYPE",
      "Message": "unknown type: UnknownType",
      "Hint": "Available types include: Base64, Boolean, Date, DateTime, Email, Float, ...",
      "Source": "{\"unknown\": UnknownType, \"badargs\": Integer(1, 100, 200)}",
      "SourceLines": [
        "{\"unknown\": UnknownType, \"badargs\": Integer(1, 100, 200)}"
      ]
    },
    {
      "Position": {
        "Offset": 36,
        "Line": 1,
        "Column": 37
      },
      "Path": "$.badargs",
      "Code": "INVALID_ARG_COUNT",
      "Message": "Integer accepts at most 2 arguments, got 3",
      "Hint": "Expected Integer to have between 1 and 2 arguments",
      "Source": "{\"unknown\": UnknownType, \"badargs\": Integer(1, 100, 200)}",
      "SourceLines": [
        "{\"unknown\": UnknownType, \"badargs\": Integer(1, 100, 200)}"
      ]
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

// Use custom types in schemas
schema := `{"ssn": SSN, "phone": PhoneNumber, "card": Luhn()}`
ast, _ := shape.Parse(parser.FormatJSONV, schema)
result := v.ValidateAll(ast, schema)
```

## CLI Tool

Shape includes a command-line tool for validating schema files.

### Installation

```bash
go install github.com/shapestone/shape/cmd/shape-validate@latest
```

### Basic Usage

```bash
# Validate a single file
shape-validate schema.jsonv

# Validate multiple files
shape-validate schema1.jsonv schema2.xmlv schema3.yamlv
```

### Flags

```
-f, --format string       Schema format (jsonv, xmlv, yamlv, csvv, propsv, textv, auto) [default: auto]
-o, --output string       Output format (text, json, quiet) [default: text]
--no-color                Disable colored output
--register-type string    Register custom types (comma-separated)
-v, --verbose             Verbose output
--version                 Show version
```

### Examples

```bash
# JSON output
shape-validate -o json schema.jsonv

# Register custom types
shape-validate --register-type SSN,CreditCard,PhoneNumber schema.jsonv

# Quiet mode (exit code only)
shape-validate -o quiet schema.jsonv && echo "Valid!" || echo "Invalid!"

# Verbose mode
shape-validate -v schema.jsonv

# Disable colors
shape-validate --no-color schema.jsonv

# Specify format explicitly
shape-validate -f yamlv config.yml
```

### Exit Codes

- `0` - Schema is valid
- `1` - Schema has validation errors
- `2` - Parse error (syntax error)
- `3` - File not found or I/O error

### CI/CD Integration

```bash
# In your CI pipeline
shape-validate -o quiet schemas/*.jsonv || exit 1
```

## Performance

Validation is designed to be fast with minimal overhead:

**Benchmark Results (Apple M1 Max):**

- Simple schema (2 properties): ~1.9 µs
- Complex schema (nested, 7 properties): ~2.8 µs
- Deep nesting (5 levels): ~2.4 µs
- With errors (4 errors): ~24.7 µs

**All validation operations complete in under 1 millisecond.**

Run benchmarks:
```bash
go test -bench=. -benchmem ./pkg/validator/
```

## Error Codes

See [error-codes.md](error-codes.md) for a complete reference of all error codes.

## Examples

See [examples.md](examples.md) for more usage examples.

## API Reference

### ValidateAll Function

```go
func ValidateAll(node ast.SchemaNode, sourceText ...string) *validator.ValidationResult
```

Validates a schema AST and returns all errors found. The `sourceText` parameter is optional but recommended for better error messages with source context.

**Returns:** `*ValidationResult` with:
- `Valid bool` - True if no errors found
- `Errors []ValidationError` - All errors found

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
- [Usage Examples](examples.md)
- [Main README](../../README.md)
