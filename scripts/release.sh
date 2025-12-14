#!/bin/bash
# Release script for Shape
# Usage: ./scripts/release.sh v0.2.0

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.2.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format v0.0.0"
    exit 1
fi

echo "========================================="
echo "Shape Release Script"
echo "Version: $VERSION"
echo "========================================="
echo

# Check we're on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Error: Must be on main branch to release"
    echo "Current branch: $CURRENT_BRANCH"
    exit 1
fi

# Check working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory is not clean"
    git status
    exit 1
fi

# Pull latest changes
echo "→ Pulling latest changes from origin/main..."
git pull origin main

# Run tests
echo
echo "→ Running tests..."
make test

# Run linter
echo
echo "→ Running linter..."
make lint || echo "Warning: Linter found issues (continuing anyway)"

# Run benchmarks
echo
echo "→ Running benchmarks..."
go test -bench=. -benchmem ./pkg/shape/ > /dev/null

# Verify dependencies
echo
echo "→ Verifying dependencies..."
go mod tidy
go mod verify

# Check if CHANGELOG has entry for this version
if ! grep -q "## \[${VERSION#v}\]" CHANGELOG.md; then
    echo
    echo "Warning: CHANGELOG.md does not have entry for ${VERSION#v}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Confirm release
echo
echo "========================================="
echo "Ready to release $VERSION"
echo "This will:"
echo "  1. Create git tag $VERSION"
echo "  2. Push tag to origin"
echo "  3. Trigger GitHub Actions release workflow"
echo "========================================="
echo
read -p "Proceed with release? (y/N) " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled"
    exit 0
fi

# Create and push tag
echo
echo "→ Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

echo "→ Pushing tag to origin..."
git push origin "$VERSION"

echo
echo "========================================="
echo "✓ Release $VERSION initiated!"
echo "========================================="
echo
echo "GitHub Actions will now:"
echo "  - Run tests and benchmarks"
echo "  - Create GitHub release"
echo "  - Extract release notes from CHANGELOG"
echo
echo "Monitor progress at:"
echo "https://github.com/shapestone/shape-core/actions"
echo
echo "Release will appear at:"
echo "https://github.com/shapestone/shape-core/releases/tag/$VERSION"
echo
