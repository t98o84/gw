# Contributing to gw

Thank you for your interest in contributing to gw! This document provides guidelines for contributing to the project.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include as many details as possible:

- Use the bug report template
- Provide a clear and descriptive title
- Describe the exact steps to reproduce the problem
- Provide specific examples to demonstrate the steps
- Describe the behavior you observed and what behavior you expected
- Include screenshots if applicable
- Include your environment details (OS, Go version, gw version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- Use the feature request template
- Provide a clear and descriptive title
- Provide a detailed description of the suggested enhancement
- Explain why this enhancement would be useful
- List examples of how the enhancement would be used

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Follow the development environment setup** described in [AGENTS.md](AGENTS.md)
3. **Use Docker for development**: Always develop, build, and test in Docker environment
   ```bash
   docker compose run --rm dev go test ./...
   docker compose run --rm dev go build -o gw .
   ```
4. **Write tests**: Add tests for any new functionality
5. **Follow Go conventions**: Use `go fmt` and ensure `golangci-lint` passes
6. **Update documentation**: Update README.md or other docs if needed
7. **Write clear commit messages**: Follow conventional commits format
   - `feat: add new feature`
   - `fix: fix bug`
   - `docs: update documentation`
   - `test: add tests`
   - `refactor: refactor code`
8. **Create a Pull Request** with a clear title and description

## Development Environment

### Prerequisites

- Docker and Docker Compose
- Git
- (Optional) Go 1.23+ for local development

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/gw.git
cd gw

# Run tests in Docker
docker compose run --rm dev go test ./...

# Build in Docker (for macOS Apple Silicon)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o gw ."

# Format code
docker compose run --rm dev go fmt ./...
```

### Running Tests

```bash
# Run all tests
docker compose run --rm dev go test ./...

# Run tests with verbose output
docker compose run --rm dev go test ./... -v

# Run tests with coverage
docker compose run --rm dev go test -race -coverprofile=coverage.txt ./...
```

### Code Style

- Follow standard Go formatting (`go fmt`)
- Pass `golangci-lint` checks
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise

## Project Structure

```
gw/
├── main.go              # Entry point
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command
│   ├── add.go           # gw add command
│   ├── rm.go            # gw rm command
│   ├── ls.go            # gw ls command
│   ├── sw.go            # gw sw command
│   ├── exec.go          # gw exec command
│   ├── fd.go            # gw fd command
│   ├── init.go          # gw init command
│   ├── close.go         # gw close command
│   └── fzf.go           # fzf helper functions
├── internal/
│   ├── git/             # Git operations
│   ├── github/          # GitHub API
│   ├── config/          # Configuration management
│   ├── fzf/             # fzf integration
│   ├── shell/           # Shell command execution
│   └── errors/          # Error handling
└── examples/            # Configuration examples
```

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) to enable automatic CHANGELOG generation:

```
<type>: <description>

[optional body]

[optional footer]
```

### Types for CHANGELOG

These types will appear in the CHANGELOG:

- `feat:` - A new feature (→ **Added** in CHANGELOG)
- `fix:` - A bug fix (→ **Fixed** in CHANGELOG)
- `change:` - Changes to existing functionality (→ **Changed** in CHANGELOG)
- `remove:` - Removed features (→ **Removed** in CHANGELOG)
- `security:` - Security fixes (→ **Security** in CHANGELOG)

### Types Excluded from CHANGELOG

These types are for development and won't appear in the CHANGELOG:

- `docs:` - Documentation only changes
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `build:` - Build system changes

### Examples

```bash
# New feature (will appear in CHANGELOG under "Added")
feat: add support for custom worktree paths

# Bug fix (will appear in CHANGELOG under "Fixed")
fix: handle empty branch names correctly

# Change to existing functionality (will appear in CHANGELOG under "Changed")
change: improve error messages for invalid configurations

# Security fix (will appear in CHANGELOG under "Security")
security: update dependencies to fix CVE-2024-xxxxx

# Documentation (won't appear in CHANGELOG)
docs: update installation instructions

# Test (won't appear in CHANGELOG)
test: add tests for config validation
```

### Important Notes

- Use descriptive commit messages that explain **what** and **why**
- Start with lowercase after the colon
- Keep the first line under 72 characters
- Use the body to provide additional context if needed
- Reference issues with `#issue-number` in the body or footer

For more details on the release process and CHANGELOG generation, see [.github/RELEASE_PROCESS.md](.github/RELEASE_PROCESS.md).

## Release Process

Releases are automated through GitHub Actions:

1. Update version in code if needed
2. Create and push a version tag:
   ```bash
   git tag v0.x.x
   git push origin v0.x.x
   ```
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries for multiple platforms
   - Create a GitHub Release
   - Update Homebrew tap

## Questions?

Feel free to open an issue for any questions about contributing!
