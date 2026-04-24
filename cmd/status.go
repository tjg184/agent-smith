package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long: `Display the current configuration status including:
  - Active profile (if any)
  - Detected targets (OpenCode, Claude Code, etc.)
  - Component counts in ~/.agent-smith/
  - Quick summary of system state

This provides a dashboard view of your agent-smith installation.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleStatus()
		},
	}
	rootCmd.AddCommand(statusCmd)
}
