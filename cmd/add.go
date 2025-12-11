package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/shell"
)

var addConfig = NewConfig()

var addCmd = &cobra.Command{
	Use:     "add <branch>",
	Aliases: []string{"a"},
	Short:   "Create a new worktree",
	Long: `Create a new worktree for the specified branch.

The worktree will be created in a sibling directory with the naming convention:
  <repo-name>-<branch-suffix>

Examples:
  gw add feature/hoge
    Creates ../ex-repo-feature-hoge/ and checks out feature/hoge

  gw add -b feature/new
    Creates a new branch and worktree

  gw add --pr 123
    Creates a worktree for PR #123

  gw add
    Interactive branch selection with fzf`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&addConfig.AddCreateBranch, "branch", "b", false, "Create a new branch")
	addCmd.Flags().StringVar(&addConfig.AddPRIdentifier, "pr", "", "PR number or URL to create worktree for")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	selector := fzf.NewSelector(shell.NewRealExecutor())
	return runAddWithSelector(cmd, args, selector, addConfig)
}

func runAddWithSelector(cmd *cobra.Command, args []string, selector fzf.Selector, cfg *Config) error {
	// Validate config
	if err := cfg.Validate(); err != nil {
		return err
	}

	repoName, err := git.GetRepoName()
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	// Create options
	opts := &addOptions{
		createBranch: cfg.AddCreateBranch,
		prIdentifier: cfg.AddPRIdentifier,
		selector:     selector,
	}

	// Determine branch
	branch, err := determineBranch(args, opts, repoName)
	if err != nil {
		return err
	}
	if branch == "" {
		return nil // User cancelled
	}

	// Check if worktree already exists
	existing, err := checkExistingWorktree(branch)
	if err != nil {
		return err
	}
	if existing != nil {
		fmt.Printf("Worktree already exists: %s\n", existing.Path)
		return nil
	}

	// Ensure branch exists or can be created
	fromPR := opts.prIdentifier != ""
	if err := ensureBranchExists(branch, cfg.AddCreateBranch, fromPR); err != nil {
		return err
	}

	// Create the worktree
	return createWorktree(repoName, branch, cfg.AddCreateBranch)
}
