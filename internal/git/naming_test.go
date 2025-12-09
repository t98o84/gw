package git

import (
	"fmt"
	"testing"

	"github.com/t98o84/gw/internal/shell"
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

func TestManager_WorktreePath(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		branch   string
		mock     *shell.MockExecutor
		want     string
		wantErr  bool
	}{
		{
			name:     "simple branch",
			repoName: "myrepo",
			branch:   "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "/path/to/myrepo-main",
			wantErr: false,
		},
		{
			name:     "feature branch with slash",
			repoName: "ex-repo",
			branch:   "feature/hoge",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" {
						return []byte("/home/user/repos/ex-repo\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "/home/user/repos/ex-repo-feature-hoge",
			wantErr: false,
		},
		{
			name:     "GetRepoRoot fails",
			repoName: "myrepo",
			branch:   "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("not a git repo")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.WorktreePath(tt.repoName, tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.WorktreePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.WorktreePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_FindWorktree(t *testing.T) {
	tests := []struct {
		name         string
		identifier   string
		mock         *shell.MockExecutor
		wantWorktree *Worktree
		wantErr      bool
	}{
		{
			name:       "find by branch name",
			identifier: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/myrepo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/myrepo", Branch: "main", Commit: "abc123", IsMain: true},
			wantErr:      false,
		},
		{
			name:       "find by directory name",
			identifier: "myrepo-feature-test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/myrepo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /path/to/myrepo-feature-test\nHEAD def456\nbranch refs/heads/feature/test\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/myrepo-feature-test", Branch: "feature/test", Commit: "def456", IsMain: false},
			wantErr:      false,
		},
		{
			name:       "find by branch with slash",
			identifier: "feature/test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/myrepo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /path/to/myrepo-feature-test\nHEAD def456\nbranch refs/heads/feature/test\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/myrepo-feature-test", Branch: "feature/test", Commit: "def456", IsMain: false},
			wantErr:      false,
		},
		{
			name:       "worktree not found",
			identifier: "nonexistent",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/myrepo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: nil,
			wantErr:      false,
		},
		{
			name:       "GetRepoName fails",
			identifier: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			wantWorktree: nil,
			wantErr:      true,
		},
		{
			name:       "List fails",
			identifier: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/myrepo\n"), nil
					}
					if name == "git" && args[0] == "worktree" {
						return nil, fmt.Errorf("git error")
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.FindWorktree(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.FindWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantWorktree == nil {
				if got != nil {
					t.Errorf("Manager.FindWorktree() = %+v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Manager.FindWorktree() = nil, want %+v", tt.wantWorktree)
					return
				}
				if got.Path != tt.wantWorktree.Path || got.Branch != tt.wantWorktree.Branch ||
					got.Commit != tt.wantWorktree.Commit || got.IsMain != tt.wantWorktree.IsMain {
					t.Errorf("Manager.FindWorktree() = %+v, want %+v", got, tt.wantWorktree)
				}
			}
		})
	}
}
