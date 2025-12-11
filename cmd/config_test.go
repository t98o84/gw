package cmd

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("creates config with default values", func(t *testing.T) {
		cfg := NewConfig()

		if cfg.SwPrintPath {
			t.Errorf("Expected SwPrintPath to be false, got %v", cfg.SwPrintPath)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	t.Run("valid config with no flags set", func(t *testing.T) {
		cfg := &Config{}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("valid config with only SwPrintPath", func(t *testing.T) {
		cfg := &Config{
			SwPrintPath: true,
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestConfigFieldAccess(t *testing.T) {
	t.Run("can access and modify fields", func(t *testing.T) {
		cfg := NewConfig()

		cfg.SwPrintPath = true
		if !cfg.SwPrintPath {
			t.Error("Failed to set SwPrintPath")
		}
	})
}
