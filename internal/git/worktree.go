package git

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/shell"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
	IsMain bool
}

// Manager manages git operations with dependency injection
type Manager struct {
	executor shell.Executor
}

// NewManager creates a new Manager with the given executor
func NewManager(executor shell.Executor) *Manager {
	return &Manager{executor: executor}
}

// defaultManager is used for backward compatibility
var defaultManager = NewManager(shell.NewRealExecutor())

// GetRepoRoot returns the root directory of the git repository
func (m *Manager) GetRepoRoot() (string, error) {
	out, err := m.executor.Execute("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", errors.NewNotAGitRepoError(".", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoRoot is a package-level wrapper for backward compatibility
func GetRepoRoot() (string, error) {
	return defaultManager.GetRepoRoot()
}

// GetRepoName returns the name of the repository (directory name)
func (m *Manager) GetRepoName() (string, error) {
	root, err := m.GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// GetRepoName is a package-level wrapper for backward compatibility
func GetRepoName() (string, error) {
	return defaultManager.GetRepoName()
}

// GetMainWorktreePath returns the path to the main (bare) worktree
func (m *Manager) GetMainWorktreePath() (string, error) {
	out, err := m.executor.Execute("git", "rev-parse", "--git-common-dir")
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

// GetMainWorktreePath is a package-level wrapper for backward compatibility
func GetMainWorktreePath() (string, error) {
	return defaultManager.GetMainWorktreePath()
}

// List returns all worktrees for the current repository
func (m *Manager) List() ([]Worktree, error) {
	out, err := m.executor.Execute("git", "worktree", "list", "--porcelain")
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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse worktree list: %w", err)
	}

	// Mark the main worktree
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}

	return worktrees, nil
}

// List is a package-level wrapper for backward compatibility
func List() ([]Worktree, error) {
	return defaultManager.List()
}

// Add creates a new worktree
func (m *Manager) Add(path string, branch string, createBranch bool) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path)
	} else {
		args = append(args, path, branch)
	}

	_, err := m.executor.Execute("git", args...)
	if err != nil {
		return errors.NewCommandExecutionError("git", args, err)
	}
	return nil
}

// Add is a package-level wrapper for backward compatibility
func Add(path string, branch string, createBranch bool) error {
	return defaultManager.Add(path, branch, createBranch)
}

// Remove removes a worktree
func (m *Manager) Remove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	_, err := m.executor.Execute("git", args...)
	if err != nil {
		return errors.NewCommandExecutionError("git", args, err)
	}
	return nil
}

// Remove is a package-level wrapper for backward compatibility
func Remove(path string, force bool) error {
	return defaultManager.Remove(path, force)
}

// Exists checks if a worktree exists for the given path or branch
func (m *Manager) Exists(pathOrBranch string) (*Worktree, error) {
	worktrees, err := m.List()
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

// Exists is a package-level wrapper for backward compatibility
func Exists(pathOrBranch string) (*Worktree, error) {
	return defaultManager.Exists(pathOrBranch)
}

// BranchExists checks if a branch exists
func (m *Manager) BranchExists(branch string) (bool, error) {
	_, err := m.executor.Execute("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil {
		// Check if error has ExitCode method (both *exec.ExitError and our test mock)
		type exitCoder interface {
			ExitCode() int
		}
		if exitErr, ok := err.(exitCoder); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BranchExists is a package-level wrapper for backward compatibility
func BranchExists(branch string) (bool, error) {
	return defaultManager.BranchExists(branch)
}

// RemoteBranchExists checks if a remote branch exists
func (m *Manager) RemoteBranchExists(branch string) (bool, error) {
	out, err := m.executor.Execute("git", "ls-remote", "--heads", "origin", branch)
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

// RemoteBranchExists is a package-level wrapper for backward compatibility
func RemoteBranchExists(branch string) (bool, error) {
	return defaultManager.RemoteBranchExists(branch)
}

// FetchBranch fetches a branch from origin
func (m *Manager) FetchBranch(branch string) error {
	args := []string{"fetch", "origin", branch + ":" + branch}
	_, err := m.executor.Execute("git", args...)
	if err != nil {
		return errors.NewCommandExecutionError("git", args, err)
	}
	return nil
}

// FetchBranch is a package-level wrapper for backward compatibility
func FetchBranch(branch string) error {
	return defaultManager.FetchBranch(branch)
}

// ListBranches returns all local and remote branches
func (m *Manager) ListBranches() ([]string, error) {
	// Get local branches
	localOut, err := m.executor.Execute("git", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	// Get remote branches
	remoteOut, err := m.executor.Execute("git", "branch", "-r", "--format=%(refname:short)")
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

// ListBranches is a package-level wrapper for backward compatibility
func ListBranches() ([]string, error) {
	return defaultManager.ListBranches()
}

// GetCurrentBranch returns the name of the current branch
func (m *Manager) GetCurrentBranch() (string, error) {
	args := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	out, err := m.executor.Execute("git", args...)
	if err != nil {
		return "", errors.NewCommandExecutionError("git", args, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch is a package-level wrapper for backward compatibility
func GetCurrentBranch() (string, error) {
	return defaultManager.GetCurrentBranch()
}

// IsBranchMerged checks if a branch is merged into the current branch
func (m *Manager) IsBranchMerged(branch string) (bool, error) {
	args := []string{"branch", "--merged", "HEAD", "--format=%(refname:short)"}
	out, err := m.executor.Execute("git", args...)
	if err != nil {
		return false, errors.NewCommandExecutionError("git", args, err)
	}

	mergedBranches := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, mergedBranch := range mergedBranches {
		if strings.TrimSpace(mergedBranch) == branch {
			return true, nil
		}
	}
	return false, nil
}

// IsBranchMerged is a package-level wrapper for backward compatibility
func IsBranchMerged(branch string) (bool, error) {
	return defaultManager.IsBranchMerged(branch)
}

// DeleteBranch deletes a git branch
func (m *Manager) DeleteBranch(branchName string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, branchName)

	_, err := m.executor.Execute("git", args...)
	if err != nil {
		return errors.NewCommandExecutionError("git", args, err)
	}
	return nil
}

// DeleteBranch is a package-level wrapper for backward compatibility
func DeleteBranch(branchName string, force bool) error {
	return defaultManager.DeleteBranch(branchName, force)
}

// GetChangedFiles returns all files with differences (modified, staged, untracked) in the specified directory
func (m *Manager) GetChangedFiles(path string) ([]string, error) {
	args := []string{"-C", path, "status", "--porcelain", "-z"}
	out, err := m.executor.Execute("git", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get git status (-C %s): %w", path, err)
	}

	var files []string
	parts := bytes.Split(out, []byte{0})
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if len(p) < 4 {
			continue
		}

		// Parse git status output: XY filename
		// X = index status, Y = working tree status
		filePath := string(p[3:])
		if filePath == "" {
			continue
		}

		// For renamed/copied files in -z mode, the format is:
		// R<space><newpath><NUL><oldpath><NUL>
		// The newpath is already extracted above (p[3:])
		// Skip the next part which is the oldpath
		if p[0] == 'R' || p[0] == 'C' {
			i++ // Skip the oldpath in the next part
		}

		files = append(files, filePath)
	}

	return files, nil
}

// GetChangedFiles is a package-level wrapper for backward compatibility
func GetChangedFiles(path string) ([]string, error) {
	return defaultManager.GetChangedFiles(path)
}

// GetIgnoredFiles returns all gitignored files (including those in global gitignore) in the specified directory
func (m *Manager) GetIgnoredFiles(path string) ([]string, error) {
	args := []string{"-C", path, "ls-files", "--others", "--ignored", "--exclude-standard", "-z"}
	out, err := m.executor.Execute("git", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ignored files (-C %s): %w", path, err)
	}

	var files []string
	parts := bytes.Split(out, []byte{0})
	for _, p := range parts {
		filePath := strings.TrimSpace(string(p))
		if filePath == "" {
			continue
		}
		files = append(files, filePath)
	}

	return files, nil
}

// GetIgnoredFiles is a package-level wrapper for backward compatibility
func GetIgnoredFiles(path string) ([]string, error) {
	return defaultManager.GetIgnoredFiles(path)
}
