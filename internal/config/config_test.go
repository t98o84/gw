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
	if cfg.Add.Sync != false {
		t.Errorf("NewConfig() Add.Sync = %v, want false", cfg.Add.Sync)
	}
	if cfg.Add.SyncIgnored != false {
		t.Errorf("NewConfig() Add.SyncIgnored = %v, want false", cfg.Add.SyncIgnored)
	}
	if cfg.Close.Force != false {
		t.Errorf("NewConfig() Close.Force = %v, want false", cfg.Close.Force)
	}
	if cfg.Rm.Force != false {
		t.Errorf("NewConfig() Rm.Force = %v, want false", cfg.Rm.Force)
	}
	if cfg.Rm.Branch != false {
		t.Errorf("NewConfig() Rm.Branch = %v, want false", cfg.Rm.Branch)
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
		name              string
		config            *Config
		openFlag          *bool
		editorFlag        *string
		closeYesFlag      *bool
		rmYesFlag         *bool
		rmBranchFlag      *bool
		noOpenFlag        bool
		noSyncFlag        bool
		noSyncIgnoredFlag bool
		closeNoYesFlag    bool
		rmNoYesFlag       bool
		rmNoBranchFlag    bool
		wantOpen          bool
		wantSync          bool
		wantSyncIgnored   bool
		wantEditor        string
		wantCloseForce    bool
		wantRmForce       bool
		wantRmBranch      bool
	}{
		{
			name: "no flags provided",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "open flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          boolPtr(true),
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          true,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "editor flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        stringPtr("code"),
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          true,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "code",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "close yes flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      boolPtr(true),
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    true,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "rm yes flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         boolPtr(true),
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       true,
			wantRmBranch:      false,
		},
		{
			name: "rm branch flag overrides config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      boolPtr(true),
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      true,
		},
		{
			name: "all flags override config",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          boolPtr(true),
			editorFlag:        stringPtr("emacs"),
			closeYesFlag:      boolPtr(true),
			rmYesFlag:         boolPtr(true),
			rmBranchFlag:      boolPtr(true),
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          true,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "emacs",
			wantCloseForce:    true,
			wantRmForce:       true,
			wantRmBranch:      true,
		},
		{
			name: "empty editor flag doesn't override",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Close: CloseConfig{
					Force: true,
				},
				Rm: RmConfig{
					Force:  true,
					Branch: true,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        stringPtr(""),
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          true,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    true,
			wantRmForce:       true,
			wantRmBranch:      true,
		},
		{
			name: "--no-open overrides config and --open flag",
			config: &Config{
				Add: AddConfig{
					Open: true,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          boolPtr(true),
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        true,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "--no-yes overrides config and --yes flag for close",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: true,
				},
				Rm: RmConfig{
					Force:  false,
					Branch: false,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      boolPtr(true),
			rmYesFlag:         nil,
			rmBranchFlag:      nil,
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    true,
			rmNoYesFlag:       false,
			rmNoBranchFlag:    false,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
		{
			name: "--no-yes and --no-branch override config and flags for rm",
			config: &Config{
				Add: AddConfig{
					Open: false,
				},
				Close: CloseConfig{
					Force: false,
				},
				Rm: RmConfig{
					Force:  true,
					Branch: true,
				},
				Editor: "vim",
			},
			openFlag:          nil,
			editorFlag:        nil,
			closeYesFlag:      nil,
			rmYesFlag:         boolPtr(true),
			rmBranchFlag:      boolPtr(true),
			noOpenFlag:        false,
			noSyncFlag:        false,
			noSyncIgnoredFlag: false,
			closeNoYesFlag:    false,
			rmNoYesFlag:       true,
			rmNoBranchFlag:    true,
			wantOpen:          false,
			wantSync:          false,
			wantSyncIgnored:   false,
			wantEditor:        "vim",
			wantCloseForce:    false,
			wantRmForce:       false,
			wantRmBranch:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.config.MergeWithFlags(
				tt.openFlag,
				tt.editorFlag,
				tt.closeYesFlag,
				tt.rmYesFlag,
				tt.rmBranchFlag,
				nil,
				nil,
				tt.noOpenFlag,
				tt.noSyncFlag,
				tt.noSyncIgnoredFlag,
				tt.closeNoYesFlag,
				tt.rmNoYesFlag,
				tt.rmNoBranchFlag,
			)
			if merged.Add.Open != tt.wantOpen {
				t.Errorf("MergeWithFlags() Add.Open = %v, want %v", merged.Add.Open, tt.wantOpen)
			}
			if merged.Add.Sync != tt.wantSync {
				t.Errorf("MergeWithFlags() Add.Sync = %v, want %v", merged.Add.Sync, tt.wantSync)
			}
			if merged.Add.SyncIgnored != tt.wantSyncIgnored {
				t.Errorf("MergeWithFlags() Add.SyncIgnored = %v, want %v", merged.Add.SyncIgnored, tt.wantSyncIgnored)
			}
			if merged.Editor != tt.wantEditor {
				t.Errorf("MergeWithFlags() Editor = %v, want %v", merged.Editor, tt.wantEditor)
			}
			if merged.Close.Force != tt.wantCloseForce {
				t.Errorf("MergeWithFlags() Close.Force = %v, want %v", merged.Close.Force, tt.wantCloseForce)
			}
			if merged.Rm.Force != tt.wantRmForce {
				t.Errorf("MergeWithFlags() Rm.Force = %v, want %v", merged.Rm.Force, tt.wantRmForce)
			}
			if merged.Rm.Branch != tt.wantRmBranch {
				t.Errorf("MergeWithFlags() Rm.Branch = %v, want %v", merged.Rm.Branch, tt.wantRmBranch)
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
