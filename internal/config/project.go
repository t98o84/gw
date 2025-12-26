package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the project-specific configuration from gw.yaml
type ProjectConfig struct {
	Hooks HooksConfig `yaml:"hooks"`
}

// HooksConfig represents the hooks configuration
type HooksConfig struct {
	PreAdd     []Hook `yaml:"pre_add,omitempty"`
	PostAdd    []Hook `yaml:"post_add,omitempty"`
	PreRemove  []Hook `yaml:"pre_remove,omitempty"`
	PostRemove []Hook `yaml:"post_remove,omitempty"`
}

// Hook represents a single hook action
type Hook struct {
	Command string            `yaml:"command"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// FindProjectConfig searches for gw.yaml in the repository root directory
func FindProjectConfig(repoRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(repoRoot, "gw.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No project config is not an error
		}
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &cfg, nil
}
