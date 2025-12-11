package cmd

import (
	"testing"

	"github.com/t98o84/gw/internal/errors"
)

func TestNewConfig(t *testing.T) {
	t.Run("creates config with default values", func(t *testing.T) {
		cfg := NewConfig()

		if cfg.AddCreateBranch {
			t.Errorf("Expected AddCreateBranch to be false, got %v", cfg.AddCreateBranch)
		}
		if cfg.AddPRIdentifier != "" {
			t.Errorf("Expected AddPRIdentifier to be empty, got %q", cfg.AddPRIdentifier)
		}
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

	t.Run("valid config with only AddCreateBranch", func(t *testing.T) {
		cfg := &Config{
			AddCreateBranch: true,
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("valid config with only AddPRIdentifier", func(t *testing.T) {
		cfg := &Config{
			AddPRIdentifier: "123",
		}
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

	t.Run("invalid config with both AddCreateBranch and AddPRIdentifier", func(t *testing.T) {
		cfg := &Config{
			AddCreateBranch: true,
			AddPRIdentifier: "123",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("Expected an error, got nil")
		}
		if !errors.IsInvalidInputError(err) {
			t.Errorf("Expected InvalidInputError, got %T", err)
		}
	})

	t.Run("valid config with all Add flags false", func(t *testing.T) {
		cfg := &Config{
			AddCreateBranch: false,
			AddPRIdentifier: "",
			SwPrintPath:     true,
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestConfigFieldAccess(t *testing.T) {
	t.Run("can access and modify fields", func(t *testing.T) {
		cfg := NewConfig()

		cfg.AddCreateBranch = true
		if !cfg.AddCreateBranch {
			t.Error("Failed to set AddCreateBranch")
		}

		cfg.AddPRIdentifier = "456"
		if cfg.AddPRIdentifier != "456" {
			t.Errorf("Expected AddPRIdentifier to be '456', got %q", cfg.AddPRIdentifier)
		}

		cfg.SwPrintPath = true
		if !cfg.SwPrintPath {
			t.Error("Failed to set SwPrintPath")
		}
	})
}
