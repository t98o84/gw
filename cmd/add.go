package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/config"
	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/shell"
)

var (
	// Global configuration loaded from config file
	globalConfig *config.Config
	// Command-line flags
	flagAddOpen   bool
	flagEditor    string
	flagAddBranch bool
	flagAddPR     string
)

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
	// Load configuration from config file
	globalConfig = config.LoadOrDefault()

	addCmd.Flags().BoolVarP(&flagAddBranch, "branch", "b", false, "Create a new branch")
	addCmd.Flags().StringVarP(&flagAddPR, "pr", "p", "", "PR number or URL to create worktree for")
	addCmd.Flags().BoolVar(&flagAddOpen, "open", false, "Open worktree in editor after creation")
	addCmd.Flags().StringVarP(&flagEditor, "editor", "e", "", "Editor command to use (e.g., code, vim)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	selector := fzf.NewSelector(shell.NewRealExecutor())
	return runAddWithSelector(cmd, args, selector)
}

func runAddWithSelector(cmd *cobra.Command, args []string, selector fzf.Selector) error {
	// Validate flags
	if flagAddBranch && flagAddPR != "" {
		return fmt.Errorf("cannot use --branch and --pr together")
	}

	// Merge config with flags (flags take precedence)
	var openFlagPtr *bool
	if cmd.Flags().Changed("open") {
		openFlagPtr = &flagAddOpen
	}
	var editorFlagPtr *string
	if cmd.Flags().Changed("editor") {
		editorFlagPtr = &flagEditor
	}
	mergedConfig := globalConfig.MergeWithFlags(openFlagPtr, editorFlagPtr)

	// Validate config
	if err := mergedConfig.Validate(); err != nil {
		return err
	}

	repoName, err := git.GetRepoName()
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	// Create options
	opts := &addOptions{
		createBranch: flagAddBranch,
		prIdentifier: flagAddPR,
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
	if err := ensureBranchExists(branch, flagAddBranch, fromPR); err != nil {
		return err
	}

	// Get editor command from merged config
	editorCmd := mergedConfig.GetEditor()

	// Create the worktree
	return createWorktree(repoName, branch, flagAddBranch, editorCmd)
}
