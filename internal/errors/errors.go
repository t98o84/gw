package errors

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// BranchNotFoundError represents an error when a branch cannot be found
type BranchNotFoundError struct {
	Branch string
	Err    error
}

func (e *BranchNotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("branch %s not found: %v", e.Branch, e.Err)
	}
	return fmt.Sprintf("branch %s not found", e.Branch)
}

func (e *BranchNotFoundError) Unwrap() error {
	return e.Err
}

func (e *BranchNotFoundError) Is(target error) bool {
	_, ok := target.(*BranchNotFoundError)
	return ok
}

// NewBranchNotFoundError creates a new BranchNotFoundError
func NewBranchNotFoundError(branch string, err error) *BranchNotFoundError {
	return &BranchNotFoundError{Branch: branch, Err: err}
}

// WorktreeExistsError represents an error when a worktree already exists
type WorktreeExistsError struct {
	Path   string
	Branch string
	Err    error
}

func (e *WorktreeExistsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("worktree already exists at %s for branch %s: %v", e.Path, e.Branch, e.Err)
	}
	return fmt.Sprintf("worktree already exists at %s for branch %s", e.Path, e.Branch)
}

func (e *WorktreeExistsError) Unwrap() error {
	return e.Err
}

func (e *WorktreeExistsError) Is(target error) bool {
	_, ok := target.(*WorktreeExistsError)
	return ok
}

// NewWorktreeExistsError creates a new WorktreeExistsError
func NewWorktreeExistsError(path, branch string, err error) *WorktreeExistsError {
	return &WorktreeExistsError{Path: path, Branch: branch, Err: err}
}

// GitHubAPIError represents an error from the GitHub API
type GitHubAPIError struct {
	Operation string
	Status    int
	Err       error
}

func (e *GitHubAPIError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("GitHub API error during %s (status %d): %v", e.Operation, e.Status, e.Err)
	}
	return fmt.Sprintf("GitHub API error during %s: %v", e.Operation, e.Err)
}

func (e *GitHubAPIError) Unwrap() error {
	return e.Err
}

func (e *GitHubAPIError) Is(target error) bool {
	_, ok := target.(*GitHubAPIError)
	return ok
}

// NewGitHubAPIError creates a new GitHubAPIError
func NewGitHubAPIError(operation string, status int, err error) *GitHubAPIError {
	return &GitHubAPIError{Operation: operation, Status: status, Err: err}
}

// CommandExecutionError represents an error when executing a command
type CommandExecutionError struct {
	Command string
	Args    []string
	Stderr  string
	Err     error
}

func (e *CommandExecutionError) Error() string {
	var sb strings.Builder
	
	if len(e.Args) > 0 {
		sb.WriteString(fmt.Sprintf("command execution failed: %s %v", e.Command, e.Args))
	} else {
		sb.WriteString(fmt.Sprintf("command execution failed: %s", e.Command))
	}
	
	if e.Stderr != "" {
		sb.WriteString(fmt.Sprintf("\n%s", strings.TrimSpace(e.Stderr)))
	} else if e.Err != nil {
		sb.WriteString(fmt.Sprintf(": %v", e.Err))
	}
	
	return sb.String()
}

func (e *CommandExecutionError) Unwrap() error {
	return e.Err
}

func (e *CommandExecutionError) Is(target error) bool {
	_, ok := target.(*CommandExecutionError)
	return ok
}

// NewCommandExecutionError creates a new CommandExecutionError
func NewCommandExecutionError(command string, args []string, err error) *CommandExecutionError {
	stderr := ""
	
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
	}
	
	return &CommandExecutionError{
		Command: command,
		Args:    args,
		Stderr:  stderr,
		Err:     err,
	}
}

// InvalidInputError represents an error when user input is invalid
type InvalidInputError struct {
	Input  string
	Reason string
	Err    error
}

func (e *InvalidInputError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid input %q: %s: %v", e.Input, e.Reason, e.Err)
	}
	return fmt.Sprintf("invalid input %q: %s", e.Input, e.Reason)
}

func (e *InvalidInputError) Unwrap() error {
	return e.Err
}

func (e *InvalidInputError) Is(target error) bool {
	_, ok := target.(*InvalidInputError)
	return ok
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(input, reason string, err error) *InvalidInputError {
	return &InvalidInputError{Input: input, Reason: reason, Err: err}
}

// WorktreeNotFoundError represents an error when a worktree cannot be found
type WorktreeNotFoundError struct {
	Identifier string // Can be a branch name, path suffix, or directory name
	Err        error
}

func (e *WorktreeNotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("worktree not found: %s: %v", e.Identifier, e.Err)
	}
	return fmt.Sprintf("worktree not found: %s", e.Identifier)
}

func (e *WorktreeNotFoundError) Unwrap() error {
	return e.Err
}

func (e *WorktreeNotFoundError) Is(target error) bool {
	_, ok := target.(*WorktreeNotFoundError)
	return ok
}

// NewWorktreeNotFoundError creates a new WorktreeNotFoundError
func NewWorktreeNotFoundError(identifier string, err error) *WorktreeNotFoundError {
	return &WorktreeNotFoundError{Identifier: identifier, Err: err}
}

// NotAGitRepoError represents an error when a directory is not a git repository
type NotAGitRepoError struct {
	Path string
	Err  error
}

func (e *NotAGitRepoError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("not a git repository: %s: %v", e.Path, e.Err)
	}
	return fmt.Sprintf("not a git repository: %s", e.Path)
}

func (e *NotAGitRepoError) Unwrap() error {
	return e.Err
}

func (e *NotAGitRepoError) Is(target error) bool {
	_, ok := target.(*NotAGitRepoError)
	return ok
}

// NewNotAGitRepoError creates a new NotAGitRepoError
func NewNotAGitRepoError(path string, err error) *NotAGitRepoError {
	return &NotAGitRepoError{Path: path, Err: err}
}

// FzfNotInstalledError represents an error when fzf is not installed
type FzfNotInstalledError struct {
	Err error
}

func (e *FzfNotInstalledError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("fzf is not installed: %v", e.Err)
	}
	return "fzf is not installed"
}

func (e *FzfNotInstalledError) Unwrap() error {
	return e.Err
}

func (e *FzfNotInstalledError) Is(target error) bool {
	_, ok := target.(*FzfNotInstalledError)
	return ok
}

// NewFzfNotInstalledError creates a new FzfNotInstalledError
func NewFzfNotInstalledError(err error) *FzfNotInstalledError {
	return &FzfNotInstalledError{Err: err}
}

// NotInWorktreeError represents an error when the current directory is not in a worktree
type NotInWorktreeError struct {
	Path string
	Err  error
}

func (e *NotInWorktreeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("not in a worktree: %s: %v", e.Path, e.Err)
	}
	return fmt.Sprintf("not in a worktree: %s", e.Path)
}

func (e *NotInWorktreeError) Unwrap() error {
	return e.Err
}

func (e *NotInWorktreeError) Is(target error) bool {
	_, ok := target.(*NotInWorktreeError)
	return ok
}

// NewNotInWorktreeError creates a new NotInWorktreeError
func NewNotInWorktreeError(path string, err error) *NotInWorktreeError {
	return &NotInWorktreeError{Path: path, Err: err}
}

// Helper functions to check error types

// IsBranchNotFoundError checks if an error is a BranchNotFoundError
func IsBranchNotFoundError(err error) bool {
	return errors.Is(err, &BranchNotFoundError{})
}

// IsWorktreeExistsError checks if an error is a WorktreeExistsError
func IsWorktreeExistsError(err error) bool {
	return errors.Is(err, &WorktreeExistsError{})
}

// IsGitHubAPIError checks if an error is a GitHubAPIError
func IsGitHubAPIError(err error) bool {
	return errors.Is(err, &GitHubAPIError{})
}

// IsCommandExecutionError checks if an error is a CommandExecutionError
func IsCommandExecutionError(err error) bool {
	return errors.Is(err, &CommandExecutionError{})
}

// IsInvalidInputError checks if an error is an InvalidInputError
func IsInvalidInputError(err error) bool {
	return errors.Is(err, &InvalidInputError{})
}

// IsWorktreeNotFoundError checks if an error is a WorktreeNotFoundError
func IsWorktreeNotFoundError(err error) bool {
	return errors.Is(err, &WorktreeNotFoundError{})
}

// IsNotAGitRepoError checks if an error is a NotAGitRepoError
func IsNotAGitRepoError(err error) bool {
	return errors.Is(err, &NotAGitRepoError{})
}

// IsFzfNotInstalledError checks if an error is a FzfNotInstalledError
func IsFzfNotInstalledError(err error) bool {
	return errors.Is(err, &FzfNotInstalledError{})
}

// IsNotInWorktreeError checks if an error is a NotInWorktreeError
func IsNotInWorktreeError(err error) bool {
	return errors.Is(err, &NotInWorktreeError{})
}
