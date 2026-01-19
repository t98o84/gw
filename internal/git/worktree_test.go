package git

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/t98o84/gw/internal/shell"
)

func TestWorktree_IsMainField(t *testing.T) {
	wt := Worktree{
		Path:   "/path/to/repo",
		Branch: "main",
		Commit: "abc1234",
		IsMain: true,
	}

	if !wt.IsMain {
		t.Error("Expected IsMain to be true")
	}

	wt2 := Worktree{
		Path:   "/path/to/repo-feature",
		Branch: "feature/hoge",
		Commit: "def5678",
		IsMain: false,
	}

	if wt2.IsMain {
		t.Error("Expected IsMain to be false")
	}
}

func TestWorktree_Fields(t *testing.T) {
	wt := Worktree{
		Path:   "/path/to/repo",
		Branch: "feature/test",
		Commit: "abc123def456",
		IsMain: false,
	}

	if wt.Path != "/path/to/repo" {
		t.Errorf("Expected Path to be '/path/to/repo', got %q", wt.Path)
	}
	if wt.Branch != "feature/test" {
		t.Errorf("Expected Branch to be 'feature/test', got %q", wt.Branch)
	}
	if wt.Commit != "abc123def456" {
		t.Errorf("Expected Commit to be 'abc123def456', got %q", wt.Commit)
	}
}

func TestManager_GetRepoRoot(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    string
		wantErr bool
	}{
		{
			name: "success",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && len(args) == 2 && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
						return []byte("/path/to/repo\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "/path/to/repo",
			wantErr: false,
		},
		{
			name: "not a git repo",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, &exec.ExitError{}
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.GetRepoRoot()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetRepoRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.GetRepoRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetCurrentBranch(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    string
		wantErr bool
	}{
		{
			name: "current branch success",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--abbrev-ref" && args[2] == "HEAD" {
						return []byte("feature/test\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "feature/test",
			wantErr: false,
		},
		{
			name: "git command fails",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.GetCurrentBranch()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetCurrentBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.GetCurrentBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_IsBranchMerged(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		mock    *shell.MockExecutor
		want    bool
		wantErr bool
	}{
		{
			name:   "branch is merged",
			branch: "feature/test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" && args[1] == "--merged" {
						return []byte("main\nfeature/test\nfeature/other\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "branch is not merged",
			branch: "feature/unmerged",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" && args[1] == "--merged" {
						return []byte("main\nfeature/test\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "git command fails",
			branch: "feature/test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.IsBranchMerged(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.IsBranchMerged() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.IsBranchMerged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_DeleteBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		force      bool
		mock       *shell.MockExecutor
		wantErr    bool
	}{
		{
			name:       "delete branch without force",
			branchName: "feature/test",
			force:      false,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" && args[1] == "-d" && args[2] == "feature/test" {
						return []byte(""), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:       "delete branch with force",
			branchName: "feature/test",
			force:      true,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" && args[1] == "-D" && args[2] == "feature/test" {
						return []byte(""), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:       "delete branch fails",
			branchName: "feature/test",
			force:      false,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("branch not merged")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			err := m.DeleteBranch(tt.branchName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.DeleteBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_GetRepoName(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    string
		wantErr bool
	}{
		{
			name: "success",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" {
						return []byte("/path/to/my-repo\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "my-repo",
			wantErr: false,
		},
		{
			name: "GetRepoRoot fails",
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
			got, err := m.GetRepoName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetRepoName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.GetRepoName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_List(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    []Worktree
		wantErr bool
	}{
		{
			name: "single worktree",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want: []Worktree{
				{Path: "/path/to/repo", Branch: "main", Commit: "abc123", IsMain: true},
			},
			wantErr: false,
		},
		{
			name: "multiple worktrees",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /path/to/repo-feature\nHEAD def456\nbranch refs/heads/feature/test\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want: []Worktree{
				{Path: "/path/to/repo", Branch: "main", Commit: "abc123", IsMain: true},
				{Path: "/path/to/repo-feature", Branch: "feature/test", Commit: "def456", IsMain: false},
			},
			wantErr: false,
		},
		{
			name: "command fails",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git command failed")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Manager.List() returned %d worktrees, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Path != tt.want[i].Path || got[i].Branch != tt.want[i].Branch || got[i].Commit != tt.want[i].Commit || got[i].IsMain != tt.want[i].IsMain {
					t.Errorf("Manager.List()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestManager_BranchExists(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		mock    *shell.MockExecutor
		want    bool
		wantErr bool
	}{
		{
			name:   "branch exists",
			branch: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "show-ref" {
						return []byte(""), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "branch does not exist",
			branch: "nonexistent",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "show-ref" {
						// Simulate exit code 1 for non-existent branch
						return nil, &testExitError{exitCode: 1}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "git command error",
			branch: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.BranchExists(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.BranchExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.BranchExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_RemoteBranchExists(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		mock    *shell.MockExecutor
		want    bool
		wantErr bool
	}{
		{
			name:   "remote branch exists",
			branch: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "ls-remote" {
						return []byte("abc123\trefs/heads/main\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "remote branch does not exist",
			branch: "nonexistent",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "ls-remote" {
						return []byte(""), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "git command error",
			branch: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.RemoteBranchExists(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.RemoteBranchExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.RemoteBranchExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// testExitError is a mock ExitError implementation for testing
type testExitError struct {
	exitCode int
}

func (e *testExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.exitCode)
}

func (e *testExitError) ExitCode() int {
	return e.exitCode
}

func TestManager_GetMainWorktreePath(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    string
		wantErr bool
	}{
		{
			name: "main worktree with .git",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" && args[1] == "--git-common-dir" {
						return []byte("/path/to/repo/.git\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "/path/to/repo",
			wantErr: false,
		},
		{
			name: "worktree subdirectory",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "rev-parse" {
						return []byte("/path/to/repo/.git/worktrees/feature\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    "/path/to/repo/.git/worktrees",
			wantErr: false,
		},
		{
			name: "git command fails",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.GetMainWorktreePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetMainWorktreePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.GetMainWorktreePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Add(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		branch       string
		createBranch bool
		from         string
		mock         *shell.MockExecutor
		wantErr      bool
	}{
		{
			name:         "add with existing branch",
			path:         "/path/to/worktree",
			branch:       "feature/test",
			createBranch: false,
			from:         "",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" && args[1] == "add" {
						if len(args) == 4 && args[2] == "/path/to/worktree" && args[3] == "feature/test" {
							return []byte(""), nil
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:         "add with new branch",
			path:         "/path/to/worktree",
			branch:       "feature/new",
			createBranch: true,
			from:         "",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" && args[1] == "add" {
						if len(args) == 5 && args[2] == "-b" && args[3] == "feature/new" && args[4] == "/path/to/worktree" {
							return []byte(""), nil
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:         "add with new branch from specific commit",
			path:         "/path/to/worktree",
			branch:       "feature/new",
			createBranch: true,
			from:         "origin/main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" && args[1] == "add" {
						if len(args) == 6 && args[2] == "-b" && args[3] == "feature/new" && args[4] == "/path/to/worktree" && args[5] == "origin/main" {
							return []byte(""), nil
						}
					}
					return nil, fmt.Errorf("unexpected command: %v", args)
				},
			},
			wantErr: false,
		},
		{
			name:         "add fails",
			path:         "/path/to/worktree",
			branch:       "feature/test",
			createBranch: false,
			from:         "",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("failed to add worktree")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			err := m.Add(tt.path, tt.branch, tt.createBranch, tt.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_Remove(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		force   bool
		mock    *shell.MockExecutor
		wantErr bool
	}{
		{
			name:  "remove without force",
			path:  "/path/to/worktree",
			force: false,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" && args[1] == "remove" {
						if len(args) == 3 && args[2] == "/path/to/worktree" {
							return []byte(""), nil
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:  "remove with force",
			path:  "/path/to/worktree",
			force: true,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" && args[1] == "remove" {
						if len(args) == 4 && args[2] == "--force" && args[3] == "/path/to/worktree" {
							return []byte(""), nil
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:  "remove fails",
			path:  "/path/to/worktree",
			force: false,
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("failed to remove worktree")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			err := m.Remove(tt.path, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_FetchBranch(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		mock    *shell.MockExecutor
		wantErr bool
	}{
		{
			name:   "fetch success",
			branch: "feature/test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "fetch" && args[1] == "origin" && args[2] == "feature/test:feature/test" {
						return []byte(""), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantErr: false,
		},
		{
			name:   "fetch fails",
			branch: "feature/test",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("fetch error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			err := m.FetchBranch(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.FetchBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_ListBranches(t *testing.T) {
	tests := []struct {
		name    string
		mock    *shell.MockExecutor
		want    []string
		wantErr bool
	}{
		{
			name: "list branches success",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" {
						if len(args) == 2 && args[1] == "--format=%(refname:short)" {
							// Local branches
							return []byte("main\nfeature/test\n"), nil
						}
						if len(args) == 3 && args[1] == "-r" {
							// Remote branches
							return []byte("origin/main\norigin/feature/remote\norigin/HEAD -> origin/main\n"), nil
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    []string{"main", "feature/test", "feature/remote"},
			wantErr: false,
		},
		{
			name: "local branch command fails",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" && len(args) == 2 {
						return nil, fmt.Errorf("git error")
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "remote branch command fails",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "branch" {
						if len(args) == 2 {
							return []byte("main\n"), nil
						}
						if len(args) == 3 && args[1] == "-r" {
							return nil, fmt.Errorf("git error")
						}
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.ListBranches()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.ListBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Manager.ListBranches() returned %d branches, want %d", len(got), len(tt.want))
					return
				}
				for i, branch := range tt.want {
					if got[i] != branch {
						t.Errorf("Manager.ListBranches()[%d] = %v, want %v", i, got[i], branch)
					}
				}
			}
		})
	}
}

func TestManager_Exists(t *testing.T) {
	tests := []struct {
		name         string
		identifier   string
		mock         *shell.MockExecutor
		wantWorktree *Worktree
		wantErr      bool
	}{
		{
			name:       "exists by path",
			identifier: "/path/to/repo",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/repo", Branch: "main", Commit: "abc123", IsMain: true},
			wantErr:      false,
		},
		{
			name:       "exists by branch",
			identifier: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/repo", Branch: "main", Commit: "abc123", IsMain: true},
			wantErr:      false,
		},
		{
			name:       "exists by basename",
			identifier: "repo",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: &Worktree{Path: "/path/to/repo", Branch: "main", Commit: "abc123", IsMain: true},
			wantErr:      false,
		},
		{
			name:       "does not exist",
			identifier: "nonexistent",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					if name == "git" && args[0] == "worktree" {
						return []byte("worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\n"), nil
					}
					return nil, fmt.Errorf("unexpected command")
				},
			},
			wantWorktree: nil,
			wantErr:      false,
		},
		{
			name:       "list fails",
			identifier: "main",
			mock: &shell.MockExecutor{
				ExecuteFunc: func(name string, args ...string) ([]byte, error) {
					return nil, fmt.Errorf("git error")
				},
			},
			wantWorktree: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.mock)
			got, err := m.Exists(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantWorktree == nil {
				if got != nil {
					t.Errorf("Manager.Exists() = %+v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Manager.Exists() = nil, want %+v", tt.wantWorktree)
					return
				}
				if got.Path != tt.wantWorktree.Path || got.Branch != tt.wantWorktree.Branch ||
					got.Commit != tt.wantWorktree.Commit || got.IsMain != tt.wantWorktree.IsMain {
					t.Errorf("Manager.Exists() = %+v, want %+v", got, tt.wantWorktree)
				}
			}
		})
	}
}
