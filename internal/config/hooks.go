package config

import (
	"fmt"
	"os"
	"os/exec"
)

// HookType represents the type of hook to execute
type HookType string

const (
	HookPreAdd     HookType = "pre_add"
	HookPostAdd    HookType = "post_add"
	HookPreRemove  HookType = "pre_remove"
	HookPostRemove HookType = "post_remove"
)

// ExecuteHooks executes hooks of the specified type
func ExecuteHooks(projectConfig *ProjectConfig, hookType HookType, worktreePath, branch, repoRoot string) error {
	if projectConfig == nil {
		return nil
	}

	var hooks []Hook
	switch hookType {
	case HookPreAdd:
		hooks = projectConfig.Hooks.PreAdd
	case HookPostAdd:
		hooks = projectConfig.Hooks.PostAdd
	case HookPreRemove:
		hooks = projectConfig.Hooks.PreRemove
	case HookPostRemove:
		hooks = projectConfig.Hooks.PostRemove
	default:
		return fmt.Errorf("unknown hook type: %s", hookType)
	}

	if len(hooks) == 0 {
		return nil
	}

	// Determine the working directory for the hook commands.
	// post_add and pre_remove run inside the target worktree, since it exists
	// at that point and is the natural place to operate on. pre_add (worktree
	// not yet created) and post_remove (worktree already deleted) run in the
	// main repository root.
	workDir := hookWorkDir(hookType, worktreePath, repoRoot)

	for i, hook := range hooks {
		if err := executeHook(hook, workDir, worktreePath, branch, repoRoot, i); err != nil {
			return fmt.Errorf("hook %d failed: %w", i+1, err)
		}
	}

	return nil
}

// hookWorkDir returns the directory in which a hook's command should run.
func hookWorkDir(hookType HookType, worktreePath, repoRoot string) string {
	switch hookType {
	case HookPostAdd, HookPreRemove:
		return worktreePath
	default: // HookPreAdd, HookPostRemove
		return repoRoot
	}
}

func executeHook(hook Hook, workDir, worktreePath, branch, repoRoot string, index int) error {
	return executeCommandHook(hook, workDir, worktreePath, branch, repoRoot, index)
}

func executeCommandHook(hook Hook, workDir, worktreePath, branch, repoRoot string, index int) error {
	if hook.Command == "" {
		return fmt.Errorf("command hook requires 'command' field")
	}

	fmt.Printf("⚙️  Hook %d: Executing command: %s\n", index+1, hook.Command)

	cmd := exec.Command("sh", "-c", hook.Command)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables.
	cmd.Env = os.Environ()
	// Add gw-specific environment variables first, as defaults.
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("GW_WORKTREE_PATH=%s", worktreePath),
		fmt.Sprintf("GW_BRANCH=%s", branch),
		fmt.Sprintf("GW_REPO_ROOT=%s", repoRoot),
	)
	// Add user-defined environment variables last so they take precedence.
	// os/exec dedups duplicate keys with a last-one-wins rule, so a user can
	// override gw-provided variables (GW_*) by setting them in the env field.
	for key, value := range hook.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	fmt.Printf("✅ Hook %d: Command completed successfully\n", index+1)
	return nil
}
