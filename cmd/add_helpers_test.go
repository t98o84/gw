package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
)

// setupMocks initializes all mock functions to their default implementations
func setupMocks() {
	mockGetPRBranch = nil
	mockListBranches = nil
	mockFindWorktree = nil
	mockBranchExists = nil
	mockRemoteBranchExists = nil
	mockFetchBranch = nil
	mockWorktreePath = nil
	mockAdd = nil
	mockOpenInEditor = nil
}

// resetMocks is called after each test
func resetMocks() {
	setupMocks()
}

// mockSelector for testing
type mockSelector struct {
	selectBranchFunc    func(branches []string) (string, error)
	selectWorktreeFunc  func(worktrees []*git.Worktree, excludeMain bool) (*git.Worktree, error)
	selectWorktreesFunc func(worktrees []*git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error)
	isAvailableFunc     func() bool
}

func (m *mockSelector) SelectBranch(branches []string) (string, error) {
	if m.selectBranchFunc != nil {
		return m.selectBranchFunc(branches)
	}
	if len(branches) > 0 {
		return branches[0], nil
	}
	return "", nil
}

func (m *mockSelector) SelectWorktree(worktrees []*git.Worktree, excludeMain bool) (*git.Worktree, error) {
	if m.selectWorktreeFunc != nil {
		return m.selectWorktreeFunc(worktrees, excludeMain)
	}
	for _, wt := range worktrees {
		if !excludeMain || !wt.IsMain {
			return wt, nil
		}
	}
	return nil, nil
}

func (m *mockSelector) SelectWorktrees(worktrees []*git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error) {
	if m.selectWorktreesFunc != nil {
		return m.selectWorktreesFunc(worktrees, excludeMain, multi)
	}
	var selected []*git.Worktree
	for _, wt := range worktrees {
		if !excludeMain || !wt.IsMain {
			selected = append(selected, wt)
			if !multi {
				break
			}
		}
	}
	return selected, nil
}

func (m *mockSelector) IsAvailable() bool {
	if m.isAvailableFunc != nil {
		return m.isAvailableFunc()
	}
	return true
}

// TestDetermineBranch tests the determineBranch function
func TestDetermineBranch(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		opts        *addOptions
		repoName    string
		setupMock   func()
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:     "PR flag specified - success",
			args:     []string{},
			opts:     &addOptions{prIdentifier: "123"},
			repoName: "test-repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					if prIdentifier == "123" && repoName == "test-repo" {
						return "feature/pr-branch", nil
					}
					return "", errors.New("unexpected call")
				}
			},
			want:    "feature/pr-branch",
			wantErr: false,
		},
		{
			name:     "PR flag specified - error",
			args:     []string{},
			opts:     &addOptions{prIdentifier: "invalid"},
			repoName: "test-repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					return "", errors.New("PR not found")
				}
			},
			wantErr:     true,
			errContains: "PR not found",
		},
		{
			name: "interactive selection - success",
			args: []string{},
			opts: &addOptions{
				selector: &mockSelector{
					selectBranchFunc: func(branches []string) (string, error) {
						return "selected-branch", nil
					},
				},
			},
			repoName: "test-repo",
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{"main", "feature/test", "selected-branch"}, nil
				}
			},
			want:    "selected-branch",
			wantErr: false,
		},
		{
			name: "interactive selection - list branches error",
			args: []string{},
			opts: &addOptions{
				selector: &mockSelector{},
			},
			repoName: "test-repo",
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return nil, errors.New("git error")
				}
			},
			wantErr:     true,
			errContains: "git error",
		},
		{
			name: "interactive selection - selector error",
			args: []string{},
			opts: &addOptions{
				selector: &mockSelector{
					selectBranchFunc: func(branches []string) (string, error) {
						return "", errors.New("user cancelled")
					},
				},
			},
			repoName: "test-repo",
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{"main", "feature/test"}, nil
				}
			},
			wantErr:     true,
			errContains: "user cancelled",
		},
		{
			name:      "branch name provided in args",
			args:      []string{"my-branch"},
			opts:      &addOptions{},
			repoName:  "test-repo",
			setupMock: func() {},
			want:      "my-branch",
			wantErr:   false,
		},
		{
			name:      "multiple args - uses first",
			args:      []string{"first-branch", "second-branch"},
			opts:      &addOptions{},
			repoName:  "test-repo",
			setupMock: func() {},
			want:      "first-branch",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			got, err := determineBranch(tt.args, tt.opts, tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("determineBranch() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("determineBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetBranchFromPR tests the getBranchFromPR function
func TestGetBranchFromPR(t *testing.T) {
	tests := []struct {
		name         string
		prIdentifier string
		repoName     string
		setupMock    func()
		want         string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "successful PR branch retrieval",
			prIdentifier: "123",
			repoName:     "test-repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					return "feature/pr-123", nil
				}
			},
			want:    "feature/pr-123",
			wantErr: false,
		},
		{
			name:         "PR not found",
			prIdentifier: "999",
			repoName:     "test-repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					return "", errors.New("PR not found")
				}
			},
			wantErr: true,
		},
		{
			name:         "invalid PR identifier",
			prIdentifier: "invalid",
			repoName:     "test-repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					return "", errors.New("invalid PR identifier")
				}
			},
			wantErr: true,
		},
		{
			name:         "PR URL format",
			prIdentifier: "https://github.com/owner/repo/pull/456",
			repoName:     "repo",
			setupMock: func() {
				mockGetPRBranch = func(prIdentifier, repoName string) (string, error) {
					return "feature/url-pr", nil
				}
			},
			want:    "feature/url-pr",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			got, err := getBranchFromPR(tt.prIdentifier, tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBranchFromPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("getBranchFromPR() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("getBranchFromPR() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSelectBranchInteractive tests the selectBranchInteractive function
func TestSelectBranchInteractive(t *testing.T) {
	tests := []struct {
		name        string
		selector    fzf.Selector
		setupMock   func()
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful branch selection",
			selector: &mockSelector{
				selectBranchFunc: func(branches []string) (string, error) {
					if len(branches) == 3 {
						return "feature/selected", nil
					}
					return "", errors.New("unexpected branches")
				},
			},
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{"main", "develop", "feature/selected"}, nil
				}
			},
			want:    "feature/selected",
			wantErr: false,
		},
		{
			name:     "list branches error",
			selector: &mockSelector{},
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return nil, errors.New("git command failed")
				}
			},
			wantErr:     true,
			errContains: "failed to list branches",
		},
		{
			name: "empty branch list",
			selector: &mockSelector{
				selectBranchFunc: func(branches []string) (string, error) {
					if len(branches) == 0 {
						return "", errors.New("no branches")
					}
					return branches[0], nil
				},
			},
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{}, nil
				}
			},
			wantErr:     true,
			errContains: "no branches",
		},
		{
			name: "user cancelled selection",
			selector: &mockSelector{
				selectBranchFunc: func(branches []string) (string, error) {
					return "", errors.New("selection cancelled")
				},
			},
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{"main", "develop"}, nil
				}
			},
			wantErr:     true,
			errContains: "selection cancelled",
		},
		{
			name:     "selector returns first branch by default",
			selector: &mockSelector{},
			setupMock: func() {
				mockListBranches = func() ([]string, error) {
					return []string{"main", "develop", "feature/test"}, nil
				}
			},
			want:    "main",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			got, err := selectBranchInteractive(tt.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectBranchInteractive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("selectBranchInteractive() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("selectBranchInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckExistingWorktree tests the checkExistingWorktree function
func TestCheckExistingWorktree(t *testing.T) {
	tests := []struct {
		name        string
		branch      string
		setupMock   func()
		want        *git.Worktree
		wantErr     bool
		errContains string
	}{
		{
			name:   "worktree exists",
			branch: "feature/test",
			setupMock: func() {
				mockFindWorktree = func(branch string) (*git.Worktree, error) {
					return &git.Worktree{
						Path:   "/path/to/repo-feature-test",
						Branch: "feature/test",
						IsMain: false,
					}, nil
				}
			},
			want: &git.Worktree{
				Path:   "/path/to/repo-feature-test",
				Branch: "feature/test",
				IsMain: false,
			},
			wantErr: false,
		},
		{
			name:   "worktree does not exist",
			branch: "feature/new",
			setupMock: func() {
				mockFindWorktree = func(branch string) (*git.Worktree, error) {
					return nil, nil
				}
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "error checking worktree",
			branch: "feature/error",
			setupMock: func() {
				mockFindWorktree = func(branch string) (*git.Worktree, error) {
					return nil, errors.New("git worktree list failed")
				}
			},
			wantErr:     true,
			errContains: "failed to check existing worktree",
		},
		{
			name:   "main worktree",
			branch: "main",
			setupMock: func() {
				mockFindWorktree = func(branch string) (*git.Worktree, error) {
					return &git.Worktree{
						Path:   "/path/to/repo",
						Branch: "main",
						IsMain: true,
					}, nil
				}
			},
			want: &git.Worktree{
				Path:   "/path/to/repo",
				Branch: "main",
				IsMain: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			got, err := checkExistingWorktree(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkExistingWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("checkExistingWorktree() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr {
				if (got == nil) != (tt.want == nil) {
					t.Errorf("checkExistingWorktree() = %v, want %v", got, tt.want)
				}
				if got != nil && tt.want != nil {
					if got.Path != tt.want.Path || got.Branch != tt.want.Branch || got.IsMain != tt.want.IsMain {
						t.Errorf("checkExistingWorktree() = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

// TestEnsureBranchExists tests the ensureBranchExists function
func TestEnsureBranchExists(t *testing.T) {
	tests := []struct {
		name         string
		branch       string
		createBranch bool
		fromPR       bool
		setupMock    func()
		wantErr      bool
		errContains  string
	}{
		{
			name:         "create new branch - not from PR",
			branch:       "new-branch",
			createBranch: true,
			fromPR:       false,
			setupMock:    func() {},
			wantErr:      false,
		},
		{
			name:         "branch exists locally",
			branch:       "existing-branch",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return true, nil
				}
			},
			wantErr: false,
		},
		{
			name:         "branch exists on remote - fetch success",
			branch:       "remote-branch",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
				mockRemoteBranchExists = func(branch string) (bool, error) {
					return true, nil
				}
				mockFetchBranch = func(branch string) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:         "branch exists on remote - fetch error",
			branch:       "remote-branch",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
				mockRemoteBranchExists = func(branch string) (bool, error) {
					return true, nil
				}
				mockFetchBranch = func(branch string) error {
					return errors.New("fetch failed")
				}
			},
			wantErr:     true,
			errContains: "fetch failed",
		},
		{
			name:         "branch does not exist - no create flag",
			branch:       "nonexistent",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
				mockRemoteBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:         "branch does not exist - with create flag from PR",
			branch:       "pr-branch",
			createBranch: true,
			fromPR:       true,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
				mockRemoteBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
			},
			wantErr: false,
		},
		{
			name:         "error checking local branch",
			branch:       "error-branch",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, errors.New("git show-ref failed")
				}
			},
			wantErr:     true,
			errContains: "failed to check branch",
		},
		{
			name:         "error checking remote branch",
			branch:       "error-remote",
			createBranch: false,
			fromPR:       false,
			setupMock: func() {
				mockBranchExists = func(branch string) (bool, error) {
					return false, nil
				}
				mockRemoteBranchExists = func(branch string) (bool, error) {
					return false, errors.New("git ls-remote failed")
				}
			},
			wantErr:     true,
			errContains: "failed to check remote branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			err := ensureBranchExists(tt.branch, tt.createBranch, tt.fromPR)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureBranchExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ensureBranchExists() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

// TestCreateWorktree tests the createWorktree function
func TestCreateWorktree(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		branch       string
		createBranch bool
		openEditor   string
		setupMock    func()
		wantErr      bool
		errContains  string
	}{
		{
			name:         "successful worktree creation",
			repoName:     "test-repo",
			branch:       "feature/test",
			createBranch: false,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-feature-test", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					if path == "/path/to/test-repo-feature-test" && branch == "feature/test" {
						return nil
					}
					return errors.New("unexpected parameters")
				}
			},
			wantErr: false,
		},
		{
			name:         "successful worktree creation with editor",
			repoName:     "test-repo",
			branch:       "feature/test",
			createBranch: false,
			openEditor:   "code",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-feature-test", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					if path == "/path/to/test-repo-feature-test" && branch == "feature/test" {
						return nil
					}
					return errors.New("unexpected parameters")
				}
				mockOpenInEditor = func(editor, path string) error {
					if editor == "code" && path == "/path/to/test-repo-feature-test" {
						return nil
					}
					return errors.New("unexpected parameters")
				}
			},
			wantErr: false,
		},
		{
			name:         "worktree creation succeeds even if editor fails",
			repoName:     "test-repo",
			branch:       "feature/test",
			createBranch: false,
			openEditor:   "invalid-editor",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-feature-test", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					return nil
				}
				mockOpenInEditor = func(editor, path string) error {
					return errors.New("editor not found")
				}
			},
			wantErr: false, // Should not fail even if editor fails
		},
		{
			name:         "worktree path generation error",
			repoName:     "test-repo",
			branch:       "feature/test",
			createBranch: false,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "", errors.New("failed to get repo root")
				}
			},
			wantErr:     true,
			errContains: "failed to generate worktree path",
		},
		{
			name:         "worktree creation error",
			repoName:     "test-repo",
			branch:       "feature/test",
			createBranch: false,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-feature-test", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					return errors.New("git worktree add failed")
				}
			},
			wantErr:     true,
			errContains: "git worktree add failed",
		},
		{
			name:         "create new branch with worktree",
			repoName:     "test-repo",
			branch:       "new-feature",
			createBranch: true,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-new-feature", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					if createBranch && path == "/path/to/test-repo-new-feature" && branch == "new-feature" {
						return nil
					}
					return errors.New("unexpected parameters")
				}
			},
			wantErr: false,
		},
		{
			name:         "branch with special characters",
			repoName:     "my-repo",
			branch:       "feature/fix-bug#123",
			createBranch: false,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/my-repo-feature-fix-bug#123", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					return nil
				}
			},
			wantErr: false,
		}, {
			name:         "create new branch from specific commit",
			repoName:     "test-repo",
			branch:       "feature/new",
			createBranch: true,
			openEditor:   "",
			setupMock: func() {
				mockWorktreePath = func(repoName, branch string) (string, error) {
					return "/path/to/test-repo-feature-new", nil
				}
				mockAdd = func(path string, branch string, createBranch bool, from string) error {
					if from != "origin/main" {
						return fmt.Errorf("expected from to be origin/main, got %s", from)
					}
					if !createBranch {
						return errors.New("expected createBranch to be true")
					}
					return nil
				}
			},
			wantErr: false,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			// Use origin/main for the new test case
			from := ""
			if tt.name == "create new branch from specific commit" {
				from = "origin/main"
			}

			err := createWorktree(tt.repoName, tt.branch, tt.createBranch, from, tt.openEditor, syncNone)
			if (err != nil) != tt.wantErr {
				t.Errorf("createWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("createWorktree() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

// TestOpenInEditor tests the openInEditor function
func TestOpenInEditor(t *testing.T) {
	tests := []struct {
		name        string
		editor      string
		path        string
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful editor open with mock",
			editor: "code",
			path:   "/path/to/worktree",
			setupMock: func() {
				mockOpenInEditor = func(editor, path string) error {
					if editor == "code" && path == "/path/to/worktree" {
						return nil
					}
					return errors.New("unexpected parameters")
				}
			},
			wantErr: false,
		},
		{
			name:   "editor not found",
			editor: "nonexistent-editor",
			path:   "/path/to/worktree",
			setupMock: func() {
				mockOpenInEditor = func(editor, path string) error {
					return errors.New("editor command 'nonexistent-editor' not found in PATH")
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:   "editor start failed",
			editor: "vim",
			path:   "/invalid/path",
			setupMock: func() {
				mockOpenInEditor = func(editor, path string) error {
					return errors.New("failed to start editor 'vim': permission denied")
				}
			},
			wantErr:     true,
			errContains: "failed to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks()
			defer resetMocks()
			tt.setupMock()

			err := openInEditor(tt.editor, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("openInEditor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("openInEditor() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
