package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage installed repositories",
		Long:  `View and manage repositories installed via 'agent-smith install all'.`,
	}

	repoListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed repositories",
		Long: `List all repositories installed via 'agent-smith install all'.

Shows each repo URL, component count, and which is currently active.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesList([]string{}, false, "repo")
		},
	}

	repoCmd.AddCommand(repoListCmd)
	rootCmd.AddCommand(repoCmd)
}
