package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/errors"
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

// Execute runs the root command and handles any errors.
// This is the main entry point for the CLI application.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
		os.Exit(1)
	}
}

// handleError provides user-friendly error messages based on the error type.
// It prints the error and helpful hints to stderr.
func handleError(err error) {
	// Handle specific error types with user-friendly messages
	switch {
	case errors.IsBranchNotFoundError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Use -b flag to create a new branch\n")
		return
	case errors.IsWorktreeNotFoundError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Use 'gw ls' to list available worktrees\n")
		return
	case errors.IsWorktreeExistsError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Use 'gw ls' to list existing worktrees\n")
		return
	case errors.IsNotAGitRepoError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Navigate to a git repository directory\n")
		return
	case errors.IsFzfNotInstalledError(err):
		fmt.Fprintf(os.Stderr, "Error: fzf is not installed\n")
		fmt.Fprintf(os.Stderr, "Hint: Install fzf for interactive selection:\n")
		fmt.Fprintf(os.Stderr, "  macOS:   brew install fzf\n")
		fmt.Fprintf(os.Stderr, "  Ubuntu:  apt install fzf\n")
		fmt.Fprintf(os.Stderr, "  Other:   https://github.com/junegunn/fzf#installation\n")
		fmt.Fprintf(os.Stderr, "Or specify the target explicitly as an argument.\n")
		return
	case errors.IsNotInWorktreeError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: The current directory is not within a git worktree\n")
		return
	case errors.IsGitHubAPIError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Check your GitHub token or PR identifier\n")
		return
	case errors.IsCommandExecutionError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	case errors.IsInvalidInputError(err):
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	default:
		// Generic error
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
