package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup [session-name]",
		Short: "Remove worktree and optionally delete branch",
		Long: `Remove a worktree session and optionally delete its associated branch.

Without arguments: Shows an interactive list of sessions to cleanup
With session name: Removes the specified session
With --all flag: Removes all worktrees except main/master (bulk cleanup)

Examples:
  ccswitch cleanup                  # Interactive selection
  ccswitch cleanup my-feature       # Remove specific session
  ccswitch cleanup --all            # Remove all worktrees (with confirmation)`,
		Args:  cobra.MaximumNArgs(1),
		Run:   cleanupSession,
	}
	
	cmd.Flags().Bool("all", false, "Remove ALL worktrees except main/master (bulk cleanup)")
	
	return cmd
}

func cleanupSession(cmd *cobra.Command, args []string) {
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
		fmt.Println(ui.InfoStyle.Render("No active sessions to cleanup"))
		return
	}

	// Check if --all flag is set
	cleanupAll, _ := cmd.Flags().GetBool("all")
	
	if cleanupAll {
		cleanupAllSessions(manager, sessions)
		return
	}

	var sessionName string
	if len(args) > 0 {
		sessionName = args[0]
	} else {
		// Show numbered list for selection
		fmt.Println(ui.TitleStyle.Render("🗑️  Select session to cleanup:"))
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

	// Ask about branch deletion
	fmt.Printf("Delete branch %s? (y/N): ", targetSession.Branch)
	scanner := bufio.NewScanner(os.Stdin)
	deleteBranch := false
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		deleteBranch = true
	}

	// Remove the session
	if err := manager.RemoveSession(targetSession.Path, deleteBranch, targetSession.Branch); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to cleanup session: %v\n"), err)
		return
	}

	fmt.Printf(ui.SuccessStyle.Render("✓ Cleaned up session: %s\n"), sessionName)
}

func cleanupAllSessions(manager *session.Manager, sessions []git.SessionInfo) {
	// Filter out the main session and any session on main/master branch
	var worktreeSessions []git.SessionInfo
	for _, s := range sessions {
		// Skip the main session (primary repository) and any worktree on main/master branch
		if s.Name != "main" && s.Branch != "main" && s.Branch != "master" {
			worktreeSessions = append(worktreeSessions, s)
		}
	}

	if len(worktreeSessions) == 0 {
		fmt.Println(ui.InfoStyle.Render("No worktree sessions to cleanup"))
		return
	}

	// Show what will be deleted
	fmt.Println(ui.TitleStyle.Render("⚠️  You are about to remove the following worktrees:"))
	fmt.Println()
	for _, session := range worktreeSessions {
		fmt.Printf("  • %s (%s)\n", session.Name, ui.InfoStyle.Render(session.Branch))
	}
	fmt.Println()
	
	// Confirm deletion
	fmt.Print("Press Enter to continue or Ctrl+C to cancel...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	
	// Ask about branch deletion
	fmt.Println()
	fmt.Print("Delete associated branches as well? (y/N): ")
	deleteBranches := false
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		deleteBranches = true
	}
	
	fmt.Println()
	
	// Remove each session
	successCount := 0
	for _, session := range worktreeSessions {
		if err := manager.RemoveSession(session.Path, deleteBranches, session.Branch); err != nil {
			fmt.Printf(ui.ErrorStyle.Render("✗ Failed to remove %s: %v\n"), session.Name, err)
		} else {
			fmt.Printf(ui.SuccessStyle.Render("✓ Successfully removed: %s\n"), session.Name)
			successCount++
		}
	}
	
	// Summary
	fmt.Println()
	if successCount == len(worktreeSessions) {
		fmt.Printf(ui.SuccessStyle.Render("✅ All %d worktrees removed successfully!\n"), successCount)
	} else {
		fmt.Printf(ui.InfoStyle.Render("Removed %d out of %d worktrees\n"), successCount, len(worktreeSessions))
	}
	
	// Switch to main/master branch
	switchToMainBranch()
}

func switchToMainBranch() {
	// Try to switch to main first, then master if main doesn't exist
	branches := []string{"main", "master"}
	
	for _, branch := range branches {
		cmd := exec.Command("git", "checkout", branch)
		_, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf(ui.SuccessStyle.Render("✓ Switched to %s branch\n"), branch)
			return
		}
	}
	
	// If we couldn't switch to main or master, just inform the user
	fmt.Println(ui.InfoStyle.Render("ℹ Could not switch to main/master branch"))
}