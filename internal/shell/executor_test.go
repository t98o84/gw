package shell

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestRealExecutor_Execute_Success(t *testing.T) {
	executor := NewRealExecutor()

	// Use a simple command that works on all platforms
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmd = "echo"
		args = []string{"hello"}
	}

	output, err := executor.Execute(cmd, args...)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestRealExecutor_Execute_CapturesStderr(t *testing.T) {
	executor := NewRealExecutor()

	// Use a command that will fail
	_, err := executor.Execute("nonexistent-command-that-should-fail")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	// Check that we got an ExitError with Stderr populated
	if exitErr, ok := err.(*exec.ExitError); ok {
		// The error should contain stderr output
		if len(exitErr.Stderr) == 0 {
			t.Error("Expected stderr to be captured in ExitError")
		}
	}
}

func TestRealExecutor_ExecuteWithStdio_Success(t *testing.T) {
	executor := NewRealExecutor()

	// Use a simple command that works on all platforms
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "exit", "0"}
	} else {
		cmd = "true"
		args = []string{}
	}

	err := executor.ExecuteWithStdio(cmd, args...)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRealExecutor_LookPath_Success(t *testing.T) {
	executor := NewRealExecutor()

	// Look for a command that should exist on all platforms
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
	} else {
		cmd = "sh"
	}

	path, err := executor.LookPath(cmd)
	if err != nil {
		t.Fatalf("Expected to find %s, got error: %v", cmd, err)
	}

	if path == "" {
		t.Errorf("Expected non-empty path for %s", cmd)
	}
}

func TestRealExecutor_LookPath_NotFound(t *testing.T) {
	executor := NewRealExecutor()

	_, err := executor.LookPath("nonexistent-command-that-should-not-exist")
	if err == nil {
		t.Error("Expected an error when looking for non-existent command")
	}
}
