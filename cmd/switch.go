package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [list|<number>]",
		Short: "Switch to an existing session or list all sessions",
		Long: `Switch to an existing session interactively or by number.

Usage:
  ccswitch switch         # Interactive session selector
  ccswitch switch list    # List all sessions with numbers
  ccswitch switch <num>   # Switch to session by number`,
		Run: switchSession,
	}
	
	return cmd
}

func switchSession(cmd *cobra.Command, args []string) {
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

	// Check if we should list sessions
	if len(args) > 0 && args[0] == "list" {
		listSessionsWithNumbers(sessions, currentDir)
		return
	}

	// Check if a number was provided
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err == nil && num > 0 && num <= len(sessions) {
			selected := &sessions[num-1]
			performSwitch(selected)
			return
		}
		fmt.Printf(ui.ErrorStyle.Render("✗ Invalid session number: %s\n"), args[0])
		fmt.Printf(ui.InfoStyle.Render("Use 'ccswitch switch list' to see available sessions\n"))
		return
	}

	// Use interactive selector
	selector := ui.NewSessionSelector(sessions)
	p := tea.NewProgram(selector)
	
	if _, err := p.Run(); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to run selector: %v\n"), err)
		return
	}
	
	if selector.IsQuit() {
		return
	}
	
	selected := selector.GetSelected()
	if selected == nil {
		return
	}

	performSwitch(selected)
}

func listSessionsWithNumbers(sessions []git.SessionInfo, currentDir string) {
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
	fmt.Println(ui.InfoStyle.Render("Use 'ccswitch switch <number>' to switch to a session"))
}

func performSwitch(selected *git.SessionInfo) {
	// Output success message with consistent formatting
	fmt.Printf("%s %s\n", ui.SuccessStyle.Render("✓ Switched to session:"), selected.Name)
	fmt.Printf("Branch: %s\n", selected.Branch)
	fmt.Printf("Location: %s\n", selected.Path)

	// Output the cd command for shell evaluation
	fmt.Printf("\ncd %s\n", selected.Path)
	
	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		fmt.Println(ui.InfoStyle.Render("💡 Note: Shell integration is not active."))
		fmt.Println(utils.GetShellIntegrationInstructions())
	}
}