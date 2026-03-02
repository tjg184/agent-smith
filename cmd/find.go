package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// Create 'find' parent command with subcommands
	findCmd := &cobra.Command{
		Use:   "find",
		Short: "Search for components in remote registries",
		Long: `Search for skills across the skills.sh registry.

This command queries the skills.sh API to discover components you can install.
Results are ranked by popularity (install count).

EXAMPLES:
  # Search for PRD-related skills
  agent-smith find skill prd

  # Search for TypeScript skills
  agent-smith find skill typescript

  # Limit results
  agent-smith find skill react --limit 10

  # Get JSON output for scripting
  agent-smith find skill api --json`,
	}

	findSkillCmd := &cobra.Command{
		Use:   "skill <query>",
		Short: "Search for skills by keyword",
		Long: `Search the skills.sh registry for skills matching your query.

The query must be at least 2 characters and can include skill names,
topics, or keywords. Results show installation instructions using
agent-smith commands.

EXAMPLES:
  # Search for PRD-related skills
  agent-smith find skill prd

  # Search for TypeScript skills
  agent-smith find skill typescript

  # Search for React skills with custom limit
  agent-smith find skill react --limit 5

  # Get machine-readable JSON output
  agent-smith find skill api --json`,
		Args: exactArgsWithHelp(1, "agent-smith find skill <query>"),
		Run: func(cmd *cobra.Command, args []string) {
			limit, _ := cmd.Flags().GetInt("limit")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			handleFindSkill(args[0], limit, jsonOutput)
		},
	}
	findSkillCmd.Flags().IntP("limit", "l", 20, "Maximum number of results to display")
	findSkillCmd.Flags().Bool("json", false, "Output results as JSON for scripting")
	findCmd.AddCommand(findSkillCmd)

	rootCmd.AddCommand(findCmd)
}
