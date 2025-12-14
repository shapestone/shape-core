# Releasing Shape

Quick reference for releasing new versions of Shape.

## Automated Release Process

Shape uses GitHub Actions to automate releases when you push a version tag.

### Quick Release (Recommended)

```bash
# 1. Ensure you're on main branch with latest changes
git checkout main
git pull origin main

# 2. Update CHANGELOG.md with new version section

# 3. Run the release script
./scripts/release.sh v0.2.0
```

The script will:
- ✅ Verify you're on main branch
- ✅ Check working directory is clean
- ✅ Run all tests
- ✅ Run linter
- ✅ Run benchmarks
- ✅ Verify dependencies
- ✅ Check CHANGELOG has version entry
- ✅ Create and push git tag

Then GitHub Actions will automatically:
- ✅ Run full test suite
- ✅ Run benchmarks
- ✅ Create GitHub release
- ✅ Extract release notes from CHANGELOG.md

### Manual Release

If you prefer manual control:

```bash
# 1. Verify everything is ready
make test
make lint
go test -bench=. -benchmem ./pkg/shape/
go mod tidy && go mod verify

# 2. Create and push tag
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions takes over from here.

## Release Workflow Details

The `.github/workflows/release.yml` workflow triggers on tag pushes matching `v*.*.*`:

1. **Pre-Release Tests** - Runs full test suite with race detection
2. **Benchmark Verification** - Ensures benchmarks still pass
3. **Create Release** - Automatically creates GitHub release with:
   - Release notes extracted from CHANGELOG.md
   - Auto-generated notes from commits
   - Marked as pre-release if tag contains `alpha`, `beta`, or `rc`

## Version Naming

- **Stable releases:** `v0.2.0`, `v1.0.0`, `v1.2.3`
- **Pre-releases:** `v0.2.0-alpha.1`, `v0.2.0-beta.1`, `v0.2.0-rc.1`
- **Patch releases:** `v0.2.1`, `v0.2.2`

## CHANGELOG Format

For automatic release note extraction, use this format in `CHANGELOG.md`:

```markdown
## [0.2.0] - 2025-10-09

### Added
- Feature 1
- Feature 2

### Changed
- Change 1

### Fixed
- Bug fix 1

### Performance
- Performance improvement 1
```

The workflow extracts everything between `## [0.2.0]` and the next `## [` heading.

## Checklist

Before releasing, verify:

- [ ] All tests passing (`make test`)
- [ ] No lint errors (`make lint`)
- [ ] Benchmarks passing (`go test -bench=. -benchmem ./pkg/shape/`)
- [ ] CHANGELOG.md updated with version section
- [ ] README.md updated (if needed)
- [ ] Documentation updated (if API changes)
- [ ] Examples updated (if needed)
- [ ] On main branch with clean working directory
- [ ] All features complete and documented

## Rollback

If you need to rollback a release:

```bash
# Delete remote tag
git push --delete origin v0.2.0

# Delete local tag
git tag -d v0.2.0

# Delete GitHub release (use GitHub UI)
```

Then fix issues and re-release as `v0.2.1`.

## Post-Release

After release is published:

1. ✅ Verify release at https://github.com/shapestone/shape-core/releases
2. ✅ Test installation: `go get github.com/shapestone/shape-core@v0.2.0`
3. ✅ Announce on project channels
4. ✅ Update dependent projects

## Monitoring

Watch the release process at:
- **Actions:** https://github.com/shapestone/shape-core/actions
- **Releases:** https://github.com/shapestone/shape-core/releases

## Support

For issues with the release process, see:
- `.github/workflows/release.yml` - GitHub Actions workflow
- `scripts/release.sh` - Release automation script
- `CHANGELOG.md` - Version history and release notes
