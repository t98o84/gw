---
name: gw
description: >-
  Manage git worktrees with the gw CLI instead of raw `git worktree`. Use when
  the user wants to create, list, remove, switch between, or run commands in git
  worktrees, work on a PR or another branch in an isolated directory, or run
  several branches in parallel. Triggers: worktree, git worktree, gw add/ls/rm/exec/sw/close,
  checkout a PR in a new dir, 別ブランチを並行で作業, PR をチェックアウト, ワークツリー.
license: MIT
compatibility: Requires git. fzf is needed only for interactive selection; gh or a GitHub token only for PR worktrees.
metadata:
  homepage: https://github.com/t98o84/gw
---

# gw — git worktree wrapper

`gw` wraps `git worktree` with intuitive naming, PR checkout, fzf selection, hooks,
and shell integration. When this skill is active, prefer `gw` over calling
`git worktree ...` directly.

This document describes the **actual runtime behavior** of the CLI. Where the CLI's
own README or `--help` text disagrees with the code, the notes below follow the code.

## Prerequisites

- **git** — always required.
- **fzf** — only needed when you run a command with no target so it prompts interactively.
  Agents should avoid this by always passing an explicit target (see "Driving gw as an agent").
- **gh** or a token — only for `--pr`. Token is read from `GITHUB_TOKEN`, then `GH_TOKEN`,
  then `gh auth token` (in that order).

## Command reference

| Command | Alias | Purpose |
|---------|-------|---------|
| `gw add [branch] [from]` | `a` | Create a worktree for a branch (optionally based on `from`) |
| `gw ls` | `l` | List worktrees |
| `gw rm [name...]` | `r` | Remove one or more worktrees |
| `gw exec [name] <cmd...>` | `e` | Run a command inside a worktree |
| `gw sw [name]` | `s` | cd into a worktree (needs shell integration) |
| `gw close` | `c` | Remove the current worktree and cd back to main (needs shell integration) |
| `gw fd` | `f` | fzf-pick a worktree and print its branch (or path with `-p`) |
| `gw init <bash\|zsh\|fish>` | `i` | Print the shell integration script |

Every command accepts its alias, e.g. `gw a`, `gw r`.

## Naming convention (exact)

A worktree directory is a **sibling of the repository root**, named
`<repoName>-<suffix>`, where:

- `repoName` = basename of the main repo root.
- `suffix` = the branch name with every `/`, `\`, and `:` replaced by `-`.

Example: in repo `ex-repo`, `gw add feature/hoge` creates `../ex-repo-feature-hoge/`
for branch `feature/hoge`.

When a command takes a `name`, you may pass any of: the branch name (`feature/hoge`),
the suffix (`feature-hoge`), the full directory name (`ex-repo-feature-hoge`), or an
absolute path. All resolve to the same worktree.

## Creating worktrees — `gw add`

```
gw add [flags] [branch] [from]
```

- `branch` (positional, optional): branch to check out. Omit it to fzf-pick a branch.
- `from` (positional, optional): base ref for a new branch (used with `-b`). It has the
  highest priority; if omitted, `add.from` from config is used.

Flags:

- `-b, --branch` — create a **new** branch.
- `-p, --pr <number|URL>` — create a worktree from a GitHub PR's head branch.
- `--open` — open the worktree in the editor after creation (no short flag).
- `-e, --editor <cmd>` — editor command (e.g. `code`, `vim`). Auto-open also requires
  `add.open=true` (or `--open`).
- `-s, --sync` — copy all changed files (modified/staged/untracked) from the main worktree.
- `-i, --sync-ignored` — copy gitignored files from the main worktree. Independent of
  `-s`; combine `-s -i` to copy both changed and gitignored files.
- `--no-open`, `--no-sync`, `--no-sync-ignored` — force-disable the matching option even
  if enabled in config (see "Config precedence").

Mutually exclusive (error if combined): `-b`+`-p`, and each `--x`+`--no-x` pair.

Behavior notes:

- On success it prints `✓ Worktree created: <path>` to **stdout**. Capture this line to
  learn the created path, or compute it from the naming rule, or run `gw ls -p`.
- If the worktree already exists it prints `Worktree already exists: <path>` and exits 0
  (no error).
- For an existing remote-only branch (without `-b`), it fetches it: `git fetch origin <b>:<b>`.
- `--sync`/`--sync-ignored`/editor failures only warn; the worktree is still created.

Examples:

```
gw add feature/hoge                  # existing branch → ../repo-feature-hoge/
gw add -b feature/new                # new branch from current HEAD
gw add -b feature/new origin/main    # new branch based on origin/main
gw add --pr 123                      # PR #123 head branch
gw add --pr https://github.com/o/r/pull/123
gw add feature/hoge --open -e code   # create and open in editor
```

## Listing — `gw ls`

- `gw ls` prints tab-separated `directory<TAB>branch<TAB>shortHash(7)`; the main worktree
  gets a trailing `\t(main)`; a detached HEAD shows `(detached)` as the branch.
- `gw ls -p` / `--path` prints only the full path of each worktree, one per line. This is
  the reliable, parse-friendly form for scripts and agents.

## Removing — `gw rm`

```
gw rm [flags] [name...]
```

- Pass one or more names to remove them. With no names it fzf-multi-selects (Tab to pick
  several); the main worktree is excluded and cannot be removed.
- `-f, --force` — remove even if the worktree is dirty (and force-delete an unmerged branch
  with `-b`).
- `-y, --yes` — **alias for `--force`**. Despite the help text, there is **no interactive
  confirmation prompt anywhere**; `-y`/`-f` simply enable force behavior.
- `-b, --branch` — also delete the branch after removing the worktree.
- `--no-force`/`--no-yes`, `--no-branch` — force-disable the matching option even if enabled
  in config.

Branch-deletion safety (with `-b`): refuses to delete `main`/`master`, refuses the current
branch, and refuses an unmerged branch unless `-f`/`--force` is given.

```
gw rm feature/hoge feature/fuga      # remove multiple
gw rm -b feature/hoge                # remove worktree and its (merged) branch
gw rm -f -b feature/hoge             # also delete an unmerged branch
```

## Running commands — `gw exec`

```
gw exec [name] <command...>
```

Runs `<command...>` with the working directory set to the target worktree, inheriting
stdin/stdout/stderr, and propagates the command's exit code. If the first argument is not a
known worktree, all arguments are treated as the command and a worktree is fzf-selected.

```
gw exec feature/hoge git status
gw exec feature-hoge npm install
```

Because it only changes the child process's directory, `gw exec` works fine for
non-interactive agents (unlike `sw`/`close`).

## Switching & closing — `gw sw` / `gw close` (need shell integration)

`sw` and `close` change **your shell's** current directory, which a plain binary cannot do.
They only work after installing shell integration:

```
eval "$(gw init bash)"   # ~/.bashrc
eval "$(gw init zsh)"    # ~/.zshrc
gw init fish | source    # ~/.config/fish/config.fish
```

`gw init` defines a `gw()` shell function that intercepts `sw`/`close`, reads the target
path via `gw ... --print-path`, and runs `cd` (and, for `close`, `gw rm`).

- `gw sw [name]` — cd into a worktree (fzf-pick if no name).
- `gw close` — from **inside a non-main worktree**, cd back to the main worktree and remove
  the current one. It cannot be run from the main worktree. Flags: `-y/--yes` (and its alias
  `--force`) auto-confirm the removal; `--no-yes`/`--no-force` disable that. There is **no
  `-b`/`--branch` on `close`** — it does not delete branches.

Without shell integration, `sw`/`close` just print the `cd ...` command to run manually.

## Finding — `gw fd`

`gw fd` fzf-picks a worktree and prints its **branch name** to stdout (`gw fd -p` prints the
full path instead). Note: the help text says "directory name" but the code prints the branch.

## Project hooks — `gw.yaml`

Place a `gw.yaml` at the **repository root** to run hooks around the worktree lifecycle:

```yaml
hooks:
  pre_add:     [{ command: "..." }]   # before creation; failure ABORTS the add
  post_add:    [{ command: "...", env: { KEY: value } }]  # after creation; failure only warns
  pre_remove:  [{ command: "..." }]   # before removal; failure ABORTS the rm
  post_remove: [{ command: "..." }]   # after removal; failure only warns
```

Each hook runs as `sh -c "<command>"`, in the order defined. Injected environment variables
(exactly these three):

- `GW_WORKTREE_PATH` — absolute path of the target worktree.
- `GW_BRANCH` — branch name.
- `GW_REPO_ROOT` — absolute path of the main repo root.

Working directory of each hook:

- `post_add` and `pre_remove` run **inside the worktree** (`$GW_WORKTREE_PATH`).
- `pre_add` and `post_remove` run in the **main repo root** (`$GW_REPO_ROOT`) — the worktree
  does not exist yet / has already been removed.

Because absolute paths are always provided, a hook can act on either location regardless of
its working directory. A hook's `env:` block **can override** the `GW_*` variables (last
value wins).

## User config — `config.yaml`

A single global file (not per-repo):

- Linux/macOS: `$XDG_CONFIG_HOME/gw/config.yaml`, else `~/.config/gw/config.yaml`
- Windows: `%APPDATA%\gw\config.yaml`

```yaml
add:
  open: false          # auto-open editor after add (also gates `editor`)
  sync: false
  sync_ignored: false
  from: ""             # default base ref for new branches (e.g. origin/main)
close:
  force: false         # `gw close` auto-confirms removal
rm:
  force: false         # `gw rm` forces removal
  branch: false        # `gw rm` also deletes the branch
editor: ""             # e.g. code, vim
```

### Config precedence (lowest → highest)

1. Config-file values.
2. Regular flags (`--open`, `--sync`, `-b`, `-y`, `--editor`, `from` positional, …).
3. `--no-*` flags — always win, forcing the value to `false`.

So `--no-open`/`--no-sync`/`--no-sync-ignored`/`--no-yes`/`--no-force`/`--no-branch` override
both config and any regular flag. The `from` positional argument overrides `add.from`.

## Driving gw as an agent

Non-interactive agents should:

- **Always pass an explicit target** (branch/name/command) so gw never falls back to fzf,
  which requires a TTY and will block or fail.
- **Do not rely on `gw sw`/`gw close`** — they only affect an interactive shell that has run
  `gw init`. In a one-shot subprocess they cannot change your directory. Instead:
  - to get a worktree path: `gw ls -p` (or capture `✓ Worktree created: <path>` from `gw add`);
  - to run something in a worktree: `gw exec <name> <cmd...>`;
  - to remove a worktree: `gw rm <name>`.
- **Read `gw ls -p`** for machine-parseable output rather than the tab-separated default.
- **PR checkout**: ensure `GITHUB_TOKEN`/`GH_TOKEN` is set or `gh auth login` is done before
  `gw add --pr`.

## Known pitfalls

- **No confirmation prompts exist.** `-y`/`--yes` on `rm`/`close` are just aliases for force;
  they do not "skip a prompt" (there is none).
- **`gw close` cannot delete branches** — it has no `-b`/`--branch`. Use `gw rm -b <name>`
  from the main worktree if you need to drop the branch too.
- **`gw close -y` under bash/zsh is unreliable**: the bash/zsh integration does not forward the
  `-y` flag to the underlying `gw rm` (only the fish integration does), so forced close of a
  dirty worktree may not take effect there. Plain `gw close` (clean worktree) works on all shells.
- **`gw fd`** prints the branch name by default, not the directory name.
- **PR worktrees** are named from the PR's head branch via the normal naming rule — there is no
  PR-specific directory name.
