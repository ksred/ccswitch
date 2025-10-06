package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the ccswitch configuration
type Config struct {
	Branch struct {
		Prefix string `yaml:"prefix"`
	} `yaml:"branch"`
	Worktree struct {
		RelativePath string `yaml:"relative_path"`
	} `yaml:"worktree"`
	UI struct {
		ShowEmoji   bool   `yaml:"show_emoji"`
		ColorScheme string `yaml:"color_scheme"`
	} `yaml:"ui"`
	Git struct {
		DefaultBranch string `yaml:"default_branch"`
		AutoFetch     bool   `yaml:"auto_fetch"`
	} `yaml:"git"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Branch.Prefix = "feature/"
	cfg.Worktree.RelativePath = "../"
	cfg.UI.ShowEmoji = true
	cfg.UI.ColorScheme = "default"
	cfg.Git.DefaultBranch = "main"
	cfg.Git.AutoFetch = false
	return cfg
}

// Load loads configuration from file or returns default
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	configPath := filepath.Join(homeDir, ".ccswitch", "config.yaml")

	// Check if config file exists
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	// Parse YAML
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}

	// Apply defaults for any missing values
	if cfg.Branch.Prefix == "" {
		cfg.Branch.Prefix = "feature/"
	}
	if cfg.Worktree.RelativePath == "" {
		cfg.Worktree.RelativePath = "../"
	}
	if cfg.UI.ColorScheme == "" {
		cfg.UI.ColorScheme = "default"
	}
	if cfg.Git.DefaultBranch == "" {
		cfg.Git.DefaultBranch = "main"
	}

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".ccswitch")
	if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
		return mkdirErr
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ccswitch", "config.yaml")
}
