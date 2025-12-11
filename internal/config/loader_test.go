package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variable to use temp directory
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	default:
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	}

	tests := []struct {
		name       string
		setupFile  func(t *testing.T)
		wantErr    bool
		wantConfig *Config
	}{
		{
			name: "config file not found",
			setupFile: func(t *testing.T) {
				// Don't create any file
			},
			wantErr: true,
		},
		{
			name: "valid config file",
			setupFile: func(t *testing.T) {
				configDir := filepath.Join(tmpDir, configDirName)
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				configPath := filepath.Join(configDir, configFileName)
				content := `add:
  open: true
editor: code
`
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantErr: false,
			wantConfig: &Config{
				Add: AddConfig{
					Open: true,
				},
				Editor: "code",
			},
		},
		{
			name: "invalid yaml",
			setupFile: func(t *testing.T) {
				configDir := filepath.Join(tmpDir, configDirName)
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				configPath := filepath.Join(configDir, configFileName)
				content := `invalid: [yaml content`
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFile(t)

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantConfig != nil {
				if cfg.Add.Open != tt.wantConfig.Add.Open {
					t.Errorf("Load() Add.Open = %v, want %v", cfg.Add.Open, tt.wantConfig.Add.Open)
				}
				if cfg.Editor != tt.wantConfig.Editor {
					t.Errorf("Load() Editor = %v, want %v", cfg.Editor, tt.wantConfig.Editor)
				}
			}

			// Clean up for next test
			configPath, _ := GetConfigPath()
			os.Remove(configPath)
		})
	}
}

func TestLoadOrDefault(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variable to use temp directory
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	default:
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	}

	t.Run("returns default config when file not found", func(t *testing.T) {
		cfg := LoadOrDefault()
		if cfg == nil {
			t.Fatal("LoadOrDefault() returned nil")
		}
		if cfg.Add.Open != false {
			t.Errorf("LoadOrDefault() Add.Open = %v, want false", cfg.Add.Open)
		}
		if cfg.Editor != "" {
			t.Errorf("LoadOrDefault() Editor = %v, want empty string", cfg.Editor)
		}
	})

	t.Run("loads valid config file", func(t *testing.T) {
		configDir := filepath.Join(tmpDir, configDirName)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}
		configPath := filepath.Join(configDir, configFileName)
		content := `add:
  open: true
editor: vim
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		cfg := LoadOrDefault()
		if cfg == nil {
			t.Fatal("LoadOrDefault() returned nil")
		}
		if cfg.Add.Open != true {
			t.Errorf("LoadOrDefault() Add.Open = %v, want true", cfg.Add.Open)
		}
		if cfg.Editor != "vim" {
			t.Errorf("LoadOrDefault() Editor = %v, want vim", cfg.Editor)
		}
	})
}

func TestSave(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variable to use temp directory
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	default:
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	}

	t.Run("saves config successfully", func(t *testing.T) {
		cfg := &Config{
			Add: AddConfig{
				Open: true,
			},
			Editor: "emacs",
		}

		if err := Save(cfg); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file was created
		configPath, _ := GetConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created")
		}

		// Load and verify contents
		loadedCfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedCfg.Add.Open != cfg.Add.Open {
			t.Errorf("Saved config Add.Open = %v, want %v", loadedCfg.Add.Open, cfg.Add.Open)
		}
		if loadedCfg.Editor != cfg.Editor {
			t.Errorf("Saved config Editor = %v, want %v", loadedCfg.Editor, cfg.Editor)
		}
	})
}
