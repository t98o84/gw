package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/github"
)

var (
	addCreateBranch bool
	addPRIdentifier string
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

  gw add -pr 123
    Creates a worktree for PR #123

  gw add
    Interactive branch selection with fzf`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&addCreateBranch, "branch", "b", false, "Create a new branch")
	addCmd.Flags().StringVar(&addPRIdentifier, "pr", "", "PR number or URL to create worktree for")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	var branch string

	// Handle PR flag
	if addPRIdentifier != "" {
		repoName, err := git.GetRepoName()
		if err != nil {
			return fmt.Errorf("failed to get repository name: %w", err)
		}

		prBranch, err := github.GetPRBranch(addPRIdentifier, repoName)
		if err != nil {
			return fmt.Errorf("failed to get PR branch: %w", err)
		}
		branch = prBranch
	} else if len(args) == 0 {
		// Interactive branch selection with fzf
		selectedBranch, err := selectBranchWithFzf()
		if err != nil {
			return err
		}
		if selectedBranch == "" {
			return nil // User cancelled
		}
		branch = selectedBranch
	} else {
		branch = args[0]
	}

	repoName, err := git.GetRepoName()
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	// Generate worktree path
	wtPath, err := git.WorktreePath(repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Check if worktree already exists
	existing, err := git.FindWorktree(branch)
	if err != nil {
		return fmt.Errorf("failed to check existing worktree: %w", err)
	}
	if existing != nil {
		fmt.Printf("Worktree already exists: %s\n", existing.Path)
		return nil
	}

	// If not creating a new branch, check if branch exists
	if !addCreateBranch && addPRIdentifier == "" {
		exists, err := git.BranchExists(branch)
		if err != nil {
			return fmt.Errorf("failed to check branch: %w", err)
		}
		if !exists {
			// Try to fetch from remote
			remoteExists, err := git.RemoteBranchExists(branch)
			if err != nil {
				return fmt.Errorf("failed to check remote branch: %w", err)
			}
			if remoteExists {
				fmt.Printf("Fetching branch %s from origin...\n", branch)
				if err := git.FetchBranch(branch); err != nil {
					return fmt.Errorf("failed to fetch branch: %w", err)
				}
			} else {
				return fmt.Errorf("branch %s does not exist (use -b to create)", branch)
			}
		}
	}

	// For PR branches, we might need to fetch first
	if addPRIdentifier != "" {
		exists, err := git.BranchExists(branch)
		if err != nil {
			return fmt.Errorf("failed to check branch: %w", err)
		}
		if !exists {
			fmt.Printf("Fetching branch %s from origin...\n", branch)
			if err := git.FetchBranch(branch); err != nil {
				return fmt.Errorf("failed to fetch branch: %w", err)
			}
		}
	}

	// Create worktree
	fmt.Printf("Creating worktree at %s for branch %s...\n", wtPath, branch)
	if err := git.Add(wtPath, branch, addCreateBranch); err != nil {
		return err
	}

	fmt.Printf("✓ Worktree created: %s\n", wtPath)
	return nil
}

// selectBranchWithFzf shows an interactive branch selector using fzf
func selectBranchWithFzf() (string, error) {
	// Check if fzf is available
	_, err := exec.LookPath("fzf")
	if err != nil {
		return "", fmt.Errorf("fzf is not installed. Please install fzf for interactive selection, or specify a branch name")
	}

	// Get all branches (local and remote)
	branches, err := git.ListBranches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	if len(branches) == 0 {
		return "", fmt.Errorf("no branches found")
	}

	// Run fzf
	fzfCmd := exec.Command("fzf", "--height=40%", "--reverse", "--prompt=Select branch: ")
	fzfCmd.Stdin = strings.NewReader(strings.Join(branches, "\n"))
	fzfCmd.Stderr = os.Stderr

	out, err := fzfCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			// User pressed Ctrl+C
			return "", nil
		}
		return "", nil // fzf cancelled
	}

	return strings.TrimSpace(string(out)), nil
}
