package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
	IsMain bool
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoName returns the name of the repository (directory name)
func GetRepoName() (string, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// GetMainWorktreePath returns the path to the main (bare) worktree
func GetMainWorktreePath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git common dir: %w", err)
	}
	gitDir := strings.TrimSpace(string(out))

	// If it's a bare repo or main worktree, git-common-dir returns .git or the repo path
	// We need to get the parent of .git
	if filepath.Base(gitDir) == ".git" {
		return filepath.Dir(gitDir), nil
	}

	// For worktrees, git-common-dir returns the path to the main repo's .git
	return filepath.Dir(gitDir), nil
}

// List returns all worktrees for the current repository
func List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
		}
	}
	// Add the last worktree if exists
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	// Mark the main worktree
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}

	return worktrees, nil
}

// Add creates a new worktree
func Add(path string, branch string, createBranch bool) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if !createBranch {
		args = append(args, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}
	return nil
}

// Remove removes a worktree
func Remove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

// Exists checks if a worktree exists for the given path or branch
func Exists(pathOrBranch string) (*Worktree, error) {
	worktrees, err := List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Path == pathOrBranch || wt.Branch == pathOrBranch || filepath.Base(wt.Path) == pathOrBranch {
			return &wt, nil
		}
	}
	return nil, nil
}

// BranchExists checks if a branch exists
func BranchExists(branch string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// RemoteBranchExists checks if a remote branch exists
func RemoteBranchExists(branch string) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

// FetchBranch fetches a branch from origin
func FetchBranch(branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch+":"+branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListBranches returns all local and remote branches
func ListBranches() ([]string, error) {
	// Get local branches
	localCmd := exec.Command("git", "branch", "--format=%(refname:short)")
	localOut, err := localCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	// Get remote branches
	remoteCmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	remoteOut, err := remoteCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	branchSet := make(map[string]bool)
	var branches []string

	// Add local branches
	for _, line := range strings.Split(strings.TrimSpace(string(localOut)), "\n") {
		if line != "" && !branchSet[line] {
			branchSet[line] = true
			branches = append(branches, line)
		}
	}

	// Add remote branches (strip origin/ prefix)
	for _, line := range strings.Split(strings.TrimSpace(string(remoteOut)), "\n") {
		if line == "" {
			continue
		}
		// Skip HEAD pointer
		if strings.Contains(line, "HEAD") {
			continue
		}
		// Strip origin/ prefix
		branch := strings.TrimPrefix(line, "origin/")
		if !branchSet[branch] {
			branchSet[branch] = true
			branches = append(branches, branch)
		}
	}

	return branches, nil
}
