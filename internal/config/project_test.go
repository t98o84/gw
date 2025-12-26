package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() string
		expectedNil bool
		expectError bool
	}{
		{
			name: "no config file",
			setupFunc: func() string {
				dir := t.TempDir()
				return dir
			},
			expectedNil: true,
			expectError: false,
		},
		{
			name: "valid config with copy hook",
			setupFunc: func() string {
				dir := t.TempDir()
				content := `hooks:
  post_add:
    - command: cp .env.example .env
`
				err := os.WriteFile(filepath.Join(dir, "gw.yaml"), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expectedNil: false,
			expectError: false,
		},
		{
			name: "valid config with command hook",
			setupFunc: func() string {
				dir := t.TempDir()
				content := `hooks:
  post_add:
    - command: npm install
      env:
        NODE_ENV: development
`
				err := os.WriteFile(filepath.Join(dir, "gw.yaml"), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expectedNil: false,
			expectError: false,
		},
		{
			name: "invalid yaml",
			setupFunc: func() string {
				dir := t.TempDir()
				content := `hooks:
  post_add:
    - type: copy
      from: .env.example
      invalid yaml here
`
				err := os.WriteFile(filepath.Join(dir, "gw.yaml"), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expectedNil: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := tt.setupFunc()
			cfg, err := FindProjectConfig(repoRoot)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedNil {
				if cfg != nil {
					t.Errorf("expected nil config but got: %+v", cfg)
				}
			} else {
				if cfg == nil {
					t.Error("expected non-nil config but got nil")
				}
			}
		})
	}
}

func TestProjectConfigStructure(t *testing.T) {
	dir := t.TempDir()
	content := `hooks:
  post_add:
    - command: cp .env.example .env
    - command: npm install
      env:
        NODE_ENV: development
`
	err := os.WriteFile(filepath.Join(dir, "gw.yaml"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := FindProjectConfig(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("config is nil")
	}

	if len(cfg.Hooks.PostAdd) != 2 {
		t.Errorf("expected 2 hooks, got %d", len(cfg.Hooks.PostAdd))
	}

	// Check first hook (copy command)
	copyHook := cfg.Hooks.PostAdd[0]
	if copyHook.Command != "cp .env.example .env" {
		t.Errorf("expected command 'cp .env.example .env', got '%s'", copyHook.Command)
	}

	// Check second hook (npm install)
	cmdHook := cfg.Hooks.PostAdd[1]
	if cmdHook.Command != "npm install" {
		t.Errorf("expected command 'npm install', got '%s'", cmdHook.Command)
	}
	if cmdHook.Env["NODE_ENV"] != "development" {
		t.Errorf("expected NODE_ENV 'development', got '%s'", cmdHook.Env["NODE_ENV"])
	}
}
