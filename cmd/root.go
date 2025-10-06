package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccswitch",
		Short: "Manage development sessions across git worktrees",
		Long: `ccswitch helps you work on multiple features simultaneously without the 
context-switching overhead of stashing changes or switching branches.

Key commands:
  ccswitch                    Create a new work session
  ccswitch checkout <branch>  Checkout an existing branch into a new worktree
  ccswitch list               Show and switch between sessions
  ccswitch switch <session>   Switch to a specific session
  ccswitch cleanup            Remove a session interactively
  ccswitch cleanup --all      Remove ALL worktrees at once (bulk cleanup)
  ccswitch pr                 Create a pull request for current session`,
		Run: createSession,
	}

	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newCheckoutCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newSwitchCmd())
	rootCmd.AddCommand(newCleanupCmd())
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newPRCmd())
	rootCmd.AddCommand(newShellInitCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}
