# Release Checklist for v0.2.0

This document outlines the steps to prepare and release version 0.2.0 of the Shape library.

## Pre-Release Verification

### Code Quality

- [x] All 5 v0.2.0 features implemented and tested
  - [x] Complete format detection for all 6 formats
  - [x] Native YAMLV parser (replaced yaml.v3)
  - [x] Schema validation with built-in types and functions
  - [x] AST optimization (string interning)
  - [x] Custom validator registration

- [x] All tests passing
  ```bash
  make test
  ```

- [x] Test coverage ≥ 95%
  ```bash
  make coverage
  ```

- [x] No lint errors (go vet passes)
  ```bash
  make lint
  ```

- [x] All benchmarks running successfully
  ```bash
  go test -bench=. -benchmem ./pkg/shape/
  ```

### Documentation

- [x] README.md updated with:
  - [x] v0.2.0 features documented
  - [x] Updated performance numbers (YAMLV 5-6x faster!)
  - [x] Schema validation section
  - [x] Custom validator examples
  - [x] v0.2.0 marked as completed

- [x] CHANGELOG.md updated with v0.2.0 entry
  - [x] All features listed
  - [x] Performance improvements documented
  - [x] Breaking changes (none!)
  - [x] Migration guide included

- [x] BENCHMARKS.md updated with v0.2.0 results
  - [x] YAMLV native parser benchmarks
  - [x] Updated performance recommendations

### Dependencies

- [x] go.mod file correct
  - [x] Module name: `github.com/shapestone/shape`
  - [x] Go version: 1.25
  - [x] **Only 1 dependency:** `github.com/google/uuid v1.6.0`
  - [x] yaml.v3 dependency removed ✅

- [x] go.sum file up to date
  ```bash
  go mod tidy
  go mod verify
  ```

### New Features Verification

#### Format Detection
- [x] All 6 formats auto-detected correctly
- [x] ParseAuto tests passing for all formats
- [x] Minimal overhead (~100-150ns)

#### Native YAMLV Parser
- [x] yaml.v3 dependency removed
- [x] All YAMLV tests passing
- [x] 5-6x performance improvement verified
- [x] 3-5x memory reduction verified

#### Schema Validation
- [x] 15 built-in types recognized
- [x] 7 built-in functions validated
- [x] Visitor pattern implementation
- [x] 45 test cases passing

#### AST Optimization
- [x] String interning implemented
- [x] Thread-safe with RWMutex
- [x] 23 common names pre-populated

#### Custom Validator Registration
- [x] RegisterType() working
- [x] RegisterFunction() working
- [x] Method chaining implemented
- [x] Built-ins protected from unregistration
- [x] 10 test functions passing

## Release Process

### 1. Update CHANGELOG

```bash
# Edit CHANGELOG.md and add v0.2.0 section
```

### 2. Final Verification

```bash
# Clean build
make clean
make build

# Run all tests with race detection
make test

# Run benchmarks
go test -bench=. -benchmem ./pkg/shape/

# Check coverage
make coverage

# Run linters
make lint

# Verify dependencies
go mod tidy
go mod verify
```

### 3. Create PR: develop → main

```bash
# Ensure develop is up to date
git checkout develop
git pull origin develop

# Create PR to main via GitHub CLI or web interface
gh pr create --base main --head develop --title "Release v0.2.0" --body "Release v0.2.0 with all features complete"
```

### 4. Merge and Tag

After PR is approved and merged:

```bash
# Switch to main and pull latest
git checkout main
git pull origin main

# Create and push tag
git tag -a v0.2.0 -m "Release v0.2.0: Performance improvements and schema validation"
git push origin v0.2.0
```

**GitHub Actions will automatically:**
- Run all tests
- Run benchmarks
- Create GitHub release
- Extract release notes from CHANGELOG.md

### 5. Verify Release

- [ ] Check GitHub Actions workflow succeeded
- [ ] Verify release appears at https://github.com/shapestone/shape/releases
- [ ] Verify release notes are correct
- [ ] Test installation: `go get github.com/shapestone/shape@v0.2.0`

## Post-Release

- [ ] Announce release on project channels
- [ ] Update any dependent projects (data-validator)
- [ ] Monitor for issues and feedback
- [ ] Begin planning v1.0.0 features

## Rollback Procedure

If critical issues are found after release:

```bash
# Delete remote tag
git push --delete origin v0.2.0

# Delete local tag
git tag -d v0.2.0

# Delete GitHub release (use GitHub UI)

# Fix issues, then re-release as v0.2.1
```

## Success Criteria

Release is successful when:

- [x] All v0.2.0 features implemented
- [x] All tests passing on main branch
- [x] Code coverage ≥ 95%
- [x] Only 1 external dependency (google/uuid)
- [x] YAMLV 5-6x performance improvement verified
- [ ] CHANGELOG updated
- [ ] GitHub release published
- [ ] Tag pushed to repository
- [ ] Release verified installable

## v0.2.0 Highlights

### Performance Wins
- **YAMLV:** 5-6x faster with native parser
- **Memory:** 3-5x reduction for YAMLV
- **Dependencies:** Removed yaml.v3 (only google/uuid remains)

### New Features
- Complete format auto-detection (all 6 formats)
- Schema validation with extensible API
- Custom validator registration
- AST string interning optimization

### Breaking Changes
- None! Fully backward compatible with v0.1.0

---

**Prepared by:** Shapestone Team
**Date:** 2025-10-09
**Target Release Date:** TBD
