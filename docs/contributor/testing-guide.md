# Testing Guide

This guide covers testing practices, standards, and workflows for Shape contributors.

## Testing Philosophy

Shape maintains high test coverage (95%+) through:

1. **Unit Tests** - Test individual functions and components in isolation
2. **Integration Tests** - Test interactions between components
3. **Benchmark Tests** - Measure and track performance
4. **Table-Driven Tests** - Parameterized test cases for comprehensive coverage
5. **Property-Based Testing** - When appropriate for validation logic

## Test Coverage Requirements

### Coverage Targets

- **Overall:** 95%+ coverage
- **New Code:** 100% coverage for new features
- **Bug Fixes:** Tests that reproduce the bug
- **Public API:** 100% coverage for exported functions

### Checking Coverage

```bash
# Overall coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./pkg/ast/
go test -cover ./pkg/tokenizer/
go test -cover ./internal/parser/jsonv/
```

## Writing Tests

### Test File Organization

Tests live alongside the code they test:

```
pkg/ast/
├── node.go         # Implementation
├── node_test.go    # Tests
├── visitor.go      # Implementation
└── visitor_test.go # Tests
```

### Test Function Naming

```go
// Format: TestFunctionName_Scenario
func TestParse_ValidInput(t *testing.T) { }
func TestParse_InvalidSyntax(t *testing.T) { }
func TestParse_EmptyInput(t *testing.T) { }

// Format: TestStructName_MethodName_Scenario
func TestTokenizer_Next_ValidToken(t *testing.T) { }
func TestTokenizer_Next_EOF(t *testing.T) { }
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    ast.SchemaNode
        wantErr bool
    }{
        {
            name:  "valid object",
            input: `{"id": UUID}`,
            want:  ast.NewObjectNode(map[string]ast.SchemaNode{
                "id": ast.NewTypeNode("UUID"),
            }),
            wantErr: false,
        },
        {
            name:    "invalid syntax",
            input:   `{"id": }`,
            want:    nil,
            wantErr: true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(parser.FormatJSONV, tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Testing Error Cases

Always test error paths:

```go
func TestParse_ErrorCases(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantErrMsg  string
    }{
        {
            name:       "unclosed object",
            input:      `{"id": UUID`,
            wantErrMsg: "expected '}'",
        },
        {
            name:       "invalid token",
            input:      `{"id": @@@}`,
            wantErrMsg: "unexpected token",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Parse(parser.FormatJSONV, tt.input)

            if err == nil {
                t.Fatal("expected error, got nil")
            }

            if !strings.Contains(err.Error(), tt.wantErrMsg) {
                t.Errorf("expected error containing %q, got %q",
                    tt.wantErrMsg, err.Error())
            }
        })
    }
}
```

### Testing with Subtests

Use subtests for logical grouping:

```go
func TestValidator(t *testing.T) {
    t.Run("type validation", func(t *testing.T) {
        t.Run("known types", func(t *testing.T) {
            // Test known types
        })

        t.Run("unknown types", func(t *testing.T) {
            // Test unknown types
        })
    })

    t.Run("function validation", func(t *testing.T) {
        t.Run("valid arguments", func(t *testing.T) {
            // Test valid arguments
        })

        t.Run("invalid arguments", func(t *testing.T) {
            // Test invalid arguments
        })
    })
}
```

## Testing Specific Components

### Testing Parsers

```go
func TestJSONVParser_ParseObject(t *testing.T) {
    parser := NewParser()

    tests := []struct {
        name    string
        input   string
        want    *ast.ObjectNode
        wantErr bool
    }{
        {
            name:  "simple object",
            input: `{"id": UUID}`,
            want: ast.NewObjectNode(map[string]ast.SchemaNode{
                "id": ast.NewTypeNode("UUID"),
            }),
            wantErr: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.Parse(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if err == nil && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Testing AST Nodes

```go
func TestTypeNode(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        node := ast.NewTypeNode("UUID")

        if node.TypeName != "UUID" {
            t.Errorf("TypeName = %q, want %q", node.TypeName, "UUID")
        }

        if node.Type() != ast.NodeTypeType {
            t.Errorf("Type() = %v, want %v", node.Type(), ast.NodeTypeType)
        }
    })

    t.Run("visitor", func(t *testing.T) {
        node := ast.NewTypeNode("UUID")
        visitor := &mockVisitor{}

        err := node.Accept(visitor)
        if err != nil {
            t.Fatalf("Accept() error = %v", err)
        }

        if !visitor.visitedType {
            t.Error("Visitor.VisitType() was not called")
        }
    })
}
```

### Testing Validators

```go
func TestSchemaValidator_ValidateType(t *testing.T) {
    v := validator.NewSchemaValidator()

    tests := []struct {
        name     string
        typeName string
        wantErr  bool
    }{
        {name: "UUID", typeName: "UUID", wantErr: false},
        {name: "Email", typeName: "Email", wantErr: false},
        {name: "Unknown", typeName: "UnknownType", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            node := ast.NewTypeNode(tt.typeName)
            err := v.VisitType(node)

            if (err != nil) != tt.wantErr {
                t.Errorf("VisitType() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing Tokenizers

```go
func TestTokenizer_Tokenize(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        wantTokens []tokenizer.Token
        wantErr    bool
    }{
        {
            name:  "simple tokens",
            input: `{ "id": UUID }`,
            wantTokens: []tokenizer.Token{
                {Type: "LBrace", Value: "{"},
                {Type: "String", Value: `"id"`},
                {Type: "Colon", Value: ":"},
                {Type: "Identifier", Value: "UUID"},
                {Type: "RBrace", Value: "}"},
            },
            wantErr: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tok := tokenizer.NewTokenizer(matchers...)
            tok.Initialize(tt.input)

            tokens, err := tok.Tokenize()

            if (err != nil) != tt.wantErr {
                t.Errorf("Tokenize() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if len(tokens) != len(tt.wantTokens) {
                t.Fatalf("got %d tokens, want %d", len(tokens), len(tt.wantTokens))
            }

            for i, want := range tt.wantTokens {
                got := tokens[i]
                if got.Type != want.Type || got.Value != want.Value {
                    t.Errorf("token[%d] = {%q, %q}, want {%q, %q}",
                        i, got.Type, got.Value, want.Type, want.Value)
                }
            }
        })
    }
}
```

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkParse_Simple(b *testing.B) {
    input := `{"id": UUID, "name": String(1, 100)}`

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := shape.Parse(parser.FormatJSONV, input)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkParse_Complex(b *testing.B) {
    input := `{
        "user": {
            "id": UUID,
            "profile": {
                "name": String(1, 100),
                "email": Email
            },
            "roles": [String(1, 30)]
        }
    }`

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := shape.Parse(parser.FormatJSONV, input)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. ./...

# Specific package
go test -bench=. ./pkg/shape/

# With memory stats
go test -bench=. -benchmem ./...

# Multiple runs for accuracy
go test -bench=. -count=10 ./...

# Compare results
go test -bench=. ./... > old.txt
# Make changes
go test -bench=. ./... > new.txt
benchcmp old.txt new.txt
```

## Test Helpers and Utilities

### Creating Test Helpers

```go
// testutil/helpers.go
package testutil

import (
    "testing"
    "github.com/shapestone/shape/pkg/ast"
)

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func AssertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got %v, want %v", got, want)
    }
}

func MustParse(t *testing.T, format parser.Format, input string) ast.SchemaNode {
    t.Helper()
    node, err := shape.Parse(format, input)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    return node
}
```

### Mock Objects

```go
type mockVisitor struct {
    visitedType     bool
    visitedFunction bool
    visitedObject   bool
}

func (m *mockVisitor) VisitType(n *ast.TypeNode) error {
    m.visitedType = true
    return nil
}

func (m *mockVisitor) VisitFunction(n *ast.FunctionNode) error {
    m.visitedFunction = true
    return nil
}

// ... other visitor methods
```

## Integration Tests

### Cross-Format Tests

```go
func TestCrossFormat_SameAST(t *testing.T) {
    // Same schema in different formats should produce same AST
    schemas := map[parser.Format]string{
        parser.FormatJSONV: `{"id": UUID}`,
        parser.FormatYAMLV: "id: UUID",
        parser.FormatPropsV: "id=UUID",
    }

    var referenceAST string

    for format, input := range schemas {
        node, err := shape.Parse(format, input)
        if err != nil {
            t.Fatalf("Parse(%s) error: %v", format, err)
        }

        astStr := node.String()

        if referenceAST == "" {
            referenceAST = astStr
        } else if astStr != referenceAST {
            t.Errorf("Format %s produced different AST\nGot:  %s\nWant: %s",
                format, astStr, referenceAST)
        }
    }
}
```

## Test Organization

### Package-Level Tests

```
pkg/ast/
├── array_test.go      # Tests for ArrayNode
├── function_test.go   # Tests for FunctionNode
├── literal_test.go    # Tests for LiteralNode
├── object_test.go     # Tests for ObjectNode
├── type_test.go       # Tests for TypeNode
└── visitor_test.go    # Tests for Visitor interface
```

### Test Data

Use `testdata` directories for fixture files:

```
internal/parser/jsonv/
├── parser.go
├── parser_test.go
└── testdata/
    ├── valid/
    │   ├── simple.jsonv
    │   ├── nested.jsonv
    │   └── array.jsonv
    └── invalid/
        ├── unclosed.jsonv
        └── bad-syntax.jsonv
```

## Running Tests

### Standard Test Run

```bash
# All tests
go test ./...

# Verbose
go test -v ./...

# With race detector
go test -race ./...

# Specific package
go test ./pkg/ast/
```

### Continuous Integration

```bash
# CI test script
#!/bin/bash
set -e

echo "Running tests..."
go test -race -cover ./...

echo "Running linter..."
golangci-lint run

echo "Checking coverage..."
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Fail if coverage < 95%
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 95" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 95%"
    exit 1
fi
```

## Best Practices

### DO

- ✅ Write tests before fixing bugs (TDD for bug fixes)
- ✅ Use table-driven tests for multiple scenarios
- ✅ Test both success and error cases
- ✅ Test edge cases (empty input, nil values, boundary conditions)
- ✅ Use descriptive test names
- ✅ Keep tests simple and focused
- ✅ Use t.Helper() in test helper functions
- ✅ Run tests with `-race` flag

### DON'T

- ❌ Skip writing tests for "simple" code
- ❌ Test implementation details (test behavior, not internals)
- ❌ Write tests that depend on execution order
- ❌ Use global state in tests
- ❌ Write overly complex tests
- ❌ Ignore failing tests (fix or remove them)

## Coverage Tools

### Generate Coverage Report

```bash
# HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Coverage by Function

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Debugging Tests

### Verbose Output

```bash
# See test output
go test -v ./pkg/ast/

# See specific test
go test -v -run TestTypeName ./pkg/ast/
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug specific test
dlv test ./pkg/ast/ -- -test.run TestTypeName
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Coverage](https://blog.golang.org/cover)
- [Testing Best Practices](https://golang.org/doc/effective_go#testing)
- [Benchmarking](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

## Questions?

- Review existing tests for examples
- Check [GitHub Discussions](https://github.com/shapestone/shape/discussions)
- Ask in PR reviews

Happy testing! 🧪
