package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/config"
	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/git"
)

var closeConfig = struct {
	PrintPath bool
	Yes       bool
}{}

var closeCmd = &cobra.Command{
	Use:     "close",
	Aliases: []string{"c"},
	Short:   "Close the current worktree and switch to the main worktree",
	Long: `Close the current worktree and switch to the main worktree.

This command must be run from within a non-main worktree. It will:
1. Switch to the main worktree
2. Remove the current worktree

Note: This command requires shell integration. Run 'gw init <shell>' to set up.

Examples:
  gw close       # Close current worktree and switch to main`,
	Args: cobra.NoArgs,
	RunE: runClose,
}

func init() {
	closeCmd.Flags().BoolVar(&closeConfig.PrintPath, "print-path", false, "Print the path instead of changing directory (used by shell wrapper)")
	closeCmd.Flags().BoolVarP(&closeConfig.Yes, "yes", "y", false, "Automatically confirm worktree deletion (used by shell wrapper)")
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.LoadOrDefault()

	// Merge with command-line flags (flags take precedence)
	var yesFlagPtr *bool
	if cmd.Flags().Changed("yes") {
		yesFlagPtr = &closeConfig.Yes
	}
	mergedConfig := cfg.MergeWithFlags(nil, nil, yesFlagPtr, nil, nil, nil, nil)

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get all worktrees
	worktrees, err := git.List()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Find the current worktree
	currentWT := findCurrentWorktree(cwd, worktrees)
	if currentWT == nil {
		return errors.NewNotInWorktreeError(cwd, nil)
	}

	// Check if trying to close the main worktree
	if currentWT.IsMain {
		return errors.NewInvalidInputError("main worktree", "cannot close the main worktree", nil)
	}

	// Get main worktree path
	mainPath, err := git.GetMainWorktreePath()
	if err != nil {
		return fmt.Errorf("failed to get main worktree path: %w", err)
	}

	if closeConfig.PrintPath {
		// Print the main worktree path for shell wrapper to use (stdout)
		fmt.Println(mainPath)
		// Print the current worktree path on stderr for the shell wrapper to remove
		fmt.Fprintf(os.Stderr, "%s\n", currentWT.Path)
		// Print -y flag status on stderr (second line)
		if mergedConfig.Close.Force {
			fmt.Fprintf(os.Stderr, "-y\n")
		} else {
			fmt.Fprintf(os.Stderr, "\n")
		}
		return nil
	}

	// Without shell integration, we can't actually change directory
	// Print instructions
	fmt.Fprintf(os.Stderr, "To close this worktree and switch to main, run:\n")
	fmt.Fprintf(os.Stderr, "  cd %s && gw rm %s\n\n", mainPath, currentWT.Path)
	fmt.Fprintf(os.Stderr, "For automatic directory switching and worktree removal, set up shell integration:\n")
	fmt.Fprintf(os.Stderr, "  eval \"$(gw init bash)\"   # for bash\n")
	fmt.Fprintf(os.Stderr, "  eval \"$(gw init zsh)\"    # for zsh\n")
	fmt.Fprintf(os.Stderr, "  gw init fish | source    # for fish\n")

	return nil
}

// findCurrentWorktree finds the worktree containing the given directory path
func findCurrentWorktree(currentPath string, worktrees []git.Worktree) *git.Worktree {
	// Clean the current path to ensure consistent comparison
	currentPath = filepath.Clean(currentPath)

	var bestMatch *git.Worktree
	longestMatchLen := 0

	for i := range worktrees {
		wtPath := filepath.Clean(worktrees[i].Path)

		// Check if current path starts with worktree path
		// Use HasPrefix with path separator to avoid partial matches
		if currentPath == wtPath || strings.HasPrefix(currentPath, wtPath+string(filepath.Separator)) {
			// Keep the longest matching path (most specific)
			if len(wtPath) > longestMatchLen {
				bestMatch = &worktrees[i]
				longestMatchLen = len(wtPath)
			}
		}
	}
	return bestMatch
}
