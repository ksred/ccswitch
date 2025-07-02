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

func createSession(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(ui.ErrorStyle.Render("✗ Failed to get current directory"))
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
		fmt.Println(ui.ErrorStyle.Render("✗ Description cannot be empty"))
		return
	}

	// Create the session
	if err := manager.CreateSession(description); err != nil {
		fmt.Printf(ui.ErrorStyle.Render("✗ %s\n"), err)
		
		// Provide helpful tips based on error
		hint := errors.ErrorHint(err)
		if hint != "" {
			fmt.Printf(ui.InfoStyle.Render("  Tip: %s\n"), hint)
		}
		
		// Special handling for branch exists error
		if errors.IsBranchExists(err) {
			cfg, _ := config.Load()
			branchName := cfg.Branch.Prefix + utils.Slugify(description)
			fmt.Printf(ui.InfoStyle.Render("  Branch: %s\n"), branchName)
		}
		return
	}

	// Success!
	sessionName := utils.Slugify(description)
	cfg, _ := config.Load()
	branchName := cfg.Branch.Prefix + sessionName
	repoName := filepath.Base(currentDir)
	
	fmt.Printf(ui.SuccessStyle.Render("✓ Created session: %s\n"), sessionName)
	fmt.Printf(ui.InfoStyle.Render("  Branch: %s\n"), branchName)
	fmt.Printf(ui.InfoStyle.Render("  Location: ~/.ccswitch/worktrees/%s/%s\n"), repoName, sessionName)
	
	// Output the switch command
	fmt.Printf("\n# Run this to enter the session:\n")
	fmt.Printf("ccswitch switch %s\n", sessionName)
}