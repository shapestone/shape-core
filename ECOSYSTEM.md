# Shape Ecosystem

Shape is parser infrastructure that other projects build upon. This document lists the parser projects and related tools in the Shape ecosystem.

## Parser Projects

Parser projects use Shape's infrastructure (AST, tokenizer, validator, grammar) to implement parsers for specific formats.

### Data Format Parsers

These parsers handle standard data formats:

- **[shape-json](https://github.com/shapestone/shape-json)** - JSON parser with JSONPath query support
- **[shape-yaml](https://github.com/shapestone/shape-yaml)** - YAML parser and validator
- **[shape-xml](https://github.com/shapestone/shape-xml)** - XML parser and validator
- **[shape-csv](https://github.com/shapestone/shape-csv)** - CSV parser with schema validation
- **[shape-props](https://github.com/shapestone/shape-props)** - Java properties file parser

## Related Projects

### Shapestone Ecosystem

Shape is part of the broader Shapestone ecosystem:

- **[Shapestone](https://github.com/shapestone)** - Organization homepage
- **[Inkling](https://github.com/shapestone/inkling)** - Uses Shape's tokenizer for diagram parsing

## Using Shape Infrastructure

To build your own parser using Shape:

1. See the [Parser Implementation Guide](docs/PARSER_IMPLEMENTATION_GUIDE.md)
2. Review [shape-json](https://github.com/shapestone/shape-json) as a reference implementation
3. Check the [Architecture Documentation](docs/architecture/ARCHITECTURE.md)

## Contributing

To contribute a parser project to the Shape ecosystem:

1. Build your parser using Shape infrastructure
2. Follow the patterns established by existing parsers
3. Submit a PR to add your parser to this document

## License

All Shape ecosystem projects use Apache License 2.0 unless otherwise specified.

Copyright © 2020-2025 Shapestone
