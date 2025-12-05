package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gw",
	Short: "Git worktree wrapper - simplify git worktree management",
	Long: `gw is a CLI tool that simplifies git worktree management.

It provides easy commands to create, list, remove, and switch between worktrees
with intuitive naming conventions and fzf integration.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
