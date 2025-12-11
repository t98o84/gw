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
	if cfg.Delete.Force != false {
		t.Errorf("NewConfig() Delete.Force = %v, want false", cfg.Delete.Force)
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
		yesFlag    *bool
		wantOpen   bool
		wantEditor string
		wantForce  bool
	}{
		{
			name: "no flags provided",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Delete: DeleteConfig{
					Force: false,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: nil,
			yesFlag:    nil,
			wantOpen:   false,
			wantEditor: "vim",
			wantForce:  false,
		},
		{
			name: "open flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Delete: DeleteConfig{
					Force: false,
				},
				Editor: "vim",
			},
			openFlag:   boolPtr(true),
			editorFlag: nil,
			yesFlag:    nil,
			wantOpen:   true,
			wantEditor: "vim",
			wantForce:  false,
		},
		{
			name: "editor flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Delete: DeleteConfig{
					Force: false,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: stringPtr("code"),
			yesFlag:    nil,
			wantOpen:   true,
			wantEditor: "code",
			wantForce:  false,
		},
		{
			name: "yes flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Delete: DeleteConfig{
					Force: false,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: nil,
			yesFlag:    boolPtr(true),
			wantOpen:   false,
			wantEditor: "vim",
			wantForce:  true,
		},
		{
			name: "all flags override config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Delete: DeleteConfig{
					Force: false,
				},
				Editor: "vim",
			},
			openFlag:   boolPtr(true),
			editorFlag: stringPtr("emacs"),
			yesFlag:    boolPtr(true),
			wantOpen:   true,
			wantEditor: "emacs",
			wantForce:  true,
		},
		{
			name: "empty editor flag doesn't override",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Delete: DeleteConfig{
					Force: true,
				},
				Editor: "vim",
			},
			openFlag:   nil,
			editorFlag: stringPtr(""),
			yesFlag:    nil,
			wantOpen:   true,
			wantEditor: "vim",
			wantForce:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.config.MergeWithFlags(tt.openFlag, tt.editorFlag, tt.yesFlag)
			if merged.Add.Open != tt.wantOpen {
				t.Errorf("MergeWithFlags() Add.Open = %v, want %v", merged.Add.Open, tt.wantOpen)
			}
			if merged.Editor != tt.wantEditor {
				t.Errorf("MergeWithFlags() Editor = %v, want %v", merged.Editor, tt.wantEditor)
			}
			if merged.Delete.Force != tt.wantForce {
				t.Errorf("MergeWithFlags() Delete.Force = %v, want %v", merged.Delete.Force, tt.wantForce)
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
