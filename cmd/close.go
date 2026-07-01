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
	Force     bool
	Branch    bool
	NoYes     bool
	NoForce   bool
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
  gw close          # Close current worktree and switch to main
  gw close -b       # Also delete the associated branch
  gw close -f -b    # Force close and delete an unmerged branch`,
	Args: cobra.NoArgs,
	RunE: runClose,
}

func init() {
	closeCmd.Flags().BoolVar(&closeConfig.PrintPath, "print-path", false, "Print the path instead of changing directory (used by shell wrapper)")
	closeCmd.Flags().BoolVarP(&closeConfig.Yes, "yes", "y", false, "Force worktree removal even if dirty (forwarded to gw rm by the shell wrapper)")
	closeCmd.Flags().BoolVarP(&closeConfig.Force, "force", "f", false, "Alias for --yes")
	closeCmd.Flags().BoolVarP(&closeConfig.Branch, "branch", "b", false, "Also delete the associated git branch (forwarded to gw rm by the shell wrapper)")
	closeCmd.Flags().BoolVar(&closeConfig.NoYes, "no-yes", false, "Disable force removal (overrides config and --yes)")
	closeCmd.Flags().BoolVar(&closeConfig.NoForce, "no-force", false, "Alias for --no-yes")
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.LoadOrDefault()

	// Validate flag conflicts
	if (closeConfig.Yes || closeConfig.Force) && (closeConfig.NoYes || closeConfig.NoForce) {
		return fmt.Errorf("cannot use --yes/--force and --no-yes/--no-force together")
	}

	// Merge with command-line flags (flags take precedence)
	var yesFlagPtr *bool
	if cmd.Flags().Changed("yes") || cmd.Flags().Changed("force") {
		yesValue := closeConfig.Yes || closeConfig.Force
		yesFlagPtr = &yesValue
	}
	noYesValue := closeConfig.NoYes || closeConfig.NoForce
	mergedConfig := cfg.MergeWithFlags(
		nil,
		nil,
		yesFlagPtr,
		nil,
		nil,
		nil,
		nil,
		nil,
		false,
		false,
		false,
		noYesValue,
		false,
		false,
	)

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

	// Determine the flags to forward to 'gw rm' via the shell wrapper.
	// Force (-y) forces removal of a dirty worktree and deletion of an
	// unmerged branch; branch (-b) also deletes the associated git branch.
	yesFlag := ""
	if mergedConfig.Close.Force {
		yesFlag = "-y"
	}
	branchFlag := ""
	if closeConfig.Branch {
		branchFlag = "-b"
	}

	if closeConfig.PrintPath {
		// Print the main worktree path for shell wrapper to use (stdout)
		fmt.Println(mainPath)
		// Print the current worktree path on stderr for the shell wrapper to remove
		fmt.Fprintf(os.Stderr, "%s\n", currentWT.Path)
		// Print the flags to forward to 'gw rm' on their own stderr lines
		// (line 2 = force, line 3 = branch). Keeping each flag on a separate
		// line lets the wrapper forward them without word-splitting a joined
		// string, which zsh does not do by default.
		fmt.Fprintf(os.Stderr, "%s\n", yesFlag)
		fmt.Fprintf(os.Stderr, "%s\n", branchFlag)
		return nil
	}

	// Without shell integration, we can't actually change directory
	// Print instructions
	var rmFlags []string
	if yesFlag != "" {
		rmFlags = append(rmFlags, yesFlag)
	}
	if branchFlag != "" {
		rmFlags = append(rmFlags, branchFlag)
	}
	rmInvocation := "gw rm"
	if len(rmFlags) > 0 {
		rmInvocation += " " + strings.Join(rmFlags, " ")
	}
	fmt.Fprintf(os.Stderr, "To close this worktree and switch to main, run:\n")
	fmt.Fprintf(os.Stderr, "  cd %s && %s %s\n\n", mainPath, rmInvocation, currentWT.Path)
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
