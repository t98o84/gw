package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var fnPrintPath bool

var fnCmd = &cobra.Command{
	Use:     "fn",
	Aliases: []string{"f"},
	Short:   "Find a worktree with fzf",
	Long: `Find a worktree using fzf interactive search.

Prints the selected worktree name or path.

Examples:
  gw fn           # Interactive search, print worktree name
  gw fn -p        # Interactive search, print full path`,
	Args: cobra.NoArgs,
	RunE: runFn,
}

func init() {
	fnCmd.Flags().BoolVarP(&fnPrintPath, "path", "p", false, "Print the full path instead of directory name")
	rootCmd.AddCommand(fnCmd)
}

func runFn(cmd *cobra.Command, args []string) error {
	wt, err := selectWorktreeWithFzf(false)
	if err != nil {
		return err
	}
	if wt == nil {
		return nil // User cancelled
	}

	if fnPrintPath {
		fmt.Println(wt.Path)
	} else {
		fmt.Println(wt.Branch)
	}

	return nil
}
