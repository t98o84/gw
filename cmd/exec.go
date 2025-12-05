package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/git"
)

var execCmd = &cobra.Command{
	Use:     "exec <name> <command...>",
	Aliases: []string{"e"},
	Short:   "Execute a command in a worktree",
	Long: `Execute a command in the specified worktree directory.

The current directory is preserved after command execution.

Examples:
  gw exec feature/hoge git status
  gw exec feature-hoge npm install
  gw exec ex-repo-feature-hoge make build
  gw exec git status
    Interactive worktree selection with fzf, then run 'git status'`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	var wt *git.Worktree
	var command []string
	var err error

	// Try to find the first arg as a worktree
	if len(args) >= 2 {
		wt, err = git.FindWorktree(args[0])
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}
		if wt != nil {
			command = args[1:]
		} else {
			// First arg is not a worktree, treat all args as command
			command = args
		}
	} else {
		// Only command provided, no worktree specified
		command = args
	}

	// If no worktree found, use fzf to select
	if wt == nil {
		wt, err = selectWorktreeWithFzf(false)
		if err != nil {
			return err
		}
		if wt == nil {
			return nil // User cancelled
		}
	}

	// Execute command in the worktree directory
	execCommand := exec.Command(command[0], command[1:]...)
	execCommand.Dir = wt.Path
	execCommand.Stdin = os.Stdin
	execCommand.Stdout = os.Stdout
	execCommand.Stderr = os.Stderr

	if err := execCommand.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}
