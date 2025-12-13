package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/t98o84/gw/internal/errors"
	"github.com/t98o84/gw/internal/fzf"
	"github.com/t98o84/gw/internal/git"
	"github.com/t98o84/gw/internal/github"
)

// syncMode represents the synchronization mode for worktree creation
type syncMode int

const (
	syncNone syncMode = iota
	syncAll
	syncIgnored
)

// Mock functions for testing - nil in production
var (
	mockGetPRBranch        func(prIdentifier, repoName string) (string, error)
	mockListBranches       func() ([]string, error)
	mockFindWorktree       func(branch string) (*git.Worktree, error)
	mockBranchExists       func(branch string) (bool, error)
	mockRemoteBranchExists func(branch string) (bool, error)
	mockFetchBranch        func(branch string) error
	mockWorktreePath       func(repoName, branch string) (string, error)
	mockAdd                func(path string, branch string, createBranch bool) error
	mockOpenInEditor       func(editor, path string) error
)

// addOptions contains options for worktree creation
type addOptions struct {
	createBranch bool
	prIdentifier string
	selector     fzf.Selector
}

// determineBranch determines which branch to use based on args and options
func determineBranch(args []string, opts *addOptions, repoName string) (string, error) {
	// Handle PR flag
	if opts.prIdentifier != "" {
		branch, err := getBranchFromPR(opts.prIdentifier, repoName)
		if err != nil {
			return "", err
		}
		return branch, nil
	}

	// Interactive selection if no args
	if len(args) == 0 {
		branch, err := selectBranchInteractive(opts.selector)
		if err != nil {
			return "", err
		}
		return branch, nil
	}

	// Use provided branch name
	return args[0], nil
}

// getBranchFromPR retrieves branch name from PR identifier
func getBranchFromPR(prIdentifier, repoName string) (string, error) {
	if mockGetPRBranch != nil {
		return mockGetPRBranch(prIdentifier, repoName)
	}
	branch, err := github.GetPRBranch(prIdentifier, repoName)
	if err != nil {
		return "", fmt.Errorf("failed to get PR branch: %w", err)
	}
	return branch, nil
}

// selectBranchInteractive shows interactive branch selector
func selectBranchInteractive(selector fzf.Selector) (string, error) {
	var branches []string
	var err error
	if mockListBranches != nil {
		branches, err = mockListBranches()
	} else {
		branches, err = git.ListBranches()
	}
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	branch, err := selector.SelectBranch(branches)
	if err != nil {
		return "", err
	}

	return branch, nil
}

// checkExistingWorktree checks if worktree already exists for the branch
func checkExistingWorktree(branch string) (*git.Worktree, error) {
	var existing *git.Worktree
	var err error
	if mockFindWorktree != nil {
		existing, err = mockFindWorktree(branch)
	} else {
		existing, err = git.FindWorktree(branch)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check existing worktree: %w", err)
	}
	return existing, nil
}

// ensureBranchExists checks and fetches branch if necessary
func ensureBranchExists(branch string, createBranch bool, fromPR bool) error {
	// If creating a new branch (and not from PR), it will be created with worktree
	if createBranch && !fromPR {
		return nil
	}

	// Check if branch exists locally
	var exists bool
	var err error
	if mockBranchExists != nil {
		exists, err = mockBranchExists(branch)
	} else {
		exists, err = git.BranchExists(branch)
	}
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}

	if !exists {
		// Try to fetch from remote
		var remoteExists bool
		if mockRemoteBranchExists != nil {
			remoteExists, err = mockRemoteBranchExists(branch)
		} else {
			remoteExists, err = git.RemoteBranchExists(branch)
		}
		if err != nil {
			return fmt.Errorf("failed to check remote branch: %w", err)
		}

		if remoteExists {
			fmt.Printf("Fetching branch %s from origin...\n", branch)
			if mockFetchBranch != nil {
				err = mockFetchBranch(branch)
			} else {
				err = git.FetchBranch(branch)
			}
			if err != nil {
				return err
			}
		} else if !createBranch {
			return errors.NewBranchNotFoundError(branch, nil)
		}
	}

	return nil
}

// syncFiles synchronizes files from main worktree to the new worktree
func syncFiles(wtPath string, mode syncMode) error {
	mainWtPath, err := getMainWorktreePath()
	if err != nil {
		return fmt.Errorf("failed to get main worktree path: %w", err)
	}

	switch mode {
	case syncAll:
		return syncAllDiffs(mainWtPath, wtPath)
	case syncIgnored:
		return syncIgnoredFiles(mainWtPath, wtPath)
	default:
		return nil
	}
}

// getMainWorktreePath returns the path of the main worktree
func getMainWorktreePath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// syncAllDiffs syncs all files with differences between main worktree and HEAD
func syncAllDiffs(mainWtPath, newWtPath string) error {
	fmt.Println("Syncing all changed files...")

	// Get all modified, untracked, and staged files
	cmd := exec.Command("git", "-C", mainWtPath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	copiedCount := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse git status output
		if len(line) < 4 {
			continue
		}

		filePath := strings.TrimSpace(line[3:])
		if filePath == "" {
			continue
		}

		// Handle renamed files (e.g., "R  old.txt -> new.txt")
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			filePath = strings.TrimSpace(parts[1])
		}

		srcPath := filepath.Join(mainWtPath, filePath)
		dstPath := filepath.Join(newWtPath, filePath)

		// Check if source file exists (skip deleted files)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Printf("  Warning: Failed to copy %s: %v\n", filePath, err)
			continue
		}
		copiedCount++
	}

	fmt.Printf("✓ Synced %d changed files\n", copiedCount)
	return nil
}

// syncIgnoredFiles syncs gitignored files from main worktree
func syncIgnoredFiles(mainWtPath, newWtPath string) error {
	fmt.Println("Syncing gitignored files...")

	// Get list of all ignored files (including those in global gitignore)
	// Using 'git status --ignored --porcelain' to include both local and global ignores
	cmd := exec.Command("git", "-C", mainWtPath, "status", "--ignored", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list ignored files: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	copiedCount := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse git status output - look for ignored files ("!!" prefix)
		if len(line) < 4 || !strings.HasPrefix(line, "!!") {
			continue
		}

		filePath := strings.TrimSpace(line[3:])
		if filePath == "" {
			continue
		}

		srcPath := filepath.Join(mainWtPath, filePath)
		dstPath := filepath.Join(newWtPath, filePath)

		// Check if source is a file (skip directories)
		info, err := os.Stat(srcPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Printf("  Warning: Failed to copy %s: %v\n", filePath, err)
			continue
		}
		copiedCount++
	}

	fmt.Printf("✓ Synced %d gitignored files\n", copiedCount)
	return nil
}

// copyFile copies a file from src to dst, creating directories as needed
func copyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// determineSyncMode determines which sync mode to use
func determineSyncMode(configSync, configSyncIgnored, flagSyncAll, flagSyncIgnored bool) syncMode {
	if flagSyncAll {
		return syncAll
	}
	if flagSyncIgnored {
		return syncIgnored
	}
	if configSync {
		return syncAll
	}
	if configSyncIgnored {
		return syncIgnored
	}
	return syncNone
}

// createWorktree creates a new worktree for the given branch
func createWorktree(repoName, branch string, createBranch bool, openEditor string, mode syncMode) error {
	var wtPath string
	var err error
	if mockWorktreePath != nil {
		wtPath, err = mockWorktreePath(repoName, branch)
	} else {
		wtPath, err = git.WorktreePath(repoName, branch)
	}
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	fmt.Printf("Creating worktree at %s for branch %s...\n", wtPath, branch)
	if mockAdd != nil {
		err = mockAdd(wtPath, branch, createBranch)
	} else {
		err = git.Add(wtPath, branch, createBranch)
	}
	if err != nil {
		return err
	}

	fmt.Printf("✓ Worktree created: %s\n", wtPath)

	// Sync files if requested
	if mode != syncNone {
		if err := syncFiles(wtPath, mode); err != nil {
			fmt.Printf("⚠ Warning: Failed to sync files: %v\n", err)
		}
	}

	// エディターで開く処理
	if openEditor != "" {
		if err := openInEditor(openEditor, wtPath); err != nil {
			// エディター起動失敗は警告のみ（ワークツリー作成は成功扱い）
			fmt.Printf("⚠ Warning: Failed to open editor: %v\n", err)
		}
	}

	return nil
}

// openInEditor opens the specified path in the given editor
func openInEditor(editor, path string) error {
	if mockOpenInEditor != nil {
		return mockOpenInEditor(editor, path)
	}

	// エディターコマンドの存在確認
	_, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("editor command '%s' not found in PATH", editor)
	}

	// エディターコマンドの実行（バックグラウンド）
	cmd := exec.Command(editor, path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start editor '%s': %w", editor, err)
	}

	fmt.Printf("✓ Opening in %s: %s\n", editor, path)

	// Release the process so it's not a child of this process
	if err := cmd.Process.Release(); err != nil {
		// Non-critical error, just log it
		fmt.Printf("⚠ Warning: Failed to release editor process: %v\n", err)
	}

	return nil
}
