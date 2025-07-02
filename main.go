package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ccsplit",
		Short: "Manage Claude Code sessions across git worktrees",
		Run:   createSession,
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		Run:   listSessions,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "cleanup [session-name]",
		Short: "Remove worktree and optionally delete branch",
		Args:  cobra.MaximumNArgs(1),
		Run:   cleanupSession,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "switch [session-name]",
		Short: "Switch to an existing session",
		Args:  cobra.MaximumNArgs(1),
		Run:   switchSession,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show ccswitch configuration and paths",
		Run:   showInfo,
	})

	rootCmd.Execute()
}

func createSession(cmd *cobra.Command, args []string) {
	// Check for uncommitted changes
	if output, err := runCmdWithOutput("git", "status", "--porcelain"); err == nil && output != "" {
		fmt.Println(errorStyle.Render("✗ You have uncommitted changes. Please commit or stash them first."))
		fmt.Println(infoStyle.Render("  Tip: Use 'git stash' to temporarily save changes"))
		return
	}

	fmt.Print(titleStyle.Render("🚀 What are you working on? "))
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	
	description := strings.TrimSpace(scanner.Text())
	if description == "" {
		fmt.Println(errorStyle.Render("✗ Description cannot be empty"))
		return
	}

	branchName := "feature/" + slugify(description)
	
	// Get the repository name for organization
	repoName := filepath.Base(getCurrentDir())
	
	// Create worktree in user's home directory: ~/.ccswitch/worktrees/repo-name/session-name
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf(errorStyle.Render("✗ Failed to get home directory: %v\n"), err)
		return
	}
	
	sessionName := slugify(description)
	worktreeBasePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName)
	worktreePath := filepath.Join(worktreeBasePath, sessionName)

	// Check if worktree directory already exists
	absWorktreePath, _ := filepath.Abs(worktreePath)
	if _, err := os.Stat(absWorktreePath); err == nil {
		fmt.Printf(errorStyle.Render("✗ Directory already exists: %s\n"), absWorktreePath)
		fmt.Println(infoStyle.Render("  Tip: Use a different description or remove the existing directory"))
		return
	}

	// Check if we're already on the branch we want to create
	currentBranch, _ := runCmdWithOutput("git", "branch", "--show-current")
	currentBranch = strings.TrimSpace(currentBranch)
	
	if currentBranch == branchName {
		fmt.Printf(errorStyle.Render("✗ You're already on branch: %s\n"), branchName)
		fmt.Println(infoStyle.Render("  Tip: Switch to main/master branch first, or use a different description"))
		return
	}

	// Check if branch already exists  
	output, err := runCmdWithOutput("git", "rev-parse", "--verify", "refs/heads/"+branchName)
	if err == nil && strings.TrimSpace(output) != "" {
		fmt.Printf(errorStyle.Render("✗ Branch already exists: %s\n"), branchName)
		
		// Check if it has a worktree
		worktrees := parseWorktrees(getWorktreeOutput())
		var existingPath string
		for _, wt := range worktrees {
			if wt.branch == branchName {
				existingPath = wt.path
				break
			}
		}
		
		if existingPath != "" {
			fmt.Printf(infoStyle.Render("  Already checked out at: %s\n"), existingPath)
			
			// If it's the current directory, suggest switching branches first
			currentDir := getCurrentDir()
			if existingPath == currentDir {
				fmt.Println(infoStyle.Render("  Tip: You're in this branch. Switch to main/master first"))
			} else {
				fmt.Println(infoStyle.Render("  Tip: Use 'ccswitch switch' to go to this session"))
			}
		} else {
			fmt.Println(infoStyle.Render("  Tip: Use 'git branch -D " + branchName + "' to delete it first"))
		}
		return
	}

	// Ensure the worktree base directory exists
	absWorktreeBasePath, _ := filepath.Abs(worktreeBasePath)
	if err := os.MkdirAll(absWorktreeBasePath, 0755); err != nil {
		fmt.Printf(errorStyle.Render("✗ Failed to create worktree directory: %v\n"), err)
		return
	}

	// Create branch without checking it out
	if output, err := runCmdWithOutput("git", "branch", branchName); err != nil {
		fmt.Printf(errorStyle.Render("✗ Failed to create branch: %v\n"), err)
		if output != "" {
			fmt.Printf(infoStyle.Render("  Git output: %s\n"), strings.TrimSpace(output))
			
			// Check if it's because the branch already exists
			if strings.Contains(output, "already exists") {
				fmt.Println(infoStyle.Render("  Tip: The branch was created in a previous attempt"))
				fmt.Printf(infoStyle.Render("  Run: git branch -D %s\n"), branchName)
			}
		}
		return
	}

	// Create worktree (this will check out the branch in the worktree)
	if output, err := runCmdWithOutput("git", "worktree", "add", worktreePath, branchName); err != nil {
		fmt.Printf(errorStyle.Render("✗ Failed to create worktree: %v\n"), err)
		if output != "" {
			fmt.Printf(infoStyle.Render("  Git output: %s\n"), strings.TrimSpace(output))
		}
		// Try to clean up the branch we just created
		runCmd("git", "branch", "-d", branchName)
		return
	}

	fmt.Printf(successStyle.Render("✓ Created session: %s\n"), sessionName)
	fmt.Printf(infoStyle.Render("  Branch: %s\n"), branchName)
	fmt.Printf(infoStyle.Render("  Location: ~/.ccswitch/worktrees/%s/%s\n"), repoName, sessionName)
	
	// Output the cd command for shell evaluation
	fmt.Printf("\n# Run this to enter the session:\n")
	fmt.Printf("cd %s\n", worktreePath)
}

func listSessions(cmd *cobra.Command, args []string) {
	sessions := getActiveSessions()
	if len(sessions) == 0 {
		fmt.Println(infoStyle.Render("No active sessions"))
		// Debug output
		if os.Getenv("CCSWITCH_DEBUG") == "1" {
			fmt.Println("\nDebug: Worktree output:")
			fmt.Println(getWorktreeOutput())
		}
		return
	}

	fmt.Println(titleStyle.Render("📁 Active Sessions:"))
	fmt.Println()
	
	// Find the longest session name for alignment
	maxLen := 0
	for _, session := range sessions {
		if len(session.sessionName) > maxLen {
			maxLen = len(session.sessionName)
		}
	}
	
	// Display sessions in a simple list
	for i, session := range sessions {
		// Session number
		fmt.Printf("  %d. ", i+1)
		
		// Session name (padded for alignment)
		fmt.Printf("%-*s  ", maxLen, successStyle.Render(session.sessionName))
		
		// Branch info
		fmt.Printf("%s  ", infoStyle.Render(session.branch))
		
		// Path (relative if possible)
		relPath, err := filepath.Rel(getCurrentDir(), session.path)
		if err != nil {
			relPath = session.path
		}
		fmt.Printf("%s\n", infoStyle.Render(relPath))
	}
	
	fmt.Println()
	fmt.Println(infoStyle.Render("Use 'ccswitch switch <name>' to switch to a session"))
}

func cleanupSession(cmd *cobra.Command, args []string) {
	sessions := getActiveSessions()
	if len(sessions) == 0 {
		fmt.Println(infoStyle.Render("No active sessions to cleanup"))
		return
	}

	var sessionName string
	if len(args) > 0 {
		sessionName = args[0]
	} else {
		// Show numbered list for selection
		fmt.Println(titleStyle.Render("🗑️  Select session to cleanup:"))
		fmt.Println()
		
		for i, session := range sessions {
			fmt.Printf("  %d. %s (%s)\n", i+1, session.sessionName, infoStyle.Render(session.branch))
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
			fmt.Println(errorStyle.Render("✗ Invalid selection"))
			return
		}
		
		sessionName = sessions[choice-1].sessionName
	}

	// Find the session
	var session *worktree
	for _, s := range sessions {
		if s.sessionName == sessionName {
			session = &s
			break
		}
	}

	if session == nil {
		fmt.Printf(errorStyle.Render("✗ Session not found: %s\n"), sessionName)
		return
	}

	// Remove worktree
	if err := runCmd("git", "worktree", "remove", session.path); err != nil {
		fmt.Printf(errorStyle.Render("✗ Failed to remove worktree: %v\n"), err)
		return
	}

	// Ask about branch deletion
	fmt.Printf("Delete branch %s? (y/N): ", session.branch)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		runCmd("git", "branch", "-D", session.branch)
		fmt.Printf(successStyle.Render("✓ Removed session and branch: %s\n"), sessionName)
	} else {
		fmt.Printf(successStyle.Render("✓ Removed session: %s\n"), sessionName)
	}
}

func slugify(text string) string {
	reg := regexp.MustCompile(`[^\w\s-]`)
	text = reg.ReplaceAllString(text, "")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, "-")
	return strings.ToLower(strings.Trim(text, "-"))
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil
	return cmd.Run()
}

func runCmdWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type worktree struct {
	path        string
	branch      string
	sessionName string
}


func getWorktreeOutput() string {
	output, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func getActiveSessions() []worktree {
	output := getWorktreeOutput()
	if output == "" {
		return nil
	}

	worktrees := parseWorktrees(output)
	if len(worktrees) == 0 {
		return nil
	}
	
	var sessions []worktree
	
	// Debug output
	if os.Getenv("CCSWITCH_DEBUG") == "1" {
		fmt.Println("Debug: All worktrees:")
		for _, wt := range worktrees {
			fmt.Printf("  - Path: %s, Branch: %s\n", wt.path, wt.branch)
		}
	}
	
	// The main worktree is typically:
	// 1. The one with branch "main" or "master"
	// 2. The first worktree if there's only one
	// 3. The one that contains the .git directory (not .git file)
	
	// First, try to find main/master branch
	var mainWorktreePath string
	for _, wt := range worktrees {
		if wt.branch == "main" || wt.branch == "master" {
			mainWorktreePath = wt.path
			break
		}
	}
	
	// If no main/master found and only one worktree, that's the main one
	if mainWorktreePath == "" && len(worktrees) == 1 {
		mainWorktreePath = worktrees[0].path
	}
	
	// If still not found, check for .git directory
	if mainWorktreePath == "" {
		for _, wt := range worktrees {
			gitPath := filepath.Join(wt.path, ".git")
			if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
				mainWorktreePath = wt.path
				break
			}
		}
	}
	
	if os.Getenv("CCSWITCH_DEBUG") == "1" {
		fmt.Printf("Debug: Main worktree identified as: %s\n", mainWorktreePath)
	}
	
	// Add all worktrees except the main one
	for _, wt := range worktrees {
		if mainWorktreePath == "" || wt.path != mainWorktreePath {
			// Extract session name from path
			sessionName := filepath.Base(wt.path)
			wt.sessionName = sessionName
			sessions = append(sessions, wt)
			
			if os.Getenv("CCSWITCH_DEBUG") == "1" {
				fmt.Printf("Debug: Added session: %s\n", sessionName)
			}
		}
	}
	
	return sessions
}

func parseWorktrees(output string) []worktree {
	var worktrees []worktree
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	var current worktree
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.path != "" {
				worktrees = append(worktrees, current)
			}
			current = worktree{path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch ") {
			current.branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	if current.path != "" {
		worktrees = append(worktrees, current)
	}
	
	return worktrees
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func showInfo(cmd *cobra.Command, args []string) {
	homeDir, _ := os.UserHomeDir()
	ccSwitchDir := filepath.Join(homeDir, ".ccswitch")
	worktreesDir := filepath.Join(ccSwitchDir, "worktrees")
	
	fmt.Println(titleStyle.Render("📊 ccswitch Information"))
	fmt.Println()
	
	// Paths
	fmt.Println(successStyle.Render("Paths:"))
	fmt.Printf("  Config directory: %s\n", infoStyle.Render(ccSwitchDir))
	fmt.Printf("  Worktrees stored in: %s\n", infoStyle.Render(worktreesDir))
	fmt.Println()
	
	// Current repository
	currentRepo := filepath.Base(getCurrentDir())
	fmt.Println(successStyle.Render("Current Repository:"))
	fmt.Printf("  Name: %s\n", infoStyle.Render(currentRepo))
	fmt.Printf("  Path: %s\n", infoStyle.Render(getCurrentDir()))
	fmt.Println()
	
	// Statistics
	fmt.Println(successStyle.Render("Statistics:"))
	
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
	
	fmt.Printf("  Total repositories: %s\n", infoStyle.Render(fmt.Sprintf("%d", repoCount)))
	fmt.Printf("  Total sessions: %s\n", infoStyle.Render(fmt.Sprintf("%d", totalWorktrees)))
	
	// Disk usage
	var totalSize int64
	filepath.Walk(worktreesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	
	fmt.Printf("  Disk usage: %s\n", infoStyle.Render(formatBytes(totalSize)))
	fmt.Println()
	
	fmt.Println(infoStyle.Render("Tip: Use 'ccswitch list' to see active sessions for this repository"))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func switchSession(cmd *cobra.Command, args []string) {
	sessions := getActiveSessions()
	if len(sessions) == 0 {
		fmt.Println(infoStyle.Render("No active sessions to switch to"))
		return
	}

	var sessionName string
	if len(args) > 0 {
		sessionName = args[0]
	} else {
		// Show numbered list for selection
		fmt.Println(titleStyle.Render("🔀 Select session to switch to:"))
		fmt.Println()
		
		for i, session := range sessions {
			fmt.Printf("  %d. %s (%s)\n", i+1, session.sessionName, infoStyle.Render(session.branch))
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
			fmt.Println(errorStyle.Render("✗ Invalid selection"))
			return
		}
		
		sessionName = sessions[choice-1].sessionName
	}

	// Find the session
	var session *worktree
	for _, s := range sessions {
		if s.sessionName == sessionName {
			session = &s
			break
		}
	}

	if session == nil {
		fmt.Printf(errorStyle.Render("✗ Session not found: %s\n"), sessionName)
		return
	}

	// Check if the worktree still exists
	if _, err := os.Stat(session.path); os.IsNotExist(err) {
		fmt.Printf(errorStyle.Render("✗ Session directory not found: %s\n"), session.path)
		fmt.Println(infoStyle.Render("  Tip: Run 'ccswitch cleanup' to remove stale sessions"))
		return
	}

	fmt.Printf(successStyle.Render("✓ Switching to session: %s\n"), sessionName)
	fmt.Printf(infoStyle.Render("  Branch: %s\n"), session.branch)
	fmt.Printf(infoStyle.Render("  Path: %s\n"), session.path)
	
	// Output the cd command for shell evaluation
	fmt.Printf("\n# Run this to enter the session:\n")
	fmt.Printf("cd %s\n", session.path)
}
