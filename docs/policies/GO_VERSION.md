# Go Version Policy

## Current Standard: Go 1.27

All repositories in the Shape ecosystem use **Go 1.27** as the minimum required version.

## Affected Repositories

- **shape-core** - Core tokenizer and AST infrastructure
- **shape-json** - JSON data format parser
- **shape-yaml** - YAML data format parser (planned)
- **shape-xml** - XML data format parser (planned)
- **shape-csv** - CSV data format parser (planned)
- **shape-properties** - Java properties format parser (planned)

## Rationale

### 1. Stability
Go 1.27 is a stable, production-ready release from the Go team with:
- Proven reliability in production environments
- Complete documentation and community support
- Stable API guarantees per Go's compatibility promise

### 2. CI/CD Support
Go 1.27 is fully supported by:
- GitHub Actions (`actions/setup-go@v5`)
- golangci-lint and other Go tooling
- Docker official Go images
- All major cloud providers' Go runtimes

### 3. Ecosystem Alignment
Using a single Go version across all repositories ensures:
- Consistent build behavior
- No version-related bugs when repos depend on each other
- Simplified developer onboarding
- Unified CI/CD configuration

### 4. Developer Experience
Developers only need to install one Go version to work on any Shape project, eliminating:
- Version confusion
- Tool compatibility issues
- "Works on my machine" problems

## Version Update Process

When updating the Shape ecosystem Go version:

### 1. Proposal
- Open an issue in `shape-core` proposing the Go version upgrade
- Include rationale (new features needed, security fixes, etc.)
- Tag as `ecosystem-coordination`

### 2. Discussion Period
- Allow 1 week for community feedback
- Discuss migration path and breaking changes
- Identify any blockers

### 3. Testing
- Create test branches in all repositories
- Verify all repos build successfully with new version
- Run full test suites with `-race` flag
- Check CI/CD workflows pass

### 4. Coordination
- Pick a coordination date
- Update all repos simultaneously via PRs
- Update this policy document
- Merge in dependency order: shape-core first, then parsers

### 5. Documentation
- Update all README.md files
- Update CONTRIBUTING.md with new requirements
- Publish release notes explaining the upgrade

## CI/CD Configuration

All `.github/workflows/*.yml` files must use:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.27'
    check-latest: true
```

The `check-latest: true` flag ensures the latest patch version (e.g., 1.27.3) is used.

## Local Development

### Required Version
Developers must use **Go 1.27.x or later**.

### Installation
Download from the official Go website:
https://go.dev/dl/

### Verification
```bash
go version
# Should show: go version go1.27.x ...
```

### IDE Configuration
Configure your IDE to use Go 1.27:
- **VS Code**: Set `"go.goroot"` in settings
- **GoLand**: File → Settings → Go → GOROOT
- **vim-go**: Set `g:go_version_warning = 0` if using 1.27+

## go.mod Requirements

All `go.mod` files must specify:

```go
module github.com/shapestone/shape-[project]

go 1.27
```

Do not specify patch versions (e.g., `go 1.27.3`) - only major.minor.

## Exceptions

### Pre-release Versions
Do not use pre-release Go versions (rc1, beta, etc.) in:
- Production code
- Main branches
- Published releases

Experimental branches may use pre-release versions for testing.

### Older Versions
Do not support Go versions older than 1.27 as:
- Shape uses modern Go features
- Older versions lack critical security fixes
- Tooling may not support them

## Migration Guide

When upgrading from a previous Go version:

### 1. Update go.mod
```bash
# Change go 1.XX to go 1.27
sed -i 's/go 1\..*/go 1.27/' go.mod
go mod tidy
```

### 2. Update CI/CD
Update all workflow files:
```yaml
go-version: '1.27'
```

### 3. Test Locally
```bash
go clean -cache
go test -race ./...
make lint
```

### 4. Update Documentation
Update any version references in:
- README.md
- CONTRIBUTING.md
- Installation guides

## History

| Version | Start Date | Rationale |
|---------|-----------|-----------|
| 1.27    | 2026-10   | Move to current Go release (shape-core v0.10.0) |
| 1.23    | 2024-12   | Ecosystem standardization, CI/CD compatibility |

## References

- [Go Release History](https://go.dev/doc/devel/release)
- [Go Release Policy](https://go.dev/doc/devel/release#policy)
- [actions/setup-go Documentation](https://github.com/actions/setup-go)

## Questions?

For questions about this policy:
1. Open a discussion in shape-core
2. Tag issues with `go-version` label
