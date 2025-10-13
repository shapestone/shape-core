# Validation Examples

Practical examples of using Shape's semantic validation features.

## Basic Validation

### Example 1: Simple Valid Schema

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{
        "id": UUID,
        "name": String(1, 100),
        "age": Integer(18, 120)
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("Schema is valid!")
    }
}
```

### Example 2: Schema with Errors

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{
        "unknown": UnknownType,
        "badArgs": Integer(1, 100, 200)
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if !result.Valid {
        // Print colored output
        fmt.Println(result.FormatColored())

        // Or access errors programmatically
        for i, err := range result.GetErrors() {
            fmt.Printf("Error %d: %s at %s\n", i+1, err.Message, err.Path)
        }
    }
}
```

## Custom Types

### Example 3: Register Custom Types

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
    "github.com/shapestone/shape/pkg/validator"
)

func main() {
    // Create validator and register custom types
    v := validator.NewSchemaValidator()
    v.RegisterType("SSN", validator.TypeDescriptor{
        Name:        "SSN",
        Description: "Social Security Number (XXX-XX-XXXX)",
    })
    v.RegisterType("PhoneNumber", validator.TypeDescriptor{
        Name:        "PhoneNumber",
        Description: "US Phone Number",
    })
    v.RegisterType("ZipCode", validator.TypeDescriptor{
        Name:        "ZipCode",
        Description: "US Zip Code",
    })

    schema := `{
        "ssn": SSN,
        "phone": PhoneNumber,
        "zip": ZipCode
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := v.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("Schema is valid with custom types!")
    }
}
```

## Custom Functions

### Example 4: Register Custom Functions

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
    "github.com/shapestone/shape/pkg/validator"
)

func main() {
    v := validator.NewSchemaValidator()

    // Register custom function with no arguments
    v.RegisterFunction("Luhn", validator.FunctionDescriptor{
        Name:        "Luhn",
        Description: "Luhn checksum validation",
        MinArgs:     0,
        MaxArgs:     0,
    })

    // Register custom function with arguments
    v.RegisterFunction("Between", validator.FunctionDescriptor{
        Name:        "Between",
        Description: "Value must be between two dates",
        MinArgs:     2,
        MaxArgs:     2,
    })

    schema := `{
        "creditCard": Luhn(),
        "created": Between("2020-01-01", "2025-12-31")
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := v.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("Schema is valid with custom functions!")
    }
}
```

## Different Output Formats

### Example 5: Plain Text Output

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{"unknown": UnknownType}`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if !result.Valid {
        // Write plain text to log file
        f, _ := os.Create("validation.log")
        defer f.Close()
        f.WriteString(result.FormatPlain())

        fmt.Println("Validation errors written to validation.log")
    }
}
```

### Example 6: JSON Output

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{"unknown": UnknownType}`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)

    // Get JSON output
    jsonBytes, err := result.ToJSON()
    if err != nil {
        log.Fatal(err)
    }

    // Parse and use programmatically
    var jsonResult struct {
        Valid      bool   `json:"valid"`
        ErrorCount int    `json:"errorCount"`
        Errors     []struct {
            Code    string `json:"Code"`
            Message string `json:"Message"`
            Path    string `json:"Path"`
        } `json:"errors"`
    }

    json.Unmarshal(jsonBytes, &jsonResult)

    fmt.Printf("Valid: %v\n", jsonResult.Valid)
    fmt.Printf("Errors: %d\n", jsonResult.ErrorCount)
    for _, err := range jsonResult.Errors {
        fmt.Printf("  - [%s] %s at %s\n", err.Code, err.Message, err.Path)
    }
}
```

## Multiple Formats

### Example 7: Validate YAMLV Schema

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `
user:
  id: UUID
  name: String(1, 100)
  email: Email
  profile:
    bio: String(0, 500)
    avatar: URL
`

    ast, err := shape.Parse(parser.FormatYAMLV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("YAMLV schema is valid!")
    } else {
        fmt.Println(result.FormatColored())
    }
}
```

### Example 8: Validate XMLV Schema

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `
<user>
    <id>UUID</id>
    <name>String(1, 100)</name>
    <email>Email</email>
    <age>Integer(18, 120)</age>
</user>
`

    ast, err := shape.Parse(parser.FormatXMLV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("XMLV schema is valid!")
    } else {
        fmt.Println(result.FormatColored())
    }
}
```

## Error Handling

### Example 9: Filter Errors by Code

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
    "github.com/shapestone/shape/pkg/validator"
)

func main() {
    schema := `{
        "unknown1": UnknownType,
        "unknown2": AnotherUnknown,
        "badArgs": Integer(1, 100, 200)
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)

    // Filter errors by code
    unknownTypes := result.ErrorsByCode(validator.ErrCodeUnknownType)
    argCountErrors := result.ErrorsByCode(validator.ErrCodeInvalidArgCount)

    fmt.Printf("Unknown types: %d\n", len(unknownTypes))
    for _, err := range unknownTypes {
        fmt.Printf("  - %s at %s\n", err.Message, err.Path)
    }

    fmt.Printf("Invalid arg count: %d\n", len(argCountErrors))
    for _, err := range argCountErrors {
        fmt.Printf("  - %s at %s\n", err.Message, err.Path)
    }
}
```

### Example 10: Get First Error Only

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{
        "unknown1": UnknownType,
        "unknown2": AnotherUnknown
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)

    // Get only the first error
    if firstErr := result.FirstError(); firstErr != nil {
        fmt.Printf("First error: %s at %s\n", firstErr.Message, firstErr.Path)
    }
}
```

## Complex Schemas

### Example 11: Nested Objects and Arrays

```go
package main

import (
    "fmt"
    "log"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    schema := `{
        "user": {
            "id": UUID,
            "username": String(3, 20),
            "email": Email,
            "profile": {
                "bio": String(0, 500),
                "avatar": URL,
                "location": {
                    "city": String(1, 100),
                    "country": String(2, 100)
                }
            },
            "tags": [String(1, 50)],
            "friends": [{
                "id": UUID,
                "name": String(1, 100)
            }]
        }
    }`

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatal(err)
    }

    result := shape.ValidateAll(ast, schema)
    if result.Valid {
        fmt.Println("Complex nested schema is valid!")
    }
}
```

## Integration with File System

### Example 12: Validate Files

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func main() {
    // Read schema from file
    content, err := os.ReadFile("schema.jsonv")
    if err != nil {
        log.Fatal(err)
    }

    schema := string(content)

    ast, err := shape.Parse(parser.FormatJSONV, schema)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }

    result := shape.ValidateAll(ast, schema)
    if !result.Valid {
        fmt.Println(result.FormatColored())
        os.Exit(1)
    }

    fmt.Println("Schema is valid!")
}
```

## CLI Usage Examples

### Example 13: Basic CLI Usage

```bash
# Validate a single file
shape-validate schema.jsonv

# Validate multiple files
shape-validate schemas/*.jsonv

# Validate with specific format
shape-validate -f yamlv config.yml
```

### Example 14: CI/CD Pipeline

```bash
#!/bin/bash

# Validate all schemas in CI pipeline
shape-validate -o quiet schemas/*.jsonv
if [ $? -ne 0 ]; then
    echo "Schema validation failed!"
    exit 1
fi

echo "All schemas valid!"
```

### Example 15: JSON Output for Tools

```bash
# Generate JSON report
shape-validate -o json schema.jsonv > validation-report.json

# Parse JSON with jq
cat validation-report.json | jq '.errorCount'
cat validation-report.json | jq '.errors[] | "\(.Code): \(.Message)"'
```

## Performance Testing

### Example 16: Benchmark Validation

```go
package main

import (
    "fmt"
    "testing"

    "github.com/shapestone/shape/internal/parser"
    "github.com/shapestone/shape/pkg/shape"
)

func BenchmarkValidation(b *testing.B) {
    schema := `{
        "id": UUID,
        "name": String(1, 100),
        "email": Email,
        "age": Integer(18, 120)
    }`

    ast, _ := shape.Parse(parser.FormatJSONV, schema)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := shape.ValidateAll(ast, schema)
        if !result.Valid {
            b.Fatal("expected valid")
        }
    }
}

func main() {
    result := testing.Benchmark(BenchmarkValidation)
    fmt.Printf("Validation took: %v per operation\n", result.NsPerOp())
}
```

## See Also

- [Validation README](README.md)
- [Error Codes Reference](error-codes.md)
- [Main README](../../README.md)
