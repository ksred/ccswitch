package cmd

import (
	"fmt"

	"github.com/ksred/ccswitch/internal/version"
	"github.com/spf13/cobra"
)

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Show detailed version information including build details and runtime information.",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			fmt.Println(info.String())
		},
	}

	return cmd
}