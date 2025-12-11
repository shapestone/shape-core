# Validation Error Codes

Complete reference for all semantic validation error codes in Shape.

## Error Code Format

Error codes follow the pattern: `ERROR [CODE]: message`

Example:
```
ERROR [UNKNOWN_TYPE]: unknown type: CountryCode
```

## Error Codes

### UNKNOWN_TYPE

**Description:** A type reference is not registered in the validator.

**When it occurs:**
- Schema references a type that doesn't exist
- Type name is misspelled
- Custom type not registered

**Example:**
```go
{"country": CountryCode}  // CountryCode is not a built-in type
```

**Error Message:**
```
ERROR [UNKNOWN_TYPE]: unknown type: CountryCode
```

**Hint:**
- Suggests similar type names using Levenshtein distance
- Lists available types if no close match found

**Resolution:**
1. Check if type name is misspelled
2. Use a built-in type (UUID, Email, String, etc.)
3. Register custom type:
   ```go
   v := validator.NewSchemaValidator()
   v.RegisterType("CountryCode", validator.TypeDescriptor{
       Name:        "CountryCode",
       Description: "ISO 3166 country code",
   })
   ```

---

### UNKNOWN_FUNCTION

**Description:** A function reference is not registered in the validator.

**When it occurs:**
- Schema calls a function that doesn't exist
- Function name is misspelled
- Custom function not registered

**Example:**
```go
{"id": NotAFunction(1, 2)}  // NotAFunction is not registered
```

**Error Message:**
```
ERROR [UNKNOWN_FUNCTION]: unknown function: NotAFunction
```

**Hint:**
- Suggests similar function names using Levenshtein distance
- Lists available functions if no close match found

**Resolution:**
1. Check if function name is misspelled
2. Use a built-in function (String, Integer, Enum, etc.)
3. Register custom function:
   ```go
   v := validator.NewSchemaValidator()
   v.RegisterFunction("NotAFunction", validator.FunctionDescriptor{
       Name:        "NotAFunction",
       Description: "Custom function",
       MinArgs:     2,
       MaxArgs:     2,
   })
   ```

---

### INVALID_ARG_COUNT

**Description:** Function called with wrong number of arguments.

**When it occurs:**
- Too few arguments provided
- Too many arguments provided

**Example (Too Few):**
```go
{"name": String()}  // String requires at least 1 argument
```

**Error Message:**
```
ERROR [INVALID_ARG_COUNT]: String requires at least 1 arguments, got 0
```

**Example (Too Many):**
```go
{"age": Integer(1, 100, 200)}  // Integer accepts at most 2 arguments
```

**Error Message:**
```
ERROR [INVALID_ARG_COUNT]: Integer accepts at most 2 arguments, got 3
```

**Hint:**
```
Expected Integer to have between 1 and 2 arguments
```

**Resolution:**
1. Check function signature
2. Provide correct number of arguments
3. For custom functions, adjust `MinArgs` and `MaxArgs` when registering

**Built-in Function Arguments:**

| Function | Min Args | Max Args | Example |
|----------|----------|----------|---------|
| String   | 1        | 2        | `String(1, 100)` |
| Integer  | 1        | 2        | `Integer(1, 100)` |
| Float    | 1        | 2        | `Float(0.0, 100.0)` |
| Enum     | 1        | unlimited | `Enum("red", "green", "blue")` |
| Pattern  | 1        | 1        | `Pattern("^[A-Z]+$")` |
| Length   | 1        | 2        | `Length(1, 100)` |
| Range    | 1        | 2        | `Range(1, 100)` |

---

### INVALID_ARG_VALUE

**Description:** Function argument value is invalid.

**When it occurs:**
- min > max in range functions
- Argument type is incorrect
- Custom validation fails

**Example:**
```go
{"age": Integer(100, 1)}  // min (100) > max (1)
```

**Error Message:**
```
ERROR [INVALID_ARG_VALUE]: Integer: min (100) cannot be greater than max (1)
```

**Hint:**
```
Check the argument types and values
```

**Resolution:**
1. Ensure min ≤ max for range functions
2. Check argument types (e.g., Pattern expects string)
3. For custom functions, fix validation logic

---

## Error Structure

Each validation error includes:

```go
type ValidationError struct {
    Position    Position      // Line, column, offset in source
    Path        string        // JSONPath to error location (e.g., "$.user.age")
    Code        ErrorCode     // Error code (UNKNOWN_TYPE, etc.)
    Message     string        // Human-readable error message
    Hint        string        // Suggestion for fixing the error
    Source      string        // Full source text (optional)
    SourceLines []string      // Lines around error for context (optional)
}
```

### Position

```go
type Position struct {
    Offset int  // Byte offset in source
    Line   int  // Line number (1-indexed)
    Column int  // Column number (1-indexed)
}
```

### JSONPath

The `Path` field uses JSONPath notation to indicate where the error occurred:

- `$` - Root
- `$.user` - Property "user" in root object
- `$.user.age` - Property "age" in nested "user" object
- `$.tags[]` - Array elements in "tags" property
- `$.users[].profile` - Property "profile" in array elements

## Multiple Errors

`ValidateAll()` returns ALL errors found, not just the first one:

```go
schema := `{
    "unknown1": UnknownType,
    "unknown2": AnotherUnknown,
    "badArgs": Integer(1, 100, 200),
    "badFunc": NotAFunction(1, 2)
}`

result := shape.ValidateAll(ast, schema)
// result.ErrorCount() == 4
```

## Error Hints

Shape provides helpful hints for each error:

### Type Hints

- **Close match found:** `Did you mean 'Email'?` (for 'Emai1')
- **No close match:** `Available types include: Base64, Boolean, Date, ...`

### Function Hints

- **Close match found:** `Did you mean 'Integer'?` (for 'Intger')
- **No close match:** `Available functions include: String, Integer, Float, ...`

### Argument Hints

- **Invalid count:** `Expected Integer to have between 1 and 2 arguments`
- **Invalid value:** `Check the argument types and values`

## Programmatic Access

```go
// Check specific error codes
for _, err := range result.GetErrors() {
    switch err.Code {
    case validator.ErrCodeUnknownType:
        // Handle unknown type
    case validator.ErrCodeUnknownFunction:
        // Handle unknown function
    case validator.ErrCodeInvalidArgCount:
        // Handle invalid arg count
    case validator.ErrCodeInvalidArgValue:
        // Handle invalid arg value
    }
}

// Filter errors by code
unknownTypes := result.ErrorsByCode(validator.ErrCodeUnknownType)
fmt.Printf("Found %d unknown types\n", len(unknownTypes))
```

## See Also

- [Validation README](README.md)
- [Usage Examples](examples.md)
- [Main README](../../README.md)
