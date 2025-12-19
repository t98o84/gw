package errors

import (
	"errors"
	"os/exec"
	"testing"
)

func TestBranchNotFoundError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewBranchNotFoundError("feature/test", nil)
		expected := "branch feature/test not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("git command failed")
		err := NewBranchNotFoundError("feature/test", wrappedErr)
		expected := "branch feature/test not found: git command failed"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewBranchNotFoundError("main", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("unwrap returns nil when no wrapped error", func(t *testing.T) {
		err := NewBranchNotFoundError("main", nil)
		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil when no wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewBranchNotFoundError("main", nil)
		if !err.Is(&BranchNotFoundError{}) {
			t.Error("Is() should return true for BranchNotFoundError type")
		}
		if err.Is(&WorktreeExistsError{}) {
			t.Error("Is() should return false for different error type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewBranchNotFoundError("main", nil)
		if !IsBranchNotFoundError(err) {
			t.Error("IsBranchNotFoundError() should return true")
		}
	})
}

func TestWorktreeExistsError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewWorktreeExistsError("/path/to/worktree", "feature/test", nil)
		expected := "worktree already exists at /path/to/worktree for branch feature/test"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("directory conflict")
		err := NewWorktreeExistsError("/path/to/worktree", "feature/test", wrappedErr)
		expected := "worktree already exists at /path/to/worktree for branch feature/test: directory conflict"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewWorktreeExistsError("/path", "main", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewWorktreeExistsError("/path", "main", nil)
		if !err.Is(&WorktreeExistsError{}) {
			t.Error("Is() should return true for WorktreeExistsError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewWorktreeExistsError("/path", "main", nil)
		if !IsWorktreeExistsError(err) {
			t.Error("IsWorktreeExistsError() should return true")
		}
	})
}

func TestGitHubAPIError(t *testing.T) {
	t.Run("error message with status code", func(t *testing.T) {
		wrappedErr := errors.New("not found")
		err := NewGitHubAPIError("GetPR", 404, wrappedErr)
		expected := "GitHub API error during GetPR (status 404): not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message without status code", func(t *testing.T) {
		wrappedErr := errors.New("network error")
		err := NewGitHubAPIError("GetPR", 0, wrappedErr)
		expected := "GitHub API error during GetPR: network error"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewGitHubAPIError("GetPR", 500, wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewGitHubAPIError("GetPR", 404, nil)
		if !err.Is(&GitHubAPIError{}) {
			t.Error("Is() should return true for GitHubAPIError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewGitHubAPIError("GetPR", 404, nil)
		if !IsGitHubAPIError(err) {
			t.Error("IsGitHubAPIError() should return true")
		}
	})
}

func TestCommandExecutionError(t *testing.T) {
	t.Run("error message with args", func(t *testing.T) {
		wrappedErr := errors.New("exit status 1")
		err := NewCommandExecutionError("git", []string{"branch", "-a"}, nil, wrappedErr)
		expected := "command execution failed: git [branch -a]: exit status 1"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message without args", func(t *testing.T) {
		wrappedErr := errors.New("command not found")
		err := NewCommandExecutionError("fzf", []string{}, nil, wrappedErr)
		expected := "command execution failed: fzf: command not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewCommandExecutionError("git", []string{"status"}, nil, wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewCommandExecutionError("git", []string{}, nil, nil)
		if !err.Is(&CommandExecutionError{}) {
			t.Error("Is() should return true for CommandExecutionError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewCommandExecutionError("git", []string{}, nil, nil)
		if !IsCommandExecutionError(err) {
			t.Error("IsCommandExecutionError() should return true")
		}
	})

	t.Run("error message with output", func(t *testing.T) {
		// Create a mock ExitError
		exitErr := &exec.ExitError{}
		output := []byte("fatal: not a git repository")
		err := NewCommandExecutionError("git", []string{"status"}, output, exitErr)

		// Check that output is captured
		if err.Output != "fatal: not a git repository" {
			t.Errorf("Expected output to be captured, got %q", err.Output)
		}

		// Check that error message includes output
		errorMsg := err.Error()
		expectedMsg := "command execution failed: git [status]\nfatal: not a git repository"
		if errorMsg != expectedMsg {
			t.Errorf("Expected %q, got %q", expectedMsg, errorMsg)
		}
	})

	t.Run("error message with output trimmed", func(t *testing.T) {
		// Create a mock ExitError with output that has leading/trailing whitespace
		exitErr := &exec.ExitError{}
		output := []byte("\n  error message  \n")
		err := NewCommandExecutionError("git", []string{"status"}, output, exitErr)

		// Check that error message has trimmed output
		errorMsg := err.Error()
		expectedMsg := "command execution failed: git [status]\nerror message"
		if errorMsg != expectedMsg {
			t.Errorf("Expected %q, got %q", expectedMsg, errorMsg)
		}
	})

	t.Run("error message without output", func(t *testing.T) {
		// Regular error (not ExitError)
		regularErr := errors.New("some error")
		err := NewCommandExecutionError("git", []string{"status"}, nil, regularErr)

		// Check that output is empty
		if err.Output != "" {
			t.Errorf("Expected output to be empty, got %q", err.Output)
		}

		// Check that error message includes the error
		errorMsg := err.Error()
		expectedMsg := "command execution failed: git [status]: some error"
		if errorMsg != expectedMsg {
			t.Errorf("Expected %q, got %q", expectedMsg, errorMsg)
		}
	})
}

func TestInvalidInputError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewInvalidInputError("abc", "must be a number", nil)
		expected := `invalid input "abc": must be a number`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("parse error")
		err := NewInvalidInputError("123abc", "must be a number", wrappedErr)
		expected := `invalid input "123abc": must be a number: parse error`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewInvalidInputError("test", "reason", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("unwrap returns nil when no wrapped error", func(t *testing.T) {
		err := NewInvalidInputError("test", "reason", nil)
		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil when no wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewInvalidInputError("test", "reason", nil)
		if !err.Is(&InvalidInputError{}) {
			t.Error("Is() should return true for InvalidInputError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewInvalidInputError("test", "reason", nil)
		if !IsInvalidInputError(err) {
			t.Error("IsInvalidInputError() should return true")
		}
	})
}

func TestErrorTypeDiscrimination(t *testing.T) {
	t.Run("different error types are distinguishable", func(t *testing.T) {
		branchErr := NewBranchNotFoundError("main", nil)
		worktreeErr := NewWorktreeExistsError("/path", "main", nil)
		githubErr := NewGitHubAPIError("GetPR", 404, nil)
		commandErr := NewCommandExecutionError("git", []string{}, nil, nil)
		inputErr := NewInvalidInputError("test", "reason", nil)

		if IsBranchNotFoundError(worktreeErr) {
			t.Error("worktreeErr should not be a BranchNotFoundError")
		}
		if IsWorktreeExistsError(branchErr) {
			t.Error("branchErr should not be a WorktreeExistsError")
		}
		if IsGitHubAPIError(commandErr) {
			t.Error("commandErr should not be a GitHubAPIError")
		}
		if IsCommandExecutionError(githubErr) {
			t.Error("githubErr should not be a CommandExecutionError")
		}
		if IsInvalidInputError(branchErr) {
			t.Error("branchErr should not be an InvalidInputError")
		}
		if !IsInvalidInputError(inputErr) {
			t.Error("inputErr should be an InvalidInputError")
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("errors.Is() works with wrapped errors", func(t *testing.T) {
		originalErr := NewBranchNotFoundError("main", nil)
		wrappedOnce := errors.New("wrapped: " + originalErr.Error())

		// Direct check
		if !IsBranchNotFoundError(originalErr) {
			t.Error("Should identify original error")
		}

		// Wrapped error won't be identified by type (expected behavior)
		if IsBranchNotFoundError(wrappedOnce) {
			t.Error("Should not identify wrapped string error")
		}
	})

	t.Run("custom errors preserve wrapped error chain", func(t *testing.T) {
		baseErr := errors.New("base error")
		customErr := NewBranchNotFoundError("main", baseErr)

		if !errors.Is(customErr, baseErr) {
			t.Error("errors.Is() should find wrapped base error")
		}
	})
}

func TestErrorFieldAccess(t *testing.T) {
	t.Run("BranchNotFoundError fields are accessible", func(t *testing.T) {
		err := NewBranchNotFoundError("feature/test", nil)
		if err.Branch != "feature/test" {
			t.Errorf("Expected branch to be 'feature/test', got %q", err.Branch)
		}
	})

	t.Run("WorktreeExistsError fields are accessible", func(t *testing.T) {
		err := NewWorktreeExistsError("/path/to/worktree", "main", nil)
		if err.Path != "/path/to/worktree" {
			t.Errorf("Expected path to be '/path/to/worktree', got %q", err.Path)
		}
		if err.Branch != "main" {
			t.Errorf("Expected branch to be 'main', got %q", err.Branch)
		}
	})

	t.Run("GitHubAPIError fields are accessible", func(t *testing.T) {
		err := NewGitHubAPIError("GetPR", 404, nil)
		if err.Operation != "GetPR" {
			t.Errorf("Expected operation to be 'GetPR', got %q", err.Operation)
		}
		if err.Status != 404 {
			t.Errorf("Expected status to be 404, got %d", err.Status)
		}
	})

	t.Run("CommandExecutionError fields are accessible", func(t *testing.T) {
		err := NewCommandExecutionError("git", []string{"branch", "-a"}, nil, nil)
		if err.Command != "git" {
			t.Errorf("Expected command to be 'git', got %q", err.Command)
		}
		if len(err.Args) != 2 || err.Args[0] != "branch" || err.Args[1] != "-a" {
			t.Errorf("Expected args to be [branch -a], got %v", err.Args)
		}
	})

	t.Run("InvalidInputError fields are accessible", func(t *testing.T) {
		err := NewInvalidInputError("test-input", "must be numeric", nil)
		if err.Input != "test-input" {
			t.Errorf("Expected input to be 'test-input', got %q", err.Input)
		}
		if err.Reason != "must be numeric" {
			t.Errorf("Expected reason to be 'must be numeric', got %q", err.Reason)
		}
	})
}

func TestWorktreeNotFoundError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewWorktreeNotFoundError("feature/test", nil)
		expected := "worktree not found: feature/test"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("directory does not exist")
		err := NewWorktreeNotFoundError("feature-test", wrappedErr)
		expected := "worktree not found: feature-test: directory does not exist"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewWorktreeNotFoundError("/path", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewWorktreeNotFoundError("/path", nil)
		if !err.Is(&WorktreeNotFoundError{}) {
			t.Error("Is() should return true for WorktreeNotFoundError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewWorktreeNotFoundError("/path", nil)
		if !IsWorktreeNotFoundError(err) {
			t.Error("IsWorktreeNotFoundError() should return true")
		}
	})
}

func TestNotAGitRepoError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewNotAGitRepoError("/path/to/dir", nil)
		expected := "not a git repository: /path/to/dir"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New(".git directory not found")
		err := NewNotAGitRepoError("/path/to/dir", wrappedErr)
		expected := "not a git repository: /path/to/dir: .git directory not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewNotAGitRepoError("/path", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewNotAGitRepoError("/path", nil)
		if !err.Is(&NotAGitRepoError{}) {
			t.Error("Is() should return true for NotAGitRepoError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewNotAGitRepoError("/path", nil)
		if !IsNotAGitRepoError(err) {
			t.Error("IsNotAGitRepoError() should return true")
		}
	})
}

func TestFzfNotInstalledError(t *testing.T) {
	t.Run("error message without wrapped error", func(t *testing.T) {
		err := NewFzfNotInstalledError(nil)
		expected := "fzf is not installed"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error message with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("command not found")
		err := NewFzfNotInstalledError(wrappedErr)
		expected := "fzf is not installed: command not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped")
		err := NewFzfNotInstalledError(wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("Unwrap() should return the wrapped error")
		}
	})

	t.Run("Is() method works correctly", func(t *testing.T) {
		err := NewFzfNotInstalledError(nil)
		if !err.Is(&FzfNotInstalledError{}) {
			t.Error("Is() should return true for FzfNotInstalledError type")
		}
	})

	t.Run("errors.Is() works with helper function", func(t *testing.T) {
		err := NewFzfNotInstalledError(nil)
		if !IsFzfNotInstalledError(err) {
			t.Error("IsFzfNotInstalledError() should return true")
		}
	})
}
