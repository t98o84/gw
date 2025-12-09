package fzf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/shell"
)

// Selector is an interface for interactive selection
type Selector interface {
	// SelectBranch shows branch selector and returns selected branch
	SelectBranch(branches []string) (string, error)

	// SelectWorktree shows worktree selector and returns selected worktree
	SelectWorktree(worktrees []git.Worktree, excludeMain bool) (*git.Worktree, error)

	// SelectWorktrees shows worktree selector with multi-select support
	SelectWorktrees(worktrees []git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error)

	// IsAvailable checks if fzf is installed
	IsAvailable() bool
}

// fzfExecutor is a function type for executing fzf commands
type fzfExecutor func(args []string, input string) (string, error)

// FzfSelector implements Selector using fzf
type FzfSelector struct {
	executor    shell.Executor
	fzfExecutor fzfExecutor
}

// NewSelector creates a new FzfSelector
func NewSelector(executor shell.Executor) Selector {
	s := &FzfSelector{executor: executor}
	s.fzfExecutor = s.defaultFzfExecutor
	return s
}

// IsAvailable checks if fzf is installed
func (s *FzfSelector) IsAvailable() bool {
	_, err := s.executor.LookPath("fzf")
	return err == nil
}

// SelectBranch shows an interactive branch selector using fzf
func (s *FzfSelector) SelectBranch(branches []string) (string, error) {
	if !s.IsAvailable() {
		return "", fmt.Errorf("fzf is not installed. Please install fzf for interactive selection, or specify a branch name")
	}

	if len(branches) == 0 {
		return "", fmt.Errorf("no branches found")
	}

	// Prepare fzf command
	args := []string{"--height=40%", "--reverse", "--prompt=Select branch: "}
	input := strings.Join(branches, "\n")

	// Execute fzf
	output, err := s.fzfExecutor(args, input)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// SelectWorktree shows an interactive worktree selector using fzf
func (s *FzfSelector) SelectWorktree(worktrees []git.Worktree, excludeMain bool) (*git.Worktree, error) {
	selected, err := s.SelectWorktrees(worktrees, excludeMain, false)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, nil
	}
	return selected[0], nil
}

// SelectWorktrees shows an interactive worktree selector with multi-select support
func (s *FzfSelector) SelectWorktrees(worktrees []git.Worktree, excludeMain bool, multi bool) ([]*git.Worktree, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("fzf is not installed. Please install fzf for interactive selection, or specify a worktree name")
	}

	if len(worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees found")
	}

	// Build list for fzf
	var items []string
	wtMap := make(map[string]*git.Worktree)
	for i := range worktrees {
		wt := &worktrees[i]
		if excludeMain && wt.IsMain {
			continue
		}
		dirName := filepath.Base(wt.Path)
		label := dirName
		if wt.IsMain {
			label = dirName + " (main)"
		}
		items = append(items, label)
		wtMap[label] = wt
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no worktrees available for selection")
	}

	// Prepare fzf command
	var args []string
	if multi {
		args = []string{"--height=40%", "--reverse", "--multi", "--prompt=Select worktrees (Tab to multi-select): "}
	} else {
		args = []string{"--height=40%", "--reverse", "--prompt=Select worktree: "}
	}
	input := strings.Join(items, "\n")

	// Execute fzf
	output, err := s.fzfExecutor(args, input)
	if err != nil {
		return nil, err
	}

	// Parse selected items
	if output == "" {
		return nil, nil
	}

	selectedLabels := strings.Split(output, "\n")
	var selected []*git.Worktree
	for _, label := range selectedLabels {
		label = strings.TrimSpace(label)
		if label != "" {
			if wt, ok := wtMap[label]; ok {
				selected = append(selected, wt)
			}
		}
	}

	return selected, nil
}

// defaultFzfExecutor is the default implementation that uses exec.Command
func (s *FzfSelector) defaultFzfExecutor(args []string, input string) (string, error) {
	return executeFzf(args, input)
}

// executeFzf executes fzf with given arguments and input
// Returns the selected output or empty string if cancelled
func executeFzf(args []string, input string) (string, error) {
	// Create command
	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	// Execute
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 130 means Ctrl+C (user cancelled)
			if exitErr.ExitCode() == 130 {
				return "", nil
			}
			// Exit code 1 means no selection
			if exitErr.ExitCode() == 1 {
				return "", nil
			}
		}
		return "", fmt.Errorf("fzf execution failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
