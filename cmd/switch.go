package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [session-name]",
		Short: "Switch to an existing session",
		Args:  cobra.MaximumNArgs(1),
		Run:   switchSession,
	}
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

	var sessionName string
	if len(args) > 0 {
		sessionName = args[0]
	} else {
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
		
		sessionName = selected.Name
	}

	// Find the session
	var targetSession *git.SessionInfo
	for _, s := range sessions {
		if s.Name == sessionName {
			s := s // Create a copy to take address of
			targetSession = &s
			break
		}
	}

	if targetSession == nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Session not found: %s\n"), sessionName)
		return
	}

	// Output the cd command for shell evaluation
	fmt.Printf("cd %s\n", targetSession.Path)
	
	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		fmt.Println(ui.InfoStyle.Render("💡 Note: Shell integration is not active."))
		fmt.Println(ui.InfoStyle.Render(utils.GetShellIntegrationInstructions()))
	}
}