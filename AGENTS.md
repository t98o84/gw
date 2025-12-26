# AGENTS.md - Development Guide for AI Agents

This document provides guidelines for AI agents (GitHub Copilot, Claude, etc.) working on this project.

## Development Environment

### ⚠️ Important: Use Docker Environment

**Always use Docker environment for development, building, and testing, not local environment.**

```bash
# Enter development container
docker compose run --rm dev sh

# Or execute commands directly
docker compose run --rm dev go test ./...
docker compose run --rm dev go build -o gw .
```

### Command Examples

```bash
# Run tests
docker compose run --rm dev go test ./...

# Detailed test output
docker compose run --rm dev go test ./... -v

# Build (for Linux)
docker compose run --rm dev go build -o gw .

# Build (for macOS Apple Silicon)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o gw ."

# Build (for macOS Intel)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o gw ."

# go mod tidy
docker compose run --rm dev go mod tidy

# Format
docker compose run --rm dev go fmt ./...
```

## Project Structure

```text
gw/
├── main.go              # Entry point
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command
│   ├── add.go           # gw add - Create worktree
│   ├── rm.go            # gw rm - Remove worktree (multiple selection support)
│   ├── ls.go            # gw ls - List worktrees
│   ├── sw.go            # gw sw - Switch worktree
│   ├── exec.go          # gw exec - Execute command in worktree
│   ├── fd.go            # gw fd - Search worktree with fzf
│   ├── init.go          # gw init - Output shell integration script
│   └── fzf.go           # fzf helper functions
├── internal/
│   ├── git/             # Git operations
│   │   ├── worktree.go  # git worktree operations
│   │   └── naming.go    # Naming convention conversion
│   └── github/          # GitHub API
│       └── pr.go        # Get branch from PR
├── go.mod
├── go.sum
├── Dockerfile
└── compose.yaml
```

## Coding Conventions

### Language
- Code comments: English
- Commit messages: English
- Documentation (README, etc.): English

### Style
- Follow standard Go formatting (`go fmt`)
- Wrap errors appropriately (`fmt.Errorf("context: %w", err)`)
- Use `os/exec` for executing external commands

### Testing
- Test files follow `*_test.go` naming convention
- Don't directly call functions requiring interactive input like fzf
- Table-driven tests are recommended

## Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/google/go-github](https://github.com/google/go-github) - GitHub API client
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) - OAuth2 authentication

## Notes

1. **fzf-related tests**: Do not directly call fzf in tests as it requires interactive input
2. **git commands**: Execute through the `internal/git` package. Tests may run outside a git repository
3. **GitHub API**: Use `GITHUB_TOKEN`, `GH_TOKEN` environment variables, or `gh auth token` for authentication
