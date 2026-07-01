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
    # Forward only the force flags that 'gw close' understands; other args
    # (e.g. -b) are ignored so they don't trigger an unknown-flag error.
    local has_yes="" has_no="" close_arg="" arg
    for arg in "${@:2}"; do
      case "$arg" in
        -y|--yes|--force) has_yes=1 ;;
        --no-yes|--no-force) has_no=1 ;;
      esac
    done
    if [ -n "$has_no" ]; then close_arg="--no-yes"; elif [ -n "$has_yes" ]; then close_arg="-y"; fi

    # Capture stderr (worktree path and -y flag) and stdout (main path) separately
    local stderr_output main_path worktree_to_remove yes_flag
    stderr_output="$(command gw close --print-path $close_arg 2>&1 >/dev/null)"
    main_path="$(command gw close --print-path $close_arg 2>/dev/null)"

    # Parse stderr output: line 1 = worktree path, line 2 = -y flag
    worktree_to_remove="$(sed -n '1p' <<< "$stderr_output")"
    yes_flag="$(sed -n '2p' <<< "$stderr_output")"

    if [ -n "$main_path" ] && [ -n "$worktree_to_remove" ]; then
      if [ "$yes_flag" = "-y" ]; then
        cd "$main_path" && command gw rm -y "$worktree_to_remove"
      else
        cd "$main_path" && command gw rm "$worktree_to_remove"
      fi
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
    # Forward only the force flags that 'gw close' understands; other args
    # (e.g. -b) are ignored so they don't trigger an unknown-flag error.
    set -l has_yes 0
    set -l has_no 0
    for a in $argv[2..]
      switch $a
        case -y --yes --force
          set has_yes 1
        case --no-yes --no-force
          set has_no 1
      end
    end
    set -l close_arg
    if test $has_no -eq 1
      set close_arg --no-yes
    else if test $has_yes -eq 1
      set close_arg -y
    end

    # Capture stderr (worktree path and -y flag) and stdout (main path).
    # command substitution splits on newlines, so each stderr line becomes a
    # list element: [1] = worktree path, [2] = -y flag (absent when not forced).
    set -l stderr_output (command gw close --print-path $close_arg 2>&1 >/dev/null)
    set -l main_path (command gw close --print-path $close_arg 2>/dev/null)

    set -l worktree_to_remove $stderr_output[1]
    set -l yes_flag ""
    if test (count $stderr_output) -ge 2
      set yes_flag $stderr_output[2]
    end

    if test -n "$main_path" -a -n "$worktree_to_remove"
      if test "$yes_flag" = "-y"
        cd "$main_path"; and command gw rm -y "$worktree_to_remove"
      else
        cd "$main_path"; and command gw rm "$worktree_to_remove"
      end
    end
  else
    command gw $argv
  end
end
`
