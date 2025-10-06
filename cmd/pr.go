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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Check if gh CLI is available
	if !isGitHubCLIAvailable() {
		ui.Error("✗ GitHub CLI (gh) is not installed or not in PATH")
		ui.Info("  Install GitHub CLI: https://cli.github.com/")
		return
	}

	// Check if we're in a git repository
	if !git.IsGitRepository(currentDir) {
		ui.Error("✗ Not in a git repository")
		return
	}

	// Get current branch
	currentBranch, err := git.GetCurrentBranch(currentDir)
	if err != nil {
		ui.Errorf("✗ Failed to get current branch: %v", err)
		return
	}

	// Check if we're on main/master branch
	if currentBranch == "main" || currentBranch == "master" {
		ui.Error("✗ Cannot create PR from main/master branch")
		ui.Info("  Switch to a feature branch first using 'ccswitch list'")
		return
	}

	// Check if we're in a ccswitch session
	manager := session.NewManager(currentDir)
	sessions, err := manager.ListSessions()
	if err != nil {
		ui.Errorf("✗ Failed to list sessions: %v", err)
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
		ui.Error("✗ Not in a ccswitch session directory")
		ui.Info("  Use 'ccswitch list' to enter a session first")
		return
	}

	ui.Titlef("🚀 Creating pull request for session: %s", currentSession.Name)
	ui.Infof("  Branch: %s", currentBranch)

	// Check if branch has commits ahead of main
	hasCommits, err := checkBranchHasCommits(currentDir, currentBranch)
	if err != nil {
		ui.Errorf("✗ Failed to check branch commits: %v", err)
		return
	}

	if !hasCommits {
		ui.Error("✗ No commits found on this branch")
		ui.Info("  Make some commits before creating a PR")
		return
	}

	// Push the branch if needed
	ui.Info("📤 Pushing branch to remote...")
	if pushErr := pushBranch(currentDir, currentBranch); pushErr != nil {
		ui.Errorf("✗ Failed to push branch: %v", pushErr)
		return
	}

	// Create PR using gh CLI
	ui.Info("📝 Creating pull request...")
	prURL, err := createPRWithGH(currentDir, currentSession.Name)
	if err != nil {
		ui.Errorf("✗ Failed to create PR: %v", err)
		return
	}

	ui.Successf("✓ Pull request created successfully!")
	ui.Infof("  URL: %s", prURL)

	// Open in browser
	ui.Info("🌐 Opening PR in browser...")
	if err := openInBrowser(prURL); err != nil {
		ui.Errorf("✗ Failed to open browser: %v", err)
		ui.Info("  You can manually open the URL above")
	}
}

func isGitHubCLIAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func checkBranchHasCommits(dir, branch string) (bool, error) {
	cmd := exec.Command("git", "rev-list", "--count", "main.."+branch) // #nosec G204
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
	title = cases.Title(language.English).String(title)

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
