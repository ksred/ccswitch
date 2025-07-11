package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and switch to sessions interactively",
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