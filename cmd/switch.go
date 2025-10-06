package cmd

import (
	"fmt"
	"os"

	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <session>",
		Short: "Switch to a specific session",
		Long: `Switch to a specific session by name.

The session name can be a partial match or the full name.
If multiple sessions match, the first one is selected.`,
		Args: cobra.ExactArgs(1),
		Run:  switchSession,
	}
}

func switchSession(cmd *cobra.Command, args []string) {
	sessionName := args[0]

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get sessions
	sessions, err := manager.ListSessions()
	if err != nil {
		ui.Errorf("✗ Failed to list sessions: %v", err)
		return
	}

	if len(sessions) == 0 {
		ui.Info("No active sessions")
		return
	}

	// Find the session
	var selected *git.SessionInfo
	for _, s := range sessions {
		if s.Name == sessionName || s.Branch == sessionName {
			selected = &s
			break
		}
	}

	if selected == nil {
		ui.Errorf("✗ Session '%s' not found", sessionName)
		ui.Info("Available sessions:")
		for _, s := range sessions {
			fmt.Printf("  %s (%s)\n", s.Name, s.Branch)
		}
		return
	}

	// Output success message with consistent formatting
	ui.Successf("✓ Switched to session: %s", selected.Name)
	fmt.Printf("Branch: %s\n", selected.Branch)
	fmt.Printf("Location: %s\n", selected.Path)

	// Output the cd command for shell evaluation
	fmt.Printf("\ncd %s\n", selected.Path)

	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		ui.Info("💡 Note: Shell integration is not active.")
		fmt.Println(utils.GetShellIntegrationInstructions())
	}
}
