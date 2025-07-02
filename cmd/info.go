package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show ccswitch configuration and paths",
		Run:   showInfo,
	}
}

func showInfo(cmd *cobra.Command, args []string) {
	homeDir, _ := os.UserHomeDir()
	ccSwitchDir := filepath.Join(homeDir, ".ccswitch")
	worktreesDir := filepath.Join(ccSwitchDir, "worktrees")
	
	fmt.Println(ui.TitleStyle.Render("📊 ccswitch Information"))
	fmt.Println()
	
	// Paths
	fmt.Println(ui.SuccessStyle.Render("Paths:"))
	fmt.Printf("  Config directory: %s\n", ui.InfoStyle.Render(ccSwitchDir))
	fmt.Printf("  Worktrees stored in: %s\n", ui.InfoStyle.Render(worktreesDir))
	fmt.Println()
	
	// Current repository
	currentDir, _ := os.Getwd()
	currentRepo := filepath.Base(currentDir)
	fmt.Println(ui.SuccessStyle.Render("Current Repository:"))
	fmt.Printf("  Name: %s\n", ui.InfoStyle.Render(currentRepo))
	fmt.Printf("  Path: %s\n", ui.InfoStyle.Render(currentDir))
	fmt.Println()
	
	// Statistics
	fmt.Println(ui.SuccessStyle.Render("Statistics:"))
	
	// Count total worktrees
	totalWorktrees := 0
	repoCount := 0
	
	if entries, err := os.ReadDir(worktreesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				repoCount++
				repoDir := filepath.Join(worktreesDir, entry.Name())
				if sessions, err := os.ReadDir(repoDir); err == nil {
					totalWorktrees += len(sessions)
				}
			}
		}
	}
	
	fmt.Printf("  Total repositories: %s\n", ui.InfoStyle.Render(fmt.Sprintf("%d", repoCount)))
	fmt.Printf("  Total worktrees: %s\n", ui.InfoStyle.Render(fmt.Sprintf("%d", totalWorktrees)))
	fmt.Println()
	
	// Version info
	fmt.Println(ui.SuccessStyle.Render("Version:"))
	fmt.Printf("  ccswitch: %s\n", ui.InfoStyle.Render("1.0.0"))
}