package cmd

import (
	"fmt"

	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/github"
)

// addOptions contains options for worktree creation
type addOptions struct {
	branch       string
	createBranch bool
	prIdentifier string
	selector     fzf.Selector
}

// determineBranch determines which branch to use based on args and options
func determineBranch(args []string, opts *addOptions, repoName string) (string, error) {
	// Handle PR flag
	if opts.prIdentifier != "" {
		branch, err := getBranchFromPR(opts.prIdentifier, repoName)
		if err != nil {
			return "", err
		}
		return branch, nil
	}

	// Interactive selection if no args
	if len(args) == 0 {
		branch, err := selectBranchInteractive(opts.selector)
		if err != nil {
			return "", err
		}
		return branch, nil
	}

	// Use provided branch name
	return args[0], nil
}

// getBranchFromPR retrieves branch name from PR identifier
func getBranchFromPR(prIdentifier, repoName string) (string, error) {
	branch, err := github.GetPRBranch(prIdentifier, repoName)
	if err != nil {
		return "", fmt.Errorf("failed to get PR branch: %w", err)
	}
	return branch, nil
}

// selectBranchInteractive shows interactive branch selector
func selectBranchInteractive(selector fzf.Selector) (string, error) {
	branches, err := git.ListBranches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	branch, err := selector.SelectBranch(branches)
	if err != nil {
		return "", err
	}

	return branch, nil
}

// checkExistingWorktree checks if worktree already exists for the branch
func checkExistingWorktree(branch string) (*git.Worktree, error) {
	existing, err := git.FindWorktree(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing worktree: %w", err)
	}
	return existing, nil
}

// ensureBranchExists checks and fetches branch if necessary
func ensureBranchExists(branch string, createBranch bool, fromPR bool) error {
	// If creating a new branch (and not from PR), it will be created with worktree
	if createBranch && !fromPR {
		return nil
	}

	// Check if branch exists locally
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
		} else if !createBranch {
			return fmt.Errorf("branch %s does not exist (use -b to create)", branch)
		}
	}

	return nil
}

// createWorktree creates a new worktree for the given branch
func createWorktree(repoName, branch string, createBranch bool) error {
	wtPath, err := git.WorktreePath(repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	fmt.Printf("Creating worktree at %s for branch %s...\n", wtPath, branch)
	if err := git.Add(wtPath, branch, createBranch); err != nil {
		return err
	}

	fmt.Printf("✓ Worktree created: %s\n", wtPath)
	return nil
}
