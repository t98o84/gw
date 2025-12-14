package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/git"
)

var swConfig = NewConfig()

var swCmd = &cobra.Command{
	Use:     "sw [flags] [name]",
	Aliases: []string{"s"},
	Short:   "Switch to a worktree directory",
	Long: `Switch to a worktree directory.

If no name is specified and fzf is available, an interactive selector will be shown.

Note: This command requires shell integration. Run 'gw init <shell>' to set up.

Examples:
  gw sw feature/hoge
  gw sw feature-hoge
  gw sw              # Interactive selection with fzf`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSw,
}

func init() {
	swCmd.Flags().BoolVar(&swConfig.SwPrintPath, "print-path", false, "Print the path instead of changing directory (used by shell wrapper)")
	rootCmd.AddCommand(swCmd)
}

func runSw(cmd *cobra.Command, args []string) error {
	var wt *git.Worktree
	var err error

	if len(args) == 0 {
		// Interactive selection with fzf
		wt, err = selectWorktreeWithFzf(false)
		if err != nil {
			return err
		}
		if wt == nil {
			return nil // User cancelled
		}
	} else {
		identifier := args[0]

		// Find the worktree
		wt, err = git.FindWorktree(identifier)
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt == nil {
			return errors.NewWorktreeNotFoundError(identifier, nil)
		}
	}

	if swConfig.SwPrintPath {
		// Just print the path for shell wrapper to use
		fmt.Println(wt.Path)
		return nil
	}

	// Without shell integration, we can't actually change directory
	// Print instructions
	fmt.Fprintf(os.Stderr, "To switch to %s, run:\n", wt.Path)
	fmt.Fprintf(os.Stderr, "  cd %s\n\n", wt.Path)
	fmt.Fprintf(os.Stderr, "For automatic directory switching, set up shell integration:\n")
	fmt.Fprintf(os.Stderr, "  eval \"$(gw init bash)\"   # for bash\n")
	fmt.Fprintf(os.Stderr, "  eval \"$(gw init zsh)\"    # for zsh\n")
	fmt.Fprintf(os.Stderr, "  gw init fish | source    # for fish\n")

	return nil
}
