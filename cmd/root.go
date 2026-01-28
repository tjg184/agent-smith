package cmd

import (
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

func init() {
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
  agent-smith install skill openai/cookbook gpt-skill --profile work`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleAddSkill(args[0], args[1], profile)
		},
	}
	installSkillCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
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
  agent-smith install agent openai/assistant coding-agent --profile work`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleAddAgent(args[0], args[1], profile)
		},
	}
	installAgentCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
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
  agent-smith install command cli-tools/formatter json-formatter --profile work`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleAddCommand(args[0], args[1], profile)
		},
	}
	installCommandCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agents/")
	installCmd.AddCommand(installCommandCmd)

	installCmd.AddCommand(&cobra.Command{
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
  agent-smith install all /path/to/local/repo`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddAll(args[0])
		},
	})

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
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "all" {
				handleUpdateAll()
			} else {
				if len(args) != 2 {
					cmd.PrintErrln("Error: update requires both type and name (or use 'update all')")
					os.Exit(1)
				}
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

Use 'agent-smith profiles activate <name>' to activate a profile, then run
'link' commands to apply it.

FLAGS (apply to all subcommands):
  --target, -t <target>  - Specify target to link to (opencode, claudecode, or all)
  --all-targets          - Explicitly link to all detected targets (default behavior)`,
	}

	// Add subcommands to 'link' command
	linkSkillCmd := &cobra.Command{
		Use:   "skill [name]",
		Short: "Link a skill or all skills to detected targets",
		Long: `Link a specific skill or all skills to detected targets.

This command links a downloaded skill from ~/.agents/skills/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific skill to all detected targets (default)
  agent-smith link skill mcp-builder

  # Link a specific skill to OpenCode only
  agent-smith link skill mcp-builder --target opencode

  # Link all skills to all detected targets
  agent-smith link skill

  # Link all skills to Claude Code only
  agent-smith link skill --target claudecode`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			if len(args) == 0 {
				// Link all skills
				handleLinkType("skills", targetFilter)
			} else {
				// Link specific skill
				handleLink("skills", args[0], targetFilter)
			}
		},
	}
	linkSkillCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkSkillCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkSkillCmd)

	linkAgentCmd := &cobra.Command{
		Use:   "agent [name]",
		Short: "Link an agent or all agents to detected targets",
		Long: `Link a specific agent or all agents to detected targets.

This command links a downloaded agent from ~/.agents/agents/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific agent to all detected targets (default)
  agent-smith link agent coding-assistant

  # Link a specific agent to OpenCode only
  agent-smith link agent coding-assistant --target opencode

  # Link all agents to all detected targets
  agent-smith link agent

  # Link all agents to Claude Code only
  agent-smith link agent --target claudecode`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			if len(args) == 0 {
				// Link all agents
				handleLinkType("agents", targetFilter)
			} else {
				// Link specific agent
				handleLink("agents", args[0], targetFilter)
			}
		},
	}
	linkAgentCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkAgentCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkAgentCmd)

	linkCommandCmd := &cobra.Command{
		Use:   "command [name]",
		Short: "Link a command or all commands to detected targets",
		Long: `Link a specific command or all commands to detected targets.

This command links a downloaded command from ~/.agents/commands/ to the appropriate
directories for OpenCode, Claude Code, or other supported targets.

EXAMPLES:
  # Link a specific command to all detected targets (default)
  agent-smith link command json-formatter

  # Link a specific command to OpenCode only
  agent-smith link command json-formatter --target opencode

  # Link all commands to all detected targets
  agent-smith link command

  # Link all commands to Claude Code only
  agent-smith link command --target claudecode`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			if len(args) == 0 {
				// Link all commands
				handleLinkType("commands", targetFilter)
			} else {
				// Link specific command
				handleLink("commands", args[0], targetFilter)
			}
		},
	}
	linkCommandCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkCommandCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	linkCmd.AddCommand(linkCommandCmd)

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
  1. Activate a profile: agent-smith profiles activate <name>
  2. Apply to editor: agent-smith link all

This gives you explicit control over when changes are applied to your editor.

EXAMPLES:
  # Link all components to all detected targets (default)
  agent-smith link all

  # Link all components to OpenCode only
  agent-smith link all --target opencode

  # Link all components to all targets explicitly
  agent-smith link all --all-targets`,
		Args: cobra.NoArgs,
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
		Args: cobra.NoArgs,
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
		Args: cobra.NoArgs,
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
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleLinkStatus()
		},
	}
	linkCmd.AddCommand(linkStatusCmd)

	rootCmd.AddCommand(linkCmd)

	unlinkCmd := &cobra.Command{
		Use:   "unlink <type|all> [name]",
		Short: "Remove a linked component or all components from targets",
		Long: `Remove a specific linked component or all linked components from detected targets.

USAGE:
  agent-smith unlink <type> <name>  - Unlink a specific component
  agent-smith unlink <type>         - Unlink all components of a specific type
  agent-smith unlink all            - Unlink all components (with --force to skip confirmation)

COMPONENT TYPES:
  skills    - Remove linked skills
  agents    - Remove linked agents
  commands  - Remove linked commands

SAFETY:
  - Symlinks are removed immediately
  - Copied directories require confirmation before deletion
  - Source files in ~/.agents/ are never touched
  - 'unlink all' and 'unlink <type>' prompt for confirmation unless --force is used

EXAMPLES:
  # Unlink a specific skill
  agent-smith unlink skills mcp-builder

  # Unlink all skills
  agent-smith unlink skills

  # Unlink all agents with confirmation
  agent-smith unlink agents

  # Unlink all commands without confirmation
  agent-smith unlink commands --force

  # Unlink all components with confirmation
  agent-smith unlink all

  # Unlink all components without confirmation
  agent-smith unlink all --force`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			if args[0] == "all" {
				handleUnlinkAll(force)
			} else if len(args) == 1 && isValidComponentType(args[0]) {
				handleUnlinkType(args[0], force)
			} else if len(args) == 2 {
				handleUnlink(args[0], args[1])
			} else {
				cmd.PrintErrln("Error: unlink requires type and name, or just type, or 'all'")
				os.Exit(1)
			}
		},
	}
	unlinkCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt (only for 'unlink all')")
	rootCmd.AddCommand(unlinkCmd)

	// Create 'profiles' parent command with subcommands
	profilesCmd := &cobra.Command{
		Use:   "profiles",
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
		Args: cobra.NoArgs,
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
  agent-smith profiles show my-profile
  
  # View contents before activating
  agent-smith profiles show work-profile
  agent-smith profiles activate work-profile`,
		Args: cobra.ExactArgs(1),
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
  agent-smith profiles activate <profile-name>`,
		Args: cobra.ExactArgs(1),
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
  agent-smith profiles delete my-profile

  # If the profile is active, deactivate it first
  agent-smith profiles deactivate
  agent-smith profiles delete my-profile`,
		Args: cobra.ExactArgs(1),
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
		Args: cobra.ExactArgs(1),
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
		Args: cobra.NoArgs,
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
  agent-smith profiles add skills my-profile gpt-skill

  # Add an agent to a profile
  agent-smith profiles add agents work-profile coding-agent

  # Add a command to a profile
  agent-smith profiles add commands dev-profile test-runner`,
		Args: cobra.ExactArgs(3),
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
  agent-smith profiles remove skills my-profile gpt-skill

  # Remove an agent from a profile
  agent-smith profiles remove agents work-profile coding-agent

  # Remove a command from a profile
  agent-smith profiles remove commands dev-profile test-runner`,
		Args: cobra.ExactArgs(3),
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
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleStatus()
		},
	}
	rootCmd.AddCommand(statusCmd)

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// These functions will be implemented in main.go to keep existing logic
var (
	handleAddSkill           func(repoURL, name, profile string)
	handleAddAgent           func(repoURL, name, profile string)
	handleAddCommand         func(repoURL, name, profile string)
	handleAddAll             func(repoURL string)
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
	handleProfilesList       func()
	handleProfilesShow       func(profileName string)
	handleProfilesCreate     func(profileName string)
	handleProfilesDelete     func(profileName string)
	handleProfilesActivate   func(profileName string)
	handleProfilesDeactivate func()
	handleProfilesAdd        func(componentType, profileName, componentName string)
	handleProfilesRemove     func(componentType, profileName, componentName string)
	handleStatus             func()
)

func SetHandlers(
	addSkill func(repoURL, name, profile string),
	addAgent func(repoURL, name, profile string),
	addCommand func(repoURL, name, profile string),
	addAll func(repoURL string),
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
	profilesList func(),
	profilesShow func(profileName string),
	profilesCreate func(profileName string),
	profilesDelete func(profileName string),
	profilesActivate func(profileName string),
	profilesDeactivate func(),
	profilesAdd func(componentType, profileName, componentName string),
	profilesRemove func(componentType, profileName, componentName string),
	status func(),
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
	handleProfilesList = profilesList
	handleProfilesShow = profilesShow
	handleProfilesCreate = profilesCreate
	handleProfilesDelete = profilesDelete
	handleProfilesActivate = profilesActivate
	handleProfilesDeactivate = profilesDeactivate
	handleProfilesAdd = profilesAdd
	handleProfilesRemove = profilesRemove
	handleStatus = status
}
