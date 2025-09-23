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
	
	ui.Title("📊 ccswitch Information")
	fmt.Println()
	
	// Paths
	ui.Success("Paths:")
	ui.Infof("  Config directory: %s", ccSwitchDir)
	ui.Infof("  Worktrees stored in: %s", worktreesDir)
	fmt.Println()
	
	// Current repository
	currentDir, _ := os.Getwd()
	currentRepo := filepath.Base(currentDir)
	ui.Success("Current Repository:")
	ui.Infof("  Name: %s", currentRepo)
	ui.Infof("  Path: %s", currentDir)
	fmt.Println()
	
	// Statistics
	ui.Success("Statistics:")
	
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
	
	ui.Infof("  Total repositories: %d", repoCount)
	ui.Infof("  Total worktrees: %d", totalWorktrees)
	fmt.Println()
	
	// Version info
	ui.Success("Version:")
	ui.Infof("  ccswitch: %s", "1.0.0")
}