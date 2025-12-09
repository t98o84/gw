package git

import (
	"path/filepath"
	"strings"
)

// BranchToSuffix converts a branch name to a directory suffix
// e.g., "feature/hoge" -> "feature-hoge"
func BranchToSuffix(branch string) string {
	// Replace / with -
	suffix := strings.ReplaceAll(branch, "/", "-")
	// Replace any other problematic characters
	suffix = strings.ReplaceAll(suffix, "\\", "-")
	suffix = strings.ReplaceAll(suffix, ":", "-")
	return suffix
}

// WorktreePath generates the worktree directory path
// e.g., for repo "ex-repo" and branch "feature/hoge" -> "../ex-repo-feature-hoge"
func (m *Manager) WorktreePath(repoName, branch string) (string, error) {
	repoRoot, err := m.GetRepoRoot()
	if err != nil {
		return "", err
	}

	suffix := BranchToSuffix(branch)
	dirName := repoName + "-" + suffix
	return filepath.Join(filepath.Dir(repoRoot), dirName), nil
}

// WorktreePath is a package-level wrapper for backward compatibility
func WorktreePath(repoName, branch string) (string, error) {
	return defaultManager.WorktreePath(repoName, branch)
}

// WorktreeDirName generates just the directory name for a worktree
func WorktreeDirName(repoName, branch string) string {
	suffix := BranchToSuffix(branch)
	return repoName + "-" + suffix
}

// ParseWorktreeIdentifier parses various forms of worktree identifier
// and returns the matching worktree if found
// Supported formats:
//   - branch name: "feature/hoge"
//   - suffix: "feature-hoge"
//   - full dir name: "ex-repo-feature-hoge"
//   - full path: "/path/to/ex-repo-feature-hoge"
func ParseWorktreeIdentifier(identifier string, repoName string) string {
	// If it's already a full path, return basename
	if filepath.IsAbs(identifier) {
		return filepath.Base(identifier)
	}

	// If it starts with repo name, it's already a full dir name
	if strings.HasPrefix(identifier, repoName+"-") {
		return identifier
	}

	// Convert branch format to suffix format
	suffix := BranchToSuffix(identifier)

	// Return full dir name
	return repoName + "-" + suffix
}

// FindWorktree finds a worktree by various identifier formats
func (m *Manager) FindWorktree(identifier string) (*Worktree, error) {
	repoName, err := m.GetRepoName()
	if err != nil {
		return nil, err
	}

	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	// Normalize the identifier
	targetDirName := ParseWorktreeIdentifier(identifier, repoName)

	for _, wt := range worktrees {
		dirName := filepath.Base(wt.Path)
		// Match by directory name
		if dirName == targetDirName {
			return &wt, nil
		}
		// Also try matching by branch name directly
		if wt.Branch == identifier {
			return &wt, nil
		}
		// Match by suffix (without repo name prefix)
		if dirName == identifier {
			return &wt, nil
		}
	}

	return nil, nil
}

// FindWorktree is a package-level wrapper for backward compatibility
func FindWorktree(identifier string) (*Worktree, error) {
	return defaultManager.FindWorktree(identifier)
}
