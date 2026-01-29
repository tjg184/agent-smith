package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-smith",
	Short: "Agent Smith - A CLI tool for managing AI agents, skills, and commands",
	Long: `Agent Smith is a powerful CLI tool for downloading, managing, and executing
AI agents, skills, and commands from git repositories.

It provides npm-like functionality for AI components, allowing you to:
- Download and install agents, skills, and commands
- Update and manage installed components
- Link components to supported targets (OpenCode, Claude Code, etc.)`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// isValidComponentType checks if a string is a valid component type
func isValidComponentType(componentType string) bool {
	return componentType == "skills" || componentType == "agents" || componentType == "commands"
}

// validateComponentType validates that a component type is valid and returns a helpful error if not
func validateComponentType(componentType string) error {
	if !isValidComponentType(componentType) {
		return fmt.Errorf("invalid component type '%s'\n\nValid component types:\n  - skills\n  - agents\n  - commands\n\nExample:\n  agent-smith update skills my-skill", componentType)
	}
	return nil
}

// exactArgsWithHelp returns a custom validator that provides helpful error messages
func exactArgsWithHelp(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		return nil
	}
}

// noArgsWithHelp returns a custom validator for commands that accept no arguments
func noArgsWithHelp(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("this command does not accept arguments\n\nUsage: %s\n\nRun '%s --help' for more information", cmd.Use, cmd.CommandPath())
	}
	return nil
}

// rangeArgsWithHelp returns a custom validator for commands with a range of arguments
func rangeArgsWithHelp(min, max int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > max {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		return nil
	}
}

// exactArgsWithComponentTypeValidation returns a validator that checks both argument count and component type
func exactArgsWithComponentTypeValidation(n int, componentTypeIndex int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// First check argument count
		if len(args) < n {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}

		// Then validate component type if specified
		if componentTypeIndex >= 0 && componentTypeIndex < len(args) {
			if err := validateComponentType(args[componentTypeIndex]); err != nil {
				return err
			}
		}

		return nil
	}
}

func init() {
	// Hide completion command from help output
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Create 'install' parent command with subcommands
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install components from git repositories",
		Long: `Install components (skills, agents, commands) from git repositories.

This command provides a modern interface for downloading and installing AI components
from any git repository (GitHub, GitLab, Bitbucket, or private repositories).

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo`,
	}

	// Add subcommands to 'install' command
	installSkillCmd := &cobra.Command{
		Use:   "skill <repository-url> <skill-name>",
		Short: "Download a skill from a git repository",
		Long: `Download and install a skill from a git repository to your local agents directory.

This command fetches a skill from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/skills/<skill-name>. The skill can include multiple components
and will be automatically detected if it contains a SKILL.md file.

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install skill openai/cookbook gpt-skill

  # Download using full URL
  agent-smith install skill https://github.com/example/repo my-skill

  # Download from local repository
  agent-smith install skill /path/to/local/skill local-skill

  # Install directly to a profile
  agent-smith install skill openai/cookbook gpt-skill --profile work

  # Install to custom directory for testing (isolated from ~/.agents/)
  agent-smith install skill ./my-skill test-skill --target-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install skill <repository-url> <skill-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			targetDir, _ := cmd.Flags().GetString("target-dir")
			handleAddSkill(args[0], args[1], profile, targetDir)
		},
	}
	installSkillCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
	installSkillCmd.Flags().StringP("target-dir", "t", "", "Install to a custom directory (isolated from ~/.agents/)")
	installCmd.AddCommand(installSkillCmd)

	installAgentCmd := &cobra.Command{
		Use:   "agent <repository-url> <agent-name>",
		Short: "Download an agent from a git repository",
		Long: `Download and install an AI agent from a git repository to your local agents directory.

This command fetches an agent from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/agents/<agent-name>. The agent will be automatically detected
based on path patterns and file extensions.

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install agent openai/assistant coding-agent

  # Download using full URL
  agent-smith install agent https://github.com/example/agent my-agent

  # Download from local repository
  agent-smith install agent /path/to/local/agent local-agent

  # Install directly to a profile
  agent-smith install agent openai/assistant coding-agent --profile work

  # Install to custom directory for testing (isolated from ~/.agents/)
  agent-smith install agent ./my-agent test-agent --target-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install agent <repository-url> <agent-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			targetDir, _ := cmd.Flags().GetString("target-dir")
			handleAddAgent(args[0], args[1], profile, targetDir)
		},
	}
	installAgentCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
	installAgentCmd.Flags().StringP("target-dir", "t", "", "Install to a custom directory (isolated from ~/.agents/)")
	installCmd.AddCommand(installAgentCmd)

	installCommandCmd := &cobra.Command{
		Use:   "command <repository-url> <command-name>",
		Short: "Download a command from a git repository",
		Long: `Download and install a command-line tool from a git repository to your local agents directory.

This command fetches a command from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/commands/<command-name>. The command will be automatically detected
based on path patterns and file extensions.

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install command cli-tools/formatter json-formatter

  # Download using full URL
  agent-smith install command https://github.com/example/tool my-tool

  # Download from local repository
  agent-smith install command /path/to/local/command local-cmd

  # Install directly to a profile
  agent-smith install command cli-tools/formatter json-formatter --profile work

  # Install to custom directory for testing (isolated from ~/.agents/)
  agent-smith install command ./my-command test-command --target-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install command <repository-url> <command-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			targetDir, _ := cmd.Flags().GetString("target-dir")
			handleAddCommand(args[0], args[1], profile, targetDir)
		},
	}
	installCommandCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
	installCommandCmd.Flags().StringP("target-dir", "t", "", "Install to a custom directory (isolated from ~/.agents/)")
	installCmd.AddCommand(installCommandCmd)

	installAllCmd := &cobra.Command{
		Use:   "all <repository-url>",
		Short: "Download all components from a git repository",
		Long: `Download and install all components (skills, agents, and commands) from a git repository.

This command fetches a repository and automatically detects all AI components
within it, then downloads them to their respective directories. Components are
detected based on the presence of SKILL.md files or path patterns.

EXAMPLES:
  # Download all components from GitHub using shorthand
  agent-smith install all openai/cookbook

  # Download using full URL
  agent-smith install all https://github.com/example/monorepo

  # Download from local repository
  agent-smith install all /path/to/local/repo

  # Install to a custom target directory (project-local)
  agent-smith install all openai/cookbook --target-dir ./tools

  # Install to a custom directory with tilde expansion
  agent-smith install all openai/cookbook --target-dir ~/my-project/agents`,
		Args: exactArgsWithHelp(1, "agent-smith install all <repository-url>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetDir, _ := cmd.Flags().GetString("target-dir")
			handleAddAll(args[0], targetDir)
		},
	}
	installAllCmd.Flags().StringP("target-dir", "t", "", "Install to a custom directory instead of ~/.agents/")
	installCmd.AddCommand(installAllCmd)

	rootCmd.AddCommand(installCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "update <type|all> [name]",
		Short: "Check and update a component or all components",
		Long: `Check and update a specific component or all downloaded components.

USAGE:
  agent-smith update <type> <name>  - Update a specific component
  agent-smith update all            - Update all downloaded components

EXAMPLES:
  # Update a specific skill
  agent-smith update skills mcp-builder

  # Update all components
  agent-smith update all`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing required arguments\n\nUsage: agent-smith update <type|all> [name]\n\nRun '%s --help' for more information", cmd.CommandPath())
			}
			if len(args) > 2 {
				return fmt.Errorf("too many arguments provided\n\nUsage: agent-smith update <type|all> [name]\n\nRun '%s --help' for more information", cmd.CommandPath())
			}
			// If first arg is not "all", validate it's a valid component type and require name
			if args[0] != "all" {
				if err := validateComponentType(args[0]); err != nil {
					return err
				}
				if len(args) != 2 {
					return fmt.Errorf("missing component name\n\nUsage: agent-smith update <type> <name>\n\nExample:\n  agent-smith update skills my-skill\n\nRun '%s --help' for more information", cmd.CommandPath())
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "all" {
				handleUpdateAll()
			} else {
				handleUpdate(args[0], args[1])
			}
		},
	})

	// Create 'link' parent command with subcommands
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Link components to detected targets",
		Long: `Link components (skills, agents, commands) to detected targets.

This command provides a modern interface for linking downloaded AI components
to supported targets (OpenCode, Claude Code, etc.).

PROFILE AWARENESS:
When a profile is active, link commands automatically use components from the
active profile directory instead of ~/.agents/. This allows you to control which
components are linked to your editor.

  - With active profile: Sources from ~/.agents/profiles/<profile>/
  - No active profile: Sources from ~/.agents/ (base installation)

Use 'agent-smith profile activate <name>' to activate a profile, then run
'link' commands to apply it.

FLAGS (apply to all subcommands):
  --target, -t <target>  - Specify target to link to (opencode, claudecode, or all)
  --all-targets          - Explicitly link to all detected targets (default behavior)`,
	}

	// Add subcommands to 'link' command
	// Singular commands - operate on ONE component
	linkSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Link a specific skill to detected targets",
		Long: `Link a specific skill to detected targets.

This command links a downloaded skill from ~/.agents/skills/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific skill to all detected targets (default)
  agent-smith link skill mcp-builder

  # Link a specific skill to OpenCode only
  agent-smith link skill mcp-builder --target opencode`,
		Args: exactArgsWithHelp(1, "agent-smith link skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link specific skill
			handleLink("skills", args[0], targetFilter)
		},
	}
	linkSkillCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkSkillCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkSkillCmd)

	// Plural command - operate on ALL skills
	linkSkillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Link all skills to detected targets",
		Long: `Link all skills to detected targets.

This command links all downloaded skills from ~/.agents/skills/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link all skills to all detected targets
  agent-smith link skills

  # Link all skills to Claude Code only
  agent-smith link skills --target claudecode`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link all skills
			handleLinkType("skills", targetFilter)
		},
	}
	linkSkillsCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkSkillsCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkSkillsCmd)

	linkAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Link a specific agent to detected targets",
		Long: `Link a specific agent to detected targets.

This command links a downloaded agent from ~/.agents/agents/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific agent to all detected targets (default)
  agent-smith link agent coding-assistant

  # Link a specific agent to OpenCode only
  agent-smith link agent coding-assistant --target opencode`,
		Args: exactArgsWithHelp(1, "agent-smith link agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link specific agent
			handleLink("agents", args[0], targetFilter)
		},
	}
	linkAgentCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkAgentCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkAgentCmd)

	// Plural command - operate on ALL agents
	linkAgentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "Link all agents to detected targets",
		Long: `Link all agents to detected targets.

This command links all downloaded agents from ~/.agents/agents/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link all agents to all detected targets
  agent-smith link agents

  # Link all agents to Claude Code only
  agent-smith link agents --target claudecode`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link all agents
			handleLinkType("agents", targetFilter)
		},
	}
	linkAgentsCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkAgentsCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkAgentsCmd)

	linkCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Link a specific command to detected targets",
		Long: `Link a specific command to detected targets.

This command links a downloaded command from ~/.agents/commands/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific command to all detected targets (default)
  agent-smith link command json-formatter

  # Link a specific command to OpenCode only
  agent-smith link command json-formatter --target opencode`,
		Args: exactArgsWithHelp(1, "agent-smith link command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link specific command
			handleLink("commands", args[0], targetFilter)
		},
	}
	linkCommandCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkCommandCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkCommandCmd)

	// Plural command - operate on ALL commands
	linkCommandsCmd := &cobra.Command{
		Use:   "commands",
		Short: "Link all commands to detected targets",
		Long: `Link all commands to detected targets.

This command links all downloaded commands from ~/.agents/commands/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link all commands to all detected targets
  agent-smith link commands

  # Link all commands to Claude Code only
  agent-smith link commands --target claudecode`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			// Link all commands
			handleLinkType("commands", targetFilter)
		},
	}
	linkCommandsCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkCommandsCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkCommandsCmd)

	linkAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Link all components to detected targets",
		Long: `Link all downloaded components (skills, agents, and commands) to detected targets.

This command links all components to the appropriate directories for OpenCode,
Claude Code, or other supported targets.

PROFILE AWARENESS:
  - With active profile: Links components from the active profile
  - No active profile: Links all components from ~/.agents/ (base installation)

TWO-STEP WORKFLOW:
  1. Activate a profile: agent-smith profile activate <name>
  2. Apply to editor: agent-smith link all

This gives you explicit control over when changes are applied to your editor.

EXAMPLES:
  # Link all components to all detected targets (default)
  agent-smith link all

  # Link all components to OpenCode only
  agent-smith link all --target opencode

  # Link all components to all targets explicitly
  agent-smith link all --all-targets`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			handleLinkAll(targetFilter)
		},
	}
	linkAllCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkAllCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkAllCmd)

	linkAutoCmd := &cobra.Command{
		Use:   "auto",
		Short: "Automatically detect and link components from current repository",
		Long: `Automatically detect and link components from the current repository.

This command scans the current working directory for AI components (skills, agents,
and commands) and automatically links them to detected targets. It uses pattern
detection to identify component files:
  - Skills: Files named SKILL.md
  - Agents: Files in /agents/ directories with .md extension
  - Commands: Files in /commands/ directories with .md extension

The detection also honors frontmatter metadata in markdown files, using the 'name'
field if present.

EXAMPLES:
  # Auto-detect and link all components in current repository
  agent-smith link auto

  # Typically used from within a repository containing component definitions
  cd /path/to/my-components
  agent-smith link auto`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleAutoLink()
		},
	}
	linkCmd.AddCommand(linkAutoCmd)

	linkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all linked components across all targets",
		Long: `List all components (skills, agents, and commands) currently linked to detected targets.

This command shows the status of each linked component, including whether it's
a symlink or copied directory, and whether the link is valid or broken.

EXAMPLES:
  # List all linked components
  agent-smith link list

The output shows:
  ✓ - Valid symlink
  ◆ - Copied directory
  ✗ - Broken link`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleListLinks()
		},
	}
	linkCmd.AddCommand(linkListCmd)

	linkStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show link status across all targets in a matrix view",
		Long: `Show the status of all components across all detected targets in a matrix format.

This command displays a table showing which components are linked to which targets,
making it easy to see what is installed where at a glance.

EXAMPLES:
  # Show link status matrix
  agent-smith link status

The output shows:
  ✓ - Valid symlink
  ◆ - Copied directory
  ✗ - Broken link
  - - Not linked
  ? - Unknown status`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleLinkStatus()
		},
	}
	linkCmd.AddCommand(linkStatusCmd)

	rootCmd.AddCommand(linkCmd)

	// Create 'unlink' parent command with subcommands
	unlinkCmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove linked components from targets",
		Long: `Remove linked components (skills, agents, commands) from detected targets.

This command provides a modern interface for unlinking downloaded AI components
from supported targets (OpenCode, Claude Code, etc.).

SAFETY:
  - Symlinks are removed immediately
  - Copied directories require confirmation before deletion
  - Source files in ~/.agents/ are never touched
  - Bulk operations (skills, agents, commands, all) prompt for confirmation unless --force is used`,
	}

	// Singular commands - operate on ONE component
	unlinkSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Unlink a specific skill from targets",
		Long: `Unlink a specific skill from detected targets.

This command removes the linked skill from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/skills/ are never touched.

EXAMPLES:
  # Unlink a specific skill
  agent-smith unlink skill mcp-builder`,
		Args: exactArgsWithHelp(1, "agent-smith unlink skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleUnlink("skills", args[0])
		},
	}
	unlinkCmd.AddCommand(unlinkSkillCmd)

	unlinkAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Unlink a specific agent from targets",
		Long: `Unlink a specific agent from detected targets.

This command removes the linked agent from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/agents/ are never touched.

EXAMPLES:
  # Unlink a specific agent
  agent-smith unlink agent coding-assistant`,
		Args: exactArgsWithHelp(1, "agent-smith unlink agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleUnlink("agents", args[0])
		},
	}
	unlinkCmd.AddCommand(unlinkAgentCmd)

	unlinkCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Unlink a specific command from targets",
		Long: `Unlink a specific command from detected targets.

This command removes the linked command from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/commands/ are never touched.

EXAMPLES:
  # Unlink a specific command
  agent-smith unlink command json-formatter`,
		Args: exactArgsWithHelp(1, "agent-smith unlink command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleUnlink("commands", args[0])
		},
	}
	unlinkCmd.AddCommand(unlinkCommandCmd)

	// Plural commands - operate on ALL components of a type
	unlinkSkillsCmd := &cobra.Command{
		Use:   "skills [name]",
		Short: "Unlink all skills from targets, or a specific skill if name provided",
		Long: `Unlink all skills from detected targets, or a specific skill if name is provided.

This command removes all linked skills from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/skills/ are never touched.

For backward compatibility, you can also provide a skill name to unlink just
that specific skill (equivalent to 'unlink skill <name>').

EXAMPLES:
  # Unlink all skills with confirmation
  agent-smith unlink skills

  # Unlink all skills without confirmation
  agent-smith unlink skills --force

  # Unlink a specific skill (backward compatibility)
  agent-smith unlink skills mcp-builder`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink skills [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			// Backward compatibility: if a name is provided, unlink that specific skill
			if len(args) == 1 {
				handleUnlink("skills", args[0])
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkType("skills", force)
			}
		},
	}
	unlinkSkillsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCmd.AddCommand(unlinkSkillsCmd)

	unlinkAgentsCmd := &cobra.Command{
		Use:   "agents [name]",
		Short: "Unlink all agents from targets, or a specific agent if name provided",
		Long: `Unlink all agents from detected targets, or a specific agent if name is provided.

This command removes all linked agents from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/agents/ are never touched.

For backward compatibility, you can also provide an agent name to unlink just
that specific agent (equivalent to 'unlink agent <name>').

EXAMPLES:
  # Unlink all agents with confirmation
  agent-smith unlink agents

  # Unlink all agents without confirmation
  agent-smith unlink agents --force

  # Unlink a specific agent (backward compatibility)
  agent-smith unlink agents coding-assistant`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink agents [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			// Backward compatibility: if a name is provided, unlink that specific agent
			if len(args) == 1 {
				handleUnlink("agents", args[0])
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkType("agents", force)
			}
		},
	}
	unlinkAgentsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCmd.AddCommand(unlinkAgentsCmd)

	unlinkCommandsCmd := &cobra.Command{
		Use:   "commands [name]",
		Short: "Unlink all commands from targets, or a specific command if name provided",
		Long: `Unlink all commands from detected targets, or a specific command if name is provided.

This command removes all linked commands from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/commands/ are never touched.

For backward compatibility, you can also provide a command name to unlink just
that specific command (equivalent to 'unlink command <name>').

EXAMPLES:
  # Unlink all commands with confirmation
  agent-smith unlink commands

  # Unlink all commands without confirmation
  agent-smith unlink commands --force

  # Unlink a specific command (backward compatibility)
  agent-smith unlink commands json-formatter`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink commands [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			// Backward compatibility: if a name is provided, unlink that specific command
			if len(args) == 1 {
				handleUnlink("commands", args[0])
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkType("commands", force)
			}
		},
	}
	unlinkCommandsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCmd.AddCommand(unlinkCommandsCmd)

	unlinkAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Unlink all components from targets",
		Long: `Unlink all components (skills, agents, and commands) from detected targets.

This command removes all linked components from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agents/ are never touched.

EXAMPLES:
  # Unlink all components with confirmation
  agent-smith unlink all

  # Unlink all components without confirmation
  agent-smith unlink all --force`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			handleUnlinkAll(force)
		},
	}
	unlinkAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCmd.AddCommand(unlinkAllCmd)

	rootCmd.AddCommand(unlinkCmd)

	// Create 'uninstall' parent command with subcommands
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed components from the system",
		Long: `Remove installed components (skills, agents, commands) from ~/.agents/.

This command removes components from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the component directory from filesystem
  3. Removing the entry from lock files

SAFETY:
  - Components are automatically unlinked before removal
  - Source directories in ~/.agents/ are permanently deleted
  - Lock file entries are removed to maintain consistency`,
	}

	// Individual component uninstall commands
	uninstallSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Remove a specific skill",
		Long: `Remove a specific skill from ~/.agents/skills/.

This command removes the skill from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the skill directory from filesystem
  3. Removing the entry from .skill-lock.json

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
	uninstallSkillCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agents/")
	uninstallCmd.AddCommand(uninstallSkillCmd)

	uninstallAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Remove a specific agent",
		Long: `Remove a specific agent from ~/.agents/agents/.

This command removes the agent from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the agent directory from filesystem
  3. Removing the entry from .agent-lock.json

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
	uninstallAgentCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agents/")
	uninstallCmd.AddCommand(uninstallAgentCmd)

	uninstallCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Remove a specific command",
		Long: `Remove a specific command from ~/.agents/commands/.

This command removes the command from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the command directory from filesystem
  3. Removing the entry from .command-lock.json

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
	uninstallCommandCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agents/")
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

	// Create 'profiles' parent command with subcommands
	profilesCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles for context switching",
		Long: `Manage profiles to switch between different sets of agents, skills, and commands.
		
Profiles allow you to organize and switch between different configurations
of AI components. Each profile can contain its own set of agents, skills,
and commands, making it easy to switch contexts for different projects or tasks.`,
	}

	profilesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available profiles",
		Long: `List all available profiles found in ~/.agents/profiles/.

This command shows all valid profiles (those containing at least one component
directory), indicates which profile is currently active, and displays component
counts for each profile.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesList()
		},
	}

	profilesShowCmd := &cobra.Command{
		Use:   "show <profile-name>",
		Short: "Show detailed information about a profile",
		Long: `Display detailed information about a specific profile.

This command shows:
  - Profile name and active status
  - Profile location on disk
  - List of all agents in the profile
  - List of all skills in the profile
  - List of all commands in the profile

Use this before activating a profile to see exactly what components it contains.

EXAMPLES:
  # Show details of a profile
  agent-smith profile show my-profile
  
  # View contents before activating
  agent-smith profile show work-profile
  agent-smith profile activate work-profile`,
		Args: exactArgsWithHelp(1, "agent-smith profile show <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesShow(args[0])
		},
	}

	profilesCreateCmd := &cobra.Command{
		Use:   "create <profile-name>",
		Short: "Create a new empty profile",
		Long: `Create a new profile with empty component directories.

This command creates a new profile directory structure at ~/.agents/profiles/<profile-name>/
with the following subdirectories:
  - agents/
  - skills/
  - commands/

After creation, you can add components to the profile and activate it with:
  agent-smith profile activate <profile-name>`,
		Args: exactArgsWithHelp(1, "agent-smith profile create <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesCreate(args[0])
		},
	}

	profilesDeleteCmd := &cobra.Command{
		Use:   "delete <profile-name>",
		Short: "Delete a profile",
		Long: `Delete a profile and all its contents.

This command permanently removes a profile directory and all components within it.
The profile must be deactivated before it can be deleted.

WARNING: This operation cannot be undone. All components in the profile will be permanently deleted.

EXAMPLES:
  # Delete a profile
  agent-smith profile delete my-profile

  # If the profile is active, deactivate it first
  agent-smith profile deactivate
  agent-smith profile delete my-profile`,
		Args: exactArgsWithHelp(1, "agent-smith profile delete <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesDelete(args[0])
		},
	}

	profilesActivateCmd := &cobra.Command{
		Use:   "activate <profile-name>",
		Short: "Activate a specific profile",
		Long: `Activate a profile without immediately affecting your editor.

This command will:
1. Update the active profile state
2. Deactivate any currently active profile

This does NOT immediately modify your editor configuration. To apply this profile
to your editor, run:
  agent-smith link all

Only one profile can be active at a time. The active profile persists across sessions.`,
		Args: exactArgsWithHelp(1, "agent-smith profile activate <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesActivate(args[0])
		},
	}

	profilesDeactivateCmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate the current profile",
		Long: `Deactivate the currently active profile without immediately affecting your editor.

This command will:
1. Clear the active profile state
2. Return to base state (no profile active)

This does NOT immediately modify your editor configuration. To apply this change
to your editor, run:
  agent-smith link all

This allows you to control when changes are applied to your editor.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesDeactivate()
		},
	}

	profilesAddCmd := &cobra.Command{
		Use:   "add <type> <profile-name> <component-name>",
		Short: "Add an existing component to a profile",
		Long: `Add an existing component from ~/.agents/ to a profile.

This command copies a component (skill, agent, or command) from your base
~/.agents/ directory to a specific profile. The component must already exist
in ~/.agents/ before it can be added to a profile.

COMPONENT TYPES:
  skills   - Copy a skill to the profile
  agents   - Copy an agent to the profile
  commands - Copy a command to the profile

EXAMPLES:
  # Add a skill to a profile
  agent-smith profile add skills my-profile gpt-skill

  # Add an agent to a profile
  agent-smith profile add agents work-profile coding-agent

  # Add a command to a profile
  agent-smith profile add commands dev-profile test-runner`,
		Args: exactArgsWithComponentTypeValidation(3, 0, "agent-smith profile add <type> <profile-name> <component-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesAdd(args[0], args[1], args[2])
		},
	}

	// profiles remove - Remove a component from a profile
	profilesRemoveCmd := &cobra.Command{
		Use:   "remove [component-type] [profile] [name]",
		Short: "Remove a component from a profile",
		Long: `Remove a component from a profile by deleting it from the profile directory.

COMPONENT TYPES:
  skills   - Remove a skill from the profile
  agents   - Remove an agent from the profile
  commands - Remove a command from the profile

EXAMPLES:
  # Remove a skill from a profile
  agent-smith profile remove skills my-profile gpt-skill

  # Remove an agent from a profile
  agent-smith profile remove agents work-profile coding-agent

  # Remove a command from a profile
  agent-smith profile remove commands dev-profile test-runner`,
		Args: exactArgsWithComponentTypeValidation(3, 0, "agent-smith profile remove <type> <profile-name> <component-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesRemove(args[0], args[1], args[2])
		},
	}

	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	profilesCmd.AddCommand(profilesActivateCmd)
	profilesCmd.AddCommand(profilesDeactivateCmd)
	profilesCmd.AddCommand(profilesAddCmd)
	profilesCmd.AddCommand(profilesRemoveCmd)
	rootCmd.AddCommand(profilesCmd)

	// Add status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status and active profile",
		Long: `Display the current configuration status including:
  - Active profile (if any)
  - Detected targets (OpenCode, Claude Code, etc.)
  - Component counts in ~/.agents/
  - Quick summary of system state

This provides a dashboard view of your agent-smith installation.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleStatus()
		},
	}
	rootCmd.AddCommand(statusCmd)

	// Create 'target' parent command with subcommands
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

	rootCmd.AddCommand(targetCmd)

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// These functions will be implemented in main.go to keep existing logic
var (
	handleAddSkill           func(repoURL, name, profile, targetDir string)
	handleAddAgent           func(repoURL, name, profile, targetDir string)
	handleAddCommand         func(repoURL, name, profile, targetDir string)
	handleAddAll             func(repoURL string, targetDir string)
	handleUpdate             func(componentType, componentName string)
	handleUpdateAll          func()
	handleLink               func(componentType, componentName, targetFilter string)
	handleLinkAll            func(targetFilter string)
	handleLinkType           func(componentType, targetFilter string)
	handleAutoLink           func()
	handleListLinks          func()
	handleLinkStatus         func()
	handleUnlink             func(componentType, componentName string)
	handleUnlinkAll          func(force bool)
	handleUnlinkType         func(componentType string, force bool)
	handleUninstall          func(componentType, componentName, profile string)
	handleUninstallAll       func(repoURL string, force bool)
	handleProfilesList       func()
	handleProfilesShow       func(profileName string)
	handleProfilesCreate     func(profileName string)
	handleProfilesDelete     func(profileName string)
	handleProfilesActivate   func(profileName string)
	handleProfilesDeactivate func()
	handleProfilesAdd        func(componentType, profileName, componentName string)
	handleProfilesRemove     func(componentType, profileName, componentName string)
	handleStatus             func()
	handleTargetAdd          func(name, path string)
)

func SetHandlers(
	addSkill func(repoURL, name, profile, targetDir string),
	addAgent func(repoURL, name, profile, targetDir string),
	addCommand func(repoURL, name, profile, targetDir string),
	addAll func(repoURL string, targetDir string),
	update func(componentType, componentName string),
	updateAll func(),
	link func(componentType, componentName, targetFilter string),
	linkAll func(targetFilter string),
	linkType func(componentType, targetFilter string),
	autoLink func(),
	listLinks func(),
	linkStatus func(),
	unlink func(componentType, componentName string),
	unlinkAll func(force bool),
	unlinkType func(componentType string, force bool),
	uninstall func(componentType, componentName, profile string),
	uninstallAll func(repoURL string, force bool),
	profilesList func(),
	profilesShow func(profileName string),
	profilesCreate func(profileName string),
	profilesDelete func(profileName string),
	profilesActivate func(profileName string),
	profilesDeactivate func(),
	profilesAdd func(componentType, profileName, componentName string),
	profilesRemove func(componentType, profileName, componentName string),
	status func(),
	targetAdd func(name, path string),
) {
	handleAddSkill = addSkill
	handleAddAgent = addAgent
	handleAddCommand = addCommand
	handleAddAll = addAll
	handleUpdate = update
	handleUpdateAll = updateAll
	handleLink = link
	handleLinkAll = linkAll
	handleLinkType = linkType
	handleAutoLink = autoLink
	handleListLinks = listLinks
	handleLinkStatus = linkStatus
	handleUnlink = unlink
	handleUnlinkAll = unlinkAll
	handleUnlinkType = unlinkType
	handleUninstall = uninstall
	handleUninstallAll = uninstallAll
	handleProfilesList = profilesList
	handleProfilesShow = profilesShow
	handleProfilesCreate = profilesCreate
	handleProfilesDelete = profilesDelete
	handleProfilesActivate = profilesActivate
	handleProfilesDeactivate = profilesDeactivate
	handleProfilesAdd = profilesAdd
	handleProfilesRemove = profilesRemove
	handleStatus = status
	handleTargetAdd = targetAdd
}
