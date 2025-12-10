package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/git"
)

var rmForce bool

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
  gw rm
    Interactive worktree selection with fzf (Tab to multi-select)`,
	RunE: runRm,
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force removal even if worktree is dirty")
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
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

	// Remove all selected worktrees
	for _, wt := range worktrees {
		if wt.IsMain {
			fmt.Printf("⚠ Skipping main worktree: %s\n", wt.Path)
			continue
		}

		fmt.Printf("Removing worktree: %s\n", wt.Path)
		if err := git.Remove(wt.Path, rmForce); err != nil {
			return fmt.Errorf("failed to remove %s: %w", wt.Path, err)
		}
		fmt.Printf("✓ Worktree removed: %s\n", wt.Path)
	}

	return nil
}
