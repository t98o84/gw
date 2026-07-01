# AGENTS.md - Development Guide for AI Agents

This document provides guidelines for AI agents (GitHub Copilot, Claude, etc.) working on this project.

## Development Environment

### ⚠️ Important: Use Docker Environment

**Always use Docker environment for development, building, and testing, not local environment.**

```bash
# Enter development container
docker compose run --rm dev sh

# Or execute commands directly
docker compose run --rm dev go test -race ./...
docker compose run --rm dev go build -o gw .
```

### Command Examples

```bash
# Run tests (race detector, matches CI)
docker compose run --rm dev go test -race ./...

# Detailed test output
docker compose run --rm dev go test -race ./... -v

# Vet
docker compose run --rm dev go vet ./...

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
│   ├── close.go         # gw close - Close current worktree and switch to main
│   ├── exec.go          # gw exec - Execute command in worktree
│   ├── fd.go            # gw fd - Search worktree with fzf
│   ├── init.go          # gw init - Output shell integration script
│   ├── fzf.go           # fzf helper functions
│   ├── add_helpers.go   # gw add helper functions
│   └── config.go        # Command flag configuration
├── internal/
│   ├── git/             # Git operations
│   │   ├── worktree.go  # git worktree operations
│   │   └── naming.go    # Naming convention conversion
│   ├── github/          # GitHub API
│   │   └── pr.go        # Get branch from PR
│   ├── config/          # Configuration and hooks
│   │   ├── config.go    # Config types and flag merge
│   │   ├── loader.go    # Load config from files (Load)
│   │   ├── hooks.go     # Hook types and execution (post_add, pre_remove, ...)
│   │   ├── project.go   # Project config from gw.yaml
│   │   └── path.go      # Config file path resolution
│   ├── shell/           # External command execution
│   │   ├── executor.go  # Executor interface + real implementation
│   │   └── mock.go      # Executor mock for tests
│   ├── errors/          # Typed domain errors
│   │   └── errors.go    # Error types (BranchNotFoundError, ...)
│   └── fzf/             # fzf integration
│       └── selector.go  # Selector interface for interactive choice
├── go.mod
├── go.sum
├── Dockerfile
└── compose.yaml
```

## Shell Integration

`gw sw` and `gw close` must change the **parent shell's** working directory,
which a child process cannot do on its own. `gw init <shell>` therefore emits a
`gw()` wrapper function that runs the real binary, reads the path it prints, and
`cd`s the shell itself (see `cmd/init.go`).

### Supported shells

| Shell | `gw init` arg | Setup |
|-------|---------------|-------|
| bash  | `bash` | `eval "$(gw init bash)"` in `~/.bashrc` |
| zsh   | `zsh`  | `eval "$(gw init zsh)"` in `~/.zshrc` |
| fish  | `fish` | `gw init fish \| source` in `~/.config/fish/config.fish` |

POSIX `sh`/dash is **not** supported: the shared wrapper uses `${@:2}` array
slicing, a bash/zsh extension that `sh` cannot run. `runInit` returns
`unsupported shell` for anything other than the three above.

### Per-shell syntax caveats

When editing the embedded scripts in `cmd/init.go`, keep these in mind:

- **bash and zsh share one script** (`bashZshInit`). Use only syntax valid in
  both: POSIX `[ ... ]` tests plus `${@:2}` array slicing (valid in bash and
  zsh, not in sh).
- **The wrapper receives paths over stdout/stderr, not arguments.**
  `gw close --print-path` prints the main-worktree path on **stdout** (for the
  wrapper to `cd`) and the current-worktree path on the **first stderr line**
  (for `gw rm`); a **second stderr line** carries `-y` (or is empty). Keep
  stdout and stderr separated when editing.
- **fish uses a separate script** (`fishInit`) with different syntax:
  `function ... end`, `$argv`, `set -l`, `; and`. It parses both stderr lines
  with `sed -n '1p'`/`'2p'` and forwards `-y` to `gw rm`; the bash/zsh script
  currently reads only the worktree path. A change on one side usually needs a
  mirrored change on the other.
- **Mind the bash vs zsh word-splitting difference** when changing how these
  captured values are passed on: an unquoted expansion is word-split in bash
  but not in zsh, so the shared script must not rely on that behaviour.

To add a shell, add a `case` in `runInit` (`cmd/init.go`) and a matching
integration-script constant.

## Coding Conventions

### Language
- Code comments: English
- Commit messages: [Conventional Commits](https://www.conventionalcommits.org/). The type prefix (`feat:`/`fix:`/`docs:`, ...) must be English—git-cliff parses it for the CHANGELOG (see `.cliff.toml` and `.github/RELEASE_PROCESS.md` for the type list). The description and body may be Japanese or English.
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
- [github.com/google/go-github/v66](https://github.com/google/go-github) - GitHub API client
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) - OAuth2 authentication
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing (config files)

## Notes

1. **fzf-related tests**: Do not directly call fzf in tests as it requires interactive input
2. **git commands**: Execute through the `internal/git` package. Tests may run outside a git repository
3. **GitHub API**: Use `GITHUB_TOKEN`, `GH_TOKEN` environment variables, or `gh auth token` for authentication
4. **Linting**: CI runs `golangci-lint` (default config) and `go vet` on every PR. `golangci-lint` is **not** bundled in the dev image (it only has `git` and `fzf`); install it locally (see the golangci-lint docs) and run `golangci-lint run` before pushing to avoid CI lint failures.
5. **Configuration**: User/project config and hooks live in `internal/config`. See `examples/config.yaml` and `examples/gw.yaml` for the format; hooks such as `post_add`/`pre_remove` are documented in the README.
