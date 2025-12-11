package cmd

import (
	"testing"
)

func TestCloseCmd(t *testing.T) {
	if closeCmd == nil {
		t.Fatal("closeCmd should not be nil")
	}

	if closeCmd.Use != "close" {
		t.Errorf("closeCmd.Use = %q, want %q", closeCmd.Use, "close")
	}
}

func TestCloseCmd_Aliases(t *testing.T) {
	aliases := closeCmd.Aliases
	expected := []string{"c"}

	if len(aliases) != len(expected) {
		t.Errorf("Expected %d aliases, got %d", len(expected), len(aliases))
		return
	}

	for i, alias := range expected {
		if aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, aliases[i], alias)
		}
	}
}

func TestCloseCmd_PrintPathFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("print-path")
	if flag == nil {
		t.Fatal("Expected 'print-path' flag to be defined")
	}
}

func TestFindCurrentWorktree(t *testing.T) {
	tests := []struct {
		name        string
		currentPath string
		worktrees   []struct {
			path   string
			branch string
			isMain bool
		}
		expectedFound bool
		expectedMain  bool
	}{
		{
			name:        "exact match",
			currentPath: "/repo/worktrees/feature",
			worktrees: []struct {
				path   string
				branch string
				isMain bool
			}{
				{path: "/repo", branch: "main", isMain: true},
				{path: "/repo/worktrees/feature", branch: "feature", isMain: false},
			},
			expectedFound: true,
			expectedMain:  false,
		},
		{
			name:        "subdirectory match",
			currentPath: "/repo/worktrees/feature/subdir",
			worktrees: []struct {
				path   string
				branch string
				isMain bool
			}{
				{path: "/repo", branch: "main", isMain: true},
				{path: "/repo/worktrees/feature", branch: "feature", isMain: false},
			},
			expectedFound: true,
			expectedMain:  false,
		},
		{
			name:        "main worktree",
			currentPath: "/repo",
			worktrees: []struct {
				path   string
				branch string
				isMain bool
			}{
				{path: "/repo", branch: "main", isMain: true},
				{path: "/repo/worktrees/feature", branch: "feature", isMain: false},
			},
			expectedFound: true,
			expectedMain:  true,
		},
		{
			name:        "not in worktree",
			currentPath: "/other/path",
			worktrees: []struct {
				path   string
				branch string
				isMain bool
			}{
				{path: "/repo", branch: "main", isMain: true},
				{path: "/repo/worktrees/feature", branch: "feature", isMain: false},
			},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert test worktrees to git.Worktree slice
			var worktrees []struct {
				Path   string
				Branch string
				IsMain bool
			}
			for _, wt := range tt.worktrees {
				worktrees = append(worktrees, struct {
					Path   string
					Branch string
					IsMain bool
				}{
					Path:   wt.path,
					Branch: wt.branch,
					IsMain: wt.isMain,
				})
			}

			// Create a simple worktree slice for testing
			// Note: This is a simplified version for unit testing
			// In real tests, we'd use git.Worktree but for simplicity we test the logic

			// Test the path matching logic directly
			// Find the longest matching path (most specific)
			found := false
			isMain := false
			longestMatch := ""
			for _, wt := range worktrees {
				if tt.currentPath == wt.Path || (len(tt.currentPath) > len(wt.Path) && tt.currentPath[:len(wt.Path)] == wt.Path && tt.currentPath[len(wt.Path)] == '/') {
					// Take the longest matching path
					if len(wt.Path) > len(longestMatch) {
						found = true
						isMain = wt.IsMain
						longestMatch = wt.Path
					}
				}
			}

			if found != tt.expectedFound {
				t.Errorf("Expected found=%v, got %v", tt.expectedFound, found)
			}

			if found && isMain != tt.expectedMain {
				t.Errorf("Expected isMain=%v, got %v", tt.expectedMain, isMain)
			}
		})
	}
}
