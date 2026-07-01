package cmd

import (
	"testing"
)

func TestInitCmd(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	if initCmd.Use != "init <shell>" {
		t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init <shell>")
	}
}

func TestInitCmd_ValidArgs(t *testing.T) {
	validArgs := initCmd.ValidArgs
	expected := []string{"bash", "zsh", "fish"}

	if len(validArgs) != len(expected) {
		t.Errorf("Expected %d valid args, got %d", len(expected), len(validArgs))
	}

	for i, arg := range expected {
		if i < len(validArgs) && validArgs[i] != arg {
			t.Errorf("ValidArgs[%d] = %q, want %q", i, validArgs[i], arg)
		}
	}
}

func TestBashZshInit_Content(t *testing.T) {
	// Verify the shell script contains expected elements
	if bashZshInit == "" {
		t.Error("bashZshInit should not be empty")
	}

	expectedElements := []string{
		"gw()",
		"gw sw",
		"--print-path",
		"cd",
	}

	for _, elem := range expectedElements {
		if !containsString(bashZshInit, elem) {
			t.Errorf("bashZshInit should contain %q", elem)
		}
	}
}

func TestFishInit_Content(t *testing.T) {
	// Verify the shell script contains expected elements
	if fishInit == "" {
		t.Error("fishInit should not be empty")
	}

	expectedElements := []string{
		"function gw",
		"gw sw",
		"--print-path",
		"cd",
	}

	for _, elem := range expectedElements {
		if !containsString(fishInit, elem) {
			t.Errorf("fishInit should contain %q", elem)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestBashZshInit_ClosePathAndYesSeparation guards against regressions where the
// bash/zsh wrapper captured the whole stderr (worktree path + "-y") into a single
// variable and passed "<path>\n-y" as one argument to `gw rm` (issue #34).
func TestBashZshInit_ClosePathAndYesSeparation(t *testing.T) {
	requiredElements := []string{
		"sed -n '1p'",             // extract worktree path (line 1)
		"sed -n '2p'",             // extract -y flag (line 2)
		"command gw rm -y",        // forward -y when close.force is enabled
		`for arg in "${@:2}"`,     // inspect CLI args (e.g. `gw close -y`)
		"-y|--yes|--force",        // whitelist the force flags close understands
		"--no-yes|--no-force",     // ...and their negations
		"--print-path $close_arg", // forward only whitelisted flags to close
	}

	for _, elem := range requiredElements {
		if !containsString(bashZshInit, elem) {
			t.Errorf("bashZshInit should contain %q to separate the worktree path from the -y flag", elem)
		}
	}
}

// TestFishInit_ClosePathAndYesSeparation guards against the fish wrapper joining
// the stderr list with spaces (echo "$list" | sed) instead of reading the path and
// -y flag as separate list elements (issue #34).
func TestFishInit_ClosePathAndYesSeparation(t *testing.T) {
	requiredElements := []string{
		"$stderr_output[1]",        // worktree path
		"$stderr_output[2]",        // -y flag
		"command gw rm -y",         // forward -y when close.force is enabled
		"for a in $argv[2..]",      // inspect CLI args
		"case -y --yes --force",    // whitelist the force flags close understands
		"case --no-yes --no-force", // ...and their negations
		"--print-path $close_arg",  // forward only whitelisted flags to close
	}

	for _, elem := range requiredElements {
		if !containsString(fishInit, elem) {
			t.Errorf("fishInit should contain %q to read the worktree path and -y flag separately", elem)
		}
	}
}

func TestRunInit_UnsupportedShell(t *testing.T) {
	err := runInit(initCmd, []string{"powershell"})
	if err == nil {
		t.Error("Expected error for unsupported shell")
	}
}

func TestRunInit_Bash(t *testing.T) {
	// This will print to stdout, but should not error
	err := runInit(initCmd, []string{"bash"})
	if err != nil {
		t.Errorf("Unexpected error for bash: %v", err)
	}
}

func TestRunInit_Zsh(t *testing.T) {
	err := runInit(initCmd, []string{"zsh"})
	if err != nil {
		t.Errorf("Unexpected error for zsh: %v", err)
	}
}

func TestRunInit_Fish(t *testing.T) {
	err := runInit(initCmd, []string{"fish"})
	if err != nil {
		t.Errorf("Unexpected error for fish: %v", err)
	}
}
