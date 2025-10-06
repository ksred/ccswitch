package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksred/ccswitch/internal/config"
	"github.com/ksred/ccswitch/internal/errors"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		Run:   createSession,
	}
}

func createSession(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get description from user
	fmt.Print(ui.TitleStyle.Render("🚀 What are you working on? "))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	description := strings.TrimSpace(scanner.Text())
	if description == "" {
		ui.Error("✗ Description cannot be empty")
		return
	}

	// Create the session
	if err := manager.CreateSession(description); err != nil {
		ui.Errorf("✗ %s", err)

		// Provide helpful tips based on error
		hint := errors.ErrorHint(err)
		if hint != "" {
			ui.Infof("  Tip: %s", hint)
		}

		// Special handling for branch exists error
		if errors.IsBranchExists(err) {
			cfg, _ := config.Load()
			branchName := cfg.Branch.Prefix + utils.Slugify(description)
			ui.Infof("  Branch: %s", branchName)
		}
		return
	}

	// Success!
	sessionName := utils.Slugify(description)
	cfg, _ := config.Load()
	branchName := cfg.Branch.Prefix + sessionName
	repoName := filepath.Base(currentDir)

	// Get the full worktree path
	homeDir, _ := os.UserHomeDir()
	worktreePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName, sessionName)

	ui.Successf("✓ Created session: %s", sessionName)
	ui.Infof("Branch: %s", branchName)
	ui.Infof("Location: ~/.ccswitch/worktrees/%s/%s", repoName, sessionName)

	// Output the cd command for the shell wrapper to execute on a separate line
	fmt.Printf("\ncd %s\n", worktreePath)

	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		ui.Info("💡 Note: Shell integration is not active.")
		ui.Info(utils.GetShellIntegrationInstructions())
	}
}
