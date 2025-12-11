package cmd

// Config holds all configuration values for the CLI commands
type Config struct {
	// Sw command flags
	SwPrintPath bool
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		SwPrintPath: false,
	}
}

// Validate checks if the configuration is valid.
// Returns nil if the configuration is valid.
func (c *Config) Validate() error {
	return nil
}
