package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed components from the system",
		Long: `Remove installed components (skills, agents, commands) from ~/.agent-smith/.

This command removes components from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the component directory from filesystem
  3. Removing the entry from lock files

SAFETY:
  - Components are automatically unlinked before removal
  - Source directories in ~/.agent-smith/ are permanently deleted
  - Lock file entries are removed to maintain consistency`,
	}

	// Individual component uninstall commands
	uninstallSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Remove a specific skill",
		Long: `Remove a specific skill from ~/.agent-smith/skills/.

This command removes the skill from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the skill directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific skill
  agent-smith uninstall skill mcp-builder

  # Remove from a profile
  agent-smith uninstall skill mcp-builder --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("skills", args[0], profile)
		},
	}
	uninstallSkillCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallSkillCmd)

	uninstallAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Remove a specific agent",
		Long: `Remove a specific agent from ~/.agent-smith/agents/.

This command removes the agent from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the agent directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific agent
  agent-smith uninstall agent coding-assistant

  # Remove from a profile
  agent-smith uninstall agent coding-assistant --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("agents", args[0], profile)
		},
	}
	uninstallAgentCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallAgentCmd)

	uninstallCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Remove a specific command",
		Long: `Remove a specific command from ~/.agent-smith/commands/.

This command removes the command from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the command directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific command
  agent-smith uninstall command json-formatter

  # Remove from a profile
  agent-smith uninstall command json-formatter --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("commands", args[0], profile)
		},
	}
	uninstallCommandCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallCommandCmd)

	// Bulk uninstall from repository
	uninstallAllCmd := &cobra.Command{
		Use:   "all <repository-url>",
		Short: "Remove all components from a repository",
		Long: `Remove all components installed from a specific repository.

This command finds all components (skills, agents, commands) that were installed
from the specified repository and removes them from the system.

The repository URL can be in any of these formats:
  - GitHub shorthand: owner/repo
  - Full HTTPS URL: https://github.com/owner/repo
  - SSH URL: git@github.com:owner/repo.git

EXAMPLES:
  # Remove all components from a repository
  agent-smith uninstall all anthropics/skills

  # Remove without confirmation prompt
  agent-smith uninstall all https://github.com/anthropics/skills --force

SAFETY:
  - Shows a list of components before removal
  - Prompts for confirmation (unless --force flag is used)
  - Automatically unlinks components from all targets
  - Continues with remaining components if some fail`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall all <repository-url>"),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			handleUninstallAll(args[0], force)
		},
	}
	uninstallAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	uninstallCmd.AddCommand(uninstallAllCmd)

	rootCmd.AddCommand(uninstallCmd)
}
