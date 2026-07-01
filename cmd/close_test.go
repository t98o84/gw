package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestCloseCmd_ForceFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("Expected 'force' flag to be defined")
	}
}

func TestCloseCmd_ForceFlagShorthand(t *testing.T) {
	flag := closeCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("Expected 'force' flag to be defined")
	}
	if flag.Shorthand != "f" {
		t.Errorf("Expected shorthand 'f', got '%s'", flag.Shorthand)
	}
}

func TestCloseCmd_BranchFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("Expected 'branch' flag to be defined")
	}
}

func TestCloseCmd_BranchFlagShorthand(t *testing.T) {
	flag := closeCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("Expected 'branch' flag to be defined")
	}
	if flag.Shorthand != "b" {
		t.Errorf("Expected shorthand 'b', got '%s'", flag.Shorthand)
	}
}

func TestCloseCmd_NoYesFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("no-yes")
	if flag == nil {
		t.Fatal("Expected 'no-yes' flag to be defined")
	}
}

func TestCloseCmd_NoForceFlag(t *testing.T) {
	flag := closeCmd.Flags().Lookup("no-force")
	if flag == nil {
		t.Fatal("Expected 'no-force' flag to be defined")
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

// TestRunClose_PrintPathProtocol drives runClose against a real git worktree and
// verifies the --print-path protocol the shell wrapper relies on: stdout carries
// the main worktree path, and stderr carries the worktree path (line 1), the
// force flag (line 2) and the branch flag (line 3). This exercises the core
// flag-forwarding behaviour of issue #36 end to end.
func TestRunClose_PrintPathProtocol(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Isolate config so a user's ~/.config/gw/config.yaml can't flip close.force.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// EvalSymlinks so paths match what git reports (e.g. /var -> /private/var on macOS).
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "init", "-q", "-b", "main")
	gitCmd(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	wt := filepath.Join(root, "wt")
	gitCmd(t, repo, "worktree", "add", "-q", wt, "-b", "feature")

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWd)
		resetCloseFlags(t)
	})
	if err := os.Chdir(wt); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		yes        bool
		branch     bool
		wantForce  string
		wantBranch string
	}{
		{"no flags", false, false, "", ""},
		{"force only", true, false, "-y", ""},
		{"branch only", false, true, "", "-b"},
		{"force and branch", true, true, "-y", "-b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCloseFlags(t)
			mustSetFlag(t, "print-path", "true")
			if tt.yes {
				mustSetFlag(t, "yes", "true")
			}
			if tt.branch {
				mustSetFlag(t, "branch", "true")
			}

			stdout, stderr := captureStdoutStderr(t, func() {
				if err := runClose(closeCmd, nil); err != nil {
					t.Fatalf("runClose returned error: %v", err)
				}
			})

			if gotMain := strings.TrimSpace(stdout); gotMain != repo {
				t.Errorf("stdout main path = %q, want %q", gotMain, repo)
			}

			lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
			if lineAt(lines, 0) != wt {
				t.Errorf("stderr line 1 (worktree path) = %q, want %q", lineAt(lines, 0), wt)
			}
			if got := lineAt(lines, 1); got != tt.wantForce {
				t.Errorf("stderr line 2 (force flag) = %q, want %q", got, tt.wantForce)
			}
			if got := lineAt(lines, 2); got != tt.wantBranch {
				t.Errorf("stderr line 3 (branch flag) = %q, want %q", got, tt.wantBranch)
			}
		})
	}
}

// lineAt returns lines[i] or "" when the index is out of range (trailing empty
// protocol lines are dropped by TrimRight).
func lineAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

func mustSetFlag(t *testing.T, name, value string) {
	t.Helper()
	if err := closeCmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set flag %s=%s: %v", name, value, err)
	}
}

func resetCloseFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"print-path", "yes", "force", "branch", "no-yes", "no-force"} {
		if err := closeCmd.Flags().Set(name, "false"); err != nil {
			t.Fatalf("reset flag %s: %v", name, err)
		}
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// captureStdoutStderr swaps os.Stdout/os.Stderr for pipes while fn runs and
// returns what was written. Output is expected to be small (well under the pipe
// buffer), so reading after fn returns cannot deadlock.
func captureStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = origOut, origErr }()

	fn()

	_ = outW.Close()
	_ = errW.Close()
	outBytes, _ := io.ReadAll(outR)
	errBytes, _ := io.ReadAll(errR)
	return string(outBytes), string(errBytes)
}
