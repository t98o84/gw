package git

import (
	"testing"
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
