package cmd

import (
	"github.com/spf13/cobra"
)

func makeUnlinkRun(componentType string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		targetFilter, _ := cmd.Flags().GetString("target")
		profile, _ := cmd.Flags().GetString("profile")
		handleUnlinkWithProfile(componentType, args[0], targetFilter, profile)
	}
}

func makeUnlinkBulkRun(componentType string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		targetFilter, _ := cmd.Flags().GetString("target")
		profile, _ := cmd.Flags().GetString("profile")
		if len(args) == 1 {
			handleUnlinkWithProfile(componentType, args[0], targetFilter, profile)
		} else {
			force, _ := cmd.Flags().GetBool("force")
			handleUnlinkTypeWithProfile(componentType, targetFilter, force, profile)
		}
	}
}

func init() {
	unlinkCmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove linked components from targets",
		Long: `Remove linked components (skills, agents, commands) from detected targets.

This command provides a modern interface for unlinking downloaded AI components
from supported targets (OpenCode, Claude Code, etc.).

SAFETY:
  - Symlinks are removed immediately
  - Copied directories require confirmation before deletion
  - Source files in ~/.agent-smith/ are never touched
  - Bulk operations (skills, agents, commands, all) prompt for confirmation unless --force is used`,
	}

	unlinkSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Unlink a specific skill from targets",
		Long: `Unlink a specific skill from detected targets.

This command removes the linked skill from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/skills/ are never touched.

EXAMPLES:
  # Unlink a specific skill from all targets
  agent-smith unlink skill mcp-builder

  # Unlink a specific skill from OpenCode only
  agent-smith unlink skill mcp-builder --target opencode

  # Unlink a skill from a specific profile
  agent-smith unlink skill mcp-builder --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink skill <name>"),
		Run:  makeUnlinkRun("skills"),
	}
	addUnlinkFlags(unlinkSkillCmd)
	unlinkCmd.AddCommand(unlinkSkillCmd)

	unlinkAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Unlink a specific agent from targets",
		Long: `Unlink a specific agent from detected targets.

This command removes the linked agent from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/agents/ are never touched.

EXAMPLES:
  # Unlink a specific agent from all targets
  agent-smith unlink agent coding-assistant

  # Unlink a specific agent from OpenCode only
  agent-smith unlink agent coding-assistant --target opencode

  # Unlink an agent from a specific profile
  agent-smith unlink agent coding-assistant --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink agent <name>"),
		Run:  makeUnlinkRun("agents"),
	}
	addUnlinkFlags(unlinkAgentCmd)
	unlinkCmd.AddCommand(unlinkAgentCmd)

	unlinkCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Unlink a specific command from targets",
		Long: `Unlink a specific command from detected targets.

This command removes the linked command from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/commands/ are never touched.

EXAMPLES:
  # Unlink a specific command from all targets
  agent-smith unlink command json-formatter

  # Unlink a specific command from OpenCode only
  agent-smith unlink command json-formatter --target opencode

  # Unlink a command from a specific profile
  agent-smith unlink command json-formatter --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink command <name>"),
		Run:  makeUnlinkRun("commands"),
	}
	addUnlinkFlags(unlinkCommandCmd)
	unlinkCmd.AddCommand(unlinkCommandCmd)

	unlinkSkillsCmd := &cobra.Command{
		Use:   "skills [name]",
		Short: "Unlink all skills from targets, or a specific skill if name provided",
		Long: `Unlink all skills from detected targets, or a specific skill if name is provided.

This command removes all linked skills from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/skills/ are never touched.

For backward compatibility, you can also provide a skill name to unlink just
that specific skill (equivalent to 'unlink skill <name>').

EXAMPLES:
  # Unlink all skills with confirmation
  agent-smith unlink skills

  # Unlink all skills without confirmation
  agent-smith unlink skills --force

  # Unlink all skills from OpenCode only
  agent-smith unlink skills --target opencode

  # Unlink all skills from a specific profile
  agent-smith unlink skills --profile work

  # Unlink a specific skill (backward compatibility)
  agent-smith unlink skills mcp-builder`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink skills [name]"),
		Run:  makeUnlinkBulkRun("skills"),
	}
	addForceFlag(unlinkSkillsCmd)
	addUnlinkFlags(unlinkSkillsCmd)
	unlinkCmd.AddCommand(unlinkSkillsCmd)

	unlinkAgentsCmd := &cobra.Command{
		Use:   "agents [name]",
		Short: "Unlink all agents from targets, or a specific agent if name provided",
		Long: `Unlink all agents from detected targets, or a specific agent if name is provided.

This command removes all linked agents from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/agents/ are never touched.

For backward compatibility, you can also provide an agent name to unlink just
that specific agent (equivalent to 'unlink agent <name>').

EXAMPLES:
  # Unlink all agents with confirmation
  agent-smith unlink agents

  # Unlink all agents without confirmation
  agent-smith unlink agents --force

  # Unlink all agents from OpenCode only
  agent-smith unlink agents --target opencode

  # Unlink all agents from a specific profile
  agent-smith unlink agents --profile work

  # Unlink a specific agent (backward compatibility)
  agent-smith unlink agents coding-assistant`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink agents [name]"),
		Run:  makeUnlinkBulkRun("agents"),
	}
	addForceFlag(unlinkAgentsCmd)
	addUnlinkFlags(unlinkAgentsCmd)
	unlinkCmd.AddCommand(unlinkAgentsCmd)

	unlinkCommandsCmd := &cobra.Command{
		Use:   "commands [name]",
		Short: "Unlink all commands from targets, or a specific command if name provided",
		Long: `Unlink all commands from detected targets, or a specific command if name is provided.

This command removes all linked commands from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/commands/ are never touched.

For backward compatibility, you can also provide a command name to unlink just
that specific command (equivalent to 'unlink command <name>').

EXAMPLES:
  # Unlink all commands with confirmation
  agent-smith unlink commands

  # Unlink all commands without confirmation
  agent-smith unlink commands --force

  # Unlink all commands from OpenCode only
  agent-smith unlink commands --target opencode

  # Unlink all commands from a specific profile
  agent-smith unlink commands --profile work

  # Unlink a specific command (backward compatibility)
  agent-smith unlink commands json-formatter`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink commands [name]"),
		Run:  makeUnlinkBulkRun("commands"),
	}
	addForceFlag(unlinkCommandsCmd)
	addUnlinkFlags(unlinkCommandsCmd)
	unlinkCmd.AddCommand(unlinkCommandsCmd)

	unlinkAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Unlink all components from targets",
		Long: `Unlink all components (skills, agents, and commands) from detected targets.

By default, only components from the currently active profile are unlinked.
Use --all-profiles to unlink components from all profiles.
Use --profile to unlink from a specific profile without switching to it.

This command removes all linked components from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/ are never touched.

EXAMPLES:
  # Unlink all components from current profile with confirmation
  agent-smith unlink all

  # Unlink all components from current profile without confirmation
  agent-smith unlink all --force

  # Unlink all components from all profiles
  agent-smith unlink all --all-profiles

  # Unlink all components from a specific profile
  agent-smith unlink all --profile work

  # Unlink all components from OpenCode only
  agent-smith unlink all --target opencode`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			force, _ := cmd.Flags().GetBool("force")
			allProfiles, _ := cmd.Flags().GetBool("all-profiles")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkAllWithProfile(targetFilter, force, allProfiles, profile)
		},
	}
	addForceFlag(unlinkAllCmd)
	unlinkAllCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAllCmd.Flags().Bool("all-profiles", false, "Unlink components from all profiles (default: current profile only)")
	unlinkAllCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkAllCmd)

	rootCmd.AddCommand(unlinkCmd)
}
