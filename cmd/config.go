package cmd

import (
	"fmt"

	"github.com/ksred/ccswitch/internal/config"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show ccswitch configuration",
		Run:   showConfig,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		Run:   showConfigPath,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create default config file",
		Run:   initConfig,
	})

	return cmd
}

func showConfig(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		ui.Errorf("✗ Failed to load config: %v", err)
		return
	}

	ui.Title("⚙️  ccswitch Configuration")
	fmt.Println()

	ui.Success("Branch:")
	ui.Infof("  Prefix: %s", cfg.Branch.Prefix)
	fmt.Println()

	ui.Success("Worktree:")
	ui.Infof("  Relative path: %s", cfg.Worktree.RelativePath)
	fmt.Println()

	ui.Success("UI:")
	ui.Infof("  Show emoji: %v", cfg.UI.ShowEmoji)
	ui.Infof("  Color scheme: %s", cfg.UI.ColorScheme)
	fmt.Println()

	ui.Success("Git:")
	ui.Infof("  Default branch: %s", cfg.Git.DefaultBranch)
	ui.Infof("  Auto fetch: %v", cfg.Git.AutoFetch)
	fmt.Println()

	configPath := config.GetConfigPath()
	ui.Infof("Config file: %s", configPath)
}

func showConfigPath(cmd *cobra.Command, args []string) {
	fmt.Println(config.GetConfigPath())
}

func initConfig(cmd *cobra.Command, args []string) {
	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		ui.Errorf("✗ Failed to create config: %v", err)
		return
	}

	configPath := config.GetConfigPath()
	ui.Successf("✓ Created default config at: %s", configPath)
	fmt.Println()
	fmt.Println("You can now edit this file to customize ccswitch behavior.")
}