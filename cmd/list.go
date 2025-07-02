package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		Run:   listSessions,
	}
}

func listSessions(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ Failed to get current directory"))
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get sessions
	sessions, err := manager.ListSessions()
	if err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to list sessions: %v\n"), err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println(ui.InfoStyle.Render("No active sessions"))
		return
	}

	fmt.Println(ui.TitleStyle.Render("📁 Active Sessions:"))
	fmt.Println()

	// Find the longest session name for alignment
	maxLen := 0
	for _, session := range sessions {
		if len(session.Name) > maxLen {
			maxLen = len(session.Name)
		}
	}

	// Display sessions in a simple list
	for i, session := range sessions {
		// Session number
		fmt.Printf("  %d. ", i+1)
		
		// Session name (padded for alignment)
		fmt.Printf("%-*s  ", maxLen, ui.SuccessStyle.Render(session.Name))
		
		// Branch info
		fmt.Printf("%s  ", ui.InfoStyle.Render(session.Branch))
		
		// Path (relative if possible)
		relPath, err := filepath.Rel(currentDir, session.Path)
		if err != nil {
			relPath = session.Path
		}
		fmt.Printf("%s\n", ui.InfoStyle.Render(relPath))
	}
	
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("Use 'ccswitch switch <name>' to switch to a session"))
}