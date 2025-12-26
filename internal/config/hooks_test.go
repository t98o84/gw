package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteHooks(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func() (string, *ProjectConfig, string, string)
		expectError  bool
		validateFunc func(*testing.T, string)
	}{
		{
			name: "nil project config",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				return dir, nil, "main", dir
			},
			expectError: false,
		},
		{
			name: "empty hooks",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{},
					},
				}
				return dir, cfg, "main", dir
			},
			expectError: false,
		},
		{
			name: "command hook success",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: "echo 'test' > test.txt",
							},
						},
					},
				}
				return dir, cfg, "feature/test", dir
			},
			expectError: false,
			validateFunc: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, "test.txt"))
				if err != nil {
					t.Errorf("failed to read command output: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file from command")
				}
			},
		},
		{
			name: "command with env vars",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: "echo $TEST_VAR > output.txt",
								Env: map[string]string{
									"TEST_VAR": "hello",
								},
							},
						},
					},
				}
				return dir, cfg, "feature/test", dir
			},
			expectError: false,
			validateFunc: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, "output.txt"))
				if err != nil {
					t.Errorf("failed to read command output: %v", err)
				}
				if !strings.Contains(string(content), "hello") {
					t.Errorf("expected 'hello' in output, got: %s", string(content))
				}
			},
		},
		{
			name: "gw environment variables are available",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				repoRoot := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: "echo \"$GW_BRANCH|$GW_WORKTREE_PATH|$GW_REPO_ROOT\" > env_test.txt",
							},
						},
					},
				}
				return dir, cfg, "feature/test", repoRoot
			},
			expectError: false,
			validateFunc: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, "env_test.txt"))
				if err != nil {
					t.Errorf("failed to read command output: %v", err)
				}
				output := strings.TrimSpace(string(content))
				if !strings.Contains(output, "feature/test") {
					t.Errorf("expected GW_BRANCH in output, got: %s", output)
				}
				if !strings.Contains(output, dir) {
					t.Errorf("expected GW_WORKTREE_PATH in output, got: %s", output)
				}
			},
		},
		{
			name: "multiple hooks",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: "echo 'first' > first.txt",
							},
							{
								Command: "echo 'second' > second.txt",
							},
						},
					},
				}
				return dir, cfg, "main", dir
			},
			expectError: false,
			validateFunc: func(t *testing.T, dir string) {
				if _, err := os.Stat(filepath.Join(dir, "first.txt")); os.IsNotExist(err) {
					t.Error("first.txt file not created")
				}
				if _, err := os.Stat(filepath.Join(dir, "second.txt")); os.IsNotExist(err) {
					t.Error("second.txt file not created")
				}
			},
		},
		{
			name: "multiline command",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: `echo "Line 1" > output.txt
echo "Line 2" >> output.txt
echo "Line 3" >> output.txt`,
							},
						},
					},
				}
				return dir, cfg, "feature/multiline", dir
			},
			expectError: false,
			validateFunc: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, "output.txt"))
				if err != nil {
					t.Errorf("failed to read command output: %v", err)
				}
				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				if len(lines) != 3 {
					t.Errorf("expected 3 lines, got %d", len(lines))
				}
				if !strings.Contains(string(content), "Line 1") ||
					!strings.Contains(string(content), "Line 2") ||
					!strings.Contains(string(content), "Line 3") {
					t.Errorf("expected all three lines in output, got: %s", string(content))
				}
			},
		},
		{
			name: "failing command",
			setupFunc: func() (string, *ProjectConfig, string, string) {
				dir := t.TempDir()
				cfg := &ProjectConfig{
					Hooks: HooksConfig{
						PostAdd: []Hook{
							{
								Command: "exit 1",
							},
						},
					},
				}
				return dir, cfg, "main", dir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cfg, branch, repoRoot := tt.setupFunc()
			err := ExecuteHooks(cfg, HookPostAdd, dir, branch, repoRoot)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, dir)
			}
		})
	}
}

func TestExecuteCommandHook(t *testing.T) {
	tests := []struct {
		name        string
		hook        Hook
		expectError bool
	}{
		{
			name: "missing command field",
			hook: Hook{
				Command: "",
			},
			expectError: true,
		},
		{
			name: "command with env vars",
			hook: Hook{
				Command: "echo $TEST_VAR > output.txt",
				Env: map[string]string{
					"TEST_VAR": "hello",
				},
			},
			expectError: false,
		},
		{
			name: "failing command",
			hook: Hook{
				Command: "exit 1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			err := executeCommandHook(tt.hook, dir, "test-branch", dir, 0)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
