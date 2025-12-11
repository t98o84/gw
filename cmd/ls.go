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
		if lsPrintPath {
			// -p flag specified, output full path only
			fmt.Println(wt.Path)
		} else {
			// -p flag not specified, output detailed information
			name := filepath.Base(wt.Path)
			branch := wt.Branch
			if branch == "" {
				branch = "(detached)"
			}
			output := fmt.Sprintf("%s\t%s\t%s", name, branch, shortHash(wt.Commit))
			if wt.IsMain {
				output += "\t(main)"
			}
			fmt.Println(output)
		}
	}

	return nil
}

// shortHash returns the first 7 characters of a commit hash
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
