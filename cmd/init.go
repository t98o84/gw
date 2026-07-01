package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init <shell>",
	Aliases: []string{"i"},
	Short:   "Print shell initialization script",
	Long: `Print shell initialization script for directory switching support.

Supported shells: bash, zsh, fish

Add to your shell configuration:
  bash: eval "$(gw init bash)"   # Add to ~/.bashrc
  zsh:  eval "$(gw init zsh)"    # Add to ~/.zshrc
  fish: gw init fish | source    # Add to ~/.config/fish/config.fish`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE:      runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash", "zsh":
		fmt.Print(bashZshInit)
	case "fish":
		fmt.Print(fishInit)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	return nil
}

const bashZshInit = `# gw shell integration
gw() {
  if [ "$1" = "sw" ] || [ "$1" = "s" ]; then
    local target
    target="$(command gw sw --print-path "${@:2}")"
    if [ -n "$target" ]; then
      cd "$target"
    fi
  elif [ "$1" = "close" ] || [ "$1" = "c" ]; then
    # 'gw close' understands the force (-y/-f) and branch (-b) flags, so forward
    # every argument. It echoes the flags to pass on to 'gw rm' on separate
    # stderr lines (line 2 = force, line 3 = branch).
    local stderr_output main_path worktree_to_remove yes_flag branch_flag
    stderr_output="$(command gw close --print-path "${@:2}" 2>&1 >/dev/null)"
    main_path="$(command gw close --print-path "${@:2}" 2>/dev/null)"

    # Parse stderr: line 1 = worktree path, line 2 = force flag, line 3 = branch flag
    worktree_to_remove="$(sed -n '1p' <<< "$stderr_output")"
    yes_flag="$(sed -n '2p' <<< "$stderr_output")"
    branch_flag="$(sed -n '3p' <<< "$stderr_output")"

    if [ -n "$main_path" ] && [ -n "$worktree_to_remove" ]; then
      # Each flag holds a single token (or empty), so unquoted expansion yields
      # one word (or none) in both bash and zsh.
      cd "$main_path" && command gw rm $yes_flag $branch_flag "$worktree_to_remove"
    fi
  else
    command gw "$@"
  fi
}
`

const fishInit = `# gw shell integration
function gw
  if test "$argv[1]" = "sw" -o "$argv[1]" = "s"
    set -l target (command gw sw --print-path $argv[2..])
    if test -n "$target"
      cd "$target"
    end
  else if test "$argv[1]" = "close" -o "$argv[1]" = "c"
    # 'gw close' understands the force (-y/-f) and branch (-b) flags, so forward
    # every argument. It echoes the flags to pass on to 'gw rm' on separate
    # stderr lines (line 2 = force, line 3 = branch).
    # command substitution splits on newlines: [1] = worktree path, and the
    # remaining elements are the flags for 'gw rm'.
    set -l stderr_output (command gw close --print-path $argv[2..] 2>&1 >/dev/null)
    set -l main_path (command gw close --print-path $argv[2..] 2>/dev/null)

    set -l worktree_to_remove $stderr_output[1]
    # Collect the non-empty flag lines (order preserved: force then branch).
    set -l rm_flags
    for f in $stderr_output[2..-1]
      if test -n "$f"
        set -a rm_flags $f
      end
    end

    if test -n "$main_path" -a -n "$worktree_to_remove"
      cd "$main_path"; and command gw rm $rm_flags "$worktree_to_remove"
    end
  else
    command gw $argv
  end
end
`
