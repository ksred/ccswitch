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
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to load config: %v\n"), err)
		return
	}

	fmt.Println(ui.TitleStyle.Render("⚙️  ccswitch Configuration"))
	fmt.Println()

	fmt.Println(ui.SuccessStyle.Render("Branch:"))
	fmt.Printf("  Prefix: %s\n", ui.InfoStyle.Render(cfg.Branch.Prefix))
	fmt.Println()

	fmt.Println(ui.SuccessStyle.Render("Worktree:"))
	fmt.Printf("  Relative path: %s\n", ui.InfoStyle.Render(cfg.Worktree.RelativePath))
	fmt.Println()

	fmt.Println(ui.SuccessStyle.Render("UI:"))
	fmt.Printf("  Show emoji: %s\n", ui.InfoStyle.Render(fmt.Sprintf("%v", cfg.UI.ShowEmoji)))
	fmt.Printf("  Color scheme: %s\n", ui.InfoStyle.Render(cfg.UI.ColorScheme))
	fmt.Println()

	fmt.Println(ui.SuccessStyle.Render("Git:"))
	fmt.Printf("  Default branch: %s\n", ui.InfoStyle.Render(cfg.Git.DefaultBranch))
	fmt.Printf("  Auto fetch: %s\n", ui.InfoStyle.Render(fmt.Sprintf("%v", cfg.Git.AutoFetch)))
	fmt.Println()

	configPath := config.GetConfigPath()
	fmt.Printf("Config file: %s\n", ui.InfoStyle.Render(configPath))
}

func showConfigPath(cmd *cobra.Command, args []string) {
	fmt.Println(config.GetConfigPath())
}

func initConfig(cmd *cobra.Command, args []string) {
	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to create config: %v\n"), err)
		return
	}

	configPath := config.GetConfigPath()
	fmt.Printf(ui.SuccessStyle.Render("✓ Created default config at: %s\n"), configPath)
	fmt.Println()
	fmt.Println("You can now edit this file to customize ccswitch behavior.")
}