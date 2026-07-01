# gw - Git Worktree Wrapper

[![CI](https://github.com/t98o84/gw/actions/workflows/ci.yaml/badge.svg)](https://github.com/t98o84/gw/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/t98o84/gw)](https://goreportcard.com/report/github.com/t98o84/gw)
[![Go Version](https://img.shields.io/github/go-mod/go-version/t98o84/gw)](https://github.com/t98o84/gw/blob/main/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/t98o84/gw)](https://github.com/t98o84/gw/releases)

A CLI tool for managing Git worktrees simply.

## Features

- 📁 Intuitive worktree creation (`gw add feature/hoge` → `../repo-feature-hoge/`)
- 🔀 Flexible specification of branch names, suffixes, and directory names
- 🐙 Worktree creation from GitHub PRs
- 🔍 Interactive worktree selection with fzf
- 🚀 Smooth directory navigation with shell integration

## Installation

### Homebrew (macOS/Linux)

```bash
brew install t98o84/tap/gw
```

### Go

```bash
go install github.com/t98o84/gw@latest
```

### Binary

Download from [Releases](https://github.com/t98o84/gw/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/t98o84/gw.git
cd gw

# Build with Docker (macOS Apple Silicon)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o gw ."

# Build with Docker (macOS Intel)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o gw ."

# Build with Docker (Linux)
docker compose run --rm dev go build -o gw .

# Copy to a directory in PATH
sudo cp gw /usr/local/bin/
# or
mkdir -p ~/.local/bin && cp gw ~/.local/bin/
```

If Go is installed locally:

```bash
go install github.com/t98o84/gw@latest
```

## Shell Integration Setup

To navigate directories with `gw sw`, add the following to your shell configuration:

### Bash

```bash
# Add to ~/.bashrc
eval "$(gw init bash)"
```

### Zsh

```bash
# Add to ~/.zshrc
eval "$(gw init zsh)"
```

### Fish

```fish
# Add to ~/.config/fish/config.fish
gw init fish | source
```

## Usage

### Configuration File

`gw` supports YAML configuration files. The configuration file paths are:

- **Linux/macOS**: `~/.config/gw/config.yaml` (or `$XDG_CONFIG_HOME/gw/config.yaml`)
- **Windows**: `%APPDATA%\gw\config.yaml`

### Project Configuration (Hook Feature)

You can define hooks that are automatically executed during the worktree lifecycle by placing `gw.yaml` in the project root.

#### Hook Types

- **pre_add**: Executed before worktree creation (validation, preparation, etc.)
- **post_add**: Executed after worktree creation (setup, initialization, etc.)
- **pre_remove**: Executed before worktree deletion (backup, cleanup, etc.)
- **post_remove**: Executed after worktree deletion (notification, final cleanup, etc.)

#### Example gw.yaml

```yaml
hooks:
  # Before worktree creation
  pre_add:
    # Branch name validation
    - command: |
        if ! echo "$GW_BRANCH" | grep -qE '^(feature|fix|hotfix)/'; then
          echo "Branch name must start with feature/, fix/, or hotfix/"
          exit 1
        fi
  
  # After worktree creation
  post_add:
    # Copy files
    - command: cp .env.example .env
    
    # Execute commands
    - command: npm install
      env:
        NODE_ENV: development
    
    # Multiple commands are also possible
    - command: |
        bundle install
        rake db:migrate
    
    # Use gw environment variables
    - command: echo "Setup complete for branch $GW_BRANCH"
  
  # Before worktree deletion
  pre_remove:
    # Backup data (pre_remove runs inside the worktree that is about to be
    # removed, so write the archive to the repo root to keep it)
    - command: |
        echo "Backing up data from $GW_WORKTREE_PATH"
        tar -czf "$GW_REPO_ROOT/backup-$GW_BRANCH-$(date +%Y%m%d).tar.gz" .
  
  # After worktree deletion
  post_remove:
    - command: echo "Cleaned up worktree for $GW_BRANCH"
```

#### Command Execution

All hooks specify shell commands in the `command` field.

#### Basic command
```yaml
- command: npm install
```

#### Execute command with environment variables
```yaml
- command: npm install
  env:
    NODE_ENV: development
```

#### Multi-line commands
```yaml
- command: |
    echo "Setting up worktree..."
    bundle install
    rake db:migrate
```

#### Available Environment Variables

gw automatically sets the following environment variables:

- `GW_WORKTREE_PATH`: Absolute path of the created worktree
- `GW_BRANCH`: Branch name
- `GW_REPO_ROOT`: Absolute path of the main repository root directory

These environment variables can be referenced in commands:

```yaml
hooks:
  post_add:
    - command: echo "Worktree created at $GW_WORKTREE_PATH for branch $GW_BRANCH"
    - command: ln -s $GW_REPO_ROOT/.env.local .env
```

You can also add custom environment variables in the `env` field (and even override gw's environment variables).

#### Working Directory

Each hook runs in a working directory that matches its point in the worktree lifecycle:

- **post_add / pre_remove**: run inside the target worktree (`$GW_WORKTREE_PATH`), since it exists at that point. For example, `post_add: cp .env.example .env` copies the file into the new worktree.
- **pre_add / post_remove**: run in the main repository root (`$GW_REPO_ROOT`), because the worktree does not exist yet (`pre_add`) or has already been removed (`post_remove`).

Because absolute paths (`$GW_WORKTREE_PATH`, `$GW_REPO_ROOT`) are always available, you can operate on either location from any hook regardless of the working directory.

#### Hook Execution Order and Error Handling

Hooks are executed in the order they are defined within each type.

- **pre_add / pre_remove**: If a hook fails, the entire operation is aborted
- **post_add / post_remove**: If a hook fails, only a warning is displayed, and the operation is treated as successful

#### Usage Examples

```bash
# Place gw.yaml in the project root
cat << 'EOF' > gw.yaml
hooks:
  pre_add:
    - command: |
        if ! echo "$GW_BRANCH" | grep -qE '^(feature|fix)/'; then
          echo "❌ Branch must start with feature/ or fix/"
          exit 1
        fi
  post_add:
    - command: cp .env.example .env
    - command: npm install
EOF

# When creating a worktree, hooks are automatically executed
gw add feature/new-feature
# Output:
# Executing pre-add hooks...
# ⚙️  Hook 1: Executing command
# ✅ Hook 1: Command completed successfully
# Creating worktree at ../repo-feature-new-feature/ for branch feature/new-feature...
# ✓ Worktree created: ../repo-feature-new-feature/
#
# Executing post-add hooks...
# ⚙️  Hook 1: Executing command: cp .env.example .env
# ✅ Hook 1: Command completed successfully
# ⚙️  Hook 2: Executing command: npm install
# ... (npm install output)
# ✅ Hook 2: Command completed successfully

# When removing a worktree, hooks are also executed
gw rm feature/new-feature
# Output:
# Executing pre-remove hooks...
# ⚙️  Hook 1: Backing up data from /path/to/worktree
# ✅ Hook 1: Command completed successfully
# Removing worktree: /path/to/worktree
# ✓ Worktree removed: /path/to/worktree
#
# Executing post-remove hooks...
# ⚙️  Hook 1: Cleaned up worktree for feature/new-feature
# ✅ Hook 1: Command completed successfully

# Invalid branch name (rejected by pre_add)
gw add invalid-branch
# Output:
# Executing pre-add hooks...
# ⚙️  Hook 1: Executing command
# ❌ Branch must start with feature/ or fix/
# ❌ Hook 1: Command failed with exit code 1
# Error: pre-add hook failed
```

### User Configuration File

#### Configuration Example

```yaml
add:
  open: true  # Automatically open in editor after worktree creation
  sync: false  # Sync changed files from main worktree (excludes gitignored)
  sync_ignored: false  # Also sync gitignored files (can be combined with sync)
rm:
  branch: false  # Also delete branch when removing worktree
  force: false  # Force removal of dirty/locked worktrees (and unmerged branches with -b)
close:
  force: false  # Force removal of dirty/locked worktrees (and unmerged branches with -b)
editor: code  # Editor command to use
```

#### Configuration Items

- `add.open` (boolean): Whether to automatically open in editor after worktree creation (default: `false`)
- `add.sync` (boolean): Whether to sync changed files (modified, staged, untracked; excludes gitignored) from main worktree (default: `false`)
- `add.sync_ignored` (boolean): Whether to also sync gitignored files (default: `false`). Independent of `add.sync`; enabling both syncs changed files **and** gitignored files.
- `add.from` (string): Base branch/commit used when creating a new branch with the `-b` flag (e.g., `origin/main`, `develop`, `HEAD~1`). The command-line argument takes precedence over this value (default: `""` - uses current branch)
- `rm.branch` (boolean): Whether to also delete associated branch when removing worktree (default: `false`)
- `rm.force` (boolean): Whether to force removal of dirty/locked worktrees and unmerged branches, passing `--force` to `git worktree remove` (and `git branch -D`). There is no interactive confirmation prompt (default: `false`)
- `close.force` (boolean): Whether to force removal when closing; forwarded to `gw rm` as `-y` (default: `false`)
- `editor` (string): Editor command to use (e.g., `code`, `vim`, `emacs`)

**Note**: Flag precedence is as follows: `--no-*` flags > regular flags > configuration file

#### About --no-* Flags

You can disable options enabled in the configuration file when executing commands:

- `--no-open`: Don't open even with `add.open=true`
- `--no-sync`: Don't sync even with `add.sync=true`
- `--no-sync-ignored`: Don't sync gitignored files even with `add.sync_ignored=true`
- `--no-yes` / `--no-force`: Disable force removal for the command it is passed to. `gw rm --no-yes` overrides `rm.force=true`; `gw close --no-yes` overrides `close.force=true` (it is not forwarded to `gw rm`, so a separately configured `rm.force=true` still applies)
- `--no-branch`: Don't delete branch even with `rm.branch=true`

```bash
# Example: Don't open even with add.open=true in config
gw add --no-open feature/hoge

# Example: Keep branch even with rm.branch=true in config
gw rm --no-branch feature/hoge
```

### Creating Worktrees

```bash
# Create a worktree for an existing branch
gw add feature/hoge
# => Creates ../ex-repo-feature-hoge/

# Create a new branch and worktree
gw add -b feature/new

# Create a worktree from a PR branch
gw add --pr 123
gw add -p 123
gw add --pr https://github.com/owner/repo/pull/123
gw add -p https://github.com/owner/repo/pull/123

# Open in editor after creating worktree (command-line flag)
gw add --open --editor code feature/hoge
gw add --open -e vim feature/hoge

# If add.open=true and editor=code are set in config file
# Editor opens automatically even without flags
gw add feature/hoge

# Don't open even with add.open=true in config (--no-open flag)
gw add --no-open feature/hoge

# Combining options is also possible
gw add -b --open --editor code feature/new
gw add --pr 123 --open -e vim
```

### Listing Worktrees

```bash
gw ls
# Output format: <directory name>\t<branch name>\t<commit hash>\t<main marker>
# Note: The tabs (\t) below are intentional - they represent the actual tab-separated output format
<!-- markdownlint-disable MD010 -->
# ex-repo	main	a1b2c3d	(main)
# ex-repo-feature-hoge	feature/hoge	b4e5f6c
# ex-repo-fix-foo	fix/foo	c7d8e9f
<!-- markdownlint-enable MD010 -->

# Output full paths only
gw ls -p
# /path/to/ex-repo
# /path/to/ex-repo-feature-hoge
# /path/to/ex-repo-fix-foo
```

### Removing Worktrees

```bash
# All of these specify the same worktree
gw rm feature/hoge
gw rm feature-hoge
gw rm ex-repo-feature-hoge

# Remove multiple worktrees at once
gw rm feature/hoge feature/fuga fix/foo

# Also delete branch (-b/--branch option)
gw rm -b feature/hoge
gw rm --branch feature-hoge

# Force delete (also delete unmerged branches)
gw rm -f -b feature/hoge

# Without arguments, select interactively with fzf (Tab for multiple selection)
gw rm
```

**Note**: The following safety checks are applied when deleting branches:
- Cannot delete `main` or `master` branch
- Cannot delete the current branch
- Cannot delete unmerged branches without `-f`/`--force` flag

### Executing Commands in Worktrees

```bash
gw exec feature/hoge git status
gw exec feature-hoge npm install

# Omit worktree name to select with fzf
gw exec git status
```

### Navigating to Worktrees

```bash
# Navigate to the specified worktree
gw sw feature/hoge

# Select interactively with fzf
gw sw
```

### Closing Current Worktree

```bash
# Close current worktree and return to main worktree
gw close

# Force removal (dirty/locked worktree) and close
gw close -y
gw close --yes

# Also delete branch
gw close -b
gw close --branch

# Force close (also delete unmerged branches)
gw close -f -b
```

**Note**: The `gw close` command:
- Cannot be executed from main worktree (`main` or `master`)
- Requires shell integration (setup with `gw init` required)
- Can force removal of dirty/locked worktrees by setting `close.force=true` in config file

## Command List

| Command | Alias | Description |
|---------|-------|-------------|
| `gw add <branch>` | `gw a` | Create worktree |
| `gw add` | `gw a` | Branch selection with fzf (no arguments) |
| `gw add -b <branch>` | `gw a -b` | Create new branch + worktree |
| `gw add --pr <url\|number>` | `gw a --pr`, `gw a -p` | Create worktree from PR branch |
| `gw add --open` | `gw a --open` | Open in editor after worktree creation |
| `gw add --no-open` | `gw a --no-open` | Don't open in editor (ignore config) |
| `gw add --editor <cmd>` | `gw a -e` | Specify editor command to use |
| `gw add --sync` | `gw a --sync` | Sync files from main worktree |
| `gw add --no-sync` | `gw a --no-sync` | Don't sync files (ignore config) |
| `gw add --sync-ignored` | `gw a --sync-ignored` | Also sync gitignored files |
| `gw add --no-sync-ignored` | `gw a --no-sync-ignored` | Don't sync gitignored files (ignore config) |
| `gw ls` | `gw l` | List worktrees |
| `gw ls -p` | `gw l -p` | Display only full paths of worktrees |
| `gw rm [name...]` | `gw r` | Remove worktree(s) (no arguments or multiple) |
| `gw rm` | `gw r` | Select with fzf (no arguments, Tab for multiple) |
| `gw rm -b <name>` | `gw r -b` | Remove worktree and branch |
| `gw rm --no-branch <name>` | `gw r --no-branch` | Don't delete branch (ignore config) |
| `gw rm --yes/-y` | `gw r -y` | Force removal of dirty/locked worktree |
| `gw rm --no-yes/--no-force` | `gw r --no-yes` | Disable force removal (ignore config) |
| `gw exec [name] <cmd...>` | `gw e` | Execute command in target worktree (fzf without arguments) |
| `gw sw [name]` | `gw s` | Navigate to target worktree (fzf without arguments) |
| `gw close [flags]` | `gw c` | Close current worktree and return to main |
| `gw close -b` | `gw c -b` | Close and delete worktree and branch |
| `gw close -y/--yes` | `gw c -y` | Close and force removal of dirty/locked worktree |
| `gw close --no-yes/--no-force` | `gw c --no-yes` | Disable force removal (ignore config) |
| `gw fd` | `gw f` | Search worktrees with fzf (output branch name) |
| `gw fd -p` | `gw f -p` | Search worktrees with fzf (output full path) |
| `gw init <shell>` | `gw i` | Output shell initialization script |

## AI Agent Skill

This repository ships an [Agent Skill](https://agentskills.io) at
[`skills/gw/SKILL.md`](skills/gw/SKILL.md) that teaches AI coding agents how to drive
`gw` correctly (commands, naming convention, hooks, config precedence, and known
pitfalls) instead of calling `git worktree` directly.

The skill is **tool-agnostic** — it follows the Agent Skills open standard, so it works
across Claude Code, GitHub Copilot, Cursor, Codex, Gemini CLI, and other skills-compatible
agents.

### Install for Claude Code

**Option A — plugin marketplace:**

```shell
/plugin marketplace add t98o84/gw
/plugin install gw@gw
```

**Option B — copy the skill (minimal):**

```bash
# Personal (all projects)
mkdir -p ~/.claude/skills && cp -r skills/gw ~/.claude/skills/gw

# Or project-local (this repo only)
mkdir -p .claude/skills && cp -r skills/gw .claude/skills/gw
```

### Install with the GitHub CLI (`gh skill`, multi-agent)

The `gh skill` GitHub CLI extension discovers `skills/*/SKILL.md` and supports many agents
via `--agent`:

```bash
gh skill search gw
gh skill preview t98o84/gw gw
gh skill install t98o84/gw gw --agent claude-code   # other agents: change --agent
```

> Note: `gh skill` is in preview and its interface may change.

### npm (optional)

Distributing this skill via npm is not required. If you want an npm-style flow, use a
third-party universal loader that reads the skill straight from GitHub — no npm publish on
`gw`'s side:

```bash
npx openskills install github:t98o84/gw
```

## Required Tools

- `git`
- `fzf` (for interactive selection)
- `gh` (for PR integration)

## License

MIT
