package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/config"
	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/git"
)

var rmConfig = struct {
	Force    bool
	Yes      bool
	Branch   bool
	NoYes    bool
	NoForce  bool
	NoBranch bool
}{}

var rmCmd = &cobra.Command{
	Use:     "rm [name...]",
	Aliases: []string{"r"},
	Short:   "Remove worktrees",
	Long: `Remove one or more worktrees by name.

The name can be specified in various formats:
  - Branch name: feature/hoge
  - Suffix: feature-hoge
  - Full directory name: ex-repo-feature-hoge

Examples:
  gw rm feature/hoge
  gw rm feature/hoge feature/fuga   # Remove multiple worktrees
  gw rm feature-hoge
  gw rm ex-repo-feature-hoge
  gw rm -b feature/hoge             # Also delete the branch
  gw rm
    Interactive worktree selection with fzf (Tab to multi-select)`,
	RunE: runRm,
}

func init() {
	rmCmd.Flags().BoolVarP(&rmConfig.Force, "force", "f", false, "Force removal even if worktree is dirty")
	rmCmd.Flags().BoolVarP(&rmConfig.Yes, "yes", "y", false, "Skip confirmation prompt (alias for --force)")
	rmCmd.Flags().BoolVarP(&rmConfig.Branch, "branch", "b", false, "Also delete the associated git branch")
	// Negation flags
	rmCmd.Flags().BoolVar(&rmConfig.NoYes, "no-yes", false, "Force disable automatic confirmation (overrides config and --yes)")
	rmCmd.Flags().BoolVar(&rmConfig.NoForce, "no-force", false, "Alias for --no-yes")
	rmCmd.Flags().BoolVar(&rmConfig.NoBranch, "no-branch", false, "Force disable branch deletion (overrides config and --branch)")
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.LoadOrDefault()

	// Validate flag conflicts
	if (rmConfig.Force || rmConfig.Yes) && (rmConfig.NoYes || rmConfig.NoForce) {
		return fmt.Errorf("cannot use --yes/--force and --no-yes/--no-force together")
	}
	if rmConfig.Branch && rmConfig.NoBranch {
		return fmt.Errorf("cannot use --branch and --no-branch together")
	}

	// Merge with command-line flags (flags take precedence)
	var forceFlagPtr *bool
	if cmd.Flags().Changed("force") || cmd.Flags().Changed("yes") {
		forceValue := rmConfig.Force || rmConfig.Yes
		forceFlagPtr = &forceValue
	}
	var branchFlagPtr *bool
	if cmd.Flags().Changed("branch") {
		branchFlagPtr = &rmConfig.Branch
	}
	noYesValue := rmConfig.NoYes || rmConfig.NoForce
	mergedConfig := cfg.MergeWithFlags(
		nil,
		nil,
		nil,
		forceFlagPtr,
		branchFlagPtr,
		nil,
		nil,
		false,
		false,
		false,
		false,
		noYesValue,
		rmConfig.NoBranch,
	)

	var worktrees []*git.Worktree

	if len(args) == 0 {
		// Interactive selection with fzf (exclude main worktree, multi-select enabled)
		selected, err := selectWorktreesWithFzf(true, true)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return nil // User cancelled
		}
		worktrees = selected
	} else {
		// Find all specified worktrees
		for _, identifier := range args {
			wt, err := git.FindWorktree(identifier)
			if err != nil {
				return fmt.Errorf("failed to find worktree: %w", err)
			}
			if wt == nil {
				return errors.NewWorktreeNotFoundError(identifier, nil)
			}
			if wt.IsMain {
				return errors.NewInvalidInputError(identifier, "cannot remove the main worktree", nil)
			}
			worktrees = append(worktrees, wt)
		}
	}

	// Get current branch and main worktree path if we need to delete branches
	var currentBranch string
	var mainWorktreePath string
	if mergedConfig.Rm.Branch {
		var err error
		currentBranch, err = git.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		// Get main worktree path to avoid cwd issues
		allWorktrees, err := git.List()
		if err != nil {
			return fmt.Errorf("failed to list worktrees: %w", err)
		}
		for _, wt := range allWorktrees {
			if wt.IsMain {
				mainWorktreePath = wt.Path
				break
			}
		}
		// Ensure we found the main worktree path when branch deletion is enabled
		if mainWorktreePath == "" {
			return fmt.Errorf("failed to determine main worktree path: aborting branch deletion")
		}
	}

	// Remove all selected worktrees
	for _, wt := range worktrees {
		if wt.IsMain {
			fmt.Printf("⚠ Skipping main worktree: %s\n", wt.Path)
			continue
		}

		fmt.Printf("Removing worktree: %s\n", wt.Path)
		if err := git.Remove(wt.Path, mergedConfig.Rm.Force); err != nil {
			return fmt.Errorf("failed to remove %s: %w", wt.Path, err)
		}
		fmt.Printf("✓ Worktree removed: %s\n", wt.Path)

		// Delete branch if requested
		if mergedConfig.Rm.Branch && wt.Branch != "" {
			deleted, err := deleteBranchSafely(wt.Branch, currentBranch, mainWorktreePath, mergedConfig.Rm.Force)
			if err != nil {
				fmt.Printf("⚠ Failed to delete branch %s: %v\n", wt.Branch, err)
			} else if !deleted {
				fmt.Printf("ℹ Branch %s not found, skipping\n", wt.Branch)
			} else {
				fmt.Printf("✓ Branch deleted: %s\n", wt.Branch)
			}
		}
	}

	return nil
}

// deleteBranchSafely deletes a branch with safety checks.
// Returns (true, nil) if the branch was deleted successfully,
// (false, nil) if the branch doesn't exist,
// or (false, err) if an error occurred.
func deleteBranchSafely(branchName, currentBranch, mainWorktreePath string, force bool) (bool, error) {
	// Safety check: don't delete main or master branches
	if branchName == "main" || branchName == "master" {
		return false, fmt.Errorf("refusing to delete %s branch", branchName)
	}

	// Safety check: don't delete the current branch
	if branchName == currentBranch {
		return false, fmt.Errorf("refusing to delete the current branch (%s)", branchName)
	}

	// Change to main worktree directory to avoid "getwd: no such file or directory" errors
	// when the current directory is inside a worktree that was just deleted
	if mainWorktreePath != "" {
		oldDir, err := os.Getwd()
		if err == nil {
			defer func() {
				if err := os.Chdir(oldDir); err != nil {
					// Log but don't fail - this is a best-effort restoration
					fmt.Fprintf(os.Stderr, "warning: failed to restore directory: %v\n", err)
				}
			}()
		}
		if err := os.Chdir(mainWorktreePath); err != nil {
			return false, fmt.Errorf("failed to change to main worktree directory: %w", err)
		}
	}

	// Check if branch exists
	exists, err := git.BranchExists(branchName)
	if err != nil {
		return false, fmt.Errorf("failed to check if branch exists: %w", err)
	}
	if !exists {
		// Branch doesn't exist, nothing to do
		return false, nil
	}

	// If not forcing, check if branch is merged
	if !force {
		merged, err := git.IsBranchMerged(branchName)
		if err != nil {
			return false, fmt.Errorf("failed to check if branch is merged: %w", err)
		}
		if !merged {
			return false, fmt.Errorf("branch is not merged (use -f/--force to delete anyway)")
		}
	}

	// Delete the branch
	if err := git.DeleteBranch(branchName, force); err != nil {
		// Check if the error message indicates the branch is not fully merged
		errMsg := err.Error()
		if !force && strings.Contains(errMsg, "not fully merged") {
			return false, fmt.Errorf("branch is not merged (use -f/--force to delete anyway)")
		}
		return false, fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}

	return true, nil
}
