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
	flagAddOpen     bool
	flagEditor      string
	flagAddBranch   bool
	flagAddPR       string
	flagSyncAll     bool
	flagSyncIgnored bool
	// Negation flags (--no-*)
	flagNoOpen        bool
	flagNoSync        bool
	flagNoSyncIgnored bool
)

var addCmd = &cobra.Command{
	Use:     "add [flags] [branch] [from]",
	Aliases: []string{"a"},
	Short:   "Create a new worktree",
	Long: `Create a new worktree for the specified branch.

The worktree will be created in a sibling directory with the naming convention:
  <repo-name>-<branch-suffix>

Hooks:
  You can configure project-specific hooks in gw.yaml at the repository root.
  Available hooks: pre_add, post_add
  
  Hooks receive these environment variables:
    - GW_WORKTREE_PATH: Path to the worktree
    - GW_BRANCH: Branch name
    - GW_REPO_ROOT: Repository root path
  
  Example gw.yaml:
    hooks:
      pre_add:
        - name: "Validate branch name"
          command: echo "Creating worktree for $GW_BRANCH"
      post_add:
        - name: "Install dependencies"
          command: cd "$GW_WORKTREE_PATH" && npm install

Examples:
  gw add feature/hoge
    Creates ../ex-repo-feature-hoge/ and checks out feature/hoge

  gw add -b feature/new origin/main
    Creates a new branch from origin/main and worktree

  gw add -b feature/new origin/develop
    Creates a new branch from origin/develop and worktree
    Note: The 'from' argument only applies when creating a new branch (-b flag)

  gw add --pr 123
    Creates a worktree for PR #123

  gw add
    Interactive branch selection with fzf`,
	Args: cobra.MaximumNArgs(2),
	RunE: runAdd,
}

func init() {
	// Load configuration from config file
	globalConfig = config.LoadOrDefault()

	addCmd.Flags().BoolVarP(&flagAddBranch, "branch", "b", false, "Create a new branch")
	addCmd.Flags().StringVarP(&flagAddPR, "pr", "p", "", "PR number or URL to create worktree for")
	addCmd.Flags().BoolVar(&flagAddOpen, "open", false, "Open worktree in editor after creation")
	addCmd.Flags().StringVarP(&flagEditor, "editor", "e", "", "Editor command to use (e.g., code, vim)")
	addCmd.Flags().BoolVarP(&flagSyncAll, "sync", "s", false, "Sync all changed files from main worktree")
	addCmd.Flags().BoolVarP(&flagSyncIgnored, "sync-ignored", "i", false, "Sync gitignored files from main worktree")
	// Negation flags
	addCmd.Flags().BoolVar(&flagNoOpen, "no-open", false, "Force disable opening worktree in editor (overrides config and --open)")
	addCmd.Flags().BoolVar(&flagNoSync, "no-sync", false, "Force disable syncing changed files (overrides config and --sync)")
	addCmd.Flags().BoolVar(&flagNoSyncIgnored, "no-sync-ignored", false, "Force disable syncing gitignored files (overrides config and --sync-ignored)")
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
	if flagSyncAll && flagSyncIgnored {
		return fmt.Errorf("cannot use --sync and --sync-ignored together")
	}

	// Validate --no-* flag conflicts
	if flagAddOpen && flagNoOpen {
		return fmt.Errorf("cannot use --open and --no-open together")
	}
	if flagSyncAll && flagNoSync {
		return fmt.Errorf("cannot use --sync and --no-sync together")
	}
	if flagSyncIgnored && flagNoSyncIgnored {
		return fmt.Errorf("cannot use --sync-ignored and --no-sync-ignored together")
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
	var syncFlagPtr *bool
	if cmd.Flags().Changed("sync") || cmd.Flags().Changed("sync-ignored") {
		// If either flag is set, determine sync behavior
		syncEnabled := flagSyncAll
		syncFlagPtr = &syncEnabled
	}
	var syncIgnoredFlagPtr *bool
	if cmd.Flags().Changed("sync-ignored") {
		syncIgnoredFlagPtr = &flagSyncIgnored
	}

	// Extract from argument (second argument) - this takes highest priority
	var from string
	var fromFlagPtr *string
	if len(args) >= 2 {
		from = args[1]
		fromFlagPtr = &from
	}

	mergedConfig := globalConfig.MergeWithFlags(
		openFlagPtr,
		editorFlagPtr,
		nil,
		nil,
		nil,
		syncFlagPtr,
		syncIgnoredFlagPtr,
		fromFlagPtr,
		flagNoOpen,
		flagNoSync,
		flagNoSyncIgnored,
		false,
		false,
		false,
	)

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

	// Use from from merged config if not specified in args
	if from == "" {
		from = mergedConfig.Add.From
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

	// Determine sync mode
	syncMode := determineSyncMode(mergedConfig.Add.Sync, mergedConfig.Add.SyncIgnored, flagSyncAll, flagSyncIgnored)

	// Create the worktree
	return createWorktree(repoName, branch, flagAddBranch, from, editorCmd, syncMode)
}
