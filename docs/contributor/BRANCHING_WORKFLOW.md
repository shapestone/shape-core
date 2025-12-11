# Branching Workflow

This document describes the branching and development workflow for the Shape project.

## Branch Structure

### Main Branches

- **`main`** - Production-ready code, protected branch
  - Contains only released, stable code
  - All commits tagged with version numbers
  - Direct pushes are **not allowed**
  - Changes only via pull requests from `develop` or hotfix branches

- **`develop`** - Integration branch for ongoing development
  - Default branch for day-to-day work
  - All feature branches merge here first
  - Periodically merged to `main` for releases

### Working Branches

- **Feature branches** - `feature/description` or `phase-N-description`
  - Created from `develop`
  - Merged back to `develop` via PR
  - Deleted after merge

- **Hotfix branches** - `hotfix/description`
  - Created from `main` for urgent production fixes
  - Merged to both `main` and `develop`
  - Tagged with patch version

## Workflow Steps

### 1. Daily Development Work

```bash
# Start from develop
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/add-format-detection

# Make changes, commit as you work
git add .
git commit -m "Add format detection for YAMLV"

# Push to remote
git push -u origin feature/add-format-detection

# Create PR to develop
gh pr create --base develop --title "Add YAMLV format detection" --body "Description..."

# Self-merge PR (no approvers required)
gh pr merge --squash

# Clean up
git checkout develop
git pull origin develop
git branch -d feature/add-format-detection
```

### 2. Release Process

```bash
# When ready to release from develop
git checkout develop
git pull origin develop

# Run full test suite
make test
go test -bench=. ./pkg/shape/

# Create release branch (optional, for release prep)
git checkout -b release/v0.2.0

# Update version numbers, CHANGELOG.md
# ... make final tweaks ...

# Create PR to main
gh pr create --base main --title "Release v0.2.0" --body "$(cat CHANGELOG.md)"

# Merge to main
gh pr merge --squash

# Tag the release
git checkout main
git pull origin main
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# Create GitHub release
gh release create v0.2.0 --title "Shape v0.2.0" --notes-file CHANGELOG.md

# Merge back to develop
git checkout develop
git merge main
git push origin develop
```

### 3. Hotfix Process

```bash
# Critical bug found in production
git checkout main
git pull origin main

# Create hotfix branch
git checkout -b hotfix/fix-parser-crash

# Fix the bug, add tests
git add .
git commit -m "Fix parser crash on empty input"

# Push and create PR to main
git push -u origin hotfix/fix-parser-crash
gh pr create --base main --title "Hotfix: Fix parser crash" --body "..."

# Merge to main
gh pr merge --squash

# Tag patch release
git checkout main
git pull origin main
git tag -a v0.1.1 -m "Release v0.1.1: Fix parser crash"
git push origin v0.1.1

# Also merge to develop
git checkout develop
git merge main
git push origin develop

# Clean up
git branch -d hotfix/fix-parser-crash
git push origin --delete hotfix/fix-parser-crash
```

## Branch Protection Rules

### Main Branch Protection
✅ **Require pull requests** - No direct pushes
✅ **No approvers required** - Can self-merge PRs
✅ **No force pushes** - History is immutable
✅ **No deletions** - Branch cannot be deleted
❌ **Admins not exempt** - Rules apply to everyone

### Best Practices

1. **Keep `develop` stable** - All tests must pass before merging
2. **Small, focused PRs** - Easier to review and merge
3. **Descriptive branch names** - `feature/add-yamlv-auto-detect` not `fix-bug`
4. **Update CHANGELOG.md** - Document all changes for releases
5. **Clean up branches** - Delete after merging
6. **Sync regularly** - Pull from `develop` daily to avoid conflicts

## Current Status

- **Main branch:** Protected ✅
- **Develop branch:** Created ✅
- **Latest release:** v0.1.0
- **Current work:** On `develop` branch

## Common Commands

```bash
# Check current branch
git branch

# See all branches
git branch -a

# Switch to develop
git checkout develop

# Update from remote
git pull origin develop

# Create feature branch
git checkout -b feature/my-feature

# Push and create PR
git push -u origin feature/my-feature
gh pr create --base develop

# View PRs
gh pr list

# Merge PR
gh pr merge <number> --squash

# Delete local branch
git branch -d feature/my-feature

# Delete remote branch
git push origin --delete feature/my-feature
```

## Questions?

See the [Contributing Guide](contributing.md) for more details on how to contribute to Shape.
