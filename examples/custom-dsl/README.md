# Custom DSL Example

This example demonstrates how to use Shape's tokenizer to build a custom domain-specific language (DSL).

## The DSL

A simple configuration language:

```
server {
    host: "localhost"
    port: 8080
    enabled: true
}

database {
    connection: "postgresql://localhost/mydb"
    pool_size: 10
}
```

## Structure

- `main.go` - Entry point that demonstrates tokenization
- `tokenizer.go` - Custom token matchers for the config DSL
- `parser.go` - Parser that converts tokens into a config AST
- `ast.go` - AST types for the configuration

## Running

```bash
cd examples/custom-dsl
go run .
```

## Key Concepts Demonstrated

1. **Custom Matchers:** How to write matchers for keywords, identifiers, literals
2. **Tokenizer Setup:** How to configure the tokenizer with your matchers
3. **Position Tracking:** Using token positions for error messages
4. **Pattern Matching:** Using Shape's pattern combinators

## Learn More

- [Custom DSL Guide](../../docs/CUSTOM_DSL_GUIDE.md)
- [Tokenizer Documentation](../../pkg/tokenizer/README.md)
