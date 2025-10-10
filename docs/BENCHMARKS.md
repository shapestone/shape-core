# Performance Benchmarks

This document contains performance benchmarks for all 6 format parsers in the Shape library.

## Test Environment

- **CPU:** Apple M1 Max
- **Platform:** darwin/arm64
- **Go Version:** 1.25

## Benchmark Results

### Simple Schemas (2 properties)

| Format | Time/op | Memory | Allocs |
|--------|---------|--------|--------|
| YAMLV  | 0.7µs   | 1.0KB  | 24     |
| CSVV   | 2.7µs   | 4.2KB  | 87     |
| XMLV   | 3.2µs   | 4.0KB  | 97     |
| PropsV | 3.2µs   | 5.4KB  | 100    |
| TEXTV  | 3.7µs   | 6.6KB  | 117    |
| JSONV  | 4.8µs   | 9.5KB  | 157    |

**Winner:** YAMLV (6.8x faster than JSONV!)

### Medium Schemas (nested objects, arrays, ~7 properties)

| Format | Time/op | Memory | Allocs |
|--------|---------|--------|--------|
| YAMLV  | 2.7µs   | 4.1KB  | 80     |
| CSVV   | 6.6µs   | 10.6KB | 211    |
| PropsV | 12.1µs  | 19.0KB | 338    |
| XMLV   | 12.6µs  | 16.0KB | 373    |
| TEXTV  | 13.0µs  | 23.0KB | 395    |
| JSONV  | 20.6µs  | 40.7KB | 657    |

**Winner:** YAMLV (7.6x faster than JSONV!)

### Large Schemas (deeply nested, ~25 properties)

| Format | Time/op | Memory  | Allocs |
|--------|---------|---------|--------|
| YAMLV  | 8.9µs   | 11.8KB  | 242    |
| CSVV   | 21.6µs  | 35.7KB  | 683    |
| PropsV | 43.3µs  | 69.8KB  | 1151   |
| XMLV   | 44.3µs  | 53.9KB  | 1243   |
| TEXTV  | 52.5µs  | 83.4KB  | 1344   |
| JSONV  | 70.0µs  | 134.2KB | 2148   |

**Winner:** YAMLV (7.9x faster than JSONV!)

### ParseAuto Performance (format detection overhead)

| Format | Time/op | Memory | Allocs |
|--------|---------|--------|--------|
| YAMLV  | 0.8µs   | 1.0KB  | 26     |
| JSONV  | 5.0µs   | 9.5KB  | 157    |

**Note:** ParseAuto adds minimal overhead (~100-150ns) compared to direct Parse calls.

## Analysis

### Performance Trends

1. **YAMLV is now the fastest parser!** (v0.2.0 native parser)
   - 6.8-7.9x faster than JSONV across all schema sizes
   - 2.4-3.0x faster than CSVV (previous champion)
   - Line-based parsing with minimal allocations
   - Native implementation without external dependencies

2. **CSVV remains very fast** for simple schemas
   - Second fastest parser overall
   - Simple structure benefits from straightforward row/column parsing
   - Minimal nesting complexity

3. **JSONV is slowest** with highest memory usage
   - Most complex tokenization (strings, numbers, nested structures)
   - Highest allocation count due to recursive parsing
   - Still acceptable performance for most use cases (5-70µs)

4. **Mid-range formats** (PropsV, XMLV, TEXTV) have similar performance
   - PropsV: 3.2-43µs depending on complexity
   - XMLV: 3.2-44µs depending on complexity
   - TEXTV: 3.7-52µs depending on complexity

5. **Memory usage scales with schema complexity**
   - Simple: 1-10KB
   - Medium: 4-41KB
   - Large: 12-134KB

6. **Allocation counts vary by parser design**
   - YAMLV has fewest allocations (24-242)
   - CSVV has moderate allocations (87-683)
   - JSONV has most allocations (157-2148)

### Recommendations

**For Performance-Critical Applications:**
- Use **YAMLV** for all schema types (fastest, lowest memory, cleanest syntax!)
- Use **CSVV** for flat schemas if YAML syntax is not preferred
- Use **XMLV** for nested schemas if XML is required
- JSONV is acceptable for most use cases despite being slowest

**For Developer Experience:**
- Use **YAMLV** for readable nested schemas (best performance + clean syntax)
- Use **JSONV** for complex nested schemas (most familiar syntax)
- Use **TEXTV** for simple flat schemas (minimal syntax)
- Use **CSVV** for tabular data representation

**For Auto-Detection:**
- ParseAuto adds minimal overhead (~100-150ns)
- Safe to use in all scenarios without performance concerns
- YAMLV detection is extremely fast (<1µs)

## Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./pkg/shape/

# Run specific format benchmarks
go test -bench=JSONV -benchmem ./pkg/shape/
go test -bench=YAMLV -benchmem ./pkg/shape/

# Run with more iterations for accuracy
go test -bench=. -benchmem -benchtime=10s ./pkg/shape/

# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./pkg/shape/
go tool pprof cpu.prof
```

## Benchmark Methodology

Each benchmark measures:
- **Time/op:** Average time per Parse() call
- **Memory:** Total bytes allocated per operation
- **Allocs:** Number of memory allocations per operation

Benchmarks are run with:
- `b.ResetTimer()` to exclude setup overhead
- Error checking to ensure parsing succeeds
- Identical schemas across formats (where possible)

## v0.2.0 Optimizations Completed

1. **YAMLV Native Parser** ✅ - Replaced yaml.v3 with custom line-based parser
   - 5-6x performance improvement across all schema sizes
   - 3-5x reduction in memory usage
   - 2-3x reduction in allocations
   - Now the fastest parser in Shape!

2. **AST String Interning** ✅ - Intern common type and function names
   - Reduces string allocations for repeated type names
   - Pre-populates 15 common types (UUID, Email, String, etc.)
   - Thread-safe with RWMutex for concurrent access

## Future Optimization Opportunities

1. **JSONV:** Reduce allocations in recursive parsing
2. **TEXTV:** Optimize string splitting for dot notation
3. **All formats:** Consider object pooling for AST nodes
4. **Tokenizer:** Optimize rune processing in hot paths
5. **Small Object Optimization:** Use slices instead of maps for objects with few properties
