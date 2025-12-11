# Local Development Setup

This guide will help you set up your local development environment for contributing to Shape.

## Prerequisites

### Required

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)

### Recommended

- **golangci-lint** - For code linting ([Installation](https://golangci-lint.run/usage/install/))
- **Make** - For build automation (usually pre-installed on macOS/Linux)
- **IDE with Go support** - VS Code, GoLand, or similar

## Initial Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR-USERNAME/shape.git
cd shape

# Add upstream remote
git remote add upstream https://github.com/shapestone/shape.git
```

### 2. Verify Go Installation

```bash
go version
# Should output: go version go1.25 or later
```

### 3. Install Dependencies

Shape has minimal dependencies (only `google/uuid` for production code):

```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### 4. Install Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
golangci-lint --version
```

## Project Structure

```
shape/
├── cmd/                    # Command-line tools
│   └── shape-validate/    # Schema validation CLI
├── docs/                   # Documentation
│   ├── architecture/      # Architecture docs
│   ├── contributor/       # Contributor guides (this file)
│   └── validation/        # Validation documentation
├── examples/              # Usage examples
│   ├── basic/            # Basic usage
│   ├── advanced/         # Advanced usage
│   ├── multi-format/     # Multi-format examples
│   └── custom-dsl/       # Custom DSL example
├── internal/             # Private packages
│   └── parser/           # Parser implementations
├── pkg/                  # Public API
│   ├── ast/             # Abstract Syntax Tree
│   ├── formats/         # Data format parsers
│   ├── shape/           # Main parsing API
│   ├── tokenizer/       # Public tokenizer API
│   └── validator/       # Schema validation
└── Makefile             # Build automation
```

## Building

### Using Make

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run linters
make lint

# Build CLI tools
make build

# Clean build artifacts
make clean
```

### Using Go Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/tokenizer/

# Build CLI
go build -o shape-validate ./cmd/shape-validate/

# Run linter
golangci-lint run
```

## Running Tests

### Basic Testing

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/ast/

# Verbose output
go test -v ./...

# With race detection
go test -race ./...
```

### Coverage

```bash
# Coverage report
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./pkg/ast/
go test -cover ./pkg/tokenizer/
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Specific benchmarks
go test -bench=. ./pkg/shape/

# With memory stats
go test -bench=. -benchmem ./...
```

## Code Quality

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix

# Specific directories
golangci-lint run ./pkg/...
```

### Formatting

```bash
# Format all code
go fmt ./...

# Check formatting
gofmt -l .

# Format specific files
go fmt ./pkg/ast/*.go
```

## Working with Examples

### Running Examples

```bash
# Basic example
go run examples/basic/main.go

# Advanced example
go run examples/advanced/main.go

# Multi-format example
go run examples/multi-format/main.go

# Custom DSL example
cd examples/custom-dsl
go run main.go
```

### Building Examples

```bash
# Build all examples
go build -o bin/basic examples/basic/main.go
go build -o bin/advanced examples/advanced/main.go
go build -o bin/multi-format examples/multi-format/main.go
```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Edit code, add tests, update documentation.

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Check coverage
go test -cover ./...

# Run linter
golangci-lint run

# Format code
go fmt ./...
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with clear message
git commit -m "feat: add new feature"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/) format.

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Troubleshooting

### Go Module Issues

```bash
# Reset go.mod and go.sum
go mod tidy

# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

### Build Issues

```bash
# Clean build cache
go clean -cache

# Rebuild everything
go build -a ./...
```

### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./pkg/ast/

# Run with race detector
go test -race ./...
```

## IDE Configuration

### VS Code

Create `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true
}
```

### GoLand/IntelliJ IDEA

1. Enable Go modules support: Preferences → Go → Go Modules → Enable Go modules integration
2. Set up golangci-lint: Preferences → Tools → File Watchers → Add golangci-lint
3. Enable auto-formatting on save

## Keeping Your Fork Updated

```bash
# Fetch upstream changes
git fetch upstream

# Update main branch
git checkout main
git merge upstream/main

# Push to your fork
git push origin main

# Rebase feature branch (if needed)
git checkout feature/your-feature-name
git rebase main
```

## Performance Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=. ./pkg/shape/

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=. ./pkg/shape/

# Analyze profile
go tool pprof mem.prof
```

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)

## Getting Help

- **Issues:** [GitHub Issues](https://github.com/shapestone/shape/issues)
- **Discussions:** [GitHub Discussions](https://github.com/shapestone/shape/discussions)
- **Documentation:** [docs/](../)
- **Examples:** [examples/](../../examples/)

## Next Steps

After setting up your environment:

1. Read [CONTRIBUTING.md](../../CONTRIBUTING.md)
2. Review [Testing Guide](testing-guide.md)
3. Check [Architecture Documentation](../architecture/ARCHITECTURE.md)
4. Browse [existing issues](https://github.com/shapestone/shape/issues)
5. Join discussions on [GitHub Discussions](https://github.com/shapestone/shape/discussions)

Happy coding! 🚀
