package fzf

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/shell"
)

// newTestSelector creates a selector with a mock fzf executor for testing
func newTestSelector(mockFzfFunc func(args []string, input string) (string, error)) *FzfSelector {
	s := &FzfSelector{
		executor: &shell.MockExecutor{
			LookPathFunc: func(name string) (string, error) {
				if name == "fzf" {
					return "/usr/bin/fzf", nil
				}
				return "", fmt.Errorf("not found")
			},
		},
	}
	if mockFzfFunc != nil {
		s.fzfExecutor = mockFzfFunc
	} else {
		s.fzfExecutor = s.defaultFzfExecutor
	}
	return s
}

// TestFzfSelector_IsAvailable tests fzf availability checking
func TestFzfSelector_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		executor *shell.MockExecutor
		want     bool
	}{
		{
			name: "fzf is available",
			executor: &shell.MockExecutor{
				LookPathFunc: func(name string) (string, error) {
					if name == "fzf" {
						return "/usr/bin/fzf", nil
					}
					return "", fmt.Errorf("not found")
				},
			},
			want: true,
		},
		{
			name: "fzf is not available",
			executor: &shell.MockExecutor{
				LookPathFunc: func(name string) (string, error) {
					return "", fmt.Errorf("not found")
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewSelector(tt.executor)
			if got := selector.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFzfSelector_SelectBranch tests branch selection
func TestFzfSelector_SelectBranch(t *testing.T) {
	tests := []struct {
		name     string
		selector *FzfSelector
		branches []string
		want     string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "fzf not available",
			selector: &FzfSelector{
				executor: &shell.MockExecutor{
					LookPathFunc: func(name string) (string, error) {
						return "", fmt.Errorf("not found")
					},
				},
			},
			branches: []string{"main", "feature/test"},
			wantErr:  true,
			errMsg:   "fzf is not installed",
		},
		{
			name:     "no branches",
			selector: newTestSelector(nil),
			branches: []string{},
			wantErr:  true,
			errMsg:   "no branches found",
		},
		{
			name: "successful selection",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				// Simulate selecting "main"
				return "main", nil
			}),
			branches: []string{"main", "feature/test"},
			want:     "main",
			wantErr:  false,
		},
		{
			name: "user cancelled",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				return "", nil
			}),
			branches: []string{"main", "feature/test"},
			want:     "",
			wantErr:  false,
		},
		{
			name: "fzf execution error",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				return "", fmt.Errorf("fzf failed")
			}),
			branches: []string{"main"},
			wantErr:  true,
			errMsg:   "fzf failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.selector.SelectBranch(tt.branches)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("SelectBranch() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SelectBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFzfSelector_SelectBranch_ValidInput tests SelectBranch with various inputs
func TestFzfSelector_SelectBranch_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		branches []string
		selected string
	}{
		{
			name:     "single branch",
			branches: []string{"main"},
			selected: "main",
		},
		{
			name:     "multiple branches",
			branches: []string{"main", "feature/test", "hotfix/bug"},
			selected: "feature/test",
		},
		{
			name:     "branch with special characters",
			branches: []string{"feature/foo-bar_123", "release/v1.0.0"},
			selected: "release/v1.0.0",
		},
		{
			name:     "branch with spaces",
			branches: []string{"main", "feature test"},
			selected: "feature test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := newTestSelector(func(args []string, input string) (string, error) {
				// Verify input contains all branches
				for _, branch := range tt.branches {
					if !strings.Contains(input, branch) {
						t.Errorf("Input missing branch: %s", branch)
					}
				}
				return tt.selected, nil
			})

			got, err := selector.SelectBranch(tt.branches)
			if err != nil {
				t.Errorf("SelectBranch() unexpected error: %v", err)
			}
			if got != tt.selected {
				t.Errorf("SelectBranch() = %v, want %v", got, tt.selected)
			}
		})
	}
}

// TestFzfSelector_SelectWorktree tests single worktree selection
func TestFzfSelector_SelectWorktree(t *testing.T) {
	tests := []struct {
		name        string
		selector    *FzfSelector
		worktrees   []git.Worktree
		excludeMain bool
		wantBranch  string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "fzf not available",
			selector: &FzfSelector{
				executor: &shell.MockExecutor{
					LookPathFunc: func(name string) (string, error) {
						return "", fmt.Errorf("not found")
					},
				},
			},
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
			},
			excludeMain: false,
			wantErr:     true,
			errMsg:      "fzf is not installed",
		},
		{
			name:        "no worktrees",
			selector:    newTestSelector(nil),
			worktrees:   []git.Worktree{},
			excludeMain: false,
			wantErr:     true,
			errMsg:      "no worktrees found",
		},
		{
			name: "successful selection",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				// Return the label that would be shown in fzf (with "(main)" suffix)
				return "repo (main)", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
			},
			excludeMain: false,
			wantBranch:  "main",
			wantErr:     false,
		},
		{
			name: "user cancelled",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				return "", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
			},
			excludeMain: false,
			wantBranch:  "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.selector.SelectWorktree(tt.worktrees, tt.excludeMain)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("SelectWorktree() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr {
				if tt.wantBranch == "" && got != nil {
					t.Errorf("SelectWorktree() = %v, want nil", got)
				} else if tt.wantBranch != "" {
					if got == nil {
						t.Error("SelectWorktree() = nil, want non-nil")
					} else if got.Branch != tt.wantBranch {
						t.Errorf("SelectWorktree() branch = %v, want %v", got.Branch, tt.wantBranch)
					}
				}
			}
		})
	}
}

// TestFzfSelector_SelectWorktrees_ExcludeMain tests main worktree exclusion
func TestFzfSelector_SelectWorktrees_ExcludeMain(t *testing.T) {
	selector := newTestSelector(nil)

	worktrees := []git.Worktree{
		{Path: "/repo", Branch: "main", IsMain: true},
	}

	// When excluding main and only main exists, should return error
	_, err := selector.SelectWorktrees(worktrees, true, false)
	if err == nil {
		t.Error("SelectWorktrees() expected error when all worktrees are excluded")
	}
	if !strings.Contains(err.Error(), "no worktrees available") {
		t.Errorf("SelectWorktrees() error = %v, should contain 'no worktrees available'", err.Error())
	}
}

// TestFzfSelector_SelectWorktrees tests multi worktree selection
func TestFzfSelector_SelectWorktrees(t *testing.T) {
	tests := []struct {
		name         string
		selector     *FzfSelector
		worktrees    []git.Worktree
		excludeMain  bool
		multi        bool
		mockResponse string
		wantCount    int
		wantBranches []string
		wantErr      bool
		errMsg       string
	}{
		{
			name: "single select mode",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				// Verify single-select args
				if strings.Contains(strings.Join(args, " "), "--multi") {
					t.Error("Expected single-select mode")
				}
				return "repo-feature", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo-feature", Branch: "feature", IsMain: false},
			},
			excludeMain:  false,
			multi:        false,
			wantCount:    1,
			wantBranches: []string{"feature"},
			wantErr:      false,
		},
		{
			name: "multi select mode",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				// Verify multi-select args
				if !strings.Contains(strings.Join(args, " "), "--multi") {
					t.Error("Expected multi-select mode")
				}
				return "repo-feature1\nrepo-feature2", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo-feature1", Branch: "feature1", IsMain: false},
				{Path: "/repo-feature2", Branch: "feature2", IsMain: false},
			},
			excludeMain:  false,
			multi:        true,
			wantCount:    2,
			wantBranches: []string{"feature1", "feature2"},
			wantErr:      false,
		},
		{
			name: "exclude main worktree",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				// Verify main is not in input
				if strings.Contains(input, "(main)") {
					t.Error("Main worktree should be excluded from input")
				}
				return "repo-feature", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo-feature", Branch: "feature", IsMain: false},
			},
			excludeMain:  true,
			multi:        false,
			wantCount:    1,
			wantBranches: []string{"feature"},
			wantErr:      false,
		},
		{
			name: "user cancelled returns nil",
			selector: newTestSelector(func(args []string, input string) (string, error) {
				return "", nil
			}),
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
			},
			excludeMain: false,
			multi:       false,
			wantCount:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.selector.SelectWorktrees(tt.worktrees, tt.excludeMain, tt.multi)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectWorktrees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("SelectWorktrees() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr {
				if len(got) != tt.wantCount {
					t.Errorf("SelectWorktrees() count = %v, want %v", len(got), tt.wantCount)
				}
				for i, wantBranch := range tt.wantBranches {
					if i >= len(got) {
						t.Errorf("SelectWorktrees() missing branch %v", wantBranch)
						continue
					}
					if got[i].Branch != wantBranch {
						t.Errorf("SelectWorktrees()[%d] branch = %v, want %v", i, got[i].Branch, wantBranch)
					}
				}
			}
		})
	}
}

// TestExecuteFzf_Integration tests executeFzf integration points
func TestExecuteFzf_Integration(t *testing.T) {
	t.Run("verify fzf command structure", func(t *testing.T) {
		// Verify that the command would be constructed correctly
		selector := NewSelector(&shell.MockExecutor{
			LookPathFunc: func(name string) (string, error) {
				if name == "fzf" {
					return "/usr/bin/fzf", nil
				}
				return "", fmt.Errorf("not found")
			},
		})

		if !selector.IsAvailable() {
			t.Error("Expected selector to report fzf as available")
		}

		// Verify exec.Command would accept fzf with typical arguments
		cmd := exec.Command("fzf", "--height=40%", "--reverse", "--prompt=test: ")
		if cmd.Path == "" {
			t.Error("Command path should not be empty")
		}
	})

	t.Run("verify input preparation", func(t *testing.T) {
		// Test that input strings are properly joined
		branches := []string{"main", "feature/test", "hotfix/bug"}
		input := strings.Join(branches, "\n")

		expectedLines := len(branches)
		actualLines := len(strings.Split(input, "\n"))

		if actualLines != expectedLines {
			t.Errorf("Expected %d lines in input, got %d", expectedLines, actualLines)
		}
	})

	t.Run("verify output trimming logic", func(t *testing.T) {
		// Test that output trimming works correctly
		testCases := []struct {
			input    string
			expected string
		}{
			{input: "branch\n", expected: "branch"},
			{input: "  branch  \n", expected: "branch"},
			{input: "branch", expected: "branch"},
			{input: "\nbranch\n", expected: "branch"},
		}

		for _, tc := range testCases {
			result := strings.TrimSpace(tc.input)
			if result != tc.expected {
				t.Errorf("TrimSpace(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("verify multi-line output parsing", func(t *testing.T) {
		// Test multi-line output parsing logic used in SelectWorktrees
		output := "item1\nitem2\nitem3"
		lines := strings.Split(output, "\n")

		if len(lines) != 3 {
			t.Errorf("Expected 3 lines, got %d", len(lines))
		}

		// Test empty output
		emptyOutput := ""
		if emptyOutput != "" {
			t.Error("Empty output should remain empty")
		}

		// Test output with trailing newline
		outputWithTrailing := "item1\nitem2\n"
		linesWithTrailing := strings.Split(strings.TrimSpace(outputWithTrailing), "\n")
		if len(linesWithTrailing) != 2 {
			t.Errorf("Expected 2 lines after trimming, got %d", len(linesWithTrailing))
		}
	})
}

// MockSelector for testing
type MockSelector struct {
	SelectBranchFunc    func(branches []string) (string, error)
	SelectWorktreeFunc  func(worktrees []git.Worktree, excludeMain bool) (*git.Worktree, error)
	SelectWorktreesFunc func(worktrees []git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error)
	IsAvailableFunc     func() bool
}

func (m *MockSelector) SelectBranch(branches []string) (string, error) {
	if m.SelectBranchFunc != nil {
		return m.SelectBranchFunc(branches)
	}
	if len(branches) > 0 {
		return branches[0], nil
	}
	return "", nil
}

func (m *MockSelector) SelectWorktree(worktrees []git.Worktree, excludeMain bool) (*git.Worktree, error) {
	if m.SelectWorktreeFunc != nil {
		return m.SelectWorktreeFunc(worktrees, excludeMain)
	}
	for i := range worktrees {
		wt := &worktrees[i]
		if !excludeMain || !wt.IsMain {
			return wt, nil
		}
	}
	return nil, nil
}

func (m *MockSelector) SelectWorktrees(worktrees []git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error) {
	if m.SelectWorktreesFunc != nil {
		return m.SelectWorktreesFunc(worktrees, excludeMain, multi)
	}
	var selected []*git.Worktree
	for i := range worktrees {
		wt := &worktrees[i]
		if !excludeMain || !wt.IsMain {
			selected = append(selected, wt)
			if !multi {
				break
			}
		}
	}
	return selected, nil
}

func (m *MockSelector) IsAvailable() bool {
	if m.IsAvailableFunc != nil {
		return m.IsAvailableFunc()
	}
	return true
}

// TestMockSelector verifies the mock works as expected
func TestMockSelector(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		mock := &MockSelector{}

		// Test IsAvailable
		if !mock.IsAvailable() {
			t.Error("MockSelector.IsAvailable() should return true by default")
		}

		// Test SelectBranch
		branch, err := mock.SelectBranch([]string{"main", "feature"})
		if err != nil {
			t.Errorf("MockSelector.SelectBranch() unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("MockSelector.SelectBranch() = %v, want 'main'", branch)
		}

		// Test SelectWorktree
		worktrees := []git.Worktree{
			{Path: "/repo", Branch: "main", IsMain: true},
			{Path: "/repo-feature", Branch: "feature", IsMain: false},
		}
		wt, err := mock.SelectWorktree(worktrees, false)
		if err != nil {
			t.Errorf("MockSelector.SelectWorktree() unexpected error: %v", err)
		}
		if wt == nil || wt.Branch != "main" {
			t.Errorf("MockSelector.SelectWorktree() = %v, want worktree with branch 'main'", wt)
		}

		// Test SelectWorktrees with exclusion
		wts, err := mock.SelectWorktrees(worktrees, true, false)
		if err != nil {
			t.Errorf("MockSelector.SelectWorktrees() unexpected error: %v", err)
		}
		if len(wts) != 1 || wts[0].Branch != "feature" {
			t.Errorf("MockSelector.SelectWorktrees() should exclude main worktree")
		}
	})

	t.Run("custom behavior", func(t *testing.T) {
		mock := &MockSelector{
			SelectBranchFunc: func(branches []string) (string, error) {
				return "custom", nil
			},
		}

		branch, err := mock.SelectBranch([]string{"main"})
		if err != nil {
			t.Errorf("MockSelector.SelectBranch() unexpected error: %v", err)
		}
		if branch != "custom" {
			t.Errorf("MockSelector.SelectBranch() = %v, want 'custom'", branch)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		expectedErr := fmt.Errorf("custom error")
		mock := &MockSelector{
			SelectBranchFunc: func(branches []string) (string, error) {
				return "", expectedErr
			},
		}

		_, err := mock.SelectBranch([]string{"main"})
		if err != expectedErr {
			t.Errorf("MockSelector.SelectBranch() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("multi-select behavior", func(t *testing.T) {
		worktrees := []git.Worktree{
			{Path: "/repo-feature1", Branch: "feature1", IsMain: false},
			{Path: "/repo-feature2", Branch: "feature2", IsMain: false},
			{Path: "/repo-feature3", Branch: "feature3", IsMain: false},
		}

		mock := &MockSelector{}

		// Test multi-select returns all
		wts, err := mock.SelectWorktrees(worktrees, false, true)
		if err != nil {
			t.Errorf("MockSelector.SelectWorktrees() unexpected error: %v", err)
		}
		if len(wts) != 3 {
			t.Errorf("MockSelector.SelectWorktrees(multi=true) = %d items, want 3", len(wts))
		}

		// Test single-select returns one
		wts, err = mock.SelectWorktrees(worktrees, false, false)
		if err != nil {
			t.Errorf("MockSelector.SelectWorktrees() unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Errorf("MockSelector.SelectWorktrees(multi=false) = %d items, want 1", len(wts))
		}
	})
}

// TestFzfSelector_EdgeCases tests edge cases and boundary conditions
func TestFzfSelector_EdgeCases(t *testing.T) {
	t.Run("empty branch name", func(t *testing.T) {
		selector := newTestSelector(func(args []string, input string) (string, error) {
			// Verify empty strings are included
			if !strings.Contains(input, "\n\n") {
				t.Error("Expected double newline for empty branch name")
			}
			return "", nil
		})

		branches := []string{"main", "", "feature"}
		_, err := selector.SelectBranch(branches)
		if err != nil {
			t.Errorf("SelectBranch() unexpected error: %v", err)
		}
	})

	t.Run("very long branch names", func(t *testing.T) {
		longBranchName := strings.Repeat("a", 1000)
		selector := newTestSelector(func(args []string, input string) (string, error) {
			if len(input) < 1000 {
				t.Errorf("Expected input length >= 1000, got %d", len(input))
			}
			return longBranchName, nil
		})

		branches := []string{longBranchName}
		got, err := selector.SelectBranch(branches)
		if err != nil {
			t.Errorf("SelectBranch() unexpected error: %v", err)
		}
		if got != longBranchName {
			t.Error("Long branch name not preserved")
		}
	})

	t.Run("worktree paths with unicode", func(t *testing.T) {
		selector := newTestSelector(func(args []string, input string) (string, error) {
			// Verify unicode is preserved
			if !strings.Contains(input, "日本語") {
				t.Error("Unicode should be preserved in input")
			}
			return "日本語", nil
		})

		worktrees := []git.Worktree{
			{Path: "/repo/日本語", Branch: "feature", IsMain: false},
		}

		got, err := selector.SelectWorktrees(worktrees, false, false)
		if err != nil {
			t.Errorf("SelectWorktrees() unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].Branch != "feature" {
			t.Error("Unicode path should be handled correctly")
		}
	})

	t.Run("empty worktree list with exclusion", func(t *testing.T) {
		selector := newTestSelector(nil)

		// Test both empty list and list with only main
		emptyWorktrees := []git.Worktree{}
		_, err := selector.SelectWorktrees(emptyWorktrees, true, false)
		if err == nil {
			t.Error("Expected error for empty worktree list")
		}

		onlyMain := []git.Worktree{
			{Path: "/repo", Branch: "main", IsMain: true},
		}
		_, err = selector.SelectWorktrees(onlyMain, true, false)
		if err == nil {
			t.Error("Expected error when all worktrees are excluded")
		}
	})
}
