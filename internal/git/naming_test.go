package git

import (
	"testing"
)

func TestBranchToSuffix(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "simple branch name",
			branch:   "main",
			expected: "main",
		},
		{
			name:     "branch with forward slash",
			branch:   "feature/hoge",
			expected: "feature-hoge",
		},
		{
			name:     "branch with multiple forward slashes",
			branch:   "feature/sub/hoge",
			expected: "feature-sub-hoge",
		},
		{
			name:     "branch with backslash",
			branch:   "feature\\hoge",
			expected: "feature-hoge",
		},
		{
			name:     "branch with colon",
			branch:   "feature:hoge",
			expected: "feature-hoge",
		},
		{
			name:     "branch with mixed special chars",
			branch:   "feature/sub:hoge\\fuga",
			expected: "feature-sub-hoge-fuga",
		},
		{
			name:     "empty branch",
			branch:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BranchToSuffix(tt.branch)
			if result != tt.expected {
				t.Errorf("BranchToSuffix(%q) = %q, want %q", tt.branch, result, tt.expected)
			}
		})
	}
}

func TestWorktreeDirName(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		branch   string
		expected string
	}{
		{
			name:     "simple branch",
			repoName: "myrepo",
			branch:   "main",
			expected: "myrepo-main",
		},
		{
			name:     "feature branch with slash",
			repoName: "ex-repo",
			branch:   "feature/hoge",
			expected: "ex-repo-feature-hoge",
		},
		{
			name:     "complex branch name",
			repoName: "project",
			branch:   "feature/sub/task",
			expected: "project-feature-sub-task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WorktreeDirName(tt.repoName, tt.branch)
			if result != tt.expected {
				t.Errorf("WorktreeDirName(%q, %q) = %q, want %q", tt.repoName, tt.branch, result, tt.expected)
			}
		})
	}
}

func TestParseWorktreeIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		repoName   string
		expected   string
	}{
		{
			name:       "branch name with slash",
			identifier: "feature/hoge",
			repoName:   "ex-repo",
			expected:   "ex-repo-feature-hoge",
		},
		{
			name:       "suffix format",
			identifier: "feature-hoge",
			repoName:   "ex-repo",
			expected:   "ex-repo-feature-hoge",
		},
		{
			name:       "full dir name already",
			identifier: "ex-repo-feature-hoge",
			repoName:   "ex-repo",
			expected:   "ex-repo-feature-hoge",
		},
		{
			name:       "absolute path",
			identifier: "/path/to/ex-repo-feature-hoge",
			repoName:   "ex-repo",
			expected:   "ex-repo-feature-hoge",
		},
		{
			name:       "simple branch name",
			identifier: "main",
			repoName:   "myrepo",
			expected:   "myrepo-main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWorktreeIdentifier(tt.identifier, tt.repoName)
			if result != tt.expected {
				t.Errorf("ParseWorktreeIdentifier(%q, %q) = %q, want %q", tt.identifier, tt.repoName, result, tt.expected)
			}
		})
	}
}
