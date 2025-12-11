package cmd

import (
	"github.com/t98o84/gw/internal/errors"
)

// Config holds all configuration values for the CLI commands
type Config struct {
	// Add command flags
	AddCreateBranch bool
	AddPRIdentifier string

	// Sw command flags
	SwPrintPath bool
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		AddCreateBranch: false,
		AddPRIdentifier: "",
		SwPrintPath:     false,
	}
}

// Validate checks if the configuration is valid.
// It returns an error if:
//   - Both --branch and --pr flags are specified for the add command
//
// Returns nil if the configuration is valid.
func (c *Config) Validate() error {
	// Validate Add command configuration
	if c.AddCreateBranch && c.AddPRIdentifier != "" {
		return errors.NewInvalidInputError(
			"--branch and --pr",
			"cannot be used together",
			nil,
		)
	}

	return nil
}
