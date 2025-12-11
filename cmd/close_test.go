package cmd

import (
	"testing"

	"github.com/t98o84/gw/internal/git"
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

func TestCloseCmd_YesFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected 'yes' flag to be defined")
	}
}

func TestCloseCmd_YesFlagShorthand(t *testing.T) {
	flag := closeCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected 'yes' flag to be defined")
	}
	if flag.Shorthand != "y" {
		t.Errorf("Expected shorthand 'y', got '%s'", flag.Shorthand)
	}
}

func TestFindCurrentWorktree(t *testing.T) {
	tests := []struct {
		name          string
		currentPath   string
		worktrees     []git.Worktree
		expectedFound bool
		expectedMain  bool
	}{
		{
			name:        "exact match",
			currentPath: "/repo/worktrees/feature",
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo/worktrees/feature", Branch: "feature", IsMain: false},
			},
			expectedFound: true,
			expectedMain:  false,
		},
		{
			name:        "subdirectory match",
			currentPath: "/repo/worktrees/feature/subdir",
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo/worktrees/feature", Branch: "feature", IsMain: false},
			},
			expectedFound: true,
			expectedMain:  false,
		},
		{
			name:        "main worktree",
			currentPath: "/repo",
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo/worktrees/feature", Branch: "feature", IsMain: false},
			},
			expectedFound: true,
			expectedMain:  true,
		},
		{
			name:        "not in worktree",
			currentPath: "/other/path",
			worktrees: []git.Worktree{
				{Path: "/repo", Branch: "main", IsMain: true},
				{Path: "/repo/worktrees/feature", Branch: "feature", IsMain: false},
			},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the actual helper function
			result := findCurrentWorktree(tt.currentPath, tt.worktrees)

			found := result != nil
			if found != tt.expectedFound {
				t.Errorf("Expected found=%v, got %v", tt.expectedFound, found)
			}

			if found && result.IsMain != tt.expectedMain {
				t.Errorf("Expected isMain=%v, got %v", tt.expectedMain, result.IsMain)
			}
		})
	}
}
