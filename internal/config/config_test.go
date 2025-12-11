package config

import "testing"

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if cfg.Add.Open != false {
		t.Errorf("NewConfig() Add.Open = %v, want false", cfg.Add.Open)
	}
	if cfg.Editor != "" {
		t.Errorf("NewConfig() Editor = %v, want empty string", cfg.Editor)
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := NewConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestConfig_MergeWithFlags(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		openFlag   *bool
		editorFlag *string
		wantOpen   bool
		wantEditor string
	}{
		{
			name: "no flags provided",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: nil,
			wantOpen:   false,
			wantEditor: "vim",
		},
		{
			name: "open flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Editor: "vim",
			},
			openFlag:   boolPtr(true),
			editorFlag: nil,
			wantOpen:   true,
			wantEditor: "vim",
		},
		{
			name: "editor flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: stringPtr("code"),
			wantOpen:   true,
			wantEditor: "code",
		},
		{
			name: "both flags override config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Editor: "vim",
			},
			openFlag:   boolPtr(true),
			editorFlag: stringPtr("emacs"),
			wantOpen:   true,
			wantEditor: "emacs",
		},
		{
			name: "empty editor flag doesn't override",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: stringPtr(""),
			wantOpen:   true,
			wantEditor: "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.config.MergeWithFlags(tt.openFlag, tt.editorFlag)
			if merged.Add.Open != tt.wantOpen {
				t.Errorf("MergeWithFlags() Add.Open = %v, want %v", merged.Add.Open, tt.wantOpen)
			}
			if merged.Editor != tt.wantEditor {
				t.Errorf("MergeWithFlags() Editor = %v, want %v", merged.Editor, tt.wantEditor)
			}
		})
	}
}

func TestConfig_GetEditor(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "open is true, editor is set",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Editor: "code",
			},
			want: "code",
		},
		{
			name: "open is false, editor is set",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Editor: "code",
			},
			want: "",
		},
		{
			name: "open is true, editor is empty",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Editor: "",
			},
			want: "",
		},
		{
			name: "open is false, editor is empty",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Editor: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetEditor()
			if got != tt.want {
				t.Errorf("GetEditor() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for creating pointers
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
