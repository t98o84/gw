package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() func()
		wantErr bool
	}{
		{
			name: "default config path",
			setup: func() func() {
				return func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			path, err := GetConfigPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if path == "" {
					t.Error("GetConfigPath() returned empty path")
				}
				if filepath.Base(path) != configFileName {
					t.Errorf("GetConfigPath() filename = %v, want %v", filepath.Base(path), configFileName)
				}
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() func()
		wantErr bool
	}{
		{
			name: "default config dir",
			setup: func() func() {
				return func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			dir, err := getConfigDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if dir == "" {
					t.Error("getConfigDir() returned empty dir")
				}
				if filepath.Base(dir) != configDirName {
					t.Errorf("getConfigDir() dirname = %v, want %v", filepath.Base(dir), configDirName)
				}
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Override the config dir for testing
	originalGOOS := runtime.GOOS
	t.Cleanup(func() {
		// Restore original GOOS (though we can't actually change it)
		_ = originalGOOS
	})

	// Set environment variable to use temp directory
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	default:
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	}

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Check that directory was created
	configDir := filepath.Join(tmpDir, configDirName)
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path exists but is not a directory")
	}
}
