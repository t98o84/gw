package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/git"
)

var lsPrintPath bool

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"l"},
	Short:   "List all worktrees",
	Long: `List all worktrees for the current repository.

Shows the directory name for each worktree.`,
	RunE: runLs,
}

func init() {
	lsCmd.Flags().BoolVarP(&lsPrintPath, "path", "p", false, "Print the full path instead of directory name")
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	worktrees, err := git.List()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found")
		return nil
	}

	for _, wt := range worktrees {
		var name string
		if lsPrintPath {
			name = wt.Path
		} else {
			name = filepath.Base(wt.Path)
		}

		// Get short commit hash (first 7 characters)
		shortCommit := wt.Commit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}

		// Format: name branch commit (main)
		branch := wt.Branch
		if branch == "" {
			branch = "(detached)"
		}

		if wt.IsMain {
			fmt.Printf("%s %s %s (main)\n", name, branch, shortCommit)
		} else {
			fmt.Printf("%s %s %s\n", name, branch, shortCommit)
		}
	}

	return nil
}
