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
context-switching overhead of stashing changes or switching branches.`,
		Run: createSession,
	}

	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newCleanupCmd())
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newPRCmd())
	rootCmd.AddCommand(newShellInitCmd())

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}