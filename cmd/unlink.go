package cmd

import (
	"github.com/spf13/cobra"
)

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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("skills", args[0], targetFilter, profile)
		},
	}
	unlinkSkillCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkSkillCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("agents", args[0], targetFilter, profile)
		},
	}
	unlinkAgentCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAgentCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("commands", args[0], targetFilter, profile)
		},
	}
	unlinkCommandCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkCommandCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific skill
			if len(args) == 1 {
				handleUnlinkWithProfile("skills", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("skills", targetFilter, force, profile)
			}
		},
	}
	unlinkSkillsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkSkillsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkSkillsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific agent
			if len(args) == 1 {
				handleUnlinkWithProfile("agents", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("agents", targetFilter, force, profile)
			}
		},
	}
	unlinkAgentsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkAgentsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAgentsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific command
			if len(args) == 1 {
				handleUnlinkWithProfile("commands", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("commands", targetFilter, force, profile)
			}
		},
	}
	unlinkCommandsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCommandsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkCommandsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
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
	unlinkAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkAllCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAllCmd.Flags().Bool("all-profiles", false, "Unlink components from all profiles (default: current profile only)")
	unlinkAllCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkAllCmd)

	rootCmd.AddCommand(unlinkCmd)
}
