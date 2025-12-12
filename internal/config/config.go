package config

// Config represents the application configuration.
type Config struct {
	Add    AddConfig   `yaml:"add"`
	Close  CloseConfig `yaml:"close"`
	Rm     RmConfig    `yaml:"rm"`
	Editor string      `yaml:"editor,omitempty"`
}

// AddConfig represents the configuration for the add command.
type AddConfig struct {
	Open        bool `yaml:"open"`
	Sync        bool `yaml:"sync"`
	SyncIgnored bool `yaml:"sync_ignored"`
}

// CloseConfig represents the configuration for the close command.
type CloseConfig struct {
	Force bool `yaml:"force"`
}

// RmConfig represents the configuration for the rm command.
type RmConfig struct {
	Force bool `yaml:"force"`
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{
		Add: AddConfig{
			Open:        false,
			Sync:        false,
			SyncIgnored: false,
		},
		Close: CloseConfig{
			Force: false,
		},
		Rm: RmConfig{
			Force: false,
		},
		Editor: "",
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Currently no validation rules, but this method is here for future use
	return nil
}

// MergeWithFlags merges the configuration with command-line flags.
// Flags take precedence over config file values.
func (c *Config) MergeWithFlags(openFlag *bool, editorFlag *string, closeYesFlag *bool, rmYesFlag *bool, syncFlag *bool, syncIgnoredFlag *bool) *Config {
	merged := &Config{
		Add:    c.Add,
		Close:  c.Close,
		Rm:     c.Rm,
		Editor: c.Editor,
	}

	if openFlag != nil {
		merged.Add.Open = *openFlag
	}

	if syncFlag != nil {
		merged.Add.Sync = *syncFlag
	}

	if syncIgnoredFlag != nil {
		merged.Add.SyncIgnored = *syncIgnoredFlag
	}

	if editorFlag != nil && *editorFlag != "" {
		merged.Editor = *editorFlag
	}

	if closeYesFlag != nil {
		merged.Close.Force = *closeYesFlag
	}

	if rmYesFlag != nil {
		merged.Rm.Force = *rmYesFlag
	}

	return merged
}

// GetEditor returns the editor command if add.open is true, otherwise empty string.
func (c *Config) GetEditor() string {
	if !c.Add.Open {
		return ""
	}
	return c.Editor
}
