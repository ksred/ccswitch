package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test default values
	if cfg.Branch.Prefix != "feature/" {
		t.Errorf("Default Branch.Prefix = %q, expected %q", cfg.Branch.Prefix, "feature/")
	}
	if cfg.Worktree.RelativePath != "../" {
		t.Errorf("Default Worktree.RelativePath = %q, expected %q", cfg.Worktree.RelativePath, "../")
	}
	if !cfg.UI.ShowEmoji {
		t.Error("Default UI.ShowEmoji should be true")
	}
	if cfg.UI.ColorScheme != "default" {
		t.Errorf("Default UI.ColorScheme = %q, expected %q", cfg.UI.ColorScheme, "default")
	}
	if cfg.Git.DefaultBranch != "main" {
		t.Errorf("Default Git.DefaultBranch = %q, expected %q", cfg.Git.DefaultBranch, "main")
	}
	if cfg.Git.AutoFetch {
		t.Error("Default Git.AutoFetch should be false")
	}
}

func TestLoadWithNoConfigFile(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no config file failed: %v", err)
	}

	// Should return default config
	defaultCfg := DefaultConfig()
	if cfg.Branch.Prefix != defaultCfg.Branch.Prefix {
		t.Errorf("Load() without config file should return default config")
	}
}

func TestLoadWithValidConfigFile(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir := filepath.Join(tempDir, ".ccswitch")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write a test config file
	configContent := `branch:
  prefix: "custom/"
worktree:
  relative_path: "/custom/path"
ui:
  show_emoji: false
  color_scheme: "dark"
git:
  default_branch: "develop"
  auto_fetch: true`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check loaded values
	if cfg.Branch.Prefix != "custom/" {
		t.Errorf("Branch.Prefix = %q, expected %q", cfg.Branch.Prefix, "custom/")
	}
	if cfg.Worktree.RelativePath != "/custom/path" {
		t.Errorf("Worktree.RelativePath = %q, expected %q", cfg.Worktree.RelativePath, "/custom/path")
	}
	if cfg.UI.ShowEmoji {
		t.Error("UI.ShowEmoji should be false")
	}
	if cfg.UI.ColorScheme != "dark" {
		t.Errorf("UI.ColorScheme = %q, expected %q", cfg.UI.ColorScheme, "dark")
	}
	if cfg.Git.DefaultBranch != "develop" {
		t.Errorf("Git.DefaultBranch = %q, expected %q", cfg.Git.DefaultBranch, "develop")
	}
	if !cfg.Git.AutoFetch {
		t.Error("Git.AutoFetch should be true")
	}
}

func TestLoadWithInvalidYAML(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir := filepath.Join(tempDir, ".ccswitch")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write invalid YAML
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		// Should return default config even with error
		t.Logf("Load() returned error as expected: %v", err)
	}

	// Should still return default config
	if cfg.Branch.Prefix != "feature/" {
		t.Error("Should return default config when YAML parsing fails")
	}
}

func TestLoadWithPartialConfig(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir := filepath.Join(tempDir, ".ccswitch")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write partial config (missing some fields)
	configContent := `branch:
  prefix: "hotfix/"`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check that specified value is loaded
	if cfg.Branch.Prefix != "hotfix/" {
		t.Errorf("Branch.Prefix = %q, expected %q", cfg.Branch.Prefix, "hotfix/")
	}

	// Check that defaults are applied for missing values
	if cfg.Worktree.RelativePath != "../" {
		t.Errorf("Missing Worktree.RelativePath should default to %q", "../")
	}
	if cfg.UI.ColorScheme != "default" {
		t.Errorf("Missing UI.ColorScheme should default to %q", "default")
	}
}

func TestSave(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{}
	cfg.Branch.Prefix = "test/"
	cfg.UI.ShowEmoji = false
	cfg.Git.DefaultBranch = "master"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check that file was created
	configPath := filepath.Join(tempDir, ".ccswitch", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load the saved config
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.Branch.Prefix != "test/" {
		t.Errorf("Saved Branch.Prefix = %q, expected %q", loadedCfg.Branch.Prefix, "test/")
	}
	if loadedCfg.UI.ShowEmoji {
		t.Error("Saved UI.ShowEmoji should be false")
	}
	if loadedCfg.Git.DefaultBranch != "master" {
		t.Errorf("Saved Git.DefaultBranch = %q, expected %q", loadedCfg.Git.DefaultBranch, "master")
	}
}

func TestGetConfigPath(t *testing.T) {
	// Create a temporary directory for HOME
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".ccswitch", "config.yaml")
	actual := GetConfigPath()

	if actual != expected {
		t.Errorf("GetConfigPath() = %q, expected %q", actual, expected)
	}
}
