package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newPRCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pr",
		Short: "Create a pull request for the current session",
		Run:   createPullRequest,
	}
}

func createPullRequest(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ Failed to get current directory"))
		return
	}

	// Check if gh CLI is available
	if !isGitHubCLIAvailable() {
		fmt.Println(ui.ErrorStyle.Render("✗ GitHub CLI (gh) is not installed or not in PATH"))
		fmt.Println(ui.InfoStyle.Render("  Install GitHub CLI: https://cli.github.com/"))
		return
	}

	// Check if we're in a git repository
	if !git.IsGitRepository(currentDir) {
		fmt.Println(ui.ErrorStyle.Render("✗ Not in a git repository"))
		return
	}

	// Get current branch
	currentBranch, err := git.GetCurrentBranch(currentDir)
	if err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to get current branch: %v\n"), err)
		return
	}

	// Check if we're on main/master branch
	if currentBranch == "main" || currentBranch == "master" {
		fmt.Println(ui.ErrorStyle.Render("✗ Cannot create PR from main/master branch"))
		fmt.Println(ui.InfoStyle.Render("  Switch to a feature branch first using 'ccswitch list'"))
		return
	}

	// Check if we're in a ccswitch session
	manager := session.NewManager(currentDir)
	sessions, err := manager.ListSessions()
	if err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to list sessions: %v\n"), err)
		return
	}

	var currentSession *git.SessionInfo
	for _, s := range sessions {
		if s.Path == currentDir {
			s := s // Create a copy to take address of
			currentSession = &s
			break
		}
	}

	if currentSession == nil {
		fmt.Println(ui.ErrorStyle.Render("✗ Not in a ccswitch session directory"))
		fmt.Println(ui.InfoStyle.Render("  Use 'ccswitch list' to enter a session first"))
		return
	}

	fmt.Printf(ui.TitleStyle.Render("🚀 Creating pull request for session: %s\n"), currentSession.Name)
	fmt.Printf(ui.InfoStyle.Render("  Branch: %s\n"), currentBranch)

	// Check if branch has commits ahead of main
	hasCommits, err := checkBranchHasCommits(currentDir, currentBranch)
	if err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to check branch commits: %v\n"), err)
		return
	}

	if !hasCommits {
		fmt.Println(ui.ErrorStyle.Render("✗ No commits found on this branch"))
		fmt.Println(ui.InfoStyle.Render("  Make some commits before creating a PR"))
		return
	}

	// Push the branch if needed
	fmt.Println(ui.InfoStyle.Render("📤 Pushing branch to remote..."))
	if err := pushBranch(currentDir, currentBranch); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to push branch: %v\n"), err)
		return
	}

	// Create PR using gh CLI
	fmt.Println(ui.InfoStyle.Render("📝 Creating pull request..."))
	prURL, err := createPRWithGH(currentDir, currentSession.Name)
	if err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to create PR: %v\n"), err)
		return
	}

	fmt.Printf(ui.SuccessStyle.Render("✓ Pull request created successfully!\n"))
	fmt.Printf(ui.InfoStyle.Render("  URL: %s\n"), prURL)

	// Open in browser
	fmt.Println(ui.InfoStyle.Render("🌐 Opening PR in browser..."))
	if err := openInBrowser(prURL); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ Failed to open browser: %v\n"), err)
		fmt.Println(ui.InfoStyle.Render("  You can manually open the URL above"))
	}
}

func isGitHubCLIAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func checkBranchHasCommits(dir, branch string) (bool, error) {
	cmd := exec.Command("git", "rev-list", "--count", "main.."+branch)
	cmd.Dir = dir
	
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	
	count := strings.TrimSpace(string(output))
	return count != "0", nil
}

func pushBranch(dir, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

func createPRWithGH(dir, sessionName string) (string, error) {
	// Generate PR title from session name
	title := strings.ReplaceAll(sessionName, "-", " ")
	title = strings.Title(title)
	
	cmd := exec.Command("gh", "pr", "create", "--title", title, "--body", "Created from ccswitch session: "+sessionName, "--web")
	cmd.Dir = dir
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	// Extract URL from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://github.com/") {
			return line, nil
		}
	}
	
	return string(output), nil
}

func openInBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch {
	case exec.Command("which", "open").Run() == nil: // macOS
		cmd = exec.Command("open", url)
	case exec.Command("which", "xdg-open").Run() == nil: // Linux
		cmd = exec.Command("xdg-open", url)
	case exec.Command("which", "cmd").Run() == nil: // Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unable to detect platform to open browser")
	}
	
	return cmd.Run()
}