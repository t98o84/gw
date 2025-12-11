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
    # Capture stderr (worktree path) and stdout (main path) separately
    local worktree_to_remove target
    worktree_to_remove="$(command gw close --print-path 2>&1 >/dev/null)"
    target="$(command gw close --print-path 2>/dev/null)"
    
    if [ -n "$target" ] && [ -n "$worktree_to_remove" ]; then
      cd "$target" && command gw rm "$worktree_to_remove"
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
    # Capture stderr (worktree path and -y flag) and stdout (main path)
    set -l stderr_output (command gw close --print-path 2>&1 >/dev/null)
    set -l main_path (command gw close --print-path 2>/dev/null)
    
    # Parse stderr output: line 1 = worktree path, line 2 = -y flag
    set -l worktree_to_remove (echo "$stderr_output" | sed -n '1p')
    set -l yes_flag (echo "$stderr_output" | sed -n '2p')
    
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
