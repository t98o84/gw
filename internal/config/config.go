package config

// Config represents the application configuration.
type Config struct {
	Add    AddConfig `yaml:"add"`
	Editor string    `yaml:"editor,omitempty"`
}

// AddConfig represents the configuration for the add command.
type AddConfig struct {
	Open bool `yaml:"open"`
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{
		Add: AddConfig{
			Open: false,
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
func (c *Config) MergeWithFlags(openFlag *bool, editorFlag *string) *Config {
	merged := &Config{
		Add:    c.Add,
		Editor: c.Editor,
	}

	if openFlag != nil {
		merged.Add.Open = *openFlag
	}

	if editorFlag != nil && *editorFlag != "" {
		merged.Editor = *editorFlag
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
