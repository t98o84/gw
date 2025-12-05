package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/t98o84/gw/internal/git"
)

// selectWorktreeWithFzf shows an interactive worktree selector using fzf
// excludeMain: if true, excludes the main worktree from the list
func selectWorktreeWithFzf(excludeMain bool) (*git.Worktree, error) {
	worktrees, err := selectWorktreesWithFzf(excludeMain, false)
	if err != nil {
		return nil, err
	}
	if len(worktrees) == 0 {
		return nil, nil
	}
	return worktrees[0], nil
}

// selectWorktreesWithFzf shows an interactive worktree selector using fzf with multi-select support
// excludeMain: if true, excludes the main worktree from the list
// multi: if true, allows selecting multiple worktrees (use Tab to select)
func selectWorktreesWithFzf(excludeMain bool, multi bool) ([]*git.Worktree, error) {
	// Check if fzf is available
	_, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf("fzf is not installed. Please install fzf for interactive selection, or specify a worktree name")
	}

	// Get worktrees
	worktrees, err := git.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
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

	// Run fzf
	fzfArgs := []string{"--height=40%", "--reverse"}
	if multi {
		fzfArgs = append(fzfArgs, "--multi", "--prompt=Select worktrees (Tab to multi-select): ")
	} else {
		fzfArgs = append(fzfArgs, "--prompt=Select worktree: ")
	}
	fzfCmd := exec.Command("fzf", fzfArgs...)
	fzfCmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	fzfCmd.Stderr = os.Stderr

	out, err := fzfCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			// User pressed Ctrl+C
			return nil, nil
		}
		return nil, nil // fzf cancelled
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return nil, nil
	}

	// Parse selected items (one per line when multi-select)
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
