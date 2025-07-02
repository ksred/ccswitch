package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
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
		// Show numbered list for selection
		fmt.Println(ui.TitleStyle.Render("📂 Select session to switch to:"))
		fmt.Println()
		
		for i, session := range sessions {
			fmt.Printf("  %d. %s (%s)\n", i+1, session.Name, ui.InfoStyle.Render(session.Branch))
		}
		
		fmt.Println()
		fmt.Print("Enter number (or q to quit): ")
		
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "q" || input == "" {
			return
		}
		
		// Parse number
		var choice int
		if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(sessions) {
			fmt.Println(ui.ErrorStyle.Render("✗ Invalid selection"))
			return
		}
		
		sessionName = sessions[choice-1].Name
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
}