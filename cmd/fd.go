package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var fdPrintPath bool

var fdCmd = &cobra.Command{
	Use:     "fd",
	Aliases: []string{"f"},
	Short:   "Find a worktree with fzf",
	Long: `Find a worktree using fzf interactive search.

Prints the selected worktree name or path.

Examples:
  gw fd           # Interactive search, print worktree name
  gw fd -p        # Interactive search, print full path`,
	Args: cobra.NoArgs,
	RunE: runFd,
}

func init() {
	fdCmd.Flags().BoolVarP(&fdPrintPath, "path", "p", false, "Print the full path instead of directory name")
	rootCmd.AddCommand(fdCmd)
}

func runFd(cmd *cobra.Command, args []string) error {
	wt, err := selectWorktreeWithFzf(false)
	if err != nil {
		return err
	}
	if wt == nil {
		return nil // User cancelled
	}

	if fdPrintPath {
		fmt.Println(wt.Path)
	} else {
		fmt.Println(wt.Branch)
	}

	return nil
}
