package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksred/ccswitch/internal/errors"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkout <branch>",
		Short: "Checkout an existing branch into a new worktree",
		Args:  cobra.ExactArgs(1),
		Run:   checkoutSession,
	}
}

func checkoutSession(cmd *cobra.Command, args []string) {
	branchName := strings.TrimSpace(args[0])

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Checkout the session
	if err := manager.CheckoutSession(branchName); err != nil {
		ui.Errorf("✗ %s", err)

		// Provide helpful tips based on error
		hint := errors.ErrorHint(err)
		if hint != "" {
			ui.Infof("  Tip: %s", hint)
		}

		return
	}

	// Success!
	sessionName := utils.Slugify(branchName)
	repoName := filepath.Base(currentDir)

	// Get the full worktree path
	homeDir, _ := os.UserHomeDir()
	worktreePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName, sessionName)

	ui.Successf("✓ Checked out session: %s", sessionName)
	ui.Infof("Branch: %s", branchName)
	ui.Infof("Location: ~/.ccswitch/worktrees/%s/%s", repoName, sessionName)

	// Output the cd command for the shell wrapper to execute on a separate line
	fmt.Printf("\ncd %s\n", worktreePath)

	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		ui.Infof("💡 Note: Shell integration is not active.")
		ui.Info(utils.GetShellIntegrationInstructions())
	}
}