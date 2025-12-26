# Release Process

This document describes the release process for the gw (Git Worktree Wrapper) project.

## Overview

Releases are automated through GitHub Actions with the following features:

- Automatic binary builds for multiple platforms (Linux, macOS, Windows)
- Automatic GitHub Release creation
- Automatic Homebrew tap updates
- CHANGELOG generation using git-cliff in Keep a Changelog format
- Docker-based CI/CD for consistency

## Versioning

We follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: Backwards-compatible new features
- **PATCH** version: Backwards-compatible bug fixes

## Commit Message Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) with the following types:

- `feat:` - New features (→ Added in CHANGELOG)
- `fix:` - Bug fixes (→ Fixed in CHANGELOG)
- `change:` - Changes to existing functionality (→ Changed in CHANGELOG)
- `remove:` - Removed features (→ Removed in CHANGELOG)
- `security:` - Security fixes (→ Security in CHANGELOG)
- `docs:` - Documentation only (skipped in CHANGELOG)
- `test:` - Tests only (skipped in CHANGELOG)
- `chore:` - Maintenance (skipped in CHANGELOG)
- `ci:` - CI/CD changes (skipped in CHANGELOG)
- `style:` - Code style (skipped in CHANGELOG)
- `refactor:` - Code refactoring (skipped in CHANGELOG)

### Examples

```bash
feat: add support for custom worktree paths
fix: handle empty branch names correctly
change: improve error messages for invalid configurations
security: update dependencies to fix CVE-2024-xxxxx
docs: update installation instructions
```

## Release Steps

### 1. Prepare for Release

Ensure all changes are committed and merged to `main`:

```bash
git checkout main
git pull origin main
```

### 2. Create and Push Release Tag

Create a new version tag following semantic versioning:

```bash
# For a new feature release
git tag v0.4.0

# For a bug fix release
git tag v0.3.12

# Push the tag
git push origin v0.4.0
```

### 3. Automated Release Process

Once the tag is pushed, GitHub Actions automatically:

1. **Run Tests** - Ensure all tests pass
2. **Run Linters** - Verify code quality
3. **Build Binaries** - Build for all platforms
4. **Generate CHANGELOG** - Create changelog using git-cliff
5. **Create GitHub Release** - With binaries and changelog
6. **Update Homebrew Tap** - For macOS users

The release process typically takes 3-5 minutes.

### 4. Manual CHANGELOG Update (Optional)

If you want to update the CHANGELOG without creating a release:

1. Go to Actions → Update CHANGELOG
2. Click "Run workflow"
3. Optionally specify a tag
4. Click "Run workflow"

This will commit the CHANGELOG with `[skip ci]` to avoid triggering another CI run.

## CHANGELOG Format

The CHANGELOG follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format:

```markdown
## [0.4.0] - 2025-12-26

### Added
- New feature description

### Fixed
- Bug fix description

### Changed
- Changed functionality description
```

### CHANGELOG Categories

- **Added** - New features (`feat:` commits)
- **Fixed** - Bug fixes (`fix:` commits)
- **Changed** - Changes to existing functionality (`change:` commits)
- **Removed** - Removed features (`remove:` commits)
- **Security** - Security fixes (`security:` commits)

## Docker-Based Release Environment

All release operations use Docker containers to ensure consistency:

```bash
# Build release image locally
docker build -f Dockerfile.release -t gw-release:latest .

# Generate CHANGELOG locally
docker run --rm -v "$PWD:/workspace" gw-release:latest \
  git-cliff -o CHANGELOG.md

# Test GoReleaser locally (without publishing)
docker run --rm -v "$PWD:/workspace" \
  -e GITHUB_TOKEN=your_token \
  gw-release:latest \
  goreleaser release --snapshot --clean
```

## Troubleshooting

### Release Fails

If a release fails:

1. Check the GitHub Actions logs
2. Fix any issues
3. Delete the tag locally and remotely:
   ```bash
   git tag -d v0.4.0
   git push origin :refs/tags/v0.4.0
   ```
4. Re-create and push the tag

### CHANGELOG Issues

If the CHANGELOG is incorrect:

1. Fix commit messages if needed
2. Re-run the "Update CHANGELOG" workflow
3. Or manually edit CHANGELOG.md and commit with `[skip ci]`

### Homebrew Tap Update Fails

If Homebrew tap update fails:

1. Check that `HOMEBREW_TAP_GITHUB_TOKEN` secret is set
2. Verify the token has write access to the homebrew-tap repository
3. Manually update the tap if needed

## Testing Before Release

Always test your changes before creating a release:

```bash
# Run all tests in Docker
docker compose run --rm dev go test ./...

# Run with coverage
docker compose run --rm dev go test -race -coverprofile=coverage.txt ./...

# Run linter
docker compose run --rm dev golangci-lint run

# Build locally
docker compose run --rm dev go build -o gw .
```

## Release Checklist

Before creating a release:

- [ ] All tests pass
- [ ] All linters pass
- [ ] Documentation is up to date
- [ ] Commit messages follow conventions
- [ ] Version number follows semantic versioning
- [ ] CHANGELOG has been reviewed (optional)

## Post-Release

After a successful release:

1. Verify the GitHub Release page
2. Test installation via Homebrew:
   ```bash
   brew upgrade t98o84/tap/gw
   gw --version
   ```
3. Announce the release if significant

## Resources

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [git-cliff Documentation](https://git-cliff.org/)
