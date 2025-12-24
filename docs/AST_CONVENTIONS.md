# AST Conventions for Format-Specific Features

This document defines conventions for mapping format-specific features to Shape's universal AST.

## General Principles

1. **Resolve references during parsing** - Don't store YAML anchors, XPath variables, etc. in the AST
2. **Use naming conventions** - Prefix special keys (`@` for XML attributes, `#` for special content)
3. **Store format metadata** - Use node positions and comments for encoding, namespaces, etc.
4. **Preserve semantics** - Choose representation that preserves format meaning and enables round-tripping

---

## XML Conventions

### Attributes vs Elements

**Convention:** Prefix attribute keys with `@`

```xml
<user id="123" name="Alice"/>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "@id":   *ast.LiteralNode{value: "123"},
        "@name": *ast.LiteralNode{value: "Alice"},
    }
}
```

**Rationale:** Distinguishes attributes from child elements with the same name.

### Text Content

**Convention:** Use `#text` key for element text content

```xml
<title>Hello World</title>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "#text": *ast.LiteralNode{value: "Hello World"},
    }
}
```

**For mixed content:**
```xml
<p>This is <strong>bold</strong> text.</p>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "#text": *ast.LiteralNode{value: "This is  text."},
        "strong": *ast.LiteralNode{value: "bold"},
    }
}
```

**Note:** Order may be lost with this representation. For order-sensitive content, consider using array representation.

### Namespaces

**Convention:** Include namespace prefix in property name

```xml
<root xmlns:ns="http://example.com">
    <ns:element>value</ns:element>
</root>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "@xmlns:ns": *ast.LiteralNode{value: "http://example.com"},
        "ns:element": *ast.LiteralNode{value: "value"},
    }
}
```

**Rationale:** Preserves namespace information without requiring special node types.

### CDATA Sections

**Convention:** Use `#cdata` key

```xml
<script><![CDATA[
    if (x < 10) { /* code */ }
]]></script>
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "#cdata": *ast.LiteralNode{value: "\n    if (x < 10) { /* code */ }\n"},
    }
}
```

### Processing Instructions

**Convention:** Use `?target` key or store in metadata

```xml
<?xml-stylesheet type="text/css" href="style.css"?>
```

**Option 1 - As property:**
```go
*ast.ObjectNode{
    properties: {
        "?xml-stylesheet": *ast.LiteralNode{value: "type=\"text/css\" href=\"style.css\""},
    }
}
```

**Option 2 - In metadata/comment (preferred for document-level PIs)**

### Comments

**Convention:** Preserve as node metadata or ignore during parsing

XML comments are typically not part of the data model and can be safely ignored unless round-tripping is required.

---

## YAML Conventions

### Anchors and Aliases

**Convention:** Resolve during parsing - expand aliases to referenced content

```yaml
defaults: &defaults
  timeout: 30
  retries: 3

production:
  <<: *defaults
  host: prod.example.com
```

**Maps to (anchors resolved):**
```go
*ast.ObjectNode{
    properties: {
        "defaults": *ast.ObjectNode{
            properties: {
                "timeout": *ast.LiteralNode{value: int64(30)},
                "retries": *ast.LiteralNode{value: int64(3)},
            }
        },
        "production": *ast.ObjectNode{
            properties: {
                "timeout": *ast.LiteralNode{value: int64(30)},  // Expanded from anchor
                "retries": *ast.LiteralNode{value: int64(3)},   // Expanded from anchor
                "host": *ast.LiteralNode{value: "prod.example.com"},
            }
        }
    }
}
```

**Rationale:** AST represents the resolved data structure, not the YAML syntax.

### Tags

**Convention:** Use type information for common tags, ignore or store in metadata for custom tags

```yaml
explicit: !!str 123
implicit: 123
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "explicit": *ast.LiteralNode{value: "123"},       // String (due to !!str tag)
        "implicit": *ast.LiteralNode{value: int64(123)},  // Integer (inferred)
    }
}
```

**For custom tags:**
- Option 1: Store in node metadata
- Option 2: Represent as TypeNode for validation purposes

### Multi-Document Files

**Convention:** Return array of AST roots

```yaml
---
document: 1
---
document: 2
```

**Parse function returns:**
```go
[]ast.SchemaNode{
    *ast.ObjectNode{properties: {"document": *ast.LiteralNode{value: int64(1)}}},
    *ast.ObjectNode{properties: {"document": *ast.LiteralNode{value: int64(2)}}},
}
```

**API:**
```go
func ParseMultiDocument(input string) ([]ast.SchemaNode, error)
```

---

## JSON Conventions

### Arrays

**Convention:** Two approaches depending on use case

**Option 1:** ObjectNode with numeric string keys (preserves position, uniform with objects)
```json
["a", "b", "c"]
```

```go
*ast.ObjectNode{
    properties: {
        "0": *ast.LiteralNode{value: "a"},
        "1": *ast.LiteralNode{value: "b"},
        "2": *ast.LiteralNode{value: "c"},
    }
}
```

**Option 2:** ArrayNode with element schema (for validation)
```go
*ast.ArrayNode{
    elementSchema: *ast.TypeNode{typeName: "string"},
    position: pos,
}
```

**Recommendation:** Use Option 1 for data representation, Option 2 for validation schemas.

### Object Key Order

**Convention:** ObjectNode uses Go map (unordered). If order matters, store in metadata.

```json
{"first": 1, "second": 2, "third": 3}
```

**Standard representation (order not guaranteed):**
```go
*ast.ObjectNode{
    properties: {
        "first": *ast.LiteralNode{value: int64(1)},
        "second": *ast.LiteralNode{value: int64(2)},
        "third": *ast.LiteralNode{value: int64(3)},
    }
}
```

**With order preservation (if needed):**
Store key order in metadata or use ordered representation in future AST version.

### null vs undefined

**Convention:** Use `nil` value in LiteralNode for explicit `null`, omit property for undefined

```json
{
  "explicit": null,
  "missing": undefined
}
```

```go
*ast.ObjectNode{
    properties: {
        "explicit": *ast.LiteralNode{value: nil},  // null present
        // "missing" property not in map            // undefined/absent
    }
}
```

---

## CSV Conventions

### Headers

**Option 1:** Store in metadata
```go
*ast.ObjectNode{
    // metadata: {"csv_headers": []string{"name", "age", "city"}}
    properties: {
        "0": *ast.ObjectNode{
            properties: {
                "name": *ast.LiteralNode{value: "Alice"},
                "age": *ast.LiteralNode{value: "30"},
                "city": *ast.LiteralNode{value: "NYC"},
            }
        },
    }
}
```

**Option 2:** First row as special entry
Use property key "headers" or numeric "-1" for header row.

### Rows and Columns

**Convention:** Each row is ObjectNode with column values

```csv
name,age,city
Alice,30,NYC
Bob,25,LA
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "0": *ast.ObjectNode{
            properties: {
                "name": *ast.LiteralNode{value: "Alice"},
                "age": *ast.LiteralNode{value: "30"},
                "city": *ast.LiteralNode{value: "NYC"},
            }
        },
        "1": *ast.ObjectNode{
            properties: {
                "name": *ast.LiteralNode{value: "Bob"},
                "age": *ast.LiteralNode{value: "25"},
                "city": *ast.LiteralNode{value: "LA"},
            }
        }
    }
}
```

### Type Inference

**Convention:** Parse as strings by default, optionally infer types

```csv
name,age,active
Alice,30,true
```

**Without type inference:**
```go
properties: {
    "age": *ast.LiteralNode{value: "30"},      // String
    "active": *ast.LiteralNode{value: "true"}, // String
}
```

**With type inference:**
```go
properties: {
    "age": *ast.LiteralNode{value: int64(30)},  // Integer
    "active": *ast.LiteralNode{value: true},    // Boolean
}
```

---

## Properties Files Conventions

### Flat Properties

```properties
server.host=localhost
server.port=8080
```

**Maps to:**
```go
*ast.ObjectNode{
    properties: {
        "server.host": *ast.LiteralNode{value: "localhost"},
        "server.port": *ast.LiteralNode{value: "8080"},
    }
}
```

### Hierarchical Representation (Optional)

**Alternative - parse dot notation into nested objects:**
```go
*ast.ObjectNode{
    properties: {
        "server": *ast.ObjectNode{
            properties: {
                "host": *ast.LiteralNode{value: "localhost"},
                "port": *ast.LiteralNode{value: "8080"},
            }
        }
    }
}
```

**Recommendation:** Provide both representations via API flags.

---

## Summary Table

| Format | Feature | Convention | Example Key |
|--------|---------|-----------|-------------|
| **XML** | Attribute | `@` prefix | `@id`, `@class` |
| **XML** | Text content | `#text` key | `#text` |
| **XML** | CDATA | `#cdata` key | `#cdata` |
| **XML** | Namespace | Include prefix | `ns:element` |
| **XML** | Namespace declaration | `@xmlns:*` | `@xmlns:ns` |
| **XML** | Processing instruction | `?target` key | `?xml-stylesheet` |
| **YAML** | Anchor | Resolve during parse | N/A (expanded) |
| **YAML** | Alias | Resolve during parse | N/A (expanded) |
| **YAML** | Tag | Type inference | N/A (affects value type) |
| **YAML** | Multi-doc | Return array | N/A (multiple roots) |
| **JSON** | Array | Numeric string keys | `"0"`, `"1"`, `"2"` |
| **CSV** | Header row | Metadata or special key | `metadata["csv_headers"]` |
| **CSV** | Row | Numeric string key | `"0"`, `"1"`, `"2"` |
| **Properties** | Dot notation | Flat or hierarchical | `"server.host"` or nested |

---

## Future Enhancements

As the Shape ecosystem evolves, consider:

1. **OrderedObjectNode** - For formats where property order matters (XML elements, some JSON use cases)
2. **MixedContentNode** - For XML mixed content with preserved order
3. **Metadata field** - Generic metadata storage on SchemaNode interface
4. **Namespace-aware nodes** - First-class namespace support beyond naming conventions

These enhancements would be additive and backward-compatible with existing conventions.
