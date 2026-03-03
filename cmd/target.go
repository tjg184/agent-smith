package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	targetCmd := &cobra.Command{
		Use:   "target",
		Short: "Manage custom targets",
		Long: `Manage custom targets for linking components.

Custom targets allow you to link components to additional editors or tools
beyond the built-in OpenCode and Claude Code targets.`,
	}

	targetAddCmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Register a new custom target",
		Long: `Register a new custom target for linking components.

This command adds a custom target to your configuration file, allowing you
to link components to additional editors or tools beyond the built-in targets.

The target will use the following subdirectories (relative to the path):
  - skills/   - For skills
  - agents/   - For agents
  - commands/ - For commands

EXAMPLES:
  # Add a custom target for Cursor
  agent-smith target add cursor ~/.cursor

  # Add a custom target for VS Code
  agent-smith target add vscode ~/.vscode/agent-smith

After adding a target, you can link components to it using:
  agent-smith link all --target <name>`,
		Args: exactArgsWithHelp(2, "agent-smith target add <name> <path>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetAdd(args[0], args[1])
		},
	}
	targetCmd.AddCommand(targetAddCmd)

	targetRemoveCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Unregister a custom target",
		Long: `Unregister a custom target from your configuration.

This command removes a custom target from your configuration file. Only custom
targets can be removed - built-in targets (opencode, claudecode) cannot be removed.

EXAMPLES:
  # Remove a custom target
  agent-smith target remove cursor

  # Remove a custom target for VS Code
  agent-smith target remove vscode

Note: This only removes the target from the configuration. It does not unlink
any components that are currently linked to this target.`,
		Args: exactArgsWithHelp(1, "agent-smith target remove <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetRemove(args[0])
		},
	}
	targetCmd.AddCommand(targetRemoveCmd)

	targetListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available targets",
		Long: `List all available targets for linking components.

This command displays all targets that are configured in your system, including:
  - Built-in targets (OpenCode, Claude Code)
  - Custom targets from your configuration

For each target, it shows:
  - Target name
  - Base directory path
  - Whether the directory currently exists

EXAMPLES:
  # List all available targets
  agent-smith target list`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetList()
		},
	}
	targetCmd.AddCommand(targetListCmd)

	rootCmd.AddCommand(targetCmd)
}
