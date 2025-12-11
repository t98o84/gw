package cmd

import (
	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/shell"
)

// selectWorktreeWithFzf shows an interactive worktree selector using fzf
// excludeMain: if true, excludes the main worktree from the list
func selectWorktreeWithFzf(excludeMain bool) (*git.Worktree, error) {
	selector := fzf.NewSelector(shell.NewRealExecutor())
	return selectWorktreeWithSelector(selector, excludeMain)
}

func selectWorktreeWithSelector(selector fzf.Selector, excludeMain bool) (*git.Worktree, error) {
	worktrees, err := git.List()
	if err != nil {
		return nil, err
	}
	// Convert []Worktree to []*Worktree
	worktreePtrs := make([]*git.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}
	return selector.SelectWorktree(worktreePtrs, excludeMain)
}

// selectWorktreesWithFzf shows an interactive worktree selector using fzf with multi-select support
// excludeMain: if true, excludes the main worktree from the list
// multi: if true, allows selecting multiple worktrees (use Tab to select)
func selectWorktreesWithFzf(excludeMain bool, multi bool) ([]*git.Worktree, error) {
	selector := fzf.NewSelector(shell.NewRealExecutor())
	return selectWorktreesWithSelector(selector, excludeMain, multi)
}

func selectWorktreesWithSelector(selector fzf.Selector, excludeMain bool, multi bool) ([]*git.Worktree, error) {
	worktrees, err := git.List()
	if err != nil {
		return nil, err
	}
	// Convert []Worktree to []*Worktree
	worktreePtrs := make([]*git.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}
	return selector.SelectWorktrees(worktreePtrs, excludeMain, multi)
}
